package transactions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	eth "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	ethrpc "github.com/ethereum/go-ethereum/rpc"

	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/rpcfilters"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

const (
	// PendingTransactionUpdate is emitted when a pending transaction is updated (added or deleted)
	EventPendingTransactionUpdate walletevent.EventType = "pending-transaction-update"
	// Caries StatusChangedPayload in message
	EventPendingTransactionStatusChanged walletevent.EventType = "pending-transaction-status-changed"

	pendingCheckInterval = 10 * time.Second
)

var (
	ErrStillPending = errors.New("transaction is still pending")
)

type TxStatus = string

// Values for status column in pending_transactions
const (
	Pending TxStatus = "Pending"
	Done    TxStatus = "Done"
)

type AutoDeleteType = bool

const (
	AutoDelete AutoDeleteType = true
	Keep       AutoDeleteType = false
)

type StatusChangedPayload struct {
	ChainID common.ChainID `json:"chainId"`
	Hash    eth.Hash       `json:"hash"`
	Status  *TxStatus      `json:"status,omitempty"`
}

type PendingTxTracker struct {
	db        *sql.DB
	rpcClient rpc.ClientInterface

	rpcFilter *rpcfilters.Service
	eventFeed *event.Feed

	taskRunner *ConditionalRepeater
}

func NewPendingTxTracker(db *sql.DB, rpcClient rpc.ClientInterface, rpcFilter *rpcfilters.Service, eventFeed *event.Feed) *PendingTxTracker {
	tm := &PendingTxTracker{
		db:        db,
		rpcClient: rpcClient,
		eventFeed: eventFeed,
		rpcFilter: rpcFilter,
	}
	tm.taskRunner = NewConditionalRepeater(pendingCheckInterval, func(ctx context.Context) bool {
		return tm.fetchTransactions(ctx)
	})
	return tm
}

type txStatusRes struct {
	// TODO - 11861: propagate real status
	Status TxStatus
	hash   eth.Hash
}

func (tm *PendingTxTracker) fetchTransactions(ctx context.Context) bool {
	res := WorkDone

	txs, err := tm.GetAllPending()
	if err != nil {
		log.Error("Failed to get pending transactions", "error", err)
		return WorkDone
	}

	txsMap := make(map[common.ChainID][]eth.Hash)
	for _, tx := range txs {
		chainID := tx.ChainID
		txsMap[chainID] = append(txsMap[chainID], tx.Hash)
	}

	// Batch request for each chain
	for chainID, txs := range txsMap {
		log.Debug("Processing pending transactions", "chainID", chainID, "count", len(txs))
		batchRes, err := fetchBatchTxStatus(ctx, tm.rpcClient, chainID, txs)
		if err != nil {
			log.Error("Failed to batch fetch pending transactions status for", "chainID", chainID, "error", err)
			continue
		}
		updateRes, err := tm.updateDBStatus(ctx, chainID, batchRes)
		if err != nil {
			log.Error("Failed to update pending transactions status for", "chainID", chainID, "error", err)
			continue
		}

		if len(updateRes) != len(batchRes) {
			res = WorkNotDone
		}

		tm.emitNotifications(chainID, updateRes)
	}

	return res
}

// fetchBatchTxStatus will exclude the still pending or errored request from the result
func fetchBatchTxStatus(ctx context.Context, rpcClient rpc.ClientInterface, chainID common.ChainID, hashes []eth.Hash) ([]txStatusRes, error) {
	chainClient, err := rpcClient.AbstractEthClient(chainID)
	if err != nil {
		log.Error("Failed to get chain client", "error", err)
		return nil, err
	}

	reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	batch := make([]ethrpc.BatchElem, 0, len(hashes))
	for _, hash := range hashes {
		jsonRes := make(map[string]interface{})
		batch = append(batch, ethrpc.BatchElem{
			Method: "eth_getTransactionByHash",
			Args:   []interface{}{hash},
			Result: &jsonRes,
		})
	}

	err = chainClient.BatchCallContext(reqCtx, batch)
	if err != nil {
		log.Error("Transactions request fail", "error", err)
		return nil, err
	}

	res := make([]txStatusRes, 0, len(batch))
	for i, b := range batch {
		isPending := true
		err := b.Error
		if err != nil {
			log.Error("Failed to get transaction", "error", err, "hash", hashes[i])
			continue
		} else {
			jsonRes := *(b.Result.(*map[string]interface{}))
			if jsonRes != nil {
				if blNo, ok := jsonRes["blockNumber"]; ok {
					isPending = blNo == nil
				}
			}
		}

		if !isPending {
			res = append(res, txStatusRes{
				hash: hashes[i],
			})
		}
	}
	return res, nil
}

// updateDBStatus returns entries that were updated only
func (tm *PendingTxTracker) updateDBStatus(ctx context.Context, chainID common.ChainID, statuses []txStatusRes) ([]eth.Hash, error) {
	res := make([]eth.Hash, 0, len(statuses))
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	updateStmt, err := tx.PrepareContext(ctx, `UPDATE pending_transactions SET status = ? WHERE network_id = ? AND hash = ?`)
	if err != nil {
		rollErr := tx.Rollback()
		if rollErr != nil {
			err = fmt.Errorf("failed to rollback transaction due to: %w", err)
		}
		return nil, fmt.Errorf("failed to prepare update statement: %w", err)
	}

	checkAutoDelStmt, err := tx.PrepareContext(ctx, `SELECT auto_delete FROM pending_transactions WHERE network_id = ? AND hash = ?`)
	if err != nil {
		rollErr := tx.Rollback()
		if rollErr != nil {
			err = fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return nil, fmt.Errorf("failed to prepare auto delete statement: %w", err)
	}

	notifyFunctions := make([]func(), 0, len(statuses))
	for _, br := range statuses {
		row := checkAutoDelStmt.QueryRowContext(ctx, chainID, br.hash)
		var autoDel bool
		err = row.Scan(&autoDel)
		if err != nil {
			if err == sql.ErrNoRows {
				log.Warn("Missing entry while checking for auto_delete", "hash", br.hash)
			} else {
				log.Error("Failed to retrieve auto_delete for pending transaction", "error", err, "hash", br.hash)
			}
			continue
		}

		if autoDel {
			notifyFn, err := tm.DeleteBySQLTx(tx, chainID, br.hash)
			if err != nil && err != ErrStillPending {
				log.Error("Failed to delete pending transaction", "error", err, "hash", br.hash)
				continue
			}
			notifyFunctions = append(notifyFunctions, notifyFn)
		} else {
			// If the entry was not deleted, update the status
			// TODO - #11861: fix status - `br.status`
			txStatus := Done

			res, err := updateStmt.ExecContext(ctx, txStatus, chainID, br.hash)
			if err != nil {
				log.Error("Failed to update pending transaction status", "error", err, "hash", br.hash)
				continue
			}
			affected, err := res.RowsAffected()
			if err != nil {
				log.Error("Failed to get updated rows", "error", err, "hash", br.hash)
				continue
			}

			if affected == 0 {
				log.Warn("Missing entry to update for", "hash", br.hash)
				continue
			}
		}

		res = append(res, br.hash)
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	for _, fn := range notifyFunctions {
		fn()
	}

	return res, nil
}

func (tm *PendingTxTracker) emitNotifications(chainID common.ChainID, changes []eth.Hash) {
	if tm.eventFeed != nil {
		for _, hash := range changes {
			status := StatusChangedPayload{
				ChainID: chainID,
				Hash:    hash,
				// TODO - #11861: status
			}

			jsonPayload, err := json.Marshal(status)
			if err != nil {
				log.Error("Failed to marshal pending transaction status", "error", err, "hash", hash)
				continue
			}
			tm.eventFeed.Send(walletevent.Event{
				Type:    EventPendingTransactionStatusChanged,
				ChainID: uint64(chainID),
				Message: string(jsonPayload),
			})
		}
	}
}

// PendingTransaction called with autoDelete = false will keep the transaction in the database until it is confirmed by the caller using Delete
func (tm *PendingTxTracker) TrackPendingTransaction(chainID common.ChainID, hash eth.Hash, from eth.Address, trType PendingTrxType, autoDelete AutoDeleteType) error {
	err := tm.addPending(&PendingTransaction{
		ChainID:    chainID,
		Hash:       hash,
		From:       from,
		Timestamp:  uint64(time.Now().Unix()),
		Type:       trType,
		AutoDelete: &autoDelete,
	})
	if err != nil {
		return err
	}

	tm.taskRunner.RunUntilDone()

	return nil
}

func (tm *PendingTxTracker) Start() error {
	tm.taskRunner.RunUntilDone()
	return nil
}

// APIs returns a list of new APIs.
func (tm *PendingTxTracker) APIs() []ethrpc.API {
	return []ethrpc.API{
		{
			Namespace: "pending",
			Version:   "0.1.0",
			Service:   tm,
			Public:    true,
		},
	}
}

// Protocols returns a new protocols list. In this case, there are none.
func (tm *PendingTxTracker) Protocols() []p2p.Protocol {
	return []p2p.Protocol{}
}

func (tm *PendingTxTracker) Stop() error {
	tm.taskRunner.Stop()
	return nil
}

type PendingTrxType string

const (
	RegisterENS               PendingTrxType = "RegisterENS"
	ReleaseENS                PendingTrxType = "ReleaseENS"
	SetPubKey                 PendingTrxType = "SetPubKey"
	BuyStickerPack            PendingTrxType = "BuyStickerPack"
	WalletTransfer            PendingTrxType = "WalletTransfer"
	DeployCommunityToken      PendingTrxType = "DeployCommunityToken"
	AirdropCommunityToken     PendingTrxType = "AirdropCommunityToken"
	RemoteDestructCollectible PendingTrxType = "RemoteDestructCollectible"
	BurnCommunityToken        PendingTrxType = "BurnCommunityToken"
	DeployOwnerToken          PendingTrxType = "DeployOwnerToken"
)

type PendingTransaction struct {
	Hash               eth.Hash       `json:"hash"`
	Timestamp          uint64         `json:"timestamp"`
	Value              bigint.BigInt  `json:"value"`
	From               eth.Address    `json:"from"`
	To                 eth.Address    `json:"to"`
	Data               string         `json:"data"`
	Symbol             string         `json:"symbol"`
	GasPrice           bigint.BigInt  `json:"gasPrice"`
	GasLimit           bigint.BigInt  `json:"gasLimit"`
	Type               PendingTrxType `json:"type"`
	AdditionalData     string         `json:"additionalData"`
	ChainID            common.ChainID `json:"network_id"`
	MultiTransactionID int64          `json:"multi_transaction_id"`

	// nil will insert the default value (Pending) in DB
	Status *TxStatus `json:"status,omitempty"`
	// nil will insert the default value (true) in DB
	AutoDelete *bool `json:"autoDelete,omitempty"`
}

const selectFromPending = `SELECT hash, timestamp, value, from_address, to_address, data,
								symbol, gas_price, gas_limit, type, additional_data,
								network_id, COALESCE(multi_transaction_id, 0), status, auto_delete
							FROM pending_transactions
							`

func rowsToTransactions(rows *sql.Rows) (transactions []*PendingTransaction, err error) {
	for rows.Next() {
		transaction := &PendingTransaction{
			Value:    bigint.BigInt{Int: new(big.Int)},
			GasPrice: bigint.BigInt{Int: new(big.Int)},
			GasLimit: bigint.BigInt{Int: new(big.Int)},
		}

		transaction.Status = new(TxStatus)
		transaction.AutoDelete = new(bool)
		err := rows.Scan(&transaction.Hash,
			&transaction.Timestamp,
			(*bigint.SQLBigIntBytes)(transaction.Value.Int),
			&transaction.From,
			&transaction.To,
			&transaction.Data,
			&transaction.Symbol,
			(*bigint.SQLBigIntBytes)(transaction.GasPrice.Int),
			(*bigint.SQLBigIntBytes)(transaction.GasLimit.Int),
			&transaction.Type,
			&transaction.AdditionalData,
			&transaction.ChainID,
			&transaction.MultiTransactionID,
			transaction.Status,
			transaction.AutoDelete,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (tm *PendingTxTracker) GetAllPending() ([]*PendingTransaction, error) {
	log.Debug("Getting all pending transactions")

	rows, err := tm.db.Query(selectFromPending+"WHERE status = ?", Pending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToTransactions(rows)
}

func (tm *PendingTxTracker) GetPendingByAddress(chainIDs []uint64, address eth.Address) ([]*PendingTransaction, error) {
	log.Debug("Getting pending transaction by address", "chainIDs", chainIDs, "address", address)

	if len(chainIDs) == 0 {
		return nil, errors.New("at least 1 chainID is required")
	}

	inVector := strings.Repeat("?, ", len(chainIDs)-1) + "?"
	var parameters []interface{}
	for _, c := range chainIDs {
		parameters = append(parameters, c)
	}

	parameters = append(parameters, address)

	rows, err := tm.db.Query(fmt.Sprintf(selectFromPending+"WHERE network_id in (%s) AND from_address = ?", inVector), parameters...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToTransactions(rows)
}

// GetPendingEntry returns sql.ErrNoRows if no pending transaction is found for the given identity
func (tm *PendingTxTracker) GetPendingEntry(chainID common.ChainID, hash eth.Hash) (*PendingTransaction, error) {
	log.Debug("Getting pending transaction", "chainID", chainID, "hash", hash)

	rows, err := tm.db.Query(selectFromPending+"WHERE network_id = ? AND hash = ?", chainID, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	trs, err := rowsToTransactions(rows)
	if err != nil {
		return nil, err
	}

	if len(trs) == 0 {
		return nil, sql.ErrNoRows
	}
	return trs[0], nil
}

// StoreAndTrackPendingTx store the details of a pending transaction and track it until it is mined
func (tm *PendingTxTracker) StoreAndTrackPendingTx(transaction *PendingTransaction) error {
	err := tm.addPending(transaction)
	if err != nil {
		return err
	}

	tm.taskRunner.RunUntilDone()

	return err
}

func (tm *PendingTxTracker) addPending(transaction *PendingTransaction) error {
	insert, err := tm.db.Prepare(`INSERT OR REPLACE INTO pending_transactions
                                      (network_id, hash, timestamp, value, from_address, to_address,
                                       data, symbol, gas_price, gas_limit, type, additional_data, multi_transaction_id, status, auto_delete)
                                      VALUES
                                      (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? , ?)`)
	if err != nil {
		return err
	}
	_, err = insert.Exec(
		transaction.ChainID,
		transaction.Hash,
		transaction.Timestamp,
		(*bigint.SQLBigIntBytes)(transaction.Value.Int),
		transaction.From,
		transaction.To,
		transaction.Data,
		transaction.Symbol,
		(*bigint.SQLBigIntBytes)(transaction.GasPrice.Int),
		(*bigint.SQLBigIntBytes)(transaction.GasLimit.Int),
		transaction.Type,
		transaction.AdditionalData,
		transaction.MultiTransactionID,
		transaction.Status,
		transaction.AutoDelete,
	)
	// Notify listeners of new pending transaction (used in activity history)
	if err == nil {
		tm.notifyPendingTransactionListeners(transaction.ChainID, []eth.Address{transaction.From, transaction.To}, transaction.Timestamp)
	}
	if tm.rpcFilter != nil {
		tm.rpcFilter.TriggerTransactionSentToUpstreamEvent(&rpcfilters.PendingTxInfo{
			Hash:    transaction.Hash,
			Type:    string(transaction.Type),
			From:    transaction.From,
			ChainID: uint64(transaction.ChainID),
		})
	}
	return err
}

func (tm *PendingTxTracker) notifyPendingTransactionListeners(chainID common.ChainID, addresses []eth.Address, timestamp uint64) {
	if tm.eventFeed != nil {
		tm.eventFeed.Send(walletevent.Event{
			Type:     EventPendingTransactionUpdate,
			ChainID:  uint64(chainID),
			Accounts: addresses,
			At:       int64(timestamp),
		})
	}
}

// DeleteBySQLTx returns ErrStillPending if the transaction is still pending
func (tm *PendingTxTracker) DeleteBySQLTx(tx *sql.Tx, chainID common.ChainID, hash eth.Hash) (notify func(), err error) {
	row := tx.QueryRow(`SELECT from_address, to_address, timestamp, status FROM pending_transactions WHERE network_id = ? AND hash = ?`, chainID, hash)
	var from, to eth.Address
	var timestamp uint64
	var status TxStatus
	err = row.Scan(&from, &to, &timestamp, &status)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`DELETE FROM pending_transactions WHERE network_id = ? AND hash = ?`, chainID, hash)
	if err != nil {
		return nil, err
	}

	if err == nil && status == Pending {
		err = ErrStillPending
	}
	return func() {
		tm.notifyPendingTransactionListeners(chainID, []eth.Address{from, to}, timestamp)
	}, err
}

func GetTransferData(tx *sql.Tx, chainID common.ChainID, hash eth.Hash) (txType *PendingTrxType, MTID *int64, err error) {
	row := tx.QueryRow(`SELECT type, multi_transaction_id FROM pending_transactions WHERE network_id = ? AND hash = ?`, chainID, hash, txType)
	txType = new(PendingTrxType)
	MTID = new(int64)
	err = row.Scan(txType, MTID)
	if err != nil {
		return nil, nil, err
	}
	return txType, MTID, nil
}

// Watch returns sql.ErrNoRows if no pending transaction is found for the given identity
// tx.Status is not nill if err is nil
func (tm *PendingTxTracker) Watch(ctx context.Context, chainID common.ChainID, hash eth.Hash) (*TxStatus, error) {
	log.Debug("Watching transaction", "chainID", chainID, "hash", hash)

	tx, err := tm.GetPendingEntry(chainID, hash)
	if err != nil {
		return nil, err
	}

	return tx.Status, nil
}

// Delete returns ErrStillPending if the deleted transaction was still pending
// The transactions are suppose to be deleted by the client only after they are confirmed
func (tm *PendingTxTracker) Delete(ctx context.Context, chainID common.ChainID, transactionHash eth.Hash) error {
	log.Debug("Delete pending transaction to confirm it", "chainID", chainID, "hash", transactionHash)

	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	notifyFn, err := tm.DeleteBySQLTx(tx, chainID, transactionHash)
	if err != nil && err != ErrStillPending {
		rollErr := tx.Rollback()
		if rollErr != nil {
			return fmt.Errorf("failed to rollback transaction due to error: %w", err)
		}
		return err
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction: %w", commitErr)
	}
	notifyFn()
	return err
}
package transfer

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/account"
	"github.com/status-im/status-go/eth-node/types"
	"github.com/status-im/status-go/multiaccounts/accounts"
	"github.com/status-im/status-go/params"
	"github.com/status-im/status-go/rpc"
	"github.com/status-im/status-go/services/wallet/bigint"
	"github.com/status-im/status-go/services/wallet/bridge"
	wallet_common "github.com/status-im/status-go/services/wallet/common"
	"github.com/status-im/status-go/transactions"
)

type MultiTransactionIDType int64

const (
	NoMultiTransactionID = MultiTransactionIDType(0)
)

type TransactionManager struct {
	db             *sql.DB
	gethManager    *account.GethManager
	transactor     *transactions.Transactor
	config         *params.NodeConfig
	accountsDB     *accounts.Database
	pendingManager *transactions.TransactionManager
	// rpcFilterSrvc  *rpcfilters.Service
}

func NewTransactionManager(db *sql.DB, gethManager *account.GethManager, transactor *transactions.Transactor,
	config *params.NodeConfig, accountsDB *accounts.Database,
	pendingTxManager *transactions.TransactionManager) *TransactionManager {
	// rpcFiltersSrvc *rpcfilters.Service) *TransactionManager {
	// pendingTxManager *transactions.TransactionManager) *TransactionManager {

	return &TransactionManager{
		db:             db,
		gethManager:    gethManager,
		transactor:     transactor,
		config:         config,
		accountsDB:     accountsDB,
		pendingManager: pendingTxManager,
		// rpcFilterSrvc:  rpcFiltersSrvc,
	}
}

type MultiTransactionType uint8

// TODO: extend with know types
const (
	MultiTransactionSend = iota
	MultiTransactionSwap
	MultiTransactionBridge
)

type MultiTransaction struct {
	ID          uint                 `json:"id"`
	Timestamp   uint64               `json:"timestamp"`
	FromAddress common.Address       `json:"fromAddress"`
	ToAddress   common.Address       `json:"toAddress"`
	FromAsset   string               `json:"fromAsset"`
	ToAsset     string               `json:"toAsset"`
	FromAmount  *hexutil.Big         `json:"fromAmount"`
	ToAmount    *hexutil.Big         `json:"toAmount"`
	Type        MultiTransactionType `json:"type"`
}

type MultiTransactionCommand struct {
	FromAddress common.Address       `json:"fromAddress"`
	ToAddress   common.Address       `json:"toAddress"`
	FromAsset   string               `json:"fromAsset"`
	ToAsset     string               `json:"toAsset"`
	FromAmount  *hexutil.Big         `json:"fromAmount"`
	Type        MultiTransactionType `json:"type"`
}

type MultiTransactionCommandResult struct {
	ID     int64                   `json:"id"`
	Hashes map[uint64][]types.Hash `json:"hashes"`
}

type TransactionIdentity struct {
	ChainID wallet_common.ChainID `json:"chainId"`
	Hash    common.Hash           `json:"hash"`
	Address common.Address        `json:"address"`
}

const multiTransactionColumns = "from_address, from_asset, from_amount, to_address, to_asset, to_amount, type, timestamp"

func insertMultiTransaction(db *sql.DB, multiTransaction *MultiTransaction) (MultiTransactionIDType, error) {
	insert, err := db.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO multi_transactions (%s)
											VALUES(?, ?, ?, ?, ?, ?, ?, ?)`, multiTransactionColumns))
	if err != nil {
		return NoMultiTransactionID, err
	}
	result, err := insert.Exec(
		multiTransaction.FromAddress,
		multiTransaction.FromAsset,
		multiTransaction.FromAmount.String(),
		multiTransaction.ToAddress,
		multiTransaction.ToAsset,
		multiTransaction.ToAmount.String(),
		multiTransaction.Type,
		time.Now().Unix(),
	)
	if err != nil {
		return NoMultiTransactionID, err
	}
	defer insert.Close()
	multiTransactionID, err := result.LastInsertId()
	return MultiTransactionIDType(multiTransactionID), err
}

func (tm *TransactionManager) InsertMultiTransaction(multiTransaction *MultiTransaction) (MultiTransactionIDType, error) {
	return insertMultiTransaction(tm.db, multiTransaction)
}

func (tm *TransactionManager) CreateMultiTransactionFromCommand(ctx context.Context, command *MultiTransactionCommand,
	data []*bridge.TransactionBridge, bridges map[string]bridge.Bridge, password string, rpcClient *rpc.Client) (
	*MultiTransactionCommandResult, error) {

	multiTransaction := multiTransactionFromCommand(command)

	multiTransactionID, err := insertMultiTransaction(tm.db, multiTransaction)
	if err != nil {
		return nil, err
	}

	hashes, err := tm.sendTransactions(multiTransaction, data, bridges, password, rpcClient)
	if err != nil {
		return nil, err
	}

	err = tm.storePendingTransactions(multiTransactionID, multiTransaction, hashes, data)
	if err != nil {
		return nil, err
	}

	return &MultiTransactionCommandResult{
		ID:     int64(multiTransactionID),
		Hashes: hashes,
	}, nil
}

func (tm *TransactionManager) storePendingTransactions(multiTransactionID MultiTransactionIDType,
	multiTransaction *MultiTransaction, hashes map[uint64][]types.Hash, data []*bridge.TransactionBridge) error {

	txs := createPendingTransactions(hashes, data, multiTransactionID, multiTransaction)
	for _, tx := range txs {
		err := tm.pendingManager.AddPending(tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func createPendingTransactions(hashes map[uint64][]types.Hash, data []*bridge.TransactionBridge,
	multiTransactionID MultiTransactionIDType, multiTransaction *MultiTransaction) []*transactions.PendingTransaction {

	txs := make([]*transactions.PendingTransaction, 0)
	for _, tx := range data {
		for _, hash := range hashes[tx.ChainID] {
			pendingTransaction := &transactions.PendingTransaction{
				Hash:               common.Hash(hash),
				Timestamp:          uint64(time.Now().Unix()),
				Value:              bigint.BigInt{Int: multiTransaction.FromAmount.ToInt()},
				From:               common.Address(tx.From()),
				To:                 common.Address(tx.To()),
				Data:               tx.Data().String(),
				Type:               transactions.WalletTransfer,
				ChainID:            tx.ChainID,
				MultiTransactionID: int64(multiTransactionID),
				Symbol:             multiTransaction.FromAsset,
			}
			txs = append(txs, pendingTransaction)
		}
	}
	return txs
}

func multiTransactionFromCommand(command *MultiTransactionCommand) *MultiTransaction {

	log.Info("Creating multi transaction", "command", command)

	multiTransaction := &MultiTransaction{
		FromAddress: command.FromAddress,
		ToAddress:   command.ToAddress,
		FromAsset:   command.FromAsset,
		ToAsset:     command.ToAsset,
		FromAmount:  command.FromAmount,
		ToAmount:    new(hexutil.Big),
		Type:        command.Type,
	}

	return multiTransaction
}

func (tm *TransactionManager) sendTransactions(multiTransaction *MultiTransaction,
	data []*bridge.TransactionBridge, bridges map[string]bridge.Bridge, password string, rpcClient *rpc.Client) (
	map[uint64][]types.Hash, error) {

	log.Info("Making transactions", "multiTransaction", multiTransaction)

	selectedAccount, err := tm.getVerifiedWalletAccount(multiTransaction.FromAddress.Hex(), password)
	if err != nil {
		return nil, err
	}

	hashes := make(map[uint64][]types.Hash)
	for _, tx := range data {
		hash, err := bridges[tx.BridgeName].Send(tx, selectedAccount)
		if err != nil {
			return nil, err
		}
		hashes[tx.ChainID] = append(hashes[tx.ChainID], hash)
	}
	return hashes, nil
}

func (tm *TransactionManager) GetMultiTransactions(ctx context.Context, ids []MultiTransactionIDType) ([]*MultiTransaction, error) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, v := range ids {
		placeholders[i] = "?"
		args[i] = v
	}

	stmt, err := tm.db.Prepare(fmt.Sprintf(`SELECT rowid, %s
											FROM multi_transactions
											WHERE rowid in (%s)`,
		multiTransactionColumns,
		strings.Join(placeholders, ",")))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var multiTransactions []*MultiTransaction
	for rows.Next() {
		multiTransaction := &MultiTransaction{}
		var fromAmount string
		var toAmount string
		err := rows.Scan(
			&multiTransaction.ID,
			&multiTransaction.FromAddress,
			&multiTransaction.FromAsset,
			&fromAmount,
			&multiTransaction.ToAddress,
			&multiTransaction.ToAsset,
			&toAmount,
			&multiTransaction.Type,
			&multiTransaction.Timestamp,
		)
		if err != nil {
			return nil, err
		}

		multiTransaction.FromAmount = new(hexutil.Big)
		_, ok := (*big.Int)(multiTransaction.FromAmount).SetString(fromAmount, 0)
		if !ok {
			return nil, errors.New("failed to convert fromAmount to big.Int: " + fromAmount)
		}

		multiTransaction.ToAmount = new(hexutil.Big)
		_, ok = (*big.Int)(multiTransaction.ToAmount).SetString(toAmount, 0)
		if !ok {
			return nil, errors.New("failed to convert toAmount to big.Int: " + toAmount)
		}

		multiTransactions = append(multiTransactions, multiTransaction)
	}

	return multiTransactions, nil
}

func (tm *TransactionManager) getVerifiedWalletAccount(address, password string) (*account.SelectedExtKey, error) {
	exists, err := tm.accountsDB.AddressExists(types.HexToAddress(address))
	if err != nil {
		log.Error("failed to query db for a given address", "address", address, "error", err)
		return nil, err
	}

	if !exists {
		log.Error("failed to get a selected account", "err", transactions.ErrInvalidTxSender)
		return nil, transactions.ErrAccountDoesntExist
	}

	key, err := tm.gethManager.VerifyAccountPassword(tm.config.KeyStoreDir, address, password)
	if err != nil {
		log.Error("failed to verify account", "account", address, "error", err)
		return nil, err
	}

	return &account.SelectedExtKey{
		Address:    key.Address,
		AccountKey: key,
	}, nil
}

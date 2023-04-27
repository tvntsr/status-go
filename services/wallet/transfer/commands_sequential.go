package transfer

import (
	"context"
	"database/sql"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/services/wallet/async"
	"github.com/status-im/status-go/services/wallet/walletevent"
)

type ethHistoricalSingleAccCommand struct {
	// db           *Database
	eth          Downloader
	address      common.Address
	chainClient  *chain.ClientWithFallback
	balanceCache *balanceCache
	feed         *event.Feed
	foundHeaders []*DBHeader
	error        error
	noLimit      bool

	from                         *Block
	to, resultingFrom, stopBlock *big.Int
}

func (c *ethHistoricalSingleAccCommand) Command() async.Command {
	return async.FiniteCommand{
		Interval: 5 * time.Second,
		Runable:  c.Run,
	}.Run
}

func (c *ethHistoricalSingleAccCommand) Run(ctx context.Context) (err error) {
	log.Info("eth historical single acc downloader starting", "address", c.address, "from", c.from.Number, "to", c.to, "noLimit", c.noLimit)

	start := time.Now()
	if c.from.Number != nil && c.from.Balance != nil {
		c.balanceCache.addBalanceToCache(c.address, c.from.Number, c.from.Balance)
	}
	if c.from.Number != nil && c.from.Nonce != nil {
		c.balanceCache.addNonceToCache(c.address, c.from.Number, c.from.Nonce)
	}
	from, headers, stopBlock, err := findBlocksWithEthTransfers2(ctx, c.chainClient, c.balanceCache, c.eth, c.address, c.from.Number, c.to, c.noLimit)

	if err != nil {
		c.error = err
		// return nil // TODO FIND OUT WHY SLOW if return err!!!
		log.Error("failed to find blocks with transfers", "error", err)
		return err
	}

	c.foundHeaders = headers
	c.resultingFrom = from
	c.stopBlock = stopBlock

	log.Info("eth historical single acc downloader finished successfully", "address", c.address,
		"from", from, "to", c.to, "total blocks", len(headers), "time", time.Since(start), "stopBlock", stopBlock)

	// err = c.db.ProcessBlocks(c.address, from, c.to, headers, ethTransfer)
	// if err != nil {
	// 	log.Error("failed to save found blocks with transfers", "error", err)
	// 	return err
	// }
	log.Debug("eth transfers were persisted. command is closed")
	return nil
}

type findBlocksCommand struct {
	account       common.Address
	db            *Database
	chainClient   *chain.ClientWithFallback
	balanceCache  *balanceCache
	feed          *event.Feed
	fromBlock     *Block
	toBlockNumber *big.Int
	foundHeaders  []*DBHeader
	noLimit       bool
	error         error
	resFromBlock  *Block
	stopBlock     *big.Int
}

func (c *findBlocksCommand) Command() async.Command {
	return async.FiniteCommand{
		Interval: 5 * time.Second,
		Runable:  c.Run,
	}.Run
}

func (c *findBlocksCommand) Run(parent context.Context) (err error) {
	log.Info("start findBlocksCommand")

	fromBlock := c.fromBlock

	// Continue from last block
	if c.resFromBlock != nil {
		c.toBlockNumber = c.resFromBlock.Number
	}

	newFromBlock, ethHeaders, stopBlock, err := c.fastIndex(parent, c.balanceCache, fromBlock, c.toBlockNumber)
	log.Info("findBlocksCommand", "stopBlock", stopBlock)
	if err != nil {
		c.error = err
		// return err // In case c.noLimit is true, hystrix "max concurrency" may be reached and we will not be able to index ETH transfers
		return nil
	}
	if c.noLimit {
		// newFromByAddress = map[common.Address]*big.Int{}
		newFromBlock = c.fromBlock // TODO: #10246 check if it is correct
	}
	erc20Headers, err := c.fastIndexErc20(parent, newFromBlock.Number, c.toBlockNumber)
	// erc20Headers, err := []*DBHeader{}, nil
	if err != nil {
		c.error = err
		// return err
		return nil
	}

	allHeaders := append(ethHeaders, erc20Headers...)

	uniqHeadersByHash := map[common.Hash]*DBHeader{}
	for _, header := range allHeaders {
		uniqHeader, ok := uniqHeadersByHash[header.Hash]
		if ok {
			if len(header.Erc20Transfers) > 0 {
				uniqHeader.Erc20Transfers = append(uniqHeader.Erc20Transfers, header.Erc20Transfers...)
			}
			uniqHeadersByHash[header.Hash] = uniqHeader
		} else {
			uniqHeadersByHash[header.Hash] = header
		}
	}

	uniqHeaders := []*DBHeader{}
	for _, header := range uniqHeadersByHash {
		uniqHeaders = append(uniqHeaders, header)
	}

	foundHeaders := uniqHeaders

	maxBlockNumber := big.NewInt(0)
	for _, header := range allHeaders {
		if header.Number.Cmp(maxBlockNumber) == 1 {
			maxBlockNumber = header.Number
		}
	}

	lastBlockNumber := c.toBlockNumber
	log.Info("saving headers", "len", len(uniqHeaders), "lastBlockNumber", lastBlockNumber,
		"balance", c.balanceCache.ReadCachedBalance(c.account, lastBlockNumber),
		"nonce", c.balanceCache.ReadCachedNonce(c.account, lastBlockNumber))
	to := &Block{
		Number:  lastBlockNumber,
		Balance: c.balanceCache.ReadCachedBalance(c.account, lastBlockNumber),
		Nonce:   c.balanceCache.ReadCachedNonce(c.account, lastBlockNumber),
	}
	// err = c.db.ProcessBlocks(c.chainClient.ChainID, address, newFromByAddress[address], to, uniqHeaders)
	err = c.db.ProcessBlocks(c.chainClient.ChainID, c.account, newFromBlock.Number, to, uniqHeaders)
	if err != nil {
		return err
	}

	sort.SliceStable(foundHeaders, func(i, j int) bool {
		return foundHeaders[i].Number.Cmp(foundHeaders[j].Number) == 1
	})
	c.foundHeaders = foundHeaders
	c.resFromBlock = newFromBlock
	c.stopBlock = stopBlock

	log.Info("end findBlocksCommand", "c.stopBlock", c.stopBlock, "stopBlock", stopBlock)
	return
}

// run fast indexing for every accont up to canonical chain head minus safety depth.
// every account will run it from last synced header.
func (c *findBlocksCommand) fastIndex(ctx context.Context, bCache *balanceCache,
	// fromByAddress map[common.Address]*Block, toByAddress map[common.Address]*big.Int) (map[common.Address]*big.Int,
	fromBlock *Block, toBlockNumber *big.Int) (*Block, []*DBHeader, *big.Int, error) {

	log.Info("fast indexer started", "accounts", c.account, "from", fromBlock.Number, "to", toBlockNumber)

	start := time.Now()
	group := async.NewGroup(ctx)

	command := &ethHistoricalSingleAccCommand{
		// db:           c.db,
		chainClient:  c.chainClient,
		balanceCache: bCache,
		address:      c.account,
		eth: &ETHDownloader{
			chainClient: c.chainClient,
			accounts:    []common.Address{c.account},
			signer:      types.NewLondonSigner(c.chainClient.ToBigInt()),
			db:          c.db,
		},
		feed:    c.feed,
		from:    fromBlock,
		to:      toBlockNumber,
		noLimit: c.noLimit,
	}
	group.Add(command.Command())

	select {
	case <-ctx.Done():
		return nil, nil, command.stopBlock, ctx.Err()
	case <-group.WaitAsync():
		if command.error != nil {
			return nil, nil, command.stopBlock, command.error
		}
		resultingFrom := &Block{Number: command.resultingFrom}
		headers := command.foundHeaders
		log.Info("fast indexer finished", "in", time.Since(start), "stopBlock", command.stopBlock)
		return resultingFrom, headers, command.stopBlock, nil
	}
}

// run fast indexing for every accont up to canonical chain head minus safety depth.
// every account will run it from last synced header.
func (c *findBlocksCommand) fastIndexErc20(ctx context.Context, fromBlockNumber *big.Int,
	toBlockNumber *big.Int) ([]*DBHeader, error) {
	// func (c *findBlocksCommand) fastIndexErc20(ctx context.Context, fromByAddress map[common.Address]*Block, toByAddress map[common.Address]*big.Int) (map[common.Address][]*DBHeader, error) {
	start := time.Now()
	group := async.NewGroup(ctx)

	erc20 := &erc20HistoricalCommand{
		erc20:       NewERC20TransfersDownloader(c.chainClient, []common.Address{c.account}, types.NewLondonSigner(c.chainClient.ToBigInt())),
		chainClient: c.chainClient,
		feed:        c.feed,
		address:     c.account,
		// from:        fromByAddress[address].Number,
		from:         fromBlockNumber,
		to:           toBlockNumber,
		foundHeaders: []*DBHeader{},
	}
	group.Add(erc20.Command())
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-group.WaitAsync():
		headers := erc20.foundHeaders
		log.Info("fast indexer Erc20 finished", "in", time.Since(start))
		return headers, nil
	}
}

type loadAllTransfersCommand struct {
	accounts []common.Address
	db       *Database
	// blockDAO           *BlockRangeSequentialDAO
	blockDAO           *BlockDAO
	eth                *ETHDownloader
	erc20              *ERC20TransfersDownloader
	chainClient        *chain.ClientWithFallback
	feed               *event.Feed
	errorsCount        int
	nonArchivalRPCNode bool
	transactionManager *TransactionManager
}

func (c *loadAllTransfersCommand) Run(parent context.Context) error {
	log.Info("start load all transfers command")
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	head, err := c.chainClient.HeaderByNumber(ctx, nil)
	cancel()
	if err != nil {
		// if c.NewError(err) {
		// 	return nil
		// }
		return err
	}

	if c.feed != nil {
		c.feed.Send(walletevent.Event{
			Type:     EventFetchingRecentHistory,
			Accounts: c.accounts,
		})
	}

	log.Info("current head is", "block number", head.Number)
	// lastKnownEthBlocks, err := c.blockDAO.getLastKnownBlocks(c.chainClient.ChainID, c.accounts)
	// if err != nil {
	// 	log.Error("failed to load last head from database", "error", err)
	// 	// if c.NewError(err) {
	// 	// 	return nil
	// 	// }
	// 	return err
	// }

	// firstKnownEthBlocks, err := c.blockDAO.getFirstKnownBlocks(c.chainClient.ChainID, c.accounts)
	// if err != nil {
	// 	log.Error("failed to load first block from database", "error", err)
	// 	// if c.NewError(err) {
	// 	// 	return nil
	// 	// }
	// 	return err
	// }

	// fromMap := map[common.Address]*big.Int{}
	//
	// if !c.nonArchivalRPCNode {
	// 	fromMap, err = findFirstRanges(parent, accountsWithoutHistory, head.Number, c.chainClient)
	// 	if err != nil {
	// 		if c.NewError(err) {
	// 			return nil
	// 		}
	// 		return err
	// 	}
	// }

	// target := head.Number
	// fromByAddress := map[common.Address]*Block{}
	// toByAddress := map[common.Address]*big.Int{}

	// for _, address := range c.accounts {
	// 	from, _ := lastKnownEthBlocks[address]
	// 	// from, ok := lastKnownEthBlocks[address]
	// 	// TODO get last known block for accounts without history, but only a last block
	// 	// if !ok {
	// 	// 	from = &LastBlock{Number: fromMap[address]}
	// 	// }
	// 	if c.nonArchivalRPCNode {
	// 		from = &Block{Number: big.NewInt(0).Sub(target, big.NewInt(100))} // Last 100 blocks? Why?
	// 	}

	// 	fromByAddress[address] = from
	// 	toByAddress[address] = target
	// }

	// bCache := newBalanceCache()
	// for _, address := range c.accounts {
	// cmnd := &findBlocksCommand{
	// 	accounts:      c.accounts,
	// 	db:            c.db,
	// 	chainClient:   c.chainClient,
	// 	balanceCache:  bCache,
	// 	feed:          c.feed,
	// 	fromByAddress: fromByAddress,
	// 	toByAddress:   toByAddress,
	// }

	// err = cmnd.Command()(parent)
	// if err != nil {
	// 	if c.NewError(err) {
	// 		return nil
	// 	}
	// 	return err
	// }

	// if cmnd.error != nil {
	// 	if c.NewError(cmnd.error) {
	// 		return nil
	// 	}
	// 	return cmnd.error
	// }

	// loadTransfers2(ctx, c.accounts, c.blockDAO, c.db, c.chainClient,
	// 	firstKnownEthBlocks, lastKnownEthBlocks, c.transactionManager)

	// _, err = c.LoadTransfers(parent, 40)
	// if err != nil {
	// 	if c.NewError(err) {
	// 		return nil
	// 	}
	// 	return err
	// }

	// if c.feed != nil {
	// 	events := map[common.Address]walletevent.Event{}
	// 	for _, address := range c.accounts {
	// 		event := walletevent.Event{
	// 			Type:     EventNewTransfers,
	// 			Accounts: []common.Address{address},
	// 		}
	// 		for _, header := range cmnd.foundHeaders[address] {
	// 			if event.BlockNumber == nil || header.Number.Cmp(event.BlockNumber) == 1 {
	// 				event.BlockNumber = header.Number
	// 			}
	// 		}
	// 		if event.BlockNumber != nil {
	// 			events[address] = event
	// 		}
	// 	}

	// 	for _, event := range events {
	// 		c.feed.Send(event)
	// 	}

	// 	c.feed.Send(walletevent.Event{
	// 		Type:        EventRecentHistoryReady,
	// 		Accounts:    c.accounts,
	// 		BlockNumber: target,
	// 	})
	// }

	log.Info("end loadAllTransfers command")
	return err
}

// TODO REMOVE IT????
func (c *loadAllTransfersCommand) Command() async.Command {
	return async.InfiniteCommand{
		Interval: 5 * time.Second,
		Runable:  c.Run,
	}.Run
}

type transfersCommand2 struct {
	db                 *Database
	eth                *ETHDownloader
	blockNum           *big.Int
	address            common.Address
	chainClient        *chain.ClientWithFallback
	fetchedTransfers   []Transfer
	transactionManager *TransactionManager
}

func (c *transfersCommand2) Command() async.Command {
	return async.FiniteCommand{
		Interval: 5 * time.Second,
		Runable:  c.Run,
	}.Run
}

func (c *transfersCommand2) Run(ctx context.Context) (err error) {
	// log.Info("transfersCommands2 start", "address", c.address, "blockNum", c.blockNum)

	// startTs := time.Now()

	allTransfers, err := c.eth.GetTransfersByNumber(ctx, c.blockNum)
	if err != nil {
		log.Warn("GetTransfersByNumber error", "error", err)
		return err
	}

	// Update MultiTransactionID from pending entry
	for index := range allTransfers {
		transfer := &allTransfers[index]
		if transfer.MultiTransactionID == NoMultiTransactionID {
			entry, err := c.transactionManager.GetPendingEntry(c.chainClient.ChainID, transfer.ID)
			if err != nil {
				if err == sql.ErrNoRows {
					// log.Info("Pending transaction not found for", "chainID", c.chainClient.ChainID, "transferID", transfer.ID)
				} else {
					return err
				}
			} else {
				transfer.MultiTransactionID = entry.MultiTransactionID
				if transfer.Receipt != nil && transfer.Receipt.Status == types.ReceiptStatusSuccessful {
					// TODO: Nim logic was deleting pending previously, should we notify UI about it?
					err := c.transactionManager.DeletePending(c.chainClient.ChainID, transfer.ID)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if len(allTransfers) > 0 {
		err = c.db.SaveTransfersOnly(c.chainClient.ChainID, c.address, allTransfers)
		if err != nil {
			log.Error("SaveTransfers error", "error", err)
			return err
		}

		err = markBlockHeadersAsLoaded(c.chainClient.ChainID, c.db, c.address, []*DBHeader{{Number: c.blockNum}})
		if err != nil {
			log.Error("markHeadersAsLoaded error", "error", err)
			return
		}

		// return
	}

	c.fetchedTransfers = allTransfers
	// log.Info("transfers loaded", "address", c.address, "len", len(allTransfers), "in", time.Since(startTs))
	// log.Info("transfers loaded", "in", time.Since(startTs))
	return nil
}

func loadTransfers2(ctx context.Context, account common.Address, blockDAO *BlockRangeSequentialDAO, db *Database,
	// chainClient *chain.ClientWithFallback, blocks []*big.Int,
	chainClient *chain.ClientWithFallback, headers []*DBHeader,
	transactionManager *TransactionManager) ([]Transfer, error) {

	// log.Info("loadTransfers2 start", "account", account, "headers len", len(headers))
	start := time.Now()
	group := async.NewGroup(ctx)

	commands := []*transfersCommand2{}
	// for _, address := range accounts {
	// blocks, ok := blocksByAddress[address]

	// if !ok {
	// 	startBlockTs := time.Now()

	// 	// TODO Need to fetch only oldest loaded block, to be able to go backwards
	// 	// block, _ := blockDAO.GetFirstKnownBlock(chainClient.ChainID, address)
	// 	log.Info("loadTransfers blocks found for", "address", address, "in", time.Since(startBlockTs))
	// }

	for _, block := range headers {
		transfers := &transfersCommand2{
			db:          db,
			chainClient: chainClient,
			address:     account,
			eth: &ETHDownloader{
				chainClient: chainClient,
				accounts:    []common.Address{account},
				signer:      types.NewLondonSigner(chainClient.ToBigInt()),
				db:          db,
			},
			blockNum:           block.Number,
			transactionManager: transactionManager,
		}
		commands = append(commands, transfers)
		group.Add(transfers.Command())
	}
	// log.Info("loadTransfers all blocks found", "in", time.Since(start))
	// allBlocksFoundTs := time.Now()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-group.WaitAsync():
		allTransfers := []Transfer{}
		for _, command := range commands {
			if len(command.fetchedTransfers) == 0 {
				continue
			}

			allTransfers = append(allTransfers, command.fetchedTransfers...)
		}
		// log.Info("loadTransfers transfers only finished", "in", time.Since(allBlocksFoundTs))
		log.Info("loadTransfers complete finished", "in", time.Since(start), "transfers", len(allTransfers))
		return allTransfers, nil
	}
}

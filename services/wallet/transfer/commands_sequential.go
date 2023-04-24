package transfer

import (
	"context"
	"math/big"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/services/wallet/async"
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
	log.Info("eth historical single acc downloader starting", "address", c.address)

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
		// return nil
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

// type erc20HistoricalCommand struct {
// 	db          *Database
// 	erc20       BatchDownloader
// 	address     common.Address
// 	chainClient *chain.ClientWithFallback
// 	feed        *event.Feed

// 	iterator     *IterativeDownloader
// 	to           *big.Int
// 	from         *big.Int
// 	foundHeaders []*DBHeader
// }

// func (c *erc20HistoricalCommand) Command() async.Command {
// 	return async.FiniteCommand{
// 		Interval: 5 * time.Second,
// 		Runable:  c.Run,
// 	}.Run
// }

// func getErc20BatchSize(chainID uint64) *big.Int {
// 	if isBinanceChain(chainID) {
// 		return binanceChainErc20BatchSize
// 	}

// 	if chainID == goerliChainID {
// 		return goerliErc20BatchSize
// 	}

// 	if chainID == goerliArbitrumChainID {
// 		return goerliErc20ArbitrumBatchSize
// 	}

// 	return erc20BatchSize
// }

// func (c *erc20HistoricalCommand) Run(ctx context.Context) (err error) {
// 	start := time.Now()
// 	if c.iterator == nil {
// 		c.iterator, err = SetupIterativeDownloader(
// 			c.db, c.chainClient, c.address,
// 			c.erc20, getErc20BatchSize(c.chainClient.ChainID), c.to, c.from)
// 		if err != nil {
// 			log.Error("failed to setup historical downloader for erc20")
// 			return err
// 		}
// 	}
// 	for !c.iterator.Finished() {
// 		headers, _, _, err := c.iterator.Next(ctx)
// 		if err != nil {
// 			log.Error("failed to get next batch", "error", err)
// 			return err
// 		}
// 		c.foundHeaders = append(c.foundHeaders, headers...)

// 		/*err = c.db.ProcessBlocks(c.address, from, to, headers, erc20Transfer)
// 		if err != nil {
// 			c.iterator.Revert()
// 			log.Error("failed to save downloaded erc20 blocks with transfers", "error", err)
// 			return err
// 		}*/
// 	}
// 	log.Info("wallet historical downloader for erc20 transfers finished", "in", time.Since(start))
// 	return nil
// }

// // controlCommand implements following procedure (following parts are executed sequeantially):
// // - verifies that the last header that was synced is still in the canonical chain
// // - runs fast indexing for each account separately
// // - starts listening to new blocks and watches for reorgs
// type controlCommand struct {
// 	accounts           []common.Address
// 	db                 *Database
// 	blockDAO           *BlockDAO
// 	eth                *ETHDownloader
// 	erc20              *ERC20TransfersDownloader
// 	chainClient        *chain.ClientWithFallback
// 	feed               *event.Feed
// 	errorsCount        int
// 	nonArchivalRPCNode bool
// 	transactionManager *TransactionManager
// }

// // controlCommand implements following procedure (following parts are executed sequeantially):
// // - verifies that the last header that was synced is still in the canonical chain
// // - runs fast indexing for each account separately
// // - starts listening to new blocks and watches for reorgs
// type loadAllTransfersCommand struct {
// 	accounts           []common.Address
// 	db                 *Database
// 	blockDAO           *BlockRangeSequentialDAO
// 	eth                *ETHDownloader
// 	erc20              *ERC20TransfersDownloader
// 	chainClient        *chain.ClientWithFallback
// 	feed               *event.Feed
// 	errorsCount        int
// 	nonArchivalRPCNode bool
// 	transactionManager *TransactionManager
// }

// func (c *loadAllTransfersCommand) Run(parent context.Context) error {
// 	log.Info("start load all transfers command")
// 	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
// 	head, err := c.chainClient.HeaderByNumber(ctx, nil)
// 	cancel()
// 	if err != nil {
// 		// if c.NewError(err) {
// 		// 	return nil
// 		// }
// 		return err
// 	}

// 	if c.feed != nil {
// 		c.feed.Send(walletevent.Event{
// 			Type:     EventFetchingRecentHistory,
// 			Accounts: c.accounts,
// 		})
// 	}

// 	log.Info("current head is", "block number", head.Number)
// 	lastKnownEthBlocks, err := c.blockDAO.getLastKnownBlocks(c.chainClient.ChainID, c.accounts)
// 	if err != nil {
// 		log.Error("failed to load last head from database", "error", err)
// 		// if c.NewError(err) {
// 		// 	return nil
// 		// }
// 		return err
// 	}

// 	firstKnownEthBlocks, err := c.blockDAO.getFirstKnownBlocks(c.chainClient.ChainID, c.accounts)
// 	if err != nil {
// 		log.Error("failed to load first block from database", "error", err)
// 		// if c.NewError(err) {
// 		// 	return nil
// 		// }
// 		return err
// 	}

// 	// fromMap := map[common.Address]*big.Int{}
// 	//
// 	// if !c.nonArchivalRPCNode {
// 	// 	fromMap, err = findFirstRanges(parent, accountsWithoutHistory, head.Number, c.chainClient)
// 	// 	if err != nil {
// 	// 		if c.NewError(err) {
// 	// 			return nil
// 	// 		}
// 	// 		return err
// 	// 	}
// 	// }

// 	// target := head.Number
// 	// fromByAddress := map[common.Address]*Block{}
// 	// toByAddress := map[common.Address]*big.Int{}

// 	// for _, address := range c.accounts {
// 	// 	from, _ := lastKnownEthBlocks[address]
// 	// 	// from, ok := lastKnownEthBlocks[address]
// 	// 	// TODO get last known block for accounts without history, but only a last block
// 	// 	// if !ok {
// 	// 	// 	from = &LastBlock{Number: fromMap[address]}
// 	// 	// }
// 	// 	if c.nonArchivalRPCNode {
// 	// 		from = &Block{Number: big.NewInt(0).Sub(target, big.NewInt(100))} // Last 100 blocks? Why?
// 	// 	}

// 	// 	fromByAddress[address] = from
// 	// 	toByAddress[address] = target
// 	// }

// 	// bCache := newBalanceCache()
// 	// cmnd := &findBlocksCommand{
// 	// 	accounts:      c.accounts,
// 	// 	db:            c.db,
// 	// 	chainClient:   c.chainClient,
// 	// 	balanceCache:  bCache,
// 	// 	feed:          c.feed,
// 	// 	fromByAddress: fromByAddress,
// 	// 	toByAddress:   toByAddress,
// 	// }

// 	// err = cmnd.Command()(parent)
// 	// if err != nil {
// 	// 	if c.NewError(err) {
// 	// 		return nil
// 	// 	}
// 	// 	return err
// 	// }

// 	// if cmnd.error != nil {
// 	// 	if c.NewError(cmnd.error) {
// 	// 		return nil
// 	// 	}
// 	// 	return cmnd.error
// 	// }

// 	loadTransfers2(ctx, c.accounts, c.blockDAO, c.db, c.chainClient,
// 		firstKnownEthBlocks, lastKnownEthBlocks, c.transactionManager)
// 	// _, err = c.LoadTransfers(parent, 40)
// 	// if err != nil {
// 	// 	if c.NewError(err) {
// 	// 		return nil
// 	// 	}
// 	// 	return err
// 	// }

// 	// if c.feed != nil {
// 	// 	events := map[common.Address]walletevent.Event{}
// 	// 	for _, address := range c.accounts {
// 	// 		event := walletevent.Event{
// 	// 			Type:     EventNewTransfers,
// 	// 			Accounts: []common.Address{address},
// 	// 		}
// 	// 		for _, header := range cmnd.foundHeaders[address] {
// 	// 			if event.BlockNumber == nil || header.Number.Cmp(event.BlockNumber) == 1 {
// 	// 				event.BlockNumber = header.Number
// 	// 			}
// 	// 		}
// 	// 		if event.BlockNumber != nil {
// 	// 			events[address] = event
// 	// 		}
// 	// 	}

// 	// 	for _, event := range events {
// 	// 		c.feed.Send(event)
// 	// 	}

// 	// 	c.feed.Send(walletevent.Event{
// 	// 		Type:        EventRecentHistoryReady,
// 	// 		Accounts:    c.accounts,
// 	// 		BlockNumber: target,
// 	// 	})
// 	// }

// 	log.Info("end loadAllTransfers command")
// 	return err
// }

// // TODO REMOVE IT????
// func (c *loadAllTransfersCommand) Command() async.Command {
// 	return async.InfiniteCommand{
// 		Interval: 5 * time.Second,
// 		Runable:  c.Run,
// 	}.Run
// }

// func (c *controlCommand) LoadTransfers(ctx context.Context, limit int) (map[common.Address][]Transfer, error) {
// 	return loadTransfers(ctx, c.accounts, c.blockDAO, c.db, c.chainClient, limit, make(map[common.Address][]*big.Int), c.transactionManager)
// }

// // TODO #10246 Rename controlCommand to something meaningful like loadTransfersCommand
// func (c *controlCommand) Run(parent context.Context) error {
// 	log.Info("start control command")
// 	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
// 	head, err := c.chainClient.HeaderByNumber(ctx, nil)
// 	cancel()
// 	if err != nil {
// 		if c.NewError(err) {
// 			return nil
// 		}
// 		return err
// 	}

// 	if c.feed != nil {
// 		c.feed.Send(walletevent.Event{
// 			Type:     EventFetchingRecentHistory,
// 			Accounts: c.accounts,
// 		})
// 	}

// 	log.Info("current head is", "block number", head.Number)

// 	// Get last known block for each account
// 	// lastKnownEthBlocks, accountsWithoutHistory, err := c.blockDAO.GetLastKnownBlockByAddresses(c.chainClient.ChainID, c.accounts)
// 	lastKnownEthBlocks, _, err := c.blockDAO.GetLastKnownBlockByAddresses(c.chainClient.ChainID, c.accounts)
// 	if err != nil {
// 		log.Error("failed to load last head from database", "error", err)
// 		if c.NewError(err) {
// 			return nil
// 		}
// 		return err
// 	}

// 	// For accounts without history, find the block where 20 < headNonce - nonce < 25 (blocks have between 20-25 transactions)
// 	// fromMap := map[common.Address]*big.Int{}

// 	// if !c.nonArchivalRPCNode {
// 	// 	fromMap, err = findFirstRanges(parent, accountsWithoutHistory, head.Number, c.chainClient)
// 	// 	if err != nil {
// 	// 		if c.NewError(err) {
// 	// 			return nil
// 	// 		}
// 	// 		return err
// 	// 	}
// 	// }

// 	// Set "fromByAddress" from the information we have
// 	target := head.Number
// 	fromByAddress := map[common.Address]*Block{}
// 	toByAddress := map[common.Address]*big.Int{}

// 	for _, address := range c.accounts {
// 		from, ok := lastKnownEthBlocks[address]
// 		if !ok {
// 			// from = &Block{Number: fromMap[address]}
// 			from = &Block{Number: big.NewInt(0)}
// 		}
// 		// if c.nonArchivalRPCNode {
// 		// 	from = &Block{Number: big.NewInt(0).Sub(target, big.NewInt(100))}
// 		// }

// 		fromByAddress[address] = from
// 		toByAddress[address] = target
// 	}

// 	log.Info("loading blocks", "from", fromByAddress[c.accounts[0]].Number, "to", toByAddress[c.accounts[0]])

// 	bCache := newBalanceCache()
// 	cmnd := &findBlocksCommand{
// 		accounts:      c.accounts,
// 		db:            c.db,
// 		chainClient:   c.chainClient,
// 		balanceCache:  bCache,
// 		feed:          c.feed,
// 		fromByAddress: fromByAddress,
// 		toByAddress:   toByAddress,
// 	}

// 	err = cmnd.Command()(parent)
// 	if err != nil {
// 		if c.NewError(err) {
// 			return nil
// 		}
// 		return err
// 	}

// 	if cmnd.error != nil {
// 		if c.NewError(cmnd.error) {
// 			return nil
// 		}
// 		return cmnd.error
// 	}

// 	_, err = c.LoadTransfers(parent, 40)
// 	if err != nil {
// 		if c.NewError(err) {
// 			return nil
// 		}
// 		return err
// 	}

// 	if c.feed != nil {
// 		events := map[common.Address]walletevent.Event{}
// 		for _, address := range c.accounts {
// 			event := walletevent.Event{
// 				Type:     EventNewTransfers,
// 				Accounts: []common.Address{address},
// 			}
// 			for _, header := range cmnd.foundHeaders[address] {
// 				if event.BlockNumber == nil || header.Number.Cmp(event.BlockNumber) == 1 {
// 					event.BlockNumber = header.Number
// 				}
// 			}
// 			if event.BlockNumber != nil {
// 				events[address] = event
// 			}
// 		}

// 		for _, event := range events {
// 			c.feed.Send(event)
// 		}

// 		c.feed.Send(walletevent.Event{
// 			Type:        EventRecentHistoryReady,
// 			Accounts:    c.accounts,
// 			BlockNumber: target,
// 		})
// 	}

// 	log.Info("end control command")
// 	return err
// }

// func nonArchivalNodeError(err error) bool {
// 	return strings.Contains(err.Error(), "missing trie node") ||
// 		strings.Contains(err.Error(), "project ID does not have access to archive state")
// }

// func (c *controlCommand) NewError(err error) bool {
// 	c.errorsCount++
// 	log.Error("controlCommand error", "error", err, "counter", c.errorsCount)
// 	if nonArchivalNodeError(err) {
// 		log.Info("Non archival node detected")
// 		c.nonArchivalRPCNode = true
// 		c.feed.Send(walletevent.Event{
// 			Type: EventNonArchivalNodeDetected,
// 		})
// 	}
// 	if c.errorsCount >= 3 {
// 		c.feed.Send(walletevent.Event{
// 			Type:    EventFetchingHistoryError,
// 			Message: err.Error(),
// 		})
// 		return true
// 	}
// 	return false
// }

// func (c *controlCommand) Command() async.Command {
// 	return async.FiniteCommand{
// 		Interval: 5 * time.Second,
// 		Runable:  c.Run,
// 	}.Run
// }

// type transfersCommand2 struct {
// 	db                 *Database
// 	eth                *ETHDownloader
// 	firstKnownBlock    *Block
// 	lastKnownBlock     *Block
// 	address            common.Address
// 	chainClient        *chain.ClientWithFallback
// 	fetchedTransfers   []Transfer
// 	transactionManager *TransactionManager
// }

// func (c *transfersCommand2) Command() async.Command {
// 	return async.FiniteCommand{
// 		Interval: 5 * time.Second,
// 		Runable:  c.Run,
// 	}.Run
// }

// func (c *transfersCommand2) Run(ctx context.Context) (err error) {
// 	startTs := time.Now()

// 	// allTransfers, err := getTransfersByBlocks(ctx, c.db, c.eth, c.blocks)
// 	// if err != nil {
// 	// 	log.Info("getTransfersByBlocks error", "error", err)
// 	// 	return err
// 	// }

// 	// // Update MultiTransactionID from pending entry
// 	// for index := range allTransfers {
// 	// 	transfer := &allTransfers[index]
// 	// 	if transfer.MultiTransactionID == NoMultiTransactionID {
// 	// 		entry, err := c.transactionManager.GetPendingEntry(c.chainClient.ChainID, transfer.ID)
// 	// 		if err != nil {
// 	// 			if err == sql.ErrNoRows {
// 	// 				// log.Info("Pending transaction not found for", "chainID", c.chainClient.ChainID, "transferID", transfer.ID)
// 	// 			} else {
// 	// 				return err
// 	// 			}
// 	// 		} else {
// 	// 			transfer.MultiTransactionID = entry.MultiTransactionID
// 	// 			if transfer.Receipt != nil && transfer.Receipt.Status == types.ReceiptStatusSuccessful {
// 	// 				// TODO: Nim logic was deleting pending previously, should we notify UI about it?
// 	// 				err := c.transactionManager.DeletePending(c.chainClient.ChainID, transfer.ID)
// 	// 				if err != nil {
// 	// 					return err
// 	// 				}
// 	// 			}
// 	// 		}
// 	// 	}
// 	// }

// 	// if len(allTransfers) > 0 {
// 	// 	err = c.db.SaveTransfers(c.chainClient.ChainID, c.address, allTransfers, c.blocks)
// 	// 	if err != nil {
// 	// 		log.Error("SaveTransfers error", "error", err)
// 	// 		return err
// 	// 	}
// 	// }

// 	// c.fetchedTransfers = allTransfers
// 	// log.Debug("transfers loaded", "address", c.address, "len", len(allTransfers), "in", time.Since(startTs))
// 	log.Debug("transfers loaded", "in", time.Since(startTs))
// 	return nil
// }

// type transfersCommand struct {
// 	db                 *Database
// 	eth                *ETHDownloader
// 	block              *big.Int
// 	address            common.Address
// 	chainClient        *chain.ClientWithFallback
// 	fetchedTransfers   []Transfer
// 	transactionManager *TransactionManager
// }

// func (c *transfersCommand) Command() async.Command {
// 	return async.FiniteCommand{
// 		Interval: 5 * time.Second,
// 		Runable:  c.Run,
// 	}.Run
// }

// func (c *transfersCommand) Run(ctx context.Context) (err error) {
// 	startTs := time.Now()

// 	allTransfers, err := getTransfersByBlocks(ctx, c.db, c.eth, []*big.Int{c.block})
// 	if err != nil {
// 		log.Info("getTransfersByBlocks error", "error", err)
// 		return err
// 	}

// 	// Update MultiTransactionID from pending entry
// 	for index := range allTransfers {
// 		transfer := &allTransfers[index]
// 		if transfer.MultiTransactionID == NoMultiTransactionID {
// 			entry, err := c.transactionManager.GetPendingEntry(c.chainClient.ChainID, transfer.ID)
// 			if err != nil {
// 				if err == sql.ErrNoRows {
// 					// log.Info("Pending transaction not found for", "chainID", c.chainClient.ChainID, "transferID", transfer.ID)
// 				} else {
// 					return err
// 				}
// 			} else {
// 				transfer.MultiTransactionID = entry.MultiTransactionID
// 				if transfer.Receipt != nil && transfer.Receipt.Status == types.ReceiptStatusSuccessful {
// 					// TODO: Nim logic was deleting pending previously, should we notify UI about it?
// 					err := c.transactionManager.DeletePending(c.chainClient.ChainID, transfer.ID)
// 					if err != nil {
// 						return err
// 					}
// 				}
// 			}
// 		}
// 	}

// 	if len(allTransfers) > 0 {
// 		err = c.db.SaveTransfers(c.chainClient.ChainID, c.address, allTransfers, []*big.Int{c.block})
// 		if err != nil {
// 			log.Error("SaveTransfers error", "error", err)
// 			return err
// 		}
// 	}

// 	c.fetchedTransfers = allTransfers
// 	log.Debug("transfers loaded", "address", c.address, "len", len(allTransfers), "in", time.Since(startTs))
// 	return nil
// }

// // type loadTransfersCommand struct {
// // 	accounts                []common.Address
// // 	db                      *Database
// // 	block                   *Block
// // 	chainClient             *chain.ClientWithFallback
// // 	blocksByAddress         map[common.Address][]*big.Int
// // 	foundTransfersByAddress map[common.Address][]Transfer
// // 	transactionManager      *TransactionManager
// // }

// // func (c *loadTransfersCommand) Command() async.Command {
// // 	return async.FiniteCommand{
// // 		Interval: 5 * time.Second,
// // 		Runable:  c.Run,
// // 	}.Run
// // }

// // func (c *loadTransfersCommand) LoadTransfers(ctx context.Context, limit int, blocksByAddress map[common.Address][]*big.Int, transactionManager *TransactionManager) (map[common.Address][]Transfer, error) {
// // 	return loadTransfers(ctx, c.accounts, c.block, c.db, c.chainClient, limit, blocksByAddress, c.transactionManager)
// // }

// // func (c *loadTransfersCommand) Run(parent context.Context) (err error) {
// // 	transfersByAddress, err := c.LoadTransfers(parent, 40, c.blocksByAddress, c.transactionManager)
// // 	if err != nil {
// // 		return err
// // 	}
// // 	c.foundTransfersByAddress = transfersByAddress

// // 	return
// // }

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
		return err // In case c.noLimit is true, hystrix "max concurrency" may be reached and we will not be able to index ETH transfers
	}
	if c.noLimit {
		// newFromByAddress = map[common.Address]*big.Int{}
		newFromBlock = c.fromBlock // TODO: #10246 check if it is correct
	}
	erc20Headers, err := c.fastIndexErc20(parent, newFromBlock.Number, c.toBlockNumber)
	// erc20Headers, err := []*DBHeader{}, nil
	if err != nil {
		c.error = err
		return err
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

// func loadTransfers(ctx context.Context, accounts []common.Address, blockDAO *BlockDAO, db *Database,
// 	chainClient *chain.ClientWithFallback, limit int, blocksByAddress map[common.Address][]*big.Int,
// 	transactionManager *TransactionManager) (map[common.Address][]Transfer, error) {

// 	log.Info("loadTransfers start", "accounts", accounts, "block", blockDAO, "limit", limit)
// 	start := time.Now()
// 	group := async.NewGroup(ctx)

// 	commands := []*transfersCommand{}
// 	for _, address := range accounts {
// 		blocks, ok := blocksByAddress[address]

// 		if !ok {
// 			startBlockTs := time.Now()
// 			blocks, _ = blockDAO.GetBlocksByAddress(chainClient.ChainID, address, numberOfBlocksCheckedPerIteration)
// 			log.Info("loadTransfers blocks found for", "address", address, "in", time.Since(startBlockTs))
// 		}
// 		for _, block := range blocks {
// 			transfers := &transfersCommand{
// 				db:          db,
// 				chainClient: chainClient,
// 				address:     address,
// 				eth: &ETHDownloader{
// 					chainClient: chainClient,
// 					accounts:    []common.Address{address},
// 					signer:      types.NewLondonSigner(chainClient.ToBigInt()),
// 					db:          db,
// 				},
// 				block:              block,
// 				transactionManager: transactionManager,
// 			}
// 			commands = append(commands, transfers)
// 			group.Add(transfers.Command())
// 		}
// 	}
// 	log.Info("loadTransfers all blocks found", "in", time.Since(start))
// 	allBlocksFoundTs := time.Now()

// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	case <-group.WaitAsync():
// 		transfersByAddress := map[common.Address][]Transfer{}
// 		for _, command := range commands {
// 			if len(command.fetchedTransfers) == 0 {
// 				continue
// 			}

// 			transfers, ok := transfersByAddress[command.address]
// 			if !ok {
// 				transfers = []Transfer{}
// 			}

// 			for _, transfer := range command.fetchedTransfers {
// 				transfersByAddress[command.address] = append(transfers, transfer)
// 			}
// 		}
// 		log.Info("loadTransfers transfers found finished", "in", time.Since(allBlocksFoundTs))
// 		log.Info("loadTransfers finished", "in", time.Since(start))
// 		return transfersByAddress, nil
// 	}
// }

// func loadTransfers2(ctx context.Context, accounts []common.Address, blockDAO *BlockRangeSequentialDAO, db *Database,
// 	chainClient *chain.ClientWithFallback, firstKnownBlocks map[common.Address]*Block,
// 	lastKnownBlocks map[common.Address]*Block, transactionManager *TransactionManager) (map[common.Address][]Transfer, error) {

// 	log.Info("loadTransfers start", "accounts", accounts)
// 	start := time.Now()
// 	group := async.NewGroup(ctx)

// 	commands := []*transfersCommand2{}
// 	for _, address := range accounts {
// 		// blocks, ok := blocksByAddress[address]

// 		// if !ok {
// 		// 	startBlockTs := time.Now()

// 		// 	// TODO Need to fetch only oldest loaded block, to be able to go backwards
// 		// 	// block, _ := blockDAO.GetFirstKnownBlock(chainClient.ChainID, address)
// 		// 	log.Info("loadTransfers blocks found for", "address", address, "in", time.Since(startBlockTs))
// 		// }

// 		// for _, block := range blocks {
// 		transfers := &transfersCommand2{
// 			db:          db,
// 			chainClient: chainClient,
// 			address:     address,
// 			eth: &ETHDownloader{
// 				chainClient: chainClient,
// 				accounts:    []common.Address{address},
// 				signer:      types.NewLondonSigner(chainClient.ToBigInt()),
// 				db:          db,
// 			},
// 			firstKnownBlock:    firstKnownBlocks[address],
// 			lastKnownBlock:     lastKnownBlocks[address],
// 			transactionManager: transactionManager,
// 		}
// 		commands = append(commands, transfers)
// 		group.Add(transfers.Command())
// 		// }
// 	}
// 	log.Info("loadTransfers all blocks found", "in", time.Since(start))
// 	allBlocksFoundTs := time.Now()

// 	select {
// 	case <-ctx.Done():
// 		return nil, ctx.Err()
// 	case <-group.WaitAsync():
// 		transfersByAddress := map[common.Address][]Transfer{}
// 		for _, command := range commands {
// 			if len(command.fetchedTransfers) == 0 {
// 				continue
// 			}

// 			transfers, ok := transfersByAddress[command.address]
// 			if !ok {
// 				transfers = []Transfer{}
// 			}

// 			for _, transfer := range command.fetchedTransfers {
// 				transfersByAddress[command.address] = append(transfers, transfer)
// 			}
// 		}
// 		log.Info("loadTransfers transfers found finished", "in", time.Since(allBlocksFoundTs))
// 		log.Info("loadTransfers finished", "in", time.Since(start))
// 		return transfersByAddress, nil
// 	}
// }

// func isBinanceChain(chainID uint64) bool {
// 	return chainID == binancChainID || chainID == binanceTestChainID
// }

// func getLowestFrom(chainID uint64, to *big.Int) *big.Int {
// 	from := big.NewInt(0)
// 	if isBinanceChain(chainID) && big.NewInt(0).Sub(to, from).Cmp(binanceChainMaxInitialRange) == 1 {
// 		from = big.NewInt(0).Sub(to, binanceChainMaxInitialRange)
// 	}

// 	return from
// }

// // Finds the latest range up to initialTo where the number of transactions is between 20 and 25
// func findFirstRange(c context.Context, account common.Address, initialTo *big.Int, client *chain.ClientWithFallback) (*big.Int, error) {
// 	from := getLowestFrom(client.ChainID, initialTo)
// 	to := initialTo
// 	goal := uint64(20)

// 	if from.Cmp(to) == 0 {
// 		return to, nil
// 	}

// 	firstNonce, err := client.NonceAt(c, account, to) // this is the latest nonce actually
// 	log.Info("find range with 20 <= len(tx) <= 25", "account", account, "firstNonce", firstNonce, "from", from, "to", to)

// 	if err != nil {
// 		return nil, err
// 	}

// 	if firstNonce <= goal {
// 		return from, nil
// 	}

// 	nonceDiff := firstNonce
// 	iterations := 0
// 	for iterations < 50 {
// 		iterations = iterations + 1

// 		if nonceDiff > goal {
// 			// from = (from + to) / 2
// 			from = from.Add(from, to)
// 			from = from.Div(from, big.NewInt(2))
// 		} else {
// 			// from = from - (to - from) / 2
// 			// to = from
// 			diff := big.NewInt(0).Sub(to, from)
// 			diff.Div(diff, big.NewInt(2))
// 			to = big.NewInt(from.Int64())
// 			from.Sub(from, diff)
// 		}
// 		fromNonce, err := client.NonceAt(c, account, from)
// 		if err != nil {
// 			return nil, err
// 		}
// 		nonceDiff = firstNonce - fromNonce

// 		log.Info("next nonce", "from", from, "n", fromNonce, "diff", firstNonce-fromNonce)

// 		if goal <= nonceDiff && nonceDiff <= (goal+5) {
// 			log.Info("range found", "account", account, "from", from, "to", to)
// 			return from, nil
// 		}
// 	}

// 	log.Info("range found", "account", account, "from", from, "to", to)

// 	return from, nil
// }

// // Finds the latest ranges up to initialTo where the number of transactions is between 20 and 25
// func findFirstRanges(c context.Context, accounts []common.Address, initialTo *big.Int, client *chain.ClientWithFallback) (map[common.Address]*big.Int, error) {
// 	res := map[common.Address]*big.Int{}

// 	for _, address := range accounts {
// 		from, err := findFirstRange(c, address, initialTo, client)
// 		if err != nil {
// 			return nil, err
// 		}

// 		res[address] = from
// 	}

// 	return res, nil
// }

// func getTransfersByBlocks(ctx context.Context, db *Database, downloader *ETHDownloader, blocks []*big.Int) ([]Transfer, error) {
// 	allTransfers := []Transfer{}

// 	for _, block := range blocks {
// 		transfers, err := downloader.GetTransfersByNumber(ctx, block)
// 		if err != nil {
// 			return nil, err
// 		}
// 		log.Debug("loadTransfers", "block", block, "new transfers", len(transfers))
// 		if len(transfers) > 0 {
// 			allTransfers = append(allTransfers, transfers...)
// 		}
// 	}

// 	return allTransfers, nil
// }

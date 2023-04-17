package transfer

import (
	"context"
	"errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/services/wallet/async"
)

var errAlreadyRunning = errors.New("already running")

// HeaderReader interface for reading headers using block number or hash.
type HeaderReader interface {
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

// BalanceReader interface for reading balance at a specifeid address.
type BalanceReader interface {
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type HistoryFetcher interface {
	start() error
	stop()

	setChainClients(chainClients map[uint64]*chain.ClientWithFallback)
	setAccounts(accounts []common.Address)
}

type DefaultFetchStrategy struct {
	db                 *Database
	blockDAO           *BlockDAO
	feed               *event.Feed
	mu                 sync.Mutex
	group              *async.Group
	transactionManager *TransactionManager
	chainClients       map[uint64]*chain.ClientWithFallback
	accounts           []common.Address
}

func (r *DefaultFetchStrategy) newControlCommand(chainClient *chain.ClientWithFallback, accounts []common.Address) *controlCommand {
	signer := types.NewLondonSigner(chainClient.ToBigInt())
	ctl := &controlCommand{
		db:          r.db,
		chainClient: chainClient,
		accounts:    accounts,
		blockDAO:    r.blockDAO,
		eth: &ETHDownloader{
			chainClient: chainClient,
			accounts:    accounts,
			signer:      signer,
			db:          r.db,
		},
		erc20:              NewERC20TransfersDownloader(chainClient, accounts, signer),
		feed:               r.feed,
		errorsCount:        0,
		transactionManager: r.transactionManager,
	}

	return ctl
}

func (s *DefaultFetchStrategy) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.group != nil {
		return errAlreadyRunning
	}
	s.group = async.NewGroup(context.Background())

	for _, chainClient := range s.chainClients {
		ctl := s.newControlCommand(chainClient, s.accounts)
		s.group.Add(ctl.Command())
	}

	return nil
}

// Stop stops reactor loop and waits till it exits.
func (s *DefaultFetchStrategy) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.group == nil {
		return
	}
	s.group.Stop()
	s.group.Wait()
	s.group = nil
}

func (s *DefaultFetchStrategy) setChainClients(chainClients map[uint64]*chain.ClientWithFallback) {
	s.chainClients = chainClients
}

func (s *DefaultFetchStrategy) setAccounts(accounts []common.Address) {
	s.accounts = accounts
}

type SequentialFetchStrategy struct {
	db                 *Database
	blockDAO           *BlockRangeSequentialDAO
	feed               *event.Feed
	mu                 sync.Mutex
	group              *async.Group
	transactionManager *TransactionManager
	chainClients       map[uint64]*chain.ClientWithFallback
	accounts           []common.Address
}

func (r *SequentialFetchStrategy) newCommand(chainClient *chain.ClientWithFallback, accounts []common.Address) *loadAllTransfersCommand {
	signer := types.NewLondonSigner(chainClient.ToBigInt())
	ctl := &loadAllTransfersCommand{
		db:          r.db,
		chainClient: chainClient,
		accounts:    accounts,
		blockDAO:    r.blockDAO,
		eth: &ETHDownloader{
			chainClient: chainClient,
			accounts:    accounts,
			signer:      signer,
			db:          r.db,
		},
		erc20:              NewERC20TransfersDownloader(chainClient, accounts, signer),
		feed:               r.feed,
		errorsCount:        0,
		transactionManager: r.transactionManager,
	}
	return ctl
}

func (s *SequentialFetchStrategy) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.group != nil {
		return errAlreadyRunning
	}
	s.group = async.NewGroup(context.Background())

	// Run the command on interval (5 seconds)
	for _, chainClient := range s.chainClients {
		ctl := s.newCommand(chainClient, s.accounts)
		s.group.Add(ctl.Command())
	}

	return nil
}

// Stop stops reactor loop and waits till it exits.
func (s *SequentialFetchStrategy) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.group == nil {
		return
	}
	s.group.Stop()
	s.group.Wait()
	s.group = nil
}

func (s *SequentialFetchStrategy) setChainClients(chainClients map[uint64]*chain.ClientWithFallback) {
	s.chainClients = chainClients
}

func (s *SequentialFetchStrategy) setAccounts(accounts []common.Address) {
	s.accounts = accounts
}

// Reactor listens to new blocks and stores transfers into the database.
type Reactor struct {
	strategy HistoryFetcher
}

// func newReactor(strategy HistoryFetcher, chainClients map[uint64]*chain.ClientWithFallback, accounts []common.Address) *Reactor {
func newReactor(strategy HistoryFetcher) *Reactor {
	return &Reactor{
		strategy: strategy,
	}
}

// Start runs reactor loop in background.
func (r *Reactor) start(chainClients map[uint64]*chain.ClientWithFallback, accounts []common.Address) error {
	r.strategy.setChainClients(chainClients)
	r.strategy.setAccounts(accounts)
	return r.strategy.start()
}

// Stop stops reactor loop and waits till it exits.
func (r *Reactor) stop() {
	r.strategy.stop()
}

func (r *Reactor) restart(chainClients map[uint64]*chain.ClientWithFallback, accounts []common.Address) error {
	r.stop()
	return r.start(chainClients, accounts)
}

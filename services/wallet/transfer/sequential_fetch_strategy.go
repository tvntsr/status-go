package transfer

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/status-im/status-go/services/wallet/async"
)

type SequentialFetchStrategy struct {
	db *Database
	// blockDAO           *BlockRangeSequentialDAO
	blockDAO           *BlockDAO
	feed               *event.Feed
	mu                 sync.Mutex
	group              *async.Group
	transactionManager *TransactionManager
	chainClients       map[uint64]*chain.ClientWithFallback
	accounts           []common.Address
}

func (r *SequentialFetchStrategy) newCommand(chainClient *chain.ClientWithFallback,
	accounts []common.Address) *loadAllTransfersCommand {

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

// func (s *SequentialFetchStrategy) setChainClients(chainClients map[uint64]*chain.ClientWithFallback) {
// 	s.chainClients = chainClients
// }

// func (s *SequentialFetchStrategy) setAccounts(accounts []common.Address) {
// 	s.accounts = accounts
// }

func (s *SequentialFetchStrategy) kind() FetchStrategyType {
	return SequentialFetchStrategyType
}

func (s *SequentialFetchStrategy) getTransfersByAddress(ctx context.Context, chainID uint64, address common.Address, toBlock *big.Int,
	limit int64, fetchMore bool) ([]Transfer, error) {

	// TODO: implement
	return []Transfer{}, nil
}

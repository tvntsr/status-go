package transfer

import (
	"context"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/async"
)

// NewConcurrentDownloader creates ConcurrentDownloader instance.
func NewConcurrentDownloader(ctx context.Context) *ConcurrentDownloader {
	runner := async.NewAtomicGroup(ctx)
	result := &Result{}
	return &ConcurrentDownloader{runner, result}
}

type ConcurrentDownloader struct {
	*async.AtomicGroup
	*Result
}

type Result struct {
	mu          sync.Mutex
	transfers   []Transfer
	headers     []*DBHeader
	blockRanges [][]*big.Int
}

var errDownloaderStuck = errors.New("eth downloader is stuck")
var errBeginOfAccHistoryReached = errors.New("Reached begin of account history")

func (r *Result) Push(transfers ...Transfer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transfers = append(r.transfers, transfers...)
}

func (r *Result) Get() []Transfer {
	r.mu.Lock()
	defer r.mu.Unlock()
	rst := make([]Transfer, len(r.transfers))
	copy(rst, r.transfers)
	return rst
}

func (r *Result) PushHeader(block *DBHeader) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.headers = append(r.headers, block)
}

func (r *Result) GetHeaders() []*DBHeader {
	r.mu.Lock()
	defer r.mu.Unlock()
	rst := make([]*DBHeader, len(r.headers))
	copy(rst, r.headers)
	return rst
}

func (r *Result) PushRange(blockRange []*big.Int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.blockRanges = append(r.blockRanges, blockRange)
}

func (r *Result) GetRanges() [][]*big.Int {
	r.mu.Lock()
	defer r.mu.Unlock()
	rst := make([][]*big.Int, len(r.blockRanges))
	copy(rst, r.blockRanges)
	r.blockRanges = [][]*big.Int{}

	return rst
}

// Downloader downloads transfers from single block using number.
type Downloader interface {
	GetTransfersByNumber(context.Context, *big.Int) ([]Transfer, error)
}

// Returns new block ranges that contain transfers and found block headers that contain transfers.
func checkRanges(parent context.Context, client BalanceReader, cache BalanceCache, downloader Downloader,
	account common.Address, ranges [][]*big.Int) ([][]*big.Int, []*DBHeader, error) {

	log.Info("checkRanges start", "account", account.Hex(), "ranges len", len(ranges))

	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()

	c := NewConcurrentDownloader(ctx)

	for _, blocksRange := range ranges {
		from := blocksRange[0]
		to := blocksRange[1]

		c.Add(func(ctx context.Context) error {
			if from.Cmp(to) >= 0 {
				return nil
			}
			log.Debug("eth transfers comparing blocks", "from", from, "to", to)
			lb, err := cache.BalanceAt(ctx, client, account, from)
			if err != nil {
				return err
			}
			hb, err := cache.BalanceAt(ctx, client, account, to)
			if err != nil {
				return err
			}
			if lb.Cmp(hb) == 0 {
				log.Debug("balances are equal", "from", from, "to", to)

				hn, err := cache.NonceAt(ctx, client, account, to)
				if err != nil {
					return err
				}
				// if nonce is zero in a newer block then there is no need to check an older one
				if *hn == 0 {
					log.Debug("zero nonce", "to", to)
					return nil
				}

				ln, err := cache.NonceAt(ctx, client, account, from)
				if err != nil {
					return err
				}
				if *ln == *hn {
					log.Debug("transaction count is also equal", "from", from, "to", to)
					return nil
				}
			}
			if new(big.Int).Sub(to, from).Cmp(one) == 0 {
				header, err := client.HeaderByNumber(ctx, to)
				if err != nil {
					return err
				}
				c.PushHeader(toDBHeader(header))
				return nil
			}
			mid := new(big.Int).Add(from, to)
			mid = mid.Div(mid, two)
			_, err = cache.BalanceAt(ctx, client, account, mid)
			if err != nil {
				return err
			}
			// log.Info("balances are not equal", "from", from, "mid", mid, "to", to)

			c.PushRange([]*big.Int{from, mid})
			c.PushRange([]*big.Int{mid, to})
			return nil
		})

	}

	select {
	case <-c.WaitAsync():
	case <-ctx.Done():
		return nil, nil, errDownloaderStuck
	}

	if c.Error() != nil {
		return nil, nil, errors.Wrap(c.Error(), "failed to dowload transfers using concurrent downloader")
	}

	log.Info("checkRanges end", "account", account.Hex())

	return c.GetRanges(), c.GetHeaders(), nil
}

// Returns new block ranges that contain transfers and found block headers that contain transfers.
func checkRanges2(parent context.Context, client BalanceReader, cache BalanceCache, downloader Downloader,
	account common.Address, ranges [][]*big.Int, stopBlock *big.Int) ([][]*big.Int, []*DBHeader, *big.Int, error) {

	log.Info("checkRanges2 start", "account", account.Hex(), "ranges len", len(ranges))

	ctx, cancel := context.WithTimeout(parent, 30*time.Second)
	defer cancel()

	c := NewConcurrentDownloader(ctx)

	newStopBlock := stopBlock

	for _, blocksRange := range ranges {
		from := blocksRange[0]
		to := blocksRange[1]

		if to.Cmp(newStopBlock) <= 0 {
			log.Warn("'to' block is less than 'stop' block _", "to", to, "stopBlock", stopBlock)
			continue
		}

		c.Add(func(ctx context.Context) error {
			if from.Cmp(to) >= 0 {
				return nil
			}

			if to.Cmp(stopBlock) <= 0 {
				log.Warn("'to' block is less than 'stop' block", "to", to, "stopBlock", stopBlock)
				return nil
			}

			log.Debug("eth transfers comparing blocks", "from", from, "to", to)
			lb, err := cache.BalanceAt(ctx, client, account, from)
			if err != nil {
				return err
			}
			hb, err := cache.BalanceAt(ctx, client, account, to)
			if err != nil {
				return err
			}
			if lb.Cmp(hb) == 0 {
				log.Debug("balances are equal", "from", from, "to", to)

				hn, err := cache.NonceAt(ctx, client, account, to)
				if err != nil {
					return err
				}
				// if nonce is zero in a newer block then there is no need to check an older one
				if *hn == 0 {
					log.Debug("zero nonce", "to", to)

					if hb.Cmp(big.NewInt(0)) == 0 {
						if to.Cmp(newStopBlock) > 0 {
							log.Info("found possible start block, we should not go back", "block", to)
							newStopBlock = to // increase newStopBlock if we found a new higher block
						} else {
							log.Info("found possible start block, but it is less than current", "block", to)
						}
					}

					return nil
				}

				ln, err := cache.NonceAt(ctx, client, account, from)
				if err != nil {
					return err
				}
				if *ln == *hn {
					log.Debug("transaction count is also equal", "from", from, "to", to)
					return nil
				}
			}
			if new(big.Int).Sub(to, from).Cmp(one) == 0 {
				header, err := client.HeaderByNumber(ctx, to)
				if err != nil {
					return err
				}
				c.PushHeader(toDBHeader(header))
				return nil
			}
			mid := new(big.Int).Add(from, to)
			mid = mid.Div(mid, two)
			_, err = cache.BalanceAt(ctx, client, account, mid)
			if err != nil {
				return err
			}
			// log.Info("balances are not equal", "from", from, "mid", mid, "to", to)

			c.PushRange([]*big.Int{mid, to})
			c.PushRange([]*big.Int{from, mid})
			return nil
		})
	}

	select {
	case <-c.WaitAsync():
	case <-ctx.Done():
		return nil, nil, newStopBlock, errDownloaderStuck
	}

	if c.Error() != nil {
		return nil, nil, stopBlock, errors.Wrap(c.Error(), "failed to dowload transfers using concurrent downloader")
	}

	log.Info("checkRanges2 end", "account", account.Hex(), "newStopBlock", newStopBlock)

	return c.GetRanges(), c.GetHeaders(), newStopBlock, nil
}

func findBlocksWithEthTransfers2(parent context.Context, client BalanceReader, cache BalanceCache, downloader Downloader,
	account common.Address, low, high *big.Int, noLimit bool) (from *big.Int, headers []*DBHeader, resStopBlock *big.Int, err error) {

	ranges := [][]*big.Int{{low, high}}
	minBlock := big.NewInt(low.Int64())
	headers = []*DBHeader{}
	var lvl = 1
	var stopBlock *big.Int = big.NewInt(0)

	for len(ranges) > 0 && lvl <= 30 {
		log.Debug("check blocks ranges", "lvl", lvl, "ranges len", len(ranges))
		lvl++
		// Check if there are transfers in blocks in ranges. To do that, nonce and balance is checked
		// the block ranges that have transfers are returned
		newRanges, newHeaders, stpBlock, err := checkRanges2(parent, client, cache, downloader, account, ranges, stopBlock)
		stopBlock = stpBlock
		if err != nil {
			return nil, nil, stopBlock, err
		}

		headers = append(headers, newHeaders...)

		if len(newRanges) > 0 {
			log.Debug("found new ranges", "account", account, "lvl", lvl, "new ranges len", len(newRanges))
		}
		if len(newRanges) > 60 && !noLimit {
			sort.SliceStable(newRanges, func(i, j int) bool {
				return newRanges[i][0].Cmp(newRanges[j][0]) == 1
			})

			newRanges = newRanges[:60]
			minBlock = newRanges[len(newRanges)-1][0]
		}

		ranges = newRanges
	}

	return minBlock, headers, stopBlock, err
}

func findBlocksWithEthTransfers(parent context.Context, client BalanceReader, cache BalanceCache, downloader Downloader,
	account common.Address, low, high *big.Int, noLimit bool) (from *big.Int, headers []*DBHeader, err error) {
	log.Info("findBlocksWithEthTranfers start", "account", account, "low", low, "high", high, "noLimit", noLimit)

	ranges := [][]*big.Int{{low, high}}
	minBlock := big.NewInt(low.Int64())
	headers = []*DBHeader{}
	var lvl = 1
	// for len(ranges) > 0 && lvl <= 30 { // Why 30??
	for len(ranges) > 0 {
		log.Info("check blocks ranges", "lvl", lvl, "ranges len", len(ranges))
		lvl++
		newRanges, newHeaders, err := checkRanges(parent, client, cache, downloader, account, ranges)

		log.Info("check blocks ranges", "newRanges len", len(newRanges))
		if err != nil {
			log.Info("check ranges end", "err", err)
			return nil, nil, err
		}

		headers = append(headers, newHeaders...)

		if len(newRanges) > 0 {
			log.Info("found new ranges", "account", account, "lvl", lvl, "new ranges len", len(newRanges))
		}
		if len(newRanges) > 60 && !noLimit { // Why 60??
			sort.SliceStable(newRanges, func(i, j int) bool {
				return newRanges[i][0].Cmp(newRanges[j][0]) == 1
			})

			newRanges = newRanges[:60]
			minBlock = newRanges[len(newRanges)-1][0]
		}

		ranges = newRanges
	}

	log.Info("findBlocksWithEthTranfers end", "account", account, "minBlock", minBlock, "headers len", len(headers))
	return minBlock, headers, err
}

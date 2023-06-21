package rpcfilters

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth/filters"
	"github.com/ethereum/go-ethereum/rpc"
)

func TestFilterLiveness(t *testing.T) {
	api := &PublicAPI{
		filters:              make(map[rpc.ID]filter),
		filterLivenessLoop:   10 * time.Millisecond,
		filterLivenessPeriod: 15 * time.Millisecond,
		client:               func() ContextCaller { return &callTracker{} },
		chainID:              func() uint64 { return 1 },
	}
	id, err := api.NewFilter(filters.FilterCriteria{})
	require.NoError(t, err)
	quit := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		api.timeoutLoop(quit)
		wg.Done()
	}()
	tick := time.Tick(10 * time.Millisecond)
	after := time.After(100 * time.Millisecond)
	func() {
		for {
			select {
			case <-after:
				assert.FailNow(t, "filter wasn't removed")
				close(quit)
				return
			case <-tick:
				api.filtersMu.Lock()
				_, exist := api.filters[id]
				api.filtersMu.Unlock()
				if !exist {
					close(quit)
					return
				}
			}
		}
	}()
	wg.Wait()
}

func TestGetFilterChangesResetsTimer(t *testing.T) {
	api := &PublicAPI{
		filters:              make(map[rpc.ID]filter),
		filterLivenessLoop:   10 * time.Millisecond,
		filterLivenessPeriod: 15 * time.Millisecond,
		client:               func() ContextCaller { return &callTracker{} },
		chainID:              func() uint64 { return 1 },
	}
	id, err := api.NewFilter(filters.FilterCriteria{})
	require.NoError(t, err)

	api.filtersMu.Lock()
	f := api.filters[id]
	require.True(t, f.deadline().Stop())
	fake := make(chan time.Time, 1)
	fake <- time.Time{}
	f.deadline().C = fake
	api.filtersMu.Unlock()

	require.False(t, f.deadline().Stop())
	// GetFilterChanges will Reset deadline
	_, err = api.GetFilterChanges(id)
	require.NoError(t, err)
	require.True(t, f.deadline().Stop())
}

func TestGetFilterLogs(t *testing.T) {
	t.Skip("Skipping due to flakiness: https://github.com/status-im/status-go/issues/1281")

	tracker := new(callTracker)
	api := &PublicAPI{
		filters: make(map[rpc.ID]filter),
		client:  func() ContextCaller { return tracker },
		chainID: func() uint64 { return 1 },
	}
	block := big.NewInt(10)
	id, err := api.NewFilter(filters.FilterCriteria{
		FromBlock: block,
	})
	require.NoError(t, err)
	logs, err := api.GetFilterLogs(context.TODO(), id)
	require.NoError(t, err)
	require.Empty(t, logs)
	require.Len(t, tracker.criteria, 1)
	rst, err := hexutil.DecodeBig(tracker.criteria[0]["fromBlock"].(string))
	require.NoError(t, err)
	require.Equal(t, block, rst)
}

func TestNewPendingTransactionFilter(t *testing.T) {

	tracker := new(callTracker)
	api := &PublicAPI{
		filters:                        make(map[rpc.ID]filter),
		client:                         func() ContextCaller { return tracker },
		chainID:                        func() uint64 { return 1 },
		transactionSentToUpstreamEvent: newTransactionSentToUpstreamEvent(),
	}

	require.NoError(t, api.transactionSentToUpstreamEvent.Start())
	defer api.transactionSentToUpstreamEvent.Stop()

	id := api.NewPendingTransactionFilter()

	hashes, err := api.GetFilterChanges(id)
	require.NoError(t, err)

	require.Equal(t, 0, len(hashes.([]common.Hash)))

	var wg sync.WaitGroup
	wg.Add(1)
	// Do it async, otherwise no event will be received
	go func() {
		time.Sleep(100 * time.Millisecond)
		t.Log("triggering transactionSentToUpstreamEvent")
		// api.transactionSentToUpstreamEvent.Trigger(types.HexToHash("0xAA"))
		api.transactionSentToUpstreamEvent.Trigger(common.HexToHash("0xAA"))
		wg.Done()
	}()
	wg.Wait()

	time.Sleep(1000 * time.Millisecond)
	t.Log("getting filter changes")
	hashes, err = api.GetFilterChanges(id)
	require.NoError(t, err)
	require.Equal(t, 1, len(hashes.([]common.Hash)))
}

// var transactionHashes = []types.Hash{types.HexToHash("0xAA"), types.HexToHash("0xBB"), types.HexToHash("0xCC")}

// func TestTransactionSentToUpstreamEventMultipleSubscribe(t *testing.T) {
// 	event := newTransactionSentToUpstreamEvent()
// 	require.NoError(t, event.Start())
// 	defer event.Stop()

// 	var subscriptionChannels []chan types.Hash
// 	for i := 0; i < 3; i++ {
// 		id, channel := event.Subscribe()
// 		// test id assignment
// 		require.Equal(t, i, id)
// 		// test numberOfSubscriptions
// 		require.Equal(t, event.numberOfSubscriptions(), i+1)
// 		subscriptionChannels = append(subscriptionChannels, channel)
// 	}

// 	var wg sync.WaitGroup

// 	wg.Add(9)
// 	go func() {
// 		for _, channel := range subscriptionChannels {
// 			ch := channel
// 			go func() {
// 				for _, expectedHash := range transactionHashes {
// 					select {
// 					case receivedHash := <-ch:
// 						require.Equal(t, expectedHash, receivedHash)
// 					case <-time.After(1 * time.Second):
// 						assert.Fail(t, "timeout")
// 					}
// 					wg.Done()
// 				}
// 			}()
// 		}
// 	}()

// 	for _, hashToTrigger := range transactionHashes {
// 		event.Trigger(hashToTrigger)
// 	}
// 	wg.Wait()
// }

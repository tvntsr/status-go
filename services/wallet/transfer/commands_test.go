package transfer

import (
	"context"
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
	"github.com/status-im/status-go/rpc/chain"
	"github.com/stretchr/testify/require"
)

// func setupTestDB(t *testing.T) (*Database, *Block, func()) {
// 	db, err := appdatabase.InitializeDB(sqlite.InMemory, "wallet-tests", sqlite.ReducedKDFIterationsNumber)
// 	require.NoError(t, err)
// 	return NewDB(db), &Block{db}, func() {
// 		require.NoError(t, db.Close())
// 	}
// }

func Test_loadTransfers(t *testing.T) {
	db, _, stop := setupTestDB(t)
	defer stop()

	// rpcClient, _ := rpc.NewClient(s.client, chainID, params.UpstreamRPCConfig{}, nil, nil)
	// rpcClient.UpstreamChainID = chainID

	INFURA_TOKEN := "c4bf68a9de2d49bbb04447a73ffd3e0b"
	URL := "https://mainnet.infura.io/v3/" + INFURA_TOKEN
	gethRPCClient, err := gethrpc.Dial(URL)
	require.NoError(t, err)

	var chainID uint64 = 1
	c := chain.NewClient(gethRPCClient, nil, chainID)

	txManager, tdbClose := setupTestTransactionDB(t)
	defer tdbClose()

	type args struct {
		ctx                context.Context
		accounts           []common.Address
		block              *BlockDAO
		db                 *Database
		chainClient        *chain.ClientWithFallback
		limit              int
		blocksByAddress    map[common.Address][]*big.Int
		transactionManager *TransactionManager
	}
	tests := []struct {
		name    string
		args    args
		want    map[common.Address][]Transfer
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			args: args{
				ctx:                context.Background(),
				accounts:           []common.Address{common.HexToAddress("0xE4eDb277e41dc89aB076a1F049f4a3EfA700bCE8")},
				block:              &BlockDAO{db.client},
				db:                 db,
				chainClient:        c,
				limit:              20,
				transactionManager: txManager,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadTransfers(tt.args.ctx, tt.args.accounts, tt.args.block, tt.args.db, tt.args.chainClient, tt.args.limit, tt.args.blocksByAddress, tt.args.transactionManager)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadTransfers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// fmt.Printf("loadTransfers() = %v", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadTransfers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_controlCommand_Run(t *testing.T) {
	db, _, stop := setupTestDB(t)
	defer stop()

	// rpcClient, _ := rpc.NewClient(s.client, chainID, params.UpstreamRPCConfig{}, nil, nil)
	// rpcClient.UpstreamChainID = chainID

	INFURA_TOKEN := "c4bf68a9de2d49bbb04447a73ffd3e0b"
	URL := "https://mainnet.infura.io/v3/" + INFURA_TOKEN
	gethRPCClient, err := gethrpc.Dial(URL)
	require.NoError(t, err)

	var chainID uint64 = 1
	chainClient := chain.NewClient(gethRPCClient, nil, chainID)

	txManager, tdbClose := setupTestTransactionDB(t)
	defer tdbClose()

	signer := types.NewLondonSigner(chainClient.ToBigInt())
	accounts := []common.Address{common.HexToAddress("0xE4eDb277e41dc89aB076a1F049f4a3EfA700bCE8")}

	type fields struct {
		accounts           []common.Address
		db                 *Database
		block              *BlockDAO
		eth                *ETHDownloader
		erc20              *ERC20TransfersDownloader
		chainClient        *chain.ClientWithFallback
		feed               *event.Feed
		errorsCount        int
		nonArchivalRPCNode bool
		transactionManager *TransactionManager
	}
	type args struct {
		parent context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				accounts: accounts,
				db:       db,
				block:    &BlockDAO{db.client},
				eth: &ETHDownloader{
					chainClient: chainClient,
					accounts:    accounts,
					signer:      signer,
					db:          db,
				},
				erc20: NewERC20TransfersDownloader(chainClient, accounts, signer),
				// feed:               r.feed,
				errorsCount:        0,
				chainClient:        chainClient,
				transactionManager: txManager,
			},
			args: args{
				parent: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &controlCommand{
				accounts:           tt.fields.accounts,
				db:                 tt.fields.db,
				blockDAO:           tt.fields.block,
				eth:                tt.fields.eth,
				erc20:              tt.fields.erc20,
				chainClient:        tt.fields.chainClient,
				feed:               tt.fields.feed,
				errorsCount:        tt.fields.errorsCount,
				nonArchivalRPCNode: tt.fields.nonArchivalRPCNode,
				transactionManager: tt.fields.transactionManager,
			}
			if err := c.Run(tt.args.parent); (err != nil) != tt.wantErr {
				t.Errorf("controlCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_findAndCheckBlockRangeCommand_fastIndexErc20(t *testing.T) {
	db, _, stop := setupTestDB(t)
	defer stop()

	// rpcClient, _ := rpc.NewClient(s.client, chainID, params.UpstreamRPCConfig{}, nil, nil)
	// rpcClient.UpstreamChainID = chainID

	INFURA_TOKEN := "c4bf68a9de2d49bbb04447a73ffd3e0b"
	URL := "https://mainnet.infura.io/v3/" + INFURA_TOKEN
	gethRPCClient, err := gethrpc.Dial(URL)
	require.NoError(t, err)

	var chainID uint64 = 1
	chainClient := chain.NewClient(gethRPCClient, nil, chainID)

	// txManager, tdbClose := setupTestTransactionDB(t)
	// defer tdbClose()

	// signer := types.NewLondonSigner(chainClient.ToBigInt())
	accounts := []common.Address{common.HexToAddress("0xE4eDb277e41dc89aB076a1F049f4a3EfA700bCE8")}

	type fields struct {
		accounts      []common.Address
		db            *Database
		chainClient   *chain.ClientWithFallback
		balanceCache  *balanceCache
		feed          *event.Feed
		fromByAddress map[common.Address]*Block
		toByAddress   map[common.Address]*big.Int
		foundHeaders  map[common.Address][]*DBHeader
		noLimit       bool
		error         error
	}
	type args struct {
		ctx           context.Context
		fromByAddress map[common.Address]*big.Int
		// fromByAddress map[common.Address]*Block
		toByAddress map[common.Address]*big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[common.Address][]*DBHeader
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				accounts:     accounts,
				db:           db,
				chainClient:  chainClient,
				balanceCache: newBalanceCache(),
				// noLimit:      true,
			},
			args: args{
				ctx: context.Background(),
				fromByAddress: map[common.Address]*big.Int{
					accounts[0]: big.NewInt(17071951),
				},
				// fromByAddress: map[common.Address]*Block{
				// 	accounts[0]: {Number: big.NewInt(17071951)},
				// },
				// toByAddress: map[common.Address]*big.Int{
				// 	accounts[0]: big.NewInt(17072951), // some last block number at the moment
				// },
			},
			// want: map[common.Address][]*DBHeader{
			// 	accounts[0]: {
			// 		{
			// 			0xc0000da3f0 0xc0000da620 0xc0000da8c0 0xc0000daa80 0xc0000dac40 0xc0000dad90 0xc0003160e0}
			// 	},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &findAndCheckBlockRangeCommand{
				accounts:      tt.fields.accounts,
				db:            tt.fields.db,
				chainClient:   tt.fields.chainClient,
				balanceCache:  tt.fields.balanceCache,
				feed:          tt.fields.feed,
				fromByAddress: tt.fields.fromByAddress,
				toByAddress:   tt.fields.toByAddress,
				foundHeaders:  tt.fields.foundHeaders,
				noLimit:       tt.fields.noLimit,
				error:         tt.fields.error,
			}
			got, err := c.fastIndexErc20(tt.args.ctx, tt.args.fromByAddress, tt.args.toByAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAndCheckBlockRangeCommand.fastIndexErc20() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("findAndCheckBlockRangeCommand.fastIndexErc20() got:")
			for _, header := range got[accounts[0]] {
				t.Log("header", "Number", header.Number, "hash", header.Hash, "ts", header.Timestamp)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAndCheckBlockRangeCommand.fastIndexErc20() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_findAndCheckBlockRangeCommand_fastIndex(t *testing.T) {
	db, _, stop := setupTestDB(t)
	defer stop()

	INFURA_TOKEN := "c4bf68a9de2d49bbb04447a73ffd3e0b"
	URL := "https://mainnet.infura.io/v3/" + INFURA_TOKEN
	gethRPCClient, err := gethrpc.Dial(URL)
	require.NoError(t, err)

	var chainID uint64 = 1
	chainClient := chain.NewClient(gethRPCClient, nil, chainID)
	accounts := []common.Address{common.HexToAddress("0xE4eDb277e41dc89aB076a1F049f4a3EfA700bCE8")}

	type fields struct {
		accounts      []common.Address
		db            *Database
		chainClient   *chain.ClientWithFallback
		balanceCache  *balanceCache
		feed          *event.Feed
		fromByAddress map[common.Address]*Block
		toByAddress   map[common.Address]*big.Int
		foundHeaders  map[common.Address][]*DBHeader
		noLimit       bool
		error         error
	}
	type args struct {
		ctx           context.Context
		bCache        *balanceCache
		fromByAddress map[common.Address]*Block
		toByAddress   map[common.Address]*big.Int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[common.Address]*big.Int
		want1   map[common.Address][]*DBHeader
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				accounts:     accounts,
				db:           db,
				chainClient:  chainClient,
				balanceCache: newBalanceCache(),
			},
			args: args{
				ctx:    context.Background(),
				bCache: newBalanceCache(),
				fromByAddress: map[common.Address]*Block{
					accounts[0]: {Number: big.NewInt(17071951)},
				},
				toByAddress: map[common.Address]*big.Int{
					accounts[0]: big.NewInt(17072951), // some last block number at the moment
				},
			},
			// want: map[common.Address][]*DBHeader{
			// 	accounts[0]: {
			// 		{
			// 			0xc0000da3f0 0xc0000da620 0xc0000da8c0 0xc0000daa80 0xc0000dac40 0xc0000dad90 0xc0003160e0}
			// 	},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &findAndCheckBlockRangeCommand{
				accounts:      tt.fields.accounts,
				db:            tt.fields.db,
				chainClient:   tt.fields.chainClient,
				balanceCache:  tt.fields.balanceCache,
				feed:          tt.fields.feed,
				fromByAddress: tt.fields.fromByAddress,
				toByAddress:   tt.fields.toByAddress,
				foundHeaders:  tt.fields.foundHeaders,
				noLimit:       tt.fields.noLimit,
				error:         tt.fields.error,
			}
			got, got1, err := c.fastIndex(tt.args.ctx, tt.args.bCache, tt.args.fromByAddress, tt.args.toByAddress)
			if (err != nil) != tt.wantErr {
				t.Errorf("findAndCheckBlockRangeCommand.fastIndex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Log("findAndCheckBlockRangeCommand.fastIndex() got:")
			for _, header := range got1[accounts[0]] {
				t.Log("header", "Number", header.Number, "hash", header.Hash, "ts", header.Timestamp)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("findAndCheckBlockRangeCommand.fastIndex() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("findAndCheckBlockRangeCommand.fastIndex() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func Test_findAndCheckBlockRangeCommand_Run(t *testing.T) {
	db, _, stop := setupTestDB(t)
	defer stop()

	INFURA_TOKEN := "c4bf68a9de2d49bbb04447a73ffd3e0b"
	URL := "https://mainnet.infura.io/v3/" + INFURA_TOKEN
	gethRPCClient, err := gethrpc.Dial(URL)
	require.NoError(t, err)

	var chainID uint64 = 1
	chainClient := chain.NewClient(gethRPCClient, nil, chainID)
	accounts := []common.Address{common.HexToAddress("0xE4eDb277e41dc89aB076a1F049f4a3EfA700bCE8")}

	type fields struct {
		accounts      []common.Address
		db            *Database
		chainClient   *chain.ClientWithFallback
		balanceCache  *balanceCache
		feed          *event.Feed
		fromByAddress map[common.Address]*Block
		// fromByAddress map[common.Address]*big.Int
		toByAddress  map[common.Address]*big.Int
		foundHeaders map[common.Address][]*DBHeader
		noLimit      bool
		error        error
	}
	type args struct {
		parent context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "test1",
			fields: fields{
				accounts:     accounts,
				db:           db,
				chainClient:  chainClient,
				balanceCache: newBalanceCache(),
				fromByAddress: map[common.Address]*Block{
					accounts[0]: {Number: big.NewInt(17071951)},
				},
				// fromByAddress: map[common.Address]*big.Int{
				// 	accounts[0]: big.NewInt(17071951),
				// },
				toByAddress: map[common.Address]*big.Int{
					accounts[0]: big.NewInt(17072951), // some last block number at the moment
				},
				noLimit: false,
			},
			args: args{
				parent: context.Background(),
			},
			// want: map[common.Address][]*DBHeader{
			// 	accounts[0]: {
			// 		{
			// 			0xc0000da3f0 0xc0000da620 0xc0000da8c0 0xc0000daa80 0xc0000dac40 0xc0000dad90 0xc0003160e0}
			// 	},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &findAndCheckBlockRangeCommand{
				accounts:      tt.fields.accounts,
				db:            tt.fields.db,
				chainClient:   tt.fields.chainClient,
				balanceCache:  tt.fields.balanceCache,
				feed:          tt.fields.feed,
				fromByAddress: tt.fields.fromByAddress,
				toByAddress:   tt.fields.toByAddress,
				foundHeaders:  tt.fields.foundHeaders,
				noLimit:       tt.fields.noLimit,
				error:         tt.fields.error,
			}
			if err := c.Run(tt.args.parent); (err != nil) != tt.wantErr {
				t.Errorf("findAndCheckBlockRangeCommand.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			t.Log("findAndCheckBlockRangeCommand.Run() len:", len(c.foundHeaders[accounts[0]]))
			for _, header := range c.foundHeaders[accounts[0]] {
				t.Log("header", "Number", header.Number, "hash", header.Hash, "ts", header.Timestamp)
			}
		})
	}
}

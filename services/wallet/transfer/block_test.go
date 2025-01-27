package transfer

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/status-im/status-go/t/helpers"
	"github.com/status-im/status-go/walletdatabase"

	"github.com/ethereum/go-ethereum/common"
)

func setupTestTransferDB(t *testing.T) (*BlockDAO, func()) {
	db, err := helpers.SetupTestMemorySQLDB(walletdatabase.DbInitializer{})
	require.NoError(t, err)
	return &BlockDAO{db}, func() {
		require.NoError(t, db.Close())
	}
}

func TestInsertRange(t *testing.T) {
	b, stop := setupTestTransferDB(t)
	defer stop()

	r := &BlocksRange{
		from: big.NewInt(0),
		to:   big.NewInt(10),
	}
	nonce := uint64(199)
	balance := big.NewInt(7657)
	account := common.Address{2}

	err := b.insertRange(777, account, r.from, r.to, balance, nonce)
	require.NoError(t, err)

	block, err := b.GetLastKnownBlockByAddress(777, account)
	require.NoError(t, err)

	require.Equal(t, 0, block.Number.Cmp(r.to))
	require.Equal(t, 0, block.Balance.Cmp(balance))
	require.Equal(t, nonce, uint64(*block.Nonce))
}

func TestGetNewRanges(t *testing.T) {
	ranges := []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(0),
			to:   big.NewInt(10),
		},
		&BlocksRange{
			from: big.NewInt(10),
			to:   big.NewInt(20),
		},
	}

	n, d := getNewRanges(ranges)
	require.Equal(t, 1, len(n))
	newRange := n[0]
	require.Equal(t, int64(0), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	require.Equal(t, 2, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(0),
			to:   big.NewInt(11),
		},
		&BlocksRange{
			from: big.NewInt(10),
			to:   big.NewInt(20),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 1, len(n))
	newRange = n[0]
	require.Equal(t, int64(0), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	require.Equal(t, 2, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(0),
			to:   big.NewInt(20),
		},
		&BlocksRange{
			from: big.NewInt(5),
			to:   big.NewInt(15),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 1, len(n))
	newRange = n[0]
	require.Equal(t, int64(0), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	require.Equal(t, 2, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(5),
			to:   big.NewInt(15),
		},
		&BlocksRange{
			from: big.NewInt(5),
			to:   big.NewInt(20),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 1, len(n))
	newRange = n[0]
	require.Equal(t, int64(5), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	require.Equal(t, 2, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(5),
			to:   big.NewInt(10),
		},
		&BlocksRange{
			from: big.NewInt(15),
			to:   big.NewInt(20),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 0, len(n))
	require.Equal(t, 0, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(0),
			to:   big.NewInt(10),
		},
		&BlocksRange{
			from: big.NewInt(10),
			to:   big.NewInt(20),
		},
		&BlocksRange{
			from: big.NewInt(30),
			to:   big.NewInt(40),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 1, len(n))
	newRange = n[0]
	require.Equal(t, int64(0), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	require.Equal(t, 2, len(d))

	ranges = []*BlocksRange{
		&BlocksRange{
			from: big.NewInt(0),
			to:   big.NewInt(10),
		},
		&BlocksRange{
			from: big.NewInt(10),
			to:   big.NewInt(20),
		},
		&BlocksRange{
			from: big.NewInt(30),
			to:   big.NewInt(40),
		},
		&BlocksRange{
			from: big.NewInt(40),
			to:   big.NewInt(50),
		},
	}

	n, d = getNewRanges(ranges)
	require.Equal(t, 2, len(n))
	newRange = n[0]
	require.Equal(t, int64(0), newRange.from.Int64())
	require.Equal(t, int64(20), newRange.to.Int64())
	newRange = n[1]
	require.Equal(t, int64(30), newRange.from.Int64())
	require.Equal(t, int64(50), newRange.to.Int64())
	require.Equal(t, 4, len(d))
}

func TestInsertZeroBalance(t *testing.T) {
	db, _, err := helpers.SetupTestSQLDB(walletdatabase.DbInitializer{}, "zero-balance")
	require.NoError(t, err)

	b := &BlockDAO{db}
	r := &BlocksRange{
		from: big.NewInt(0),
		to:   big.NewInt(10),
	}
	nonce := uint64(199)
	balance := big.NewInt(0)
	account := common.Address{2}

	err = b.insertRange(777, account, r.from, r.to, balance, nonce)
	require.NoError(t, err)

	block, err := b.GetLastKnownBlockByAddress(777, account)
	require.NoError(t, err)

	require.Equal(t, 0, block.Number.Cmp(r.to))
	require.Equal(t, big.NewInt(0).Int64(), block.Balance.Int64())
	require.Equal(t, nonce, uint64(*block.Nonce))
}

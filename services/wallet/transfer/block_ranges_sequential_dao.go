package transfer

import (
	"database/sql"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/status-go/services/wallet/bigint"
)

type BlockRangeSequentialDAO struct {
	db *sql.DB
}

func (b *BlockRangeSequentialDAO) GetFirstKnownBlock(chainID uint64, address common.Address) (block *Block, err error) {
	query := `SELECT blk_first FROM blocks_ranges_sequential
	WHERE address = ?
	AND network_id = ?
	ORDER BY blk_first
	LIMIT 1`

	rows, err := b.db.Query(query, address, chainID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		block = &Block{Number: &big.Int{}}
		err = rows.Scan((*bigint.SQLBigInt)(block.Number))
		if err != nil {
			return nil, err
		}

		return block, nil
	}

	return nil, nil
}

func (b *BlockRangeSequentialDAO) GetLastKnownBlock(chainID uint64, address common.Address) (block *Block, err error) {
	query := `SELECT blk_last FROM blocks_ranges_sequential
	WHERE address = ?
	AND network_id = ?
	ORDER BY blk_last DESC
	LIMIT 1`

	rows, err := b.db.Query(query, address, chainID)
	if err != nil {
		return
	}
	defer rows.Close()

	if rows.Next() {
		block = &Block{Number: &big.Int{}}
		err = rows.Scan((*bigint.SQLBigInt)(block.Number))
		if err != nil {
			return nil, err
		}
		return block, nil
	}

	return nil, nil
}

func (b *BlockRangeSequentialDAO) getLastKnownBlocks(chainID uint64, addresses []common.Address) (map[common.Address]*Block, error) {
	result := map[common.Address]*Block{}
	for _, address := range addresses {
		block, error := b.GetLastKnownBlock(chainID, address)
		if error != nil {
			return nil, error
		}

		if block != nil {
			result[address] = block
		}
	}

	return result, nil
}

func (b *BlockRangeSequentialDAO) getFirstKnownBlocks(chainID uint64, addresses []common.Address) (map[common.Address]*Block, error) {
	result := map[common.Address]*Block{}
	for _, address := range addresses {
		block, error := b.GetFirstKnownBlock(chainID, address)
		if error != nil {
			return nil, error
		}

		if block != nil {
			result[address] = block
		}
	}

	return result, nil
}

// func deleteRange(chainID uint64, creator statementCreator, account common.Address, first *big.Int, last *big.Int) error {
// 	log.Info("delete blocks range", "account", account, "network", chainID, "first", first, "last", last)
// 	delete, err := creator.Prepare(`DELETE FROM blocks_ranges_sequential
//                                         WHERE address = ?
//                                         AND network_id = ?
//                                         AND blk_first = ?
//                                         AND blk_last = ?`)
// 	if err != nil {
// 		log.Info("some error", "error", err)
// 		return err
// 	}

// 	_, err = delete.Exec(account, chainID, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(last))
// 	return err
// }

func (b *BlockRangeSequentialDAO) deleteRange(chainID uint64, account common.Address, first *big.Int, last *big.Int) error {
	log.Info("delete blocks range", "account", account, "network", chainID, "first", first, "last", last)
	delete, err := b.db.Prepare(`DELETE FROM blocks_ranges_sequential
                                        WHERE address = ?
                                        AND network_id = ?
                                        AND blk_first = ?
                                        AND blk_last = ?`)
	if err != nil {
		log.Info("some error", "error", err)
		return err
	}

	_, err = delete.Exec(account, chainID, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(last))
	return err
}

// func insertRange(chainID uint64, creator statementCreator, account common.Address, first *big.Int, last *big.Int) error {
// 	log.Info("insert blocks range", "account", account, "network", chainID, "first", first, "last", last)
// 	insert, err := creator.Prepare("INSERT INTO blocks_ranges_sequential (network_id, address, blk_first, blk_last) VALUES (?, ?, ?, ?)")
// 	if err != nil {
// 		return err
// 	}

// 	_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(last))
// 	return err
// }

func (b *BlockRangeSequentialDAO) insertRange(chainID uint64, account common.Address, first *big.Int, last *big.Int) error {
	log.Debug("insert blocks range", "account", account, "network id", chainID, "first", first, "last", last)
	insert, err := b.db.Prepare("INSERT INTO blocks_ranges_sequential (network_id, address, blk_first, blk_last) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(last))
	return err
}

// func upsertRange(chainID uint64, creator statementCreator, account common.Address, first *big.Int, last *Block) (err error) {
// 	log.Debug("upsert blocks range", "account", account, "network id", chainID, "first", first, "last", last.Number)
// 	update, err := creator.Prepare(`UPDATE blocks_ranges_sequential
//                 SET blk_last = ?,
//                 WHERE address = ?
//                 AND network_id = ?
//                 AND blk_last = ?`)

// 	if err != nil {
// 		return err
// 	}

// 	res, err := update.Exec((*bigint.SQLBigInt)(to.Number), account, chainID, (*bigint.SQLBigInt)(first))

// 	if err != nil {
// 		return err
// 	}
// 	affected, err := res.RowsAffected()
// 	if err != nil {
// 		return err
// 	}
// 	if affected == 0 {
// 		insert, err := creator.Prepare("INSERT INTO blocks_ranges_sequential (network_id, address, blk_first, blk_last) VALUES (?, ?, ?, ?)")
// 		if err != nil {
// 			return err
// 		}

// 		_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(to.Number))
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return
// }

func (b *BlockRangeSequentialDAO) upsertRange(chainID uint64, creator statementCreator, account common.Address, first *big.Int, last *Block) (err error) {
	log.Debug("upsert blocks range", "account", account, "network id", chainID, "first", first, "last", last.Number)
	update, err := b.db.Prepare(`UPDATE blocks_ranges_sequential
                SET blk_last = ?,
                WHERE address = ?
                AND network_id = ?
                AND blk_last = ?`)

	if err != nil {
		return err
	}

	res, err := update.Exec((*bigint.SQLBigInt)(last.Number), account, chainID, (*bigint.SQLBigInt)(first))

	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		insert, err := creator.Prepare("INSERT INTO blocks_ranges_sequential (network_id, address, blk_first, blk_last) VALUES (?, ?, ?, ?)")
		if err != nil {
			return err
		}

		_, err = insert.Exec(chainID, account, (*bigint.SQLBigInt)(first), (*bigint.SQLBigInt)(last.Number))
		if err != nil {
			return err
		}
	}

	return
}

package executor

import (
	"bytes"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/pkg/errors"
)

// 生成 key -> id 格式的本地数据库数据， 在下个版本这个文件可以全部删除
// 由于数据库精简需要保存具体数据

// 将手动生成的local db 的代码和用table 生成的local db的代码分离出来
// 手动生成的local db, 将不生成任意资产标价的数据， 保留用coins 生成交易的数据， 来兼容为升级的app 应用
// 希望有全量数据的， 需要调用新的rpc

// 下个版本可以删除
const (
	sellOrderSHTAS = "LODB-trade-sellorder-shtas:"
	sellOrderASTS  = "LODB-trade-sellorder-asts:"
	sellOrderATSS  = "LODB-trade-sellorder-atss:"
	sellOrderTSPAS = "LODB-trade-sellorder-tspas:"
	buyOrderSHTAS  = "LODB-trade-buyorder-shtas:"
	buyOrderASTS   = "LODB-trade-buyorder-asts:"
	buyOrderATSS   = "LODB-trade-buyorder-atss:"
	buyOrderTSPAS  = "LODB-trade-buyorder-tspas:"
	// Addr-Status-Type-Height-Key
	orderASTHK = "LODB-trade-order-asthk:"
)

// UpdateLocalDBPart1 手动生成KV，需要在原有数据库中删除
func UpdateLocalDBPart1(localDB dbm.DB) error {
	prefixes := []string{
		sellOrderSHTAS,
		sellOrderASTS,
		sellOrderATSS,
		sellOrderTSPAS,
		buyOrderSHTAS,
		buyOrderASTS,
		buyOrderATSS,
		buyOrderTSPAS,
		orderASTHK,
	}

	for _, prefix := range prefixes {
		err := delOnePrefix(localDB, prefix)
		if err != nil {
			tradelog.Error("UpdateLocalDBPart1 failed", "err", err)
			return errors.Cause(err)
		}
	}
	return nil
}

func delOnePrefix(localDB dbm.DB, prefix string) (err error) {
	allDeleted := false
	total := 10000
	for !allDeleted {
		allDeleted, err = delOnePrefixLimit(localDB, prefix, total)
		if err != nil {
			return err
		}
	}
	return nil
}

// 删除指定前缀的N个记录
//  DB interface -> included IteratorDB interface
//  IteratorDB interface -> Iterator func return Iterator interface
//  Iterator interface -> Key func
//  先写入固定的 start, end， 所有key 在这中间， 然后慢慢处理
func delOnePrefixLimit(localDB dbm.DB, prefix string, total int) (allDeleted bool, err error) {
	start := []byte(prefix)
	err = localDB.SetSync(start, []byte(""))
	if err != nil {
		return allDeleted, errors.Wrap(err, "SetSync for prefix start:"+prefix)
	}

	keys := make([][]byte, total)

	count := 0
	it := localDB.Iterator(start, nil, false)
	for it.Rewind(); it.Valid(); it.Next() {
		keys[count] = it.Key()
		count++
		if count == total {
			break
		}
	}

	batch := localDB.NewBatch(false)
	for i := 0; i < count; i++ {
		// 保护下， 避免测试其他bug，误删其他key， 破坏数据库
		if !bytes.HasPrefix(keys[i], start) {
			tradelog.Error("delOnePrefixLimit delete key not match prefix", "prefix", prefix, "key", string(keys[i]))
			panic("bug: " + "delOnePrefixLimit delete key not match prefix: " + prefix + " " + string(keys[i]))
		}
		batch.Delete(keys[i])
	}
	err = batch.Write()
	if err != nil {
		return allDeleted, errors.Wrap(err, "batch.Write when delete prefix keys: "+prefix)
	}

	return count < total, nil
}

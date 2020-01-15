package executor

import (
	"bytes"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
	"github.com/pkg/errors"
)

const (
	tradeLocaldbVersioin = "LODB-trade-version"
)

// 由于数据库精简需要保存具体数据
// 生成 key -> id 格式的本地数据库数据， 在下个版本这个文件可以全部删除
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

// Upgrade 实现升级接口
func (t *trade) Upgrade() error {
	localDB := t.GetLocalDB()
	return TradeUpdateLocalDBV2(localDB, 0)
}

// TradeUpdateLocalDBV2 trade 本地数据库升级
// from 1 to 2
func TradeUpdateLocalDBV2(localDB dbm.KVDB, total int) error {
	// 外部不指定， 强制分批执行
	if total <= 0 {
		total = 10000
	}
	toVersion := 2
	version, err := getVersion(localDB)
	if err != nil {
		tradelog.Error("TradeUpdateLocalDBV2 get version", "err", err)
		return errors.Cause(err)
	}
	if version >= toVersion {
		return nil
	}

	err = UpdateLocalDBPart2(localDB, total)
	if err != nil {
		return err
	}
	// TODO input DB to KVDB
	err = UpdateLocalDBPart1(nil, total)
	if err != nil {
		return err
	}
	return setVersion(localDB, toVersion)
}

// UpdateLocalDBPart1 手动生成KV，需要在原有数据库中删除
func UpdateLocalDBPart1(localDB dbm.DB, total int) error {
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
		err := delOnePrefix(localDB, prefix, total)
		if err != nil {
			tradelog.Error("UpdateLocalDBPart1 failed", "err", err)
			return errors.Cause(err)
		}
	}
	return nil
}

func delOnePrefix(localDB dbm.DB, prefix string, total int) (err error) {
	allDeleted := false
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
		tradelog.Debug("delOnePrefixLimit", "KEY", string(keys[i]))
		batch.Delete(keys[i])
	}
	err = batch.Write()
	if err != nil {
		return allDeleted, errors.Wrap(err, "batch.Write when delete prefix keys: "+prefix)
	}

	return count < total, nil
}

// UpdateLocalDBPart2 升级order
// order 从 v1 升级到 v2
// 通过tableV1 删除， 通过tableV2 添加, 无需通过每个区块扫描对应的交易
func UpdateLocalDBPart2(kvdb dbm.KVDB, total int) error {
	return upgradeOrder(kvdb, total)
}

func upgradeOrder(kvdb dbm.KVDB, total int) (err error) {
	allDeleted := false
	for !allDeleted {
		allDeleted, err = upgradeOrderLimit(kvdb, total)
		if err != nil {
			return err
		}
	}
	return nil
}

func upgradeOrderLimit(kvdb dbm.KVDB, total int) (allDeleted bool, err error) {
	tab2 := NewOrderTableV2(kvdb)
	tab := NewOrderTable(kvdb)
	q1 := tab.GetQuery(kvdb)

	var order1 pty.LocalOrder
	rows, err := q1.List("key", &order1, []byte(""), int32(total), 0)
	if err != nil && err != types.ErrNotFound {
		return false, errors.Wrap(err, "upgradeOrderLimit list from order v1 table")
	}
	for _, row := range rows {
		o1, ok := row.Data.(*pty.LocalOrder)
		if !ok {
			return false, errors.Wrap(types.ErrTypeAsset, "decode order v1")
		}
		err = tab2.Add(o1)
		if err != nil {
			return false, errors.Wrap(err, "upgradeOrderLimit add to order v2 table")
		}
		err = tab.Del([]byte(o1.GetKey()))
		if err != nil {
			return false, errors.Wrap(err, "upgradeOrderLimit add to order v2 table")
		}
	}
	kvs, err := tab2.Save()
	if err != nil {
		return false, errors.Wrap(err, "upgradeOrderLimit save-add to order v2 table")
	}
	kvs2, err := tab.Save()
	if err != nil {
		return false, errors.Wrap(err, "upgradeOrderLimit save-del to order v1 table")
	}
	kvs = append(kvs, kvs2...)

	kvdb.Begin()
	for _, kv := range kvs {
		tradelog.Debug("upgradeOrderLimit", "KEY", string(kv.GetKey()))
		err = kvdb.Set(kv.GetKey(), kv.GetValue())
		if err != nil {
			break
		}
	}
	if err != nil {
		kvdb.Rollback()
		return false, errors.Wrap(err, "upgradeOrderLimit kvdb set")
	}
	err = kvdb.Commit()
	if err != nil {
		kvdb.Rollback()
		return false, errors.Wrap(err, "upgradeOrderLimit kvdb set")
	}
	return len(rows) < total, nil
}

// localdb Version
func getVersion(kvdb dbm.KV) (int, error) {
	value, err := kvdb.Get([]byte(tradeLocaldbVersioin))
	if err != nil && err != types.ErrNotFound {
		return 1, err
	}
	if err == types.ErrNotFound {
		return 1, nil
	}
	var v types.Int32
	err = types.Decode(value, &v)
	if err != nil {
		return 1, err
	}
	return int(v.Data), nil
}

func setVersion(kvdb dbm.KV, version int) error {
	v := types.Int32{Data: int32(version)}
	x := types.Encode(&v)
	return kvdb.Set([]byte(tradeLocaldbVersioin), x)
}

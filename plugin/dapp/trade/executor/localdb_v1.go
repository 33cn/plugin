package executor

import (
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
	err := TradeUpdateLocalDBV2(localDB)
	if err != nil {
		tradelog.Error("Upgrade failed", "err", err)
		return errors.Cause(err)
	}
	return nil
}

// TradeUpdateLocalDBV2 trade 本地数据库升级
// from 1 to 2
func TradeUpdateLocalDBV2(localDB dbm.KVDB) error {
	toVersion := 2
	version, err := getVersion(localDB)
	if err != nil {
		errors.Wrap(err, "TradeUpdateLocalDBV2 get version")
		return err
	}
	if version >= toVersion {
		return nil
	}

	err = UpdateLocalDBPart2(localDB)
	if err != nil {
		errors.Wrap(err, "TradeUpdateLocalDBV2 UpdateLocalDBPart2")
		return err
	}

	err = UpdateLocalDBPart1(localDB)
	if err != nil {
		errors.Wrap(err, "TradeUpdateLocalDBV2 UpdateLocalDBPart1")
		return err
	}
	err = setVersion(localDB, toVersion)
	if err != nil {
		errors.Wrap(err, "TradeUpdateLocalDBV2 setVersion")
		return err
	}
	return nil
}

// UpdateLocalDBPart1 手动生成KV，需要在原有数据库中删除
func UpdateLocalDBPart1(localDB dbm.KVDB) error {
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
			errors.Wrapf(err, "UpdateLocalDBPart1 delOnePrefix: %s", prefix)
			return err
		}
	}
	return nil
}

// delOnePrefix  删除指定前缀的记录
func delOnePrefix(localDB dbm.KVDB, prefix string) error {
	start := []byte(prefix)
	keys, err := localDB.List(start, nil, 0, dbm.ListASC|dbm.ListKeyOnly)
	if err != nil {
		return err
	}
	for _, key := range keys {
		err = localDB.Set(key, nil)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateLocalDBPart2 升级order
// order 从 v1 升级到 v2
// 通过tableV1 删除， 通过tableV2 添加, 无需通过每个区块扫描对应的交易
func UpdateLocalDBPart2(kvdb dbm.KVDB) error {
	return upgradeOrder(kvdb)
}

func upgradeOrder(kvdb dbm.KVDB) (err error) {
	tab2 := NewOrderTableV2(kvdb)
	tab := NewOrderTable(kvdb)
	q1 := tab.GetQuery(kvdb)

	var order1 pty.LocalOrder
	rows, err := q1.List("key", &order1, []byte(""), 0, 0)
	if err != nil {
		if err == types.ErrNotFound {
			return nil
		}
		return errors.Wrap(err, "upgradeOrderLimit list from order v1 table")
	}

	for _, row := range rows {
		o1, ok := row.Data.(*pty.LocalOrder)
		if !ok {
			return errors.Wrap(types.ErrTypeAsset, "decode order v1")
		}
		err = tab2.Add(o1)
		if err != nil {
			return errors.Wrap(err, "upgradeOrderLimit add to order v2 table")
		}
		err = tab.Del([]byte(o1.GetKey()))
		if err != nil {
			return errors.Wrap(err, "upgradeOrderLimit add to order v2 table")
		}
	}

	kvs, err := tab2.Save()
	if err != nil {
		return errors.Wrap(err, "upgradeOrderLimit save-add to order v2 table")
	}
	kvs2, err := tab.Save()
	if err != nil {
		return errors.Wrap(err, "upgradeOrderLimit save-del to order v1 table")
	}
	kvs = append(kvs, kvs2...)

	for _, kv := range kvs {
		tradelog.Debug("upgradeOrderLimit", "KEY", string(kv.GetKey()))
		err = kvdb.Set(kv.GetKey(), kv.GetValue())
		if err != nil {
			err = errors.Wrap(err, "upgradeOrderLimit sed localdb")
			break
		}
	}
	if err != nil {
		return errors.Wrap(err, "upgradeOrderLimit kvdb set")
	}
	return nil
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

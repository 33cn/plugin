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
func (t *trade) Upgrade() (*types.LocalDBSet, error) {
	localDB := t.GetLocalDB()
	// 获得默认的coins symbol， 更新数据时用
	coinSymbol := t.GetAPI().GetConfig().GetCoinSymbol()
	kvs, err := UpgradeLocalDBV2(localDB, t.GetAPI().GetConfig().GetCoinExec(), coinSymbol)
	if err != nil {
		tradelog.Error("Upgrade failed", "err", err)
		return nil, errors.Cause(err)
	}
	return kvs, nil
}

// UpgradeLocalDBV2 trade 本地数据库升级
// from 1 to 2
func UpgradeLocalDBV2(localDB dbm.KVDB, coinExec, coinSymbol string) (*types.LocalDBSet, error) {
	toVersion := 2
	tradelog.Info("UpgradeLocalDBV2 upgrade start", "to_version", toVersion)
	version, err := getVersion(localDB)
	if err != nil {
		return nil, errors.Wrap(err, "UpgradeLocalDBV2 get version")
	}
	if version >= toVersion {
		tradelog.Debug("UpgradeLocalDBV2 not need to upgrade", "current_version", version, "to_version", toVersion)
		return nil, nil
	}

	var kvset types.LocalDBSet
	kvs, err := UpgradeLocalDBPart2(localDB, coinExec, coinSymbol)
	if err != nil {
		return nil, errors.Wrap(err, "UpgradeLocalDBV2 UpgradeLocalDBPart2")
	}
	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	}

	kvs, err = UpgradeLocalDBPart1(localDB)
	if err != nil {
		return nil, errors.Wrap(err, "UpgradeLocalDBV2 UpgradeLocalDBPart1")
	}
	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	}

	kvs, err = setVersion(localDB, toVersion)
	if err != nil {
		return nil, errors.Wrap(err, "UpgradeLocalDBV2 setVersion")
	}
	if len(kvs) > 0 {
		kvset.KV = append(kvset.KV, kvs...)
	}

	tradelog.Info("UpgradeLocalDBV2 upgrade done")
	return &kvset, nil
}

// UpgradeLocalDBPart1 手动生成KV，需要在原有数据库中删除
func UpgradeLocalDBPart1(localDB dbm.KVDB) ([]*types.KeyValue, error) {
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

	var allKvs []*types.KeyValue
	for _, prefix := range prefixes {
		kvs, err := delOnePrefix(localDB, prefix)
		if err != nil {
			return nil, errors.Wrapf(err, "UpdateLocalDBPart1 delOnePrefix: %s", prefix)
		}
		if len(kvs) > 0 {
			allKvs = append(allKvs, kvs...)
		}

	}
	return allKvs, nil
}

// delOnePrefix  删除指定前缀的记录
func delOnePrefix(localDB dbm.KVDB, prefix string) ([]*types.KeyValue, error) {
	start := []byte(prefix)
	keys, err := localDB.List(start, nil, 0, dbm.ListASC|dbm.ListKeyOnly)
	if err != nil {
		if err == types.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	var kvs []*types.KeyValue
	tradelog.Debug("delOnePrefix", "len", len(keys), "prefix", prefix)
	for _, key := range keys {
		err = localDB.Set(key, nil)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, &types.KeyValue{Key: key, Value: nil})
	}

	return kvs, nil
}

// UpgradeLocalDBPart2 升级order
// order 从 v1 升级到 v2
// 通过tableV1 删除， 通过tableV2 添加, 无需通过每个区块扫描对应的交易
func UpgradeLocalDBPart2(kvdb dbm.KVDB, coinExec, coinSymbol string) ([]*types.KeyValue, error) {
	return upgradeOrder(kvdb, coinExec, coinSymbol)
}

func upgradeOrder(kvdb dbm.KVDB, coinExec, coinSymbol string) ([]*types.KeyValue, error) {
	tab2 := NewOrderTableV2(kvdb)
	tab := NewOrderTable(kvdb)
	q1 := tab.GetQuery(kvdb)

	var order1 pty.LocalOrder
	rows, err := q1.List("key", &order1, []byte(""), 0, 0)
	if err != nil {
		if err == types.ErrNotFound {
			return nil, nil
		}
		return nil, errors.Wrap(err, "upgradeOrder list from order v1 table")
	}

	tradelog.Debug("upgradeOrder", "len", len(rows))
	for _, row := range rows {
		o1, ok := row.Data.(*pty.LocalOrder)
		if !ok {
			return nil, errors.Wrap(types.ErrTypeAsset, "decode order v1")
		}

		o2 := types.Clone(o1).(*pty.LocalOrder)
		upgradeLocalOrder(o2, coinExec, coinSymbol)
		err = tab2.Add(o2)
		if err != nil {
			return nil, errors.Wrap(err, "upgradeOrder add to order v2 table")
		}

		err = tab.Del([]byte(o1.TxIndex))
		if err != nil {
			return nil, errors.Wrapf(err, "upgradeOrder del from order v1 table, key: %s", o1.TxIndex)
		}
	}

	kvs, err := tab2.Save()
	if err != nil {
		return nil, errors.Wrap(err, "upgradeOrder save-add to order v2 table")
	}
	kvs2, err := tab.Save()
	if err != nil {
		return nil, errors.Wrap(err, "upgradeOrder save-del to order v1 table")
	}
	kvs = append(kvs, kvs2...)

	for _, kv := range kvs {
		tradelog.Debug("upgradeOrder", "KEY", string(kv.GetKey()))
		err = kvdb.Set(kv.GetKey(), kv.GetValue())
		if err != nil {
			return nil, errors.Wrapf(err, "upgradeOrder set localdb key: %s", string(kv.GetKey()))
		}
	}

	return kvs, nil
}

// upgradeLocalOrder 处理两个fork前的升级数据
// 1. 支持任意资产
// 2. 支持任意资产定价
func upgradeLocalOrder(order *pty.LocalOrder, coinExec, coinSymbol string) {
	if order.AssetExec == "" {
		order.AssetExec = defaultAssetExec
	}
	if order.PriceExec == "" {
		order.PriceExec = coinExec
		order.PriceSymbol = coinSymbol
	}
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

func setVersion(kvdb dbm.KV, version int) ([]*types.KeyValue, error) {
	v := types.Int32{Data: int32(version)}
	x := types.Encode(&v)
	err := kvdb.Set([]byte(tradeLocaldbVersioin), x)
	return []*types.KeyValue{{Key: []byte(tradeLocaldbVersioin), Value: x}}, err
}

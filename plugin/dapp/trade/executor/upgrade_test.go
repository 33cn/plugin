package executor

import (
	"bytes"
	"testing"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
)

func Test_Upgrade(t *testing.T) {
	dir, db, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	assert.NotNil(t, localdb)

	// test empty db
	_, err := callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// test again
	setVersion(localdb, 1)
	_, err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// test with data

	// create for test
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
	localdb.Set([]byte(prefixes[0]+"xxxx1"), []byte("xx1"))
	localdb.Set([]byte(prefixes[0]+"xxxx2"), []byte("xx2"))
	localdb.Set([]byte(prefixes[0]+"xxxx3"), []byte("xx3"))
	localdb.Set([]byte(prefixes[1]+"xxxx3"), []byte("xx3"))

	//tabV2 := NewOrderTableV2(localdb)
	tabV1 := NewOrderTable(localdb)
	tabV1.Add(order1)
	tabV1.Add(order2)
	tabV1.Add(order3)
	kvs, err := tabV1.Save()
	assert.Nil(t, err)
	for _, kv := range kvs {
		localdb.Set(kv.Key, kv.Value)
	}

	// 初次升级
	setVersion(localdb, 1)
	kvset, err := callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)
	assert.NotNil(t, kvset)

	// 已经是升级后的版本了， 不需要再升级
	kvset, err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)
	assert.Nil(t, kvset)

	// 先修改版本去升级，但数据已经升级了， 所以处理数据量为0
	setVersion(localdb, 1)
	kvset, err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)
	// 只有version 升级
	assert.Equal(t, 1, len(kvset.KV))

	// just print log
	//assert.NotNil(t, nil)
}

func callUpgradeLocalDBV2(localdb dbm.KVDB) (*types.LocalDBSet, error) {
	return UpgradeLocalDBV2(localdb, "coins", "bty")
}

// 测试更新后是否删除完全， asset 设置
func Test_UpgradeOrderAsset(t *testing.T) {
	dir, db, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	assert.NotNil(t, localdb)

	tabV1 := NewOrderTable(localdb)
	tabV1.Add(order3)
	kvs, err := tabV1.Save()
	assert.Nil(t, err)
	for _, kv := range kvs {
		localdb.Set(kv.Key, kv.Value)
	}

	kvset, err := callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)
	assert.NotNil(t, kvset)

	v1, err := localdb.List([]byte("LODB-trade-order"), nil, 0, dbm.ListASC|dbm.ListWithKey)
	assert.Nil(t, err)
	assert.NotNil(t, v1)

	primaryKey := "000000000000300001"
	prefix := "LODB-trade-order_v2"
	for _, v := range v1 {
		var kv types.KeyValue
		err := types.Decode(v, &kv)
		assert.Nil(t, err)

		// 前缀都是v2， 删除完成测试
		if !bytes.Equal([]byte("LODB-trade-order_v2-d-000000000000300001"), kv.Key) {
			assert.Equal(t, []byte(primaryKey), kv.Value)
			assert.True(t, bytes.HasPrefix(kv.Key, []byte(prefix)))
		}
	}

	// assert 前缀测试
	v, err := localdb.Get([]byte("LODB-trade-order_v2-m-asset-coins.bty_token.CCNY-000000000000300001"))
	assert.Nil(t, err)
	assert.Equal(t, primaryKey, string(v))

	// just print log
	//assert.NotNil(t, nil)
}

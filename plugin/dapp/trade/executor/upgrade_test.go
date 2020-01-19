package executor

import (
	"testing"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
)

func Test_Upgrade(t *testing.T) {
	dir, db, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	assert.NotNil(t, localdb)

	// test empty db
	err := callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// test again
	setVersion(localdb, 1)
	err = callUpgradeLocalDBV2(localdb)
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
	err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// 已经是升级后的版本了， 不需要再升级
	err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// 先修改版本去升级，但数据已经升级了， 所以处理数据量为0
	setVersion(localdb, 1)
	err = callUpgradeLocalDBV2(localdb)
	assert.Nil(t, err)

	// just print log
	//assert.NotNil(t, nil)
}

func callUpgradeLocalDBV2(localdb dbm.KVDB) error {
	return UpgradeLocalDBV2(localdb, "bty")
}

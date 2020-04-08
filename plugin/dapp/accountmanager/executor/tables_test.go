package executor

import (
	"testing"
	"time"

	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

func TestAccountTable(t *testing.T) {
	_, _, kvdb := util.CreateTestDB()
	table := NewAccountTable(kvdb)
	now := time.Now().Unix()
	row1 := &et.Account{Index: now * int64(types.MaxTxsPerBlock), AccountID: "harry2015", Status: 0, ExpireTime: now + 10, Addr: "xxxx"}
	row2 := &et.Account{Index: now*int64(types.MaxTxsPerBlock) + 1, AccountID: "harry2020", Status: 0, ExpireTime: now, Addr: "xxxx"}
	table.Add(row1)
	table.Add(row2)
	kvs, err := table.Save()
	if err != nil {
		t.Error(err)
	}
	for _, kv := range kvs {
		kvdb.Set(kv.Key, kv.Value)
	}
	time.Sleep(2 * time.Second)
	list, err := findAccountListByIndex(kvdb, time.Now().Unix()+10, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(list)
	list, err = findAccountListByStatus(kvdb, et.Normal, 0, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(list)
	row1.Status = et.Frozen
	err = table.Replace(row1)
	if err != nil {
		t.Error(err)
	}
	kvs, err = table.Save()
	if err != nil {
		t.Error(err)
	}
	for _, kv := range kvs {
		kvdb.Set(kv.Key, kv.Value)
	}
	list, err = findAccountListByStatus(kvdb, et.Frozen, 0, "")
	if err != nil {
		t.Error(err)
	}
	t.Log(list)

}

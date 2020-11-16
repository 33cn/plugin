package state

import (
	"strings"

	chain33db "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
)

var (
	currentExecTxHash string
	localDB           chain33db.DB
)

func newMemDB() chain33db.DB {
	memdb, err := chain33db.NewGoMemDB("", "", 0)
	if err != nil {
		panic(err)
	}
	return memdb
}

func setCurrentTx(txhashNew string) {
	if 0 == strings.Compare(txhashNew, currentExecTxHash) {
		return
	}
	currentExecTxHash = txhashNew
	localDB = newMemDB()
}

func getLocalValue(key []byte, txHash string) ([]byte, error) {
	setCurrentTx(txHash)
	return localDB.Get(key)
}

func setLocalValue(key, value []byte, txHash string) error {
	setCurrentTx(txHash)
	return localDB.Set(key, value)
}

//注意该接口只能在执行本地交易查询时使用，否则会破坏数据
func GetAllLocalKeyValues(txhashNew string) []*types.KeyValue {
	if txhashNew != currentExecTxHash {
		return nil
	}

	goMemDB, ok := localDB.(*chain33db.GoMemDB)
	if !ok {
		return nil
	}

	var kvs []*types.KeyValue
	it := goMemDB.DB().NewIterator(nil)
	for it.Next() {
		kvs = append(kvs, &types.KeyValue{Key: it.Key(), Value: it.Value()})
	}
	it.Release()
	return kvs
}

//该函数只是方便用来帮助进行单元测试，不可以在正常业务逻辑中使用
func SetCurrentTx4UT(txhashNew string) {
	currentExecTxHash = txhashNew
}

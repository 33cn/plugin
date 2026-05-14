package executor

import "fmt"

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-lightclient-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-lightclient-"
)

// statedb

func btcLastHeaderKey() []byte {
	return []byte(KeyPrefixStateDB + "btc-lastheader")
}

// localdb

func btcHeaderKey(height uint64) []byte {
	return []byte(KeyPrefixLocalDB + fmt.Sprintf("btc-header-%010d", height))
}

func btcHeaderHashHeightKey(hash string) []byte {
	return []byte(KeyPrefixLocalDB + "btc-hash2height-" + hash)
}

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
)

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-evmxgo-statedb"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-evmxgo-"

	evmxgoCreatedSTONewLocal = "LODB-evmxgo-create-sto-"
)

func calcEvmxgoKey(value string) (key []byte) {
	return []byte(fmt.Sprintf(KeyPrefixStateDB+"-%s", value))
}

func calcEvmxgoKeyLocal() []byte {
	return []byte(evmxgoCreatedSTONewLocal)
}

func calcEvmxgoStatusKeyLocal(token string) []byte {
	return []byte(fmt.Sprintf(evmxgoCreatedSTONewLocal+"%s", token))
}

//存储地址上收币的信息
func calcAddrKey(token string, addr string) []byte {
	return []byte(fmt.Sprintf("LODB-evmxgo-%s-Addr:%s", token, address.FormatAddrKey(addr)))
}

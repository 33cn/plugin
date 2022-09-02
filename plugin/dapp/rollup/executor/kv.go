package executor

import "fmt"

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-rollup-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-rollup-"
)

func formatRollupStatusKey(title string) []byte {
	return []byte(KeyPrefixStateDB + fmt.Sprintf("%s-status", title))
}

func formatCommitRoundInfoKey(title string, round int64) []byte {
	return []byte(KeyPrefixStateDB + fmt.Sprintf("%s-roundinfo-%018d", title, round))
}

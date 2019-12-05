package executor

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-storage-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-storage-"
)

// Key Storage to save key
func Key(txHash string) (key []byte) {
	key = append(key, []byte(KeyPrefixStateDB)...)
	key = append(key, []byte(txHash)...)
	return key
}

package executor

import "fmt"

// 生成 key -> id 格式的本地数据库数据， 在下个版本这个文件可以全部删除
// 由于数据库精简需要保存具体数据

// 将手动生成的local db 的代码和用table 生成的local db的代码分离出来
// 手动生成的local db, 将不生成任意资产标价的数据， 保留用coins 生成交易的数据， 来兼容为升级的app 应用
// 希望有全量数据的， 需要调用新的rpc

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

// UpdateLocalDBPart1 手动生成KV，需要在原有数据库中删除
// TODO
func UpdateLocalDBPart1() {
	prefix := []string{
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
	fmt.Printf("%+v", prefix)
}

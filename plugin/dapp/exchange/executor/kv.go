package executor

import (
	"fmt"

	"github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 用户合约存取kv数据时，key值前缀需要满足一定规范
 * 即key = keyPrefix + userKey
 * 需要字段前缀查询时，使用’-‘作为分割符号
 */

var (
	//KeyPrefixStateDB state db key必须前缀
	KeyPrefixStateDB = "mavl-exchange-"
	//KeyPrefixLocalDB local db的key必须前缀
	KeyPrefixLocalDB = "LODB-exchange-"
)

//状态数据库中存储具体挂单信息
func calcOrderKey(orderID string) []byte {
	key := fmt.Sprintf("%s"+"orderID:%s", KeyPrefixStateDB, orderID)
	return []byte(key)
}
func calcMarketDepthPrefix(left, right *types.Asset, op int32) []byte {
	key := fmt.Sprintf("%s"+"depth-%s-%s-%d:", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), op)
	return []byte(key)
}

//市场深度
func calcMarketDepthKey(left, right *types.Asset, op int32, price float32) []byte {
	// 设置精度为1e8
	key := fmt.Sprintf("%s"+"depth-%s-%s-%d:%016d", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), op, int64(Truncate(price)*float32(1e8)))
	return []byte(key)
}

func calcMarketDepthOrderPrefix(left, right *types.Asset, op int32, price float32) []byte {
	// 设置精度为1e8
	key := fmt.Sprintf("%s"+"order-%s-%s-%d:%016d", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), op, int64(Truncate(price)*float32(1e8)))
	return []byte(key)
}

// localdb中存储市场挂单ID
func calcMarketDepthOrderKey(left, right *types.Asset, op int32, price float32, index int64) []byte {
	// 设置精度为1e8
	key := fmt.Sprintf("%s"+"order-%s-%s-%d:%016d:%018d", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), op, int64(Truncate(price)*float32(1e8)), index)
	return []byte(key)
}

//最新已经成交的订单，这里状态固定都是完成状态,这个主要给外部使用，可以查询最新得成交信息
func calcCompletedOrderKey(left, right *types.Asset, index int64) []byte {
	// 设置精度为1e8
	key := fmt.Sprintf("%s"+"completed-%s-%s-%d:%018d", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), types.Completed, index)
	return []byte(key)
}
func calcCompletedOrderPrefix(left, right *types.Asset) []byte {
	// 设置精度为1e8
	key := fmt.Sprintf("%s"+"completed-%s-%s-%d:", KeyPrefixLocalDB, left.GetSymbol(), right.GetSymbol(), types.Completed)
	return []byte(key)
}

//根据地址和订单状态，去查询订单列表,包含所有交易对
func calcUserOrderIDPrefix(status int32, addr string) []byte {
	key := fmt.Sprintf("%s"+"addr:%s:%d:", KeyPrefixLocalDB, addr, status)
	return []byte(key)
}
func calcUserOrderIDKey(status int32, addr string, index int64) []byte {
	key := fmt.Sprintf("%s"+"addr:%s:%d:%018d", KeyPrefixLocalDB, addr, status, index)
	return []byte(key)
}

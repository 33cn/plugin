package executor

import (
	"strconv"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

// 生成 key -> id 格式的本地数据库数据， 在下个版本这个文件可以全部删除
// 由于数据库精简需要保存具体数据

// 将手动生成的local db 的代码和用table 生成的local db的代码分离出来
// 手动生成的local db, 将不生成任意资产标价的数据， 保留用coins 生成交易的数据， 来兼容为升级的app 应用
// 希望有全量数据的， 需要调用新的rpc

func genSellOrderKeyValue(kv []*types.KeyValue, sellorder *pty.SellOrder, status int32, value []byte) []*types.KeyValue {
	newkey := calcTokenSellOrderKey(sellorder.TokenSymbol, sellorder.Address, status, sellorder.SellID, sellorder.Height)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesSellOrderKeyStatus(sellorder.TokenSymbol, sellorder.Address, status, sellorder.SellID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesSellOrderKeyToken(sellorder.TokenSymbol, sellorder.Address, status, sellorder.SellID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcTokensSellOrderKeyStatus(sellorder.TokenSymbol, status,
		calcPriceOfToken(sellorder.PricePerBoardlot, sellorder.AmountPerBoardlot), sellorder.Address, sellorder.SellID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	st, ty := fromStatus(status)
	newkey = calcOnesOrderKey(sellorder.Address, st, ty, sellorder.Height, sellorder.SellID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	return kv
}

func genBuyMarketOrderKeyValue(kv []*types.KeyValue, receipt *pty.ReceiptBuyBase,
	status int32, height int64, value []byte) []*types.KeyValue {

	keyID := receipt.TxHash

	newkey := calcTokenBuyOrderKey(receipt.TokenSymbol, receipt.Owner, status, keyID, height)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesBuyOrderKeyStatus(receipt.TokenSymbol, receipt.Owner, status, keyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesBuyOrderKeyToken(receipt.TokenSymbol, receipt.Owner, status, keyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	priceBoardlot, err := strconv.ParseFloat(receipt.PricePerBoardlot, 64)
	if err != nil {
		panic(err)
	}
	priceBoardlotInt64 := int64(priceBoardlot * float64(types.TokenPrecision))
	AmountPerBoardlot, err := strconv.ParseFloat(receipt.AmountPerBoardlot, 64)
	if err != nil {
		panic(err)
	}
	AmountPerBoardlotInt64 := int64(AmountPerBoardlot * float64(types.Coin))
	price := calcPriceOfToken(priceBoardlotInt64, AmountPerBoardlotInt64)

	newkey = calcTokensBuyOrderKeyStatus(receipt.TokenSymbol, status,
		price, receipt.Owner, keyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	st, ty := fromStatus(status)
	newkey = calcOnesOrderKey(receipt.Owner, st, ty, height, keyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	return kv
}

func genBuyLimitOrderKeyValue(kv []*types.KeyValue, buyOrder *pty.BuyLimitOrder, status int32, value []byte) []*types.KeyValue {
	newkey := calcTokenBuyOrderKey(buyOrder.TokenSymbol, buyOrder.Address, status, buyOrder.BuyID, buyOrder.Height)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesBuyOrderKeyStatus(buyOrder.TokenSymbol, buyOrder.Address, status, buyOrder.BuyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcOnesBuyOrderKeyToken(buyOrder.TokenSymbol, buyOrder.Address, status, buyOrder.BuyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	newkey = calcTokensBuyOrderKeyStatus(buyOrder.TokenSymbol, status,
		calcPriceOfToken(buyOrder.PricePerBoardlot, buyOrder.AmountPerBoardlot), buyOrder.Address, buyOrder.BuyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	st, ty := fromStatus(status)
	newkey = calcOnesOrderKey(buyOrder.Address, st, ty, buyOrder.Height, buyOrder.BuyID)
	kv = append(kv, &types.KeyValue{Key: newkey, Value: value})

	return kv
}

package spot

import (
	"math/big"

	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

//SafeMul Safe multiplication of large numbers, prevent overflow
func SafeMul(x, y, coinPrecision int64) int64 {
	res := big.NewInt(0).Mul(big.NewInt(x), big.NewInt(y))
	res = big.NewInt(0).Div(res, big.NewInt(coinPrecision))
	return res.Int64()
}

//SafeAdd Safe add
func SafeAdd(x, y int64) int64 {
	res := big.NewInt(0).Add(big.NewInt(x), big.NewInt(y))
	return res.Int64()
}

//Calculate the average transaction price
func caclAVGPrice(order *et.SpotOrder, price int64, amount int64) int64 {
	x := big.NewInt(0).Mul(big.NewInt(order.AVGPrice), big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()))
	y := big.NewInt(0).Mul(big.NewInt(price), big.NewInt(amount))
	total := big.NewInt(0).Add(x, y)
	div := big.NewInt(0).Add(big.NewInt(order.GetLimitOrder().Amount-order.GetBalance()), big.NewInt(amount))
	avg := big.NewInt(0).Div(total, div)
	return avg.Int64()
}

//计Calculation fee
func calcMtfFee(cost int64, rate int32, coinPrecision int64) int64 {
	fee := big.NewInt(0).Mul(big.NewInt(cost), big.NewInt(int64(rate)))
	fee = big.NewInt(0).Div(fee, big.NewInt(coinPrecision))
	return fee.Int64()
}

//CalcActualCost Calculate actual cost
func CalcActualCost(op int32, amount int64, price, coinPrecision int64) int64 {
	if op == et.OpBuy {
		return SafeMul(amount, price, coinPrecision)
	}
	return amount
}

package wallet

import (
	"math/big"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func GetNftOrderMsg(payload *zt.SpotNftOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetNftTakerOrderMsg(payload *zt.SpotNftTakerOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetNftOrder2Msg(payload *zt.SpotNftOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetNftTakerOrder2Msg(payload *zt.SpotNftTakerOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetLimitOrderMsg(payload *zt.SpotLimitOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetAssetLimitOrderMsg(payload *zt.SpotAssetLimitOrder) *zt.ZkMsg {
	return GetSpotSwapMsg(payload.Order)
}

func GetSpotSwapMsg(order *zt.ZkOrder) *zt.ZkMsg {
	var pubData []uint

	binaryData := make([]uint, zt.MsgWidth)

	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(zt.TySwapAction), zt.TxTypeBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(order.AccountID), zt.AccountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(uint64(order.TokenSell)), zt.TokenBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(new(big.Int).SetUint64(uint64(order.TokenBuy)), zt.TokenBitWidth)...)

	ratio1, _ := big.NewInt(0).SetString(order.Ratio1, 10)
	ratio2, _ := big.NewInt(0).SetString(order.Ratio2, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(ratio1, zt.AmountBitWidth)...)
	pubData = append(pubData, getBigEndBitsWithFixLen(ratio2, zt.AmountBitWidth)...)

	//merkel tree 上的精度都是1e18 和eth保持一致， 这里的amount是1e8精度
	amount, _ := big.NewInt(0).SetString(order.Amount, 10)
	pubData = append(pubData, getBigEndBitsWithFixLen(amount, zt.AmountBitWidth)...)
	copy(binaryData, pubData)

	return &zt.ZkMsg{
		First:  setBeBitsToVal(binaryData[:zt.MsgFirstWidth]),
		Second: setBeBitsToVal(binaryData[zt.MsgFirstWidth : zt.MsgFirstWidth+zt.MsgSecondWidth]),
		Third:  setBeBitsToVal(binaryData[zt.MsgFirstWidth+zt.MsgSecondWidth:]),
	}
}

package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// CheckTx 实现自定义检验交易接口，供框架调用
func SpotCheckTx(cfg *types.Chain33Config, tx *types.Transaction, index int) error {
	//发送交易的时候就检查payload,做严格的参数检查
	var exchange et.ZksyncAction
	types.Decode(tx.GetPayload(), &exchange)
	if exchange.Ty == et.TyLimitOrderAction {
		limitOrder := exchange.GetLimitOrder()
		return et.CheckLimitOrder(cfg, limitOrder)
	}
	if exchange.Ty == et.TyNftOrderAction {
		nftOrder := exchange.GetNftOrder()
		if !et.CheckIsNFTToken(nftOrder.LeftAsset) || !et.CheckIsNormalToken(nftOrder.RightAsset) {
			return types.ErrTypeAsset
		}
	}
	if exchange.Ty == et.TyMarketOrderAction {
		return types.ErrActionNotSupport
	}
	return nil
}

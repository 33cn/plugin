package executor

import (
	"time"

	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

// 限价交易
func (e *zksync) Exec_LimitOrder(payload *et.SpotLimitOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	// checkTx will check payload and zk Signature
	start := time.Now()
	action := NewZkSpotDex(e, tx, index)
	r, err := action.LimitOrder(&e.DriverBase, payload, "")
	if err != nil {
		return r, err
	}
	// 构造 LimitOrder 的结算清单
	list := GetSpotMatch(r)
	end := time.Now()
	zlog.Error("zkspot Exec_LimitOrder.LimitOrder", "cost", end.Sub(start))

	action2 := NewAction(e, tx, index)
	r2, err := action2.SpotMatch(payload, list)
	if err != nil {
		return r, err
	}
	end2 := time.Now()
	zlog.Error("zkspot Exec_LimitOrder.SpotMatch", "cost", end2.Sub(start))

	return mergeReceipt(r, r2), nil
}

// 限价交易
func (e *zksync) Exec_AssetLimitOrder(payload *et.SpotAssetLimitOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	// checkTx will check payload and zk Signature
	start := time.Now()
	action := NewZkSpotDex(e, tx, index)
	r, err := action.AssetLimitOrder(&e.DriverBase, payload, "")
	if err != nil {
		return r, err
	}
	// 构造 LimitOrder 的结算清单
	list := GetSpotMatch(r)
	end := time.Now()
	zlog.Error("zkspot Exec_AssetLimitOrder.LimitOrder", "cost", end.Sub(start))

	action2 := NewAction(e, tx, index)
	r2, err := action2.AssetMatch(payload, list)
	if err != nil {
		return r, err
	}
	end2 := time.Now()
	zlog.Error("zkspot Exec_AssetLimitOrder.SpotMatch", "cost", end2.Sub(start))

	return mergeReceipt(r, r2), nil
}

// 撤单
func (e *zksync) Exec_RevokeOrder(payload *et.SpotRevokeOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewZkSpotDex(e, tx, index)
	return action.RevokeOrder(payload, "")
}

/*

//市价交易
func (e *zksync) Exec_MarketOrder(payload *et.SpotMarketOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	//TODO marketOrder
	return nil, types.ErrActionNotSupport
}

// 绑定委托交易地址
func (e *zksync) Exec_ExchangeBind(payload *et.SpotExchangeBind, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewZkSpotDex(e, tx, index)
	return actiondb.ExchangeBind(payload)
}


// 委托交易
func (e *zksync) Exec_EntrustOrder(payload *et.SpotEntrustOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewZkSpotDex(e, tx, index)
	r, err := action.EntrustOrder(&e.DriverBase, payload)
	if err != nil {
		return r, err
	}
	// 构造 LimitOrder 的结算清单
	list := GetSpotMatch(r)

	action2 := NewAction(e, tx, index)
	// TODO 这里参数如何统一
	limitOrder := &et.SpotLimitOrder{
		LeftAsset:  payload.LeftAsset,
		RightAsset: payload.RightAsset,
		Price:      payload.Price,
		Amount:     payload.Amount,
		Op:         payload.Op,
		Order:      payload.Order,
	}
	r2, err := action2.SpotMatch(limitOrder, list)
	if err != nil {
		return r, err
	}
	return mergeReceipt(r, r2), nil
}

// 委托撤单
func (e *zksync) Exec_EntrustRevokeOrder(payload *et.SpotEntrustRevokeOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := NewZkSpotDex(e, tx, index)
	ee := a.newEntrust()
	err := ee.CheckBind(payload.Addr)
	if err != nil {
		return nil, err
	}
	p := et.SpotRevokeOrder{
		OrderID: payload.OrderID,
	}
	action := NewZkSpotDex(e, tx, index)
	return action.RevokeOrder(&p, action.txinfo.From)
}
*/
// 限价交易
func (e *zksync) Exec_NftOrder(payload *et.SpotNftOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Error("Exec_NftOrder", "err", err, "stack", et.GetStack())
		}
	}()
	action := NewZkSpotDex(e, tx, index)
	return action.NftOrder(&e.DriverBase, payload, "", int32(et.AssetType_ZkNft))
}

// 限价交易
func (e *zksync) Exec_NftOrder2(payload *et.SpotNftOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Error("Exec_NftOrder2", "err", err, "stack", et.GetStack())
		}
	}()
	action := NewZkSpotDex(e, tx, index)
	return action.NftOrder(&e.DriverBase, payload, "", int32(et.AssetType_EvmNft))
}

// 限价交易
func (e *zksync) Exec_NftTakerOrder(payload *et.SpotNftTakerOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	return e.nftTakerOrder(payload, tx, index, et.TyNftOrderAction)
}

// 限价交易
func (e *zksync) Exec_NftTakerOrder2(payload *et.SpotNftTakerOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	return e.nftTakerOrder(payload, tx, index, et.TyNftTakerOrder2Action)
}

func (e *zksync) nftTakerOrder(payload *et.SpotNftTakerOrder, tx *types.Transaction, index int, nftType int) (*types.Receipt, error) {
	defer func() {
		if err := recover(); err != nil {
			zlog.Error("Exec_NftTakerOrder", "err", err, "stack", et.GetStack())
		}
	}()
	// checkTx will check payload and zk Signature
	start := time.Now()
	action := NewZkSpotDex(e, tx, index)
	r, err := action.NftTakerOrder(&e.DriverBase, payload, "", nftType)
	if err != nil {
		return r, err
	}
	// 构造 LimitOrder 的结算清单
	list := GetSpotMatch(r)
	end := time.Now()
	zlog.Error("zkspot Exec_NftTakerOrder.NftTakerOrder", "cost", end.Sub(start))

	action2 := NewAction(e, tx, index)
	r2, err := action2.SpotNftMatch(payload, list)
	if err != nil {
		return r, err
	}
	end2 := time.Now()
	zlog.Error("zkspot Exec_LimitOrder.SpotMatch", "cost", end2.Sub(start))

	return mergeReceipt(r, r2), nil
}

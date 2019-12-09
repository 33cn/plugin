package executor

import (
	"github.com/33cn/chain33/types"
	exchangetypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (e *exchange) ExecLocal_LimitOrder(payload *exchangetypes.LimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case exchangetypes.TyLimitOrderLog:
				receipt := &exchangetypes.ReceiptExchange{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				kv := e.updateIndex(receipt)
				dbSet.KV = append(dbSet.KV, kv...)
			}
		}
	}
	return e.addAutoRollBack(tx, dbSet.KV), nil
}

func (e *exchange) ExecLocal_MarketOrder(payload *exchangetypes.MarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case exchangetypes.TyMarketOrderLog:
				receipt := &exchangetypes.ReceiptExchange{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				kv := e.updateIndex(receipt)
				dbSet.KV = append(dbSet.KV, kv...)
			}
		}
	}
	return e.addAutoRollBack(tx, dbSet.KV), nil
}

func (e *exchange) ExecLocal_RevokeOrder(payload *exchangetypes.RevokeOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case exchangetypes.TyRevokeOrderLog:
				receipt := &exchangetypes.ReceiptExchange{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				kv := e.updateIndex(receipt)
				dbSet.KV = append(dbSet.KV, kv...)
			}
		}
	}
	return e.addAutoRollBack(tx, dbSet.KV), nil
}

//设置自动回滚
func (e *exchange) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

func (e *exchange) updateIndex(receipt *exchangetypes.ReceiptExchange) (kvs []*types.KeyValue) {
	switch receipt.Order.Status {
	case exchangetypes.Ordered:
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		price := receipt.GetOrder().GetLimitOrder().GetPrice()
		oderID := receipt.GetOrder().OrderID
		index := receipt.GetIndex()
		addr := receipt.GetOrder().Addr
		var markDepth exchangetypes.MarketDepth
		err := findObject(e.GetLocalDB(), calcMarketDepthKey(left, right, op, price), &markDepth)
		if err == types.ErrNotFound {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = receipt.Order.Balance
		} else {
			markDepth.Amount = markDepth.Amount + receipt.Order.Balance
		}

		//marketDepth
		kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, op, price), Value: types.Encode(&markDepth)})
		//orderID
		kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthOrderKey(left, right, op, price, receipt.Index), Value: types.Encode(&exchangetypes.OrderID{ID: oderID, Index: index})})
		//addr orderID
		kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Ordered, addr, index), Value: types.Encode(&exchangetypes.OrderID{ID: oderID, Index: index})})

		if len(receipt.MatchOrders) > 0 {
			//撮合交易更新
			var balance int64
			for i, matchOrder := range receipt.MatchOrders {
				if matchOrder.Status == exchangetypes.Completed {
					// 删除原有状态orderID
					kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthOrderKey(left, right, matchOrder.GetLimitOrder().Op, price, matchOrder.Index), Value: nil})
					//删除原有状态orderID
					kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Ordered, matchOrder.Addr, matchOrder.Index), Value: nil})
					//更新状态为已完成,索引index,改为当前的index
					kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Completed, matchOrder.Addr, index+int64(i+1)), Value: types.Encode(&exchangetypes.OrderID{ID: matchOrder.OrderID, Index: index + int64(i+1)})})
					//calcCompletedOrderKey
					kvs = append(kvs, &types.KeyValue{Key: calcCompletedOrderKey(left, right, index+int64(i+1)), Value: types.Encode(&exchangetypes.OrderID{ID: matchOrder.OrderID, Index: index + int64(i+1)})})
				}
				if matchOrder.Status == exchangetypes.Ordered {
					//只需统一更改市场深度状态，其他不需要处理
					balance = balance + matchOrder.Balance
				}
			}
			//更改匹配市场深度
			var matchDepth exchangetypes.MarketDepth
			err = findObject(e.GetLocalDB(), calcMarketDepthKey(left, right, OpSwap(op), price), &matchDepth)
			if err == types.ErrNotFound {
				matchDepth.Price = price
				matchDepth.LeftAsset = left
				matchDepth.RightAsset = right
				matchDepth.Op = OpSwap(op)
				matchDepth.Amount = balance
			} else {
				matchDepth.Amount = matchDepth.Amount - receipt.Order.Executed
			}
			//marketDepth
			kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, OpSwap(op), price), Value: types.Encode(&matchDepth)})
			if matchDepth.Amount == 0 {
				//删除
				kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, OpSwap(op), price), Value: nil})
			}
		}
		return
	case exchangetypes.Completed:
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		price := receipt.GetOrder().GetLimitOrder().GetPrice()
		oderID := receipt.GetOrder().OrderID
		index := receipt.GetIndex()
		addr := receipt.GetOrder().Addr

		//user orderID
		kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Completed, addr, index), Value: types.Encode(&exchangetypes.OrderID{ID: oderID, Index: index})})

		//calcCompletedOrderKey
		kvs = append(kvs, &types.KeyValue{Key: calcCompletedOrderKey(left, right, index), Value: types.Encode(&exchangetypes.OrderID{ID: oderID, Index: index})})

		if len(receipt.MatchOrders) > 0 {
			//撮合交易更新
			var balance int64
			for i, matchOrder := range receipt.MatchOrders {
				if matchOrder.Status == exchangetypes.Completed {
					// 删除原有状态orderID
					kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthOrderKey(left, right, matchOrder.GetLimitOrder().Op, price, matchOrder.Index), Value: nil})
					//删除原有状态orderID
					kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Ordered, matchOrder.Addr, matchOrder.Index), Value: nil})
					//更新状态为已完成,更新索引
					kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Completed, matchOrder.Addr, index+int64(i+1)), Value: types.Encode(&exchangetypes.OrderID{ID: matchOrder.OrderID, Index: index + int64(i+1)})})
					//calcCompletedOrderKey
					kvs = append(kvs, &types.KeyValue{Key: calcCompletedOrderKey(left, right, index+int64(i+1)), Value: types.Encode(&exchangetypes.OrderID{ID: matchOrder.OrderID, Index: index + int64(i+1)})})

				}
				if matchOrder.Status == exchangetypes.Ordered {
					//只需统一更改市场深度状态，其他不需要处理
					balance = balance + matchOrder.Balance
				}
			}
			//更改match市场深度
			var matchDepth exchangetypes.MarketDepth
			err := findObject(e.GetLocalDB(), calcMarketDepthKey(left, right, OpSwap(op), price), &matchDepth)
			if err == types.ErrNotFound {
				matchDepth.Price = price
				matchDepth.LeftAsset = left
				matchDepth.RightAsset = right
				matchDepth.Op = OpSwap(op)
				matchDepth.Amount = balance
			} else {
				matchDepth.Amount = matchDepth.Amount - receipt.Order.Executed
			}
			//marketDepth
			kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, OpSwap(op), price), Value: types.Encode(&matchDepth)})
			if matchDepth.Amount == 0 {
				//删除
				kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, OpSwap(op), price), Value: nil})
			}
		}
		return
	case exchangetypes.Revoked:
		//只有状态时ordered状态的订单才能被撤回
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		price := receipt.GetOrder().GetLimitOrder().GetPrice()
		oderID := receipt.GetOrder().OrderID
		index := receipt.GetIndex()
		addr := receipt.GetOrder().Addr
		var marketDepth exchangetypes.MarketDepth
		err := findObject(e.GetLocalDB(), calcMarketDepthKey(left, right, op, price), &marketDepth)
		if err == nil {
			//marketDepth
			marketDepth.Amount = marketDepth.Amount - receipt.GetOrder().Balance
			elog.Error("revoked", "recept.order.balance", receipt.GetOrder().Balance)
			kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, op, price), Value: types.Encode(&marketDepth)})
		}
		elog.Error("revoked", "marketDepth.Amount", marketDepth.Amount)
		if marketDepth.Amount == 0 {
			//删除
			kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthKey(left, right, op, price), Value: nil})
		}
		// 删除原有状态orderID
		kvs = append(kvs, &types.KeyValue{Key: calcMarketDepthOrderKey(left, right, op, price, receipt.GetOrder().Index), Value: nil})
		//删除原有状态orderID
		kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Ordered, addr, receipt.GetOrder().Index), Value: nil})
		//添加撤销的订单
		kvs = append(kvs, &types.KeyValue{Key: calcUserOrderIDKey(exchangetypes.Revoked, addr, index), Value: types.Encode(&exchangetypes.OrderID{ID: oderID, Index: index})})
	}
	return
}
func OpSwap(op int32) int32 {
	if op == exchangetypes.OpBuy {
		return exchangetypes.OpSell
	} else {
		return exchangetypes.OpBuy
	}
}

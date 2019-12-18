package executor

import (
	"github.com/33cn/chain33/types"
	ety "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (e *exchange) ExecLocal_LimitOrder(payload *ety.LimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case ety.TyLimitOrderLog:
				receipt := &ety.ReceiptExchange{}
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

func (e *exchange) ExecLocal_MarketOrder(payload *ety.MarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case ety.TyMarketOrderLog:
				receipt := &ety.ReceiptExchange{}
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

func (e *exchange) ExecLocal_RevokeOrder(payload *ety.RevokeOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case ety.TyRevokeOrderLog:
				receipt := &ety.ReceiptExchange{}
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

func (e *exchange) updateIndex(receipt *ety.ReceiptExchange) (kvs []*types.KeyValue) {
	completedTable := NewCompletedOrderTable(e.GetLocalDB())
	marketTable := NewMarketDepthTable(e.GetLocalDB())
	userTable := NewUserOrderTable(e.GetLocalDB())
	orderTable := NewMarketOrderTable(e.GetLocalDB())
	switch receipt.Order.Status {
	case ety.Ordered:
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		price := receipt.GetOrder().GetLimitOrder().GetPrice()
		order := receipt.GetOrder()
		index := receipt.GetIndex()
		var markDepth ety.MarketDepth
		depth, err := queryMarketDepth(e.GetLocalDB(), left, right, op, price)
		if err == types.ErrNotFound {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = receipt.Order.Balance
		} else {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = depth.Amount + receipt.Order.Balance
		}
		//marketDepth
		err = marketTable.Replace(&markDepth)
		if err != nil {
			elog.Error("updateIndex", "marketTable.Replace", err.Error())
			return nil
		}
		err = orderTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "orderTable.Replace", err.Error())
			return nil
		}
		err = userTable.Replace(order)
		if err != nil {
			return nil
		}
		if len(receipt.MatchOrders) > 0 {
			//撮合交易更新
			cache := make(map[float64]int64)
			for i, matchOrder := range receipt.MatchOrders {
				if matchOrder.Status == ety.Completed {
					// 删除原有状态orderID
					err = orderTable.DelRow(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "orderTable.DelRow", err.Error())
						return nil
					}
					//删除原有状态orderID
					matchOrder.Status = ety.Ordered
					userTable.DelRow(matchOrder)
					//更新状态为已完成,索引index,改为当前的index
					matchOrder.Status = ety.Completed
					matchOrder.Index = index + int64(i+1)
					userTable.Replace(matchOrder)
					//calcCompletedOrderKey
					completedTable.Replace(matchOrder)
				}
				if matchOrder.Status == ety.Ordered {
					//更新数据
					err = orderTable.Replace(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "orderTable.Replace", err.Error())
						return nil
					}
				}
				executed := cache[matchOrder.GetLimitOrder().Price]
				executed = executed + matchOrder.Executed
				cache[matchOrder.GetLimitOrder().Price] = executed
			}

			//更改匹配市场深度
			for pr, executed := range cache {
				var matchDepth ety.MarketDepth
				depth, err := queryMarketDepth(e.GetLocalDB(), left, right, OpSwap(op), pr)
				if err == types.ErrNotFound {
					continue
				} else {
					matchDepth.Price = pr
					matchDepth.LeftAsset = left
					matchDepth.RightAsset = right
					matchDepth.Op = OpSwap(op)
					matchDepth.Amount = depth.Amount - executed
				}
				//marketDepth
				err = marketTable.Replace(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.Replace", err.Error())
					return nil
				}
				if matchDepth.Amount <= 0 {
					//删除
					err = marketTable.DelRow(&matchDepth)
					if err != nil {
						elog.Error("updateIndex", "marketTable.DelRow", err.Error())
						return nil
					}
				}
			}
		}
	case ety.Completed:
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		index := receipt.GetIndex()
		err := userTable.Replace(receipt.GetOrder())
		if err != nil {
			elog.Error("updateIndex", "userTable.Replace", err.Error())
			return nil
		}
		err = completedTable.Replace(receipt.Order)
		if err != nil {
			elog.Error("updateIndex", "completedTable.Replace", err.Error())
			return nil
		}
		cache := make(map[float64]int64)
		if len(receipt.MatchOrders) > 0 {
			//撮合交易更新
			for i, matchOrder := range receipt.MatchOrders {
				if matchOrder.Status == ety.Completed {
					// 删除原有状态orderID
					err = orderTable.DelRow(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "orderTable.DelRow", err.Error())
						return nil
					}
					//删除原有状态orderID
					matchOrder.Status = ety.Ordered
					err = userTable.DelRow(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "userTable.DelRow", err.Error())
						return nil
					}
					//更新状态为已完成,更新索引
					matchOrder.Status = ety.Completed
					matchOrder.Index = index + int64(i+1)
					err = userTable.Replace(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "userTable.Replace", err.Error())
						return nil
					}
					//calcCompletedOrderKey
					err = completedTable.Replace(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "completedTable.Replace", err.Error())
						return nil
					}
				}

				if matchOrder.Status == ety.Ordered {
					//更新数据
					err = orderTable.Replace(matchOrder)
					if err != nil {
						elog.Error("updateIndex", "orderTable.Replace", err.Error())
						return nil
					}
				}
				executed := cache[matchOrder.GetLimitOrder().Price]
				executed = executed + matchOrder.Executed
				cache[matchOrder.GetLimitOrder().Price] = executed
			}
			//更改match市场深度
			for pr, executed := range cache {
				var matchDepth ety.MarketDepth
				depth, err := queryMarketDepth(e.GetLocalDB(), left, right, OpSwap(op), pr)
				if err == types.ErrNotFound {
					continue
				} else {
					matchDepth.Price = pr
					matchDepth.LeftAsset = left
					matchDepth.RightAsset = right
					matchDepth.Op = OpSwap(op)
					matchDepth.Amount = depth.Amount - executed
				}
				//marketDepth
				err = marketTable.Replace(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.Replace", err.Error())
					return nil
				}
				if matchDepth.Amount <= 0 {
					//删除
					err = marketTable.DelRow(&matchDepth)
					if err != nil {
						elog.Error("updateIndex", "marketTable.DelRow", err.Error())
						return nil
					}
				}
			}
		}
	case ety.Revoked:
		//只有状态时ordered状态的订单才能被撤回
		left := receipt.GetOrder().GetLimitOrder().GetLeftAsset()
		right := receipt.GetOrder().GetLimitOrder().GetRightAsset()
		op := receipt.GetOrder().GetLimitOrder().GetOp()
		price := receipt.GetOrder().GetLimitOrder().GetPrice()
		order := receipt.GetOrder()
		index := receipt.GetIndex()
		var marketDepth ety.MarketDepth
		depth, err := queryMarketDepth(e.GetLocalDB(), left, right, op, price)
		if err == nil {
			//marketDepth
			marketDepth.Amount = depth.Amount - receipt.GetOrder().Balance
			err = marketTable.Replace(&marketDepth)
			if err != nil {
				elog.Error("updateIndex", "marketTable.Replace", err.Error())
				return nil
			}
		}
		if marketDepth.Amount == 0 {
			//删除
			err = marketTable.DelRow(&marketDepth)
			if err != nil {
				elog.Error("updateIndex", "marketTable.DelRow", err.Error())
				return nil
			}
		}
		// 删除原有状态orderID
		err = orderTable.DelRow(order)
		if err != nil {
			elog.Error("updateIndex", "orderTable.DelRow", err.Error())
			return nil
		}
		//删除原有状态orderID
		order.Status = ety.Ordered
		err = userTable.DelRow(order)
		if err != nil {
			elog.Error("updateIndex", "userTable.DelRow", err.Error())
			return nil
		}
		order.Status = ety.Revoked
		order.Index = index
		//添加撤销的订单
		err = userTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "userTable.Replace", err.Error())
			return nil
		}
	}

	kv, err := marketTable.Save()
	if err != nil {
		elog.Error("updateIndex", "marketTable.Save", err.Error())
		return nil
	}
	kvs = append(kvs, kv...)
	kv, err = userTable.Save()
	if err != nil {
		elog.Error("updateIndex", "userTable.Save", err.Error())
		return nil
	}
	kvs = append(kvs, kv...)
	kv, err = orderTable.Save()
	if err != nil {
		elog.Error("updateIndex", "orderTable.Save", err.Error())
		return nil
	}
	kvs = append(kvs, kv...)
	kv, err = completedTable.Save()
	if err != nil {
		elog.Error("updateIndex", "completedTable.Save", err.Error())
		return nil
	}
	kvs = append(kvs, kv...)

	return
}

func OpSwap(op int32) int32 {
	if op == ety.OpBuy {
		return ety.OpSell
	}
	return ety.OpBuy
}

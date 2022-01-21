package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	ety "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (e *exchange) ExecLocal_LimitOrder(payload *ety.LimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocal(tx, receiptData, index)
}

func (e *exchange) ExecLocal_MarketOrder(payload *ety.MarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocal(tx, receiptData, index)
}

func (e *exchange) ExecLocal_RevokeOrder(payload *ety.RevokeOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocal(tx, receiptData, index)
}

func (e *exchange) ExecLocal_EntrustOrder(payload *ety.LimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocal(tx, receiptData, index)
}

func (e *exchange) ExecLocal_EntrustRevokeOrder(payload *ety.MarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocal(tx, receiptData, index)
}

func (e *exchange) interExecLocal(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	historyTable := NewHistoryOrderTable(e.GetLocalDB())
	marketTable := NewMarketDepthTable(e.GetLocalDB())
	orderTable := NewMarketOrderTable(e.GetLocalDB())
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case ety.TyMarketOrderLog, ety.TyRevokeOrderLog, ety.TyLimitOrderLog:
				receipt := &ety.ReceiptExchange{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				e.updateIndex(marketTable, orderTable, historyTable, receipt)
				//dbSet.KV = append(dbSet.KV, kv...)
			}
		}
	}
	//刷新KV
	var kvs []*types.KeyValue
	kv, err := marketTable.Save()
	if err != nil {
		elog.Error("updateIndex", "marketTable.Save", err.Error())
		return nil, nil
	}
	kvs = append(kvs, kv...)

	kv, err = orderTable.Save()
	if err != nil {
		elog.Error("updateIndex", "orderTable.Save", err.Error())
		return nil, nil
	}
	kvs = append(kvs, kv...)

	kv, err = historyTable.Save()
	if err != nil {
		elog.Error("updateIndex", "historyTable.Save", err.Error())
		return nil, nil
	}
	kvs = append(kvs, kv...)
	dbSet.KV = append(dbSet.KV, kvs...)
	dbSet = e.addAutoRollBack(tx, dbSet.KV)
	localDB := e.GetLocalDB()
	for _, kv1 := range dbSet.KV {
		err := localDB.Set(kv1.Key, kv1.Value)
		if err != nil {
			elog.Error("updateIndex", "localDB.Set", err.Error())
			return dbSet, err
		}
	}
	return dbSet, nil
}

//设置自动回滚
func (e *exchange) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {
	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

func (e *exchange) updateIndex(marketTable, orderTable, historyTable *table.Table, receipt *ety.ReceiptExchange) (kvs []*types.KeyValue) {
	switch receipt.Order.Status {
	case ety.Ordered:
		err := e.updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
		err = e.updateMatchOrders(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetMatchOrders(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	case ety.Completed:
		err := e.updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
		err = e.updateMatchOrders(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetMatchOrders(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	case ety.Revoked:
		err := e.updateOrder(marketTable, orderTable, historyTable, receipt.GetOrder(), receipt.GetIndex())
		if err != nil {
			return nil
		}
	}

	return
}

func (e *exchange) updateOrder(marketTable, orderTable, historyTable *table.Table, order *ety.Order, index int64) error {
	left := order.GetLimitOrder().GetLeftAsset()
	right := order.GetLimitOrder().GetRightAsset()
	op := order.GetLimitOrder().GetOp()
	price := order.GetLimitOrder().GetPrice()
	switch order.Status {
	case ety.Ordered:
		var markDepth ety.MarketDepth
		depth, err := queryMarketDepth(marketTable, left, right, op, price)
		if err == types.ErrNotFound {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = order.Balance
		} else {
			markDepth.Price = price
			markDepth.LeftAsset = left
			markDepth.RightAsset = right
			markDepth.Op = op
			markDepth.Amount = depth.Amount + order.Balance
		}
		//marketDepth
		err = marketTable.Replace(&markDepth)
		if err != nil {
			elog.Error("updateIndex", "marketTable.Replace", err.Error())
			return err
		}
		err = orderTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "orderTable.Replace", err.Error())
			return err
		}

	case ety.Completed:
		err := historyTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "historyTable.Replace", err.Error())
			return err
		}
	case ety.Revoked:
		//只有状态时ordered状态的订单才能被撤回
		var marketDepth ety.MarketDepth
		depth, err := queryMarketDepth(marketTable, left, right, op, price)
		if err != nil {
			elog.Error("updateIndex", "ety.Revoked queryMarketDepth", err.Error())
			return err
		}
		//marketDepth
		marketDepth.Price = price
		marketDepth.LeftAsset = left
		marketDepth.RightAsset = right
		marketDepth.Op = op
		marketDepth.Amount = depth.Amount - order.Balance

		if marketDepth.Amount > 0 {
			err = marketTable.Replace(&marketDepth)
			if err != nil {
				elog.Error("updateIndex", "marketTable.Replace", err.Error())
				return err
			}
		}
		if marketDepth.Amount <= 0 {
			//删除
			err = marketTable.DelRow(&marketDepth)
			if err != nil {
				elog.Error("updateIndex", "marketTable.DelRow", err.Error())
				return err
			}
		}
		//删除原有状态orderID
		primaryKey := []byte(fmt.Sprintf("%022d", order.OrderID))
		err = orderTable.Del(primaryKey)
		if err != nil {
			elog.Error("updateIndex", "orderTable.Del", err.Error())
			return err
		}
		order.Status = ety.Revoked
		order.Index = index
		//添加撤销的订单
		err = historyTable.Replace(order)
		if err != nil {
			elog.Error("updateIndex", "historyTable.Replace", err.Error())
			return err
		}
	}
	return nil
}
func (e *exchange) updateMatchOrders(marketTable, orderTable, historyTable *table.Table, order *ety.Order, matchOrders []*ety.Order, index int64) error {
	left := order.GetLimitOrder().GetLeftAsset()
	right := order.GetLimitOrder().GetRightAsset()
	op := order.GetLimitOrder().GetOp()
	if len(matchOrders) > 0 {
		//撮合交易更新
		cache := make(map[int64]int64)
		for i, matchOrder := range matchOrders {
			if matchOrder.Balance == 0 && matchOrder.Executed == 0 {
				var matchDepth ety.MarketDepth
				matchDepth.Price = matchOrder.AVGPrice
				matchDepth.LeftAsset = left
				matchDepth.RightAsset = right
				matchDepth.Op = OpSwap(op)
				matchDepth.Amount = 0
				err := marketTable.DelRow(&matchDepth)
				if err != nil && err != types.ErrNotFound {
					elog.Error("updateIndex", "marketTable.DelRow", err.Error())
					return err
				}
				continue
			}
			if matchOrder.Status == ety.Completed {
				// 删除原有状态orderID
				err := orderTable.DelRow(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "orderTable.DelRow", err.Error())
					return err
				}
				//索引index,改为当前的index
				matchOrder.Index = index + int64(i+1)
				err = historyTable.Replace(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "historyTable.Replace", err.Error())
					return err
				}
			} else if matchOrder.Status == ety.Ordered {
				//更新数据
				err := orderTable.Replace(matchOrder)
				if err != nil {
					elog.Error("updateIndex", "orderTable.Replace", err.Error())
					return err
				}
			}
			executed := cache[matchOrder.GetLimitOrder().Price]
			executed = executed + matchOrder.Executed
			cache[matchOrder.GetLimitOrder().Price] = executed
		}

		//更改匹配市场深度
		for pr, executed := range cache {
			var matchDepth ety.MarketDepth
			depth, err := queryMarketDepth(marketTable, left, right, OpSwap(op), pr)
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
			if matchDepth.Amount > 0 {
				err = marketTable.Replace(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.Replace", err.Error())
					return err
				}
			}
			if matchDepth.Amount <= 0 {
				//删除
				err = marketTable.DelRow(&matchDepth)
				if err != nil {
					elog.Error("updateIndex", "marketTable.DelRow", err.Error())
					return err
				}
			}
		}
	}
	return nil
}

//OpSwap ...
func OpSwap(op int32) int32 {
	if op == ety.OpBuy {
		return ety.OpSell
	}
	return ety.OpBuy
}

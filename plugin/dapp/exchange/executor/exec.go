package executor

import (
	"fmt"
	"github.com/33cn/chain33/types"
	exchangetypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (e *exchange) Exec_LimitOrder(payload *exchangetypes.LimitOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.LimitOrder(payload)
}

func (e *exchange) Exec_MarketOrder(payload *exchangetypes.MarketOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	//TODO marketOrder
	return nil, fmt.Errorf("%s", "not support MarketOrder..")
}

func (e *exchange) Exec_RevokeOrder(payload *exchangetypes.RevokeOrder, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(e, tx, index)
	return action.RevokeOrder(payload)
}

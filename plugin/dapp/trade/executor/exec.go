// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

func (t *trade) Exec_SellLimit(sell *pty.TradeForSell, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeSell(sell)
}

func (t *trade) Exec_BuyMarket(buy *pty.TradeForBuy, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeBuy(buy)
}

func (t *trade) Exec_RevokeSell(revoke *pty.TradeForRevokeSell, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeRevokeSell(revoke)
}

func (t *trade) Exec_BuyLimit(buy *pty.TradeForBuyLimit, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeBuyLimit(buy)
}

func (t *trade) Exec_SellMarket(sell *pty.TradeForSellMarket, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeSellMarket(sell)
}

func (t *trade) Exec_RevokeBuy(revoke *pty.TradeForRevokeBuy, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newTradeAction(t, tx)
	return action.tradeRevokeBuyLimit(revoke)
}

package spot

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

// LeftToken: seller -> buyer
// RightToken: buyer -> seller
// RightToken: buyer, seller -> fee-bank

type SpotTrader struct {
	cfg *types.Chain33Config

	fee     *SpotFee
	accFeeX AssetAccount

	AccID uint64
	from  string
	accX  *AssetAccounts

	order   *Order
	matches *et.ReceiptSpotMatch
}

func (s *SpotTrader) GetOrder() *Order {
	return s.order
}

func (s *SpotTrader) GetAccout() *AssetAccounts {
	return s.accX
}

func (s *SpotTrader) CheckTokenAmountForLimitOrder(tid uint64, total int64) error {
	err := s.accX.sellAcc.CheckBalance(total)
	if err != nil {
		elog.Error("check balance", "total", total, "err", err)
		return et.ErrAssetBalance
	}
	return nil
}

func (s *SpotTrader) FrozenForLimitOrder(orderx *Order) (*types.Receipt, error) {
	precision := s.cfg.GetCoinPrecision()
	asset, amount := orderx.calcFrozenToken(precision)

	receipt, err := s.accX.sellAcc.Frozen(int64(amount))
	if err != nil {
		elog.Error("FrozenForLimitOrder", "asset.Ty", asset.Ty, "err", err, "need", amount)
		return nil, et.ErrAssetBalance
	}
	return receipt, nil
}

func (s *SpotTrader) Trade(maker *spotMaker) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	balance := s.calcTradeBalance(maker.order.order)
	matchDetail := s.calcTradeInfo(maker, balance)

	receipt3, kvs3, err := maker.orderTraded(matchDetail, s.order.order)
	if err != nil {
		elog.Error("maker.orderTraded", "err", err)
		return receipt3, kvs3, err
	}

	taker := spotTaker{SpotTrader: s}
	receipt2, kvs2, err := taker.orderTraded(matchDetail, maker.order.order)
	if err != nil {
		elog.Error("taker.orderTraded", "err", err)
		return receipt2, kvs2, err
	}

	var receipt []*types.ReceiptLog
	var kvs []*types.KeyValue
	if s.AccID == maker.AccID {
		receipt, kvs, err = s.selfSettlement(maker, matchDetail)
	} else {
		receipt, kvs, err = s.settlement(maker, matchDetail)
	}
	if err != nil {
		elog.Error("settlement", "err", err)
		return receipt, kvs, err
	}

	kvs = append(kvs, kvs2...)
	kvs = append(kvs, kvs3...)
	receipt = append(receipt, receipt2...)
	receipt = append(receipt, receipt3...)

	return receipt, kvs, nil
}

func (s *SpotTrader) calcTradeBalance(order *et.SpotOrder) int64 {
	if order.GetBalance() >= s.order.order.GetBalance() {
		return s.order.order.GetBalance()
	}
	return order.GetBalance()
}

func (s *SpotTrader) calcTradeInfo(maker *spotMaker, balance int64) *et.MatchInfo {
	var info et.MatchInfo
	info.Matched = balance
	info.LeftBalance = balance
	info.RightBalance = SafeMul(balance, maker.order.order.GetLimitOrder().Price, s.cfg.GetCoinPrecision())
	info.FeeTaker = SafeMul(info.RightBalance, int64(s.order.order.TakerRate), s.cfg.GetCoinPrecision())
	info.FeeMaker = SafeMul(info.RightBalance, int64(maker.order.order.Rate), s.cfg.GetCoinPrecision())
	return &info
}

// account 是一个对象代表一个人的一个资产
// dexAccount 是一个对象代表一个人的所有资产
// settlement
// LeftToken: seller -> buyer
// RightToken: buyer -> seller
// RightToken: buyer, seller -> fee-bank
func (s *SpotTrader) settlement(maker *spotMaker, tradeBalance *et.MatchInfo) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	//if s.acc.acc.Id == maker.acc.acc.Id {
	//	return s.selfSettlement(maker, tradeBalance)
	//}
	var err error
	var receipt, receipt2 *types.Receipt
	if s.order.GetOp() == et.OpSell {
		receipt, err = s.accX.sellAcc.Transfer(maker.accX.buyAcc, tradeBalance.LeftBalance)
		if err != nil {
			elog.Error("settlement", "sell.doTranfer1", err)
			return nil, nil, err
		}
		receipt2, err = maker.accX.sellAcc.TransferFrozen(s.accX.buyAcc, tradeBalance.RightBalance)
		if err != nil {
			elog.Error("settlement", "sell.doFrozenTranfer2", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
		receipt2, err = s.accX.sellAcc.Transfer(s.accFeeX, tradeBalance.FeeTaker)
		if err != nil {
			elog.Error("settlement", "sell-fee.doTranfer", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
		receipt2, err = maker.accX.buyAcc.TransferFrozen(s.accFeeX, tradeBalance.FeeMaker)
		if err != nil {
			elog.Error("settlement", "sell-fee.doFrozenTranfer3", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
	} else {
		receipt, err = s.accX.sellAcc.Transfer(maker.accX.buyAcc, tradeBalance.RightBalance)
		if err != nil {
			elog.Error("settlement", "buy.doTranfer1", err)
			return nil, nil, err
		}
		receipt2, err = maker.accX.sellAcc.TransferFrozen(s.accX.buyAcc, tradeBalance.LeftBalance)
		if err != nil {
			elog.Error("settlement", "buy.doFrozenTranfer2", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
		receipt2, err = s.accX.sellAcc.Transfer(s.accFeeX, tradeBalance.FeeTaker)
		if err != nil {
			elog.Error("settlement", "buy-fee.doTranfer1", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
		receipt2, err = maker.accX.buyAcc.Transfer(s.accFeeX, tradeBalance.FeeMaker)
		if err != nil {
			elog.Error("settlement", "buy-fee.doTranfer2", err)
			return nil, nil, err
		}
		receipt = et.MergeReceipt(receipt, receipt2)
	}

	re := et.ReceiptSpotTrade{
		Match: tradeBalance,
		//Prev: &et.TradeAccounts{
		//	Taker: copyAcc,
		//	Maker: copyAccMaker,
		//	Fee:   copyFeeAcc,
		//},
		//Current: &et.TradeAccounts{
		//	Taker: s.acc.acc,
		//	Maker: maker.acc.acc,
		//	Fee:   s.accFee.acc,
		//},
		MakerOrder: maker.order.order.GetLimitOrder().Order, // TODO
	}

	log1 := types.ReceiptLog{
		Ty:  et.TySpotTradeLog,
		Log: types.Encode(&re),
	}
	receipt.Logs = append(receipt.Logs, &log1)
	return receipt.Logs, receipt.KV, nil
}

// taker/maker the same user
func (s *SpotTrader) selfSettlement(maker *spotMaker, tradeBalance *et.MatchInfo) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	var err error
	var receipt, receipt2 *types.Receipt

	// taker 是buy,  maker 是sell, Left 是冻结的. takerFee + makerFee 是活动的
	// taker 是sell, maker 是 buy, Right 是冻结的. makerFee 是冻结的. takerFee是活动的
	if s.order.GetOp() == et.OpSell {
		rightAmount := tradeBalance.RightBalance
		rightAmount += tradeBalance.FeeMaker
		receipt, err = s.accX.buyAcc.UnFrozen(int64(rightAmount))
		if err != nil {
			return nil, nil, err
		}
		receipt2, err = s.accX.buyAcc.Transfer(s.accFeeX, int64(tradeBalance.FeeTaker+tradeBalance.FeeMaker))
		if err != nil {
			return nil, nil, err
		}
	} else {
		receipt, err = s.accX.sellAcc.UnFrozen(int64(tradeBalance.LeftBalance))
		if err != nil {
			return nil, nil, err
		}
		receipt2, err = s.accX.sellAcc.Transfer(s.accFeeX, int64(tradeBalance.FeeTaker+tradeBalance.FeeMaker))
		if err != nil {
			return nil, nil, err
		}
	}
	receipt = et.MergeReceipt(receipt, receipt2)

	re := et.ReceiptSpotTrade{
		Match: tradeBalance,
		//Prev: &et.TradeAccounts{
		//	Taker: copyAcc,
		//	Maker: copyAcc,
		//	Fee:   copyFeeAcc,
		//},
		//Current: &et.TradeAccounts{
		//	Taker: s.acc.acc,
		//	Maker: s.acc.acc,
		//	Fee:   s.accFee.acc,
		//},
		MakerOrder: maker.order.order.GetLimitOrder().Order,
	}

	log1 := types.ReceiptLog{
		Ty:  et.TySpotTradeLog,
		Log: types.Encode(&re),
	}
	receipt.Logs = append(receipt.Logs, &log1)
	return receipt.Logs, receipt.KV, nil
}

type spotTaker struct {
	*SpotTrader
}

type spotMaker struct {
	*SpotTrader
}

// s.order is taker, order is maker
func (s *spotTaker) orderTraded(matchDetail *et.MatchInfo, order *et.SpotOrder) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	matched := matchDetail.Matched

	// fee and AVGPrice
	s.order.order.DigestedFee += matchDetail.FeeTaker
	s.order.order.AVGPrice = caclAVGPrice(s.order.order, s.order.order.GetLimitOrder().Price, matched)

	// status
	if matched == s.order.order.GetBalance() {
		s.order.order.Status = et.Completed
	} else {
		s.order.order.Status = et.Ordered
	}

	// order matched
	s.order.order.Executed = matched
	s.order.order.Balance -= matched

	s.matches.Order = s.order.order
	s.matches.MatchOrders = append(s.matches.MatchOrders, order)
	// receipt-log, order-kvs 在匹配完成后一次性生成, 不需要生成多次
	// kvs := GetOrderKvSet(s.order)
	// logs += s.matches
	return []*types.ReceiptLog{}, []*types.KeyValue{}, nil
}

// 2 -> 1 update, 2 kv
func (m *spotMaker) orderTraded(matchDetail *et.MatchInfo, takerOrder *et.SpotOrder) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	matched := matchDetail.Matched

	// fee and AVGPrice
	m.order.order.DigestedFee += matchDetail.FeeMaker
	m.order.order.AVGPrice = caclAVGPrice(m.order.order, m.order.order.GetLimitOrder().Price, matched)

	m.order.order.UpdateTime = takerOrder.UpdateTime

	// status
	if matched == m.order.order.GetBalance() {
		m.order.order.Status = et.Completed
	} else {
		m.order.order.Status = et.Ordered
	}

	// order matched
	m.order.order.Executed = matched
	m.order.order.Balance -= matched
	kvs := m.order.repo.GetOrderKvSet(m.order.order)
	return []*types.ReceiptLog{}, kvs, nil
}

func (taker *SpotTrader) matchModel(matchorder *Order, statedb dbm.KV, s *Spot) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	matched := taker.calcTradeBalance(matchorder.order)
	elog.Info("try match", "activeId", taker.order.order.OrderID, "passiveId", matchorder.order.OrderID, "activeAddr", taker.order.order.Addr, "passiveAddr",
		matchorder.order.Addr, "amount", matched, "price", taker.order.order.GetLimitOrder().Price)

	accMatch, err := s.LoadTrader(matchorder.order.Addr, matchorder.GetZkOrder().GetAccountID(), taker.accX.sell, taker.accX.buy)
	if err != nil {
		return nil, nil, err
	}
	maker := spotMaker{
		SpotTrader: accMatch,
	}

	logs, kvs, err = taker.Trade(&maker)
	elog.Info("try match2", "activeId", taker.order.order.OrderID, "passiveId", matchorder.order.OrderID, "activeAddr", taker.order.order.Addr, "passiveAddr",
		matchorder.order.Addr, "amount", matched, "price", taker.order.GetPrice())
	return logs, kvs, err
}

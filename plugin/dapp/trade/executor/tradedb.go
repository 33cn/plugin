// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

type sellDB struct {
	pty.SellOrder
}

func newSellDB(sellOrder pty.SellOrder) (selldb *sellDB) {
	selldb = &sellDB{sellOrder}
	if pty.InvalidStartTime != selldb.Starttime {
		selldb.Status = pty.TradeOrderStatusNotStart
	}
	return
}

func (selldb *sellDB) save(db dbm.KV) []*types.KeyValue {
	set := selldb.getKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}

	return set
}

func (selldb *sellDB) getSellLogs(tradeType int32, txhash string, cfg *types.Chain33Config) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = tradeType
	base := &pty.ReceiptSellBase{
		TokenSymbol:       selldb.TokenSymbol,
		Owner:             selldb.Address,
		AmountPerBoardlot: types.FormatAmount2FixPrecisionDisplay(selldb.AmountPerBoardlot, cfg.GetTokenPrecision()),
		MinBoardlot:       selldb.MinBoardlot,
		PricePerBoardlot:  types.FormatAmount2FixPrecisionDisplay(selldb.PricePerBoardlot, cfg.GetCoinPrecision()),
		TotalBoardlot:     selldb.TotalBoardlot,
		SoldBoardlot:      selldb.SoldBoardlot,
		Starttime:         selldb.Starttime,
		Stoptime:          selldb.Stoptime,
		Crowdfund:         selldb.Crowdfund,
		SellID:            selldb.SellID,
		Status:            pty.SellOrderStatus[selldb.Status],
		BuyID:             "",
		TxHash:            txhash,
		Height:            selldb.Height,
		AssetExec:         selldb.AssetExec,
		PriceExec:         selldb.GetPriceExec(),
		PriceSymbol:       selldb.GetPriceSymbol(),
	}
	if pty.TyLogTradeSellLimit == tradeType {
		receiptTrade := &pty.ReceiptTradeSellLimit{Base: base}
		log.Log = types.Encode(receiptTrade)

	} else if pty.TyLogTradeSellRevoke == tradeType {
		receiptTrade := &pty.ReceiptTradeSellRevoke{Base: base}
		log.Log = types.Encode(receiptTrade)
	}

	return log
}

func (selldb *sellDB) getBuyLogs(buyerAddr string, boardlotcnt int64, txhash string, cfg *types.Chain33Config) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogTradeBuyMarket
	base := &pty.ReceiptBuyBase{
		TokenSymbol:       selldb.TokenSymbol,
		Owner:             buyerAddr,
		AmountPerBoardlot: types.FormatAmount2FixPrecisionDisplay(selldb.AmountPerBoardlot, cfg.GetTokenPrecision()),
		MinBoardlot:       selldb.MinBoardlot,
		PricePerBoardlot:  types.FormatAmount2FixPrecisionDisplay(selldb.PricePerBoardlot, cfg.GetCoinPrecision()),
		TotalBoardlot:     boardlotcnt,
		BoughtBoardlot:    boardlotcnt,
		BuyID:             "",
		Status:            pty.SellOrderStatus[pty.TradeOrderStatusBoughtOut],
		SellID:            selldb.SellID,
		TxHash:            txhash,
		Height:            selldb.Height,
		AssetExec:         selldb.AssetExec,
		PriceExec:         selldb.PriceExec,
		PriceSymbol:       selldb.PriceSymbol,
	}

	receipt := &pty.ReceiptTradeBuyMarket{Base: base}
	log.Log = types.Encode(receipt)
	return log
}

func getSellOrderFromID(sellID []byte, db dbm.KV) (*pty.SellOrder, error) {
	value, err := db.Get(sellID)
	if err != nil {
		tradelog.Error("getSellOrderFromID", "Failed to get value from db with sellid", string(sellID))
		return nil, err
	}

	var sellOrder pty.SellOrder
	if err = types.Decode(value, &sellOrder); err != nil {
		tradelog.Error("getSellOrderFromID", "Failed to decode sell order", string(sellID))
		return nil, err
	}
	return &sellOrder, nil
}

func (selldb *sellDB) getKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&selldb.SellOrder)
	key := []byte(selldb.SellID)
	kvset = append(kvset, &types.KeyValue{Key: key, Value: value})
	return kvset
}

type buyDB struct {
	pty.BuyLimitOrder
}

func newBuyDB(sellOrder pty.BuyLimitOrder) (buydb *buyDB) {
	buydb = &buyDB{sellOrder}
	return
}

func (buydb *buyDB) save(db dbm.KV) []*types.KeyValue {
	set := buydb.getKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}

	return set
}

func (buydb *buyDB) getKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&buydb.BuyLimitOrder)
	key := []byte(buydb.BuyID)
	kvset = append(kvset, &types.KeyValue{Key: key, Value: value})
	return kvset
}

func (buydb *buyDB) getBuyLogs(tradeType int32, txhash string, cfg *types.Chain33Config) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = tradeType
	base := &pty.ReceiptBuyBase{
		TokenSymbol:       buydb.TokenSymbol,
		Owner:             buydb.Address,
		AmountPerBoardlot: types.FormatAmount2FixPrecisionDisplay(buydb.AmountPerBoardlot, cfg.GetTokenPrecision()),
		MinBoardlot:       buydb.MinBoardlot,
		PricePerBoardlot:  types.FormatAmount2FixPrecisionDisplay(buydb.PricePerBoardlot, cfg.GetCoinPrecision()),
		TotalBoardlot:     buydb.TotalBoardlot,
		BoughtBoardlot:    buydb.BoughtBoardlot,
		BuyID:             buydb.BuyID,
		Status:            pty.SellOrderStatus[buydb.Status],
		SellID:            "",
		TxHash:            txhash,
		Height:            buydb.Height,
		AssetExec:         buydb.AssetExec,
		PriceExec:         buydb.PriceExec,
		PriceSymbol:       buydb.PriceSymbol,
	}
	if pty.TyLogTradeBuyLimit == tradeType {
		receiptTrade := &pty.ReceiptTradeBuyLimit{Base: base}
		log.Log = types.Encode(receiptTrade)

	} else if pty.TyLogTradeBuyRevoke == tradeType {
		receiptTrade := &pty.ReceiptTradeBuyRevoke{Base: base}
		log.Log = types.Encode(receiptTrade)
	}

	return log
}

func getBuyOrderFromID(buyID []byte, db dbm.KV) (*pty.BuyLimitOrder, error) {
	value, err := db.Get(buyID)
	if err != nil {
		tradelog.Error("getBuyOrderFromID", "Failed to get value from db with buyID", string(buyID))
		return nil, err
	}

	var buy pty.BuyLimitOrder
	if err = types.Decode(value, &buy); err != nil {
		tradelog.Error("getBuyOrderFromID", "Failed to decode buy order", string(buyID))
		return nil, err
	}
	return &buy, nil
}

func (buydb *buyDB) getSellLogs(sellerAddr string, sellID string, boardlotCnt int64, txhash string, cfg *types.Chain33Config) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogTradeSellMarket
	base := &pty.ReceiptSellBase{
		TokenSymbol:       buydb.TokenSymbol,
		Owner:             sellerAddr,
		AmountPerBoardlot: types.FormatAmount2FixPrecisionDisplay(buydb.AmountPerBoardlot, cfg.GetTokenPrecision()),
		MinBoardlot:       buydb.MinBoardlot,
		PricePerBoardlot:  types.FormatAmount2FixPrecisionDisplay(buydb.PricePerBoardlot, cfg.GetCoinPrecision()),
		TotalBoardlot:     boardlotCnt,
		SoldBoardlot:      boardlotCnt,
		Starttime:         0,
		Stoptime:          0,
		Crowdfund:         false,
		SellID:            "",
		Status:            pty.SellOrderStatus[pty.TradeOrderStatusSoldOut],
		BuyID:             buydb.BuyID,
		TxHash:            txhash,
		Height:            buydb.Height,
		AssetExec:         buydb.AssetExec,
		PriceExec:         buydb.PriceExec,
		PriceSymbol:       buydb.PriceSymbol,
	}
	receiptSellMarket := &pty.ReceiptSellMarket{Base: base}
	log.Log = types.Encode(receiptSellMarket)

	return log
}

type tradeAction struct {
	db        dbm.KV
	txhash    string
	fromaddr  string
	blocktime int64
	height    int64
	execaddr  string
	api       client.QueueProtocolAPI
}

func newTradeAction(t *trade, tx *types.Transaction) *tradeAction {
	hash := hex.EncodeToString(tx.Hash())
	fromaddr := tx.From()
	return &tradeAction{t.GetStateDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), dapp.ExecAddress(string(tx.Execer)), t.GetAPI()}
}

func (action *tradeAction) tradeSell(sell *pty.TradeForSell) (*types.Receipt, error) {
	if sell.TotalBoardlot < 0 || sell.PricePerBoardlot < 0 || sell.MinBoardlot < 0 || sell.AmountPerBoardlot < 0 {
		return nil, types.ErrInvalidParam
	}
	cfg := action.api.GetConfig()

	if !checkAsset(cfg, action.height, sell.AssetExec, sell.TokenSymbol) {
		return nil, types.ErrInvalidParam
	}
	if !checkPrice(cfg, action.height, sell.PriceExec, sell.PriceExec) {
		return nil, types.ErrInvalidParam
	}
	if !notSameAsset(cfg, action.height, sell.AssetExec, sell.TokenSymbol, sell.PriceExec, sell.PriceExec) {
		return nil, pty.ErrAssetAndPriceSame
	}

	accDB, err := createAccountDB(cfg, action.height, action.db, sell.AssetExec, sell.TokenSymbol)
	if err != nil {
		return nil, err
	}
	//确认发起此次出售或者众筹的余额是否足够
	totalAmount := sell.GetTotalBoardlot() * sell.GetAmountPerBoardlot()
	receipt, err := accDB.ExecFrozen(action.fromaddr, action.execaddr, totalAmount)
	if err != nil {
		tradelog.Error("trade sell ", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", totalAmount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	sellOrder := pty.SellOrder{
		TokenSymbol:       sell.GetTokenSymbol(),
		Address:           action.fromaddr,
		AmountPerBoardlot: sell.GetAmountPerBoardlot(),
		MinBoardlot:       sell.GetMinBoardlot(),
		PricePerBoardlot:  sell.GetPricePerBoardlot(),
		TotalBoardlot:     sell.GetTotalBoardlot(),
		SoldBoardlot:      0,
		Starttime:         sell.GetStarttime(),
		Stoptime:          sell.GetStoptime(),
		Crowdfund:         sell.GetCrowdfund(),
		SellID:            calcTokenSellID(action.txhash),
		Status:            pty.TradeOrderStatusOnSale,
		Height:            action.height,
		AssetExec:         sell.AssetExec,
		PriceExec:         sell.GetPriceExec(),
		PriceSymbol:       sell.GetPriceSymbol(),
	}

	tokendb := newSellDB(sellOrder)
	sellOrderKV := tokendb.save(action.db)
	logs = append(logs, receipt.Logs...)
	logs = append(logs, tokendb.getSellLogs(pty.TyLogTradeSellLimit, action.txhash, cfg))
	kv = append(kv, receipt.KV...)
	kv = append(kv, sellOrderKV...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *tradeAction) tradeBuy(buyOrder *pty.TradeForBuy) (*types.Receipt, error) {
	cfg := action.api.GetConfig()
	if cfg.IsDappFork(action.height, pty.TradeX, pty.ForkTradeIDX) {
		buyOrder.SellID = calcTokenSellID(buyOrder.SellID)
	}
	if buyOrder.BoardlotCnt < 0 || !strings.HasPrefix(buyOrder.SellID, sellIDPrefix) {
		return nil, types.ErrInvalidParam
	}

	sellidByte := []byte(buyOrder.SellID)
	sellOrder, err := getSellOrderFromID(sellidByte, action.db)
	if err != nil {
		return nil, pty.ErrTSellOrderNotExist
	}

	if sellOrder.Status == pty.TradeOrderStatusNotStart && sellOrder.Starttime > action.blocktime {
		return nil, pty.ErrTSellOrderNotStart
	} else if sellOrder.Status == pty.TradeOrderStatusSoldOut {
		return nil, pty.ErrTSellOrderSoldout
	} else if sellOrder.Status == pty.TradeOrderStatusOnSale && sellOrder.TotalBoardlot-sellOrder.SoldBoardlot < buyOrder.BoardlotCnt {
		return nil, pty.ErrTSellOrderNotEnough
	} else if sellOrder.Status == pty.TradeOrderStatusRevoked {
		return nil, pty.ErrTSellOrderRevoked
	} else if sellOrder.Status == pty.TradeOrderStatusExpired {
		return nil, pty.ErrTSellOrderExpired
	} else if sellOrder.Status == pty.TradeOrderStatusOnSale && buyOrder.BoardlotCnt < sellOrder.MinBoardlot {
		return nil, pty.ErrTCntLessThanMinBoardlot
	}

	priceAcc, err := createPriceDB(cfg, action.height, action.db, sellOrder.PriceExec, sellOrder.PriceSymbol)
	if err != nil {
		tradelog.Error("createPriceDB", "addrFrom", action.fromaddr, "height", action.height,
			"price", sellOrder.PriceExec+"-"+sellOrder.PriceSymbol, "err", err)
		return nil, err
	}
	//首先购买费用的划转
	receiptFromAcc, err := priceAcc.ExecTransfer(action.fromaddr, sellOrder.Address, action.execaddr, buyOrder.BoardlotCnt*sellOrder.PricePerBoardlot)
	if err != nil {
		tradelog.Error("account.Transfer ", "addrFrom", action.fromaddr, "addrTo", sellOrder.Address,
			"amount", buyOrder.BoardlotCnt*sellOrder.PricePerBoardlot)
		return nil, err
	}
	//然后实现购买token的转移,因为这部分token在之前的卖单生成时已经进行冻结
	accDB, err := createAccountDB(cfg, action.height, action.db, sellOrder.AssetExec, sellOrder.TokenSymbol)
	if err != nil {
		tradelog.Error("createAccountDB", "addrFrom", action.fromaddr, "height", action.height,
			"price", sellOrder.AssetExec+"-"+sellOrder.TokenSymbol, "err", err)
		return nil, err
	}
	receiptFromExecAcc, err := accDB.ExecTransferFrozen(sellOrder.Address, action.fromaddr, action.execaddr, buyOrder.BoardlotCnt*sellOrder.AmountPerBoardlot)
	if err != nil {
		tradelog.Error("account.ExecTransfer token ", "error info", err, "addrFrom", sellOrder.Address,
			"addrTo", action.fromaddr, "execaddr", action.execaddr,
			"amount", buyOrder.BoardlotCnt*sellOrder.AmountPerBoardlot)
		//因为未能成功将对应的token进行转账，所以需要将购买方的账户资金进行回退
		priceAcc.ExecTransfer(sellOrder.Address, action.fromaddr, action.execaddr, buyOrder.BoardlotCnt*sellOrder.PricePerBoardlot)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	tradelog.Debug("tradeBuy", "Soldboardlot before this buy", sellOrder.SoldBoardlot)
	sellOrder.SoldBoardlot += buyOrder.BoardlotCnt
	tradelog.Debug("tradeBuy", "Soldboardlot after this buy", sellOrder.SoldBoardlot)
	if sellOrder.SoldBoardlot == sellOrder.TotalBoardlot {
		sellOrder.Status = pty.TradeOrderStatusSoldOut
	}
	sellTokendb := newSellDB(*sellOrder)
	sellOrderKV := sellTokendb.save(action.db)

	logs = append(logs, receiptFromAcc.Logs...)
	logs = append(logs, receiptFromExecAcc.Logs...)
	logs = append(logs, sellTokendb.getSellLogs(pty.TyLogTradeSellLimit, action.txhash, cfg))
	logs = append(logs, sellTokendb.getBuyLogs(action.fromaddr, buyOrder.BoardlotCnt, action.txhash, cfg))
	kv = append(kv, receiptFromAcc.KV...)
	kv = append(kv, receiptFromExecAcc.KV...)
	kv = append(kv, sellOrderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *tradeAction) tradeRevokeSell(revoke *pty.TradeForRevokeSell) (*types.Receipt, error) {
	cfg := action.api.GetConfig()
	if cfg.IsDappFork(action.height, pty.TradeX, pty.ForkTradeIDX) {
		revoke.SellID = calcTokenSellID(revoke.SellID)
	}
	if !strings.HasPrefix(revoke.SellID, sellIDPrefix) {
		return nil, types.ErrInvalidParam
	}
	sellidByte := []byte(revoke.SellID)
	sellOrder, err := getSellOrderFromID(sellidByte, action.db)
	if err != nil {
		return nil, pty.ErrTSellOrderNotExist
	}

	if sellOrder.Status == pty.TradeOrderStatusSoldOut {
		return nil, pty.ErrTSellOrderSoldout
	} else if sellOrder.Status == pty.TradeOrderStatusRevoked {
		return nil, pty.ErrTSellOrderRevoked
	} else if sellOrder.Status == pty.TradeOrderStatusExpired {
		return nil, pty.ErrTSellOrderExpired
	}

	if action.fromaddr != sellOrder.Address {
		return nil, pty.ErrTSellOrderRevoke
	}
	//然后实现购买token的转移,因为这部分token在之前的卖单生成时已经进行冻结
	accDB, err := createAccountDB(cfg, action.height, action.db, sellOrder.AssetExec, sellOrder.TokenSymbol)
	if err != nil {
		return nil, err
	}
	tradeRest := (sellOrder.TotalBoardlot - sellOrder.SoldBoardlot) * sellOrder.AmountPerBoardlot
	receiptFromExecAcc, err := accDB.ExecActive(sellOrder.Address, action.execaddr, tradeRest)
	if err != nil {
		tradelog.Error("account.ExecActive token ", "addrFrom", sellOrder.Address, "execaddr", action.execaddr, "amount", tradeRest)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	sellOrder.Status = pty.TradeOrderStatusRevoked
	tokendb := newSellDB(*sellOrder)
	sellOrderKV := tokendb.save(action.db)

	logs = append(logs, receiptFromExecAcc.Logs...)
	logs = append(logs, tokendb.getSellLogs(pty.TyLogTradeSellRevoke, action.txhash, cfg))
	kv = append(kv, receiptFromExecAcc.KV...)
	kv = append(kv, sellOrderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//不同合约之间查询的需求后面要考虑，现在先重复处理一下，原则上不能直接引用其他合约的代码
//后面可能会有一套查询规则 和 写规则, 合约对其他合约只读
func calcTokenKey(token string) (key []byte) {
	tokenCreated := "mavl-token-"
	return []byte(fmt.Sprintf(tokenCreated+"%s", token))
}

func checkTokenExist(token string, db dbm.KV) bool {
	_, err := db.Get(calcTokenKey(token))
	return err == nil
}

func (action *tradeAction) tradeBuyLimit(buy *pty.TradeForBuyLimit) (*types.Receipt, error) {
	// ErrTokenNotExist error token symbol not exist
	errTokenNotExist := errors.New("ErrTokenSymbolNotExist")

	if buy.TotalBoardlot < 0 || buy.PricePerBoardlot < 0 || buy.MinBoardlot < 0 || buy.AmountPerBoardlot < 0 {
		return nil, types.ErrInvalidParam
	}
	// 这个检查会比较鸡肋, 按目前的想法的能支持更多的资产， 各种资产检查不一样
	// 可以先让订单成功, 如果不合适, 自己撤单也行
	// 或后续跨合约注册一个检测的函数
	if buy.AssetExec == "" || buy.AssetExec == defaultAssetExec {
		// check token exist
		if !checkTokenExist(buy.TokenSymbol, action.db) {
			return nil, errTokenNotExist
		}
	}

	cfg := action.api.GetConfig()

	if !checkAsset(cfg, action.height, buy.AssetExec, buy.TokenSymbol) {
		return nil, types.ErrInvalidParam
	}
	if !checkPrice(cfg, action.height, buy.PriceExec, buy.PriceExec) {
		return nil, types.ErrInvalidParam
	}
	if !notSameAsset(cfg, action.height, buy.AssetExec, buy.TokenSymbol, buy.PriceExec, buy.PriceExec) {
		return nil, pty.ErrAssetAndPriceSame
	}

	priceAcc, err := createPriceDB(cfg, action.height, action.db, buy.PriceExec, buy.PriceSymbol)
	if err != nil {
		return nil, err
	}
	// check enough bty
	amount := buy.PricePerBoardlot * buy.TotalBoardlot
	receipt, err := priceAcc.ExecFrozen(action.fromaddr, action.execaddr, amount)
	if err != nil {
		tradelog.Error("trade tradeBuyLimit ", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", amount)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	buyOrder := pty.BuyLimitOrder{
		TokenSymbol:       buy.GetTokenSymbol(),
		Address:           action.fromaddr,
		AmountPerBoardlot: buy.GetAmountPerBoardlot(),
		MinBoardlot:       buy.GetMinBoardlot(),
		PricePerBoardlot:  buy.GetPricePerBoardlot(),
		TotalBoardlot:     buy.GetTotalBoardlot(),
		BoughtBoardlot:    0,
		BuyID:             calcTokenBuyID(action.txhash),
		Status:            pty.TradeOrderStatusOnBuy,
		Height:            action.height,
		AssetExec:         buy.AssetExec,
		PriceExec:         buy.PriceExec,
		PriceSymbol:       buy.PriceSymbol,
	}

	tokendb := newBuyDB(buyOrder)
	buyOrderKV := tokendb.save(action.db)
	logs = append(logs, receipt.Logs...)
	logs = append(logs, tokendb.getBuyLogs(pty.TyLogTradeBuyLimit, action.txhash, cfg))
	kv = append(kv, receipt.KV...)
	kv = append(kv, buyOrderKV...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func (action *tradeAction) tradeSellMarket(sellOrder *pty.TradeForSellMarket) (*types.Receipt, error) {
	cfg := action.api.GetConfig()
	if cfg.IsDappFork(action.height, pty.TradeX, pty.ForkTradeIDX) {
		sellOrder.BuyID = calcTokenBuyID(sellOrder.BuyID)
	}
	if sellOrder.BoardlotCnt < 0 || !strings.HasPrefix(sellOrder.BuyID, buyIDPrefix) {
		return nil, types.ErrInvalidParam
	}

	idByte := []byte(sellOrder.BuyID)
	buyOrder, err := getBuyOrderFromID(idByte, action.db)
	if err != nil {
		tradelog.Error("getBuyOrderFromID failed", "err", err)
		return nil, pty.ErrTBuyOrderNotExist
	}

	if buyOrder.Status == pty.TradeOrderStatusBoughtOut {
		return nil, pty.ErrTBuyOrderSoldout
	} else if buyOrder.Status == pty.TradeOrderStatusRevoked {
		return nil, pty.ErrTBuyOrderRevoked
	} else if buyOrder.Status == pty.TradeOrderStatusOnBuy && buyOrder.TotalBoardlot-buyOrder.BoughtBoardlot < sellOrder.BoardlotCnt {
		return nil, pty.ErrTBuyOrderNotEnough
	} else if buyOrder.Status == pty.TradeOrderStatusOnBuy && sellOrder.BoardlotCnt < buyOrder.MinBoardlot {
		return nil, pty.ErrTCntLessThanMinBoardlot
	}

	// 打token
	accDB, err := createAccountDB(cfg, action.height, action.db, buyOrder.AssetExec, buyOrder.TokenSymbol)
	if err != nil {
		tradelog.Error("createAccountDB failed", "err", err, "order", buyOrder)
		return nil, err
	}
	amountToken := sellOrder.BoardlotCnt * buyOrder.AmountPerBoardlot
	tradelog.Debug("tradeSellMarket", "step1 cnt", sellOrder.BoardlotCnt, "amountToken", amountToken)
	receiptFromExecAcc, err := accDB.ExecTransfer(action.fromaddr, buyOrder.Address, action.execaddr, amountToken)
	if err != nil {
		tradelog.Error("account.ExecTransfer token ", "error info", err, "addrFrom", buyOrder.Address,
			"addrTo", action.fromaddr, "execaddr", action.execaddr,
			"amountToken", amountToken)
		return nil, err
	}

	//首先购买费用的划转
	priceAcc, err := createPriceDB(cfg, action.height, action.db, buyOrder.PriceExec, buyOrder.PriceSymbol)
	if err != nil {
		return nil, err
	}
	amount := sellOrder.BoardlotCnt * buyOrder.PricePerBoardlot
	tradelog.Debug("tradeSellMarket", "step2 cnt", sellOrder.BoardlotCnt, "price", buyOrder.PricePerBoardlot, "amount", amount)
	receiptFromAcc, err := priceAcc.ExecTransferFrozen(buyOrder.Address, action.fromaddr, action.execaddr, amount)
	if err != nil {
		tradelog.Error("account.Transfer ", "addrFrom", buyOrder.Address, "addrTo", action.fromaddr,
			"amount", amount)
		// 因为未能成功将对应的币进行转账，所以需要回退
		accDB.ExecTransfer(buyOrder.Address, action.fromaddr, action.execaddr, amountToken)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	tradelog.Debug("tradeBuy", "BoughtBoardlot before this buy", buyOrder.BoughtBoardlot)
	buyOrder.BoughtBoardlot += sellOrder.BoardlotCnt
	tradelog.Debug("tradeBuy", "BoughtBoardlot after this buy", buyOrder.BoughtBoardlot)
	if buyOrder.BoughtBoardlot == buyOrder.TotalBoardlot {
		buyOrder.Status = pty.TradeOrderStatusBoughtOut
	}
	buyTokendb := newBuyDB(*buyOrder)
	sellOrderKV := buyTokendb.save(action.db)

	logs = append(logs, receiptFromAcc.Logs...)
	logs = append(logs, receiptFromExecAcc.Logs...)
	logs = append(logs, buyTokendb.getBuyLogs(pty.TyLogTradeBuyLimit, action.txhash, cfg))
	logs = append(logs, buyTokendb.getSellLogs(action.fromaddr, action.txhash, sellOrder.BoardlotCnt, action.txhash, cfg))
	kv = append(kv, receiptFromAcc.KV...)
	kv = append(kv, receiptFromExecAcc.KV...)
	kv = append(kv, sellOrderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *tradeAction) tradeRevokeBuyLimit(revoke *pty.TradeForRevokeBuy) (*types.Receipt, error) {
	cfg := action.api.GetConfig()
	if cfg.IsDappFork(action.height, pty.TradeX, pty.ForkTradeIDX) {
		revoke.BuyID = calcTokenBuyID(revoke.BuyID)
	}
	if !strings.HasPrefix(revoke.BuyID, buyIDPrefix) {
		return nil, types.ErrInvalidParam
	}
	buyIDByte := []byte(revoke.BuyID)
	buyOrder, err := getBuyOrderFromID(buyIDByte, action.db)
	if err != nil {
		return nil, pty.ErrTBuyOrderNotExist
	}

	if buyOrder.Status == pty.TradeOrderStatusBoughtOut {
		return nil, pty.ErrTBuyOrderSoldout
	} else if buyOrder.Status == pty.TradeOrderStatusBuyRevoked {
		return nil, pty.ErrTBuyOrderRevoked
	}

	if action.fromaddr != buyOrder.Address {
		return nil, pty.ErrTBuyOrderRevoke
	}

	priceAcc, err := createPriceDB(cfg, action.height, action.db, buyOrder.PriceExec, buyOrder.PriceSymbol)
	if err != nil {
		return nil, err
	}
	//然后实现购买token的转移,因为这部分token在之前的卖单生成时已经进行冻结
	tradeRest := (buyOrder.TotalBoardlot - buyOrder.BoughtBoardlot) * buyOrder.PricePerBoardlot
	//tradelog.Info("tradeRevokeBuyLimit", "total-b", buyOrder.TotalBoardlot, "price", buyOrder.PricePerBoardlot, "amount", tradeRest)
	receiptFromExecAcc, err := priceAcc.ExecActive(buyOrder.Address, action.execaddr, tradeRest)
	if err != nil {
		tradelog.Error("account.ExecActive bty ", "addrFrom", buyOrder.Address, "execaddr", action.execaddr, "amount", tradeRest)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	buyOrder.Status = pty.TradeOrderStatusBuyRevoked
	tokendb := newBuyDB(*buyOrder)
	sellOrderKV := tokendb.save(action.db)

	logs = append(logs, receiptFromExecAcc.Logs...)
	logs = append(logs, tokendb.getBuyLogs(pty.TyLogTradeBuyRevoke, action.txhash, cfg))
	kv = append(kv, receiptFromExecAcc.KV...)
	kv = append(kv, sellOrderKV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

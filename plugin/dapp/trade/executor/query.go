// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

// 目前设计trade 的query， 有两个部分的大分类
// 1. 按token 分
//    可以用于 token的挂单查询 (按价格排序)： OnBuy/OnSale
//    token 的历史行情 （按价格排序）: SoldOut/BoughtOut--> TODO 是否需要按时间（区块高度排序更合理）
// 2. 按 addr 分。 用于客户个人的钱包
//    自己未完成的交易 （按地址状态来）
//    自己的历史交易 （addr 的所有订单）
//
// 由于现价买/卖是没有orderID的， 用txhash 代替作为key
// key 有两种 orderID， txhash (0xAAAAAAAAAAAAAAA)

// 根据token 分页显示未完成成交卖单
func (t *trade) Query_GetTokenSellOrderByStatus(req *pty.ReqTokenSellOrder) (types.Message, error) {
	return t.GetTokenSellOrderByStatus(req, req.Status)
}

// 根据token 分页显示未完成成交买单
func (t *trade) Query_GetTokenBuyOrderByStatus(req *pty.ReqTokenBuyOrder) (types.Message, error) {
	if req.Status == 0 {
		req.Status = pty.TradeOrderStatusOnBuy
	}
	return t.GetTokenBuyOrderByStatus(req, req.Status)
}

// addr part
// addr(-token) 的所有订单， 不分页
func (t *trade) Query_GetOnesSellOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesSellOrder(req)
}

func (t *trade) Query_GetOnesBuyOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesBuyOrder(req)
}

// 按 用户状态来 addr-status
func (t *trade) Query_GetOnesSellOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesSellOrdersWithStatus(req)
}

func (t *trade) Query_GetOnesBuyOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesBuyOrdersWithStatus(req)
}

func (t *trade) Query_GetOnesOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesOrderWithStatus(req)
}

func (t *trade) Query_GetOneOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOneOrder(req)
}

func (t *trade) GetOnesSellOrder(addrTokens *pty.ReqAddrAssets) (types.Message, error) {
	var keys [][]byte
	if 0 == len(addrTokens.Token) {
		values, err := t.GetLocalDB().List(calcOnesSellOrderPrefixAddr(addrTokens.Addr), nil, 0, 0)
		if err != nil {
			return nil, err
		}
		if len(values) != 0 {
			tradelog.Debug("trade Query", "get number of sellID", len(values))
			keys = append(keys, values...)
		}
	} else {
		for _, token := range addrTokens.Token {
			values, err := t.GetLocalDB().List(calcOnesSellOrderPrefixToken(token, addrTokens.Addr), nil, 0, 0)
			tradelog.Debug("trade Query", "Begin to list addr with token", token, "got values", len(values))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			if len(values) != 0 {
				keys = append(keys, values...)
			}
		}
	}

	var replys pty.ReplyTradeOrders
	for _, key := range keys {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		replys.Orders = append(replys.Orders, reply)
	}
	return &replys, nil
}

func (t *trade) GetOnesBuyOrder(addrTokens *pty.ReqAddrAssets) (types.Message, error) {
	var keys [][]byte
	if 0 == len(addrTokens.Token) {
		values, err := t.GetLocalDB().List(calcOnesBuyOrderPrefixAddr(addrTokens.Addr), nil, 0, 0)
		if err != nil {
			return nil, err
		}
		if len(values) != 0 {
			tradelog.Debug("trade Query", "get number of buy keys", len(values))
			keys = append(keys, values...)
		}
	} else {
		for _, token := range addrTokens.Token {
			values, err := t.GetLocalDB().List(calcOnesBuyOrderPrefixToken(token, addrTokens.Addr), nil, 0, 0)
			tradelog.Debug("trade Query", "Begin to list addr with token", token, "got values", len(values))
			if err != nil && err != types.ErrNotFound {
				return nil, err
			}
			if len(values) != 0 {
				keys = append(keys, values...)
			}
		}
	}

	var replys pty.ReplyTradeOrders
	for _, key := range keys {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		replys.Orders = append(replys.Orders, reply)
	}

	return &replys, nil
}

func (t *trade) GetOnesSellOrdersWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	var sellIDs [][]byte
	values, err := t.GetLocalDB().List(calcOnesSellOrderPrefixStatus(req.Addr, req.Status), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(values) != 0 {
		tradelog.Debug("trade Query", "get number of sellID", len(values))
		sellIDs = append(sellIDs, values...)
	}

	var replys pty.ReplyTradeOrders
	for _, key := range sellIDs {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		replys.Orders = append(replys.Orders, reply)
	}

	return &replys, nil
}

func (t *trade) GetOnesBuyOrdersWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	var sellIDs [][]byte
	values, err := t.GetLocalDB().List(calcOnesBuyOrderPrefixStatus(req.Addr, req.Status), nil, 0, 0)
	if err != nil {
		return nil, err
	}
	if len(values) != 0 {
		tradelog.Debug("trade Query", "get number of buy keys", len(values))
		sellIDs = append(sellIDs, values...)
	}
	var replys pty.ReplyTradeOrders
	for _, key := range sellIDs {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		replys.Orders = append(replys.Orders, reply)
	}

	return &replys, nil
}

func (t *trade) GetTokenSellOrderByStatus(req *pty.ReqTokenSellOrder, status int32) (types.Message, error) {
	if req.Count <= 0 || (req.Direction != 1 && req.Direction != 0) {
		return nil, types.ErrInvalidParam
	}

	fromKey := []byte("")
	if len(req.FromKey) != 0 {
		sell := t.replyReplySellOrderfromID([]byte(req.FromKey))
		if sell == nil {
			tradelog.Error("GetTokenSellOrderByStatus", "key not exist", req.FromKey)
			return nil, types.ErrInvalidParam
		}
		fromKey = calcTokensSellOrderKeyStatus(sell.TokenSymbol, sell.Status,
			calcPriceOfToken(sell.PricePerBoardlot, sell.AmountPerBoardlot), sell.Owner, sell.Key)
	}
	values, err := t.GetLocalDB().List(calcTokensSellOrderPrefixStatus(req.TokenSymbol, status), fromKey, req.Count, req.Direction)
	if err != nil {
		return nil, err
	}
	var replys pty.ReplyTradeOrders
	for _, key := range values {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		replys.Orders = append(replys.Orders, reply)
	}
	return &replys, nil
}

func (t *trade) GetTokenBuyOrderByStatus(req *pty.ReqTokenBuyOrder, status int32) (types.Message, error) {
	if req.Count <= 0 || (req.Direction != 1 && req.Direction != 0) {
		return nil, types.ErrInvalidParam
	}

	fromKey := []byte("")
	if len(req.FromKey) != 0 {
		buy := t.replyReplyBuyOrderfromID([]byte(req.FromKey))
		if buy == nil {
			tradelog.Error("GetTokenBuyOrderByStatus", "key not exist", req.FromKey)
			return nil, types.ErrInvalidParam
		}
		fromKey = calcTokensBuyOrderKeyStatus(buy.TokenSymbol, buy.Status,
			calcPriceOfToken(buy.PricePerBoardlot, buy.AmountPerBoardlot), buy.Owner, buy.Key)
	}
	tradelog.Debug("GetTokenBuyOrderByStatus", "fromKey ", fromKey)

	// List Direction 是升序， 买单是要降序， 把高价买的放前面， 在下一页操作时， 显示买价低的。
	direction := 1 - req.Direction
	values, err := t.GetLocalDB().List(calcTokensBuyOrderPrefixStatus(req.TokenSymbol, status), fromKey, req.Count, direction)
	if err != nil {
		return nil, err
	}
	var replys pty.ReplyTradeOrders
	for _, key := range values {
		reply := t.loadOrderFromKey(key)
		if reply == nil {
			continue
		}
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		replys.Orders = append(replys.Orders, reply)
	}
	return &replys, nil
}

// query reply utils
func (t *trade) replyReplySellOrderfromID(key []byte) *pty.ReplySellOrder {
	tradelog.Debug("trade Query", "id", string(key), "check-prefix", sellIDPrefix)
	if strings.HasPrefix(string(key), sellIDPrefix) {
		if sellorder, err := getSellOrderFromID(key, t.GetStateDB()); err == nil {
			tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
			return sellOrder2reply(sellorder)
		}
	} else { // txhash as key
		txResult, err := getTx(key, t.GetLocalDB(), t.GetAPI())
		tradelog.Debug("GetOnesSellOrder ", "load txhash", string(key))
		if err != nil {
			return nil
		}
		return txResult2sellOrderReply(txResult)
	}
	return nil
}

func (t *trade) replyReplyBuyOrderfromID(key []byte) *pty.ReplyBuyOrder {
	tradelog.Debug("trade Query", "id", string(key), "check-prefix", buyIDPrefix)
	if strings.HasPrefix(string(key), buyIDPrefix) {
		if buyOrder, err := getBuyOrderFromID(key, t.GetStateDB()); err == nil {
			tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
			return buyOrder2reply(buyOrder)
		}
	} else { // txhash as key
		txResult, err := getTx(key, t.GetLocalDB(), t.GetAPI())
		tradelog.Debug("replyReplyBuyOrderfromID ", "load txhash", string(key))
		if err != nil {
			return nil
		}
		return txResult2buyOrderReply(txResult)
	}
	return nil
}

func sellOrder2reply(sellOrder *pty.SellOrder) *pty.ReplySellOrder {
	reply := &pty.ReplySellOrder{
		TokenSymbol:       sellOrder.TokenSymbol,
		Owner:             sellOrder.Address,
		AmountPerBoardlot: sellOrder.AmountPerBoardlot,
		MinBoardlot:       sellOrder.MinBoardlot,
		PricePerBoardlot:  sellOrder.PricePerBoardlot,
		TotalBoardlot:     sellOrder.TotalBoardlot,
		SoldBoardlot:      sellOrder.SoldBoardlot,
		BuyID:             "",
		Status:            sellOrder.Status,
		SellID:            sellOrder.SellID,
		TxHash:            strings.Replace(sellOrder.SellID, sellIDPrefix, "0x", 1),
		Height:            sellOrder.Height,
		Key:               sellOrder.SellID,
		AssetExec:         sellOrder.AssetExec,
	}
	return reply
}

func txResult2sellOrderReply(txResult *types.TxResult) *pty.ReplySellOrder {
	logs := txResult.Receiptdate.Logs
	tradelog.Debug("txResult2sellOrderReply", "show logs", logs)
	for _, log := range logs {
		if log.Ty == pty.TyLogTradeSellMarket {
			var receipt pty.ReceiptSellMarket
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			amount, err := strconv.ParseFloat(receipt.Base.AmountPerBoardlot, 64)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			price, err := strconv.ParseFloat(receipt.Base.PricePerBoardlot, 64)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}

			txhash := common.ToHex(txResult.GetTx().Hash())
			reply := &pty.ReplySellOrder{
				TokenSymbol:       receipt.Base.TokenSymbol,
				Owner:             receipt.Base.Owner,
				AmountPerBoardlot: int64(amount * float64(types.TokenPrecision)),
				MinBoardlot:       receipt.Base.MinBoardlot,
				PricePerBoardlot:  int64(price * float64(types.Coin)),
				TotalBoardlot:     receipt.Base.TotalBoardlot,
				SoldBoardlot:      receipt.Base.SoldBoardlot,
				BuyID:             receipt.Base.BuyID,
				Status:            pty.SellOrderStatus2Int[receipt.Base.Status],
				SellID:            "",
				TxHash:            txhash,
				Height:            receipt.Base.Height,
				Key:               txhash,
				AssetExec:         receipt.Base.AssetExec,
			}
			tradelog.Debug("txResult2sellOrderReply", "show reply", reply)
			return reply
		}
	}
	return nil
}

func buyOrder2reply(buyOrder *pty.BuyLimitOrder) *pty.ReplyBuyOrder {
	reply := &pty.ReplyBuyOrder{
		TokenSymbol:       buyOrder.TokenSymbol,
		Owner:             buyOrder.Address,
		AmountPerBoardlot: buyOrder.AmountPerBoardlot,
		MinBoardlot:       buyOrder.MinBoardlot,
		PricePerBoardlot:  buyOrder.PricePerBoardlot,
		TotalBoardlot:     buyOrder.TotalBoardlot,
		BoughtBoardlot:    buyOrder.BoughtBoardlot,
		BuyID:             buyOrder.BuyID,
		Status:            buyOrder.Status,
		SellID:            "",
		TxHash:            strings.Replace(buyOrder.BuyID, buyIDPrefix, "0x", 1),
		Height:            buyOrder.Height,
		Key:               buyOrder.BuyID,
		AssetExec:         buyOrder.AssetExec,
	}
	return reply
}

func txResult2buyOrderReply(txResult *types.TxResult) *pty.ReplyBuyOrder {
	logs := txResult.Receiptdate.Logs
	tradelog.Debug("txResult2sellOrderReply", "show logs", logs)
	for _, log := range logs {
		if log.Ty == pty.TyLogTradeBuyMarket {
			var receipt pty.ReceiptTradeBuyMarket
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			amount, err := strconv.ParseFloat(receipt.Base.AmountPerBoardlot, 64)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			price, err := strconv.ParseFloat(receipt.Base.PricePerBoardlot, 64)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			txhash := common.ToHex(txResult.GetTx().Hash())
			reply := &pty.ReplyBuyOrder{
				TokenSymbol:       receipt.Base.TokenSymbol,
				Owner:             receipt.Base.Owner,
				AmountPerBoardlot: int64(amount * float64(types.TokenPrecision)),
				MinBoardlot:       receipt.Base.MinBoardlot,
				PricePerBoardlot:  int64(price * float64(types.Coin)),
				TotalBoardlot:     receipt.Base.TotalBoardlot,
				BoughtBoardlot:    receipt.Base.BoughtBoardlot,
				BuyID:             "",
				Status:            pty.SellOrderStatus2Int[receipt.Base.Status],
				SellID:            receipt.Base.SellID,
				TxHash:            txhash,
				Height:            receipt.Base.Height,
				Key:               txhash,
				AssetExec:         receipt.Base.AssetExec,
			}
			tradelog.Debug("txResult2sellOrderReply", "show reply", reply)
			return reply
		}
	}
	return nil
}

const (
	orderStatusInvalid = iota
	orderStatusOn
	orderStatusDone
	orderStatusRevoke
)

const (
	orderTypeInvalid = iota
	orderTypeSell
	orderTypeBuy
)

func fromStatus(status int32) (st, ty int32) {
	switch status {
	case pty.TradeOrderStatusOnSale:
		return orderStatusOn, orderTypeSell
	case pty.TradeOrderStatusSoldOut:
		return orderStatusDone, orderTypeSell
	case pty.TradeOrderStatusRevoked:
		return orderStatusRevoke, orderTypeSell
	case pty.TradeOrderStatusOnBuy:
		return orderStatusOn, orderTypeBuy
	case pty.TradeOrderStatusBoughtOut:
		return orderStatusDone, orderTypeBuy
	case pty.TradeOrderStatusBuyRevoked:
		return orderStatusRevoke, orderTypeBuy
	}
	return orderStatusInvalid, orderTypeInvalid
}

// SellMarkMarket BuyMarket 没有tradeOrder 需要调用这个函数进行转化
// BuyRevoke, SellRevoke 也需要
// SellLimit/BuyLimit 有order 但order 里面没有 bolcktime， 直接访问 order 还需要再次访问 block， 还不如直接访问交易
func buyBase2Order(base *pty.ReceiptBuyBase, txHash string, blockTime int64) *pty.ReplyTradeOrder {
	amount, err := strconv.ParseFloat(base.AmountPerBoardlot, 64)
	if err != nil {
		tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
		return nil
	}
	price, err := strconv.ParseFloat(base.PricePerBoardlot, 64)
	if err != nil {
		tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
		return nil
	}
	key := txHash
	if len(base.BuyID) > 0 {
		key = base.BuyID
	}
	//txhash := common.ToHex(txResult.GetTx().Hash())
	reply := &pty.ReplyTradeOrder{
		TokenSymbol:       base.TokenSymbol,
		Owner:             base.Owner,
		AmountPerBoardlot: int64(amount * float64(types.TokenPrecision)),
		MinBoardlot:       base.MinBoardlot,
		PricePerBoardlot:  int64(price * float64(types.Coin)),
		TotalBoardlot:     base.TotalBoardlot,
		TradedBoardlot:    base.BoughtBoardlot,
		BuyID:             base.BuyID,
		Status:            pty.SellOrderStatus2Int[base.Status],
		SellID:            base.SellID,
		TxHash:            txHash,
		Height:            base.Height,
		Key:               key,
		BlockTime:         blockTime,
		IsSellOrder:       false,
		AssetExec:         base.AssetExec,
	}
	tradelog.Debug("txResult2sellOrderReply", "show reply", reply)
	return reply
}

func sellBase2Order(base *pty.ReceiptSellBase, txHash string, blockTime int64) *pty.ReplyTradeOrder {
	amount, err := strconv.ParseFloat(base.AmountPerBoardlot, 64)
	if err != nil {
		tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
		return nil
	}
	price, err := strconv.ParseFloat(base.PricePerBoardlot, 64)
	if err != nil {
		tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
		return nil
	}
	//txhash := common.ToHex(txResult.GetTx().Hash())
	key := txHash
	if len(base.SellID) > 0 {
		key = base.SellID
	}
	reply := &pty.ReplyTradeOrder{
		TokenSymbol:       base.TokenSymbol,
		Owner:             base.Owner,
		AmountPerBoardlot: int64(amount * float64(types.TokenPrecision)),
		MinBoardlot:       base.MinBoardlot,
		PricePerBoardlot:  int64(price * float64(types.Coin)),
		TotalBoardlot:     base.TotalBoardlot,
		TradedBoardlot:    base.SoldBoardlot,
		BuyID:             base.BuyID,
		Status:            pty.SellOrderStatus2Int[base.Status],
		SellID:            base.SellID,
		TxHash:            txHash,
		Height:            base.Height,
		Key:               key,
		BlockTime:         blockTime,
		IsSellOrder:       true,
		AssetExec:         base.AssetExec,
	}
	tradelog.Debug("txResult2sellOrderReply", "show reply", reply)
	return reply
}

func txResult2OrderReply(txResult *types.TxResult) *pty.ReplyTradeOrder {
	logs := txResult.Receiptdate.Logs
	tradelog.Debug("txResult2sellOrderReply", "show logs", logs)
	for _, log := range logs {
		if log.Ty == pty.TyLogTradeBuyMarket {
			var receipt pty.ReceiptTradeBuyMarket
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			return buyBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		} else if log.Ty == pty.TyLogTradeBuyRevoke {
			var receipt pty.ReceiptTradeBuyRevoke
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			return buyBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		} else if log.Ty == pty.TyLogTradeSellRevoke {
			var receipt pty.ReceiptTradeSellRevoke
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show revoke 1 ", receipt)
			return sellBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		} else if log.Ty == pty.TyLogTradeSellMarket {
			var receipt pty.ReceiptSellMarket
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			return sellBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		}
	}
	return nil
}

func limitOrderTxResult2Order(txResult *types.TxResult) *pty.ReplyTradeOrder {
	logs := txResult.Receiptdate.Logs
	tradelog.Debug("txResult2sellOrderReply", "show logs", logs)
	for _, log := range logs {
		if log.Ty == pty.TyLogTradeSellLimit {
			var receipt pty.ReceiptTradeSellLimit
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			return sellBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		} else if log.Ty == pty.TyLogTradeBuyLimit {
			var receipt pty.ReceiptTradeBuyLimit
			err := types.Decode(log.Log, &receipt)
			if err != nil {
				tradelog.Error("txResult2sellOrderReply", "decode receipt", err)
				return nil
			}
			tradelog.Debug("txResult2sellOrderReply", "show logs 1 ", receipt)
			return buyBase2Order(receipt.Base, common.ToHex(txResult.GetTx().Hash()), txResult.Blocktime)
		}
	}
	return nil
}

func (t *trade) loadOrderFromKey(key []byte) *pty.ReplyTradeOrder {
	tradelog.Debug("trade Query", "id", string(key), "check-prefix", sellIDPrefix)
	if strings.HasPrefix(string(key), sellIDPrefix) {
		txHash := strings.Replace(string(key), sellIDPrefix, "0x", 1)
		txResult, err := getTx([]byte(txHash), t.GetLocalDB(), t.GetAPI())
		tradelog.Debug("loadOrderFromKey ", "load txhash", txResult)
		if err != nil {
			return nil
		}
		reply := limitOrderTxResult2Order(txResult)

		sellOrder, err := getSellOrderFromID(key, t.GetStateDB())
		tradelog.Debug("trade Query", "getSellOrderFromID", string(key))
		if err != nil {
			return nil
		}
		reply.TradedBoardlot = sellOrder.SoldBoardlot
		reply.Status = sellOrder.Status
		return reply
	} else if strings.HasPrefix(string(key), buyIDPrefix) {
		txHash := strings.Replace(string(key), buyIDPrefix, "0x", 1)
		txResult, err := getTx([]byte(txHash), t.GetLocalDB(), t.GetAPI())
		tradelog.Debug("loadOrderFromKey ", "load txhash", txResult)
		if err != nil {
			return nil
		}
		reply := limitOrderTxResult2Order(txResult)

		buyOrder, err := getBuyOrderFromID(key, t.GetStateDB())
		if err != nil {
			return nil
		}
		reply.TradedBoardlot = buyOrder.BoughtBoardlot
		reply.Status = buyOrder.Status
		return reply
	}
	txResult, err := getTx(key, t.GetLocalDB(), t.GetAPI())
	tradelog.Debug("loadOrderFromKey ", "load txhash", string(key))
	if err != nil {
		return nil
	}
	return txResult2OrderReply(txResult)
}

func (t *trade) GetOnesOrderWithStatusV1(req *pty.ReqAddrAssets) (types.Message, error) {
	fromKey := []byte("")
	if len(req.FromKey) != 0 {
		order := t.loadOrderFromKey([]byte(req.FromKey))
		if order == nil {
			tradelog.Error("GetOnesOrderWithStatus", "key not exist", req.FromKey)
			return nil, types.ErrInvalidParam
		}
		st, ty := fromStatus(order.Status)
		fromKey = calcOnesOrderKey(order.Owner, st, ty, order.Height, order.Key)
	}

	orderStatus, orderType := fromStatus(req.Status)
	if orderStatus == orderStatusInvalid || orderType == orderTypeInvalid {
		return nil, types.ErrInvalidParam
	}

	keys, err := t.GetLocalDB().List(calcOnesOrderPrefixStatus(req.Addr, orderStatus), fromKey, req.Count, req.Direction)
	if err != nil {
		return nil, err
	}

	var replys pty.ReplyTradeOrders
	for _, key := range keys {
		reply := t.loadOrderFromKey(key)
		if reply != nil {
			replys.Orders = append(replys.Orders, reply)
			tradelog.Debug("trade Query", "height of key", reply.Height,
				"len of reply.Selloders", len(replys.Orders))
		}
	}
	return &replys, nil
}

func (t *trade) GetOnesOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	orderStatus, orderType := fromStatus(req.Status)
	if orderStatus == orderStatusInvalid || orderType == orderTypeInvalid {
		return nil, types.ErrInvalidParam
	}

	// 使用 owner isFinished 组合
	var order pty.LocalOrder
	if orderStatus == orderStatusOn {
		order.IsFinished = false
	} else {
		order.IsFinished = true
	}
	order.Owner = req.Addr
	if len(req.FromKey) > 0 {
		order.TxIndex = req.FromKey
	}
	rows, err := list(t.GetLocalDB(), "owner_isFinished", &order, req.Count, req.Direction)
	if err != nil {
		tradelog.Error("GetOnesOrderWithStatus", "err", err)
		return nil, err
	}
	var replys pty.ReplyTradeOrders
	for _, row := range rows {
		o, ok := row.Data.(*pty.LocalOrder)
		if !ok {
			tradelog.Error("GetOnesOrderWithStatus", "err", "bad row type")
			return nil, types.ErrTypeAsset
		}
		reply := fmtReply(o)
		replys.Orders = append(replys.Orders, reply)
	}
	return &replys, nil
}

func fmtReply(order *pty.LocalOrder) *pty.ReplyTradeOrder {
	return &pty.ReplyTradeOrder{
		TokenSymbol:       order.AssetSymbol,
		Owner:             order.Owner,
		AmountPerBoardlot: order.AmountPerBoardlot,
		MinBoardlot:       order.MinBoardlot,
		PricePerBoardlot:  order.PricePerBoardlot,
		TotalBoardlot:     order.TotalBoardlot,
		TradedBoardlot:    order.TradedBoardlot,
		BuyID:             order.BuyID,
		Status:            order.Status,
		SellID:            order.SellID,
		TxHash:            order.TxHash[0],
		Height:            order.Height,
		Key:               order.TxIndex,
		BlockTime:         order.BlockTime,
		IsSellOrder:       order.IsSellOrder,
		AssetExec:         order.AssetExec,
	}
}

func (t *trade) GetOneOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	query := NewOrderTable(t.GetLocalDB())
	tradelog.Debug("query GetData dbg", "primary", req.FromKey)
	row, err := query.GetData([]byte(req.FromKey))
	if err != nil {
		tradelog.Error("query GetData failed", "key", req.FromKey, "err", err)
		return nil, err
	}

	o, ok := row.Data.(*pty.LocalOrder)
	if !ok {
		tradelog.Error("query GetData failed", "err", "bad row type")
		return nil, types.ErrTypeAsset
	}
	reply := fmtReply(o)

	return reply, nil
}

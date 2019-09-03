package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/trade/types"
)

// 1.8 根据token 分页显示未完成成交卖单
func (t *trade) Query_GetTokenSellOrderByStatus(req *pty.ReqTokenSellOrder) (types.Message, error) {
	return t.GetTokenSellOrderByStatus(req, req.Status)
}

// GetTokenSellOrderByStatus by status
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

// 1.3 根据token 分页显示未完成成交买单
func (t *trade) Query_GetTokenBuyOrderByStatus(req *pty.ReqTokenBuyOrder) (types.Message, error) {
	if req.Status == 0 {
		req.Status = pty.TradeOrderStatusOnBuy
	}
	return t.GetTokenBuyOrderByStatus(req, req.Status)
}

// GetTokenBuyOrderByStatus by status
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

// addr part
// 1.4 addr(-token) 的所有订单， 不分页
func (t *trade) Query_GetOnesSellOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesSellOrder(req)
}

// 1.1 addr(-token) 的所有订单， 不分页
func (t *trade) Query_GetOnesBuyOrder(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesBuyOrder(req)
}

// GetOnesSellOrder by address or address-token
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

// GetOnesBuyOrder by address or address-token
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

// 1.5 没找到
// 按 用户状态来 addr-status
func (t *trade) Query_GetOnesSellOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesSellOrdersWithStatus(req)
}

// 1.2 按 用户状态来 addr-status
func (t *trade) Query_GetOnesBuyOrderWithStatus(req *pty.ReqAddrAssets) (types.Message, error) {
	return t.GetOnesBuyOrdersWithStatus(req)
}

// GetOnesSellOrdersWithStatus by address-status
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

// GetOnesBuyOrdersWithStatus by address-status
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

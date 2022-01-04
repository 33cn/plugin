package client

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var cfg = types.NewChain33Config(types.GetDefaultCfgstring())

type ExchangeClient struct {
	client   Cli
	txHeight int64
}

func NewExchangCient(cli Cli) *ExchangeClient {
	return &ExchangeClient{
		client:   cli,
		txHeight: types.LowAllowPackHeight,
	}
}

func (c *ExchangeClient) QueryMarketDepth(msg proto.Message) (types.Message, error) {
	data, err := c.client.Query(et.FuncNameQueryMarketDepth, msg)
	if err != nil {
		return nil, err
	}
	var resp et.MarketDepthList
	err = types.Decode(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ExchangeClient) QueryHistoryOrderList(msg proto.Message) (types.Message, error) {
	data, err := c.client.Query(et.FuncNameQueryHistoryOrderList, msg)
	if err != nil {
		return nil, err
	}
	var resp et.OrderList
	err = types.Decode(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ExchangeClient) QueryOrder(msg proto.Message) (types.Message, error) {
	data, err := c.client.Query(et.FuncNameQueryOrder, msg)
	if err != nil {
		return nil, err
	}
	var resp et.Order
	err = types.Decode(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ExchangeClient) QueryOrderList(msg proto.Message) (types.Message, error) {
	data, err := c.client.Query(et.FuncNameQueryOrderList, msg)
	if err != nil {
		return nil, err
	}
	var resp et.OrderList
	err = types.Decode(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *ExchangeClient) LimitOrder(msg proto.Message, hexKey string) (*et.ReceiptExchange, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("LimitOrder", msg)
	if err != nil {
		return nil, err
	}
	logs, err := c.client.Send(tx, hexKey)
	if err != nil {
		return nil, err
	}
	var resp et.ReceiptExchange
	for _, l := range logs {
		if l.Ty == et.TyLimitOrderLog {
			err = types.Decode(l.Log, &resp)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return &resp, nil
}

//TODO marketOrder
func (c *ExchangeClient) MarketOrder(msg proto.Message, hexKey string) (*et.ReceiptExchange, error) {
	return nil, errors.New("Unopen")
	//ety := types.LoadExecutorType(et.ExchangeX)
	//tx, err := ety.Create("MarketOrder", msg)
	//if err != nil {
	//	return nil, err
	//}
	//logs, err := c.client.Send(tx, hexKey)
	//if err != nil {
	//	return nil, err
	//}
	//var resp et.ReceiptExchange
	//for _, l := range logs {
	//	if l.Ty == et.TyMarketOrderLog {
	//		err = types.Decode(l.Log, &resp)
	//		if err != nil {
	//			return nil, err
	//		}
	//		break
	//	}
	//}
	//return &resp, nil
}

func (c *ExchangeClient) RevokeOrder(msg proto.Message, hexKey string) (*et.ReceiptExchange, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("RevokeOrder", msg)
	if err != nil {
		return nil, err
	}
	logs, err := c.client.Send(tx, hexKey)
	if err != nil {
		return nil, err
	}
	var resp et.ReceiptExchange
	for _, l := range logs {
		if l.Ty == et.TyRevokeOrderLog {
			err = types.Decode(l.Log, &resp)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return &resp, nil
}

func (c *ExchangeClient) ExchangeBind(msg proto.Message, hexKey string) (*et.ReceiptExchangeBind, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("ExchangeBind", msg)
	if err != nil {
		return nil, err
	}
	logs, err := c.client.Send(tx, hexKey)
	if err != nil {
		return nil, err
	}
	var resp et.ReceiptExchangeBind
	for _, l := range logs {
		if l.Ty == et.TyExchangeBindLog {
			err = types.Decode(l.Log, &resp)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return &resp, nil
}

func (c *ExchangeClient) EntrustOrder(msg proto.Message, hexKey string) (*et.ReceiptExchange, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("EntrustOrder", msg)
	if err != nil {
		return nil, err
	}
	logs, err := c.client.Send(tx, hexKey)
	if err != nil {
		return nil, err
	}
	var resp et.ReceiptExchange
	for _, l := range logs {
		if l.Ty == et.TyLimitOrderLog {
			err = types.Decode(l.Log, &resp)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return &resp, nil
}

func (c *ExchangeClient) EntrustRevokeOrder(msg proto.Message, hexKey string) (*et.ReceiptExchange, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("EntrustRevokeOrder", msg)
	if err != nil {
		return nil, err
	}
	logs, err := c.client.Send(tx, hexKey)
	if err != nil {
		return nil, err
	}
	var resp et.ReceiptExchange
	for _, l := range logs {
		if l.Ty == et.TyRevokeOrderLog {
			err = types.Decode(l.Log, &resp)
			if err != nil {
				return nil, err
			}
			break
		}
	}
	return &resp, nil
}

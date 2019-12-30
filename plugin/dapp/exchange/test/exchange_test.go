package test

import (
	"context"
	"testing"

	"github.com/33cn/plugin/plugin/dapp/exchange/executor"

	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	tt "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"
)

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	cli      Cli
	orderId  int64
	orderId2 int64
	coin     = "bty"
	token    = "CCNY"
)

func init() {
	//cli = NewExecCli()
	cli = NewGRPCCli(":8802")
}

func TestLimitOrder2(t *testing.T) {
	req := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      4,
		Amount:     10 * types.Coin,
		Op:         et.OpBuy,
	}
	testLimitOrder(t, req, Nodes[0], PrivKeyA)
}

func TestLimitOrder(t *testing.T) {
	//A 挂买 4x10
	req := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      4,
		Amount:     10 * types.Coin,
		Op:         et.OpBuy,
	}
	_, err := doLimitOrder(req, PrivKeyA)
	assert.Nil(t, err)
}

func TestOrderList(t *testing.T) {
	orderList, err := getOrderList(et.Ordered, Nodes[0], "")
	assert.Nil(t, err)
	t.Log(orderList)
	orderId = orderList.List[0].OrderID
}

func TestGetOrder(t *testing.T) {
	order, err := getOrder(orderId)
	assert.Nil(t, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	assert.Equal(t, 10*types.Coin, order.Balance)
}

func TestMarketDepth(t *testing.T) {
	depth, err := getMarketDepth(&et.QueryMarketDepth{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Op:         et.OpBuy,
	})
	assert.Nil(t, err)
	t.Log(depth)
	assert.Equal(t, 10*types.Coin, depth.List[0].Amount)
}

func TestMatch(t *testing.T) {
	//B 挂卖 4x5
	req2 := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      4,
		Amount:     5 * types.Coin,
		Op:         et.OpSell,
	}
	_, err := doLimitOrder(req2, PrivKeyB)
	assert.Nil(t, err)

	orderList, err := getOrderList(et.Completed, Nodes[1], "")
	assert.Nil(t, err)
	t.Log(orderList)
	orderId2 = orderList.List[0].OrderID

	//订单1的状态应该还是ordered
	order, err := getOrder(orderId)
	assert.Nil(t, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	assert.Equal(t, 5*types.Coin, order.Balance)

	//order2状态是completed
	order, err = getOrder(orderId2)
	assert.Nil(t, err)
	assert.Equal(t, int32(et.Completed), order.Status)

	//买盘还有 5
	depth, err := getMarketDepth(&et.QueryMarketDepth{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Op:         et.OpBuy,
	})
	assert.Nil(t, err)
	t.Log(depth)
	assert.Equal(t, 5*types.Coin, depth.List[0].Amount)
}

func TestHistoryOrderList(t *testing.T) {
	historyq := &et.QueryHistoryOrderList{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
	}
	historyOrderList, err := getHistoryOrderList(historyq)
	assert.Nil(t, err)
	t.Log(historyOrderList)
	assert.Equal(t, orderId2, historyOrderList.List[0].OrderID)
}

func TestRevokeOrder(t *testing.T) {
	//A 撤回未完成订单
	_, err := doRevokeOrder(orderId, PrivKeyA)
	assert.Nil(t, err)

	//根据订单号，查询订单详情
	order, err := getOrder(orderId)
	assert.Nil(t, err)
	assert.Equal(t, int32(et.Revoked), order.Status)
	assert.Equal(t, 5*types.Coin, order.Balance)

	//查询市场深度，买盘应该为空
	depth, err := getMarketDepth(&et.QueryMarketDepth{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Op:         et.OpBuy,
	})
	assert.Nil(t, depth)
	depth, err = getMarketDepth(&et.QueryMarketDepth{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Op:         et.OpSell,
	})
	assert.Nil(t, depth)
}

func BenchmarkOrder(b *testing.B) {
	req := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      1,
		Amount:     10 * types.Coin,
		Op:         et.OpSell,
	}
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, _ := ety.Create("LimitOrder", req)
	for i := 0; i < b.N; i++ {
		cli.Send(tx, PrivKeyA)
	}
}

func doLimitOrder(req *et.LimitOrder, privKey string) ([]*types.ReceiptLog, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("LimitOrder", req)
	if err != nil {
		return nil, err
	}
	return cli.Send(tx, privKey)
}

func doRevokeOrder(orderID int64, privKey string) ([]*types.ReceiptLog, error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err := ety.Create("RevokeOrder", &et.RevokeOrder{OrderID: orderID})
	if err != nil {
		return nil, err
	}
	return cli.Send(tx, privKey)
}

func getOrderList(status int32, addr string, primaryKey string) (*et.OrderList, error) {
	msg, err := cli.Query(et.FuncNameQueryOrderList, &et.QueryOrderList{Status: status, Address: addr, PrimaryKey: primaryKey})
	if err != nil {
		return nil, err
	}

	var resp et.OrderList
	err = types.Decode(msg, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func getOrder(orderID int64) (*et.Order, error) {
	msg, err := cli.Query(et.FuncNameQueryOrder, &et.QueryOrder{OrderID: orderID})
	if err != nil {
		return nil, err
	}

	var resp et.Order
	err = types.Decode(msg, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func getMarketDepth(query *et.QueryMarketDepth) (*et.MarketDepthList, error) {
	msg, err := cli.Query(et.FuncNameQueryMarketDepth, query)
	if err != nil {
		return nil, err
	}

	var resp et.MarketDepthList
	err = types.Decode(msg, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func getHistoryOrderList(query *et.QueryHistoryOrderList) (*et.OrderList, error) {
	msg, err := cli.Query(et.FuncNameQueryHistoryOrderList, query)
	if err != nil {
		return nil, err
	}

	var resp et.OrderList
	err = types.Decode(msg, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func getExcerBalance(addr string) (*types.Account, error) {
	var addrs []string
	addrs = append(addrs, addr)
	params := &types.ReqBalance{
		Addresses: addrs,
		Execer:    et.ExchangeX,
	}

	accs, err := cli.(*GRPCCli).client.GetBalance(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return accs.Acc[0], nil
}

func getExcerTokenBalance(addr string, symbol string) (map[string]*types.Account, error) {
	var addrs []string
	addrs = append(addrs, addr)
	param := &tt.ReqTokenBalance{
		Addresses:   addrs,
		TokenSymbol: symbol,
		Execer:      et.ExchangeX,
	}
	msg, err := cli.Query("token.GetAccountTokenBalance", param)
	if err != nil {
		return nil, err
	}

	var resp tt.ReplyAccountTokenAssets
	err = types.Decode(msg, &resp)
	if err != nil {
		return nil, err
	}

	assets := make(map[string]*types.Account)
	for _, v := range resp.TokenAssets {
		assets[v.Symbol] = v.Account
	}
	return assets, nil
}

func testLimitOrder(t *testing.T, req *et.LimitOrder, addr string, privkey string) {
	accPrev, err := getExcerBalance(addr)
	assert.Nil(t, err)
	t.Log(accPrev)

	tokenPrev, err := getExcerTokenBalance(addr, "CCNY")
	assert.Nil(t, err)
	t.Log(tokenPrev)

	_, err = doLimitOrder(req, privkey)
	assert.Nil(t, err)

	accAfter, err := getExcerBalance(addr)
	assert.Nil(t, err)
	t.Log(accAfter)

	tokenAfter, err := getExcerTokenBalance(addr, "CCNY")
	assert.Nil(t, err)
	t.Log(tokenAfter)

	cost := executor.SafeMul(req.Amount, req.Price)
	t.Log(req.Amount, req.Price, cost)
	// bty/ccny
	if req.Op == et.OpBuy {
		// bty
		assert.Equal(t, accAfter.Balance, accPrev.Balance)
		assert.Equal(t, accAfter.Frozen, accPrev.Frozen)
		// ccny
		assert.Equal(t, tokenAfter[token].Balance, tokenPrev[token].Balance-cost)
		assert.Equal(t, tokenAfter[token].Frozen, tokenPrev[token].Frozen+cost)
	} else {
		// bty
		assert.Equal(t, accAfter.Balance, accPrev.Balance-req.Amount)
		assert.Equal(t, accAfter.Frozen, accPrev.Frozen+req.Amount)
		// ccny
		assert.Equal(t, tokenAfter[token].Balance, tokenPrev[token].Balance)
		assert.Equal(t, tokenAfter[token].Frozen, tokenPrev[token].Frozen)
	}
}

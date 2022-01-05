package client_test

import (
	"testing"

	"github.com/33cn/chain33/types"
	excli "github.com/33cn/plugin/plugin/dapp/exchange/client"
	etypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

var (
	addrA    = "1BsJugqWiF47x2c915YW2TkBKVLU9GmvEn"
	privKeyA = "0x13169cd9ecf0d3e4a78d1a97a9abb506cb12b2f45c127a8a33c84f389a38e674"
	addrB    = "1FpPrLgyuR6reqj8LQ3HNVHkjtA7hAcvHo"
	privKeyB = "0xf15b088b01051caccb668e00e3c891d044cdec30de856af72c02c0abe1ba90ec"
	grpcAddr = "127.0.0.1:8802"
	execer   = "token"
)

func TestQueryMarketDepth(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.QueryMarketDepth{
		LeftAsset:  &etypes.Asset{Symbol: "BTC", Execer: execer},
		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
		Op:         2,
		PrimaryKey: "",
		Count:      10,
	}
	resp, err := client.QueryMarketDepth(req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("QueryMarketDepth", resp.(*etypes.MarketDepthList).List)
}

func TestQueryHistoryOrderList(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.QueryHistoryOrderList{
		LeftAsset:  &etypes.Asset{Symbol: "BTC", Execer: execer},
		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
		PrimaryKey: "",
		Count:      10,
		Direction:  0,
	}
	resp, err := client.QueryHistoryOrderList(req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("QueryHistory", resp.(*etypes.OrderList).List)
}

func TestQueryOrder(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.QueryOrder{
		OrderID: 28000000000,
	}
	resp, err := client.QueryOrder(req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("QueryOrder", resp.(*etypes.Order))
}

func TestQueryOrderList(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.QueryOrderList{
		Status:     1,
		Address:    "1FpPrLgyuR6reqj8LQ3HNVHkjtA7hAcvHo",
		PrimaryKey: "",
		Count:      10,
		Direction:  0,
	}
	resp, err := client.QueryOrderList(req)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("QueryOrderList", resp.(*etypes.OrderList).List)
}

func TestLimitOrder(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.LimitOrder{
		LeftAsset:  &etypes.Asset{Symbol: "BTC", Execer: execer},
		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
		Op:         etypes.OpSell,
		Price:      400 * types.DefaultCoinPrecision,
		Amount:     1 * types.DefaultCoinPrecision,
	}

	resp, err := client.LimitOrder(req, privKeyA)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("LimitOrder", resp)
}

//TODO marketOrder
//func TestMarketOrder(t *testing.T) {
//	cli := excli.NewGRPCCli(grpcAddr)
//	client := excli.NewExchangCient(cli)
//
//	req := &etypes.MarketOrder{
//		LeftAsset:  &etypes.Asset{Symbol: "BTC", Execer: execer},
//		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
//		Op:         etypes.OpSell,
//		Amount:     4 * types.DefaultCoinPrecision,
//	}
//
//	resp, err := client.MarketOrder(req, privKey)
//	if err != nil {
//		t.Log(err)
//		return
//	}
//	t.Log("MarketOrder", resp)
//}

func TestRevokeOrder(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.RevokeOrder{
		OrderID: 88000000000,
	}

	resp, err := client.RevokeOrder(req, privKeyA)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("RevokeOrder", resp)
}

func TestExchangeBind(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.ExchangeBind{
		ExchangeAddress: addrA,
		EntrustAddress:  addrB,
	}

	resp, err := client.ExchangeBind(req, privKeyA)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("ExchangeBind", resp)
}

func TestEntrustOrder(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.EntrustOrder{
		LeftAsset:  &etypes.Asset{Symbol: "BTC", Execer: execer},
		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
		Op:         etypes.OpSell,
		Price:      400 * types.DefaultCoinPrecision,
		Amount:     1 * types.DefaultCoinPrecision,
		Addr:       addrA,
	}

	resp, err := client.EntrustOrder(req, privKeyB)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("EntrustOrder", resp)
}

func TestEntrustRevokeOrder(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.EntrustRevokeOrder{
		OrderID: 42000000000,
		Addr:    addrA,
	}

	resp, err := client.EntrustRevokeOrder(req, privKeyB)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("EntrustRevokeOrder", resp)
}

package client_test

import (
	"testing"

	"github.com/33cn/chain33/types"
	excli "github.com/33cn/plugin/plugin/dapp/exchange/client"
	etypes "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

var (
	privKey  = "0x13169cd9ecf0d3e4a78d1a97a9abb506cb12b2f45c127a8a33c84f389a38e674" //1BsJugqWiF47x2c915YW2TkBKVLU9GmvEn
	grpcAddr = "127.0.0.1:8802"
	execer   = "token"
)

func TestQueryMarketDepth(t *testing.T) {
	cli := excli.NewGRPCCli(grpcAddr)
	client := excli.NewExchangCient(cli)

	req := &etypes.QueryMarketDepth{
		LeftAsset:  &etypes.Asset{Symbol: "BTY", Execer: execer},
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
		LeftAsset:  &etypes.Asset{Symbol: "BTY", Execer: execer},
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
		OrderID: 46000000000,
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
		LeftAsset:  &etypes.Asset{Symbol: "BTY", Execer: execer},
		RightAsset: &etypes.Asset{Symbol: "USDT", Execer: execer},
		Op:         etypes.OpSell,
		Price:      4 * types.DefaultCoinPrecision,
		Amount:     4 * types.DefaultCoinPrecision,
	}

	resp, err := client.LimitOrder(req, privKey)
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
//		LeftAsset:  &etypes.Asset{Symbol: "BTY", Execer: execer},
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

	resp, err := client.RevokeOrder(req, privKey)
	if err != nil {
		t.Log(err)
		return
	}
	t.Log("RevokeOrder", resp)
}

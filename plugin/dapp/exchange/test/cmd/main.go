package main

import (
	"fmt"
	"sync"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/exchange/test"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

var (
	cli      *test.GRPCCli
	Genesis  = "0x3990969DF92A5914F7B71EEB9A4E58D6E255F32BF042FEA5318FC8B3D50EE6E8" // 1CbEVT9RnM5oZhWMj4fxUrJX94VtRotzvs
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = []string{
		"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
		"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
		"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k",
		"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs",
	}
)

// 批量测试前，先确保测试账户有足够的币和钱
func main() {
	cli = test.NewGRPCCli("localhost:8802")
	var group sync.WaitGroup
	group.Add(2000)
	for i:=0;i<1000;i++{
		go buy(&group)
		go sell(&group)
	}
	group.Wait()

}

func sell(group *sync.WaitGroup) {
	req := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      1*types.Coin,
		Amount:     types.Coin,
		Op:         et.OpSell,
	}
	ety := types.LoadExecutorType(et.ExchangeX)

	// 卖 2000 次，需 2000*1=2000 个 bty
	for i := 0; i < 2; i++ {
		fmt.Println("sell ", i)
		tx, err := ety.Create("LimitOrder", req)
		if err != nil {
			panic(err)
		}
		//go cli.SendTx(tx, PrivKeyA)
		cli.SendTx(tx, PrivKeyA)
	}
	group.Done()

}

func buy(group *sync.WaitGroup) {
	req := &et.LimitOrder{
		LeftAsset:  &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"},
		Price:      1*types.Coin,
		Amount:     types.Coin,
		Op:         et.OpBuy,
	}
	ety := types.LoadExecutorType(et.ExchangeX)
	// 买 2000 次，需 2000*1=2000 个 ccny
	for i := 0; i < 2; i++ {
		fmt.Println("buy ", i)
		tx, err := ety.Create("LimitOrder", req)
		if err != nil {
			panic(err)
		}
		//go cli.SendTx(tx, PrivKeyB)
		cli.SendTx(tx, PrivKeyB)
	}
	group.Done()
}

package executor

import (
	"testing"

	"github.com/33cn/chain33/common/db"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
	"github.com/stretchr/testify/assert"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
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

func TestExchange(t *testing.T) {
	//环境准备
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	Init(et.ExchangeX, cfg, nil)
	total := 100 * types.DefaultCoinPrecision
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}

	accountC := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[2],
	}
	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[3],
	}
	dir, stateDB, kvdb := util.CreateTestDB()
	//defer util.CloseTestDB(dir, stateDB)
	execAddr := address.ExecAddress(et.ExchangeX)

	accA, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accC, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accC.SaveExecAccount(execAddr, &accountC)

	accD, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accD.SaveExecAccount(execAddr, &accountD)

	accA1, _ := account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accA1.SaveExecAccount(execAddr, &accountA)

	accB1, _ := account.NewAccountDB(cfg, "paracross", "coins.bty", stateDB)
	accB1.SaveExecAccount(execAddr, &accountB)

	accC1, _ := account.NewAccountDB(cfg, "paracross", "token.CCNY", stateDB)
	accC1.SaveExecAccount(execAddr, &accountC)

	accD1, _ := account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accD1.SaveExecAccount(execAddr, &accountD)

	env := &execEnv{
		10,
		1,
		1539918074,
	}

	/*
	  买卖单价格相同，测试正常撮合流程，查询功能是否可用
	 用例说明：
	   1.先挂数量是10的买单。
	   2.然后再挂数量是5的吃单
	   3.最后撤销未成交部分的买单
	*/

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 4, Amount: 10 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyA, stateDB, kvdb, env)
	//根据地址状态查看订单,最新得订单号永远是在list[0],第一位
	orderList, err := Exec_QueryOrderList(et.Ordered, Nodes[0], "", stateDB, kvdb)
	assert.Equal(t, nil, err)

	orderID1 := orderList.List[0].OrderID
	//根据订单号，查询订单详情
	order, err := Exec_QueryOrder(orderID1, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	assert.Equal(t, 10*types.DefaultCoinPrecision, order.GetBalance())

	//根据op查询市场深度
	marketDepthList, err := Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, 10*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	// 吃半单
	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 4, Amount: 5 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyB, stateDB, kvdb, env)
	//根据地址状态查看订单,最新得订单号永远是在list[0],第一位
	orderList, err = Exec_QueryOrderList(et.Completed, Nodes[1], "", stateDB, kvdb)
	assert.Equal(t, nil, err)
	orderID2 := orderList.List[0].OrderID
	//查询订单1详情
	order, err = Exec_QueryOrder(orderID1, stateDB, kvdb)
	assert.Equal(t, nil, err)
	//订单1的状态应该还是ordered
	assert.Equal(t, int32(et.Ordered), order.Status)
	assert.Equal(t, 5*types.DefaultCoinPrecision, order.Balance)

	//order2状态是completed
	order, err = Exec_QueryOrder(orderID2, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)
	//根据op查询市场深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	//市场深度应该改变
	assert.Equal(t, 5*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	//QueryHistoryOrderList
	orderList, err = Exec_QueryHistoryOrder(&et.QueryHistoryOrderList{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, orderID2, orderList.List[0].OrderID)
	//撤回未完成的订单
	Exec_RevokeOrder(t, orderID1, PrivKeyA, stateDB, kvdb, env)
	//根据订单号，查询订单详情
	order, err = Exec_QueryOrder(orderID1, stateDB, kvdb)
	assert.Equal(t, nil, err)
	//订单1的状态应该Revoked
	assert.Equal(t, int32(et.Revoked), order.Status)
	assert.Equal(t, 5*types.DefaultCoinPrecision, order.Balance)

	//根据op查询市场深度
	//查看bty,CCNY买市场深度,查不到买单深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.NotEqual(t, nil, err)
	//根据原有状态去查看买单是否被改变
	//原有ordered状态的数据应该被删除
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[0], "", stateDB, kvdb)
	assert.Equal(t, types.ErrNotFound, err)

	/*
			买卖单价格相同，测试正常撮合流程，查询功能是否可用
			反向测试
			 用例说明：
			   1.先挂数量是10的卖单。
		       2.然后再挂数量是10的卖单
		       3.再挂数量是5的买单
		       4.再挂数量是15的买单
	*/
	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"}, RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 50000000, Amount: 10 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA, stateDB, kvdb, env)
	//根据地址状态查看订单
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[0], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID3 := orderList.List[0].OrderID

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 50000000, Amount: 10 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA, stateDB, kvdb, env)
	//根据地址状态查看订单
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[0], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID4 := orderList.List[0].OrderID

	//根据op查询市场深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Op: et.OpSell}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	//市场深度应该改变
	assert.Equal(t, 20*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 50000000, Amount: 5 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB, stateDB, kvdb, env)
	//根据地址状态查看订单
	orderList, err = Exec_QueryOrderList(et.Completed, Nodes[1], "", stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), orderList.List[1].Status)
	//同价格按先进先出得原则吃单
	//查询订单3详情
	order, err = Exec_QueryOrder(orderID3, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	//订单余额
	assert.Equal(t, 5*types.DefaultCoinPrecision, order.Balance)
	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Price: 50000000, Amount: 15 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB, stateDB, kvdb, env)
	//order3,order4挂单全部被吃完
	order, err = Exec_QueryOrder(orderID3, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)
	order, err = Exec_QueryOrder(orderID4, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)
	//根据op查询市场深度,这时候应该查不到
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "paracross", Symbol: "coins.bty"}, Op: et.OpSell}, stateDB, kvdb)
	assert.Equal(t, types.ErrNotFound, err)

	/*
	 低于市场价的卖单测试 /高于市场价格的买单测试
	 用例说明：
	   1.先挂数量是5,价格为4的买单
	   2.再挂数量是10,价格为3的卖单
	   3.再挂数量是5,价格为4的卖单
	   4.再挂数量是5,价格为5的卖单
	   5.挂数量是15,价格为4.5的买单
	   6.挂单数量是2,价格为1的卖单
	   7.挂单数量是8,价格为1的卖单
	   8.撤回未成交的卖单
	*/

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 400000000, Amount: 5 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyD, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[3], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID6 := orderList.List[0].OrderID

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 300000000, Amount: 10 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyC, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[2], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID7 := orderList.List[0].OrderID

	//此时订单6应该被吃掉
	order, err = Exec_QueryOrder(orderID6, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)

	order, err = Exec_QueryOrder(orderID7, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	//查看账户余额,按被动方买单的价格成交
	acc := accD1.LoadExecAccount(Nodes[3], execAddr)
	assert.Equal(t, 80*types.DefaultCoinPrecision, acc.Balance)

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 400000000, Amount: 5 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyC, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[2], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID8 := orderList.List[0].OrderID
	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 500000000, Amount: 5 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyC, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[2], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID9 := orderList.List[0].OrderID

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 450000000, Amount: 15 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyD, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[3], "", stateDB, kvdb)
	//orderID10 := orderList.List[0].OrderID
	assert.Equal(t, 5*types.DefaultCoinPrecision, orderList.List[0].Balance)
	assert.Nil(t, err)
	//order7和order8价格在吃单范围内
	order, err = Exec_QueryOrder(orderID7, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)
	order, err = Exec_QueryOrder(orderID8, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Completed), order.Status)

	order, err = Exec_QueryOrder(orderID9, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(et.Ordered), order.Status)
	assert.Equal(t, 5*types.DefaultCoinPrecision, order.Balance)
	//余额检查
	acc = accD1.LoadExecAccount(Nodes[3], execAddr)
	//100-5*4-(10-5)*3-5*4-(15-5-5)*4.5 = 22.5
	assert.Equal(t, int64(2250000000), acc.Balance)
	acc = accC.LoadExecAccount(Nodes[2], execAddr)
	assert.Equal(t, 80*types.DefaultCoinPrecision, acc.Balance)

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 100000000, Amount: 2 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyC, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Completed, Nodes[2], "", stateDB, kvdb)
	assert.Nil(t, err)
	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 100000000, Amount: 8 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyC, stateDB, kvdb, env)
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[2], "", stateDB, kvdb)
	assert.Nil(t, err)
	orderID10 := orderList.List[0].OrderID
	assert.Equal(t, int32(et.Ordered), orderList.List[0].Status)
	assert.Equal(t, 5*types.DefaultCoinPrecision, orderList.List[0].Balance)
	//余额检查
	acc = accD1.LoadExecAccount(Nodes[3], execAddr)
	// Nodes[3]为成交  Nodes[2]冻结10
	assert.Equal(t, int64(2250000000), acc.Balance)
	acc = accC.LoadExecAccount(Nodes[2], execAddr)
	assert.Equal(t, 70*types.DefaultCoinPrecision, acc.Balance)
	//orderID9和order10未成交
	Exec_RevokeOrder(t, orderID9, PrivKeyC, stateDB, kvdb, env)
	Exec_RevokeOrder(t, orderID10, PrivKeyC, stateDB, kvdb, env)
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpSell}, stateDB, kvdb)
	assert.NotEqual(t, nil, err)
	acc = accC.LoadExecAccount(Nodes[2], execAddr)
	assert.Equal(t, 80*types.DefaultCoinPrecision, acc.Balance)

	//清理环境,重建数据库
	util.CloseTestDB(dir, stateDB)
	total = 1000 * types.DefaultCoinPrecision
	accountA = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}

	dir, stateDB, kvdb = util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)
	//execAddr := address.ExecAddress(et.ExchangeX)

	accA, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accA1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accA1.SaveExecAccount(execAddr, &accountA)

	accB1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accB1.SaveExecAccount(execAddr, &accountB)

	env = &execEnv{
		10,
		1,
		1539918074,
	}
	/*
	  撮合深度测试：
	  用例说明:
	    1.先挂200单，价格为1数量为1的买单
	    2.然后再挂价格为1,数量为200的卖单
	    3.相同得地址不能交易,不会撮合
	    4.不同的地址没有权限进行订单撤销
	*/

	for i := 0; i < 200; i++ {
		Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 100000000, Amount: 1 * types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyA, stateDB, kvdb, env)
	}

	Exec_LimitOrder(t, &et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: 100000000, Amount: 200 * types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyB, stateDB, kvdb, env)
	if et.MaxMatchCount > 200 {
		return
	}
	orderList, err = Exec_QueryOrderList(et.Ordered, Nodes[1], "", stateDB, kvdb)
	orderID := orderList.List[0].OrderID
	assert.Equal(t, nil, err)
	assert.Equal(t, (200-et.MaxMatchCount)*types.DefaultCoinPrecision, orderList.List[0].Balance)
	//根据op查询市场深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, (200-et.MaxMatchCount)*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpSell}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, (200-et.MaxMatchCount)*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	//根据状态地址查询订单信息
	//分页查询
	var count int
	var primaryKey string
	for {
		orderList, err := Exec_QueryOrderList(et.Completed, Nodes[0], primaryKey, stateDB, kvdb)
		if err != nil {
			break
		}
		count = count + len(orderList.List)
		if orderList.PrimaryKey == "" {
			break
		}
		primaryKey = orderList.PrimaryKey
	}
	assert.Equal(t, et.MaxMatchCount, count)

	//分页查询查看历史订单
	//根据状态地址查询订单信息
	count = 0
	primaryKey = ""
	for {
		query := &et.QueryHistoryOrderList{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, PrimaryKey: primaryKey}
		orderList, err := Exec_QueryHistoryOrder(query, stateDB, kvdb)
		if err != nil {
			break
		}
		count = count + len(orderList.List)
		if orderList.PrimaryKey == "" {
			break
		}
		primaryKey = orderList.PrimaryKey
	}
	assert.Equal(t, et.MaxMatchCount, count)
	//不同的地址没有权限进行订单撤销
	err = Exec_RevokeOrder(t, orderID, PrivKeyA, stateDB, kvdb, env)
	assert.NotEqual(t, nil, err)
	err = Exec_RevokeOrder(t, orderID, PrivKeyB, stateDB, kvdb, env)
	assert.Equal(t, nil, err)

	//清理环境,重建数据库
	util.CloseTestDB(dir, stateDB)
	total = 1000 * types.DefaultCoinPrecision
	accountA = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}

	dir, stateDB, kvdb = util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)
	//execAddr := address.ExecAddress(et.ExchangeX)

	accA, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accA1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accA1.SaveExecAccount(execAddr, &accountA)

	accB1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accB1.SaveExecAccount(execAddr, &accountB)

	env = &execEnv{
		10,
		1,
		1539918074,
	}
	/*
	  //批量测试,同一个区块内出现多笔可以撮合的买卖交易
	  用例说明:
	    1.在同一个区块内,出现如下：
	        10笔买单
	        20笔卖单
	        50笔买单
	        20笔卖单
	        50笔买单
	        100笔卖单

	*/
	acc = accB1.LoadExecAccount(Nodes[1], execAddr)
	assert.Equal(t, 1000*types.DefaultCoinPrecision, acc.Balance)
	acc = accA.LoadExecAccount(Nodes[0], execAddr)
	assert.Equal(t, 1000*types.DefaultCoinPrecision, acc.Balance)
	var txs []*types.Transaction
	for i := 0; i < 10; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 20; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 50; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 20; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}
	for i := 0; i < 50; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 100; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}
	err = Exec_Block(t, stateDB, kvdb, env, txs...)
	assert.Equal(t, nil, err)
	acc = accB1.LoadExecAccount(Nodes[1], execAddr)
	assert.Equal(t, 890*types.DefaultCoinPrecision, acc.Balance)
	acc = accA.LoadExecAccount(Nodes[0], execAddr)
	assert.Equal(t, 860*types.DefaultCoinPrecision, acc.Balance)
	assert.Equal(t, 30*types.DefaultCoinPrecision, acc.Frozen)

	//根据op查询市场深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.NotEqual(t, nil, err)
	//assert.Equal(t, (200-et.MaxMatchCount)*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpSell}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, 30*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	//清理环境,重建数据库
	util.CloseTestDB(dir, stateDB)
	total = 1000 * types.DefaultCoinPrecision
	accountA = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB = types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}

	dir, stateDB, kvdb = util.CreateTestDB()
	defer util.CloseTestDB(dir, stateDB)
	//execAddr := address.ExecAddress(et.ExchangeX)

	accA, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ = account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accA1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accA1.SaveExecAccount(execAddr, &accountA)

	accB1, _ = account.NewAccountDB(cfg, "token", "CCNY", stateDB)
	accB1.SaveExecAccount(execAddr, &accountB)

	env = &execEnv{
		10,
		1,
		1539918074,
	}
	/*
			  //批量测试,同个区块内出现多笔可以撮合的买卖交易
			  用例说明:
			    1.在同一区块内,出现如下：
		            100笔卖单
			        50笔买单
			        20笔卖单
			        100笔买单
	*/
	acc = accB1.LoadExecAccount(Nodes[1], execAddr)
	assert.Equal(t, 1000*types.DefaultCoinPrecision, acc.Balance)
	acc = accA.LoadExecAccount(Nodes[0], execAddr)
	assert.Equal(t, 1000*types.DefaultCoinPrecision, acc.Balance)
	txs = nil
	for i := 0; i < 100; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 50; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}

	for i := 0; i < 20; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpSell}, PrivKeyA)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}
	for i := 0; i < 100; i++ {
		tx, err := CreateLimitOrder(&et.LimitOrder{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
			RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Price: types.DefaultCoinPrecision, Amount: types.DefaultCoinPrecision, Op: et.OpBuy}, PrivKeyB)
		assert.Equal(t, nil, err)
		txs = append(txs, tx)
	}
	err = Exec_Block(t, stateDB, kvdb, env, txs...)
	assert.Equal(t, nil, err)
	acc = accB1.LoadExecAccount(Nodes[1], execAddr)
	assert.Equal(t, 850*types.DefaultCoinPrecision, acc.Balance)
	assert.Equal(t, 30*types.DefaultCoinPrecision, acc.Frozen)
	acc = accA.LoadExecAccount(Nodes[0], execAddr)
	assert.Equal(t, 880*types.DefaultCoinPrecision, acc.Balance)

	//根据op查询市场深度
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpSell}, stateDB, kvdb)
	assert.NotEqual(t, nil, err)
	marketDepthList, err = Exec_QueryMarketDepth(&et.QueryMarketDepth{LeftAsset: &et.Asset{Symbol: "bty", Execer: "coins"},
		RightAsset: &et.Asset{Execer: "token", Symbol: "CCNY"}, Op: et.OpBuy}, stateDB, kvdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, 30*types.DefaultCoinPrecision, marketDepthList.List[0].GetAmount())

	//根据状态地址查询订单信息
	//分页查询
	count = 0
	primaryKey = ""
	for {
		orderList, err := Exec_QueryOrderList(et.Completed, Nodes[1], primaryKey, stateDB, kvdb)
		if err != nil {
			break
		}
		count = count + len(orderList.List)
		if orderList.PrimaryKey == "" {
			break
		}
		primaryKey = orderList.PrimaryKey
	}
	assert.Equal(t, 120, count)

	count = 0
	primaryKey = ""
	for {
		orderList, err := Exec_QueryOrderList(et.Ordered, Nodes[1], primaryKey, stateDB, kvdb)
		if err != nil {
			break
		}
		count = count + len(orderList.List)
		if orderList.PrimaryKey == "" {
			break
		}
		primaryKey = orderList.PrimaryKey
	}
	assert.Equal(t, 30, count)

}

func CreateLimitOrder(limitOrder *et.LimitOrder, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err = ety.Create("LimitOrder", limitOrder)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}
func CreateRevokeOrder(orderID int64, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.ExchangeX)
	tx, err = ety.Create("RevokeOrder", &et.RevokeOrder{OrderID: orderID})
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.ExchangeX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

//模拟区块中交易得执行过程
func Exec_Block(t *testing.T, stateDB db.DB, kvdb db.KVDB, env *execEnv, txs ...*types.Transaction) error {
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := NewExchange()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	e := exec.(*exchange)

	for index, tx := range txs {
		err := e.CheckTx(tx, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}

	}

	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 20
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for index, tx := range txs {
		receipt, err := exec.Exec(tx, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, index)
		if err != nil {
			t.Log(err.Error())
			return err
		}
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
		//save to database
		util.SaveKVList(stateDB, set.KV)
		assert.Equal(t, types.ExecOk, int(receipt.Ty))
	}
	return nil
}
func Exec_LimitOrder(t *testing.T, limitOrder *et.LimitOrder, privKey string, stateDB db.DB, kvdb db.KVDB, env *execEnv) error {
	tx, err := CreateLimitOrder(limitOrder, privKey)
	if err != nil {
		return err
	}
	return Exec_Block(t, stateDB, kvdb, env, tx)
}

func Exec_RevokeOrder(t *testing.T, orderID int64, privKey string, stateDB db.DB, kvdb db.KVDB, env *execEnv) error {
	tx, err := CreateRevokeOrder(orderID, privKey)
	if err != nil {
		return err
	}
	return Exec_Block(t, stateDB, kvdb, env, tx)
}

func Exec_QueryOrderList(status int32, addr string, primaryKey string, stateDB db.KV, kvdb db.KVDB) (*et.OrderList, error) {
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := NewExchange()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据地址状态查看订单
	msg, err := exec.Query(et.FuncNameQueryOrderList, types.Encode(&et.QueryOrderList{Status: status, Address: addr, PrimaryKey: primaryKey}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.OrderList), nil
}
func Exec_QueryOrder(orderID int64, stateDB db.KV, kvdb db.KVDB) (*et.Order, error) {
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := NewExchange()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据orderID查看订单信息
	msg, err := exec.Query(et.FuncNameQueryOrder, types.Encode(&et.QueryOrder{OrderID: orderID}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.Order), err
}

func Exec_QueryMarketDepth(query *et.QueryMarketDepth, stateDB db.KV, kvdb db.KVDB) (*et.MarketDepthList, error) {
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := NewExchange()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据QueryMarketDepth查看市场深度
	msg, err := exec.Query(et.FuncNameQueryMarketDepth, types.Encode(query))
	if err != nil {
		return nil, err
	}
	return msg.(*et.MarketDepthList), err
}

func Exec_QueryHistoryOrder(query *et.QueryHistoryOrderList, stateDB db.KV, kvdb db.KVDB) (*et.OrderList, error) {
	cfg := types.NewChain33Config(et.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := NewExchange()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	//根据QueryMarketDepth查看市场深度
	msg, err := exec.Query(et.FuncNameQueryHistoryOrderList, types.Encode(query))
	return msg.(*et.OrderList), err
}
func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		return tx, err
	}

	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}

	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}

	tx.Sign(int32(signType), privKey)
	return tx, nil
}

func TestCheckPrice(t *testing.T) {
	t.Log(CheckPrice(1e8))
	t.Log(CheckPrice(-1))
	t.Log(CheckPrice(1e17))
	t.Log(CheckPrice(0))
}

func TestRawMeta(t *testing.T) {
	HistoryOrderRow := NewHistoryOrderRow()
	t.Log(HistoryOrderRow.Get("index"))
	MarketDepthRow := NewMarketDepthRow()
	t.Log(MarketDepthRow.Get("price"))
	marketOrderRow := NewOrderRow()
	t.Log(marketOrderRow.Get("orderID"))
}

func TestKV(t *testing.T) {
	a := &types.KeyValue{Key: []byte("1111111"), Value: nil}
	t.Log(a.Key, a.Value)
}

func TestSafeMul(t *testing.T) {
	t.Log(SafeMul(1e8, 1e7, types.DefaultCoinPrecision))
	t.Log(SafeMul(1e10, 1e16, types.DefaultCoinPrecision))
	t.Log(SafeMul(1e7, 1e6, types.DefaultCoinPrecision))
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/33cn/plugin/plugin/dapp/ticket/executor"
	pty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// 回归测试：TicketClose 对重复 ticketId 重复付款。
//
// TicketClose 前半段只做校验并构造 tickets 切片，后半段才逐个 ExecActive 付款并 Save。
// 由于前半段从不写库，同一个 ticketId 在 payload 里出现 N 次时后半段就会读到同一份未变更的
// 票状态，于是【同一张票被支付 N 次，而库里只记一次关闭】：
//
//	for i := 0; i < len(tclose.TicketId); i++ {   // 只校验，不写库
//	    ticket, err := readTicket(action.db, tclose.TicketId[i])
//	    ...
//	    tickets[i] = &DB{*ticket, prevstatus}
//	}
//	for i := 0; i < len(tickets); i++ {           // 每个下标各付一次
//	    receipt1, err := action.coinsAccount.ExecActive(t.ReturnAddress, action.execaddr, retValue)
//	    ...
//	    t.Save(action.db)
//	}
//
// 支付来源是 (ReturnAddress, ticket执行地址) 下的冻结池，而该池同时是该地址【所有】未关闭票的
// 押金。入账方恒为票自身的 ReturnAddress，所以钱进不了攻击者口袋(见
// TestTicketCloseDupPaysOnlyReturnAddress)；真实危害在别处，均由下列用例实测界定：
//   - 提现时间锁按【单张票】校验、押金池按【返还地址】共享，于是一张到期票重复即可解冻
//     同地址未到期票的押金(TestTicketCloseDupBypassesWithdrawLock)；
//   - 已挖矿票关闭时还会从【全链共享】的开发基金池重复解冻 CoinDevFund，
//     与本地址押金无关(TestTicketCloseDupDrainsDevFundPool)；
//   - 共享押金池被抽干后同地址其余票暂时无法关闭，直到挖矿奖励回填
//     (TestTicketCloseDupDrainsSharedPool)。
//
// 修复：分叉 ForkTicketCloseDup 生效后，payload 中重复的 ticketId 直接拒绝，pre-fork 行为不变
// (保证历史区块可原样重放)。
const (
	dupClosePrice = 100
	// dupCloseTicketID 与 dupCloseTicketID2 是两张不同的票，用于批量关闭/非相邻重复的场景
	dupCloseTicketID  = "dup-close-ticket-0000000000"
	dupCloseTicketID2 = "dup-close-ticket-0000000001"
)

// dupCloseTestCfg 本文件专用的隔离配置。
//
// 不能复用 mock33 的配置：那是 TestMain 里启动、仍在后台挖矿的节点正在读的对象，
// 在测试里对它 SetDappFork 会与节点的 IsFork 读取竞争(forks map 并发读写，-race 可复现)。
// 这里在包初始化期(节点启动前)用同一份 testdata 配置构造独立实例，之后只由本文件的测试读写。
var dupCloseTestCfg = func() *types.Chain33Config {
	data, err := os.ReadFile(filepath.Join("testdata", "chain33.cfg.toml"))
	if err != nil {
		panic(err)
	}
	return types.NewChain33Config(string(data))
}()

// dupCloseEnv 一套隔离的执行器环境：独立 memdb、预置的票与冻结池、待执行的 tclose 交易
type dupCloseEnv struct {
	driver    dapp.Driver
	kvdb      dbm.KV
	cfg       *types.Chain33Config
	execAddr  string
	addr      string
	tx        *types.Transaction
	blockTime int64
	height    int64
}

// newDupCloseEnv 签名者自持票：矿工地址与返还地址都是签名者自己
func newDupCloseEnv(t *testing.T, ids []string, pool int64) *dupCloseEnv {
	priv, err := FromPrivkey(PrivKeyA)
	require.Nil(t, err)
	return newDupCloseEnvCross(t, ids, pool, priv, "", "")
}

// newDupCloseEnvCross minerAddr/returnAddr 为空时取签名者地址。
// 指定后用于模拟「签名者是绑定矿工、票主(返还地址)另有其人」的场景。
// 冻结池归属于 returnAddr —— 与 ExecActive(t.ReturnAddress, ...) 的入账方一致。
func newDupCloseEnvCross(t *testing.T, ids []string, pool int64, signer crypto.PrivKey, minerAddr, returnAddr string) *dupCloseEnv {
	cfg := dupCloseTestCfg
	memDB, err := dbm.NewGoMemDB("dup-close", "dup-close", 0)
	require.Nil(t, err)
	kvdb := dbm.NewKVDB(memDB)

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	blockTime := int64(1539918074)
	tx := createCloseTx(t, cfg, ids, signer)
	// 默认用交易自身的 From 作为票的矿工/返还地址，保证能通过 action.fromaddr 校验
	if minerAddr == "" {
		minerAddr = tx.From()
	}
	if returnAddr == "" {
		returnAddr = tx.From()
	}
	env := &dupCloseEnv{
		cfg:       cfg,
		kvdb:      kvdb,
		execAddr:  dapp.ExecAddress("ticket"),
		addr:      returnAddr,
		tx:        tx,
		blockTime: blockTime,
		height:    1000,
	}

	driver, err := dapp.LoadDriver("ticket", -1)
	require.Nil(t, err)
	driver.SetAPI(api)
	driver.SetEnv(1000, blockTime, 0)
	driver.SetStateDB(kvdb)
	driver.SetLocalDB(kvdb)
	env.driver = driver

	// 把 payload 里出现过的每张票都建成已开票、可关闭的普通票(CreateTime 取 0，远超 withdrawTime)
	for _, id := range uniqueIDs(ids) {
		ticket := &pty.Ticket{
			TicketId:      id,
			MinerAddress:  minerAddr,
			ReturnAddress: returnAddr,
			CreateTime:    0,
			Status:        pty.TicketOpened,
			Price:         dupClosePrice,
		}
		require.Nil(t, kvdb.Set(executor.Key(id), types.Encode(ticket)))
	}

	// 冻结池：该返还地址所有未关闭票的押金之和
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(kvdb)
	acc.SaveExecAccount(env.execAddr, &types.Account{Addr: returnAddr, Balance: 0, Frozen: pool})
	return env
}

// loadAccount 读取 (返还地址, ticket执行地址) 下的子账户
func (env *dupCloseEnv) loadAccount() *types.Account {
	return env.accountOf(env.addr)
}

// accountOf 读取任意地址的 ticket 子账户
func (env *dupCloseEnv) accountOf(addr string) *types.Account {
	acc := account.NewCoinsAccount(env.cfg)
	acc.SetDB(env.kvdb)
	return acc.LoadExecAccount(addr, env.execAddr)
}

// ticketStatus 读取指定票在库里落定的状态
func (env *dupCloseEnv) ticketStatus(t *testing.T, id string) int32 {
	ticket, err := readTicketFromDB(env.kvdb, id)
	require.Nil(t, err)
	return ticket.Status
}

func createCloseTx(t *testing.T, cfg *types.Chain33Config, ids []string, priv crypto.PrivKey) *types.Transaction {
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create("Tclose", &pty.TicketClose{TicketId: ids})
	require.Nil(t, err)
	tx, err = types.FormatTx(cfg, "ticket", tx)
	require.Nil(t, err)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

// TestTicketCloseDupTicketIDPreFork 记录漏洞本身：分叉前同一张票被重复支付，
// 库里只记一次关闭(证明重复不是幂等的)。
func TestTicketCloseDupTicketIDPreFork(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	// 池中放两张票的押金，重复的 id 每次都能从池里支付成功
	env := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID}, 2*dupClosePrice)
	receipt, err := env.driver.Exec(env.tx, 2)
	assert.Nil(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	acc := env.loadAccount()
	// 一张票付了两次，冻结池被掏空
	assert.Equal(t, int64(2*dupClosePrice), acc.Balance)
	assert.Equal(t, int64(0), acc.Frozen)
	assert.Equal(t, int32(pty.TicketClosed), env.ticketStatus(t, dupCloseTicketID))
}

// TestTicketCloseDupTicketIDPostFork 校验修复：分叉后重复 ticketId 被拒绝，且不产生任何支付或落库。
func TestTicketCloseDupTicketIDPostFork(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, 0)

	env := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID}, 2*dupClosePrice)
	_, err := env.driver.Exec(env.tx, 2)
	assert.Equal(t, pty.ErrDupTicketID, err)

	// 拒绝发生在任何付款之前：余额与冻结池原样保留，票仍是可关闭状态
	acc := env.loadAccount()
	assert.Equal(t, int64(0), acc.Balance)
	assert.Equal(t, int64(2*dupClosePrice), acc.Frozen)
	assert.Equal(t, int32(pty.TicketOpened), env.ticketStatus(t, dupCloseTicketID))

	// 同一笔交易里隔开的重复(非相邻)同样要拒绝
	env2 := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID2, dupCloseTicketID}, 3*dupClosePrice)
	_, err = env2.driver.Exec(env2.tx, 2)
	assert.Equal(t, pty.ErrDupTicketID, err)
	assert.Equal(t, int32(pty.TicketOpened), env2.ticketStatus(t, dupCloseTicketID2))
}

// TestTicketCloseNormalPostFork 正常路径不受影响：单张关闭与多张不同票的批量关闭照常付款。
func TestTicketCloseNormalPostFork(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, 0)

	// 单张
	env := newDupCloseEnv(t, []string{dupCloseTicketID}, dupClosePrice)
	receipt, err := env.driver.Exec(env.tx, 2)
	assert.Nil(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	acc := env.loadAccount()
	assert.Equal(t, int64(dupClosePrice), acc.Balance)
	assert.Equal(t, int64(0), acc.Frozen)
	assert.Equal(t, int32(pty.TicketClosed), env.ticketStatus(t, dupCloseTicketID))

	// 多张不同的票：每张各付一次，总额正确
	env2 := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID2}, 2*dupClosePrice)
	receipt, err = env2.driver.Exec(env2.tx, 2)
	assert.Nil(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	acc2 := env2.loadAccount()
	assert.Equal(t, int64(2*dupClosePrice), acc2.Balance)
	assert.Equal(t, int64(0), acc2.Frozen)
	assert.Equal(t, int32(pty.TicketClosed), env2.ticketStatus(t, dupCloseTicketID))
	assert.Equal(t, int32(pty.TicketClosed), env2.ticketStatus(t, dupCloseTicketID2))
}

// TestTicketForkConfigDeclared 新增 dapp fork 必须同时补齐仓库里所有 [fork.sub.ticket] 配置。
// chain33 解析配置时会逐个核对代码里注册的 fork 与 toml 中的条目，缺一个就直接 panic：
//
//	exec ticket name ForkTicketCloseDup not config in config file
//
// 节点因此无法启动(DisableForkCheck=true 的配置除外，但那类配置不覆盖发布用的链配置)。
func TestTicketForkConfigDeclared(t *testing.T) {
	repoRoot := filepath.Join("..", "..", "..", "..")
	checkConfig := func(rel string) {
		data, err := os.ReadFile(filepath.Join(repoRoot, rel))
		require.Nil(t, err, "read %s", rel)
		text := string(data)
		require.Contains(t, text, "[fork.sub.ticket]", "%s no longer holds a ticket fork section", rel)
		assert.Contains(t, text, pty.ForkTicketCloseDup,
			"config %s has a [fork.sub.ticket] section but no %s entry", rel, pty.ForkTicketCloseDup)
	}

	// 随仓库发布的配置(.toml/.readme)：缺条目会让节点在 initForkConfig 处直接 panic
	var configs []string
	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if ext := filepath.Ext(path); ext != ".toml" && ext != ".readme" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if !strings.Contains(string(data), "[fork.sub.ticket]") {
			return nil
		}
		rel, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		configs = append(configs, rel)
		return nil
	})
	require.Nil(t, err)
	// 防止断言被改名/搬走后空转
	require.NotEmpty(t, configs, "no [fork.sub.ticket] config found, nothing was checked")
	for _, rel := range configs {
		checkConfig(rel)
	}

	// 单测里内联的链配置同样会走 initForkConfig，一并核对
	checkConfig(filepath.Join("plugin", "dapp", "ticket", "rpc", "rpc_test.go"))
}

// TestTicketCloseDupPaysOnlyReturnAddress 界定 H-1 的影响边界：重复付款的入账方恒为票自身的
// ReturnAddress，攻击者无法把多付的钱转给自己或任意第三方。
//
// TicketClose 的 fromaddr 校验只放行两类票：自己是 ReturnAddress 的票，或自己是其绑定矿工的票。
// 两类情况下 ExecActive 的入账地址都是 t.ReturnAddress：
//   - 攻击者就是 ReturnAddress：多取的是自己池里的钱，总额仍以自己存入的押金为上限，没有净收益；
//   - 攻击者只是绑定矿工(本用例)：钱全部进票主自己的子账户，攻击者一分不得。
func TestTicketCloseDupPaysOnlyReturnAddress(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	// 分叉前，让重复付款真正发生，以便观察资金流向
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	victim := string(Nodes[0])   //票主：持有冻结池
	attacker := string(Nodes[1]) //绑定矿工：能发起 tclose，但不持有资产
	attackerPriv, err := FromPrivkey(PrivKeyB)
	require.Nil(t, err)

	// 票主池中有 3 张票的押金(来自他其他未关闭的票)
	env := newDupCloseEnvCross(t, []string{dupCloseTicketID, dupCloseTicketID, dupCloseTicketID},
		3*dupClosePrice, attackerPriv, attacker, victim)

	receipt, err := env.driver.Exec(env.tx, 2)
	require.Nil(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	// 票主：池被多付掏空，钱全部落在票主自己的可用余额里
	victimAcc := env.accountOf(victim)
	assert.Equal(t, int64(3*dupClosePrice), victimAcc.Balance, "多付的钱只进票主自己的子账户")
	assert.Equal(t, int64(0), victimAcc.Frozen)

	// 攻击者：分文未得
	attackerAcc := env.accountOf(attacker)
	assert.Equal(t, int64(0), attackerAcc.Balance, "攻击者不应有任何入账")
	assert.Equal(t, int64(0), attackerAcc.Frozen)
}

// TestTicketCloseDupExtractionCappedByPool 重复付款的提取额以池内余额为上限：
// ExecActive 在 Frozen < amount 时返回 ErrNoBalance，因此无法凭重复取出超过该地址已存入的押金。
// 注意：本用例直接调用 driver.Exec，失败的交易不会回滚已写入的 kv，
// 故只断言错误本身，不断言失败后的状态。
func TestTicketCloseDupExtractionCappedByPool(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	// 池里只有 2 张票的押金，却重复 5 次：第 3 次付款必然失败
	ids := []string{dupCloseTicketID, dupCloseTicketID, dupCloseTicketID, dupCloseTicketID, dupCloseTicketID}
	env := newDupCloseEnv(t, ids, 2*dupClosePrice)

	_, err := env.driver.Exec(env.tx, 2)
	assert.Equal(t, types.ErrNoBalance, err, "超过池余额的重复付款必须失败，不能凭空多付")
}

// TestTicketCloseDupBypassesWithdrawLock 实测本漏洞真正的协议违背点：
// 提现时间锁(ticketWithdrawTime)是按【单张票】校验的，而押金池是按【返还地址】共享的。
// 因此只需要一张已到期的票，靠重复它就能把同一地址下【尚未到期】的票的押金一并解冻出来 ——
// 即绕过了时间锁对整池资金的约束。
func TestTicketCloseDupBypassesWithdrawLock(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	// 池 = A(已到期) + B(未到期) 两张票的押金
	env := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID}, 2*dupClosePrice)
	withdrawTime := pty.GetTicketMinerParam(cfg, env.height).TicketWithdrawTime
	require.Greater(t, withdrawTime, int64(0), "配置里 ticketWithdrawTime 必须为正才有锁可绕")

	// B 尚未到期：CreateTime 取当前区块时间
	ticketB := &pty.Ticket{
		TicketId:      dupCloseTicketID2,
		MinerAddress:  env.addr,
		ReturnAddress: env.addr,
		CreateTime:    env.blockTime,
		Status:        pty.TicketOpened,
		Price:         dupClosePrice,
	}
	require.Nil(t, env.kvdb.Set(executor.Key(dupCloseTicketID2), types.Encode(ticketB)))

	// 前置对照：B 自己单独关闭必须被时间锁拒绝
	_, err := env.driver.Exec(createCloseTx(t, cfg, []string{dupCloseTicketID2}, mustPriv(t, PrivKeyA)), 2)
	assert.Equal(t, pty.ErrTime, err, "未到期的票不可单独关闭")

	// 攻击：只重复已到期的 A，即可把整个池(含 B 的押金)解冻
	_, err = env.driver.Exec(env.tx, 2)
	assert.Nil(t, err, "重复已到期票关闭成功")
	acc := env.loadAccount()
	assert.Equal(t, int64(2*dupClosePrice), acc.Balance, "B 的押金被一并解冻")
	assert.Equal(t, int64(0), acc.Frozen)
	assert.Equal(t, int32(pty.TicketOpened), env.ticketStatus(t, dupCloseTicketID2), "B 本身从未到期")
}

// TestTicketCloseDupDrainsDevFundPool 实测本漏洞影响面【超出攻击者自己】的那条路径：
// 关闭一张已挖矿(prevstatus==2)的票时，除了付给票主，还会从【全链共享的】开发基金池
// (chain33Cfg.GetFundAddr(), ticket 执行地址) 解冻一份 CoinDevFund —— 且这一步在付款循环内，
// 因此重复 N 次就是重复解冻 N 份，消耗的是所有人的钱，与本地址押金多少无关。
// (该分支在 ForkTicketFundAddrV1 之前生效，见 testdata 中 ForkTicketFundAddrV1=-1)
func TestTicketCloseDupDrainsDevFundPool(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	devFund := pty.GetTicketMinerParam(cfg, 1000).CoinDevFund
	require.Greater(t, devFund, int64(0), "coinDevFund 必须为正")
	fundAddr := cfg.GetFundAddr()
	require.NotEmpty(t, fundAddr)

	// 一张已挖矿的票，重复关闭 3 次
	const repeats = 3
	ids := []string{dupCloseTicketID, dupCloseTicketID, dupCloseTicketID}
	env := newDupCloseEnv(t, ids, repeats*dupClosePrice)
	mined := &pty.Ticket{
		TicketId:      dupCloseTicketID,
		MinerAddress:  env.addr,
		ReturnAddress: env.addr,
		CreateTime:    0,
		MinerTime:     0,
		Status:        pty.TicketMined,
		Price:         dupClosePrice,
		MinerValue:    0,
	}
	require.Nil(t, env.kvdb.Set(executor.Key(dupCloseTicketID), types.Encode(mined)))

	// 开发基金池里是别人的钱：给足 10 份，观察被重复消耗掉几份
	const fundPool = 10
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(env.kvdb)
	acc.SaveExecAccount(env.execAddr, &types.Account{Addr: fundAddr, Balance: 0, Frozen: fundPool * devFund})
	require.Equal(t, fundPool*devFund, acc.LoadExecAccount(fundAddr, env.execAddr).Frozen)

	_, err := env.driver.Exec(env.tx, 2)
	require.Nil(t, err, "重复关闭已挖矿票应当成功")

	// 循环里每个下标各解冻一份：本该只解冻 1 份，实际解冻了 repeats 份，
	// 多出的部分直接从全链共享的基金池里拿走 —— 与本地址押金无关，影响的不是攻击者自己
	after := acc.LoadExecAccount(fundAddr, env.execAddr)
	assert.Equal(t, (fundPool-repeats)*devFund, after.Frozen, "基金池被重复消耗")
	assert.Equal(t, repeats*devFund, after.Balance, "共解冻出 3 份 CoinDevFund(应为 1 份)")
}

// TestTicketCloseDupFundDrainBlocksOtherHolders 实测「消耗全链共享基金池」真正伤到【别人】的机制：
//
// 基金池 = 历次挖矿时 ExecDepositFrozen(GetFundAddr(), execAddr, CoinDevFund) 增发并冻结的币，
// 全链只有这一个(fundKeyAddr 的子账户)。关闭一张【已挖矿】的票时，除了付给票主，还会从这个
// 池解冻一份 CoinDevFund —— 而这一步在付款循环内、且失败会让整笔交易返回错误。
//
// 于是攻击者重复关闭自己的票 N 次，就把基金池多抽走 (N-1) 份。池被抽干后，
// **其他持票人**关闭自己的票时会卡在这个基金解冻上：ExecActive 返回 ErrNoBalance，
// 整笔 tclose 失败 —— 他们的押金(TicketPrice)与挖矿奖励(MinerValue)全部取不回来。
// 这些人跟攻击者毫无关系，这是本漏洞唯一能伤害到第三方的路径。
func TestTicketCloseDupFundDrainBlocksOtherHolders(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	devFund := pty.GetTicketMinerParam(cfg, 1000).CoinDevFund
	require.Greater(t, devFund, int64(0))
	fundAddr := cfg.GetFundAddr()
	require.NotEmpty(t, fundAddr)

	// 攻击者(PrivKeyA)：一张已挖矿的票，重复关闭 2 次
	const repeats = 2
	env := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID}, repeats*dupClosePrice)
	require.Nil(t, env.kvdb.Set(executor.Key(dupCloseTicketID), types.Encode(&pty.Ticket{
		TicketId: dupCloseTicketID, MinerAddress: env.addr, ReturnAddress: env.addr,
		Status: pty.TicketMined, Price: dupClosePrice,
	})))

	// 受害者(PrivKeyB)：另一个毫不相干的持票人，同样持有一张已挖矿的票
	victimPriv, err := FromPrivkey(PrivKeyB)
	require.Nil(t, err)
	victimAddr := createCloseTx(t, cfg, []string{dupCloseTicketID}, victimPriv).From()
	require.NotEqual(t, env.addr, victimAddr)
	require.Nil(t, env.kvdb.Set(executor.Key(dupCloseTicketID2), types.Encode(&pty.Ticket{
		TicketId: dupCloseTicketID2, MinerAddress: victimAddr, ReturnAddress: victimAddr,
		Status: pty.TicketMined, Price: dupClosePrice, MinerValue: dupClosePrice,
	})))
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(env.kvdb)
	acc.SaveExecAccount(env.execAddr, &types.Account{Addr: victimAddr, Frozen: 2 * dupClosePrice})

	// 基金池此刻只有 repeats 份 —— 刚好够攻击者全部抽走
	acc.SaveExecAccount(env.execAddr, &types.Account{Addr: fundAddr, Frozen: repeats * devFund})

	// 攻击：重复关闭自己的票，把全链基金池抽干
	_, err = env.driver.Exec(env.tx, 2)
	require.Nil(t, err, "攻击者的重复关闭成功")
	assert.Equal(t, int64(0), acc.LoadExecAccount(fundAddr, env.execAddr).Frozen, "全链基金池被抽干")

	// 受害：关闭自己的票 —— 基金池已空，整笔交易失败，押金与奖励都取不回
	_, err = env.driver.Exec(createCloseTx(t, cfg, []string{dupCloseTicketID2}, victimPriv), 2)
	assert.Equal(t, types.ErrNoBalance, err, "基金池空的期间，其他人的票关不掉")
	assert.Equal(t, int32(pty.TicketMined), env.ticketStatus(t, dupCloseTicketID2),
		"受害者的票仍卡在已挖矿状态，押金+奖励无法取回")
}

// TestTicketCloseDupDrainsSharedPool 实测 4.2 的危害与其边界：
// 抽干的是【同地址共享】的押金池，受害的其余票在被挖矿奖励重新填满之前无法关闭；
// 但该池可回填，故不是永久锁死。
func TestTicketCloseDupDrainsSharedPool(t *testing.T) {
	cfg := dupCloseTestCfg
	origin := cfg.GetDappFork(pty.TicketX, pty.ForkTicketCloseDup)
	defer cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, origin)
	cfg.SetDappFork(pty.TicketX, pty.ForkTicketCloseDup, types.MaxHeight)

	// 池 = 两张票的押金。攻击只重复关闭 A，受害的是同一地址的另一张票 B。
	// 环境按 A 重复两次的 payload 建立，再补建 B(它不在 payload 里)。
	env := newDupCloseEnv(t, []string{dupCloseTicketID, dupCloseTicketID}, 2*dupClosePrice)
	ticketB := &pty.Ticket{
		TicketId:      dupCloseTicketID2,
		MinerAddress:  env.addr,
		ReturnAddress: env.addr,
		CreateTime:    0,
		Status:        pty.TicketOpened,
		Price:         dupClosePrice,
	}
	require.Nil(t, env.kvdb.Set(executor.Key(dupCloseTicketID2), types.Encode(ticketB)))

	// 攻击：重复关闭 A，把共享池抽干
	_, err := env.driver.Exec(env.tx, 2)
	require.Nil(t, err, "重复关闭 A 应当成功并抽干共享池")
	assert.Equal(t, int64(0), env.loadAccount().Frozen, "共享池已被抽干")
	assert.Equal(t, int32(pty.TicketClosed), env.ticketStatus(t, dupCloseTicketID))

	// 受害：此后 B 无法关闭(池内已不足一张票的押金)
	closeB := createCloseTx(t, cfg, []string{dupCloseTicketID2}, mustPriv(t, PrivKeyA))
	_, err = env.driver.Exec(closeB, 2)
	assert.Equal(t, types.ErrNoBalance, err, "池被抽干后受害票无法关闭")
	assert.Equal(t, int32(pty.TicketOpened), env.ticketStatus(t, dupCloseTicketID2))

	// 恢复路径：挖矿奖励走 ExecDepositFrozen(ReturnAddress, execaddr, MinerValue)，存入的正是
	// 同一个池(见 TicketMiner)，因此池可回填 —— 故受害票是被锁住而非永久锁死
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(env.kvdb)
	_, err = acc.ExecDepositFrozen(env.addr, env.execAddr, dupClosePrice)
	require.Nil(t, err, "挖矿奖励回填共享池")

	_, err = env.driver.Exec(closeB, 2)
	assert.Nil(t, err, "池回填后受害票恢复可关闭")
	assert.Equal(t, int32(pty.TicketClosed), env.ticketStatus(t, dupCloseTicketID2))
}

func mustPriv(t *testing.T, hexKey string) crypto.PrivKey {
	priv, err := FromPrivkey(hexKey)
	require.Nil(t, err)
	return priv
}

// uniqueIDs 去重，用于按 payload 建票(重复的 id 只需建一次)
func uniqueIDs(ids []string) []string {
	seen := make(map[string]struct{}, len(ids))
	var out []string
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// readTicketFromDB 按票 id 读库，与执行器内部 readTicket 一致
func readTicketFromDB(kvdb dbm.KV, id string) (*pty.Ticket, error) {
	data, err := kvdb.Get(executor.Key(id))
	if err != nil {
		return nil, err
	}
	var ticket pty.Ticket
	err = types.Decode(data, &ticket)
	if err != nil {
		return nil, err
	}
	return &ticket, nil
}

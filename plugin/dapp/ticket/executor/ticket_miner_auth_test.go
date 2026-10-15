// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor_test

import (
	"fmt"
	"os"
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

// 审计项 H-2 的回归用例：TicketOpen 必须校验 MinerAddress 归属。
//
// 票据的本地索引 LODB-ticket-tl:<MinerAddress>:<status> 与共识选票都用 MinerAddress
// 来标识归属，而 MinerAddress 原本没有任何归属校验：任何签名者只要把 ReturnAddress
// 填成自己，就能把票据挂到任意矿工名下，让该矿工节点把它当自己的票挖出来，出块奖励
// 却增发给票据的 ReturnAddress（出票人）。修复后（ForkTicketMinerAuth 生效）：
//   - MinerAddress 必须是合法地址；
//   - MinerAddress 只能是签名者自己，或 ReturnAddress 通过 Tbind 明确授权的矿工；
//   - 同时修掉 Count > len(PubHashes) 的越界 panic。
const (
	// minerAuthHeight 高于 testdata 的 ForkTicketId(1600000)，ForkTicketMinerAuth 默认 0 已生效
	minerAuthHeight = int64(1700000)
	minerAuthTime   = int64(1539918074)
	// minerAuthLateFork 用于覆盖“分叉尚未生效”的旧路径
	minerAuthLateFork = int64(2000000)
)

type authEnv struct {
	driver dapp.Driver
	db     dbm.KVDB
	close  func()
}

func newAuthEnv(t *testing.T, cfg *types.Chain33Config, height, blockTime int64) *authEnv {
	dir, ldb, kvdb := util.CreateTestDB()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	driver, err := dapp.LoadDriver("ticket", height)
	require.NoError(t, err)
	driver.SetAPI(api)
	driver.SetEnv(height, blockTime, uint64(blockTime))
	driver.SetStateDB(kvdb)
	driver.SetLocalDB(kvdb)
	return &authEnv{driver: driver, db: kvdb, close: func() { util.CloseTestDB(dir, ldb) }}
}

// fundTicketExec 把票面资金放进地址在 ticket 执行器下的子账户（等价于链上 coins->ticket）
func (e *authEnv) fundTicketExec(t *testing.T, cfg *types.Chain33Config, addr string, amount int64) {
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(e.db)
	acc.SaveExecAccount(address.ExecAddress(pty.TicketX), &types.Account{Addr: addr, Balance: amount, Frozen: 0})
}

// execBalance 读回地址在 ticket 执行器下的子账户余额
func (e *authEnv) execBalance(cfg *types.Chain33Config, addr string) *types.Account {
	acc := account.NewCoinsAccount(cfg)
	acc.SetDB(e.db)
	return acc.LoadExecAccount(addr, address.ExecAddress(pty.TicketX))
}

func authAddr(priv crypto.PrivKey) string {
	return address.PubKeyToAddr(address.DefaultID, priv.PubKey().Bytes())
}

func authTx(t *testing.T, cfg *types.Chain33Config, action string, payload proto.Message, priv crypto.PrivKey) *types.Transaction {
	ety := types.LoadExecutorType("ticket")
	tx, err := ety.Create(action, payload)
	require.NoError(t, err)
	tx, err = types.FormatTx(cfg, "ticket", tx)
	require.NoError(t, err)
	tx.Sign(types.SECP256K1, priv)
	return tx
}

// authPubHashes 生成 count 个与指定私钥绑定的 pubHash（开票人自己的 privHash 哈希）
func authPubHashes(priv crypto.PrivKey, count int32, seed int64) [][]byte {
	hashes := make([][]byte, count)
	for i := 0; i < int(count); i++ {
		privHash := common.Sha256([]byte(fmt.Sprintf("%x:%d:%d", priv.Bytes(), i, seed)))
		hashes[i] = common.Sha256(privHash)
	}
	return hashes
}

// authNewTickets 从回执里取出新建票据的 ticketId 与归属地址
func authNewTickets(t *testing.T, receipt *types.Receipt) (ids, addrs []string) {
	for _, l := range receipt.Logs {
		if l.Ty != pty.TyLogNewTicket {
			continue
		}
		var rt pty.ReceiptTicket
		require.NoError(t, types.Decode(l.Log, &rt))
		ids = append(ids, rt.TicketId)
		addrs = append(addrs, rt.Addr)
	}
	return ids, addrs
}

// authCfgForkAt 基于 testdata 配置，把 ForkTicketMinerAuth 改成指定高度
func authCfgForkAt(t *testing.T, height int64) *types.Chain33Config {
	data, err := os.ReadFile("testdata/chain33.cfg.toml")
	require.NoError(t, err)
	cfgStr := strings.Replace(string(data), "ForkTicketMinerAuth = 0",
		fmt.Sprintf("ForkTicketMinerAuth = %d", height), 1)
	require.NotEqual(t, string(data), cfgStr, "testdata 应显式配置 ForkTicketMinerAuth")
	cfg := types.NewChain33Config(cfgStr)
	require.True(t, cfg.IsDappFork(height, pty.TicketX, pty.ForkTicketMinerAuth))
	require.False(t, cfg.IsDappFork(height-1, pty.TicketX, pty.ForkTicketMinerAuth))
	return cfg
}

// Test_TicketOpen_RejectsForeignMinerAddress 修复后第三方不能把票据挂到别的矿工名下
func Test_TicketOpen_RejectsForeignMinerAddress(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	require.True(t, cfg.IsDappFork(minerAuthHeight, pty.TicketX, pty.ForkTicketMinerAuth),
		"ForkTicketMinerAuth 默认注册为 0，应当已生效")

	attacker, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)
	victim, err := FromPrivkey(PrivKeyB)
	require.NoError(t, err)
	attackerAddr, victimAddr := authAddr(attacker), authAddr(victim)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, attackerAddr, 4*price)

	// 攻击者用自己的地址做 ReturnAddress，把 MinerAddress 填成受害矿工
	topen := &pty.TicketOpen{
		MinerAddress:  victimAddr,
		ReturnAddress: attackerAddr,
		Count:         1,
		RandSeed:      1,
		PubHashes:     authPubHashes(attacker, 1, 1),
	}
	tx := authTx(t, cfg, "Topen", topen, attacker)
	require.Equal(t, attackerAddr, tx.From())

	receipt, err := env.driver.Exec(tx, 0)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrMinerAddr, err)
	// 校验发生在冻结之前：攻击者的票面资金没有被扣走
	assert.Equal(t, 4*price, env.execBalance(cfg, attackerAddr).GetBalance())
}

// Test_TicketOpen_AllowsOwnMinerAddress 自挖路径（MinerAddress = ReturnAddress = 签名者）不受影响
func Test_TicketOpen_AllowsOwnMinerAddress(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	self, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)
	selfAddr := authAddr(self)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, selfAddr, 4*price)

	topen := &pty.TicketOpen{
		MinerAddress:  selfAddr,
		ReturnAddress: selfAddr,
		Count:         2,
		RandSeed:      7,
		PubHashes:     authPubHashes(self, 2, 7),
	}
	tx := authTx(t, cfg, "Topen", topen, self)
	require.Equal(t, selfAddr, tx.From())

	receipt, err := env.driver.Exec(tx, 0)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	ids, addrs := authNewTickets(t, receipt)
	require.Len(t, ids, 2)
	for i, id := range ids {
		assert.Equal(t, selfAddr, addrs[i])
		assert.True(t, strings.HasPrefix(id, selfAddr+":"), "ticketId 前缀应为开票矿工地址: %s", id)
		// 票据确实落到状态库
		var saved pty.Ticket
		var found bool
		for _, kv := range receipt.KV {
			if string(kv.Key) == string(ex.Key(id)) {
				require.NoError(t, types.Decode(kv.Value, &saved))
				found = true
			}
		}
		assert.True(t, found, "票据应写入状态库: %s", id)
		assert.Equal(t, selfAddr, saved.GetMinerAddress())
	}
}

// Test_TicketOpen_AllowsBoundMinerAddress ReturnAddress 明确绑定（Tbind）的矿工仍可开票/被代开票
func Test_TicketOpen_AllowsBoundMinerAddress(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	owner, err := FromPrivkey(PrivKeyA) // 票主：ReturnAddress，也是签名者
	require.NoError(t, err)
	miner, err := FromPrivkey(PrivKeyB) // 被授权的矿工
	require.NoError(t, err)
	ownerAddr, minerAddr := authAddr(owner), authAddr(miner)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, ownerAddr, 8*price)

	// 票主签名 Tbind：授权 miner 挖自己的票
	bindTx := authTx(t, cfg, "Tbind", &pty.TicketBind{MinerAddress: minerAddr, ReturnAddress: ownerAddr}, owner)
	bindReceipt, err := env.driver.Exec(bindTx, 0)
	require.NoError(t, err)
	require.NotNil(t, bindReceipt)
	require.Equal(t, int32(types.ExecOk), bindReceipt.Ty)

	// 1) 票主自己开票，MinerAddress 填被授权的矿工
	topen := &pty.TicketOpen{
		MinerAddress:  minerAddr,
		ReturnAddress: ownerAddr,
		Count:         1,
		RandSeed:      11,
		PubHashes:     authPubHashes(miner, 1, 11),
	}
	receipt, err := env.driver.Exec(authTx(t, cfg, "Topen", topen, owner), 0)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)
	ids, addrs := authNewTickets(t, receipt)
	require.Len(t, ids, 1)
	assert.Equal(t, minerAddr, addrs[0])

	// 2) 被授权矿工代开票（fromaddr = miner != ReturnAddress）
	topen2 := &pty.TicketOpen{
		MinerAddress:  minerAddr,
		ReturnAddress: ownerAddr,
		Count:         1,
		RandSeed:      12,
		PubHashes:     authPubHashes(miner, 1, 12),
	}
	receipt2, err := env.driver.Exec(authTx(t, cfg, "Topen", topen2, miner), 0)
	require.NoError(t, err)
	require.NotNil(t, receipt2)
	assert.Equal(t, int32(types.ExecOk), receipt2.Ty)
}

// Test_TicketOpen_RejectsInvalidMinerAddress MinerAddress 必须是合法地址
func Test_TicketOpen_RejectsInvalidMinerAddress(t *testing.T) {
	cfg := mock33.GetAPI().GetConfig()
	self, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)
	selfAddr := authAddr(self)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, selfAddr, 4*price)

	topen := &pty.TicketOpen{
		MinerAddress:  "not-a-chain-address",
		ReturnAddress: selfAddr,
		Count:         1,
		RandSeed:      3,
		PubHashes:     authPubHashes(self, 1, 3),
	}
	receipt, err := env.driver.Exec(authTx(t, cfg, "Topen", topen, self), 0)
	assert.Nil(t, receipt)
	require.Error(t, err)
	assert.NotEqual(t, pty.ErrMinerAddr, err, "应来自地址合法性校验")
	assert.NotEqual(t, pty.ErrMinerNotPermit, err)
}

// Test_TicketOpen_RejectsPubHashesCountMismatch 修复前 Count > len(PubHashes) 会越界 panic
func Test_TicketOpen_RejectsPubHashesCountMismatch(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("TicketOpen 不应 panic: %v", r)
		}
	}()

	cfg := mock33.GetAPI().GetConfig()
	self, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)
	selfAddr := authAddr(self)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, selfAddr, 4*price)

	topen := &pty.TicketOpen{
		MinerAddress:  selfAddr,
		ReturnAddress: selfAddr,
		Count:         2, // PubHashes 只给 1 个
		RandSeed:      5,
		PubHashes:     authPubHashes(self, 1, 5),
	}
	receipt, err := env.driver.Exec(authTx(t, cfg, "Topen", topen, self), 0)
	assert.Nil(t, receipt)
	assert.Equal(t, pty.ErrOpenTicketPubHash, err)
}

// Test_TicketOpen_LegacyPathBeforeFork 分叉高度之前仍按旧逻辑放行：
// 说明 ForkTicketMinerAuth 必须配置在安全的未来高度，升级窗口内不要撤下旧规则。
func Test_TicketOpen_LegacyPathBeforeFork(t *testing.T) {
	cfg := authCfgForkAt(t, minerAuthLateFork)
	require.False(t, cfg.IsDappFork(minerAuthHeight, pty.TicketX, pty.ForkTicketMinerAuth))

	attacker, err := FromPrivkey(PrivKeyA)
	require.NoError(t, err)
	victim, err := FromPrivkey(PrivKeyB)
	require.NoError(t, err)
	attackerAddr, victimAddr := authAddr(attacker), authAddr(victim)

	env := newAuthEnv(t, cfg, minerAuthHeight, minerAuthTime)
	defer env.close()

	price := pty.GetTicketMinerParam(cfg, minerAuthHeight).TicketPrice
	env.fundTicketExec(t, cfg, attackerAddr, 4*price)

	topen := &pty.TicketOpen{
		MinerAddress:  victimAddr,
		ReturnAddress: attackerAddr,
		Count:         1,
		RandSeed:      1,
		PubHashes:     authPubHashes(attacker, 1, 1),
	}
	receipt, err := env.driver.Exec(authTx(t, cfg, "Topen", topen, attacker), 0)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	assert.Equal(t, int32(types.ExecOk), receipt.Ty)

	ids, addrs := authNewTickets(t, receipt)
	require.Len(t, ids, 1)
	assert.Equal(t, victimAddr, addrs[0], "分叉前仍会把票挂到任意矿工名下（因此分叉高度必须留足升级窗口）")
}

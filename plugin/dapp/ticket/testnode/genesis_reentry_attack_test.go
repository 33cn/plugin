// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testnode

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"

	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// TestGenesisReentryRejectedOnTicketNode 在真实票共识测试节点上验证审计项 H-2 的修复链路，
// 全程走 mempool -> 出块 -> 执行器，而不是直接调 driver.Exec：
//
//  1. 默认配置下 ForkTicketId=0（创世即生效），即报告里的 post-fork 场景；
//  2. 攻击者发一笔 TicketAction{Ty=Genesis} 交易；修复前 Ticket.Exec_Genesis 缺少
//     height != 0 保护（coins/coinsx/token 都有），高度 >0 依然执行成功并造出
//     IsGenesis=true、3 段 ticketId 的“创世票”（MinerAddress=受害矿工、ReturnAddress=攻击者），
//     既绕过 pubHash 校验又绕过成熟期，受害矿工会替攻击者挖票并把奖励增发到攻击者地址；
//  3. 修复后该交易仍然可以进块，但 Exec_Genesis 在派发前被 chain33 框架层拒绝
//     （DriverBase.Exec 的 checkGenesisHeight + 系统分叉 ForkGenesisOnlyAtHeightZero，
//     本仓库配置取 0、默认配置 title=local 亦全为 0），回执为 ExecErr，
//     链上不产生任何票据状态，受害矿工也就无从挖出伪造票。
func TestGenesisReentryRejectedOnTicketNode(t *testing.T) {
	// 关闭手续费：攻击者地址的 plain coins 为 0（资产都在 ticket 子账户里）
	cfgStr := strings.Replace(types.GetDefaultCfgstring(), "minTxFeeRate=100000", "minTxFeeRate=0", 1)
	cfgStr = strings.Replace(cfgStr, `loglevel = "debug"`, `loglevel = "error"`, 1)
	cfg := types.NewChain33Config(cfgStr)
	cfg.GetModuleConfig().Consensus.Name = "ticket"
	cfg.GetModuleConfig().Wallet.MinFee = 0
	cfg.SetMinFee(0)

	mock33 := testnode.NewWithConfig(cfg, nil)
	defer mock33.Close()
	require.NoError(t, mock33.WaitHeightTimeout(3, 120))

	// 0) 环境确认：ForkTicketId 已生效 => 报告里的 post-fork 场景
	height := mock33.GetLastBlock().Height
	require.True(t, cfg.IsDappFork(height, ty.TicketX, "ForkTicketId"),
		"默认配置 ForkTicketId=0，创世即生效，属于 post-fork 场景")
	params := ty.GetTicketMinerParam(cfg, height)
	t.Logf("node height=%d ForkTicketId=%v ticketPrice=%d coinReward=%d",
		height, cfg.IsDappFork(height, ty.TicketX, "ForkTicketId"), params.TicketPrice, params.CoinReward)

	// 角色：攻击者持伪造票的 ReturnAddress；受害矿工是节点钱包持有私钥的地址
	attackerPriv := util.TestPrivkeyList[2]
	attackerAddr := address.PubKeyToAddr(address.DefaultID, attackerPriv.PubKey().Bytes())
	victimPriv := util.TestPrivkeyList[0]
	victimAddr := address.PubKeyToAddr(address.DefaultID, victimPriv.PubKey().Bytes())
	t.Logf("attacker=%s victim=%s", attackerAddr, victimAddr)

	// 1) 攻击者提交伪造“创世票”交易（MinerAddress=受害矿工，ReturnAddress=自己）
	//    守卫在余额校验之前生效，因此这里不依赖攻击者是否有票面资金。
	ety := types.LoadExecutorType("ticket")
	genesisAction, err := ety.Create("Genesis", &ty.TicketGenesis{
		MinerAddress:  victimAddr,
		ReturnAddress: attackerAddr,
		Count:         1,
	})
	require.NoError(t, err)
	genesisTx, err := types.FormatTx(cfg, "ticket", genesisAction)
	require.NoError(t, err)
	genesisTx.Sign(types.SECP256K1, attackerPriv)
	txHash := mock33.SendTx(genesisTx)
	txHashHex := common.ToHex(txHash)

	// 2) 交易上链，但执行失败：回执必须是 ExecErr 而不是 ExecOk
	var detail *types.TransactionDetail
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		detail, err = mock33.GetAPI().QueryTx(&types.ReqHash{Hash: txHash})
		if err == nil && detail != nil && detail.GetReceipt() != nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	require.NotNil(t, detail)
	require.NotNil(t, detail.GetReceipt(), "交易必须上链并返回回执")
	// 失败交易的回执不会是 ExecOk(收费为 0 时通常是 ExecPack)；关键是动作没有执行成功
	assert.NotEqual(t, int32(types.ExecOk), detail.GetReceipt().GetTy(),
		"高度 %d 上的 Exec_Genesis 必须执行失败", detail.GetHeight())
	t.Logf("genesis reentry tx=%s height=%d receiptTy=%d rejected", txHashHex, detail.GetHeight(), detail.GetReceipt().GetTy())

	// 3) 链上不产生伪造创世票（3 段 ticketId 由 MinerAddress + txHash 确定）
	firstID := fmt.Sprintf("%s:%s:%010d", victimAddr, txHashHex, 0)
	require.Equal(t, 2, strings.Count(firstID, ":"), "伪造票的 ticketId 形态: %s", firstID)
	reply, err := mock33.GetAPI().Query(ty.TicketX, "TicketInfos", &ty.TicketInfos{TicketIds: []string{firstID}})
	if err == nil {
		assert.Empty(t, reply.(*ty.ReplyTicketList).GetTickets(), "被拒绝的 Genesis 不得写入任何票据状态")
	}

	// 4) 受害矿工名下也没有这笔交易产生的票据
	listReply, err := mock33.GetAPI().Query(ty.TicketX, "TicketList", &ty.TicketList{Addr: victimAddr, Status: ty.TicketOpened})
	if err == nil {
		assert.Zero(t, countForged(listReply.(*ty.ReplyTicketList).GetTickets(), txHashHex),
			"伪造票不得落到受害矿工名下")
	}

	// 5) 继续出块：受害矿工不会挖出伪造票
	require.NoError(t, mock33.WaitHeightTimeout(height+5, 60))
	mined := minedForged(t, mock33, victimAddr, txHashHex, attackerAddr)
	assert.Zero(t, mined.count, "受害矿工不得挖出伪造创世票")
	t.Logf("after %d more blocks: minedForged=%d (attack blocked)", mock33.GetLastBlock().Height-height, mined.count)
}

type forgedMined struct {
	count  int
	ids    []string
	values []int64
}

func countForged(tickets []*ty.Ticket, txHashHex string) int {
	var n int
	for _, tk := range tickets {
		if strings.Contains(tk.GetTicketId(), txHashHex) {
			n++
		}
	}
	return n
}

func minedForged(t *testing.T, mock33 *testnode.Chain33Mock, victimAddr, txHashHex, attackerAddr string) forgedMined {
	var res forgedMined
	reply, err := mock33.GetAPI().Query(ty.TicketX, "TicketList", &ty.TicketList{Addr: victimAddr, Status: ty.TicketMined})
	if err != nil {
		return res
	}
	for _, tk := range reply.(*ty.ReplyTicketList).GetTickets() {
		if !strings.Contains(tk.GetTicketId(), txHashHex) || tk.GetReturnAddress() != attackerAddr {
			continue
		}
		res.count++
		res.ids = append(res.ids, tk.GetTicketId())
		res.values = append(res.values, tk.GetMinerValue())
	}
	return res
}

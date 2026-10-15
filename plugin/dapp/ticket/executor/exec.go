// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// Exec_Genesis exec genesis
func (t *Ticket) Exec_Genesis(payload *ty.TicketGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
	// 创世动作只允许在高度 0 执行, 与 coins/coinsx/token 对齐:
	// 否则可在任意高度重入 GenesisInit 造出 isGenesis=true 的 3 段 ticketId 票据,
	// 绕过 pubHash 校验/成熟期/提现等待, 把出块奖励记到攻击者 ReturnAddress。
	if t.GetHeight() != 0 {
		tlog.Error("ticket genesis is only allowed at height 0", "height", t.GetHeight())
		return nil, types.ErrReRunGenesis
	}
	if payload.Count <= 0 {
		return nil, ty.ErrTicketCount
	}

	payload.MinerAddress = common.FmtEthAddressWithFork(payload.MinerAddress, t.GetAPI().GetConfig(), t.GetHeight())
	payload.ReturnAddress = common.FmtEthAddressWithFork(payload.ReturnAddress, t.GetAPI().GetConfig(), t.GetHeight())
	actiondb := NewAction(t, tx)
	return actiondb.GenesisInit(payload)
}

// Exec_Topen exec open
func (t *Ticket) Exec_Topen(payload *ty.TicketOpen, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.Count <= 0 {
		tlog.Error("topen ", "value", payload)
		return nil, ty.ErrTicketCount
	}
	payload.MinerAddress = common.FmtEthAddressWithFork(payload.MinerAddress, t.GetAPI().GetConfig(), t.GetHeight())
	payload.ReturnAddress = common.FmtEthAddressWithFork(payload.ReturnAddress, t.GetAPI().GetConfig(), t.GetHeight())
	actiondb := NewAction(t, tx)
	return actiondb.TicketOpen(payload)
}

// Exec_Tbind exec bind
func (t *Ticket) Exec_Tbind(payload *ty.TicketBind, tx *types.Transaction, index int) (*types.Receipt, error) {
	payload.MinerAddress = common.FmtEthAddressWithFork(payload.MinerAddress, t.GetAPI().GetConfig(), t.GetHeight())
	payload.ReturnAddress = common.FmtEthAddressWithFork(payload.ReturnAddress, t.GetAPI().GetConfig(), t.GetHeight())
	actiondb := NewAction(t, tx)
	return actiondb.TicketBind(payload)
}

// Exec_Tclose exec close
func (t *Ticket) Exec_Tclose(payload *ty.TicketClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	payload.MinerAddress = common.FmtEthAddressWithFork(payload.MinerAddress, t.GetAPI().GetConfig(), t.GetHeight())
	actiondb := NewAction(t, tx)
	return actiondb.TicketClose(payload)
}

// Exec_Miner exec miner
func (t *Ticket) Exec_Miner(payload *ty.TicketMiner, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.TicketMiner(payload, index)
}

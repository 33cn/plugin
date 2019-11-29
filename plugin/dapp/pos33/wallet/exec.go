// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// On_ClosePos33Tickets close ticket
func (policy *ticketPolicy) On_ClosePos33Tickets(req *ty.Pos33TicketClose) (types.Message, error) {
	operater := policy.getWalletOperate()
	reply, err := policy.forceClosePos33Ticket(operater.GetBlockHeight()+1, req.MinerAddress)
	if err != nil {
		bizlog.Error("onClosePos33Tickets", "forceClosePos33Ticket error", err.Error())
	} else {
		go func() {
			if len(reply.Hashes) > 0 {
				operater.WaitTxs(reply.Hashes)
				FlushPos33Ticket(policy.getAPI())
			}
		}()
	}
	return reply, err
}

// On_WalletGetPos33Tickets get ticket
func (policy *ticketPolicy) On_WalletGetPos33Tickets(req *types.ReqNil) (types.Message, error) {
	tickets, privs, err := policy.getPos33TicketsByStatus(1)
	tks := &ty.ReplyWalletPos33Tickets{Tickets: tickets, Privkeys: privs}
	return tks, err
}

// On_WalletAutoMiner auto mine
func (policy *ticketPolicy) On_WalletAutoMiner(req *ty.Pos33MinerFlag) (types.Message, error) {
	policy.store.SetAutoMinerFlag(req.Flag)
	policy.setAutoMining(req.Flag)
	FlushPos33Ticket(policy.getAPI())
	return &types.Reply{IsOk: true}, nil
}

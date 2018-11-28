// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

func (t *Ticket) execDelLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	for _, item := range receiptData.Logs {
		//这三个是ticket 的log
		if item.Ty == ty.TyLogNewTicket || item.Ty == ty.TyLogMinerTicket || item.Ty == ty.TyLogCloseTicket {
			var ticketlog ty.ReceiptTicket
			err := types.Decode(item.Log, &ticketlog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.delTicket(&ticketlog)
			dbSet.KV = append(dbSet.KV, kv...)
		} else if item.Ty == ty.TyLogTicketBind {
			var ticketlog ty.ReceiptTicketBind
			err := types.Decode(item.Log, &ticketlog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.delTicketBind(&ticketlog)
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}
	return dbSet, nil
}

// ExecDelLocal_Genesis exec del local genesis
func (t *Ticket) ExecDelLocal_Genesis(payload *ty.TicketGenesis, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.execDelLocal(receiptData)
}

// ExecDelLocal_Topen exec del local open
func (t *Ticket) ExecDelLocal_Topen(payload *ty.TicketOpen, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.execDelLocal(receiptData)
}

// ExecDelLocal_Tbind exec del local bind
func (t *Ticket) ExecDelLocal_Tbind(payload *ty.TicketBind, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.execDelLocal(receiptData)
}

// ExecDelLocal_Tclose exec del local close
func (t *Ticket) ExecDelLocal_Tclose(payload *ty.TicketClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.execDelLocal(receiptData)
}

// ExecDelLocal_Miner exec del local miner
func (t *Ticket) ExecDelLocal_Miner(payload *ty.TicketMiner, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return t.execDelLocal(receiptData)
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

func (t *Pos33Ticket) execLocal(receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	n := 0
	for _, item := range receiptData.Logs {
		//这三个是ticket 的log
		if item.Ty == ty.TyLogNewPos33Ticket || item.Ty == ty.TyLogMinerPos33Ticket || item.Ty == ty.TyLogClosePos33Ticket {
			var ticketlog ty.ReceiptPos33Ticket
			err := types.Decode(item.Log, &ticketlog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.savePos33Ticket(&ticketlog)
			dbSet.KV = append(dbSet.KV, kv...)
		} else if item.Ty == ty.TyLogPos33TicketBind {
			var ticketlog ty.ReceiptPos33TicketBind
			err := types.Decode(item.Log, &ticketlog)
			if err != nil {
				panic(err) //数据错误了，已经被修改了
			}
			kv := t.savePos33TicketBind(&ticketlog)
			dbSet.KV = append(dbSet.KV, kv...)
		}
		// save all ticket count
		if item.Ty == ty.TyLogNewPos33Ticket {
			n++
		} else if item.Ty == ty.TyLogClosePos33Ticket {
			n--
		}
	}
	kv := t.saveAllPos33TicketCount(n)
	dbSet.KV = append(dbSet.KV, kv...)
	return dbSet, nil
}

// ExecLocal_Genesis exec local genesis
func (t *Pos33Ticket) ExecLocal_Genesis(payload *ty.Pos33TicketGenesis, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	tlog.Info("ExecLocal_Genesis", "height", t.GetHeight())
	return t.execLocal(receiptData)
}

// ExecLocal_Topen exec local open
func (t *Pos33Ticket) ExecLocal_Topen(payload *ty.Pos33TicketOpen, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	tlog.Info("ExecLocal_Topne", "height", t.GetHeight())
	return t.execLocal(receiptData)
}

// ExecLocal_Tbind exec local bind
func (t *Pos33Ticket) ExecLocal_Tbind(payload *ty.Pos33TicketBind, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	tlog.Info("ExecLocal_Tbind", "height", t.GetHeight())
	return t.execLocal(receiptData)
}

// ExecLocal_Tclose exec local close
func (t *Pos33Ticket) ExecLocal_Tclose(payload *ty.Pos33TicketClose, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	tlog.Info("ExecLocal_Tclose", "height", t.GetHeight())
	return t.execLocal(receiptData)
}

// ExecLocal_Miner exec local miner
func (t *Pos33Ticket) ExecLocal_Pminer(payload *ty.Pos33Miner, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	tlog.Info("ExecLocal_Pminer", "height", t.GetHeight())
	dbSet, err := t.execLocal(receiptData)
	if err != nil {
		return nil, err
	}
	kv := t.chechAndUpdateTicketCount()
	dbSet.KV = append(dbSet.KV, kv...)
	return dbSet, nil
}

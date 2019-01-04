/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"github.com/33cn/chain33/types"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
)

func (o *oracle) execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	table := oty.NewTable(o.GetLocalDB())
	for _, item := range receipt.Logs {
		if item.Ty >= oty.TyLogEventPublish && item.Ty <= oty.TyLogResultPublish {
			var oraclelog oty.ReceiptOracle
			err := types.Decode(item.Log, &oraclelog)
			if err != nil {
				return nil, err
			}
			err = table.Replace(&oraclelog)
			if err != nil {
				return nil, err
			}
			kvs, err := table.Save()
			if err != nil {
				return nil, err
			}
			set.KV = append(set.KV, kvs...)
		}
	}
	return set, nil
}

func (o *oracle) ExecLocal_EventPublish(payload *oty.EventPublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execLocal(receiptData)
}

func (o *oracle) ExecLocal_EventAbort(payload *oty.EventAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execLocal(receiptData)
}

func (o *oracle) ExecLocal_ResultPrePublish(payload *oty.ResultPrePublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execLocal(receiptData)
}

func (o *oracle) ExecLocal_ResultAbort(payload *oty.ResultAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execLocal(receiptData)
}

func (o *oracle) ExecLocal_ResultPublish(payload *oty.ResultPublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execLocal(receiptData)
}

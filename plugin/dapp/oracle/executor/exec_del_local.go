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

func (o *oracle) execDelLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	table := oty.NewTable(o.GetLocalDB())
	for _, item := range receipt.Logs {
		var oraclelog oty.ReceiptOracle
		err := types.Decode(item.Log, &oraclelog)
		if err != nil {
			return nil, err
		}

		//回滚时如果状态为EventPublished则删除记录，否则回滚至上一状态
		if oraclelog.Status == oty.EventPublished {
			err = table.Del([]byte(oraclelog.EventID))
			if err != nil {
				return nil, err
			}
		} else {
			oraclelog.Status = oraclelog.PreStatus
			err = table.Replace(&oraclelog)
			if err != nil {
				return nil, err
			}
		}

		kvs, err := table.Save()
		if err != nil {
			return nil, err
		}

		set.KV = append(set.KV, kvs...)
	}
	return set, nil
}

func (o *oracle) ExecDelLocal_EventPublish(payload *oty.EventPublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execDelLocal(receiptData)
}

func (o *oracle) ExecDelLocal_EventAbort(payload *oty.EventAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execDelLocal(receiptData)
}

func (o *oracle) ExecDelLocal_ResultPrePublish(payload *oty.ResultPrePublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execDelLocal(receiptData)
}

func (o *oracle) ExecDelLocal_ResultAbort(payload *oty.ResultAbort, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execDelLocal(receiptData)
}

func (o *oracle) ExecDelLocal_ResultPublish(payload *oty.ResultPublish, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return o.execDelLocal(receiptData)
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

func (c *Blackwhite) execLocal(receiptData *types.ReceiptData) ([]*types.KeyValue, error) {
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case gt.TyLogBlackwhiteCreate,
			gt.TyLogBlackwhitePlay,
			gt.TyLogBlackwhiteShow,
			gt.TyLogBlackwhiteTimeout,
			gt.TyLogBlackwhiteDone:
			{
				var receipt gt.ReceiptBlackwhiteStatus
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := c.saveHeightIndex(&receipt)
				set = append(set, kv...)
			}
		case gt.TyLogBlackwhiteLoopInfo:
			{
				var res gt.ReplyLoopResults
				err := types.Decode(log.Log, &res)
				if err != nil {
					return nil, err
				}
				kv := c.saveLoopResult(&res)
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	return set, nil
}

// ExecLocal_Create 执行生成创建游戏产生的本地数据库
func (c *Blackwhite) ExecLocal_Create(payload *gt.BlackwhiteCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecLocal_Play 执行生成参与游戏产生的本地数据库
func (c *Blackwhite) ExecLocal_Play(payload *gt.BlackwhitePlay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecLocal_Show 执行生成出示密钥产生的本地数据库
func (c *Blackwhite) ExecLocal_Show(payload *gt.BlackwhiteShow, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecLocal_TimeoutDone 执行生成超时产生的本地数据库
func (c *Blackwhite) ExecLocal_TimeoutDone(payload *gt.BlackwhiteTimeoutDone, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

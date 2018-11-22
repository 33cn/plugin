// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

func (c *Blackwhite) execDelLocal(receiptData *types.ReceiptData) ([]*types.KeyValue, error) {
	retKV := make([]*types.KeyValue, 0)
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case gt.TyLogBlackwhiteCreate:
			{
				var receipt gt.ReceiptBlackwhiteStatus
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv := c.delHeightIndex(&receipt)
				retKV = append(retKV, kv...)
				break
			}
		case gt.TyLogBlackwhitePlay:
		case gt.TyLogBlackwhiteShow:
		case gt.TyLogBlackwhiteTimeout:
		case gt.TyLogBlackwhiteDone:
			{
				var receipt gt.ReceiptBlackwhiteStatus
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				//状态数据库由于默克尔树特性，之前生成的索引无效，故不需要回滚，只回滚localDB
				kv := c.delHeightIndex(&receipt)
				retKV = append(retKV, kv...)

				kv = c.saveRollHeightIndex(&receipt)
				retKV = append(retKV, kv...)
				break
			}
		case gt.TyLogBlackwhiteLoopInfo:
			{
				var res gt.ReplyLoopResults
				err := types.Decode(log.Log, &res)
				if err != nil {
					return nil, err
				}
				kv := c.delLoopResult(&res)
				retKV = append(retKV, kv...)
			}
		default:
			return nil, types.ErrNotSupport
		}
	}
	return retKV, nil
}

// ExecDelLocal_Create 执行删除创建游戏产生的本地数据库
func (c *Blackwhite) ExecDelLocal_Create(payload *gt.BlackwhiteCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecDelLocal_Play 执行删除参与游戏产生的本地数据库
func (c *Blackwhite) ExecDelLocal_Play(payload *gt.BlackwhitePlay, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecDelLocal_Show 执行删除出示密钥产生的本地数据库
func (c *Blackwhite) ExecDelLocal_Show(payload *gt.BlackwhiteShow, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecDelLocal_TimeoutDone 执行删除超时产生的本地数据库
func (c *Blackwhite) ExecDelLocal_TimeoutDone(payload *gt.BlackwhiteTimeoutDone, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := c.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

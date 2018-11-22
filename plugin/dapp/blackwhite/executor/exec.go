// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

// Exec_Create 创建游戏
func (c *Blackwhite) Exec_Create(payload *gt.BlackwhiteCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(c, tx, int32(index))
	return action.Create(payload)
}

// Exec_Play 参与游戏
func (c *Blackwhite) Exec_Play(payload *gt.BlackwhitePlay, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(c, tx, int32(index))
	return action.Play(payload)
}

// Exec_Show 出示密钥
func (c *Blackwhite) Exec_Show(payload *gt.BlackwhiteShow, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(c, tx, int32(index))
	return action.Show(payload)
}

// Exec_TimeoutDone 超时处理
func (c *Blackwhite) Exec_TimeoutDone(payload *gt.BlackwhiteTimeoutDone, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(c, tx, int32(index))
	return action.TimeoutDone(payload)
}

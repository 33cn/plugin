// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

// Query_GetBlackwhiteRoundInfo 查询游戏信息
func (c *Blackwhite) Query_GetBlackwhiteRoundInfo(in *gt.ReqBlackwhiteRoundInfo) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return c.GetBlackwhiteRoundInfo(in)
}

// Query_GetBlackwhiteByStatusAndAddr 查询符合状态以及地址的游戏信息
func (c *Blackwhite) Query_GetBlackwhiteByStatusAndAddr(in *gt.ReqBlackwhiteRoundList) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return c.GetBwRoundListInfo(in)
}

// Query_GetBlackwhiteloopResult 查询游戏中每轮次的比赛结果
func (c *Blackwhite) Query_GetBlackwhiteloopResult(in *gt.ReqLoopResult) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return c.GetBwRoundLoopResult(in)
}

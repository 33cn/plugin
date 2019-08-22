// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrVotePeriod 非投票期间
	ErrVotePeriod = errors.New("ErrVotePeriod")
	// ErrProposalStatus 状态错误
	ErrProposalStatus = errors.New("ErrProposalStatus")
	// ErrRepeatVoteAddr 重复投票地址
	ErrRepeatVoteAddr = errors.New("ErrRepeatVoteAddr")
	// ErrRevokeProposalPeriod 非取消提案期间
	ErrRevokeProposalPeriod = errors.New("ErrRevokeProposalPeriod")
	// ErrRevokeProposalPower 不能取消
	ErrRevokeProposalPower = errors.New("ErrRevokeProposalPower")
	// ErrTerminatePeriod 不能终止
	ErrTerminatePeriod = errors.New("ErrTerminatePeriod")
	// ErrNoActiveBoard 没有有效董事会
	ErrNoActiveBoard = errors.New("ErrNoActiveBoard")
	// ErrNoAutonomyExec 非Autonomy执行器
	ErrNoAutonomyExec = errors.New("ErrNoAutonomyExec")
	// ErrNoPeriodAmount 当前没有足够额度
	ErrNoPeriodAmount = errors.New("ErrNoPeriodAmount")
	// ErrMinerAddr 无效挖矿地址
	ErrMinerAddr = errors.New("ErrMinerAddr")
	// ErrBindAddr 无效绑定地址
	ErrBindAddr = errors.New("ErrBindAddr")
	// ErrChangeBoardAddr 无效修改董事会成员地址
	ErrChangeBoardAddr = errors.New("ErrChangeBoardAddr")
	// ErrBoardNumber 董事会成员数错误
	ErrBoardNumber = errors.New("ErrBoardNumber")
	// ErrRepeatAddr 重复地址
	ErrRepeatAddr = errors.New("ErrRepeatAddr")
)

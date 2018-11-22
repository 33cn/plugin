// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "errors"

var (
	// ErrIncorrectStatus 所处游戏状态不正确
	ErrIncorrectStatus = errors.New("ErrIncorrectStatus")
	// ErrRepeatPlayerAddr 重复玩家
	ErrRepeatPlayerAddr = errors.New("ErrRepeatPlayerAddress")
	// ErrNoTimeoutDone 还未超时
	ErrNoTimeoutDone = errors.New("ErrNoTimeoutDone")
	// ErrNoExistAddr 不存在地址，未参与游戏
	ErrNoExistAddr = errors.New("ErrNoExistAddress")
	// ErrNoLoopSeq 查询的轮次大于决出胜负的轮次
	ErrNoLoopSeq = errors.New("ErrBlackwhiteFinalloopLessThanSeq")
)

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import "github.com/33cn/chain33/common/log/log15"

var tlog = log15.New("module", MixX)

const (
	MaxTreeLeaves = 1024
)

// 执行器的日志类型
const (
	// TyLogParacrossCommit commit log key
	TyLogMixDeposit              = 750
	TyLogMixWithdraw             = 751
	TyLogMixTransfer             = 752
	TyLogMixAuth                 = 753
	TyLogMixConfigVk             = 754
	TyLogMixConfigAuth           = 755
	TyLogCurrentCommitTreeLeaves = 756
	TyLogCurrentCommitTreeRoots  = 756
	TyLogCommitTreeRootLeaves    = 757
	TyLogCommitTreeArchiveRoots  = 758
	TyLogNulliferSet             = 759
	TyLogAuthorizeSet            = 760
	TyLogAuthorizeSpendSet       = 761
)

//action type
const (
	MixActionConfig = iota
	MixActionDeposit
	MixActionWithdraw
	MixActionTransfer
	MixActionAuth
)

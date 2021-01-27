// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var tlog = log15.New("module", MixX)

const (
	MaxTreeLeaves = 1024
)

// 执行器的日志类型
const (
	// TyLogParacrossCommit commit log key
	TyLogMixLocalDeposit         = 750
	TyLogMixLocalNullifier       = 751
	TyLogMixLocalAuth            = 752
	TyLogMixLocalAuthSpend       = 753
	TyLogMixConfigVk             = 754
	TyLogMixConfigAuth           = 755
	TyLogCurrentCommitTreeLeaves = 756
	TyLogCurrentCommitTreeRoots  = 757
	TyLogCommitTreeRootLeaves    = 758
	TyLogCommitTreeArchiveRoots  = 759
	TyLogNulliferSet             = 760
	TyLogAuthorizeSet            = 761
	TyLogAuthorizeSpendSet       = 762
	TyLogMixConfigPaymentKey     = 763
)

//action type
const (
	MixActionConfig = iota
	MixActionDeposit
	MixActionWithdraw
	MixActionTransfer
	MixActionAuth
)

//curve H point
const (
	PointHX = "10190477835300927557649934238820360529458681672073866116232821892325659279502"
	PointHY = "7969140283216448215269095418467361784159407896899334866715345504515077887397"
)

//mix transfer tx fee
const Privacy2PrivacyTxFee = types.Coin
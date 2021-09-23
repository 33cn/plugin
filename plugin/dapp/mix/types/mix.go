// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/common/log/log15"
)

var tlog = log15.New("module", MixX)

const (
	//MaxTreeLeaves = 1024
	TreeLevel = 10
	//MimcHashSeed 电路不支持作为公共输入，设为全局常数
	MimcHashSeed = "19172955941344617222923168298456110557655645809646772800021167670156933290312"
)

// 执行器的日志类型
const (
	// TyLogParacrossCommit commit log key
	TyLogMixLocalDeposit       = 750
	TyLogMixLocalNullifier     = 751
	TyLogMixLocalAuth          = 752
	TyLogMixLocalAuthSpend     = 753
	TyLogMixConfigVk           = 754
	TyLogMixConfigAuth         = 755
	TyLogSubLeaves             = 756
	TyLogSubRoots              = 757
	TyLogArchiveRootLeaves     = 758
	TyLogCommitTreeArchiveRoot = 759
	TyLogCommitTreeStatus      = 760
	TyLogNulliferSet           = 761
	TyLogAuthorizeSet          = 762
	TyLogAuthorizeSpendSet     = 763
	TyLogMixConfigPaymentKey   = 764
)

//action type
const (
	MixActionConfig = iota
	MixActionDeposit
	MixActionWithdraw
	MixActionTransfer
	MixActionAuth
)

//circuits default file name
const (
	DepositPk = "circuit_deposit.pk"
	DepositVk = "circuit_deposit.vk"

	WithdrawPk = "circuit_withdraw.pk"
	WithdrawVk = "circuit_withdraw.vk"

	AuthPk = "circuit_auth.pk"
	AuthVk = "circuit_auth.vk"

	TransInputPk = "circuit_transfer_input.pk"
	TransInputVk = "circuit_transfer_input.vk"

	TransOutputPk = "circuit_transfer_output.pk"
	TransOutputVk = "circuit_transfer_output.vk"
)

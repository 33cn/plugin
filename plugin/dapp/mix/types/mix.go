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
	DepositCircuit = "circuit_deposit.r1cs"
	DepositPk      = "circuit_deposit.pk"
	DepositVk      = "circuit_deposit.vk"

	WithdrawCircuit = "circuit_withdraw.r1cs"
	WithdrawPk      = "circuit_withdraw.pk"
	WithdrawVk      = "circuit_withdraw.vk"

	AuthCircuit = "circuit_auth.r1cs"
	AuthPk      = "circuit_auth.pk"
	AuthVk      = "circuit_auth.vk"

	TransInputCircuit = "circuit_transfer_input.r1cs"
	TransInputPk      = "circuit_transfer_input.pk"
	TransInputVk      = "circuit_transfer_input.vk"

	TransOutputCircuit = "circuit_transfer_output.r1cs"
	TransOutputPk      = "circuit_transfer_output.pk"
	TransOutputVk      = "circuit_transfer_output.vk"
)

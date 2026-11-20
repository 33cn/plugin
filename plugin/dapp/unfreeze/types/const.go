// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// unfreeze action ty
const (
	UnfreezeActionCreate = iota + 1
	UnfreezeActionWithdraw
	UnfreezeActionTerminate

	//log for unfreeze
	TyLogCreateUnfreeze    = 2001 // TODO 修改具体编号
	TyLogWithdrawUnfreeze  = 2002
	TyLogTerminateUnfreeze = 2003
)

const (
	// Action_CreateUnfreeze Action 名字
	Action_CreateUnfreeze = "createUnfreeze"
	// Action_WithdrawUnfreeze Action 名字
	Action_WithdrawUnfreeze = "withdrawUnfreeze"
	// Action_TerminateUnfreeze Action 名字
	Action_TerminateUnfreeze = "terminateUnfreeze"
)

const (
	// FuncName_QueryUnfreezeWithdraw 查询方法名
	FuncName_QueryUnfreezeWithdraw = "QueryUnfreezeWithdraw"
)

// 包的名字可以通过配置文件来配置
// 建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
// 如果发生重名，可以通过配置文件修改这些名字
var (
	PackageName    = "chain33.unfreeze"
	RPCName        = "Chain33.Unfreeze"
	UnfreezeX      = "unfreeze"
	ExecerUnfreeze = []byte(UnfreezeX)

	FixAmountX      = "FixAmount"
	LeftProportionX = "LeftProportion"
	SupportMeans    = []string{"FixAmount", "LeftProportion"}

	ForkTerminatePartX = "ForkTerminatePart"
	ForkUnfreezeIDX    = "ForkUnfreezeIDX"
	// ForkUnfreezeDecayBoundX 解冻衰减计算有界化分叉。启用后 calcFrozen 不再按用户提供的
	// StartTime 迭代，而是先做 O(1) 收敛判定，最多 maxDecaySteps 次；同时把 v1 历史仓位里
	// TenThousandth >= 10000 的退化参数统一按"已完全解锁"处理。
	ForkUnfreezeDecayBoundX = "ForkUnfreezeDecayBound"
)

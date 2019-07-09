// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// autonomy action ty
const (
	AutonomyActionPropBoard       = iota + 1
	AutonomyActionRvkPropBoard
	AutonomyActionVotePropBoard
	AutonomyActionTmintPropBoard

	AutonomyActionPropProject
	AutonomyActionRvkPropProject
	AutonomyActionVotePropProject
	AutonomyActionTmintPropProject

	AutonomyActionPropRule
	AutonomyActionRvkPropRule
	AutonomyActionVotePropRule
	AutonomyActionTmintPropRule

	//log for autonomy
	TyLogPropBoard      = 2101
	TyLogRvkPropBoard   = 2102
	TyLogVotePropBoard  = 2103
	TyLogTmintPropBoard = 2104

	TyLogPropProject      = 2111
	TyLogRvkPropProject   = 2112
	TyLogVotePropProject  = 2113
	TyLogTmintPropProject = 2114

	TyLogPropRule      = 2121
	TyLogRvkPropRule   = 2122
	TyLogVotePropRule  = 2123
	TyLogTmintPropRule = 2124
)

const (
	// Action_PropBoard Action 名字
	Action_PropBoard = "propBoard"
	// Action_RvkPropBoard Action 名字
	Action_RvkPropBoard = "rvkPropBoard"
	// Action_TmintPropBoard Action 名字
	Action_TmintPropBoard = "tmintPropBoard"

	// Action_PropProject Action 名字
	Action_PropProject = "propProject"
	// Action_RvkPropProject Action 名字
	Action_RvkPropProject = "rvkPropProject"
	// Action_TmintPropProject Action 名字
	Action_TmintPropProject = "tmintPropProject"

	// Action_PropRule Action 名字
	Action_PropRule= "propRule"
	// Action_RvkPropRule Action 名字
	Action_RvkPropRule = "rvkPropRule"
	// Action_TmintPropRule Action 名字
	Action_TmintPropRule = "tmintPropRule"
)

// status
const (
	AutonomyStatusProposalBoard = iota + 1
	AutonomyStatusRvkPropBoard
	AutonomyStatusVotePropBoard
	AutonomyStatusTmintPropBoard
)

const (
	// FuncNameQueryProposalBoard 查询方法名
	FuncNameQueryProposalBoard   = "QueryProposalBoard"
	FuncNameQueryProposalProject = "QueryProposalProject"
	FuncNameQueryProposalRule    = "QueryProposalRule"
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	PackageName    = "chain33.autonomy"
	RPCName        = "Chain33.Autonomy"
	AutonomyX      = "autonomy"
	ExecerAutonomy = []byte(AutonomyX)
)

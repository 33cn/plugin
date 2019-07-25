// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// autonomy action ty
const (
	AutonomyActionPropBoard = iota + 1
	AutonomyActionRvkPropBoard
	AutonomyActionVotePropBoard
	AutonomyActionTmintPropBoard

	AutonomyActionPropProject
	AutonomyActionRvkPropProject
	AutonomyActionVotePropProject
	AutonomyActionPubVotePropProject
	AutonomyActionTmintPropProject

	AutonomyActionPropRule
	AutonomyActionRvkPropRule
	AutonomyActionVotePropRule
	AutonomyActionTmintPropRule

	AutonomyActionTransfer
	AutonomyActionCommentProp

	//log for autonomy
	TyLogPropBoard      = 2101
	TyLogRvkPropBoard   = 2102
	TyLogVotePropBoard  = 2103
	TyLogTmintPropBoard = 2104

	TyLogPropProject        = 2111
	TyLogRvkPropProject     = 2112
	TyLogVotePropProject    = 2113
	TyLogPubVotePropProject = 2114
	TyLogTmintPropProject   = 2115

	TyLogPropRule      = 2121
	TyLogRvkPropRule   = 2122
	TyLogVotePropRule  = 2123
	TyLogTmintPropRule = 2124

	TyLogCommentProp = 2131
)

// status
const (
	AutonomyStatusProposalBoard = iota + 1
	AutonomyStatusRvkPropBoard
	AutonomyStatusVotePropBoard
	AutonomyStatusTmintPropBoard
)

const (
	AutonomyStatusProposalProject = iota + 1
	AutonomyStatusRvkPropProject
	AutonomyStatusVotePropProject
	AutonomyStatusPubVotePropProject
	AutonomyStatusTmintPropProject
)

const (
	AutonomyStatusProposalRule = iota + 1
	AutonomyStatusRvkPropRule
	AutonomyStatusVotePropRule
	AutonomyStatusTmintPropRule
)

const (
	// GetProposalBoard 用于在cmd里面的区分不同的查询
	GetProposalBoard = "GetProposalBoard"
	// ListProposalBoard
	ListProposalBoard = "ListProposalBoard"
	// GetProposalProject 用于在cmd里面的区分不同的查询
	GetProposalProject = "GetProposalProject"
	// ListProposalProject
	ListProposalProject = "ListProposalProject"
	// GetProposalRule 用于在cmd里面的区分不同的查询
	GetProposalRule = "GetProposalRule"
	// ListProposalRule
	ListProposalRule = "ListProposalRule"
	// ListProposalComment
	ListProposalComment = "ListProposalComment"
)

//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	AutonomyX      = "autonomy"
	ExecerAutonomy = []byte(AutonomyX)
)

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

//dpos action ty
const (
	DposVoteActionRegist = iota + 1
	DposVoteActionCancelRegist
	DposVoteActionReRegist
	DposVoteActionVote
	DposVoteActionCancelVote
	DposVoteActionRegistVrfM
	DposVoteActionRegistVrfRP

	CandidatorStatusRegist = iota + 1
	CandidatorStatusVoted
	CandidatorStatusCancelVoted
	CandidatorStatusCancelRegist
	CandidatorStatusReRegist

	VrfStatusMRegist = iota + 1
	VrfStatusRPRegist
)

//log ty
const (
	TyLogCandicatorRegist   = 1001
	TyLogCandicatorVoted    = 1002
	TyLogCandicatorCancelVoted  = 1003
	TyLogCandicatorCancelRegist  = 1004
	TyLogCandicatorReRegist  = 1005
	TyLogVrfMRegist  = 1006
	TyLogVrfRPRegist  = 1007
)

const (
	VoteFrozenTime = 3 * 24 * 3600
	RegistFrozenCoins = 1000000000000
)
//包的名字可以通过配置文件来配置
//建议用github的组织名称，或者用户名字开头, 再加上自己的插件的名字
//如果发生重名，可以通过配置文件修改这些名字
var (
	DPosX = "dpos"
	ExecerDposVote = []byte(DPosX)
)

const (
	//FuncNameQueryCandidatorByPubkeys func name
	FuncNameQueryCandidatorByPubkeys = "QueryCandidatorByPubkeys"

	//FuncNameQueryCandidatorByTopN func name
	FuncNameQueryCandidatorByTopN = "QueryCandidatorByTopN"

	//FuncNameQueryVrfByTime func name
	FuncNameQueryVrfByTime = "QueryVrfByTime"

	//FuncNameQueryVrfByCycle func name
	FuncNameQueryVrfByCycle = "QueryVrfByCycle"

	//FuncNameQueryVote func name
	FuncNameQueryVote = "QueryVote"

	//CreateRegistTx 创建注册候选节点交易
	CreateRegistTx = "Regist"

	//CreateCancelRegistTx 创建取消候选节点的交易
	CreateCancelRegistTx = "CancelRegist"

	//CreateReRegistTx 创建重新注册候选节点的交易
	CreateReRegistTx = "ReRegist"

	//CreateVoteTx 创建为候选节点投票的交易
	CreateVoteTx = "Vote"

	//CreateCancelVoteTx 创建取消对候选节点投票的交易
	CreateCancelVoteTx = "CancelVote"

	//CreateRegistVrfMTx 创建注册Vrf的M信息的交易
	CreateRegistVrfMTx = "RegistVrfM"

	//CreateRegistVrfRPTx 创建注册Vrf的R/P信息的交易
	CreateRegistVrfRPTx = "RegistVrfRP"

	//QueryVrfByTime 创建根据time查询Vrf信息查询
	QueryVrfByTime = 1

	//QueryVrfByCycle 创建根据cycle查询Vrf信息查询
	QueryVrfByCycle = 2
)

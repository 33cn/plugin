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
	DposVoteActionRecordCB
	DPosVoteActionRegistTopNCandidator

	CandidatorStatusRegist = iota + 1
	CandidatorStatusVoted
	CandidatorStatusCancelVoted
	CandidatorStatusCancelRegist
	CandidatorStatusReRegist

	VrfStatusMRegist = iota + 1
	VrfStatusRPRegist

	CBStatusRecord = iota + 1

	TopNCandidatorStatusRegist = iota + 1
)

//log ty
const (
	TyLogCandicatorRegist        = 1001
	TyLogCandicatorVoted         = 1002
	TyLogCandicatorCancelVoted   = 1003
	TyLogCandicatorCancelRegist  = 1004
	TyLogCandicatorReRegist      = 1005
	TyLogVrfMRegist              = 1006
	TyLogVrfRPRegist             = 1007
	TyLogCBInfoRecord            = 1008
	TyLogTopNCandidatorRegist    = 1009
)

const (
	VoteFrozenTime = 3 * 24 * 3600
	RegistFrozenCoins = 1000000000000

	VoteTypeNone          int32 = 1
	VoteTypeVote          int32 = 2
	VoteTypeCancelVote    int32 = 3
	VoteTypeCancelAllVote int32 = 4

	TopNCandidatorsVoteInit int64 = 0
	TopNCandidatorsVoteMajorOK int64 = 1
	TopNCandidatorsVoteMajorFail int64 = 2
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

	//FuncNameQueryVrfByCycleForTopN func name
	FuncNameQueryVrfByCycleForTopN = "QueryVrfByCycleForTopN"

	//FuncNameQueryVrfByCycleForPubkeys func name
	FuncNameQueryVrfByCycleForPubkeys = "QueryVrfByCycleForPubkeys"

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

	//CreateRecordCBTx 创建记录CB信息的交易
	CreateRecordCBTx = "RecordCB"

	//QueryVrfByTime 根据time查询Vrf信息
	QueryVrfByTime = 1

	//QueryVrfByCycle 根据cycle查询Vrf信息
	QueryVrfByCycle = 2

	//QueryVrfByCycleForTopN 根据cycle查询当前topN的候选节点的Vrf信息
	QueryVrfByCycleForTopN = 3

	//QueryVrfByCycleForPubkeys 根据cycle查询指定pubkey的多个候选节点的Vrf信息
	QueryVrfByCycleForPubkeys = 4

	//FuncNameQueryCBInfoByCycle func name
	FuncNameQueryCBInfoByCycle = "QueryCBInfoByCycle"

	//FuncNameQueryCBInfoByHeight func name
	FuncNameQueryCBInfoByHeight = "QueryCBInfoByHeight"

	//FuncNameQueryCBInfoByHash func name
	FuncNameQueryCBInfoByHash = "QueryCBInfoByHash"

	//FuncNameQueryLatestCBInfoByHeight func name
	FuncNameQueryLatestCBInfoByHeight = "QueryLatestCBInfoByHeight"

	//QueryCBInfoByCycle 根据cycle查询cycle boundary信息
	QueryCBInfoByCycle = 1

	//QueryCBInfoByHeight 根据stopHeight查询cycle boundary信息
	QueryCBInfoByHeight = 2

	//QueryCBInfoByHash 根据stopHash查询cycle boundary信息
	QueryCBInfoByHash = 3

	//QueryCBInfoByHeight 根据stopHeight查询cycle boundary信息
	QueryLatestCBInfoByHeight = 4

	//FuncNameQueryTopNByVersion func name
	FuncNameQueryTopNByVersion = "QueryTopNByVersion"
)

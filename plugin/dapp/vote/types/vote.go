package types

import (
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyCreateGroupAction
	TyUpdateGroupAction
	TyCreateVoteAction
	TyCommitVoteAction
	TyCloseVoteAction
	TyUpdateMemberAction

	NameCreateGroupAction  = "CreateGroup"
	NameUpdateGroupAction  = "UpdateGroup"
	NameCreateVoteAction   = "CreateVote"
	NameCommitVoteAction   = "CommitVote"
	NameCloseVoteAction    = "CloseVote"
	NameUpdateMemberAction = "UpdateMember"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyCreateGroupLog
	TyUpdateGroupLog
	TyCreateVoteLog
	TyCommitVoteLog
	TyCloseVoteLog
	TyUpdateMemberLog

	NameCreateGroupLog  = "CreateGroupLog"
	NameUpdateGroupLog  = "UpdateGroupLog"
	NameCreateVoteLog   = "CreateVoteLog"
	NameCommitVoteLog   = "CommitVoteLog"
	NameCloseVoteLog    = "CloseVoteLog"
	NameUpdateMemberLog = "UpdateMemberLog"
)

var (
	//VoteX 执行器名称定义
	VoteX = "vote"
	//定义actionMap
	actionMap = map[string]int32{
		NameCreateGroupAction:  TyCreateGroupAction,
		NameUpdateGroupAction:  TyUpdateGroupAction,
		NameCreateVoteAction:   TyCreateVoteAction,
		NameCommitVoteAction:   TyCommitVoteAction,
		NameCloseVoteAction:    TyCloseVoteAction,
		NameUpdateMemberAction: TyUpdateMemberAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyCreateGroupLog:  {Ty: reflect.TypeOf(GroupInfo{}), Name: NameCreateGroupLog},
		TyUpdateGroupLog:  {Ty: reflect.TypeOf(GroupInfo{}), Name: NameUpdateGroupLog},
		TyCreateVoteLog:   {Ty: reflect.TypeOf(VoteInfo{}), Name: NameCreateVoteLog},
		TyCommitVoteLog:   {Ty: reflect.TypeOf(CommitInfo{}), Name: NameCommitVoteLog},
		TyCloseVoteLog:    {Ty: reflect.TypeOf(VoteInfo{}), Name: NameCloseVoteLog},
		TyUpdateMemberLog: {Ty: reflect.TypeOf(MemberInfo{}), Name: NameUpdateMemberLog},
	}
	tlog = log.New("module", "vote.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(VoteX))
	//注册合约启用高度
	types.RegFork(VoteX, InitFork)
	types.RegExec(VoteX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(VoteX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(VoteX, NewType(cfg))
}

type voteType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *voteType {
	c := &voteType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (v *voteType) GetPayload() types.Message {
	return &VoteAction{}
}

// GeTypeMap 获取合约action的id和name信息
func (v *voteType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (v *voteType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

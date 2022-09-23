package types

import (
	"reflect"
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

const (

	// RollupCommitTimeout rollup提交超时秒数, 超过该值未提交下一个round数据, 即为超时
	RollupCommitTimeout = 600
)

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyCommitAction

	NameCommitAction = "Commit"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyCommitRoundInfoLog
	TyRollupStatusLog

	NameCommitRoundInfoLog = "CommitRoundInfoLog"
	NameRollupStatusLog    = "RollupStatusLog"
)

var (
	//RollupX 执行器名称定义
	RollupX = "rollup"
	//定义actionMap
	actionMap = map[string]int32{
		NameCommitAction: TyCommitAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyCommitRoundInfoLog: {Ty: reflect.TypeOf(CommitRoundInfo{}), Name: NameCommitRoundInfoLog},
		TyRollupStatusLog:    {Ty: reflect.TypeOf(RollupStatus{}), Name: NameRollupStatusLog},
	}
	tlog = log.New("module", "rollup.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(RollupX))
	//注册合约启用高度
	types.RegFork(RollupX, InitFork)
	types.RegExec(RollupX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(RollupX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(RollupX, NewType(cfg))
}

type rollupType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *rollupType {
	c := &rollupType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (r *rollupType) GetPayload() types.Message {
	return &RollupAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (r *rollupType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (r *rollupType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

// FormatHexPubKey format
func FormatHexPubKey(pubKey string) string {
	if strings.HasPrefix(pubKey, "0x") {
		return pubKey[2:]
	}
	return pubKey
}

package types

import (
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
	TyRegisterAction
	TyResetAction
	TyTransferAction
	TySuperviseAction
	TyApplyAction

	NameRegisterAction  = "Register"
	NameResetAction     = "Reset"
	NameTransferAction  = "Transfer"
	NameSuperviseAction = "Supervise"
	NameApplyAction     = "Apply"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyRegisterLog
	TyApplyLog
	TyTransferLog
	TySuperviseLog
)

//状态
const (
	UnknownStatus = iota
	Normal
	Frozen
	Locked
	Expired
)

var (
	//AccountmanagerX 执行器名称定义
	AccountmanagerX = "accountmanager"
	//定义actionMap
	actionMap = map[string]int32{
		NameRegisterAction:  TyRegisterAction,
		NameResetAction:     TyResetAction,
		NameApplyAction:     TyApplyAction,
		NameTransferAction:  TyTransferAction,
		NameSuperviseAction: TySuperviseAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//LogID:	{Ty: reflect.TypeOf(LogStruct), Name: LogName},
	}
	tlog = log.New("module", "accountmanager.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(AccountmanagerX))
	//注册合约启用高度
	types.RegFork(AccountmanagerX, InitFork)
	types.RegExec(AccountmanagerX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(AccountmanagerX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(AccountmanagerX, NewType(cfg))
}

type accountmanagerType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *accountmanagerType {
	c := &accountmanagerType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (a *accountmanagerType) GetPayload() types.Message {
	return &AccountmanagerAction{}
}

// GeTypeMap 获取合约action的id和name信息
func (a *accountmanagerType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (a *accountmanagerType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

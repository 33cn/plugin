package types

import (
	"reflect"

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
	NameResetAction     = "ResetKey"
	NameTransferAction  = "Transfer"
	NameSuperviseAction = "Supervise"
	NameApplyAction     = "Apply"

	FuncNameQueryAccountByID      = "QueryAccountByID"
	FuncNameQueryAccountsByStatus = "QueryAccountsByStatus"
	FuncNameQueryExpiredAccounts  = "QueryExpiredAccounts"
	FuncNameQueryAccountByAddr    = "QueryAccountByAddr"
	FuncNameQueryBalanceByID      = "QueryBalanceByID"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyRegisterLog
	TyResetLog
	TyTransferLog
	TySuperviseLog
	TyApplyLog
)

//状态
const (
	Normal = int32(iota)
	Frozen
	Locked
	Expired
)

//supervior op
const (
	UnknownSupervisorOp = int32(iota)
	Freeze
	UnFreeze
	AddExpire
	Authorize
)

//apply  op
const (
	UnknownApplyOp = int32(iota)
	RevokeReset
	EnforceReset
)

//list ...
const (
	ListDESC = int32(0)
	ListASC  = int32(1)
	ListSeek = int32(2)
)
const (
	//Count 单次list还回条数
	Count = int32(10)
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
		TyRegisterLog:  {Ty: reflect.TypeOf(AccountReceipt{}), Name: "TyRegisterLog"},
		TyResetLog:     {Ty: reflect.TypeOf(TransferReceipt{}), Name: "TyResetLog"},
		TyTransferLog:  {Ty: reflect.TypeOf(AccountReceipt{}), Name: "TyTransferLog"},
		TySuperviseLog: {Ty: reflect.TypeOf(SuperviseReceipt{}), Name: "TySuperviseLog"},
		TyApplyLog:     {Ty: reflect.TypeOf(AccountReceipt{}), Name: "TyApplyLog"},
	}
	//tlog = log.New("module", "accountmanager.types")
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

//AccountmanagerType ...
type AccountmanagerType struct {
	types.ExecTypeBase
}

//NewType ...
func NewType(cfg *types.Chain33Config) *AccountmanagerType {
	c := &AccountmanagerType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (a *AccountmanagerType) GetPayload() types.Message {
	return &AccountmanagerAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (a *AccountmanagerType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (a *AccountmanagerType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

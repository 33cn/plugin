package types

import (
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"reflect"
	//"reflect"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknowAction = iota + 100
	TyMintAction
	TyTransferAction
	TyConfirmAction

	NameMintAction     = "Mint"
	NameTransferAction = "Transfer"
	NameConfirmAction  = "Confirm"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyMintLog
	TyTransferLog
	TyConfirmLog
	TyPendingTxLog

	NamePendingTxLog = "PendingTxLog"
)

var (
	//RgbxX 执行器名称定义
	RgbxX = "rgbx"
	//定义actionMap
	actionMap = map[string]int32{
		NameMintAction:     TyMintAction,
		NameTransferAction: TyTransferAction,
		NameConfirmAction:  TyConfirmAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyPendingTxLog: {Ty: reflect.TypeOf(PendingTx{}), Name: NamePendingTxLog},
	}
	tlog = log.New("module", "rgbx.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(RgbxX))
	//注册合约启用高度
	types.RegFork(RgbxX, InitFork)
	types.RegExec(RgbxX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(RgbxX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(RgbxX, NewType(cfg))
}

type rgbxType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *rgbxType {
	c := &rgbxType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (r *rgbxType) GetPayload() types.Message {
	return &RgbxAction{}
}

// GeTypeMap 获取合约action的id和name信息
func (r *rgbxType) GetTypeMap() map[string]int32 {
	return actionMap
}

// GetLogMap 获取合约log相关信息
func (r *rgbxType) GetLogMap() map[int64]*types.LogInfo {
	return logMap
}

// GetActionName get action name by action type
func GetActionName(ty int32) string {
	if ty == TyMintAction {
		return NameMintAction
	} else if ty == TyTransferAction {
		return NameTransferAction
	} else if ty == TyConfirmAction {
		return NameConfirmAction
	} else {
		return "unknownAction"
	}
}

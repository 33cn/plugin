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
	TyTransferAction
	TyWithdrawAction
	TyTransferToExecAction
	TyMintAction
	TyBurnAction

	NameTransferAction       = "Transfer"
	NameWithdrawAction       = "Withdraw"
	NameTransferToExecAction = "TransferToExec"
	NameMintAction           = "Mint"
	NameBurnAction           = "Burn"
	NameMintMapAction        = "MintMap"
	NameBurnMapAction        = "BurnMap"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyTransferLog
	TyWithdrawLog
	TyTransferToExecLog
	TyMintLog
	TyBurnLog
)

var (
	//EvmxgoX 执行器名称定义
	EvmxgoX = "evmxgo"
	//定义actionMap
	actionMap = map[string]int32{
		NameTransferAction:       TyTransferAction,
		NameWithdrawAction:       TyWithdrawAction,
		NameTransferToExecAction: TyTransferToExecAction,
		NameMintAction:           TyMintAction,
		NameBurnAction:           TyBurnAction,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		//LogID:	{Ty: reflect.TypeOf(LogStruct), Name: LogName},
	}
	tlog = log.New("module", "evmxgo.types")
)

// init defines a register function
func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(EvmxgoX))
	//注册合约启用高度
	types.RegFork(EvmxgoX, InitFork)
	types.RegExec(EvmxgoX, InitExecutor)
}

// InitFork defines register fork
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(EvmxgoX, "Enable", 0)
}

// InitExecutor defines register executor
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(EvmxgoX, NewType(cfg))
}

type evmxgoType struct {
	types.ExecTypeBase
}

func NewType(cfg *types.Chain33Config) *evmxgoType {
	c := &evmxgoType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload 获取合约action结构
func (e *evmxgoType) GetPayload() types.Message {
	return &EvmxgoAction{}
}

// GetTypeMap 获取合约action的id和name信息
func (e *evmxgoType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Transfer":        ActionTransfer,
		"Withdraw":        ActionWithdraw,
		"TransferToExec":  EvmxgoActionTransferToExec,
		"Mint":            EvmxgoActionMint,
		"Burn":            EvmxgoActionBurn,
		NameMintMapAction: EvmxgoActionMintMap,
		NameBurnMapAction: EvmxgoActionBurnMap,
	}
}

// GetLogMap 获取合约log相关信息
func (e *evmxgoType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogEvmxgoTransfer:        {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenTransfer"},
		TyLogEvmxgoDeposit:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenDeposit"},
		TyLogEvmxgoExecTransfer:    {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecTransfer"},
		TyLogEvmxgoExecWithdraw:    {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecWithdraw"},
		TyLogEvmxgoExecDeposit:     {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecDeposit"},
		TyLogEvmxgoExecFrozen:      {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecFrozen"},
		TyLogEvmxgoExecActive:      {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecActive"},
		TyLogEvmxgoGenesisTransfer: {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenGenesisTransfer"},
		TyLogEvmxgoGenesisDeposit:  {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenGenesisDeposit"},
		TyLogEvmxgoMint:            {Ty: reflect.TypeOf(ReceiptEvmxgoAmount{}), Name: "LogMintToken"},
		TyLogEvmxgoBurn:            {Ty: reflect.TypeOf(ReceiptEvmxgoAmount{}), Name: "LogBurnToken"},
		TyLogEvmxgoMintMap:         {Ty: reflect.TypeOf(ReceiptEvmxgoAmount{}), Name: "LogMintMapToken"},
		TyLogEvmxgoBurnMap:         {Ty: reflect.TypeOf(ReceiptEvmxgoAmount{}), Name: "LogBurnMapToken"},
	}

}

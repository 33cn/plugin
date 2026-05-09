package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
	//"reflect"
)

/*
 * 交易相关类型定义
 * 交易action通常有对应的log结构，用于交易回执日志记录
 * 每一种action和log需要用id数值和name名称加以区分
 */

// action类型id和name，这些常量可以自定义修改
const (
	TyUnknownAction = iota + 100
	TyMintAction
	TyTransferAction
	TyConfirmAction
	TyCommitDKGAction
	TyDepositAsset
	TyWithDrawAsset

	NameMintAction      = "Mint"
	NameTransferAction  = "Transfer"
	NameConfirmAction   = "Confirm"
	NameCommitDKGAction = "CommitDKG"
	// 须与 RgbxAction protobuf oneof 字段名一致（chain33 ExecTypeBase.SetChild 用 strings.ToLower 映射），
	// 例如 oneof deposit -> 此处为 "Deposit"，不能为 "DepositAsset"。
	NameDepositAssetAction  = "Deposit"
	NameWithdrawAssetAction = "Withdraw"
)

// log类型id值
const (
	TyUnknownLog = iota + 100
	TyAssetLog
	TyPendingTxLog
	TyCommitDKGLog
	TyDepositAssetLog
	TyWithdrawAssetLog

	NameAssetLog         = "AssetLog"
	NamePendingTxLog     = "PendingTxLog"
	NameCommitDKGLog     = "CommitDKGLog"
	NameDepositAssetLog  = "DepositAssetLog"
	NameWithdrawAssetLog = "WithdrawAssetLog"
)

var (
	//RgbxX 执行器名称定义
	RgbxX = "rgbx"
	//定义actionMap
	actionMap = map[string]int32{
		NameMintAction:          TyMintAction,
		NameTransferAction:      TyTransferAction,
		NameConfirmAction:       TyConfirmAction,
		NameCommitDKGAction:     TyCommitDKGAction,
		NameDepositAssetAction:  TyDepositAsset,
		NameWithdrawAssetAction: TyWithDrawAsset,
	}
	//定义log的id和具体log类型及名称，填入具体自定义log类型
	logMap = map[int64]*types.LogInfo{
		TyAssetLog:     {Ty: reflect.TypeOf(RgbxAsset{}), Name: NameAssetLog},
		TyPendingTxLog: {Ty: reflect.TypeOf(PendingTx{}), Name: NamePendingTxLog},
		TyCommitDKGLog: {Ty: reflect.TypeOf(types.ReqAddrs{}), Name: NameCommitDKGLog},
	}
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
	switch ty {
	case TyMintAction:
		return NameMintAction
	case TyTransferAction:
		return NameTransferAction
	case TyConfirmAction:
		return NameConfirmAction
	case TyCommitDKGAction:
		return NameCommitDKGAction
	case TyDepositAsset:
		return NameDepositAssetAction
	case TyWithDrawAsset:
		return NameWithdrawAssetAction
	default:
		return "unknownAction"
	}
}

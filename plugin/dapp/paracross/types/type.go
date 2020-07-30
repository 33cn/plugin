// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	// ParaX paracross exec name
	ParaX = "paracross"
	glog  = log.New("module", ParaX)
	// ForkCommitTx main chain support paracross commit tx
	ForkCommitTx = "ForkParacrossCommitTx"
	// MainForkParacrossCommitTx 平行链配置项对应主链的ForkCommitTx 高度
	MainForkParacrossCommitTx = "mainForkParacrossCommitTx"
	// ForkLoopCheckCommitTxDone 循环检查共识交易done的fork
	ForkLoopCheckCommitTxDone = "ForkLoopCheckCommitTxDone"
	// MainLoopCheckCommitTxDoneForkHeight 平行链的配置项，对应主链的ForkLoopCheckCommitTxDone高度
	MainLoopCheckCommitTxDoneForkHeight = "mainLoopCheckCommitTxDoneForkHeight"
	// ForkParaSelfConsStages 平行链自共识分阶段共识
	ForkParaSelfConsStages = "ForkParaSelfConsStages"
	// ForkParaAssetTransferRbk 平行链资产转移平行链失败主链回滚
	ForkParaAssetTransferRbk = "ForkParaAssetTransferRbk"

	// ParaConsSubConf sub
	ParaConsSubConf = "consensus.sub.para"
	//ParaPrefixConsSubConf prefix
	ParaPrefixConsSubConf = "config." + ParaConsSubConf
	//ParaSelfConsInitConf self stage init config
	ParaSelfConsInitConf = "paraSelfConsInitDisable"
	//ParaSelfConsConfPreContract self consens enable string as ["0-100"] config pre stage contract
	ParaSelfConsConfPreContract = "selfConsensEnablePreContract"
	//ParaFilterIgnoreTxGroup adapt 6.1.0 to check para tx in group
	ParaFilterIgnoreTxGroup = "filterIgnoreParaTxGroup"
)

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(ParaX))
	types.RegFork(ParaX, InitFork)
	types.RegExec(ParaX, InitExecutor)

}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(ParaX, "Enable", 0)
	cfg.RegisterDappFork(ParaX, "ForkParacrossWithdrawFromParachain", 1298600)
	cfg.RegisterDappFork(ParaX, ForkCommitTx, 1850000)
	cfg.RegisterDappFork(ParaX, ForkLoopCheckCommitTxDone, 3230000)
	cfg.RegisterDappFork(ParaX, ForkParaAssetTransferRbk, 4500000)

	//只在平行链启用
	cfg.RegisterDappFork(ParaX, ForkParaSelfConsStages, types.MaxHeight)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(ParaX, NewType(cfg))
}

// GetExecName get para exec name
func GetExecName(cfg *types.Chain33Config) string {
	return cfg.ExecName(ParaX)
}

// ParacrossType base paracross type
type ParacrossType struct {
	types.ExecTypeBase
}

// NewType get paracross type
func NewType(cfg *types.Chain33Config) *ParacrossType {
	c := &ParacrossType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (p *ParacrossType) GetName() string {
	return ParaX
}

// GetLogMap get receipt log map
func (p *ParacrossType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogParacrossCommit:           {Ty: reflect.TypeOf(ReceiptParacrossCommit{}), Name: "LogParacrossCommit"},
		TyLogParacrossCommitDone:       {Ty: reflect.TypeOf(ReceiptParacrossDone{}), Name: "LogParacrossCommitDone"},
		TyLogParacrossCommitRecord:     {Ty: reflect.TypeOf(ReceiptParacrossRecord{}), Name: "LogParacrossCommitRecord"},
		TyLogParaAssetWithdraw:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetWithdraw"},
		TyLogParaAssetTransfer:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetTransfer"},
		TyLogParaAssetDeposit:          {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaAssetDeposit"},
		TyLogParaCrossAssetTransfer:    {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogParaCrossAssetTransfer"},
		TyLogParacrossMiner:            {Ty: reflect.TypeOf(ReceiptParacrossMiner{}), Name: "LogParacrossMiner"},
		TyLogParaNodeConfig:            {Ty: reflect.TypeOf(ReceiptParaNodeConfig{}), Name: "LogParaNodeConfig"},
		TyLogParaNodeStatusUpdate:      {Ty: reflect.TypeOf(ReceiptParaNodeAddrStatUpdate{}), Name: "LogParaNodeAddrStatUpdate"},
		TyLogParaNodeGroupAddrsUpdate:  {Ty: reflect.TypeOf(types.ReceiptConfig{}), Name: "LogParaNodeGroupAddrsUpdate"},
		TyLogParaNodeVoteDone:          {Ty: reflect.TypeOf(ReceiptParaNodeVoteDone{}), Name: "LogParaNodeVoteDone"},
		TyLogParaNodeGroupConfig:       {Ty: reflect.TypeOf(ReceiptParaNodeGroupConfig{}), Name: "LogParaNodeGroupConfig"},
		TyLogParaNodeGroupStatusUpdate: {Ty: reflect.TypeOf(ReceiptParaNodeGroupConfig{}), Name: "LogParaNodeGroupStatusUpdate"},
		TyLogParaSelfConsStageConfig:   {Ty: reflect.TypeOf(ReceiptSelfConsStageConfig{}), Name: "LogParaSelfConsStageConfig"},
		TyLogParaStageVoteDone:         {Ty: reflect.TypeOf(ReceiptSelfConsStageVoteDone{}), Name: "LogParaSelfConfStageVoteDoen"},
		TyLogParaStageGroupUpdate:      {Ty: reflect.TypeOf(ReceiptSelfConsStagesUpdate{}), Name: "LogParaSelfConfStagesUpdate"},
	}
}

// GetTypeMap get action type
func (p *ParacrossType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Commit":             ParacrossActionCommit,
		"Miner":              ParacrossActionMiner,
		"AssetTransfer":      ParacrossActionAssetTransfer,
		"AssetWithdraw":      ParacrossActionAssetWithdraw,
		"Transfer":           ParacrossActionTransfer,
		"Withdraw":           ParacrossActionWithdraw,
		"TransferToExec":     ParacrossActionTransferToExec,
		"CrossAssetTransfer": ParacrossActionCrossAssetTransfer,
		"NodeConfig":         ParacrossActionNodeConfig,
		"NodeGroupConfig":    ParacrossActionNodeGroupApply,
		"SelfStageConfig":    ParacrossActionSelfStageConfig,
	}
}

// GetPayload paracross get action payload
func (p *ParacrossType) GetPayload() types.Message {
	return &ParacrossAction{}
}

// CreateTx paracross create tx by different action
func (p ParacrossType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	cfg := p.GetConfig()
	//保留老的ParacrossAssetTransfer接口，默认的AssetTransfer　也可以
	if action == "ParacrossAssetTransfer" || action == "ParacrossAssetWithdraw" {
		var param types.CreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawAssetTransferTx(cfg, &param)
	} else if action == "Transfer" || action == "Withdraw" || action == "TransferToExec" {
		//transfer/withdraw/toExec 需要特殊处理主链上的tx.to场景
		return p.CreateRawTransferTx(action, message)
	}
	return p.ExecTypeBase.CreateTx(action, message)
}

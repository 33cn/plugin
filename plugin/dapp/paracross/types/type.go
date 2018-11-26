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

	/*
		logInfo = map[int64]*types.LogInfo{
			TyLogParacrossCommit:  {reflect.TypeOf(ReceiptParacrossCommit{}), "LogParacrossCommit"},
			TyLogParacrossCommitDone:  {reflect.TypeOf(ReceiptParacrossDone{}), "LogParacrossDone"},
			TyLogParacrossCommitRecord: {reflect.TypeOf(ReceiptParacrossRecord{}), "LogParacrossCommitRecord"},
			TyLogParaAssetTransfer: {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParacrossAssetTransfer"},
			TyLogParaAssetWithdraw:   {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParacrossAssetWithdraw"},
			TyLogParacrossMiner:  {reflect.TypeOf(ReceiptParacrossMiner{}), "LogParacrossMiner"},
			TyLogParaAssetDeposit: {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParacrossAssetDeposit"},
		}*/

	// init query rpc
	/* TODO-TODO
	types.RegisterRPCQueryHandle("ParacrossGetTitle", &ParacrossGetTitle{})
	types.RegisterRPCQueryHandle("paracrossListTitles", &paracrossListTitles{})
	types.RegisterRPCQueryHandle("paracrossGetTitleHeight", &paracrossGetTitleHeight{})
	types.RegisterRPCQueryHandle("paracrossGetAssetTxResult", &paracrossGetAssetTxResult{})
	*/
)

func init() {
	// init executor type
	types.AllowUserExec = append(types.AllowUserExec, []byte(ParaX))
	types.RegistorExecutor(ParaX, NewType())
	types.RegisterDappFork(ParaX, "Enable", 0)
}

// GetExecName get para exec name
func GetExecName() string {
	return types.ExecName(ParaX)
}

// ParacrossType base paracross type
type ParacrossType struct {
	types.ExecTypeBase
}

// NewType get paracross type
func NewType() *ParacrossType {
	c := &ParacrossType{}
	c.SetChild(c)
	return c
}

// GetLogMap get receipt log map
func (p *ParacrossType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogParacrossCommit:       {reflect.TypeOf(ReceiptParacrossCommit{}), "LogParacrossCommit"},
		TyLogParacrossCommitDone:   {reflect.TypeOf(ReceiptParacrossDone{}), "LogParacrossCommitDone"},
		TyLogParacrossCommitRecord: {reflect.TypeOf(ReceiptParacrossRecord{}), "LogParacrossCommitRecord"},
		TyLogParaAssetWithdraw:     {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParaAssetWithdraw"},
		TyLogParaAssetTransfer:     {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParaAssetTransfer"},
		TyLogParaAssetDeposit:      {reflect.TypeOf(types.ReceiptAccountTransfer{}), "LogParaAssetDeposit"},
		TyLogParacrossMiner:        {reflect.TypeOf(ReceiptParacrossMiner{}), "LogParacrossMiner"},
	}
}

// GetTypeMap get action type
func (p *ParacrossType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Commit":         ParacrossActionCommit,
		"Miner":          ParacrossActionMiner,
		"AssetTransfer":  ParacrossActionAssetTransfer,
		"AssetWithdraw":  ParacrossActionAssetWithdraw,
		"Transfer":       ParacrossActionTransfer,
		"Withdraw":       ParacrossActionWithdraw,
		"TransferToExec": ParacrossActionTransferToExec,
	}
}

// GetPayload paracross get action payload
func (p *ParacrossType) GetPayload() types.Message {
	return &ParacrossAction{}
}

// CreateTx paracross create tx by different action
func (p ParacrossType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	if action == "ParacrossCommit" {
		var param paracrossCommitTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return createRawParacrossCommitTx(&param)
	} else if action == "ParacrossAssetTransfer" || action == "ParacrossAssetWithdraw" {
		var param types.CreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawAssetTransferTx(&param)

	} else if action == "ParacrossTransfer" || action == "ParacrossWithdraw" || action == "ParacrossTransferToExec" {
		var param types.CreateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			glog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawTransferTx(&param)

	}

	return nil, types.ErrNotSupport
}

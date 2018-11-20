// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	elog = log.New("module", "exectype.evm")

	actionName = map[string]int32{
		"EvmCreate": EvmCreateAction,
		"EvmCall":   EvmCallAction,
	}
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerEvm)
	// init executor type
	types.RegistorExecutor(ExecutorName, NewType())
	types.RegisterDappFork(ExecutorName, "ForkEVMState", 650000)
	types.RegisterDappFork(ExecutorName, "ForkEVMKVHash", 1000000)
	types.RegisterDappFork(ExecutorName, "Enable", 500000)
}

// EvmType EVM类型定义
type EvmType struct {
	types.ExecTypeBase
}

// NewType 新建EVM类型对象
func NewType() *EvmType {
	c := &EvmType{}
	c.SetChild(c)
	return c
}

// GetPayload 获取消息负载结构
func (evm *EvmType) GetPayload() types.Message {
	return &EVMContractAction{}
}

// ActionName 获取ActionName
func (evm EvmType) ActionName(tx *types.Transaction) string {
	// 这个需要通过合约交易目标地址来判断Action
	// 如果目标地址为空，或为evm的固定合约地址，则为创建合约，否则为调用合约
	if strings.EqualFold(tx.To, address.ExecAddress(types.ExecName(ExecutorName))) {
		return "createEvmContract"
	}
	return "callEvmContract"
}

// GetTypeMap 获取类型映射
func (evm *EvmType) GetTypeMap() map[string]int32 {
	return actionName
}

// GetRealToAddr 获取实际地址
func (evm EvmType) GetRealToAddr(tx *types.Transaction) string {
	if string(tx.Execer) == ExecutorName {
		return tx.To
	}
	var action EVMContractAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return tx.To
	}
	return tx.To
}

// Amount 获取金额
func (evm EvmType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

// CreateTx 创建交易对象
func (evm EvmType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	elog.Debug("evm.CreateTx", "action", action)
	if action == "CreateCall" {
		var param CreateCallTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return createRawEvmCreateCallTx(&param)
	}
	return nil, types.ErrNotSupport
}

// GetLogMap 获取日志类型映射
func (evm *EvmType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

func createRawEvmCreateCallTx(parm *CreateCallTx) (*types.Transaction, error) {
	if parm == nil {
		elog.Error("createRawEvmCreateCallTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	bCode, err := common.FromHex(parm.Code)
	if err != nil {
		elog.Error("createRawEvmCreateCallTx", "parm.Code", parm.Code)
		return nil, err
	}

	action := &EVMContractAction{
		Amount:   parm.Amount,
		Code:     bCode,
		GasLimit: parm.GasLimit,
		GasPrice: parm.GasPrice,
		Note:     parm.Note,
		Alias:    parm.Alias,
	}
	tx := &types.Transaction{}
	if parm.IsCreate {
		tx = &types.Transaction{
			Execer:  []byte(types.ExecName(ExecutorName)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(types.ExecName(ExecutorName)),
		}
	} else {
		tx = &types.Transaction{
			Execer:  []byte(types.ExecName(parm.Name)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(types.ExecName(parm.Name)),
		}
	}
	tx, err = types.FormatTx(string(tx.Execer), tx)
	if err != nil {
		return nil, err
	}
	if tx.Fee < parm.Fee {
		tx.Fee += parm.Fee
	}
	return tx, nil
}

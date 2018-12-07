// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"strings"

	"errors"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/golang/protobuf/proto"
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

	types.RegisterDappFork(ExecutorName, EVMEnable, 500000)
	// EVM合约中的数据分散存储，支持大数据量
	types.RegisterDappFork(ExecutorName, ForkEVMState, 650000)
	// EVM合约状态数据生成哈希，保存在主链的StateDB中
	types.RegisterDappFork(ExecutorName, ForkEVMKVHash, 1000000)
	// EVM合约支持ABI绑定和调用
	types.RegisterDappFork(ExecutorName, ForkEVMABI, 1250000)
	// EEVM合约用户金额冻结
	types.RegisterDappFork(ExecutorName, ForkEVMFrozen, 1300000)
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

// GetName 获取执行器名称
func (evm *EvmType) GetName() string {
	return ExecutorName
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
		return createEvmTx(&param)
	}
	return nil, types.ErrNotSupport
}

// GetLogMap 获取日志类型映射
func (evm *EvmType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

func createEvmTx(param *CreateCallTx) (*types.Transaction, error) {
	if param == nil {
		elog.Error("createEvmTx", "param", param)
		return nil, types.ErrInvalidParam
	}

	// 调用格式判断规则：
	// 十六进制格式默认使用原方式调用，其它格式，使用ABI方式调用
	// 为了方便区分，在ABI格式前加0x00000000

	action := &EVMContractAction{
		Amount:   param.Amount,
		GasLimit: param.GasLimit,
		GasPrice: param.GasPrice,
		Note:     param.Note,
		Alias:    param.Alias,
	}
	// Abi数据和二进制代码必须指定一个，优先判断ABI
	if len(param.Abi) > 0 {
		action.Abi = strings.TrimSpace(param.Abi)
	}
	if len(param.Code) > 0 {
		bCode, err := common.FromHex(param.Code)
		if err != nil {
			elog.Error("create evm Tx error, code is invalid", "param.Code", param.Code)
			return nil, err
		}
		action.Code = bCode
	}

	if param.IsCreate {
		if len(action.Abi) > 0 && len(action.Code) == 0 {
			elog.Error("create evm Tx error, code is empty")
			return nil, errors.New("code must be set in create tx")
		}

		return createRawTx(action, "", param.Fee)
	}
	return createRawTx(action, param.Name, param.Fee)
}

func createRawTx(action proto.Message, name string, fee int64) (*types.Transaction, error) {
	tx := &types.Transaction{}
	if len(name) == 0 {
		tx = &types.Transaction{
			Execer:  []byte(types.ExecName(ExecutorName)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(types.ExecName(ExecutorName)),
		}
	} else {
		tx = &types.Transaction{
			Execer:  []byte(types.ExecName(name)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(types.ExecName(name)),
		}
	}
	tx, err := types.FormatTx(string(tx.Execer), tx)
	if err != nil {
		return nil, err
	}

	if tx.Fee < fee {
		tx.Fee = fee
	}

	return tx, nil
}

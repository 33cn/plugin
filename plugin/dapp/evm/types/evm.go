// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
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
		"Exec":  EvmExecAction,
		"Update": EvmUpdateAction,
		"Destroy": EvmDestroyAction,
		"Freeze":  EvmFreezeAction,
		"Release": EvmReleaseAction,
	}
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, ExecerEvm)
	types.RegFork(ExecutorName, InitFork)
	types.RegExec(ExecutorName, InitExecutor)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(ExecutorName, EVMEnable, 500000)
	// EVM合约中的数据分散存储，支持大数据量
	cfg.RegisterDappFork(ExecutorName, ForkEVMState, 650000)
	// EVM合约状态数据生成哈希，保存在主链的StateDB中
	cfg.RegisterDappFork(ExecutorName, ForkEVMKVHash, 1000000)
	// EVM合约支持ABI绑定和调用
	cfg.RegisterDappFork(ExecutorName, ForkEVMABI, 1250000)
	// EEVM合约用户金额冻结
	cfg.RegisterDappFork(ExecutorName, ForkEVMFrozen, 1300000)
	// EEVM 黄皮v1分叉高度
	cfg.RegisterDappFork(ExecutorName, ForkEVMYoloV1, 9500000)
	// EVM合约支持交易组
	cfg.RegisterDappFork(ExecutorName, ForkEVMTxGroup, 0)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(ExecutorName, NewType(cfg))
}

// EvmType EVM类型定义
type EvmType struct {
	types.ExecTypeBase
}

// NewType 新建EVM类型对象
func NewType(cfg *types.Chain33Config) *EvmType {
	c := &EvmType{}
	c.SetChild(c)
	c.SetConfig(cfg)
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
	if action == "Exec" {
		var param CreateCallTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return createEvmTx(evm.GetConfig(), &param)
	} else if action == "Update" {
		var param UpdateTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("UpdateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		return createEvmUpdateTx(evm.GetConfig(), &param)
	} else if action == "Destroy" {
		var param DestroyTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("DestroyTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		v := &EVMContractDestroy{
			Addr:     param.Addr,
		}
		destroy := &EVMContractAction{
			Ty:    EvmDestroyAction,
			Value: &EVMContractAction_Destroy{v},
		}

		return createRawTx(evm.GetConfig(), destroy, "", param.Fee)
	} else if action == "Freeze" {
		var param FreezeTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("FreezeTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		v := &EVMContractFreeze{
			Addr:     param.Addr,
		}
		freeze := &EVMContractAction{
			Ty:    EvmFreezeAction,
			Value: &EVMContractAction_Freeze{v},
		}
		return createRawTx(evm.GetConfig(), freeze, "", param.Fee)
	}  else if action == "Release" {
		var param ReleaseTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			elog.Error("ReleaseTx", "Error", err)
			return nil, types.ErrInvalidParam
		}

		v := &EVMContractRelease{
			Addr:     param.Addr,
		}
		release := &EVMContractAction{
			Ty:    EvmReleaseAction,
			Value: &EVMContractAction_Release{v},
		}
		return createRawTx(evm.GetConfig(), release, "", param.Fee)
	}

	return nil, types.ErrNotSupport
}

// GetLogMap 获取日志类型映射
func (evm *EvmType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

func createEvmTx(cfg *types.Chain33Config, param *CreateCallTx) (*types.Transaction, error) {
	if param == nil {
		elog.Error("createEvmTx", "param", param)
		return nil, types.ErrInvalidParam
	}

	// 调用格式判断规则：
	// 十六进制格式默认使用原方式调用，其它格式，使用ABI方式调用
	// 为了方便区分，在ABI格式前加0x00000000

	action := &EVMContractExec{
		Amount:   param.Amount,
		GasLimit: param.GasLimit,
		GasPrice: param.GasPrice,
		Note:     param.Note,
		Alias:    param.Alias,
	}
	if len(param.Code) > 0 {
		bCode, err := common.FromHex(param.Code)
		if err != nil {
			elog.Error("create evm Tx error, code is invalid", "param.Code", param.Code)
			return nil, err
		}
		action.Code = bCode
	}

	if len(param.Para) > 0 {
		para, err := common.FromHex(param.Para)
		if err != nil {
			elog.Error("create evm Tx error, code is invalid", "param.Code", param.Code)
			return nil, err
		}
		action.Para = para
	}

	if param.IsCreate {
		if len(action.Code) == 0 {
			elog.Error("create evm Tx error, code is empty")
			return nil, errors.New("code must be set in create tx")
		}

		return createRawTx(cfg, action, "", param.Fee)
	}
	return createRawTx(cfg, action, param.Name, param.Fee)
}

func createEvmUpdateTx(cfg *types.Chain33Config, param *UpdateTx) (*types.Transaction, error) {
	if param == nil {
		elog.Error("createEvmUpdateTx", "param", param)
		return nil, types.ErrInvalidParam
	}

	// 调用格式判断规则：
	// 十六进制格式默认使用原方式调用，其它格式，使用ABI方式调用
	// 为了方便区分，在ABI格式前加0x00000000

	action := &EVMContractUpdate{
		Addr:     param.Addr,
		Amount:   param.Amount,
		GasLimit: param.GasLimit,
		GasPrice: param.GasPrice,
		Note:     param.Note,
		Alias:    param.Alias,
	}
	if len(param.Code) > 0 {
		bCode, err := common.FromHex(param.Code)
		if err != nil {
			elog.Error("create evm Tx error, code is invalid", "param.Code", param.Code)
			return nil, err
		}
		action.Code = bCode
	}

	return createRawTx(cfg, action, "", param.Fee)
}

func createRawTx(cfg *types.Chain33Config, action proto.Message, name string, fee int64) (*types.Transaction, error) {
	tx := &types.Transaction{}
	if len(name) == 0 {
		tx = &types.Transaction{
			Execer:  []byte(cfg.ExecName(ExecutorName)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(cfg.ExecName(ExecutorName)),
		}
	} else {
		tx = &types.Transaction{
			Execer:  []byte(cfg.ExecName(name)),
			Payload: types.Encode(action),
			To:      address.ExecAddress(cfg.ExecName(name)),
		}
	}
	tx, err := types.FormatTx(cfg, string(tx.Execer), tx)
	if err != nil {
		return nil, err
	}

	if tx.Fee < fee {
		tx.Fee = fee
	}

	return tx, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"strings"

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
	types.RegFork(ExecutorName, InitFork)
	types.RegExec(ExecutorName, InitExecutor)
}

// InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(ExecutorName, EVMEnable, 0)
	// EVM合约中的数据分散存储，支持大数据量
	cfg.RegisterDappFork(ExecutorName, ForkEVMState, 0)
	// EVM合约状态数据生成哈希，保存在主链的StateDB中
	cfg.RegisterDappFork(ExecutorName, ForkEVMKVHash, 0)
	// EVM合约支持ABI绑定和调用
	cfg.RegisterDappFork(ExecutorName, ForkEVMABI, 0)
	// EEVM合约用户金额冻结
	cfg.RegisterDappFork(ExecutorName, ForkEVMFrozen, 0)
	// EEVM 黄皮v1分叉高度
	cfg.RegisterDappFork(ExecutorName, ForkEVMYoloV1, 0)
	// EVM合约支持交易组
	cfg.RegisterDappFork(ExecutorName, ForkEVMTxGroup, 0)
	cfg.RegisterDappFork(ExecutorName, ForkEVMMixAddress, 0)
	cfg.RegisterDappFork(ExecutorName, ForkIntrinsicGas, 0)
	cfg.RegisterDappFork(ExecutorName, ForkEVMAddressInit, 0)
	cfg.RegisterDappFork(ExecutorName, ForkEvmExecNonce, 0)
	cfg.RegisterDappFork(ExecutorName, ForkEvmExecNonceV2, 0)

}

// InitExecutor ...
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

// ActionName 获取ActionName
func (evm EvmType) ActionName(tx *types.Transaction) string {
	// 这个需要通过合约交易目标地址来判断Action
	// 如果目标地址为空，或为evm的固定合约地址，则为创建合约，否则为调用合约
	cfg := evm.GetConfig()

	var action EVMContractAction
	err := types.Decode(tx.Payload, &action)
	if err == nil {
		if strings.EqualFold(tx.To, address.ExecAddress(cfg.ExecName(ExecutorName))) && len(action.Code) > 0 {
			return "createEvmContract"
		}
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

// GetLogMap 获取日志类型映射
func (evm *EvmType) GetLogMap() map[int64]*types.LogInfo {
	return logInfo
}

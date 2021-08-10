// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

const (
	// EvmCreateAction 创建或调用合约
	EvmExecAction = 1
	// EvmDestroyAction 销毁合约
	EvmUpdateAction = 2
	// EvmDestroyAction 销毁合约
	EvmDestroyAction = 3
	// EvmFreezeAction 冻结合约
	EvmFreezeAction = 4
	// EvmReleaseAction 解冻合约
	EvmReleaseAction = 5

	// TyLogContractData  合约代码变更日志
	TyLogContractData = 601
	// TyLogContractState  合约状态数据变更日志
	TyLogContractState = 602
	// TyLogCallContract  合约状态数据变更日志
	TyLogCallContract = 603
	// TyLogEVMStateChangeItem  合约状态数据变更项日志
	TyLogEVMStateChangeItem = 604
	// TyLogEVMEventData 合约生成新的event日志数据
	TyLogEVMEventData = 605
	// TyLogEVMStatisticDataInit 合约统计数据初始化日志
	TyLogEVMStatisticDataInit = 606
	// TyLogEVMStatisticData 合约统计数据变更日志
	TyLogEVMStatisticData = 607

	// MaxGasLimit  最大Gas消耗上限
	MaxGasLimit = 10000000
)

const (
	// EVMEnable 启用EVM
	EVMEnable = "Enable"
	// ForkEVMState EVM合约中的数据分散存储，支持大数据量
	ForkEVMState = "ForkEVMState"
	// ForkEVMKVHash EVM合约状态数据生成哈希，保存在主链的StateDB中
	ForkEVMKVHash = "ForkEVMKVHash"
	// ForkEVMABI EVM合约支持ABI绑定和调用
	ForkEVMABI = "ForkEVMABI"
	// ForkEVMFrozen EVM合约用户金额冻结
	ForkEVMFrozen = "ForkEVMFrozen"
	// ForkEVMYoloV1 YoloV1虚拟机指令分叉
	ForkEVMYoloV1 = "ForkEVMYoloV1"
	//ForkEVMTxGroup 交易组中的交易通过GAS检查
	ForkEVMTxGroup = "ForkEVMTxGroup"
)

var (
	// EvmPrefix  本执行器前缀
	EvmPrefix = "user.evm."
	// ExecutorName  本执行器名称
	ExecutorName = "evm"

	// ExecerEvm EVM执行器名称
	ExecerEvm = []byte(ExecutorName)
	// UserPrefix 执行器前缀
	UserPrefix = []byte(EvmPrefix)

	logInfo = map[int64]*types.LogInfo{
		TyLogCallContract:         {Ty: reflect.TypeOf(ReceiptEVMContract{}), Name: "LogCallContract"},
		TyLogContractData:         {Ty: reflect.TypeOf(EVMContractData{}), Name: "LogContractData"},
		TyLogContractState:        {Ty: reflect.TypeOf(EVMContractState{}), Name: "LogContractState"},
		TyLogEVMStateChangeItem:   {Ty: reflect.TypeOf(EVMStateChangeItem{}), Name: "LogEVMStateChangeItem"},
		TyLogEVMStatisticDataInit: {Ty: reflect.TypeOf(ReceiptEvmStatistic{}), Name: "LogEVMStatisticDataInit"},
		TyLogEVMStatisticData:     {Ty: reflect.TypeOf(ReceiptEvmStatistic{}), Name: "LogEVMStatisticData"},
	}
)

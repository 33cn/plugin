// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

const (
	EvmCreateAction = 1
	EvmCallAction   = 2

	// 合约代码变更日志
	TyLogContractData = 601
	// 合约状态数据变更日志
	TyLogContractState = 602
	// 合约状态数据变更日志
	TyLogCallContract = 603
	// 合约状态数据变更项日志
	TyLogEVMStateChangeItem = 604

	// 查询方法
	FuncCheckAddrExists = "CheckAddrExists"
	FuncEstimateGas     = "EstimateGas"
	FuncEvmDebug        = "EvmDebug"

	// 最大Gas消耗上限
	MaxGasLimit = 10000000
)

var (
	// 本执行器前缀
	EvmPrefix = "user.evm."
	// 本执行器名称
	ExecutorName = "evm"

	ExecerEvm  = []byte(ExecutorName)
	UserPrefix = []byte(EvmPrefix)

	logInfo = map[int64]*types.LogInfo{
		TyLogCallContract:       {Ty: reflect.TypeOf(ReceiptEVMContract{}), Name: "LogCallContract"},
		TyLogContractData:       {Ty: reflect.TypeOf(EVMContractData{}), Name: "LogContractData"},
		TyLogContractState:      {Ty: reflect.TypeOf(EVMContractState{}), Name: "LogContractState"},
		TyLogEVMStateChangeItem: {Ty: reflect.TypeOf(EVMStateChangeItem{}), Name: "LogEVMStateChangeItem"},
	}
)

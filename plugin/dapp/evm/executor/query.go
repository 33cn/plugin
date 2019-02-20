// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"math/big"
	"strings"

	"errors"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// Query_CheckAddrExists 检查合约地址是否存在，此操作不会改变任何状态，所以可以直接从statedb查询
func (evm *EVMExecutor) Query_CheckAddrExists(in *evmtypes.CheckEVMAddrReq) (types.Message, error) {
	evm.CheckInit()
	addrStr := in.Addr
	if len(addrStr) == 0 {
		return nil, model.ErrAddrNotExists
	}

	var addr common.Address
	// 合约名称
	if strings.HasPrefix(addrStr, types.ExecName(evmtypes.EvmPrefix)) {
		addr = common.ExecAddress(addrStr)
	} else {
		// 合约地址
		nAddr := common.StringToAddress(addrStr)
		if nAddr == nil {
			return nil, model.ErrAddrNotExists
		}
		addr = *nAddr
	}

	exists := evm.GetMStateDB().Exist(addr.String())
	ret := &evmtypes.CheckEVMAddrResp{Contract: exists}
	if exists {
		account := evm.GetMStateDB().GetAccount(addr.String())
		if account != nil {
			ret.ContractAddr = account.Addr
			ret.ContractName = account.GetExecName()
			ret.AliasName = account.GetAliasName()
		}
	}
	return ret, nil
}

// Query_EstimateGas 此方法用来估算合约消耗的Gas，不能修改原有执行器的状态数据
func (evm *EVMExecutor) Query_EstimateGas(in *evmtypes.EstimateEVMGasReq) (types.Message, error) {
	evm.CheckInit()
	var (
		caller common.Address
	)

	// 如果未指定调用地址，则直接使用一个虚拟的地址发起调用
	if len(in.Caller) > 0 {
		callAddr := common.StringToAddress(in.Caller)
		if callAddr != nil {
			caller = *callAddr
		}
	} else {
		caller = common.ExecAddress(types.ExecName(evmtypes.ExecutorName))
	}

	to := common.StringToAddress(in.To)
	if to == nil {
		to = common.StringToAddress(EvmAddress)
	}
	msg := common.NewMessage(caller, to, 0, in.Amount, evmtypes.MaxGasLimit, 1, in.Code, "estimateGas", in.Abi)
	txHash := common.BigToHash(big.NewInt(evmtypes.MaxGasLimit)).Bytes()

	receipt, err := evm.innerExec(msg, txHash, 1, evmtypes.MaxGasLimit, false)
	if err != nil {
		return nil, err
	}

	if receipt.Ty == types.ExecOk {
		callData := getCallReceipt(receipt.GetLogs())
		if callData != nil {
			result := &evmtypes.EstimateEVMGasResp{}
			result.Gas = callData.UsedGas
			return result, nil
		}
	}
	return nil, errors.New("contract call error")
}

// 从日志中查找调用结果
func getCallReceipt(logs []*types.ReceiptLog) *evmtypes.ReceiptEVMContract {
	if len(logs) == 0 {
		return nil
	}
	for _, v := range logs {
		if v.Ty == evmtypes.TyLogCallContract {
			var res evmtypes.ReceiptEVMContract
			err := types.Decode(v.Log, &res)
			if err != nil {
				return nil
			}
			return &res
		}
	}
	return nil
}

// Query_EvmDebug 此方法用来估算合约消耗的Gas，不能修改原有执行器的状态数据
func (evm *EVMExecutor) Query_EvmDebug(in *evmtypes.EvmDebugReq) (types.Message, error) {
	evm.CheckInit()
	optype := in.Optype

	if optype < 0 {
		evmDebug = false
	} else if optype > 0 {
		evmDebug = true
	}
	ret := &evmtypes.EvmDebugResp{DebugStatus: fmt.Sprintf("%v", evmDebug)}
	return ret, nil
}

// Query_Query 此方法用来调用合约的只读接口，不修改原有执行器的状态数据
func (evm *EVMExecutor) Query_Query(in *evmtypes.EvmQueryReq) (types.Message, error) {
	evm.CheckInit()

	ret := &evmtypes.EvmQueryResp{}
	ret.Address = in.Address
	ret.Input = in.Input
	ret.Caller = in.Caller

	var (
		caller common.Address
	)

	to := common.StringToAddress(in.Address)
	if to == nil {
		ret.JsonData = fmt.Sprintf("invalid address:%v", in.Address)
		return ret, nil
	}

	// 如果未指定调用地址，则直接使用一个虚拟的地址发起调用
	if len(in.Caller) > 0 {
		callAddr := common.StringToAddress(in.Caller)
		if callAddr != nil {
			caller = *callAddr
		}
	} else {
		caller = common.ExecAddress(types.ExecName(evmtypes.ExecutorName))
	}

	msg := common.NewMessage(caller, common.StringToAddress(in.Address), 0, 0, evmtypes.MaxGasLimit, 1, nil, "estimateGas", in.Input)
	txHash := common.BigToHash(big.NewInt(evmtypes.MaxGasLimit)).Bytes()

	receipt, err := evm.innerExec(msg, txHash, 1, evmtypes.MaxGasLimit, true)
	if err != nil {
		ret.JsonData = fmt.Sprintf("%v", err)
		return ret, nil
	}
	if receipt.Ty == types.ExecOk {
		callData := getCallReceipt(receipt.GetLogs())
		if callData != nil {
			ret.RawData = common.Bytes2Hex(callData.Ret)
			ret.JsonData = callData.JsonRet
			return ret, nil
		}
	}
	return ret, nil
}

// Query_QueryABI 此方法用来查询合约绑定的ABI数据，不修改原有执行器的状态数据
func (evm *EVMExecutor) Query_QueryABI(in *evmtypes.EvmQueryAbiReq) (types.Message, error) {
	evm.CheckInit()

	addr := common.StringToAddress(in.GetAddress())
	if addr == nil {
		return nil, fmt.Errorf("invalid address: %v", in.GetAddress())
	}

	abiData := evm.mStateDB.GetAbi(addr.String())

	return &evmtypes.EvmQueryAbiResp{Address: in.GetAddress(), Abi: abiData}, nil
}

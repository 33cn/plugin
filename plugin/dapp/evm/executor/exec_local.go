// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/model"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// ExecLocal 处理本地区块新增逻辑
func (evm *EVMExecutor) ExecLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := evm.DriverBase.ExecLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}

	// 调用失败信息统计
	if receipt.GetTy() != types.ExecOk {
		action := evm.GetActionName(tx)
		if action == "exec" {
			for _, logItem := range receipt.Logs {
				var action evmtypes.EVMContractAction
				err := types.Decode(tx.Payload, &action)
				if err != nil {
					return nil, err
				}

				lcdbKey := GetStatisticKey(action.GetExec().ContractAddr)
				evmStatData, err := evm.GetLocalDB().Get(lcdbKey)
				if err != nil {
					continue
				}

				var evmstat evmtypes.EVMContractStatistic
				err = types.Decode(evmStatData, &evmstat)
				if err != nil {
					continue
				}

				evmstat.CallTimes++
				failReason := string(logItem.Log)
				if failReason == model.ErrOutOfGas.Error() {
					evmstat.FailReason[model.StatisticGasError]++
				} else if failReason == model.ErrFrozen.Error() || failReason == model.ErrDestruct.Error() ||
					failReason == model.ErrDepth.Error() || failReason == model.ErrInsufficientBalance.Error() {
					evmstat.FailReason[model.StatisticExecError]++
				} else {
					evmstat.FailReason[model.StatisticEVMError]++
				}

				set.KV = append(set.KV, &types.KeyValue{Key: GetStatisticKey(action.GetExec().ContractAddr), Value: types.Encode(&evmstat)})
			}
		}
		return set, nil
	}

	// 调用统计信息记录
	for _, logItem := range receipt.Logs {
		if evmtypes.TyLogEVMStatisticDataInit == logItem.Ty { // 调用统计信息初始化
			data := logItem.Log
			var changeItem evmtypes.ReceiptEvmStatistic
			err = types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}
			var evmstat evmtypes.EVMContractStatistic
			evmstat.Caller = make([]string, 0)
			evmstat.FailReason = make(map[string]uint64)
			evmstat.FailReason[model.StatisticEVMError] = 0
			evmstat.FailReason[model.StatisticExecError] = 0
			evmstat.FailReason[model.StatisticGasError] = 0
			evmstat.PrevAddr = changeItem.PreAddr
			set.KV = append(set.KV, &types.KeyValue{Key: GetStatisticKey(changeItem.Addr), Value: types.Encode(&evmstat)})
		} else if evmtypes.TyLogEVMStatisticData == logItem.Ty { // 调用统计信息记录
			data := logItem.Log
			var changeItem evmtypes.ReceiptEvmStatistic
			err = types.Decode(data, &changeItem)
			if err != nil {
				return set, err
			}

			lcdbKey := GetStatisticKey(changeItem.Addr)
			evmStatData, err := evm.GetLocalDB().Get(lcdbKey)
			if err != nil {
				return set, err
			}
			var evmstat evmtypes.EVMContractStatistic
			err = types.Decode(evmStatData, &evmstat)
			if err != nil {
				return set, err
			}
			evmstat.CallTimes++
			evmstat.SuccseccTimes++

			var exist bool
			for _, callerAddr := range evmstat.Caller {
				if callerAddr == changeItem.Caller {
					exist = true
					break
				}
			}
			if !exist {
				evmstat.Caller = append(evmstat.Caller, changeItem.Caller)
			}
			set.KV = append(set.KV, &types.KeyValue{Key: lcdbKey, Value: types.Encode(&evmstat)})
		}
	}

	cfg := evm.GetAPI().GetConfig()
	if cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMState) {
		// 需要将Exec中生成的合约状态变更信息写入localdb
		for _, logItem := range receipt.Logs {
			if evmtypes.TyLogEVMStateChangeItem == logItem.Ty {
				data := logItem.Log
				var changeItem evmtypes.EVMStateChangeItem
				err = types.Decode(data, &changeItem)
				if err != nil {
					return set, err
				}
				//转换老的log的key-> 新的key
				key := []byte(changeItem.Key)
				if bytes.HasPrefix(key, []byte("mavl-")) {
					key[0] = 'L'
					key[1] = 'O'
					key[2] = 'D'
					key[3] = 'B'
				}
				set.KV = append(set.KV, &types.KeyValue{Key: key, Value: changeItem.CurrentValue})
			}
		}
	}
	set.KV = evm.AddRollbackKV(tx, []byte(evmtypes.ExecutorName), set.KV)
	return set, err
}

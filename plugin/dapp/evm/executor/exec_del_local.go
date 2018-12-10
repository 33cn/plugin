// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// ExecDelLocal 处理区块回滚
func (evm *EVMExecutor) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := evm.DriverBase.ExecDelLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}

	if types.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMState) {
		// 需要将Exec中生成的合约状态变更信息从localdb中恢复
		for _, logItem := range receipt.Logs {
			if evmtypes.TyLogEVMStateChangeItem == logItem.Ty {
				data := logItem.Log
				var changeItem evmtypes.EVMStateChangeItem
				err = types.Decode(data, &changeItem)
				if err != nil {
					return set, err
				}
				set.KV = append(set.KV, &types.KeyValue{Key: []byte(changeItem.Key), Value: changeItem.PreValue})
			}
		}
	}

	return set, err
}

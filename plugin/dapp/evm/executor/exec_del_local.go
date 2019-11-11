// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
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
	cfg := evm.GetAPI().GetConfig()
	if cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMState) {
		// 需要将Exec中生成的合约状态变更信息从localdb中恢复
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
				set.KV = append(set.KV, &types.KeyValue{Key: []byte(key), Value: changeItem.PreValue})
			}
		}
	}

	return set, err
}

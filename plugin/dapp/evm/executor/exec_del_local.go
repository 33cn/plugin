// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// ExecDelLocal 处理区块回滚
func (evm *EVMExecutor) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := evm.DriverBase.ExecDelLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}

	// 以太坊类型交易, 直接调用自动回滚处理
	if types.IsEthSignID(tx.GetSignature().GetTy()) {

		kvs, err := evm.DelRollbackKV(tx, []byte(evmtypes.ExecutorName))
		if err != nil {
			elog.Error("ExecDelLocal", "txHash", common.ToHex(tx.Hash()), "err", err)
			return nil, err
		}
		set.KV = kvs
		return set, nil
	}

	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	cfg := evm.GetAPI().GetConfig()
	if cfg.IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEVMState) {
		kvs, err := evm.DelRollbackKV(tx, []byte(evmtypes.ExecutorName))
		if err != nil {
			return nil, err
		}
		set.KV = kvs
	}
	return set, err
}

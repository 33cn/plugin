// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// ExecDelLocal 处理区块回滚
func (evm *EVMExecutor) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := evm.DriverBase.ExecDelLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}

	defer func(lSet *types.LocalDBSet) {
		if types.IsEthSignID(tx.GetSignature().GetTy()) {
			nonceLocalKey := secp256k1eth.CaculCoinsEvmAccountKey(tx.From())
			nonceV, err := evm.GetLocalDB().Get(nonceLocalKey)
			if err == nil {
				var evmNonce types.EvmAccountNonce
				types.Decode(nonceV, &evmNonce)
				if evmNonce.GetNonce() == tx.GetNonce()+1 {
					evmNonce.Nonce--
					if evmNonce.GetNonce() < 0 {
						evmNonce.Nonce = 0
					}
					if lSet != nil {
						lSet.KV = append(lSet.KV, &types.KeyValue{Key: nonceLocalKey, Value: types.Encode(&evmNonce)})
					}
				}

			}

		}
	}(set)

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

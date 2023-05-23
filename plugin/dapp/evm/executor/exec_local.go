// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"errors"

	"github.com/33cn/chain33/common"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/system/crypto/secp256k1eth"
	"github.com/33cn/chain33/types"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
)

var (
	errInvalidEvmNonce = errors.New("errInvalidEvmNonce")
)

func (evm *EVMExecutor) handleEvmNonce(tx *types.Transaction, dbSet *types.LocalDBSet) error {

	if types.IsEthSignID(tx.GetSignature().GetTy()) {
		fromAddr := tx.From()
		nonceLocalKey := secp256k1eth.CaculCoinsEvmAccountKey(fromAddr)
		evmNonce := &types.EvmAccountNonce{}
		nonceV, nonceErr := evm.GetLocalDB().Get(nonceLocalKey)
		if nonceErr == nil {
			_ = types.Decode(nonceV, evmNonce)
		}

		log.Debug("ExecLocal handleEvmNonce", "height", evm.GetHeight(), "txHash", common.ToHex(tx.Hash()),
			"from", fromAddr, "expect", evmNonce.GetNonce(), "actual", tx.GetNonce())

		if evm.GetAPI().GetConfig().IsDappFork(evm.GetHeight(), "evm", evmtypes.ForkEvmExecNonce) &&
			evmNonce.GetNonce() != tx.GetNonce() { //nonce 错误 返回异常
			return errInvalidEvmNonce
		}
		evmNonce.Addr = fromAddr
		evmNonce.Nonce++
		if dbSet != nil {
			dbSet.KV = append(dbSet.KV, &types.KeyValue{Key: nonceLocalKey, Value: types.Encode(evmNonce)})
		}
	}
	return nil
}

// ExecLocal 处理本地区块新增逻辑
func (evm *EVMExecutor) ExecLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (set *types.LocalDBSet, err error) {
	set, err = evm.DriverBase.ExecLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}

	// 校验及设置evm nonce
	if err = evm.handleEvmNonce(tx, set); err != nil {
		return nil, err
	}

	if receipt.GetTy() != types.ExecOk {
		// 失败交易也需要记录evm nonce, 增加自动回滚处理
		if types.IsEthSignID(tx.GetSignature().GetTy()) {
			set.KV = evm.AddRollbackKV(tx, []byte(evmtypes.ExecutorName), set.KV)
		}
		return set, nil
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

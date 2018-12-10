// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

//Exec_MultiSigAccCreate 创建多重签名账户
func (m *MultiSig) Exec_MultiSigAccCreate(payload *mty.MultiSigAccCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigAccCreate(payload)
}

//Exec_MultiSigOwnerOperate 多重签名账户owner属性的修改：owner的add/del/replace等
func (m *MultiSig) Exec_MultiSigOwnerOperate(payload *mty.MultiSigOwnerOperate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigOwnerOperate(payload)
}

//Exec_MultiSigAccOperate 多重签名账户属性的修改：weight权重以及每日限额的修改
func (m *MultiSig) Exec_MultiSigAccOperate(payload *mty.MultiSigAccOperate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigAccOperate(payload)
}

//Exec_MultiSigConfirmTx 多重签名账户上交易的确认和撤销
func (m *MultiSig) Exec_MultiSigConfirmTx(payload *mty.MultiSigConfirmTx, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigConfirmTx(payload)
}

//Exec_MultiSigExecTransferTo 合约中外部账户转账到多重签名账户，Addr --->multiSigAddr
func (m *MultiSig) Exec_MultiSigExecTransferTo(payload *mty.MultiSigExecTransferTo, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigExecTransferTo(payload)
}

//Exec_MultiSigExecTransferFrom 合约中多重签名账户转账到外部账户，multiSigAddr--->Addr
func (m *MultiSig) Exec_MultiSigExecTransferFrom(payload *mty.MultiSigExecTransferFrom, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(m, tx, int32(index))
	return action.MultiSigExecTransferFrom(payload)
}

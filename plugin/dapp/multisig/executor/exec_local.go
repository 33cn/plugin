// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

//ExecLocal_MultiSigAccCreate 创建多重签名账户,根据payload和receiptData信息获取相关信息并保存到db中
func (m *MultiSig) ExecLocal_MultiSigAccCreate(payload *mty.MultiSigAccCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.execLocalMultiSigReceipt(receiptData, tx, true)
	if err != nil {
		multisiglog.Error("ExecLocal_MultiSigAccCreate", "err", err)
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

//ExecLocal_MultiSigOwnerOperate 多重签名账户owner属性的修改：owner的add/del/replace/modify等
func (m *MultiSig) ExecLocal_MultiSigOwnerOperate(payload *mty.MultiSigOwnerOperate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.execLocalMultiSigReceipt(receiptData, tx, true)
	if err != nil {
		multisiglog.Error("ExecLocal_MultiSigOwnerOperate", "err", err)
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

//ExecLocal_MultiSigAccOperate 多重签名账户属性的修改：weight权重以及每日限额的修改
func (m *MultiSig) ExecLocal_MultiSigAccOperate(payload *mty.MultiSigAccOperate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.execLocalMultiSigReceipt(receiptData, tx, true)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

//ExecLocal_MultiSigConfirmTx 多重签名账户上交易的确认和撤销
func (m *MultiSig) ExecLocal_MultiSigConfirmTx(payload *mty.MultiSigConfirmTx, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.execLocalMultiSigReceipt(receiptData, tx, true)
	if err != nil {
		multisiglog.Error("ExecLocal_MultiSigConfirmTx", "err", err)
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

//ExecLocal_MultiSigExecTransferTo 合约中外部账户转账到多重签名账户，Addr --->multiSigAddr
func (m *MultiSig) ExecLocal_MultiSigExecTransferTo(payload *mty.MultiSigExecTransferTo, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.saveMultiSigTransfer(tx, mty.IsSubmit, true)
	if err != nil {
		multisiglog.Error("ExecLocal_MultiSigExecTransferTo", "err", err)
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

//ExecLocal_MultiSigExecTransferFrom 合约中多重签名账户转账到外部账户，multiSigAddr--->Addr
func (m *MultiSig) ExecLocal_MultiSigExecTransferFrom(payload *mty.MultiSigExecTransferFrom, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.GetTy() != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}

	kv, err := m.execLocalMultiSigReceipt(receiptData, tx, true)
	if err != nil {
		multisiglog.Error("ExecLocal_MultiSigExecTransferFrom", "err", err)
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

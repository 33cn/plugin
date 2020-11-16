package executor

import (
	"github.com/33cn/chain33/types"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
)

// ExecDelLocal_CreateJvmContract 本地撤销执行创建Jvm合约
func (jvm *JVMExecutor) ExecDelLocal_CreateJvmContract(payload *jvmTypes.CreateJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execDelLocal(tx, receipt, index)
}

// ExecDelLocal_CallJvmContract 本地撤销执行调用Jvm合约
func (jvm *JVMExecutor) ExecDelLocal_CallJvmContract(payload *jvmTypes.CallJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execDelLocal(tx, receipt, index)
}

// ExecDelLocal_UpdateJvmContract 本地撤销执行更新Jvm合约
func (jvm *JVMExecutor) ExecDelLocal_UpdateJvmContract(payload *jvmTypes.UpdateJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execDelLocal(tx, receipt, index)
}

func (Jvm *JVMExecutor) execDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	kvs, err := Jvm.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	set.KV = kvs
	return set, nil
}

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	jvmState "github.com/33cn/plugin/plugin/dapp/jvm/executor/state"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
)

// ExecLocal_CreateJvmContract 本地执行创建Jvm合约
func (jvm *JVMExecutor) ExecLocal_CreateJvmContract(payload *jvmTypes.CreateJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execLocal(tx, receipt, index)
}

// ExecLocal_CallJvmContract 本地执行调用Jvm合约
func (jvm *JVMExecutor) ExecLocal_CallJvmContract(payload *jvmTypes.CallJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execLocal(tx, receipt, index)
}

// ExecLocal_UpdateJvmContract 本地执行更新Jvm合约
func (jvm *JVMExecutor) ExecLocal_UpdateJvmContract(payload *jvmTypes.UpdateJvmContract, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return jvm.execLocal(tx, receipt, index)
}

func (Jvm *JVMExecutor) execLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	kvs := jvmState.GetAllLocalKeyValues(common.ToHex(tx.Hash()))
	set.KV = Jvm.AddRollbackKV(tx, tx.Execer, kvs)
	return set, nil
}

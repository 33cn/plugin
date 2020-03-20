package executor

import (
	"github.com/33cn/chain33/types"
	accountmanagertypes "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (a *accountmanager) ExecLocal_Register(payload *accountmanagertypes.Register, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return a.addAutoRollBack(tx, dbSet.KV), nil
}
func (a *accountmanager) ExecLocal_Reset(payload *accountmanagertypes.Reset, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

func (a *accountmanager) ExecLocal_Apply(payload *accountmanagertypes.Apply, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

func (a *accountmanager) ExecLocal_Transfer(payload *accountmanagertypes.Transfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

func (a *accountmanager) ExecLocal_Supervise(payload *accountmanagertypes.Supervise, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//设置自动回滚
func (a *accountmanager) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

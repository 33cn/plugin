package executor

import (
	"github.com/33cn/chain33/types"
	storagetypes "github.com/33cn/plugin/plugin/dapp/storage/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (s *storage) ExecLocal_ContentStorage(payload *storagetypes.ContentOnlyNotaryStorage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return s.addAutoRollBack(tx, dbSet.KV), nil
}

func (s *storage) ExecLocal_HashStorage(payload *storagetypes.HashOnlyNotaryStorage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return s.addAutoRollBack(tx, dbSet.KV), nil
}

func (s *storage) ExecLocal_LinkStorage(payload *storagetypes.LinkNotaryStorage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return s.addAutoRollBack(tx, dbSet.KV), nil
}

func (s *storage) ExecLocal_EncryptStorage(payload *storagetypes.EncryptNotaryStorage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return s.addAutoRollBack(tx, dbSet.KV), nil
}

func (s *storage) ExecLocal_EncryptShareStorage(payload *storagetypes.EncryptShareNotaryStorage, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code
	return s.addAutoRollBack(tx, dbSet.KV), nil
}

//设置自动回滚
func (s *storage) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = s.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

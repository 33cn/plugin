package executor

import (
	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (r *rollup) ExecLocal_CommitBatch(payload *rolluptypes.CheckPoint, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	//implement code, add customize kv to dbSet...

	//auto gen for localdb auto rollback
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

//当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (r *rollup) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = r.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

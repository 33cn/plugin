package executor

import (
	"github.com/33cn/chain33/types"
)

/*
 * 实现区块回退时本地执行的数据清除
 */

// ExecDelLocal localdb kv数据自动回滚接口
func (r *rollup) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kvs, err := r.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

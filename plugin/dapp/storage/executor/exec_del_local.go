package executor

import (
	"github.com/33cn/chain33/types"
)

/*
 * 实现区块回退时本地执行的数据清除
 */

// ExecDelLocal 回退自动删除，重写基类
func (s *storage) ExecDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kvs, err := s.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

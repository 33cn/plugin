package executor

import (
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func (z *zksync) execAutoLocalZksync(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receiptData.Ty != types.ExecOk {
		return nil, types.ErrInvalidParam
	}
	set, err := z.execLocalZksync(tx, receiptData, index)
	if err != nil {
		return set, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = z.AddRollbackKV(tx, tx.Execer, set.KV)
	return dbSet, nil
}

func (z *zksync) execLocalZksync(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	table := NewZksyncInfoTable(z.GetLocalDB())

	var zksyncInfo zt.OperationInfo
	err := types.Decode(receiptData.Logs[0].GetLog(), &zksyncInfo)
	if err != nil {
		return nil, err
	}

	err = table.Add(&zksyncInfo)
	if err != nil {
		return nil, err
	}

	kvs, err := table.Save()
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

func (z *zksync) execAutoDelLocal(tx *types.Transaction, receiptData *types.ReceiptData) (*types.LocalDBSet, error) {
	kvs, err := z.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

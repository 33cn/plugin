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
	infoTable := NewZksyncInfoTable(z.GetLocalDB())

	dbSet := &types.LocalDBSet{}
	for _, log := range receiptData.Logs {
		switch log.Ty {
		case
			zt.TyDepositLog,
			zt.TyWithdrawLog,
			zt.TyTreeToContractLog,
			zt.TyContractToTreeLog,
			zt.TyTransferLog,
			zt.TyTransferToNewLog,
			zt.TySetPubKeyLog,
			zt.TyForceExitLog,
			zt.TyFullExitLog,
			zt.TySwapLog:
			var zklog zt.ZkReceiptLog
			err := types.Decode(log.GetLog(), &zklog)
			if err != nil {
				return nil, err
			}
			if zklog.OperationInfo == nil || zklog.OperationInfo.AccountID == 0 {
				continue
			}
			err = infoTable.Replace(zklog.OperationInfo)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, zklog.LocalKvs...)
		}
	}
	kvs, err := infoTable.Save()
	if err != nil {
		return nil, err
	}
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

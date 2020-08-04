package executor

import (
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

//ExecLocal_Register ...
func (a *Accountmanager) ExecLocal_Register(payload *et.Register, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case et.TyRegisterLog:
				receipt := &et.AccountReceipt{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				accountTable := NewAccountTable(a.GetLocalDB())
				err := accountTable.Add(receipt.Account)
				if err != nil {
					return nil, err
				}
				kvs, err := accountTable.Save()
				if err != nil {
					return nil, err
				}
				dbSet.KV = append(dbSet.KV, kvs...)

			}
		}
	}
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//ExecLocal_ResetKey ...
func (a *Accountmanager) ExecLocal_ResetKey(payload *et.ResetKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case et.TyResetLog:
				receipt := &et.AccountReceipt{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				accountTable := NewAccountTable(a.GetLocalDB())
				err := accountTable.Replace(receipt.Account)
				if err != nil {
					return nil, err
				}
				kvs, err := accountTable.Save()
				if err != nil {
					return nil, err
				}
				dbSet.KV = append(dbSet.KV, kvs...)
			}
		}
	}
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//ExecLocal_Apply ...
func (a *Accountmanager) ExecLocal_Apply(payload *et.Apply, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case et.TyApplyLog:
				receipt := &et.AccountReceipt{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				accountTable := NewAccountTable(a.GetLocalDB())
				err := accountTable.Replace(receipt.Account)
				if err != nil {
					return nil, err
				}
				kvs, err := accountTable.Save()
				if err != nil {
					return nil, err
				}
				dbSet.KV = append(dbSet.KV, kvs...)
			}
		}
	}
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//ExecLocal_Transfer ...
func (a *Accountmanager) ExecLocal_Transfer(payload *et.Transfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case et.TyResetLog:
				receipt := &et.TransferReceipt{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				accountTable := NewAccountTable(a.GetLocalDB())
				err := accountTable.Replace(receipt.FromAccount)
				if err != nil {
					return nil, err
				}
				kvs, err := accountTable.Save()
				if err != nil {
					return nil, err
				}
				dbSet.KV = append(dbSet.KV, kvs...)
			}
		}
	}
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//ExecLocal_Supervise ...
func (a *Accountmanager) ExecLocal_Supervise(payload *et.Supervise, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receiptData.Ty == types.ExecOk {
		for _, log := range receiptData.Logs {
			switch log.Ty {
			case et.TySuperviseLog:
				receipt := &et.SuperviseReceipt{}
				if err := types.Decode(log.Log, receipt); err != nil {
					return nil, err
				}
				accountTable := NewAccountTable(a.GetLocalDB())
				//当时续期操作得话，需要重建
				if receipt.Op == et.AddExpire {
					for _, account := range receipt.Accounts {
						err := accountTable.DelRow(account)
						if err != nil {
							return nil, err
						}
						//重置主键
						account.Index = receipt.Index
						err = accountTable.Replace(account)
						if err != nil {
							return nil, err
						}
					}
					kvs, err := accountTable.Save()
					if err != nil {
						return nil, err
					}
					dbSet.KV = append(dbSet.KV, kvs...)
				} else {
					for _, account := range receipt.Accounts {
						err := accountTable.Replace(account)
						if err != nil {
							return nil, err
						}
					}
					kvs, err := accountTable.Save()
					if err != nil {
						return nil, err
					}
					dbSet.KV = append(dbSet.KV, kvs...)
				}

			}
		}
	}
	return a.addAutoRollBack(tx, dbSet.KV), nil
}

//设置自动回滚
func (a *Accountmanager) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {
	dbSet := &types.LocalDBSet{}
	dbSet.KV = a.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

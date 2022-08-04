package executor

import (
	"github.com/33cn/chain33/types"
	ety "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

/*
* Local execution of transaction related data, data is not on the chain
* Non-critical data, local storage (localDB), used for auxiliary query, high efficiency
 */

func (e *zksync) ExecLocal_LimitOrder(payload *ety.SpotLimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_MarketOrder(payload *ety.SpotMarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_RevokeOrder(payload *ety.SpotRevokeOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_EntrustOrder(payload *ety.SpotLimitOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_EntrustRevokeOrder(payload *ety.SpotMarketOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_SpotNftOrder(payload *ety.SpotNftOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_SpotNftTradeOrder(payload *ety.SpotNftTakerOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) ExecLocal_SpotNftTradeOrder2(payload *ety.SpotNftTakerOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}
func (e *zksync) ExecLocal_SpotNftOrder2(payload *ety.SpotNftOrder, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.interExecLocalWithZk(tx, receiptData, index)
}

func (e *zksync) interExecLocalWithZk(tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet, err := e.execLocalZksync(tx, receiptData, index)
	if err != nil {
		return dbSet, err
	}
	action := NewZkSpotDex(e, tx, index)
	set2, err := action.execLocal(tx, receiptData, index)
	if err != nil {
		return dbSet, err
	}
	dbSet.KV = append(dbSet.KV, set2.KV...)

	dbSet = e.addAutoRollBack(tx, dbSet.KV)
	localDB := e.GetLocalDB()
	for _, kv1 := range dbSet.KV {
		//elog.Info("updateIndex", "localDB.Set", string(kv1.Key))
		err := localDB.Set(kv1.Key, kv1.Value)
		if err != nil {
			zlog.Error("updateIndex", "localDB.Set", err.Error())
			return dbSet, err
		}
	}
	return dbSet, nil
}

// Set automatic rollback
func (e *zksync) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {
	dbSet := &types.LocalDBSet{}
	dbSet.KV = e.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

package executor

import (
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (r *rgbx) addPendingTxKV(dbSet *types.LocalDBSet, logData []byte, index int) {

	dbSet.KV = append(dbSet.KV, &types.KeyValue{
		Key:   formatPendingTxKey(r.GetHeight(), int64(index)),
		Value: logData,
	})
}

func (r *rgbx) ExecLocal_Mint(_ *rtypes.MintAsset, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}

	for _, log := range receiptData.Logs {

		if log.Ty == rtypes.TyPendingTxLog {
			r.addPendingTxKV(dbSet, log.Log, index)
		}
	}

	//auto gen for localdb auto rollback
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

func (r *rgbx) ExecLocal_Transfer(_ *rtypes.TransferAsset, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	for _, log := range receiptData.Logs {

		if log.Ty == rtypes.TyPendingTxLog {
			r.addPendingTxKV(dbSet, log.Log, index)
		}
	}

	//auto gen for localdb auto rollback
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

func (r *rgbx) ExecLocal_Confirm(confirm *rtypes.ConfirmTx, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	// remove pending tx record
	if !confirm.Timeout {
		dbSet.KV = append(dbSet.KV,
			&types.KeyValue{Key: formatPendingTxKey(confirm.TxBlockHeight, confirm.TxIndex),
				Value: types.Encode(&rtypes.PendingTx{Confirmed: true})})
	}
	dbSet.KV = append(dbSet.KV, &types.KeyValue{Key: []byte(confirmedHeightKey),
		Value: types.Encode(&types.Int64{Data: confirm.ConfirmedBlockHeight})})

	//auto gen for localdb auto rollback
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

// 当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (r *rgbx) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = r.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

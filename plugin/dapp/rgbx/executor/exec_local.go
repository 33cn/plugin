package executor

import (
	"errors"

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

func (r *rgbx) ExecLocal_Withdraw(_ *rtypes.WithdrawAsset, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	from := tx.From()
	for _, log := range receiptData.Logs {
		if log.Ty == rtypes.TyPendingTxLog {
			r.addPendingTxKV(dbSet, log.Log, index)

			list := &rtypes.TxBlockIndexList{}
			err := readDB(r.GetLocalDB(), formatPendingTxFromKey(from), list)
			if err != nil && !errors.Is(err, types.ErrNotFound) {
				return nil, err
			}
			list.BlockIndexList = append(list.BlockIndexList, &rtypes.TxBlockIndex{
				BlockHeight: r.GetHeight(),
				TxIndex:     int64(index),
			})
			dbSet.KV = append(dbSet.KV, &types.KeyValue{
				Key:   formatPendingTxFromKey(from),
				Value: types.Encode(list),
			})
		}
	}
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

func (r *rgbx) ExecLocal_Deposit(_ *rtypes.DepositAsset, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	for _, log := range receiptData.Logs {
		if log.Ty == rtypes.TyPendingTxLog {
			r.addPendingTxKV(dbSet, log.Log, index)
		}
	}
	return r.addAutoRollBack(tx, dbSet.KV), nil
}

func (r *rgbx) ExecLocal_Confirm(confirm *rtypes.ConfirmTx, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	// remove pending tx record
	if !confirm.Timeout {
		pending := &rtypes.PendingTx{}
		err := readDB(r.GetLocalDB(), formatPendingTxKey(confirm.TxBlockHeight, confirm.TxIndex), pending)
		if err != nil && !errors.Is(err, types.ErrNotFound) {
			return nil, err
		}
		if err == nil && pending.GetFromAddress() != "" {
			list := &rtypes.TxBlockIndexList{}
			err = readDB(r.GetLocalDB(), formatPendingTxFromKey(pending.GetFromAddress()), list)
			if err != nil && !errors.Is(err, types.ErrNotFound) {
				return nil, err
			}
			if err == nil {
				filtered := make([]*rtypes.TxBlockIndex, 0, len(list.GetBlockIndexList()))
				for _, item := range list.GetBlockIndexList() {
					if item.GetBlockHeight() == confirm.GetTxBlockHeight() && item.GetTxIndex() == confirm.GetTxIndex() {
						continue
					}
					filtered = append(filtered, item)
				}
				if len(filtered) == 0 {
					dbSet.KV = append(dbSet.KV, &types.KeyValue{Key: formatPendingTxFromKey(pending.GetFromAddress()), Value: nil})
				} else {
					list.BlockIndexList = filtered
					dbSet.KV = append(dbSet.KV, &types.KeyValue{
						Key:   formatPendingTxFromKey(pending.GetFromAddress()),
						Value: types.Encode(list),
					})
				}
			}
		}
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

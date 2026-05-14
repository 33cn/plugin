package executor

import (
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
)

/*
 * 实现交易相关数据本地执行，数据不上链
 * 非关键数据，本地存储(localDB), 用于辅助查询，效率高
 */

func (l *lightclient) ExecLocal_BtcHeaders(headers *ltypes.BtcHeaders, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}

	for _, h := range headers.GetHeaders() {

		dbSet.KV = append(dbSet.KV,
			&types.KeyValue{
				Key:   btcHeaderKey(h.GetHeight()),
				Value: types.Encode(h),
			}, &types.KeyValue{
				Key:   btcHeaderHashHeightKey(h.GetHash()),
				Value: types.Encode(&types.Int64{Data: int64(h.GetHeight())}),
			})
	}

	//auto gen for localdb auto rollback
	return l.addAutoRollBack(tx, dbSet.KV), nil
}

// 当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (l *lightclient) addAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = l.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

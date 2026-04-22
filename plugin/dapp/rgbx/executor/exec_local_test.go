package executor

import (
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func testSetKV(t *testing.T, db db.DB, kvSet *types.LocalDBSet, del bool) {

	var err error
	for _, kv := range kvSet.GetKV() {
		if !del {
			err = db.Set(kv.Key, kv.Value)
			continue
		}
		if kv.Value == nil {
			err = db.Delete(kv.Key)
		} else {
			err = db.Set(kv.Key, kv.Value)
		}
		require.Nil(t, err)
	}
}

func TestRgbx_ExecLocal_Mint(t *testing.T) {

	r := newRgbx()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameMintAction, &rtypes.MintAsset{})
	require.Nil(t, err)

	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	kvSet, err := r.ExecLocal(tx, &types.ReceiptData{Logs: []*types.ReceiptLog{{
		Ty: rtypes.TyPendingTxLog, Log: types.Encode(&rtypes.PendingTx{ActionType: rtypes.TyMintAction})}}}, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, false)

	pendTx := &rtypes.PendingTx{}
	key := formatPendingTxKey(0, 0)
	require.Nil(t, readDB(local, key, pendTx))
	require.Equal(t, int32(rtypes.TyMintAction), pendTx.ActionType)
	kvSet, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, true)
	_, err = db.Get(key)
	require.Equal(t, types.ErrNotFound, err)
}

func TestRgbx_ExecLocal_Transfer(t *testing.T) {
	r := newRgbx()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameTransferAction, &rtypes.TransferAsset{})
	require.Nil(t, err)

	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	kvSet, err := r.ExecLocal(tx, &types.ReceiptData{Logs: []*types.ReceiptLog{{
		Ty: rtypes.TyPendingTxLog, Log: types.Encode(&rtypes.PendingTx{ActionType: rtypes.TyTransferAction})}}}, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, false)

	pendTx := &rtypes.PendingTx{}
	key := formatPendingTxKey(0, 0)
	require.Nil(t, readDB(local, key, pendTx))
	require.Equal(t, int32(rtypes.TyTransferAction), pendTx.ActionType)
	kvSet, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, true)
	_, err = db.Get(key)
	require.Equal(t, types.ErrNotFound, err)

}

func TestRgbx_ExecLocal_Confirm(t *testing.T) {

	r := newRgbx()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameConfirmAction, &rtypes.ConfirmTx{ConfirmedBlockHeight: 1})
	require.Nil(t, err)

	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	kvSet, err := r.ExecLocal(tx, &types.ReceiptData{Logs: []*types.ReceiptLog{{
		Ty: rtypes.TyPendingTxLog, Log: types.Encode(&rtypes.PendingTx{ActionType: rtypes.TyConfirmAction})}}}, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, false)

	pendTx := &rtypes.PendingTx{}
	key := formatPendingTxKey(0, 0)
	require.Nil(t, readDB(local, key, pendTx))
	require.True(t, pendTx.Confirmed)
	heightData := &types.Int64{}
	require.Nil(t, readDB(local, []byte(confirmedHeightKey), heightData))
	require.Equal(t, int64(1), heightData.Data)

	kvSet, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, true)
	_, err = db.Get(key)
	require.Equal(t, types.ErrNotFound, err)
	_, err = db.Get([]byte(confirmedHeightKey))
	require.Equal(t, types.ErrNotFound, err)
}

func TestRgbx_ExecLocal_DepositAsset(t *testing.T) {
	r := newRgbx()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameDepositAssetAction, &rtypes.DepositAsset{})
	require.Nil(t, err)

	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	kvSet, err := r.ExecLocal(tx, &types.ReceiptData{Logs: []*types.ReceiptLog{{
		Ty: rtypes.TyPendingTxLog, Log: types.Encode(&rtypes.PendingTx{ActionType: rtypes.TyDepositAsset})}}}, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, false)

	pendTx := &rtypes.PendingTx{}
	key := formatPendingTxKey(0, 0)
	require.Nil(t, readDB(local, key, pendTx))
	require.Equal(t, int32(rtypes.TyDepositAsset), pendTx.ActionType)

	kvSet, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, true)
	_, err = db.Get(key)
	require.Equal(t, types.ErrNotFound, err)
}

func TestRgbx_ExecLocal_WithdrawAsset(t *testing.T) {
	r := newRgbx()
	tx, err := r.GetExecutorType().CreateTransaction(rtypes.NameWithdrawAssetAction, &rtypes.WithdrawAsset{})
	require.Nil(t, err)

	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	kvSet, err := r.ExecLocal(tx, &types.ReceiptData{Logs: []*types.ReceiptLog{{
		Ty: rtypes.TyPendingTxLog, Log: types.Encode(&rtypes.PendingTx{ActionType: rtypes.TyWithDrawAsset})}}}, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, false)

	pendTx := &rtypes.PendingTx{}
	key := formatPendingTxKey(0, 0)
	require.Nil(t, readDB(local, key, pendTx))
	require.Equal(t, int32(rtypes.TyWithDrawAsset), pendTx.ActionType)

	kvSet, err = r.ExecDelLocal(tx, nil, 0)
	require.Nil(t, err)
	testSetKV(t, db, kvSet, true)
	_, err = db.Get(key)
	require.Equal(t, types.ErrNotFound, err)
}

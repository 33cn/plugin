package executor

import (
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/stretchr/testify/require"
)

type kvSetter interface {
	Set(key, value []byte) error
	Delete(key []byte) error
}

func applyStateKV(t *testing.T, stateDB kvSetter, kvs []*types.KeyValue) {
	for _, kv := range kvs {
		var errState error
		if kv.Value == nil {
			if stateDB != nil {
				errState = stateDB.Delete(kv.Key)
			}
			require.Nil(t, errState)
		} else {
			if stateDB != nil {
				errState = stateDB.Set(kv.Key, kv.Value)
			}
			require.Nil(t, errState)
		}
	}
}

type kvSetOnly interface {
	Set(key, value []byte) error
}

func applyLocalKV(t *testing.T, localDB kvSetOnly, kvs []*types.KeyValue) {
	for _, kv := range kvs {
		if localDB == nil || kv.Value == nil {
			continue
		}
		require.Nil(t, localDB.Set(kv.Key, kv.Value))
	}
}

func TestCrossChainDepositWithdrawConfirmExec(t *testing.T) {
	r := newRgbx()
	dir, state, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)

	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(testCfg)
	r.SetAPI(api)
	r.SetStateDB(state)
	r.SetLocalDB(local)

	userAddr, userPriv := util.Genaddress()

	deposit := &rtypes.DepositAsset{
		Amount:         1000,
		DepositAddress: userAddr,
		AssetSymbol:    "btc",
	}
	depositTx, err := r.GetExecutorType().CreateTransaction(rtypes.NameDepositAssetAction, deposit)
	require.Nil(t, err)
	depositTx.Sign(types.SECP256K1, userPriv)
	depositReceipt, err := r.Exec(depositTx, 0)
	require.Nil(t, err)
	applyStateKV(t, state, depositReceipt.KV)

	depositLocal, err := r.ExecLocal(depositTx, &types.ReceiptData{Ty: depositReceipt.Ty, Logs: depositReceipt.Logs}, 0)
	require.Nil(t, err)
	applyLocalKV(t, local, depositLocal.KV)

	accDB, err := r.(*rgbx).newAccount("xbtc")
	require.Nil(t, err)
	userAcc := accDB.LoadAccount(userAddr)
	require.Equal(t, int64(1000), userAcc.GetBalance())

	withdraw := &rtypes.WithdrawAsset{
		Amount:          600,
		FeeRate:         10,
		DestinationAddr: "tb1qwj4z7vmxq0x9mep74jkh0exj4rlzu7xhj4h3kl",
		AssetSymbol:     "BTC",
	}
	withdrawTx, err := r.GetExecutorType().CreateTransaction(rtypes.NameWithdrawAssetAction, withdraw)
	require.Nil(t, err)
	withdrawTx.Sign(types.SECP256K1, userPriv)
	withdrawReceipt, err := r.Exec(withdrawTx, 1)
	require.Nil(t, err)
	applyStateKV(t, state, withdrawReceipt.KV)

	withdrawLocal, err := r.ExecLocal(withdrawTx, &types.ReceiptData{Ty: withdrawReceipt.Ty, Logs: withdrawReceipt.Logs}, 1)
	require.Nil(t, err)
	applyLocalKV(t, local, withdrawLocal.KV)
	listResp, err := r.(*rgbx).Query_ListPendingTxByFrom(&types.ReqString{Data: userAddr})
	require.Nil(t, err)
	require.Len(t, listResp.(*rtypes.PendingTxs).GetPendingList(), 1)

	userAcc = accDB.LoadAccount(userAddr)
	require.Equal(t, int64(400), userAcc.GetBalance())
	lockAddr := r.(*rgbx).crossChainLockAddress(accDB)
	require.Equal(t, int64(600), accDB.LoadAccount(lockAddr).GetBalance())

	confirm := &rtypes.ConfirmTx{
		ActionType:           rtypes.TyWithDrawAsset,
		TxBlockHeight:        0,
		TxIndex:              1,
		TxHash:               withdrawTx.Hash(),
		ConfirmedBlockHeight: 2,
	}
	confirmTx, err := r.GetExecutorType().CreateTransaction(rtypes.NameConfirmAction, confirm)
	require.Nil(t, err)
	confirmTx.Sign(types.SECP256K1, testPriv)
	confirmReceipt, err := r.Exec(confirmTx, 2)
	require.Nil(t, err)
	applyStateKV(t, state, confirmReceipt.KV)
	confirmLocal, err := r.ExecLocal(confirmTx, &types.ReceiptData{Ty: confirmReceipt.Ty, Logs: confirmReceipt.Logs}, 2)
	require.Nil(t, err)
	applyLocalKV(t, local, confirmLocal.KV)

	userAcc = accDB.LoadAccount(userAddr)
	require.Equal(t, int64(400), userAcc.GetBalance())
	require.Equal(t, int64(0), accDB.LoadAccount(lockAddr).GetBalance())
	listResp, err = r.(*rgbx).Query_ListPendingTxByFrom(&types.ReqString{Data: userAddr})
	require.Nil(t, err)
	require.Len(t, listResp.(*rtypes.PendingTxs).GetPendingList(), 0)
}

func TestFormatCrossChainAccountSymbolByConfig(t *testing.T) {
	origin := rgbxCfg.CrossChainAssetPrefix
	defer func() { rgbxCfg.CrossChainAssetPrefix = origin }()

	rgbxCfg.CrossChainAssetPrefix = "rb"
	require.Equal(t, "rbBTC", formatCrossChainSymbol("btc"))

	rgbxCfg.CrossChainAssetPrefix = "X"
	require.Equal(t, "XBTC", formatCrossChainSymbol("btc"))
}

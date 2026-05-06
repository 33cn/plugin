package executor

import (
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRgbx_Query_ListPendingTx(t *testing.T) {

	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	funcName := "ListPendingTx"
	_, err := r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{}))
	require.Equal(t, types.ErrInvalidParam, err)
	_, err = r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{Count: maxListCount + 1}))
	require.Equal(t, types.ErrInvalidParam, err)

	key1, key2, key3 := formatPendingTxKey(0, 1), formatPendingTxKey(0, 2), formatPendingTxKey(0, 3)
	require.Nil(t, db.Set(key1, []byte("invalidData")))
	require.Nil(t, db.Set(key2, types.Encode(&rtypes.PendingTx{TxIndex: 2})))
	require.Nil(t, db.Set(key3, types.Encode(&rtypes.PendingTx{TxIndex: 3})))
	msg, err := r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{Count: 10}))
	require.Nil(t, err)
	require.Equal(t, 2, len(msg.(*rtypes.PendingTxs).GetPendingList()))
	msg, err = r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{Count: 10, StartIndex: 3}))
	require.Nil(t, err)
	require.Equal(t, 0, len(msg.(*rtypes.PendingTxs).GetPendingList()))
}

func TestRgbx_Query_ListPendingTx_EndHeight(t *testing.T) {
	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	funcName := "ListPendingTx"

	require.NoError(t, db.Set(formatPendingTxKey(1, 1), types.Encode(&rtypes.PendingTx{
		TxBlockHeight: 1,
		TxIndex:       1,
	})))
	require.NoError(t, db.Set(formatPendingTxKey(2, 1), types.Encode(&rtypes.PendingTx{
		TxBlockHeight: 2,
		TxIndex:       1,
	})))
	require.NoError(t, db.Set(formatPendingTxKey(3, 1), types.Encode(&rtypes.PendingTx{
		TxBlockHeight: 3,
		TxIndex:       1,
	})))

	msg, err := r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{Count: 10, EndHeight: 2}))
	require.NoError(t, err)
	list := msg.(*rtypes.PendingTxs).GetPendingList()
	require.Len(t, list, 2)
	require.Equal(t, int64(1), list[0].GetTxBlockHeight())
	require.Equal(t, int64(2), list[1].GetTxBlockHeight())

	msg, err = r.Query(funcName, types.Encode(&rtypes.ReqListPendingTx{Count: 10, EndHeight: 1}))
	require.NoError(t, err)
	list = msg.(*rtypes.PendingTxs).GetPendingList()
	require.Len(t, list, 1)
	require.Equal(t, int64(1), list[0].GetTxBlockHeight())
}

func TestRgbx_Query_GetPendingTx(t *testing.T) {
	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	funcName := "GetPendingTx"
	require.Nil(t, db.Set(formatPendingTxKey(0, 1), types.Encode(&rtypes.PendingTx{TxIndex: 1})))
	_, err := r.Query(funcName, types.Encode(&rtypes.ReqGetPendingTx{}))
	require.Equal(t, types.ErrNotFound, err)
	msg, err := r.Query(funcName, types.Encode(&rtypes.ReqGetPendingTx{Index: 1}))
	require.Equal(t, int64(1), msg.(*rtypes.PendingTx).GetTxIndex())
}

func TestRgbx_Query_GetConfirmedHeight(t *testing.T) {
	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	r.SetLocalDB(local)
	funcName := "GetConfirmedHeight"
	_, err := r.Query(funcName, nil)
	require.Nil(t, err)
	require.Nil(t, db.Set([]byte(confirmedHeightKey), types.Encode(&types.Int64{Data: 1})))
	msg, err := r.Query(funcName, nil)
	require.Equal(t, int64(1), msg.(*types.Int64).GetData())
}

func TestRgbx_Query_GetAsset(t *testing.T) {

	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(local)
	funcName := "GetAsset"
	_, err := r.Query(funcName, nil)
	require.Equal(t, types.ErrNotFound, err)
	require.Nil(t, db.Set(formatAssetKey("test"), types.Encode(&rtypes.RgbxAsset{Symbol: "test"})))
	msg, err := r.Query(funcName, types.Encode(&types.ReqString{Data: "test"}))
	require.Equal(t, "test", msg.(*rtypes.RgbxAsset).Symbol)
}

func TestRgbx_Query_GetCrossChainInfo(t *testing.T) {
	r := newRgbx()
	dir, db, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, db)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(local)
	funcName := "GetCrossChainInfo"

	msg, err := r.Query(funcName, types.Encode(&types.ReqString{Data: "btc"}))
	require.NoError(t, err)
	require.Equal(t, "", msg.(*rtypes.CrossChainInfo).GetAssetSymbol())

	info := &rtypes.CrossChainInfo{AssetSymbol: "BTC", TssAddress: "tb1qxx"}
	require.NoError(t, db.Set(formatCrossChainInfoKey("btc"), types.Encode(info)))
	msg, err = r.Query(funcName, types.Encode(&types.ReqString{Data: "btc"}))
	require.NoError(t, err)
	require.Equal(t, "BTC", msg.(*rtypes.CrossChainInfo).GetAssetSymbol())
	require.Equal(t, "tb1qxx", msg.(*rtypes.CrossChainInfo).GetTssAddress())

	require.NoError(t, db.Set(formatCrossChainInfoKey("bad"), []byte("not-protobuf")))
	_, err = r.Query(funcName, types.Encode(&types.ReqString{Data: "bad"}))
	require.Error(t, err)
}

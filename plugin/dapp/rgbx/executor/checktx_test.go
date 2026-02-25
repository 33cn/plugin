package executor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

var testCommitAddr string
var testPriv crypto.PrivKey
var testCfg *types.Chain33Config

func init() {
	testCommitAddr, testPriv = util.Genaddress()
	rgbxCfg.CommitAddress = testCommitAddr
	testCfg = types.NewChain33Config(types.GetDefaultCfgstring())
	Init(driverName, testCfg, nil)
}

type testCase struct {
	action    types.Message
	expectErr error
}

func Test_CheckTx(t *testing.T) {

	r := newRgbx()

	action := &rtypes.RgbxAction{}
	tx := &types.Transaction{Payload: []byte("testdata")}
	require.Equal(t, types.ErrActionNotSupport, r.CheckTx(tx, 0))

	tx.Payload = types.Encode(action)
	require.Equal(t, types.ErrActionNotSupport, r.CheckTx(tx, 0))
}

func testCheck(t *testing.T, driver dapp.Driver, tx *types.Transaction, action types.Message, expectErr error, idx int) {

	tx.Payload = types.Encode(action)
	require.Equalf(t, expectErr, driver.CheckTx(tx, 0), "testcase: %d", idx)
}

func Test_checkMint(t *testing.T) {

	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyMintAction

	tx := &types.Transaction{}
	mintAction := &rtypes.RgbxAction_Mint{}
	action.Value = mintAction

	tcArr := []*testCase{
		{
			expectErr: ErrInvalidSymbolLength,
			action:    &rtypes.MintAsset{Symbol: ""},
		},
		{
			expectErr: ErrInvalidSymbolLength,
			action:    &rtypes.MintAsset{Symbol: "aaaabbbbccccdddde"},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.MintAsset{Symbol: "test"},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.MintAsset{Symbol: "test", TotalAmount: rtypes.MaxAssetAmount + 1},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.MintAsset{Symbol: "test", Type: 1, TotalAmount: 2},
		},
		{
			expectErr: ErrInvalidMetaHashLength,
			action:    &rtypes.MintAsset{Symbol: "test", TotalAmount: 1, MetaHash: []byte(strings.Repeat("abcd", 9))},
		},
		{
			expectErr: ErrDuplicateAssetSymbol,
			action:    &rtypes.MintAsset{Symbol: "test", TotalAmount: 1, MetaHash: []byte("hash")},
		},
		{
			expectErr: ErrNilGenesisOut,
			action:    &rtypes.MintAsset{Symbol: "test1", TotalAmount: 1, MetaHash: []byte("hash"), GenesisOut: &rtypes.OutPoint{Hash: "hash"}},
		},
		{
			expectErr: nil,
			action: &rtypes.MintAsset{Symbol: "test1", TotalAmount: 1, MetaHash: []byte("hash"), GenesisOut: &rtypes.OutPoint{
				Hash:     "hash",
				PkScript: []byte("pubkey"),
			}},
		},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	r.SetStateDB(state)
	err := state.Set(formatAssetKey("test"), []byte("test"))
	require.Nil(t, err)

	for idx, tc := range tcArr {
		mintAction.Mint = tc.action.(*rtypes.MintAsset)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

func Test_checkTransfer(t *testing.T) {
	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyTransferAction
	addr, priv := util.Genaddress()
	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, priv)
	value := &rtypes.RgbxAction_Transfer{}
	action.Value = value
	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	tcArr := []*testCase{
		{
			expectErr: types.ErrInvalidAddress,
			action:    &rtypes.TransferAsset{From: "f4dfbea13c:0"},
		},
		{
			expectErr: types.ErrInvalidAddress,
			action:    &rtypes.TransferAsset{To: utxoAddr, ChangeAddr: "invalidaddr"},
		},
		{
			expectErr: ErrFromUtxoPkScriptNotSet,
			action:    &rtypes.TransferAsset{From: utxoAddr, To: addr},
		},
		{
			expectErr: ErrAssetNotExist,
			action:    &rtypes.TransferAsset{To: addr},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "normal"},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "normal", Amount: 1},
		},
		{
			expectErr: ErrInvalidAssetSender,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "collect"},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{From: utxoAddr, To: tx.From(), Symbol: "collect", FromPkScript: []byte("pubkey")},
		},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	r.SetStateDB(state)
	err := state.Set(formatAssetKey("normal"), types.Encode(&rtypes.RgbxAsset{}))
	require.Nil(t, err)
	err = state.Set(formatAssetKey("collect"), types.Encode(&rtypes.RgbxAsset{
		Type:  1,
		Owner: utxoAddr,
	}))
	require.Nil(t, err)

	for idx, tc := range tcArr {
		value.Transfer = tc.action.(*rtypes.TransferAsset)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

func Test_checkConfirm(t *testing.T) {
	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyConfirmAction
	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, testPriv)
	value := &rtypes.RgbxAction_Confirm{}
	action.Value = value
	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"

	btcTx := wire.MsgTx{}
	out, err := wire.NewOutPointFromString(utxoAddr)
	require.Nil(t, err)
	btcTx.TxIn = append(btcTx.TxIn,
		&wire.TxIn{PreviousOutPoint: *out},
		&wire.TxIn{PreviousOutPoint: wire.OutPoint{
			Hash:  out.Hash,
			Index: 1,
		}})
	btcTx.TxOut = append(btcTx.TxOut, wire.NewTxOut(0, []byte("testScript")))
	buf := bytes.NewBuffer(make([]byte, 0, btcTx.SerializeSizeStripped()))
	err = btcTx.SerializeNoWitness(buf)
	require.Nil(t, err)

	tcArr := []*testCase{
		{
			expectErr: ErrPendingTxNotExist,
			action:    &rtypes.ConfirmTx{TxIndex: 1},
		},
		{
			expectErr: ErrTxAlreadyConfirmed,
			action:    &rtypes.ConfirmTx{TxBlockHeight: 1},
		},
		{
			expectErr: ErrConfirmedHashNotEqual,
			action:    &rtypes.ConfirmTx{TxHash: []byte("hash")},
		},
		{
			expectErr: nil,
			action:    &rtypes.ConfirmTx{Timeout: true},
		},
		{
			expectErr: ErrDecodeBtcTx,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: []byte("invalidBtcTxData")}},
		},
		{
			expectErr: ErrInvalidSpendingTxIn,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: buf.Bytes(), SpendingInputIdx: 2}},
		},
		{
			expectErr: ErrSpendingInputNotEqual,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: buf.Bytes(), SpendingInputIdx: 1}},
		},
		{
			expectErr: nil,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: buf.Bytes(), OpRetOutputIdx: -1}},
		},
		{
			expectErr: ErrOpRetOutputPkScriptNotEqual,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: buf.Bytes()}},
		},
		{
			expectErr: nil,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{SpendingTx: buf.Bytes(), OpRetOutputPkScript: []byte("testScript")}},
		},
	}

	dir, state, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	r.SetStateDB(state)
	r.SetLocalDB(local)
	require.Nil(t, state.Set(formatPayloadKey(nil), types.Encode(&rtypes.PendingTx{})))
	require.Nil(t, local.Set(formatPendingTxKey(0, 0), types.Encode(&rtypes.PendingTx{Utxo: &rtypes.OutPoint{Hash: out.Hash.String()}})))
	require.Nil(t, local.Set(formatPendingTxKey(1, 0), types.Encode(&rtypes.PendingTx{Confirmed: true})))

	for idx, tc := range tcArr {
		value.Confirm = tc.action.(*rtypes.ConfirmTx)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

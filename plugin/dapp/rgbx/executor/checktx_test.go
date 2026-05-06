package executor

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/mock"
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
			expectErr: ErrInvalidFromUtxo,
			action:    &rtypes.TransferAsset{FromUtxo: "f4dfbea13c:0", To: addr, Amount: 1, Symbol: "normal"},
		},
		{
			expectErr: types.ErrInvalidAddress,
			action:    &rtypes.TransferAsset{To: utxoAddr, ChangeAddr: "invalidaddr", Amount: 1, Symbol: "normal"},
		},
		{
			expectErr: ErrInvalidFromUtxo,
			action:    &rtypes.TransferAsset{FromUtxo: utxoAddr, To: addr, Amount: 1, Symbol: "normal"},
		},
		{
			expectErr: ErrAssetNotExist,
			action:    &rtypes.TransferAsset{To: addr, Amount: 1},
		},
		{
			expectErr: ErrInvalidAssetAmount,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "normal"},
		},
		{
			expectErr: types.ErrInsufficientBalance,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "normal", Amount: 1},
		},
		{
			expectErr: ErrInvalidAssetSender,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "collect", Amount: 1},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "xbtc", Amount: 1},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{FromUtxo: utxoAddr, To: addr, Symbol: "xbtc", Amount: 1},
		},
		{
			expectErr: types.ErrInsufficientBalance,
			action:    &rtypes.TransferAsset{To: addr, Symbol: "xbtc", Amount: 2},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{FromUtxo: utxoAddr, To: tx.From(), Symbol: "collect", Amount: 1, FromUtxoPkScript: []byte("pubkey")},
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
	err = state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{AssetSymbol: "BTC"}))
	require.Nil(t, err)
	// checkCrossChainTransfer 现在使用 newAccount，symbol 为 "xbtc"
	crossAcc, err := r.(*rgbx).newAccount("xbtc")
	require.Nil(t, err)
	_, err = crossAcc.Mint(tx.From(), 1)
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
			action:    &rtypes.ConfirmTx{TxBlockHeight: 2, TxIndex: 0, TxHash: []byte("hash")},
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
	require.Nil(t, state.Set(formatPayloadKey([]byte("hash")), types.Encode(&rtypes.MintAsset{Symbol: "x", TotalAmount: 1})))
	require.Nil(t, local.Set(formatPendingTxKey(0, 0), types.Encode(&rtypes.PendingTx{Utxo: &rtypes.OutPoint{Hash: out.Hash.String()}})))
	require.Nil(t, local.Set(formatPendingTxKey(2, 0), types.Encode(&rtypes.PendingTx{
		Utxo:   &rtypes.OutPoint{Hash: out.Hash.String()},
		TxHash: []byte("other"),
	})))
	require.Nil(t, local.Set(formatPendingTxKey(1, 0), types.Encode(&rtypes.PendingTx{Confirmed: true})))

	for idx, tc := range tcArr {
		value.Confirm = tc.action.(*rtypes.ConfirmTx)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

func newTestnetWitnessAddr(t *testing.T) (addr string, pkScript []byte) {
	t.Helper()
	priv, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	pub := priv.PubKey().SerializeCompressed()
	waddr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pub), &chaincfg.TestNet3Params)
	require.NoError(t, err)
	pk, err := txscript.PayToAddrScript(waddr)
	require.NoError(t, err)
	return waddr.String(), pk
}

func Test_checkWithdraw(t *testing.T) {
	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyWithDrawAsset
	userAddr, userPriv := util.Genaddress()
	validDest, _ := newTestnetWitnessAddr(t)
	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, userPriv)
	value := &rtypes.RgbxAction_Withdraw{}
	action.Value = value

	tcArr := []*testCase{
		{expectErr: ErrInvalidWithdrawAmount, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: minBtcWithdrawAmount - 1, DestinationAddr: validDest, FeeRate: 1,
		}},
		{expectErr: ErrInvalidWithdrawFeeRate, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: minBtcWithdrawAmount, DestinationAddr: validDest, FeeRate: 0,
		}},
		{expectErr: ErrInvalidWithdrawFeeRate, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: minBtcWithdrawAmount, DestinationAddr: validDest, FeeRate: maxBtcFeeRate + 1,
		}},
		{expectErr: ErrInvalidWithdrawDestination, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: minBtcWithdrawAmount, DestinationAddr: "not-a-btc-address", FeeRate: 1,
		}},
		{expectErr: types.ErrInsufficientBalance, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: 10001, DestinationAddr: validDest, FeeRate: 1,
		}},
		{expectErr: nil, action: &rtypes.WithdrawAsset{
			AssetSymbol: "btc", Amount: minBtcWithdrawAmount, DestinationAddr: validDest, FeeRate: 10,
		}},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
	r.SetStateDB(state)
	require.Nil(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{AssetSymbol: "BTC"})))
	// checkWithdraw 现在使用 newAccount(withdraw.GetAssetSymbol())，即 newAccount("btc")
	// formatSymbol("btc") -> "BTC"，所以账户 symbol 是 "BTC"
	acc, err := r.(*rgbx).newAccount("xbtc")
	require.Nil(t, err)
	_, err = acc.Mint(userAddr, 10000)
	require.Nil(t, err)

	for idx, tc := range tcArr {
		if idx == 1 {
			require.Nil(t, state.Delete(formatCrossChainInfoKey("btc")))
		}
		if idx == 2 {
			require.Nil(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{AssetSymbol: "BTC"})))
		}
		value.Withdraw = tc.action.(*rtypes.WithdrawAsset)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

func Test_checkDeposit(t *testing.T) {
	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyDepositAsset
	tx := &types.Transaction{}
	value := &rtypes.RgbxAction_Deposit{}
	action.Value = value

	depAddr, _ := util.Genaddress()
	dupProofData := []byte("dup-tx-bytes")
	var minimalBtcTx wire.MsgTx
	minimalBtcTx.Version = 2
	buf := bytes.NewBuffer(make([]byte, 0, minimalBtcTx.SerializeSizeStripped()))
	require.NoError(t, minimalBtcTx.SerializeNoWitness(buf))

	tcArr := []*testCase{
		{expectErr: ErrInvalidDepositAmount, action: &rtypes.DepositAsset{
			AssetSymbol: "btc", Amount: 0, DepositAddress: depAddr, TxProof: &rtypes.BtcTxProof{TxData: []byte{1}},
		}},
		{expectErr: ErrInvalidDepositAddress, action: &rtypes.DepositAsset{
			AssetSymbol: "btc", Amount: 1, DepositAddress: "", TxProof: &rtypes.BtcTxProof{TxData: []byte{1}},
		}},
		{expectErr: ErrInvalidBtcTxProof, action: &rtypes.DepositAsset{
			AssetSymbol: "btc", Amount: 1, DepositAddress: depAddr, TxProof: nil,
		}},
		{expectErr: ErrInvalidBtcTxProof, action: &rtypes.DepositAsset{
			AssetSymbol: "btc", Amount: 1, DepositAddress: depAddr, TxProof: &rtypes.BtcTxProof{TxData: []byte{0xff}},
		}},
		{expectErr: ErrDuplicateDepositProof, action: &rtypes.DepositAsset{
			AssetSymbol: "btc", Amount: 1, DepositAddress: depAddr, TxProof: &rtypes.BtcTxProof{TxData: dupProofData},
		}},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	r.SetStateDB(state)
	require.Nil(t, state.Set(formatDepositUsedKey(dupProofData), []byte("1")))

	for idx, tc := range tcArr {
		value.Deposit = tc.action.(*rtypes.DepositAsset)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}

	// decode ok, header query fails
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(nil, errors.New("no header"))
	value.Deposit = &rtypes.DepositAsset{
		AssetSymbol:    "btc",
		Amount:         1,
		DepositAddress: depAddr,
		TxProof: &rtypes.BtcTxProof{
			TxData:      buf.Bytes(),
			BlockHeight: 1,
			BlockHash:   "00",
			TxIndex:     0,
			MerkleProof: nil,
		},
	}
	testCheck(t, r, tx, action, ErrGetBtcHeader, len(tcArr))
}

func Test_checkCommitDKG(t *testing.T) {
	r := newRgbx()
	action := &rtypes.RgbxAction{}
	action.Ty = rtypes.TyCommitDKGAction
	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, testPriv)
	value := &rtypes.RgbxAction_CommitDKG{}
	action.Value = value

	dkgAddr, validPk := newTestnetWitnessAddr(t)

	tcArr := []*testCase{
		{expectErr: ErrInvalidDkgAddress, action: &rtypes.CommitDKG{
			AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: []byte{0x01},
		}},
		{expectErr: ErrGetGuardianNodeAddress, action: &rtypes.CommitDKG{
			AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: validPk,
		}},
		{expectErr: ErrInvalidGuardianCommitter, action: &rtypes.CommitDKG{
			AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: validPk,
		}},
		{expectErr: ErrDuplicateDKGCommit, action: &rtypes.CommitDKG{
			AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: validPk,
		}},
		{expectErr: nil, action: &rtypes.CommitDKG{
			AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: validPk,
		}},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(testCfg)
	api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
	r.SetStateDB(state)

	for idx, tc := range tcArr {
		switch idx {
		case 1:
			api.ExpectedCalls = nil
			api.On("GetConfig").Return(testCfg)
			api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
			api.On("Query", paratypes.ParaX, "GetNodeGroupStatus", mock.Anything).Return(nil, errors.New("query fail"))
		case 2:
			api.ExpectedCalls = nil
			api.On("GetConfig").Return(testCfg)
			api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
			api.On("Query", paratypes.ParaX, "GetNodeGroupStatus", mock.Anything).Return(
				&paratypes.ParaNodeGroupStatus{TargetAddrs: "other1,other2"}, nil)
		case 3:
			api.ExpectedCalls = nil
			api.On("GetConfig").Return(testCfg)
			api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
			api.On("Query", paratypes.ParaX, "GetNodeGroupStatus", mock.Anything).Return(
				&paratypes.ParaNodeGroupStatus{TargetAddrs: testCommitAddr}, nil)
			require.Nil(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{AssetSymbol: "BTC"})))
		case 4:
			api.ExpectedCalls = nil
			api.On("GetConfig").Return(testCfg)
			api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
			api.On("Query", paratypes.ParaX, "GetNodeGroupStatus", mock.Anything).Return(
				&paratypes.ParaNodeGroupStatus{TargetAddrs: testCommitAddr}, nil)
			require.Nil(t, state.Delete(formatCrossChainInfoKey("btc")))
		}
		value.CommitDKG = tc.action.(*rtypes.CommitDKG)
		testCheck(t, r, tx, action, tc.expectErr, idx)
	}
}

func Test_decodeBtcAddressScript(t *testing.T) {

	params := lighttypes.GetBtcChainParams("regtest")

	priv, err := btcec.NewPrivateKey()
	require.Nil(t, err)
	pub := priv.PubKey().SerializeCompressed()
	waddr, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pub), params)
	require.Nil(t, err)
	fmt.Println(waddr.String())
	_, err = btcutil.DecodeAddress(waddr.String(), params)
	require.Nil(t, err)
}

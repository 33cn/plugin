package executor

import (
	"bytes"
	"errors"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_btcProof2String_and_merkelProof2String(t *testing.T) {
	proof := &rtypes.BtcTxProof{
		BlockHeight: 12,
		BlockHash:   "abc",
		TxIndex:     3,
		TxData:      []byte{1, 2},
	}
	s := btcProof2String(proof)
	require.Contains(t, s, "12")
	require.Contains(t, s, "abc")
	require.Contains(t, s, "3")
	require.Contains(t, s, "0102")

	require.Equal(t, "", merkelProof2String(nil))
	require.Contains(t, merkelProof2String([][]byte{{0xaa}, {0xbb}}), "aa")
}

func Test_hasExpectedOpReturnData_and_commitments(t *testing.T) {
	data := []byte("rgbx:test")
	script, err := txscript.NullDataScript(data)
	require.NoError(t, err)
	tx := &wire.MsgTx{}
	tx.TxOut = append(tx.TxOut, wire.NewTxOut(0, script))
	require.True(t, hasExpectedOpReturnData(tx, data))
	require.False(t, hasExpectedOpReturnData(tx, []byte("other")))

	chain33Hash := []byte{9, 8, 7}
	wdData := append([]byte(withdrawCommitmentPrefix), chain33Hash...)
	wdScript, err := txscript.NullDataScript(wdData)
	require.NoError(t, err)
	txWd := &wire.MsgTx{}
	txWd.TxOut = append(txWd.TxOut, wire.NewTxOut(0, wdScript))
	require.True(t, hasWithdrawCommitment(txWd, chain33Hash))

	depAddr, _ := util.Genaddress()
	depData := append([]byte(depositCommitmentPrefix), []byte(depAddr)...)
	depScript, err := txscript.NullDataScript(depData)
	require.NoError(t, err)
	txDep := &wire.MsgTx{}
	txDep.TxOut = append(txDep.TxOut, wire.NewTxOut(0, depScript))
	require.True(t, hasDepositCommitment(txDep, depAddr))

	utxo := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	out, err := wire.NewOutPointFromString(utxo)
	require.NoError(t, err)
	tx2 := &wire.MsgTx{}
	tx2.TxIn = []*wire.TxIn{{PreviousOutPoint: *out}}
	require.True(t, hasDepositCommitment(tx2, utxo))
	require.False(t, hasDepositCommitment(tx2, "0000000000000000000000000000000000000000000000000000000000000000:1"))
}

func Test_rgbx_validateDepositTxContent(t *testing.T) {
	r := newRgbx()
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	r.SetStateDB(state)

	tssScript := []byte{0x51, 0x52}
	deposit := &rtypes.DepositAsset{AssetSymbol: "btc", Amount: 1000}
	btcTx := &wire.MsgTx{}
	btcTx.TxOut = append(btcTx.TxOut, wire.NewTxOut(1000, tssScript))

	err := r.(*rgbx).validateDepositTxContent("h1", deposit, btcTx)
	require.Equal(t, ErrGetCrossChainInfo, err)

	require.NoError(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{
		AssetSymbol: "BTC",
		PkScript:    tssScript,
	})))

	err = r.(*rgbx).validateDepositTxContent("h1", deposit, btcTx)
	require.NoError(t, err)

	deposit.Amount = 999
	err = r.(*rgbx).validateDepositTxContent("h1", deposit, btcTx)
	require.Equal(t, ErrInvalidDepositAmount, err)
}

func Test_rgbx_validateWithdrawTxContent(t *testing.T) {
	r := newRgbx()
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)
	r.SetStateDB(state)

	destAddr, destScript := newTestnetWitnessAddr(t)
	tssScript := []byte{0x51, 0x20, 0xab, 0xcd}

	err := r.(*rgbx).validateWithdrawTxContent("h1", nil, &wire.MsgTx{})
	require.Equal(t, ErrPendingTxNotExist, err)

	pending := &rtypes.PendingTx{AssetSymbol: "btc", TargetAddress: destAddr, Amount: 50000}
	err = r.(*rgbx).validateWithdrawTxContent("h1", pending, &wire.MsgTx{})
	require.Equal(t, ErrInvalidCrossChainInfo, err)

	require.NoError(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{
		AssetSymbol: "BTC",
		PkScript:    tssScript,
	})))

	pendingBad := &rtypes.PendingTx{AssetSymbol: "btc", TargetAddress: "not-valid", Amount: 50000}
	err = r.(*rgbx).validateWithdrawTxContent("h1", pendingBad, &wire.MsgTx{})
	require.Equal(t, ErrInvalidWithdrawDestination, err)

	btcTx := &wire.MsgTx{}
	btcTx.TxOut = append(btcTx.TxOut,
		wire.NewTxOut(40000, destScript),
		wire.NewTxOut(10000, []byte{0x76}), // unexpected script
	)
	err = r.(*rgbx).validateWithdrawTxContent("h1", pending, btcTx)
	require.Equal(t, ErrInvalidWithdrawDestinationScript, err)

	btcTx2 := &wire.MsgTx{}
	btcTx2.TxOut = append(btcTx2.TxOut,
		wire.NewTxOut(40000, destScript),
		wire.NewTxOut(10000, tssScript),
	)
	err = r.(*rgbx).validateWithdrawTxContent("h1", pending, btcTx2)
	require.NoError(t, err)

	btcTx3 := &wire.MsgTx{}
	btcTx3.TxOut = append(btcTx3.TxOut, wire.NewTxOut(0, destScript)) // destAmount 0
	btcTx3.TxOut = append(btcTx3.TxOut, wire.NewTxOut(60000, tssScript))
	err = r.(*rgbx).validateWithdrawTxContent("h1", pending, btcTx3)
	require.Equal(t, ErrInvalidWithdrawAmount, err)

	btcTx4 := &wire.MsgTx{}
	btcTx4.TxOut = append(btcTx4.TxOut, wire.NewTxOut(60000, destScript)) // > pending amount
	btcTx4.TxOut = append(btcTx4.TxOut, wire.NewTxOut(10000, tssScript))
	err = r.(*rgbx).validateWithdrawTxContent("h1", pending, btcTx4)
	require.Equal(t, ErrInvalidWithdrawAmount, err)
}

func newBtcTxProofFixture(t *testing.T) (*rgbx, *rtypes.BtcTxProof, string) {
	t.Helper()
	r := newRgbx().(*rgbx)
	var btcTx wire.MsgTx
	btcTx.Version = 2
	btcTx.TxOut = append(btcTx.TxOut, wire.NewTxOut(1000, []byte{0x51}))
	buf := new(bytes.Buffer)
	require.NoError(t, btcTx.SerializeNoWitness(buf))
	txID := btcTx.TxHash()
	leaves := [][]byte{txID.CloneBytes()}
	root := merkle.GetMerkleRoot(leaves)
	_, branch := merkle.GetMerkleRootAndBranch(leaves, 0)
	rootHash, err := chainhash.NewHash(root)
	require.NoError(t, err)
	proof := &rtypes.BtcTxProof{
		TxData:      buf.Bytes(),
		BlockHeight: 100,
		BlockHash:   "deadbeef",
		TxIndex:     0,
		MerkleProof: branch,
	}
	return r, proof, rootHash.String()
}

func Test_rgbx_validateBtcTxProof_emptyAndDecode(t *testing.T) {
	r := newRgbx().(*rgbx)
	_, err := r.validateBtcTxProof("tx1", nil)
	require.Equal(t, ErrInvalidBtcTxProof, err)
	_, err = r.validateBtcTxProof("tx1", &rtypes.BtcTxProof{TxData: []byte{0xff}})
	require.Equal(t, ErrInvalidBtcTxProof, err)
}

func Test_rgbx_validateBtcTxProof_getHeaderError(t *testing.T) {
	r, proof, _ := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.MatchedBy(func(req types.Message) bool {
		h, ok := req.(*ltypes.ReqGetBtcHeader)
		return ok && h != nil && h.Height == 100
	})).Return(nil, errors.New("nope"))
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", proof)
	require.Equal(t, ErrGetBtcHeader, err)
}

func Test_rgbx_validateBtcTxProof_headerMismatch(t *testing.T) {
	r, proof, rootStr := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       rootStr,
		Height:     101,
		MerkleRoot: rootStr,
	}, nil)
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", proof)
	require.Equal(t, ErrInvalidBtcProofBlock, err)
}

func Test_rgbx_validateBtcTxProof_headerHashMismatch(t *testing.T) {
	r, proof, rootStr := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       rootStr,
		Height:     100,
		MerkleRoot: rootStr,
	}, nil)
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", proof)
	require.Equal(t, ErrInvalidBtcProofBlock, err)
}

func Test_rgbx_validateBtcTxProof_invalidMerkleRootStr(t *testing.T) {
	r, proof, _ := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       "deadbeef",
		Height:     100,
		MerkleRoot: "gg",
	}, nil)
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", proof)
	require.Equal(t, ErrInvalidBtcProofMerkle, err)
}

func Test_rgbx_validateBtcTxProof_invalidHeaderType(t *testing.T) {
	r, proof, _ := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&types.ReplyString{Data: "x"}, nil)
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", proof)
	require.Equal(t, ErrGetBtcHeader, err)
}

func Test_rgbx_validateBtcTxProof_merkleBranchMismatch(t *testing.T) {
	r, proof, rootStr := newBtcTxProofFixture(t)
	badProof := &rtypes.BtcTxProof{
		TxData:      proof.TxData,
		BlockHeight: proof.BlockHeight,
		BlockHash:   proof.BlockHash,
		TxIndex:     proof.TxIndex,
		MerkleProof: [][]byte{bytes.Repeat([]byte{1}, 32)},
	}
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       "deadbeef",
		Height:     100,
		MerkleRoot: rootStr,
	}, nil)
	r.SetAPI(api)
	_, err := r.validateBtcTxProof("tx1", badProof)
	require.Equal(t, ErrInvalidBtcProofMerkle, err)
}

func Test_rgbx_validateBtcTxProof_success(t *testing.T) {
	r, proof, rootStr := newBtcTxProofFixture(t)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(types.NewChain33Config(types.GetDefaultCfgstring()))
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       "deadbeef",
		Height:     100,
		MerkleRoot: rootStr,
	}, nil)
	r.SetAPI(api)
	tx, err := r.validateBtcTxProof("tx1", proof)
	require.NoError(t, err)
	require.Equal(t, int32(2), tx.Version)
}

func Test_rgbx_checkWithdrawConfirm(t *testing.T) {
	r := newRgbx()
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(&types.ReplyString{Data: "testnet3"}, nil)

	confirm := &rtypes.ConfirmTx{TxHash: []byte{1, 2, 3}}
	pending := &rtypes.PendingTx{AssetSymbol: "btc", Amount: 1000}
	destAddr, destPk := newTestnetWitnessAddr(t)
	pending.TargetAddress = destAddr
	tssScript := []byte{0x51, 0x20, 0x11, 0x22}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	r.SetStateDB(state)
	require.NoError(t, state.Set(formatCrossChainInfoKey("btc"), types.Encode(&rtypes.CrossChainInfo{
		AssetSymbol: "BTC",
		PkScript:    tssScript,
	})))

	commitData := append([]byte(withdrawCommitmentPrefix), confirm.GetTxHash()...)
	commitScript, err := txscript.NullDataScript(commitData)
	require.NoError(t, err)
	var btcTx wire.MsgTx
	btcTx.Version = 2
	btcTx.TxOut = append(btcTx.TxOut,
		wire.NewTxOut(0, commitScript),
		wire.NewTxOut(500, destPk),
		wire.NewTxOut(500, tssScript),
	)
	buf := new(bytes.Buffer)
	require.NoError(t, btcTx.SerializeNoWitness(buf))
	txID := btcTx.TxHash()
	leaves := [][]byte{txID.CloneBytes()}
	root := merkle.GetMerkleRoot(leaves)
	_, branch := merkle.GetMerkleRootAndBranch(leaves, 0)
	rootHash, err := chainhash.NewHash(root)
	require.NoError(t, err)
	proof := &rtypes.BtcTxProof{
		TxData:      buf.Bytes(),
		BlockHeight: 1,
		BlockHash:   "dead",
		TxIndex:     0,
		MerkleProof: branch,
	}
	confirm.BtcTxProof = proof

	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       "dead",
		Height:     1,
		MerkleRoot: rootHash.String(),
	}, nil)

	err = r.(*rgbx).checkWithdrawConfirm("a", "b", confirm, pending)
	require.NoError(t, err)

	confirmMismatch := &rtypes.ConfirmTx{TxHash: []byte{9, 9, 9}, BtcTxProof: proof}
	err = r.(*rgbx).checkWithdrawConfirm("a", "b", confirmMismatch, pending)
	require.Equal(t, ErrInvalidBtcProofCommitment, err)
}

func Test_rgbx_decodeBtcAddressScript_empty(t *testing.T) {
	r := newRgbx()
	_, err := r.(*rgbx).decodeBtcAddressScript("")
	require.Equal(t, types.ErrInvalidAddress, err)
}

func Test_rgbx_getBtcNetName_queryError(t *testing.T) {
	r := newRgbx()
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	api.On("Query", ltypes.LightclientX, "GetBtcNetName", mock.Anything).Return(nil, errors.New("down"))
	_, err := r.(*rgbx).getBtcNetName()
	require.Error(t, err)
}

package executor

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func testExec(t *testing.T, driver dapp.Driver, actionName string, action types.Message, expectErr error, index int) *types.Receipt {

	tx, err := driver.GetExecutorType().CreateTransaction(actionName, action)
	require.Nilf(t, err, "testcase %d", index)
	tx.Sign(types.SECP256K1, testPriv)
	recp, err := driver.Exec(tx, 0)
	require.Equalf(t, expectErr, err, "testcase %d", index)
	return recp
}

func TestRgbx_Exec_Mint(t *testing.T) {

	r := newRgbx()
	mint := &rtypes.MintAsset{}
	testExec(t, r, rtypes.NameMintAction, mint, nil, 0)
}

func TestRgbx_Exec_Transfer(t *testing.T) {

	r := newRgbx()
	addr2, _ := util.Genaddress()
	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	tcArr := []*testCase{
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{FromUtxo: utxoAddr},
		},
		{
			expectErr: ErrAssetNotExist,
			action:    &rtypes.TransferAsset{Symbol: "test"},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{Symbol: "collect"},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{Symbol: "normal", Amount: 1},
		},
		{
			expectErr: types.ErrNoBalance,
			action:    &rtypes.TransferAsset{Symbol: "normal", Amount: 1},
		},
		{
			expectErr: nil,
			action:    &rtypes.TransferAsset{Symbol: "xbtc", To: addr2, Amount: 1},
		},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	require.Nil(t, state.Set(formatAssetKey("normal"), types.Encode(&rtypes.RgbxAsset{})))
	require.Nil(t, state.Set(formatAssetKey("collect"), types.Encode(&rtypes.RgbxAsset{Type: 1})))

	acc, err := r.(*rgbx).newAccount("normal")
	require.Nil(t, err)
	_, err = acc.Mint(testCommitAddr, 1)
	require.Nil(t, err)
	crossAcc, err := r.(*rgbx).newAccount("xbtc")
	require.Nil(t, err)
	_, err = crossAcc.Mint(testCommitAddr, 1)
	require.Nil(t, err)

	for idx, tc := range tcArr {
		testExec(t, r, rtypes.NameTransferAction, tc.action, tc.expectErr, idx)
	}
}

func TestRgbx_Exec_Confirm(t *testing.T) {

	r := newRgbx()
	addr, _ := util.Genaddress()
	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	normal, normal1, collect, collect1 := "normal", "normal1", "collect", "collect1"

	buildTxWithOpReturn := func(data []byte) []byte {
		tx := wire.NewMsgTx(wire.TxVersion)
		tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
		script, _ := txscript.NullDataScript(data)
		tx.AddTxOut(wire.NewTxOut(0, script))
		buf := new(bytes.Buffer)
		_ = tx.SerializeNoWitness(buf)
		return buf.Bytes()
	}

	tcArr := []*testCase{
		{
			expectErr: nil,
			action:    &rtypes.ConfirmTx{Timeout: true},
		},
		{
			expectErr: nil,
			action:    &rtypes.ConfirmTx{UtxoProof: &rtypes.UtxoSpendingProof{}},
		},
		{
			expectErr: nil,
			action: &rtypes.ConfirmTx{
				ActionType: rtypes.TyMintAction,
				TxHash:     []byte(normal),
				UtxoProof:  &rtypes.UtxoSpendingProof{OpRetOutputIdx: 0},
				BtcTxProof: &rtypes.BtcTxProof{TxData: buildTxWithOpReturn([]byte(normal))},
			},
		},
		{
			expectErr: nil,
			action: &rtypes.ConfirmTx{
				ActionType: rtypes.TyMintAction,
				TxHash:     []byte(collect),
				UtxoProof:  &rtypes.UtxoSpendingProof{OpRetOutputIdx: 0},
				BtcTxProof: &rtypes.BtcTxProof{TxData: buildTxWithOpReturn([]byte(collect))},
			},
		},
		{
			expectErr: nil,
			action: &rtypes.ConfirmTx{
				TxHash:     []byte(normal1),
				UtxoProof:  &rtypes.UtxoSpendingProof{OpRetOutputIdx: 0},
				BtcTxProof: &rtypes.BtcTxProof{TxData: buildTxWithOpReturn([]byte(normal1))},
			},
		},
		{
			expectErr: nil,
			action: &rtypes.ConfirmTx{
				TxHash:     []byte(collect1),
				UtxoProof:  &rtypes.UtxoSpendingProof{OpRetOutputIdx: 0},
				BtcTxProof: &rtypes.BtcTxProof{TxData: buildTxWithOpReturn([]byte(collect1))},
			},
		},
	}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)

	require.Nil(t, state.Set(formatPayloadKey([]byte(normal)), types.Encode(&rtypes.MintAsset{Symbol: normal, TotalAmount: 1})))
	require.Nil(t, state.Set(formatPayloadKey([]byte(collect)), types.Encode(&rtypes.MintAsset{Symbol: collect, Type: 1, TotalAmount: 1})))

	require.Nil(t, state.Set(formatAssetKey(normal1), types.Encode(&rtypes.RgbxAsset{})))
	require.Nil(t, state.Set(formatAssetKey(collect1), types.Encode(&rtypes.RgbxAsset{Type: 1, Symbol: collect1})))

	require.Nil(t, state.Set(formatPayloadKey([]byte(normal1)),
		types.Encode(&rtypes.TransferAsset{Symbol: normal1, Amount: 1, FromUtxo: utxoAddr, To: addr})))
	require.Nil(t, state.Set(formatPayloadKey([]byte(collect1)),
		types.Encode(&rtypes.TransferAsset{Symbol: collect1, To: addr})))

	accDB, err := r.(*rgbx).newAccount(normal1)
	require.Nil(t, err)
	_, err = accDB.Mint(utxoAddr, 2)
	require.Nil(t, err)

	for idx, tc := range tcArr {
		recp := testExec(t, r, rtypes.NameConfirmAction, tc.action, tc.expectErr, idx)
		if len(recp.GetKV()) > 0 {
			util.SaveKVList(state, recp.KV)
		}
	}
	// check mint
	asset := &rtypes.RgbxAsset{}
	require.Nil(t, readDB(state, formatAssetKey(normal), asset))
	require.Equal(t, formatSymbol(normal), asset.Symbol)
	require.Nil(t, readDB(state, formatAssetKey(collect), asset))
	require.Equal(t, formatSymbol(collect), asset.Symbol)
	require.Equal(t, rtypes.Collectible, rtypes.AssetType(asset.Type))
	collectSpendHash := chainhash.DoubleHashH(buildTxWithOpReturn([]byte(collect))).String()
	owner := rtypes.FormatUtxo(collectSpendHash, 1)
	require.Equal(t, owner, asset.Owner)

	// check transfer
	require.Nil(t, readDB(state, formatAssetKey(collect1), asset))
	require.Equal(t, addr, asset.Owner)
	require.Equal(t, int64(0), accDB.LoadAccount(utxoAddr).Balance)
	require.Equal(t, int64(1), accDB.LoadAccount(addr).Balance)
	normal1SpendHash := chainhash.DoubleHashH(buildTxWithOpReturn([]byte(normal1))).String()
	changeAddr := rtypes.FormatUtxo(normal1SpendHash, 1)
	require.Equal(t, int64(1), accDB.LoadAccount(changeAddr).Balance)
}

func TestRgbx_Exec_Deposit(t *testing.T) {
	r := newRgbx()
	deposit := &rtypes.DepositAsset{
		Amount:         100,
		DepositAddress: testCommitAddr,
		AssetSymbol:    "btc",
		TxProof:        &rtypes.BtcTxProof{},
	}
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	testExec(t, r, rtypes.NameDepositAssetAction, deposit, nil, 0)
}

func TestRgbx_Exec_Withdraw(t *testing.T) {
	r := newRgbx()
	withdraw := &rtypes.WithdrawAsset{
		Amount:          600,
		FeeRate:         10,
		DestinationAddr: "tb1qw508d6qejxtdg4y5r3zarvary0c5xw7kygt080",
		AssetSymbol:     "BTC",
	}
	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	r.SetStateDB(state)
	// withdraw 使用 newAccount 直接操作账户，symbol 为 "BTC"
	// 但为了测试通过，需要设置 cross chain info 和账户余额
	require.NoError(t, state.Set(formatCrossChainInfoKey("BTC"), types.Encode(&rtypes.CrossChainInfo{AssetSymbol: "BTC"})))
	acc, err := r.(*rgbx).newAccount("xBTC")
	require.Nil(t, err)
	_, err = acc.Mint(testCommitAddr, 1000)
	require.Nil(t, err)
	testExec(t, r, rtypes.NameWithdrawAssetAction, withdraw, nil, 0)
}

func TestRgbx_Exec_CommitDKG(t *testing.T) {
	r := newRgbx()
	dkgAddr, pkScript := newTestnetWitnessAddr(t)
	commit := &rtypes.CommitDKG{AssetSymbol: "btc", DkgAddress: dkgAddr, PkScript: pkScript}

	dir, state, _ := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	r.SetAPI(api)
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api.On("GetConfig").Return(cfg)
	api.On("Query", paratypes.ParaX, "GetNodeGroupStatus", mock.Anything).Return(
		&paratypes.ParaNodeGroupStatus{TargetAddrs: testCommitAddr}, nil)
	r.SetStateDB(state)

	recp := testExec(t, r, rtypes.NameCommitDKGAction, commit, nil, 0)
	require.NotNil(t, recp)
	require.NotEmpty(t, recp.KV)
}

// TestRgbx_Exec_Confirm_WithMerkleProof walks the full confirm happy path with a
// real BtcTxProof (BlockHeight/BlockHash/TxIndex/MerkleProof): checkConfirm must
// validate the BTC merkle proof, then Exec_Confirm must perform the state
// transition and mint the native asset.
func TestRgbx_Exec_Confirm_WithMerkleProof(t *testing.T) {
	r := newRgbx().(*rgbx)
	utxoAddr := "74503993e7c8d4280f6fbb99ae5aaa92231a1981a358e40f97e2b4f4dfbea13c:0"
	out, err := wire.NewOutPointFromString(utxoAddr)
	require.NoError(t, err)

	pendingHash := []byte("merkle-mint")
	symbol := "nativemerkle"

	// BTC spend tx: spends the genesis utxo, output 0 = OP_RETURN commitment
	// to the pending chain33 tx hash, output 1 = value (owner utxo).
	var spendTx wire.MsgTx
	spendTx.Version = 2
	spendTx.TxIn = append(spendTx.TxIn, &wire.TxIn{PreviousOutPoint: *out})
	commitment, err := txscript.NullDataScript(pendingHash)
	require.NoError(t, err)
	spendTx.TxOut = append(spendTx.TxOut, wire.NewTxOut(0, commitment))
	spendTx.TxOut = append(spendTx.TxOut, wire.NewTxOut(5000, []byte{0x51}))

	buf := new(bytes.Buffer)
	require.NoError(t, spendTx.SerializeNoWitness(buf))
	txData := buf.Bytes()
	txID := spendTx.TxHash()

	// Build a 2-tx block merkle tree so the proof carries a real branch.
	// Compute the branch before GetMerkleRoot, which mutates its leaf slice.
	dummyID := chainhash.DoubleHashH(bytes.Repeat([]byte{0xaa}, 32))
	leaves := [][]byte{txID.CloneBytes(), dummyID.CloneBytes()}
	_, branch := merkle.GetMerkleRootAndBranch(leaves, 0)
	root := merkle.GetMerkleRoot(leaves)
	rootHash, err := chainhash.NewHash(root)
	require.NoError(t, err)

	confirm := &rtypes.ConfirmTx{
		ActionType:    rtypes.TyMintAction,
		TxBlockHeight: 10,
		TxIndex:       0,
		TxHash:        pendingHash,
		UtxoProof:     &rtypes.UtxoSpendingProof{SpendingInputIdx: 0, OpRetOutputIdx: 0},
		BtcTxProof:    &rtypes.BtcTxProof{TxData: txData, BlockHeight: 100, BlockHash: "deadbeef", TxIndex: 0, MerkleProof: branch},
	}

	dir, state, local := util.CreateTestDB()
	defer util.CloseTestDB(dir, state)
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig").Return(testCfg)
	api.On("Query", ltypes.LightclientX, "GetBtcHeader", mock.Anything).Return(&ltypes.BtcHeader{
		Hash:       "deadbeef",
		Height:     100,
		MerkleRoot: rootHash.String(),
	}, nil)
	r.SetAPI(api)
	r.SetStateDB(state)
	r.SetLocalDB(local)

	// Pending payload (MintAsset) in state, pending tx in local db.
	require.NoError(t, state.Set(formatPayloadKey(pendingHash), types.Encode(&rtypes.MintAsset{Symbol: symbol, TotalAmount: 1000})))
	require.NoError(t, local.Set(formatPendingTxKey(10, 0), types.Encode(&rtypes.PendingTx{
		Utxo:   &rtypes.OutPoint{Hash: out.Hash.String()},
		TxHash: pendingHash,
	})))

	tx := &types.Transaction{}
	tx.Sign(types.SECP256K1, testPriv)
	txHash := hex.EncodeToString(tx.Hash())

	// 1) merkle proof verification
	err = r.checkConfirm(tx.From(), txHash, confirm)
	require.NoError(t, err)

	// 2) state transition + mint
	receipt, err := r.Exec_Confirm(confirm, tx, 0)
	require.NoError(t, err)
	require.NotNil(t, receipt)
	require.NotEmpty(t, receipt.KV)
	util.SaveKVList(state, receipt.KV)

	// 3) verify the asset was minted with the spend tx as genesis.
	// For a Normal asset the owner utxo holds the balance in the account DB
	// (asset.Owner is only populated for collectibles).
	asset := &rtypes.RgbxAsset{}
	require.NoError(t, readDB(state, formatAssetKey(symbol), asset))
	require.Equal(t, formatSymbol(symbol), asset.Symbol)
	spendHash := chainhash.DoubleHashH(txData).String()
	require.Equal(t, spendHash, asset.GenesisBtcTxHash)

	owner := rtypes.FormatUtxo(spendHash, uint32(confirm.UtxoProof.GetOpRetOutputIdx()+1))
	accDB, err := r.newAccount(symbol)
	require.NoError(t, err)
	require.Equal(t, int64(1000), accDB.LoadAccount(owner).Balance)
}

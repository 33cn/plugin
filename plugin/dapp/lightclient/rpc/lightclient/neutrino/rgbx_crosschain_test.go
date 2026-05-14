package neutrino

import (
	"bytes"
	"math"
	"testing"

	lighttypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
)

func Test_encodeOutPoint_decodeOutPoint_roundTrip(t *testing.T) {
	hash, err := chainhash.NewHashFromStr("000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f")
	require.NoError(t, err)
	op := wire.OutPoint{Hash: *hash, Index: 1}
	encoded := encodeOutPoint(&op)
	decoded, err := decodeOutPoint(encoded)
	require.NoError(t, err)
	require.Equal(t, op.String(), decoded.String())
}

func Test_pending2WithdrawRequest(t *testing.T) {
	chain33Hash := []byte{0xab, 0xcd}
	pending := &rtypes.PendingTx{
		TxHash:        chain33Hash,
		Amount:        100_000,
		FeeRate:       10,
		TargetAddress: "bcrt1qexample",
		ActionType:    rtypes.TyWithDrawAsset,
	}
	req := pending2WithdrawRequest(pending)
	require.Equal(t, pending.GetTxHash(), req.chain33WithDrawHash)
	require.Equal(t, int64(100_000), int64(req.amount))
	require.Equal(t, int64(10), int64(req.feeRate))
	require.Equal(t, "bcrt1qexample", req.toAddress)
}

func Test_rgbx_createConfirmPayload_timeout(t *testing.T) {
	r := newRGBX()
	r.pendingCache.addTx("deadbeef", &rtypes.PendingTx{TxBlockHeight: 100})
	pendTx := &rtypes.PendingTx{
		ActionType:    rtypes.TyTransferAction,
		TxBlockHeight: 50,
		TxIndex:       2,
		TxHash:        []byte{1, 2, 3, 4},
	}
	info := &utxoSpendInfo{pendingTxHash: "deadbeef", timeout: true}
	confirm, err := r.createConfirmPayload(info, pendTx)
	require.NoError(t, err)
	require.True(t, confirm.Timeout)
	require.Equal(t, pendTx.ActionType, confirm.ActionType)
	require.Equal(t, pendTx.TxBlockHeight, confirm.TxBlockHeight)
	require.Equal(t, pendTx.TxIndex, confirm.TxIndex)
	require.Equal(t, int64(99), confirm.ConfirmedBlockHeight) // min height 100 - 1
	require.Nil(t, confirm.UtxoProof)
}

func Test_rgbx_createConfirmPayload_withSpendingTxAndOpReturn(t *testing.T) {
	r := newRGBX()
	r.pendingCache.addTx("abc", &rtypes.PendingTx{TxBlockHeight: 10})

	spendTx := wire.NewMsgTx(wire.TxVersion)
	spendTx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	opRet, err := txscript.NullDataScript([]byte("rgbx:test"))
	require.NoError(t, err)
	spendTx.AddTxOut(wire.NewTxOut(0, opRet))
	spendTx.AddTxOut(wire.NewTxOut(1000, []byte{txscript.OP_0, 0x14}))

	info := &utxoSpendInfo{
		pendingTxHash:      "abc",
		spendingInputIndex: 0,
		spendingTx:         spendTx,
	}
	pendTx := &rtypes.PendingTx{
		ActionType:    rtypes.TyMintAction,
		TxBlockHeight: 5,
		TxIndex:       1,
		TxHash:        []byte{0x11},
	}
	confirm, err := r.createConfirmPayload(info, pendTx)
	require.NoError(t, err)
	require.False(t, confirm.Timeout)
	require.NotNil(t, confirm.UtxoProof)
	require.Equal(t, uint32(0), confirm.UtxoProof.SpendingInputIdx)
	require.GreaterOrEqual(t, confirm.UtxoProof.OpRetOutputIdx, int32(0))
	require.Equal(t, byte(txscript.OP_RETURN), confirm.UtxoProof.OpRetOutputPkScript[0])

	var back wire.MsgTx
	require.NoError(t, back.DeserializeNoWitness(bytes.NewReader(confirm.UtxoProof.SpendingTx)))
	require.Equal(t, spendTx.TxHash().String(), back.TxHash().String())
}

func Test_estimateBtcFee(t *testing.T) {
	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	tx.AddTxOut(wire.NewTxOut(1000, []byte{txscript.OP_0, 0x14}))
	fee := estimateBtcFee(tx, 10)
	require.Greater(t, int64(fee), int64(0))
}

func Test_tssService_parseTxFromNotify(t *testing.T) {
	ts := &tssService{}
	_, _, err := ts.parseTxFromNotify(nil)
	require.Error(t, err)

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(wire.NewTxIn(&wire.OutPoint{}, nil, nil))
	buf := bytes.NewBuffer(nil)
	require.NoError(t, tx.SerializeNoWitness(buf))
	notify := &lighttypes.TssSignNotify{
		BtcTxData:    buf.Bytes(),
		InputAmounts: []int64{1000},
	}
	got, amounts, err := ts.parseTxFromNotify(notify)
	require.NoError(t, err)
	require.Len(t, amounts, 1)
	require.Equal(t, tx.TxHash().String(), got.TxHash().String())

	notify.InputAmounts = nil
	_, _, err = ts.parseTxFromNotify(notify)
	require.Error(t, err)
}

func Test_pendingTxCache_getMinPendingHeight_empty(t *testing.T) {
	c := newPendingTxCache(8)
	require.Equal(t, int64(math.MaxInt64), c.getMinPendingHeight())
}

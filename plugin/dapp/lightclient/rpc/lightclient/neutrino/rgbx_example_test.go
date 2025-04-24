package neutrino

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	// bcrt1qp4v07g5u0s0x6a2llrz3lxr8z7arz7vt5wcjpp
	priv1 = "4964acfcd8f48deecda252e34a6d6cde93beceebbe7ec3391a3006edd3314fd1"

	btcFee      = int64(0.01 * 1e8)
	totalAmount = int64(49.99 * 1e8)
)

func newTxSigHashes(tx *wire.MsgTx, pkScript []byte, amount int64) *txscript.TxSigHashes {

	prevOutFetcher := txscript.NewCannedPrevOutputFetcher(pkScript, amount)
	return txscript.NewTxSigHashes(tx, prevOutFetcher)
}

func Test_createRgbxTx(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	privByte, err := hex.DecodeString(priv1)
	require.NoError(t, err)
	priv1, pub1 := btcec.PrivKeyFromBytes(privByte)

	hash, err := chainhash.NewHashFromStr("b4c0e95b0a43b7d6dc3b39bd9ef5e7b0304e5fbd8ea4670136a9c169a3615535")
	require.NoError(t, err)

	txIn := &wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  *hash,
			Index: 1,
		},
		Sequence: 0xffffffff,
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	tx.AddTxIn(txIn)
	chain33Hash := "abd8bffa6c4a4018c1dfe062c943c6fbb2df1ecf3cae970ebfede4195a6ede0b"
	chain33HashData, err := hex.DecodeString(chain33Hash)
	require.NoError(t, err)
	opRet, err := txscript.NullDataScript(chain33HashData)
	require.NoError(t, err)
	tx.AddTxOut(wire.NewTxOut(0, opRet))
	p2pkh, err := btcutil.NewAddressWitnessPubKeyHash(btcutil.Hash160(pub1.SerializeCompressed()), &chaincfg.RegressionNetParams)
	require.NoError(t, err)
	p2wpkhScript, err := txscript.PayToAddrScript(p2pkh)
	require.NoError(t, err)
	tx.AddTxOut(wire.NewTxOut(totalAmount-btcFee, p2wpkhScript))

	hashCache := newTxSigHashes(tx, p2wpkhScript, totalAmount)
	witness, err := txscript.WitnessSignature(tx, hashCache, 0,
		totalAmount, p2wpkhScript, txscript.SigHashAll, priv1, true)
	require.NoError(t, err)
	txIn.Witness = witness
	buf := bytes.NewBuffer(make([]byte, 0, tx.SerializeSize()))
	require.NoError(t, tx.Serialize(buf))
	fmt.Println(hex.EncodeToString(buf.Bytes()))
}

func Test_btcExtendKey(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	//desc := "wpkh(tprv8ZgxMBicQKsPdWEKjojF9ZKgvZqo9LnH2gBctcW3rByRkP4UbzJdJJXbhagRmjAtYFPB2ZpqAT8iBTxVXgb24TUnjWMbLFeQpfDzDrAtPWB/84h/1h/0h/0/*)#2rz2j3d5"
	masterPriv := "tprv8ZgxMBicQKsPdWEKjojF9ZKgvZqo9LnH2gBctcW3rByRkP4UbzJdJJXbhagRmjAtYFPB2ZpqAT8iBTxVXgb24TUnjWMbLFeQpfDzDrAtPWB"

	masterKey, err := hdkeychain.NewKeyFromString(masterPriv)
	require.NoError(t, err)
	///84h/1h/0h/0/*
	masterKey, _ = masterKey.Derive(hdkeychain.HardenedKeyStart + 84)
	masterKey, _ = masterKey.Derive(hdkeychain.HardenedKeyStart + 1)
	masterKey, _ = masterKey.Derive(hdkeychain.HardenedKeyStart + 0)
	masterKey, _ = masterKey.Derive(0)

	key, err := masterKey.Derive(1)
	require.NoError(t, err)

	priv, err := key.ECPrivKey()

	require.NoError(t, err)
	fmt.Println(hex.EncodeToString(priv.Serialize()))
	witnessProg := btcutil.Hash160(priv.PubKey().SerializeCompressed())
	addr, err := btcutil.NewAddressWitnessPubKeyHash(witnessProg, &chaincfg.RegressionNetParams)
	require.NoError(t, err)
	script, err := txscript.PayToAddrScript(addr)
	require.NoError(t, err)

	fmt.Println("script:", hex.EncodeToString(script))
	address := addr.String()
	fmt.Println(address)

}

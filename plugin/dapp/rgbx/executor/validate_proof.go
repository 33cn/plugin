package executor

import (
	"bytes"

	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const withdrawCommitmentPrefix = "rgbx:withdraw:"

func (r *rgbx) checkWithdrawConfirm(txHash, confirmHash string, confirm *rtypes.ConfirmTx) error {

	proof := confirm.GetBtcTxProof()
	if proof == nil || len(proof.GetTxData()) == 0 {
		elog.Error("checkWithdrawConfirm empty btc proof", "txHash", txHash, "confirmHash", confirmHash)
		return ErrInvalidBtcTxProof
	}
	var withdrawTx wire.MsgTx
	if err := withdrawTx.DeserializeNoWitness(bytes.NewReader(proof.GetTxData())); err != nil {
		elog.Error("checkWithdrawConfirm decode withdraw tx", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return ErrInvalidBtcTxProof
	}
	if !hasWithdrawCommitment(&withdrawTx, confirm.GetTxHash()) {
		elog.Error("checkWithdrawConfirm commitment mismatch", "txHash", txHash, "confirmHash", confirmHash)
		return ErrInvalidBtcProofCommitment
	}

	blockHashStr, err := btcBlockHashToString(proof.GetBlockHash())
	if err != nil {
		elog.Error("checkWithdrawConfirm invalid block hash bytes", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return ErrInvalidBtcBlockHash
	}
	header, err := r.getBtcHeader(proof.GetBlockHeight())
	if err != nil {
		elog.Error("checkWithdrawConfirm get btc header", "txHash", txHash, "confirmHash", confirmHash,
			"blockHash", blockHashStr, "height", proof.GetBlockHeight(), "err", err)
		return ErrGetBtcHeader
	}
	if header.GetHash() != blockHashStr || header.GetHeight() != proof.GetBlockHeight() {
		elog.Error("checkWithdrawConfirm header mismatch", "txHash", txHash, "confirmHash", confirmHash,
			"expectHash", blockHashStr, "actualHash", header.GetHash(),
			"expectHeight", proof.GetBlockHeight(), "actualHeight", header.GetHeight())
		return ErrInvalidBtcProofBlock
	}

	txID := withdrawTx.TxHash()
	merkleRoot, err := calcBtcMerkleRoot(txID, proof.GetMerkleProof(), uint32(proof.GetTxIndex()))
	if err != nil {
		elog.Error("checkWithdrawConfirm calc merkle root", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return ErrCalcBtcMerkleRoot
	}
	headerMerkleRoot, err := chainhash.NewHashFromStr(header.GetMerkleRoot())
	if err != nil {
		elog.Error("checkWithdrawConfirm invalid header merkleRoot", "txHash", txHash, "confirmHash", confirmHash,
			"merkleRoot", header.GetMerkleRoot(), "err", err)
		return ErrInvalidBtcProofMerkle
	}
	if !bytes.Equal(merkleRoot, headerMerkleRoot.CloneBytes()) {
		elog.Error("checkWithdrawConfirm merkle root not match", "txHash", txHash, "confirmHash", confirmHash,
			"blockHash", blockHashStr)
		return ErrInvalidBtcProofMerkle
	}
	return nil
}

func (r *rgbx) getBtcHeader(height uint64) (*ltypes.BtcHeader, error) {

	msg, err := r.GetAPI().Query(ltypes.LightclientX, "GetBtcHeader", &ltypes.ReqGetBtcHeader{Height: height})
	if err != nil {
		elog.Error("getBtcHeader query", "height", height, "err", err)
		return nil, err
	}
	header, ok := msg.(*ltypes.BtcHeader)
	if !ok || header == nil {
		elog.Error("getBtcHeader invalid header", "height", height)
		return nil, types.ErrInvalidParam
	}
	return header, nil
}

func btcBlockHashToString(hashBytes []byte) (string, error) {
	if len(hashBytes) != chainhash.HashSize {
		return "", types.ErrInvalidParam
	}
	var blockHash chainhash.Hash
	copy(blockHash[:], hashBytes)
	return blockHash.String(), nil
}

func hasWithdrawCommitment(tx *wire.MsgTx, chain33TxHash []byte) bool {
	expectData := append([]byte(withdrawCommitmentPrefix), chain33TxHash...)
	for _, out := range tx.TxOut {
		if len(out.PkScript) == 0 || out.PkScript[0] != txscript.OP_RETURN {
			continue
		}
		pushedData, err := txscript.PushedData(out.PkScript)
		if err != nil {
			continue
		}
		for _, data := range pushedData {
			if bytes.Equal(data, expectData) {
				return true
			}
		}
	}
	return false
}

func calcBtcMerkleRoot(txHash chainhash.Hash, merkleProof [][]byte, txIndex uint32) ([]byte, error) {
	current := txHash.CloneBytes()
	for i, sibling := range merkleProof {
		if len(sibling) != chainhash.HashSize {
			return nil, types.ErrInvalidParam
		}
		data := make([]byte, 0, chainhash.HashSize*2)
		if (txIndex>>uint(i))&1 == 0 {
			data = append(data, current...)
			data = append(data, sibling...)
		} else {
			data = append(data, sibling...)
			data = append(data, current...)
		}
		current = chainhash.DoubleHashB(data)
	}
	return current, nil
}

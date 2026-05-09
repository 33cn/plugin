package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

const withdrawCommitmentPrefix = "rgbx:withdraw:"
const depositCommitmentPrefix = "rgbx:deposit:"

func btcProof2String(txProof *rtypes.BtcTxProof) string {
	return fmt.Sprintf("btcBlockHeight: %d, btcBlockHash: %s, btcTxIndex: %d, btcTxData: %s",
		txProof.GetBlockHeight(), txProof.GetBlockHash(),
		txProof.GetTxIndex(), hex.EncodeToString(txProof.GetTxData()))
}

func merkelProof2String(merkleProof [][]byte) string {
	result := ""
	for _, proof := range merkleProof {
		result += hex.EncodeToString(proof) + "|"
	}
	return result
}

func (r *rgbx) checkWithdrawConfirm(txHash, confirmHash string, confirm *rtypes.ConfirmTx, pendingTx *rtypes.PendingTx) error {
	btcTx, err := r.validateBtcTxProof(txHash, confirm.GetBtcTxProof())
	if err != nil {
		elog.Error("checkWithdrawConfirm validate btc tx proof", "txHash", txHash, "confirmHash", confirmHash,
			"btcProof", btcProof2String(confirm.GetBtcTxProof()), "err", err)
		return err
	}
	if !hasWithdrawCommitment(btcTx, confirm.GetTxHash()) {
		elog.Error("checkWithdrawConfirm commitment mismatch", "txHash", txHash, "confirmHash", confirmHash,
			"btcProof", btcProof2String(confirm.GetBtcTxProof()))
		return ErrInvalidBtcProofCommitment
	}
	if err = r.validateWithdrawTxContent(txHash, pendingTx, btcTx); err != nil {
		return err
	}
	return nil
}

func (r *rgbx) validateDepositTxContent(txHash string, deposit *rtypes.DepositAsset, btcTx *wire.MsgTx) error {
	info, err := r.getCrossChainInfo(deposit.GetAssetSymbol())
	if err != nil {
		elog.Error("validateDepositTxContent getCrossChainInfo", "txHash", txHash, "symbol", deposit.GetAssetSymbol(), "err", err)
		return ErrGetCrossChainInfo
	}
	var amount int64
	for _, out := range btcTx.TxOut {
		if bytes.Equal(out.PkScript, info.GetPkScript()) {
			amount += out.Value
		}
	}
	if amount != deposit.GetAmount() {
		elog.Error("validateDepositTxContent amount mismatch", "txHash", txHash,
			"expect", deposit.GetAmount(), "actual", amount)
		return ErrInvalidDepositAmount
	}
	return nil
}

func (r *rgbx) validateWithdrawTxContent(txHash string, pendingTx *rtypes.PendingTx, btcTx *wire.MsgTx) error {
	if pendingTx == nil {
		return ErrPendingTxNotExist
	}
	info, err := r.getCrossChainInfo(pendingTx.GetAssetSymbol())
	if err != nil {
		elog.Error("validateWithdrawTxContent getCrossChainInfo", "txHash", txHash, "symbol", pendingTx.GetAssetSymbol(), "err", err)
		return ErrInvalidCrossChainInfo
	}
	destScript, err := r.decodeBtcAddressScript(pendingTx.GetTargetAddress())
	if err != nil {
		elog.Error("validateWithdrawTxContent decode target address", "txHash", txHash, "address", pendingTx.GetTargetAddress(), "err", err)
		return ErrInvalidWithdrawDestination
	}
	var destAmount int64
	for _, out := range btcTx.TxOut {
		if len(out.PkScript) > 0 && out.PkScript[0] == txscript.OP_RETURN {
			continue
		}
		if bytes.Equal(out.PkScript, destScript) {
			destAmount += out.Value
			continue
		}
		if !bytes.Equal(out.PkScript, info.GetPkScript()) {
			elog.Error("validateWithdrawTxContent unexpected output script", "txHash", txHash, "script", hex.EncodeToString(out.PkScript))
			return ErrInvalidWithdrawDestinationScript
		}
	}
	if destAmount <= 0 || destAmount > pendingTx.GetAmount() {
		elog.Error("validateWithdrawTxContent dest amount invalid", "txHash", txHash,
			"destAmount", destAmount, "expectAmount", pendingTx.GetAmount())
		return ErrInvalidWithdrawAmount
	}

	return nil
}

func (r *rgbx) getCrossChainInfo(symbol string) (*rtypes.CrossChainInfo, error) {
	if symbol == "" {
		symbol = rtypes.BTCSymbol
	}
	info := &rtypes.CrossChainInfo{}
	err := readDB(r.GetStateDB(), formatCrossChainInfoKey(symbol), info)
	return info, err
}

func (r *rgbx) decodeBtcAddressScript(addr string) ([]byte, error) {
	if addr == "" {
		return nil, types.ErrInvalidAddress
	}
	netName, err := r.getBtcNetName()
	if err != nil {
		return nil, err
	}
	params := ltypes.GetBtcChainParams(netName)
	decoded, err := btcutil.DecodeAddress(addr, params)
	if err != nil {
		return nil, err
	}
	return txscript.PayToAddrScript(decoded)
}

func (r *rgbx) validateBtcTxProof(txHash string, proof *rtypes.BtcTxProof) (*wire.MsgTx, error) {
	if proof == nil || len(proof.GetTxData()) == 0 {
		elog.Error("validateBtcTxProof empty btc proof", "txHash", txHash)
		return nil, ErrInvalidBtcTxProof
	}
	var btcTx wire.MsgTx
	if err := btcTx.DeserializeNoWitness(bytes.NewReader(proof.GetTxData())); err != nil {
		elog.Error("validateBtcTxProof decode btc tx", "txHash", txHash, "err", err)
		return nil, ErrInvalidBtcTxProof
	}

	blockHashStr := proof.GetBlockHash()
	header, err := r.getBtcHeader(proof.GetBlockHeight())
	if err != nil {
		elog.Error("validateBtcTxProof get btc header", "txHash", txHash,
			"blockHash", blockHashStr, "height", proof.GetBlockHeight(), "err", err)
		return nil, ErrGetBtcHeader
	}
	if header.GetHash() != blockHashStr || header.GetHeight() != proof.GetBlockHeight() {
		elog.Error("validateBtcTxProof header mismatch", "txHash", txHash,
			"expectHash", blockHashStr, "actualHash", header.GetHash(),
			"expectHeight", proof.GetBlockHeight(), "actualHeight", header.GetHeight())
		return nil, ErrInvalidBtcProofBlock
	}

	txID := btcTx.TxHash()
	merkleRoot := merkle.GetMerkleRootFromBranch(proof.GetMerkleProof(), txID.CloneBytes(), proof.GetTxIndex())
	headerMerkleRoot, err := chainhash.NewHashFromStr(header.GetMerkleRoot())
	if err != nil {
		elog.Error("validateBtcTxProof invalid header merkleRoot", "txHash", txHash,
			"merkleRoot", header.GetMerkleRoot(), "err", err)
		return nil, ErrInvalidBtcProofMerkle
	}
	if !bytes.Equal(merkleRoot, headerMerkleRoot.CloneBytes()) {
		elog.Error("validateBtcTxProof merkle root not match", "txHash", txHash,
			"expectMerkleRoot", header.GetMerkleRoot(), "actualMerkleRoot", hex.EncodeToString(merkleRoot),
			"merkleProof", merkelProof2String(proof.GetMerkleProof()))
		return nil, ErrInvalidBtcProofMerkle
	}
	return &btcTx, nil
}

func (r *rgbx) getBtcNetName() (string, error) {
	msg, err := r.GetAPI().Query(ltypes.LightclientX, "GetBtcNetName", &types.ReqNil{})
	if err != nil {
		elog.Error("getBtcNetName query", "err", err)
		return "", err
	}
	return msg.(*types.ReplyString).Data, nil
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

func hasWithdrawCommitment(tx *wire.MsgTx, chain33TxHash []byte) bool {
	expectData := append([]byte(withdrawCommitmentPrefix), chain33TxHash...)
	return hasExpectedOpReturnData(tx, expectData)
}

func hasDepositCommitment(tx *wire.MsgTx, depositAddress string) bool {
	if rtypes.IsUtxoAddress(depositAddress) {
		if len(tx.TxIn) > 0 {
			firstInputUtxo := tx.TxIn[0].PreviousOutPoint
			return depositAddress == rtypes.FormatUtxo(firstInputUtxo.Hash.String(), firstInputUtxo.Index)
		}
		return false
	}
	expectData := append([]byte(depositCommitmentPrefix), []byte(depositAddress)...)
	return hasExpectedOpReturnData(tx, expectData)
}

func hasExpectedOpReturnData(tx *wire.MsgTx, expectData []byte) bool {
	expectScript, err := txscript.NullDataScript(expectData)
	if err != nil {
		elog.Error("hasExpectedOpReturnData null data script", "err", err)
		return false
	}
	for _, out := range tx.TxOut {
		if len(out.PkScript) == 0 || out.PkScript[0] != txscript.OP_RETURN {
			continue
		}
		if bytes.Equal(out.PkScript, expectScript) {
			return true
		}
	}
	return false
}

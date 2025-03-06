package executor

import (
	"bytes"
	"encoding/hex"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

var ()

// CheckTx 实现自定义检验交易接口，供框架调用
func (r *rgbx) CheckTx(tx *types.Transaction, index int) error {

	txHash := hex.EncodeToString(tx.Hash())
	action := &rtypes.RgbxAction{}
	err := types.Decode(tx.GetPayload(), action)
	if err != nil {
		elog.Error("CheckTx", "txHash", txHash, "Decode payload error", err)
		return types.ErrDecode
	}

	switch action.Ty {
	case rtypes.TyMintAction:

	case rtypes.TyTransferAction:
	case rtypes.TyConfirmAction:
	default:
		err = types.ErrActionNotSupport

	}
	if err != nil {
		elog.Error("rgbx CheckTx", "txHash", txHash, "actionName", tx.ActionName(), "err", err)
	}
	return err
}

func (r *rgbx) checkMint(txHash string, mint *rtypes.MintAsset) error {

	if len(mint.GetSymbol()) <= 1 || len(mint.GetSymbol()) >= MaxAssetSymbolLength {
		elog.Error("checkMint", "txHash", txHash,
			"symbol", mint.Symbol, "symbolLen", len(mint.GetSymbol()))
		return ErrInvalidSymbolLength
	}

	ty := Type(mint.GetType())
	if (ty != Normal && mint.GetTotalAmount() != 1) ||
		mint.GetTotalAmount() > MaxAssetAmount {
		elog.Error("checkMint", "txHash", txHash, "symbol", mint.Symbol,
			"amount", mint.GetTotalAmount(), "type", ty.String())
		return ErrInvalidAssetAmount
	}

	if len(mint.GetMetaHash()) != MetaHashLen {
		elog.Error("checkMint", "txHash", txHash, "symbol", mint.Symbol,
			"metaHashLen", len(mint.GetMetaHash()))
		return ErrInvalidMetaHashLength
	}
	if mint.GetGenesisOut() == nil {
		elog.Error("checkMint nil out", "txHash", txHash, "symbol", mint.Symbol)
		return ErrNilGenesisOut
	}
	_, err := r.GetStateDB().Get(formatAssetKey(mint.GetSymbol()))
	if types.ErrNotFound != err {
		elog.Error("checkMint duplicate asset", "txHash", txHash, "symbol", mint.Symbol)
		return ErrDuplicateAssetSymbol
	}

	return nil
}

func (r *rgbx) checkTransfer(txHash string, transfer *rtypes.TransferAsset) error {

	if transfer.GetFrom().Address() == "" || transfer.GetTo().Address() == "" {
		elog.Error("checkTransfer address", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"from", transfer.GetFrom().Address(), "to", transfer.GetTo().Address())
		return types.ErrInvalidAddress
	}

	_, err := r.GetStateDB().Get(formatAssetKey(transfer.GetSymbol()))
	if err != nil {
		elog.Error("checkTransfer get asset", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"err", err)
		return ErrAssetNotExist
	}

	return nil
}

// TODO check from address
func (r *rgbx) checkConfirm(txHash string, confirm *rtypes.ConfirmTx) error {

	confirmTxHash := hex.EncodeToString(confirm.TxHash)
	action := rtypes.GetActionName(confirm.GetActionType())

	pendingTx := &rtypes.PendingTx{}
	err := readDB(r.GetLocalDB(), formatPendingTxKey(confirm.TxBlockHeight, confirm.TxIndex), pendingTx)
	if err != nil {
		elog.Error("checkConfirm read pending tx", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash,
			"height", confirm.TxBlockHeight, "index", confirm.TxIndex, "err", err)
		return ErrPendingTxNotExist
	}

	if pendingTx.Confirmed {
		elog.Error("checkConfirm tx already confirmed", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash)
		return ErrTxAlreadyConfirmed
	}

	_, err = r.GetStateDB().Get(formatPayloadKey(confirm.GetTxHash()))
	if !bytes.Equal(confirm.GetTxHash(), pendingTx.GetTxHash()) {
		elog.Error("checkConfirm tx hash not equal", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash,
			"expectConfirmHash", hex.EncodeToString(pendingTx.GetTxHash()))
		return ErrConfirmedHashNotEqual
	}

	if confirm.Timeout {
		elog.Debug("checkConfirm timeout", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash)
		return nil
	}

	spendingTxHash := chainhash.DoubleHashH(confirm.GetProof().GetSpendingTx()).String()

	spendingTx := wire.MsgTx{}
	err = spendingTx.DeserializeNoWitness(bytes.NewReader(confirm.GetProof().GetSpendingTx()))
	if err != nil {
		elog.Error("checkConfirm decode spending tx", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"rawSpendingTx", hex.EncodeToString(confirm.GetProof().GetSpendingTx()),
			"decode err", err)
		return ErrDecodeBtcTx
	}

	// check input
	expectInput := pendingTx.Utxo.ToString()
	actualInput := spendingTx.TxIn[int(confirm.GetProof().GetSpendingInputIdx())].PreviousOutPoint.String()
	if expectInput != actualInput {
		elog.Error("checkConfirm input utxo not equal", "action", action,
			"expectInput", expectInput, "actualInput", actualInput)
		return ErrSpendingInputNotEqual
	}

	// 表示op_return输出不存在，即utxo已经在btc链花费, 但没有构建rgbx所约束的op_return输出
	if confirm.GetProof().GetOpRetOutputIdx() < 0 {
		elog.Debug("checkConfirm opReturn output not exist",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "spendingTxHash", spendingTxHash)
		return nil
	}

	// 提供的op_return pkScript参数非法，和btc原始交易中的输出不符
	if !bytes.Equal(confirm.GetProof().OpRetOutputPkScript,
		spendingTx.TxOut[int(confirm.GetProof().GetOpRetOutputIdx())].PkScript) {
		elog.Error("checkConfirm opReturn pkScript not equal",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "spendingTxHash", spendingTxHash)
		return ErrOpRetOutputPkScriptNotEqual
	}

	//check op_return commitment
	commitment, _ := txscript.NullDataScript(confirm.GetTxHash())
	opRetOutScript := spendingTx.TxOut[int(confirm.GetProof().GetOpRetOutputIdx())].PkScript
	if !bytes.Equal(opRetOutScript, commitment) {

		elog.Error("checkConfirm op return commitment", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash,
			"spendingTxHash", spendingTxHash, "commit", hex.EncodeToString(opRetOutScript),
			"expectCommit", hex.EncodeToString(commitment))
		return ErrInvalidOpRetCommitment
	}

	return nil
}

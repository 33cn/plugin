package executor

import (
	"bytes"
	"encoding/hex"
	"errors"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var (
	ErrInvalidSymbolLength         = errors.New("invalid asset symbol length")
	ErrInvalidAssetAmount          = errors.New("invalid asset amount")
	ErrInvalidMetaHashLength       = errors.New("invalid meta hash length")
	ErrNilGenesisOut               = errors.New("nil genesis output")
	ErrDuplicateAssetSymbol        = errors.New("duplicate asset symbol")
	ErrAssetNotExist               = errors.New("asset not exist")
	ErrDecodeBtcTx                 = errors.New("decode btc tx error")
	ErrPendingTxNotExist           = errors.New("pending tx not exist")
	ErrTxAlreadyConfirmed          = errors.New("tx already confirmed")
	ErrConfirmedHashNotEqual       = errors.New("confirmed hash not equal")
	ErrSpendingInputNotEqual       = errors.New("spending input not equal")
	ErrOpRetOutputPkScriptNotEqual = errors.New("ErrOpRetOutputPkScriptNotEqual")
	ErrInvalidCommitAddress        = errors.New("ErrInvalidCommitAddress")
	ErrFromUtxoPkScriptNotSet      = errors.New("ErrFromUtxoPkScriptNotSet")
	ErrInvalidAssetPrecision       = errors.New("ErrInvalidAssetPrecision")
	ErrInvalidAssetSender          = errors.New("ErrInvalidAssetSender")
	ErrInvalidSpendingTxIn         = errors.New("ErrInvalidSpendingTxIn")
)

// CheckTx 实现自定义检验交易接口，供框架调用
func (r *rgbx) CheckTx(tx *types.Transaction, index int) error {

	txHash := hex.EncodeToString(tx.Hash())
	action := &rtypes.RgbxAction{}
	err := types.Decode(tx.GetPayload(), action)
	if err != nil {
		elog.Error("CheckTx", "txHash", txHash, "Decode payload error", err)
		return types.ErrActionNotSupport
	}

	switch action.Ty {
	case rtypes.TyMintAction:
		err = r.checkMint(txHash, action.GetMint())
	case rtypes.TyTransferAction:
		err = r.checkTransfer(tx, txHash, action.GetTransfer())
	case rtypes.TyConfirmAction:
		err = r.checkConfirm(tx.From(), txHash, action.GetConfirm())
	default:
		err = types.ErrActionNotSupport

	}
	if err != nil {
		elog.Error("rgbx CheckTx", "txHash", txHash, "actionName", tx.ActionName(),
			"err", err, "action", string(types.MustPBToJSON(action)))
	}
	return err
}

func (r *rgbx) checkMint(txHash string, mint *rtypes.MintAsset) error {

	if len(mint.GetSymbol()) < 1 || len(mint.GetSymbol()) > rtypes.MaxAssetSymbolLength {
		elog.Error("checkMint", "txHash", txHash,
			"symbol", mint.Symbol, "symbolLen", len(mint.GetSymbol()))
		return ErrInvalidSymbolLength
	}

	ty := rtypes.AssetType(mint.GetType())
	if mint.GetTotalAmount() <= 0 || mint.GetTotalAmount() > rtypes.MaxAssetAmount ||
		(ty == rtypes.Collectible && mint.GetTotalAmount() != 1) {
		elog.Error("checkMint", "txHash", txHash, "symbol", mint.Symbol,
			"amount", mint.GetTotalAmount(), "type", ty.String())
		return ErrInvalidAssetAmount
	}
	if ty != rtypes.Collectible && mint.GetPrecision() > rtypes.MaxPrecision {
		elog.Error("checkMint", "txHash", txHash, "symbol", mint.Symbol,
			"precision", mint.GetPrecision(), "maxPrecision", rtypes.MaxPrecision)
		return ErrInvalidAssetPrecision
	}

	if len(mint.GetMetaHash()) > rtypes.MetaHashLen {
		elog.Error("checkMint", "txHash", txHash, "symbol", mint.Symbol,
			"metaHashLen", len(mint.GetMetaHash()))
		return ErrInvalidMetaHashLength
	}

	_, err := r.GetStateDB().Get(formatAssetKey(mint.GetSymbol()))
	if !errors.Is(err, types.ErrNotFound) {
		elog.Error("checkMint duplicate asset", "txHash", txHash, "symbol", mint.Symbol)
		return ErrDuplicateAssetSymbol
	}

	if mint.GetGenesisOut().GetHash() == "" || mint.GetGenesisOut().GetPkScript() == nil {
		elog.Error("checkMint invalid genesis out", "txHash", txHash, "symbol", mint.Symbol)
		return ErrNilGenesisOut
	}

	return nil
}

func (r *rgbx) checkTransfer(tx *types.Transaction, txHash string, transfer *rtypes.TransferAsset) error {

	fromAddr := transfer.GetFrom()
	if fromAddr == "" {
		fromAddr = tx.From()
	}
	if address.CheckAddress(fromAddr, -1) != nil ||
		address.CheckAddress(transfer.GetTo(), -1) != nil ||
		(transfer.GetChangeAddr() != "" && address.CheckAddress(transfer.GetChangeAddr(), -1) != nil) {
		elog.Error("checkTransfer address", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"from", fromAddr, "to", transfer.GetTo(), "changeAddr", transfer.GetChangeAddr())
		return types.ErrInvalidAddress
	}

	if rtypes.IsUtxoAddress(fromAddr) && len(transfer.GetFromPkScript()) == 0 {
		elog.Error("checkTransfer pkScript is nil", "txHash", txHash,
			"symbol", transfer.GetSymbol(), "from", transfer.GetFrom())
		return ErrFromUtxoPkScriptNotSet
	}

	asset := &rtypes.RgbxAsset{}
	err := readDB(r.GetStateDB(), formatAssetKey(transfer.GetSymbol()), asset)
	if err != nil {
		elog.Error("checkTransfer get asset", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"err", err)
		return ErrAssetNotExist
	}

	assetTy := rtypes.AssetType(asset.GetType())
	if assetTy == rtypes.Normal &&
		(transfer.GetAmount() > asset.GetTotalAmount() || transfer.GetAmount() <= 0) {
		elog.Error("checkTransfer", "txHash", txHash,
			"symbol", transfer.GetSymbol(), "amount", transfer.GetAmount(), "total", asset.GetTotalAmount())
		return ErrInvalidAssetAmount
	}

	if assetTy == rtypes.Collectible && fromAddr != asset.Owner {
		elog.Error("checkTransfer invalid owner", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"from", fromAddr, "assetOwner", asset.Owner)
		return ErrInvalidAssetSender
	}

	return nil
}

func (r *rgbx) checkConfirm(fromAddr, txHash string, confirm *rtypes.ConfirmTx) error {

	confirmTxHash := hex.EncodeToString(confirm.TxHash)
	action := rtypes.GetActionName(confirm.GetActionType())

	if fromAddr != rgbxCfg.CommitAddress {
		elog.Error("checkConfirm fromAddr", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash,
			"fromAddr", fromAddr, "commitAddr", rgbxCfg.CommitAddress)
		return ErrInvalidCommitAddress
	}

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

	btcSpendHash := chainhash.DoubleHashH(confirm.GetProof().GetSpendingTx()).String()
	spendingTx := wire.MsgTx{}
	err = spendingTx.DeserializeNoWitness(bytes.NewReader(confirm.GetProof().GetSpendingTx()))
	if err != nil {
		elog.Error("checkConfirm decode spending tx", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"btcSpendingTx", hex.EncodeToString(confirm.GetProof().GetSpendingTx()),
			"decode err", err)
		return ErrDecodeBtcTx
	}

	spendingInputIdx := int(confirm.GetProof().GetSpendingInputIdx())
	if spendingInputIdx >= len(spendingTx.TxIn) {
		elog.Error("checkConfirm spending tx input", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"inputIdx", spendingInputIdx, "txInLen", len(spendingTx.TxIn), "btcSpendHash", btcSpendHash)
		return ErrInvalidSpendingTxIn
	}

	// check input
	expectInput := pendingTx.Utxo.ToString()
	actualInput := spendingTx.TxIn[int(confirm.GetProof().GetSpendingInputIdx())].PreviousOutPoint.String()
	if expectInput != actualInput {
		elog.Error("checkConfirm input utxo not equal", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"expectInput", expectInput, "actualInput", actualInput, "btcSpendHash", btcSpendHash)
		return ErrSpendingInputNotEqual
	}

	opRetOutIdx := int(confirm.GetProof().GetOpRetOutputIdx())
	// 表示op_return输出不存在，即utxo已经在btc链花费, 但没有构建rgbx所约束的op_return输出
	if opRetOutIdx < 0 || opRetOutIdx >= len(spendingTx.TxOut) {
		elog.Debug("checkConfirm opReturn output not exist",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "btcSpendHash", btcSpendHash)
		return nil
	}

	// 提供的op_return pkScript参数非法，和btc原始交易中的输出不符
	if !bytes.Equal(confirm.GetProof().OpRetOutputPkScript,
		spendingTx.TxOut[int(confirm.GetProof().GetOpRetOutputIdx())].PkScript) {
		elog.Error("checkConfirm opReturn pkScript not equal",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "btcSpendHash", btcSpendHash)
		return ErrOpRetOutputPkScriptNotEqual
	}

	return nil
}

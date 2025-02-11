package executor

import (
	"bytes"
	"encoding/hex"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
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

	val, err := r.GetStateDB().Get(formatPayloadKey(confirm.GetTxHash()))

	if err != nil {
		elog.Error("checkConfirm get payload", "txHash", txHash,
			"confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"action", rtypes.GetActionName(confirm.GetActionType()), "err", err)
		return ErrConfirmedTxNotExist
	}

	if confirm.Timeout {
		elog.Debug("checkConfirm timeout", "txHash", txHash,
			"confirmTxHash", hex.EncodeToString(confirm.GetTxHash()))
		return nil
	}

	//TODO check op_return commitment
	if !bytes.Equal(confirm.Witness.GetOpRetOut().PkScript, confirm.GetTxHash()) {

		elog.Error("checkConfirm op return commitment", "txHash", txHash,
			"confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"opCommitment", hex.EncodeToString(confirm.Witness.GetOpRetOut().PkScript))
		return ErrInvalidOpRetCommitment
	}

	if confirm.ActionType == rtypes.TyMintAction {
		return r.checkConfirmMint(txHash, confirm, val)
	} else if confirm.ActionType == rtypes.TyTransferAction {

	}

	return nil
}

func (r *rgbx) checkConfirmMint(txHash string, confirm *rtypes.ConfirmTx, payload []byte) error {

	mint := &rtypes.MintAsset{}
	err := types.Decode(payload, mint)
	if err != nil {
		elog.Error("checkConfirmMint decode payload", "txHash", txHash,
			"confirmTxHash", hex.EncodeToString(confirm.GetTxHash()))
		return types.ErrDecode
	}

	if mint.GetGenesisOut().ToString() != confirm.GetWitness().GetIn().ToString() {
		elog.Error("checkConfirmMint genesis out", "txHash", txHash,
			"confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"expect", mint.GetGenesisOut(), "actual", confirm.GetWitness().GetIn())
		return ErrGenesisOutNotEqual
	}
	return nil
}

func (r *rgbx) checkConfirmTransfer(txHash string, transfer *rtypes.TransferAsset, payload []byte) error {

	return nil
}

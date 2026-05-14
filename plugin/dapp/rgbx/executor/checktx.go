package executor

import (
	"bytes"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

var (
	ErrInvalidSymbolLength              = errors.New("invalid asset symbol length")
	ErrInvalidAssetAmount               = errors.New("invalid asset amount")
	ErrInvalidMetaHashLength            = errors.New("invalid meta hash length")
	ErrNilGenesisOut                    = errors.New("nil genesis output")
	ErrDuplicateAssetSymbol             = errors.New("duplicate asset symbol")
	ErrAssetNotExist                    = errors.New("asset not exist")
	ErrDecodeBtcTx                      = errors.New("decode btc tx error")
	ErrPendingTxNotExist                = errors.New("pending tx not exist")
	ErrConfirmPayloadNotExist           = errors.New("confirm payload not exist")
	ErrTxAlreadyConfirmed               = errors.New("tx already confirmed")
	ErrConfirmedHashNotEqual            = errors.New("confirmed hash not equal")
	ErrSpendingInputNotEqual            = errors.New("spending input not equal")
	ErrOpRetOutputPkScriptNotEqual      = errors.New("ErrOpRetOutputPkScriptNotEqual")
	ErrInvalidCommitAddress             = errors.New("ErrInvalidCommitAddress")
	ErrFromUtxoPkScriptNotSet           = errors.New("ErrFromUtxoPkScriptNotSet")
	ErrInvalidAssetPrecision            = errors.New("ErrInvalidAssetPrecision")
	ErrInvalidAssetSender               = errors.New("ErrInvalidAssetSender")
	ErrInvalidFromUtxo                  = errors.New("invalid from utxo")
	ErrInvalidSpendingTxIn              = errors.New("ErrInvalidSpendingTxIn")
	ErrInvalidWithdrawAmount            = errors.New("invalid withdraw amount")
	ErrInvalidWithdrawDestination       = errors.New("invalid withdraw destination")
	ErrInvalidWithdrawDestinationScript = errors.New("invalid withdraw destination script")
	ErrInvalidDepositAmount             = errors.New("invalid deposit amount")
	ErrInvalidDepositAddress            = errors.New("invalid deposit address")
	ErrInvalidDepositCommitment         = errors.New("invalid deposit opreturn commitment")
	ErrInvalidWithdrawFeeRate           = errors.New("invalid withdraw fee rate")
	ErrInvalidAssetSymbol               = errors.New("invalid asset symbol")
	ErrInvalidBtcTxProof                = errors.New("invalid btc tx proof")
	ErrWithdrawConfirmTimeoutNotAllowed = errors.New("withdraw confirm timeout not allowed")
	ErrInvalidBtcProofIndex             = errors.New("invalid btc proof tx index")
	ErrInvalidBtcBlockHash              = errors.New("invalid btc block hash")
	ErrGetBtcHeader                     = errors.New("get btc header error")
	ErrInvalidBtcProofBlock             = errors.New("invalid btc proof block info")
	ErrInvalidBtcProofCommitment        = errors.New("invalid btc withdraw commitment")
	ErrInvalidBtcProofMerkle            = errors.New("invalid btc merkle proof")
	ErrCalcBtcMerkleRoot                = errors.New("calc btc merkle root error")
	ErrInvalidCrossChainInfo            = errors.New("invalid cross chain info")
	ErrNewAccountDB                     = errors.New("new account db error")
	ErrGetCrossChainInfo                = errors.New("get cross chain info error")
	ErrDuplicateDepositProof            = errors.New("duplicate deposit proof")
	ErrInvalidGuardianCommitter         = errors.New("invalid guardian committer")
	ErrDuplicateDKGCommit               = errors.New("duplicate dkg commit")
	ErrGetGuardianNodeAddress           = errors.New("get guardian node address error")
	ErrGetDkgConfirmations              = errors.New("get dkg confirmations error")
	ErrInvalidDkgAddress                = errors.New("invalid dkg address")
)

const (
	maxBtcFeeRate        = int64(1000)
	minBtcWithdrawAmount = int64(546)
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
	case rtypes.TyCommitDKGAction:
		err = r.checkCommitDKG(txHash, tx.From(), action.GetCommitDKG())
	case rtypes.TyDepositAsset:
		err = r.checkDeposit(txHash, action.GetDeposit())
	case rtypes.TyWithDrawAsset:
		err = r.checkWithdraw(tx.From(), txHash, action.GetWithdraw())
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

	if isCrossChainSymbol(mint.GetSymbol()) {
		return ErrInvalidAssetSymbol
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

	if transfer.GetAmount() <= 0 {
		elog.Error("checkTransfer amount", "txHash", txHash, "symbol", transfer.GetSymbol(), "amount", transfer.GetAmount())
		return ErrInvalidAssetAmount
	}
	fromAddr := tx.From()
	if isCrossChainSymbol(transfer.GetSymbol()) {
		return r.checkCrossChainTransfer(txHash, fromAddr, transfer)
	}
	fromUtxo := transfer.GetFromUtxo()
	if fromUtxo != "" {
		if !rtypes.IsUtxoAddress(fromUtxo) || len(transfer.GetFromUtxoPkScript()) == 0 {
			elog.Error("checkTransfer invalid fromUtxo", "txHash", txHash, "symbol", transfer.GetSymbol(), "fromUtxo", fromUtxo)
			return ErrInvalidFromUtxo
		}
		fromAddr = fromUtxo
	}
	if address.CheckAddress(transfer.GetTo(), -1) != nil ||
		(transfer.GetChangeAddr() != "" && address.CheckAddress(transfer.GetChangeAddr(), -1) != nil) {
		elog.Error("checkTransfer address", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"from", fromAddr, "to", transfer.GetTo(), "changeAddr", transfer.GetChangeAddr())
		return types.ErrInvalidAddress
	}

	asset := &rtypes.RgbxAsset{}
	err := readDB(r.GetStateDB(), formatAssetKey(transfer.GetSymbol()), asset)
	if err != nil {
		elog.Error("checkTransfer get asset", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"err", err)
		return ErrAssetNotExist
	}

	assetTy := rtypes.AssetType(asset.GetType())
	if assetTy == rtypes.Normal {
		accDb, err := r.newAccount(transfer.GetSymbol())
		if err != nil {
			elog.Error("checkTransfer newAccount", "txHash", txHash, "symbol", transfer.GetSymbol(), "err", err)
			return ErrNewAccountDB
		}
		balance := accDb.LoadAccount(fromAddr).GetBalance()
		if balance < transfer.GetAmount() {
			elog.Error("checkTransfer insufficient balance", "txHash", txHash, "from", fromAddr,
				"symbol", transfer.GetSymbol(), "need", transfer.GetAmount(), "balance", balance)
			return types.ErrInsufficientBalance
		}
	} else if fromAddr != asset.Owner {
		elog.Error("checkTransfer invalid owner", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"from", fromAddr, "assetOwner", asset.Owner)
		return ErrInvalidAssetSender
	}
	return nil
}

func (r *rgbx) checkCrossChainTransfer(txHash, fromAddr string, transfer *rtypes.TransferAsset) error {

	accDB, err := r.newAccount(transfer.GetSymbol())
	if err != nil {
		elog.Error("checkCrossChainTransfer newCrossChainAccount", "txHash", txHash, "symbol", transfer.GetSymbol(), "err", err)
		return ErrNewAccountDB
	}
	if accDB.LoadAccount(fromAddr).GetBalance() < transfer.GetAmount() {
		elog.Error("checkCrossChainTransfer insufficient balance", "txHash", txHash, "from", fromAddr,
			"symbol", transfer.GetSymbol(), "need", transfer.GetAmount())
		return types.ErrInsufficientBalance
	}
	return nil
}

func (r *rgbx) checkCommitDKG(txHash, fromAddr string, commitDKG *rtypes.CommitDKG) error {

	symbol := commitDKG.GetAssetSymbol()
	pkScript, err := r.decodeBtcAddressScript(commitDKG.GetDkgAddress())
	if err != nil || !bytes.Equal(pkScript, commitDKG.GetPkScript()) {
		elog.Error("checkCommitDKG decode btc address script", "txHash", txHash,
			"symbol", symbol, "dkgAddress", commitDKG.GetDkgAddress(),
			"pkScript", hex.EncodeToString(commitDKG.GetPkScript()), "expectPkScript", hex.EncodeToString(pkScript), "err", err)
		return ErrInvalidDkgAddress
	}
	guardianAddrs, err := r.getGuardianNodeAddress(rgbxCfg.GuardianParachainTitle)
	if err != nil {
		elog.Error("checkCommitDKG getGuardianNodeAddress", "txHash", txHash, "symbol", symbol, "err", err)
		return ErrGetGuardianNodeAddress
	}
	if !strings.Contains(guardianAddrs, fromAddr) {
		elog.Error("checkCommitDKG invalid committer", "txHash", txHash, "symbol", symbol, "fromAddr", fromAddr)
		return ErrInvalidGuardianCommitter
	}

	_, err = r.GetStateDB().Get(formatCrossChainInfoKey(symbol))
	if err == nil {
		elog.Error("checkCommitDKG duplicate cross chain info", "txHash", txHash, "symbol", symbol)
		return ErrDuplicateDKGCommit
	}
	return nil
}

func (r *rgbx) checkWithdraw(fromAddr, txHash string, withdraw *rtypes.WithdrawAsset) error {

	symbol := ensureCrossChainSymbol(withdraw.GetAssetSymbol())
	if withdraw.GetAmount() < minBtcWithdrawAmount {
		elog.Error("checkWithdraw amount", "txHash", txHash, "amount", withdraw.GetAmount())
		return ErrInvalidWithdrawAmount
	}

	if _, err := r.decodeBtcAddressScript(withdraw.GetDestinationAddr()); err != nil {
		elog.Error("checkWithdraw invalid btc destination", "txHash", txHash, "address", withdraw.GetDestinationAddr(), "err", err)
		return ErrInvalidWithdrawDestination
	}
	if withdraw.GetFeeRate() < 1 || withdraw.GetFeeRate() > maxBtcFeeRate {
		elog.Error("checkWithdraw feeRate", "txHash", txHash, "feeRate", withdraw.GetFeeRate())
		return ErrInvalidWithdrawFeeRate
	}
	accDB, err := r.newAccount(symbol)
	if err != nil {
		elog.Error("checkWithdraw newAccount", "txHash", txHash, "symbol", withdraw.GetAssetSymbol(), "err", err)
		return err
	}
	balance := accDB.LoadAccount(fromAddr).GetBalance()
	if balance < withdraw.GetAmount() {
		elog.Error("checkWithdraw insufficient balance", "txHash", txHash, "from", fromAddr,
			"symbol", symbol, "need", withdraw.GetAmount(), "balance", balance)
		return types.ErrInsufficientBalance
	}
	return nil
}

func (r *rgbx) checkDeposit(txHash string, deposit *rtypes.DepositAsset) error {
	if deposit.GetAmount() <= 0 {
		elog.Error("checkDeposit amount", "txHash", txHash, "amount", deposit.GetAmount())
		return ErrInvalidDepositAmount
	}
	addr := deposit.GetDepositAddress()
	if addr == "" || (!rtypes.IsUtxoAddress(addr) && address.CheckAddress(addr, -1) != nil) {
		elog.Error("checkDeposit address invalid", "txHash", txHash, "address", addr)
		return ErrInvalidDepositAddress
	}
	_, err := r.GetStateDB().Get(formatDepositUsedKey(deposit.GetTxProof().GetTxData()))
	if !errors.Is(err, types.ErrNotFound) {
		elog.Error("checkDeposit duplicate proof", "txHash", txHash, "symbol", deposit.GetAssetSymbol(), "err", err)
		return ErrDuplicateDepositProof
	}
	btcTx, err := r.validateBtcTxProof(txHash, deposit.GetTxProof())
	if err != nil {
		elog.Error("checkDeposit validate btc tx proof", "txHash", txHash, "btcProof", btcProof2String(deposit.GetTxProof()), "err", err)
		return err
	}
	if !hasDepositCommitment(btcTx, addr) {
		elog.Error("checkDeposit commitment mismatch", "txHash", txHash, "depositAddress", addr)
		return ErrInvalidDepositCommitment
	}
	if err = r.validateDepositTxContent(txHash, deposit, btcTx); err != nil {
		return err
	}
	return nil
}

func (r *rgbx) checkConfirm(fromAddr, txHash string, confirm *rtypes.ConfirmTx) error {

	confirmTxHash := hex.EncodeToString(confirm.TxHash)
	action := rtypes.GetActionName(confirm.GetActionType())

	if rgbxCfg.CommitAddress != "" && fromAddr != rgbxCfg.CommitAddress {
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
	if err != nil {
		elog.Error("checkConfirm get payload", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash, "err", err)
		return ErrConfirmPayloadNotExist
	}
	if !bytes.Equal(confirm.GetTxHash(), pendingTx.GetTxHash()) {
		elog.Error("checkConfirm tx hash not equal", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash,
			"expectConfirmHash", hex.EncodeToString(pendingTx.GetTxHash()))
		return ErrConfirmedHashNotEqual
	}

	if confirm.GetActionType() == rtypes.TyWithDrawAsset {
		if confirm.Timeout {
			elog.Error("checkConfirm timeout not supported for withdraw", "action", action,
				"txHash", txHash, "confirmTxHash", confirmTxHash)
			return ErrWithdrawConfirmTimeoutNotAllowed
		}
		return r.checkWithdrawConfirm(txHash, confirmTxHash, confirm, pendingTx)
	}

	if confirm.Timeout {
		elog.Debug("checkConfirm timeout", "action", action,
			"txHash", txHash, "confirmTxHash", confirmTxHash)
		return nil
	}

	btcSpendHash := chainhash.DoubleHashH(confirm.GetUtxoProof().GetSpendingTx()).String()
	spendingTx := wire.MsgTx{}
	err = spendingTx.DeserializeNoWitness(bytes.NewReader(confirm.GetUtxoProof().GetSpendingTx()))
	if err != nil {
		elog.Error("checkConfirm decode spending tx", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"btcSpendingTx", hex.EncodeToString(confirm.GetUtxoProof().GetSpendingTx()),
			"decode err", err)
		return ErrDecodeBtcTx
	}

	spendingInputIdx := int(confirm.GetUtxoProof().GetSpendingInputIdx())
	if spendingInputIdx >= len(spendingTx.TxIn) {
		elog.Error("checkConfirm spending tx input", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"inputIdx", spendingInputIdx, "txInLen", len(spendingTx.TxIn), "btcSpendHash", btcSpendHash)
		return ErrInvalidSpendingTxIn
	}

	// check input
	expectInput := pendingTx.Utxo.ToString()
	actualInput := spendingTx.TxIn[int(confirm.GetUtxoProof().GetSpendingInputIdx())].PreviousOutPoint.String()
	if expectInput != actualInput {
		elog.Error("checkConfirm input utxo not equal", "action", action,
			"txHash", txHash, "confirmTxHash", hex.EncodeToString(confirm.GetTxHash()),
			"expectInput", expectInput, "actualInput", actualInput, "btcSpendHash", btcSpendHash)
		return ErrSpendingInputNotEqual
	}

	opRetOutIdx := int(confirm.GetUtxoProof().GetOpRetOutputIdx())
	// 表示op_return输出不存在，即utxo已经在btc链花费, 但没有构建rgbx所约束的op_return输出
	if opRetOutIdx < 0 || opRetOutIdx >= len(spendingTx.TxOut) {
		elog.Debug("checkConfirm opReturn output not exist",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "btcSpendHash", btcSpendHash)
		return nil
	}

	// 提供的op_return pkScript参数非法，和btc原始交易中的输出不符
	if !bytes.Equal(confirm.GetUtxoProof().OpRetOutputPkScript,
		spendingTx.TxOut[int(confirm.GetUtxoProof().GetOpRetOutputIdx())].PkScript) {
		elog.Error("checkConfirm opReturn pkScript not equal",
			"action", action, "txHash", txHash,
			"confirmTxHash", confirmTxHash, "btcSpendHash", btcSpendHash)
		return ErrOpRetOutputPkScriptNotEqual
	}

	return nil
}

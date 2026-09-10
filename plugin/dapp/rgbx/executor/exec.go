package executor

import (
	"bytes"
	"encoding/hex"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (r *rgbx) Exec_Mint(mint *rtypes.MintAsset, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}

	txHash := hex.EncodeToString(tx.Hash())
	elog.Debug("Exec_Mint", "txHash", txHash, "symbol", mint.Symbol,
		"amount", mint.TotalAmount, "gensisOut", mint.GetGenesisOut().ToString())
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatPayloadKey(tx.Hash()),
		Value: types.Encode(mint),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: rtypes.TyPendingTxLog,
		Log: types.Encode(&rtypes.PendingTx{
			ActionType:    rtypes.TyMintAction,
			Timestamp:     r.GetBlockTime(),
			TxBlockHeight: r.GetHeight(),
			TxIndex:       int64(index),
			TxHash:        tx.Hash(),
			Utxo:          mint.GetGenesisOut(),
		}),
	})

	return receipt, nil
}

func (r *rgbx) Exec_Transfer(transfer *rtypes.TransferAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	txHash := hex.EncodeToString(tx.Hash())

	fromAddr := tx.From()
	elog.Debug("Exec_Transfer", "txHash", txHash, "symbol", transfer.Symbol, "amount", transfer.Amount,
		"from", fromAddr, "to", transfer.GetTo(),
		"changeAddr", transfer.GetChangeAddr(), "fromUtxo", transfer.GetFromUtxo())

	if isCrossChainSymbol(transfer.GetSymbol()) {
		accDB, err := r.newAccount(transfer.GetSymbol())
		if err != nil {
			elog.Error("Exec_Transfer newCrossChainAccount", "txHash", txHash, "from", tx.From(),
				"to", transfer.To, "symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
			return nil, err
		}
		receipt, err := accDB.Transfer(fromAddr, transfer.GetTo(), transfer.GetAmount())
		if err != nil {
			elog.Error("Exec_Transfer cross transfer", "txHash", txHash, "from", tx.From(),
				"to", transfer.To, "symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
			return nil, err
		}
		return receipt, nil
	}

	// from是btc utxo, 记录并等待confirm交易
	if transfer.GetFromUtxo() != "" {
		receipt := &types.Receipt{
			Ty: types.ExecOk,
			KV: []*types.KeyValue{{Key: formatPayloadKey(tx.Hash()), Value: types.Encode(transfer)}},
		}
		// check tx 阶段已经校验过地址
		utxo, _ := rtypes.NewOutPointFromString(transfer.GetFromUtxo())
		utxo.PkScript = transfer.GetFromUtxoPkScript()
		receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
			Ty: rtypes.TyPendingTxLog,
			Log: types.Encode(&rtypes.PendingTx{
				ActionType:    rtypes.TyTransferAction,
				Timestamp:     r.GetBlockTime(),
				TxBlockHeight: r.GetHeight(),
				TxIndex:       int64(index),
				TxHash:        tx.Hash(),
				Utxo:          utxo,
			}),
		})
		return receipt, nil
	}

	asset := &rtypes.RgbxAsset{}
	err := readDB(r.GetStateDB(), formatAssetKey(transfer.GetSymbol()), asset)
	if err != nil {
		elog.Error("Exec_Transfer get asset", "txHash", txHash, "symbol", transfer.GetSymbol(),
			"err", err)
		return nil, ErrAssetNotExist
	}

	if asset.Type == uint32(rtypes.Collectible) {
		return r.assetReceipt(asset, transfer.GetTo()), nil
	}

	accDB, err := r.newAccount(transfer.GetSymbol())
	if err != nil {
		elog.Error("Exec_Transfer newAccount", "txHash", txHash, "from", fromAddr,
			"to", transfer.To, "symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	receipt, err := accDB.Transfer(fromAddr, transfer.GetTo(), transfer.GetAmount())
	if err != nil {
		elog.Error("Exec_Transfer transfer", "txHash", txHash, "from", fromAddr,
			"to", transfer.To, "symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	return receipt, nil
}

func (r *rgbx) assetReceipt(asset *rtypes.RgbxAsset, owner string) *types.Receipt {

	if asset.Type == uint32(rtypes.Collectible) {
		asset.Owner = owner
	}
	assetVal := types.Encode(asset)
	receipt := &types.Receipt{
		Ty:   types.ExecOk,
		KV:   []*types.KeyValue{{Key: formatAssetKey(asset.Symbol), Value: assetVal}},
		Logs: []*types.ReceiptLog{{Ty: rtypes.TyAssetLog, Log: assetVal}},
	}
	return receipt
}

func (r *rgbx) Exec_Confirm(confirm *rtypes.ConfirmTx, tx *types.Transaction, index int) (*types.Receipt, error) {

	txHash := hex.EncodeToString(tx.Hash())
	confirmHash := hex.EncodeToString(confirm.GetTxHash())
	action := rtypes.GetActionName(confirm.GetActionType())
	elog.Debug("Exec_Confirm", "opRetOutIdx", confirm.GetUtxoProof().GetOpRetOutputIdx(),
		"timeout", confirm.GetTimeout(), "txHash", txHash, "confirmHash", confirmHash,
		"action", action)
	if confirm.GetTimeout() {
		return &types.Receipt{Ty: types.ExecOk}, nil
	}
	if confirm.ActionType == rtypes.TyWithDrawAsset {
		return r.confirmWithdrawSettlement(confirm, txHash, confirmHash)
	}

	btcTxData := confirm.GetBtcTxProof().GetTxData()
	// 绑定资产的utxo已经在btc链上花费，但op return不存在或承诺数据不正确，
	// 交易仅做标记并返回，相关资产永久冻结，无法转移
	var btcTx wire.MsgTx
	opRetIdx := confirm.GetUtxoProof().GetOpRetOutputIdx()
	if opRetIdx >= 0 && len(btcTxData) > 0 {
		if err := btcTx.DeserializeNoWitness(bytes.NewReader(btcTxData)); err != nil {
			elog.Error("Exec_Confirm deserialize btc tx", "action", action,
				"txHash", txHash, "confirmHash", confirmHash, "err", err)
			return nil, err
		}
	}
	// spendHash 取 merkle 证明所绑定交易的 txid（解析后重序列化的 no-witness 哈希），
	// 不用原始字节直接哈希——原始字节可拼接尾随字节改变 DoubleHashH 结果，
	// 导致 spendHash 与 merkle 验证过的 txid 不一致，污染 GenesisBtcTxHash 与资产 owner
	spendHash := btcTx.TxHash().String()

	commitment, _ := txscript.NullDataScript(confirm.GetTxHash())
	if opRetIdx < 0 || int(opRetIdx) >= len(btcTx.TxOut) ||
		!bytes.Equal(commitment, btcTx.TxOut[opRetIdx].PkScript) {

		elog.Warn("Exec_Confirm op return commitment", "action", action,
			"txHash", txHash, "confirmHash", confirmHash, "opRetIdx", opRetIdx,
			"spendHash", spendHash, "txOutLen", len(btcTx.TxOut),
			"expectCommit", hex.EncodeToString(commitment))
		return &types.Receipt{Ty: types.ExecOk}, nil
	}

	if confirm.ActionType == rtypes.TyMintAction {
		return r.mintAsset(confirm, txHash, confirmHash, spendHash)
	}
	return r.transferAsset(confirm, txHash, confirmHash, spendHash)
}

func (r *rgbx) confirmWithdrawSettlement(confirm *rtypes.ConfirmTx, txHash, confirmHash string) (*types.Receipt, error) {
	withdraw := &rtypes.WithdrawAsset{}
	if err := readDB(r.GetStateDB(), formatPayloadKey(confirm.GetTxHash()), withdraw); err != nil {
		elog.Error("confirmWithdrawSettlement read payload", "txHash", txHash, "confirmHash", confirmHash, "err", err)
		return nil, err
	}
	symbol := ensureCrossChainSymbol(withdraw.GetAssetSymbol())
	accDB, err := r.newAccount(symbol)
	if err != nil {
		return nil, err
	}
	lockAddr := r.crossChainLockAddress(accDB)
	receipt, err := accDB.Burn(lockAddr, withdraw.GetAmount())
	if err != nil {
		elog.Error("confirmWithdrawSettlement burn lock", "txHash", txHash, "confirmHash", confirmHash,
			"lockAddr", lockAddr, "symbol", withdraw.GetAssetSymbol(), "amount", withdraw.GetAmount(), "err", err)
		return nil, err
	}
	return receipt, nil
}

func (r *rgbx) mintAsset(confirm *rtypes.ConfirmTx, txHash, confirmHash, spendHash string) (*types.Receipt, error) {

	mint := &rtypes.MintAsset{}
	err := readDB(r.GetStateDB(), formatPayloadKey(confirm.GetTxHash()), mint)

	if err != nil {
		elog.Error("mintAsset readDB", "txHash", txHash,
			"confirmHash", confirmHash, "err", err)
		return nil, err
	}
	assetTy := rtypes.AssetType(mint.GetType())
	log.Debug("mintAsset", "symbol", mint.Symbol, "amount", mint.TotalAmount,
		"txHash", txHash, "confirmHash", confirmHash, "assetTy", assetTy.String(),
		"spendingTxHash", spendHash, "metaHash", mint.MetaHash)
	asset := &rtypes.RgbxAsset{
		Symbol:           formatSymbol(mint.Symbol),
		Type:             mint.Type,
		TotalAmount:      mint.TotalAmount,
		MetaHash:         mint.MetaHash,
		GenesisBtcTxHash: spendHash,
		Precision:        mint.Precision,
	}
	// 默认opReturn的下一个utxo作为资产所有者， 如果不存在，资产将被永久冻结，无法转移
	owner := rtypes.FormatUtxo(spendHash, uint32(confirm.GetUtxoProof().GetOpRetOutputIdx()+1))
	receipt := r.assetReceipt(asset, owner)
	if assetTy == rtypes.Collectible {
		return receipt, nil
	}

	// Normal asset
	accDB, err := r.newAccount(mint.GetSymbol())
	if err != nil {
		elog.Error("Exec_Transfer newAccount", "txHash", txHash,
			"from", confirmHash, "err", err)
		return nil, err
	}
	mintReceipt, err := accDB.Mint(owner, mint.GetTotalAmount())
	if err != nil {
		elog.Error("mintAsset mint", "txHash", txHash,
			"confirmHash", confirmHash, "symbol", mint.Symbol, "owner", owner,
			"amount", mint.TotalAmount, "err", err)
		return nil, err
	}

	receipt.KV = append(receipt.KV, mintReceipt.KV...)
	receipt.Logs = append(receipt.Logs, mintReceipt.Logs...)
	return receipt, nil

}

func (r *rgbx) transferAsset(confirm *rtypes.ConfirmTx, txHash, confirmHash, spendHash string) (*types.Receipt, error) {

	transfer := &rtypes.TransferAsset{}
	err := readDB(r.GetStateDB(), formatPayloadKey(confirm.GetTxHash()), transfer)
	if err != nil {
		elog.Error("transferAsset readDB", "txHash", txHash,
			"confirmHash", confirmHash, "err", err)
		return nil, err
	}

	if isCrossChainSymbol(transfer.GetSymbol()) {
		return nil, types.ErrNotSupport
	}

	changeAddress := transfer.GetChangeAddr()
	// 未指定找零地址时， 则使用opReturn的下一个utxo， 如果不存在，资产将被永久冻结，无法转移
	if changeAddress == "" {
		changeAddress = rtypes.FormatUtxo(spendHash, uint32(confirm.GetUtxoProof().GetOpRetOutputIdx()+1))
	}

	log.Debug("transferAsset", "symbol", transfer.Symbol, "amount", transfer.Amount,
		"txHash", txHash, "confirmHash", confirmHash, "spendHash", spendHash,
		"from", transfer.FromUtxo, "to", transfer.To, "change", changeAddress)

	asset := &rtypes.RgbxAsset{}
	err = readDB(r.GetStateDB(), formatAssetKey(transfer.GetSymbol()), asset)
	if err != nil {
		elog.Error("transferAsset get asset", "txHash", txHash, "confirmHash", confirmHash,
			"symbol", transfer.GetSymbol(), "err", err)
		return nil, ErrAssetNotExist
	}
	if asset.Type == uint32(rtypes.Collectible) {
		return r.assetReceipt(asset, transfer.GetTo()), nil
	}
	accDB, err := r.newAccount(transfer.GetSymbol())
	if err != nil {
		elog.Error("transferAsset newAccount", "txHash", txHash,
			"confirmHash", confirmHash, "from", transfer.FromUtxo, "to", transfer.To,
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}

	changeAmount := accDB.LoadAccount(transfer.GetFromUtxo()).GetBalance() - transfer.GetAmount()
	receipt, err := accDB.Transfer(transfer.GetFromUtxo(), transfer.GetTo(), transfer.GetAmount())
	if err != nil {
		elog.Error("transferAsset transfer", "txHash", txHash, "confirmHash", confirmHash,
			"from", transfer.FromUtxo, "to", transfer.To,
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}

	// handle change
	if changeAmount <= 0 {
		elog.Debug("transferAsset transfer zero change", "txHash", txHash, "confirmHash", confirmHash)
		return receipt, nil
	}

	changeReceipt, err := accDB.Transfer(transfer.GetFromUtxo(), changeAddress, changeAmount)
	if err != nil {
		elog.Error("transferAsset change", "txHash", txHash, "confirmHash", confirmHash,
			"from", transfer.FromUtxo, "changeAddr", changeAddress,
			"symbol", transfer.Symbol, "amount", changeAmount, "err", err)
		return nil, err
	}

	receipt.KV = append(receipt.KV, changeReceipt.KV...)
	receipt.Logs = append(receipt.Logs, changeReceipt.Logs...)
	return receipt, nil
}

package executor

import (
	"encoding/hex"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (r *rgbx) Exec_Mint(mint *rtypes.MintAsset, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatPayloadKey(tx.Hash()),
		Value: types.Encode(mint),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: rtypes.TyPendingTxLog,
		Log: types.Encode(&rtypes.PendingTx{
			ActionType: rtypes.TyMintAction,
			Timestamp:  r.GetBlockTime(),
			TxHash:     tx.Hash(),
			Utxo:       mint.GetGenesisOut(),
		}),
	})

	return receipt, nil
}

func (r *rgbx) Exec_Transfer(transfer *rtypes.TransferAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	txHash := hex.EncodeToString(tx.Hash())
	elog.Debug("Exec_Transfer", "txHash", txHash, "symbol", transfer.Symbol, "amount", transfer.Amount,
		"from", transfer.GetFrom().Address(), "to", transfer.GetTo().Address())

	// from是btc utxo, 记录并等待confirm交易
	if transfer.GetFrom().GetUtxo() != nil {
		receipt := &types.Receipt{
			Ty: types.ExecOk,
			KV: []*types.KeyValue{{Key: formatPayloadKey(tx.Hash()), Value: types.Encode(transfer)}},
		}
		receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
			Ty: rtypes.TyPendingTxLog,
			Log: types.Encode(&rtypes.PendingTx{
				ActionType: rtypes.TyTransferAction,
				Timestamp:  r.GetBlockTime(),
				TxHash:     tx.Hash(),
				Utxo:       transfer.GetFrom().GetUtxo(),
			}),
		})
		return receipt, nil
	}
	accDB, err := r.newAccount(transfer.GetSymbol())
	if err != nil {
		elog.Error("Exec_Transfer newAccount", "txHash", txHash,
			"from", transfer.From.Address(), "to", transfer.To.Address(),
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	receipt, err := accDB.Transfer(transfer.GetFrom().Address(), transfer.GetTo().Address(), transfer.GetAmount())
	if err != nil {
		elog.Error("Exec_Transfer transfer", "txHash", txHash,
			"from", transfer.From.Address(), "to", transfer.To.Address(),
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	return receipt, nil
}

func (r *rgbx) Exec_Confirm(confirm *rtypes.ConfirmTx, tx *types.Transaction, index int) (*types.Receipt, error) {

	txHash := hex.EncodeToString(tx.Hash())
	confirmHash := hex.EncodeToString(confirm.GetTxHash())
	if confirm.GetTimeout() {
		elog.Debug("Exec_Confirm timeout", "txHash", txHash,
			"confirmTx", confirmHash, "action", rtypes.GetActionName(confirm.GetActionType()))
		return &types.Receipt{Ty: types.ExecOk}, nil
	}

	if confirm.ActionType == rtypes.TyMintAction {
		return r.mintAsset(confirm, txHash, confirmHash)
	}
	return r.transferAsset(confirm, txHash, confirmHash)
}

func (r *rgbx) mintAsset(confirm *rtypes.ConfirmTx, txHash, confirmHash string) (*types.Receipt, error) {

	mint := &rtypes.MintAsset{}

	err := readDB(r.GetStateDB(), formatPayloadKey(confirm.GetTxHash()), mint)

	if err != nil {
		elog.Error("mintAsset readDB", "txHash", txHash,
			"confirmTX", confirmHash, "err", err)
		return nil, err
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	asset := &rtypes.Asset{
		Symbol:      mint.Symbol,
		Type:        mint.Type,
		TotalAmount: mint.TotalAmount,
		MetaHash:    mint.MetaHash,
		GenesisTx:   confirm.Witness.GetTxHash(),
	}
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatAssetKey(mint.Symbol),
		Value: types.Encode(asset),
	})

	ownerUtxo := &rtypes.OutPoint{
		Hash:  confirm.Witness.GetTxHash(),
		Index: 2,
	}

	accDB, err := r.newAccount(mint.GetSymbol())
	if err != nil {
		elog.Error("Exec_Transfer newAccount", "txHash", txHash,
			"from", confirmHash, "err", err)
		return nil, err
	}
	mintReceipt, err := accDB.Mint(ownerUtxo.ToString(), mint.GetTotalAmount())
	if err != nil {
		elog.Error("mintAsset mint", "txHash", txHash,
			"confirmTx", confirmHash, "symbol", mint.Symbol, "owner", ownerUtxo.ToString(),
			"amount", mint.TotalAmount, "err", err)
		return nil, err
	}

	receipt.KV = append(receipt.KV, mintReceipt.KV...)
	receipt.Logs = append(receipt.Logs, mintReceipt.Logs...)
	return receipt, nil

}

func (r *rgbx) transferAsset(confirm *rtypes.ConfirmTx, txHash, confirmHash string) (*types.Receipt, error) {

	transfer := &rtypes.TransferAsset{}

	err := readDB(r.GetStateDB(), formatPayloadKey(confirm.GetTxHash()), transfer)

	if err != nil {
		elog.Error("transferAsset readDB", "txHash", txHash,
			"confirmTX", confirmHash, "err", err)
		return nil, err
	}

	accDB, err := r.newAccount(transfer.GetSymbol())
	if err != nil {
		elog.Error("transferAsset newAccount", "txHash", txHash, "confirmTx", confirmHash,
			"from", transfer.From.Address(), "to", transfer.To.Address(),
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	receipt, err := accDB.Transfer(transfer.GetFrom().Address(), transfer.GetTo().Address(), transfer.GetAmount())
	if err != nil {
		elog.Error("transferAsset transfer", "txHash", txHash, "confirmTx", confirmHash,
			"from", transfer.From.Address(), "to", transfer.To.Address(),
			"symbol", transfer.Symbol, "amount", transfer.Amount, "err", err)
		return nil, err
	}
	return receipt, nil

}

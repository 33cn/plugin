package executor

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/33cn/chain33/types"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

const defaultGuardianParachainTitle = "user.p.rgbxguardians."

func (r *rgbx) Exec_CommitDKG(commit *rtypes.CommitDKG, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := hex.EncodeToString(tx.Hash())
	commitAddr := tx.From()
	addrs := &types.ReqAddrs{}
	err := readDB(r.GetStateDB(), formatDkgConfirmationsKey(commit.DkgAddress), addrs)
	if err != nil && !errors.Is(err, types.ErrNotFound) {
		elog.Error("Exec_CommitDKG", "txHash", txHash, "symbol", commit.AssetSymbol, "dkgAddr", commit.DkgAddress, "readDB err", err)
		return nil, err
	}
	for _, addr := range addrs.Addrs {
		if commitAddr == addr {
			return receipt, nil
		}
	}
	addrs.Addrs = append(addrs.Addrs, commitAddr)
	guardianAddrs, err := r.getGuardianNodeAddress(rgbxCfg.GuardianParachainTitle)
	if err != nil {
		elog.Error("Exec_CommitDKG", "txHash", txHash, "getGuardianNodeAddress err", err)
		return nil, err
	}

	encodeAddrs := types.Encode(addrs)
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatDkgConfirmationsKey(commit.GetDkgAddress()),
		Value: encodeAddrs,
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  rtypes.TyCommitDKGLog,
		Log: encodeAddrs,
	})

	if len(strings.Split(guardianAddrs, ",")) == len(addrs.Addrs) {

		info := &rtypes.CrossChainInfo{
			AssetSymbol: formatSymbol(commit.AssetSymbol),
			TssAddress:  commit.DkgAddress,
			PkScript:    commit.PkScript,
		}
		receipt.KV = append(receipt.KV, &types.KeyValue{
			Key:   formatCrossChainInfoKey(commit.GetAssetSymbol()),
			Value: types.Encode(info),
		})
	}

	return receipt, nil

}

func (r *rgbx) Exec_DepositAsset(deposit *rtypes.DepositAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := tx.Hash()
	accDB, err := r.newCrossChainAccount(deposit.GetAssetSymbol())
	if err != nil {
		elog.Error("Exec_DepositAsset newCrossChainAccount", "txHash", hex.EncodeToString(txHash), "symbol", deposit.GetAssetSymbol(),
			"err", err)
		return nil, err
	}
	depositReceipt, err := accDB.Mint(deposit.GetDepositAddress(), deposit.GetAmount())
	if err != nil {
		elog.Error("Exec_DepositAsset Mint", "txHash", hex.EncodeToString(txHash), "symbol", deposit.GetAssetSymbol(),
			"depositAddr", deposit.GetDepositAddress(), "amount", deposit.GetAmount(), "err", err)
		return nil, err
	}
	receipt.KV = append(receipt.KV, depositReceipt.KV...)
	receipt.Logs = append(receipt.Logs, depositReceipt.Logs...)

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatPayloadKey(txHash),
		Value: types.Encode(deposit),
	}, &types.KeyValue{
		Key:   formatDepositUsedKey(deposit.GetTxProof().GetTxData()),
		Value: []byte("used"),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: rtypes.TyPendingTxLog,
		Log: types.Encode(&rtypes.PendingTx{
			ActionType:    rtypes.TyDepositAsset,
			Timestamp:     r.GetBlockTime(),
			TxBlockHeight: r.GetHeight(),
			TxIndex:       int64(index),
			TxHash:        tx.Hash(),
			Amount:        deposit.GetAmount(),
			AssetSymbol:   formatSymbol(deposit.GetAssetSymbol()),
			TargetAddress: deposit.GetDepositAddress(),
		}),
	})

	return receipt, nil

}

func (r *rgbx) Exec_WithdrawAsset(withdraw *rtypes.WithdrawAsset, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := tx.Hash()
	accDB, err := r.newCrossChainAccount(withdraw.GetAssetSymbol())
	if err != nil {
		return nil, err
	}
	lockAddr := r.crossChainLockAddress(accDB)
	lockReceipt, err := accDB.Transfer(tx.From(), lockAddr, withdraw.GetAmount())
	if err != nil {
		elog.Error("Exec_WithdrawAsset lock transfer", "txHash", hex.EncodeToString(txHash), "from", tx.From(),
			"symbol", withdraw.GetAssetSymbol(), "amount", withdraw.GetAmount(), "err", err)
		return nil, err
	}
	receipt.KV = append(receipt.KV, lockReceipt.KV...)
	receipt.Logs = append(receipt.Logs, lockReceipt.Logs...)

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatPayloadKey(txHash),
		Value: types.Encode(withdraw),
	})
	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: rtypes.TyPendingTxLog,
		Log: types.Encode(&rtypes.PendingTx{
			ActionType:    rtypes.TyWithDrawAsset,
			Timestamp:     r.GetBlockTime(),
			TxBlockHeight: r.GetHeight(),
			TxIndex:       int64(index),
			TxHash:        txHash,
			FromAddress:   tx.From(),
			AssetSymbol:   formatSymbol(withdraw.GetAssetSymbol()),
			TargetAddress: withdraw.GetDestinationAddr(),
			Amount:        withdraw.GetAmount(),
			FeeRate:       withdraw.GetFeeRate(),
		}),
	})
	return receipt, nil
}

func (r *rgbx) getGuardianNodeAddress(title string) (string, error) {

	params := &paratypes.ReqParacrossNodeInfo{Title: title}
	resp, err := r.GetAPI().Query(paratypes.ParaX, "GetNodeGroupStatus", params)
	if err != nil {
		elog.Error("getGuardianNodeAddress", "title", title, "err", err)
		return "", err
	}

	status := resp.(*paratypes.ParaNodeGroupStatus)
	return status.TargetAddrs, nil
}

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

	symbol := formatSymbol(commit.GetAssetSymbol())
	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := hex.EncodeToString(tx.Hash())
	commitAddr := tx.From()
	addrs := &types.ReqAddrs{}
	err := readDB(r.GetStateDB(), formatDkgConfirmationsKey(commit.DkgAddress), addrs)
	if err != nil && !errors.Is(err, types.ErrNotFound) {
		elog.Error("Exec_CommitDKG", "txHash", txHash, "symbol", symbol, "dkgAddr", commit.DkgAddress, "readDB err", err)
		return nil, ErrGetDkgConfirmations
	}
	for _, addr := range addrs.Addrs {
		if commitAddr == addr {
			return receipt, nil
		}
	}
	addrs.Addrs = append(addrs.Addrs, commitAddr)
	guardianAddrs, err := r.getGuardianNodeAddress(rgbxCfg.GuardianParachainTitle)
	if err != nil {
		elog.Error("Exec_CommitDKG", "txHash", txHash, "symbol", symbol, "getGuardianNodeAddress err", err)
		return nil, ErrGetGuardianNodeAddress
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
			AssetSymbol:   symbol,
			WrappedSymbol: formatCrossChainSymbol(symbol),
			TssAddress:    commit.DkgAddress,
			PkScript:      commit.PkScript,
		}
		receipt.KV = append(receipt.KV, &types.KeyValue{
			Key:   formatCrossChainInfoKey(symbol),
			Value: types.Encode(info),
		})
	}

	return receipt, nil

}

func (r *rgbx) Exec_Deposit(deposit *rtypes.DepositAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := tx.Hash()
	symbol := ensureCrossChainSymbol(deposit.GetAssetSymbol())
	accDB, err := r.newAccount(symbol)
	if err != nil {
		elog.Error("Exec_Deposit newCrossChainAccount", "txHash", hex.EncodeToString(txHash), "symbol", symbol,
			"err", err)
		return nil, err
	}
	depositReceipt, err := accDB.Mint(deposit.GetDepositAddress(), deposit.GetAmount())
	if err != nil {
		elog.Error("Exec_Deposit Mint", "txHash", hex.EncodeToString(txHash), "symbol", symbol,
			"depositAddr", deposit.GetDepositAddress(), "amount", deposit.GetAmount(), "err", err)
		return nil, err
	}
	receipt.KV = append(receipt.KV, depositReceipt.KV...)
	receipt.Logs = append(receipt.Logs, depositReceipt.Logs...)

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatDepositUsedKey(deposit.GetTxProof().GetTxData()),
		Value: []byte("used"),
	})

	return receipt, nil

}

func (r *rgbx) Exec_Withdraw(withdraw *rtypes.WithdrawAsset, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := tx.Hash()
	symbol := ensureCrossChainSymbol(withdraw.GetAssetSymbol())
	accDB, err := r.newAccount(symbol)
	if err != nil {
		return nil, err
	}
	lockAddr := r.crossChainLockAddress(accDB)
	lockReceipt, err := accDB.Transfer(tx.From(), lockAddr, withdraw.GetAmount())
	if err != nil {
		elog.Error("Exec_Withdraw lock transfer", "txHash", hex.EncodeToString(txHash), "from", tx.From(),
			"symbol", symbol, "amount", withdraw.GetAmount(), "err", err)
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

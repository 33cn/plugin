package executor

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/33cn/chain33/types"
	paratypes "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
)

const guardianParachainTitle = "user.p.rgbxguardians."

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
			break
		}
	}
	addrs.Addrs = append(addrs.Addrs, commitAddr)
	guardianAddrs, err := r.getGuardianNodeAddress(guardianParachainTitle)
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
			Symbol:         commit.AssetSymbol,
			DepositAddress: commit.DkgAddress,
		}
		receipt.KV = append(receipt.KV, &types.KeyValue{
			Key:   formatCrossChainDepositAddressKey(commit.GetAssetSymbol()),
			Value: types.Encode(info),
		})
	}

	return receipt, nil

}

func (r *rgbx) Exec_DepositAsset(deposit *rtypes.DepositAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
	txHash := tx.Hash()

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatPayloadKey(txHash),
		Value: types.Encode(deposit),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty: rtypes.TyPendingTxLog,
		Log: types.Encode(&rtypes.PendingTx{
			ActionType:    rtypes.TyDepositAsset,
			Timestamp:     r.GetBlockTime(),
			TxBlockHeight: r.GetHeight(),
			TxIndex:       int64(index),
			TxHash:        tx.Hash(),
			TargetAddress: deposit.SenderAddr,
		}),
	})

	return receipt, nil

}

func (r *rgbx) Exec_WithdrawAsset(confirm *rtypes.WithdrawAsset, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}
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

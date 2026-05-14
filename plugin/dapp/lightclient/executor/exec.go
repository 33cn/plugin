package executor

import (
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (l *lightclient) Exec_BtcHeaders(headers *ltypes.BtcHeaders, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}

	prevHeader, err := getBtcLastHeader(l.GetStateDB())
	if err != nil && err != types.ErrNotFound {
		elog.Error("Exec_BtcHeaders", "getBtcLastHeader err", err)
		return nil, ErrBtcGetLastHeader
	}

	commitHeader := headers.GetHeaders()[len(headers.GetHeaders())-1]
	elog.Debug("Exec_BtcHeaders", "lastHeight", prevHeader.GetHeight(), "lastHash", prevHeader.GetHash(),
		"commitHeight", commitHeader.GetHeight(), "commitHash", commitHeader.GetHash())
	log := &ltypes.BtcHeadersLog{}
	log.LastHeight = prevHeader.GetHeight()
	log.LastHash = prevHeader.GetHash()

	log.CommitHash = commitHeader.GetHash()
	log.CommitHeight = commitHeader.GetHeight()
	log.Confirmations = commitHeader.GetConfirmations()

	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   btcLastHeaderKey(),
		Value: types.Encode(commitHeader),
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  ltypes.TyBtcHeadersLog,
		Log: types.Encode(log),
	})

	return receipt, nil
}

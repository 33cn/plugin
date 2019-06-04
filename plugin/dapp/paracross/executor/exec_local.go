// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"

	"encoding/hex"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//ExecLocal_Commit commit tx local db process
func (e *Paracross) ExecLocal_Commit(payload *pt.ParacrossCommitAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParacrossCommit {
			var g pt.ReceiptParacrossCommit
			types.Decode(log.Log, &g)

			var r pt.ParacrossTx
			r.TxHash = common.ToHex(tx.Hash())
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: types.Encode(&r)})
		} else if log.Ty == pt.TyLogParacrossCommitDone {
			var g pt.ReceiptParacrossDone
			types.Decode(log.Log, &g)

			key := calcLocalTitleKey(g.Title)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})

			key = calcLocalHeightKey(g.Title, g.Height)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})
			if !types.IsPara() {
				r, err := e.saveLocalParaTxs(tx, false)
				if err != nil {
					return nil, err
				}
				set.KV = append(set.KV, r.KV...)
			}

		} else if log.Ty == pt.TyLogParacrossCommitRecord {
			var g pt.ReceiptParacrossRecord
			types.Decode(log.Log, &g)

			var r pt.ParacrossTx
			r.TxHash = common.ToHex(tx.Hash())
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: types.Encode(&r)})
		}
	}
	return &set, nil
}

//ExecLocal_NodeConfig node config add process
func (e *Paracross) ExecLocal_NodeConfig(payload *pt.ParaNodeAddrConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParaNodeConfig {
			var g pt.ReceiptParaNodeConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalNodeTitleStatus(g.Current.Title, g.Prev.Status, g.Current.Id), Value: nil})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key:   calcLocalNodeTitleStatus(g.Current.Title, g.Current.Status, g.Current.Id),
				Value: types.Encode(g.Current)})
		} else if log.Ty == pt.TyLogParaNodeVoteDone {
			var g pt.ReceiptParaNodeVoteDone
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			key := calcLocalNodeTitleDone(g.Title, g.TargetAddr)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})
		}
	}
	return &set, nil
}

//ExecLocal_NodeGroupConfig node group config add process
func (e *Paracross) ExecLocal_NodeGroupConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParaNodeGroupConfig {
			var g pt.ReceiptParaNodeGroupConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalNodeGroupStatusTitle(g.Prev.Status, g.Current.Title, g.Current.Id), Value: nil})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalNodeGroupStatusTitle(g.Current.Status, g.Current.Title, g.Current.Id), Value: types.Encode(g.Current)})
		}
		if log.Ty == pt.TyLogParaNodeConfig {
			var g pt.ReceiptParaNodeConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalNodeTitleStatus(g.Current.Title, g.Prev.Status, g.Current.Id), Value: nil})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key:   calcLocalNodeTitleStatus(g.Current.Title, g.Current.Status, g.Current.Id),
				Value: types.Encode(g.Current)})
		}
	}
	return &set, nil
}

//ExecLocal_AssetTransfer asset transfer local proc
func (e *Paracross) ExecLocal_AssetTransfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet

	//  主链转出记录，
	//  转入在 commit done 时记录， 因为没有日志里没有当时tx信息
	r, err := e.initLocalAssetTransfer(tx, true, false)
	if err != nil {
		return nil, err
	}
	set.KV = append(set.KV, r)

	return &set, nil
}

//ExecLocal_AssetWithdraw asset withdraw process
func (e *Paracross) ExecLocal_AssetWithdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

func setMinerTxResult(payload *pt.ParacrossMinerAction, txs []*types.Transaction, receipts []*types.ReceiptData) error {
	isCommitTx := make(map[string]bool)
	var curTxHashs, paraTxHashs, crossTxHashs [][]byte
	for _, tx := range txs {
		hash := tx.Hash()
		curTxHashs = append(curTxHashs, hash)
		//对user.p.xx.paracross ,actionTy==commit 的tx不需要再发回主链
		if types.IsMyParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			var payload pt.ParacrossAction
			err := types.Decode(tx.Payload, &payload)
			if err != nil {
				clog.Error("setMinerTxResult", "txHash", common.ToHex(hash))
				return err
			}
			if payload.Ty == pt.ParacrossActionCommit {
				isCommitTx[string(hash)] = true
			}
		}
		//跨链交易包含了主链交易，需要过滤出来
		if types.IsMyParaExecName(string(tx.Execer)) && !isCommitTx[string(hash)] {
			paraTxHashs = append(paraTxHashs, hash)
		}
	}
	totalCrossTxHashs := FilterParaMainCrossTxHashes(types.GetTitle(), txs)
	for _, crossHash := range totalCrossTxHashs {
		if !isCommitTx[string(crossHash)] {
			crossTxHashs = append(crossTxHashs, crossHash)
		}
	}
	payload.Status.TxHashs = paraTxHashs
	payload.Status.TxResult = util.CalcBitMap(paraTxHashs, curTxHashs, receipts)
	payload.Status.CrossTxHashs = crossTxHashs
	payload.Status.CrossTxResult = util.CalcBitMap(crossTxHashs, curTxHashs, receipts)

	return nil
}

func setMinerTxResultFork(status *pt.ParacrossNodeStatus, txs []*types.Transaction, receipts []*types.ReceiptData) error {
	isCommitTx := make(map[string]bool)
	var curTxHashs [][]byte
	for _, tx := range txs {
		hash := tx.Hash()
		curTxHashs = append(curTxHashs, hash)

		if types.IsMyParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
			var payload pt.ParacrossAction
			err := types.Decode(tx.Payload, &payload)
			if err != nil {
				clog.Error("setMinerTxResultFork", "txHash", common.ToHex(hash))
				return err
			}
			if payload.Ty == pt.ParacrossActionCommit {
				isCommitTx[string(hash)] = true
			}
		}
	}

	status.TxCounts = uint32(len(curTxHashs))
	//有tx且全部是user.p.x.paracross的commit tx时候设为0
	status.NonCommitTxCounts = 1
	if len(curTxHashs) != 0 && len(curTxHashs) == len(isCommitTx) {
		status.NonCommitTxCounts = 0
	}
	crossTxHashs := FilterParaCrossTxHashes(types.GetTitle(), txs)

	//主链自己过滤平行链tx， 对平行链执行失败的tx主链无法识别，主链和平行链需要获取相同的最初的tx map
	//全部平行链tx结果
	status.TxResult = []byte(hex.EncodeToString(util.CalcBitMap(curTxHashs, curTxHashs, receipts)))
	//跨链tx结果
	status.CrossTxResult = []byte(hex.EncodeToString(util.CalcBitMap(crossTxHashs, curTxHashs, receipts)))

	status.TxHashs = [][]byte{CalcTxHashsHash(curTxHashs)}
	status.CrossTxHashs = [][]byte{CalcTxHashsHash(crossTxHashs)}

	return nil
}

//ExecLocal_Miner miner tx local db process
func (e *Paracross) ExecLocal_Miner(payload *pt.ParacrossMinerAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if index != 0 {
		return nil, pt.ErrParaMinerBaseIndex
	}

	var set types.LocalDBSet
	txs := e.GetTxs()

	forkHeight := getDappForkHeight(pt.ForkCommitTx)

	//removed the 0 vote tx
	if payload.Status.MainBlockHeight >= forkHeight {
		err := setMinerTxResultFork(payload.Status, txs[1:], e.GetReceipt()[1:])
		if err != nil {
			return nil, err
		}
	} else {
		err := setMinerTxResult(payload, txs[1:], e.GetReceipt()[1:])
		if err != nil {
			return nil, err
		}
	}

	set.KV = append(set.KV, &types.KeyValue{
		Key:   pt.CalcMinerHeightKey(payload.Status.Title, payload.Status.Height),
		Value: types.Encode(payload.Status)})

	return &set, nil
}

// ExecLocal_Transfer asset transfer local db process
func (e *Paracross) ExecLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_Withdraw asset withdraw local db process
func (e *Paracross) ExecLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecLocal_TransferToExec transfer asset to exec local db process
func (e *Paracross) ExecLocal_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//ExecDelLocal_Commit consensus commit tx del local db process
func (e *Paracross) ExecDelLocal_Commit(payload *pt.ParacrossCommitAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	cfg := e.GetAPI().GetConfig()
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParacrossCommit { //} || log.Ty == types.TyLogParacrossCommitRecord {
			var g pt.ReceiptParacrossCommit
			types.Decode(log.Log, &g)

			var r pt.ParacrossTx
			r.TxHash = common.ToHex(tx.Hash())
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: nil})
		} else if log.Ty == pt.TyLogParacrossCommitDone {
			var g pt.ReceiptParacrossDone
			types.Decode(log.Log, &g)
			g.Height = g.Height - 1

			key := calcLocalTitleKey(g.Title)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})

			key = calcLocalHeightKey(g.Title, g.Height)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: nil})

			if !cfg.IsPara() && g.Height > 0 {
				r, err := e.saveLocalParaTxs(tx, true)
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
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: nil})
		}
	}
	return &set, nil
}

// ExecDelLocal_NodeConfig node config tx delete process
func (e *Paracross) ExecDelLocal_NodeConfig(payload *pt.ParaNodeAddrConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	cfg := e.GetAPI().GetConfig()
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParaNodeConfig {
			var g pt.ReceiptParaNodeConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalNodeTitleStatus(g.Current.Title, g.Prev.Status, g.Prev.Id), Value: types.Encode(g.Prev)})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalNodeTitleStatus(g.Current.Title, g.Current.Status, g.Current.Id), Value: nil})
		} else if log.Ty == pt.TyLogParaNodeVoteDone {
			var g pt.ReceiptParaNodeVoteDone
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			key := calcLocalNodeTitleDone(g.Title, g.TargetAddr)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: nil})
		} else if log.Ty == pt.TyLogParacrossCommitDone {
			var g pt.ReceiptParacrossDone
			types.Decode(log.Log, &g)
			g.Height = g.Height - 1

			key := calcLocalTitleKey(g.Title)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})

			key = calcLocalHeightKey(g.Title, g.Height)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: nil})

			if !cfg.IsPara() && g.Height > 0 {
				r, err := e.saveLocalParaTxsFork(&g, true)
				if err != nil {
					return nil, err
				}
				set.KV = append(set.KV, r.KV...)
			}
		}
	}
	return &set, nil
}

// ExecDelLocal_NodeGroupConfig node group config tx delete process
func (e *Paracross) ExecDelLocal_NodeGroupConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
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
					Key: calcLocalNodeGroupStatusTitle(g.Prev.Status, g.Current.Title, g.Current.Id), Value: types.Encode(g.Prev)})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalNodeGroupStatusTitle(g.Current.Status, g.Current.Title, g.Current.Id), Value: nil})
		} else if log.Ty == pt.TyLogParaNodeConfig {
			var g pt.ReceiptParaNodeConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalNodeTitleStatus(g.Current.Title, g.Prev.Status, g.Prev.Id), Value: types.Encode(g.Prev)})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalNodeTitleStatus(g.Current.Title, g.Current.Status, g.Current.Id), Value: nil})
		}
	}
	return &set, nil
}

func (e *Paracross) ExecDelLocal_SupervisionNodeConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParaSupervisionNodeConfig {
			var g pt.ReceiptParaNodeGroupConfig
			err := types.Decode(log.Log, &g)
			if err != nil {
				return nil, err
			}
			if g.Prev != nil {
				set.KV = append(set.KV, &types.KeyValue{
					Key: calcLocalSupervisionNodeStatusTitle(g.Current.Title, g.Prev.Status, g.Current.TargetAddrs, g.Current.Id), Value: types.Encode(g.Prev)})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalSupervisionNodeStatusTitle(g.Current.Title, g.Current.Status, g.Current.TargetAddrs, g.Current.Id), Value: nil})
		}
	}
	return &set, nil
}

//ExecDelLocal_AssetTransfer asset transfer del local db process
func (e *Paracross) ExecDelLocal_AssetTransfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet

	//  主链转出记录，
	//  转入在 commit done 时记录， 因为没有日志里没有当时tx信息
	r, err := e.initLocalAssetTransfer(tx, true, nil)
	if err != nil {
		return nil, err
	}
	set.KV = append(set.KV, r)

	return &set, nil
}

//ExecDelLocal_AssetWithdraw asset withdraw local db process
func (e *Paracross) ExecDelLocal_AssetWithdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecDelLocal_CrossAssetTransfer asset transfer del local db process
func (e *Paracross) ExecDelLocal_CrossAssetTransfer(payload *pt.CrossAssetTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	r, err := e.initLocalAssetTransfer(tx, true, nil)
	if err != nil {
		return nil, err
	}
	set.KV = append(set.KV, r)

	return &set, nil
}

//ExecDelLocal_Miner miner tx del local db process
func (e *Paracross) ExecDelLocal_Miner(payload *pt.ParacrossMinerAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if index != 0 {
		return nil, pt.ErrParaMinerBaseIndex
	}

	var set types.LocalDBSet
	set.KV = append(set.KV, &types.KeyValue{Key: pt.CalcMinerHeightKey(payload.Status.Title, payload.Status.Height), Value: nil})

	return &set, nil
}

//ExecDelLocal_Transfer asset transfer del local process
func (e *Paracross) ExecDelLocal_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecDelLocal_Withdraw asset withdraw del local db process
func (e *Paracross) ExecDelLocal_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecDelLocal_TransferToExec asset transfer to exec del local db process
func (e *Paracross) ExecDelLocal_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return nil, nil
}

//ExecDelLocal_SelfStageConfig transfer asset to exec local db process
func (e *Paracross) ExecDelLocal_SelfStageConfig(payload *pt.ParaStageConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.execAutoDelLocal(tx, receiptData)
}

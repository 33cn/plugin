// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//Exec_Commit consensus commit tx exec process
func (e *Paracross) Exec_Commit(payload *pt.ParacrossCommitAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	receipt, err := a.Commit(payload)
	if err != nil {
		clog.Error("Paracross commit failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_AssetTransfer asset transfer exec process
func (e *Paracross) Exec_AssetTransfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	clog.Debug("Paracross.Exec", "transfer", "")
	_, err := e.checkTxGroup(tx, index)
	if err != nil {
		clog.Error("ParacrossActionAssetTransfer", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	a := newAction(e, tx)
	receipt, err := a.AssetTransfer(payload)
	if err != nil {
		clog.Error("Paracross AssetTransfer failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_AssetWithdraw asset withdraw exec process
func (e *Paracross) Exec_AssetWithdraw(payload *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	_, err := e.checkTxGroup(tx, index)
	if err != nil {
		clog.Error("ParacrossActionAssetWithdraw", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	a := newAction(e, tx)
	receipt, err := a.AssetWithdraw(payload)
	if err != nil {
		clog.Error("ParacrossActionAssetWithdraw failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_CrossAssetTransfer parallel chain asset transfer exec process
func (e *Paracross) Exec_CrossAssetTransfer(payload *pt.CrossAssetTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	_, err := e.checkTxGroup(tx, index)
	if err != nil {
		clog.Error("ParacrossActionCrossAssetTransfer", "get tx group failed", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	a := newAction(e, tx)
	receipt, err := a.CrossAssetTransfer(payload)
	if err != nil {
		clog.Error("Paracross CrossAssetTransfer failed", "error", err, "hash", hex.EncodeToString(tx.Hash()))
		return nil, err
	}
	return receipt, nil
}

//Exec_Miner miner tx exec process
func (e *Paracross) Exec_Miner(payload *pt.ParacrossMinerAction, tx *types.Transaction, index int) (*types.Receipt, error) {
	if index != 0 && payload.AddIssueCoins <= 0 {
		return nil, pt.ErrParaMinerBaseIndex
	}
	cfg := e.GetAPI().GetConfig()
	if !cfg.IsPara() {
		return nil, types.ErrNotSupport
	}
	a := newAction(e, tx)
	return a.Miner(payload)
}

//Exec_Transfer exec asset transfer process
func (e *Paracross) Exec_Transfer(payload *types.AssetsTransfer, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.Transfer(payload, tx, index)
}

//Exec_Withdraw exec asset withdraw
func (e *Paracross) Exec_Withdraw(payload *types.AssetsWithdraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.Withdraw(payload, tx, index)
}

//Exec_TransferToExec exec transfer asset
func (e *Paracross) Exec_TransferToExec(payload *types.AssetsTransferToExec, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.TransferToExec(payload, tx, index)
}

//Exec_NodeConfig exec super node config
func (e *Paracross) Exec_NodeConfig(payload *pt.ParaNodeAddrConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.NodeConfig(payload)
}

//Exec_NodeGroupConfig node group config process
func (e *Paracross) Exec_NodeGroupConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.NodeGroupConfig(payload)
}

//Exec_SelfStageConfig node group config process
func (e *Paracross) Exec_SelfStageConfig(payload *pt.ParaStageConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.SelfStageConfig(payload)
}

//Exec_ParaBindMiner node group config process
func (e *Paracross) Exec_ParaBindMiner(payload *pt.ParaBindMinerCmd, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.bindMiner(payload)
}

//Exec_SupervisionNodeConfig exec Supervision node config
func (e *Paracross) Exec_SupervisionNodeConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(e, tx)
	return a.SupervisionNodeConfig(payload)
}

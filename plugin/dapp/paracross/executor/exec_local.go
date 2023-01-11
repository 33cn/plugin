// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"
	"math/big"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//ExecLocal_Commit commit tx local db process
func (e *Paracross) ExecLocal_Commit(payload *pt.ParacrossCommitAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	cfg := e.GetAPI().GetConfig()
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
			if !cfg.IsPara() && g.Height > 0 {
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
		} else if log.Ty == pt.TyLogParacrossCommitDone {
			var g pt.ReceiptParacrossDone
			types.Decode(log.Log, &g)

			key := calcLocalTitleKey(g.Title)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})

			key = calcLocalHeightKey(g.Title, g.Height)
			set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&g)})
			if !cfg.IsPara() && g.Height > 0 {
				r, err := e.saveLocalParaTxsFork(&g, false)
				if err != nil {
					return nil, err
				}
				set.KV = append(set.KV, r.KV...)
			}

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
		} else if log.Ty == pt.TyLogParaNodeConfig {
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

func (e *Paracross) ExecLocal_SupervisionNodeConfig(payload *pt.ParaNodeGroupConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
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
					Key: calcLocalSupervisionNodeStatusTitle(g.Current.Title, g.Prev.Status, g.Current.TargetAddrs, g.Current.Id), Value: nil})
			}

			set.KV = append(set.KV, &types.KeyValue{
				Key: calcLocalSupervisionNodeStatusTitle(g.Current.Title, g.Current.Status, g.Current.TargetAddrs, g.Current.Id), Value: types.Encode(g.Current)})
		}
	}
	return &set, nil
}

//ExecLocal_AssetTransfer asset transfer local proc
func (e *Paracross) ExecLocal_AssetTransfer(payload *types.AssetsTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet

	//  主链转出记录，
	//  转入在 commit done 时记录， 因为没有日志里没有当时tx信息
	asset, err := e.getAssetTransferInfo(tx, payload.Cointoken, false)
	if err != nil {
		return nil, err
	}
	r, err := e.initLocalAssetTransfer(tx, false, asset)
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

//ExecLocal_CrossAssetTransfer asset transfer local proc
func (e *Paracross) ExecLocal_CrossAssetTransfer(payload *pt.CrossAssetTransfer, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var set types.LocalDBSet
	cfg := e.GetAPI().GetConfig()
	act, err := getCrossAction(payload, string(tx.Execer))
	if err != nil {
		clog.Crit("local CrossAssetTransfer getCrossAction failed", "error", err)
		return nil, err
	}
	//  主链转出和平行链提取记录，
	//  主链提取和平行链转出在 commit done 时记录
	if !cfg.IsPara() && (act == pt.ParacrossMainAssetWithdraw || act == pt.ParacrossParaAssetTransfer) {
		return nil, nil
	}
	asset, err := e.getCrossAssetTransferInfo(payload, tx, act)
	if err != nil {
		return nil, err
	}
	r, err := e.initLocalAssetTransfer(tx, false, asset)
	if err != nil {
		return nil, err
	}
	set.KV = append(set.KV, r)

	return &set, nil
}

func setMinerTxResult(cfg *types.Chain33Config, payload *pt.ParacrossMinerAction, txs []*types.Transaction, receipts []*types.ReceiptData) error {
	isCommitTx := make(map[string]bool)
	var curTxHashs, paraTxHashs, crossTxHashs [][]byte
	for _, tx := range txs {
		hash := tx.Hash()
		curTxHashs = append(curTxHashs, hash)
		//对user.p.xx.paracross ,actionTy==commit 的tx不需要再发回主链
		if cfg.IsMyParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
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
		if cfg.IsMyParaExecName(string(tx.Execer)) && !isCommitTx[string(hash)] {
			paraTxHashs = append(paraTxHashs, hash)
		}
	}
	totalCrossTxHashs := FilterParaMainCrossTxHashes(cfg.GetTitle(), txs)
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

//带版本号编码的bitmap，版本号占4bit位，除去版本号位，最高位为跨链交易个数，后面的则是交易结果bitmap
//比如00011110， 版本号为0001,1110的最高位1索引为3，代表后面3个交易，三个交易110代表第0个交易是失败的，其余的是ok的
//如果没有数量表示，在所有跨链交易都是失败的时候，返回的是个空值,无法区分是失败还是无交易
func getCrossAssetTxBitMap(crossAssetTxHashs, allTxHashs [][]byte, receipts []*types.ReceiptData) string {
	rst := pt.ParaCrossStatusBitMapVer1
	if len(crossAssetTxHashs) > 0 {
		crossTxBitmap := util.CalcBitMap(crossAssetTxHashs, allTxHashs, receipts)
		val := big.NewInt(0)
		val.SetBytes(crossTxBitmap)
		val.SetBit(val, len(crossAssetTxHashs), 1)
		rst += val.Text(2)
	}
	return rst
}

func setMinerTxResultFork(cfg *types.Chain33Config, status *pt.ParacrossNodeStatus, txs []*types.Transaction, receipts []*types.ReceiptData) error {
	isCommitTx := make(map[string]bool)
	var allTxHashs [][]byte
	for _, tx := range txs {
		hash := tx.Hash()
		allTxHashs = append(allTxHashs, hash)

		if cfg.IsMyParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
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

	//有tx且全部是user.p.x.paracross的commit tx时候是空块，发送时候需要等有实际交易时候再发
	//如果当前区块除了minerTx没有交易，len(allTxHash)=0, 也认为是没有实际的交易
	status.NonCommitTxCounts = 1
	if len(allTxHashs) == len(isCommitTx) {
		status.NonCommitTxCounts = 0
	}

	//主链自己过滤平行链tx， 对平行链执行失败的tx主链无法识别，主链和平行链需要获取相同的最初的tx map
	//全部平行链tx结果
	status.TxResult = []byte(hex.EncodeToString(util.CalcSingleBitMap(allTxHashs, receipts)))

	//获取跨链资产转移交易信息
	crossAssetTxHashs, err := FilterParaCrossAssetTxHashes(txs)
	if err != nil {
		return err
	}
	status.CrossTxResult = []byte(getCrossAssetTxBitMap(crossAssetTxHashs, allTxHashs, receipts))

	//ForkLoopCheckCommitTxDone 后只保留全部txreseult 结果
	if !pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		//跨链tx结果
		crossTxHashs := FilterParaCrossTxHashes(txs)
		status.CrossTxResult = []byte(hex.EncodeToString(util.CalcBitMap(crossTxHashs, allTxHashs, receipts)))
		status.TxHashs = [][]byte{CalcTxHashsHash(allTxHashs)}
		status.CrossTxHashs = [][]byte{CalcTxHashsHash(crossTxHashs)}
	}

	return nil
}

//ExecLocal_Miner miner tx local db process
func (e *Paracross) ExecLocal_Miner(payload *pt.ParacrossMinerAction, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if index != 0 {
		return nil, pt.ErrParaMinerBaseIndex
	}

	var set types.LocalDBSet
	txs := e.GetTxs()
	cfg := e.GetAPI().GetConfig()

	//removed the 0 vote tx
	if pt.IsParaForkHeight(cfg, payload.Status.MainBlockHeight, pt.ForkCommitTx) {
		err := setMinerTxResultFork(cfg, payload.Status, txs[1:], e.GetReceipt()[1:])
		if err != nil {
			return nil, err
		}
	} else {
		err := setMinerTxResult(cfg, payload, txs[1:], e.GetReceipt()[1:])
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

//ExecLocal_SelfStageConfig transfer asset to exec local db process
func (e *Paracross) ExecLocal_SelfStageConfig(payload *pt.ParaStageConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.execAutoLocalStage(tx, receiptData, index)
}

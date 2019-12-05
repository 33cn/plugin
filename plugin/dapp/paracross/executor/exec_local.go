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
	cfg := e.GetAPI().GetConfig()
	validCommit := false
	for _, log := range receiptData.Logs {
		if log.Ty == pt.TyLogParacrossCommit {
			validCommit = true
			var g pt.ReceiptParacrossCommit
			types.Decode(log.Log, &g)

			var r pt.ParacrossTx
			r.TxHash = common.ToHex(tx.Hash())
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: types.Encode(&r)})
		} else if log.Ty == pt.TyLogParacrossCommitDone {
			validCommit = true
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
			validCommit = true
			var g pt.ReceiptParacrossRecord
			types.Decode(log.Log, &g)

			var r pt.ParacrossTx
			r.TxHash = common.ToHex(tx.Hash())
			set.KV = append(set.KV, &types.KeyValue{Key: calcLocalTxKey(g.Status.Title, g.Status.Height, tx.From()), Value: types.Encode(&r)})
		}
	}
	if validCommit {
		clog.Debug("ExecLocal_Commit", "adding a public key for title:",payload.Status.Title,
			"addr:", tx.From(), "public key:", common.ToHex(tx.Signature.Pubkey))

		e.setNodePubkey(payload.Status.Title, tx.From(), tx.Signature.Pubkey, &set)

		clog.Debug("ExecLocal_Commit", "adding a public key with key:",string(set.KV[len(set.KV) - 1].Key))
	}

	return &set, nil
}
//在此处收集超级节点的公钥
func (e *Paracross) setNodePubkey(title, addr string, pubKey []byte, set *types.LocalDBSet) error {
	localDB := e.GetLocalDB()
	var paraNodeAddrPubKey pt.ParaNodeAddrPubKey
	key := calcParaSuperNodePubKey(title)
	clog.Debug("setNodePubkey", "key:", string(key))
	nodePubKey, error := localDB.Get(key)
	if nil != error {
		paraNodeAddrPubKey.Addr2Pubkey  = make(map[string][]byte)
		paraNodeAddrPubKey.Addr2Pubkey[addr] = pubKey
		set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&paraNodeAddrPubKey)})
		clog.Debug("setNodePubkey", "Set for the 1st time with key:", string(set.KV[len(set.KV) - 1].Key))
		return nil
	}
	//如果存在时，不用确认是否解码成功，也不用确认该地址对应的公钥是否进行保存，
	if err := types.Decode(nodePubKey, &paraNodeAddrPubKey); nil != err {
		clog.Error("setNodePubkey", "Failed to decode due to :", err.Error())
		paraNodeAddrPubKey.Addr2Pubkey  = make(map[string][]byte)
		paraNodeAddrPubKey.Addr2Pubkey[addr] = pubKey
		set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&paraNodeAddrPubKey)})
		return nil
	}
	if _, ok := paraNodeAddrPubKey.Addr2Pubkey[addr]; ok {
		return nil
	}

	paraNodeAddrPubKey.Addr2Pubkey[addr] = pubKey
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&paraNodeAddrPubKey)})
	clog.Debug("setNodePubkey", "Succeed to add public key for:", string(set.KV[len(set.KV) - 1].Key))
	return nil
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
		} else if pt.TyLogParaNodeGroupAddrsUpdate == log.Ty {
			var receiptConfig types.ReceiptConfig
			if err := types.Decode(log.Log, &receiptConfig); nil != err {
				return nil, err
			}
			prev, current := receiptConfig.Prev, receiptConfig.Current
			prevNodes := prev.GetArr()
			currentNodes := current.GetArr()
			//只处理减少的情况
			if len(prevNodes.Value) > len(currentNodes.Value) {
				localDB := e.GetLocalDB()
				var paraNodeAddrPubKey pt.ParaNodeAddrPubKey
				key := calcParaSuperNodePubKey(payload.Title)
				nodePubKey, err := localDB.Get(key)
				if nil != err {
					clog.Error("ExecLocal_NodeConfig", "failed get info from local db with key:", string(key))
					continue
				}
				if err := types.Decode(nodePubKey, &paraNodeAddrPubKey); nil != err {
					return nil, err
				}

				paraNodeAddrPubKeyNew := pt.ParaNodeAddrPubKey{
					Addr2Pubkey:make(map[string][]byte),
				}
				//遍历当前所有的配置节点，获取相应的公钥填充到新的公钥信息列表中
				for _, node := range currentNodes.Value {
					if pubkey, exist := paraNodeAddrPubKey.Addr2Pubkey[node]; exist {
						paraNodeAddrPubKeyNew.Addr2Pubkey[node] = pubkey
					}
				}
				set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(&paraNodeAddrPubKeyNew)})
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

func setMinerTxResultFork(cfg *types.Chain33Config, status *pt.ParacrossNodeStatus, txs []*types.Transaction, receipts []*types.ReceiptData) error {
	isCommitTx := make(map[string]bool)
	var curTxHashs [][]byte
	for _, tx := range txs {
		hash := tx.Hash()
		curTxHashs = append(curTxHashs, hash)

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

	//有tx且全部是user.p.x.paracross的commit tx时候设为0
	status.NonCommitTxCounts = 1
	if len(curTxHashs) != 0 && len(curTxHashs) == len(isCommitTx) {
		status.NonCommitTxCounts = 0
	}

	//主链自己过滤平行链tx， 对平行链执行失败的tx主链无法识别，主链和平行链需要获取相同的最初的tx map
	//全部平行链tx结果
	status.TxResult = []byte(hex.EncodeToString(util.CalcSingleBitMap(curTxHashs, receipts)))
	clog.Debug("setMinerTxResultFork", "height", status.Height, "txResult", string(status.TxResult))

	//ForkLoopCheckCommitTxDone 后只保留全部txreseult 结果
	if !pt.IsParaForkHeight(cfg, status.MainBlockHeight, pt.ForkLoopCheckCommitTxDone) {
		//跨链tx结果
		crossTxHashs := FilterParaCrossTxHashes(txs)
		status.CrossTxResult = []byte(hex.EncodeToString(util.CalcBitMap(crossTxHashs, curTxHashs, receipts)))
		status.TxHashs = [][]byte{CalcTxHashsHash(curTxHashs)}
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

//ExecLocal_SelfConsensStageConfig transfer asset to exec local db process
func (e *Paracross) ExecLocal_SelfStageConfig(payload *pt.ParaStageConfig, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return e.execAutoLocalStage(tx, receiptData, index)
}

package executor

import (
	"bytes"
	"encoding/hex"
	"errors"

	rexec "github.com/33cn/plugin/plugin/dapp/rollup/executor"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

//Exec_RollupCrossTx exec commit rollup
func (p *Paracross) Exec_RollupCrossTx(commit *pt.RollupCrossTx, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(p, tx)
	return a.rollupCrossTx(commit)
}

//当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (p *Paracross) setAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = p.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

//ExecLocal_RollupCrossTx exec local commit rollup
func (p *Paracross) ExecLocal_RollupCrossTx(commit *pt.RollupCrossTx, tx *types.Transaction,
	receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {

	dbSet := &types.LocalDBSet{}

	rollupLog := &pt.RollupCrossTxLog{}
	err := types.Decode(receiptData.Logs[0].Log, rollupLog)
	if err != nil {
		clog.Error("ExecLocal_RollupCrossTx", "commitRound", commit.GetCommitRound(),
			"txHash", hex.EncodeToString(tx.Hash()), "decode err", err)
		return nil, types.ErrDecode
	}

	crossTxHashes, crossTxs, err := getRollupCrossTxs(p.GetAPI(), commit.GetChainTitle(), commit.GetTxIndices())

	if err != nil {
		clog.Error("ExecLocal_RollupCrossTx", "commitRound", commit.GetCommitRound(), "getRollupCrossTxs err", err)
		return nil, ErrGetRollupCrossTx
	}
	crossTxResults, _ := common.FromHex(rollupLog.CrossTxResults)
	for i, crossTx := range crossTxs {
		execOK := util.BitMapBit(crossTxResults, uint32(i))
		paraHeight := commit.GetTxIndices()[i].BlockHeight
		set, err := p.updateLocalParaTx(commit.GetChainTitle(), paraHeight, crossTx, execOK, false)
		if err != nil {
			clog.Error("ExecLocal_RollupCrossTx", "title", commit.GetChainTitle(), "height", paraHeight,
				"txIndex", i, "txHash", hex.EncodeToString(crossTxHashes[i]),
				"execOK", execOK, "err", err)
			return nil, err
		}

		dbSet.KV = append(dbSet.KV, set.KV...)
	}

	return p.setAutoRollBack(tx, dbSet.KV), nil
}

//ExecDelLocal_RollupCrossTx exec local commit rollup
func (p *Paracross) ExecDelLocal_RollupCrossTx(_ *pt.RollupCrossTx, tx *types.Transaction,
	_ *types.ReceiptData, _ int) (*types.LocalDBSet, error) {
	kvs, err := p.DelRollbackKV(tx, tx.Execer)
	if err != nil {
		return nil, err
	}
	dbSet := &types.LocalDBSet{}
	dbSet.KV = append(dbSet.KV, kvs...)
	return dbSet, nil
}

var (
	ErrInvalidCommitRound   = errors.New("ErrInvalidCommitRound")
	ErrInvalidChain         = errors.New("ErrInvalidChain")
	ErrGetRollupCrossTx     = errors.New("ErrGetRollupCrossTx")
	ErrGetRollupCommitRound = errors.New("ErrGetRollupCommitRound")
	ErrCrossTxCheckHash     = errors.New("ErrCrossTxCheckHash")
)

func (a *action) rollupCrossTx(commit *pt.RollupCrossTx) (*types.Receipt, error) {
	clog.Debug("rollupCrossTx", "title", commit.GetChainTitle(),
		"commitRound", commit.GetCommitRound(), "txHash", common.ToHex(a.txhash))
	if a.api.GetConfig().IsPara() {
		return nil, ErrInvalidChain
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	status, err := rexec.GetRollupStatus(a.db, commit.GetChainTitle())

	if err != nil || status.CommitRound != commit.GetCommitRound() {

		clog.Error("rollupCrossTx", "currRound", status.GetCommitRound(),
			"commitRound", commit.GetCommitRound(), "getRollupStatus err", err)
		return nil, ErrInvalidCommitRound
	}

	roundInfo, err := rexec.GetRoundInfo(a.db, commit.GetChainTitle(), commit.GetCommitRound())
	if err != nil {
		clog.Error("rollupCrossTx", "commitRound", commit.GetCommitRound(), "getRollupCommitRound err", err)
		return nil, ErrGetRollupCommitRound
	}

	crossTxHashes, crossTxs, err := getRollupCrossTxs(a.api, commit.GetChainTitle(), commit.GetTxIndices())

	if err != nil {
		clog.Error("rollupCrossTx", "commitRound", commit.GetCommitRound(), "getRollupCrossTxs err", err)
		return nil, ErrGetRollupCrossTx
	}

	checkHash := common.ToHex(CalcTxHashsHash(crossTxHashes))
	if roundInfo.CrossTxCheckHash != checkHash {
		clog.Error("rollupCrossTx", "commitRound", commit.GetCommitRound(),
			"calcHash", checkHash, "commitHash", roundInfo.CrossTxCheckHash)

		for _, hash := range crossTxHashes {
			clog.Error("RollupCrossTx cross tx info", "txhash", common.ToHex(hash))
		}
		return nil, ErrCrossTxCheckHash
	}

	crossTxResults, _ := common.FromHex(roundInfo.CrossTxResults)

	rollupLog := &pt.RollupCrossTxLog{
		CommitRound:      commit.GetCommitRound(),
		ChainTitle:       commit.GetChainTitle(),
		CrossTxCheckHash: checkHash,
		CrossTxResults:   roundInfo.CrossTxResults,
		CrossTxHashes:    make([]string, 0, len(crossTxs)),
	}
	for _, tx := range crossTxs {
		rollupLog.CrossTxHashes = append(rollupLog.CrossTxHashes, hex.EncodeToString(tx.Hash()))
	}

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  pt.TyLogParaRollupCrossTx,
		Log: types.Encode(rollupLog),
	})

	rep, err := a.execCrossTxs(commit.GetChainTitle(), commit.GetCommitRound(), crossTxs, crossTxResults)
	if err != nil {

		clog.Error("RollupCrossTx", "commitRound", commit.GetCommitRound(), "execCrossTxs err", err)
		return nil, err
	}

	return mergeReceipt(receipt, rep), nil
}

func getTx(api client.QueueProtocolAPI, hash []byte) (*types.Transaction, error) {

	detail, err := api.QueryTx(&types.ReqHash{Hash: hash})
	return detail.GetTx(), err
}

func getRollupCrossTxs(api client.QueueProtocolAPI, paraTitle string, idxArr []*pt.CrossTxIndex) ([][]byte, []*types.Transaction, error) {

	blkCrossTxCache := make(map[int64][]*types.Transaction, 4)
	crossTxs := make([]*types.Transaction, 0, len(idxArr))
	crossTxHashes := make([][]byte, 0, len(idxArr))
	cfg := api.GetConfig()
	for _, txIdx := range idxArr {

		// first get from cache
		blkCrossTxs, ok := blkCrossTxCache[txIdx.BlockHeight]
		if !ok && txIdx.BlockHeight > 0{

			// get block from blockchain
			detail, err := getBlockByHeight(api, txIdx.BlockHeight, true)
			if err != nil {
				clog.Error("getRollupCrossTxs", "height", txIdx.BlockHeight, "getBlock err", err)
				return nil, nil, err
			}

			blkCrossTxs = FilterParaCrossTxs(FilterTxsForPara(cfg, detail.FilterParaTxsByTitle(cfg, paraTitle)))
			blkCrossTxCache[txIdx.BlockHeight] = blkCrossTxs
		}
		var crossTx *types.Transaction
		if txIdx.BlockHeight > 0 && int(txIdx.FilterIndex) < len(blkCrossTxs) {
			crossTx = blkCrossTxs[txIdx.FilterIndex]
		}
		// 通过索引无法获取或者获取的交易数据不对, 可能是主链回滚导致索引信息混乱, 需要通过交易哈希查询交易
		if crossTx == nil || !bytes.Equal(crossTx.Hash(), txIdx.TxHash) {

			txHash := hex.EncodeToString(txIdx.TxHash)
			clog.Debug("getRollupCrossTxs", "paraTitle", paraTitle,
				"filterIdx", txIdx.FilterIndex, "len", len(blkCrossTxs),
				"height", txIdx.BlockHeight, "txHash", txHash)

			tx, err := getTx(api, txIdx.TxHash)
			if err != nil {
				clog.Error("getRollupCrossTxs", "txHash", txHash, "getTx err", err)
				return nil, nil, err
			}
			crossTx = tx
		}

		crossTxs = append(crossTxs, crossTx)
		crossTxHashes = append(crossTxHashes, txIdx.TxHash)
	}

	return crossTxHashes, crossTxs, nil
}

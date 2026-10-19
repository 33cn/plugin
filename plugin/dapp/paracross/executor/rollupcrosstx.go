package executor

import (
	"bytes"
	"encoding/hex"
	"errors"

	rexec "github.com/33cn/plugin/plugin/dapp/rollup/executor"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// Exec_RollupCrossTx exec commit rollup
func (p *Paracross) Exec_RollupCrossTx(commit *pt.RollupCrossTx, tx *types.Transaction, index int) (*types.Receipt, error) {
	a := newAction(p, tx)
	return a.rollupCrossTx(commit)
}

// 当区块回滚时，框架支持自动回滚localdb kv，需要对exec-local返回的kv进行封装
func (p *Paracross) setAutoRollBack(tx *types.Transaction, kv []*types.KeyValue) *types.LocalDBSet {

	dbSet := &types.LocalDBSet{}
	dbSet.KV = p.AddRollbackKV(tx, tx.Execer, kv)
	return dbSet
}

// ExecLocal_RollupCrossTx exec local commit rollup
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

	crossTxHashes, crossTxs, err := getRollupCrossTxs(p.GetAPI(), p.GetMainHeight(), commit.GetChainTitle(), commit.GetTxIndices())

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

// ExecDelLocal_RollupCrossTx exec local commit rollup
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
	ErrRollupRoundSettled   = errors.New("ErrRollupRoundSettled")
	ErrGetRollupCrossDone   = errors.New("ErrGetRollupCrossDone")
)

func (a *action) rollupCrossTx(commit *pt.RollupCrossTx) (*types.Receipt, error) {
	clog.Debug("rollupCrossTx", "title", commit.GetChainTitle(),
		"commitRound", commit.GetCommitRound(), "txHash", common.ToHex(a.txhash))
	cfg := a.api.GetConfig()
	if cfg.IsPara() {
		return nil, ErrInvalidChain
	}
	//分叉后rollup的chainTitle必须是规范平行链title(user.p.<name>.): 非规范title(例如真实链title去掉结尾'.')
	//会与真实平行链使用不同的节点组状态key，却指向同一个execer，必须拒绝
	if cfg.IsDappFork(a.exec.GetMainHeight(), pt.ParaX, pt.ForkParaTitleExecMatch) &&
		!pt.IsValidParaTitle(commit.GetChainTitle()) {
		clog.Error("rollupCrossTx invalid chainTitle", "chainTitle", commit.GetChainTitle(),
			"txHash", common.ToHex(a.txhash))
		return nil, rexec.ErrChainTitle
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	status, err := rexec.GetRollupStatus(a.db, commit.GetChainTitle())

	if err != nil || status.CommitRound != commit.GetCommitRound() {

		clog.Error("rollupCrossTx", "currRound", status.GetCommitRound(),
			"commitRound", commit.GetCommitRound(), "getRollupStatus err", err)
		return nil, ErrInvalidCommitRound
	}

	//分叉后同一轮的结算只允许执行一次(见 rollup types.ForkSecurityHardening, 该分叉注册在 rollup 名下,
	//paracross 与 rollup 共用同一个配置项)。
	//此前准入只校验"轮次相等", 而当前轮次要到【下一轮】提交时才推进, 因此在被打断之前任何人都能重放
	//这笔结算(改 nonce 即新交易), 反复从该 title 的托管池支取/重复铸币。
	settleOnce := cfg.IsDappFork(a.exec.GetMainHeight(), rtypes.RollupX, rtypes.ForkSecurityHardening)
	if settleOnce {
		doneRound, err := getRollupCrossDone(a.db, commit.GetChainTitle())
		if err != nil {
			clog.Error("rollupCrossTx getRollupCrossDone", "chainTitle", commit.GetChainTitle(),
				"commitRound", commit.GetCommitRound(), "txHash", common.ToHex(a.txhash), "err", err)
			return nil, ErrGetRollupCrossDone
		}
		if commit.GetCommitRound() <= doneRound {
			clog.Error("rollupCrossTx round already settled", "chainTitle", commit.GetChainTitle(),
				"commitRound", commit.GetCommitRound(), "doneRound", doneRound, "txHash", common.ToHex(a.txhash))
			return nil, ErrRollupRoundSettled
		}
	}

	roundInfo, err := rexec.GetRoundInfo(a.db, commit.GetChainTitle(), commit.GetCommitRound())
	if err != nil {
		clog.Error("rollupCrossTx", "commitRound", commit.GetCommitRound(), "getRollupCommitRound err", err)
		return nil, ErrGetRollupCommitRound
	}

	crossTxHashes, crossTxs, err := getRollupCrossTxs(a.api, a.exec.GetMainHeight(), commit.GetChainTitle(), commit.GetTxIndices())

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

	if settleOnce {
		//结算成功才推进水位: 该 KV 随本笔交易一起落库, 交易失败回滚时水位也不会前进
		receipt.KV = append(receipt.KV, &types.KeyValue{
			Key:   calcRollupCrossDoneKey(commit.GetChainTitle()),
			Value: types.Encode(&types.Int64{Data: commit.GetCommitRound()}),
		})
	}

	return mergeReceipt(receipt, rep), nil
}

// getRollupCrossDone 读取某 title 已结算的最大 rollup 轮次, 未记录过时返回 0
func getRollupCrossDone(db dbm.KV, title string) (int64, error) {
	value, err := db.Get(calcRollupCrossDoneKey(title))
	if err != nil {
		if isNotFound(err) {
			return 0, nil
		}
		return 0, err
	}
	var round types.Int64
	if err := types.Decode(value, &round); err != nil {
		return 0, err
	}
	return round.Data, nil
}

func getTx(api client.QueueProtocolAPI, hash []byte) (*types.Transaction, error) {

	detail, err := api.QueryTx(&types.ReqHash{Hash: hash})
	return detail.GetTx(), err
}

func getRollupCrossTxs(api client.QueueProtocolAPI, mainHeight int64, paraTitle string, idxArr []*pt.CrossTxIndex) ([][]byte, []*types.Transaction, error) {

	blkCrossTxCache := make(map[int64][]*types.Transaction, 4)
	crossTxs := make([]*types.Transaction, 0, len(idxArr))
	crossTxHashes := make([][]byte, 0, len(idxArr))
	cfg := api.GetConfig()
	for _, txIdx := range idxArr {

		// first get from cache
		blkCrossTxs, ok := blkCrossTxCache[txIdx.BlockHeight]
		if !ok && txIdx.BlockHeight > 0 {

			// get block from blockchain
			detail, err := getBlockByHeight(api, txIdx.BlockHeight, true)
			if err != nil {
				clog.Error("getRollupCrossTxs", "height", txIdx.BlockHeight, "getBlock err", err)
				return nil, nil, err
			}

			//分叉后按title精确匹配，非规范title(如真实链title去掉结尾'.')不能索引到其他平行链的跨链交易
			blkCrossTxs = filterCrossTxsByExactTitle(cfg, mainHeight, paraTitle,
				FilterParaCrossTxs(FilterTxsForPara(cfg, detail.FilterParaTxsByTitle(cfg, paraTitle))))
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

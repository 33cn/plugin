package executor

import (
	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (r *rollup) Exec_Commit(commit *rolluptypes.CheckPoint, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	commitRound := commit.GetCommitRound()
	status, err := r.getRollupStatus(commit.GetChainTitle())
	if err != nil {
		elog.Error("Exec_CommitBatch", "title", commit.GetChainTitle(),
			"round", commitRound, "get status err", err)
		return nil, ErrGetRollupStatus
	}

	parentHash := common.ToHex(commit.GetBatch().GetBlockHeaders()[0].ParentHash)

	// check parent block hash with last commit round
	// 首次提交没有status记录, lastBlockHash为空
	if len(status.CommitBlockHash) > 0 && status.CommitBlockHash != parentHash {

		elog.Error("Exec_CommitBatch", "title", commit.GetChainTitle(),
			"round", commitRound, "currLastHash", status.CommitBlockHash,
			"parentHash", parentHash)
		return nil, ErrParentHashNotEqual
	}

	headers := commit.GetBatch().GetBlockHeaders()
	roundInfo := &rolluptypes.CommitRoundInfo{
		CommitRound:      commitRound,
		ParentBlockHash:  parentHash,
		LastBlockHash:    calcBlockHash(headers[len(headers)-1]),
		LastBlockHeight:  headers[len(headers)-1].Height,
		CrossTxCheckHash: common.ToHex(commit.GetBatch().GetCrossTxCheckHash()),
		CrossTxResults:   common.ToHex(commit.GetBatch().GetCrossTxResults()),
	}

	encodeVal := types.Encode(roundInfo)
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatCommitRoundInfoKey(commit.GetChainTitle(), commitRound),
		Value: encodeVal,
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  rolluptypes.TyCommitRoundInfoLog,
		Log: encodeVal,
	})

	status.Timestamp = r.GetBlockTime()
	status.CommitRound = roundInfo.CommitRound
	status.CommitBlockHeight = roundInfo.LastBlockHeight
	status.CommitBlockHash = roundInfo.LastBlockHash
	status.CommitAddr = tx.From()
	status.CrossTxSyncedHeight = commit.CrossTxSyncedHeight
	encodeVal = types.Encode(status)
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatRollupStatusKey(commit.GetChainTitle()),
		Value: encodeVal,
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  rolluptypes.TyRollupStatusLog,
		Log: encodeVal,
	})

	return receipt, nil
}

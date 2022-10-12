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
		return nil, errGetRollupStatus
	}

	parentHash := common.ToHex(commit.GetBatch().GetBlockHeaders()[0].ParentHash)
	isNextRound := status.CommitRound+1 == commitRound
	// check parent block hash with last commit round
	// 首次提交没有status记录, lastBlockHash为空
	if len(status.CommitBlockHash) > 0 && isNextRound &&
		status.CommitBlockHash != parentHash {

		elog.Error("Exec_CommitBatch", "title", commit.GetChainTitle(),
			"round", commitRound, "currLastHash", status.CommitBlockHash,
			"parentHash", parentHash)
		return nil, errParentHashNotEqual
	}

	headers := commit.GetBatch().GetBlockHeaders()
	roundInfo := &rolluptypes.CommitRoundInfo{
		CommitRound:     commitRound,
		ParentBlockHash: parentHash,
		LastBlockHash:   calcBlockHash(headers[len(headers)-1]),
		LastBlockHeight: headers[len(headers)-1].Height,
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

	// 向后遍历statedb, 检测后续round是否已经提交到链上, 进行串连(batch在提交时是无序的)

	for isNextRound {
		commitRound++

		nextInfo, err := r.getRoundInfo(commit.GetChainTitle(), commitRound)
		if err != nil {
			break
		}
		if roundInfo.LastBlockHash == nextInfo.ParentBlockHash {
			roundInfo = nextInfo
		} else {
			elog.Error("Exec_CommitBatch ParentHashNotMatch", "commitRound", commitRound,
				"expectHash", roundInfo.LastBlockHash,
				"actualHash", nextInfo.ParentBlockHash)
		}
	}

	if isNextRound {

		status.Timestamp = r.GetBlockTime()
		status.CommitRound = roundInfo.CommitRound
		status.CommitBlockHeight = roundInfo.LastBlockHeight
		status.CommitBlockHash = roundInfo.LastBlockHash
		status.CommitAddr = tx.From()
		encodeVal = types.Encode(status)
		receipt.KV = append(receipt.KV, &types.KeyValue{
			Key:   formatRollupStatusKey(commit.GetChainTitle()),
			Value: encodeVal,
		})

		receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
			Ty:  rolluptypes.TyRollupStatusLog,
			Log: encodeVal,
		})
	}

	return receipt, nil
}

package executor

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (r *rollup) Exec_CommitBatch(commit *rolluptypes.CommitBatch, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	commitRound := commit.GetCommitRound()
	status, err := r.getRollupStatus(commit.GetChainTitle())
	if err != nil {
		elog.Error("Exec_CommitBatch", "title", commit.GetChainTitle(),
			"round", commitRound, "get status err", err)
		return nil, errGetRollupStatus
	}

	parentHash := commit.GetBatch().GetBlockHeaders()[0].ParentHash
	isNextRound := status.CurrCommitRound+1 == commitRound
	// check parent block hash with last commit round
	// 首次提交没有status记录, lastBlockHash为空
	if status.LastBlockHash != nil && isNextRound &&
		!bytes.Equal(status.LastBlockHash, parentHash) {

		elog.Error("Exec_CommitBatch", "title", commit.GetChainTitle(),
			"round", commitRound, "currLastHash", hex.EncodeToString(status.LastBlockHash),
			"parentHash", hex.EncodeToString(parentHash))
		return nil, errParentHashNotEqual
	}

	headers := commit.GetBatch().GetBlockHeaders()
	roundInfo := &rolluptypes.CommitRoundInfo{
		CommitRound:     commitRound,
		FirstBlockHash:  calcBlockHash(headers[0]),
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
		if bytes.Equal(roundInfo.LastBlockHash, nextInfo.FirstBlockHash) {
			roundInfo = nextInfo
		}
	}

	if isNextRound {

		status.Timestamp = r.GetBlockTime()
		status.CurrCommitRound = roundInfo.CommitRound
		status.CurrCommitBlockHeight = roundInfo.LastBlockHeight
		status.LastBlockHash = roundInfo.LastBlockHash

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

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"

	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (r *rollup) Exec_Commit(cp *rolluptypes.CheckPoint, tx *types.Transaction, index int) (*types.Receipt, error) {

	receipt := &types.Receipt{Ty: types.ExecOk}

	commitRound := cp.GetCommitRound()
	status, err := GetRollupStatus(r.GetStateDB(), cp.GetChainTitle())
	if err != nil {
		elog.Error("Exec_CommitBatch", "title", cp.GetChainTitle(),
			"round", commitRound, "get status err", err)
		return nil, ErrGetRollupStatus
	}

	headers := cp.GetBatch().GetBlockHeaders()
	roundInfo := &rolluptypes.CommitRoundInfo{
		CommitRound:      commitRound,
		FirstBlockHeight: headers[0].Height,
		LastBlockHeight:  headers[len(headers)-1].Height,
		CommitTxCount:    int32(len(cp.GetBatch().GetTxList())),
		CrossTxCheckHash: common.ToHex(cp.GetBatch().GetCrossTxCheckHash()),
		CrossTxResults:   common.ToHex(cp.GetBatch().GetCrossTxResults()),
	}

	blkHashes := make([][]byte, len(headers))
	for i, h := range headers {
		roundInfo.CommitTxCount += int32(h.TxCount)
		blkHashes[i] = sha256Hash(h)
	}
	roundInfo.BlockRootHash = common.ToHex(merkle.GetMerkleRoot(blkHashes))

	encodeVal := types.Encode(roundInfo)
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatCommitRoundInfoKey(cp.GetChainTitle(), commitRound),
		Value: encodeVal,
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  rolluptypes.TyCommitRoundInfoLog,
		Log: encodeVal,
	})

	status.Timestamp = r.GetBlockTime()
	status.CommitRound = roundInfo.CommitRound
	status.CommitBlockHeight = roundInfo.LastBlockHeight
	status.CommitBlockHash = calcBlockHash(headers[len(headers)-1])
	status.CommitAddr = tx.From()
	status.CrossTxSyncedHeight = cp.CrossTxSyncedHeight
	encodeVal = types.Encode(status)
	receipt.KV = append(receipt.KV, &types.KeyValue{
		Key:   formatRollupStatusKey(cp.GetChainTitle()),
		Value: encodeVal,
	})

	receipt.Logs = append(receipt.Logs, &types.ReceiptLog{
		Ty:  rolluptypes.TyRollupStatusLog,
		Log: encodeVal,
	})

	return receipt, nil
}

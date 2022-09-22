package rollup

import (
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) GetCommitBatch(blocks []*types.Block) *rolluptypes.BlockBatch {

	batch := &rolluptypes.BlockBatch{}

	batch.BlockHeaders = make([]*types.Header, 0, len(blocks))
	batch.TxList = make([][]byte, 0, minCommitTxCount)
	batch.PubKeyList = make([][]byte, 0, minCommitTxCount)
	batch.TxAddrIDList = make([]byte, 0, minCommitTxCount)
	signs := make([]crypto.Signature, 0, minCommitTxCount)
	blsDriver := r.val.blsDriver
	for _, block := range blocks {

		header := block.GetHeader(r.chainCfg)
		header.Hash = nil
		batch.BlockHeaders = append(batch.BlockHeaders, header)
		for _, tx := range block.Txs {

			ctx := types.CloneTx(tx)
			batch.PubKeyList = append(batch.PubKeyList, ctx.Signature.Pubkey)
			// 本地已执行区块, 签名信息合法, 无需错误处理
			sign, _ := blsDriver.SignatureFromBytes(ctx.Signature.GetSignature())
			signs = append(signs, sign)
			batch.TxAddrIDList = append(batch.TxAddrIDList, byte(types.ExtractAddressID(ctx.Signature.Ty)))
			ctx.Signature = nil
			batch.TxList = append(batch.TxList, types.Encode(ctx))

		}
	}

	aggreDriver := r.val.blsDriver.(crypto.AggregateCrypto)
	aggreSign, err := aggreDriver.Aggregate(signs)
	if err != nil {
		rlog.Error("GetCommitBatch", "aggregate sign err", aggreSign)
		return nil
	}
	batch.AggregateTxSign = aggreSign.Bytes()
	return batch
}

// 提交共识
func (r *RollUp) handleCommitBatch() {

	ticker := time.NewTicker(time.Duration(r.cfg.CommitInterval) * time.Second)
	nextCommitRound := r.val.getNextCommitRound()
	for {

		select {

		case <-r.ctx.Done():
			ticker.Stop()
			return
		//case <-
		case <-ticker.C:

			batch := r.cache.getAggregateBatch(nextCommitRound, r.val.aggregateSign)
			// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
			if batch == nil {
				continue
			}

			// build commit tx

			nextCommitRound += int64(r.val.getValidatorCount())
		}

		// 其他节点未提交导致超时
		currRound, timeout := r.val.isRollupCommitTimeout()
		if timeout {

			batch := r.cache.getAggregateBatch(currRound+1, r.val.aggregateSign)
			// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
			if batch == nil {
				continue
			}

		}

	}

}

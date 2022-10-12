package rollup

import (
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) buildBlockBatch(blocks []*types.Block) *rolluptypes.BlockBatch {

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
		rlog.Error("buildBlockBatch", "aggregate sign err", aggreSign)
		return nil
	}
	batch.AggregateTxSign = aggreSign.Bytes()
	return batch
}

// 提交共识
func (r *RollUp) handleCommitCheckPoint() {

	ticker := time.NewTicker(time.Duration(r.cfg.CommitInterval) * time.Second)
	defer ticker.Stop()
	for {

		select {

		case <-r.ctx.Done():
			return
		case <-ticker.C:

			nextCommitRound, ok := r.val.isMyCommitTurn()
			if !ok {
				continue
			}
			cp := r.cache.getPreparedCheckPoint(nextCommitRound, r.val.aggregateSign)
			// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
			if cp == nil {
				continue
			}

			// build commit tx

		}

	}

}

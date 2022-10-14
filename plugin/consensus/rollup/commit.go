package rollup

import (
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) buildBlockBatch(blocks []*types.Block) *rtypes.BlockBatch {

	batch := &rtypes.BlockBatch{}

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
	var alreadyCommitRound int64
	defer ticker.Stop()
	for {

		select {

		case <-r.ctx.Done():
			return
		case <-ticker.C:

			nextCommitRound, ok := r.val.isMyCommitTurn()
			if !ok || nextCommitRound <= alreadyCommitRound {
				continue
			}
			cp := r.cache.getPreparedCheckPoint(nextCommitRound, r.val.aggregateSign)
			// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
			if cp == nil {
				continue
			}

			// commit

			if err := r.commitCheckPoint(cp); err != nil {
				rlog.Error("commitCheckPoint err", "round", cp.GetCommitRound(), "err", err)
				continue
			}

			alreadyCommitRound = cp.GetCommitRound()
		}

	}

}

func (r *RollUp) commitCheckPoint(cp *rtypes.CheckPoint) error {

	tx, err := r.createTx(rtypes.RollupX, rtypes.NameCommitAction, types.Encode(cp))

	if err != nil {
		return errors.Wrap(err, "createTx")
	}

	tx.Fee, err = tx.GetRealFee(r.getProperFeeRate())

	if err != nil {
		return errors.Wrap(err, "setTxFee")
	}

	tx.Sign(types.EncodeSignID(secp256k1.ID, address.GetDefaultAddressID()), r.val.signTxKey)

	err = r.sendTx(tx)

	if err != nil {
		return errors.Wrap(err, "sendTx")
	}

	return nil
}

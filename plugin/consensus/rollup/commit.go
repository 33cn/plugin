package rollup

import (
	"bytes"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) buildCommitData(blocks []*types.Block) (*rtypes.BlockBatch, *pt.CommitRollupCrossTx) {

	batch := &rtypes.BlockBatch{}
	batch.BlockHeaders = make([]*types.Header, 0, len(blocks))
	batch.TxList = make([][]byte, 0, minCommitTxCount)
	batch.PubKeyList = make([][]byte, 0, minCommitTxCount)
	batch.TxAddrIDList = make([]byte, 0, minCommitTxCount)
	signs := make([]crypto.Signature, 0, minCommitTxCount)
	blsDriver := r.val.blsDriver

	crossInfo := &pt.CommitRollupCrossTx{}
	crossTxHashes := make([][]byte, 0, 8)
	for _, block := range blocks {

		header := block.GetHeader(r.chainCfg)
		header.Hash = nil
		batch.BlockHeaders = append(batch.BlockHeaders, header)
		for _, tx := range block.Txs {

			// 过滤跨链交易
			if types.IsParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
				crossTxHashes = append(crossTxHashes, tx.Hash())
			}

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
		rlog.Error("buildCommitData", "aggregate sign err", aggreSign)
		return nil, nil
	}
	batch.AggregateTxSign = aggreSign.Bytes()
	batch.CrossTxCheckHash = calcCrossTxCheckHash(crossTxHashes)
	crossInfo.TxIndices = r.cross.removePackedCrossTx(crossTxHashes)
	return batch, crossInfo
}

// 提交共识
func (r *RollUp) handleCommit() {

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
			commit := r.cache.getPreparedCommit(nextCommitRound, r.val.aggregateSign)
			// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
			if commit == nil {
				continue
			}

			// commit
			commitRound := commit.cp.GetCommitRound()
			if err := r.commit2MainChain(commit); err != nil {
				rlog.Error("handleCommit", "round", commitRound, "err", err)
				continue
			}

			alreadyCommitRound = commitRound
		}

	}

}

func (r *RollUp) commit2MainChain(info *commitInfo) error {

	tx1, err1 := r.createTx(rtypes.RollupX, rtypes.NameCommitAction, types.Encode(info.cp))
	tx2, err2 := r.createTx(pt.ParaX, pt.NameCommitCrossTxAction, types.Encode(info.crossTx))
	if err1 != nil || err2 != nil {
		rlog.Error("commit2MainChain", "err1", err1, "err2", err2)
		return errors.New("ErrCreateTx")
	}
	gtx, err := types.CreateTxGroup([]*types.Transaction{tx1, tx2}, r.getProperFeeRate())
	if err != nil {
		return errors.Wrapf(err, "createGroupTx")
	}

	for index := range gtx.GetTxs() {
		gtx.SignN(index, types.EncodeSignID(secp256k1.ID, address.GetDefaultAddressID()), r.val.signTxKey)
	}

	err = r.sendTx2MainChain(gtx.Tx())
	if err != nil {
		return errors.Wrap(err, "sendTx2MainChain")
	}

	return nil
}

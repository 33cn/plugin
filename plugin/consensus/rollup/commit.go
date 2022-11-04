package rollup

import (
	"bytes"
	"math/big"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

func (r *RollUp) buildCommitData(details []*types.BlockDetail) (*rtypes.BlockBatch, *pt.CommitRollup) {

	batch := &rtypes.BlockBatch{}
	batch.BlockHeaders = make([]*types.Header, 0, len(details))
	batch.TxList = make([][]byte, 0, minCommitTxCount)
	batch.PubKeyList = make([][]byte, 0, minCommitTxCount)
	batch.TxAddrIDList = make([]byte, 0, minCommitTxCount)
	signs := make([]crypto.Signature, 0, minCommitTxCount)
	blsDriver := r.val.blsDriver

	crossInfo := &pt.CommitRollup{}
	crossTxHashes := make([][]byte, 0, 8)
	crossTxRst := big.NewInt(0)
	for _, detail := range details {

		header := detail.Block.GetHeader(r.chainCfg)
		header.Hash = nil
		batch.BlockHeaders = append(batch.BlockHeaders, header)
		for i, tx := range detail.Block.Txs {

			// 过滤跨链交易
			if types.IsParaExecName(string(tx.Execer)) && bytes.HasSuffix(tx.Execer, []byte(pt.ParaX)) {
				if detail.Receipts[i].Ty == types.ExecOk {
					crossTxRst.SetBit(crossTxRst, len(crossTxHashes), 1)
				}
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
	if len(crossTxHashes) > 0 {
		batch.CrossTxResults = crossTxRst.Bytes()
		batch.CrossTxCheckHash = calcCrossTxCheckHash(crossTxHashes)
	}
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
				rlog.Error("handleCommit", "round", commitRound,
					"crossTx", len(commit.crossTx.TxIndices), "err", err)
				continue
			}

			alreadyCommitRound = commitRound
		}

	}

}

func (r *RollUp) commit2MainChain(info *commitInfo) error {

	tx, err := r.createTx(rtypes.RollupX, rtypes.NameCommitAction, types.Encode(info.cp))
	if err != nil {
		return errors.Wrapf(err, "createCommitCheckPointTx")
	}
	tx.Fee, _ = tx.GetRealFee(r.getProperFeeRate())
	tx.Sign(types.EncodeSignID(secp256k1.ID, address.GetDefaultAddressID()), r.val.signTxKey)
	// 提交跨链交易, 构建交易组
	if len(info.crossTx.TxIndices) > 0 {
		tx2, err := r.createTx(pt.ParaX, pt.NameCommitRollupAction, types.Encode(info.crossTx))
		if err != nil {
			return errors.New("ErrCreateCommitCrossTx")
		}
		gtx, err := types.CreateTxGroup([]*types.Transaction{tx, tx2}, r.getProperFeeRate())
		if err != nil {
			return errors.Wrapf(err, "createGroupTx")
		}

		for index := range gtx.GetTxs() {
			gtx.SignN(index, types.EncodeSignID(secp256k1.ID, address.GetDefaultAddressID()), r.val.signTxKey)
		}
		tx = gtx.Tx()
	}

	err = r.sendTx2MainChain(tx)
	if err != nil {
		return errors.Wrap(err, "sendTx2MainChain")
	}

	return nil
}

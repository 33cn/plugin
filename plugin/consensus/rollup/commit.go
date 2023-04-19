package rollup

import (
	"encoding/hex"
	"math/big"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

const (
	maxCommitDataSize = types.MaxTxSize * 3 / 4
)

func (r *RollUp) buildCommitData(details []*types.BlockDetail, commitRound int64,
	fragIndex *int32) ([]*rtypes.BlockBatch, []*pt.RollupCrossTx) {

	// 全量提交, 打包交易数据
	if r.cfg.FullDataCommit {
		return r.buildFullData(details, commitRound, fragIndex)
	}

	// 精简提交
	batch := &rtypes.BlockBatch{}
	batch.BlockHeaders = make([]*types.Header, 0, len(details))
	crossInfo := &pt.RollupCrossTx{}
	crossTxHashes := make([][]byte, 0, 8)
	crossTxRst := big.NewInt(0)

	for _, detail := range details {

		header := detail.Block.GetHeader(r.chainCfg)
		header.Hash = nil
		batch.BlockHeaders = append(batch.BlockHeaders, header)
		for i, tx := range detail.Block.Txs {

			// 过滤跨链交易
			if isCrossChainTx(tx) {
				txHash := tx.Hash()
				rlog.Debug("buildCommitData cross tx", "txHash", hex.EncodeToString(txHash))
				if detail.Receipts[i].Ty == types.ExecOk {
					crossTxRst.SetBit(crossTxRst, len(crossTxHashes), 1)
				}
				crossTxHashes = append(crossTxHashes, txHash)
			}
		}
	}

	if len(crossTxHashes) > 0 {
		batch.CrossTxResults = crossTxRst.Bytes()
		batch.CrossTxCheckHash = calcCrossTxCheckHash(crossTxHashes)
	}
	crossInfo.TxIndices = r.cross.removePackedCrossTx(crossTxHashes)

	return []*rtypes.BlockBatch{batch}, []*pt.RollupCrossTx{crossInfo}
}

func newCommitData(header *types.Header) (*rtypes.BlockBatch, *pt.RollupCrossTx) {

	batch := &rtypes.BlockBatch{}
	batch.BlockHeaders = make([]*types.Header, 0, 8)
	batch.TxList = make([][]byte, 0, 128)
	batch.PubKeyList = make([][]byte, 0, 128)
	batch.TxAddrIDList = make([]byte, 0, 128)

	if header != nil {
		batch.BlockHeaders = append(batch.BlockHeaders, header)
	}

	crossInfo := &pt.RollupCrossTx{}
	return batch, crossInfo
}

// 构造提交数据, 包括区块交易数据, 跨链交易信息数据
func (r *RollUp) buildFullData(details []*types.BlockDetail, commitRound int64,
	fragIndex *int32) ([]*rtypes.BlockBatch, []*pt.RollupCrossTx) {

	batchList := make([]*rtypes.BlockBatch, 0, 1)
	crossList := make([]*pt.RollupCrossTx, 0, 1)
	batch, crossInfo := newCommitData(nil)

	signs := make([]crypto.Signature, 0, 128)
	blsDriver := r.val.blsDriver
	crossTxHashes := make([][]byte, 0, 8)
	crossTxRst := big.NewInt(0)
	commitSize := 0
	aggreDriver := blsDriver.(crypto.AggregateCrypto)

	// 提交数据封装
	sealBlockBatch := func(fragIndex int) error {

		batch.BlockFragIndex = int32(fragIndex)
		aggreSign, err := aggreDriver.Aggregate(signs)
		if err != nil {
			rlog.Error("buildFullData", "round", commitRound, "aggregate sign err", err)
			return err
		}
		batch.AggregateTxSign = aggreSign.Bytes()
		if len(crossTxHashes) > 0 {
			batch.CrossTxResults = crossTxRst.Bytes()
			batch.CrossTxCheckHash = calcCrossTxCheckHash(crossTxHashes)
		}
		crossInfo.TxIndices = r.cross.removePackedCrossTx(crossTxHashes)
		batchList = append(batchList, batch)
		crossList = append(crossList, crossInfo)
		return nil
	}
	// 区块分割后接续, 从下标位置读取交易, 只有首次启动构建存在断点拼接情况
	if *fragIndex > 0 {
		details[0].Block.Txs = details[0].Block.Txs[*fragIndex:]
		*fragIndex = 0
	}

	for _, detail := range details {

		header := detail.Block.GetHeader(r.chainCfg)
		header.Hash = nil
		batch.BlockHeaders = append(batch.BlockHeaders, header)
		for i, tx := range detail.Block.Txs {

			ctx := types.CloneTx(tx)
			ctx.Signature = nil
			txData := types.Encode(ctx)

			// 超过最大容量, 区块分割
			if commitSize+len(txData) > maxCommitDataSize {

				if sealBlockBatch(i) != nil {
					return nil, nil
				}
				batch, crossInfo = newCommitData(header)
				crossTxHashes = crossTxHashes[:0]
				crossTxRst = big.NewInt(0)
				commitSize = 0
			}

			commitSize += len(txData)

			// 过滤跨链交易
			if isCrossChainTx(tx) {
				txHash := tx.Hash()
				rlog.Debug("buildFullData cross tx", "txHash", hex.EncodeToString(txHash))
				if detail.Receipts[i].Ty == types.ExecOk {
					crossTxRst.SetBit(crossTxRst, len(crossTxHashes), 1)
				}
				crossTxHashes = append(crossTxHashes, tx.Hash())
			}

			batch.PubKeyList = append(batch.PubKeyList, tx.Signature.Pubkey)
			// 本地已执行区块, 签名信息合法, 无需错误处理
			sign, _ := blsDriver.SignatureFromBytes(tx.Signature.GetSignature())
			signs = append(signs, sign)
			batch.TxAddrIDList = append(batch.TxAddrIDList, byte(types.ExtractAddressID(ctx.Signature.Ty)))
			batch.TxList = append(batch.TxList, txData)
		}
	}

	if sealBlockBatch(0) != nil {
		return nil, nil
	}
	return batchList, crossList
}

func (r *RollUp) handleBuildBatch() {

	fragIndex := r.initFragIndex
	for {

		select {
		case <-r.ctx.Done():
			return
		default:
		}
		blockDetails, prepared := r.getNextBatchBlocks(r.nextBuildHeight)
		// 区块内未达到最低批量数量, 需要继续等待
		if !prepared || len(blockDetails) == 0 {
			rlog.Debug("handleBuildBatch", "height", r.nextBuildHeight,
				"round", r.nextBuildRound, "msg", "wait more local block")
			time.Sleep(time.Second * 3)
			continue
		}

		rlog.Debug("handleBuildBatch commit height", "nextBuildRound", r.nextBuildRound,
			"start", blockDetails[0].GetBlock().GetHeight(),
			"end", blockDetails[len(blockDetails)-1].GetBlock().GetHeight())

		batchList, crossList := r.buildCommitData(blockDetails, r.nextBuildRound, &fragIndex)

		for i, blkBatch := range batchList {
			crossInfo := crossList[i]

			cp := &rtypes.CheckPoint{
				ChainTitle:          r.chainCfg.GetTitle(),
				CommitRound:         r.nextBuildRound,
				Batch:               blkBatch,
				CrossTxSyncedHeight: r.cross.refreshSyncedHeight(),
			}
			crossInfo.ChainTitle = r.chainCfg.GetTitle()
			crossInfo.CommitRound = r.nextBuildRound
			commit := &commitInfo{
				cp:      cp,
				crossTx: crossInfo,
			}

			r.nextBuildRound++
			sign := r.val.sign(cp.GetCommitRound(), cp.GetBatch())

			r.cache.addCommitInfo(commit)
			r.cache.addValidatorSign(true, sign)
			r.tryPubMsg(psValidatorSignTopic, types.Encode(sign), sign.CommitRound)

		}

		r.nextBuildHeight += int64(len(blockDetails))

	}
}

// 提交共识
func (r *RollUp) handleCommit() {

	var alreadyCommitRound int64
	for {

		select {
		case <-r.ctx.Done():
			return
		default:
		}

		nextCommitRound, ok := r.val.isMyCommitTurn(r.cfg.MaxCommitInterval)
		if !ok || nextCommitRound <= alreadyCommitRound {
			rlog.Debug("handleCommit not ok", "round", nextCommitRound, "alreadyCommit", alreadyCommitRound)
			time.Sleep(2 * time.Second)
			continue
		}
		commit := r.cache.getPreparedCommit(nextCommitRound, r.val.aggregateSign)
		// cache中不存在或 验证者签名数量未达到要求, 需要继续等待
		if commit == nil {
			rlog.Debug("handleCommit not ready", "round", nextCommitRound)
			time.Sleep(2 * time.Second)
			continue
		}

		// commit
		commitRound := commit.cp.GetCommitRound()
		if err := r.commit2MainChain(commit); err != nil {
			rlog.Error("handleCommit", "round", commitRound,
				"crossTx", len(commit.crossTx.TxIndices), "err", err)
			time.Sleep(time.Second * 2)
			continue
		}

		alreadyCommitRound = commitRound
	}
}

func (r *RollUp) commit2MainChain(info *commitInfo) error {

	tx, err := r.createTx(rtypes.RollupX, rtypes.NameCommitAction, types.Encode(info.cp))
	if err != nil {
		return errors.Wrapf(err, "createCommitCheckPointTx")
	}
	tx.Fee, _ = tx.GetRealFee(r.getProperFeeRate())
	tx.Sign(types.EncodeSignID(secp256k1.ID, r.cfg.AddressID), r.val.signTxKey)
	// 提交跨链交易, 构建交易组
	if len(info.crossTx.TxIndices) > 0 {
		tx2, err := r.createTx(pt.ParaX, pt.NameRollupCrossTxAction, types.Encode(info.crossTx))
		if err != nil {
			return errors.New("ErrCreateCommitCrossTx")
		}
		gtx, err := types.CreateTxGroup([]*types.Transaction{tx, tx2}, r.getProperFeeRate())
		if err != nil {
			return errors.Wrapf(err, "createGroupTx")
		}

		for index := range gtx.GetTxs() {
			gtx.SignN(index, types.EncodeSignID(secp256k1.ID, r.cfg.AddressID), r.val.signTxKey)
		}
		tx = gtx.Tx()
	}
	rlog.Debug("commit2MainChain", "round", info.cp.GetCommitRound(),
		"crossLen", len(info.crossTx.GetTxIndices()), "txHash", hex.EncodeToString(tx.Hash()))
	err = r.sendTx2MainChain(tx)
	if err != nil {
		return errors.Wrap(err, "sendTx2MainChain")
	}

	return nil
}

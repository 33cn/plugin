package rollup

import (
	"encoding/hex"
	"strings"
	"time"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"

	"github.com/33cn/plugin/plugin/dapp/paracross/executor"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

// 本地获取批量提交数据, 需要确保数据一致性
func (r *RollUp) getNextBatchBlocks(startHeight int64) ([]*types.BlockDetail, bool) {

	req := &types.ReqBlocks{
		Start:    startHeight,
		End:      startHeight + minCommitTxCount,
		IsDetail: true,
	}
	header, err := r.base.GetAPI().GetLastHeader()
	if err != nil || header.GetHeight() < startHeight {
		rlog.Debug("getNextBatchBlocks", "startHeight", startHeight, "err", err)
		return nil, false
	}
	details, err := r.base.GetAPI().GetBlocks(req)
	if err != nil || len(details.GetItems()) == 0 {
		rlog.Error("getNextBatchBlocks", "req", req.String(), "err", err)
		return nil, false
	}

	blkDetails := details.GetItems()
	batchPrepared := false
	// 全量数据提交模式, 以交易为单位, 满足最低提交数量原则
	if r.cfg.FullDataCommit {

		txCount := 0
		for i, blk := range blkDetails {
			txCount += len(blk.GetBlock().GetTxs())
			if txCount >= minCommitTxCount {
				blkDetails = blkDetails[:i+1]
				batchPrepared = true
				break
			}
		}
	} else {
		// 精简提交模式, 只提交区块头数据, 以区块为单位, 满足最低提交数量原则
		if len(blkDetails) >= minCommitBlkCount {
			batchPrepared = true
			blkDetails = blkDetails[:minCommitBlkCount]
		}
	}
	firstBlockTime := blkDetails[0].GetBlock().GetBlockTime()
	// 本地不产生区块时触发, 增加10s延迟判定, 避免临界情况导致判定不一致
	if !batchPrepared && types.Now().Unix()-firstBlockTime > r.cfg.MaxCommitInterval+10 {
		batchPrepared = true
	}

	// 提交间隔共识, 单次提交首尾区块时间间隔限制最大值
	for i := 1; batchPrepared && i < len(blkDetails); i++ {
		if blkDetails[i].GetBlock().GetBlockTime()-firstBlockTime > r.cfg.MaxCommitInterval {
			blkDetails = blkDetails[:i]
			break
		}
	}

	return blkDetails, batchPrepared
}

func (r *RollUp) sendP2PMsg(ty int64, data interface{}) error {
	msg := r.client.NewMessage("p2p", ty, data)
	err := r.client.Send(msg, true)
	if err != nil {
		return errors.Wrapf(err, "ty=%d", ty)
	}
	resp, err := r.client.WaitTimeout(msg, time.Second*5)
	if err != nil {
		return errors.Wrapf(err, "wait ty=%d", ty)
	}
	reply, ok := resp.GetData().(*types.Reply)
	if !ok {
		return types.ErrTypeAsset
	}
	if !reply.GetIsOk() {
		return errors.New(string(reply.GetMsg()))
	}
	return nil
}

func (r *RollUp) isChainSync() bool {

	reply, err := r.base.GetAPI().IsSync()

	if err != nil {
		rlog.Error("isChainSync", "err", err)
		return false
	}

	return reply.GetIsOk()
}

func shortHash(hash []byte) string {
	return types.CalcTxShortHash(hash)
}

func filterParaTx(cfg *types.Chain33Config, detail *types.ParaTxDetail) []*types.Transaction {
	return executor.FilterTxsForPara(cfg, detail)
}

// 检测是否跨链交易
func isCrossChainTx(tx *types.Transaction) bool {

	execer := string(tx.GetExecer())
	if strings.HasSuffix(execer, pt.ParaX) && types.IsParaExecName(execer) {

		var payload pt.ParacrossAction
		err := types.Decode(tx.Payload, &payload)
		if err != nil {
			rlog.Error("isCrossChainTx decode tx payload", "txhash", hex.EncodeToString(tx.Hash()), "err", err.Error())
			return false
		}
		if payload.Ty == pt.ParacrossActionCrossAssetTransfer {
			return true
		}
	}

	return false
}

// 检测为包含跨链交易的交易组
func isCrossChainGroupTx(txs ...*types.Transaction) bool {

	for _, tx := range txs {
		if isCrossChainTx(tx) {
			return true
		}
	}
	return false
}

// 平行链交易过滤出跨链交易, 解析交易组情况
func filterParaCrossTx(paraTxs []*types.Transaction) []*types.Transaction {

	if len(paraTxs) <= 0 {
		return nil
	}
	crossTxs := make([]*types.Transaction, 0, len(paraTxs))
	for i := 0; i < len(paraTxs); i++ {

		groupCount := int(paraTxs[i].GetGroupCount())
		// 交易组情况, 如果其中包含跨链交易,将交易组所有交易添加
		if groupCount > 1 && groupCount+i <= len(paraTxs) {

			if isCrossChainGroupTx(paraTxs[i : groupCount+i]...) {
				rlog.Info("filterParaCrossTx", "groupCount", groupCount, "group tx hash", hex.EncodeToString(paraTxs[i].Hash()))
				crossTxs = append(crossTxs, paraTxs[i:groupCount+i]...)
			}
			i += groupCount - 1
		} else if isCrossChainTx(paraTxs[i]) {
			rlog.Info("filterParaCrossTx", "cross tx hash", hex.EncodeToString(paraTxs[i].Hash()))
			crossTxs = append(crossTxs, paraTxs[i])
		}
	}

	return crossTxs

}

func calcCrossTxCheckHash(hashList [][]byte) []byte {

	return executor.CalcTxHashsHash(hashList)
}

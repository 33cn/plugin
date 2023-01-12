package rollup

import (
	"time"

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

	// 满足最大提交间隔原则
	firstBlockTime := blkDetails[0].GetBlock().GetBlockTime()
	for i := 1; batchPrepared && i < len(blkDetails); i++ {
		if blkDetails[i].GetBlock().GetBlockTime()-firstBlockTime > r.cfg.MaxCommitInterval {
			blkDetails = blkDetails[:i]
			break
		}
	}
	// 本地不产生区块时触发, 增加10s延迟判定, 避免临界情况导致判定不一致
	if !batchPrepared && types.Now().Unix()-firstBlockTime > r.cfg.MaxCommitInterval+10 {
		batchPrepared = true
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

func shortHash(hash []byte) string {
	return types.CalcTxShortHash(hash)
}

func filterParaTx(cfg *types.Chain33Config, detail *types.ParaTxDetail) []*types.Transaction {
	return executor.FilterTxsForPara(cfg, detail)
}

func filterParaCrossTx(txs []*types.Transaction) []*types.Transaction {
	return executor.FilterParaCrossTxs(txs)
}

func calcCrossTxCheckHash(hashList [][]byte) []byte {

	return executor.CalcTxHashsHash(hashList)
}

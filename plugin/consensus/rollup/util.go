package rollup

import (
	"time"

	"github.com/33cn/plugin/plugin/dapp/paracross/executor"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

func (r *RollUp) getNextBatchBlocks(startHeight int64) []*types.BlockDetail {

	req := &types.ReqBlocks{
		Start: startHeight,
		End:   startHeight + minCommitCount - 1,
	}

	details, err := r.base.GetAPI().GetBlocks(req)
	for err != nil {
		rlog.Error("getNextBatchBlocks", "req", req.String(), "err", err)
		return nil
	}

	// 全量数据提交, 以交易为单位, 需满足最低提交数量
	if r.cfg.FullDataCommit {
		txCount := 0
		for i, detail := range details.GetItems() {
			txCount += len(detail.GetBlock().GetTxs())
			if txCount >= minCommitCount {
				return details.GetItems()[:i]
			}
		}
	} else {
		// 精简模式, 只提交区块头数据, 以区块为单位, 需满足最低提交数量
		if len(details.GetItems()) == minCommitCount {
			return details.GetItems()
		}
	}
	// 满足最大提交间隔, 触发提交
	if len(details.GetItems()) > 0 &&
		types.Now().Unix()-details.GetItems()[0].Block.BlockTime >= r.cfg.MaxCommitInterval {
		return details.GetItems()
	}

	// 未达到提交阈值
	return nil
}

func (r *RollUp) sendP2PMsg(ty int64, data interface{}) error {
	msg := r.base.GetQueueClient().NewMessage("p2p", ty, data)
	err := r.base.GetQueueClient().Send(msg, true)
	if err != nil {
		return errors.Wrapf(err, "ty=%d", ty)
	}
	resp, err := r.base.GetQueueClient().WaitTimeout(msg, time.Second*5)
	if err != nil {
		return errors.Wrapf(err, "wait ty=%d", ty)
	}

	if resp.GetData().(*types.Reply).IsOk {
		return nil
	}
	return errors.New(string(resp.GetData().(*types.Reply).GetMsg()))
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

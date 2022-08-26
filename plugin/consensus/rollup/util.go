package rollup

import (
	"time"

	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
)

func (r *RollUp) getNextBatchBlocks(startHeight int64) []*types.Block {

	req := &types.ReqBlocks{
		Start: startHeight,
		End:   startHeight + minCommitTxCount,
	}

	details, err := r.base.GetAPI().GetBlocks(req)
	for err != nil {
		rlog.Error("getNextBatchBlocks", "req", req.String(), "err", err)
		return nil
	}

	blocks := make([]*types.Block, 0, minCommitTxCount)
	txCount := 0
	for _, detail := range details.GetItems() {
		blocks = append(blocks, detail.GetBlock())
		txCount += len(detail.GetBlock().GetTxs())
		if txCount >= minCommitTxCount {
			break
		}
	}

	if txCount < minCommitTxCount {
		return nil
	}

	return blocks
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

func (r *RollUp) buildCommitTx() {

}

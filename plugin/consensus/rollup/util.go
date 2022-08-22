package rollup

import (
	"github.com/33cn/chain33/types"
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

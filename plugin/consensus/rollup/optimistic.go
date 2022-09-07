package rollup

import (
	"github.com/33cn/chain33/types"
	rolluptypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

type optimistic struct {
}

func (op *optimistic) GetCommitBatch(blocks []*types.Block) *rolluptypes.CommitBatch {

	batch := &rolluptypes.CommitBatch{}
   	blocks[0].Hash()
	types.Headers{}
	return batch
}

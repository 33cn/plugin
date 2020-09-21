package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func (w *Wasm) Query_Check(query *types2.QueryCheckContract) (types.Message, error) {
	if query == nil {
		return nil, types.ErrInvalidParam
	}
	return &types.Reply{IsOk: w.contractExist(query.Name)}, nil
}

package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func (w *Wasm) Query_Check(query *types2.QueryCheckContract) (types.Message, error) {
	if query == nil {
		return nil, types.ErrInvalidParam
	}
	_, err := w.GetStateDB().Get(contractKey(query.Name))
	if err == nil {
		return &types.Reply{IsOk: true}, nil
	}
	if err == types.ErrNotFound {
		return &types.Reply{}, nil
	}
	return nil, err
}

func (w *Wasm) Query_QueryStateDB(query *types2.QueryContractDB) (types.Message, error) {
	if query == nil {
		return nil, types.ErrInvalidParam
	}
	key := append(calcStatePrefix(query.Contract), query.Key...)
	v, err := w.GetStateDB().Get(key)
	if err != nil {
		return nil, err
	}
	return &types.ReplyString{Data: string(v)}, nil
}

func (w *Wasm) Query_QueryLocalDB(query *types2.QueryContractDB) (types.Message, error) {
	if query == nil {
		return nil, types.ErrInvalidParam
	}
	key := append(calcLocalPrefix(query.Contract), query.Key...)
	v, err := w.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}
	return &types.ReplyString{Data: string(v)}, nil
}

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

//func (w *Wasm) Query_CreateTransaction(query *types2.QueryCreateTransaction) (types.Message, error) {
//	if query == nil {
//		return nil, types.ErrInvalidParam
//	}
//	if !validateName(query.Name) {
//		return nil, types2.ErrInvalidContractName
//	}
//
//	var action *types2.WasmAction
//	switch query.Ty {
//	case types2.WasmActionCreate:
//		action = &types2.WasmAction{
//			Value: &types2.WasmAction_Create{
//				Create: &types2.WasmCreate{
//					Name: query.Name,
//					Code: query.Code,
//				},
//			},
//		}
//	case types2.WasmActionCall:
//		action = &types2.WasmAction{
//			Value: &types2.WasmAction_Call{
//				Call: &types2.WasmCall{
//					Contract:   query.Name,
//					Method:     query.Method,
//					Parameters: query.Parameters,
//				},
//			},
//		}
//	}
//	if action == nil {
//		return nil, types.ErrInvalidParam
//	}
//	cfg := w.GetAPI().GetConfig()
//	tx := &types.Transaction{
//		Payload: types.Encode(action),
//	}
//	tx, err := types.FormatTx(w.GetAPI().GetConfig(), cfg.ExecName(types2.WasmX), tx)
//	if err != nil {
//		return nil, err
//	}
//	if tx.Fee < query.Fee {
//		tx.Fee = query.Fee
//	}
//	resp := types.ReplyString{
//		Data: common.ToHex(types.Encode(tx)),
//	}
//	return &resp, nil
//}

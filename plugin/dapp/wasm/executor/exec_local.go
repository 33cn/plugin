package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func (w *Wasm) ExecLocal_Create(payload *types2.WasmCreate, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (w *Wasm) ExecLocal_Update(payload *types2.WasmUpdate, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (w *Wasm) ExecLocal_Call(payload *types2.WasmCall, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	if receipt.Ty != types.ExecOk {
		return &types.LocalDBSet{}, nil
	}
	localExecer := w.userExecName(payload.Contract, true)
	var KVs []*types.KeyValue
	for _, item := range receipt.Logs {
		if item.Ty == types2.TyLogLocalData {
			var data types2.LocalDataLog
			err := types.Decode(item.Log, &data)
			if err != nil {
				return nil, err
			}
			KVs = append(KVs, &types.KeyValue{
				Key:   data.Key,
				Value: data.Value,
			})
		}
	}

	return &types.LocalDBSet{KV: w.AddRollbackKV(tx, []byte(localExecer), KVs)}, nil
}

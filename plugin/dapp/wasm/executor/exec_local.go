package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func (w *Wasm) ExecLocal_Create(payload *types2.WasmCreate, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return execLocal(receipt)
}

func (w *Wasm) ExecLocal_Call(payload *types2.WasmCall, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return execLocal(receipt)
}

func execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	if receipt.Ty != types.ExecOk {
		return nil, nil
	}
	set := &types.LocalDBSet{}
	for _, item := range receipt.Logs {
		if item.Ty == types2.TyLogLocalData {
			var data types2.LocalDataLog
			err := types.Decode(item.Log, &data)
			if err != nil {
				log.Error("execLocal", "decode error", err)
				continue
			}
			set.KV = append(set.KV, &types.KeyValue{Key: data.Key, Value: data.CurValue})
		}
	}
	return set, nil
}

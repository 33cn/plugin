package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/perlin-network/life/compiler"
	"github.com/perlin-network/life/exec"
	validation "github.com/perlin-network/life/wasm-validation"
)

var wasmCB *Wasm

func (w *Wasm) Exec_Create(payload *types2.WasmCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload == nil {
		return nil, types.ErrInvalidParam
	}

	name := payload.Name
	if !validateName(name) {
		return nil, types2.ErrInvalidContractName
	}
	if w.contractExist(name) {
		return nil, types2.ErrContractExist
	}
	code := payload.Code
	if len(code) > 1<<20 { //TODO: max size to define
		return nil, types2.ErrCodeOversize
	}
	if err := validation.ValidateWasm(code); err != nil {
		return nil, types2.ErrInvalidWasm
	}
	var receiptLogs []*types.ReceiptLog
	kv := []*types.KeyValue{{
		Key:   contractKey(name),
		Value: code,
	}}

	receiptLog := &types.ReceiptLog{
		Ty: types2.TyLogWasmCreate,
		Log: types.Encode(&types2.CreateContractLog{
			Name: name,
			Code: hex.EncodeToString(code),
		}),
	}
	receiptLogs = append(receiptLogs, receiptLog)

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: receiptLogs,
	}, nil
}

func (w *Wasm) Exec_Call(payload *types2.WasmCall, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload == nil {
		return nil, types.ErrInvalidParam
	}
	code, err := w.getContract(payload.Contract)
	if err != nil {
		return nil, err
	}
	vm, err := exec.NewVirtualMachine(code, exec.VMConfig{
		DefaultMemoryPages:   128,
		DefaultTableSize:     128,
		DisableFloatingPoint: true,
		GasLimit:             uint64(tx.Fee),
	}, new(Resolver), &compiler.SimpleGasPolicy{GasPerInstruction: 1})
	if err != nil {
		return nil, err
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(payload.Method)
	if !ok {
		return nil, types2.ErrInvalidMethod
	}

	w.contractAddr = address.ExecAddress(payload.Contract)
	w.tx = tx
	w.execAddr = address.ExecAddress(string(tx.Execer))
	wasmCB = w
	defer func() {
		wasmCB = nil
	}()
	// Run the WebAssembly module's entry function.
	ret, err := vm.RunWithGasLimit(entryID, int(tx.Fee), payload.Parameters...)
	if err != nil {
		return nil, err
	}
	var logs []*types.ReceiptLog
	logs = append(logs, &types.ReceiptLog{Ty: types2.TyLogWasmCall, Log: types.Encode(&types2.CallContractLog{
		Contract: payload.Contract,
		Method:   payload.Method,
		Result:   ret,
	})})

	logs = append(logs, &types.ReceiptLog{Ty: types2.TyLogCustom, Log: types.Encode(&types2.CustomLog{
		Info: w.logs,
	})})

	logs = append(logs, w.receiptLogs...)

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   w.kvs,
		Logs: logs,
	}, nil
}

func validateName(name string) bool {
	if !types2.NameReg.MatchString(name) || len(name) < 4 || len(name) > 20 {
		return false
	}
	return true
}

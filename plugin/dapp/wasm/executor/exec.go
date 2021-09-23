package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/perlin-network/life/compiler"
	"github.com/perlin-network/life/exec"
	validation "github.com/perlin-network/life/wasm-validation"
)

var wasmCB *Wasm

func (w *Wasm) userExecName(name string, local bool) string {
	execer := "user." + types2.WasmX + "." + name
	if local {
		cfg := w.GetAPI().GetConfig()
		execer = cfg.ExecName(execer)
	}
	return execer
}

func (w *Wasm) checkTxExec(txExec string, execName string) bool {
	cfg := w.GetAPI().GetConfig()
	return txExec == cfg.ExecName(execName)
}

func (w *Wasm) Exec_Create(payload *types2.WasmCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload == nil {
		return nil, types.ErrInvalidParam
	}
	if !w.checkTxExec(string(tx.Execer), types2.WasmX) {
		return nil, types.ErrExecNameNotMatch
	}

	name := payload.Name
	if !validateName(name) {
		return nil, types2.ErrInvalidContractName
	}
	code := payload.Code
	if len(code) > types2.MaxCodeSize {
		return nil, types2.ErrCodeOversize
	}
	if err := validation.ValidateWasm(code); err != nil {
		return nil, types2.ErrInvalidWasm
	}

	kvc := dapp.NewKVCreator(w.GetStateDB(), types.CalcStatePrefix(tx.Execer), nil)
	_, err := kvc.GetNoPrefix(contractKey(name))
	if err == nil {
		return nil, types2.ErrContractExist
	}
	if err != types.ErrNotFound {
		return nil, err
	}
	kvc.AddNoPrefix(contractKey(name), code)
	kvc.AddNoPrefix(contractCreatorKey(name), []byte(tx.From()))

	receiptLog := &types.ReceiptLog{
		Ty: types2.TyLogWasmCreate,
		Log: types.Encode(&types2.CreateContractLog{
			Name: name,
			Code: hex.EncodeToString(code),
		}),
	}

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kvc.KVList(),
		Logs: []*types.ReceiptLog{receiptLog},
	}, nil
}

func (w *Wasm) Exec_Update(payload *types2.WasmUpdate, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload == nil {
		return nil, types.ErrInvalidParam
	}
	if !w.checkTxExec(string(tx.Execer), types2.WasmX) {
		return nil, types.ErrExecNameNotMatch
	}

	name := payload.Name
	kvc := dapp.NewKVCreator(w.GetStateDB(), types.CalcStatePrefix(tx.Execer), nil)
	creator, err := kvc.GetNoPrefix(contractCreatorKey(name))
	if err != nil {
		return nil, types2.ErrContractNotExist
	}
	_, err = kvc.GetNoPrefix(contractKey(name))
	if err != nil {
		return nil, types2.ErrContractNotExist
	}
	if tx.From() != string(creator) {
		return nil, types2.ErrInvalidCreator
	}

	code := payload.Code
	if len(code) > types2.MaxCodeSize {
		return nil, types2.ErrCodeOversize
	}
	if err := validation.ValidateWasm(code); err != nil {
		return nil, types2.ErrInvalidWasm
	}

	kvc.AddNoPrefix(contractKey(name), code)

	// 删除旧合约缓存
	delete(w.VMCache, name)

	receiptLog := &types.ReceiptLog{
		Ty: types2.TyLogWasmUpdate,
		Log: types.Encode(&types2.UpdateContractLog{
			Name: name,
			Code: hex.EncodeToString(code),
		}),
	}

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kvc.KVList(),
		Logs: []*types.ReceiptLog{receiptLog},
	}, nil
}

func (w *Wasm) Exec_Call(payload *types2.WasmCall, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload == nil {
		return nil, types.ErrInvalidParam
	}
	if !w.checkTxExec(string(tx.Execer), types2.WasmX) {
		return nil, types.ErrExecNameNotMatch
	}

	w.stateKVC = dapp.NewKVCreator(w.GetStateDB(), calcStatePrefix(payload.Contract), nil)
	var vm *exec.VirtualMachine
	var ok bool
	if vm, ok = w.VMCache[payload.Contract]; !ok {
		code, err := w.stateKVC.GetNoPrefix(contractKey(payload.Contract))
		if err != nil {
			return nil, err
		}
		vm, err = exec.NewVirtualMachine(code, exec.VMConfig{
			DefaultMemoryPages:   128,
			DefaultTableSize:     128,
			DisableFloatingPoint: true,
			GasLimit:             uint64(tx.Fee),
		}, new(Resolver), &compiler.SimpleGasPolicy{GasPerInstruction: 1})
		if err != nil {
			return nil, err
		}
		w.VMCache[payload.Contract] = vm
	} else {
		vm.Config.GasLimit = uint64(tx.Fee)
		vm.Gas = 0
	}

	// Get the function ID of the entry function to be executed.
	entryID, ok := vm.GetFunctionExport(payload.Method)
	if !ok {
		return nil, types2.ErrInvalidMethod
	}

	w.contractName = payload.Contract
	w.tx = tx
	w.execAddr = address.ExecAddress(string(types.GetRealExecName(tx.Execer)))
	w.ENV = make(map[int]string)
	w.localCache = nil
	w.kvs = nil
	w.receiptLogs = nil
	w.customLogs = nil
	for i, v := range payload.Env {
		w.ENV[i] = v
	}
	wasmCB = w
	defer func() {
		wasmCB = nil
	}()
	// Run the WebAssembly module's entry function.
	ret, err := vm.RunWithGasLimit(entryID, int(tx.Fee), payload.Parameters...)
	if err != nil {
		return nil, err
	}
	var kvs []*types.KeyValue
	kvs = append(kvs, w.kvs...)
	kvs = append(kvs, w.stateKVC.KVList()...)

	var logs []*types.ReceiptLog
	logs = append(logs, &types.ReceiptLog{Ty: types2.TyLogWasmCall, Log: types.Encode(&types2.CallContractLog{
		Contract: payload.Contract,
		Method:   payload.Method,
		Result:   int32(ret),
	})})
	logs = append(logs, w.receiptLogs...)
	logs = append(logs, &types.ReceiptLog{Ty: types2.TyLogCustom, Log: types.Encode(&types2.CustomLog{
		Info: w.customLogs,
	})})
	for _, log := range w.localCache {
		logs = append(logs, &types.ReceiptLog{
			Ty:  types2.TyLogLocalData,
			Log: types.Encode(log),
		})
	}

	receipt := &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kvs,
		Logs: logs,
	}
	if int32(ret) < 0 || int16(ret) < 0 {
		receipt.Ty = types.ExecPack
	}

	return receipt, nil
}

func validateName(name string) bool {
	if !types2.NameReg.MatchString(name) || len(name) < 4 || len(name) > 20 {
		return false
	}
	return true
}

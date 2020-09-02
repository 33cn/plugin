package executor

import (
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func setStateDB(key, value []byte) {
	wasmCB.kvs = append(wasmCB.kvs, &types.KeyValue{
		Key:   wasmCB.formatStateKey(key),
		Value: value,
	})
}

func getStateDBSize(key []byte) int {
	value, err := getStateDB(key)
	if err != nil {
		return 0
	}
	return len(value)
}

func getStateDB(key []byte) ([]byte, error) {
	return wasmCB.GetStateDB().Get(wasmCB.formatStateKey(key))
}

func setLocalDB(key, value []byte) {
	preValue, _ := getLocalDB(key)
	wasmCB.receiptLogs = append(wasmCB.receiptLogs, &types.ReceiptLog{
		Ty: types2.TyLogLocalData,
		Log: types.Encode(&types2.LocalDataLog{
			Key:      wasmCB.formatLocalKey(key),
			PreValue: preValue,
			CurValue: value,
		}),
	})
}

func getLocalDBSize(key []byte) int {
	value, err := getLocalDB(key)
	if err != nil {
		return 0
	}
	return len(value)
}

func getLocalDB(key []byte) ([]byte, error) {
	return wasmCB.GetLocalDB().Get(wasmCB.formatLocalKey(key))
}

func execFrozen(addr string, amount int64) error {
	receipt, err := wasmCB.GetCoinsAccount().ExecFrozen(addr, wasmCB.execAddr, amount)
	if err != nil {
		return err
	}
	wasmCB.kvs = append(wasmCB.kvs, receipt.KV...)
	wasmCB.receiptLogs = append(wasmCB.receiptLogs, receipt.Logs...)
	return nil
}

func execActive(addr string, amount int64) error {
	receipt, err := wasmCB.GetCoinsAccount().ExecActive(addr, wasmCB.execAddr, amount)
	if err != nil {
		return err
	}
	wasmCB.kvs = append(wasmCB.kvs, receipt.KV...)
	wasmCB.receiptLogs = append(wasmCB.receiptLogs, receipt.Logs...)
	return nil
}

func execTransfer(from, to string, amount int64) error {
	receipt, err := wasmCB.GetCoinsAccount().ExecTransfer(from, to, wasmCB.execAddr, amount)
	if err != nil {
		return err
	}
	wasmCB.kvs = append(wasmCB.kvs, receipt.KV...)
	wasmCB.receiptLogs = append(wasmCB.receiptLogs, receipt.Logs...)
	return nil
}

func execTransferFrozen(from, to string, amount int64) error {
	receipt, err := wasmCB.GetCoinsAccount().ExecTransferFrozen(from, to, wasmCB.execAddr, amount)
	if err != nil {
		return err
	}
	wasmCB.kvs = append(wasmCB.kvs, receipt.KV...)
	wasmCB.receiptLogs = append(wasmCB.receiptLogs, receipt.Logs...)
	return nil
}

func getFrom() string {
	return wasmCB.tx.From()
}

func getHeight() int64 {
	return wasmCB.GetHeight()
}

func getRandom() int64 {
	req := &types.ReqRandHash{
		ExecName: "ticket",
		BlockNum: 5,
		Hash:     wasmCB.GetLastHash(),
	}
	hash, err := wasmCB.GetExecutorAPI().GetRandNum(req)
	if err != nil {
		return -1
	}
	var rand int64
	for _, c := range hash {
		rand = rand*256 + int64(c)
	}
	if rand < 0 {
		return -rand
	}
	return rand
}

func printlog(s string) {
	wasmCB.logs = append(wasmCB.logs, s)
}

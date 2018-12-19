package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
)

// Exec_Ping 执行 ping 类型交易
func (h *Echo) Exec_Ping(ping *echotypes.Ping, tx *types.Transaction, index int) (*types.Receipt, error) {
	msg := ping.Msg
	res := fmt.Sprintf("%s, ping ping ping!", msg)
	xx := &echotypes.PingLog{Msg: msg, Echo: res}
	logs := []*types.ReceiptLog{{Ty: echotypes.TyLogPing, Log: types.Encode(xx)}}
	kv := []*types.KeyValue{{Key: []byte(fmt.Sprintf(KeyPrefixPing, msg)), Value: []byte(res)}}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// Exec_Pang 执行 pang 类型交易
func (h *Echo) Exec_Pang(ping *echotypes.Pang, tx *types.Transaction, index int) (*types.Receipt, error) {
	msg := ping.Msg
	res := fmt.Sprintf("%s, pang pang pang!", msg)
	xx := &echotypes.PangLog{Msg: msg, Echo: res}
	logs := []*types.ReceiptLog{{Ty: echotypes.TyLogPang, Log: types.Encode(xx)}}
	kv := []*types.KeyValue{{Key: []byte(fmt.Sprintf(KeyPrefixPang, msg)), Value: []byte(res)}}
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

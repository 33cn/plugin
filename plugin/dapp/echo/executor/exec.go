package executor
import (
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types"
	"github.com/33cn/chain33/types"
	"fmt"
)
func (h *Echo) Exec_Ping(ping *echotypes.Ping, tx *types.Transaction, index int) (*types.Receipt, error) {
	msg := ping.Msg
	res := fmt.Sprintf("%s, ping ping ping!", msg)
	xx := &echotypes.PingLog{Msg:msg, Echo:res}
	logs := []*types.ReceiptLog{{echotypes.TyLogPing, types.Encode(xx)}}
	kv := []*types.KeyValue{{[]byte(fmt.Sprintf(KeyPrefixPing, msg)), []byte(res)}}
	receipt := &types.Receipt{types.ExecOk, kv, logs}
	return receipt, nil
}
func (h *Echo) Exec_Pang(ping *echotypes.Pang, tx *types.Transaction, index int) (*types.Receipt, error) {
	msg := ping.Msg
	res := fmt.Sprintf("%s, pang pang pang!", msg)
	xx := &echotypes.PangLog{Msg:msg, Echo:res}
	logs := []*types.ReceiptLog{{echotypes.TyLogPang, types.Encode(xx)}}
	kv := []*types.KeyValue{{[]byte(fmt.Sprintf(KeyPrefixPang, msg)), []byte(res)}}
	receipt := &types.Receipt{types.ExecOk, kv, logs}
	return receipt, nil
}
package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
)

// ExecDelLocal_Ping 交易执行成功，将本消息对应的数值减1
func (h *Echo) ExecDelLocal_Ping(ping *echotypes.Ping, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	// 这里简化处理，不做基本的零值及错误检查了
	var pingLog echotypes.PingLog
	types.Decode(receipt.Logs[0].Log, &pingLog)
	localKey := []byte(fmt.Sprintf(KeyPrefixPingLocal, pingLog.Msg))
	oldValue, err := h.GetLocalDB().Get(localKey)
	if err != nil {
		return nil, err
	}
	types.Decode(oldValue, &pingLog)
	if pingLog.Count > 0 {
		pingLog.Count--
	}
	val := types.Encode(&pingLog)
	if pingLog.Count == 0 {
		val = nil
	}
	kv := []*types.KeyValue{{Key: localKey, Value: val}}
	return &types.LocalDBSet{KV: kv}, nil
}

// ExecDelLocal_Pang 交易执行成功，将本消息对应的数值减1
func (h *Echo) ExecDelLocal_Pang(ping *echotypes.Pang, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	// 这里简化处理，不做基本的零值及错误检查了
	var pangLog echotypes.PangLog
	types.Decode(receipt.Logs[0].Log, &pangLog)
	localKey := []byte(fmt.Sprintf(KeyPrefixPangLocal, pangLog.Msg))
	oldValue, err := h.GetLocalDB().Get(localKey)
	if err != nil {
		return nil, err
	}
	types.Decode(oldValue, &pangLog)
	if pangLog.Count > 0 {
		pangLog.Count--
	}
	val := types.Encode(&pangLog)
	if pangLog.Count == 0 {
		val = nil
	}
	kv := []*types.KeyValue{{Key: localKey, Value: val}}
	return &types.LocalDBSet{KV: kv}, nil
}

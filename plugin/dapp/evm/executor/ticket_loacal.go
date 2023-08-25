package executor

import (
	"fmt"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

func (evm *EVMExecutor) saveTicket(ticketlog *ty.ReceiptTicket) (kvs []*types.KeyValue) {
	if ticketlog.PrevStatus > 0 {
		kv := delticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.PrevStatus)
		kvs = append(kvs, kv)
	}
	kvs = append(kvs, addticket(ticketlog.Addr, ticketlog.TicketId, ticketlog.Status))
	return kvs
}

func (evm *EVMExecutor) saveTicketBind(b *ty.ReceiptTicketBind) (kvs []*types.KeyValue) {
	//解除原来的绑定
	if len(b.OldMinerAddress) > 0 {
		kv := &types.KeyValue{
			Key:   calcBindMinerKey(b.OldMinerAddress, b.ReturnAddress),
			Value: nil,
		}
		//tlog.Warn("tb:del", "key", string(kv.Key))
		kvs = append(kvs, kv)
	}

	kv := &types.KeyValue{Key: calcBindReturnKey(b.ReturnAddress), Value: []byte(b.NewMinerAddress)}
	//tlog.Warn("tb:add", "key", string(kv.Key), "value", string(kv.Value))

	kvs = append(kvs, kv)
	kv = &types.KeyValue{
		Key:   calcBindMinerKey(b.GetNewMinerAddress(), b.ReturnAddress),
		Value: []byte(b.ReturnAddress),
	}
	//tlog.Warn("tb:add", "key", string(kv.Key), "value", string(kv.Value))
	kvs = append(kvs, kv)
	return kvs
}

func addticket(addr string, ticketID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcTicketKey(addr, ticketID, status)
	kv.Value = []byte(ticketID)
	return kv
}

func delticket(addr string, ticketID string, status int32) *types.KeyValue {
	kv := &types.KeyValue{}
	kv.Key = calcTicketKey(addr, ticketID, status)
	kv.Value = nil
	return kv
}

func calcTicketKey(addr string, ticketID string, status int32) []byte {
	key := fmt.Sprintf("LODB-ticket-tl:%s:%d:%s", address.FormatAddrKey(addr), status, ticketID)
	return []byte(key)
}

func calcBindReturnKey(returnAddress string) []byte {
	key := fmt.Sprintf("LODB-ticket-bind:%s", address.FormatAddrKey(returnAddress))
	return []byte(key)
}

func calcBindMinerKey(minerAddress string, returnAddress string) []byte {
	key := fmt.Sprintf("LODB-ticket-miner:%s:%s", address.FormatAddrKey(minerAddress),
		address.FormatAddrKey(returnAddress))
	return []byte(key)
}

func calcBindMinerKeyPrefix(minerAddress string) []byte {
	key := fmt.Sprintf("LODB-ticket-miner:%s", address.FormatAddrKey(minerAddress))
	return []byte(key)
}

func calcTicketPrefix(addr string, status int32) []byte {
	key := fmt.Sprintf("LODB-ticket-tl:%s:%d", address.FormatAddrKey(addr), status)
	return []byte(key)
}

package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/exchange/types"
)

var (
	exchangeBindKeyPrefix = []byte("mavl-exchange-ebind-")
)

func (a *Action) ExchangeBind(payload *et.ExchangeBind) (*types.Receipt, error) {
	if a.fromaddr != payload.GetExchangeAddress() {
		return nil, types.ErrFromAddr
	}
	// If the value is null, the binding is unbound. If the value is not null, the address format is verified
	if len(payload.GetEntrustAddress()) > 0 {
		if err := address.CheckAddress(payload.GetEntrustAddress(), a.height); err != nil {
			return nil, err
		}
	}

	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	oldbind := a.getBind(payload.GetExchangeAddress())
	log := getBindLog(payload, oldbind)
	logs = append(logs, log)

	saveBind(a.statedb, payload)
	kv := getBindKV(payload)
	kvs = append(kvs, kv...)

	kvs = append(kvs, kv...)
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

func (a *Action) EntrustOrder(payload *et.EntrustOrder) (*types.Receipt, error) {
	entrustAddr, addr := a.fromaddr, payload.Addr

	if !a.checkBind(entrustAddr, addr) {
		return nil, et.ErrBindAddr
	}

	a.fromaddr = addr
	limitOrder := &et.LimitOrder{
		LeftAsset:  payload.LeftAsset,
		RightAsset: payload.RightAsset,
		Price:      payload.Price,
		Amount:     payload.Amount,
		Op:         payload.Op,
	}
	return a.LimitOrder(limitOrder, entrustAddr)
}

func (a *Action) EntrustRevokeOrder(payload *et.EntrustRevokeOrder) (*types.Receipt, error) {
	entrustAddr, addr := a.fromaddr, payload.Addr

	if !a.checkBind(entrustAddr, addr) {
		return nil, et.ErrBindAddr
	}

	a.fromaddr = payload.Addr
	revokeOrder := &et.RevokeOrder{OrderID: payload.OrderID}
	return a.RevokeOrder(revokeOrder)
}

func (a *Action) checkBind(entrustAddr, addr string) bool {
	return a.getBind(addr) == entrustAddr
}

func (a *Action) getBind(addr string) string {
	value, err := a.statedb.Get(bindKey(addr))
	if err != nil || value == nil {
		return ""
	}
	var bind et.ExchangeBind
	err = types.Decode(value, &bind)
	if err != nil {
		panic(err)
	}
	return bind.GetEntrustAddress()
}

func bindKey(id string) (key []byte) {
	key = append(key, exchangeBindKeyPrefix...)
	key = append(key, []byte(id)...)
	return key
}

func getBindLog(payload *et.ExchangeBind, old string) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = et.TyExchangeBindLog
	r := &et.ReceiptExchangeBind{}
	r.ExchangeAddress = payload.ExchangeAddress
	r.OldEntrustAddress = old
	r.NewEntrustAddress = payload.EntrustAddress
	log.Log = types.Encode(r)
	return log
}

func saveBind(db dbm.KV, payload *et.ExchangeBind) {
	set := getBindKV(payload)
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

func getBindKV(payload *et.ExchangeBind) (kvset []*types.KeyValue) {
	value := types.Encode(payload)
	kvset = append(kvset, &types.KeyValue{Key: bindKey(payload.ExchangeAddress), Value: value})
	return kvset
}

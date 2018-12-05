// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	log "github.com/33cn/chain33/common/log/log15"

	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

var relaylog = log.New("module", "execs.relay")

var driverName = "relay"
var subconfig = types.ConfSub(driverName)

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&relay{}))
}

// Init relay register driver
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newRelay, types.GetDappFork(driverName, "Enable")) //TODO: ForkV18Relay
}

// GetName relay get driver name
func GetName() string {
	return newRelay().GetName()
}

type relay struct {
	drivers.DriverBase
}

func newRelay() drivers.Driver {
	r := &relay{}
	r.SetChild(r)
	r.SetExecutorType(types.LoadExecutorType(driverName))
	return r
}

func (r *relay) GetDriverName() string {
	return driverName
}

func (r *relay) GetPayloadValue() types.Message {
	return &ty.RelayAction{}
}

func (r *relay) CheckTx(tx *types.Transaction, index int) error {
	return nil
}

func (r *relay) GetSellOrderByStatus(addrCoins *ty.ReqRelayAddrCoins) (types.Message, error) {
	var prefixs [][]byte
	if 0 == len(addrCoins.Coins) {
		val := calcOrderPrefixStatus((int32)(addrCoins.Status))
		prefixs = append(prefixs, val)
	} else {
		for _, coin := range addrCoins.Coins {
			val := calcOrderPrefixCoinStatus(coin, (int32)(addrCoins.Status))
			prefixs = append(prefixs, val)
		}
	}

	return r.GetSellOrder(prefixs)

}

func (r *relay) GetSellRelayOrder(addrCoins *ty.ReqRelayAddrCoins) (types.Message, error) {
	var prefixs [][]byte
	if 0 == len(addrCoins.Coins) {
		val := calcOrderPrefixAddr(addrCoins.Addr)
		prefixs = append(prefixs, val)
	} else {
		for _, coin := range addrCoins.Coins {
			val := calcOrderPrefixAddrCoin(addrCoins.Addr, coin)
			prefixs = append(prefixs, val)
		}
	}

	return r.GetSellOrder(prefixs)

}

func (r *relay) GetBuyRelayOrder(addrCoins *ty.ReqRelayAddrCoins) (types.Message, error) {
	var prefixs [][]byte
	if 0 == len(addrCoins.Coins) {
		val := calcAcceptPrefixAddr(addrCoins.Addr)
		prefixs = append(prefixs, val)
	} else {
		for _, coin := range addrCoins.Coins {
			val := calcAcceptPrefixAddrCoin(addrCoins.Addr, coin)
			prefixs = append(prefixs, val)
		}
	}

	return r.GetSellOrder(prefixs)

}

func (r *relay) GetSellOrder(prefixs [][]byte) (types.Message, error) {
	var OrderIds [][]byte

	for _, prefix := range prefixs {
		values, err := r.GetLocalDB().List(prefix, nil, 0, 0)
		if err != nil {
			return nil, err
		}

		if 0 != len(values) {
			OrderIds = append(OrderIds, values...)
		}
	}

	return r.getRelayOrderReply(OrderIds)

}

func (r *relay) getRelayOrderReply(OrderIDs [][]byte) (types.Message, error) {
	orderIDGot := make(map[string]bool)

	var reply ty.ReplyRelayOrders
	for _, orderID := range OrderIDs {
		if !orderIDGot[string(orderID)] {
			if order, err := r.getSellOrderFromDb(orderID); err == nil {
				reply.Relayorders = insertOrderDescending(order, reply.Relayorders)
			}
			orderIDGot[string(orderID)] = true
		}
	}
	return &reply, nil
}

func insertOrderDescending(toBeInserted *ty.RelayOrder, orders []*ty.RelayOrder) []*ty.RelayOrder {
	if 0 == len(orders) {
		orders = append(orders, toBeInserted)
	} else {
		index := len(orders)
		for i, element := range orders {
			if toBeInserted.Amount >= element.Amount {
				index = i
				break
			}
		}

		if len(orders) == index {
			orders = append(orders, toBeInserted)
		} else {
			rear := append([]*ty.RelayOrder{}, orders[index:]...)
			orders = append(orders[0:index], toBeInserted)
			orders = append(orders, rear...)
		}
	}
	return orders
}

func (r *relay) getSellOrderFromDb(OrderID []byte) (*ty.RelayOrder, error) {
	value, err := r.GetStateDB().Get(OrderID)
	if err != nil {
		return nil, err
	}
	var order ty.RelayOrder
	types.Decode(value, &order)
	return &order, nil
}

func (r *relay) getBTCHeaderFromDb(hash []byte) (*ty.BtcHeader, error) {
	value, err := r.GetStateDB().Get(hash)
	if err != nil {
		return nil, err
	}
	var header ty.BtcHeader
	types.Decode(value, &header)
	return &header, nil
}

func (r *relay) getOrderKv(OrderID []byte, ty int32) []*types.KeyValue {
	order, _ := r.getSellOrderFromDb(OrderID)

	var kv []*types.KeyValue
	kv = deleteCreateOrderKeyValue(kv, order, int32(order.PreStatus))
	kv = getCreateOrderKeyValue(kv, order, int32(order.Status))

	return kv
}

func (r *relay) getDeleteOrderKv(OrderID []byte, ty int32) []*types.KeyValue {
	order, _ := r.getSellOrderFromDb(OrderID)
	var kv []*types.KeyValue
	kv = deleteCreateOrderKeyValue(kv, order, int32(order.Status))
	kv = getCreateOrderKeyValue(kv, order, int32(order.PreStatus))

	return kv
}

func getCreateOrderKeyValue(kv []*types.KeyValue, order *ty.RelayOrder, status int32) []*types.KeyValue {
	orderID := []byte(order.Id)

	key := calcOrderKeyStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: orderID})

	key = calcOrderKeyCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: orderID})

	key = calcOrderKeyAddrStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: orderID})

	key = calcOrderKeyAddrCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: orderID})

	key = calcAcceptKeyAddr(order, status)
	if key != nil {
		kv = append(kv, &types.KeyValue{Key: key, Value: orderID})
	}

	return kv

}

func deleteCreateOrderKeyValue(kv []*types.KeyValue, order *ty.RelayOrder, status int32) []*types.KeyValue {

	key := calcOrderKeyStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	key = calcOrderKeyCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	key = calcOrderKeyAddrStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	key = calcOrderKeyAddrCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: nil})

	key = calcAcceptKeyAddr(order, status)
	if key != nil {
		kv = append(kv, &types.KeyValue{Key: key, Value: nil})
	}

	return kv
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (r *relay) CheckReceiptExecOk() bool {
	return true
}

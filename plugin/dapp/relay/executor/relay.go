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

func Init(name string, sub []byte) {
	drivers.Register(GetName(), newRelay, types.GetDappFork(driverName, "Enable")) //TODO: ForkV18Relay
}

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

func (c *relay) GetPayloadValue() types.Message {
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

func (r *relay) getRelayOrderReply(OrderIds [][]byte) (types.Message, error) {
	OrderIdGot := make(map[string]bool)

	var reply ty.ReplyRelayOrders
	for _, OrderId := range OrderIds {
		if !OrderIdGot[string(OrderId)] {
			if order, err := r.getSellOrderFromDb(OrderId); err == nil {
				reply.Relayorders = insertOrderDescending(order, reply.Relayorders)
			}
			OrderIdGot[string(OrderId)] = true
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

func (r *relay) getSellOrderFromDb(OrderId []byte) (*ty.RelayOrder, error) {
	value, err := r.GetStateDB().Get(OrderId)
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

func (r *relay) getOrderKv(OrderId []byte, ty int32) []*types.KeyValue {
	order, _ := r.getSellOrderFromDb(OrderId)

	var kv []*types.KeyValue
	kv = deleteCreateOrderKeyValue(kv, order, int32(order.PreStatus))
	kv = getCreateOrderKeyValue(kv, order, int32(order.Status))

	return kv
}

func (r *relay) getDeleteOrderKv(OrderId []byte, ty int32) []*types.KeyValue {
	order, _ := r.getSellOrderFromDb(OrderId)
	var kv []*types.KeyValue
	kv = deleteCreateOrderKeyValue(kv, order, int32(order.Status))
	kv = getCreateOrderKeyValue(kv, order, int32(order.PreStatus))

	return kv
}

func getCreateOrderKeyValue(kv []*types.KeyValue, order *ty.RelayOrder, status int32) []*types.KeyValue {
	OrderId := []byte(order.Id)

	key := calcOrderKeyStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: OrderId})

	key = calcOrderKeyCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: OrderId})

	key = calcOrderKeyAddrStatus(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: OrderId})

	key = calcOrderKeyAddrCoin(order, status)
	kv = append(kv, &types.KeyValue{Key: key, Value: OrderId})

	key = calcAcceptKeyAddr(order, status)
	if key != nil {
		kv = append(kv, &types.KeyValue{Key: key, Value: OrderId})
	}

	return kv

}

func deleteCreateOrderKeyValue(kv []*types.KeyValue, order *ty.RelayOrder, status int32) []*types.KeyValue {

	key := calcOrderKeyStatus(order, status)
	kv = append(kv, &types.KeyValue{key, nil})

	key = calcOrderKeyCoin(order, status)
	kv = append(kv, &types.KeyValue{key, nil})

	key = calcOrderKeyAddrStatus(order, status)
	kv = append(kv, &types.KeyValue{key, nil})

	key = calcOrderKeyAddrCoin(order, status)
	kv = append(kv, &types.KeyValue{key, nil})

	key = calcAcceptKeyAddr(order, status)
	if key != nil {
		kv = append(kv, &types.KeyValue{key, nil})
	}

	return kv
}

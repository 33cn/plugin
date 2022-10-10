// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"

	"strings"

	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

const (
	relayOrderSCAIH          = "LODB-relay-sellorder-scaih:"
	relayOrderCSAIH          = "LODB-relay-sellorder-csaih:"
	relayOrderASCIH          = "LODB-relay-sellorder-ascih:"
	relayOrderACSIH          = "LODB-relay-sellorder-acsih:"
	relayBuyOrderACSIH       = "LODB-relay-buyorder-acsih:"
	orderIDPrefix            = "mavl-relay-orderid-"
	coinHashPrefix           = "mavl-relay-coinhash-"
	btcLastHead              = "mavl-relay-btclasthead"
	relayBTCHeaderHash       = "LODB-relay-btcheader-hash"
	relayBTCHeaderHeight     = "LODB-relay-btcheader-height"
	relayBTCHeaderHeightList = "LODB-relay-btcheader-height-list"
)

var (
	relayBTCHeaderLastHeight = []byte("LODB-relay-btcheader-last-height")
	relayBTCHeaderBaseHeight = []byte("LODB-relay-btcheader-base-height")
)

func calcBtcHeaderKeyHash(hash string) []byte {
	key := fmt.Sprintf(relayBTCHeaderHash+"%s", hash)
	return []byte(key)
}

func calcBtcHeaderKeyHeight(height int64) []byte {
	key := fmt.Sprintf(relayBTCHeaderHeight+"%d", height)
	return []byte(key)
}

func calcBtcHeaderKeyHeightList(height int64) []byte {
	key := fmt.Sprintf(relayBTCHeaderHeightList+"%d", height)
	return []byte(key)
}

func calcOrderKeyStatus(order *ty.RelayOrder, status int32) []byte {
	key := fmt.Sprintf(relayOrderSCAIH+"%d:%s:%s:%s:%d",
		status, order.XCoin, address.FormatAddrKey(order.CreaterAddr), order.Id, order.Height)
	return []byte(key)
}

func calcOrderKeyCoin(order *ty.RelayOrder, status int32) []byte {
	key := fmt.Sprintf(relayOrderCSAIH+"%s:%d:%s:%s:%d",
		order.XCoin, status, address.FormatAddrKey(order.CreaterAddr), order.Id, order.Height)
	return []byte(key)
}

func calcOrderKeyAddrStatus(order *ty.RelayOrder, status int32) []byte {
	key := fmt.Sprintf(relayOrderASCIH+"%s:%d:%s:%s:%d",
		address.FormatAddrKey(order.CreaterAddr), status, order.XCoin, order.Id, order.Height)
	return []byte(key)
}

func calcOrderKeyAddrCoin(order *ty.RelayOrder, status int32) []byte {
	key := fmt.Sprintf(relayOrderACSIH+"%s:%s:%d:%s:%d",
		address.FormatAddrKey(order.CreaterAddr), order.XCoin, status, order.Id, order.Height)
	return []byte(key)
}

func calcOrderPrefixStatus(status int32) []byte {
	prefix := fmt.Sprintf(relayOrderSCAIH+"%d:", status)
	return []byte(prefix)
}

func calcOrderPrefixCoinStatus(coin string, status int32) []byte {
	prefix := fmt.Sprintf(relayOrderCSAIH+"%s:%d:", coin, status)
	return []byte(prefix)
}

func calcOrderPrefixAddrCoin(addr string, coin string) []byte {
	key := fmt.Sprintf(relayOrderACSIH+"%s:%s", address.FormatAddrKey(addr), coin)
	return []byte(key)
}

func calcOrderPrefixAddr(addr string) []byte {
	return []byte(fmt.Sprintf(relayOrderACSIH+"%s", address.FormatAddrKey(addr)))
}

func calcAcceptKeyAddr(order *ty.RelayOrder, status int32) []byte {
	if order.AcceptAddr != "" {
		return []byte(fmt.Sprintf(relayBuyOrderACSIH+"%s:%s:%d:%s:%d",
			address.FormatAddrKey(order.AcceptAddr), order.XCoin, status, order.Id, order.Height))
	}
	return nil

}

func calcAcceptPrefixAddr(addr string) []byte {
	return []byte(fmt.Sprintf(relayBuyOrderACSIH+"%s", address.FormatAddrKey(addr)))
}

func calcAcceptPrefixAddrCoin(addr, coin string) []byte {
	return []byte(fmt.Sprintf(relayBuyOrderACSIH+"%s:%s", address.FormatAddrKey(addr), coin))
}

func calcRelayOrderID(hash string) string {
	return orderIDPrefix + hash
}

func calcCoinHash(hash string) string {
	return coinHashPrefix + hash
}

func getRealTxHashID(id string) string {
	ids := strings.Split(id, "-")
	return ids[len(ids)-1]

}

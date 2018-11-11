// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	rTy "github.com/33cn/plugin/plugin/dapp/relay/types"
)

func (r *relay) Query_GetRelayOrderByStatus(in *rTy.ReqRelayAddrCoins) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return r.GetSellOrderByStatus(in)
}

func (r *relay) Query_GetSellRelayOrder(in *rTy.ReqRelayAddrCoins) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return r.GetSellRelayOrder(in)
}

func (r *relay) Query_GetBuyRelayOrder(in *rTy.ReqRelayAddrCoins) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return r.GetBuyRelayOrder(in)
}

func (r *relay) Query_GetBTCHeaderList(in *rTy.ReqRelayBtcHeaderHeightList) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := newBtcStore(r.GetLocalDB())
	return db.getHeadHeightList(in)
}

func (r *relay) Query_GetBTCHeaderCurHeight(in *rTy.ReqRelayQryBTCHeadHeight) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	db := newBtcStore(r.GetLocalDB())
	return db.getBtcCurHeight(in)
}

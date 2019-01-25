// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	mtypes "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

func newStore(db db.DB) *multisigStore {
	return &multisigStore{Store: wcom.NewStore(db)}
}

// multisigStore 多重签名数据库存储操作类
type multisigStore struct {
	*wcom.Store
}

//获取指定owner拥有的多重签名地址
func (store *multisigStore) listOwnerAttrsByAddr(addr string) (*mtypes.OwnerAttrs, error) {
	if len(addr) == 0 {
		bizlog.Error("listMultisigAddrByOwnerAddr addr is nil")
		return nil, types.ErrInvalidParam
	}

	ownerAttrByte, err := store.Get(calcMultisigAddr(addr))
	if err != nil {
		bizlog.Error("listMultisigAddrByOwnerAddr", "addr", addr, "db Get error ", err)
		if err == db.ErrNotFoundInDb {
			return nil, types.ErrNotFound
		}
		return nil, err
	}
	if nil == ownerAttrByte || len(ownerAttrByte) == 0 {
		return nil, types.ErrNotFound
	}
	var ownerAttrs mtypes.OwnerAttrs
	err = types.Decode(ownerAttrByte, &ownerAttrs)
	if err != nil {
		bizlog.Error("listMultisigAddrByOwnerAddr", "proto.Unmarshal err:", err)
		return nil, types.ErrUnmarshal
	}
	return &ownerAttrs, nil
}

//获取本钱包地址拥有的所有多重签名地址
func (store *multisigStore) listOwnerAttrs() (*mtypes.OwnerAttrs, error) {

	list := store.NewListHelper()
	ownerbytes := list.PrefixScan(calcPrefixMultisigAddr())
	if len(ownerbytes) == 0 {
		bizlog.Error("listOwnerAttrs is null")
		return nil, types.ErrNotFound
	}
	var replayOwnerAttrs mtypes.OwnerAttrs
	for _, ownerattrbytes := range ownerbytes {
		var ownerAttrs mtypes.OwnerAttrs
		err := types.Decode(ownerattrbytes, &ownerAttrs)
		if err != nil {
			bizlog.Error("listOwnerAttrs", "Decode err", err)
			continue
		}
		replayOwnerAttrs.Items = append(replayOwnerAttrs.Items, ownerAttrs.Items...)
	}
	return &replayOwnerAttrs, nil
}

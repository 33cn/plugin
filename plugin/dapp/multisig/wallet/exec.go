// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/types"
)

//On_MultiSigAddresList 获取owner对应的多重签名地址列表
func (policy *multisigPolicy) On_MultiSigAddresList(req *types.ReqString) (types.Message, error) {
	policy.getWalletOperate().GetMutex().Lock()
	defer policy.getWalletOperate().GetMutex().Unlock()

	//获取本钱包中记录的所有多重签名地址
	if req.Data == "" {
		reply, err := policy.store.listOwnerAttrs()
		if err != nil {
			bizlog.Error("On_MultiSigAddresList  listOwnerAttrs", "err", err)
		}
		return reply, err
	}
	//值查询指定owner地址拥有的多重签名地址列表
	reply, err := policy.store.listOwnerAttrsByAddr(req.Data)
	if err != nil {
		bizlog.Error("On_MultiSigAddresList listOwnerAttrsByAddr", "owneraddr", req.Data, "err", err)
	}
	return reply, err
}

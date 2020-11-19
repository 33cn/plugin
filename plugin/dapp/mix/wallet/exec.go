// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func (policy *mixPolicy) On_ShowAccountPrivacyInfo(req *types.ReqString) (types.Message, error) {
	return policy.getAccountPrivacyKey(req.Data)
}

func (policy *mixPolicy) On_ShowAccountNoteInfo(req *types.ReqAddrs) (types.Message, error) {
	return policy.showAccountNoteInfo(req.Addrs)
}

func (policy *mixPolicy) On_GetRescanStatus(in *types.ReqNil) (types.Message, error) {
	return &types.ReqString{Data: policy.getRescanStatus()}, nil

}

//重新扫描所有notes
func (policy *mixPolicy) On_RescanNotes(in *types.ReqNil) (types.Message, error) {
	err := policy.tryRescanNotes()
	if err != nil {
		bizlog.Error("rescanUTXOs", "err", err.Error())
	}
	return &types.ReqString{Data: "ok"}, err
}

func (policy *mixPolicy) On_EnablePrivacy(req *types.ReqAddrs) (types.Message, error) {
	return policy.enablePrivacy(req.Addrs)
}

func (policy *mixPolicy) On_EncodeSecretData(req *mixTy.SecretData) (types.Message, error) {
	return encodeSecretData(req)
}

func (policy *mixPolicy) On_EncryptSecretData(req *mixTy.EncryptSecretData) (types.Message, error) {
	return encryptSecretData(req)
}

func (policy *mixPolicy) On_DecryptSecretData(req *mixTy.DecryptSecretData) (types.Message, error) {
	return decryptSecretData(req)
}

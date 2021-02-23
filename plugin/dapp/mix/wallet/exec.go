// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

func (policy *mixPolicy) On_ShowAccountPrivacyInfo(req *mixTy.PaymentKeysReq) (types.Message, error) {
	if len(req.Addr) == 0 && len(req.PrivKey) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "addr or privkey need be set")
	}

	//通过私钥获取
	if len(req.PrivKey) > 0 {
		prikeybyte, err := common.FromHex(req.PrivKey)
		if err != nil {
			return nil, errors.Wrapf(err, "privkey fromHex error,key=%s", req.PrivKey)
		}
		var ret mixTy.WalletAddrPrivacy
		ret.Privacy = newPrivacyKey(prikeybyte)
		if req.Detail <= 0 {
			ret.Privacy.EncryptKey.PrivKey = ""
			ret.Privacy.PaymentKey.SpendKey = ""
		}
		return &ret, nil
	}

	//通过account 从钱包获取
	keys, err := policy.getAccountPrivacyKey(req.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "get account =%s privacy key", req.Addr)
	}
	if req.Detail <= 0 {
		keys.Privacy.EncryptKey.PrivKey = ""
		keys.Privacy.PaymentKey.SpendKey = ""
	}
	return keys, nil
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

//func (policy *mixPolicy) On_EncodeSecretData(req *mixTy.SecretData) (types.Message, error) {
//	return encodeSecretData(req)
//}

func (policy *mixPolicy) On_EncryptSecretData(req *mixTy.EncryptSecretData) (types.Message, error) {
	return encryptSecretData(req)
}

func (policy *mixPolicy) On_DecryptSecretData(req *mixTy.DecryptSecretData) (types.Message, error) {
	return decryptSecretData(req)
}

//func (policy *mixPolicy) On_DepositProof(req *mixTy.CreateRawTxReq) (types.Message, error) {
//	return policy.createDepositTx(req)
//}
//
//func (policy *mixPolicy) On_WithdrawProof(req *mixTy.CreateRawTxReq) (types.Message, error) {
//	return policy.createWithdrawTx(req)
//}
//
//func (policy *mixPolicy) On_AuthProof(req *mixTy.CreateRawTxReq) (types.Message, error) {
//	return policy.createAuthTx(req)
//}
//
//func (policy *mixPolicy) On_TransferProof(req *mixTy.CreateRawTxReq) (types.Message, error) {
//	return policy.createTransferTx(req)
//}

func (policy *mixPolicy) On_CreateRawTransaction(req *mixTy.CreateRawTxReq) (types.Message, error) {
	return policy.createRawTx(req)
}

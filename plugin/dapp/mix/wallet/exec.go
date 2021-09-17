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

func (p *mixPolicy) On_ShowAccountPrivacyInfo(req *mixTy.PaymentKeysReq) (types.Message, error) {
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
		if !req.Detail {
			ret.Privacy.SecretKey.SecretPrivKey = ""
			ret.Privacy.PaymentKey.SpendKey = ""
		}
		return &ret, nil
	}

	//通过account 从钱包获取
	keys, err := p.getAccountPrivacyKey(req.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "get account =%s privacy key", req.Addr)
	}
	if !req.Detail {
		keys.Privacy.SecretKey.SecretPrivKey = ""
		keys.Privacy.PaymentKey.SpendKey = ""
	}
	return keys, nil
}

func (p *mixPolicy) On_ShowAccountNoteInfo(req *mixTy.WalletMixIndexReq) (types.Message, error) {
	return p.showAccountNoteInfo(req)
}

func (p *mixPolicy) On_GetRescanStatus(in *types.ReqNil) (types.Message, error) {
	return &types.ReqString{Data: p.getRescanStatus()}, nil

}

//重新扫描所有notes
func (p *mixPolicy) On_RescanNotes(in *types.ReqNil) (types.Message, error) {
	err := p.tryRescanNotes()
	if err != nil {
		bizlog.Error("rescanUTXOs", "err", err.Error())
	}
	return &types.ReqString{Data: "ok"}, err
}

func (p *mixPolicy) On_EnablePrivacy(req *types.ReqAddrs) (types.Message, error) {
	return p.enablePrivacy(req.Addrs)
}

func (p *mixPolicy) On_EncryptSecretData(req *mixTy.EncryptSecretData) (types.Message, error) {
	return encryptSecretData(req)
}

func (p *mixPolicy) On_DecryptSecretData(req *mixTy.DecryptSecretData) (types.Message, error) {
	return decryptSecretData(req)
}

func (p *mixPolicy) On_CreateRawTransaction(req *mixTy.CreateRawTxReq) (types.Message, error) {
	return p.createRawTx(req)
}

func (p *mixPolicy) On_CreateZkKeyFile(req *mixTy.CreateZkKeyFileReq) (types.Message, error) {
	return p.createZkKeyFile(req)
}

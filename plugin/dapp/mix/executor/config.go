// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

// IsSuperManager is supper manager or not
func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confMix := types.ConfSub(cfg, mixTy.MixX)
	for _, m := range confMix.GStrList("mixApprs") {
		if addr == m {
			return true
		}
	}
	return false
}

// need super manager
func (a *action) Config(config *mixTy.MixConfigAction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	switch config.Ty {
	case mixTy.MixConfigType_Verify:
		//必须是超级管理员才能配置
		if !isSuperManager(cfg, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not super manager,%s", a.fromaddr)
		}
		return a.ConfigAddVerifyKey(config.GetVerifyKey())
	case mixTy.MixConfigType_Auth:
		//必须是超级管理员才能配置
		if !isSuperManager(cfg, a.fromaddr) {
			return nil, errors.Wrapf(types.ErrNotAllow, "not super manager,%s", a.fromaddr)
		}
		if config.Action == mixTy.MixConfigAct_Add {
			return a.ConfigAddAuthPubKey(config.GetAuthKey())
		} else {
			return a.ConfigDeleteAuthPubKey(config.GetAuthKey())
		}
	case mixTy.MixConfigType_Payment:
		//个人配置，个人负责，可重配
		return a.ConfigPaymentPubKey(config.GetNoteAccountKey())
	}
	return nil, errors.Wrapf(types.ErrNotFound, "ty=%d", config.Ty)

}

func makeConfigVerifyKeyReceipt(data *mixTy.ZkVerifyKeys, ty int32) *types.Receipt {
	key := getVerifyKeysKey(ty)
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(data)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: mixTy.TyLogMixConfigVk, Log: types.Encode(data)},
		},
	}

}

func getVerifyKeys(db dbm.KV, ty int32) (*mixTy.ZkVerifyKeys, error) {
	key := getVerifyKeysKey(ty)
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db verify key")
	}
	var keys mixTy.ZkVerifyKeys
	err = types.Decode(v, &keys)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db verify key")
	}

	return &keys, nil
}

func (a *action) ConfigAddVerifyKey(newKey *mixTy.MixZkVerifyKey) (*types.Receipt, error) {
	keys, err := getVerifyKeys(a.db, int32(newKey.Type))
	if isNotFound(errors.Cause(err)) {
		keys := &mixTy.ZkVerifyKeys{}
		keys.Data = append(keys.Data, newKey)
		return makeConfigVerifyKeyReceipt(keys, int32(newKey.Type)), nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "AddVerifyKey,ty=%d", newKey.Type)
	}
	//逆序保存keys,保证新的key先遍历到
	keys.Data = []*mixTy.MixZkVerifyKey{newKey, keys.Data[0]}
	return makeConfigVerifyKeyReceipt(keys, int32(newKey.Type)), nil

}

func makeConfigAuthKeyReceipt(data *mixTy.AuthKeys) *types.Receipt {
	key := getAuthPubKeysKey()
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(data)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: mixTy.TyLogMixConfigAuth, Log: types.Encode(data)},
		},
	}

}

func (a *action) getAuthKeys() (*mixTy.AuthKeys, error) {
	key := getAuthPubKeysKey()
	v, err := a.db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var keys mixTy.AuthKeys
	err = types.Decode(v, &keys)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db key")
	}

	return &keys, nil
}

func (a *action) ConfigAddAuthPubKey(key string) (*types.Receipt, error) {
	keys, err := a.getAuthKeys()
	if isNotFound(errors.Cause(err)) {
		keys := &mixTy.AuthKeys{}
		keys.Keys = append(keys.Keys, key)
		return makeConfigAuthKeyReceipt(keys), nil
	}
	if err != nil {
		return nil, err
	}

	keys.Keys = append(keys.Keys, key)
	return makeConfigAuthKeyReceipt(keys), nil
}

func (a *action) ConfigDeleteAuthPubKey(key string) (*types.Receipt, error) {
	keys, err := a.getAuthKeys()
	if err != nil {
		return nil, err
	}

	var newKeys mixTy.AuthKeys
	for _, v := range keys.Keys {
		if key == v {
			continue
		}
		newKeys.Keys = append(newKeys.Keys, v)
	}

	return makeConfigAuthKeyReceipt(&newKeys), nil
}

func makeConfigPaymentKeyReceipt(data *mixTy.NoteAccountKey) *types.Receipt {
	key := calcReceivingKey(data.Addr)
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(data)},
		},
		Logs: []*types.ReceiptLog{
			{Ty: mixTy.TyLogMixConfigPaymentKey, Log: types.Encode(data)},
		},
	}

}

func GetPaymentPubKey(db dbm.KV, addr string) (*mixTy.NoteAccountKey, error) {
	key := calcReceivingKey(addr)
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var keys mixTy.NoteAccountKey
	err = types.Decode(v, &keys)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db key")
	}

	return &keys, nil
}

func (a *action) ConfigPaymentPubKey(paykey *mixTy.NoteAccountKey) (*types.Receipt, error) {
	if paykey == nil || len(paykey.NoteReceiveAddr) == 0 || len(paykey.SecretReceiveKey) == 0 || len(paykey.Addr) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "pubkey=%v", paykey)
	}
	//检查用户使用对应的addr的key，但不能确保key就是对应addr
	if paykey.Addr != a.fromaddr {
		return nil, errors.Wrapf(types.ErrInvalidParam, "register addr=%s not match with sign=%s", paykey.Addr, a.fromaddr)
	}
	//直接覆盖
	return makeConfigPaymentKeyReceipt(&mixTy.NoteAccountKey{
		Addr:             a.fromaddr,
		NoteReceiveAddr:  paykey.NoteReceiveAddr,
		SecretReceiveKey: paykey.SecretReceiveKey}), nil

}

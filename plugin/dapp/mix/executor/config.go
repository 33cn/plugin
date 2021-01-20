// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	manager "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

// IsSuperManager is supper manager or not
func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confManager := types.ConfSub(cfg, manager.ManageX)
	for _, m := range confManager.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

// need super manager
func (a *action) Config(config *mixTy.MixConfigAction) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "not super manager,%s", a.fromaddr)
	}
	switch config.Ty {
	case mixTy.MixConfigType_VerifyKey:
		return a.ConfigAddVerifyKey(config.GetVerifyKey())
	case mixTy.MixConfigType_AuthPubKey:
		if config.Action == mixTy.MixConfigAct_Add {
			return a.ConfigAddAuthPubKey(config.GetAuthPk())
		} else {
			return a.ConfigDeleteAuthPubKey(config.GetAuthPk())
		}
	case mixTy.MixConfigType_PaymentPubKey:
		return a.ConfigPaymentPubKey(config.GetPaymentKey())
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

func (a *action) ConfigAddVerifyKey(newKey *mixTy.ZkVerifyKey) (*types.Receipt, error) {
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
	keys.Data = []*mixTy.ZkVerifyKey{newKey, keys.Data[0]}
	return makeConfigVerifyKeyReceipt(keys, int32(newKey.Type)), nil

}

func makeConfigAuthKeyReceipt(data *mixTy.AuthPubKeys) *types.Receipt {
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

func (a *action) getAuthKeys() (*mixTy.AuthPubKeys, error) {
	key := getAuthPubKeysKey()
	v, err := a.db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var keys mixTy.AuthPubKeys
	err = types.Decode(v, &keys)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db key")
	}

	return &keys, nil
}

func (a *action) ConfigAddAuthPubKey(key string) (*types.Receipt, error) {
	keys, err := a.getAuthKeys()
	if isNotFound(errors.Cause(err)) {
		keys := &mixTy.AuthPubKeys{}
		keys.Data = append(keys.Data, key)
		return makeConfigAuthKeyReceipt(keys), nil
	}
	if err != nil {
		return nil, err
	}

	keys.Data = append(keys.Data, key)
	return makeConfigAuthKeyReceipt(keys), nil
}

func (a *action) ConfigDeleteAuthPubKey(key string) (*types.Receipt, error) {
	keys, err := a.getAuthKeys()
	if err != nil {
		return nil, err
	}

	var newKeys mixTy.AuthPubKeys
	for _, v := range keys.Data {
		if key == v {
			continue
		}
		newKeys.Data = append(newKeys.Data, v)
	}

	return makeConfigAuthKeyReceipt(&newKeys), nil
}

func makeConfigPaymentKeyReceipt(data *mixTy.PaymentKey) *types.Receipt {
	key := getPaymentPubKey(data.Addr)
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

func GetPaymentPubKey(db dbm.KV, addr string) (*mixTy.PaymentKey, error) {
	key := getPaymentPubKey(addr)
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var keys mixTy.PaymentKey
	err = types.Decode(v, &keys)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db key")
	}

	return &keys, nil
}

func (a *action) ConfigPaymentPubKey(paykey *mixTy.PaymentKey) (*types.Receipt, error) {
	if paykey == nil || len(paykey.PayingKey) == 0 || len(paykey.ReceivingKey.X) == 0 || len(paykey.ReceivingKey.Y) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "pubkey=%v", paykey)
	}
	//直接覆盖
	return makeConfigPaymentKeyReceipt(&mixTy.PaymentKey{
		Addr:         a.fromaddr,
		PayingKey:    paykey.PayingKey,
		ReceivingKey: paykey.ReceivingKey}), nil

}

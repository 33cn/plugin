// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"github.com/33cn/chain33/common/db"

	system "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	wcom "github.com/33cn/chain33/wallet/common"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

func newStore(db db.DB) *mixStore {
	return &mixStore{Store: wcom.NewStore(db)}
}

// mixStore 隐私交易数据库存储操作类
type mixStore struct {
	*wcom.Store
}

func (store *mixStore) getAccountPrivacy(addr string) ([]byte, error) {
	if len(addr) == 0 {
		return nil, types.ErrInvalidParam
	}

	return store.Get(calcMixAddrKey(addr))

}

func (store *mixStore) setAccountPrivacy(addr string, data []byte) error {
	if len(addr) == 0 {
		bizlog.Error("SetWalletAccountPrivacy addr is nil")
		return types.ErrInvalidParam
	}
	if len(data) == 0 {
		bizlog.Error("SetWalletAccountPrivacy privacy is nil")
		return types.ErrInvalidParam
	}

	store.GetDB().Set(calcMixAddrKey(addr), data)

	return nil
}

func (store *mixStore) enablePrivacy() {
	newbatch := store.NewBatch(true)
	newbatch.Set(calcMixPrivacyEnable(), []byte("true"))
	newbatch.Write()
}

func (store *mixStore) getPrivacyEnable() bool {
	_, err := store.Get(calcMixPrivacyEnable())
	if err != nil {
		return false
	}
	return true
}

func (store *mixStore) setRescanNoteStatus(status int32) {
	newbatch := store.NewBatch(true)
	newbatch.Set(calcRescanNoteStatus(), []byte(mixTy.MixWalletRescanStatus(status).String()))
	newbatch.Write()
}

func (store *mixStore) setKvs(set *types.LocalDBSet) {
	newbatch := store.NewBatch(true)
	for _, s := range set.KV {
		newbatch.Set(s.Key, s.Value)
	}
	newbatch.Write()
}

func (store *mixStore) getRescanNoteStatus() int32 {
	v, err := store.Get(calcRescanNoteStatus())
	if err != nil {
		return int32(mixTy.MixWalletRescanStatus_IDLE)
	}
	return mixTy.MixWalletRescanStatus_value[string(v)]
}

//AddRollbackKV add rollback kv
func (d *mixStore) AddRollbackKV(tx *types.Transaction, execer []byte, kvs []*types.KeyValue) []*types.KeyValue {
	k := types.CalcRollbackKey(types.GetRealExecName(execer), tx.Hash())
	kvc := system.NewKVCreator(d.GetDB(), types.CalcLocalPrefix(execer), k)
	kvc.AddListNoPrefix(kvs)
	kvc.AddRollbackKV()
	return kvc.KVList()
}

//DelRollbackKV del rollback kv when exec_del_local
func (d *mixStore) DelRollbackKV(tx *types.Transaction, execer []byte) ([]*types.KeyValue, error) {
	krollback := types.CalcRollbackKey(types.GetRealExecName(execer), tx.Hash())
	kvc := system.NewKVCreator(d.GetDB(), types.CalcLocalPrefix(execer), krollback)
	kvs, err := kvc.GetRollbackKVList()
	if err != nil {
		return nil, err
	}
	for _, kv := range kvs {
		kvc.AddNoPrefix(kv.Key, kv.Value)
	}
	kvc.DelRollbackKV()
	return kvc.KVList(), nil
}

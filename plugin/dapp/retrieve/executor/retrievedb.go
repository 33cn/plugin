// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	//"bytes"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"

	//log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/system/dapp"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

const (
	retrieveBackup  = 1
	retrievePrepare = 2
	retrievePerform = 3
	retrieveCancel  = 4
)

// MaxRelation when backup
const MaxRelation = 10

// DB def
type DB struct {
	rt.Retrieve
}

// NewDB instance
func NewDB(backupaddress string) *DB {
	r := &DB{}
	r.BackupAddress = backupaddress

	return r
}

// RelateDB on retrieve action
func (r *DB) RelateDB(defaultAddress string, createTime int64, delayPeriod int64) bool {
	if len(r.RetPara) >= MaxRelation {
		return false
	}
	rlog.Debug("RetrieveBackup", "RelateDB", defaultAddress)
	para := &rt.RetrievePara{DefaultAddress: defaultAddress, Status: retrieveBackup, CreateTime: createTime, PrepareTime: 0, DelayPeriod: delayPeriod}
	r.RetPara = append(r.RetPara, para)

	return true
}

// UnRelateDB on retrieve action
func (r *DB) UnRelateDB(index int) bool {
	r.RetPara = append(r.RetPara[:index], r.RetPara[index+1:]...)
	return true
}

// CheckRelation on retrieve action
func (r *DB) CheckRelation(defaultAddress string) (int, bool) {
	for i := 0; i < len(r.RetPara); i++ {
		if r.RetPara[i].DefaultAddress == defaultAddress {
			return i, true
		}
	}
	return MaxRelation, false
}

// GetKVSet for retrieve
func (r *DB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&r.Retrieve)
	kvset = append(kvset, &types.KeyValue{Key: Key(r.BackupAddress), Value: value})
	return kvset
}

// Save KV
func (r *DB) Save(db dbm.KV) {
	set := r.GetKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

// Key for retrieve
func Key(address string) (key []byte) {
	key = append(key, []byte("mavl-retrieve-")...)
	key = append(key, address...)
	return key
}

// Action def
type Action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
}

// NewRetrieveAcction gen instance
func NewRetrieveAcction(r *Retrieve, tx *types.Transaction) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{r.GetCoinsAccount(), r.GetStateDB(), hash, fromaddr,
		r.GetBlockTime(), r.GetHeight(), dapp.ExecAddress(string(tx.Execer))}
}

// RetrieveBackup Action
func (action *Action) RetrieveBackup(backupRet *rt.BackupRetrieve) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt
	var r *DB
	var newRetrieve = false
	if types.IsDappFork(action.height, rt.RetrieveX, "ForkRetrive") {
		if err := address.CheckAddress(backupRet.BackupAddress); err != nil {
			rlog.Debug("retrieve checkaddress")
			return nil, err
		}
		if err := address.CheckAddress(backupRet.DefaultAddress); err != nil {
			rlog.Debug("retrieve checkaddress")
			return nil, err
		}

		if action.fromaddr != backupRet.DefaultAddress {
			rlog.Debug("RetrieveBackup", "action.fromaddr", action.fromaddr, "backupRet.DefaultAddress", backupRet.DefaultAddress)
			return nil, rt.ErrRetrieveDefaultAddress
		}
	}
	//用备份地址检索，如果没有，就建立新的，然后检查并处理关联
	retrieve, err := readRetrieve(action.db, backupRet.BackupAddress)
	if err != nil && err != types.ErrNotFound {
		rlog.Error("RetrieveBackup", "readRetrieve", err)
		return nil, err
	} else if err == types.ErrNotFound {
		newRetrieve = true
		rlog.Debug("RetrieveBackup", "newAddress", backupRet.BackupAddress)
	}

	if newRetrieve {
		r = NewDB(backupRet.BackupAddress)
	} else {
		r = &DB{*retrieve}
	}

	if index, related := r.CheckRelation(backupRet.DefaultAddress); !related {
		if !r.RelateDB(backupRet.DefaultAddress, action.blocktime, backupRet.DelayPeriod) {
			rlog.Debug("RetrieveBackup", "index", index)
			return nil, rt.ErrRetrieveRelateLimit
		}
	} else {
		rlog.Debug("RetrieveBackup", "repeataddr")
		return nil, rt.ErrRetrieveRepeatAddress
	}

	r.Save(action.db)
	kv = append(kv, r.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// RetrievePrepare Action
func (action *Action) RetrievePrepare(preRet *rt.PrepareRetrieve) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt
	var r *DB
	var index int
	var related bool

	retrieve, err := readRetrieve(action.db, preRet.BackupAddress)
	if err != nil {
		rlog.Debug("RetrievePrepare", "readRetrieve", err)
		return nil, err
	}
	r = &DB{*retrieve}
	if action.fromaddr != r.BackupAddress {
		rlog.Debug("RetrievePrepare", "action.fromaddr", action.fromaddr, "r.BackupAddress", r.BackupAddress)
		return nil, rt.ErrRetrievePrepareAddress
	}

	if index, related = r.CheckRelation(preRet.DefaultAddress); !related {
		rlog.Debug("RetrievePrepare", "CheckRelation", preRet.DefaultAddress)
		return nil, rt.ErrRetrieveRelation
	}

	if r.RetPara[index].Status != retrieveBackup {
		rlog.Debug("RetrievePrepare", "Status", r.RetPara[index].Status)
		return nil, rt.ErrRetrieveStatus
	}
	r.RetPara[index].PrepareTime = action.blocktime
	r.RetPara[index].Status = retrievePrepare

	r.Save(action.db)
	kv = append(kv, r.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// RetrievePerform Action
func (action *Action) RetrievePerform(perfRet *rt.PerformRetrieve) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt
	var index int
	var related bool
	var acc *types.Account

	retrieve, err := readRetrieve(action.db, perfRet.BackupAddress)
	if err != nil {
		rlog.Debug("RetrievePerform", "readRetrieve", perfRet.BackupAddress)
		return nil, err
	}

	r := &DB{*retrieve}

	if index, related = r.CheckRelation(perfRet.DefaultAddress); !related {
		rlog.Debug("RetrievePerform", "CheckRelation", perfRet.DefaultAddress)
		return nil, rt.ErrRetrieveRelation
	}

	if r.BackupAddress != action.fromaddr {
		rlog.Debug("RetrievePerform", "BackupAddress", r.BackupAddress, "action.fromaddr", action.fromaddr)
		return nil, rt.ErrRetrievePerformAddress
	}

	if r.RetPara[index].Status != retrievePrepare {
		rlog.Debug("RetrievePerform", "Status", r.RetPara[index].Status)
		return nil, rt.ErrRetrieveStatus
	}
	if action.blocktime-r.RetPara[index].PrepareTime < r.RetPara[index].DelayPeriod {
		rlog.Debug("RetrievePerform", "ErrRetrievePeriodLimit")
		return nil, rt.ErrRetrievePeriodLimit
	}

	acc = action.coinsAccount.LoadExecAccount(r.RetPara[index].DefaultAddress, action.execaddr)
	rlog.Debug("RetrievePerform", "acc.Balance", acc.Balance)
	if acc.Balance > 0 {
		receipt, err = action.coinsAccount.ExecTransfer(r.RetPara[index].DefaultAddress, r.BackupAddress, action.execaddr, acc.Balance)
		if err != nil {
			rlog.Debug("RetrievePerform", "ExecTransfer", err)
			return nil, err
		}
	} else {
		return nil, rt.ErrRetrieveNoBalance
	}

	//r.RetPara[index].Status = Retrieve_Performed
	//remove the relation
	r.UnRelateDB(index)
	r.Save(action.db)
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	kv = append(kv, r.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// RetrieveCancel Action
func (action *Action) RetrieveCancel(cancel *rt.CancelRetrieve) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt
	var index int
	var related bool

	retrieve, err := readRetrieve(action.db, cancel.BackupAddress)
	if err != nil {
		rlog.Debug("RetrieveCancel", "readRetrieve err", cancel.BackupAddress)
		return nil, err
	}
	r := &DB{*retrieve}

	if index, related = r.CheckRelation(cancel.DefaultAddress); !related {
		rlog.Debug("RetrieveCancel", "CheckRelation", cancel.DefaultAddress)
		return nil, rt.ErrRetrieveRelation
	}

	if action.fromaddr != r.RetPara[index].DefaultAddress {
		rlog.Debug("RetrieveCancel", "action.fromaddr", action.fromaddr, "DefaultAddress", r.RetPara[index].DefaultAddress)
		return nil, rt.ErrRetrieveCancelAddress
	}

	if r.RetPara[index].Status != retrievePrepare {
		rlog.Debug("RetrieveCancel", "Status", r.RetPara[index].Status)
		return nil, rt.ErrRetrieveStatus
	}

	//r.RetPara[index].Status = Retrieve_Canceled
	//remove the relation
	r.UnRelateDB(index)
	r.Save(action.db)
	kv = append(kv, r.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func readRetrieve(db dbm.KV, address string) (*rt.Retrieve, error) {
	data, err := db.Get(Key(address))
	if err != nil {
		rlog.Debug("readRetrieve", "get", err)
		return nil, err
	}
	var retrieve rt.Retrieve
	//decode
	err = types.Decode(data, &retrieve)
	if err != nil {
		rlog.Debug("readRetrieve", "decode", err)
		return nil, err
	}
	return &retrieve, nil
}

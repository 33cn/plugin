// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

//DelRetrieveInfo 删除
func DelRetrieveInfo(info *rt.RetrieveQuery, Status int64, db dbm.KVDB) (*types.KeyValue, error) {
	switch Status {
	case retrieveBackup:
		kv := &types.KeyValue{calcRetrieveKey(info.BackupAddress, info.DefaultAddress), nil}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	case retrievePrepare:
		oldInfo, err := getRetrieveInfo(db, info.BackupAddress, info.DefaultAddress)
		if oldInfo == nil {
			return nil, err
		}
		info.DelayPeriod = oldInfo.DelayPeriod
		info.Status = retrieveBackup
		info.PrepareTime = 0
		value := types.Encode(info)
		kv := &types.KeyValue{calcRetrieveKey(info.BackupAddress, info.DefaultAddress), value}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	case retrievePerform:
		fallthrough
	case retrieveCancel:
		oldInfo, err := getRetrieveInfo(db, info.BackupAddress, info.DefaultAddress)
		if oldInfo == nil {
			return nil, err
		}
		info.DelayPeriod = oldInfo.DelayPeriod
		info.Status = retrievePrepare
		info.PrepareTime = oldInfo.PrepareTime
		value := types.Encode(info)
		kv := &types.KeyValue{calcRetrieveKey(info.BackupAddress, info.DefaultAddress), value}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	default:
		return nil, nil
	}
}

func (c *Retrieve) execDelLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := c.DriverBase.ExecDelLocal(tx, receipt, index)
	if err != nil {
		return nil, err
	}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	return set, nil
}

//ExecDelLocal_Backup ...
func (c *Retrieve) ExecDelLocal_Backup(backup *rt.BackupRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := c.execDelLocal(tx, receiptData, index)

	info := rt.RetrieveQuery{backup.BackupAddress, backup.DefaultAddress, backup.DelayPeriod, zeroPrepareTime, zeroRemainTime, retrieveBackup}
	kv, err := DelRetrieveInfo(&info, retrieveBackup, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

//ExecDelLocal_Prepare ...
func (c *Retrieve) ExecDelLocal_Prepare(pre *rt.PrepareRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := c.execDelLocal(tx, receiptData, index)

	info := rt.RetrieveQuery{pre.BackupAddress, pre.DefaultAddress, zeroDelay, c.GetBlockTime(), zeroRemainTime, retrievePrepare}
	kv, err := DelRetrieveInfo(&info, retrievePrepare, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

//ExecDelLocal_Perform ...
func (c *Retrieve) ExecDelLocal_Perform(perf *rt.PerformRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := c.execDelLocal(tx, receiptData, index)

	info := rt.RetrieveQuery{perf.BackupAddress, perf.DefaultAddress, zeroDelay, zeroPrepareTime, zeroRemainTime, retrievePerform}
	kv, err := DelRetrieveInfo(&info, retrievePerform, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

//ExecDelLocal_Cancel ...
func (c *Retrieve) ExecDelLocal_Cancel(cancel *rt.CancelRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set, err := c.execDelLocal(tx, receiptData, index)

	info := rt.RetrieveQuery{cancel.BackupAddress, cancel.DefaultAddress, zeroDelay, zeroPrepareTime, zeroRemainTime, retrieveCancel}
	kv, err := DelRetrieveInfo(&info, retrieveCancel, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

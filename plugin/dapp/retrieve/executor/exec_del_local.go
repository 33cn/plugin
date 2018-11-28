// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

// DelRetrieveInfo local
func DelRetrieveInfo(info *rt.RetrieveQuery, Status int64, db dbm.KVDB) (*types.KeyValue, error) {
	switch Status {
	case retrieveBackup:
		kv := &types.KeyValue{Key: calcRetrieveKey(info.BackupAddress, info.DefaultAddress), Value: nil}
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
		kv := &types.KeyValue{Key: calcRetrieveKey(info.BackupAddress, info.DefaultAddress), Value: value}
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
		kv := &types.KeyValue{Key: calcRetrieveKey(info.BackupAddress, info.DefaultAddress), Value: value}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	default:
		return nil, nil
	}
}

// ExecDelLocal_Backup Action
func (c *Retrieve) ExecDelLocal_Backup(backup *rt.BackupRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecDelLocal_Backup")

	info := rt.RetrieveQuery{BackupAddress: backup.BackupAddress, DefaultAddress: backup.DefaultAddress, DelayPeriod: backup.DelayPeriod, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrieveBackup}
	kv, err := DelRetrieveInfo(&info, retrieveBackup, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecDelLocal_Prepare Action
func (c *Retrieve) ExecDelLocal_Prepare(pre *rt.PrepareRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecDelLocal_Prepare")

	info := rt.RetrieveQuery{BackupAddress: pre.BackupAddress, DefaultAddress: pre.DefaultAddress, DelayPeriod: zeroDelay, PrepareTime: c.GetBlockTime(), RemainTime: zeroRemainTime, Status: retrievePrepare}
	kv, err := DelRetrieveInfo(&info, retrievePrepare, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecDelLocal_Perform Action
func (c *Retrieve) ExecDelLocal_Perform(perf *rt.PerformRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecDelLocal_Perform")

	info := rt.RetrieveQuery{BackupAddress: perf.BackupAddress, DefaultAddress: perf.DefaultAddress, DelayPeriod: zeroDelay, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrievePerform}
	kv, err := DelRetrieveInfo(&info, retrievePerform, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecDelLocal_Cancel Action
func (c *Retrieve) ExecDelLocal_Cancel(cancel *rt.CancelRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecDelLocal_Cancel")

	info := rt.RetrieveQuery{BackupAddress: cancel.BackupAddress, DefaultAddress: cancel.DefaultAddress, DelayPeriod: zeroDelay, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrieveCancel}
	kv, err := DelRetrieveInfo(&info, retrieveCancel, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

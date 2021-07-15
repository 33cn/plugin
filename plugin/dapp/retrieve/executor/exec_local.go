// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

func createRetrieve(backupAddress, defaultAddress string, status int32) rt.RetrieveQuery {
	return rt.RetrieveQuery{BackupAddress: backupAddress, DefaultAddress: defaultAddress, DelayPeriod: zeroDelay, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: status}
}

// SaveRetrieveInfo local
func SaveRetrieveInfo(info *rt.RetrieveQuery, Status int64, db dbm.KVDB) (*types.KeyValue, error) {
	rlog.Debug("Retrieve SaveRetrieveInfo", "backupaddr", info.BackupAddress, "defaddr", info.DefaultAddress)
	switch Status {
	case retrieveBackup:
		oldInfo, err := getRetrieveInfo(db, info.BackupAddress, info.DefaultAddress)
		if oldInfo != nil && oldInfo.Status == retrieveBackup {
			return nil, err
		}
		value := types.Encode(info)
		kv := &types.KeyValue{Key: calcRetrieveKey(info.BackupAddress, info.DefaultAddress), Value: value}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	case retrievePrepare:
		oldInfo, err := getRetrieveInfo(db, info.BackupAddress, info.DefaultAddress)
		if oldInfo == nil {
			return nil, err
		}
		info.DelayPeriod = oldInfo.DelayPeriod
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
		info.PrepareTime = oldInfo.PrepareTime
		value := types.Encode(info)
		kv := &types.KeyValue{Key: calcRetrieveKey(info.BackupAddress, info.DefaultAddress), Value: value}
		db.Set(kv.Key, kv.Value)
		return kv, nil
	default:
		return nil, nil
	}
}

// ExecLocal_Backup Action
func (c *Retrieve) ExecLocal_Backup(backup *rt.BackupRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecLocal_Backup")
	info := createRetrieve(backup.BackupAddress, backup.DefaultAddress, retrieveBackup)
	info.DelayPeriod = backup.DelayPeriod
	kv, err := SaveRetrieveInfo(&info, retrieveBackup, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecLocal_Prepare Action
func (c *Retrieve) ExecLocal_Prepare(pre *rt.PrepareRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecLocal_Prepare")

	info := createRetrieve(pre.BackupAddress, pre.DefaultAddress, retrievePrepare)
	info.PrepareTime = c.GetBlockTime()
	kv, err := SaveRetrieveInfo(&info, retrievePrepare, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecLocal_Perform Action
func (c *Retrieve) ExecLocal_Perform(perf *rt.PerformRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecLocal_Perf")

	info := createRetrieve(perf.BackupAddress, perf.DefaultAddress, retrievePerform)
	kv, err := SaveRetrieveInfo(&info, retrievePerform, c.GetLocalDB())
	if err != nil {
		return set, nil
	}
	cfg := c.GetAPI().GetConfig()
	if cfg.IsDappFork(c.GetHeight(), rt.RetrieveX, rt.ForkRetriveAssetX) {
		if len(perf.Assets) == 0 {
			perf.Assets = append(perf.Assets, &rt.AssetSymbol{Exec: cfg.GetCoinExec(), Symbol: cfg.GetCoinSymbol()})
		}
	}
	for _, asset := range perf.Assets {
		value := types.Encode(&info)
		kv := &types.KeyValue{Key: calcRetrieveAssetKey(info.BackupAddress, info.DefaultAddress, asset.Exec, asset.Symbol), Value: value}
		c.GetLocalDB().Set(kv.Key, kv.Value)
		set.KV = append(set.KV, kv)
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

// ExecLocal_Cancel Action
func (c *Retrieve) ExecLocal_Cancel(cancel *rt.CancelRetrieve, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	rlog.Debug("Retrieve ExecLocal_Cancel")

	info := createRetrieve(cancel.BackupAddress, cancel.DefaultAddress, retrieveCancel)
	kv, err := SaveRetrieveInfo(&info, retrieveCancel, c.GetLocalDB())
	if err != nil {
		return set, nil
	}

	if kv != nil {
		set.KV = append(set.KV, kv)
	}

	return set, nil
}

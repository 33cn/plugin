// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

// Exec_Backup Action
func (c *Retrieve) Exec_Backup(backup *rt.BackupRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	if backup.DelayPeriod < minPeriod {
		return nil, rt.ErrRetrievePeriodLimit
	}
	rlog.Debug("RetrieveBackup action")
	return actiondb.RetrieveBackup(backup)
}

// Exec_Perform Action
func (c *Retrieve) Exec_Perform(perf *rt.PerformRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PerformRetrieve action")
	return actiondb.RetrievePerform(perf)
}

// Exec_Prepare Action
func (c *Retrieve) Exec_Prepare(pre *rt.PrepareRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PreRetrieve action")
	return actiondb.RetrievePrepare(pre)
}

// Exec_Cancel Action
func (c *Retrieve) Exec_Cancel(cancel *rt.CancelRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PreRetrieve action")
	return actiondb.RetrieveCancel(cancel)
}

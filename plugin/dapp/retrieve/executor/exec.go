package executor

import (
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

func (c *Retrieve) Exec_Backup(backup *rt.BackupRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	if backup.DelayPeriod < minPeriod {
		return nil, rt.ErrRetrievePeriodLimit
	}
	rlog.Debug("RetrieveBackup action")
	return actiondb.RetrieveBackup(backup)
}

func (c *Retrieve) Exec_Perform(perf *rt.PerformRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PerformRetrieve action")
	return actiondb.RetrievePerform(perf)
}

func (c *Retrieve) Exec_Prepare(pre *rt.PrepareRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PreRetrieve action")
	return actiondb.RetrievePrepare(pre)
}

func (c *Retrieve) Exec_Cancel(cancel *rt.CancelRetrieve, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewRetrieveAcction(c, tx)
	rlog.Debug("PreRetrieve action")
	return actiondb.RetrieveCancel(cancel)
}

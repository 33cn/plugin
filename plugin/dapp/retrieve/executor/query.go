// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

// Query_GetRetrieveInfo get retrieve state
func (r *Retrieve) Query_GetRetrieveInfo(in *rt.ReqRetrieveInfo) (types.Message, error) {
	rlog.Debug("Retrieve Query", "backupaddr", in.BackupAddress, "defaddr", in.DefaultAddress)
	info, err := getRetrieveInfo(r.GetLocalDB(), in.BackupAddress, in.DefaultAddress)
	if info == nil {
		return nil, err
	}
	if info.Status == retrievePrepare {
		info.RemainTime = info.DelayPeriod - (r.GetBlockTime() - info.PrepareTime)
		if info.RemainTime < 0 {
			info.RemainTime = 0
		}
	}
	return info, nil
}

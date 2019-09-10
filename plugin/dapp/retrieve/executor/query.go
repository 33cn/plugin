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

	// 在指定asset 的情况下， 显示具体asset 的找回状态
	if info.Status == retrievePerform && in.GetAssetExec() != "" {
		asset, err := getRetrieveAsset(r.GetLocalDB(), in.BackupAddress, in.DefaultAddress, in.AssetExec, in.AssetSymbol)
		if asset == nil {
			// retrievePerform状态下，不存在即可以找回，但还没找回
			info.Status = retrievePrepare
			info.RemainTime = zeroRemainTime
			return info, nil
		}
		return asset, err
	}
	return info, nil
}

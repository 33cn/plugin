// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

// Query_GetUnfreezeWithdraw 查询合约可提币量
func (u *Unfreeze) Query_GetUnfreezeWithdraw(in *types.ReqString) (types.Message, error) {
	return QueryWithdraw(u.GetStateDB(), in.GetData())
}

// Query_GetUnfreeze 查询合约状态
func (u *Unfreeze) Query_GetUnfreeze(in *types.ReqString) (types.Message, error) {
	return QueryUnfreeze(u.GetStateDB(), in.GetData())
}

// QueryWithdraw 查询可提币状态
func QueryWithdraw(stateDB dbm.KV, unfreezeID string) (types.Message, error) {
	unfreeze, err := loadUnfreeze(unfreezeID, stateDB)
	if err != nil {
		uflog.Error("QueryWithdraw ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}
	currentTime := time.Now().Unix()
	reply := &pty.ReplyQueryUnfreezeWithdraw{UnfreezeID: unfreezeID}
	available, err := getWithdrawAvailable(unfreeze, currentTime)
	if err != nil {
		return nil, err
	}

	reply.AvailableAmount = available
	return reply, nil
}

func getWithdrawAvailable(unfreeze *pty.Unfreeze, calcTime int64) (int64, error) {
	means, err := newMeans(unfreeze.Means)
	if err != nil {
		return 0, err
	}
	frozen, err := means.calcFrozen(unfreeze, calcTime)
	if err != nil {
		return 0, err
	}
	_, amount := withdraw(unfreeze, frozen)
	return amount, nil
}

// QueryUnfreeze 查询合约状态
func QueryUnfreeze(stateDB dbm.KV, unfreezeID string) (types.Message, error) {
	unfreeze, err := loadUnfreeze(unfreezeID, stateDB)
	if err != nil {
		uflog.Error("QueryUnfreeze ", "unfreezeID", unfreezeID, "err", err)
		return nil, err
	}

	return unfreeze, nil
}

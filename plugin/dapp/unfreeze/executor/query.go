// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"time"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
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

// Query_ListUnfreezeByCreator 查询列表
func (u *Unfreeze) Query_ListUnfreezeByCreator(in *pty.ReqUnfreezes) (types.Message, error) {
	return ListUnfreezeByCreator(u.GetLocalDB(), in)
}

// Query_ListUnfreezeByBeneficiary 查询列表
func (u *Unfreeze) Query_ListUnfreezeByBeneficiary(in *pty.ReqUnfreezes) (types.Message, error) {
	return ListUnfreezeByBeneficiary(u.GetLocalDB(), in)
}

// QueryWithdraw 查询可提币状态
func QueryWithdraw(stateDB dbm.KV, id string) (types.Message, error) {
	id = unfreezeIDFromHex(id)
	unfreeze, err := loadUnfreeze(id, stateDB)
	if err != nil {
		uflog.Error("QueryWithdraw ", "unfreezeID", id, "err", err)
		return nil, err
	}
	currentTime := time.Now().Unix()
	reply := &pty.ReplyQueryUnfreezeWithdraw{UnfreezeID: id}
	available, err := getWithdrawAvailable(unfreeze, currentTime)
	if err != nil {
		return nil, err
	}

	reply.AvailableAmount = available
	return reply, nil
}

func getWithdrawAvailable(unfreeze *pty.Unfreeze, calcTime int64) (int64, error) {
	means, err := newMeans(unfreeze.Means, 1500000)
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
func QueryUnfreeze(stateDB dbm.KV, id string) (types.Message, error) {
	id = unfreezeIDFromHex(id)
	unfreeze, err := loadUnfreeze(id, stateDB)
	if err != nil {
		uflog.Error("QueryUnfreeze ", "unfreezeID", id, "err", err)
		return nil, err
	}

	return unfreeze, nil
}

// ListUnfreezeByCreator 查询列表实现
func ListUnfreezeByCreator(ldb dbm.KVDB, req *pty.ReqUnfreezes) (types.Message, error) {
	if len(req.Initiator) == 0 {
		return nil, types.ErrInvalidParam
	}
	u := &pty.LocalUnfreeze{Unfreeze: &pty.Unfreeze{}}
	u.Unfreeze.Initiator = req.Initiator

	if len(req.FromKey) > 0 {
		u.TxIndex = req.FromKey
	}

	rows, err := list(ldb, "init", u, req.Count, req.Direction)
	if err != nil {
		uflog.Error("ListUnfreezeByCreator ", "err", err, "params", req)
		return nil, err
	}

	return fmtLocalUnfreeze(rows)
}

// ListUnfreezeByBeneficiary 查询列表实现
func ListUnfreezeByBeneficiary(ldb dbm.KVDB, req *pty.ReqUnfreezes) (types.Message, error) {
	if len(req.Beneficiary) == 0 {
		return nil, types.ErrInvalidParam
	}
	u := &pty.LocalUnfreeze{Unfreeze: &pty.Unfreeze{}}
	u.Unfreeze.Beneficiary = req.Beneficiary

	if len(req.FromKey) > 0 {
		u.TxIndex = req.FromKey
	}

	uflog.Error("ListUnfreezeByBeneficiary ", "params", req)
	rows, err := list(ldb, "beneficiary", u, req.Count, req.Direction)
	if err != nil {
		uflog.Error("ListUnfreezeByBeneficiary ", "err", err, "params", req)
		return nil, err
	}

	return fmtLocalUnfreeze(rows)
}

func fmtLocalUnfreeze(rows []*table.Row) (*pty.ReplyUnfreezes, error) {
	var results pty.ReplyUnfreezes
	for _, row := range rows {
		r, ok := row.Data.(*pty.LocalUnfreeze)
		if !ok {
			uflog.Error("ListUnfreeze", "err", "bad row type")
			return nil, types.ErrDecode
		}
		v := &pty.ReplyUnfreeze{
			UnfreezeID:  r.Unfreeze.UnfreezeID,
			StartTime:   r.Unfreeze.StartTime,
			AssetExec:   r.Unfreeze.AssetExec,
			AssetSymbol: r.Unfreeze.AssetSymbol,
			TotalCount:  r.Unfreeze.TotalCount,
			Initiator:   r.Unfreeze.Initiator,
			Beneficiary: r.Unfreeze.Beneficiary,
			Remaining:   r.Unfreeze.Remaining,
			Means:       r.Unfreeze.Means,
			Terminated:  r.Unfreeze.Terminated,
			Key:         r.TxIndex,
		}
		if v.Means == pty.FixAmountX {
			v.MeansOpt = &pty.ReplyUnfreeze_FixAmount{FixAmount: r.Unfreeze.GetFixAmount()}
		} else if v.Means == pty.LeftProportionX {
			v.MeansOpt = &pty.ReplyUnfreeze_LeftProportion{LeftProportion: r.Unfreeze.GetLeftProportion()}
		}
		results.Unfreeze = append(results.Unfreeze, v)
	}
	return &results, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/issuance/types"
)

func (c *Issuance) Query_IssuanceInfoByID(req *pty.ReqIssuanceInfo) (types.Message, error) {
	issu,err := queryIssuanceByID(c.GetStateDB(), req.IssuanceId)
	if err != nil {
		clog.Error("Query_IssuanceInfoByID", "id", req.IssuanceId, "error", err)
		return nil, err
	}

	return &pty.RepIssuanceCurrentInfo{
		Status:             issu.Status,
		TotalBalance:       issu.TotalBalance,
		DebtCeiling:        issu.DebtCeiling,
		LiquidationRatio:   issu.LiquidationRatio,
		Balance:            issu.Balance,
	}, nil
}

func (c *Issuance) Query_IssuanceInfoByIDs(req *pty.ReqIssuanceInfos) (types.Message, error) {
	infos := &pty.RepIssuanceCurrentInfos{}
	for _, id := range req.IssuanceIds {
		issu,err := queryIssuanceByID(c.GetStateDB(), id)
		if err != nil {
			clog.Error("Query_IssuanceInfoByID", "id", id, "error", err)
			return nil, err
		}

		infos.Infos = append(infos.Infos, &pty.RepIssuanceCurrentInfo{
			Status:             issu.Status,
			TotalBalance:       issu.TotalBalance,
			DebtCeiling:        issu.DebtCeiling,
			LiquidationRatio:   issu.LiquidationRatio,
			Balance:            issu.Balance,
		})
	}

	return infos, nil
}

func (c *Issuance) Query_IssuanceByStatus(req *pty.ReqIssuanceByStatus) (types.Message, error) {
	ids := &pty.RepIssuanceIDs{}
	issuIDRecords, err := queryIssuanceByStatus(c.GetLocalDB(), req.Status)
	if err != nil {
		clog.Error("Query_IssuanceByStatus", "get issuance record error", err)
		return nil, err
	}

	for _, record := range issuIDRecords {
		ids.IDs = append(ids.IDs, record.IssuanceId)
	}

	return ids, nil
}

func (c *Issuance) Query_IssuanceRecordByID(req *pty.ReqIssuanceDebtInfo) (types.Message, error) {
	issuRecord, err := queryIssuanceRecordByID(c.GetStateDB(), req.IssuanceId, req.DebtId)
	if err != nil {
		clog.Error("Query_IssuanceRecordByID", "get issuance record error", err)
		return nil, err
	}

	ret := &pty.RepIssuanceDebtInfo{}
	ret.Record = issuRecord
	return issuRecord, nil
}

func (c *Issuance) Query_IssuanceRecordsByAddr(req *pty.ReqIssuanceRecordsByAddr) (types.Message, error) {
	records, err := queryIssuanceRecordByAddr(c.GetStateDB(), c.GetLocalDB(), req.Addr)
	if err != nil {
		clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
		return nil, err
	}

	ret := &pty.RepIssuanceRecords{}
	ret.Records = records
	return ret, nil
}

func (c *Issuance) Query_IssuanceRecordsByStatus(req *pty.ReqIssuanceRecordsByStatus) (types.Message, error) {
	records, err := queryIssuanceRecordsByStatus(c.GetStateDB(), c.GetLocalDB(), req.Status)
	if err != nil {
		clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
		return nil, err
	}

	ret := &pty.RepIssuanceRecords{}
	ret.Records = records
	return ret, nil
}
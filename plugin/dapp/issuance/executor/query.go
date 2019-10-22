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

func (c *Issuance) Query_IssuanceByAddr(req *pty.ReqIssuanceByAddr) (types.Message, error) {
	ids := &pty.RepIssuanceIDs{}
	issuIDRecords, err := queryIssuanceByAddr(c.GetLocalDB(), req.Addr)
	if err != nil {
		clog.Error("Query_IssuanceByAddr", "get issuance record error", err)
		return nil, err
	}

	for _, record := range issuIDRecords {
		ids.IDs = append(ids.IDs, record.IssuanceId)
	}

	return ids, nil
}

func (c *Issuance) Query_IssuanceDebtInfoByAddr(req *pty.ReqIssuanceDebtInfoByAddr) (types.Message, error) {
	records, err := queryIssuanceByAddr(c.GetLocalDB(), req.Addr)
	if err != nil {
		clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
		return nil, err
	}

	ret := &pty.RepIssuanceDebtInfos{}
	for _, record := range records {
		if record.IssuanceId == req.IssuanceId {
			issu, err := queryIssuanceByID(c.GetStateDB(), record.IssuanceId)
			if err != nil {
				clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
				return nil, err
			}

			for _, borrowRecord := range issu.DebtRecords {
				if borrowRecord.AccountAddr == req.Addr {
					ret.Record = append(ret.Record, borrowRecord)
				}
			}

			for _, borrowRecord := range issu.InvalidRecords {
				if borrowRecord.AccountAddr == req.Addr {
					ret.Record = append(ret.Record, borrowRecord)
				}
			}
		}
	}

	return nil, pty.ErrRecordNotExist
}

func (c *Issuance) Query_IssuanceDebtInfoByStatus(req *pty.ReqIssuanceDebtInfoByStatus) (types.Message, error) {
	records, err := queryIssuanceRecordByStatus(c.GetLocalDB(), req.Status)
	if err != nil {
		clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
		return nil, err
	}

	ret := &pty.RepIssuanceDebtInfos{}
	for _, record := range records {
		issu, err := queryIssuanceByID(c.GetStateDB(), record.IssuanceId)
		if err != nil {
			clog.Error("Query_IssuanceDebtInfoByAddr", "get issuance record error", err)
			return nil, err
		}

		for _, borrowRecord := range issu.DebtRecords {
			if borrowRecord.Status == req.Status {
				ret.Record = append(ret.Record, borrowRecord)
			}
		}

		for _, borrowRecord := range issu.InvalidRecords {
			if borrowRecord.Status == req.Status {
				ret.Record = append(ret.Record, borrowRecord)
			}
		}
	}

	return ret, nil
}
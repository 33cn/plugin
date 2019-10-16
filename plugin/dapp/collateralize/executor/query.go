// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

func (c *Collateralize) Query_CollateralizeInfoByID(req *pty.ReqCollateralizeInfo) (types.Message, error) {
	coll,err := queryCollateralizeByID(c.GetStateDB(), req.CollateralizeId)
	if err != nil {
		clog.Error("Query_CollateralizeInfoByID", "id", req.CollateralizeId, "error", err)
		return nil, err
	}

	return &pty.RepCollateralizeCurrentInfo{
		Status:             coll.Status,
		TotalBalance:       coll.TotalBalance,
		DebtCeiling:        coll.DebtCeiling,
		LiquidationRatio:   coll.LiquidationRatio,
		StabilityFee:       coll.StabilityFee,
		CreateAddr:         coll.CreateAddr,
		Balance:            coll.Balance,
	}, nil
}

func (c *Collateralize) Query_CollateralizeInfoByIDs(req *pty.ReqCollateralizeInfos) (types.Message, error) {
	infos := &pty.RepCollateralizeCurrentInfos{}
	for _, id := range req.CollateralizeIds {
		coll,err := queryCollateralizeByID(c.GetStateDB(), id)
		if err != nil {
			clog.Error("Query_CollateralizeInfoByID", "id", id, "error", err)
			return nil, err
		}

		infos.Infos = append(infos.Infos, &pty.RepCollateralizeCurrentInfo{
			Status:             coll.Status,
			TotalBalance:       coll.TotalBalance,
			DebtCeiling:        coll.DebtCeiling,
			LiquidationRatio:   coll.LiquidationRatio,
			StabilityFee:       coll.StabilityFee,
			CreateAddr:         coll.CreateAddr,
			Balance:            coll.Balance,
		})
	}

	return infos, nil
}

func (c *Collateralize) Query_CollateralizeByStatus(req *pty.ReqCollateralizeByStatus) (types.Message, error) {
	ids := &pty.RepCollateralizeIDs{}
	collIDRecords, err := queryCollateralizeByStatus(c.GetLocalDB(), pty.CollateralizeStatusCreated)
	if err != nil {
		clog.Error("Query_CollateralizeByStatus", "get collateralize record error", err)
		return nil, err
	}

	for _, record := range collIDRecords {
		ids.IDs = append(ids.IDs, record.CollateralizeId)
	}

	return ids, nil
}

func (c *Collateralize) Query_CollateralizeByAddr(req *pty.ReqCollateralizeByAddr) (types.Message, error) {
	ids := &pty.RepCollateralizeIDs{}
	collIDRecords, err := queryCollateralizeByAddr(c.GetLocalDB(), req.Addr)
	if err != nil {
		clog.Error("Query_CollateralizeByAddr", "get collateralize record error", err)
		return nil, err
	}

	for _, record := range collIDRecords {
		ids.IDs = append(ids.IDs, record.CollateralizeId)
	}

	return ids, nil
}

func (c *Collateralize) Query_CollateralizeBorrowInfoByAddr(req *pty.ReqCollateralizeBorrowInfoByAddr) (types.Message, error) {
	records, err := queryCollateralizeByAddr(c.GetLocalDB(), req.Addr)
	if err != nil {
		clog.Error("Query_CollateralizeBorrowInfoByAddr", "get collateralize record error", err)
		return nil, err
	}

	ret := &pty.RepCollateralizeBorrowInfos{}
	for _, record := range records {
		if record.CollateralizeId == req.CollateralizeId {
			coll, err := queryCollateralizeByID(c.GetStateDB(), record.CollateralizeId)
			if err != nil {
				clog.Error("Query_CollateralizeBorrowInfoByAddr", "get collateralize record error", err)
				return nil, err
			}

			for _, borrowRecord := range coll.BorrowRecords {
				if borrowRecord.AccountAddr == req.Addr {
					ret.Record = append(ret.Record, borrowRecord)
				}
			}

			for _, borrowRecord := range coll.InvalidRecords {
				if borrowRecord.AccountAddr == req.Addr {
					ret.Record = append(ret.Record, borrowRecord)
				}
			}
		}
	}

	return nil, pty.ErrRecordNotExist
}

func (c *Collateralize) Query_CollateralizeBorrowInfoByStatus(req *pty.ReqCollateralizeBorrowInfoByStatus) (types.Message, error) {
	records, err := queryCollateralizeRecordByStatus(c.GetLocalDB(), req.Status)
	if err != nil {
		clog.Error("Query_CollateralizeBorrowInfoByAddr", "get collateralize record error", err)
		return nil, err
	}

	ret := &pty.RepCollateralizeBorrowInfos{}
	for _, record := range records {
		coll, err := queryCollateralizeByID(c.GetStateDB(), record.CollateralizeId)
		if err != nil {
			clog.Error("Query_CollateralizeBorrowInfoByAddr", "get collateralize record error", err)
			return nil, err
		}

		for _, borrowRecord := range coll.BorrowRecords {
			if borrowRecord.Status == req.Status {
				ret.Record = append(ret.Record, borrowRecord)
			}
		}

		for _, borrowRecord := range coll.InvalidRecords {
			if borrowRecord.Status == req.Status {
				ret.Record = append(ret.Record, borrowRecord)
			}
		}
	}

	return ret, nil
}
// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

func (d *DPos) updateCandVote(log *dty.ReceiptCandicator) (kvs []*types.KeyValue, err error) {
	voteTable := dty.NewDposVoteTable(d.GetLocalDB())
	canTable := dty.NewDposCandidatorTable(d.GetLocalDB())

	if log.Status == dty.CandidatorStatusRegist {
		candInfo := log.CandInfo
		log.CandInfo = nil

		err = canTable.Add(candInfo)
		if err != nil {
			return nil, err
		}

		kvs, err = canTable.Save()
		if err != nil {
			return nil, err
		}
	} else if log.Status == dty.CandidatorStatusVoted {
		candInfo := log.CandInfo
		log.CandInfo = nil
		voter := candInfo.Voters[len(candInfo.Voters)-1]

		err = canTable.Replace(candInfo)
		if err != nil {
			return nil, err
		}

		kvs1, err := canTable.Save()
		if err != nil {
			return nil, err
		}

		err = voteTable.Add(voter)
		if err != nil {
			return nil, err
		}

		kvs2, err := voteTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	} else if log.Status == dty.CandidatorStatusCancelVoted {
		candInfo := log.CandInfo
		log.CandInfo = nil
		voter := &dty.DposVoter{
			FromAddr: log.FromAddr,
			Pubkey: log.Pubkey,
			Votes: log.Votes,
			Index: log.Index,
			Time: log.Time,
		}

		err = canTable.Replace(candInfo)
		if err != nil {
			return nil, err
		}

		kvs1, err := canTable.Save()
		if err != nil {
			return nil, err
		}

		err = voteTable.Add(voter)
		if err != nil {
			return nil, err
		}

		kvs2, err := voteTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	} else if log.Status == dty.CandidatorStatusReRegist || log.Status == dty.CandidatorStatusCancelRegist{
		candInfo := log.CandInfo
		log.CandInfo = nil

		err = canTable.Replace(candInfo)
		if err != nil {
			return nil, err
		}

		kvs, err = canTable.Save()
		if err != nil {
			return nil, err
		}
	}

	return kvs, nil
}

func (d *DPos) updateVrf(log *dty.ReceiptVrf) (kvs []*types.KeyValue, err error) {
	if log.Status == dty.VrfStatusMRegist {
		vrfMTable := dty.NewDposVrfMTable(d.GetLocalDB())
		vrfM := &dty.DposVrfM{
			Index: log.Index,
			Pubkey: log.Pubkey,
			Cycle: log.Cycle,
			Height: log.Height,
			M: log.M,
			Time: log.Time,
		}

		err = vrfMTable.Add(vrfM)
		if err != nil {
			return nil, err
		}

		kvs, err = vrfMTable.Save()
		if err != nil {
			return nil, err
		}
	} else if log.Status == dty.VrfStatusRPRegist {
		VrfRPTable := dty.NewDposVrfRPTable(d.GetLocalDB())
		vrfRP := &dty.DposVrfRP{
			Index: log.Index,
			Pubkey: log.Pubkey,
			Cycle: log.Cycle,
			Height: log.Height,
			R: log.R,
			P: log.P,
			M: log.M,
			Time: log.Time,
		}

		err = VrfRPTable.Add(vrfRP)
		if err != nil {
			return nil, err
		}

		kvs, err = VrfRPTable.Save()
		if err != nil {
			return nil, err
		}
	}

	return kvs, nil
}

func (d *DPos) execLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	for _, item := range receipt.Logs {
		if item.Ty >= dty.TyLogCandicatorRegist && item.Ty <= dty.TyLogCandicatorReRegist {
			var candLog dty.ReceiptCandicator
			err := types.Decode(item.Log, &candLog)
			if err != nil {
				return nil, err
			}
			kvs, err := d.updateCandVote(&candLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kvs...)
		} else if  item.Ty >= dty.TyLogVrfMRegist && item.Ty <= dty.TyLogVrfRPRegist {
			var vrfLog dty.ReceiptVrf
			err := types.Decode(item.Log, &vrfLog)
			if err != nil {
				return nil, err
			}
			kvs, err := d.updateVrf(&vrfLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kvs...)
		}
	}

	return dbSet, nil
}

//ExecLocal_Regist method
func (d *DPos) ExecLocal_Regist(payload *dty.DposCandidatorRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_CancelRegist method
func (d *DPos) ExecLocal_CancelRegist(payload *dty.DposCandidatorCancelRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_ReRegist method
func (d *DPos) ExecLocal_ReRegist(payload *dty.DposCandidatorRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_Vote method
func (d *DPos) ExecLocal_Vote(payload *dty.DposVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_CancelVote method
func (d *DPos) ExecLocal_CancelVote(payload *dty.DposCancelVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_VrfMRegist method
func (d *DPos) ExecLocal_VrfMRegist(payload *dty.DposVrfMRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}

//ExecLocal_VrfRPRegist method
func (d *DPos) ExecLocal_VrfRPRegist(payload *dty.DposVrfRPRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execLocal(receiptData)
}
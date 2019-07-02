// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

func (d *DPos) rollbackCand(cand *dty.CandidatorInfo, log *dty.ReceiptCandicator) {
	if cand == nil || log == nil {
		return
	}

	//如果状态发生了变化，则需要将状态恢复到前一状态
	if log.StatusChange {
		cand.Status = log.PreStatus
		cand.Index = cand.PreIndex
	}

	//如果投票了，则需要把投票回滚
	if log.Voted {
		cand.Votes -= log.Votes
		if log.Votes > 0 {
			//如果是投票，则回滚时将投票删除。
			cand.Voters = cand.Voters[:len(cand.Voters)-1]
		} else {
			//如果是撤销投票，则回滚时，将删除的投票还回来
			voter := &dty.DposVoter{
				FromAddr: log.FromAddr,
				Pubkey: log.Pubkey,
				Votes: -log.Votes,
				Index: log.Index,
				Time: log.Time - dty.VoteFrozenTime,
			}
			cand.Voters = append(cand.Voters, voter)
		}
	}
}

func (d *DPos) rollbackCandVote(log *dty.ReceiptCandicator) (kvs []*types.KeyValue, err error) {
	voterTable := dty.NewDposVoteTable(d.GetLocalDB())
	candTable := dty.NewDposCandidatorTable(d.GetLocalDB())
	if err != nil {
		return nil, err
	}

	if log.Status == dty.CandidatorStatusRegist {
		//注册回滚,cand表删除记录
		err = candTable.Del(log.Pubkey)
		if err != nil {
			return nil, err
		}
		kvs, err = candTable.Save()
		return kvs, err
	} else if log.Status == dty.CandidatorStatusVoted {
		//投票阶段回滚，回滚状态，回滚投票
		candInfo := log.CandInfo
		log.CandInfo = nil

		//先回滚候选节点信息
		d.rollbackCand(candInfo, log)

		err = candTable.Replace(candInfo)
		if err != nil {
			return nil, err
		}
		kvs1, err := candTable.Save()
		if err != nil {
			return nil, err
		}

		//删除投票信息
		err = voterTable.Del([]byte(fmt.Sprintf("%018d", log.Index)))
		if err != nil {
			return nil, err
		}

		kvs2, err := voterTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	} else if log.Status == dty.CandidatorStatusCancelVoted {
		//撤销投票回滚，需要将撤销的投票还回来
		candInfo := log.CandInfo
		log.CandInfo = nil

		//先回滚候选节点信息
		d.rollbackCand(candInfo, log)

		err = candTable.Replace(candInfo)
		if err != nil {
			return nil, err
		}
		kvs1, err := candTable.Save()
		if err != nil {
			return nil, err
		}

		//删除投票信息
		err = voterTable.Del([]byte(fmt.Sprintf("%018d", log.Index)))
		if err != nil {
			return nil, err
		}

		kvs2, err := voterTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	}

	return kvs, nil
}

func (d *DPos) rollbackVrf(log *dty.ReceiptVrf) (kvs []*types.KeyValue, err error) {
	if log.Status == dty.VrfStatusMRegist {
		vrfMTable := dty.NewDposVrfMTable(d.GetLocalDB())

		//注册回滚,cand表删除记录
		err = vrfMTable.Del([]byte(fmt.Sprintf("%018d", log.Index)))
		if err != nil {
			return nil, err
		}
		kvs, err = vrfMTable.Save()
		return kvs, err
	} else if log.Status == dty.VrfStatusRPRegist {
		VrfRPTable := dty.NewDposVrfRPTable(d.GetLocalDB())

		err = VrfRPTable.Del([]byte(fmt.Sprintf("%018d", log.Index)))
		if err != nil {
			return nil, err
		}
		kvs, err = VrfRPTable.Save()
		return kvs, err
	}

	return nil, nil
}

func (d *DPos) execDelLocal(receipt *types.ReceiptData) (*types.LocalDBSet, error) {
	dbSet := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return dbSet, nil
	}

	for _, log := range receipt.Logs {
		switch log.GetTy() {
		case dty.CandidatorStatusRegist, dty.CandidatorStatusVoted, dty.CandidatorStatusCancelVoted, dty.CandidatorStatusCancelRegist:
			receiptLog := &dty.ReceiptCandicator{}
			if err := types.Decode(log.Log, receiptLog); err != nil {
				return nil, err
			}
			kv, err := d.rollbackCandVote(receiptLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kv...)

		case dty.VrfStatusMRegist, dty.VrfStatusRPRegist:
			receiptLog := &dty.ReceiptVrf{}
			if err := types.Decode(log.Log, receiptLog); err != nil {
				return nil, err
			}
			kv, err := d.rollbackVrf(receiptLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kv...)
		}
	}

	return dbSet, nil
}

//ExecDelLocal_Regist method
func (d *DPos) ExecDelLocal_Regist(payload *dty.DposCandidatorRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_CancelRegist method
func (d *DPos) ExecDelLocal_CancelRegist(payload *dty.DposCandidatorCancelRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_ReRegist method
func (d *DPos) ExecDelLocal_ReRegist(payload *dty.DposCandidatorRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_Vote method
func (d *DPos) ExecDelLocal_Vote(payload *dty.DposVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_CancelVote method
func (d *DPos) ExecDelLocal_CancelVote(payload *dty.DposCancelVote, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_VrfMRegist method
func (d *DPos) ExecDelLocal_VrfMRegist(payload *dty.DposVrfMRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_VrfRPRegist method
func (d *DPos) ExecDelLocal_VrfRPRegist(payload *dty.DposVrfRPRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}
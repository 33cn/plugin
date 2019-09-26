// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
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
	if log.VoteType == dty.VoteTypeVote {
		for i := 0; i < len(cand.Voters); i++ {
			if cand.Voters[i].Index == log.Vote.Index && cand.Voters[i].FromAddr == log.Vote.FromAddr && bytes.Equal(cand.Voters[i].Pubkey, log.Vote.Pubkey) {
				cand.Voters = append(cand.Voters[0:i], cand.Voters[i+1:]...)
				break
			}
		}
	} else if log.VoteType == dty.VoteTypeCancelVote {
		cand.Voters = append(cand.Voters, log.Vote)
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

		if log.VoteType == dty.VoteTypeVote {
			err = voterTable.Del([]byte(fmt.Sprintf("%018d", log.Index)))
			if err != nil {
				return nil, err
			}
		} else if log.VoteType == dty.VoteTypeCancelVote {
			err = voterTable.Add(log.Vote)
			if err != nil {
				return nil, err
			}
		}

		kvs2, err := voterTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	} else if log.Status == dty.CandidatorStatusCancelRegist {
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

		if log.VoteType == dty.VoteTypeCancelAllVote {
			for i := 0; i < len(candInfo.Voters); i++ {
				err = voterTable.Add(candInfo.Voters[i])
				if err != nil {
					return nil, err
				}
			}
		}

		kvs2, err := voterTable.Save()
		if err != nil {
			return nil, err
		}

		kvs = append(kvs1, kvs2...)
	} else if log.Status == dty.CandidatorStatusReRegist {
		//注册回滚,cand表删除记录
		err = candTable.Del(log.Pubkey)
		if err != nil {
			return nil, err
		}
		kvs, err = candTable.Save()
		return kvs, err
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

func (d *DPos) rollbackCBInfo(log *dty.ReceiptCB) (kvs []*types.KeyValue, err error) {
	if log.Status == dty.CBStatusRecord {
		cbTable := dty.NewDposCBTable(d.GetLocalDB())

		//注册回滚,cand表删除记录
		err = cbTable.Del([]byte(fmt.Sprintf("%018d", log.Cycle)))
		if err != nil {
			return nil, err
		}
		kvs, err = cbTable.Save()
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
		case dty.TyLogCandicatorRegist, dty.TyLogCandicatorVoted, dty.TyLogCandicatorCancelVoted, dty.TyLogCandicatorCancelRegist, dty.TyLogCandicatorReRegist:
			receiptLog := &dty.ReceiptCandicator{}
			if err := types.Decode(log.Log, receiptLog); err != nil {
				return nil, err
			}
			kv, err := d.rollbackCandVote(receiptLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kv...)

		case dty.TyLogVrfMRegist, dty.TyLogVrfRPRegist:
			receiptLog := &dty.ReceiptVrf{}
			if err := types.Decode(log.Log, receiptLog); err != nil {
				return nil, err
			}
			kv, err := d.rollbackVrf(receiptLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kv...)

		case dty.TyLogCBInfoRecord:
			receiptLog := &dty.ReceiptCB{}
			if err := types.Decode(log.Log, receiptLog); err != nil {
				return nil, err
			}
			kv, err := d.rollbackCBInfo(receiptLog)
			if err != nil {
				return nil, err
			}
			dbSet.KV = append(dbSet.KV, kv...)

		case dty.TyLogTopNCandidatorRegist:
			//do nothing now
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

//ExecDelLocal_RecordCB method
func (d *DPos) ExecDelLocal_RecordCB(payload *dty.DposCBInfo, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

//ExecDelLocal_RegistTopN method
func (d *DPos) ExecDelLocal_RegistTopN(payload *dty.TopNCandidatorRegist, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return d.execDelLocal(receiptData)
}

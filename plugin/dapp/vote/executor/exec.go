package executor

import (
	"github.com/33cn/chain33/types"
	votetypes "github.com/33cn/plugin/plugin/dapp/vote/types"
)

/*
 * 实现交易的链上执行接口
 * 关键数据上链（statedb）并生成交易回执（log）
 */

func (v *vote) Exec_CreateGroup(payload *votetypes.CreateGroup, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.createGroup(payload)
}

func (v *vote) Exec_UpdateGroup(payload *votetypes.UpdateGroup, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.updateGroup(payload)
}

func (v *vote) Exec_CreateVote(payload *votetypes.CreateVote, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.createVote(payload)
}

func (v *vote) Exec_CommitVote(payload *votetypes.CommitVote, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.commitVote(payload)
}

func (v *vote) Exec_CloseVote(payload *votetypes.CloseVote, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.closeVote(payload)
}

func (v *vote) Exec_UpdateMember(payload *votetypes.UpdateMember, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newAction(v, tx, index)
	return action.updateMember(payload)
}

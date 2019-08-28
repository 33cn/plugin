// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

//Exec_Regist DPos执行器注册候选节点
func (d *DPos) Exec_Regist(payload *dty.DposCandidatorRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.Regist(payload)
}

//Exec_CancelRegist DPos执行器取消注册候选节点
func (d *DPos) Exec_CancelRegist(payload *dty.DposCandidatorCancelRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.CancelRegist(payload)
}

//Exec_ReRegist DPos执行器重新注册候选节点
func (d *DPos) Exec_ReRegist(payload *dty.DposCandidatorRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.ReRegist(payload)
}

//Exec_Vote DPos执行器为候选节点投票
func (d *DPos) Exec_Vote(payload *dty.DposVote, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.Vote(payload)
}

//Exec_CancelVote DPos执行器撤销对一个候选节点的投票
func (d *DPos) Exec_CancelVote(payload *dty.DposCancelVote, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.CancelVote(payload)
}

//Exec_RegistVrfM DPos执行器注册一个受托节点的Vrf M信息
func (d *DPos) Exec_RegistVrfM(payload *dty.DposVrfMRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.RegistVrfM(payload)
}

//Exec_RegistVrfRP DPos执行器注册一个受托节点的Vrf R/P信息
func (d *DPos) Exec_RegistVrfRP(payload *dty.DposVrfRPRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.RegistVrfRP(payload)
}

//Exec_RecordCB DPos执行器记录CycleBoundary信息
func (d *DPos) Exec_RecordCB(payload *dty.DposCBInfo, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.RecordCB(payload)
}

//Exec_RegistTopN DPos执行器注册某一cycle中的TOPN信息
func (d *DPos) Exec_RegistTopN(payload *dty.TopNCandidatorRegist, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(d, tx, index)
	return action.RegistTopN(payload)
}

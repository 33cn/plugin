// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"

	"sort"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func getStageID(db dbm.KV, id string) (*pt.SelfConsensStageInfo, error) {
	realID := calcParaSelfConsensStageIDKey(id)
	val, err := getDb(db, []byte(realID))
	if err != nil {
		return nil, err
	}

	var status pt.SelfConsensStageInfo
	err = types.Decode(val, &status)
	return &status, err
}

func makeStageConfigReceipt(prev, current *pt.SelfConsensStageInfo) *types.Receipt {
	log := &pt.ReceiptSelfConsStageConfig{
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: []byte(current.Id), Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaSelfConsStageConfig,
				Log: types.Encode(log),
			},
		},
	}
}

func makeStageVoteDoneReceipt(config *pt.SelfConsensStage, totalCount, commitCount, most int, pass string) *types.Receipt {
	log := &pt.ReceiptSelfConsStageVoteDone{
		Stage:      config,
		TotalNodes: int32(totalCount),
		TotalVote:  int32(commitCount),
		MostVote:   int32(most),
		VoteRst:    pass,
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: nil,
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaNodeVoteDone,
				Log: types.Encode(log),
			},
		},
	}
}

func makeStageGroupReceipt(prev, current *pt.SelfConsensStages) *types.Receipt {
	key := calcParaSelfConsStagesKey()
	log := &pt.ReceiptSelfConsStagesUpdate{Prev: prev, Current: current}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaStageGroupUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func getSelfConsensStages(db dbm.KV) (*pt.SelfConsensStages, error) {
	key := calcParaSelfConsStagesKey()
	item, err := db.Get(key)
	if err != nil {
		if isNotFound(err) {
			err = pt.ErrKeyNotExist
		}
		return nil, errors.Wrapf(err, "getSelfConsensStages key:%s", string(key))
	}
	var config pt.SelfConsensStages
	err = types.Decode(item, &config)
	if err != nil {
		return nil, errors.Wrap(err, "getSelfConsensStages decode config")
	}
	return &config, nil
}

func getSelfConsStagesMap(stages []*pt.SelfConsensStage) map[int64]uint32 {
	stagesMap := make(map[int64]uint32)
	for _, v := range stages {
		stagesMap[v.BlockHeight] = v.Enable
	}
	return stagesMap
}

func sortStages(stages *pt.SelfConsensStages, new *pt.SelfConsensStage) {
	stages.Items = append(stages.Items, new)
	sort.Slice(stages.Items, func(i, j int) bool { return stages.Items[i].BlockHeight < stages.Items[j].BlockHeight })
}

func updateStages(db dbm.KV, stage *pt.SelfConsensStage) (*types.Receipt, error) {
	stages, err := getSelfConsensStages(db)
	if err != nil && errors.Cause(err) != pt.ErrKeyNotExist {
		return nil, err
	}
	if stages == nil {
		stages = &pt.SelfConsensStages{}
		stages.Items = append(stages.Items, stage)
		return makeStageGroupReceipt(nil, stages), nil
	}

	var old pt.SelfConsensStages
	err = deepCopy(&old, stages)
	if err != nil {
		clog.Error("updateStages  deep copy fail", "copy", old, "stat", stages)
		return nil, err
	}
	sortStages(stages, stage)
	return makeStageGroupReceipt(&old, stages), nil

}

func selfConsensInitStage(cfg *types.Chain33Config) *types.Receipt {
	close := cfg.IsEnable("consensus.sub.para.paraSelfConsInitDisable")
	stage := &pt.SelfConsensStage{BlockHeight: 0, Enable: pt.ParaConfigYes}
	if close {
		stage.Enable = pt.ParaConfigNo
	}
	stages := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{stage}}
	return makeStageGroupReceipt(nil, stages)
}

func getSelfConsOneStage(db dbm.KV, height int64) (*pt.SelfConsensStage, error) {
	stages, err := getSelfConsensStages(db)
	if err != nil {
		return nil, err
	}

	for i := len(stages.Items) - 1; i >= 0; i-- {
		if height >= stages.Items[i].BlockHeight {
			return stages.Items[i], nil
		}
	}
	return nil, errors.Wrapf(pt.ErrKeyNotExist, "SelfConsStage not found to height:%d", height)

}

func isSelfConsOn(db dbm.KV, height int64) (bool, error) {
	stage, err := getSelfConsOneStage(db, height)
	if err != nil {
		return false, err
	}
	return stage.Enable == pt.ParaConfigYes, nil
}

func (a *action) checkValidStage(config *pt.SelfConsensStage) error {
	cfg := a.api.GetConfig()
	//0. 设置高度必须大于fork高度
	if !cfg.IsDappFork(config.BlockHeight, pt.ParaX, pt.ForkParaSelfConsStages) {
		return errors.Wrapf(types.ErrNotAllow, "checkValidStage config height:%d less than fork height", config.BlockHeight)
	}

	//1. 设置高度必须大于当前区块高度
	if config.BlockHeight <= a.height {
		return errors.Wrapf(pt.ErrHeightHasPast, "checkValidStage config height:%d less than block height:%d", config.BlockHeight, a.height)
	}

	//2. 如果已经设置到stages中，简单起见，就不能更改了，应该也不会有很大影响
	stages, err := getSelfConsensStages(a.db)
	if err != nil && errors.Cause(err) != pt.ErrKeyNotExist {
		return errors.Wrapf(err, "checkValidStage get stages")
	}
	if stages != nil {
		stageMap := getSelfConsStagesMap(stages.Items)
		if _, exist := stageMap[config.BlockHeight]; exist {
			return errors.Wrapf(err, "checkValidStage config height:%d existed", config.BlockHeight)
		}
	}

	//3. enable check
	if config.Enable != pt.ParaConfigYes && config.Enable != pt.ParaConfigNo {
		return errors.Wrapf(err, "checkValidStage config enable:%d not support", config.Enable)
	}
	return nil
}

func (a *action) stageApply(stage *pt.SelfConsensStage) (*types.Receipt, error) {
	err := a.checkValidStage(stage)
	if err != nil {
		return nil, err
	}

	stat := &pt.SelfConsensStageInfo{
		Id:         calcParaSelfConsensStageIDKey(common.ToHex(a.txhash)),
		Status:     pt.ParaApplyJoining,
		Stage:      stage,
		FromAddr:   a.fromaddr,
		ExecHeight: a.height}
	return makeStageConfigReceipt(nil, stat), nil
}

func (a *action) stageCancel(config *pt.ConfigCancelInfo) (*types.Receipt, error) {
	stat, err := getStageID(a.db, config.Id)
	if err != nil {
		return nil, err
	}

	//只能提案发起人撤销
	if a.fromaddr != stat.FromAddr {
		return nil, errors.Wrapf(types.ErrNotAllow, "stage id create by:%s,not by:%s", stat.FromAddr, a.fromaddr)
	}

	if stat.Status != pt.ParaApplyJoining && stat.Status != pt.ParaApplyVoting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "stage config id:%s,status:%d", config.Id, stat.Status)
	}

	var copyStat pt.SelfConsensStageInfo
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("selfConsensQuit  deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}

	stat.Status = pt.ParaApplyCanceled
	stat.ExecHeight = a.height
	return makeStageConfigReceipt(&copyStat, stat), nil
}

func (a *action) stageVote(config *pt.ConfigVoteInfo) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	nodes, _, err := getParacrossNodes(a.db, cfg.GetTitle())
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", cfg.GetTitle())
	}
	if !validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "not validNode:%s", a.fromaddr)
	}

	stat, err := getStageID(a.db, config.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "stageVote getStageID:%s", config.Id)
	}

	if stat.Status != pt.ParaApplyJoining && stat.Status != pt.ParaApplyVoting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, stat.Status)
	}
	//stage blockHeight　也不能小于当前vote tx height,不然没有意义
	err = a.checkValidStage(stat.Stage)
	if err != nil {
		return nil, err
	}

	var copyStat pt.SelfConsensStageInfo
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("selfConsensVote deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}
	stat.Status = pt.ParaApplyVoting
	if stat.Votes == nil {
		stat.Votes = &pt.ParaNodeVoteDetail{}
	}
	found, index := hasVoted(stat.Votes.Addrs, a.fromaddr)
	if found {
		stat.Votes.Votes[index] = pt.ParaNodeVoteStr[config.Value]
	} else {
		stat.Votes.Addrs = append(stat.Votes.Addrs, a.fromaddr)
		stat.Votes.Votes = append(stat.Votes.Votes, pt.ParaNodeVoteStr[config.Value])
	}

	//剔除已退出nodegroup的addr的投票
	updateVotes(stat.Votes, nodes)

	most, vote := getMostVote(stat.Votes)
	if !isCommitDone(nodes, most) {
		return makeStageConfigReceipt(&copyStat, stat), nil
	}
	clog.Info("paracross.stageVote  ----pass", "most", most, "pass", vote)

	receipt := &types.Receipt{Ty: types.ExecOk}
	if vote == pt.ParaVoteYes {
		r, err := updateStages(a.db, stat.Stage)
		if err != nil {
			return nil, errors.Wrapf(err, "stageVote updateStages")
		}
		receipt = mergeReceipt(receipt, r)
	}
	stat.Status = pt.ParaApplyClosed
	stat.ExecHeight = a.height
	r := makeStageConfigReceipt(&copyStat, stat)
	receipt = mergeReceipt(receipt, r)

	r = makeStageVoteDoneReceipt(stat.Stage, len(nodes), len(stat.Votes.Addrs), most, pt.ParaNodeVoteStr[vote])
	receipt = mergeReceipt(receipt, r)
	return receipt, nil

}

//SelfConsensStageConfig support self consens stage config
func (a *action) SelfStageConfig(config *pt.ParaStageConfig) (*types.Receipt, error) {
	if config.OpTy == pt.ParaOpNewApply {
		return a.stageApply(config.GetStage())

	} else if config.OpTy == pt.ParaOpCancel {
		return a.stageCancel(config.GetCancel())

	} else if config.OpTy == pt.ParaOpVote {
		return a.stageVote(config.GetVote())
	}

	return nil, pt.ErrParaUnSupportNodeOper

}

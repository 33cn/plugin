// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func getNodeAddr(db dbm.KV, key []byte) (*pt.ParaNodeAddrStatus, error) {
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeAddrStatus
	err = types.Decode(val, &status)
	return &status, err
}

func saveNodeAddr(db dbm.KV, key []byte, status types.Message) error {
	// use as a types.Message
	val := types.Encode(status)
	return db.Set(key, val)
}

func makeVoteDoneReceipt(config *pt.ParaNodeAddrConfig, totalCount, commitCount, most int, ok bool, status int32) *types.Receipt {

	log := &pt.ReceiptParaNodeVoteDone{
		Title:      config.Title,
		TargetAddr: config.Addr,
		TotalNodes: int32(totalCount),
		TotalVote:  int32(commitCount),
		MostVote:   int32(most),
		VoteRst:    ok,
		DoneStatus: status,
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

func makeNodeConfigReceipt(addr string, config *pt.ParaNodeAddrConfig, prev, current *pt.ParaNodeAddrStatus) *types.Receipt {
	key := calcParaNodeAddrKey(config.Title, config.Addr)
	log := &pt.ReceiptParaNodeConfig{
		Addr:    addr,
		Config:  config,
		Prev:    prev,
		Current: current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaNodeConfig,
				Log: types.Encode(log),
			},
		},
	}
}

func makeParaNodeGroupReiceipt(title string, prev, current *types.ConfigItem) *types.Receipt {
	key := calcParaNodeGroupKey(title)
	log := &types.ReceiptConfig{Prev: prev, Current: current}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaNodeGroupUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func (a *action) nodeAdd(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	key := calcParaNodeGroupKey(config.Title)
	nodes, _, err := getNodes(a.db, key)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	if validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeAddr existed:%s", a.fromaddr)
	}

	key = calcParaNodeAddrKey(config.Title, config.Addr)
	stat, err := getNodeAddr(a.db, key)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		clog.Info("first time add node addr", "key", string(key))
		stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeAdding,
			Title:     config.Title,
			ApplyAddr: config.Addr,
			Votes:     &pt.ParaNodeVoteDetail{}}
		saveNodeAddr(a.db, key, stat)
		return makeNodeConfigReceipt(a.fromaddr, config, nil, stat), nil
	}

	copyStatus := *stat
	if stat.Status != pt.ParacrossNodeQuited {
		clog.Error("nodeaccount.nodeAdd key exist", "key", string(key), "status", stat)
		return nil, pt.ErrParaNodeAddrExisted
	}
	stat.Status = pt.ParacrossNodeAdding
	stat.Votes = &pt.ParaNodeVoteDetail{}
	saveNodeAddr(a.db, key, stat)
	return makeNodeConfigReceipt(a.fromaddr, config, &copyStatus, stat), nil

}

func (a *action) nodeDelete(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	key := calcParaNodeGroupKey(config.Title)
	nodes, _, err := getNodes(a.db, key)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	if !validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "nodeAddr not existed:%s", a.fromaddr)
	}
	//不允许最后一个账户退出
	if len(nodes) == 1 {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupLastAddr, "nodeAddr last one:%s", a.fromaddr)
	}

	key = calcParaNodeAddrKey(config.Title, config.Addr)
	stat, err := getNodeAddr(a.db, key)
	if err != nil {
		return nil, err
	}
	//refused or quiting
	if stat.Status != pt.ParacrossNodeAdded {
		clog.Error("nodeaccount.nodeDelete wrong status", "key", string(key), "status", stat)
		return nil, errors.Wrapf(pt.ErrParaNodeGroupRefuseByVote, "nodeAddr refused by vote:%s", a.fromaddr)
	}
	copyStat := *stat
	stat.Status = pt.ParacrossNodeQuiting
	stat.Votes = &pt.ParaNodeVoteDetail{}
	saveNodeAddr(a.db, key, stat)
	return makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat), nil

}

func getMostVote(stat *pt.ParaNodeAddrStatus) (int, bool) {
	var ok, nok int
	for _, v := range stat.GetVotes().Votes {
		if v == pt.ParaNodeVotePass {
			ok++
		} else {
			nok++
		}
	}
	if ok > nok {
		return ok, true
	}
	return nok, false

}

func hasVoted(addrs []string, addr string) (bool, int) {
	return hasCommited(addrs, addr)
}

func (a *action) nodeVote(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	key := calcParaNodeGroupKey(config.Title)
	nodes, _, err := getNodes(a.db, key)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	if !validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "not validNode:%s", a.fromaddr)
	}

	if a.fromaddr == config.Addr {
		return nil, errors.Wrapf(pt.ErrParaNodeVoteSelf, "not allow to vote self:%s", a.fromaddr)
	}

	// 如果投票账户是group账户，需计算此账户之外的投票
	if validNode(config.Addr, nodes) {
		temp := make(map[string]struct{})
		for k := range nodes {
			if k != config.Addr {
				temp[k] = struct{}{}
			}
		}
		nodes = temp
	}

	key = calcParaNodeAddrKey(config.Title, config.Addr)
	stat, err := getNodeAddr(a.db, key)
	if err != nil {
		return nil, err
	}

	copyStat := *stat
	if stat.Votes == nil {
		stat.Votes = &pt.ParaNodeVoteDetail{}
	}
	found, index := hasVoted(stat.Votes.Addrs, a.fromaddr)
	if found {
		stat.Votes.Votes[index] = config.Value
	} else {
		stat.Votes.Addrs = append(stat.Votes.Addrs, a.fromaddr)
		stat.Votes.Votes = append(stat.Votes.Votes, config.Value)
	}
	receipt := makeNodeConfigReceipt(config.Addr, config, &copyStat, stat)
	most, ok := getMostVote(stat)
	if !isCommitDone(stat, nodes, most) {
		saveNodeAddr(a.db, key, stat)
		return receipt, nil
	}
	clog.Info("paracross.nodeVote commit ----pass", "most", most, "pass", ok)

	var receiptGroup *types.Receipt
	if !ok {
		stat.Status = pt.ParacrossNodeRefused
	} else {
		if stat.Status == pt.ParacrossNodeAdding {
			receiptGroup, err = unpdateNodeGroup(a.db, config.Title, config.Addr, true)
			if err != nil {
				return nil, err
			}
			stat.Status = pt.ParacrossNodeAdded
		} else if stat.Status == pt.ParacrossNodeQuiting {
			receiptGroup, err = unpdateNodeGroup(a.db, config.Title, config.Addr, false)
			if err != nil {
				return nil, err
			}
			stat.Status = pt.ParacrossNodeQuited
		}
	}

	saveNodeAddr(a.db, key, stat)
	receipt = makeNodeConfigReceipt(config.Addr, config, &copyStat, stat)
	if receiptGroup != nil {
		receipt.KV = append(receipt.KV, receiptGroup.KV...)
		receipt.Logs = append(receipt.Logs, receiptGroup.Logs...)
	}
	receiptDone := makeVoteDoneReceipt(config, len(nodes), len(stat.Votes.Addrs), most, ok, stat.Status)
	receipt.KV = append(receipt.KV, receiptDone.KV...)
	receipt.Logs = append(receipt.Logs, receiptDone.Logs...)
	return receipt, nil

}

func unpdateNodeGroup(db dbm.KV, title, addr string, add bool) (*types.Receipt, error) {
	var item types.ConfigItem

	key := calcParaNodeGroupKey(title)
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			clog.Error("unpdateNodeGroup", "decode db key", key)
			return nil, err // types.ErrBadConfigValue
		}
	}

	copyValue := *item.GetArr()
	copyItem := item
	copyItem.Value = &types.ConfigItem_Arr{Arr: &copyValue}

	if add {
		item.GetArr().Value = append(item.GetArr().Value, addr)
		item.Addr = addr
		clog.Info("unpdateNodeGroup", "add key", key, "from", copyItem.GetArr().Value, "to", item.GetArr().Value)

	} else {
		//必须保留至少1个授权账户
		if len(item.GetArr().Value) <= 1 {
			return nil, pt.ErrParaNodeGroupLastAddr
		}
		item.Addr = addr
		item.GetArr().Value = make([]string, 0)
		for _, value := range copyItem.GetArr().Value {
			clog.Info("unpdateNodeGroup", "key delete", key, "current", value)
			if value != addr {
				item.GetArr().Value = append(item.GetArr().Value, value)
			}
		}
	}

	return makeParaNodeGroupReiceipt(title, &copyItem, &item), nil
}

func (a *action) nodeTakeover(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	key := calcManageConfigNodesKey(config.Title)
	_, nodes, err := getNodes(a.db, key)
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, pt.ErrParaManageNodesNotSet
	}

	var item types.ConfigItem
	key = calcParaNodeGroupKey(config.Title)
	value, err := a.db.Get(key)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if value != nil {
		return nil, pt.ErrParaNodeGroupExisted
	}

	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr

	copyItem := item
	item.GetArr().Value = append(item.GetArr().Value, nodes...)
	item.Addr = a.fromaddr
	a.db.Set(key, types.Encode(&item))
	receipt := makeParaNodeGroupReiceipt(config.Title, &copyItem, &item)

	//update add addr
	for _, addr := range nodes {
		key = calcParaNodeAddrKey(config.Title, addr)
		stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeAdded,
			Title:     config.Title,
			ApplyAddr: addr,
			Votes:     &pt.ParaNodeVoteDetail{}}
		saveNodeAddr(a.db, key, stat)
		config.Addr = addr
		r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}
	return receipt, nil
}

func (a *action) NodeConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	if !validTitle(config.Title) {
		return nil, pt.ErrInvalidTitle
	}

	if config.Op == pt.ParaNodeAdd {
		if config.Addr != a.fromaddr {
			return nil, types.ErrFromAddr
		}
		return a.nodeAdd(config)

	} else if config.Op == pt.ParaNodeDelete {
		if config.Addr != a.fromaddr {
			return nil, types.ErrFromAddr
		}
		return a.nodeDelete(config)

	} else if config.Op == pt.ParaNodeVote {
		return a.nodeVote(config)
	} else if config.Op == pt.ParaNodeTakeover {
		return a.nodeTakeover(config)

	} else {
		return nil, pt.ErrParaUnSupportNodeOper
	}

}

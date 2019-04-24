// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/gob"

	"strings"

	dbm "github.com/33cn/chain33/common/db"
	manager "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

var (
	confManager = types.ConfSub(manager.ManageX)
	conf        = types.ConfSub(pt.ParaX)
)

func deepCopy(dst, src interface{}) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(src); err != nil {
		return err
	}
	return gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(dst)
}

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
	val := types.Encode(status)
	return db.Set(key, val)
}

func makeVoteDoneReceipt(config *pt.ParaNodeAddrConfig, totalCount, commitCount, most int, pass string, status int32) *types.Receipt {
	log := &pt.ReceiptParaNodeVoteDone{
		Title:      config.Title,
		TargetAddr: config.Addr,
		TotalNodes: int32(totalCount),
		TotalVote:  int32(commitCount),
		MostVote:   int32(most),
		VoteRst:    pass,
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

func makeParaNodeGroupApplyReiceipt(title, addr string, prev, current *pt.ParaNodeAddrStatus, logTy int32) *types.Receipt {
	key := calcParaNodeGroupApplyKey(title)
	log := &pt.ReceiptParaNodeConfig{
		Addr:    addr,
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
				Ty:  logTy,
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

func (a *action) nodeJoin(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	key := calcParaNodeGroupKey(config.Title)
	nodes, _, err := getNodes(a.db, key)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	if validNode(a.fromaddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeAddr existed:%s", a.fromaddr)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsFrozen([]string{a.fromaddr}, config.CoinsFrozen)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	key = calcParaNodeAddrKey(config.Title, config.Addr)
	stat, err := getNodeAddr(a.db, key)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		clog.Info("first time add node addr", "key", string(key))
		stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeAdding,
			Title:       config.Title,
			ApplyAddr:   config.Addr,
			Votes:       &pt.ParaNodeVoteDetail{},
			CoinsFrozen: config.CoinsFrozen}
		saveNodeAddr(a.db, key, stat)
		r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
		return receipt, nil
	}

	var copyStat pt.ParaNodeAddrStatus
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("nodeaccount.nodeJoin deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}
	if stat.Status != pt.ParacrossNodeQuited {
		clog.Error("nodeaccount.nodeJoin key exist", "key", string(key), "status", stat)
		return nil, pt.ErrParaNodeAddrExisted
	}
	stat.Status = pt.ParacrossNodeAdding
	stat.CoinsFrozen = config.CoinsFrozen
	stat.Votes = &pt.ParaNodeVoteDetail{}
	saveNodeAddr(a.db, key, stat)
	r := makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)
	return receipt, nil

}

func (a *action) nodeQuit(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
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

	if stat.Status == pt.ParacrossNodeQuiting || stat.Status == pt.ParacrossNodeQuited {
		clog.Error("nodeaccount.nodeQuit wrong status", "key", string(key), "status", stat)
		return nil, errors.Wrapf(pt.ErrParaUnSupportNodeOper, "nodeAddr %s was quit status:%d", a.fromaddr, stat.Status)
	}
	var copyStat pt.ParaNodeAddrStatus
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("nodeaccount.nodeQuit deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}
	if stat.Status == pt.ParacrossNodeAdded {
		stat.Status = pt.ParacrossNodeQuiting
		stat.Votes = &pt.ParaNodeVoteDetail{}
		saveNodeAddr(a.db, key, stat)
		return makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat), nil
	}

	//still adding status, quit directly
	receipt := &types.Receipt{Ty: types.ExecOk}
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsActive([]string{a.fromaddr}, stat.CoinsFrozen)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	stat.Status = pt.ParacrossNodeQuited
	stat.Votes = &pt.ParaNodeVoteDetail{}
	saveNodeAddr(a.db, key, stat)
	r := makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil

}

// IsSuperManager is supper manager or not
func isSuperManager(addr string) bool {
	for _, m := range confManager.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

func getMostVote(stat *pt.ParaNodeAddrStatus) (int, string) {
	var ok, nok int
	for _, v := range stat.GetVotes().Votes {
		if v == pt.ParaNodeVoteYes {
			ok++
		} else {
			nok++
		}
	}
	if ok > nok {
		return ok, pt.ParaNodeVoteYes
	}
	return nok, pt.ParaNodeVoteNo

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
	if !validNode(a.fromaddr, nodes) && !isSuperManager(a.fromaddr) {
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

	var copyStat pt.ParaNodeAddrStatus
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("nodeaccount.nodevOTE deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}

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
	receipt := makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
	most, vote := getMostVote(stat)
	if !isCommitDone(stat, nodes, most) {
		//超级用户投yes票，就可以通过，防止当前所有授权节点都忘掉私钥场景
		//超级用户且当前group里面有任一账户投yes票也可以通过是备选方案 （most >1)即可
		if !(isSuperManager(a.fromaddr) && most > 0 && vote == pt.ParaNodeVoteYes) {
			saveNodeAddr(a.db, key, stat)
			return receipt, nil
		}
	}
	clog.Info("paracross.nodeVote  ----pass", "most", most, "pass", vote)

	var receiptGroup *types.Receipt
	if vote == pt.ParaNodeVoteNo {
		// 对已经在group里面的node，直接投票remove，对正在申请中的adding or quiting状态保持不变，对quited的保持不变
		if stat.Status == pt.ParacrossNodeAdded {
			receiptGroup, err = unpdateNodeGroup(a.db, config.Title, config.Addr, false)
			if err != nil {
				return nil, err
			}
			stat.Status = pt.ParacrossNodeQuited
		}
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

			if !types.IsPara() {
				r, err := a.nodeGroupCoinsActive([]string{stat.ApplyAddr}, stat.CoinsFrozen)
				if err != nil {
					return nil, err
				}
				receipt.KV = append(receipt.KV, r.KV...)
				receipt.Logs = append(receipt.Logs, r.Logs...)
			}
		}
	}
	saveNodeAddr(a.db, key, stat)
	receipt = makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
	if receiptGroup != nil {
		receipt.KV = append(receipt.KV, receiptGroup.KV...)
		receipt.Logs = append(receipt.Logs, receiptGroup.Logs...)
	}
	receiptDone := makeVoteDoneReceipt(config, len(nodes), len(stat.Votes.Addrs), most, vote, stat.Status)
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
			clog.Info("unpdateNodeGroup", "key delete", string(key), "current", value)
			if value != addr {
				item.GetArr().Value = append(item.GetArr().Value, value)
			}
		}
	}

	return makeParaNodeGroupReiceipt(title, &copyItem, &item), nil
}

func (a *action) checkConfig(title string) error {
	if !validTitle(title) {
		return pt.ErrInvalidTitle
	}

	forkHeight := types.GetDappFork(pt.ParaX, pt.ForkCommitTx)
	if types.IsPara() {
		forkHeight = types.Conf("config.consensus.sub.para").GInt("MainForkParacrossCommitTx")
		if forkHeight == -1 || forkHeight == 0 {
			forkHeight = types.MaxHeight
		}
	}
	if a.exec.GetMainHeight() < forkHeight {
		return types.ErrNotSupport
	}

	return nil
}

//NodeConfig support super account node config
func (a *action) NodeConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	err := a.checkConfig(config.Title)
	if err != nil {
		return nil, err
	}

	if config.Op == pt.ParaNodeJoin {
		if config.Addr != a.fromaddr {
			return nil, types.ErrFromAddr
		}
		return a.nodeJoin(config)

	} else if config.Op == pt.ParaNodeQuit {
		if config.Addr != a.fromaddr {
			return nil, types.ErrFromAddr
		}
		return a.nodeQuit(config)

	} else if config.Op == pt.ParaNodeVote {
		return a.nodeVote(config)
	} else {
		return nil, pt.ErrParaUnSupportNodeOper
	}

}

//NodeConfig support super account node config
func (a *action) NodeGroupConfig(config *pt.ParaNodeGroupApply) (*types.Receipt, error) {
	err := a.checkConfig(config.Title)
	if err != nil {
		return nil, err
	}

	if len(config.Addrs) == 0 {
		return nil, types.ErrInvalidParam
	}

	if config.Op == pt.ParacrossNodeGroupApply {
		if !strings.Contains(config.Addrs, a.fromaddr) {
			clog.Error("node group apply fromaddr not one of apply addrs", "addr", a.fromaddr, "apply", config.Addrs)
			return nil, types.ErrNotAllow
		}
		err := a.checkNodeGroupExist(config.Title)
		if err != nil {
			return nil, err
		}
		return a.nodeGroupApply(config)

	} else if config.Op == pt.ParacrossNodeGroupApprove {
		err := a.checkNodeGroupExist(config.Title)
		if err != nil {
			return nil, err
		}
		return a.nodeGroupApprove(config)

	} else if config.Op == pt.ParacrossNodeGroupQuit {
		if !strings.Contains(config.Addrs, a.fromaddr) {
			clog.Error("node group apply fromaddr not one of apply addrs", "addr", a.fromaddr, "apply", config.Addrs)
			return nil, types.ErrNotAllow
		}
		return a.nodeGroupQuit(config)
	} else {
		return nil, pt.ErrParaUnSupportNodeOper
	}

}

func getAddrGroup(addr string) []string {
	if strings.Contains(addr, ",") {
		repeats := make(map[string]bool)
		var addrs []string

		s := strings.Trim(addr, " ")
		s = strings.Trim(s, ",")
		ss := strings.Split(s, ",")
		for _, v := range ss {
			v = strings.Trim(v, " ")
			if !repeats[v] {
				addrs = append(addrs, v)
				repeats[v] = true
			}
		}
		return addrs
	}

	return []string{addr}
}

func (a *action) checkNodeGroupExist(title string) error {
	key := calcParaNodeGroupKey(title)
	value, err := a.db.Get(key)
	if err != nil && !isNotFound(err) {
		return err
	}
	if value != nil {
		clog.Error("node group apply, group existed")
		return pt.ErrParaNodeGroupExisted
	}

	return nil
}

func (a *action) nodeGroupCoinsFrozen(addrs []string, configCoinsFrozen int64) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	confCoins := confManager.GInt("nodeGroupFrozenCoins")
	if configCoinsFrozen < confCoins {
		return nil, pt.ErrParaNodeGroupFrozenCoinsNotEnough
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	for _, addr := range addrs {
		receipt, err := a.coinsAccount.ExecFrozen(addr, a.execaddr, configCoinsFrozen)
		if err != nil {
			clog.Error("node group apply", "addr", addr, "execaddr", a.execaddr, "amount", configCoinsFrozen)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	receipt.KV = append(receipt.KV, kv...)
	receipt.Logs = append(receipt.Logs, logs...)

	return receipt, nil
}

func (a *action) nodeGroupCoinsActive(addrs []string, configCoinsFrozen int64) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	for _, addr := range addrs {
		receipt, err := a.coinsAccount.ExecActive(addr, a.execaddr, configCoinsFrozen)
		if err != nil {
			clog.Error("node group apply", "addr", addr, "execaddr", a.execaddr, "amount", configCoinsFrozen)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	receipt.KV = append(receipt.KV, kv...)
	receipt.Logs = append(receipt.Logs, logs...)

	return receipt, nil
}

// NodeGroupApply
func (a *action) nodeGroupApply(config *pt.ParaNodeGroupApply) (*types.Receipt, error) {
	key := calcParaNodeGroupApplyKey(title)
	status, err := getNodeAddr(a.db, key)
	if err != nil && !isNotFound(err) {
		return nil, err
	}
	if status != nil && status.Status != pt.ParacrossNodeGroupQuit {
		clog.Error("node group apply exist", "status", status.Status)
		return nil, pt.ErrParaNodeGroupExisted
	}

	addrs := getAddrGroup(config.Addrs)
	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsFrozen(addrs, config.CoinsFrozen)
		if err != nil {
			return nil, err
		}

		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeGroupApply,
		Title:       config.Title,
		ApplyAddr:   strings.Join(addrs, ","),
		CoinsFrozen: config.CoinsFrozen}
	saveNodeAddr(a.db, key, stat)
	r := makeParaNodeGroupApplyReiceipt(config.Title, a.fromaddr, nil, stat, pt.TyLogParaNodeGroupApply)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupQuit(config *pt.ParaNodeGroupApply) (*types.Receipt, error) {
	key := calcParaNodeGroupApplyKey(title)
	status, err := getNodeAddr(a.db, key)
	if err != nil {
		return nil, err
	}
	if status != nil && status.Status != pt.ParacrossNodeGroupApply {
		clog.Error("node group apply exist", "status", status.Status)
		return nil, pt.ErrParaNodeGroupStatusWrong
	}

	applyAddrs, err := checkNodeGroupAddrsMatch(status.ApplyAddr, config.Addrs)
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsActive(applyAddrs, status.CoinsFrozen)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeGroupQuit,
		Title:       config.Title,
		ApplyAddr:   status.ApplyAddr,
		CoinsFrozen: status.CoinsFrozen}
	saveNodeAddr(a.db, key, stat)
	r := makeParaNodeGroupApplyReiceipt(config.Title, a.fromaddr, nil, stat, pt.TyLogParaNodeGroupQuit)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func checkNodeGroupAddrsMatch(applyAddr, confAddr string) ([]string, error) {
	confAddrs := getAddrGroup(confAddr)
	applyAddrs := strings.Split(applyAddr, ",")

	applys := make(map[string]bool)
	configs := make(map[string]bool)
	diff := []string{}
	for _, addr := range applyAddrs {
		applys[addr] = true
	}
	for _, addr := range confAddrs {
		configs[addr] = true
	}
	if len(applys) != len(configs) {
		clog.Error("node group addrs count not match", "apply", applyAddr, "quit", confAddr)
		return nil, pt.ErrParaNodeGroupAddrNotMatch
	}

	for _, addr := range confAddrs {
		if !applys[addr] {
			diff = append(diff, addr)
		}
	}
	if len(diff) > 0 {
		clog.Error("node group addrs not match", "apply", applyAddr, "quit", confAddr)
		return nil, pt.ErrParaNodeGroupAddrNotMatch
	}
	return confAddrs, nil

}

// NodeGroupApprove super addr approve the node group apply
func (a *action) nodeGroupApprove(config *pt.ParaNodeGroupApply) (*types.Receipt, error) {
	if !isSuperManager(a.fromaddr) {
		clog.Error("node group approve not super manager", "addr", a.fromaddr)
		return nil, types.ErrNotAllow
	}

	key := calcParaNodeGroupApplyKey(title)
	status, err := getNodeAddr(a.db, key)
	if err != nil {
		return nil, err
	}
	if status.Status != pt.ParacrossNodeGroupApply {
		clog.Error("node group approve not apply status", "status", status.Status)
		return nil, pt.ErrParaNodeGroupStatusWrong
	}

	applyAddrs, err := checkNodeGroupAddrsMatch(status.ApplyAddr, config.Addrs)
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//create the node group
	r, err := a.nodeGroupCreate(config.Title, applyAddrs, config.CoinsFrozen)
	if err != nil {
		return nil, err
	}
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeGroupApprove,
		Title:       config.Title,
		ApplyAddr:   status.ApplyAddr,
		CoinsFrozen: status.CoinsFrozen}
	saveNodeAddr(a.db, key, stat)
	r = makeParaNodeGroupApplyReiceipt(config.Title, a.fromaddr, status, stat, pt.TyLogParaNodeGroupApprove)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupCreate(title string, nodes []string, coinFrozen int64) (*types.Receipt, error) {
	var item types.ConfigItem
	key := calcParaNodeGroupKey(title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr

	copyItem := item
	item.GetArr().Value = append(item.GetArr().Value, nodes...)
	item.Addr = a.fromaddr
	a.db.Set(key, types.Encode(&item))
	receipt := makeParaNodeGroupReiceipt(title, &copyItem, &item)

	//update addr status
	for _, addr := range nodes {
		key = calcParaNodeAddrKey(title, addr)
		stat := &pt.ParaNodeAddrStatus{Status: pt.ParacrossNodeAdded,
			Title:       title,
			ApplyAddr:   addr,
			Votes:       &pt.ParaNodeVoteDetail{},
			CoinsFrozen: coinFrozen}
		saveNodeAddr(a.db, key, stat)
		config := &pt.ParaNodeAddrConfig{
			Title:       title,
			Addr:        addr,
			CoinsFrozen: coinFrozen,
		}
		r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}
	return receipt, nil
}

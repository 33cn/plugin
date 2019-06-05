// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/gob"

	"strings"

	"strconv"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
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

func getNodeAddr(db dbm.KV, title, addr string) (*pt.ParaNodeAddrIdStatus, error) {
	key := calcParaNodeAddrKey(title, addr)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeAddrIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

func getNodeID(db dbm.KV, id string) (*pt.ParaNodeIdStatus, error) {
	val, err := getDb(db, []byte(id))
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

func getNodeGroupStatus(db dbm.KV, title string) (*pt.ParaNodeGroupStatus, error) {
	key := calcParaNodeGroupStatusKey(title)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeGroupStatus
	err = types.Decode(val, &status)
	return &status, err
}

func getDb(db dbm.KV, key []byte) ([]byte, error) {
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func getNodeGroupID(db dbm.KV, id string) (*pt.ParaNodeGroupStatus, error) {
	val, err := getDb(db, []byte(id))
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeGroupStatus
	err = types.Decode(val, &status)
	return &status, err
}

func makeVoteDoneReceipt(config *pt.ParaNodeIdStatus, totalCount, commitCount, most int, pass string, status int32) *types.Receipt {
	log := &pt.ReceiptParaNodeVoteDone{
		Title:      config.Title,
		TargetAddr: config.TargetAddr,
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

func makeParaNodeStatusReceipt(fromAddr string, prev, current *pt.ParaNodeAddrIdStatus) *types.Receipt {
	key := calcParaNodeAddrKey(current.Title, current.Addr)
	log := &pt.ReceiptParaNodeAddrStatUpdate{
		FromAddr: fromAddr,
		Prev:     prev,
		Current:  current,
	}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaNodeStatusUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func makeNodeConfigReceipt(fromAddr string, config *pt.ParaNodeAddrConfig, prev, current *pt.ParaNodeIdStatus) *types.Receipt {
	log := &pt.ReceiptParaNodeConfig{
		Addr:    fromAddr,
		Config:  config,
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
				Ty:  pt.TyLogParaNodeConfig,
				Log: types.Encode(log),
			},
		},
	}
}

func makeNodeGroupIDReceipt(addr string, prev, current *pt.ParaNodeGroupStatus) *types.Receipt {
	log := &pt.ReceiptParaNodeGroupConfig{
		Addr:    addr,
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
				Ty:  pt.TyLogParaNodeGroupConfig,
				Log: types.Encode(log),
			},
		},
	}
}

func makeParaNodeGroupStatusReceipt(title string, addr string, prev, current *pt.ParaNodeGroupStatus) *types.Receipt {
	key := calcParaNodeGroupStatusKey(title)
	log := &pt.ReceiptParaNodeGroupConfig{
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
				Ty:  pt.TyLogParaNodeGroupStatusUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func makeParaNodeGroupReceipt(title string, prev, current *types.ConfigItem) *types.Receipt {
	key := calcParaNodeGroupAddrsKey(title)
	log := &types.ReceiptConfig{Prev: prev, Current: current}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaNodeGroupAddrsUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func (a *action) checkValidNode(config *pt.ParaNodeAddrConfig) (bool, error) {
	nodes, _, err := getParacrossNodes(a.db, config.Title)
	if err != nil {
		return false, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	//有可能申请地址和配置地址不是同一个
	if validNode(config.Addr, nodes) {
		return true, nil
	}
	return false, nil
}

func (a *action) nodeJoin(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	addrExist, err := a.checkValidNode(config)
	if err != nil {
		return nil, err
	}
	if addrExist {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeAddr existed:%s", config.Addr)
	}

	nodeGroupStatus, err := getNodeGroupStatus(a.db, config.Title)
	if err != nil {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupNotSet, "nodegroup not exist:%s", config.Title)
	}

	if config.CoinsFrozen < nodeGroupStatus.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough,
			"coinFrozen not enough:%d,expected:%d", config.CoinsFrozen, nodeGroupStatus.CoinsFrozen)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsFrozen(a.fromaddr, config.CoinsFrozen, 1)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	addrStat, err := getNodeAddr(a.db, config.Title, config.Addr)
	if err != nil {
		if !isNotFound(err) {
			return nil, err
		}
		clog.Info("first time add node addr", "title", config.Title, "addr", config.Addr)
		stat := &pt.ParaNodeIdStatus{
			Id:          calcParaNodeIDKey(config.Title, common.ToHex(a.txhash)),
			Status:      pt.ParacrossNodeJoining,
			Title:       config.Title,
			TargetAddr:  config.Addr,
			FromAddr:    a.fromaddr,
			Votes:       &pt.ParaNodeVoteDetail{},
			CoinsFrozen: config.CoinsFrozen,
			Height:      a.height}
		r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
		return receipt, nil
	}

	if addrStat.Status == pt.ParacrossNodeQuited {
		stat := &pt.ParaNodeIdStatus{
			Id:          calcParaNodeIDKey(config.Title, common.ToHex(a.txhash)),
			Status:      pt.ParacrossNodeJoining,
			Title:       config.Title,
			TargetAddr:  config.Addr,
			FromAddr:    a.fromaddr,
			Votes:       &pt.ParaNodeVoteDetail{},
			CoinsFrozen: config.CoinsFrozen,
			Height:      a.height}
		r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
		return receipt, nil

	}
	return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeAddr existed:%s,status:%d", config.Addr, addrStat.Status)

}

func (a *action) nodeQuit(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	addrExist, err := a.checkValidNode(config)
	if err != nil {
		return nil, err
	}
	if !addrExist {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "nodeAddr not existed:%s", config.Addr)
	}

	addrStat, err := getNodeAddr(a.db, config.Title, config.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "nodeAddr:%s get error", config.Addr)
	}
	if addrStat.Status != pt.ParacrossNodeJoined {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "nodeAddr:%s status:%d", config.Addr, addrStat.Status)
	}

	stat := &pt.ParaNodeIdStatus{
		Id:         calcParaNodeIDKey(config.Title, common.ToHex(a.txhash)),
		Status:     pt.ParacrossNodeQuiting,
		Title:      config.Title,
		TargetAddr: config.Addr,
		FromAddr:   a.fromaddr,
		Votes:      &pt.ParaNodeVoteDetail{},
		Height:     a.height}
	return makeNodeConfigReceipt(a.fromaddr, config, nil, stat), nil

}

func (a *action) nodeCancel(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	stat, err := getNodeID(a.db, config.Id)
	if err != nil {
		return nil, err
	}

	//只能提案发起人撤销
	if a.fromaddr != stat.FromAddr {
		return nil, errors.Wrapf(types.ErrNotAllow, "id create by:%s,not by:%s", stat.FromAddr, a.fromaddr)
	}

	if config.Title != stat.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, stat.Title)
	}

	if stat.Status != pt.ParacrossNodeJoining && stat.Status != pt.ParacrossNodeQuiting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, stat.Status)
	}

	var copyStat pt.ParaNodeIdStatus
	err = deepCopy(&copyStat, stat)
	if err != nil {
		clog.Error("nodeaccount.nodeQuit deep copy fail", "copy", copyStat, "stat", stat)
		return nil, err
	}
	if stat.Status == pt.ParacrossNodeJoining {
		receipt := &types.Receipt{Ty: types.ExecOk}
		if !types.IsPara() {
			r, err := a.nodeGroupCoinsActive(stat.FromAddr, stat.CoinsFrozen, 1)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)
		}
		stat.Status = pt.ParacrossNodeCanceled
		stat.Height = a.height
		r := makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
		receipt = mergeReceipt(receipt, r)
		return receipt, nil
	}

	if stat.Status == pt.ParacrossNodeQuiting {
		stat.Status = pt.ParacrossNodeCanceled
		stat.Height = a.height
		return makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat), nil
	}

	return nil, errors.Wrapf(pt.ErrParaUnSupportNodeOper, "nodeid %s was quit status:%d", config.Id, stat.Status)

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

func getMostVote(stat *pt.ParaNodeIdStatus) (int, int) {
	var ok, nok int
	for _, v := range stat.GetVotes().Votes {
		if v == pt.ParaNodeVoteStr[pt.ParaNodeVoteYes] {
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

//主链配置平行链停止块数， 反应到主链上为对应平行链空块间隔×停止块数，如果主链当前高度超过平行链共识高度对应主链高度后面这个主链块数就表示通过
func (a *action) superManagerVoteProc(title string) error {
	status, err := getNodeGroupStatus(a.db, title)
	if err != nil {
		return err
	}

	confStopBlocks := conf.GInt("paraConsensusStopBlocks")
	data, err := a.exec.paracrossGetHeight(title)
	if err != nil {
		clog.Info("paracross.superManagerVoteProc get consens height", "title", title, "err", err.Error())
		return err
	}
	var consensMainHeight int64
	consensHeight := data.(*pt.ParacrossStatus).Height
	//如果group建立后一直没有共识，则从approve时候开始算
	if consensHeight == -1 {
		consensMainHeight = status.Height
	} else {
		stat, err := a.exec.paracrossGetStateTitleHeight(title, consensHeight)
		if err != nil {
			clog.Info("paracross.superManagerVoteProc get consens title height", "title", title, "conesusHeight", consensHeight, "err", err.Error())
			return err
		}
		consensMainHeight = stat.(*pt.ParacrossHeightStatus).MainHeight
	}
	//return err to stop tx pass to para chain
	if a.height <= consensMainHeight+confStopBlocks {
		return errors.Wrapf(pt.ErrParaConsensStopBlocksNotReach,
			"supermanager height not reach,current:%d less consens:%d plus confStopBlocks:%d", a.height, consensMainHeight, confStopBlocks)
	}

	return nil
}

func updateVotes(stat *pt.ParaNodeIdStatus, nodes map[string]struct{}) {
	votes := &pt.ParaNodeVoteDetail{}
	for i, addr := range stat.Votes.Addrs {
		if _, ok := nodes[addr]; ok {
			votes.Addrs = append(votes.Addrs, addr)
			votes.Votes = append(votes.Votes, stat.Votes.Votes[i])
		}
	}
	stat.Votes = votes
}

func (a *action) updateNodeAddrStatus(stat *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	addrStat, err := getNodeAddr(a.db, stat.Title, stat.TargetAddr)
	if err != nil {
		if !isNotFound(err) {
			return nil, errors.Wrapf(err, "nodeAddr:%s get error", stat.TargetAddr)
		}
		addrStat = &pt.ParaNodeAddrIdStatus{}
		addrStat.Title = stat.Title
		addrStat.Addr = stat.TargetAddr
		addrStat.Status = pt.ParacrossNodeJoined
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaNodeStatusReceipt(a.fromaddr, nil, addrStat), nil
	}

	preStat := *addrStat
	if stat.Status == pt.ParacrossNodeJoining {
		addrStat.Status = pt.ParacrossNodeJoined
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaNodeStatusReceipt(a.fromaddr, &preStat, addrStat), nil
	}

	if stat.Status == pt.ParacrossNodeQuiting {
		proposalStat, err := getNodeID(a.db, addrStat.ProposalId)
		if err != nil {
			return nil, errors.Wrapf(err, "nodeAddr:%s quiting wrong proposeid:%s", stat.TargetAddr, addrStat.ProposalId)
		}

		addrStat.Status = pt.ParacrossNodeQuited
		addrStat.QuitId = stat.Id
		receipt := makeParaNodeStatusReceipt(a.fromaddr, &preStat, addrStat)

		if !types.IsPara() {
			r, err := a.nodeGroupCoinsActive(proposalStat.FromAddr, proposalStat.CoinsFrozen, 1)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)
		}
		return receipt, nil
	}

	return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "nodeAddr:%s  get wrong status:%d", stat.TargetAddr, stat.Status)
}

func (a *action) nodeVote(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	nodes, _, err := getParacrossNodes(a.db, config.Title)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	if !validNode(a.fromaddr, nodes) && !isSuperManager(a.fromaddr) {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "not validNode:%s", a.fromaddr)
	}

	stat, err := getNodeID(a.db, config.Id)
	if err != nil {
		return nil, err
	}
	if config.Title != stat.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, stat.Title)
	}
	if stat.Status != pt.ParacrossNodeJoining && stat.Status != pt.ParacrossNodeQuiting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, stat.Status)
	}

	//已经被其他id pass 场景
	if stat.Status == pt.ParacrossNodeJoining && validNode(stat.TargetAddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "config id:%s,addr:%s", config.Id, stat.TargetAddr)
	}
	if stat.Status == pt.ParacrossNodeQuiting && !validNode(stat.TargetAddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "config id:%s,addr:%s", config.Id, stat.TargetAddr)
	}

	var copyStat pt.ParaNodeIdStatus
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
		stat.Votes.Votes[index] = pt.ParaNodeVoteStr[config.Value]
	} else {
		stat.Votes.Addrs = append(stat.Votes.Addrs, a.fromaddr)
		stat.Votes.Votes = append(stat.Votes.Votes, pt.ParaNodeVoteStr[config.Value])
	}

	//剔除已退出nodegroup的addr的投票
	updateVotes(stat, nodes)

	most, vote := getMostVote(stat)
	if !isCommitDone(stat, nodes, most) {
		superManagerPass := false
		if isSuperManager(a.fromaddr) {
			//如果主链执行失败，交易不会过滤到平行链，如果主链成功，平行链直接成功
			if !types.IsPara() {
				err := a.superManagerVoteProc(config.Title)
				if err != nil {
					return nil, err
				}
			}
			superManagerPass = true
		}

		//超级用户投yes票，共识停止了一定高度就可以通过，防止当前所有授权节点都忘掉私钥场景
		if !(superManagerPass && most > 0 && vote == pt.ParaNodeVoteYes) {
			return makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat), nil
		}
	}
	clog.Info("paracross.nodeVote  ----pass", "most", most, "pass", vote)

	receipt := &types.Receipt{Ty: types.ExecOk}
	if vote == pt.ParaNodeVoteNo {
		if stat.Status == pt.ParacrossNodeJoining {
			stat.Status = pt.ParacrossNodeClosed
			stat.Height = a.height
			//active coins
			if !types.IsPara() {
				r, err := a.nodeGroupCoinsActive(stat.FromAddr, stat.CoinsFrozen, 1)
				if err != nil {
					return nil, err
				}
				receipt = mergeReceipt(receipt, r)
			}
		} else if stat.Status == pt.ParacrossNodeQuiting {
			stat.Status = pt.ParacrossNodeClosed
			stat.Height = a.height
		}

	} else {
		if stat.Status == pt.ParacrossNodeJoining {
			r, err := unpdateNodeGroup(a.db, config.Title, stat.TargetAddr, true)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			r, err = a.updateNodeAddrStatus(stat)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			stat.Status = pt.ParacrossNodeClosed
			stat.Height = a.height
		} else if stat.Status == pt.ParacrossNodeQuiting {
			r, err := unpdateNodeGroup(a.db, config.Title, stat.TargetAddr, false)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			r, err = a.updateNodeAddrStatus(stat)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			stat.Status = pt.ParacrossNodeClosed
			stat.Height = a.height
		}
	}
	r := makeNodeConfigReceipt(a.fromaddr, config, &copyStat, stat)
	receipt = mergeReceipt(receipt, r)

	receiptDone := makeVoteDoneReceipt(stat, len(nodes), len(stat.Votes.Addrs), most, pt.ParaNodeVoteStr[vote], stat.Status)
	receipt = mergeReceipt(receipt, receiptDone)
	return receipt, nil

}

func unpdateNodeGroup(db dbm.KV, title, addr string, add bool) (*types.Receipt, error) {
	var item types.ConfigItem

	key := calcParaNodeGroupAddrsKey(title)
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
		clog.Info("unpdateNodeGroup", "add key", string(key), "from", copyItem.GetArr().Value, "to", item.GetArr().Value)

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

	return makeParaNodeGroupReceipt(title, &copyItem, &item), nil
}

func getConfigAddrs(addr string) []string {
	addr = strings.Trim(addr, " ")
	if addr == "" {
		return nil
	}
	if strings.Contains(addr, ",") {
		repeats := make(map[string]bool)
		var addrs []string

		s := strings.Trim(addr, " ")
		s = strings.Trim(s, ",")
		ss := strings.Split(s, ",")
		for _, v := range ss {
			v = strings.Trim(v, " ")
			if v != "" && !repeats[v] {
				addrs = append(addrs, v)
				repeats[v] = true
			}
		}
		return addrs
	}

	return []string{addr}
}

func (a *action) checkNodeGroupExist(title string) error {
	key := calcParaNodeGroupAddrsKey(title)
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

func (a *action) nodeGroupCoinsFrozen(createAddr string, configCoinsFrozen int64, nodeCounts int64) (*types.Receipt, error) {
	receipt := &types.Receipt{}
	confCoins := conf.GInt("nodeGroupFrozenCoins")
	if configCoinsFrozen < confCoins {
		return nil, pt.ErrParaNodeGroupFrozenCoinsNotEnough
	}
	if configCoinsFrozen == 0 {
		clog.Info("node group apply configCoinsFrozen is 0")
		return receipt, nil
	}

	realExec := string(types.GetRealExecName(a.tx.Execer))
	realExecAddr := dapp.ExecAddress(realExec)

	r, err := a.coinsAccount.ExecFrozen(createAddr, realExecAddr, nodeCounts*configCoinsFrozen)
	if err != nil {
		clog.Error("node group apply", "addr", createAddr, "realExec", realExec, "realAddr", realExecAddr, "amount", nodeCounts*configCoinsFrozen)
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupCoinsActive(createAddr string, configCoinsFrozen int64, nodeCount int64) (*types.Receipt, error) {
	receipt := &types.Receipt{}

	realExec := string(types.GetRealExecName(a.tx.Execer))
	realExecAddr := dapp.ExecAddress(realExec)

	if configCoinsFrozen == 0 {
		return receipt, nil
	}

	r, err := a.coinsAccount.ExecActive(createAddr, realExecAddr, nodeCount*configCoinsFrozen)
	if err != nil {
		clog.Error("node group apply", "addr", createAddr,
			"realExec", realExec, "realAddr", realExecAddr, "amount", configCoinsFrozen, "nodeCount", nodeCount)
		return nil, err
	}

	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

// NodeGroupApply
func (a *action) nodeGroupApply(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	addrs := getConfigAddrs(config.Addrs)
	if len(addrs) == 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "node group apply addrs null:%s", config.Addrs)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsFrozen(a.fromaddr, config.CoinsFrozen, int64(len(addrs)))
		if err != nil {
			return nil, err
		}

		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	stat := &pt.ParaNodeGroupStatus{
		Id:          calcParaNodeGroupIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParacrossNodeGroupApply,
		Title:       config.Title,
		TargetAddrs: strings.Join(addrs, ","),
		CoinsFrozen: config.CoinsFrozen,
		FromAddr:    a.fromaddr,
		Height:      a.height}
	r := makeNodeGroupIDReceipt(a.fromaddr, nil, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupModify(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}
	stat := &pt.ParaNodeGroupStatus{
		Id:          calcParaNodeGroupIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParacrossNodeGroupModify,
		Title:       config.Title,
		CoinsFrozen: config.CoinsFrozen,
		Height:      a.height}
	r := makeNodeGroupIDReceipt(a.fromaddr, nil, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupQuit(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	status, err := getNodeGroupID(a.db, config.Id)
	if err != nil {
		return nil, err
	}

	//只能提案发起人撤销
	if a.fromaddr != status.FromAddr {
		return nil, errors.Wrapf(types.ErrNotAllow, "id create by:%s,not by:%s", status.FromAddr, a.fromaddr)
	}

	if config.Title != status.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, status.Title)
	}

	//approved or quited
	if status.Status != pt.ParacrossNodeGroupApply {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "node group apply not apply:%d", status.Status)
	}

	applyAddrs := strings.Split(status.TargetAddrs, ",")

	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	if !types.IsPara() {
		r, err := a.nodeGroupCoinsActive(status.FromAddr, status.CoinsFrozen, int64(len(applyAddrs)))
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	copyStat := *status
	status.Status = pt.ParacrossNodeGroupQuit
	status.Height = a.height

	r := makeNodeGroupIDReceipt(a.fromaddr, &copyStat, status)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) nodeGroupApproveModify(config *pt.ParaNodeGroupConfig, modify *pt.ParaNodeGroupStatus) (*types.Receipt, error) {
	stat, err := getNodeGroupStatus(a.db, config.Title)
	if err != nil {
		return nil, err
	}

	//approve modify case
	if modify.CoinsFrozen < config.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "id not enough coins modify:%d,config:%d", modify.CoinsFrozen, config.CoinsFrozen)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	copyModify := *modify
	modify.Status = pt.ParacrossNodeGroupApprove
	modify.Height = a.height

	r := makeNodeGroupIDReceipt(a.fromaddr, &copyModify, modify)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	//对已经approved group和addrs不再统一修改active&frozen改动的coins，因为可能有些addr已经退出group了，没退出的，退出时候按最初设置解冻
	// 这里只修改参数，对后面再加入的节点起作用
	copyStat := *stat
	stat.Id = modify.Id
	stat.CoinsFrozen = modify.CoinsFrozen
	stat.Height = a.height

	r = makeParaNodeGroupStatusReceipt(config.Title, a.fromaddr, &copyStat, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil

}

func (a *action) nodeGroupApproveApply(config *pt.ParaNodeGroupConfig, apply *pt.ParaNodeGroupStatus) (*types.Receipt, error) {
	err := a.checkNodeGroupExist(config.Title)
	if err != nil {
		return nil, err
	}

	if apply.CoinsFrozen < config.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "id not enough coins apply:%d,config:%d", apply.CoinsFrozen, config.CoinsFrozen)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//create the node group
	r, err := a.nodeGroupCreate(apply)
	if err != nil {
		return nil, errors.Wrapf(err, "nodegroup create:title:%s,addrs:%s", config.Title, apply.TargetAddrs)
	}
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	copyStat := *apply
	apply.Status = pt.ParacrossNodeGroupApprove
	apply.Height = a.height

	r = makeNodeGroupIDReceipt(a.fromaddr, &copyStat, apply)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	r = makeParaNodeGroupStatusReceipt(config.Title, a.fromaddr, nil, apply)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil

}

// NodeGroupApprove super addr approve the node group apply
func (a *action) nodeGroupApprove(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	if !isSuperManager(a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "node group approve not super manager:%s", a.fromaddr)
	}

	id, err := getNodeGroupID(a.db, config.Id)
	if err != nil {
		return nil, err
	}

	if config.Title != id.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, id.Title)
	}

	if id.Status == pt.ParacrossNodeGroupModify {
		return a.nodeGroupApproveModify(config, id)
	}

	if id.Status == pt.ParacrossNodeGroupApply {
		return a.nodeGroupApproveApply(config, id)
	}

	return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "nodeGroupApprove id wrong status:%d,id:%s", id.Status, config.Id)

}

func (a *action) nodeGroupCreate(status *pt.ParaNodeGroupStatus) (*types.Receipt, error) {
	nodes := strings.Split(status.TargetAddrs, ",")

	var item types.ConfigItem
	key := calcParaNodeGroupAddrsKey(status.Title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, nodes...)
	item.Addr = a.fromaddr

	receipt := makeParaNodeGroupReceipt(status.Title, nil, &item)

	//update addr status
	for i, addr := range nodes {
		stat := &pt.ParaNodeIdStatus{
			Id:          status.Id + "-" + strconv.Itoa(i),
			Status:      pt.ParacrossNodeClosed,
			Title:       status.Title,
			TargetAddr:  addr,
			Votes:       &pt.ParaNodeVoteDetail{Addrs: []string{a.fromaddr}, Votes: []string{"yes"}},
			CoinsFrozen: status.CoinsFrozen,
			FromAddr:    status.FromAddr,
			Height:      a.height}

		r := makeNodeConfigReceipt(a.fromaddr, nil, nil, stat)
		receipt = mergeReceipt(receipt, r)

		r, err := a.updateNodeAddrStatus(stat)
		if err != nil {
			return nil, err
		}
		receipt = mergeReceipt(receipt, r)
	}
	return receipt, nil
}

//NodeGroupConfig support super node group config
func (a *action) NodeGroupConfig(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	if !validTitle(config.Title) {
		return nil, pt.ErrInvalidTitle
	}

	if config.Op == pt.ParacrossNodeGroupApply {
		s := strings.Trim(config.Addrs, " ")
		if len(s) == 0 {
			return nil, types.ErrInvalidParam
		}
		err := a.checkNodeGroupExist(config.Title)
		if err != nil {
			return nil, err
		}
		return a.nodeGroupApply(config)

	} else if config.Op == pt.ParacrossNodeGroupApprove {
		if config.Id == "" {
			return nil, types.ErrInvalidParam
		}
		return a.nodeGroupApprove(config)

	} else if config.Op == pt.ParacrossNodeGroupQuit {
		if config.Id == "" {
			return nil, types.ErrInvalidParam
		}
		return a.nodeGroupQuit(config)
	} else if config.Op == pt.ParacrossNodeGroupModify {

		return a.nodeGroupModify(config)
	}

	return nil, pt.ErrParaUnSupportNodeOper

}

//NodeConfig support super account node config
func (a *action) NodeConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	if !validTitle(config.Title) {
		return nil, pt.ErrInvalidTitle
	}

	if config.Op == pt.ParaNodeJoin {
		return a.nodeJoin(config)

	} else if config.Op == pt.ParaNodeQuit {
		return a.nodeQuit(config)

	} else if config.Op == pt.ParaNodeCancel {
		if config.Id == "" {
			return nil, types.ErrInvalidParam
		}
		return a.nodeCancel(config)

	} else if config.Op == pt.ParaNodeVote {
		if config.Id == "" || config.Value >= pt.ParaNodeVoteEnd {
			return nil, types.ErrInvalidParam
		}
		return a.nodeVote(config)
	}

	return nil, pt.ErrParaUnSupportNodeOper

}

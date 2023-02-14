// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"

	"strconv"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	manager "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

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

//分叉之前 id是"mavl-paracros-...0x12342308b"格式，分叉以后只支持输入为去掉了mavl-paracross前缀的交易id，系统会为id加上前缀
func getNodeIDWithFork(cfg *types.Chain33Config, db dbm.KV, title string, height int64, id string) (*pt.ParaNodeIdStatus, error) {
	if pt.IsParaForkHeight(cfg, height, pt.ForkLoopCheckCommitTxDone) {
		id = calcParaNodeIDKey(title, id)
	}
	return getNodeID(db, id)
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

//分叉之前 id是"mavl-paracros-...0x12342308b"格式，分叉以后只支持输入为去掉了mavl-paracross前缀的交易id，系统会为id加上前缀
func getNodeGroupID(cfg *types.Chain33Config, db dbm.KV, title string, height int64, id string) (*pt.ParaNodeGroupStatus, error) {
	if pt.IsParaForkHeight(cfg, height, pt.ForkLoopCheckCommitTxDone) {
		id = calcParaNodeGroupIDKey(title, id)
	}
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
	cfg := a.api.GetConfig()
	if !cfg.IsPara() {
		r, err := a.nodeGroupCoinsFrozen(a.fromaddr, config.CoinsFrozen, 1)
		if err != nil {
			return nil, err
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	addrStat, err := getNodeAddr(a.db, config.Title, config.Addr)
	if err != nil && !isNotFound(err) {
		return nil, errors.Wrapf(err, "nodeJoin get title=%s,nodeAddr=%s", config.Title, config.Addr)
	}
	if addrStat != nil && addrStat.Status != pt.ParaApplyQuited {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeJoin nodeAddr existed:%s,status:%d", config.Addr, addrStat.Status)
	}
	stat := &pt.ParaNodeIdStatus{
		Id:          calcParaNodeIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParaApplyJoining,
		Title:       config.Title,
		TargetAddr:  config.Addr,
		BlsPubKey:   config.BlsPubKey,
		FromAddr:    a.fromaddr,
		Votes:       &pt.ParaNodeVoteDetail{},
		CoinsFrozen: config.CoinsFrozen,
		Height:      a.height}
	r := makeNodeConfigReceipt(a.fromaddr, config, nil, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)
	return receipt, nil
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
	if addrStat.Status != pt.ParaApplyJoined {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "nodeAddr:%s status:%d", config.Addr, addrStat.Status)
	}

	stat := &pt.ParaNodeIdStatus{
		Id:         calcParaNodeIDKey(config.Title, common.ToHex(a.txhash)),
		Status:     pt.ParaApplyQuiting,
		Title:      config.Title,
		TargetAddr: config.Addr,
		FromAddr:   a.fromaddr,
		Votes:      &pt.ParaNodeVoteDetail{},
		Height:     a.height}
	return makeNodeConfigReceipt(a.fromaddr, config, nil, stat), nil
}

func (a *action) nodeCancel(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	stat, err := getNodeIDWithFork(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
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

	if stat.Status != pt.ParaApplyJoining && stat.Status != pt.ParaApplyQuiting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, stat.Status)
	}

	copyStat := proto.Clone(stat).(*pt.ParaNodeIdStatus)
	if stat.Status == pt.ParaApplyJoining {
		receipt := &types.Receipt{Ty: types.ExecOk}
		cfg := a.api.GetConfig()
		if !cfg.IsPara() {
			r, err := a.nodeGroupCoinsActive(stat.FromAddr, stat.CoinsFrozen, 1)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)
		}
		stat.Status = pt.ParaApplyCanceled
		stat.Height = a.height
		r := makeNodeConfigReceipt(a.fromaddr, config, copyStat, stat)
		receipt = mergeReceipt(receipt, r)
		return receipt, nil
	}

	if stat.Status == pt.ParaApplyQuiting {
		stat.Status = pt.ParaApplyCanceled
		stat.Height = a.height
		return makeNodeConfigReceipt(a.fromaddr, config, copyStat, stat), nil
	}

	return nil, errors.Wrapf(pt.ErrParaUnSupportNodeOper, "nodeid %s was quit status:%d", config.Id, stat.Status)
}

func (a *action) nodeModify(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	addrStat, err := getNodeAddr(a.db, config.Title, config.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "nodeAddr:%s get error", config.Addr)
	}

	//只能提案发起人撤销
	if a.fromaddr != config.Addr {
		return nil, errors.Wrapf(types.ErrNotAllow, "addr create by:%s,not by:%s", config.Addr, a.fromaddr)
	}

	preStat := *addrStat
	addrStat.BlsPubKey = config.BlsPubKey

	return makeParaNodeStatusReceipt(a.fromaddr, &preStat, addrStat), nil
}

// IsSuperManager is supper manager or not
func isSuperManager(cfg *types.Chain33Config, addr string) bool {
	confManager := types.ConfSub(cfg, manager.ManageX)
	for _, m := range confManager.GStrList("superManager") {
		if addr == m {
			return true
		}
	}
	return false
}

func getMostVote(stat *pt.ParaNodeVoteDetail) (int, int) {
	var ok, nok int
	for _, v := range stat.Votes {
		if v == pt.ParaNodeVoteStr[pt.ParaVoteYes] {
			ok++
		} else {
			nok++
		}
	}
	if ok > nok {
		return ok, pt.ParaVoteYes
	}
	return nok, pt.ParaVoteNo

}

func hasVoted(addrs []string, addr string) (bool, int) {
	return hasCommited(addrs, addr)
}

//主链配置平行链停止块数， 反应到主链上为对应平行链空块间隔×停止块数，如果主链当前高度超过平行链共识高度对应主链高度后面这个主链块数就表示通过
func (a *action) superManagerVoteProc(title string) error {
	status, err := getNodeGroupStatus(a.db, title)
	if err != nil {
		return errors.Wrap(err, "getNodegroupStatus")
	}
	cfg := a.api.GetConfig()
	conf := types.ConfSub(cfg, pt.ParaX)
	confStopBlocks := conf.GInt("paraConsensusStopBlocks")
	data, err := a.exec.paracrossGetHeight(title)
	if err != nil {
		clog.Info("paracross.superManagerVoteProc get consens height", "title", title, "err", err.Error())
		return errors.Wrap(err, "getTitleHeight")
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
			return errors.Wrapf(err, "getStateTitleHeight,consensHeight=%d", consensHeight)
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

func updateVotes(in *pt.ParaNodeVoteDetail, nodes map[string]struct{}) *pt.ParaNodeVoteDetail {
	votes := &pt.ParaNodeVoteDetail{}
	for i, addr := range in.Addrs {
		if _, ok := nodes[addr]; ok {
			votes.Addrs = append(votes.Addrs, addr)
			votes.Votes = append(votes.Votes, in.Votes[i])
		}
	}
	return votes
}

//由于propasal id 和quit id分开，quit id不知道对应addr　proposal id的coinfrozen信息，需要维护一个围绕addr的数据库结构信息
func (a *action) updateNodeAddrStatus(stat *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	addrStat, err := getNodeAddr(a.db, stat.Title, stat.TargetAddr)
	if err != nil {
		if !isNotFound(err) {
			return nil, errors.Wrapf(err, "nodeAddr:%s get error", stat.TargetAddr)
		}
		addrStat = &pt.ParaNodeAddrIdStatus{}
		addrStat.Title = stat.Title
		addrStat.Addr = stat.TargetAddr
		addrStat.BlsPubKey = stat.BlsPubKey
		addrStat.Status = pt.ParaApplyJoined
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaNodeStatusReceipt(a.fromaddr, nil, addrStat), nil
	}

	preStat := *addrStat
	if stat.Status == pt.ParaApplyJoining {
		addrStat.Status = pt.ParaApplyJoined
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaNodeStatusReceipt(a.fromaddr, &preStat, addrStat), nil
	}

	if stat.Status == pt.ParaApplyQuiting {
		proposalStat, err := getNodeID(a.db, addrStat.ProposalId)
		if err != nil {
			return nil, errors.Wrapf(err, "nodeAddr:%s quiting wrong proposeid:%s", stat.TargetAddr, addrStat.ProposalId)
		}

		addrStat.Status = pt.ParaApplyQuited
		addrStat.QuitId = stat.Id
		receipt := makeParaNodeStatusReceipt(a.fromaddr, &preStat, addrStat)

		cfg := a.api.GetConfig()
		if !cfg.IsPara() {
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

func (a *action) checkIsSuperManagerVote(config *pt.ParaNodeAddrConfig, nodes map[string]struct{}) (bool, error) {
	cfg := a.api.GetConfig()

	//平行链：主链成功的交易如果不是nodegroup节点，就是superManager,为了不需在平行链配置superManger,间接判断是否是superManager
	if cfg.IsPara() {
		if validNode(a.fromaddr, nodes) {
			return false, nil
		}
		return true, nil
	}

	//主链：只检查超级管理员处理
	if !isSuperManager(cfg, a.fromaddr) {
		return false, nil
	}
	err := a.superManagerVoteProc(config.Title)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (a *action) nodeVote(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	nodes, _, err := getParacrossNodes(a.db, config.Title)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}
	cfg := a.api.GetConfig()
	//只在主链检查，　只有nodegroup成员和supermanager可以vote
	if !cfg.IsPara() && !validNode(a.fromaddr, nodes) && !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "not validNode:%s", a.fromaddr)
	}

	stat, err := getNodeIDWithFork(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
	if err != nil {
		return nil, err
	}
	if config.Title != stat.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, stat.Title)
	}
	if stat.Status != pt.ParaApplyJoining && stat.Status != pt.ParaApplyQuiting {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, stat.Status)
	}

	//已经被其他id pass 场景
	if stat.Status == pt.ParaApplyJoining && validNode(stat.TargetAddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "config id:%s,addr:%s", config.Id, stat.TargetAddr)
	}
	if stat.Status == pt.ParaApplyQuiting && !validNode(stat.TargetAddr, nodes) {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "config id:%s,addr:%s", config.Id, stat.TargetAddr)
	}

	copyStat := proto.Clone(stat).(*pt.ParaNodeIdStatus)
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
	stat.Votes = updateVotes(stat.Votes, nodes)

	most, vote := getMostVote(stat.Votes)
	if !isCommitDone(len(nodes), most) {
		superManagerPass, err := a.checkIsSuperManagerVote(config, nodes)
		if err != nil {
			return nil, err
		}

		//超级用户投yes票，共识停止了一定高度就可以通过，防止当前所有授权节点都忘掉私钥场景
		if !(superManagerPass && most > 0 && vote == pt.ParaVoteYes) {
			return makeNodeConfigReceipt(a.fromaddr, config, copyStat, stat), nil
		}
	}
	clog.Info("paracross.nodeVote  ----pass", "most", most, "pass", vote)

	receipt := &types.Receipt{Ty: types.ExecOk}
	if vote == pt.ParaVoteNo {
		if stat.Status == pt.ParaApplyJoining {
			stat.Status = pt.ParaApplyClosed
			stat.Height = a.height
			//active coins
			if !cfg.IsPara() {
				r, err := a.nodeGroupCoinsActive(stat.FromAddr, stat.CoinsFrozen, 1)
				if err != nil {
					return nil, err
				}
				receipt = mergeReceipt(receipt, r)
			}
		} else if stat.Status == pt.ParaApplyQuiting {
			stat.Status = pt.ParaApplyClosed
			stat.Height = a.height
		}
	} else {
		if stat.Status == pt.ParaApplyJoining {
			r, err := updateNodeGroup(a.db, config.Title, stat.TargetAddr, true)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			r, err = a.updateNodeAddrStatus(stat)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			stat.Status = pt.ParaApplyClosed
			stat.Height = a.height
		} else if stat.Status == pt.ParaApplyQuiting {
			r, err := updateNodeGroup(a.db, config.Title, stat.TargetAddr, false)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			r, err = a.updateNodeAddrStatus(stat)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)

			if a.exec.GetMainHeight() > pt.GetDappForkHeight(cfg, pt.ForkLoopCheckCommitTxDone) {
				//node quit后，如果committx满足2/3目标，自动触发commitDone
				r, err = a.loopCommitTxDone(config.Title)
				if err != nil {
					clog.Error("updateNodeGroup.loopCommitTxDone", "title", cfg.GetTitle(), "err", err.Error())
				}
				receipt = mergeReceipt(receipt, r)
			}

			stat.Status = pt.ParaApplyClosed
			stat.Height = a.height
		}
	}
	r := makeNodeConfigReceipt(a.fromaddr, config, copyStat, stat)
	receipt = mergeReceipt(receipt, r)

	receiptDone := makeVoteDoneReceipt(stat, len(nodes), len(stat.Votes.Addrs), most, pt.ParaNodeVoteStr[vote], stat.Status)
	receipt = mergeReceipt(receipt, receiptDone)
	return receipt, nil
}

func updateNodeGroup(db dbm.KV, title, addr string, add bool) (*types.Receipt, error) {
	var item types.ConfigItem

	key := calcParaNodeGroupAddrsKey(title)
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			clog.Error("updateNodeGroup", "decode db key", key)
			return nil, err // types.ErrBadConfigValue
		}
	}

	copyValue := *item.GetArr()
	copyItem := item
	copyItem.Value = &types.ConfigItem_Arr{Arr: &copyValue}

	if add {
		item.GetArr().Value = append(item.GetArr().Value, addr)
		item.Addr = addr
		clog.Info("updateNodeGroup add", "addr", addr, "from", copyItem.GetArr().Value, "to", item.GetArr().Value)
	} else {
		//必须保留至少1个授权账户
		if len(item.GetArr().Value) <= 1 {
			return nil, pt.ErrParaNodeGroupLastAddr
		}
		item.Addr = addr
		item.GetArr().Value = make([]string, 0)
		for _, value := range copyItem.GetArr().Value {
			if value != addr {
				item.GetArr().Value = append(item.GetArr().Value, value)
			}
		}
		clog.Info("updateNodeGroup delete", "addr", addr)
	}
	err = db.Set(key, types.Encode(&item))
	if err != nil {
		return nil, errors.Wrapf(err, "updateNodeGroup set dbkey=%s", key)
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
	cfg := a.api.GetConfig()
	conf := types.ConfSub(cfg, pt.ParaX)
	confCoins := conf.GInt("nodeGroupFrozenCoins")
	if configCoinsFrozen < confCoins {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "nodeGroupCoinsFrozen apply=%d,conf=%d", configCoinsFrozen, confCoins)
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

	var blsPubKeys []string
	if len(config.BlsPubKeys) > 0 {
		blsPubKeys = getConfigAddrs(config.BlsPubKeys)
		if len(blsPubKeys) != len(addrs) {
			return nil, errors.Wrapf(types.ErrInvalidParam, "nodegroup apply blsPubkeys length=%d not match addrs=%d",
				len(blsPubKeys), len(addrs))
		}
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	cfg := a.api.GetConfig()
	if !cfg.IsPara() {
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
		BlsPubKeys:  strings.Join(blsPubKeys, ","),
		CoinsFrozen: config.CoinsFrozen,
		FromAddr:    a.fromaddr,
		Height:      a.height,
	}

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
	cfg := a.api.GetConfig()
	status, err := getNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
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
	if !cfg.IsPara() {
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
	cfg := a.api.GetConfig()
	if cfg.IsPara() && cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaSelfConsStages) {
		//不允许主链成功平行链失败导致不一致的情况，这里如果失败则手工设置init stage
		r = selfConsensInitStage(cfg)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	return receipt, nil
}

func (a *action) checkApproveOpOld(config *pt.ParaNodeGroupConfig) error {
	cfg := a.api.GetConfig()
	//fork之后采用 autonomy 检查模式
	confManager := types.ConfSub(cfg, manager.ManageX)
	autonomyExec := confManager.GStr(types.AutonomyCfgKey)
	if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaAutonomySuperGroup) && len(autonomyExec) > 0 {
		//去autonomy 合约检验是否id approved, 成功 err返回nil
		_, err := a.api.QueryChain(&types.ChainExecutor{
			Driver:   autonomyExec,
			FuncName: "IsAutonomyApprovedItem",
			Param:    types.Encode(&types.ReqMultiStrings{Datas: []string{config.AutonomyItemID, config.Id}}),
		})
		if err != nil {
			return errors.Wrapf(err, "query autonomy,approveid=%s,hashId=%s", config.AutonomyItemID, config.Id)
		}

		return nil
	}

	//fork之前检查是否from superManager
	if !isSuperManager(cfg, a.fromaddr) {
		return errors.Wrapf(types.ErrNotAllow, "node group approve not super manager:%s", a.fromaddr)
	}
	return nil
}

func (a *action) checkApproveOpNew(config *pt.ParaNodeGroupConfig, status *pt.ParaNodeGroupStatus) error {
	cfg := a.api.GetConfig()

	//from地址和apply的相同，ok
	if status.FromAddr == a.fromaddr {
		return nil
	}
	//from superManager
	if isSuperManager(cfg, a.fromaddr) {
		return nil
	}

	confManager := types.ConfSub(cfg, manager.ManageX)
	autonomyExec := confManager.GStr(types.AutonomyCfgKey)
	if len(autonomyExec) > 0 {
		//去autonomy 合约检验是否id approved, 成功 err返回nil
		_, err := a.api.QueryChain(&types.ChainExecutor{
			Driver:   autonomyExec,
			FuncName: "IsAutonomyApprovedItem",
			Param:    types.Encode(&types.ReqMultiStrings{Datas: []string{config.AutonomyItemID, config.Id}}),
		})
		if err != nil {
			return errors.Wrapf(err, "query autonomy,approveid=%s,hashId=%s", config.AutonomyItemID, config.Id)
		}
		return nil
	}

	return errors.Wrapf(types.ErrNotAllow, "from Addr=%s not applier=%s or manager", a.fromaddr, status.FromAddr)
}

//核查approve的条件，最新的分叉版本开启了自由注册模式，只要申请者和approve者是同一个用户即可
func (a *action) checkApproveOp(config *pt.ParaNodeGroupConfig, status *pt.ParaNodeGroupStatus) error {
	cfg := a.api.GetConfig()
	//最新版本开启自由注册的检查，只要申请者和approve者是同一个用户即可
	if cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaFreeRegister) {
		return a.checkApproveOpNew(config, status)
	}
	//自由注册之前的检查，之前曾经需要atonomy投票，然后获取投票id来审查通过
	return a.checkApproveOpOld(config)
}

// NodeGroupApprove super addr approve the node group apply
func (a *action) nodeGroupApprove(config *pt.ParaNodeGroupConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()

	id, err := getNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
	if err != nil {
		return nil, errors.Wrapf(err, "getNodegGroupId=%s", config.Id)
	}

	if config.Title != id.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, id.Title)
	}

	//只在主链检查， 主链检查失败不会同步到平行链，主链成功，平行链默认成功
	if !cfg.IsPara() {
		err := a.checkApproveOp(config, id)
		if err != nil {
			return nil, err
		}
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

	var blsPubKeys []string
	if len(status.BlsPubKeys) > 0 {
		blsPubKeys = strings.Split(status.BlsPubKeys, ",")
	}

	//update addr status
	for i, addr := range nodes {
		stat := &pt.ParaNodeIdStatus{
			Id:          status.Id + "-" + strconv.Itoa(i),
			Status:      pt.ParaApplyClosed,
			Title:       status.Title,
			TargetAddr:  addr,
			Votes:       &pt.ParaNodeVoteDetail{Addrs: []string{a.fromaddr}, Votes: []string{"yes"}},
			CoinsFrozen: status.CoinsFrozen,
			FromAddr:    status.FromAddr,
			Height:      a.height}
		if len(blsPubKeys) > 0 {
			stat.BlsPubKey = blsPubKeys[i]
		}
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
	cfg := a.api.GetConfig()
	if !validTitle(cfg, config.Title) {
		return nil, pt.ErrInvalidTitle
	}
	if !types.IsParaExecName(string(a.tx.Execer)) && cfg.IsDappFork(a.exec.GetMainHeight(), pt.ParaX, pt.ForkParaAssetTransferRbk) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
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
	cfg := a.api.GetConfig()
	if !validTitle(cfg, config.Title) {
		return nil, pt.ErrInvalidTitle
	}
	if !types.IsParaExecName(string(a.tx.Execer)) && cfg.IsDappFork(a.exec.GetMainHeight(), pt.ParaX, pt.ForkParaAssetTransferRbk) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}

	switch config.Op {
	case pt.ParaOpNewApply:
		return a.nodeJoin(config)
	case pt.ParaOpQuit:
		//退出nodegroup
		return a.nodeQuit(config)
	case pt.ParaOpCancel:
		//撤销未批准的申请
		if config.Id == "" {
			return nil, types.ErrInvalidParam
		}
		return a.nodeCancel(config)
	case pt.ParaOpVote:
		if config.Id == "" || config.Value >= pt.ParaVoteEnd {
			return nil, types.ErrInvalidParam
		}
		return a.nodeVote(config)
	case pt.ParaOpModify:
		//修改addr相关联的参数，只能原创地址修改
		return a.nodeModify(config)
	default:
		return nil, pt.ErrParaUnSupportNodeOper
	}
}

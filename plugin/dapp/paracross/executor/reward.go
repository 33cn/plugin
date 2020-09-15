package executor

import (
	"bytes"

	"github.com/pkg/errors"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/golang/protobuf/proto"
)

const (
	opBind   = 1
	opUnBind = 2
)

//根据挖矿共识节点地址 过滤整体共识节点映射列表， 获取委托挖矿地址
func (a *action) getBindAddrs(nodes []string, statusHeight int64) (*pt.ParaNodeBindList, error) {
	nodesMap := make(map[string]bool)
	for _, n := range nodes {
		nodesMap[n] = true
	}

	var newLists pt.ParaNodeBindList
	list, err := getBindNodeInfo(a.db)
	if err != nil {
		clog.Error("paracross getBindAddrs err", "height", statusHeight)
		return nil, err
	}
	//这样检索是按照list的映射顺序，不是按照nodes的顺序(需要循环嵌套)
	for _, m := range list.Miners {
		if nodesMap[m.SuperNode] {
			newLists.Miners = append(newLists.Miners, m)
		}
	}

	return &newLists, nil

}

func (a *action) rewardSuperNode(coinReward int64, miners []string, statusHeight int64) (*types.Receipt, int64, error) {
	//分配给矿工的单位奖励
	minerUnit := coinReward / int64(len(miners))
	var change int64
	receipt := &types.Receipt{Ty: types.ExecOk}
	if minerUnit > 0 {
		//如果不等分转到发展基金
		change = coinReward % minerUnit
		for _, addr := range miners {
			rep, err := a.coinsAccount.ExecDeposit(addr, a.execaddr, minerUnit)

			if err != nil {
				clog.Error("paracross super node reward deposit err", "height", statusHeight,
					"execAddr", a.execaddr, "minerAddr", addr, "amount", minerUnit, "err", err)
				return nil, 0, err
			}
			receipt = mergeReceipt(receipt, rep)
		}
	}
	return receipt, change, nil
}

//奖励委托挖矿账户
func (a *action) rewardBindAddr(coinReward int64, bindList *pt.ParaNodeBindList, statusHeight int64) (*types.Receipt, int64, error) {
	if coinReward <= 0 {
		return nil, 0, nil
	}

	//有可能一个bindAddr 在多个node绑定，这里会累计上去
	var bindAddrList []*pt.ParaBindMinerInfo
	for _, node := range bindList.Miners {
		info, err := getBindAddrInfo(a.db, node.SuperNode, node.Miner)
		if err != nil {
			return nil, 0, err
		}
		bindAddrList = append(bindAddrList, info)
	}

	var totalCoins int64
	for _, addr := range bindAddrList {
		totalCoins += addr.BindCoins
	}

	//分配给矿工的单位奖励
	minerUnit := coinReward / totalCoins
	var change int64
	receipt := &types.Receipt{Ty: types.ExecOk}
	if minerUnit > 0 {
		//如果不等分转到发展基金
		change = coinReward % minerUnit
		for _, miner := range bindAddrList {
			rep, err := a.coinsAccount.ExecDeposit(miner.Addr, a.execaddr, minerUnit*miner.BindCoins)
			if err != nil {
				clog.Error("paracross bind miner reward deposit err", "height", statusHeight,
					"execAddr", a.execaddr, "minerAddr", miner.Addr, "amount", minerUnit*miner.BindCoins, "err", err)
				return nil, 0, err
			}
			receipt = mergeReceipt(receipt, rep)
		}
	}
	return receipt, change, nil
}

// reward 挖矿奖励， 主要处理挖矿分配逻辑，先实现基本策略，后面根据需求进行重构
func (a *action) reward(nodeStatus *pt.ParacrossNodeStatus, stat *pt.ParacrossHeightStatus) (*types.Receipt, error) {
	//获取挖矿相关配置，这里需注意是共识的高度，而不是交易的高度
	cfg := a.api.GetConfig()
	coinReward := cfg.MGInt("mver.consensus.paracross.coinReward", nodeStatus.Height) * types.Coin
	coinBaseReward := cfg.MGInt("mver.consensus.paracross.coinBaseReward", nodeStatus.Height) * types.Coin
	fundReward := cfg.MGInt("mver.consensus.paracross.coinDevFund", nodeStatus.Height) * types.Coin
	fundAddr := cfg.MGStr("mver.consensus.fundKeyAddr", nodeStatus.Height)

	//防止coinBaseReward 设置出错场景， coinBaseReward 一定要比coinReward小
	if coinBaseReward >= coinReward {
		coinBaseReward = coinReward / 10
	}

	//超级节点地址
	nodeAddrs := getSuperNodes(stat.Details, nodeStatus.BlockHash)
	//委托地址
	bindAddrs, err := a.getBindAddrs(nodeAddrs, nodeStatus.Height)
	if err != nil {
		return nil, err
	}

	//奖励超级节点
	minderRewards := coinReward
	//如果有委托挖矿地址，则超级节点分baseReward部分，否则全部
	if len(bindAddrs.Miners) > 0 {
		minderRewards = coinBaseReward
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	r, change, err := a.rewardSuperNode(minderRewards, nodeAddrs, nodeStatus.Height)
	if err != nil {
		return nil, err
	}
	fundReward += change
	mergeReceipt(receipt, r)

	//奖励委托挖矿地址
	r, change, err = a.rewardBindAddr(coinReward-minderRewards, bindAddrs, nodeStatus.Height)
	if err != nil {
		return nil, err
	}
	fundReward += change
	mergeReceipt(receipt, r)

	//奖励发展基金
	if fundReward > 0 {
		rep, err := a.coinsAccount.ExecDeposit(fundAddr, a.execaddr, fundReward)
		if err != nil {
			clog.Error("paracross fund reward deposit err", "height", nodeStatus.Height,
				"execAddr", a.execaddr, "fundAddr", fundAddr, "amount", fundReward, "err", err)
			return nil, err
		}
		receipt = mergeReceipt(receipt, rep)
	}

	return receipt, nil
}

// getSuperNodes 获取提交共识消息的矿工地址
func getSuperNodes(detail *pt.ParacrossStatusDetails, blockHash []byte) []string {
	addrs := make([]string, 0)
	for i, hash := range detail.BlockHash {
		if bytes.Equal(hash, blockHash) {
			addrs = append(addrs, detail.Addrs[i])
		}
	}
	return addrs
}

//
func mergeReceipt(receipt1, receipt2 *types.Receipt) *types.Receipt {
	if receipt2 != nil {
		receipt1.KV = append(receipt1.KV, receipt2.KV...)
		receipt1.Logs = append(receipt1.Logs, receipt2.Logs...)
	}

	return receipt1
}

func makeAddrBindReceipt(node, addr string, prev, current *pt.ParaBindMinerInfo) *types.Receipt {
	key := calcParaBindMinerAddr(node, addr)
	log := &pt.ReceiptParaBindMinerInfo{
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
				Ty:  pt.TyLogParaBindMinerAddr,
				Log: types.Encode(log),
			},
		},
	}
}

func makeNodeBindReceipt(prev, current *pt.ParaNodeBindList) *types.Receipt {
	key := calcParaBindMinerNode()
	log := &pt.ReceiptParaNodeBindListUpdate{
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
				Ty:  pt.TyLogParaBindMinerNode,
				Log: types.Encode(log),
			},
		},
	}
}

//绑定到超级节点
func (a *action) bind2Node(node string) (*types.Receipt, error) {
	list, err := getBindNodeInfo(a.db)
	if err != nil {
		return nil, errors.Wrap(err, "bind2Node")
	}

	//由于kvmvcc内存架构，如果存储结构为nil，将回溯查找，这样在只有一个绑定时候，unbind后，有可能会回溯到更早状态，是错误的，title这里就是占位使用
	if len(list.Title) <= 0 {
		list.Title = a.api.GetConfig().GetTitle()
	}

	old := proto.Clone(list).(*pt.ParaNodeBindList)
	list.Miners = append(list.Miners, &pt.ParaNodeBindOne{SuperNode: node, Miner: a.fromaddr})

	return makeNodeBindReceipt(old, list), nil

}

//从超级节点解绑
func (a *action) unbind2Node(node string) (*types.Receipt, error) {
	list, err := getBindNodeInfo(a.db)
	if err != nil {
		return nil, errors.Wrap(err, "unbind2Node")
	}
	newList := &pt.ParaNodeBindList{Title: a.api.GetConfig().GetTitle()}
	old := proto.Clone(list).(*pt.ParaNodeBindList)

	for _, m := range list.Miners {
		if m.SuperNode == node && m.Miner == a.fromaddr {
			continue
		}
		newList.Miners = append(newList.Miners, m)
	}
	return makeNodeBindReceipt(old, newList), nil

}

func getBindNodeInfo(db dbm.KV) (*pt.ParaNodeBindList, error) {
	var list pt.ParaNodeBindList
	key := calcParaBindMinerNode()
	data, err := db.Get(key)
	if isNotFound(err) {
		return &list, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "get key failed")
	}

	err = types.Decode(data, &list)
	if err != nil {
		return nil, errors.Wrapf(err, "decode failed")
	}
	return &list, nil
}

func getBindAddrInfo(db dbm.KV, node, addr string) (*pt.ParaBindMinerInfo, error) {
	key := calcParaBindMinerAddr(node, addr)
	data, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get key failed node=%s,addr=%s", node, addr)
	}

	var info pt.ParaBindMinerInfo
	err = types.Decode(data, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "decode failed node=%s,addr=%s", node, addr)
	}
	return &info, nil
}

func (a *action) bindOp(cmd *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	if cmd.BindCoins <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner BindCoins nil from addr %s", a.fromaddr)
	}

	err := a.isValidSuperNode(cmd.TargetNode)
	if err != nil {
		return nil, err
	}

	current, err := getBindAddrInfo(a.db, cmd.TargetNode, a.fromaddr)
	if err != nil && !isNotFound(errors.Cause(err)) {
		return nil, errors.Wrap(err, "getBindAddrInfo")
	}

	//found, 修改当前的绑定
	if current != nil && current.BindStatus == opBind {
		var receipt *types.Receipt

		if cmd.BindCoins == current.BindCoins {
			return nil, errors.Wrapf(types.ErrInvalidParam, "bind coins same current=%d, cmd=%d", current.BindCoins, cmd.BindCoins)
		}

		//释放一部分coins
		if cmd.BindCoins < current.BindCoins {
			receipt, err = a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, (current.BindCoins-cmd.BindCoins)*types.Coin)
			if err != nil {
				return nil, errors.Wrapf(err, "bindOp Active addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, current.BindCoins-cmd.BindCoins)
			}
		} else {
			//冻结更多
			receipt, err = a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, (cmd.BindCoins-current.BindCoins)*types.Coin)
			if err != nil {
				return nil, errors.Wrapf(err, "bindOp frozen more addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, cmd.BindCoins-current.BindCoins)
			}
		}

		acctCopy := *current
		current.BindCoins = cmd.BindCoins
		r := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, &acctCopy, current)
		return mergeReceipt(receipt, r), nil
	}

	//not bind, 增加新绑定
	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, cmd.BindCoins*types.Coin)
	if err != nil {
		return nil, errors.Wrapf(err, "bindOp frozen addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, cmd.BindCoins)
	}

	//bind addr
	newer := &pt.ParaBindMinerInfo{
		Addr:        a.fromaddr,
		BindStatus:  opBind,
		BindCoins:   cmd.BindCoins,
		BlockTime:   a.blocktime,
		BlockHeight: a.height,
		TargetNode:  cmd.TargetNode,
	}
	rBind := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, current, newer)
	mergeReceipt(receipt, rBind)

	//增加到列表中
	rList, err := a.bind2Node(cmd.TargetNode)
	if err != nil {
		return nil, err
	}
	mergeReceipt(receipt, rList)
	return receipt, nil

}

func (a *action) unBindOp(cmd *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	acct, err := getBindAddrInfo(a.db, cmd.TargetNode, a.fromaddr)
	if err != nil {
		return nil, err
	}

	cfg := a.api.GetConfig()
	unBindHours := cfg.MGInt("mver.consensus.paracross.unBindTime", a.height)
	if a.blocktime-acct.BlockTime < unBindHours*60*60 {
		return nil, errors.Wrapf(types.ErrNotAllow, "unBindOp unbind time=%d less %d hours than bind time =%d", a.blocktime, unBindHours, acct.BlockTime)
	}

	if acct.BindStatus != opBind {
		return nil, errors.Wrapf(types.ErrNotAllow, "unBindOp,current addr is unbind status")
	}

	//unfrozen
	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, acct.BindCoins*types.Coin)
	if err != nil {
		return nil, errors.Wrapf(err, "unBindOp addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, acct.BindCoins)
	}

	//删除 bind addr
	//由于kvmvcc的原因，不能通过把一个key值=nil的方式删除，kvmvcc这样是删除了当前版本，就会查询更早的版本，&struct{}也不行，len=0 也被认为是删除了的
	acctCopy := *acct
	acct.BindStatus = opUnBind
	acct.BlockHeight = a.height
	acct.BlockTime = a.blocktime
	rUnBind := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, &acctCopy, acct)
	mergeReceipt(receipt, rUnBind)

	//从列表删除
	rUnList, err := a.unbind2Node(cmd.TargetNode)
	if err != nil {
		return nil, err
	}
	mergeReceipt(receipt, rUnList)

	return receipt, nil
}

func (a *action) bindMiner(info *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	if len(info.TargetNode) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner TargetNode should not be nil to addr %s", a.fromaddr)
	}

	//只允许平行链操作
	if !types.IsParaExecName(string(a.tx.Execer)) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}

	if info.BindAction != opBind && info.BindAction != opUnBind {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner action=%d not correct", info.BindAction)
	}

	if info.BindAction == opBind {
		return a.bindOp(info)
	}
	return a.unBindOp(info)
}

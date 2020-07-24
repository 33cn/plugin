package executor

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	opBind   = 1
	opUnBind = 2
)

//根据挖矿节点地址 获取委托挖矿地址
func (a *action) getBindAddrs(nodes []string, statusHeight int64) []*pt.ParaNodeBindList {
	var lists []*pt.ParaNodeBindList
	for _, m := range nodes {
		list, err := a.getBindNodeInfo(m)
		if isNotFound(errors.Cause(err)) {
			continue
		}
		if err != nil {
			clog.Error("paracross getBindAddrs err", "height", statusHeight, "node", m)
			continue
		}
		lists = append(lists, list)
	}

	return lists

}

func isBindAddrFound(bindAddrs []*pt.ParaNodeBindList) bool {
	if len(bindAddrs) <= 0 {
		return false
	}

	for _, addr := range bindAddrs {
		if len(addr.Miners) > 0 {
			return true
		}
	}
	return false
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
func (a *action) rewardBindAddr(coinReward int64, bindList []*pt.ParaNodeBindList, statusHeight int64) (*types.Receipt, int64, error) {
	if coinReward <= 0 {
		return nil, 0, nil
	}

	//有可能一个bindAddr 在多个node绑定，这里会累计上去
	var bindAddrList []*pt.ParaBindMinerInfo
	for _, node := range bindList {
		for _, miner := range node.Miners {
			info, err := a.getBindAddrInfo(node.SuperNode, miner)
			if err != nil {
				return nil, 0, err
			}
			bindAddrList = append(bindAddrList, info)
		}
	}

	var totalCoins int64
	for _, addr := range bindAddrList {
		totalCoins += addr.BindCount
	}

	//分配给矿工的单位奖励
	minerUnit := coinReward / totalCoins
	var change int64
	receipt := &types.Receipt{Ty: types.ExecOk}
	if minerUnit > 0 {
		//如果不等分转到发展基金
		change = coinReward % minerUnit
		for _, miner := range bindAddrList {
			rep, err := a.coinsAccount.ExecDeposit(miner.Addr, a.execaddr, minerUnit*miner.BindCount)
			if err != nil {
				clog.Error("paracross bind miner reward deposit err", "height", statusHeight,
					"execAddr", a.execaddr, "minerAddr", miner.Addr, "amount", minerUnit*miner.BindCount, "err", err)
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
	minerAddrs := getMiners(stat.Details, nodeStatus.BlockHash)
	//委托地址
	bindAddrs := a.getBindAddrs(minerAddrs, nodeStatus.Height)

	//奖励超级节点
	minderRewards := coinReward
	//如果有委托挖矿地址，则超级节点分baseReward部分，否则全部
	if isBindAddrFound(bindAddrs) {
		minderRewards = coinBaseReward
	}
	receipt := &types.Receipt{Ty: types.ExecOk}
	r, change, err := a.rewardSuperNode(minderRewards, minerAddrs, nodeStatus.Height)
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

// getMiners 获取提交共识消息的矿工地址
func getMiners(detail *pt.ParacrossStatusDetails, blockHash []byte) []string {

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
	var val []byte
	if current != nil {
		val = types.Encode(current)
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: val},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaBindMinerAddr,
				Log: types.Encode(log),
			},
		},
	}
}

func makeNodeBindReceipt(addr string, prev, current *pt.ParaNodeBindList) *types.Receipt {
	key := calcParaBindMinerNode(addr)
	log := &pt.ReceiptParaNodeBindListUpdate{
		Prev:    prev,
		Current: current,
	}
	var val []byte
	if current != nil {
		val = types.Encode(current)
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: val},
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
	key := calcParaBindMinerNode(node)
	data, err := a.db.Get(key)
	if err != nil && err != types.ErrNotFound {
		return nil, errors.Wrapf(err, "unbind2Node get failed node=%s", node)
	}
	//found
	if len(data) > 0 {
		var list pt.ParaNodeBindList
		err = types.Decode(data, &list)
		if err != nil {
			return nil, errors.Wrapf(err, "bind2Node decode failed node=%s", node)
		}
		listCopy := list
		list.Miners = append(list.Miners, a.fromaddr)
		return makeNodeBindReceipt(node, &listCopy, &list), nil
	}
	//unfound
	var list pt.ParaNodeBindList
	list.SuperNode = node
	list.Miners = append(list.Miners, a.fromaddr)
	return makeNodeBindReceipt(node, nil, &list), nil
}

//从超级节点解绑
func (a *action) unbind2Node(node string) (*types.Receipt, error) {
	list, err := a.getBindNodeInfo(node)
	if err != nil {
		return nil, errors.Wrap(err, "unbind2Node")
	}
	newList := &pt.ParaNodeBindList{SuperNode: list.SuperNode}
	for _, m := range list.Miners {
		if m != a.fromaddr {
			newList.Miners = append(newList.Miners, m)
		}
	}
	return makeNodeBindReceipt(node, list, newList), nil

}

func (a *action) getBindNodeInfo(node string) (*pt.ParaNodeBindList, error) {
	key := calcParaBindMinerNode(node)
	data, err := a.db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get key failed node=%s", node)
	}

	var list pt.ParaNodeBindList
	err = types.Decode(data, &list)
	if err != nil {
		return nil, errors.Wrapf(err, "decode failed node=%s", node)
	}
	return &list, nil
}

func (a *action) getBindAddrInfo(node, addr string) (*pt.ParaBindMinerInfo, error) {
	key := calcParaBindMinerAddr(node, addr)
	data, err := a.db.Get(key)
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

func (a *action) bindOp(info *pt.ParaBindMinerInfo) (*types.Receipt, error) {
	if len(info.Addr) > 0 && info.Addr != a.fromaddr {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner addr=%s not from addr %s", info.Addr, a.fromaddr)
	}

	if info.BindCount <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner bindCount nil from addr %s", a.fromaddr)
	}

	ok, err := a.isValidSuperNode(info.TargetAddr)
	if err != nil || !ok {
		return nil, errors.Wrapf(err, "invalid target node=%s", info.TargetAddr)
	}

	key := calcParaBindMinerAddr(info.TargetAddr, a.fromaddr)
	data, err := a.db.Get(key)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	//found, 修改当前的绑定
	if len(data) > 0 {
		var receipt *types.Receipt
		var acct pt.ParaBindMinerInfo
		err = types.Decode(data, &acct)
		if err != nil {
			return nil, errors.Wrapf(err, "bindOp decode for addr=%s", a.fromaddr)
		}

		if info.BindCount == acct.BindCount {
			return nil, errors.Wrapf(types.ErrInvalidParam, "bindOp bind coins not change current=%d, modify=%d",
				acct.BindCount, info.BindCount)
		}

		//释放一部分冻结coins
		if info.BindCount < acct.BindCount {
			receipt, err = a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, (acct.BindCount-info.BindCount)*types.Coin)
			if err != nil {
				return nil, errors.Wrapf(err, "bindOp Active addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, acct.BindCount-info.BindCount)
			}
		}
		//冻结更多
		receipt, err = a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, (info.BindCount-acct.BindCount)*types.Coin)
		if err != nil {
			return nil, errors.Wrapf(err, "bindOp frozen more addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, info.BindCount-acct.BindCount)
		}

		acctCopy := acct
		acct.BindCount = info.BindCount

		r := makeAddrBindReceipt(info.TargetAddr, a.fromaddr, &acctCopy, &acct)

		return mergeReceipt(receipt, r), nil
	}

	//not found, 增加新绑定
	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, info.BindCount*types.Coin)
	if err != nil {
		return nil, errors.Wrapf(err, "bindOp frozen addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, info.BindCount)
	}

	//bind addr
	acct := &pt.ParaBindMinerInfo{
		Addr:       a.fromaddr,
		BindStatus: opBind,
		BindCount:  info.BindCount,
		BlockTime:  a.blocktime,
		TargetAddr: info.TargetAddr,
	}
	rBind := makeAddrBindReceipt(info.TargetAddr, a.fromaddr, nil, acct)
	mergeReceipt(receipt, rBind)

	//增加到列表中
	rList, err := a.bind2Node(info.TargetAddr)
	if err != nil {
		return nil, err
	}
	mergeReceipt(receipt, rList)
	return receipt, nil

}

func (a *action) unBindOp(info *pt.ParaBindMinerInfo) (*types.Receipt, error) {
	acct, err := a.getBindAddrInfo(info.TargetAddr, a.fromaddr)
	if err != nil {
		return nil, err
	}

	cfg := a.api.GetConfig()
	unBindHours := cfg.MGInt("mver.consensus.paracross.unBindTime", a.height)
	if acct.BlockTime-a.blocktime < unBindHours*60*60 {
		return nil, errors.Wrapf(err, "unBindOp unbind time=%d less %d hours than bind time =%d", a.blocktime, unBindHours, acct.BlockTime)
	}

	//unfrozen
	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, acct.BindCount*types.Coin)
	if err != nil {
		return nil, errors.Wrapf(err, "unBindOp addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, acct.BindCount)
	}

	//删除 bind addr
	rUnBind := makeAddrBindReceipt(info.TargetAddr, a.fromaddr, acct, nil)
	mergeReceipt(receipt, rUnBind)

	//从列表删除
	rUnList, err := a.unbind2Node(info.TargetAddr)
	if err != nil {
		return nil, err
	}
	mergeReceipt(receipt, rUnList)

	return receipt, nil
}

func (a *action) bindMiner(info *pt.ParaBindMinerInfo) (*types.Receipt, error) {
	if len(info.TargetAddr) <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner targetAddr should not be nil to addr %s", a.fromaddr)
	}

	if info.BindStatus != opBind && info.BindStatus != opUnBind {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner status=%d not correct", info.BindStatus)
	}

	if info.BindStatus == opBind {
		return a.bindOp(info)
	}
	return a.unBindOp(info)
}

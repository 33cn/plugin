package executor

import (
	"github.com/pkg/errors"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	opBind   = 1
	opUnBind = 2
	opModify = 3
)

//从内存中获取bin状态的miner list
func (a *action) getBindAddrs(nodes []string, statusHeight int64) (bool, map[string][]*pt.ParaBindMinerInfo, error) {
	nodeBinders := make(map[string][]*pt.ParaBindMinerInfo)
	var foundBinder bool
	for _, node := range nodes {
		var minerList []*pt.ParaBindMinerInfo
		list, err := getBindMinerList(a.db, node)
		if err != nil {
			clog.Error("paracross getBindAddrs err", "height", statusHeight, "err", err)
			return false, nil, err
		}
		for _, l := range list {
			//过滤所有bind状态的miner
			if l.BindStatus == opBind {
				foundBinder = true
				minerList = append(minerList, l)
			}
		}
		nodeBinders[node] = minerList
	}
	return foundBinder, nodeBinders, nil
}

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

func makeAddMinerBindNodeListReceipt(miner string, node string, current *pt.ParaMinerBindNodes) *types.Receipt {
	key := calcParaMinerBindNodeList(miner)
	log := &pt.ReceiptParaMinerBindNodeList{
		Miner:   miner,
		Node:    node,
		Current: current,
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaMinerBindNodeList,
				Log: types.Encode(log),
			},
		},
	}
}

func makeAddNodeBindMinerCountReceipt(node string, prev, current *pt.ParaBindNodeInfo) *types.Receipt {
	key := calcParaNodeBindMinerCount(node)
	log := &pt.ReceiptParaBindConsensusNodeInfo{
		NodeAddr: node,
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
				Ty:  pt.TyLogParaBindMinerNode,
				Log: types.Encode(log),
			},
		},
	}
}

func makeNodeBindMinerIndexReceipt(node, miner string, index int64) *types.Receipt {
	key := calcParaNodeBindMinerIndex(node, index)
	log := &pt.ReceiptParaBindIndex{
		SelfAddr: node,
		BindAddr: miner,
		Index:    index,
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&pt.ParaBindAddr{Addr: miner})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaBindMinerIndex,
				Log: types.Encode(log),
			},
		},
	}
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

// node绑定的miner数量

func getBindMinerCount(db dbm.KV, node string) (*pt.ParaBindNodeInfo, error) {
	key := calcParaNodeBindMinerCount(node)
	data, err := db.Get(key)
	if isNotFound(err) {
		return &pt.ParaBindNodeInfo{}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "getBindNodeInfo node=%s", node)
	}

	var info pt.ParaBindNodeInfo
	err = types.Decode(data, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "decode failed node=%s", node)
	}
	return &info, nil
}

func getNodeBindMinerIndexInfo(db dbm.KV, node string, index int64) (*pt.ParaBindAddr, error) {
	key := calcParaNodeBindMinerIndex(node, index)
	data, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "db.get")
	}

	var info pt.ParaBindAddr
	err = types.Decode(data, &info)
	if err != nil {
		return nil, errors.Wrap(err, "decode failed node")
	}
	return &info, nil
}

func getBindMinerList(db dbm.KV, node string) ([]*pt.ParaBindMinerInfo, error) {
	//从db恢复
	//node 绑定挖矿地址数量
	nodeInfo, err := getBindMinerCount(db, node)
	if err != nil {
		return nil, errors.Wrapf(err, "updateGlobalBindMinerInfo.getNodeInfo node=%s", node)
	}
	if nodeInfo.BindTotalCount <= 0 {
		return nil, nil
	}

	var minerList []*pt.ParaBindMinerInfo
	for i := int64(0); i < nodeInfo.BindTotalCount; i++ {
		miner, err := getNodeBindMinerIndexInfo(db, node, i)
		if err != nil {
			return nil, errors.Wrapf(err, "getBindMinerList.getMinerIndex,node=%s,index=%d", node, i)
		}
		minerInfo, err := getBindAddrInfo(db, node, miner.Addr)
		if err != nil {
			return nil, errors.Wrapf(err, "getBindMinerList.getBindAddrInfo,node=%s,addr=%s", node, miner.Addr)
		}
		minerList = append(minerList, minerInfo)
	}
	return minerList, nil
}

func getMinerBindNodeList(db dbm.KV, miner string) (*pt.ParaMinerBindNodes, error) {
	key := calcParaMinerBindNodeList(miner)
	data, err := db.Get(key)
	if isNotFound(err) {
		return &pt.ParaMinerBindNodes{}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "getBindNodeInfo node=%s", miner)
	}

	var info pt.ParaMinerBindNodes
	err = types.Decode(data, &info)
	if err != nil {
		return nil, errors.Wrapf(err, "decode failed node=%s", miner)
	}
	return &info, nil
}

func (a *action) addMinerBindNode(miner, node string) (*types.Receipt, error) {
	nodes, err := getMinerBindNodeList(a.db, miner)
	if err != nil {
		return nil, errors.Wrapf(err, "addMinerBindNode miner=%s", miner)
	}
	nodes.Nodes = append(nodes.Nodes, node)
	return makeAddMinerBindNodeListReceipt(miner, node, nodes), nil
}

func (a *action) addNodeBindMinerCount(node, miner string) (*types.Receipt, int64, error) {
	receipt := &types.Receipt{Ty: types.ExecOk}
	bindInfo, err := getBindMinerCount(a.db, node)
	if err != nil {
		return nil, 0, err
	}

	//new index --> miner
	rIdx := makeNodeBindMinerIndexReceipt(node, miner, bindInfo.BindTotalCount)
	mergeReceipt(receipt, rIdx)

	//只增加绑定index，如果有bind miner退出或又加入，只更新对应miner状态，绑定关系不变
	//add totalCount
	rNode := makeAddNodeBindMinerCountReceipt(node, bindInfo, &pt.ParaBindNodeInfo{BindTotalCount: bindInfo.BindTotalCount + 1})
	mergeReceipt(receipt, rNode)

	return receipt, bindInfo.BindTotalCount, nil
}

func (a *action) addNewBind(cmd *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	coinPrecision := a.api.GetConfig().GetCoinPrecision()
	receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, cmd.BindCoins*coinPrecision)
	if err != nil {
		return nil, errors.Wrapf(err, "addNew bindOp frozen addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, cmd.BindCoins)
	}

	//增加node绑定miner数量
	rNode, newGlobalIndex, err := a.addNodeBindMinerCount(cmd.TargetNode, a.fromaddr)
	if err != nil {
		return nil, errors.Wrapf(err, "addBindCount for %s to %s", cmd.TargetNode, a.fromaddr)
	}
	mergeReceipt(receipt, rNode)

	//增加node绑定miner
	newer := &pt.ParaBindMinerInfo{
		Addr:          a.fromaddr,
		BindStatus:    opBind,
		BindCoins:     cmd.BindCoins,
		BlockTime:     a.blocktime,
		BlockHeight:   a.height,
		ConsensusNode: cmd.TargetNode,
		GlobalIndex:   newGlobalIndex,
	}
	//miner --> index and detail info
	rBind := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, nil, newer)
	mergeReceipt(receipt, rBind)

	//增加miner 绑定node信息
	rMiner, err := a.addMinerBindNode(a.fromaddr, cmd.TargetNode)
	if err != nil {
		return nil, errors.Wrapf(err, "addBindCount for %s to %s", a.fromaddr, cmd.TargetNode)
	}
	mergeReceipt(receipt, rMiner)

	return receipt, nil
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
		return nil, errors.Wrapf(err, "getBindAddrInfo,node=%s,from=%s", cmd.TargetNode, a.fromaddr)
	}

	//found, 修改当前的绑定
	if current != nil {
		//已经是绑定状态，不允许重复绑定
		if current.BindStatus == opBind {
			return nil, errors.Wrapf(types.ErrNotAllow, "already binded,node=%s,addr=%s", cmd.TargetNode, a.fromaddr)
		}

		//处理解绑定状态
		//新冻结资产
		coinPrecision := a.api.GetConfig().GetCoinPrecision()
		receipt, err := a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, cmd.BindCoins*coinPrecision)
		if err != nil {
			return nil, errors.Wrapf(err, "bindOp frozen addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, cmd.BindCoins)
		}
		acctCopy := *current
		current.BindStatus = opBind
		current.BindCoins = cmd.BindCoins
		current.BlockTime = a.blocktime
		current.BlockHeight = a.height
		r := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, &acctCopy, current)
		return mergeReceipt(receipt, r), nil
	}

	//增加新绑定
	return a.addNewBind(cmd)

}

func (a *action) modifyBindOp(cmd *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	if cmd.BindCoins <= 0 {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner BindCoins=0 from addr %s", a.fromaddr)
	}

	err := a.isValidSuperNode(cmd.TargetNode)
	if err != nil {
		return nil, err
	}

	current, err := getBindAddrInfo(a.db, cmd.TargetNode, a.fromaddr)
	if err != nil {
		return nil, errors.Wrapf(err, "getBindAddrInfo node=%s,binder=%s", cmd.TargetNode, a.fromaddr)
	}

	if current.BindStatus != opBind {
		return nil, errors.Wrapf(types.ErrNotAllow, "not bind status, node=%s,binder=%s", cmd.TargetNode, a.fromaddr)
	}

	var receipt *types.Receipt
	if cmd.BindCoins == current.BindCoins {
		return nil, errors.Wrapf(types.ErrInvalidParam, "bind coins same current=%d, cmd=%d", current.BindCoins, cmd.BindCoins)
	}

	coinPrecision := a.api.GetConfig().GetCoinPrecision()
	//释放一部分coins
	if cmd.BindCoins < current.BindCoins {
		receipt, err = a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, (current.BindCoins-cmd.BindCoins)*coinPrecision)
		if err != nil {
			return nil, errors.Wrapf(err, "bindOp Active addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, current.BindCoins-cmd.BindCoins)
		}
	} else {
		//冻结更多
		receipt, err = a.coinsAccount.ExecFrozen(a.fromaddr, a.execaddr, (cmd.BindCoins-current.BindCoins)*coinPrecision)
		if err != nil {
			return nil, errors.Wrapf(err, "bindOp frozen more addr=%s,execaddr=%s,coins=%d", a.fromaddr, a.execaddr, cmd.BindCoins-current.BindCoins)
		}
	}

	acctCopy := *current
	current.BindCoins = cmd.BindCoins
	r := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, &acctCopy, current)
	return mergeReceipt(receipt, r), nil
}

func (a *action) unBindOp(cmd *pt.ParaBindMinerCmd) (*types.Receipt, error) {
	minerInfo, err := getBindAddrInfo(a.db, cmd.TargetNode, a.fromaddr)
	if err != nil {
		return nil, err
	}

	if minerInfo.BindStatus != opBind {
		return nil, errors.Wrapf(types.ErrNotAllow, "unBindOp,current addr is unbind status")
	}

	cfg := a.api.GetConfig()
	unBindHours := cfg.MGInt("mver.consensus.paracross.unBindTime", a.height)
	if a.blocktime-minerInfo.BlockTime < unBindHours*60*60 {
		return nil, errors.Wrapf(types.ErrNotAllow, "unBindOp unbind time=%d less %d hours than bind time =%d", a.blocktime, unBindHours, minerInfo.BlockTime)
	}

	//unfrozen
	receipt, err := a.coinsAccount.ExecActive(a.fromaddr, a.execaddr, minerInfo.BindCoins*cfg.GetCoinPrecision())
	if err != nil {
		return nil, errors.Wrapf(err, "unBindOp addr=%s,execaddr=%s,count=%d", a.fromaddr, a.execaddr, minerInfo.BindCoins)
	}

	//删除 bind addr
	//由于kvmvcc的原因，不能通过把一个key值=nil的方式删除，kvmvcc这样是删除了当前版本，就会查询更早的版本，&struct{}也不行，len=0 也被认为是删除了的
	acctCopy := *minerInfo
	minerInfo.BindStatus = opUnBind
	minerInfo.BlockHeight = a.height
	minerInfo.BlockTime = a.blocktime
	minerInfo.BindCoins = 0
	rUnBind := makeAddrBindReceipt(cmd.TargetNode, a.fromaddr, &acctCopy, minerInfo)
	mergeReceipt(receipt, rUnBind)

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

	switch info.BindAction {
	case opBind:
		return a.bindOp(info)
	case opUnBind:
		return a.unBindOp(info)
	case opModify:
		return a.modifyBindOp(info)
	default:
		return nil, errors.Wrapf(types.ErrInvalidParam, "bindMiner action=%d not support", info.BindAction)
	}

}

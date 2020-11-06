package executor

import (
	"fmt"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func makeParaSupervisionNodeGroupReceipt(title string, prev, current *types.ConfigItem) *types.Receipt {
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	log := &types.ReceiptConfig{Prev: prev, Current: current}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaSupervisionNodeGroupAddrsUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func makeSupervisionNodeConfigReceipt(fromAddr string, config *pt.ParaNodeAddrConfig, prev, current *pt.ParaNodeIdStatus) *types.Receipt {
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
				Ty:  pt.TyLogParaSupervisionNodeConfig,
				Log: types.Encode(log),
			},
		},
	}
}

func makeParaSupervisionNodeStatusReceipt(fromAddr string, prev, current *pt.ParaNodeAddrIdStatus) *types.Receipt {
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
				Ty:  pt.TyLogParaSupervisionNodeStatusUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func getSupervisionNodeID(db dbm.KV, title string, id string) (*pt.ParaNodeIdStatus, error) {
	id = calcParaSupervisionNodeIDKey(title, id)
	val, err := getDb(db, []byte(id))
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

func (a *action) updateSupervisionNodeGroup(title, addr string, add bool) (*types.Receipt, error) {
	var item types.ConfigItem
	key := calcParaSupervisionNodeGroupAddrsKey(title)

	value, err := a.db.Get(key)
	if err != nil {
		return nil, err
	}

	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			clog.Error("updateSupervisionNodeGroup", "decode db key", key)
			return nil, err
		}
	}

	copyValue := *item.GetArr()
	copyItem := item
	copyItem.Value = &types.ConfigItem_Arr{Arr: &copyValue}

	receipt := &types.Receipt{Ty: types.ExecOk}
	item.Addr = addr
	if add {
		item.GetArr().Value = append(item.GetArr().Value, addr)
		clog.Info("updateSupervisionNodeGroup add", "addr", addr)
	} else {
		item.GetArr().Value = make([]string, 0)
		for _, value := range copyItem.GetArr().Value {
			if value != addr {
				item.GetArr().Value = append(item.GetArr().Value, value)
			}
		}
		clog.Info("updateSupervisionNodeGroup delete", "addr", addr)
	}
	err = a.db.Set(key, types.Encode(&item))
	if err != nil {
		return nil, errors.Wrapf(err, "updateNodeGroup set dbkey=%s", key)
	}
	r := makeParaSupervisionNodeGroupReceipt(title, &copyItem, &item)
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

func (a *action) checkValidSupervisionNode(config *pt.ParaNodeAddrConfig) (bool, error) {
	key := calcParaSupervisionNodeGroupAddrsKey(config.Title)
	nodes, _, err := getNodes(a.db, key)
	if err != nil && !(isNotFound(err) || errors.Cause(err) == pt.ErrTitleNotExist) {
		return false, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}

	if validNode(config.Addr, nodes) {
		return true, nil
	}
	return false, nil
}

func (a *action) checkSupervisionNodeGroupExist(title string) (bool, error) {
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	value, err := a.db.Get(key)
	if err != nil && !isNotFound(err) {
		return false, err
	}

	if value != nil {
		var item types.ConfigItem
		err = types.Decode(value, &item)
		if err != nil {
			clog.Error("updateSupervisionNodeGroup", "decode db key", key)
			return false, err
		}

		return true, nil
	}

	return false, nil
}

func (a *action) supervisionNodeGroupCreate(title, targetAddrs string) (*types.Receipt, error) {
	var item types.ConfigItem
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, targetAddrs)
	item.Addr = a.fromaddr

	receipt := makeParaSupervisionNodeGroupReceipt(title, nil, &item)
	return receipt, nil
}

//由于propasal id 和quit id分开，quit id不知道对应addr　proposal id的coinfrozen信息，需要维护一个围绕addr的数据库结构信息
func (a *action) updateSupervisionNodeAddrStatus(stat *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	addrStat, err := getNodeAddr(a.db, stat.Title, stat.TargetAddr)
	if err != nil {
		if !isNotFound(err) {
			return nil, errors.Wrapf(err, "nodeAddr:%s get error", stat.TargetAddr)
		}
		addrStat = &pt.ParaNodeAddrIdStatus{}
		addrStat.Title = stat.Title
		addrStat.Addr = stat.TargetAddr
		addrStat.BlsPubKey = stat.BlsPubKey
		addrStat.Status = pt.ParacrossSupervisionNodeApprove
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaSupervisionNodeStatusReceipt(a.fromaddr, nil, addrStat), nil
	}

	preStat := *addrStat
	if stat.Status == pt.ParacrossSupervisionNodeQuit {
		proposalStat, err := getNodeID(a.db, addrStat.ProposalId)
		if err != nil {
			return nil, errors.Wrapf(err, "nodeAddr:%s quiting wrong proposeid:%s", stat.TargetAddr, addrStat.ProposalId)
		}

		addrStat.Status = stat.Status
		addrStat.QuitId = stat.Id
		receipt := makeParaSupervisionNodeStatusReceipt(a.fromaddr, &preStat, addrStat)

		cfg := a.api.GetConfig()
		if !cfg.IsPara() {
			r, err := a.nodeGroupCoinsActive(proposalStat.FromAddr, proposalStat.CoinsFrozen, 1)
			if err != nil {
				return nil, err
			}
			receipt = mergeReceipt(receipt, r)
		}
		return receipt, nil
	} else {
		addrStat.Status = stat.Status
		addrStat.ProposalId = stat.Id
		addrStat.QuitId = ""
		return makeParaSupervisionNodeStatusReceipt(a.fromaddr, &preStat, addrStat), nil
	}
}

func (a *action) supervisionNodeApply(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	// 必须要有授权节点 监督节点才有意义 判断是否存在授权节点
	addrExist, err := a.checkValidNode(config)
	if err != nil {
		return nil, err
	}

	// 不能跟授权节点一致
	if addrExist {
		clog.Debug("supervisionNodeGroup Apply", "config.Addr", config.Addr, "err", "config.Addr existed in super group")
		return nil, pt.ErrParaNodeAddrExisted
	}

	// 判断 node 是否已经申请
	addrExist, err = a.checkValidSupervisionNode(config)
	if err != nil {
		fmt.Println("err:", err)
	}
	if addrExist {
		clog.Debug("supervisionNodeGroup Apply", "config.Addr", config.Addr, "err", "config.Addr existed in supervision group")
		return nil, pt.ErrParaSupervisionNodeAddrExisted
	}

	// 在主链上冻结金额
	receipt := &types.Receipt{Ty: types.ExecOk}
	cfg := a.api.GetConfig()
	if !cfg.IsPara() {
		r, err := a.nodeGroupCoinsFrozen(a.fromaddr, config.CoinsFrozen, 1)
		if err != nil {
			return nil, err
		}
		receipt = mergeReceipt(receipt, r)
	}

	stat := &pt.ParaNodeIdStatus{
		Id:          calcParaSupervisionNodeIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParacrossSupervisionNodeApply,
		Title:       config.Title,
		TargetAddr:  config.Addr,
		BlsPubKey:   config.BlsPubKey,
		CoinsFrozen: config.CoinsFrozen,
		FromAddr:    a.fromaddr,
		Height:      a.height,
	}
	r := makeSupervisionNodeConfigReceipt(a.fromaddr, config, nil, stat)
	receipt = mergeReceipt(receipt, r)

	return receipt, nil
}

func (a *action) supervisionNodeApprove(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	//只在主链检查
	if !cfg.IsPara() && !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "node group approve not supervision manager:%s", a.fromaddr)
	}

	apply, err := getSupervisionNodeID(a.db, config.Title, config.Id)
	if err != nil {
		return nil, err
	}
	if config.Title != apply.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, apply.Title)
	}
	if apply.CoinsFrozen < config.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "id not enough coins apply:%d,config:%d", apply.CoinsFrozen, config.CoinsFrozen)
	}

	// 判断监督账户组是否已经存在
	exist, err := a.checkSupervisionNodeGroupExist(config.Title)
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	if !exist {
		// 监督账户组不存在
		r, err := a.supervisionNodeGroupCreate(apply.Title, apply.TargetAddr)
		if err != nil {
			return nil, errors.Wrapf(err, "nodegroup create:title:%s,addrs:%s", config.Title, apply.TargetAddr)
		}
		receipt = mergeReceipt(receipt, r)
	} else {
		// 监督账户组已经存在
		r, err := a.updateSupervisionNodeGroup(config.Title, apply.TargetAddr, true)
		if err != nil {
			return nil, err
		}
		receipt = mergeReceipt(receipt, r)
	}

	copyStat := proto.Clone(apply).(*pt.ParaNodeIdStatus)
	apply.Status = pt.ParacrossSupervisionNodeApprove
	apply.Height = a.height

	r, err := a.updateSupervisionNodeAddrStatus(apply)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)

	r = makeSupervisionNodeConfigReceipt(a.fromaddr, config, copyStat, apply)
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

func (a *action) supervisionNodeQuit(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	addrExist, err := a.checkValidSupervisionNode(config)
	if err != nil {
		return nil, err
	}
	if !addrExist {
		return nil, errors.Wrapf(pt.ErrParaSupervisionNodeAddrNotExisted, "nodeAddr not existed:%s", config.Addr)
	}

	status, err := getNodeAddr(a.db, config.Title, config.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "nodeAddr:%s get error", config.Addr)
	}
	if status.Status != pt.ParacrossSupervisionNodeApprove {
		return nil, errors.Wrapf(pt.ErrParaSupervisionNodeAddrNotExisted, "nodeAddr:%s status:%d", config.Addr, status.Status)
	}

	cfg := a.api.GetConfig()
	stat := &pt.ParaNodeIdStatus{
		Id:         status.ProposalId,
		Status:     pt.ParacrossSupervisionNodeQuit,
		Title:      config.Title,
		TargetAddr: config.Addr,
		FromAddr:   a.fromaddr,
		Height:     a.height,
	}

	//只能提案发起人或超级节点可以撤销
	if a.fromaddr != status.Addr && !cfg.IsPara() && !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "id create by:%s,not by:%s", status.Addr, a.fromaddr)
	}

	if config.Title != status.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, status.Title)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}

	// node quit后，如果committx满足2/3目标，自动触发commitDone 后期增加
	r, err := a.loopCommitTxDone(config.Title)
	if err != nil {
		clog.Error("updateSupervisionNodeGroup.loopCommitTxDone", "title", config.Title, "err", err.Error())
	}
	receipt = mergeReceipt(receipt, r)

	r, err = a.updateSupervisionNodeGroup(config.Title, stat.TargetAddr, false)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)

	r, err = a.updateSupervisionNodeAddrStatus(stat)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)

	r = makeSupervisionNodeConfigReceipt(a.fromaddr, config, nil, stat)
	receipt = mergeReceipt(receipt, r)

	return receipt, nil
}

func (a *action) supervisionNodeCancel(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	status, err := getSupervisionNodeID(a.db, config.Title, config.Id)
	if err != nil {
		return nil, err
	}

	//只能提案发起人可以撤销
	if a.fromaddr != status.FromAddr {
		return nil, errors.Wrapf(types.ErrNotAllow, "id create by:%s,not by:%s", status.FromAddr, a.fromaddr)
	}
	if config.Title != status.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, status.Title)
	}
	if status.Status != pt.ParacrossSupervisionNodeApply {
		return nil, errors.Wrapf(pt.ErrParaNodeOpStatusWrong, "config id:%s,status:%d", config.Id, status.Status)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//main chain
	if !cfg.IsPara() {
		r, err := a.nodeGroupCoinsActive(status.FromAddr, status.CoinsFrozen, 1)
		if err != nil {
			return nil, err
		}
		receipt = mergeReceipt(receipt, r)
	}

	copyStat := proto.Clone(status).(*pt.ParaNodeIdStatus)
	status.Status = pt.ParacrossSupervisionNodeCancel
	status.Height = a.height

	r := makeSupervisionNodeConfigReceipt(a.fromaddr, config, copyStat, status)
	receipt = mergeReceipt(receipt, r)

	return receipt, nil
}

func (a *action) SupervisionNodeConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !validTitle(cfg, config.Title) {
		return nil, pt.ErrInvalidTitle
	}
	if !types.IsParaExecName(string(a.tx.Execer)) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}
	if (config.Op == pt.ParacrossSupervisionNodeApprove || config.Op == pt.ParacrossSupervisionNodeCancel) && config.Id == "" {
		return nil, types.ErrInvalidParam
	}

	if config.Op == pt.ParacrossSupervisionNodeApply {
		return a.supervisionNodeApply(config)
	} else if config.Op == pt.ParacrossSupervisionNodeApprove {
		return a.supervisionNodeApprove(config)
	} else if config.Op == pt.ParacrossSupervisionNodeQuit {
		// 退出 group
		return a.supervisionNodeQuit(config)
	} else if config.Op == pt.ParacrossSupervisionNodeCancel {
		// 撤销未批准的申请
		return a.supervisionNodeCancel(config)
	}

	return nil, pt.ErrParaUnSupportNodeOper
}

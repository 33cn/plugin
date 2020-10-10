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

func makeSupervisionNodeIDReceipt(addr string, prev, current *pt.ParaNodeGroupStatus) *types.Receipt {
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
				Ty:  pt.TyLogParaSupervisionNodeGroupConfig,
				Log: types.Encode(log),
			},
		},
	}
}

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

func makeStageSupervisionGroupReceipt(prev, current *pt.SelfConsensStages) *types.Receipt {
	key := []byte(fmt.Sprintf(paraSupervisionSelfConsensStages))
	log := &pt.ReceiptSelfConsStagesUpdate{Prev: prev, Current: current}
	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  pt.TyLogParaStageSupervisionGroupUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func makeParaSupervisionNodeGroupStatusReceipt(title string, addr string, prev, current *pt.ParaNodeGroupStatus) *types.Receipt {
	key := calcParaSupervisionNodeGroupStatusKey(title)
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
				Ty:  pt.TyLogParaSupervisionNodeGroupStatusUpdate,
				Log: types.Encode(log),
			},
		},
	}
}

func getSupervisionNodeGroupStatus(db dbm.KV, title string) (*pt.ParaNodeGroupStatus, error) {
	key := calcParaSupervisionNodeGroupStatusKey(title)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeGroupStatus
	err = types.Decode(val, &status)
	return &status, err
}

//func getSupervisionNodeAddr(db dbm.KV, title, addr string) (*pt.ParaNodeAddrIdStatus, error) {
//	key := calcParaSupervisionNodeAddrKey(title, addr)
//	val, err := db.Get(key)
//	if err != nil {
//		return nil, err
//	}
//
//	var status pt.ParaNodeAddrIdStatus
//	err = types.Decode(val, &status)
//	return &status, err
//}

func supervisionSelfConsentInitStage(cfg *types.Chain33Config) *types.Receipt {
	isEnable := cfg.IsEnable(pt.ParaConsSubConf + "." + pt.ParaSelfConsInitConf)
	stage := &pt.SelfConsensStage{StartHeight: 0, Enable: pt.ParaConfigYes}
	if isEnable {
		stage.Enable = pt.ParaConfigNo
	}
	stages := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{stage}}
	return makeStageSupervisionGroupReceipt(nil, stages)
}

func getSupervisionNodeGroupID(cfg *types.Chain33Config, db dbm.KV, title string, height int64, id string) (*pt.ParaNodeIdStatus, error) {
	if pt.IsParaForkHeight(cfg, height, pt.ForkLoopCheckCommitTxDone) {
		id = calcParaSupervisionNodeGroupIDKey(title, id)
	}
	val, err := getDb(db, []byte(id))
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

func updateSupervisionNodeGroup(db dbm.KV, title, addr string, add bool) (*types.Receipt, error) {
	var item types.ConfigItem

	key := calcParaSupervisionNodeGroupAddrsKey(title)
	value, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			clog.Error("updateSupervisionNodeGroup", "decode db key", key)
			return nil, err // types.ErrBadConfigValue
		}
	}

	copyValue := *item.GetArr()
	copyItem := item
	copyItem.Value = &types.ConfigItem_Arr{Arr: &copyValue}

	if add {
		item.GetArr().Value = append(item.GetArr().Value, addr)
		item.Addr = addr
		clog.Info("updateSupervisionNodeGroup add", "addr", addr, "from", copyItem.GetArr().Value, "to", item.GetArr().Value)
	} else {
		item.Addr = addr
		item.GetArr().Value = make([]string, 0)
		for _, value := range copyItem.GetArr().Value {
			if value != addr {
				item.GetArr().Value = append(item.GetArr().Value, value)
			}
		}
		clog.Info("updateSupervisionNodeGroup delete", "addr", addr)
	}
	err = db.Set(key, types.Encode(&item))
	if err != nil {
		return nil, errors.Wrapf(err, "updateNodeGroup set dbkey=%s", key)
	}
	return makeParaSupervisionNodeGroupReceipt(title, &copyItem, &item), nil
}

func (a *action) checkValidSupervisionNode(config *pt.ParaNodeAddrConfig) (bool, error) {
	nodes, _, err := getParacrossSupervisonNodes(a.db, config.Title)
	if err != nil && !isNotFound(err) {
		return false, errors.Wrapf(err, "getNodes for title:%s", config.Title)
	}

	if validNode(config.Addr, nodes) {
		return true, nil
	}
	return false, nil
}

func (a *action) checkSupervisionNodeGroupExist(title string) (error, bool) {
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	value, err := a.db.Get(key)
	if err != nil && !isNotFound(err) {
		return err, false
	}

	if value != nil {
		return nil, true
	}

	return nil, false
}

func (a *action) supervisionNodeGroupCreate(status *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	var item types.ConfigItem
	key := calcParaSupervisionNodeGroupAddrsKey(status.Title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, status.TargetAddr)
	item.Addr = a.fromaddr

	receipt := makeParaSupervisionNodeGroupReceipt(status.Title, nil, &item)

	status.Status = pt.ParacrossSupervisionNodeApprove
	r := makeSupervisionNodeConfigReceipt(a.fromaddr, nil, nil, status)
	receipt = mergeReceipt(receipt, r)

	r, err := a.updateSupervisionNodeAddrStatus(status)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)
	return receipt, nil
}

func getSupervisionNodeAddr(db dbm.KV, title, addr string) (*pt.ParaNodeAddrIdStatus, error) {
	key := calcParaNodeAddrKey(title, addr)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeAddrIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

//由于propasal id 和quit id分开，quit id不知道对应addr　proposal id的coinfrozen信息，需要维护一个围绕addr的数据库结构信息
func (a *action) updateSupervisionNodeAddrStatus(stat *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	//cfg := a.api.GetConfig()
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

		addrStat.Status = pt.ParacrossSupervisionNodeQuit
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
		stat.Status = pt.ParacrossSupervisionNodeApprove
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

	// 是否已经申请
	addrExist, err = a.checkValidSupervisionNode(config)
	if err != nil {
		fmt.Println("err:", err)
	}
	if addrExist {
		clog.Debug("supervisionNodeGroup Apply", "config.Addr", config.Addr, "err", "config.Addr existed in supervision group")
		return nil, pt.ErrParaSupervisionNodeAddrExisted
	}

	// 判断和监督组冻结金额是否一致
	nodeGroupStatus, err := getSupervisionNodeGroupStatus(a.db, config.Title)
	if err != nil && !isNotFound(err) {
		return nil, errors.Wrapf(pt.ErrParaSupervisionNodeGroupNotSet, "nodegroup not exist:%s", config.Title)
	}

	if nodeGroupStatus != nil && config.CoinsFrozen < nodeGroupStatus.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough,
			"coinFrozen not enough:%d,expected:%d", config.CoinsFrozen, nodeGroupStatus.CoinsFrozen)
	}

	// 在主链上冻结金额
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

	stat := &pt.ParaNodeIdStatus{
		Id:          calcParaSupervisionNodeGroupIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParacrossSupervisionNodeApply,
		Title:       config.Title,
		TargetAddr:  config.Addr,
		BlsPubKey:   config.BlsPubKey,
		CoinsFrozen: config.CoinsFrozen,
		FromAddr:    a.fromaddr,
		Height:      a.height,
	}

	r := makeSupervisionNodeConfigReceipt(a.fromaddr, config, nil, stat)
	//r := makeSupervisionNodeIDReceipt(a.fromaddr, nil, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)
	return receipt, nil
}

func (a *action) supervisionNodeApprove(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	//只在主链检查
	if !cfg.IsPara() && !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "node group approve not supervision manager:%s", a.fromaddr)
	}

	apply, err := getSupervisionNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
	if err != nil {
		return nil, err
	}

	if config.Title != apply.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, apply.Title)
	}

	// 判断监督账户组是否已经存在
	err, exist := a.checkSupervisionNodeGroupExist(config.Title)
	if err != nil {
		return nil, err
	}

	// 监督账户组已经不存在
	if !exist {
		if apply.CoinsFrozen < config.CoinsFrozen {
			return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "id not enough coins apply:%d,config:%d", apply.CoinsFrozen, config.CoinsFrozen)
		}

		receipt := &types.Receipt{Ty: types.ExecOk}
		//create the supervision node group
		r, err := a.supervisionNodeGroupCreate(apply)
		if err != nil {
			return nil, errors.Wrapf(err, "nodegroup create:title:%s,addrs:%s", config.Title, apply.TargetAddr)
		}
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		stat := &pt.ParaNodeGroupStatus{
			Id:          apply.Id,
			Status:      apply.Status,
			Title:       apply.Title,
			TargetAddrs: apply.TargetAddr,
			BlsPubKeys:  apply.BlsPubKey,
			CoinsFrozen: apply.CoinsFrozen,
			FromAddr:    apply.FromAddr,
			Height:      apply.Height,
		}

		copyStat := *stat
		stat.Status = pt.ParacrossSupervisionNodeApprove
		apply.Status = pt.ParacrossSupervisionNodeApprove
		apply.Height = a.height

		r = makeSupervisionNodeIDReceipt(a.fromaddr, &copyStat, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		r = makeParaSupervisionNodeGroupStatusReceipt(config.Title, a.fromaddr, nil, stat)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)

		//不允许主链成功平行链失败导致不一致的情况，这里如果失败则手工设置init stage 默认设置自共识
		if cfg.IsPara() {
			r = supervisionSelfConsentInitStage(cfg)
			receipt.KV = append(receipt.KV, r.KV...)
			receipt.Logs = append(receipt.Logs, r.Logs...)
		}

		return receipt, nil
	}

	// 监督账户组已经存在
	copyStat := proto.Clone(apply).(*pt.ParaNodeIdStatus)
	receipt := &types.Receipt{Ty: types.ExecOk}

	r, err := updateSupervisionNodeGroup(a.db, config.Title, apply.TargetAddr, true)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)

	r, err = a.updateSupervisionNodeAddrStatus(apply)
	if err != nil {
		return nil, err
	}
	receipt = mergeReceipt(receipt, r)

	apply.Status = pt.ParacrossSupervisionNodeApprove
	apply.Height = a.height

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

	status, err := getSupervisionNodeAddr(a.db, config.Title, config.Addr)
	if err != nil {
		return nil, errors.Wrapf(err, "nodeAddr:%s get error", config.Addr)
	}
	if status.Status != pt.ParacrossSupervisionNodeApprove {
		return nil, errors.Wrapf(pt.ErrParaSupervisionNodeAddrNotExisted, "nodeAddr:%s status:%d", config.Addr, status.Status)
	}

	cfg := a.api.GetConfig()
	stat := &pt.ParaNodeIdStatus{
		Id:         calcParaSupervisionNodeGroupIDKey(config.Title, common.ToHex(a.txhash)),
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
	r, err := updateSupervisionNodeGroup(a.db, config.Title, stat.TargetAddr, false)
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
	status, err := getSupervisionNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
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
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	copyStat := proto.Clone(status).(*pt.ParaNodeIdStatus)
	status.Status = pt.ParacrossSupervisionNodeCancel
	status.Height = a.height

	r := makeSupervisionNodeConfigReceipt(a.fromaddr, config, copyStat, status)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) SupervisionNodeGroupConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
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

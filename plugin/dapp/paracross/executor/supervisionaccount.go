package executor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func makeSupervisionNodeGroupIDReceipt(addr string, prev, current *pt.ParaNodeGroupStatus) *types.Receipt {
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
	key := calcParaSupervisionNodeAddrKey(current.Title, current.Addr)
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

func getSupervisionNodeAddr(db dbm.KV, title, addr string) (*pt.ParaNodeAddrIdStatus, error) {
	key := calcParaSupervisionNodeAddrKey(title, addr)
	val, err := db.Get(key)
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeAddrIdStatus
	err = types.Decode(val, &status)
	return &status, err
}

func supervisionSelfConsentInitStage(cfg *types.Chain33Config) *types.Receipt {
	isEnable := cfg.IsEnable(pt.ParaConsSubConf + "." + pt.ParaSelfConsInitConf)
	stage := &pt.SelfConsensStage{StartHeight: 0, Enable: pt.ParaConfigYes}
	if isEnable {
		stage.Enable = pt.ParaConfigNo
	}
	stages := &pt.SelfConsensStages{Items: []*pt.SelfConsensStage{stage}}
	return makeStageSupervisionGroupReceipt(nil, stages)
}

func getSupervisionNodeGroupID(cfg *types.Chain33Config, db dbm.KV, title string, height int64, id string) (*pt.ParaNodeGroupStatus, error) {
	if pt.IsParaForkHeight(cfg, height, pt.ForkLoopCheckCommitTxDone) {
		id = calcParaSupervisionNodeGroupIDKey(title, id)
	}
	val, err := getDb(db, []byte(id))
	if err != nil {
		return nil, err
	}

	var status pt.ParaNodeGroupStatus
	err = types.Decode(val, &status)
	return &status, err
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

func (a *action) checkSupervisionNodeGroupExist(title string) error {
	key := calcParaSupervisionNodeGroupAddrsKey(title)
	_, err := a.db.Get(key)
	if err != nil && !isNotFound(err) {
		return err
	}

	//if value != nil {
	//	clog.Error("node group apply, group existed")
	//	return pt.ErrParaSupervisionNodeGroupExisted
	//}

	return nil
}

func (a *action) supervisionNodeGroupCreate(status *pt.ParaNodeGroupStatus) (*types.Receipt, error) {
	nodes := strings.Split(status.TargetAddrs, ",")

	var item types.ConfigItem
	key := calcParaSupervisionNodeGroupAddrsKey(status.Title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, nodes...)
	item.Addr = a.fromaddr

	receipt := makeParaSupervisionNodeGroupReceipt(status.Title, nil, &item)

	var blsPubKeys []string
	if len(status.BlsPubKeys) > 0 {
		blsPubKeys = strings.Split(status.BlsPubKeys, ",")
	}

	//update addr status
	for i, addr := range nodes {
		stat := &pt.ParaNodeIdStatus{
			Id:          status.Id + "-" + strconv.Itoa(i),
			Status:      pt.ParacrossSupervisionNodeApprove,
			Title:       status.Title,
			TargetAddr:  addr,
			CoinsFrozen: status.CoinsFrozen,
			FromAddr:    status.FromAddr,
			Height:      a.height}
		if len(blsPubKeys) > 0 {
			stat.BlsPubKey = blsPubKeys[i]
		}
		r := makeSupervisionNodeConfigReceipt(a.fromaddr, nil, nil, stat)
		receipt = mergeReceipt(receipt, r)

		r, err := a.updateSupervisionNodeAddrStatus(stat)
		if err != nil {
			return nil, err
		}
		receipt = mergeReceipt(receipt, r)
	}
	return receipt, nil
}

//由于propasal id 和quit id分开，quit id不知道对应addr　proposal id的coinfrozen信息，需要维护一个围绕addr的数据库结构信息
func (a *action) updateSupervisionNodeAddrStatus(stat *pt.ParaNodeIdStatus) (*types.Receipt, error) {
	//cfg := a.api.GetConfig()
	addrStat, err := getSupervisionNodeAddr(a.db, stat.Title, stat.TargetAddr)
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
	stat.Status = pt.ParacrossSupervisionNodeApprove
	addrStat.ProposalId = stat.Id
	addrStat.QuitId = ""
	return makeParaSupervisionNodeStatusReceipt(a.fromaddr, &preStat, addrStat), nil
}

func (a *action) supervisionNodeGroupApply(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	// 不能跟授权节点一致
	addrExist, err := a.checkValidNode(config)
	if err != nil {
		return nil, err
	}
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

	// 判断申请节点之前没有申请或者状态不是申请退出
	addrStat, err := getSupervisionNodeAddr(a.db, config.Title, config.Addr)
	if err != nil && !isNotFound(err) {
		return nil, errors.Wrapf(err, "nodeJoin get title=%s,nodeAddr=%s", config.Title, config.Addr)
	}
	if addrStat != nil && addrStat.Status != pt.ParacrossSupervisionNodeQuit {
		return nil, errors.Wrapf(pt.ErrParaNodeAddrExisted, "nodeJoin nodeAddr existed:%s,status:%d", config.Addr, addrStat.Status)
	}

	targetAddrs := ""
	blsPubKeys := ""
	if nodeGroupStatus != nil {
		targetAddrs = nodeGroupStatus.TargetAddrs + ","
		blsPubKeys = nodeGroupStatus.BlsPubKeys + ","
	}
	targetAddrs += config.Addr
	blsPubKeys += config.BlsPubKey
	stat := &pt.ParaNodeGroupStatus{
		Id:          calcParaSupervisionNodeGroupIDKey(config.Title, common.ToHex(a.txhash)),
		Status:      pt.ParacrossSupervisionNodeApply,
		Title:       config.Title,
		TargetAddrs: targetAddrs,
		BlsPubKeys:  blsPubKeys,
		CoinsFrozen: config.CoinsFrozen,
		FromAddr:    a.fromaddr,
		Height:      a.height,
	}

	r := makeSupervisionNodeGroupIDReceipt(a.fromaddr, nil, stat)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)
	return receipt, nil
}

func (a *action) supervisionNodeGroupApprove(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	//只在主链检查
	if !cfg.IsPara() && !isSuperManager(cfg, a.fromaddr) {
		return nil, errors.Wrapf(types.ErrNotAllow, "node group approve not supervision manager:%s", a.fromaddr)
	}

	id, err := getSupervisionNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
	if err != nil {
		return nil, err
	}

	if config.Title != id.Title {
		return nil, errors.Wrapf(pt.ErrNodeNotForTheTitle, "config title:%s,id title:%s", config.Title, id.Title)
	}

	return a.supervisionNodeGroupApproveApply(config, id)
}

func (a *action) supervisionNodeGroupApproveApply(config *pt.ParaNodeAddrConfig, apply *pt.ParaNodeGroupStatus) (*types.Receipt, error) {
	err := a.checkSupervisionNodeGroupExist(config.Title)
	if err != nil {
		return nil, err
	}

	if apply.CoinsFrozen < config.CoinsFrozen {
		return nil, errors.Wrapf(pt.ErrParaNodeGroupFrozenCoinsNotEnough, "id not enough coins apply:%d,config:%d", apply.CoinsFrozen, config.CoinsFrozen)
	}

	receipt := &types.Receipt{Ty: types.ExecOk}
	//create the node group
	r, err := a.supervisionNodeGroupCreate(apply)
	if err != nil {
		return nil, errors.Wrapf(err, "nodegroup create:title:%s,addrs:%s", config.Title, apply.TargetAddrs)
	}
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	copyStat := *apply
	apply.Status = pt.ParacrossSupervisionNodeApprove
	apply.Height = a.height

	r = makeSupervisionNodeGroupIDReceipt(a.fromaddr, &copyStat, apply)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	r = makeParaSupervisionNodeGroupStatusReceipt(config.Title, a.fromaddr, nil, apply)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)
	cfg := a.api.GetConfig()
	if cfg.IsPara() && cfg.IsDappFork(a.height, pt.ParaX, pt.ForkParaSelfConsStages) {
		//不允许主链成功平行链失败导致不一致的情况，这里如果失败则手工设置init stage ???
		r = supervisionSelfConsentInitStage(cfg)
		receipt.KV = append(receipt.KV, r.KV...)
		receipt.Logs = append(receipt.Logs, r.Logs...)
	}

	return receipt, nil
}

func (a *action) supervisionNodeGroupQuit(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	status, err := getSupervisionNodeGroupID(cfg, a.db, config.Title, a.exec.GetMainHeight(), config.Id)
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
	if status.Status != pt.ParacrossSupervisionNodeApply {
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
	status.Status = pt.ParacrossSupervisionNodeQuit
	status.Height = a.height

	r := makeSupervisionNodeGroupIDReceipt(a.fromaddr, &copyStat, status)
	receipt.KV = append(receipt.KV, r.KV...)
	receipt.Logs = append(receipt.Logs, r.Logs...)

	return receipt, nil
}

func (a *action) SupervisionNodeGroupConfig(config *pt.ParaNodeAddrConfig) (*types.Receipt, error) {
	cfg := a.api.GetConfig()
	if !validTitle(cfg, config.Title) {
		return nil, pt.ErrInvalidTitle
	}
	if !types.IsParaExecName(string(a.tx.Execer)) && cfg.IsDappFork(a.exec.GetMainHeight(), pt.ParaX, pt.ForkParaSupervisionRbk) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}
	if (config.Op == pt.ParacrossSupervisionNodeApprove || config.Op == pt.ParacrossSupervisionNodeQuit) && config.Id == "" {
		return nil, types.ErrInvalidParam
	}

	if config.Op == pt.ParacrossSupervisionNodeApply {
		return a.supervisionNodeGroupApply(config)
	} else if config.Op == pt.ParacrossSupervisionNodeApprove {
		return a.supervisionNodeGroupApprove(config)
	} else if config.Op == pt.ParacrossSupervisionNodeQuit {
		return a.supervisionNodeGroupQuit(config)
	}

	return nil, pt.ErrParaUnSupportNodeOper
}

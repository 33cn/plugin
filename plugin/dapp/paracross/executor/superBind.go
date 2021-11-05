package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func getBindLog(parabind, old *pt.ParaSuperNodeBindMiner) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	if old == nil {
		log.Ty = pt.TyLogParaSuperNodeBindReturnNew
		r := &pt.ReceiptParaSuperNodeBindReturnNew{}
		r.SuperAddress = parabind.SuperAddress
		r.NewMinerAddress = parabind.MinerAddress
		r.NewMinerBlsPubKey = parabind.MinerBlsPubKey
		log.Log = types.Encode(r)
	} else if len(parabind.MinerAddress) <= 0 {
		log.Ty = pt.TyLogParaSuperNodeBindReturnUnBind
		r := &pt.ReceiptParaSuperNodeBindReturnUnBind{}
		r.SuperAddress = parabind.SuperAddress
		r.OldMinerAddress = old.MinerAddress
		r.OldMinerBlsPubKey = old.MinerBlsPubKey
		log.Log = types.Encode(r)
	} else {
		log.Ty = pt.TyLogParaSuperNodeBindReturnUpdate
		r := &pt.ReceiptParaSuperNodeBindReturnUpdate{}
		r.SuperAddress = parabind.SuperAddress
		r.NewMinerAddress = parabind.MinerAddress
		r.NewMinerBlsPubKey = parabind.MinerBlsPubKey
		r.OldMinerAddress = old.MinerAddress
		r.OldMinerBlsPubKey = old.MinerBlsPubKey
		log.Log = types.Encode(r)
	}

	return log
}

func getMinerLog(parabind, oldbind *pt.ParaSuperNodeBindMiner) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	if oldbind != nil && oldbind.MinerBlsPubKey != parabind.MinerBlsPubKey {
		log.Ty = pt.TyLogParaSuperNodeBindMinerUpdate
		r := &pt.ReceiptParaSuperNodeBindMinerUpdate{}
		r.MinerAddress = parabind.MinerAddress
		r.OldMinerBlsPubKey = oldbind.MinerBlsPubKey
		r.NewMinerBlsPubKey = parabind.MinerBlsPubKey
		r.SuperAddress = parabind.SuperAddress
		log.Log = types.Encode(r)
	} else {
		log.Ty = pt.TyLogParaSuperNodeBindMinerNew
		r := &pt.ReceiptParaSuperNodeBindMinerNew{}
		r.MinerAddress = parabind.MinerAddress
		r.MinerBlsPubKey = parabind.MinerBlsPubKey
		r.NewSuperAddress = parabind.SuperAddress
		log.Log = types.Encode(r)
	}
	return log
}

func getMinerLogRemove(miner, oldSuper string) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogParaSuperNodeBindMinerUnBind
	r := &pt.ReceiptParaSuperNodeBindMinerUnBind{}
	r.MinerAddress = miner
	r.OldSuperAddress = oldSuper
	log.Log = types.Encode(r)
	return log
}

func getBind(db dbm.KV, addr string) (*pt.ParaSuperNodeBindMiner, error) {
	value, err := db.Get(calcParaSuperNodeBindReturnAddr(addr))
	if err != nil || value == nil || string(value) == string(types.EmptyValue) {
		return nil, types.ErrNotFound
	}
	var bind pt.ParaSuperNodeBindMiner
	err = types.Decode(value, &bind)
	if err != nil {
		return nil, err
	}
	return &bind, nil
}

func getSuper(db dbm.KV, addr string) (*pt.ParaSuperNodeBindMiner, error) {
	value, err := db.Get(calcParaSuperNodeBindMinerAddr(addr))
	if err != nil || value == nil || string(value) == string(types.EmptyValue) {
		return nil, types.ErrNotFound
	}
	var bind pt.ParaSuperNodeBindMiner
	err = types.Decode(value, &bind)
	if err != nil {
		return nil, err
	}
	return &bind, nil
}

func getBindSuperNode(db dbm.KV, addr, title string) (string, error) {
	bindMiner, err := getSuper(db, addr)
	if err != nil && err != types.ErrNotFound {
		return "", err
	}
	if bindMiner != nil {
		nodes, _, err := getParacrossNodes(db, title)
		if err != nil {
			return "", errors.Wrapf(err, "getNodes for title:%s", title)
		}
		if !validNode(bindMiner.SuperAddress, nodes) {
			return "", errors.Wrapf(pt.ErrParaNodeAddrNotExisted, "invalid node=%s", bindMiner.SuperAddress)
		}

		return bindMiner.SuperAddress, nil
	}
	return "", nil
}

func (a *action) superNodeBindMiner(parabind *pt.ParaSuperNodeBindMiner) (*types.Receipt, error) {
	// 只允许平行链操作
	if !types.IsParaExecName(string(a.tx.Execer)) {
		clog.Error("superNodeBindMiner", "string(a.tx.Execer)", string(a.tx.Execer))
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}

	// "" 表示设置为空,解除绑定, 非空即绑定
	if len(parabind.MinerAddress) > 0 {
		if err := address.CheckAddress(parabind.MinerAddress); err != nil {
			return nil, err
		}

		if len(parabind.MinerBlsPubKey) <= 0 {
			return nil, errors.Wrapf(types.ErrEmpty, "minerAddress=%s, minerPriKey is empty", parabind.MinerAddress)
		}
	}

	if a.fromaddr != parabind.SuperAddress {
		clog.Error("superNodeBindMiner", "a.fromaddr != parabind.SuperAddress", a.fromaddr)
		return nil, types.ErrFromAddr
	}

	oldbind, err := getBind(a.db, parabind.SuperAddress)
	if err != nil && err != types.ErrNotFound {
		clog.Error("superNodeBindMiner", "getBind err", err)
		return nil, err
	}
	oldbindMiner := ""
	oldbindBls := ""
	if oldbind != nil {
		oldbindMiner = oldbind.MinerAddress
		oldbindBls = oldbind.MinerBlsPubKey
		clog.Debug("superNodeBindMiner", "oldbind.MinerAddress", oldbind.MinerAddress)
	}
	if oldbindMiner == parabind.MinerAddress && oldbindBls == parabind.MinerBlsPubKey {
		if len(parabind.MinerAddress) <= 0 {
			// 之前未绑定 不需要解除绑定
			return nil, errors.Wrapf(types.ErrInvalidParam, "It has not been bound before and does not need to be unbound")
		} else {
			// 这两次绑定的地址都一样 不做处理
			return nil, errors.Wrapf(types.ErrSendSameToRecv, "minerAddress=%s is same", parabind.MinerAddress)
		}
	}

	// 获取授权节点情况
	minerInfo, err := getSuper(a.db, parabind.MinerAddress)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if minerInfo != nil && minerInfo.SuperAddress != parabind.SuperAddress {
		// 委托地址已经绑定了其他超级节点,不能再被绑定,要先解除绑定
		return nil, errors.Wrapf(types.ErrInvalidParam, "minerAddress=%s is used", parabind.MinerAddress)
	}

	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	log := getBindLog(parabind, oldbind)
	logs = append(logs, log)

	value := types.Encode(parabind)
	if len(parabind.MinerAddress) <= 0 {
		kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindReturnAddr(parabind.SuperAddress), Value: types.EmptyValue})
	} else {
		log2 := getMinerLog(parabind, oldbind)
		logs = append(logs, log2)

		kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindReturnAddr(parabind.SuperAddress), Value: value})
		kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(parabind.MinerAddress), Value: value})
	}

	if oldbindMiner != "" {
		// 获取共识节点之前绑定的授权节点情况
		oldSuper, err := getSuper(a.db, oldbind.MinerAddress)
		if err != nil && err != types.ErrNotFound {
			return nil, err
		}

		// 更新之前绑定授权节点信息
		if oldSuper != nil && oldSuper.MinerAddress != "" {
			clog.Debug("superNodeBindMiner oldSuper", "oldSuper.SuperAddress", oldSuper.SuperAddress, "oldSuper.MinerAddress", oldSuper.MinerAddress)
			log3 := getMinerLogRemove(oldSuper.MinerAddress, oldSuper.SuperAddress)
			logs = append(logs, log3)

			kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(oldSuper.MinerAddress), Value: types.EmptyValue})
		}
	}

	for i := 0; i < len(kvs); i++ {
		_ = a.db.Set(kvs[i].GetKey(), kvs[i].Value)
	}

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

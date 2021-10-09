package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func getBindLog(parabind *pt.ParaSuperNodeBindMiner, old string) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogParaSuperNodeBindReturn
	r := &pt.ReceiptParaSuperNodeBindReturn{}
	r.SuperAddress = parabind.SuperAddress
	r.OldMinerAddress = old
	r.NewMinerAddress = parabind.MinerAddress
	log.Log = types.Encode(r)
	return log
}

func getMinerLog(parabind *pt.ParaSuperNodeBindMiner, old string) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogParaSuperNodeBindMiner
	r := &pt.ReceiptParaSuperNodeBindMiner{}
	r.MinerAddress = parabind.MinerAddress
	r.MinerBlsPubKey = parabind.MinerBlsPubKey
	r.OldSuperAddress = old
	r.NewSuperAddress = parabind.SuperAddress
	log.Log = types.Encode(r)
	return log
}

func getBind(db dbm.KV, addr string) (*pt.ParaSuperNodeBindMiner, error) {
	value, err := db.Get(calcParaSuperNodeBindReturnAddr(addr))
	if err != nil || value == nil {
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
	if err != nil || value == nil {
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
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
	}

	// "" 表示设置为空,解除绑定
	if len(parabind.MinerAddress) > 0 {
		if err := address.CheckAddress(parabind.MinerAddress); err != nil {
			return nil, err
		}

		if len(parabind.MinerBlsPubKey) <= 0 {
			return nil, errors.Wrapf(types.ErrEmpty, "minerAddress=%s, minerPriKey is empty", parabind.MinerAddress)
		}
	}

	if a.fromaddr != parabind.SuperAddress {
		return nil, types.ErrFromAddr
	}
	// 发起者必须是共识节点
	err := a.isValidSuperNode(a.fromaddr)
	if err != nil {
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	oldbind, err := getBind(a.db, parabind.SuperAddress)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	oldbindMiner := ""
	oldbindBls := ""
	if oldbind != nil {
		oldbindMiner = oldbind.MinerAddress
		oldbindBls = oldbind.MinerBlsPubKey
	}
	if oldbindMiner == parabind.MinerAddress && oldbindBls == parabind.MinerBlsPubKey {
		// 这两次绑定的地址都一样 不做处理
		return nil, errors.Wrapf(types.ErrSendSameToRecv, "minerAddress=%s is same", parabind.MinerAddress)
	}

	log := getBindLog(parabind, oldbindMiner)
	logs = append(logs, log)
	oldSuper, err := getSuper(a.db, parabind.MinerAddress)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if oldSuper != nil {
		log2 := getMinerLog(parabind, oldSuper.SuperAddress)
		logs = append(logs, log2)
	}

	value := types.Encode(parabind)
	kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindReturnAddr(parabind.SuperAddress), Value: value})
	kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(parabind.MinerAddress), Value: value})
	if len(parabind.MinerAddress) <= 0 {
		kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(oldbindMiner), Value: nil})
	}

	for i := 0; i < len(kvs); i++ {
		_ = a.db.Set(kvs[i].GetKey(), kvs[i].Value)
	}

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

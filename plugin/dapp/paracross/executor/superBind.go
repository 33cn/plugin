package executor

import (
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
)

func saveBind(db dbm.KV, parabind *pt.ParaSuperNodeBindMiner) {
	set := getBindKV(parabind)
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

func getBindKV(parabind *pt.ParaSuperNodeBindMiner) (kvset []*types.KeyValue) {
	value := types.Encode(parabind)
	kvset = append(kvset, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(parabind.SuperAddress), Value: value})
	return kvset
}

func getBindLog(parabind *pt.ParaSuperNodeBindMiner, old string) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogParaSuperNodeBindMiner
	r := &pt.ReceiptParaSuperNodeBindMiner{}
	r.SuperAddress = parabind.SuperAddress
	r.OldMinerAddress = old
	r.NewMinerAddress = parabind.MinerAddress
	log.Log = types.Encode(r)
	return log
}

func (a *action) getBind(addr string) string {
	value, err := a.db.Get(calcParaSuperNodeBindMinerAddr(addr))
	if err != nil || value == nil {
		return ""
	}
	var bind pt.ParaSuperNodeBindMiner
	err = types.Decode(value, &bind)
	if err != nil {
		panic(err)
	}
	return bind.MinerAddress
}

func (a *action) superNodeBindMiner(parabind *pt.ParaSuperNodeBindMiner) (*types.Receipt, error) {
	// "" 表示设置为空,解除绑定
	if len(parabind.MinerAddress) > 0 {
		if err := address.CheckAddress(parabind.MinerAddress); err != nil {
			return nil, err
		}
	}

	// 只允许平行链操作
	if !types.IsParaExecName(string(a.tx.Execer)) {
		return nil, errors.Wrapf(types.ErrInvalidParam, "exec=%s,should prefix with user.p.", string(a.tx.Execer))
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
	oldbind := a.getBind(parabind.SuperAddress)
	log := getBindLog(parabind, oldbind)
	logs = append(logs, log)
	saveBind(a.db, parabind)
	kv := getBindKV(parabind)
	kvs = append(kvs, kv...)
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

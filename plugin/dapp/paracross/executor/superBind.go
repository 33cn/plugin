package executor

import (
	"github.com/33cn/chain33/common/address"
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

func (a *action) getBind(addr string) (string, string) {
	value, err := a.db.Get(calcParaSuperNodeBindReturnAddr(addr))
	if err != nil || value == nil {
		return "", ""
	}
	var bind pt.ParaSuperNodeBindMiner
	err = types.Decode(value, &bind)
	if err != nil {
		panic(err)
	}
	return bind.MinerAddress, bind.MinerBlsPubKey
}

func (a *action) getSuper(addr string) string {
	value, err := a.db.Get(calcParaSuperNodeBindMinerAddr(addr))
	if err != nil || value == nil {
		return ""
	}
	var bind pt.ParaSuperNodeBindMiner
	err = types.Decode(value, &bind)
	if err != nil {
		panic(err)
	}
	return bind.SuperAddress
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
	oldbind, oldBls := a.getBind(parabind.SuperAddress)
	if oldbind == parabind.MinerAddress && oldBls == parabind.MinerBlsPubKey {
		// 这两次绑定的地址都一样 不做处理
		return nil, errors.Wrapf(types.ErrSendSameToRecv, "minerAddress=%s is same", parabind.MinerAddress)
	}

	log := getBindLog(parabind, oldbind)
	logs = append(logs, log)
	oldSuper := a.getSuper(parabind.MinerAddress)
	log2 := getMinerLog(parabind, oldSuper)
	logs = append(logs, log2)

	value := types.Encode(parabind)
	kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindReturnAddr(parabind.SuperAddress), Value: value})
	kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(parabind.MinerAddress), Value: value})
	if len(parabind.MinerAddress) <= 0 {
		kvs = append(kvs, &types.KeyValue{Key: calcParaSuperNodeBindMinerAddr(oldbind), Value: nil})
	}

	for i := 0; i < len(kvs); i++ {
		_ = a.db.Set(kvs[i].GetKey(), kvs[i].Value)
	}

	receipt := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipt, nil
}

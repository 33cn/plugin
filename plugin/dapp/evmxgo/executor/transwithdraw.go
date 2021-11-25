// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

func (e *evmxgo) ExecTransWithdraw(accountDB *account.DB, tx *types.Transaction, action *evmxgotypes.EvmxgoAction, index int) (*types.Receipt, error) {
	switch {
	case action.Ty == evmxgotypes.ActionTransfer && action.GetTransfer() != nil:
		transfer := action.GetTransfer()
		from := tx.From()
		//to 是 execs 合约地址
		if drivers.IsDriverAddress(transfer.To, e.GetHeight()) {
			return accountDB.TransferToExec(from, transfer.To, transfer.Amount)
		}
		return accountDB.Transfer(from, transfer.To, transfer.Amount)

	case action.Ty == evmxgotypes.ActionWithdraw && action.GetWithdraw() != nil:
		withdraw := action.GetWithdraw()
		from := tx.From()
		//to 是 execs 合约地址
		if drivers.IsDriverAddress(withdraw.To, e.GetHeight()) || isExecAddrMatch(withdraw.ExecName, withdraw.To) {
			return accountDB.TransferWithdraw(from, withdraw.To, withdraw.Amount)
		}
		return nil, types.ErrActionNotSupport

	case action.Ty == evmxgotypes.EvmxgoActionTransferToExec && action.GetTransferToExec() != nil:
		transfer := action.GetTransferToExec()
		from := tx.From()
		//to 是 execs 合约地址
		if !isExecAddrMatch(transfer.ExecName, transfer.To) {
			return nil, types.ErrToAddrNotSameToExecAddr
		}
		return accountDB.TransferToExec(from, transfer.To, transfer.Amount)

	default:
		return nil, types.ErrActionNotSupport
	}

}

func isExecAddrMatch(name string, to string) bool {
	toaddr := address.ExecAddress(name)
	return toaddr == to
}

//0: all tx
//1: from tx
//2: to tx

func (e *evmxgo) ExecLocalTransWithdraw(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	//执行成功
	var action evmxgotypes.EvmxgoAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		panic(err)
	}
	var kv *types.KeyValue
	if action.Ty == evmxgotypes.ActionTransfer && action.GetTransfer() != nil {
		transfer := action.GetTransfer()
		kv, err = updateAddrReciver(e.GetLocalDB(), transfer.Cointoken, transfer.To, transfer.Amount, true)
	} else if action.Ty == evmxgotypes.ActionWithdraw && action.GetWithdraw() != nil {
		withdraw := action.GetWithdraw()
		from := tx.From()
		kv, err = updateAddrReciver(e.GetLocalDB(), withdraw.Cointoken, from, withdraw.Amount, true)
	} else if action.Ty == evmxgotypes.EvmxgoActionTransferToExec && action.GetTransferToExec() != nil {
		transfer := action.GetTransferToExec()
		kv, err = updateAddrReciver(e.GetLocalDB(), transfer.Cointoken, transfer.To, transfer.Amount, true)
	}
	if err != nil {
		return set, nil
	}
	if kv != nil {
		set.KV = append(set.KV, kv)
	}
	return set, nil
}

func (e *evmxgo) ExecDelLocalLocalTransWithdraw(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	//执行成功
	var action evmxgotypes.EvmxgoAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		panic(err)
	}
	var kv *types.KeyValue
	if action.Ty == evmxgotypes.ActionTransfer && action.GetTransfer() != nil {
		transfer := action.GetTransfer()
		kv, err = updateAddrReciver(e.GetLocalDB(), transfer.Cointoken, transfer.To, transfer.Amount, false)
	} else if action.Ty == evmxgotypes.ActionWithdraw && action.GetWithdraw() != nil {
		withdraw := action.GetWithdraw()
		from := tx.From()
		kv, err = updateAddrReciver(e.GetLocalDB(), withdraw.Cointoken, from, withdraw.Amount, false)
	} else if action.Ty == evmxgotypes.EvmxgoActionTransferToExec && action.GetTransferToExec() != nil {
		transfer := action.GetTransferToExec()
		kv, err = updateAddrReciver(e.GetLocalDB(), transfer.Cointoken, transfer.To, transfer.Amount, false)
	}
	if err != nil {
		return set, nil
	}
	if kv != nil {
		set.KV = append(set.KV, kv)
	}
	return set, nil
}

func getAddrReciverKV(token string, addr string, reciverAmount int64) *types.KeyValue {
	reciver := &types.Int64{Data: reciverAmount}
	amountbytes := types.Encode(reciver)
	kv := &types.KeyValue{Key: calcAddrKey(token, addr), Value: amountbytes}
	return kv
}

func getAddrReciver(db dbm.KVDB, token string, addr string) (int64, error) {
	reciver := types.Int64{}
	addrReciver, err := db.Get(calcAddrKey(token, addr))
	if err != nil && err != types.ErrNotFound {
		return 0, err
	}
	if len(addrReciver) == 0 {
		return 0, nil
	}
	err = types.Decode(addrReciver, &reciver)
	if err != nil {
		return 0, err
	}
	return reciver.Data, nil
}

func setAddrReciver(db dbm.KVDB, token string, addr string, reciverAmount int64) error {
	kv := getAddrReciverKV(addr, token, reciverAmount)
	return db.Set(kv.Key, kv.Value)
}

func updateAddrReciver(cachedb dbm.KVDB, token string, addr string, amount int64, isadd bool) (*types.KeyValue, error) {
	recv, err := getAddrReciver(cachedb, token, addr)
	if err != nil && err != types.ErrNotFound {
		return nil, err
	}
	if isadd {
		recv += amount
	} else {
		recv -= amount
	}
	err = setAddrReciver(cachedb, token, addr, recv)
	if err != nil {
		return nil, err
	}
	//keyvalue
	return getAddrReciverKV(token, addr, recv), nil
}

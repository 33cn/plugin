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
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
)

func (t *token) ExecTransWithdraw(accountDB *account.DB, tx *types.Transaction, action *tokenty.TokenAction, index int) (*types.Receipt, error) {
	if (action.Ty == tokenty.ActionTransfer) && action.GetTransfer() != nil {
		transfer := action.GetTransfer()
		from := tx.From()
		//to 是 execs 合约地址
		if drivers.IsDriverAddress(tx.GetRealToAddr(), t.GetHeight()) {
			return accountDB.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
		}
		return accountDB.Transfer(from, tx.GetRealToAddr(), transfer.Amount)
	} else if (action.Ty == tokenty.ActionWithdraw) && action.GetWithdraw() != nil {
		withdraw := action.GetWithdraw()
		if !types.IsFork(t.GetHeight(), "ForkWithdraw") {
			withdraw.ExecName = ""
		}
		from := tx.From()
		//to 是 execs 合约地址
		if drivers.IsDriverAddress(tx.GetRealToAddr(), t.GetHeight()) || isExecAddrMatch(withdraw.ExecName, tx.GetRealToAddr()) {
			return accountDB.TransferWithdraw(from, tx.GetRealToAddr(), withdraw.Amount)
		}
		return nil, types.ErrActionNotSupport
	} else if (action.Ty == tokenty.ActionGenesis) && action.GetGenesis() != nil {
		genesis := action.GetGenesis()
		if t.GetHeight() == 0 {
			if drivers.IsDriverAddress(tx.GetRealToAddr(), t.GetHeight()) {
				return accountDB.GenesisInitExec(genesis.ReturnAddress, genesis.Amount, tx.GetRealToAddr())
			}
			return accountDB.GenesisInit(tx.GetRealToAddr(), genesis.Amount)
		}
		return nil, types.ErrReRunGenesis
	} else if action.Ty == tokenty.TokenActionTransferToExec && action.GetTransferToExec() != nil {
		if !types.IsFork(t.GetHeight(), "ForkTransferExec") {
			return nil, types.ErrActionNotSupport
		}
		transfer := action.GetTransferToExec()
		from := tx.From()
		//to 是 execs 合约地址
		if !isExecAddrMatch(transfer.ExecName, tx.GetRealToAddr()) {
			return nil, types.ErrToAddrNotSameToExecAddr
		}
		return accountDB.TransferToExec(from, tx.GetRealToAddr(), transfer.Amount)
	} else {
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

func (t *token) ExecLocalTransWithdraw(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	//执行成功
	var action tokenty.TokenAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		panic(err)
	}
	var kv *types.KeyValue
	if action.Ty == tokenty.ActionTransfer && action.GetTransfer() != nil {
		transfer := action.GetTransfer()
		kv, err = updateAddrReciver(t.GetLocalDB(), transfer.Cointoken, tx.GetRealToAddr(), transfer.Amount, true)
	} else if action.Ty == tokenty.ActionWithdraw && action.GetWithdraw() != nil {
		withdraw := action.GetWithdraw()
		from := tx.From()
		kv, err = updateAddrReciver(t.GetLocalDB(), withdraw.Cointoken, from, withdraw.Amount, true)
	} else if action.Ty == tokenty.ActionGenesis && action.GetGenesis() != nil {
		gen := action.GetGenesis()
		kv, err = updateAddrReciver(t.GetLocalDB(), "token", tx.GetRealToAddr(), gen.Amount, true)
	} else if action.Ty == tokenty.TokenActionTransferToExec && action.GetTransferToExec() != nil {
		transfer := action.GetTransferToExec()
		kv, err = updateAddrReciver(t.GetLocalDB(), transfer.Cointoken, tx.GetRealToAddr(), transfer.Amount, true)
	}
	if err != nil {
		return set, nil
	}
	if kv != nil {
		set.KV = append(set.KV, kv)
	}
	return set, nil
}

func (t *token) ExecDelLocalLocalTransWithdraw(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if receipt.GetTy() != types.ExecOk {
		return set, nil
	}
	//执行成功
	var action tokenty.TokenAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		panic(err)
	}
	var kv *types.KeyValue
	if action.Ty == tokenty.ActionTransfer && action.GetTransfer() != nil {
		transfer := action.GetTransfer()
		kv, err = updateAddrReciver(t.GetLocalDB(), transfer.Cointoken, tx.GetRealToAddr(), transfer.Amount, false)
	} else if action.Ty == tokenty.ActionWithdraw && action.GetWithdraw() != nil {
		withdraw := action.GetWithdraw()
		from := tx.From()
		kv, err = updateAddrReciver(t.GetLocalDB(), withdraw.Cointoken, from, withdraw.Amount, false)
	} else if action.Ty == tokenty.TokenActionTransferToExec && action.GetTransferToExec() != nil {
		transfer := action.GetTransferToExec()
		kv, err = updateAddrReciver(t.GetLocalDB(), transfer.Cointoken, tx.GetRealToAddr(), transfer.Amount, false)
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

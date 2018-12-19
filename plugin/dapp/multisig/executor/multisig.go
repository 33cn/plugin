// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

/*
multiSig合约主要实现如下功能：
//创建多重签名账户
//多重签名账户owner属性的修改：owner的add/del/replace等
//多重签名账户属性的修改：weight权重以及每日限额的修改
//多重签名账户交易的确认和撤销
//合约中外部账户转账到多重签名账户，Addr --->multiSigAddr
//合约中多重签名账户转账到外部账户，multiSigAddr--->Addr
*/

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/system/dapp"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

var multisiglog = log.New("module", "execs.multisig")

var driverName = "multisig"

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&MultiSig{}))
}

// Init multisig模块初始化
func Init(name string, sub []byte) {
	drivers.Register(GetName(), newMultiSig, types.GetDappFork(driverName, "Enable"))
}

// GetName multisig合约name
func GetName() string {
	return newMultiSig().GetName()
}

// MultiSig multisig合约实例
type MultiSig struct {
	drivers.DriverBase
}

func newMultiSig() drivers.Driver {
	m := &MultiSig{}
	m.SetChild(m)
	m.SetExecutorType(types.LoadExecutorType(driverName))
	return m
}

// GetDriverName 获取multisig合约name
func (m *MultiSig) GetDriverName() string {
	return driverName
}

// CheckTx 检测multisig合约交易,转账交易amount不能为负数
func (m *MultiSig) CheckTx(tx *types.Transaction, index int) error {
	ety := m.GetExecutorType()

	//amount check
	amount, err := ety.Amount(tx)
	if err != nil {
		return err
	}
	if amount < 0 {
		return types.ErrAmount
	}

	_, v, err := ety.DecodePayloadValue(tx)
	if err != nil {
		return err
	}
	payload := v.Interface()

	//MultiSigAccCreate 交易校验
	if ato, ok := payload.(*mty.MultiSigAccCreate); ok {
		return checkAccountCreateTx(ato)
	}

	//MultiSigOwnerOperate 交易的检测
	if ato, ok := payload.(*mty.MultiSigOwnerOperate); ok {
		return checkOwnerOperateTx(ato)
	}
	//MultiSigAccOperate 交易的检测
	if ato, ok := payload.(*mty.MultiSigAccOperate); ok {
		return checkAccountOperateTx(ato)
	}
	//MultiSigConfirmTx  交易的检测
	if ato, ok := payload.(*mty.MultiSigConfirmTx); ok {
		if err := address.CheckMultiSignAddress(ato.GetMultiSigAccAddr()); err != nil {
			return types.ErrInvalidAddress
		}
		return nil
	}

	//MultiSigExecTransferTo 交易的检测
	if ato, ok := payload.(*mty.MultiSigExecTransferTo); ok {
		if err := address.CheckMultiSignAddress(ato.GetTo()); err != nil {
			return types.ErrInvalidAddress
		}
		//assets check
		return mty.IsAssetsInvalid(ato.GetExecname(), ato.GetSymbol())
	}
	//MultiSigExecTransferFrom 交易的检测
	if ato, ok := payload.(*mty.MultiSigExecTransferFrom); ok {
		//from addr check
		if err := address.CheckMultiSignAddress(ato.GetFrom()); err != nil {
			return types.ErrInvalidAddress
		}
		//to addr check
		if err := address.CheckAddress(ato.GetTo()); err != nil {
			return types.ErrInvalidAddress
		}
		//assets check
		return mty.IsAssetsInvalid(ato.GetExecname(), ato.GetSymbol())
	}

	return nil
}
func checkAccountCreateTx(ato *mty.MultiSigAccCreate) error {
	var totalweight uint64
	var ownerCount int

	requiredWeight := ato.GetRequiredWeight()
	if requiredWeight == 0 {
		return mty.ErrInvalidWeight
	}
	owners := ato.GetOwners()
	ownersMap := make(map[string]bool)

	//创建时requiredweight权重的值不能大于所有owner权重之和
	for _, owner := range owners {
		if owner != nil {
			if err := address.CheckAddress(owner.OwnerAddr); err != nil {
				return types.ErrInvalidAddress
			}
			if owner.Weight == 0 {
				return mty.ErrInvalidWeight
			}
			if ownersMap[owner.OwnerAddr] {
				return mty.ErrOwnerExist
			}
			ownersMap[owner.OwnerAddr] = true
			totalweight += owner.Weight
			ownerCount = ownerCount + 1
		}
	}

	if ato.RequiredWeight > totalweight {
		return mty.ErrRequiredweight
	}

	//创建时最少设置两个owner
	if ownerCount < mty.MinOwnersInit {
		return mty.ErrOwnerLessThanTwo
	}
	//owner总数不能大于最大值
	if ownerCount > mty.MaxOwnersCount {
		return mty.ErrMaxOwnerCount
	}

	dailyLimit := ato.GetDailyLimit()
	//assets check
	return mty.IsAssetsInvalid(dailyLimit.GetExecer(), dailyLimit.GetSymbol())
}

func checkOwnerOperateTx(ato *mty.MultiSigOwnerOperate) error {
	OldOwner := ato.GetOldOwner()
	NewOwner := ato.GetNewOwner()
	NewWeight := ato.GetNewWeight()
	MultiSigAccAddr := ato.GetMultiSigAccAddr()
	if err := address.CheckMultiSignAddress(MultiSigAccAddr); err != nil {
		return types.ErrInvalidAddress
	}

	if ato.OperateFlag == mty.OwnerAdd {
		if err := address.CheckAddress(NewOwner); err != nil {
			return types.ErrInvalidAddress
		}
		if NewWeight <= 0 {
			return mty.ErrInvalidWeight
		}
	}
	if ato.OperateFlag == mty.OwnerDel {
		if err := address.CheckAddress(OldOwner); err != nil {
			return types.ErrInvalidAddress
		}
	}
	if ato.OperateFlag == mty.OwnerModify {
		if err := address.CheckAddress(OldOwner); err != nil {
			return types.ErrInvalidAddress
		}
		if NewWeight <= 0 {
			return mty.ErrInvalidWeight
		}
	}
	if ato.OperateFlag == mty.OwnerReplace {
		if err := address.CheckAddress(OldOwner); err != nil {
			return types.ErrInvalidAddress
		}
		if err := address.CheckAddress(NewOwner); err != nil {
			return types.ErrInvalidAddress
		}
	}
	return nil
}
func checkAccountOperateTx(ato *mty.MultiSigAccOperate) error {
	//MultiSigAccOperate MultiSigAccAddr 地址检测
	MultiSigAccAddr := ato.GetMultiSigAccAddr()
	if err := address.CheckMultiSignAddress(MultiSigAccAddr); err != nil {
		return types.ErrInvalidAddress
	}

	if ato.OperateFlag == mty.AccWeightOp {
		NewWeight := ato.GetNewRequiredWeight()
		if NewWeight <= 0 {
			return mty.ErrInvalidWeight
		}
	}
	if ato.OperateFlag == mty.AccDailyLimitOp {
		dailyLimit := ato.GetDailyLimit()
		//assets check
		return mty.IsAssetsInvalid(dailyLimit.GetExecer(), dailyLimit.GetSymbol())
	}
	return nil
}

//多重签名交易的Receipt处理
func (m *MultiSig) execLocalMultiSigReceipt(receiptData *types.ReceiptData, tx *types.Transaction, addOrRollback bool) ([]*types.KeyValue, error) {
	var set []*types.KeyValue
	for _, log := range receiptData.Logs {
		multisiglog.Info("execLocalMultiSigReceipt", "Ty", log.Ty)

		switch log.Ty {
		case mty.TyLogMultiSigAccCreate:
			{
				var receipt mty.MultiSig
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}

				kv, err := m.saveMultiSigAccCreate(&receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigOwnerAdd,
			mty.TyLogMultiSigOwnerDel:
			{
				var receipt mty.ReceiptOwnerAddOrDel
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigOwnerAddOrDel(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigOwnerModify,
			mty.TyLogMultiSigOwnerReplace:
			{
				var receipt mty.ReceiptOwnerModOrRep
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigOwnerModOrRep(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigAccWeightModify:
			{
				var receipt mty.ReceiptWeightModify
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigAccWeight(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigAccDailyLimitAdd,
			mty.TyLogMultiSigAccDailyLimitModify:
			{
				var receipt mty.ReceiptDailyLimitOperate
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigAccDailyLimit(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigConfirmTx, //只是交易确认和撤销，交易没有被执行
			mty.TyLogMultiSigConfirmTxRevoke:
			{
				var receipt mty.ReceiptConfirmTx
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigConfirmTx(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogDailyLimitUpdate: //账户的DailyLimit更新
			{
				var receipt mty.ReceiptAccDailyLimitUpdate
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveDailyLimitUpdate(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		case mty.TyLogMultiSigTx: //交易被某个owner确认并执行
			{
				var receipt mty.ReceiptMultiSigTx
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				//交易被执行需要更新tx的执行状态以及确认owner列表
				kv1, err := m.saveMultiSigTx(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv1...)

				//转账交易被执行需要更新账户的amount统计计数,需要区分submit和confirm
				if receipt.CurExecuted {
					kv2, err := m.saveMultiSigTransfer(tx, receipt.SubmitOrConfirm, addOrRollback)
					if err != nil {
						return nil, err
					}
					set = append(set, kv2...)
				}
			}
		case mty.TyLogTxCountUpdate:
			{
				var receipt mty.ReceiptTxCountUpdate
				err := types.Decode(log.Log, &receipt)
				if err != nil {
					return nil, err
				}
				kv, err := m.saveMultiSigTxCountUpdate(receipt, addOrRollback)
				if err != nil {
					return nil, err
				}
				set = append(set, kv...)
			}
		default:
			break
		}
	}
	return set, nil
}

//转账交易to地址收币数量更新，Submit直接解析tx。Confirm需要解析对应txid的交易信息
func (m *MultiSig) saveMultiSigTransfer(tx *types.Transaction, SubmitOrConfirm, addOrRollback bool) ([]*types.KeyValue, error) {
	var set []*types.KeyValue
	//执行成功解析GetPayload信息
	var action mty.MultiSigAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		panic(err)
	}
	var to string
	var execname string
	var symbol string
	var amount int64

	//addr-->multiSigAccAddr
	//multiSigAccAddr-->addr
	if SubmitOrConfirm {
		if action.Ty == mty.ActionMultiSigExecTransferTo && action.GetMultiSigExecTransferTo() != nil {
			tx := action.GetMultiSigExecTransferTo()
			to = tx.To
			execname = tx.Execname
			symbol = tx.Symbol
			amount = tx.Amount
		} else if action.Ty == mty.ActionMultiSigExecTransferFrom && action.GetMultiSigExecTransferFrom() != nil {
			tx := action.GetMultiSigExecTransferFrom()
			to = tx.To
			execname = tx.Execname
			symbol = tx.Symbol
			amount = tx.Amount
		} else {
			return set, nil
		}
	} else {
		if action.Ty != mty.ActionMultiSigConfirmTx || action.GetMultiSigConfirmTx() == nil {
			return nil, mty.ErrActionTyNoMatch
		}
		//通过需要确认的txid从数据库中获取对应的multiSigTx信息，然后根据txhash查询具体的交易详情
		multiSigConfirmTx := action.GetMultiSigConfirmTx()
		multiSigTx, err := getMultiSigTx(m.GetLocalDB(), multiSigConfirmTx.MultiSigAccAddr, multiSigConfirmTx.TxId)
		if err != nil {
			return set, err
		}
		tx, err := getTxByHash(m.GetAPI(), multiSigTx.TxHash)
		if err != nil {
			return nil, err
		}
		payload, err := getMultiSigTxPayload(tx)
		if err != nil {
			return nil, err
		}
		if multiSigTx.TxType == mty.TransferOperate {
			tx := payload.GetMultiSigExecTransferFrom()
			to = tx.To
			execname = tx.Execname
			symbol = tx.Symbol
			amount = tx.Amount
		} else {
			return set, nil
		}
	}
	kv, err := updateAddrReciver(m.GetLocalDB(), to, execname, symbol, amount, addOrRollback)
	if err != nil {
		return set, err
	}
	if kv != nil {
		set = append(set, kv)
	}
	return set, nil
}

//localdb Receipt相关消息的处理。需要区分执行的是add/Rollback
func (m *MultiSig) saveMultiSigAccCreate(multiSig *mty.MultiSig, addOrRollback bool) ([]*types.KeyValue, error) {
	multiSigAddr := multiSig.MultiSigAddr
	//增加一个多重签名账户信息到localdb中，第一次增加时账户在local中应该是不存在的。如果存在就返回错误
	oldmultiSig, err := getMultiSigAccount(m.GetLocalDB(), multiSigAddr)
	if err != nil {
		return nil, err
	}
	if addOrRollback && oldmultiSig != nil { //创建的账户已经存在报错
		multisiglog.Error("saveMultiSigAccCreate:getMultiSigAccount", "addOrRollback", addOrRollback, "MultiSigAddr", multiSigAddr, "oldmultiSig", oldmultiSig, "err", err)
		return nil, mty.ErrAccountHasExist

	} else if !addOrRollback && oldmultiSig == nil { //回滚函数的账户不经存在报错
		multisiglog.Error("saveMultiSigAccCreate:getMultiSigAccount", "addOrRollback", addOrRollback, "MultiSigAddr", multiSigAddr, "err", err)
		return nil, types.ErrAccountNotExist
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, addOrRollback)
	if err != nil {
		return nil, err
	}
	accountkv := getMultiSigAccountKV(multiSig, addOrRollback)

	//获取当前的账户计数
	lastcount, err := getMultiSigAccCount(m.GetLocalDB())
	if err != nil {
		return nil, err
	}

	//更新账户列表,回滚时索引需要--
	if !addOrRollback && lastcount > 0 {
		lastcount = lastcount - 1
	}
	accCountListkv, err := updateMultiSigAccList(m.GetLocalDB(), multiSig.MultiSigAddr, lastcount, addOrRollback)
	if err != nil {
		return nil, err
	}

	//账户计数增加一个
	accCountkv, err := updateMultiSigAccCount(m.GetLocalDB(), addOrRollback)
	if err != nil {
		return nil, err
	}
	//更新create地址创建的多重签名账户
	accAddrkv := setMultiSigAddress(m.GetLocalDB(), multiSig.CreateAddr, multiSig.MultiSigAddr, addOrRollback)

	var kvs []*types.KeyValue
	kvs = append(kvs, accCountkv)
	kvs = append(kvs, accountkv)
	kvs = append(kvs, accCountListkv)
	kvs = append(kvs, accAddrkv)

	return kvs, nil
}

//账户owner的add/del操作.需要区分add/del 交易
func (m *MultiSig) saveMultiSigOwnerAddOrDel(ownerOp mty.ReceiptOwnerAddOrDel, addOrRollback bool) ([]*types.KeyValue, error) {
	//增加一个多重签名账户信息到localdb中，第一次增加时账户在local中应该是不存在的。如果存在就返回错误
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), ownerOp.MultiSigAddr)
	multisiglog.Error("saveMultiSigOwnerAddOrDel", "ownerOp", ownerOp)

	if err != nil || multiSig == nil {
		multisiglog.Error("saveMultiSigOwnerAddOrDel", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "err", err)
		return nil, err
	}
	multisiglog.Error("saveMultiSigOwnerAddOrDel", "wonerlen ", len(multiSig.Owners))

	_, index, _, _, find := getOwnerInfoByAddr(multiSig, ownerOp.Owner.OwnerAddr)
	if addOrRollback { //正常添加交易
		if ownerOp.AddOrDel && !find { //add owner
			multiSig.Owners = append(multiSig.Owners, ownerOp.Owner)
		} else if !ownerOp.AddOrDel && find { //dell owner
			multiSig.Owners = delOwner(multiSig.Owners, index)
			//multiSig.Owners = append(multiSig.Owners[0:index], multiSig.Owners[index+1:]...)
		} else {
			multisiglog.Error("saveMultiSigOwnerAddOrDel", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "index", index, "find", find)
			return nil, mty.ErrOwnerNoMatch
		}

	} else { //回滚删除交易
		if ownerOp.AddOrDel && find { //回滚add owner
			multiSig.Owners = delOwner(multiSig.Owners, index)
			//multiSig.Owners = append(multiSig.Owners[0:index], multiSig.Owners[index+1:]...)
		} else if !ownerOp.AddOrDel && !find { //回滚 del owner
			multiSig.Owners = append(multiSig.Owners, ownerOp.Owner)
		} else {
			multisiglog.Error("saveMultiSigOwnerAddOrDel", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "index", index, "find", find)
			return nil, mty.ErrOwnerNoMatch
		}
	}
	multisiglog.Error("saveMultiSigOwnerAddOrDel", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "multiSig", multiSig)

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	accountkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, accountkv)
	return kvs, nil
}

//账户owner的mod/replace操作
func (m *MultiSig) saveMultiSigOwnerModOrRep(ownerOp mty.ReceiptOwnerModOrRep, addOrRollback bool) ([]*types.KeyValue, error) {
	//获取多重签名账户信息从db中
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), ownerOp.MultiSigAddr)

	if err != nil || multiSig == nil {
		return nil, err
	}
	if addOrRollback { //正常添加交易
		_, index, _, _, find := getOwnerInfoByAddr(multiSig, ownerOp.PrevOwner.OwnerAddr)
		if ownerOp.ModOrRep && find { //modify owner weight
			multiSig.Owners[index].Weight = ownerOp.CurrentOwner.Weight
		} else if !ownerOp.ModOrRep && find { //replace owner addr
			multiSig.Owners[index].OwnerAddr = ownerOp.CurrentOwner.OwnerAddr
		} else {
			multisiglog.Error("saveMultiSigOwnerModOrRep ", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "index", index, "find", find)
			return nil, mty.ErrOwnerNoMatch
		}

	} else { //回滚删除交易
		_, index, _, _, find := getOwnerInfoByAddr(multiSig, ownerOp.CurrentOwner.OwnerAddr)
		if ownerOp.ModOrRep && find { //回滚modify owner weight
			multiSig.Owners[index].Weight = ownerOp.PrevOwner.Weight
		} else if !ownerOp.ModOrRep && find { //回滚 replace owner addr
			multiSig.Owners[index].OwnerAddr = ownerOp.PrevOwner.OwnerAddr
		} else {
			multisiglog.Error("saveMultiSigOwnerModOrRep ", "addOrRollback", addOrRollback, "ownerOp", ownerOp, "index", index, "find", find)
			return nil, mty.ErrOwnerNoMatch
		}
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	accountkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, accountkv)
	return kvs, nil
}

//账户weight权重的mod操作
func (m *MultiSig) saveMultiSigAccWeight(accountOp mty.ReceiptWeightModify, addOrRollback bool) ([]*types.KeyValue, error) {
	//获取多重签名账户信息从db中
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), accountOp.MultiSigAddr)

	if err != nil || multiSig == nil {
		return nil, err
	}
	if addOrRollback { //正常添加交易
		multiSig.RequiredWeight = accountOp.CurrentWeight
	} else { //回滚删除交易
		multiSig.RequiredWeight = accountOp.PrevWeight
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	accountkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, accountkv)
	return kvs, nil
}

//账户DailyLimit资产每日限额的add/mod操作
func (m *MultiSig) saveMultiSigAccDailyLimit(accountOp mty.ReceiptDailyLimitOperate, addOrRollback bool) ([]*types.KeyValue, error) {
	//获取多重签名账户信息从db中
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), accountOp.MultiSigAddr)

	if err != nil || multiSig == nil {
		return nil, err
	}
	curExecer := accountOp.CurDailyLimit.Execer
	curSymbol := accountOp.CurDailyLimit.Symbol
	curDailyLimit := accountOp.CurDailyLimit
	prevDailyLimit := accountOp.PrevDailyLimit

	//在每日限额列表中查找具体资产的每日限额信息
	index, find := isDailyLimit(multiSig, curExecer, curSymbol)

	if addOrRollback { //正常添加交易
		if accountOp.AddOrModify && !find { //add DailyLimit
			multiSig.DailyLimits = append(multiSig.DailyLimits, curDailyLimit)
		} else if !accountOp.AddOrModify && find { //modifyDailyLimit
			multiSig.DailyLimits[index].DailyLimit = curDailyLimit.DailyLimit
			multiSig.DailyLimits[index].SpentToday = curDailyLimit.SpentToday
			multiSig.DailyLimits[index].LastDay = curDailyLimit.LastDay
		} else {
			multisiglog.Error("saveMultiSigAccDailyLimit", "addOrRollback", addOrRollback, "accountOp", accountOp, "index", index, "find", find)
			return nil, mty.ErrDailyLimitNoMatch
		}
	} else { //回滚删除交易
		if accountOp.AddOrModify && find { //删除已经 add 的 DailyLimit
			multiSig.DailyLimits = append(multiSig.DailyLimits[0:index], multiSig.DailyLimits[index+1:]...)
		} else if !accountOp.AddOrModify && find { //恢复前一次的状态 modifyDailyLimit
			multiSig.DailyLimits[index].DailyLimit = prevDailyLimit.DailyLimit
			multiSig.DailyLimits[index].SpentToday = prevDailyLimit.SpentToday
			multiSig.DailyLimits[index].LastDay = prevDailyLimit.LastDay
		} else {
			multisiglog.Error("saveMultiSigAccDailyLimit", "addOrRollback", addOrRollback, "accountOp", accountOp, "index", index, "find", find)
			return nil, mty.ErrDailyLimitNoMatch
		}
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	accountkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, accountkv)
	return kvs, nil
}

//多重签名账户交易的Confirm/Revoke
func (m *MultiSig) saveMultiSigConfirmTx(confirmTx mty.ReceiptConfirmTx, addOrRollback bool) ([]*types.KeyValue, error) {
	multiSigAddr := confirmTx.MultiSigTxOwner.MultiSigAddr
	txid := confirmTx.MultiSigTxOwner.Txid
	owner := confirmTx.MultiSigTxOwner.ConfirmedOwner

	//获取多重签名交易信息从db中
	multiSigTx, err := getMultiSigTx(m.GetLocalDB(), multiSigAddr, txid)
	if err != nil {
		return nil, err
	}
	if multiSigTx == nil {
		multisiglog.Error("saveMultiSigConfirmTx", "addOrRollback", addOrRollback, "confirmTx", confirmTx)
		return nil, mty.ErrTxidNotExist
	}
	index, exist := isOwnerConfirmedTx(multiSigTx, owner.OwnerAddr)
	if addOrRollback { //正常添加交易
		if confirmTx.ConfirmeOrRevoke && !exist { //add Confirmed Owner
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner, owner)
		} else if !confirmTx.ConfirmeOrRevoke && exist { //Revoke Confirmed Owner
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner[0:index], multiSigTx.ConfirmedOwner[index+1:]...)
		} else {
			multisiglog.Error("saveMultiSigConfirmTx", "addOrRollback", addOrRollback, "confirmTx", confirmTx, "index", index, "exist", exist)
			return nil, mty.ErrDailyLimitNoMatch
		}
	} else { //回滚删除交易
		if confirmTx.ConfirmeOrRevoke && exist { //回滚已经 add Confirmed Owner
			//multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner, owner)
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner[0:index], multiSigTx.ConfirmedOwner[index+1:]...)

		} else if !confirmTx.ConfirmeOrRevoke && !exist { //回滚已经 Revoke Confirmed Owner
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner, owner)
		} else {
			multisiglog.Error("saveMultiSigConfirmTx", "addOrRollback", addOrRollback, "confirmTx", confirmTx, "index", index, "exist", exist)
			return nil, mty.ErrDailyLimitNoMatch
		}
	}

	err = setMultiSigTx(m.GetLocalDB(), multiSigTx, true)
	if err != nil {
		return nil, err
	}
	txkv := getMultiSigTxKV(multiSigTx, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, txkv)
	return kvs, nil
}

//多重签名账户交易被确认执行,更新交易的执行结果以及增加确认owner
//包含转账的交易以及修改多重签名账户属性的交易
func (m *MultiSig) saveMultiSigTx(execTx mty.ReceiptMultiSigTx, addOrRollback bool) ([]*types.KeyValue, error) {
	multiSigAddr := execTx.MultiSigTxOwner.MultiSigAddr
	txid := execTx.MultiSigTxOwner.Txid
	owner := execTx.MultiSigTxOwner.ConfirmedOwner
	curExecuted := execTx.CurExecuted
	prevExecuted := execTx.PrevExecuted
	submitOrConfirm := execTx.SubmitOrConfirm

	temMultiSigTx := &mty.MultiSigTx{}
	temMultiSigTx.MultiSigAddr = multiSigAddr
	temMultiSigTx.Txid = txid
	temMultiSigTx.TxHash = execTx.TxHash
	temMultiSigTx.TxType = execTx.TxType
	temMultiSigTx.Executed = false
	//获取多重签名交易信息从db中
	multiSigTx, err := getMultiSigTx(m.GetLocalDB(), multiSigAddr, txid)
	if err != nil {
		multisiglog.Error("saveMultiSigTx getMultiSigTx ", "addOrRollback", addOrRollback, "execTx", execTx, "err", err)
		return nil, err
	}

	//Confirm的交易需要确认对应的txid已经存在
	if multiSigTx == nil && !submitOrConfirm {
		multisiglog.Error("saveMultiSigTx", "addOrRollback", addOrRollback, "execTx", execTx)
		return nil, mty.ErrTxidNotExist
	}

	//add submit的交易需要创建txid，
	if submitOrConfirm && addOrRollback {
		if multiSigTx != nil {
			multisiglog.Error("saveMultiSigTx", "addOrRollback", addOrRollback, "execTx", execTx)
			return nil, mty.ErrTxidHasExist
		}
		multiSigTx = temMultiSigTx
	}

	index, exist := isOwnerConfirmedTx(multiSigTx, owner.OwnerAddr)
	if addOrRollback { //正常添加交易
		if !exist { //add Confirmed Owner and modify Executed
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner, owner)
			if prevExecuted != multiSigTx.Executed {
				return nil, mty.ErrExecutedNoMatch
			}
			multiSigTx.Executed = curExecuted
		} else {
			multisiglog.Error("saveMultiSigTx", "addOrRollback", addOrRollback, "execTx", execTx, "index", index, "exist", exist)
			return nil, mty.ErrOwnerNoMatch
		}
	} else { //回滚删除交易
		if exist { //回滚已经 add Confirmed Owner and modify Executed
			multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner[0:index], multiSigTx.ConfirmedOwner[index+1:]...)
			multiSigTx.Executed = prevExecuted
		} else {
			multisiglog.Error("saveMultiSigTx", "addOrRollback", addOrRollback, "execTx", execTx, "index", index, "exist", exist)
			return nil, mty.ErrOwnerNoMatch
		}
	}
	//submit交易的回滚需要将对应txid的值设置成nil
	setNil := true
	if !addOrRollback && submitOrConfirm {
		setNil = false
	}

	err = setMultiSigTx(m.GetLocalDB(), multiSigTx, setNil)
	if err != nil {
		return nil, err
	}
	txkv := getMultiSigTxKV(multiSigTx, setNil)

	var kvs []*types.KeyValue
	kvs = append(kvs, txkv)
	return kvs, nil
}

//多重签名账户交易被确认执行，更新对应资产的每日限额信息，以及txcount计数
func (m *MultiSig) saveDailyLimitUpdate(execTransfer mty.ReceiptAccDailyLimitUpdate, addOrRollback bool) ([]*types.KeyValue, error) {
	multiSigAddr := execTransfer.MultiSigAddr
	curDailyLimit := execTransfer.CurDailyLimit
	prevDailyLimit := execTransfer.PrevDailyLimit
	execer := execTransfer.CurDailyLimit.Execer
	symbol := execTransfer.CurDailyLimit.Symbol

	//获取多重签名交易信息从db中
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), multiSigAddr)
	if err != nil {
		return nil, err
	}
	if multiSig == nil {
		multisiglog.Error("saveAccExecTransfer", "addOrRollback", addOrRollback, "execTransfer", execTransfer)
		return nil, types.ErrAccountNotExist
	}
	index, exist := isDailyLimit(multiSig, execer, symbol)
	if !exist {
		return nil, types.ErrAccountNotExist
	}
	if addOrRollback { //正常添加交易
		multiSig.DailyLimits[index].SpentToday = curDailyLimit.SpentToday
		multiSig.DailyLimits[index].LastDay = curDailyLimit.LastDay
	} else { //回滚删除交易
		multiSig.DailyLimits[index].SpentToday = prevDailyLimit.SpentToday
		multiSig.DailyLimits[index].LastDay = prevDailyLimit.LastDay
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	txkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, txkv)
	return kvs, nil
}

//多重签名账户交易被确认执行，更新对应资产的每日限额信息，以及txcount计数
func (m *MultiSig) saveMultiSigTxCountUpdate(accTxCount mty.ReceiptTxCountUpdate, addOrRollback bool) ([]*types.KeyValue, error) {
	multiSigAddr := accTxCount.MultiSigAddr
	curTxCount := accTxCount.CurTxCount

	//获取多重签名交易信息从db中
	multiSig, err := getMultiSigAccount(m.GetLocalDB(), multiSigAddr)
	if err != nil {
		return nil, err
	}
	if multiSig == nil {
		multisiglog.Error("saveMultiSigTxCountUpdate", "addOrRollback", addOrRollback, "accTxCount", accTxCount)
		return nil, types.ErrAccountNotExist
	}

	if addOrRollback { //正常添加交易
		if multiSig.TxCount+1 == curTxCount {
			multiSig.TxCount = curTxCount
		} else {
			multisiglog.Error("saveMultiSigTxCountUpdate", "addOrRollback", addOrRollback, "accTxCount", accTxCount, "TxCount", multiSig.TxCount)
			return nil, mty.ErrInvalidTxid
		}
	} else { //回滚删除交易
		if multiSig.TxCount == curTxCount && curTxCount > 0 {
			multiSig.TxCount = curTxCount - 1
		}
	}

	err = setMultiSigAccount(m.GetLocalDB(), multiSig, true)
	if err != nil {
		return nil, err
	}
	txkv := getMultiSigAccountKV(multiSig, true)

	var kvs []*types.KeyValue
	kvs = append(kvs, txkv)
	return kvs, nil
}

//获取多重签名账户的指定资产
func (m *MultiSig) getMultiSigAccAssets(multiSigAddr string, assets *mty.Assets) (*types.Account, error) {
	symbol := getRealSymbol(assets.Symbol)

	acc, err := account.NewAccountDB(assets.Execer, symbol, m.GetStateDB())
	if err != nil {
		return &types.Account{}, err
	}
	var acc1 *types.Account

	execaddress := dapp.ExecAddress(types.ExecName(m.GetName()))
	acc1 = acc.LoadExecAccount(multiSigAddr, execaddress)
	return acc1, nil
}

//内部共用接口

//获取指定owner的weight权重，owner所在的index，所有owners的weight权重之和，以及owner是否存在
func getOwnerInfoByAddr(multiSigAcc *mty.MultiSig, oldowner string) (uint64, int, uint64, int, bool) {
	//首先遍历所有owners，确定对应的owner已近存在.
	var findindex int
	var totalweight uint64
	var oldweight uint64
	var totalowner int
	flag := false

	for index, owner := range multiSigAcc.Owners {
		if owner.OwnerAddr == oldowner {
			flag = true
			findindex = index
			oldweight = owner.Weight
		}
		totalweight += owner.Weight
		totalowner++
	}
	//owner不存在
	if !flag {
		return 0, 0, totalweight, totalowner, false
	}
	return oldweight, findindex, totalweight, totalowner, true
}

//确认某笔交易是否已经达到确认需要的权重
func isConfirmed(requiredWeight uint64, multiSigTx *mty.MultiSigTx) bool {
	var totalweight uint64
	for _, owner := range multiSigTx.ConfirmedOwner {
		totalweight += owner.Weight
	}
	return totalweight >= requiredWeight
}

//确认某笔交易的额度是否满足每日限额,返回是否满足，以及新的newLastDay时间
func isUnderLimit(blocktime int64, amount uint64, dailyLimit *mty.DailyLimit) (bool, int64) {

	var lastDay int64
	var newSpentToday uint64

	nowtime := blocktime //types.Now().Unix()
	newSpentToday = dailyLimit.SpentToday

	//已经是新的一天了。需要更新LastDay为当前时间，SpentToday今日花费0
	if nowtime > dailyLimit.LastDay+mty.OneDaySecond {
		lastDay = nowtime
		newSpentToday = 0
	}

	if newSpentToday+amount > dailyLimit.DailyLimit || newSpentToday+amount < newSpentToday {
		return false, lastDay
	}
	return true, lastDay
}

//确定这个地址是否是此multiSigAcc多重签名账户的owner,如果是owner的话并返回weight权重
func isOwner(multiSigAcc *mty.MultiSig, ownerAddr string) (uint64, bool) {
	for _, owner := range multiSigAcc.Owners {
		if owner.OwnerAddr == ownerAddr {
			return owner.Weight, true
		}
	}
	return 0, false
}

//删除指定index的owner从owners列表中
func delOwner(Owners []*mty.Owner, index int) []*mty.Owner {
	ownerSize := len(Owners)
	multisiglog.Error("delOwner", "ownerSize", ownerSize, "index", index)

	//删除第一个owner
	if index == 0 {
		Owners = Owners[1:]
	} else if (ownerSize) == index+1 { //删除最后一个owner
		multisiglog.Error("delOwner", "ownerSize", ownerSize)
		Owners = Owners[0 : ownerSize-1]
	} else {
		Owners = append(Owners[0:index], Owners[index+1:]...)
	}
	return Owners
}

//指定资产是否设置了每日限额
func isDailyLimit(multiSigAcc *mty.MultiSig, execer, symbol string) (int, bool) {
	for index, dailyLimit := range multiSigAcc.DailyLimits {
		if dailyLimit.Execer == execer && dailyLimit.Symbol == symbol {
			return index, true
		}
	}
	return 0, false
}

//owner是否已经确认过某个txid，已经确认过就返回index
func isOwnerConfirmedTx(multiSigTx *mty.MultiSigTx, ownerAddr string) (int, bool) {
	for index, owner := range multiSigTx.ConfirmedOwner {
		if owner.OwnerAddr == ownerAddr {
			return index, true
		}
	}
	return 0, false
}

//通过txhash获取tx交易信息
func getTxByHash(api client.QueueProtocolAPI, txHash string) (*types.TransactionDetail, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		multisiglog.Error("GetTxByHash DecodeString ", "hash", txHash)
		return nil, err
	}
	txs, err := api.GetTransactionByHash(&types.ReqHashes{Hashes: [][]byte{hash}})
	if err != nil {
		multisiglog.Error("GetTxByHash", "hash", txHash)
		return nil, err
	}
	if len(txs.Txs) != 1 {
		multisiglog.Error("GetTxByHash", "len is not 1", len(txs.Txs))
		return nil, mty.ErrTxHashNoMatch
	}
	if txs.Txs == nil {
		multisiglog.Error("GetTxByHash", "tx hash not found", txHash)
		return nil, mty.ErrTxHashNoMatch
	}
	return txs.Txs[0], nil
}

//从tx交易中解析payload信息
func getMultiSigTxPayload(tx *types.TransactionDetail) (*mty.MultiSigAction, error) {
	if !bytes.HasSuffix(tx.Tx.Execer, []byte(mty.MultiSigX)) {
		multisiglog.Error("GetMultiSigTx", "tx.Tx.Execer", string(tx.Tx.Execer), "MultiSigX", mty.MultiSigX)
		return nil, mty.ErrExecerHashNoMatch
	}
	var payload mty.MultiSigAction
	err := types.Decode(tx.Tx.Payload, &payload)
	if err != nil {
		multisiglog.Error("GetMultiSigTx:Decode Payload", "error", err)
		return nil, err
	}
	multisiglog.Error("GetMultiSigTx:Decode Payload", "payload", payload)
	return &payload, nil
}

//bty 显示是大写，在底层mavl数据库中对应key值时使用小写
func getRealSymbol(symbol string) string {
	if symbol == types.BTY {
		return "bty"
	}
	return symbol
}

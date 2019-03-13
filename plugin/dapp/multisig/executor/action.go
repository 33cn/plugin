// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

//action 结构体
type action struct {
	coinsAccount *account.DB
	db           dbm.KV
	localdb      dbm.KVDB
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	index        int32
	execaddr     string
	api          client.QueueProtocolAPI
}

func newAction(t *MultiSig, tx *types.Transaction, index int32) *action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &action{t.GetCoinsAccount(), t.GetStateDB(), t.GetLocalDB(), hash, fromaddr,
		t.GetBlockTime(), t.GetHeight(), index, dapp.ExecAddress(string(tx.Execer)), t.GetAPI()}
}

//MultiSigAccCreate 创建多重签名账户
func (a *action) MultiSigAccCreate(accountCreate *mty.MultiSigAccCreate) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var totalweight uint64
	var ownerCount int
	var dailyLimit mty.DailyLimit

	//参数校验
	if accountCreate == nil {
		return nil, types.ErrInvalidParam
	}
	//创建时requiredweight权重的值不能大于所有owner权重之和
	for _, owner := range accountCreate.Owners {
		if owner != nil {
			totalweight += owner.Weight
			ownerCount = ownerCount + 1
		}
	}

	if accountCreate.RequiredWeight > totalweight {
		return nil, mty.ErrRequiredweight
	}

	//创建时最少设置两个owner
	if ownerCount < mty.MinOwnersInit {
		return nil, mty.ErrOwnerLessThanTwo
	}
	//owner总数不能大于最大值
	if ownerCount > mty.MaxOwnersCount {
		return nil, mty.ErrMaxOwnerCount
	}

	multiSigAccount := &mty.MultiSig{}
	multiSigAccount.CreateAddr = a.fromaddr
	multiSigAccount.Owners = accountCreate.Owners
	multiSigAccount.TxCount = 0
	multiSigAccount.RequiredWeight = accountCreate.RequiredWeight

	//获取资产的每日限额设置
	if accountCreate.DailyLimit != nil {
		symbol := accountCreate.DailyLimit.Symbol
		execer := accountCreate.DailyLimit.Execer
		err := mty.IsAssetsInvalid(execer, symbol)
		if err != nil {
			return nil, err
		}
		dailyLimit.Symbol = symbol
		dailyLimit.Execer = execer
		dailyLimit.DailyLimit = accountCreate.DailyLimit.DailyLimit
		dailyLimit.SpentToday = 0
		dailyLimit.LastDay = a.blocktime //types.Now().Unix()
		multiSigAccount.DailyLimits = append(multiSigAccount.DailyLimits, &dailyLimit)
	}
	//通过创建交易的txhash生成一个唯一的多重签名合约 NewAddrFromString
	addr := address.MultiSignAddress(a.txhash)
	//账户去重校验
	multiSig, err := getMultiSigAccFromDb(a.db, addr)
	if err == nil && multiSig != nil {
		return nil, mty.ErrAccountHasExist
	}

	multiSigAccount.MultiSigAddr = addr
	receiptLog := &types.ReceiptLog{}
	receiptLog.Ty = mty.TyLogMultiSigAccCreate
	receiptLog.Log = types.Encode(multiSigAccount)
	logs = append(logs, receiptLog)

	key, value := setMultiSigAccToDb(a.db, multiSigAccount)
	kv = append(kv, &types.KeyValue{Key: key, Value: value})

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//MultiSigAccOperate 多重签名账户属性的修改：weight权重以及每日限额的修改
//修改requiredweight权重值不能大于当前所有owner权重之和
func (a *action) MultiSigAccOperate(AccountOperate *mty.MultiSigAccOperate) (*types.Receipt, error) {

	if AccountOperate == nil {
		return nil, types.ErrInvalidParam
	}
	//首先从statedb中获取MultiSigAccAddr的状态信息
	multiSigAccAddr := AccountOperate.MultiSigAccAddr
	multiSigAccount, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("MultiSigAccountOperate", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, err
	}

	if multiSigAccount == nil {
		multisiglog.Error("MultiSigAccountOperate:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, types.ErrAccountNotExist
	}

	//校验交易提交者是否是本账户的owner
	owneraddr := a.fromaddr
	ownerWeight, isowner := isOwner(multiSigAccount, owneraddr)
	if !isowner {
		return nil, mty.ErrIsNotOwner
	}

	//dailylimit每日限额属性的修改需要校验assets资产的合法性
	if !AccountOperate.OperateFlag {
		execer := AccountOperate.DailyLimit.Execer
		symbol := AccountOperate.DailyLimit.Symbol
		err := mty.IsAssetsInvalid(execer, symbol)
		if err != nil {
			return nil, err
		}
	}
	//生成新的txid,并将此交易信息添加到Txs列表中
	txID := multiSigAccount.TxCount
	newMultiSigTx := &mty.MultiSigTx{}
	newMultiSigTx.Txid = txID
	newMultiSigTx.TxHash = hex.EncodeToString(a.txhash)
	newMultiSigTx.Executed = false
	newMultiSigTx.TxType = mty.AccountOperate
	newMultiSigTx.MultiSigAddr = multiSigAccAddr
	confirmOwner := &mty.Owner{OwnerAddr: owneraddr, Weight: ownerWeight}
	newMultiSigTx.ConfirmedOwner = append(newMultiSigTx.ConfirmedOwner, confirmOwner)

	return a.executeAccOperateTx(multiSigAccount, newMultiSigTx, AccountOperate, confirmOwner, true)
}

//MultiSigOwnerOperate 多重签名账户owner属性的修改：owner的add/del/replace等
//在del和replace owner时需要保证修改后所有owner的权重之和大于requiredweight值
func (a *action) MultiSigOwnerOperate(AccOwnerOperate *mty.MultiSigOwnerOperate) (*types.Receipt, error) {
	multiSigAccAddr := AccOwnerOperate.MultiSigAccAddr

	//首先从statedb中获取MultiSigAccAddr的状态信息
	multiSigAccount, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("MultiSigAccountOperate", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, err
	}
	if multiSigAccount == nil {
		multisiglog.Error("MultiSigAccountOwnerOperate:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, types.ErrAccountNotExist
	}

	//校验交易提交者是否是本账户的owner
	owneraddr := a.fromaddr
	ownerWeight, isowner := isOwner(multiSigAccount, owneraddr)
	if !isowner {
		return nil, mty.ErrIsNotOwner
	}

	//生成新的txid,并将此交易信息添加到Txs列表中
	txID := multiSigAccount.TxCount
	newMultiSigTx := &mty.MultiSigTx{}
	newMultiSigTx.Txid = txID
	newMultiSigTx.TxHash = hex.EncodeToString(a.txhash)
	newMultiSigTx.Executed = false
	newMultiSigTx.TxType = mty.OwnerOperate
	newMultiSigTx.MultiSigAddr = multiSigAccAddr
	confirmOwner := &mty.Owner{OwnerAddr: owneraddr, Weight: ownerWeight}
	newMultiSigTx.ConfirmedOwner = append(newMultiSigTx.ConfirmedOwner, confirmOwner)

	return a.executeOwnerOperateTx(multiSigAccount, newMultiSigTx, AccOwnerOperate, confirmOwner, true)
}

//MultiSigExecTransferFrom 首先判断转账的额度是否大于每日限量，小于就直接执行交易，调用ExecTransferFrozen进行转账
//大于每日限量只需要将交易信息记录
//合约中多重签名账户转账到外部账户，multiSigAddr--->Addr
func (a *action) MultiSigExecTransferFrom(multiSigAccTransfer *mty.MultiSigExecTransferFrom) (*types.Receipt, error) {

	//首先从statedb中获取MultiSigAccAddr的状态信息
	multiSigAccAddr := multiSigAccTransfer.From
	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("MultiSigAccExecTransfer", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, err
	}

	// to 地址必须不是多重签名账户地址
	multiSigAccTo, err := getMultiSigAccFromDb(a.db, multiSigAccTransfer.To)
	if multiSigAccTo != nil && err == nil {
		multisiglog.Error("MultiSigExecTransferFrom", "multiSigAccTo", multiSigAccTo, "ToAddr", multiSigAccTransfer.To)
		return nil, mty.ErrAddrNotSupport
	}

	//校验交易提交者是否是本账户的owner
	owneraddr := a.fromaddr
	ownerWeight, isowner := isOwner(multiSigAcc, owneraddr)
	if !isowner {
		return nil, mty.ErrIsNotOwner
	}

	//assete资产合法性校验
	err = mty.IsAssetsInvalid(multiSigAccTransfer.Execname, multiSigAccTransfer.Symbol)
	if err != nil {
		return nil, err
	}
	//生成新的txid,并将此交易信息添加到Txs列表中
	txID := multiSigAcc.TxCount
	newMultiSigTx := &mty.MultiSigTx{}
	newMultiSigTx.Txid = txID
	newMultiSigTx.TxHash = hex.EncodeToString(a.txhash)
	newMultiSigTx.Executed = false
	newMultiSigTx.TxType = mty.TransferOperate
	newMultiSigTx.MultiSigAddr = multiSigAccAddr
	confirmOwner := &mty.Owner{OwnerAddr: owneraddr, Weight: ownerWeight}
	newMultiSigTx.ConfirmedOwner = append(newMultiSigTx.ConfirmedOwner, confirmOwner)

	//确认并执行此交易
	return a.executeTransferTx(multiSigAcc, newMultiSigTx, multiSigAccTransfer, confirmOwner, mty.IsSubmit)
}

//MultiSigExecTransferTo 将合约中外部账户转账上的Execname.Symbol资产转到多重签名账户上，from:Addr --->to:multiSigAddr
// from地址使用tx中的签名的地址，payload中from地址不使用在 TransferTo交易中
func (a *action) MultiSigExecTransferTo(execTransfer *mty.MultiSigExecTransferTo) (*types.Receipt, error) {

	//from地址校验必须不是多重签名账户地址
	multiSigAccFrom, err := getMultiSigAccFromDb(a.db, a.fromaddr)
	if multiSigAccFrom != nil && err == nil {
		multisiglog.Error("MultiSigExecTransferTo", "multiSigAccFrom", multiSigAccFrom, "From", a.fromaddr)
		return nil, mty.ErrAddrNotSupport
	}
	// to 地址必须是多重签名账户地址
	multiSigAccTo, err := getMultiSigAccFromDb(a.db, execTransfer.To)
	if multiSigAccTo == nil || err != nil {
		multisiglog.Error("MultiSigExecTransferTo", "ToAddr", execTransfer.To)
		return nil, mty.ErrAddrNotSupport
	}
	//assete资产合法性校验
	err = mty.IsAssetsInvalid(execTransfer.Execname, execTransfer.Symbol)
	if err != nil {
		return nil, err
	}

	//将指定账户上的资产从balance转账到多重签名账户的balance上
	symbol := getRealSymbol(execTransfer.Symbol)
	newAccountDB, err := account.NewAccountDB(execTransfer.Execname, symbol, a.db)
	if err != nil {
		return nil, err
	}
	receiptExecTransfer, err := newAccountDB.ExecTransfer(a.fromaddr, execTransfer.To, a.execaddr, execTransfer.Amount)
	if err != nil {
		multisiglog.Error("MultiSigExecTransfer:ExecTransfer", "From", a.fromaddr,
			"To", execTransfer.To, "execaddr", a.execaddr,
			"amount", execTransfer.Amount, "Execer", execTransfer.Execname, "Symbol", execTransfer.Symbol, "error", err)
		return nil, err
	}
	//将多重签名账户上的balance冻结起来
	receiptExecFrozen, err := newAccountDB.ExecFrozen(execTransfer.To, a.execaddr, execTransfer.Amount)
	if err != nil {
		multisiglog.Error("MultiSigExecTransfer:ExecFrozen", "addr", execTransfer.To, "execaddr", a.execaddr,
			"amount", execTransfer.Amount, "Execer", execTransfer.Execname, "Symbol", execTransfer.Symbol, "error", err)
		return nil, err
	}

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	logs = append(logs, receiptExecTransfer.Logs...)
	logs = append(logs, receiptExecFrozen.Logs...)
	kv = append(kv, receiptExecTransfer.KV...)
	kv = append(kv, receiptExecFrozen.KV...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//MultiSigConfirmTx 多重签名账户上MultiSigAcc账户Transfer交易的确认和撤销
//确认交易需要判断权重是否满足，满足就直接执行交易，调用ExecTransferFrozen进行转账
//不满足就只更新本交易的确认owner
//撤销确认交易，只允许撤销还没有被执行的交易，只更新本交易的确认owner
func (a *action) MultiSigConfirmTx(ConfirmTx *mty.MultiSigConfirmTx) (*types.Receipt, error) {

	//首先从statedb中获取MultiSigAccAddr的状态信息
	multiSigAccAddr := ConfirmTx.MultiSigAccAddr
	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("MultiSigConfirmTx:getMultiSigAccFromDb", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, err
	}
	//校验交易提交者是否是本账户的owner
	owneraddr := a.fromaddr
	ownerWeight, isowner := isOwner(multiSigAcc, owneraddr)
	if !isowner {
		multisiglog.Error("MultiSigConfirmTx: is not Owner", "MultiSigAccAddr", multiSigAccAddr, "txFrom", owneraddr, "err", err)
		return nil, mty.ErrIsNotOwner
	}
	//TxId的合法性校验
	if ConfirmTx.TxId > multiSigAcc.TxCount {
		multisiglog.Error("MultiSigConfirmTx: Invalid Txid", "MultiSigAccTxCount", multiSigAcc.TxCount, "Confirm TxId", ConfirmTx.TxId, "err", err)
		return nil, mty.ErrInvalidTxid
	}
	//获取多重签名账户上txid对应的交易信息
	multiSigTx, err := getMultiSigAccTxFromDb(a.db, multiSigAccAddr, ConfirmTx.TxId)
	if err != nil {
		multisiglog.Error("MultiSigConfirmTx:getMultiSigAccTxFromDb", "multiSigAccAddr", multiSigAccAddr, "Confirm TxId", ConfirmTx.TxId, "err", err)
		return nil, mty.ErrTxidNotExist
	}
	//已经被执行的交易不可以再确认/撤销
	if multiSigTx.Executed {
		return nil, mty.ErrTxHasExecuted
	}
	//此owneraddr是否已经确认过此txid对应的交易
	findindex, exist := isOwnerConfirmedTx(multiSigTx, owneraddr)

	//不能重复确认同一笔交易直接返回
	if exist && ConfirmTx.ConfirmOrRevoke {
		return nil, mty.ErrDupConfirmed
	}
	//需要撤销的确认信息没有找到直接返回
	if !exist && !ConfirmTx.ConfirmOrRevoke {
		return nil, mty.ErrConfirmNotExist
	}

	owner := &mty.Owner{OwnerAddr: owneraddr, Weight: ownerWeight}

	//首先处理撤销确认交易，将owneraddr的确认信息从交易确认列表中删除
	if exist && !ConfirmTx.ConfirmOrRevoke {
		multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner[0:findindex], multiSigTx.ConfirmedOwner[findindex+1:]...)
	} else if !exist && ConfirmTx.ConfirmOrRevoke {
		//增加此owner的确认信息到multiSigTx的确认列表中
		multiSigTx.ConfirmedOwner = append(multiSigTx.ConfirmedOwner, owner)
	}

	multiSigTxOwner := &mty.MultiSigTxOwner{MultiSigAddr: multiSigAccAddr, Txid: ConfirmTx.TxId, ConfirmedOwner: owner}
	isConfirm := isConfirmed(multiSigAcc.RequiredWeight, multiSigTx)

	//权重未达到要求或者撤销确认交易，构造MultiSigConfirmTx的receiptLog
	if !isConfirm || !ConfirmTx.ConfirmOrRevoke {
		return a.confirmTransaction(multiSigTx, multiSigTxOwner, ConfirmTx.ConfirmOrRevoke)
	}
	//获取txhash对应交易详细信息
	tx, err := getTxByHash(a.api, multiSigTx.TxHash)
	if err != nil {
		return nil, err
	}
	payload, err := getMultiSigTxPayload(tx)
	if err != nil {
		return nil, err
	}

	//根据不同的交易类型调用各自的处理函数，区分 操作owner/account 和转账的交易
	if multiSigTx.TxType == mty.OwnerOperate && payload != nil {
		transfer := payload.GetMultiSigOwnerOperate()
		return a.executeOwnerOperateTx(multiSigAcc, multiSigTx, transfer, owner, false)
	} else if multiSigTx.TxType == mty.AccountOperate {
		transfer := payload.GetMultiSigAccOperate()
		return a.executeAccOperateTx(multiSigAcc, multiSigTx, transfer, owner, false)
	} else if multiSigTx.TxType == mty.TransferOperate {
		transfer := payload.GetMultiSigExecTransferFrom()
		return a.executeTransferTx(multiSigAcc, multiSigTx, transfer, owner, mty.IsConfirm)
	}
	multisiglog.Error("MultiSigConfirmTx:GetMultiSigTx", "multiSigAccAddr", multiSigAccAddr, "Confirm TxId", ConfirmTx.TxId, "TxType unknown", multiSigTx.TxType)
	return nil, mty.ErrTxTypeNoMatch
}

//多重签名账户请求权重的修改,返回新的KeyValue对和ReceiptLog信息
func (a *action) multiSigWeightModify(multiSigAccAddr string, newRequiredWeight uint64) (*types.KeyValue, *types.ReceiptLog, error) {

	multiSigAccount, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("multiSigWeightModify", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAccount == nil {
		multisiglog.Error("multiSigWeightModify:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}

	//首先获取所有owner的权重之和，新设置的newRequiredWeight不能大于owner的权重之和
	var totalweight uint64
	receiptLog := &types.ReceiptLog{}
	for _, owner := range multiSigAccount.Owners {
		if owner != nil {
			totalweight += owner.Weight
		}
	}
	if newRequiredWeight > totalweight {
		return nil, nil, mty.ErrRequiredweight
	}
	//修改RequiredWeight字段
	prevWeight := multiSigAccount.RequiredWeight
	multiSigAccount.RequiredWeight = newRequiredWeight

	//组装receiptLog
	receiptWeight := &mty.ReceiptWeightModify{}
	receiptWeight.MultiSigAddr = multiSigAccount.MultiSigAddr
	receiptWeight.PrevWeight = prevWeight
	receiptWeight.CurrentWeight = multiSigAccount.RequiredWeight
	receiptLog.Ty = mty.TyLogMultiSigAccWeightModify
	receiptLog.Log = types.Encode(receiptWeight)

	key, value := setMultiSigAccToDb(a.db, multiSigAccount)
	kv := &types.KeyValue{Key: key, Value: value}
	return kv, receiptLog, nil
}

//多重签名账户资产每日限额的添加或者修改,
func (a *action) multiSigDailyLimitOperate(multiSigAccAddr string, dailylimit *mty.SymbolDailyLimit) (*types.KeyValue, *types.ReceiptLog, error) {

	multiSigAccount, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("multiSigDailyLimitOperate", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAccount == nil {
		multisiglog.Error("multiSigDailyLimitOperate:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}

	flag := false
	var addOrModify bool
	var findindex int
	var curDailyLimit *mty.DailyLimit

	receiptLog := &types.ReceiptLog{}

	newSymbol := dailylimit.Symbol
	newExecer := dailylimit.Execer
	newDailyLimit := dailylimit.DailyLimit

	prevDailyLimit := &mty.DailyLimit{Symbol: newSymbol, Execer: newExecer, DailyLimit: 0, SpentToday: 0, LastDay: 0}
	//首先遍历获取需要修改的symbol每日限额,没有找到就添加
	for index, dailyLimit := range multiSigAccount.DailyLimits {
		if dailyLimit.Symbol == newSymbol && dailyLimit.Execer == newExecer {
			prevDailyLimit.DailyLimit = dailyLimit.DailyLimit
			prevDailyLimit.SpentToday = dailyLimit.SpentToday
			prevDailyLimit.LastDay = dailyLimit.LastDay
			flag = true
			findindex = index
			break
		}
	}
	if flag { //modify old DailyLimit
		multiSigAccount.DailyLimits[findindex].DailyLimit = newDailyLimit
		curDailyLimit = multiSigAccount.DailyLimits[findindex]
		addOrModify = false
	} else { //add new DailyLimit
		temDailyLimit := &mty.DailyLimit{}
		temDailyLimit.Symbol = newSymbol
		temDailyLimit.Execer = newExecer
		temDailyLimit.DailyLimit = newDailyLimit
		temDailyLimit.SpentToday = 0
		temDailyLimit.LastDay = a.blocktime //types.Now().Unix()
		multiSigAccount.DailyLimits = append(multiSigAccount.DailyLimits, temDailyLimit)

		curDailyLimit = temDailyLimit
		addOrModify = true
	}
	receiptDailyLimit := &mty.ReceiptDailyLimitOperate{
		MultiSigAddr:   multiSigAccount.MultiSigAddr,
		PrevDailyLimit: prevDailyLimit,
		CurDailyLimit:  curDailyLimit,
		AddOrModify:    addOrModify,
	}
	receiptLog.Ty = mty.TyLogMultiSigAccDailyLimitModify
	receiptLog.Log = types.Encode(receiptDailyLimit)

	key, value := setMultiSigAccToDb(a.db, multiSigAccount)
	kv := &types.KeyValue{Key: key, Value: value}
	return kv, receiptLog, nil
}

//多重签名账户的添加,返回新的KeyValue对和ReceiptLog信息
func (a *action) multiSigOwnerAdd(multiSigAccAddr string, AccOwnerOperate *mty.MultiSigOwnerOperate) (*types.KeyValue, *types.ReceiptLog, error) {

	//添加newowner到账户的owner中
	var newOwner mty.Owner
	newOwner.OwnerAddr = AccOwnerOperate.NewOwner
	newOwner.Weight = AccOwnerOperate.NewWeight
	return a.receiptOwnerAddOrDel(multiSigAccAddr, &newOwner, true)
}

//多重签名账户的删除,返回新的KeyValue对和ReceiptLog信息
func (a *action) multiSigOwnerDel(multiSigAccAddr string, AccOwnerOperate *mty.MultiSigOwnerOperate) (*types.KeyValue, *types.ReceiptLog, error) {
	var owner mty.Owner
	owner.OwnerAddr = AccOwnerOperate.OldOwner
	owner.Weight = 0
	return a.receiptOwnerAddOrDel(multiSigAccAddr, &owner, false)
}

//组装add/del owner的receipt信息
func (a *action) receiptOwnerAddOrDel(multiSigAccAddr string, owner *mty.Owner, addOrDel bool) (*types.KeyValue, *types.ReceiptLog, error) {
	receiptLog := &types.ReceiptLog{}

	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("receiptOwnerAddOrDel", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAcc == nil {
		multisiglog.Error("receiptOwnerAddOrDel:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}

	oldweight, index, totalWeight, totalowner, exist := getOwnerInfoByAddr(multiSigAcc, owner.OwnerAddr)

	if addOrDel {
		if exist {
			return nil, nil, mty.ErrOwnerExist
		}
		if totalowner >= mty.MaxOwnersCount {
			return nil, nil, mty.ErrMaxOwnerCount
		}
		multiSigAcc.Owners = append(multiSigAcc.Owners, owner)
		receiptLog.Ty = mty.TyLogMultiSigOwnerAdd
	} else {
		if !exist {
			return nil, nil, mty.ErrOwnerNotExist
		}
		//删除时需要确认删除后所有owners的权重之和必须大于reqweight
		if totalWeight-oldweight < multiSigAcc.RequiredWeight {
			return nil, nil, mty.ErrTotalWeightNotEnough
		}
		//最少要保留一个owner
		if totalowner <= 1 {
			return nil, nil, mty.ErrOnlyOneOwner
		}
		owner.Weight = oldweight
		receiptLog.Ty = mty.TyLogMultiSigOwnerDel
		multiSigAcc.Owners = delOwner(multiSigAcc.Owners, index)
	}

	//组装receiptLog
	receiptOwner := &mty.ReceiptOwnerAddOrDel{}
	receiptOwner.MultiSigAddr = multiSigAcc.MultiSigAddr
	receiptOwner.Owner = owner
	receiptOwner.AddOrDel = addOrDel
	receiptLog.Log = types.Encode(receiptOwner)

	key, value := setMultiSigAccToDb(a.db, multiSigAcc)
	keyValue := &types.KeyValue{Key: key, Value: value}
	return keyValue, receiptLog, nil
}

//多重签名账户owner的修改,返回新的KeyValue对和ReceiptLog信息
func (a *action) multiSigOwnerModify(multiSigAccAddr string, AccOwnerOperate *mty.MultiSigOwnerOperate) (*types.KeyValue, *types.ReceiptLog, error) {

	prev := &mty.Owner{OwnerAddr: AccOwnerOperate.OldOwner, Weight: 0}
	cur := &mty.Owner{OwnerAddr: AccOwnerOperate.OldOwner, Weight: AccOwnerOperate.NewWeight}
	return a.receiptOwnerModOrRep(multiSigAccAddr, prev, cur, true)
}

//多重签名账户owner的替换,返回新的KeyValue对和ReceiptLog信息
func (a *action) multiSigOwnerReplace(multiSigAccAddr string, AccOwnerOperate *mty.MultiSigOwnerOperate) (*types.KeyValue, *types.ReceiptLog, error) {

	prev := &mty.Owner{OwnerAddr: AccOwnerOperate.OldOwner, Weight: 0}
	cur := &mty.Owner{OwnerAddr: AccOwnerOperate.NewOwner, Weight: 0}
	return a.receiptOwnerModOrRep(multiSigAccAddr, prev, cur, false)
}

//组装修改/替换owner的receipt信息
func (a *action) receiptOwnerModOrRep(multiSigAccAddr string, prev *mty.Owner, cur *mty.Owner, modOrRep bool) (*types.KeyValue, *types.ReceiptLog, error) {
	receiptLog := &types.ReceiptLog{}

	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("receiptOwnerModOrRep", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAcc == nil {
		multisiglog.Error("receiptOwnerModOrRep:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}
	oldweight, index, totalWeight, _, exist := getOwnerInfoByAddr(multiSigAcc, prev.OwnerAddr)
	if modOrRep {
		if !exist {
			return nil, nil, mty.ErrOwnerNotExist
		}
		//修改时需要确认修改后所有owners的权重之和必须大于reqweight
		if totalWeight-oldweight+cur.Weight < multiSigAcc.RequiredWeight {
			return nil, nil, mty.ErrTotalWeightNotEnough
		}
		prev.Weight = oldweight
		multiSigAcc.Owners[index].Weight = cur.Weight
		receiptLog.Ty = mty.TyLogMultiSigOwnerModify
	} else {
		if !exist {
			return nil, nil, mty.ErrOwnerNotExist
		}
		//替换时newowner应该不存在
		_, _, _, _, find := getOwnerInfoByAddr(multiSigAcc, cur.OwnerAddr)
		if find {
			return nil, nil, mty.ErrNewOwnerExist
		}
		prev.Weight = oldweight
		cur.Weight = oldweight
		multiSigAcc.Owners[index].OwnerAddr = cur.OwnerAddr
		receiptLog.Ty = mty.TyLogMultiSigOwnerReplace
	}
	//组装receiptLog
	receiptAddOwner := &mty.ReceiptOwnerModOrRep{}
	receiptAddOwner.MultiSigAddr = multiSigAcc.MultiSigAddr
	receiptAddOwner.PrevOwner = prev
	receiptAddOwner.CurrentOwner = cur
	receiptAddOwner.ModOrRep = modOrRep
	receiptLog.Log = types.Encode(receiptAddOwner)

	key, value := setMultiSigAccToDb(a.db, multiSigAcc)
	keyValue := &types.KeyValue{Key: key, Value: value}
	return keyValue, receiptLog, nil
}

//组装AccExecTransfer的receipt信息,需要区分是在提交交易时执行的，还是在确认阶段执行的交易
func (a *action) receiptDailyLimitUpdate(multiSigAccAddr string, findindex int, curdailyLimit *mty.DailyLimit) (*types.KeyValue, *types.ReceiptLog, error) {

	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("receiptDailyLimitUpdate", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAcc == nil {
		multisiglog.Error("receiptDailyLimitUpdate:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}
	receiptLog := &types.ReceiptLog{}

	//组装receiptLog
	receipt := &mty.ReceiptAccDailyLimitUpdate{}
	receipt.MultiSigAddr = multiSigAcc.MultiSigAddr
	receipt.PrevDailyLimit = multiSigAcc.DailyLimits[findindex]
	receipt.CurDailyLimit = curdailyLimit
	receiptLog.Ty = mty.TyLogDailyLimitUpdate
	receiptLog.Log = types.Encode(receipt)

	//更新DailyLimit
	multiSigAcc.DailyLimits[findindex].SpentToday = curdailyLimit.SpentToday
	multiSigAcc.DailyLimits[findindex].LastDay = curdailyLimit.LastDay

	key, value := setMultiSigAccToDb(a.db, multiSigAcc)
	keyValue := &types.KeyValue{Key: key, Value: value}
	return keyValue, receiptLog, nil
}

//组装修改账户属性时交易计数的增加和的receipt信息
func (a *action) receiptTxCountUpdate(multiSigAccAddr string) (*types.KeyValue, *types.ReceiptLog, error) {

	multiSigAcc, err := getMultiSigAccFromDb(a.db, multiSigAccAddr)
	if err != nil {
		multisiglog.Error("receiptTxCountUpdate", "MultiSigAccAddr", multiSigAccAddr, "err", err)
		return nil, nil, err
	}

	if multiSigAcc == nil {
		multisiglog.Error("receiptTxCountUpdate:getMultiSigAccFromDb is nil", "MultiSigAccAddr", multiSigAccAddr)
		return nil, nil, types.ErrAccountNotExist
	}

	receiptLog := &types.ReceiptLog{}

	//组装receiptLog
	multiSigAcc.TxCount++

	receiptLogTxCount := &mty.ReceiptTxCountUpdate{
		MultiSigAddr: multiSigAcc.MultiSigAddr,
		CurTxCount:   multiSigAcc.TxCount,
	}

	receiptLog.Ty = mty.TyLogTxCountUpdate
	receiptLog.Log = types.Encode(receiptLogTxCount)

	key, value := setMultiSigAccToDb(a.db, multiSigAcc)
	keyValue := &types.KeyValue{Key: key, Value: value}
	return keyValue, receiptLog, nil
}

//组装MultiSigAccTx的receipt信息
func (a *action) receiptMultiSigTx(multiSigTx *mty.MultiSigTx, owner *mty.Owner, prevExecutes, subOrConfirm bool) (*types.KeyValue, *types.ReceiptLog) {
	receiptLog := &types.ReceiptLog{}

	//组装receiptLog
	receiptLogTx := &mty.ReceiptMultiSigTx{}
	multiSigTxOwner := &mty.MultiSigTxOwner{MultiSigAddr: multiSigTx.MultiSigAddr, Txid: multiSigTx.Txid, ConfirmedOwner: owner}

	receiptLogTx.MultiSigTxOwner = multiSigTxOwner
	receiptLogTx.PrevExecuted = prevExecutes
	receiptLogTx.CurExecuted = multiSigTx.Executed
	receiptLogTx.SubmitOrConfirm = subOrConfirm
	if subOrConfirm {
		receiptLogTx.TxHash = multiSigTx.TxHash
		receiptLogTx.TxType = multiSigTx.TxType
	}

	receiptLog.Ty = mty.TyLogMultiSigTx
	receiptLog.Log = types.Encode(receiptLogTx)

	key, value := setMultiSigAccTxToDb(a.db, multiSigTx)
	keyValue := &types.KeyValue{Key: key, Value: value}
	return keyValue, receiptLog
}

//确认并执行转账交易：区分submitTx和confirmtx阶段。
func (a *action) executeTransferTx(multiSigAcc *mty.MultiSig, newMultiSigTx *mty.MultiSigTx, transfer *mty.MultiSigExecTransferFrom, confOwner *mty.Owner, subOrConfirm bool) (*types.Receipt, error) {

	//获取对应资产的每日限额信息
	var findindex int
	curDailyLimit := &mty.DailyLimit{Symbol: transfer.Symbol, Execer: transfer.Execname, DailyLimit: 0, SpentToday: 0, LastDay: 0}
	for Index, dailyLimit := range multiSigAcc.DailyLimits {
		if dailyLimit.Symbol == transfer.Symbol && dailyLimit.Execer == transfer.Execname {
			curDailyLimit.DailyLimit = dailyLimit.DailyLimit
			curDailyLimit.SpentToday = dailyLimit.SpentToday
			curDailyLimit.LastDay = dailyLimit.LastDay
			findindex = Index
			break
		}
	}
	//每日限额为0不允许转账
	if curDailyLimit.DailyLimit == 0 {
		return nil, mty.ErrDailyLimitIsZero
	}

	//确认此交易额度是否在每日限额之内，或者权重已达到要求
	amount := transfer.Amount
	confirmed := isConfirmed(multiSigAcc.RequiredWeight, newMultiSigTx)
	underLimit, newlastday := isUnderLimit(a.blocktime, uint64(amount), curDailyLimit)

	//新的一天更新lastday和spenttoday的值
	if newlastday != 0 {
		curDailyLimit.LastDay = newlastday
		curDailyLimit.SpentToday = 0
	}

	prevExecuted := newMultiSigTx.Executed

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	//权重满足或者小于每日限额，允许执行此交易，如果转账交易执行失败，不应该直接返回，需要继续更新多重签名账户和tx列表的状态信息
	if confirmed || underLimit {

		//执行此交易，从多重签名账户转币到指定账户，在multiSig合约中转账
		symbol := getRealSymbol(transfer.Symbol)
		execerAccDB, err := account.NewAccountDB(transfer.Execname, symbol, a.db)
		if err != nil {
			multisiglog.Error("executeTransaction:NewAccountDB", "From", transfer.From, "To", transfer.To,
				"execaddr", a.execaddr, "amount", amount, "Execer", transfer.Execname, "Symbol", transfer.Symbol, "error", err)
			return nil, err
		}
		receiptFromMultiSigAcc, err := execerAccDB.ExecTransferFrozen(transfer.From, transfer.To, a.execaddr, amount)
		if err != nil {
			multisiglog.Error("executeTransaction:ExecTransferFrozen", "From", transfer.From, "To", transfer.To,
				"execaddr", a.execaddr, "amount", amount, "Execer", transfer.Execname, "Symbol", transfer.Symbol, "error", err)
			return nil, err
		}
		logs = append(logs, receiptFromMultiSigAcc.Logs...)
		kv = append(kv, receiptFromMultiSigAcc.KV...)

		//标识此交易已经被执行
		newMultiSigTx.Executed = true

		//增加今日已用金额, 只有在提交交易时才会使用每日限额的额度
		if !confirmed && subOrConfirm {
			curDailyLimit.SpentToday += uint64(amount)
		}
	}

	//更新multiSigAcc状态:txcount有增加在submit阶段
	if subOrConfirm {
		keyvalue, receiptlog, err := a.receiptTxCountUpdate(multiSigAcc.MultiSigAddr)
		if err != nil {
			multisiglog.Error("executeTransaction:receiptTxCountUpdate", "error", err)
		}
		kv = append(kv, keyvalue)
		logs = append(logs, receiptlog)
	}

	//更新multiSigAcc状态:对应资产的每日限额信息可能有更新
	keyvalue, receiptlog, err := a.receiptDailyLimitUpdate(multiSigAcc.MultiSigAddr, findindex, curDailyLimit)
	if err != nil {
		multisiglog.Error("executeTransaction:receiptDailyLimitUpdate", "error", err)
	}
	//更新newMultiSigTx的状态：MultiSigTx增加一个确认owner，交易的执行状态可能有更新
	keyvaluetx, receiptlogtx := a.receiptMultiSigTx(newMultiSigTx, confOwner, prevExecuted, subOrConfirm)

	logs = append(logs, receiptlog)
	logs = append(logs, receiptlogtx)
	kv = append(kv, keyvalue)
	kv = append(kv, keyvaluetx)

	//test
	multisiglog.Error("executeTransferTx", "multiSigAcc", multiSigAcc, "newMultiSigTx", newMultiSigTx)

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: logs,
	}, nil
}

//确认并执行操作账户属性的交易：区分submitTx和confirmtx阶段。
func (a *action) executeAccOperateTx(multiSigAcc *mty.MultiSig, newMultiSigTx *mty.MultiSigTx, accountOperate *mty.MultiSigAccOperate, confOwner *mty.Owner, subOrConfirm bool) (*types.Receipt, error) {

	//确认权重是否已达到要求
	confirmed := isConfirmed(multiSigAcc.RequiredWeight, newMultiSigTx)
	prevExecuted := newMultiSigTx.Executed

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	var accAttrkv *types.KeyValue
	var accAttrReceiptLog *types.ReceiptLog
	var err error

	//权重满足允许执行此交易，需要继续更新多重签名账户和tx列表的状态信息
	if confirmed {
		//修改账户RequiredWeight的操作
		if accountOperate.OperateFlag {
			accAttrkv, accAttrReceiptLog, err = a.multiSigWeightModify(multiSigAcc.MultiSigAddr, accountOperate.NewRequiredWeight)
			if err != nil {
				multisiglog.Error("executeAccOperateTx", "multiSigWeightModify", err)
				return nil, err
			}
		} else { //资产每日限额的修改
			accAttrkv, accAttrReceiptLog, err = a.multiSigDailyLimitOperate(multiSigAcc.MultiSigAddr, accountOperate.DailyLimit)
			if err != nil {
				multisiglog.Error("executeAccOperateTx", "multiSigDailyLimitOperate", err)
				return nil, err
			}
		}
		logs = append(logs, accAttrReceiptLog)
		kv = append(kv, accAttrkv)
		//标识此交易已经被执行
		newMultiSigTx.Executed = true
	}

	//更新multiSigAcc状态:txcount有增加在submit阶段
	if subOrConfirm {
		keyvalue, receiptlog, err := a.receiptTxCountUpdate(multiSigAcc.MultiSigAddr)
		if err != nil {
			multisiglog.Error("executeAccOperateTx:receiptTxCountUpdate", "error", err)
		}
		kv = append(kv, keyvalue)
		logs = append(logs, receiptlog)
	}
	//更新newMultiSigTx的状态：MultiSigTx增加一个确认owner，交易的执行状态可能有更新
	keyvaluetx, receiptlogtx := a.receiptMultiSigTx(newMultiSigTx, confOwner, prevExecuted, subOrConfirm)
	logs = append(logs, receiptlogtx)
	kv = append(kv, keyvaluetx)
	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: logs,
	}, nil
}

//确认并执行操作owner属性的交易：区分submitTx和confirmtx阶段。
func (a *action) executeOwnerOperateTx(multiSigAccount *mty.MultiSig, newMultiSigTx *mty.MultiSigTx, accountOperate *mty.MultiSigOwnerOperate, confOwner *mty.Owner, subOrConfirm bool) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	var multiSigkv *types.KeyValue
	var receiptLog *types.ReceiptLog
	var err error
	//确认权重是否已达到要求
	confirmed := isConfirmed(multiSigAccount.RequiredWeight, newMultiSigTx)
	prevExecuted := newMultiSigTx.Executed

	flag := accountOperate.OperateFlag

	//权重满足允许执行此交易，需要继续更新多重签名账户和tx列表的状态信息
	if confirmed {
		//add
		if mty.OwnerAdd == flag {
			multiSigkv, receiptLog, err = a.multiSigOwnerAdd(multiSigAccount.MultiSigAddr, accountOperate)
			if err != nil {
				multisiglog.Error("MultiSigAccountOwnerOperate", "multiSigOwnerAdd err", err)
				return nil, err
			}

		} else if mty.OwnerDel == flag {
			multiSigkv, receiptLog, err = a.multiSigOwnerDel(multiSigAccount.MultiSigAddr, accountOperate)
			if err != nil {
				multisiglog.Error("MultiSigAccountOwnerOperate", "multiSigOwnerAdd err", err)
				return nil, err
			}
		} else if mty.OwnerModify == flag { //modify owner
			multiSigkv, receiptLog, err = a.multiSigOwnerModify(multiSigAccount.MultiSigAddr, accountOperate)
			if err != nil {
				multisiglog.Error("MultiSigAccountOwnerOperate", "multiSigOwnerModify err", err)
				return nil, err
			}
		} else if mty.OwnerReplace == flag { //replace owner
			multiSigkv, receiptLog, err = a.multiSigOwnerReplace(multiSigAccount.MultiSigAddr, accountOperate)
			if err != nil {
				multisiglog.Error("MultiSigAccountOwnerOperate", "multiSigOwnerReplace err", err)
				return nil, err
			}
		} else {
			multisiglog.Error("MultiSigAccountOwnerOperate", "OperateFlag", flag)
			return nil, mty.ErrOperateType
		}
		logs = append(logs, receiptLog)
		kv = append(kv, multiSigkv)

		//标识此交易已经被执行
		newMultiSigTx.Executed = true
	}

	//更新multiSigAcc状态:txcount有增加在submit阶段
	if subOrConfirm {
		keyvalue, receiptlog, err := a.receiptTxCountUpdate(multiSigAccount.MultiSigAddr)
		if err != nil {
			multisiglog.Error("executeOwnerOperateTx:receiptTxCountUpdate", "error", err)
		}
		kv = append(kv, keyvalue)
		logs = append(logs, receiptlog)
	}
	//更新newMultiSigTx的状态：MultiSigTx增加一个确认owner，交易的执行状态可能有更新
	keyvaluetx, receiptlogtx := a.receiptMultiSigTx(newMultiSigTx, confOwner, prevExecuted, subOrConfirm)
	logs = append(logs, receiptlogtx)
	kv = append(kv, keyvaluetx)

	//test
	multisiglog.Error("executeOwnerOperateTx", "multiSigAccount", multiSigAccount, "newMultiSigTx", newMultiSigTx)

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   kv,
		Logs: logs,
	}, nil
}

//构造确认交易的receiptLog
func (a *action) confirmTransaction(multiSigTx *mty.MultiSigTx, multiSigTxOwner *mty.MultiSigTxOwner, ConfirmOrRevoke bool) (*types.Receipt, error) {
	receiptLog := &types.ReceiptLog{}

	receiptLogUnConfirmTx := &mty.ReceiptConfirmTx{MultiSigTxOwner: multiSigTxOwner, ConfirmeOrRevoke: ConfirmOrRevoke}
	if ConfirmOrRevoke {
		receiptLog.Ty = mty.TyLogMultiSigConfirmTx
	} else {
		receiptLog.Ty = mty.TyLogMultiSigConfirmTxRevoke
	}
	receiptLog.Log = types.Encode(receiptLogUnConfirmTx)

	//更新MultiSigAccTx
	key, value := setMultiSigAccTxToDb(a.db, multiSigTx)
	kv := &types.KeyValue{Key: key, Value: value}

	return &types.Receipt{
		Ty:   types.ExecOk,
		KV:   []*types.KeyValue{kv},
		Logs: []*types.ReceiptLog{receiptLog},
	}, nil
}

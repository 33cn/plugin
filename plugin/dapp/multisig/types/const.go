// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var multisiglog = log15.New("module", "execs.multisig")

// OwnerAdd : 交易操作类型
var (
	OwnerAdd     uint64 = 1
	OwnerDel     uint64 = 2
	OwnerModify  uint64 = 3
	OwnerReplace uint64 = 4
	//AccWeightOp 账户属性的操作
	AccWeightOp     = true
	AccDailyLimitOp = false
	//OwnerOperate 多重签名交易类型：转账，owner操作，account操作
	OwnerOperate    uint64 = 1
	AccountOperate  uint64 = 2
	TransferOperate uint64 = 3
	//IsSubmit ：
	IsSubmit  = true
	IsConfirm = false

	MultiSigX            = "multisig"
	OneDaySecond   int64 = 24 * 3600
	MinOwnersInit        = 2
	MinOwnersCount       = 1  //一个多重签名的账户最少要保留一个owner
	MaxOwnersCount       = 20 //一个多重签名的账户最多拥有20个owner

	Multisiglog = log15.New("module", MultiSigX)
)

// MultiSig 交易的actionid
const (
	ActionMultiSigAccCreate        = 10000
	ActionMultiSigOwnerOperate     = 10001
	ActionMultiSigAccOperate       = 10002
	ActionMultiSigConfirmTx        = 10003
	ActionMultiSigExecTransferTo   = 10004
	ActionMultiSigExecTransferFrom = 10005
)

//多重签名账户执行输出的logid
const (
	TyLogMultiSigAccCreate = 10000 //只输出多重签名的账户地址

	TyLogMultiSigOwnerAdd     = 10001 //输出add的owner：addr和weight
	TyLogMultiSigOwnerDel     = 10002 //输出del的owner：addr和weight
	TyLogMultiSigOwnerModify  = 10003 //输出modify的owner：preweight以及currentweight
	TyLogMultiSigOwnerReplace = 10004 //输出old的owner的信息：以及当前的owner信息：addr+weight

	TyLogMultiSigAccWeightModify     = 10005 //输出修改前后确认权重的值：preReqWeight和curReqWeight
	TyLogMultiSigAccDailyLimitAdd    = 10006 //输出add的DailyLimit：Symbol和DailyLimit
	TyLogMultiSigAccDailyLimitModify = 10007 //输出modify的DailyLimit：preDailyLimit以及currentDailyLimit

	TyLogMultiSigConfirmTx       = 10008 //对某笔未执行交易的确认
	TyLogMultiSigConfirmTxRevoke = 10009 //已经确认交易的撤销只针对还未执行的交易

	TyLogDailyLimitUpdate = 10010 //DailyLimit更新，DailyLimit在Submit和Confirm阶段都可能有变化
	TyLogMultiSigTx       = 10011 //在Submit提交交易阶段才会有更新
	TyLogTxCountUpdate    = 10012 //txcount只在在Submit阶段提交新的交易是才会增加计数

)

//AccAssetsResult 账户资产cli的显示，主要是amount需要转换成浮点型字符串
type AccAssetsResult struct {
	Execer   string `json:"execer,omitempty"`
	Symbol   string `json:"symbol,omitempty"`
	Currency int32  `json:"currency,omitempty"`
	Balance  string `json:"balance,omitempty"`
	Frozen   string `json:"frozen,omitempty"`
	Receiver string `json:"receiver,omitempty"`
	Addr     string `json:"addr,omitempty"`
}

//DailyLimitResult 每日限额信息的显示cli
type DailyLimitResult struct {
	Symbol     string `json:"symbol,omitempty"`
	Execer     string `json:"execer,omitempty"`
	DailyLimit string `json:"dailyLimit,omitempty"`
	SpentToday string `json:"spent,omitempty"`
	LastDay    string `json:"lastday,omitempty"`
}

//MultiSigResult 多重签名账户信息的显示cli
type MultiSigResult struct {
	CreateAddr     string              `json:"createAddr,omitempty"`
	MultiSigAddr   string              `json:"multiSigAddr,omitempty"`
	Owners         []*Owner            `json:"owners,omitempty"`
	DailyLimits    []*DailyLimitResult `json:"dailyLimits,omitempty"`
	TxCount        uint64              `json:"txCount,omitempty"`
	RequiredWeight uint64              `json:"requiredWeight,omitempty"`
}

//UnSpentAssetsResult 每日限额之内未花费额度的显示cli
type UnSpentAssetsResult struct {
	Symbol  string `json:"symbol,omitempty"`
	Execer  string `json:"execer,omitempty"`
	UnSpent string `json:"unspent,omitempty"`
}

//IsAssetsInvalid 资产的合法性检测，Symbol：必须全部大写，例如：BTY,coins.BTY。exec：必须在types.AllowUserExec中存在
func IsAssetsInvalid(exec, symbol string) error {

	//exec检测
	allowExeName := types.AllowUserExec
	nameLen := len(allowExeName)
	execValid := false
	for i := 0; i < nameLen; i++ {
		if exec == string(allowExeName[i]) {
			execValid = true
			break
		}
	}
	if !execValid {
		multisiglog.Error("IsAssetsInvalid", "exec", exec)
		return ErrInvalidExec
	}
	//Symbol不做检测
	return nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
)

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(TokenX))
	types.RegistorExecutor(TokenX, NewType())
	types.RegisterDappFork(TokenX, "Enable", 100899)
	types.RegisterDappFork(TokenX, "ForkTokenBlackList", 190000)
	types.RegisterDappFork(TokenX, "ForkBadTokenSymbol", 184000)
	types.RegisterDappFork(TokenX, "ForkTokenPrice", 560000)
}

// TokenType 执行器基类结构体
type TokenType struct {
	types.ExecTypeBase
}

// NewType 创建执行器类型
func NewType() *TokenType {
	c := &TokenType{}
	c.SetChild(c)
	return c
}

// GetPayload 获取token action
func (t *TokenType) GetPayload() types.Message {
	return &TokenAction{}
}

// GetTypeMap 根据action的name获取type
func (t *TokenType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Transfer":          ActionTransfer,
		"Genesis":           ActionGenesis,
		"Withdraw":          ActionWithdraw,
		"TokenPreCreate":    TokenActionPreCreate,
		"TokenFinishCreate": TokenActionFinishCreate,
		"TokenRevokeCreate": TokenActionRevokeCreate,
		"TransferToExec":    TokenActionTransferToExec,
	}
}

// GetLogMap 获取log的映射对应关系
func (t *TokenType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogTokenTransfer:        {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenTransfer"},
		TyLogTokenDeposit:         {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenDeposit"},
		TyLogTokenExecTransfer:    {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecTransfer"},
		TyLogTokenExecWithdraw:    {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecWithdraw"},
		TyLogTokenExecDeposit:     {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecDeposit"},
		TyLogTokenExecFrozen:      {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecFrozen"},
		TyLogTokenExecActive:      {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenExecActive"},
		TyLogTokenGenesisTransfer: {Ty: reflect.TypeOf(types.ReceiptAccountTransfer{}), Name: "LogTokenGenesisTransfer"},
		TyLogTokenGenesisDeposit:  {Ty: reflect.TypeOf(types.ReceiptExecAccountTransfer{}), Name: "LogTokenGenesisDeposit"},
		TyLogPreCreateToken:       {Ty: reflect.TypeOf(ReceiptToken{}), Name: "LogPreCreateToken"},
		TyLogFinishCreateToken:    {Ty: reflect.TypeOf(ReceiptToken{}), Name: "LogFinishCreateToken"},
		TyLogRevokeCreateToken:    {Ty: reflect.TypeOf(ReceiptToken{}), Name: "LogRevokeCreateToken"},
	}
}

// RPC_Default_Process rpc 默认处理
func (t *TokenType) RPC_Default_Process(action string, msg interface{}) (*types.Transaction, error) {
	var create *types.CreateTx
	if _, ok := msg.(*types.CreateTx); !ok {
		return nil, types.ErrInvalidParam
	}
	create = msg.(*types.CreateTx)
	if !create.IsToken {
		return nil, types.ErrNotSupport
	}
	tx, err := t.AssertCreate(create)
	if err != nil {
		return nil, err
	}
	//to地址的问题,如果是主链交易，to地址就是直接是设置to
	if !types.IsPara() {
		tx.To = create.To
	}
	return tx, err
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"reflect"

	log "github.com/33cn/chain33/common/log/log15"

	"github.com/33cn/chain33/types"
)

var tokenlog = log.New("module", "execs.token.types")

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(TokenX))
	types.RegFork(TokenX, InitFork)
	types.RegExec(TokenX, InitExecutor)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(TokenX, "Enable", 0)
	cfg.RegisterDappFork(TokenX, ForkTokenBlackListX, 0)
	cfg.RegisterDappFork(TokenX, ForkBadTokenSymbolX, 0)
	cfg.RegisterDappFork(TokenX, ForkTokenPriceX, 0)
	cfg.RegisterDappFork(TokenX, ForkTokenSymbolWithNumberX, 0)
	cfg.RegisterDappFork(TokenX, ForkTokenCheckX, 0)
	cfg.RegisterDappFork(TokenX, ForkTokenEvm, 0)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(TokenX, NewType(cfg))
}

// TokenType 执行器基类结构体
type TokenType struct {
	types.ExecTypeBase
}

// NewType 创建执行器类型
func NewType(cfg *types.Chain33Config) *TokenType {
	c := &TokenType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetName 获取执行器名称
func (t *TokenType) GetName() string {
	return TokenX
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
		"TokenMint":         TokenActionMint,
		"TokenBurn":         TokenActionBurn,
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
		TyLogTokenMint:            {Ty: reflect.TypeOf(ReceiptTokenAmount{}), Name: "LogMintToken"},
		TyLogTokenBurn:            {Ty: reflect.TypeOf(ReceiptTokenAmount{}), Name: "LogBurnToken"},
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
	cfg := t.GetConfig()
	if !cfg.IsPara() {
		tx.To = create.To
	}
	return tx, err
}

// CreateTx token 创建合约
func (t *TokenType) CreateTx(action string, msg json.RawMessage) (*types.Transaction, error) {
	tx, err := t.ExecTypeBase.CreateTx(action, msg)
	if err != nil {
		tokenlog.Error("token CreateTx failed", "err", err, "action", action, "msg", string(msg))
		return nil, err
	}
	cfg := t.GetConfig()
	if !cfg.IsPara() {
		var transfer TokenAction
		err = types.Decode(tx.Payload, &transfer)
		if err != nil {
			tokenlog.Error("token CreateTx failed", "decode payload err", err, "action", action, "msg", string(msg))
			return nil, err
		}
		if action == "Transfer" {
			tx.To = transfer.GetTransfer().To
		} else if action == "Withdraw" {
			tx.To = transfer.GetWithdraw().To
		} else if action == "TransferToExec" {
			tx.To = transfer.GetTransferToExec().To
		}
	}
	return tx, nil
}

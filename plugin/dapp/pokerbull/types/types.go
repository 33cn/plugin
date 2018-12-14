// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"encoding/json"
	"errors"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
)

var (
	llog = log.New("module", "exectype."+PokerBullX)
)

func init() {
	// init executor type
	types.RegistorExecutor(PokerBullX, NewType())
	types.AllowUserExec = append(types.AllowUserExec, ExecerPokerBull)
	types.RegisterDappFork(PokerBullX, "Enable", 0)
}

// PokerBullType 斗牛执行器类型
type PokerBullType struct {
	types.ExecTypeBase
}

// NewType 创建pokerbull执行器类型
func NewType() *PokerBullType {
	c := &PokerBullType{}
	c.SetChild(c)
	return c
}

// GetName 获取执行器名称
func (t *PokerBullType) GetName() string {
	return PokerBullX
}

// GetPayload 获取payload
func (t *PokerBullType) GetPayload() types.Message {
	return &PBGameAction{}
}

// GetTypeMap 获取类型map
func (t *PokerBullType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Start":    PBGameActionStart,
		"Continue": PBGameActionContinue,
		"Quit":     PBGameActionQuit,
		"Query":    PBGameActionQuery,
	}
}

// GetLogMap 获取日志map
func (t *PokerBullType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogPBGameStart:    {Ty: reflect.TypeOf(ReceiptPBGame{}), Name: "TyLogPBGameStart"},
		TyLogPBGameContinue: {Ty: reflect.TypeOf(ReceiptPBGame{}), Name: "TyLogPBGameContinue"},
		TyLogPBGameQuit:     {Ty: reflect.TypeOf(ReceiptPBGame{}), Name: "TyLogPBGameQuit"},
		TyLogPBGameQuery:    {Ty: reflect.TypeOf(ReceiptPBGame{}), Name: "TyLogPBGameQuery"},
	}
}

// CreateTx method
func (t *PokerBullType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("pokerbull.CreateTx", "action", action)

	if action == CreateStartTx {
		var param PBGameStart
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawPBStartTx(&param)
	} else if action == CreateContinueTx {
		var param PBGameContinue
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawPBContinueTx(&param)
	} else if action == CreatequitTx {
		var param PBGameQuit
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawPBQuitTx(&param)
	}

	return nil, types.ErrNotSupport
}

// CreateRawPBStartTx method
func CreateRawPBStartTx(head *PBGameStart) (*types.Transaction, error) {
	if head.PlayerNum > MaxPlayerNum {
		return nil, errors.New("Player number should be maximum 5")
	}

	val := &PBGameAction{
		Ty:    PBGameActionStart,
		Value: &PBGameAction_Start{Start: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawPBContinueTx method
func CreateRawPBContinueTx(head *PBGameContinue) (*types.Transaction, error) {
	val := &PBGameAction{
		Ty:    PBGameActionContinue,
		Value: &PBGameAction_Continue{Continue: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawPBQuitTx method
func CreateRawPBQuitTx(head *PBGameQuit) (*types.Transaction, error) {
	val := &PBGameAction{
		Ty:    PBGameActionQuit,
		Value: &PBGameAction_Quit{Quit: head},
	}
	tx, err := types.CreateFormatTx(types.ExecName(PokerBullX), types.Encode(val))
	if err != nil {
		return nil, err
	}
	return tx, nil
}

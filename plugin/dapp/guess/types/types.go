// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"github.com/33cn/chain33/common/address"
	"reflect"

	"github.com/33cn/chain33/types"
	log "github.com/33cn/chain33/common/log/log15"
)

var (
	llog = log.New("module", "exectype." + GuessX)
)

func init() {
	// init executor type
	types.RegistorExecutor(GuessX, NewType())
	types.AllowUserExec = append(types.AllowUserExec, ExecerGuess)
	types.RegisterDappFork(GuessX, "Enable", 0)
}

// exec
type GuessType struct {
	types.ExecTypeBase
}

func NewType() *GuessType {
	c := &GuessType{}
	c.SetChild(c)
	return c
}

func (t *GuessType) GetPayload() types.Message {
	return &GuessGameAction{}
}

func (t *GuessType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Start": GuessGameActionStart,
		"Bet": GuessGameActionBet,
		"StopBet":GuessGameActionStopBet,
		"Abort": GuessGameActionAbort,
		"Publish": GuessGameActionPublish,
		"Query": GuessGameActionQuery,
	}
}

func (t *GuessType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogGuessGameStart:    {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGameStart"},
		TyLogGuessGameBet: {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGameBet"},
		TyLogGuessGameStopBet:     {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGameStopBet"},
		TyLogGuessGameAbort:    {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGameAbort"},
		TyLogGuessGamePublish:    {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGamePublish"},
		TyLogGuessGameTimeout:    {reflect.TypeOf(ReceiptGuessGame{}), "TyLogGuessGameTimeout"},
	}
}

// CreateTx method
func (t *GuessType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	llog.Debug("Guess.CreateTx", "action", action)

	if action == "GuessStart" {
		var param GuessGameStartTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx.GuessStart", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawGuessStartTx(&param)
	} else if action == "GuessBet" {
		var param GuessGameBetTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx.GuessBet", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawGuessBetTx(&param)
	} else if action == "GuessStopBet" {
		var param GuessGameStopBetTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx.GuessStopBet", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawGuessStopBetTx(&param)
	} else if action == "GuessPublish" {
		var param GuessGamePublishTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx.GuessPublish", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawGuessPublishTx(&param)
	} else if action == "GuessAbort" {
		var param GuessGameAbortTx
		err := json.Unmarshal(message, &param)
		if err != nil {
			llog.Error("CreateTx.GuessAbort", "Error", err)
			return nil, types.ErrInvalidParam
		}
		return CreateRawGuessAbortTx(&param)
	} else {
		return nil, types.ErrNotSupport
	}
}

// CreateRawLotteryCreateTx method
func CreateRawGuessStartTx(parm *GuessGameStartTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawGuessStartTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &GuessGameStart {
		Topic: parm.Topic,
		Options: parm.Options,
		Category: parm.Category,
		MaxBetHeight: parm.MaxBetHeight,
		MaxBetsOneTime: parm.MaxBets,
		MaxBetsNumber: parm.MaxBetsNumber,
		DevFeeFactor: parm.DevFeeFactor,
		DevFeeAddr: parm.DevFeeAddr,
		PlatFeeFactor: parm.PlatFeeFactor,
		PlatFeeAddr: parm.PlatFeeAddr,
		ExpireHeight: parm.ExpireHeight,
	}

	val := &GuessGameAction{
		Ty:    GuessGameActionStart,
		Value: &GuessGameAction_Start{Start: v},
	}
	llog.Info("CreateRawGuessStartTx", "Ty", val.Ty, "GuessGameActionStart", GuessGameActionStart)
	name := types.ExecName(GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGuessBetTx method
func CreateRawGuessBetTx(parm *GuessGameBetTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawGuessBet", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &GuessGameBet{
		GameId: parm.GameId,
		Option: parm.Option,
		BetsNum: parm.BetsNum,
	}
	val := &GuessGameAction{
		Ty:    GuessGameActionBet,
		Value: &GuessGameAction_Bet{Bet: v},
	}
	name := types.ExecName(GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGuessStopBetTx method
func CreateRawGuessStopBetTx(parm *GuessGameStopBetTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawGuessBet", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &GuessGameStopBet{
		GameId: parm.GameId,
	}
	val := &GuessGameAction{
		Ty:    GuessGameActionStopBet,
		Value: &GuessGameAction_StopBet{StopBet: v},
	}
	name := types.ExecName(GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGuessPublishTx method
func CreateRawGuessPublishTx(parm *GuessGamePublishTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawGuessPublish", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &GuessGamePublish{
		GameId: parm.GameId,
		Result: parm.Result,
	}

	val := &GuessGameAction{
		Ty:    GuessGameActionPublish,
		Value: &GuessGameAction_Publish{Publish: v},
	}
	name := types.ExecName(GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// CreateRawGuessAbortTx method
func CreateRawGuessAbortTx(parm *GuessGameAbortTx) (*types.Transaction, error) {
	if parm == nil {
		llog.Error("CreateRawGuessAbortTx", "parm", parm)
		return nil, types.ErrInvalidParam
	}

	v := &GuessGameAbort {
		GameId: parm.GameId,
	}

	val := &GuessGameAction{
		Ty:    GuessGameActionAbort,
		Value: &GuessGameAction_Abort{Abort: v},
	}
	name := types.ExecName(GuessX)
	tx := &types.Transaction{
		Execer:  []byte(types.ExecName(GuessX)),
		Payload: types.Encode(val),
		Fee:     parm.Fee,
		To:      address.ExecAddress(name),
	}

	tx, err := types.FormatTx(name, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}



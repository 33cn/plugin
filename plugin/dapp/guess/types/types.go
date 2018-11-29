// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"reflect"

	"github.com/33cn/chain33/types"
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




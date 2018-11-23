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
type PokerBullType struct {
	types.ExecTypeBase
}

func NewType() *PokerBullType {
	c := &PokerBullType{}
	c.SetChild(c)
	return c
}

func (t *PokerBullType) GetPayload() types.Message {
	return &PBGameAction{}
}

func (t *PokerBullType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Start":    PBGameActionStart,
		"Continue": PBGameActionContinue,
		"Quit":     PBGameActionQuit,
		"Query":    PBGameActionQuery,
	}
}

func (t *PokerBullType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogPBGameStart:    {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameStart"},
		TyLogPBGameContinue: {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameContinue"},
		TyLogPBGameQuit:     {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameQuit"},
		TyLogPBGameQuery:    {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameQuery"},
	}
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
		"Start":    PBGameActionStart,
		"Continue": PBGameActionContinue,
		"Quit":     PBGameActionQuit,
		"Query":    PBGameActionQuery,
	}
}

func (t *PokerBullType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogPBGameStart:    {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameStart"},
		TyLogPBGameContinue: {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameContinue"},
		TyLogPBGameQuit:     {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameQuit"},
		TyLogPBGameQuery:    {reflect.TypeOf(ReceiptPBGame{}), "TyLogPBGameQuery"},
	}
}



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
	types.AllowUserExec = append(types.AllowUserExec, ExecerDposVote)
	types.RegFork(DPosX, InitFork)
	types.RegExec(DPosX, InitExecutor)
}

//InitFork ...
func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(DPosX, "Enable", 0)
}

//InitExecutor ...
func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(DPosX, NewType(cfg))
}

// DPosType struct
type DPosType struct {
	types.ExecTypeBase
}

// NewType method
func NewType(cfg *types.Chain33Config) *DPosType {
	c := &DPosType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload method
func (t *DPosType) GetPayload() types.Message {
	return &DposVoteAction{}
}

// GetTypeMap method
func (t *DPosType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Regist":       DposVoteActionRegist,
		"CancelRegist": DposVoteActionCancelRegist,
		"ReRegist":     DposVoteActionReRegist,
		"Vote":         DposVoteActionVote,
		"CancelVote":   DposVoteActionCancelVote,
		"RegistVrfM":   DposVoteActionRegistVrfM,
		"RegistVrfRP":  DposVoteActionRegistVrfRP,
		"RecordCB":     DposVoteActionRecordCB,
		"RegistTopN":   DPosVoteActionRegistTopNCandidator,
	}
}

// GetLogMap method
func (t *DPosType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogCandicatorRegist:       {Ty: reflect.TypeOf(ReceiptCandicator{}), Name: "TyLogCandicatorRegist"},
		TyLogCandicatorVoted:        {Ty: reflect.TypeOf(ReceiptCandicator{}), Name: "TyLogCandicatorVoted"},
		TyLogCandicatorCancelVoted:  {Ty: reflect.TypeOf(ReceiptCandicator{}), Name: "TyLogCandicatorCancelVoted"},
		TyLogCandicatorCancelRegist: {Ty: reflect.TypeOf(ReceiptCandicator{}), Name: "TyLogCandicatorCancelRegist"},
		TyLogCandicatorReRegist:     {Ty: reflect.TypeOf(ReceiptCandicator{}), Name: "TyLogCandicatorReRegist"},
		TyLogVrfMRegist:             {Ty: reflect.TypeOf(ReceiptVrf{}), Name: "TyLogVrfMRegist"},
		TyLogVrfRPRegist:            {Ty: reflect.TypeOf(ReceiptVrf{}), Name: "TyLogVrfRPRegist"},
		TyLogCBInfoRecord:           {Ty: reflect.TypeOf(ReceiptCB{}), Name: "TyLogCBInfoRecord"},
		TyLogTopNCandidatorRegist:   {Ty: reflect.TypeOf(ReceiptTopN{}), Name: "TyLogTopNCandidatorRegist"},
	}
}

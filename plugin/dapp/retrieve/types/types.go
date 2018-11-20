// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
)

func init() {
	// init executor type
	types.RegistorExecutor(RetrieveX, NewType())
	types.RegisterDappFork(RetrieveX, "Enable", 0)
	types.RegisterDappFork(RetrieveX, "ForkRetrive", 180000)
}

//RetrieveType retrieve
type RetrieveType struct {
	types.ExecTypeBase
}

//NewType new retrieve
func NewType() *RetrieveType {
	c := &RetrieveType{}
	c.SetChild(c)
	return c
}

//GetRealToAddr 避免老的，没有To字段的交易分叉
func (r RetrieveType) GetRealToAddr(tx *types.Transaction) string {
	if len(tx.To) == 0 {
		return address.ExecAddress(string(tx.Execer))
	}
	return tx.To
}

//GetPayload 获取payload
func (r *RetrieveType) GetPayload() types.Message {
	return &RetrieveAction{}
}

//GetName 获取name
func (r *RetrieveType) GetName() string {
	return RetrieveX
}

//GetLogMap 获取log map
func (r *RetrieveType) GetLogMap() map[int64]*types.LogInfo {
	return nil
}

//GetTypeMap 获取type map
func (r *RetrieveType) GetTypeMap() map[string]int32 {
	return actionName
}

//ActionName action name
func (r RetrieveType) ActionName(tx *types.Transaction) string {
	var action RetrieveAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		return "unknown-err"
	}
	if action.Ty == RetrieveActionPrepare && action.GetPrepare() != nil {
		return "prepare"
	} else if action.Ty == RetrieveActionPerform && action.GetPerform() != nil {
		return "perform"
	} else if action.Ty == RetrieveActionBackup && action.GetBackup() != nil {
		return "backup"
	} else if action.Ty == RetrieveActionCancel && action.GetCancel() != nil {
		return "cancel"
	}
	return "unknown"
}

//Amount 数量
func (r RetrieveType) Amount(tx *types.Transaction) (int64, error) {
	return 0, nil
}

//CreateTx 构造交易
func (r RetrieveType) CreateTx(action string, message json.RawMessage) (*types.Transaction, error) {
	return nil, nil
}

//CoinsDepositLog ...
type CoinsDepositLog struct {
}

//Name name
func (l CoinsDepositLog) Name() string {
	return "LogDeposit"
}

//Decode 解码
func (l CoinsDepositLog) Decode(msg []byte) (interface{}, error) {
	var logTmp types.ReceiptAccountTransfer
	err := types.Decode(msg, &logTmp)
	if err != nil {
		return nil, err
	}
	return logTmp, err
}

//RetrieveGetInfo ...
type RetrieveGetInfo struct {
}

//JSONToProto json -> proto
func (t *RetrieveGetInfo) JSONToProto(message json.RawMessage) ([]byte, error) {
	var req ReqRetrieveInfo
	err := json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	return types.Encode(&req), nil
}

//ProtoToJSON proto -> json
func (t *RetrieveGetInfo) ProtoToJSON(reply *types.Message) (interface{}, error) {
	return reply, nil
}

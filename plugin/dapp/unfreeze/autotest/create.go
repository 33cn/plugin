// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"encoding/json"
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type fixAmountCase struct {
	types.BaseCase
	TotalAmount string `toml:"totalAmount"`
	From        string `toml:"from"`
	To          string `toml:"to"`
	Period      int    `toml:"period"`
}

type fixAmountPack struct {
	types.BaseCasePack
	info *unfreezeInfo
}

type unfreezeInfo struct {
	unFreezeID string
	period     int
}

// SendCommand send command
func (t *fixAmountCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(t, &fixAmountPack{}, packID)
}

// GetCheckHandlerMap defines get check handle for map
func (pack *fixAmountPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 2)
	funcMap["frozen"] = pack.checkFrozen
	funcMap["unfreeze"] = pack.checkUnfreeze

	return funcMap
}

// GetDependData defines get depend data function
func (pack *fixAmountPack) GetDependData() interface{} {

	return pack.info
}

func (pack *fixAmountPack) checkFrozen(txInfo types.CheckHandlerParamType) bool {

	if len(txInfo.Receipt.Logs) < 2 {
		pack.FLog.Error("checkFrozenLog", "err", "logNumLessThan 2")
		return false
	}

	var frozenLog map[string]interface{}

	err := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &frozenLog)

	if err != nil {
		pack.FLog.Error("checkFrozenLog", "unmarshalErr", err)
	}

	interCase := pack.TCase.(*fixAmountCase)
	amount, _ := strconv.ParseFloat(interCase.TotalAmount, 64)
	b := types.CheckFrozenDeltaWithAddr(frozenLog, interCase.From, amount)
	return b
}

func (pack *fixAmountPack) checkUnfreeze(txInfo types.CheckHandlerParamType) bool {

	var freezeLog map[string]interface{}

	err := json.Unmarshal(txInfo.Receipt.Logs[2].Log, &freezeLog)

	if err != nil {
		pack.FLog.Error("checkUnfreeze", "unmarshalErr", err)
	}
	interCase := pack.TCase.(*fixAmountCase)
	info := &unfreezeInfo{}
	info.period = interCase.Period
	info.unFreezeID = freezeLog["current"].(map[string]interface{})["unfreezeID"].(string)
	pack.info = info

	amount, _ := strconv.ParseFloat(interCase.TotalAmount, 64)
	total, _ := strconv.ParseFloat(freezeLog["current"].(map[string]interface{})["totalCount"].(string), 64)
	remain, _ := strconv.ParseFloat(freezeLog["current"].(map[string]interface{})["remaining"].(string), 64)

	amount *= 1e8

	if types.IsBalanceEqualFloat(amount, total) &&
		types.IsBalanceEqualFloat(amount, remain) &&
		freezeLog["current"].(map[string]interface{})["initiator"].(string) == interCase.From &&
		freezeLog["current"].(map[string]interface{})["beneficiary"].(string) == interCase.To {
		return true
	}

	return false
}

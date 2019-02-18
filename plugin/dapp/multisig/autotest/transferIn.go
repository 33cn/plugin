// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type transferInCase struct {
	types.BaseCase
	From   string `toml:"from"`
	Amount string `toml:"amount"`
	info   *multisigInfo
}

type transferInPack struct {
	types.BaseCasePack
}

// SendCommand defines send command
func (testCase *transferInCase) SendCommand(packID string) (types.PackFunc, error) {

	if testCase.info == nil || len(testCase.info.account) == 0 {
		return nil, fmt.Errorf("nil multiSign account addr")
	}
	return types.DefaultSend(testCase, &transferInPack{}, packID)
}

// SetDependData defines set depend data function
func (testCase *transferInCase) SetDependData(depData interface{}) {

	if info, ok := depData.(*multisigInfo); ok && info != nil {
		testCase.info = info
		testCase.Command = fmt.Sprintf("%s -t %s", testCase.Command, testCase.info.account)

	}
}

// GetCheckHandlerMap defines get check handle for map
func (pack *transferInPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 1)
	funcMap["balance"] = pack.checkBalance
	return funcMap
}

func (pack *transferInPack) checkBalance(txInfo types.CheckHandlerParamType) bool {

	var fromLog map[string]interface{}
	var toLog map[string]interface{}
	var frozenLog map[string]interface{}

	err1 := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &fromLog)
	err2 := json.Unmarshal(txInfo.Receipt.Logs[2].Log, &toLog)
	err3 := json.Unmarshal(txInfo.Receipt.Logs[3].Log, &frozenLog)

	if err1 != nil || err2 != nil || err3 != nil {
		pack.FLog.Error("checkMultiSignTransferIn", "id", pack.PackID, "unmarshalErr1", err1, "unmarshalErr2", err2, "unmarshalErr3", err3)
		return false
	}

	interCase := pack.TCase.(*transferInCase)
	amount, err := strconv.ParseFloat(interCase.Amount, 64)
	if err != nil {
		pack.FLog.Error("checkMultiSignTransferIn", "id", pack.PackID, "ParseFloat", err)
		return false
	}
	return types.CheckBalanceDeltaWithAddr(fromLog, interCase.From, -amount) &&
		types.CheckBalanceDeltaWithAddr(toLog, interCase.info.account, amount) &&
		types.CheckFrozenDeltaWithAddr(frozenLog, interCase.info.account, amount) &&
		types.CheckBalanceDeltaWithAddr(frozenLog, interCase.info.account, -amount)
}

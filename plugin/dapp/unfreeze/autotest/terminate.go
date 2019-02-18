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

type unfreezeTerminateCase struct {
	types.BaseCase
	Addr string `toml:"addr"`
	info *unfreezeInfo
}

type unfreezeTerminatePack struct {
	types.BaseCasePack
}

// SendCommand defines send command
func (testCase *unfreezeTerminateCase) SendCommand(packID string) (types.PackFunc, error) {

	if testCase.info == nil || len(testCase.info.unFreezeID) == 0 {
		return nil, fmt.Errorf("can't withdraw without unFreezeID")
	}
	testCase.Command = fmt.Sprintf("%s --id %s", testCase.Command, testCase.info.unFreezeID[len("mavl-unfreeze-"):])

	return types.DefaultSend(testCase, &unfreezeTerminatePack{}, packID)
}

// SetDependData defines set depend data function
func (testCase *unfreezeTerminateCase) SetDependData(depData interface{}) {

	if info, ok := depData.(*unfreezeInfo); ok && info != nil {
		testCase.info = info
	}
}

// GetCheckHandlerMap defines get check handle for map
func (pack *unfreezeTerminatePack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 1)
	funcMap["unfreeze"] = pack.checkUnfreeze
	return funcMap
}

func (pack *unfreezeTerminatePack) checkUnfreeze(txInfo types.CheckHandlerParamType) bool {

	var activeLog map[string]interface{}
	var freezeLog map[string]interface{}

	err1 := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &activeLog)
	err2 := json.Unmarshal(txInfo.Receipt.Logs[2].Log, &freezeLog)

	if err1 != nil || err2 != nil {
		pack.FLog.Error("checkUnfreeze", "unmarshalErr1", err1, "unmarshalErr2", err2)
	}

	interCase := pack.TCase.(*unfreezeTerminateCase)
	freezePrev, _ := strconv.ParseFloat(freezeLog["prev"].(map[string]interface{})["remaining"].(string), 64)
	freezeCurr, _ := strconv.ParseFloat(freezeLog["current"].(map[string]interface{})["remaining"].(string), 64)
	delta := (freezeCurr - freezePrev) / 1e8
	from := freezeLog["current"].(map[string]interface{})["initiator"].(string)

	return types.CheckFrozenDeltaWithAddr(activeLog, from, delta) && from == interCase.Addr
}

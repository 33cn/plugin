// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type unfreezeWithdrawCase struct {
	types.BaseCase
	Addr string `toml:"addr"`
	info *unfreezeInfo
}

type unfreezeWithdrawPack struct {
	types.BaseCasePack
}

// SendCommand defines send command
func (testCase *unfreezeWithdrawCase) SendCommand(packID string) (types.PackFunc, error) {

	if testCase.info == nil || len(testCase.info.unFreezeID) == 0 {
		return nil, fmt.Errorf("can't withdraw without unFreezeID")
	}
	time.Sleep(time.Second * time.Duration(testCase.info.period))
	return types.DefaultSend(testCase, &unfreezeWithdrawPack{}, packID)
}

// SetDependData defines set depend data function
func (testCase *unfreezeWithdrawCase) SetDependData(depData interface{}) {

	if info, ok := depData.(*unfreezeInfo); ok && info != nil {
		testCase.info = info
		testCase.Command = fmt.Sprintf("%s --id %s", testCase.Command, testCase.info.unFreezeID[len("mavl-unfreeze-"):])

	}
}

// GetCheckHandlerMap defines get check handle for map
func (pack *unfreezeWithdrawPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 1)
	funcMap["unfreeze"] = pack.checkUnfreeze
	return funcMap
}

func (pack *unfreezeWithdrawPack) checkUnfreeze(txInfo types.CheckHandlerParamType) bool {

	var fromLog map[string]interface{}
	var toLog map[string]interface{}
	var freezeLog map[string]interface{}

	err1 := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &fromLog)
	err2 := json.Unmarshal(txInfo.Receipt.Logs[2].Log, &toLog)
	err3 := json.Unmarshal(txInfo.Receipt.Logs[3].Log, &freezeLog)

	if err1 != nil || err2 != nil || err3 != nil {
		pack.FLog.Error("checkUnfreeze", "unmarshalErr1", err1, "unmarshalErr2", err2, "unmarshalErr3", err3)
	}

	interCase := pack.TCase.(*unfreezeWithdrawCase)
	freezePrev, _ := strconv.ParseFloat(freezeLog["prev"].(map[string]interface{})["remaining"].(string), 64)
	freezeCurr, _ := strconv.ParseFloat(freezeLog["current"].(map[string]interface{})["remaining"].(string), 64)
	delta := (freezeCurr - freezePrev) / 1e8
	from := freezeLog["current"].(map[string]interface{})["initiator"].(string)
	to := freezeLog["current"].(map[string]interface{})["beneficiary"].(string)

	return types.CheckFrozenDeltaWithAddr(fromLog, from, delta) && types.CheckBalanceDeltaWithAddr(toLog, to, -delta) && to == interCase.Addr
}

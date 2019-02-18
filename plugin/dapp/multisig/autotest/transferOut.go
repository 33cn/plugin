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

type transferOutCase struct {
	types.BaseCase
	To     string `toml:"to"`
	Amount string `toml:"amount"`
	info   *multisigInfo
}

type transferOutPack struct {
	types.BaseCasePack
	txID string
}

// SendCommand defines send command
func (testCase *transferOutCase) SendCommand(packID string) (types.PackFunc, error) {

	if testCase.info == nil || len(testCase.info.account) == 0 {
		return nil, fmt.Errorf("nil multiSign account addr")
	}
	return types.DefaultSend(testCase, &transferOutPack{}, packID)
}

// GetDependData defines get depend data function
func (pack *transferOutPack) GetDependData() interface{} {

	return pack.txID
}

// SetDependData defines set depend data function
func (testCase *transferOutCase) SetDependData(depData interface{}) {

	if info, ok := depData.(*multisigInfo); ok && info != nil {
		testCase.info = info
		testCase.Command = fmt.Sprintf("%s -f %s", testCase.Command, testCase.info.account)

	}
}

// GetCheckHandlerMap defines get check handle for map
func (pack *transferOutPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMap, 1)
	funcMap["balance"] = pack.checkBalance
	return funcMap
}

func (pack *transferOutPack) checkBalance(txInfo types.CheckHandlerParamType) bool {

	logLen := len(txInfo.Receipt.Logs)
	var txLog map[string]interface{}
	err := json.Unmarshal(txInfo.Receipt.Logs[logLen-1].Log, &txLog)
	if err != nil {
		pack.FLog.Error("checkMultiSignTransferOut", "id", pack.PackID, "Unmarshal", err)
		return false
	}
	pack.txID = txLog["multiSigTxOwner"].(map[string]interface{})["txid"].(string)
	if logLen < 6 {
		//need confirm
		return true
	}

	var fromLog map[string]interface{}
	var toLog map[string]interface{}

	err1 := json.Unmarshal(txInfo.Receipt.Logs[1].Log, &fromLog)
	err2 := json.Unmarshal(txInfo.Receipt.Logs[2].Log, &toLog)

	if err1 != nil || err2 != nil {
		pack.FLog.Error("checkMultiSignTransferOut", "id", pack.PackID, "unmarshalErr1", err1, "unmarshalErr2", err2)
		return false
	}

	interCase := pack.TCase.(*transferOutCase)
	amount, err3 := strconv.ParseFloat(interCase.Amount, 64)
	if err3 != nil {
		pack.FLog.Error("checkMultiSignTransferOut", "id", pack.PackID, "ParseFloat", err3)
		return false
	}
	return types.CheckFrozenDeltaWithAddr(fromLog, interCase.info.account, -amount) &&
		types.CheckBalanceDeltaWithAddr(toLog, interCase.To, amount)
}

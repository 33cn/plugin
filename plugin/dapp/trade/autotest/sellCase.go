// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

// SellCase defines sell case command
type SellCase struct {
	types.BaseCase
	From   string `toml:"from"`
	Amount string `toml:"amount"`
}

// SellPack defines sell pack command
type SellPack struct {
	types.BaseCasePack
	orderInfo *SellOrderInfo
}

// SellOrderInfo sell order information
type SellOrderInfo struct {
	sellID string
}

// SendCommand send command of sellcase
func (testCase *SellCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &SellPack{}, packID)
}

// GetCheckHandlerMap defines get check handle for map
func (pack *SellPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["frozen"] = pack.checkFrozen
	funcMap["balance"] = pack.checkBalance

	return funcMap
}

// GetDependData defines get depend data function
func (pack *SellPack) GetDependData() interface{} {

	return pack.orderInfo
}

func (pack *SellPack) checkBalance(txInfo map[string]interface{}) bool {

	/*fromAddr := txInfo["tx"].(map[string]interface{})["from"].(string)
	toAddr := txInfo["tx"].(map[string]interface{})["to"].(string)*/
	feeStr := txInfo["tx"].(map[string]interface{})["fee"].(string)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	interCase := pack.TCase.(*SellCase)
	logFee := logArr[0].(map[string]interface{})["log"].(map[string]interface{})
	logSend := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	fee, _ := strconv.ParseFloat(feeStr, 64)
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)

	pack.FLog.Info("SellBalanceDetails", "TestID", pack.PackID,
		"Fee", feeStr, "SellAmount", interCase.Amount,
		"SellerBalancePrev", logSend["prev"].(map[string]interface{})["balance"].(string),
		"SellerBalanceCurr", logSend["current"].(map[string]interface{})["balance"].(string))

	//save sell order info
	sellOrderInfo := logArr[2].(map[string]interface{})["log"].(map[string]interface{})["base"].(map[string]interface{})
	pack.orderInfo = &SellOrderInfo{}
	pack.orderInfo.sellID = sellOrderInfo["sellID"].(string)

	return types.CheckBalanceDeltaWithAddr(logFee, interCase.From, -fee) &&
		types.CheckBalanceDeltaWithAddr(logSend, interCase.From, -amount)

}

func (pack *SellPack) checkFrozen(txInfo map[string]interface{}) bool {

	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	interCase := pack.TCase.(*SellCase)
	logSend := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)

	pack.FLog.Info("SellFrozenDetails", "TestID", pack.PackID,
		"SellAmount", interCase.Amount,
		"SellerFrozenPrev", logSend["prev"].(map[string]interface{})["frozen"].(string),
		"SellerFrozenCurr", logSend["current"].(map[string]interface{})["frozen"].(string))

	return types.CheckFrozenDeltaWithAddr(logSend, interCase.From, amount)

}

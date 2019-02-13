// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

// BuyCase defines buycase command
type BuyCase struct {
	types.BaseCase
	From        string `toml:"from"`
	To          string `toml:"to"`
	TokenAmount string `toml:"tokenAmount"`
	BtyAmount   string `toml:"btyAmount"`
}

// BuyPack defines buypack command
type BuyPack struct {
	types.BaseCasePack
}

// DependBuyCase defines depend buycase command
type DependBuyCase struct {
	BuyCase
	SellID string `toml:"sellID,omitempty"`
}

// DependBuyPack defines depend buy pack command
type DependBuyPack struct {
	BuyPack
}

// SendCommand defines send command function of dependbuycase
func (testCase *DependBuyCase) SendCommand(packID string) (types.PackFunc, error) {

	if len(testCase.SellID) == 0 {
		return nil, errors.New("depend sell case failed, Can't buy without sell id")
	}
	sellID := testCase.SellID[len("mavl-trade-sell-"):]
	testCase.Command = fmt.Sprintf("%s -s %s", testCase.Command, sellID)

	return types.DefaultSend(&testCase.BuyCase, &BuyPack{}, packID)
}

// SetDependData defines set depend data function
func (testCase *DependBuyCase) SetDependData(depData interface{}) {

	if orderInfo, ok := depData.(*SellOrderInfo); ok && orderInfo != nil {

		testCase.SellID = orderInfo.sellID
	}
}

// GetCheckHandlerMap defines get check handler for map
func (pack *BuyPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["frozen"] = pack.checkFrozen
	funcMap["balance"] = pack.checkBalance

	return funcMap
}

func (pack *BuyPack) checkBalance(txInfo map[string]interface{}) bool {

	/*fromAddr := txInfo["tx"].(map[string]interface{})["from"].(string)
	toAddr := txInfo["tx"].(map[string]interface{})["to"].(string)*/
	feeStr := txInfo["tx"].(map[string]interface{})["fee"].(string)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	interCase := pack.TCase.(*BuyCase)

	logFee := logArr[0].(map[string]interface{})["log"].(map[string]interface{})
	logBuyBty := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	logSellBty := logArr[2].(map[string]interface{})["log"].(map[string]interface{})
	logBuyToken := logArr[4].(map[string]interface{})["log"].(map[string]interface{})

	fee, _ := strconv.ParseFloat(feeStr, 64)
	tokenAmount, _ := strconv.ParseFloat(interCase.TokenAmount, 64)
	btyAmount, _ := strconv.ParseFloat(interCase.BtyAmount, 64)

	pack.FLog.Info("BuyBalanceDetails", "ID", pack.PackID,
		"Fee", feeStr, "TokenAmount", interCase.TokenAmount, "BtyAmount", interCase.BtyAmount,
		"SellerBtyPrev", logSellBty["prev"].(map[string]interface{})["balance"].(string),
		"SellerBtyCurr", logSellBty["current"].(map[string]interface{})["balance"].(string),
		"BuyerBtyPrev", logBuyBty["prev"].(map[string]interface{})["balance"].(string),
		"BuyerBtyCurr", logBuyBty["current"].(map[string]interface{})["balance"].(string),
		"BuyerTokenPrev", logBuyToken["prev"].(map[string]interface{})["balance"].(string),
		"BuyerTokenCurr", logBuyToken["current"].(map[string]interface{})["balance"].(string))

	return types.CheckBalanceDeltaWithAddr(logFee, interCase.From, -fee) &&
		types.CheckBalanceDeltaWithAddr(logBuyBty, interCase.From, -btyAmount) &&
		types.CheckBalanceDeltaWithAddr(logSellBty, interCase.To, btyAmount) &&
		types.CheckBalanceDeltaWithAddr(logBuyToken, interCase.From, tokenAmount)

}

func (pack *BuyPack) checkFrozen(txInfo map[string]interface{}) bool {

	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	interCase := pack.TCase.(*BuyCase)
	logSellToken := logArr[3].(map[string]interface{})["log"].(map[string]interface{})
	tokenAmount, _ := strconv.ParseFloat(interCase.TokenAmount, 64)

	pack.FLog.Info("BuyFrozenDetails", "ID", pack.PackID,
		"BuyTokenAmount", interCase.TokenAmount,
		"SellerTokenPrev", logSellToken["prev"].(map[string]interface{})["frozen"].(string),
		"SellerTokenCurr", logSellToken["current"].(map[string]interface{})["frozen"].(string))

	return types.CheckFrozenDeltaWithAddr(logSellToken, interCase.To, -tokenAmount)

}

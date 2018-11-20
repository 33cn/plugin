// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

//PubToPrivCase pub2priv case
type PubToPrivCase struct {
	types.BaseCase
	From   string `toml:"from"`
	To     string `toml:"to"`
	Amount string `toml:"amount"`
}

// PubToPrivPack public to privacy package
type PubToPrivPack struct {
	types.BaseCasePack
}

// SendCommand send command
func (testCase *PubToPrivCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &PubToPrivPack{}, packID)
}

// GetCheckHandlerMap get check handler map
func (pack *PubToPrivPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["balance"] = pack.checkBalance
	funcMap["utxo"] = pack.checkUtxo
	return funcMap
}

func (pack *PubToPrivPack) checkBalance(txInfo map[string]interface{}) bool {

	interCase := pack.TCase.(*PubToPrivCase)
	feeStr := txInfo["tx"].(map[string]interface{})["fee"].(string)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	logFee := logArr[0].(map[string]interface{})["log"].(map[string]interface{})
	logSend := logArr[1].(map[string]interface{})["log"].(map[string]interface{})

	fee, _ := strconv.ParseFloat(feeStr, 64)
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)

	pack.FLog.Info("Pub2PrivateDetails", "TestID", pack.PackID,
		"Fee", feeStr, "Amount", interCase.Amount, "FromAddr", interCase.From,
		"FromPrev", logSend["prev"].(map[string]interface{})["balance"].(string),
		"FromCurr", logSend["current"].(map[string]interface{})["balance"].(string))

	return types.CheckBalanceDeltaWithAddr(logFee, interCase.From, -fee) &&
		types.CheckBalanceDeltaWithAddr(logSend, interCase.From, -amount)
}

func (pack *PubToPrivPack) checkUtxo(txInfo map[string]interface{}) bool {

	interCase := pack.TCase.(*PubToPrivCase)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	outputLog := logArr[2].(map[string]interface{})["log"].(map[string]interface{})
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)

	//get available utxo with addr
	availUtxo, err := types.CalcUtxoAvailAmount(interCase.To, pack.TxHash)
	totalOutput := types.CalcTxUtxoAmount(outputLog, "keyoutput")
	availCheck := types.IsBalanceEqualFloat(availUtxo, amount)

	pack.FLog.Info("Pub2PrivateUtxoDetail", "TestID", pack.PackID,
		"TransferAmount", interCase.Amount, "UtxoOutput", totalOutput,
		"ToAddr", interCase.To, "UtxoAvailable", availUtxo, "CalcAvailErr", err)

	return availCheck && types.IsBalanceEqualFloat(totalOutput, amount)

}

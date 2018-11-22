// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

//PrivToPubCase pub2priv case
type PrivToPubCase struct {
	types.BaseCase
	From   string `toml:"from"`
	To     string `toml:"to"`
	Amount string `toml:"amount"`
}

// PrivToPubPack privacy to public package
type PrivToPubPack struct {
	types.BaseCasePack
}

// SendCommand send command
func (testCase *PrivToPubCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &PrivToPubPack{}, packID)
}

// GetCheckHandlerMap get check handler map
func (pack *PrivToPubPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["balance"] = pack.checkBalance
	funcMap["utxo"] = pack.checkUtxo
	return funcMap
}

func (pack *PrivToPubPack) checkBalance(txInfo map[string]interface{}) bool {

	interCase := pack.TCase.(*PrivToPubCase)
	feeStr := txInfo["tx"].(map[string]interface{})["fee"].(string)
	from := txInfo["fromaddr"].(string) //privacy contract addr
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	logFee := logArr[0].(map[string]interface{})["log"].(map[string]interface{})
	logPub := logArr[1].(map[string]interface{})["log"].(map[string]interface{})

	fee, _ := strconv.ParseFloat(feeStr, 64)
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)

	pack.FLog.Info("Private2PubDetails", "TestID", pack.PackID,
		"Fee", feeStr, "Amount", interCase.Amount,
		"FromAddr", interCase.From, "ToAddr", interCase.To,
		"ToPrev", logPub["prev"].(map[string]interface{})["balance"].(string),
		"ToCurr", logPub["current"].(map[string]interface{})["balance"].(string))

	return types.CheckBalanceDeltaWithAddr(logFee, from, -fee) &&
		types.CheckBalanceDeltaWithAddr(logPub, interCase.To, amount)
}

func (pack *PrivToPubPack) checkUtxo(txInfo map[string]interface{}) bool {

	interCase := pack.TCase.(*PrivToPubCase)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	inputLog := logArr[2].(map[string]interface{})["log"].(map[string]interface{})
	outputLog := logArr[3].(map[string]interface{})["log"].(map[string]interface{})
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)
	fee, _ := strconv.ParseFloat(txInfo["tx"].(map[string]interface{})["fee"].(string), 64)

	utxoInput := types.CalcTxUtxoAmount(inputLog, "keyinput")
	utxoOutput := types.CalcTxUtxoAmount(outputLog, "keyoutput")
	//get available utxo with addr
	availUtxo, err1 := types.CalcUtxoAvailAmount(interCase.From, pack.TxHash)
	//get spend utxo with addr
	spendUtxo, err2 := types.CalcUtxoSpendAmount(interCase.From, pack.TxHash)

	utxoCheck := types.IsBalanceEqualFloat(availUtxo, utxoOutput) && types.IsBalanceEqualFloat(spendUtxo, utxoInput)

	pack.FLog.Info("Private2PubUtxoDetail", "TestID", pack.PackID, "Fee", fee,
		"TransferAmount", interCase.Amount, "UtxoInput", utxoInput, "UtxoOutput", utxoOutput,
		"FromAddr", interCase.From, "UtxoAvailable", availUtxo, "UtxoSpend", spendUtxo,
		"CalcAvailErr", err1, "CalcSpendErr", err2)

	return types.IsBalanceEqualFloat(amount, utxoInput-utxoOutput-fee) && utxoCheck
}

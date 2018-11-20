// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"strconv"

	"github.com/33cn/chain33/cmd/autotest/types"
)

//PrivToPrivCase pub2priv case
type PrivToPrivCase struct {
	types.BaseCase
	From   string `toml:"from"`
	To     string `toml:"to"`
	Amount string `toml:"amount"`
}

// PrivToPrivPack privacy to privacy package
type PrivToPrivPack struct {
	types.BaseCasePack
}

// SendCommand send command
func (testCase *PrivToPrivCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &PrivToPrivPack{}, packID)
}

// GetCheckHandlerMap get check handler map
func (pack *PrivToPrivPack) GetCheckHandlerMap() interface{} {

	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["utxo"] = pack.checkUtxo
	return funcMap
}

func (pack *PrivToPrivPack) checkUtxo(txInfo map[string]interface{}) bool {

	interCase := pack.TCase.(*PrivToPrivCase)
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	inputLog := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	outputLog := logArr[2].(map[string]interface{})["log"].(map[string]interface{})
	amount, _ := strconv.ParseFloat(interCase.Amount, 64)
	fee, _ := strconv.ParseFloat(txInfo["tx"].(map[string]interface{})["fee"].(string), 64)

	utxoInput := types.CalcTxUtxoAmount(inputLog, "keyinput")
	utxoOutput := types.CalcTxUtxoAmount(outputLog, "keyoutput")

	fromAvail, err1 := types.CalcUtxoAvailAmount(interCase.From, pack.TxHash)
	fromSpend, err2 := types.CalcUtxoSpendAmount(interCase.From, pack.TxHash)
	toAvail, err3 := types.CalcUtxoAvailAmount(interCase.To, pack.TxHash)

	utxoCheck := types.IsBalanceEqualFloat(fromAvail, utxoInput-amount-fee) &&
		types.IsBalanceEqualFloat(toAvail, amount) &&
		types.IsBalanceEqualFloat(fromSpend, utxoInput)

	pack.FLog.Info("Private2PrivateUtxoDetail", "TestID", pack.PackID,
		"FromAddr", interCase.From, "ToAddr", interCase.To, "Fee", fee,
		"TransferAmount", interCase.Amount, "UtxoInput", utxoInput, "UtxoOutput", utxoOutput,
		"FromAvailable", fromAvail, "FromSpend", fromSpend, "ToAvailable", toAvail,
		"CalcFromAvailErr", err1, "CalcFromSpendErr", err2, "CalcToAvailErr", err3)

	return types.IsBalanceEqualFloat(fee, utxoInput-utxoOutput) && utxoCheck
}

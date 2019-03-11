// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"github.com/33cn/chain33/cmd/autotest/types"

	"strconv"
)

// TokenPreCreateCase token precreatecase command
type TokenPreCreateCase struct {
	types.BaseCase
	//From string `toml:"from"`
	//Amount string `toml:"amount"`
}

// TokenPreCreatePack defines token precreate package command
type TokenPreCreatePack struct {
	types.BaseCasePack
}

// TokenFinishCreateCase token finish create case command
type TokenFinishCreateCase struct {
	types.BaseCase
	//From string `toml:"from"`
	//Amount string `toml:"amount"`
}

// TokenFinishCreatePack token finish create pack command
type TokenFinishCreatePack struct {
	types.BaseCasePack
}

// TokenRevokeCase token revoke case command
type TokenRevokeCase struct {
	types.BaseCase
}

// TokenRevokePack token revoke pack command
type TokenRevokePack struct {
	types.BaseCasePack
}

// SendCommand defines send command function of tokenprecreatecase
func (testCase *TokenPreCreateCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &TokenPreCreatePack{}, packID)
}

// SendCommand defines send command function of tokenrevokecase
func (testCase *TokenRevokeCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &TokenRevokePack{}, packID)
}

// SendCommand send command function of tokenfinishcreatecase
func (testCase *TokenFinishCreateCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &TokenFinishCreatePack{}, packID)
}

// TokenMintCase token mint case
type TokenMintCase struct {
	types.BaseCase

	Amount string `toml:"amount"`
}

// TokenMintPack token mint pack command
type TokenMintPack struct {
	types.BaseCasePack
}

// SendCommand send command function of tokenfinishcreatecase
func (testCase *TokenMintCase) SendCommand(packID string) (types.PackFunc, error) {

	return types.DefaultSend(testCase, &TokenMintPack{}, packID)
}

// GetCheckHandlerMap get check handle for map
func (pack *TokenMintPack) GetCheckHandlerMap() interface{} {
	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["balance"] = pack.checkBalance

	return funcMap
}

func (pack *TokenMintPack) checkBalance(txInfo map[string]interface{}) bool {
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	logTokenBurn := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	logAccBurn := logArr[2].(map[string]interface{})["log"].(map[string]interface{})

	interCase := pack.TCase.(*TokenMintCase)
	amount1, _ := strconv.ParseInt(interCase.Amount, 10, 64)
	amount := amount1 * 1e8

	pack.FLog.Info("MintDetails", "TestID", pack.PackID,
		"TokenPrev", logTokenBurn["prev"].(map[string]interface{})["total"].(string),
		"TokenCurr", logTokenBurn["current"].(map[string]interface{})["total"].(string),
		"AccPrev", logAccBurn["prev"].(map[string]interface{})["balance"].(string),
		"AccCurr", logAccBurn["current"].(map[string]interface{})["balance"].(string),
		"amount", amount1)

	totalCurrent := parseInt64(logTokenBurn["current"].(map[string]interface{})["total"])
	totalPrev := parseInt64(logTokenBurn["prev"].(map[string]interface{})["total"])

	accCurrent := parseInt64(logAccBurn["current"].(map[string]interface{})["balance"])
	accPrev := parseInt64(logAccBurn["prev"].(map[string]interface{})["balance"])

	return totalCurrent-amount == totalPrev && accCurrent-amount == accPrev
}

// TokenBurnCase token mint case
type TokenBurnCase struct {
	types.BaseCase

	Amount string `toml:"amount"`
}

// TokenBurnPack token mint pack command
type TokenBurnPack struct {
	types.BaseCasePack
}

// SendCommand send command function
func (testCase *TokenBurnCase) SendCommand(packID string) (types.PackFunc, error) {
	return types.DefaultSend(testCase, &TokenBurnPack{}, packID)
}

// GetCheckHandlerMap get check handle for map
func (pack *TokenBurnPack) GetCheckHandlerMap() interface{} {
	funcMap := make(types.CheckHandlerMapDiscard, 2)
	funcMap["balance"] = pack.checkBalance

	return funcMap
}

func (pack *TokenBurnPack) checkBalance(txInfo map[string]interface{}) bool {
	logArr := txInfo["receipt"].(map[string]interface{})["logs"].([]interface{})
	logTokenBurn := logArr[1].(map[string]interface{})["log"].(map[string]interface{})
	logAccBurn := logArr[2].(map[string]interface{})["log"].(map[string]interface{})

	interCase := pack.TCase.(*TokenBurnCase)
	amount1, _ := strconv.ParseInt(interCase.Amount, 10, 64)
	amount := amount1 * 1e8

	pack.FLog.Info("BurnDetails", "TestID", pack.PackID,
		"TokenPrev", logTokenBurn["prev"].(map[string]interface{})["total"].(string),
		"TokenCurr", logTokenBurn["current"].(map[string]interface{})["total"].(string),
		"AccPrev", logAccBurn["prev"].(map[string]interface{})["balance"].(string),
		"AccCurr", logAccBurn["current"].(map[string]interface{})["balance"].(string),
		"amount", amount1)

	totalCurrent := parseInt64(logTokenBurn["current"].(map[string]interface{})["total"])
	totalPrev := parseInt64(logTokenBurn["prev"].(map[string]interface{})["total"])

	accCurrent := parseInt64(logAccBurn["current"].(map[string]interface{})["balance"])
	accPrev := parseInt64(logAccBurn["prev"].(map[string]interface{})["balance"])

	return totalCurrent+amount == totalPrev && accCurrent+amount == accPrev
}

func parseInt64(s interface{}) int64 {
	i, _ := strconv.ParseInt(s.(string), 10, 64)
	return i
}

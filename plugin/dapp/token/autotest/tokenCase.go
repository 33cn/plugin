// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.


package autotest

import (
	"github.com/33cn/chain33/cmd/autotest/types"
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

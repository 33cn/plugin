// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"github.com/33cn/chain33/cmd/autotest/types"
)

// JsCreateCase token createcase command
type JsCreateCase struct {
	types.BaseCase
}

// JsCreatePack defines  create package command
type JsCreatePack struct {
	types.BaseCasePack
}

// SendCommand defines send command function of tokenprecreatecase
func (testCase *JsCreateCase) SendCommand(packID string) (types.PackFunc, error) {
	return types.DefaultSend(testCase, &JsCreatePack{}, packID)
}

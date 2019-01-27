// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"fmt"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type confirmCase struct {
	types.BaseCase
	From   string `toml:"from"`
	Amount string `toml:"amount"`
	txID   string
	info   *multisigInfo
}

type confirmPack struct {
	types.BaseCasePack
}

// SendCommand defines send command
func (testCase *confirmCase) SendCommand(packID string) (types.PackFunc, error) {

	if testCase.txID == "" || testCase.info == nil || testCase.info.account == "" {
		return nil, fmt.Errorf("nil confirm tx id or multi sign account")
	}
	return types.DefaultSend(testCase, &confirmPack{}, packID)
}

// SetDependData defines set depend data function
func (testCase *confirmCase) SetDependData(depData interface{}) {

	if txid, ok := depData.(string); ok && txid != "" {
		testCase.txID = txid
		testCase.Command = fmt.Sprintf("%s -i %s", testCase.Command, testCase.txID)
	} else if info, ok := depData.(*multisigInfo); ok && info != nil {
		testCase.info = info
		testCase.Command = fmt.Sprintf("%s -a %s", testCase.Command, testCase.info.account)

	}
}

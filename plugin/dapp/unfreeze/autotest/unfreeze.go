// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"reflect"

	"github.com/33cn/chain33/cmd/autotest/types"
	"github.com/33cn/chain33/system/dapp/coins/autotest"
)

type unfreezeAutoTest struct {
	SimpleCaseArr        []types.SimpleCase      `toml:"SimpleCase,omitempty"`
	TransferCaseArr      []autotest.TransferCase `toml:"TransferCase,omitempty"`
	UnfreezeCreateFixArr []fixAmountCase         `toml:"UnfreezeCreateFix,omitempty"`
	UnfreezeWithdrawArr  []unfreezeWithdrawCase  `toml:"UnfreezeWithdraw,omitempty"`
	UnfreezeTerminateArr []unfreezeTerminateCase `toml:"UnfreezeTerminate,omitempty"`
}

func init() {

	types.RegisterAutoTest(unfreezeAutoTest{})

}

func (config unfreezeAutoTest) GetName() string {

	return "unfreeze"
}

func (config unfreezeAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

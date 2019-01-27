// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"reflect"

	"github.com/33cn/chain33/cmd/autotest/types"
	"github.com/33cn/chain33/system/dapp/coins/autotest"
)

type multisigAutoTest struct {
	SimpleCaseArr   []types.SimpleCase      `toml:"SimpleCase,omitempty"`
	TransferCaseArr []autotest.TransferCase `toml:"TransferCase,omitempty"`
	CreateCaseArr   []createMultisigCase    `toml:"MultiSigCreateCase"`
	TransferInArr   []transferInCase        `toml:"MultiSigTransInCase"`
	TransferOutArr  []transferOutCase       `toml:"MultiSigTransOutCase"`
	ConfirmArr      []confirmCase           `toml:"MultiSigConfirmCase"`
}

func init() {

	types.RegisterAutoTest(multisigAutoTest{})

}

func (config multisigAutoTest) GetName() string {

	return "multisig"
}

func (config multisigAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

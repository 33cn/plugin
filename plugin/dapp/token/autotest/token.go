// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package autotest

import (
	. "github.com/33cn/chain33/cmd/autotest/types"
	. "github.com/33cn/chain33/system/dapp/coins/autotest"
	"reflect"
)

type tokenAutoTest struct {
	SimpleCaseArr            []SimpleCase            `toml:"SimpleCase,omitempty"`
	TokenPreCreateCaseArr    []TokenPreCreateCase    `toml:"TokenPreCreateCase,omitempty"`
	TokenFinishCreateCaseArr []TokenFinishCreateCase `toml:"TokenFinishCreateCase,omitempty"`
	TransferCaseArr          []TransferCase          `toml:"TransferCase,omitempty"`
	WithdrawCaseArr          []WithdrawCase          `toml:"WithdrawCase,omitempty"`
	TokenRevokeCaseArr       []TokenRevokeCase       `toml:"TokenRevokeCase,omitempty"`
}

func init() {

	RegisterAutoTest(tokenAutoTest{})

}

func (config tokenAutoTest) GetName() string {

	return "token"
}

func (config tokenAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

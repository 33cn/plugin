// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"reflect"

	"github.com/33cn/chain33/cmd/autotest/types"
	coinautotest "github.com/33cn/chain33/system/dapp/coins/autotest"
	tokenautotest "github.com/33cn/plugin/plugin/dapp/token/autotest"
)

type tradeAutoTest struct {
	SimpleCaseArr            []types.SimpleCase                    `toml:"SimpleCase,omitempty"`
	TokenPreCreateCaseArr    []tokenautotest.TokenPreCreateCase    `toml:"TokenPreCreateCase,omitempty"`
	TokenFinishCreateCaseArr []tokenautotest.TokenFinishCreateCase `toml:"TokenFinishCreateCase,omitempty"`
	TransferCaseArr          []coinautotest.TransferCase           `toml:"TransferCase,omitempty"`
	SellCaseArr              []SellCase                            `toml:"SellCase,omitempty"`
	DependBuyCaseArr         []DependBuyCase                       `toml:"DependBuyCase,omitempty"`
}

func init() {

	types.RegisterAutoTest(tradeAutoTest{})

}

func (config tradeAutoTest) GetName() string {

	return "trade"
}

func (config tradeAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

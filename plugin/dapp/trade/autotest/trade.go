// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package autotest

import (
	. "github.com/33cn/chain33/cmd/autotest/types"
	. "github.com/33cn/chain33/system/dapp/coins/autotest"
	. "github.com/33cn/plugin/plugin/dapp/token/autotest"
	"reflect"

)

type tradeAutoTest struct {
	SimpleCaseArr            []SimpleCase            `toml:"SimpleCase,omitempty"`
	TokenPreCreateCaseArr    []TokenPreCreateCase    `toml:"TokenPreCreateCase,omitempty"`
	TokenFinishCreateCaseArr []TokenFinishCreateCase `toml:"TokenFinishCreateCase,omitempty"`
	TransferCaseArr          []TransferCase          `toml:"TransferCase,omitempty"`
	SellCaseArr              []SellCase              `toml:"SellCase,omitempty"`
	DependBuyCaseArr         []DependBuyCase         `toml:"DependBuyCase,omitempty"`
}

func init() {

	RegisterAutoTest(tradeAutoTest{})

}

func (config tradeAutoTest) GetName() string {

	return "trade"
}

func (config tradeAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

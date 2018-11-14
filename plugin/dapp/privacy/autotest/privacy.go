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

type privacyAutoTest struct {
	SimpleCaseArr            []SimpleCase            `toml:"SimpleCase,omitempty"`
	TokenPreCreateCaseArr    []TokenPreCreateCase    `toml:"TokenPreCreateCase,omitempty"`
	TokenFinishCreateCaseArr []TokenFinishCreateCase `toml:"TokenFinishCreateCase,omitempty"`
	TransferCaseArr          []TransferCase          `toml:"TransferCase,omitempty"`
	PubToPrivCaseArr         []PubToPrivCase         `toml:"PubToPrivCase,omitempty"`
	PrivToPrivCaseArr        []PrivToPrivCase        `toml:"PrivToPrivCase,omitempty"`
	PrivToPubCaseArr         []PrivToPubCase         `toml:"PrivToPubCase,omitempty"`
	CreateUtxosCaseArr       []CreateUtxosCase       `toml:"CreateUtxosCase,omitempty"`
}

func init() {

	RegisterAutoTest(privacyAutoTest{})

}

func (config privacyAutoTest) GetName() string {

	return "privacy"
}

func (config privacyAutoTest) GetTestConfigType() reflect.Type {

	return reflect.TypeOf(config)
}

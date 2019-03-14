// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package autotest

import (
	"reflect"

	"github.com/33cn/chain33/cmd/autotest/types"
)

type jsAutoTest struct {
	SimpleCaseArr   []types.SimpleCase `toml:"SimpleCase,omitempty"`
	JSCreateCaseArr []JsCreateCase     `toml:"jsCreateCase,omitempty"`
}

func init() {
	types.RegisterAutoTest(jsAutoTest{})
}

func (config jsAutoTest) GetName() string {
	return "js"
}

func (config jsAutoTest) GetTestConfigType() reflect.Type {
	return reflect.TypeOf(config)
}

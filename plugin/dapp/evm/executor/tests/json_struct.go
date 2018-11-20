// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tests

// VMCase 一个测试用例
type VMCase struct {
	name string
	env  EnvJSON
	exec ExecJSON
	gas  int64
	logs string
	out  string
	err  string
	pre  map[string]AccountJSON
	post map[string]AccountJSON
}

// EnvJSON 上下文信息
type EnvJSON struct {
	currentCoinbase   string
	currentDifficulty int64
	currentGasLimit   int64
	currentNumber     int64
	currentTimestamp  int64
}

// ExecJSON 调用信息
type ExecJSON struct {
	address  string
	caller   string
	code     string
	data     string
	gas      int64
	gasPrice int64
	origin   string
	value    int64
}

// AccountJSON 账户信息
type AccountJSON struct {
	balance int64
	code    string
	nonce   int64
	storage map[string]string
}

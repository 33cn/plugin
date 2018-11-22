// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

// BlackwhiteCreateTx 创建游戏结构体
type BlackwhiteCreateTx struct {
	PlayAmount  int64  `json:"amount"`
	PlayerCount int32  `json:"playerCount"`
	Timeout     int64  `json:"timeout"`
	GameName    string `json:"gameName"`
	Fee         int64  `json:"fee"`
}

// BlackwhitePlayTx 参与游戏结构体
type BlackwhitePlayTx struct {
	GameID     string   `json:"gameID"`
	Amount     int64    `json:"amount"`
	HashValues [][]byte `json:"hashValues"`
	Fee        int64    `json:"fee"`
}

// BlackwhiteShowTx 出示密钥结构体
type BlackwhiteShowTx struct {
	GameID string `json:"gameID"`
	Secret string `json:"secret"`
	Fee    int64  `json:"fee"`
}

// BlackwhiteTimeoutDoneTx 游戏超时结构体
type BlackwhiteTimeoutDoneTx struct {
	GameID string `json:"GameID"`
	Fee    int64  `json:"fee"`
}

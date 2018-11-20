// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

// GamePreCreateTx pre create game,unused
type GamePreCreateTx struct {
	//Secret     string `json:"secret"`
	//下注必须时偶数，不能时级数
	Amount int64 `json:"amount"`
	//暂时只支持sha256加密
	HashType  string `json:"hashType"`
	HashValue []byte `json:"hashValue,omitempty"`
	Fee       int64  `json:"fee"`
}

// GamePreMatchTx pre match game,unused
type GamePreMatchTx struct {
	GameID string `json:"gameID"`
	Guess  int32  `json:"guess"`
	Fee    int64  `json:"fee"`
}

// GamePreCancelTx pre cancel tx,unused
type GamePreCancelTx struct {
	GameID string `json:"gameID"`
	Fee    int64  `json:"fee"`
}

// GamePreCloseTx pre close game, unused
type GamePreCloseTx struct {
	GameID string `json:"gameID"`
	Secret string `json:"secret"`
	Result int32  `json:"result"`
	Fee    int64  `json:"fee"`
}

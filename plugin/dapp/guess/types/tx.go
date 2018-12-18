// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

type GuessGameStartTx struct {
	Topic         string `json:"topic,omitempty"`
	Options       string `json:"options,omitempty"`
	Category      string `json:"category,omitempty"`
	MaxBetHeight  int64  `json:"maxHeight,omitempty"`
	MaxBets       int64  `json:"maxBets,omitempty"`
	MaxBetsNumber int64  `json:"maxBetsNumber,omitempty"`
	DevFeeFactor  int64  `json:"devFeeFactor,omitempty"`
	DevFeeAddr    string `json:"devFeeAddr,omitempty"`
	PlatFeeFactor int64  `json:"platFeeFactor,omitempty"`
	PlatFeeAddr   string `json:"platFeeAddr,omitempty"`
	ExpireHeight  int64  `json:"expireHeight,omitempty"`
	Fee           int64  `json:"fee,omitempty"`
}

type GuessGameBetTx struct {
	GameId  string `json:"gameId,omitempty"`
	Option  string `json:"option,omitempty"`
	BetsNum int64  `json:"betsNum,omitempty"`
	Fee     int64  `json:"fee,omitempty"`
}

type GuessGameStopBetTx struct {
	GameId string `json:"gameId,omitempty"`
	Fee    int64  `json:"fee,omitempty"`
}

type GuessGamePublishTx struct {
	GameId string `json:"gameId,omitempty"`
	Result string `json:"result,omitempty"`
	Fee    int64  `json:"fee,omitempty"`
}

type GuessGameAbortTx struct {
	GameId string `json:"gameId,omitempty"`
	Fee    int64  `json:"fee,omitempty"`
}

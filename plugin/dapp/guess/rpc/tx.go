// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

type GuessGameStart struct {
	Topic                string   `json:"topic,omitempty"`
	Options              string   `json:"options,omitempty"`
	Category             string   `json:"category,omitempty"`
	MaxTime              string   `json:"maxTime,omitempty"`
	MaxHeight            int64    `json:"maxHeight,omitempty"`
	Symbol               string   `json:"symbol,omitempty"`
	Exec                 string   `json:"exec,omitempty"`
	OneBet               uint32   `json:"oneBet,omitempty"`
	MaxBets              uint32   `json:"maxBets,omitempty"`
	MaxBetsNumber        uint32   `json:"maxBetsNumber,omitempty"`
	DevFeeFactor         int64    `json:"devFeeFactor,omitempty"`
	DevFeeAddr           string   `json:"devFeeAddr,omitempty"`
	PlatFeeFactor        int64    `json:"platFeeFactor,omitempty"`
	PlatFeeAddr          string   `json:"platFeeAddr,omitempty"`
	Expire               string   `json:"expire,omitempty"`
	ExpireHeight         int64    `json:"expireHeight,omitempty"`
	Fee                  int64    `json:"fee,omitempty"`
}

type GuessGameBet struct {
	GameId               string   `json:"gameId,omitempty"`
	Option               string   `json:"option,omitempty"`
	BetsNum              uint32   `json:"betsNum,omitempty"`
	Fee                  int64    `json:"fee,omitempty"`
}

type GuessGameStopBet struct {
	GameId               string   `json:"gameId,omitempty"`
	Fee                  int64    `json:"fee,omitempty"`
}

type GuessGamePublish struct {
	GameId               string   `json:"gameId,omitempty"`
	Result               string   `json:"result,omitempty"`
	Fee                  int64    `json:"fee,omitempty"`
}

type GuessGameAbort struct {
	GameId               string   `json:"gameId,omitempty"`
	Fee                  int64    `json:"fee,omitempty"`
}

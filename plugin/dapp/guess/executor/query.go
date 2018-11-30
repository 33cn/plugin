// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
)

func (g *Guess) Query_QueryGamesByIds(in *pkt.QueryGuessGameInfos) (types.Message, error) {
	return Infos(g.GetStateDB(), in)
}

func (g *Guess) Query_QueryGameById(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	game, err := readGame(g.GetStateDB(), in.GetGameId())
	if err != nil {
		return nil, err
	}
	return &pkt.ReplyGuessGameInfo{Game: game}, nil
}

func (g *Guess) Query_QueryGameByAddr(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	gameIds, err := getGameListByAddr(g.GetLocalDB(), in.Addr, in.Index)
	if err != nil {
		return nil, err
	}

	return gameIds, nil
}

func (g *Guess) Query_QueryGameByStatus(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	gameIds, err := getGameListByStatus(g.GetLocalDB(), in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return gameIds, nil
}

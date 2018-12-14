// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

// Query_QueryGameListByIDs 根据id列表查询游戏
func (g *PokerBull) Query_QueryGameListByIDs(in *pkt.QueryPBGameInfos) (types.Message, error) {
	return Infos(g.GetStateDB(), in)
}

// Query_QueryGameByID 根据id查询游戏
func (g *PokerBull) Query_QueryGameByID(in *pkt.QueryPBGameInfo) (types.Message, error) {
	game, err := readGame(g.GetStateDB(), in.GetGameId())
	if err != nil {
		return nil, err
	}
	return &pkt.ReplyPBGame{Game: game}, nil
}

// Query_QueryGameByAddr 根据地址查询游戏
func (g *PokerBull) Query_QueryGameByAddr(in *pkt.QueryPBGameInfo) (types.Message, error) {
	gameIds, err := getGameListByAddr(g.GetLocalDB(), in.Addr, in.Index)
	if err != nil {
		return nil, err
	}
	return gameIds, nil
}

// Query_QueryGameByStatus 根据状态查询游戏
func (g *PokerBull) Query_QueryGameByStatus(in *pkt.QueryPBGameInfo) (types.Message, error) {
	gameIds, err := getGameListByStatus(g.GetLocalDB(), in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return gameIds, nil
}

// Query_QueryGameByRound 查询某一回合游戏结果
func (g *PokerBull) Query_QueryGameByRound(in *pkt.QueryPBGameByRound) (types.Message, error) {
	game, err := readGame(g.GetStateDB(), in.GetGameId())
	if err != nil {
		return nil, err
	}

	if in.Round > game.Round {
		return nil, types.ErrInvalidParam
	}

	var roundPlayers []*pkt.PBPlayer
	for _, player := range game.Players {
		var isReady bool
		if in.Round == game.Round {
			isReady = player.Ready
		} else {
			isReady = false
		}
		roundPlayer := &pkt.PBPlayer{
			Address: player.Address,
			Ready:   isReady,
		}
		roundPlayers = append(roundPlayers, roundPlayer)
	}

	var result *pkt.PBResult
	if len(game.Results) < int(in.Round) {
		result = nil
	} else {
		result = game.Results[in.Round-1]
	}

	gameInfo := &pkt.ReplyPBGameByRound{
		GameId:    game.GameId,
		Status:    game.Status,
		Result:    result,
		IsWaiting: game.IsWaiting,
		Value:     game.Value,
		Players:   roundPlayers,
		Return:    (game.Value / types.Coin) * pkt.WinnerReturn,
	}

	return gameInfo, nil
}

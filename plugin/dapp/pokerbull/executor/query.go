// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
)

// Query_QueryGameListByIds 根据id列表查询游戏
func (g *PokerBull) Query_QueryGameListByIds(in *pkt.QueryPBGameInfos) (types.Message, error) {
	return Infos(g.GetStateDB(), in)
}

// Query_QueryGameById 根据id查询游戏
func (g *PokerBull) Query_QueryGameById(in *pkt.QueryPBGameInfo) (types.Message, error) {
	game, err := readGame(g.GetStateDB(), in.GetGameId())
	if err != nil {
		return nil, err
	}
	return &pkt.ReplyPBGame{game}, nil
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

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
	records, err := getGameListByAddr(g.GetLocalDB(), in.Addr, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (g *Guess) Query_QueryGameByStatus(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	records, err := getGameListByStatus(g.GetLocalDB(), in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (g *Guess) Query_QueryGameByAdminAddr(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	records, err := getGameListByAdminAddr(g.GetLocalDB(), in.AdminAddr, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (g *Guess) Query_QueryGameByAddrStatus(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	records, err := getGameListByAddrStatus(g.GetLocalDB(), in.Addr, in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (g *Guess) Query_QueryGameByAdminStatus(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	records, err := getGameListByAdminStatus(g.GetLocalDB(), in.AdminAddr, in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (g *Guess) Query_QueryGameByCategoryStatus(in *pkt.QueryGuessGameInfo) (types.Message, error) {
	records, err := getGameListByCategoryStatus(g.GetLocalDB(), in.Category, in.Status, in.Index)
	if err != nil {
		return nil, err
	}

	return records, nil
}

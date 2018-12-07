// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/fingerguessing/types"
)

func (g *Fingerguessing) Query_QueryGameListByIds(in *gt.QueryGameInfos) (types.Message, error) {
	return Infos(g.GetStateDB(), in)
}

func (g *Fingerguessing) Query_QueryGameListCount(in *gt.QueryGameListCount) (types.Message, error) {
	return QueryGameListCount(g.GetStateDB(), in)
}

func (g *Fingerguessing) Query_QueryGameListByStatusAndAddr(in *gt.QueryGameListByStatusAndAddr) (types.Message, error) {
	return List(g.GetLocalDB(), g.GetStateDB(), in)
}

func (g *Fingerguessing) Query_QueryGameById(in *gt.QueryGameInfo) (types.Message, error) {
	game, err := readGame(g.GetStateDB(), in.GetGameId())
	if err != nil {
		return nil, err
	}
	return &gt.ReplyGame{game}, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

//Query_QueryGamesByIDs method
func (g *Guess) Query_QueryGamesByIDs(in *gty.QueryGuessGameInfos) (types.Message, error) {
	return queryGameInfos(g.GetLocalDB(), in)
}

//Query_QueryGameByID method
func (g *Guess) Query_QueryGameByID(in *gty.QueryGuessGameInfo) (types.Message, error) {
	game, err := queryGameInfo(g.GetLocalDB(), []byte(in.GetGameID()))
	if err != nil {
		return nil, err
	}

	return &gty.ReplyGuessGameInfo{Game: game}, nil
}

//Query_QueryGamesByAddr method
func (g *Guess) Query_QueryGamesByAddr(in *gty.QueryGuessGameInfo) (types.Message, error) {
	gameTable := gty.NewGuessUserTable(g.GetLocalDB())
	query := gameTable.GetQuery(g.GetLocalDB())

	return queryUserTableData(query, "addr", []byte(in.Addr), []byte(in.PrimaryKey))
}

//Query_QueryGamesByStatus method
func (g *Guess) Query_QueryGamesByStatus(in *gty.QueryGuessGameInfo) (types.Message, error) {
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	query := gameTable.GetQuery(g.GetLocalDB())

	return queryGameTableData(query, "status", []byte(fmt.Sprintf("%2d", in.Status)), []byte(in.PrimaryKey))
}

//Query_QueryGamesByAdminAddr method
func (g *Guess) Query_QueryGamesByAdminAddr(in *gty.QueryGuessGameInfo) (types.Message, error) {
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	query := gameTable.GetQuery(g.GetLocalDB())
	prefix := []byte(in.AdminAddr)
	return queryGameTableData(query, "admin", prefix, []byte(in.PrimaryKey))
}

//Query_QueryGamesByAddrStatus method
func (g *Guess) Query_QueryGamesByAddrStatus(in *gty.QueryGuessGameInfo) (types.Message, error) {
	userTable := gty.NewGuessUserTable(g.GetLocalDB())
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	tableJoin, err := table.NewJoinTable(userTable, gameTable, []string{"addr#status"})
	if err != nil {
		return nil, err
	}

	prefix := table.JoinKey([]byte(in.Addr), []byte(fmt.Sprintf("%2d", in.Status)))

	return queryJoinTableData(tableJoin, "addr#status", prefix, []byte(in.PrimaryKey))
}

//Query_QueryGamesByAdminStatus method
func (g *Guess) Query_QueryGamesByAdminStatus(in *gty.QueryGuessGameInfo) (types.Message, error) {
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	query := gameTable.GetQuery(g.GetLocalDB())
	prefix := []byte(fmt.Sprintf("%s:%2d", in.AdminAddr, in.Status))

	return queryGameTableData(query, "admin_status", prefix, []byte(in.PrimaryKey))
}

//Query_QueryGamesByCategoryStatus method
func (g *Guess) Query_QueryGamesByCategoryStatus(in *gty.QueryGuessGameInfo) (types.Message, error) {
	gameTable := gty.NewGuessGameTable(g.GetLocalDB())
	query := gameTable.GetQuery(g.GetLocalDB())
	prefix := []byte(fmt.Sprintf("%s:%2d", in.Category, in.Status))

	return queryGameTableData(query, "category_status", prefix, []byte(in.PrimaryKey))
}

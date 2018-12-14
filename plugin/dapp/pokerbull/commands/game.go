// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"strconv"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
	"github.com/spf13/cobra"
)

// PokerBullCmd 斗牛游戏命令行
func PokerBullCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pokerbull",
		Short: "poker bull game management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		PokerBullStartRawTxCmd(),
		PokerBullContinueRawTxCmd(),
		PokerBullQuitRawTxCmd(),
		PokerBullQueryResultRawTxCmd(),
	)

	return cmd
}

// PokerBullStartRawTxCmd 生成开始交易命令行
func PokerBullStartRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a new round poker bull game",
		Run:   pokerbullStart,
	}
	addPokerbullStartFlags(cmd)
	return cmd
}

func addPokerbullStartFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("value", "a", 0, "value")
	cmd.MarkFlagRequired("value")

	cmd.Flags().Uint32P("playerCount", "p", 0, "player count")
	cmd.MarkFlagRequired("playerCount")
}

func pokerbullStart(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	value, _ := cmd.Flags().GetUint64("value")
	playerCount, _ := cmd.Flags().GetUint32("playerCount")

	params := &rpctypes.CreateTxIn{
		Execer: types.ExecName(pkt.PokerBullX),
		ActionName: pkt.CreateStartTx,
		Payload: []byte(fmt.Sprintf("{\"value\":%d,\"playerNum\":%d}", int64(value) * types.Coin, int32(playerCount))),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// PokerBullContinueRawTxCmd 生成继续游戏命令行
func PokerBullContinueRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Continue a new round poker bull game",
		Run:   pokerbullContinue,
	}
	addPokerbullContinueFlags(cmd)
	return cmd
}

func addPokerbullContinueFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.MarkFlagRequired("gameID")
}

func pokerbullContinue(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")

	params := &rpctypes.CreateTxIn{
		Execer: types.ExecName(pkt.PokerBullX),
		ActionName: pkt.CreateContinueTx,
		Payload: []byte(fmt.Sprintf("{\"gameId\":\"%s\"}", gameID)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// PokerBullQuitRawTxCmd 生成继续游戏命令行
func PokerBullQuitRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "Quit game",
		Run:   pokerbullQuit,
	}
	addPokerbullQuitFlags(cmd)
	return cmd
}

func addPokerbullQuitFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.MarkFlagRequired("gameID")
}

func pokerbullQuit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")

	params := &rpctypes.CreateTxIn{
		Execer: types.ExecName(pkt.PokerBullX),
		ActionName: pkt.CreatequitTx,
		Payload: []byte(fmt.Sprintf("{\"gameId\":\"%s\"}", gameID)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// PokerBullQueryResultRawTxCmd 查询命令行
func PokerBullQueryResultRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query result",
		Run:   pokerbullQuery,
	}
	addPokerbullQueryFlags(cmd)
	return cmd
}

func addPokerbullQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.Flags().StringP("address", "a", "", "address")
	cmd.Flags().StringP("index", "i", "", "index")
	cmd.Flags().StringP("status", "s", "", "status")
	cmd.Flags().StringP("gameIDs", "d", "", "gameIDs")
	cmd.Flags().StringP("round", "r", "", "round")
}

func pokerbullQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	status, _ := strconv.ParseInt(statusStr, 10, 32)
	indexstr, _ := cmd.Flags().GetString("index")
	index, _ := strconv.ParseInt(indexstr, 10, 64)
	gameIDs, _ := cmd.Flags().GetString("gameIDs")
	round, _ := cmd.Flags().GetString("round")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.PokerBullX
	req := &pkt.QueryPBGameInfo{
		GameId: gameID,
		Addr:   address,
		Status: int32(status),
		Index:  index,
	}
	params.Payload = types.MustPBToJSON(req)
	if gameID != "" {
		if round == "" {
			params.FuncName = pkt.FuncNameQueryGameByID
			var res pkt.ReplyPBGame
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else {
			params.FuncName = pkt.FuncNameQueryGameByRound
			roundInt, err := strconv.ParseInt(round, 10, 32)
			if err != nil {
				fmt.Println(err)
				return
			}
			req := &pkt.QueryPBGameByRound{
				GameId: gameID,
				Round:  int32(roundInt),
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.ReplyPBGameByRound
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		}
	} else if address != "" {
		params.FuncName = pkt.FuncNameQueryGameByAddr
		var res pkt.PBGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if statusStr != "" {
		params.FuncName = pkt.FuncNameQueryGameByStatus
		var res pkt.PBGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if gameIDs != "" {
		params.FuncName = pkt.FuncNameQueryGameListByIDs
		var gameIDsS []string
		gameIDsS = append(gameIDsS, gameIDs)
		gameIDsS = append(gameIDsS, gameIDs)
		req := &pkt.QueryPBGameInfos{GameIds: gameIDsS}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.ReplyPBGameList
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		fmt.Println("Error: requeres at least one of gameID, address or status")
		cmd.Help()
	}
}

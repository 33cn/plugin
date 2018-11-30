// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
	"github.com/spf13/cobra"
)

func GuessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "guess",
		Short: "guess game management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		GuessStartRawTxCmd(),
		GuessBetRawTxCmd(),
		GuessAbortRawTxCmd(),
		GuessQueryRawTxCmd(),
		GuessPublishRawTxCmd(),
	)

	return cmd
}

func GuessStartRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start a new guess game",
		Run:   guessStart,
	}
	addGuessStartFlags(cmd)
	return cmd
}

func addGuessStartFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("topic", "t", "", "topic")
	cmd.MarkFlagRequired("topic")

	cmd.Flags().StringP("options", "o", "", "options")
	cmd.MarkFlagRequired("options")

	cmd.Flags().StringP("maxTime", "mt", "", "max time to bet, after this bet is forbidden")

	cmd.Flags().Uint32P("maxHeight", "h", 0, "max height to bet, after this bet is forbidden")
	cmd.MarkFlagRequired("maxHeight")

	cmd.Flags().StringP("symbol", "s", "bty", "token symbol")
	cmd.Flags().StringP("exec", "e", "coins", "excutor name")

	cmd.Flags().Uint32P("oneBet", "b", 10, "one bet number, eg:10 bty / 10 token")
	//cmd.MarkFlagRequired("oneBet")

	cmd.Flags().Uint32P("maxBets", "m", 10000, "max bets one time")
	//cmd.MarkFlagRequired("maxBets")

	cmd.Flags().Uint32P("maxBetsNumber", "n", 100000, "max bets number")
	//cmd.MarkFlagRequired("maxBetsNumber")

	cmd.Flags().Float64P("fee", "f", 0, "fee")

	cmd.Flags().StringP("feeAddr", "a", "", "fee address")

	cmd.Flags().StringP("expire", "ex", "", "expire time of the game, after this any addr can abort it")

	cmd.Flags().Uint32P("expireHeight", "eh", 0, "expire height of the game, after this any addr can abort it")
}

func guessStart(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	topic, _ := cmd.Flags().GetString("topic")
	options, _ := cmd.Flags().GetString("options")
	maxTime, _ := cmd.Flags().GetString("maxTime")
	maxHeight, _ := cmd.Flags().GetUint32("maxHeight")
	symbol, _ := cmd.Flags().GetString("symbol")
	exec, _ := cmd.Flags().GetString("exec")
	oneBet, _ := cmd.Flags().GetUint32("oneBet")
	maxBets, _ := cmd.Flags().GetUint32("maxBets")
	maxBetsNumber, _ := cmd.Flags().GetUint32("maxBetsNumber")
	fee, _ := cmd.Flags().GetFloat64("fee")
	feeAddr, _ := cmd.Flags().GetString("feeAddr")
	expire, _ := cmd.Flags().GetString("expire")
	expireHeight, _ := cmd.Flags().GetUint32("expireHeight")

	feeInt64 := uint64(fee * 1e4)

	params := &pkt.GuessStartTxReq{
		Topic:         topic,
		Options:       options,
		MaxTime:       maxTime,
		MaxHeight:     maxHeight,
		Symbol:        symbol,
		Exec:          exec,
		OneBet:        oneBet,
		MaxBets:       maxBets,
		MaxBetsNumber: maxBetsNumber,
		Fee:           feeInt64,
		FeeAddr:       feeAddr,
		Expire:        expire,
		ExpireHeight:  expireHeight,
	}

	var res string
	ctx := jsonrpc.NewRpcCtx(rpcLaddr, "guess.GuessStartTx", params, &res)
	ctx.RunWithoutMarshal()
}

func GuessBetRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bet",
		Short: "bet for one option in a guess game",
		Run:   guessBet,
	}
	addGuessBetFlags(cmd)
	return cmd
}

func addGuessBetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameId", "g", "", "game ID")
	cmd.MarkFlagRequired("gameId")
	cmd.Flags().StringP("option", "o", "", "option")
	cmd.MarkFlagRequired("option")
	cmd.Flags().Uint32P("betsNumber", "b", 1, "bets number for one option in a guess game")
	cmd.MarkFlagRequired("betsNumber")
}

func guessBet(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameId, _ := cmd.Flags().GetString("gameId")
	option, _ := cmd.Flags().GetString("option")
	betsNumber, _ := cmd.Flags().GetUint32("betsNumber")

	params := &pkt.GuessBetTxReq{
		GameId: gameId,
		Option: option,
		Bets: betsNumber,
	}

	var res string
	ctx := jsonrpc.NewRpcCtx(rpcLaddr, "guess.GuessBetTx", params, &res)
	ctx.RunWithoutMarshal()
}

func GuessAbortRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort",
		Short: "abort a guess game",
		Run:   guessAbort,
	}
	addGuessAbortFlags(cmd)
	return cmd
}

func addGuessAbortFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameId", "g", "", "game Id")
	cmd.MarkFlagRequired("gameId")
}

func guessAbort(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameId, _ := cmd.Flags().GetString("gameId")

	params := &pkt.GuessAbortTxReq{
		GameId: gameId,
	}

	var res string
	ctx := jsonrpc.NewRpcCtx(rpcLaddr, "guess.GuessAbortTx", params, &res)
	ctx.RunWithoutMarshal()
}

func GuessPublishRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish",
		Short: "publish the result of a guess game",
		Run:   guessPublish,
	}
	addGuessPublishFlags(cmd)
	return cmd
}

func addGuessPublishFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameId", "g", "", "game Id of a guess game")
	cmd.MarkFlagRequired("gameId")

	cmd.Flags().StringP("result", "r", "", "result of a guess game")
	cmd.MarkFlagRequired("result")
}

func guessPublish(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameId, _ := cmd.Flags().GetString("gameId")
	result, _ := cmd.Flags().GetString("result")

	params := &pkt.GuessPublishTxReq{
		GameId: gameId,
		Result: result,
	}

	var res string
	ctx := jsonrpc.NewRpcCtx(rpcLaddr, "guess.GuessPublishTx", params, &res)
	ctx.RunWithoutMarshal()
}


func GuessQueryRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query result",
		Run:   guessQuery,
	}
	addGuessQueryFlags(cmd)
	return cmd
}

func addGuessQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.Flags().StringP("address", "a", "", "address")
	cmd.Flags().StringP("index", "i", "", "index")
	cmd.Flags().StringP("status", "s", "", "status")
	cmd.Flags().StringP("gameIDs", "d", "", "gameIDs")
}


func guessQuery(cmd *cobra.Command, args []string) {
	/*
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	status, _ := strconv.ParseInt(statusStr, 10, 32)
	indexstr, _ := cmd.Flags().GetString("index")
	index, _ := strconv.ParseInt(indexstr, 10, 64)
	gameIDs, _ := cmd.Flags().GetString("gameIDs")

	var params types.Query4Cli
	params.Execer = pkt.GuessX
	req := &pkt.QueryGuessGameInfo{
		GameId: gameID,
		Addr:   address,
		Status: int32(status),
		Index:  index,
	}
	params.Payload = req
	if gameID != "" {
		params.FuncName = pkt.FuncName_QueryGameById
		var res pkt.ReplyGuessGameInfo
		ctx := jsonrpc.NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if address != "" {
		params.FuncName = pkt.FuncName_QueryGameByAddr
		var res pkt.PBGameRecords
		ctx := jsonrpc.NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if statusStr != "" {
		params.FuncName = pkt.FuncName_QueryGameByStatus
		var res pkt.PBGameRecords
		ctx := jsonrpc.NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if gameIDs != "" {
		params.FuncName = pkt.FuncName_QueryGameListByIds
		var gameIDsS []string
		gameIDsS = append(gameIDsS, gameIDs)
		gameIDsS = append(gameIDsS, gameIDs)
		req := &pkt.QueryPBGameInfos{gameIDsS}
		params.Payload = req
		var res pkt.ReplyGuessGameInfos
		ctx := jsonrpc.NewRpcCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		fmt.Println("Error: requeres at least one of gameID, address or status")
		cmd.Help()
	}
	*/
}

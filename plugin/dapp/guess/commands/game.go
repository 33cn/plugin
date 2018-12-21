// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"strings"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
	"github.com/spf13/cobra"
)

//GuessCmd Guess合约命令行
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
		GuessStopBetRawTxCmd(),
	)

	return cmd
}

//GuessStartRawTxCmd 构造Guess合约的start原始交易（未签名）的命令行
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

	cmd.Flags().StringP("category", "c", "default", "options")
	cmd.Flags().Int64P("maxBetHeight", "m", 0, "max height to bet, after this bet is forbidden")
	cmd.Flags().Int64P("maxBetsOneTime", "s", 10000, "max bets one time")
	cmd.Flags().Int64P("maxBetsNumber", "n", 100000, "max bets number")
	cmd.Flags().Int64P("devFeeFactor", "d", 0, "dev fee factor, unit: 1/1000")
	cmd.Flags().StringP("devFeeAddr", "f", "", "dev address to receive share")
	cmd.Flags().Int64P("platFeeFactor", "p", 0, "plat fee factor, unit: 1/1000")
	cmd.Flags().StringP("platFeeAddr", "q", "", "plat address to receive share")
	cmd.Flags().Int64P("expireHeight", "e", 0, "expire height of the game, after this any addr can abort it")
}

func guessStart(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	topic, _ := cmd.Flags().GetString("topic")
	category, _ := cmd.Flags().GetString("category")
	options, _ := cmd.Flags().GetString("options")
	maxBetHeight, _ := cmd.Flags().GetInt64("maxBetHeight")
	maxBetsOneTime, _ := cmd.Flags().GetInt64("maxBetsOneTime")
	maxBetsNumber, _ := cmd.Flags().GetInt64("maxBetsNumber")
	devFeeFactor, _ := cmd.Flags().GetInt64("devFeeFactor")
	devFeeAddr, _ := cmd.Flags().GetString("devFeeAddr")
	platFeeFactor, _ := cmd.Flags().GetInt64("platFeeFactor")
	platFeeAddr, _ := cmd.Flags().GetString("platFeeAddr")
	expireHeight, _ := cmd.Flags().GetInt64("expireHeight")

	payload := fmt.Sprintf("{\"topic\":\"%s\", \"options\":\"%s\", \"category\":\"%s\", \"maxBetHeight\":%d, \"maxBetsOneTime\":%d,\"maxBetsNumber\":%d,\"devFeeFactor\":%d,\"platFeeFactor\":%d,\"expireHeight\":%d,\"devFeeAddr\":\"%s\",\"platFeeAddr\":\"%s\"}", topic, options, category, maxBetHeight, maxBetsOneTime, maxBetsNumber, devFeeFactor, platFeeFactor, expireHeight, devFeeAddr, platFeeAddr)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.GuessX),
		ActionName: pkt.CreateStartTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()

}

//GuessBetRawTxCmd 构造Guess合约的bet原始交易（未签名）的命令行
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
	cmd.Flags().Int64P("betsNumber", "b", 1, "bets number for one option in a guess game")
	cmd.MarkFlagRequired("betsNumber")
}

func guessBet(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameId")
	option, _ := cmd.Flags().GetString("option")
	betsNumber, _ := cmd.Flags().GetInt64("betsNumber")

	payload := fmt.Sprintf("{\"gameID\":\"%s\", \"option\":\"%s\", \"betsNum\":%d}", gameID, option, betsNumber)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.GuessX),
		ActionName: pkt.CreateBetTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//GuessStopBetRawTxCmd 构造Guess合约的停止下注(stopBet)原始交易（未签名）的命令行
func GuessStopBetRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop bet",
		Short: "stop bet for a guess game",
		Run:   guessStopBet,
	}
	addGuessStopBetFlags(cmd)
	return cmd
}

func addGuessStopBetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameId", "g", "", "game ID")
	cmd.MarkFlagRequired("gameId")
	cmd.Flags().Float64P("fee", "f", 0.01, "tx fee")
}

func guessStopBet(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameId")

	payload := fmt.Sprintf("{\"gameID\":\"%s\"}", gameID)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.GuessX),
		ActionName: pkt.CreateStopBetTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//GuessAbortRawTxCmd 构造Guess合约的撤销(Abort)原始交易（未签名）的命令行
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
	gameID, _ := cmd.Flags().GetString("gameId")

	payload := fmt.Sprintf("{\"gameID\":\"%s\"}", gameID)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.GuessX),
		ActionName: pkt.CreateAbortTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//GuessPublishRawTxCmd 构造Guess合约的发布结果(Publish)原始交易（未签名）的命令行
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
	gameID, _ := cmd.Flags().GetString("gameId")
	result, _ := cmd.Flags().GetString("result")

	payload := fmt.Sprintf("{\"gameID\":\"%s\",\"result\":\"%s\"}", gameID, result)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.GuessX),
		ActionName: pkt.CreatePublishTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//GuessQueryRawTxCmd 构造Guess合约的查询(Query)命令行
func GuessQueryRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query info",
		Run:   guessQuery,
	}
	addGuessQueryFlags(cmd)
	return cmd
}

func addGuessQueryFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("type", "t", 1, "query type, 1:Ids,2:Id,3:Addr,4:Status,5:AdminAddr,6:AddrStatus,7:AdminStatus,8:CategoryStatus")
	cmd.Flags().StringP("gameId", "g", "", "game Id")
	cmd.Flags().StringP("addr", "a", "", "address")
	cmd.Flags().StringP("adminAddr", "m", "", "admin address")
	cmd.Flags().Int64P("index", "i", 0, "index")
	cmd.Flags().Int32P("status", "s", 0, "status")
	cmd.Flags().StringP("gameIDs", "d", "", "gameIDs")
	cmd.Flags().StringP("category", "c", "default", "game category")
}

func guessQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetInt32("type")
	gameID, _ := cmd.Flags().GetString("gameId")
	addr, _ := cmd.Flags().GetString("addr")
	adminAddr, _ := cmd.Flags().GetString("adminAddr")
	status, _ := cmd.Flags().GetInt32("status")
	index, _ := cmd.Flags().GetInt64("index")
	gameIDs, _ := cmd.Flags().GetString("gameIDs")
	category, _ := cmd.Flags().GetString("category")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.GuessX

	//query type,
	//1:QueryGamesByIds,
	//2:QueryGameById,
	//3:QueryGameByAddr,
	//4:QueryGameByStatus,
	//5:QueryGameByAdminAddr,
	//6:QueryGameByAddrStatus,
	//7:QueryGameByAdminStatus,
	//8:QueryGameByCategoryStatus,
	switch ty {
	case 1:
		gameIds := strings.Split(gameIDs, ";")
		req := &pkt.QueryGuessGameInfos{
			GameIDs: gameIds,
		}
		params.FuncName = pkt.FuncNameQueryGamesByIds
		params.Payload = types.MustPBToJSON(req)
		var res pkt.ReplyGuessGameInfos
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 2:
		req := &pkt.QueryGuessGameInfo{
			GameID: gameID,
		}
		params.FuncName = pkt.FuncNameQueryGameById
		params.Payload = types.MustPBToJSON(req)
		var res pkt.ReplyGuessGameInfo
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 3:
		req := &pkt.QueryGuessGameInfo{
			Addr:  addr,
			Index: index,
		}
		params.FuncName = pkt.FuncNameQueryGameByAddr
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 4:
		req := &pkt.QueryGuessGameInfo{
			Status: status,
			Index:  index,
		}
		params.FuncName = pkt.FuncNameQueryGameByStatus
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 5:
		req := &pkt.QueryGuessGameInfo{
			AdminAddr: adminAddr,
			Index:     index,
		}
		params.FuncName = pkt.FuncNameQueryGameByAdminAddr
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 6:
		req := &pkt.QueryGuessGameInfo{
			Addr:   addr,
			Status: status,
			Index:  index,
		}
		params.FuncName = pkt.FuncNameQueryGameByAddrStatus
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 7:
		req := &pkt.QueryGuessGameInfo{
			AdminAddr: adminAddr,
			Status:    status,
			Index:     index,
		}
		params.FuncName = pkt.FuncNameQueryGameByAdminStatus
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case 8:
		req := &pkt.QueryGuessGameInfo{
			Category: category,
			Status:   status,
			Index:    index,
		}
		params.FuncName = pkt.FuncNameQueryGameByCategoryStatus
		params.Payload = types.MustPBToJSON(req)
		var res pkt.GuessGameRecords
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}
}

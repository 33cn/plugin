// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"strconv"
	"strings"

	"github.com/33cn/chain33/common"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
	"github.com/spf13/cobra"
)

// BlackwhiteCmd 黑白配游戏命令行
func BlackwhiteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blackwhite",
		Short: "blackwhite game management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		BlackwhiteCreateRawTxCmd(),
		BlackwhitePlayRawTxCmd(),
		BlackwhiteShowRawTxCmd(),
		BlackwhiteTimeoutDoneTxCmd(),
		ShowBlackwhiteInfoCmd(),
	)

	return cmd
}

// BlackwhiteCreateRawTxCmd 创建黑白配游戏交易命令
func BlackwhiteCreateRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new round blackwhite game",
		Run:   blackwhiteCreate,
	}
	addBlackwhiteCreateFlags(cmd)
	return cmd
}

func addBlackwhiteCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Uint32P("playerCount", "p", 0, "player count")
	cmd.MarkFlagRequired("playerCount")
	cmd.Flags().Int64P("timeout", "t", 0, "timeout(min),default:10min")
	cmd.Flags().StringP("gameName", "g", "", "game name")
	cmd.Flags().Float64P("fee", "f", 0, "coin transaction fee")
}

func blackwhiteCreate(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	amount, _ := cmd.Flags().GetUint64("amount")
	playerCount, _ := cmd.Flags().GetUint32("playerCount")
	timeout, _ := cmd.Flags().GetInt64("timeout")
	gameName, _ := cmd.Flags().GetString("gameName")
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee * 1e4)
	amountInt64 := int64(amount)

	if timeout == 0 {
		timeout = 10
	}
	timeout = 60 * timeout

	params := &gt.BlackwhiteCreateTxReq{
		PlayAmount:  amountInt64 * types.Coin,
		PlayerCount: int32(playerCount),
		Timeout:     timeout,
		GameName:    gameName,
		Fee:         feeInt64,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "blackwhite.BlackwhiteCreateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// BlackwhitePlayRawTxCmd 参与玩黑白配游戏
func BlackwhitePlayRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "play",
		Short: "play a blackwhite game",
		Run:   blackwhitePlay,
	}
	addBlackwhitePlayFlags(cmd)
	return cmd
}

func addBlackwhitePlayFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.MarkFlagRequired("gameID")

	cmd.Flags().Uint64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("isBlackStr", "i", "", "[0-1-1-1-1-1-0-0-1-1] (1:black,0:white)")
	cmd.MarkFlagRequired("isBlackStr")

	cmd.Flags().StringP("secret", "s", "", "secret key")
	cmd.MarkFlagRequired("secret")
	cmd.Flags().Float64P("fee", "f", 0, "coin transaction fee")

}

func blackwhitePlay(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")
	amount, _ := cmd.Flags().GetUint64("amount")
	isBlackStr, _ := cmd.Flags().GetString("isBlackStr")
	secret, _ := cmd.Flags().GetString("secret")
	fee, _ := cmd.Flags().GetFloat64("fee")

	blacks := strings.Split(isBlackStr, "-")

	var hashValues [][]byte
	for i, black := range blacks {
		if black == "1" {
			hashValues = append(hashValues, common.Sha256([]byte(strconv.Itoa(i)+secret+black)))
		} else {
			white := "0"
			hashValues = append(hashValues, common.Sha256([]byte(strconv.Itoa(i)+secret+white)))
		}
	}

	feeInt64 := int64(fee * 1e4)
	amountInt64 := int64(amount)
	params := &gt.BlackwhitePlayTxReq{
		GameID:     gameID,
		Amount:     amountInt64 * types.Coin,
		HashValues: hashValues,
		Fee:        feeInt64,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "blackwhite.BlackwhitePlayTx", params, &res)
	ctx.RunWithoutMarshal()
}

// BlackwhiteShowRawTxCmd 出示密钥
func BlackwhiteShowRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show secret key",
		Run:   blackwhiteShow,
	}
	addBlackwhiteShowFlags(cmd)
	return cmd
}

func addBlackwhiteShowFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.MarkFlagRequired("gameID")

	cmd.Flags().StringP("secret", "s", "", "secret key")
	cmd.MarkFlagRequired("secret")
	cmd.Flags().Float64P("fee", "f", 0, "coin transaction fee")

}

func blackwhiteShow(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")
	secret, _ := cmd.Flags().GetString("secret")
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee * 1e4)

	params := &gt.BlackwhiteShowTxReq{
		GameID: gameID,
		Secret: secret,
		Fee:    feeInt64,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "blackwhite.BlackwhiteShowTx", params, &res)
	ctx.RunWithoutMarshal()
}

// BlackwhiteTimeoutDoneTxCmd 触发游戏超时，由外部触发
func BlackwhiteTimeoutDoneTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timeoutDone",
		Short: "timeout done the game result",
		Run:   blackwhiteTimeoutDone,
	}
	addBlackwhiteTimeoutDonelags(cmd)
	return cmd
}

func addBlackwhiteTimeoutDonelags(cmd *cobra.Command) {
	cmd.Flags().StringP("gameID", "g", "", "game ID")
	cmd.MarkFlagRequired("gameID")
	cmd.Flags().Float64P("fee", "f", 0, "coin transaction fee")
}

func blackwhiteTimeoutDone(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	gameID, _ := cmd.Flags().GetString("gameID")
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee * 1e4)

	params := &gt.BlackwhiteTimeoutDoneTxReq{
		GameID: gameID,
		Fee:    feeInt64,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "blackwhite.BlackwhiteTimeoutDoneTx", params, &res)
	ctx.RunWithoutMarshal()
}

// ShowBlackwhiteInfoCmd 显示黑白配游戏查询信息
func ShowBlackwhiteInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showInfo",
		Short: "show black white round info",
		Run:   showBlackwhiteInfo,
	}
	addshowBlackwhiteInfoflags(cmd)
	return cmd
}

func addshowBlackwhiteInfoflags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("type", "t", 0, "type")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("gameID", "g", "", "game ID")

	cmd.Flags().Uint32P("status", "s", 0, "status")
	cmd.Flags().StringP("addr", "a", "", "addr")
	cmd.Flags().Int32P("count", "c", 0, "count")
	cmd.Flags().Int32P("direction", "d", 0, "direction")
	cmd.Flags().Int64P("index", "i", 0, "index")

	cmd.Flags().Uint32P("loopSeq", "l", 0, "loopSeq")
}

func showBlackwhiteInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	typ, _ := cmd.Flags().GetUint32("type")

	gameID, _ := cmd.Flags().GetString("gameID")

	status, _ := cmd.Flags().GetUint32("status")
	addr, _ := cmd.Flags().GetString("addr")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	index, _ := cmd.Flags().GetInt64("index")

	loopSeq, _ := cmd.Flags().GetUint32("loopSeq")

	var params rpctypes.Query4Jrpc

	var rep interface{}

	params.Execer = gt.BlackwhiteX
	if 0 == typ {
		req := gt.ReqBlackwhiteRoundInfo{
			GameID: gameID,
		}
		params.FuncName = gt.GetBlackwhiteRoundInfo
		params.Payload = types.MustPBToJSON(&req)
		rep = &gt.ReplyBlackwhiteRoundInfo{}
	} else if 1 == typ {
		req := gt.ReqBlackwhiteRoundList{
			Status:    int32(status),
			Address:   addr,
			Count:     count,
			Direction: direction,
			Index:     index,
		}
		params.FuncName = gt.GetBlackwhiteByStatusAndAddr
		params.Payload = types.MustPBToJSON(&req)
		rep = &gt.ReplyBlackwhiteRoundList{}
	} else if 2 == typ {
		req := gt.ReqLoopResult{
			GameID:  gameID,
			LoopSeq: int32(loopSeq),
		}
		params.FuncName = gt.GetBlackwhiteloopResult
		params.Payload = types.MustPBToJSON(&req)
		rep = &gt.ReplyLoopResults{}
	}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

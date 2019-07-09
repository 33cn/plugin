// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"strings"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/spf13/cobra"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)


// AutonomyCmd 自治系统命令行
func AutonomyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autonomy",
		Short: "autonomy management",
		Args:  cobra.MinimumNArgs(1),
	}

	// board
	cmd.AddCommand(
		ProposalBoardCmd(),
		RevokeProposalBoardCmd(),
		VoteProposalBoardCmd(),
		TerminateProposalBoardCmd(),
		ShowProposalBoardCmd(),
	)

	return cmd
}

// ProposalBoardCmd 创建提案命令
func ProposalBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposalboard",
		Short: "create proposal board",
		Run:   proposalBoard,
	}
	addProposalBoardFlags(cmd)
	return cmd
}

func addProposalBoardFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("year", "y", 0, "year")
	cmd.Flags().Int32P("month", "m", 0, "month")
	cmd.Flags().Int32P("day", "d", 0, "day")
	cmd.Flags().Int64P("startBlock", "s", 0, "start block height")
	cmd.MarkFlagRequired("startBlock")
	cmd.Flags().Int64P("endBlock", "e", 0, "end block height")
	cmd.MarkFlagRequired("endBlock")

	cmd.Flags().StringP("boards", "b", "", "addr1-addr2......addrN (3<=N<=30)")
	cmd.MarkFlagRequired("boards")
}

func proposalBoard(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")
	boardstr, _ := cmd.Flags().GetString("gameName")

	boards := strings.Split(boardstr, "-")

	params := &auty.ProposalBoard{
		Year:  year,
		Month: month,
		Day:     day,
		Boards:    boards,
		StartBlockHeight:  startBlock,
		EndBlockHeight: endBlock,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.PropBoardTx", params, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalBoardCmd 撤销提案
func RevokeProposalBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revokeBoard",
		Short: "revoke proposal board",
		Run:   revokeProposalBoard,
	}
	addRevokeProposalBoardFlags(cmd)
	return cmd
}

func addRevokeProposalBoardFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func revokeProposalBoard(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalBoard{
		ProposalID:     ID,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.RevokeProposalBoardTx", params, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalBoardCmd 撤销提案
func VoteProposalBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteBoard",
		Short: "vote proposal board",
		Run:   voteProposalBoard,
	}
	addVoteProposalBoardFlags(cmd)
	return cmd
}

func addVoteProposalBoardFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().BoolP("approve", "ap", true, "is approve, default true")
}

func voteProposalBoard(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetBool("approve")

	params := &auty.VoteProposalBoard{
		ProposalID:     ID,
		Approve: approve,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.VoteProposalBoardTx", params, &res)
	ctx.RunWithoutMarshal()
}

// TerminateProposalBoard 撤销提案
func TerminateProposalBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminateBoard",
		Short: "terminate proposal board",
		Run:   terminateProposalBoard,
	}
	addTerminateProposalBoardFlags(cmd)
	return cmd
}

func addTerminateProposalBoardFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func terminateProposalBoard(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalBoard{
		ProposalID:     ID,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.TerminateProposalBoardTx", params, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalBoardCmd 显示提案查询信息
func ShowProposalBoardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showInfo",
		Short: "show proposal board info",
		Run:   showProposalBoard,
	}
	addShowProposalBoardflags(cmd)
	return cmd
}

func addShowProposalBoardflags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("type", "t", 0, "type")
	cmd.MarkFlagRequired("type")

	cmd.Flags().Uint32P("status", "s", 0, "status")
	cmd.Flags().Int32P("count", "c", 0, "count")
	cmd.Flags().Int32P("direction", "d", 0, "direction")
	cmd.Flags().Int64P("index", "i", 0, "index")
}

func showProposalBoard(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	typ, _ := cmd.Flags().GetUint32("type")
	status, _ := cmd.Flags().GetUint32("status")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	index, _ := cmd.Flags().GetInt64("index")

	var params rpctypes.Query4Jrpc
	var rep interface{}
	params.Execer = auty.AutonomyX
	if 0 == typ {
		req := auty.ReqQueryProposalBoard{
			Status:    int32(status),
			Count:     count,
			Direction: direction,
			Index:     index,
		}
		params.FuncName = auty.GetProposalBoard
		params.Payload = types.MustPBToJSON(&req)
		rep = &auty.ReplyQueryProposalBoard{}
	}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

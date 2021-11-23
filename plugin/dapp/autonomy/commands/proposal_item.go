// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/spf13/cobra"
)

// ProposalItemCmd 创建提案命令
func ProposalItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposalItem",
		Short: "create proposal Item",
		Run:   proposalItem,
	}
	addProposalItemFlags(cmd)
	return cmd
}

func addProposalItemFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("year", "y", 0, "year")
	cmd.Flags().Int32P("month", "m", 0, "month")
	cmd.Flags().Int32P("day", "d", 0, "day")

	cmd.Flags().StringP("itemTxHash", "i", "", "the tx to apply check")
	cmd.MarkFlagRequired("itemTxHash")
	cmd.Flags().StringP("description", "p", "", "description item")

	cmd.Flags().Int64P("startBlock", "s", 0, "start block height")
	cmd.MarkFlagRequired("startBlock")
	cmd.Flags().Int64P("endBlock", "e", 0, "end block height")
	cmd.MarkFlagRequired("endBlock")
}

func proposalItem(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	txHash, _ := cmd.Flags().GetString("itemTxHash")
	description, _ := cmd.Flags().GetString("description")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")

	params := &auty.ProposalItem{
		Year:             year,
		Month:            month,
		Day:              day,
		ItemTxHash:       txHash,
		Description:      description,
		StartBlockHeight: startBlock,
		EndBlockHeight:   endBlock,
	}

	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "PropItem",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalItemCmd 撤销提案
func RevokeProposalItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revokeItem",
		Short: "revoke proposal Item",
		Run:   revokeProposalItem,
	}
	addRevokeProposalItemFlags(cmd)
	return cmd
}

func addRevokeProposalItemFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func revokeProposalItem(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalItem{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "RvkPropItem",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// VoteProposalItemCmd 投票提案
func VoteProposalItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteItem",
		Short: "vote proposal Item",
		Run:   voteProposalItem,
	}
	addVoteProposalItemFlags(cmd)
	return cmd
}

func addVoteProposalItemFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("approve", "r", 1, "1:approve, 2:oppose, 3:quit, default 1")
}

func voteProposalItem(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetInt32("approve")

	params := &auty.VoteProposalItem{
		ProposalID: ID,
		Vote:       auty.AutonomyVoteOption(approve),
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "VotePropItem",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// TerminateProposalItemCmd 终止提案
func TerminateProposalItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminateItem",
		Short: "terminate proposal Item",
		Run:   terminateProposalItem,
	}
	addTerminateProposalItemFlags(cmd)
	return cmd
}

func addTerminateProposalItemFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func terminateProposalItem(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalItem{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "TmintPropItem",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalItemCmd 显示提案查询信息
func ShowProposalItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showItem",
		Short: "show proposal Item info",
		Run:   showProposalItem,
	}
	addShowProposalItemflags(cmd)
	return cmd
}

func addShowProposalItemflags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("type", "y", 0, "type(0:query by hash; 1:list)")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")

	cmd.Flags().Uint32P("status", "s", 0, "status")
	cmd.Flags().StringP("addr", "a", "", "address")
	cmd.Flags().Int32P("count", "c", 1, "count, default is 1")
	cmd.Flags().Int32P("direction", "d", 0, "direction, default is reserve")
	cmd.Flags().Int64P("height", "t", -1, "height, default is -1")
	cmd.Flags().Int32P("index", "i", -1, "index, default is -1")
}

func showProposalItem(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	typ, _ := cmd.Flags().GetUint32("type")
	propID, _ := cmd.Flags().GetString("proposalID")
	status, _ := cmd.Flags().GetUint32("status")
	addr, _ := cmd.Flags().GetString("addr")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	height, _ := cmd.Flags().GetInt64("height")
	index, _ := cmd.Flags().GetInt32("index")

	var params rpctypes.Query4Jrpc
	var rep interface{}
	params.Execer = auty.AutonomyX
	if 0 == typ {
		req := types.ReqString{
			Data: propID,
		}
		params.FuncName = auty.GetProposalItem
		params.Payload = types.MustPBToJSON(&req)
		rep = &auty.ReplyQueryProposalItem{}
	} else if 1 == typ {
		req := auty.ReqQueryProposalItem{
			Status:    int32(status),
			Addr:      addr,
			Count:     count,
			Direction: direction,
			Height:    height,
			Index:     index,
		}
		params.FuncName = auty.ListProposalItem
		params.Payload = types.MustPBToJSON(&req)
		rep = &auty.ReplyQueryProposalItem{}
	}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

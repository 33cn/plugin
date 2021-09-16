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

// ProposalChangeCmd 创建提案命令
func ProposalChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposalChange",
		Short: "create proposal change",
		Run:   proposalChange,
	}
	addProposalChangeFlags(cmd)
	return cmd
}

func addProposalChangeFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("year", "y", 0, "year")
	cmd.Flags().Int32P("month", "m", 0, "month")
	cmd.Flags().Int32P("day", "d", 0, "day")
	cmd.Flags().Int64P("startBlock", "s", 0, "start block height")
	cmd.MarkFlagRequired("startBlock")
	cmd.Flags().Int64P("endBlock", "e", 0, "end block height")
	cmd.MarkFlagRequired("endBlock")

	cmd.Flags().StringP("change", "c", "", "addr")
	cmd.MarkFlagRequired("change")
}

func proposalChange(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")
	changeAddrstr, _ := cmd.Flags().GetString("change")

	var changes []*auty.Change
	change := &auty.Change{Cancel: true, Addr: changeAddrstr}
	changes = append(changes, change)

	params := &auty.ProposalChange{
		Year:             year,
		Month:            month,
		Day:              day,
		Changes:          changes,
		StartBlockHeight: startBlock,
		EndBlockHeight:   endBlock,
	}

	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "PropChange",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalChangeCmd 撤销提案
func RevokeProposalChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revokeChange",
		Short: "revoke proposal change",
		Run:   revokeProposalChange,
	}
	addRevokeProposalChangeFlags(cmd)
	return cmd
}

func addRevokeProposalChangeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func revokeProposalChange(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalChange{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "RvkPropChange",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// VoteProposalChangeCmd 投票提案
func VoteProposalChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteChange",
		Short: "vote proposal change",
		Run:   voteProposalChange,
	}
	addVoteProposalChangeFlags(cmd)
	return cmd
}

func addVoteProposalChangeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("approve", "r", 1, "1:approve, 2:oppose, 3:quit, default 1")
}

func voteProposalChange(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetInt32("approve")

	params := &auty.VoteProposalChange{
		ProposalID: ID,
		Vote:       auty.AutonomyVoteOption(approve),
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "VotePropChange",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// TerminateProposalChangeCmd 终止提案
func TerminateProposalChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminateChange",
		Short: "terminate proposal change",
		Run:   terminateProposalChange,
	}
	addTerminateProposalChangeFlags(cmd)
	return cmd
}

func addTerminateProposalChangeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func terminateProposalChange(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalChange{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "TmintPropChange",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalChangeCmd 显示提案查询信息
func ShowProposalChangeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showChange",
		Short: "show proposal change info",
		Run:   showProposalChange,
	}
	addShowProposalChangeflags(cmd)
	return cmd
}

func addShowProposalChangeflags(cmd *cobra.Command) {
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

func showProposalChange(cmd *cobra.Command, args []string) {
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
		params.FuncName = auty.GetProposalChange
		params.Payload = types.MustPBToJSON(&req)
	} else if 1 == typ {
		req := auty.ReqQueryProposalChange{
			Status:    int32(status),
			Addr:      addr,
			Count:     count,
			Direction: direction,
			Height:    height,
			Index:     index,
		}
		params.FuncName = auty.ListProposalChange
		params.Payload = types.MustPBToJSON(&req)
	}
	rep = &auty.ReplyQueryProposalChange{}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

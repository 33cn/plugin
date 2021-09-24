// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	commandtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

	"strings"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/spf13/cobra"
)

// ProposalProjectCmd 创建提案命令
func ProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposalProject",
		Short: "create proposal Project",
		Run:   proposalProject,
	}
	addProposalProjectFlags(cmd)
	return cmd
}

func addProposalProjectFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("year", "y", 0, "year")
	cmd.Flags().Int32P("month", "m", 0, "month")
	cmd.Flags().Int32P("day", "d", 0, "day")

	cmd.Flags().StringP("firstStage", "f", "", "first stage proposal ID")
	cmd.Flags().StringP("lastStage", "l", "", "last stage proposal ID")
	cmd.Flags().StringP("production", "p", "", "production address")
	cmd.Flags().StringP("description", "i", "", "description project")
	cmd.Flags().StringP("contractor", "c", "", "contractor introduce")
	cmd.Flags().Int64P("amount", "a", 0, "project cost amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("amountDetail", "t", "", "project cost amount detail")
	cmd.Flags().StringP("toAddr", "o", "", "project contractor account address")
	cmd.MarkFlagRequired("toAddr")

	cmd.Flags().Int64P("startBlock", "s", 0, "start block height")
	cmd.MarkFlagRequired("startBlock")
	cmd.Flags().Int64P("endBlock", "e", 0, "end block height")
	cmd.MarkFlagRequired("endBlock")
	cmd.Flags().Int32P("projectNeedBlockNum", "n", 0, "project complete need time(unit is block number)")
}

func proposalProject(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	firstStage, _ := cmd.Flags().GetString("firstStage")
	lastStage, _ := cmd.Flags().GetString("lastStage")
	production, _ := cmd.Flags().GetString("production")
	description, _ := cmd.Flags().GetString("description")
	contractor, _ := cmd.Flags().GetString("contractor")

	amount, _ := cmd.Flags().GetInt64("amount")

	amountDetail, _ := cmd.Flags().GetString("amountDetail")
	toAddr, _ := cmd.Flags().GetString("toAddr")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")
	projectNeedBlockNum, _ := cmd.Flags().GetInt32("projectNeedBlockNum")

	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	params := &auty.ProposalProject{
		Year:                year,
		Month:               month,
		Day:                 day,
		FirstStage:          firstStage,
		LastStage:           lastStage,
		Production:          production,
		Description:         description,
		Contractor:          contractor,
		Amount:              amount * cfg.CoinPrecision,
		AmountDetail:        amountDetail,
		ToAddr:              toAddr,
		StartBlockHeight:    startBlock,
		EndBlockHeight:      endBlock,
		ProjectNeedBlockNum: projectNeedBlockNum,
	}

	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "PropProject",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalProjectCmd 撤销提案
func RevokeProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revokeProject",
		Short: "revoke proposal Project",
		Run:   revokeProposalProject,
	}
	addRevokeProposalProjectFlags(cmd)
	return cmd
}

func addRevokeProposalProjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func revokeProposalProject(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalProject{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "RvkPropProject",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// VoteProposalProjectCmd 投票提案
func VoteProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteProject",
		Short: "vote proposal Project",
		Run:   voteProposalProject,
	}
	addVoteProposalProjectFlags(cmd)
	return cmd
}

func addVoteProposalProjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("approve", "r", 1, "1:approve, 2:oppose, 3:quit, default 1")
}

func voteProposalProject(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetInt32("approve")

	params := &auty.VoteProposalProject{
		ProposalID: ID,
		Vote:       auty.AutonomyVoteOption(approve),
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "VotePropProject",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// PubVoteProposalProjectCmd 全员投票提案
func PubVoteProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pubvoteProject",
		Short: "pub vote proposal Project",
		Run:   pubVoteProposalProject,
	}
	addPubVoteProposalProjectFlags(cmd)
	return cmd
}

func addPubVoteProposalProjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("oppose", "s", 1, "is oppose, default true")
	cmd.Flags().StringP("originAddr", "o", "", "origin address: addr1-addr2......addrN")
}

func pubVoteProposalProject(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	oppose, _ := cmd.Flags().GetInt32("oppose")
	originAddr, _ := cmd.Flags().GetString("originAddr")

	var isopp bool
	if oppose == 0 {
		isopp = false
	} else {
		isopp = true
	}

	var originAddrs []string
	if len(originAddr) > 0 {
		originAddrs = strings.Split(originAddr, "-")
	}

	params := &auty.PubVoteProposalProject{
		ProposalID: ID,
		Oppose:     isopp,
		OriginAddr: originAddrs,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "PubVotePropProject",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// TerminateProposalProjectCmd 终止提案
func TerminateProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminateProject",
		Short: "terminate proposal Project",
		Run:   terminateProposalProject,
	}
	addTerminateProposalProjectFlags(cmd)
	return cmd
}

func addTerminateProposalProjectFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func terminateProposalProject(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalProject{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "TmintPropProject",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalProjectCmd 显示提案查询信息
func ShowProposalProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showProject",
		Short: "show proposal project info",
		Run:   showProposalProject,
	}
	addShowProposalProjectflags(cmd)
	return cmd
}

func addShowProposalProjectflags(cmd *cobra.Command) {
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

func showProposalProject(cmd *cobra.Command, args []string) {
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
		params.FuncName = auty.GetProposalProject
		params.Payload = types.MustPBToJSON(&req)
		rep = &auty.ReplyQueryProposalProject{}
	} else if 1 == typ {
		req := auty.ReqQueryProposalProject{
			Status:    int32(status),
			Addr:      addr,
			Count:     count,
			Direction: direction,
			Height:    height,
			Index:     index,
		}
		params.FuncName = auty.ListProposalProject
		params.Payload = types.MustPBToJSON(&req)
		rep = &auty.ReplyQueryProposalProject{}
	}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pkg/errors"

	"strings"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	commandtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	"github.com/spf13/cobra"
)

// ProposalRuleCmd 创建提案命令
func ProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposalRule",
		Short: "create proposal Rule",
		Run:   proposalRule,
	}
	addProposalRuleFlags(cmd)
	return cmd
}

func addProposalRuleFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("year", "y", 0, "year")
	cmd.Flags().Int32P("month", "m", 0, "month")
	cmd.Flags().Int32P("day", "d", 0, "day")
	cmd.Flags().Int64P("startBlock", "s", 0, "start block height")
	cmd.MarkFlagRequired("startBlock")
	cmd.Flags().Int64P("endBlock", "e", 0, "end block height")
	cmd.MarkFlagRequired("endBlock")

	// 可修改规则
	cmd.Flags().Int32P("boardApproveRatio", "r", 0, "board approve ratio(unit is %)")
	cmd.Flags().Int32P("pubOpposeRatio", "o", 0, "public oppose ratio(unit is %)")
	cmd.Flags().Int64P("proposalAmount", "p", 0, "proposal cost amount")
	cmd.Flags().Int64P("largeProjectAmount", "l", 0, "large project amount threshold")
	cmd.Flags().Int32P("publicPeriod", "u", 0, "public time")
	cmd.Flags().Int32P("pubAttendRatio", "a", 0, "public attend ratio(unit is %)")
	cmd.Flags().Int32P("pubApproveRatio", "v", 0, "public approve ratio(unit is %)")
}

func proposalRule(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")

	boardApproveRatio, _ := cmd.Flags().GetInt32("boardApproveRatio")
	pubOpposeRatio, _ := cmd.Flags().GetInt32("pubOpposeRatio")

	proposalAmount, _ := cmd.Flags().GetInt64("proposalAmount")
	largeProjectAmount, _ := cmd.Flags().GetInt64("largeProjectAmount")
	publicPeriod, _ := cmd.Flags().GetInt32("publicPeriod")
	pubAttendRatio, _ := cmd.Flags().GetInt32("pubAttendRatio")
	pubApproveRatio, _ := cmd.Flags().GetInt32("pubApproveRatio")

	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	params := &auty.ProposalRule{
		Year:  year,
		Month: month,
		Day:   day,
		RuleCfg: &auty.RuleConfig{
			BoardApproveRatio:  boardApproveRatio,
			PubOpposeRatio:     pubOpposeRatio,
			ProposalAmount:     proposalAmount * cfg.CoinPrecision,
			LargeProjectAmount: largeProjectAmount * cfg.CoinPrecision,
			PublicPeriod:       publicPeriod,
			PubAttendRatio:     pubAttendRatio,
			PubApproveRatio:    pubApproveRatio,
		},
		StartBlockHeight: startBlock,
		EndBlockHeight:   endBlock,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "PropRule",
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// RevokeProposalRuleCmd 撤销提案
func RevokeProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revokeRule",
		Short: "revoke proposal Rule",
		Run:   revokeProposalRule,
	}
	addRevokeProposalRuleFlags(cmd)
	return cmd
}

func addRevokeProposalRuleFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func revokeProposalRule(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalRule{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "RvkPropRule",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// VoteProposalRuleCmd 投票提案
func VoteProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteRule",
		Short: "vote proposal Rule",
		Run:   voteProposalRule,
	}
	addVoteProposalRuleFlags(cmd)
	return cmd
}

func addVoteProposalRuleFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("approve", "r", 1, "1:approve, 2:oppose, 3:quit, default 1")
	cmd.Flags().StringP("originAddr", "o", "", "origin address: addr1-addr2......addrN")
}

func voteProposalRule(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetInt32("approve")
	originAddr, _ := cmd.Flags().GetString("originAddr")

	var originAddrs []string
	if len(originAddr) > 0 {
		originAddrs = strings.Split(originAddr, "-")
	}

	params := &auty.VoteProposalRule{
		ProposalID: ID,
		OriginAddr: originAddrs,
		Vote:       auty.AutonomyVoteOption(approve),
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "VotePropRule",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// TerminateProposalRuleCmd 终止提案
func TerminateProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminateRule",
		Short: "terminate proposal Rule",
		Run:   terminateProposalRule,
	}
	addTerminateProposalRuleFlags(cmd)
	return cmd
}

func addTerminateProposalRuleFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
}

func terminateProposalRule(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalRule{
		ProposalID: ID,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "TmintPropRule",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalRuleCmd 显示提案查询信息
func ShowProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showRule",
		Short: "show proposal rule info",
		Run:   showProposalRule,
	}
	addShowProposalRuleflags(cmd)
	return cmd
}

func addShowProposalRuleflags(cmd *cobra.Command) {
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

func showProposalRule(cmd *cobra.Command, args []string) {
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
		params.FuncName = auty.GetProposalRule
		params.Payload = types.MustPBToJSON(&req)
	} else if 1 == typ {
		req := auty.ReqQueryProposalRule{
			Status:    int32(status),
			Addr:      addr,
			Count:     count,
			Direction: direction,
			Height:    height,
			Index:     index,
		}
		params.FuncName = auty.ListProposalRule
		params.Payload = types.MustPBToJSON(&req)
	}
	rep = &auty.ReplyQueryProposalRule{}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

// ShowActiveRuleCmd 显示提案查询信息
func ShowActiveRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showActiveRule",
		Short: "show active rule",
		Run:   showActiveRule,
	}
	return cmd
}

func showActiveRule(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	params := rpctypes.Query4Jrpc{}
	params.Execer = auty.AutonomyX
	params.FuncName = auty.GetActiveRule
	params.Payload = types.MustPBToJSON(&types.ReqString{})
	rep := &auty.RuleConfig{}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

// TransferFundCmd 资金转入自治系统合约中
func TransferFundCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transferFund",
		Short: "transfer fund",
		Run:   transferFund,
	}
	addTransferFundflags(cmd)
	return cmd
}

func addTransferFundflags(cmd *cobra.Command) {
	cmd.Flags().Int64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("note", "n", "", "note")
}

func transferFund(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	amount, _ := cmd.Flags().GetInt64("amount")
	note, _ := cmd.Flags().GetString("note")
	paraName, _ := cmd.Flags().GetString("paraName")

	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	params := &auty.TransferFund{
		Amount: amount * cfg.CoinPrecision,
		Note:   note,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "Transfer",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// CommentProposalCmd 评论提案
func CommentProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "comment proposal",
		Run:   commentProposal,
	}
	addCommentProposalflags(cmd)
	return cmd
}

func addCommentProposalflags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().StringP("repHash", "r", "", "reply Comment hash")
	cmd.Flags().StringP("comment", "c", "", "comment")
	cmd.MarkFlagRequired("comment")
}

func commentProposal(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proposalID, _ := cmd.Flags().GetString("proposalID")
	repHash, _ := cmd.Flags().GetString("repHash")
	comment, _ := cmd.Flags().GetString("comment")

	params := &auty.Comment{
		ProposalID: proposalID,
		RepHash:    repHash,
		Comment:    comment,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(auty.AutonomyX, paraName),
		ActionName: "CommentProp",
		Payload:    payLoad,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalCommentCmd 显示提案评论查询信息
func ShowProposalCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showComment",
		Short: "show proposal comment info",
		Run:   showProposalComment,
	}
	addShowProposalCommentflags(cmd)
	return cmd
}

func addShowProposalCommentflags(cmd *cobra.Command) {
	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.MarkFlagRequired("proposalID")
	cmd.Flags().Int32P("count", "c", 1, "count, default is 1")
	cmd.Flags().Int32P("direction", "d", 0, "direction, default is reserve")
	cmd.Flags().Int64P("height", "t", -1, "height, default is -1")
	cmd.Flags().Int64P("index", "i", -1, "index, default is -1")
}

func showProposalComment(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	propID, _ := cmd.Flags().GetString("proposalID")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	height, _ := cmd.Flags().GetInt64("height")
	index, _ := cmd.Flags().GetInt32("index")

	var params rpctypes.Query4Jrpc
	var rep interface{}
	params.Execer = auty.AutonomyX

	req := auty.ReqQueryProposalComment{
		ProposalID: propID,
		Count:      count,
		Direction:  direction,
		Height:     height,
		Index:      index,
	}
	params.FuncName = auty.ListProposalComment
	params.Payload = types.MustPBToJSON(&req)

	rep = &auty.ReplyQueryProposalComment{}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

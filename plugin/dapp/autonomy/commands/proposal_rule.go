// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
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
	cmd.Flags().Int32P("boardAttendRatio", "t", 0, "board attend ratio(unit is %)")
	cmd.Flags().Int32P("boardApproveRatio", "r", 0, "board approve ratio(unit is %)")
	cmd.Flags().Int32P("pubOpposeRatio", "o", 0, "public oppose ratio(unit is %)")
	cmd.Flags().Int64P("proposalAmount", "p", 0, "proposal cost amount")
	cmd.Flags().Int64P("largeProjectAmount", "l", 0, "large project amount threshold")
	cmd.Flags().Int32P("publicPeriod", "u", 0, "public time")
}

func proposalRule(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	year, _ := cmd.Flags().GetInt32("year")
	month, _ := cmd.Flags().GetInt32("month")
	day, _ := cmd.Flags().GetInt32("day")

	startBlock, _ := cmd.Flags().GetInt64("startBlock")
	endBlock, _ := cmd.Flags().GetInt64("endBlock")

	boardAttendRatio, _ := cmd.Flags().GetInt32("boardAttendRatio")
	boardApproveRatio, _ := cmd.Flags().GetInt32("boardApproveRatio")
	pubOpposeRatio, _ := cmd.Flags().GetInt32("pubOpposeRatio")

	proposalAmount, _ := cmd.Flags().GetInt64("proposalAmount")
	largeProjectAmount, _ := cmd.Flags().GetInt64("largeProjectAmount")
	publicPeriod, _ := cmd.Flags().GetInt32("publicPeriod")

	params := &auty.ProposalRule{
		Year:  year,
		Month: month,
		Day:   day,
		RuleCfg: &auty.RuleConfig{
			BoardAttendRatio:   boardAttendRatio,
			BoardApproveRatio:  boardApproveRatio,
			PubOpposeRatio:     pubOpposeRatio,
			ProposalAmount:     proposalAmount * types.Coin,
			LargeProjectAmount: largeProjectAmount * types.Coin,
			PublicPeriod:       publicPeriod,
		},
		StartBlockHeight: startBlock,
		EndBlockHeight:   endBlock,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.PropRuleTx", params, &res)
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
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalRule{
		ProposalID: ID,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.RevokeProposalRuleTx", params, &res)
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
	cmd.Flags().Int32P("approve", "r", 1, "is approve, default true")
}

func voteProposalRule(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")
	approve, _ := cmd.Flags().GetInt32("approve")
	var isapp bool
	if approve == 0 {
		isapp = false
	} else {
		isapp = true
	}

	params := &auty.VoteProposalRule{
		ProposalID: ID,
		Approve:    isapp,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.VoteProposalRuleTx", params, &res)
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
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ID, _ := cmd.Flags().GetString("proposalID")

	params := &auty.RevokeProposalRule{
		ProposalID: ID,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.TerminateProposalRuleTx", params, &res)
	ctx.RunWithoutMarshal()
}

// ShowProposalRuleCmd 显示提案查询信息
func ShowProposalRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showRuleInfo",
		Short: "show proposal rule info",
		Run:   showProposalRule,
	}
	addShowProposalRuleflags(cmd)
	return cmd
}

func addShowProposalRuleflags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("type", "t", 0, "type(0:query by hash; 1:list)")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("proposalID", "p", "", "proposal ID")
	cmd.Flags().Uint32P("status", "s", 0, "status")
	cmd.Flags().Int32P("count", "c", 1, "count, default is 1")
	cmd.Flags().Int32P("direction", "d", -1, "direction, default is reserve")
	cmd.Flags().Int64P("index", "i", -1, "index, default is -1")
}

func showProposalRule(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	typ, _ := cmd.Flags().GetUint32("type")
	propID, _ := cmd.Flags().GetString("proposalID")
	status, _ := cmd.Flags().GetUint32("status")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	index, _ := cmd.Flags().GetInt64("index")

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
			Count:     count,
			Direction: direction,
			Index:     index,
		}
		params.FuncName = auty.ListProposalRule
		params.Payload = types.MustPBToJSON(&req)
	}
	rep = &auty.ReplyQueryProposalRule{}

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

	params := &auty.TransferFund{
		Amount: amount * types.Coin,
		Note:   note,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.TransferFundTx", params, &res)
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
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proposalID, _ := cmd.Flags().GetString("proposalID")
	repHash, _ := cmd.Flags().GetString("repHash")
	comment, _ := cmd.Flags().GetString("comment")

	params := &auty.Comment{
		ProposalID: proposalID,
		RepHash: repHash,
		Comment:    comment,
	}
	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "autonomy.CommentProposalTx", params, &res)
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
	cmd.Flags().Int32P("direction", "d", -1, "direction, default is reserve")
	cmd.Flags().Int64P("index", "i", -1, "index, default is -1")
}

func showProposalComment(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	propID, _ := cmd.Flags().GetString("proposalID")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	index, _ := cmd.Flags().GetInt64("index")

	var params rpctypes.Query4Jrpc
	var rep interface{}
	params.Execer = auty.AutonomyX

	req := auty.ReqQueryProposalComment{
		ProposalID: propID,
		Count:      count,
		Direction:  direction,
		Index:      index,
	}
	params.FuncName = auty.ListProposalComment
	params.Payload = types.MustPBToJSON(&req)

	rep = &auty.ReplyQueryProposalComment{}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

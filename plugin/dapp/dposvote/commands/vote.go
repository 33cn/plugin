// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"strings"
	"time"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	"github.com/spf13/cobra"
)

//DPosCmd DPosVote合约命令行
func DPosCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dpos",
		Short: "dpos vote management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		DPosRegistCmd(),
		DPosCancelRegistCmd(),
		DPosVoteCmd(),
		DPosReRegistCmd(),
		DPosVoteCancelCmd(),
		DPosCandidatorQueryCmd(),
		DPosVoteQueryCmd(),
		DPosVrfMRegistCmd(),
		DPosVrfRPRegistCmd(),
		DPosVrfQueryCmd(),
	)

	return cmd
}

//DPosRegistCmd 构造候选节点注册的命令行
func DPosRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regist",
		Short: "regist a new candidator",
		Run:   regist,
	}
	addRegistFlags(cmd)
	return cmd
}

func addRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("ip", "i", "", "ip")
	cmd.MarkFlagRequired("address")
}

func regist(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")
	ip, _ := cmd.Flags().GetString("ip")


	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\", \"ip\":\"%s\"}", pubkey, address, ip)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()

}

//DPosCancelRegistCmd 构造候选节点去注册的命令行
func DPosCancelRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelRegist",
		Short: "cancel regist for a candidator",
		Run:   cancelRegist,
	}
	addCancelRegistFlags(cmd)
	return cmd
}

func addCancelRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")
}

func cancelRegist(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\"}", pubkey, address)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateCancelRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVoteCmd 构造为候选节点投票的命令行
func DPosVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote for one candidator",
		Run:   vote,
	}
	addVoteFlags(cmd)
	return cmd
}

func addVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey of a candidator")
	cmd.MarkFlagRequired("pubkey")
	cmd.Flags().Int64P("votes", "v", 0, "votes")
	cmd.MarkFlagRequired("votes")
	cmd.Flags().StringP("addr", "a", "", "address of voter")
	cmd.MarkFlagRequired("addr")
}

func vote(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	votes, _ := cmd.Flags().GetInt64("votes")
	addr, _ := cmd.Flags().GetString("addr")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"votes\":\"%d\", \"fromAddr\":\"%s\"}", pubkey, votes, addr)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateVoteTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVoteCancelCmd 构造撤销对候选节点投票的命令行
func DPosVoteCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancelVote",
		Short: "cancel votes to a candidator",
		Run:   cancelVote,
	}
	addCancelVoteFlags(cmd)
	return cmd
}

func addCancelVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey of a candidator")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("votes", "v", 0, "votes")
	cmd.MarkFlagRequired("votes")
}

func cancelVote(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	votes, _ := cmd.Flags().GetInt64("votes")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"votes\":\"%d\"}", pubkey, votes)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateCancelVoteTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosReRegistCmd 构造重新注册候选节点的命令行
func DPosReRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reRegist",
		Short: "re regist a canceled candidator",
		Run:   reRegist,
	}
	addReRegistFlags(cmd)
	return cmd
}

func addReRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("ip", "i", "", "ip")
	cmd.MarkFlagRequired("address")
}

func reRegist(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	address, _ := cmd.Flags().GetString("address")
	ip, _ := cmd.Flags().GetString("ip")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"address\":\"%s\", \"ip\":\"%s\"}", pubkey, address, ip)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateReRegistTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()

}

//DPosCandidatorQueryCmd 构造查询候选节点信息的命令行
func DPosCandidatorQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidatorQuery",
		Short: "query candidator info",
		Run:   candidatorQuery,
	}
	addCandidatorQueryFlags(cmd)
	return cmd
}

func addCandidatorQueryFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("type", "t", "", "topN/pubkeys")

	cmd.Flags().Int64P("top", "n", 0, "top N by votes")

	cmd.Flags().StringP("pubkeys", "k", "", "pubkeys")
}

func candidatorQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	pubkeys, _ := cmd.Flags().GetString("pubkeys")
	topN, _ := cmd.Flags().GetInt64("top")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	switch ty {
	case "topN":
		req := &dty.CandidatorQuery{
			TopN: int32(topN),
		}
		params.FuncName = dty.FuncNameQueryCandidatorByTopN
		params.Payload = types.MustPBToJSON(req)
		var res dty.CandidatorReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "pubkeys":
		keys := strings.Split(pubkeys, ";")
		req := &dty.CandidatorQuery{
		}
		for _, key := range keys {
			req.Pubkeys = append(req.Pubkeys, key)
		}
		params.FuncName = dty.FuncNameQueryCandidatorByPubkeys
		params.Payload = types.MustPBToJSON(req)
		var res dty.CandidatorReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}
}


//DPosVoteQueryCmd 构造投票信息查询的命令行
func DPosVoteQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "voteQuery",
		Short: "query vote info",
		Run:   voteQuery,
	}
	addVoteQueryFlags(cmd)
	return cmd
}

func addVoteQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkeys", "k", "", "pubkeys")
	cmd.Flags().StringP("address", "a", "", "address")
}

func voteQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkeys, _ := cmd.Flags().GetString("pubkeys")
	addr, _ := cmd.Flags().GetString("address")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	req := &dty.DposVoteQuery{
		Addr: addr,
	}

	keys := strings.Split(pubkeys, ";")
	for _, key := range keys {
		req.Pubkeys = append(req.Pubkeys, key)
	}

	params.FuncName = dty.FuncNameQueryVote
	params.Payload = types.MustPBToJSON(req)
	var res dty.DposVoteReply
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()

}

//DPosVrfMRegistCmd 构造注册VRF M信息（输入信息）的命令行
func DPosVrfMRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfMRegist",
		Short: "regist m of vrf",
		Run:   vrfM,
	}
	addVrfMFlags(cmd)
	return cmd
}

func addVrfMFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("cycle", "c", 0, "cycle no. of dpos consensus")
	cmd.MarkFlagRequired("cycle")

	cmd.Flags().StringP("m", "m", "", "input of vrf")
	cmd.MarkFlagRequired("m")
}

func vrfM(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	m, _ := cmd.Flags().GetString("m")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"cycle\":\"%d\", \"m\":\"%X\"}", pubkey, cycle, m)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateRegistVrfMTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVrfRPRegistCmd 构造VRF R/P(hash及proof)注册的命令行
func DPosVrfRPRegistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfRPRegist",
		Short: "regist r,p of vrf",
		Run:   vrfRP,
	}
	addVrfRPRegistFlags(cmd)
	return cmd
}

func addVrfRPRegistFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkey", "k", "", "pubkey")
	cmd.MarkFlagRequired("pubkey")

	cmd.Flags().Int64P("cycle", "c", 0, "cycle no. of dpos consensus")
	cmd.MarkFlagRequired("cycle")

	cmd.Flags().StringP("hash", "r", "", "hash of vrf")
	cmd.MarkFlagRequired("hash")

	cmd.Flags().StringP("proof", "p", "", "proof of vrf")
	cmd.MarkFlagRequired("proof")
}

func vrfRP(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	cycle, _ := cmd.Flags().GetInt64("cycle")
	r, _ := cmd.Flags().GetString("r")
	p, _ := cmd.Flags().GetString("p")

	payload := fmt.Sprintf("{\"pubkey\":\"%s\", \"cycle\":\"%d\", \"r\":\"%s\", \"p\":\"%s\"}", pubkey, cycle, r, p)
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(dty.DPosX),
		ActionName: dty.CreateRegistVrfRPTx,
		Payload:    []byte(payload),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//DPosVrfQueryCmd 构造VRF相关信息查询的命令行
func DPosVrfQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vrfQuery",
		Short: "query vrf info",
		Run:   vrfQuery,
	}
	addVrfQueryFlags(cmd)
	return cmd
}

func addVrfQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "query type")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("time", "d", "", "time like 2019-06-18")
	cmd.Flags().Int64P("timestamp", "s", 0, "time stamp from 1970-1-1")
	cmd.Flags().Int64P("cycle", "c", 0, "cycle,one time point belongs to a cycle")

}

func vrfQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	dtime, _ := cmd.Flags().GetString("time")
	timestamp, _ := cmd.Flags().GetInt64("timestamp")
	cycle, _ := cmd.Flags().GetInt64("cycle")

	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	switch ty {
	case "dtime":
		t, err := time.Parse("2006-01-02 15:04:05", dtime)
		if err != nil {
			fmt.Println("err time format:", dtime)
			return
		}

		req := &dty.DposVrfQuery{
			Ty: dty.QueryVrfByTime,
			Timestamp: t.Unix(),
		}

		params.FuncName = dty.FuncNameQueryVrfByTime
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "timestamp":
		if timestamp <= 0 {
			fmt.Println("err timestamp:", timestamp)
			return
		}

		req := &dty.DposVrfQuery{
			Ty: dty.QueryVrfByTime,
			Timestamp: timestamp,
		}

		params.FuncName = dty.FuncNameQueryVrfByTime
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()

	case "cycle":
		if cycle <= 0 {
			fmt.Println("err cycle:", cycle)
			return
		}

		req := &dty.DposVrfQuery{
			Ty: dty.QueryVrfByCycle,
			Cycle: cycle,
		}

		params.FuncName = dty.FuncNameQueryVrfByCycle
		params.Payload = types.MustPBToJSON(req)
		var res dty.DposVrfReply
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	}
}
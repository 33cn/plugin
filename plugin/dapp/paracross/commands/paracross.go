// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/system/dapp/commands"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/spf13/cobra"
)

//ParcCmd paracross cmd register
func ParcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "para",
		Short: "Construct para transactions",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateRawAssetTransferCmd(),
		CreateRawAssetWithdrawCmd(),
		CreateRawTransferCmd(),
		CreateRawWithdrawCmd(),
		CreateRawTransferToExecCmd(),
		superNodeCmd(),
		nodeGroupCmd(),
		paraConfigCmd(),
		GetParaInfoCmd(),
		GetParaListCmd(),
		GetParaAssetTransCmd(),
		IsSyncCmd(),
		GetHeightCmd(),
		GetBlockInfoCmd(),
		GetLocalBlockInfoCmd(),
		GetConsensDoneInfoCmd(),
	)
	return cmd
}

// CreateRawAssetTransferCmd create raw asset transfer tx
func CreateRawAssetTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset_transfer",
		Short: "Create a asset transfer to para-chain transaction",
		Run:   createAssetTransfer,
	}
	addCreateAssetTransferFlags(cmd)
	return cmd
}

func addCreateAssetTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "receiver account address")
	cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
}

func createAssetTransfer(cmd *cobra.Command, args []string) {
	txHex, err := createAssetTx(cmd, false)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(txHex)
}

// CreateRawAssetWithdrawCmd create raw asset withdraw tx
func CreateRawAssetWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset_withdraw",
		Short: "Create a asset withdraw to para-chain transaction",
		Run:   createAssetWithdraw,
	}
	addCreateAssetWithdrawFlags(cmd)
	return cmd
}

func addCreateAssetWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("to", "t", "", "receiver account address")
	cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
}

func createAssetWithdraw(cmd *cobra.Command, args []string) {
	txHex, err := createAssetTx(cmd, true)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(txHex)
}

func createAssetTx(cmd *cobra.Command, isWithdraw bool) (string, error) {
	title, _ := cmd.Flags().GetString("title")
	//这里cfg除了里面FormatTx需要外，没其他作用，平行链执行器需要的参数已经填好了，这里title就是默认空就可以，支持主链构建平行链交易
	cfg := types.GetCliSysParam(title)

	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0 {
		return "", types.ErrAmount
	}
	amountInt64 := int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4

	toAddr, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	paraName, _ := cmd.Flags().GetString("paraName")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "title is not right, title format like `user.p.guodun.`")
		return "", types.ErrInvalidParam
	}
	execName := paraName + pt.ParaX

	param := types.CreateTx{
		To:          toAddr,
		Amount:      amountInt64,
		Fee:         0,
		Note:        []byte(note),
		IsWithdraw:  isWithdraw,
		IsToken:     false,
		TokenSymbol: symbol,
		ExecName:    execName,
	}
	tx, err := pt.CreateRawAssetTransferTx(cfg, &param)
	if err != nil {
		return "", err
	}

	txHex := types.Encode(tx)
	return hex.EncodeToString(txHex), nil
}

//CreateRawTransferCmd  create raw transfer tx
func CreateRawTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a transfer transaction",
		Run:   createTransfer,
	}
	addCreateTransferFlags(cmd)
	return cmd
}

func addCreateTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "receiver account address")
	cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
	cmd.MarkFlagRequired("symbol")
}

func createTransfer(cmd *cobra.Command, args []string) {
	commands.CreateAssetTransfer(cmd, args, pt.ParaX)
}

//CreateRawTransferToExecCmd create raw transfer to exec tx
func CreateRawTransferToExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_exec",
		Short: "Create a transfer to exec transaction",
		Run:   createTransferToExec,
	}
	addCreateTransferToExecFlags(cmd)
	return cmd
}

func addCreateTransferToExecFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "coins.bty", "default for bty, symbol for token")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	cmd.MarkFlagRequired("exec")
}

func createTransferToExec(cmd *cobra.Command, args []string) {
	commands.CreateAssetSendToExec(cmd, args, pt.ParaX)
}

//CreateRawWithdrawCmd create raw withdraw tx
func CreateRawWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Create a withdraw transaction",
		Run:   createWithdraw,
	}
	addCreateWithdrawFlags(cmd)
	return cmd
}

func addCreateWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	cmd.MarkFlagRequired("exec")
}

func createWithdraw(cmd *cobra.Command, args []string) {
	commands.CreateAssetWithdraw(cmd, args, pt.ParaX)
}

func superNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "super_node",
		Short: "super node manage cmd",
	}
	cmd.AddCommand(nodeJoinCmd())
	cmd.AddCommand(nodeVoteCmd())
	cmd.AddCommand(nodeQuitCmd())
	cmd.AddCommand(nodeCancelCmd())

	cmd.AddCommand(getNodeInfoCmd())
	cmd.AddCommand(getNodeIDInfoCmd())
	cmd.AddCommand(getNodeListCmd())
	return cmd
}

func addNodeJoinFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "target join addr")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().Float64P("coins", "c", 0, "frozen coins amount, should not less nodegroup's setting")
	cmd.MarkFlagRequired("coins")

}

func createNodeJoinTx(cmd *cobra.Command, args []string) {
	opAddr, _ := cmd.Flags().GetString("addr")
	coins, _ := cmd.Flags().GetFloat64("coins")
	paraName, _ := cmd.Flags().GetString("paraName")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: 1, Addr: opAddr, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nodeJoinCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "join",
		Short: "super node apply for join nodegroup cmd",
		Run:   createNodeJoinTx,
	}
	addNodeJoinFlags(cmd)
	return cmd
}

func addNodeVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	cmd.MarkFlagRequired("id")

	cmd.Flags().Uint32P("value", "v", 1, "vote value: 1:yes,2:no")
	cmd.MarkFlagRequired("value")
}

func createNodeVoteTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	val, _ := cmd.Flags().GetUint32("value")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: 2, Id: id, Value: val}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func nodeVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "nodegroup nodes vote for new join node cmd",
		Run:   createNodeVoteTx,
	}
	addNodeVoteFlags(cmd)
	return cmd
}

func addNodeQuitFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "target quit addr")
	cmd.MarkFlagRequired("addr")

}

func createNodeQuitTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	opAddr, _ := cmd.Flags().GetString("addr")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: 3, Addr: opAddr}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func nodeQuitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "super node apply for quit nodegroup cmd",
		Run:   createNodeQuitTx,
	}
	addNodeQuitFlags(cmd)
	return cmd
}

func addNodeCancelFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	cmd.MarkFlagRequired("id")

}

func createNodeCancelTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: 4, Id: id}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func nodeCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "super node cancel join or quit action by id cmd",
		Run:   createNodeCancelTx,
	}
	addNodeCancelFlags(cmd)
	return cmd
}

// getNodeInfoCmd get node current status
func getNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr_status",
		Short: "Get node current status:10:joined,11:quited from nodegroup",
		Run:   nodeInfo,
	}
	addNodeBodyCmdFlags(cmd)
	return cmd
}

func addNodeBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "addr apply for super user")
	cmd.MarkFlagRequired("addr")

}

func nodeInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetNodeAddrInfo"
	req := pt.ReqParacrossNodeInfo{
		Title: paraName,
		Addr:  addr,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaNodeAddrIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// getNodeIDInfoCmd get node current status
func getNodeIDInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id_status",
		Short: "Get node id current vote status:1:joining,2:quiting,3:closed,4:canceled",
		Run:   nodeIDInfo,
	}
	addNodeIDBodyCmdFlags(cmd)
	return cmd
}

func addNodeIDBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "id apply for super user")
	cmd.MarkFlagRequired("id")

}

func nodeIDInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetNodeIDInfo"
	req := pt.ReqParacrossNodeInfo{
		Title: paraName,
		Id:    id,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaNodeIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// getNodeListCmd get node list by status
func getNodeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id_list",
		Short: "Get node apply id list info by status",
		Run:   nodeList,
	}
	addNodeListCmdFlags(cmd)
	return cmd
}

func addNodeListCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("status", "s", 0, "status:0:all,1:joining,2:quiting,3:closed,4:canceled")
	cmd.MarkFlagRequired("status")

}

func nodeList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	status, _ := cmd.Flags().GetInt32("status")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "ListNodeStatusInfo"
	req := pt.ReqParacrossNodeInfo{
		Title:  paraName,
		Status: status,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.RespParacrossNodeAddrs
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

func addSelfConsStageCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "g", 0, "height apply for self consensus enable or not ")
	cmd.MarkFlagRequired("height")

	cmd.Flags().Uint32P("enable", "e", 0, "if self consensus enable at height,1:enable,2:disable")
	cmd.MarkFlagRequired("enable")

}

func selfConsStage(cmd *cobra.Command, args []string) {
	height, _ := cmd.Flags().GetInt64("height")
	enable, _ := cmd.Flags().GetUint32("enable")
	paraName, _ := cmd.Flags().GetString("paraName")

	var config pt.ParaStageConfig
	config.Title = paraName
	config.Op = pt.ParaOpNewApply
	config.Value = &pt.ParaStageConfig_Stage{Stage: &pt.SelfConsensStage{StartHeight: height, Enable: enable}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "selfConsStageConfig",
		Payload:    types.MustPBToJSON(&config),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func selfConsStageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new",
		Short: "apply for para chain's self consensus stages cmd",
		Run:   selfConsStage,
	}
	addSelfConsStageCmdFlags(cmd)
	return cmd
}

func addVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	cmd.MarkFlagRequired("id")

	cmd.Flags().Uint32P("value", "v", 1, "vote value: 1:yes,2:no")
	cmd.MarkFlagRequired("value")
}

func createVoteTx(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	val, _ := cmd.Flags().GetUint32("value")
	paraName, _ := cmd.Flags().GetString("paraName")

	var config pt.ParaStageConfig
	config.Title = paraName
	config.Op = pt.ParaOpVote
	config.Value = &pt.ParaStageConfig_Vote{Vote: &pt.ConfigVoteInfo{Id: id, Value: val}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "selfConsStageConfig",
		Payload:    types.MustPBToJSON(&config),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func configVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote for config cmd",
		Run:   createVoteTx,
	}
	addVoteFlags(cmd)
	return cmd
}

func stageCancelTx(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	paraName, _ := cmd.Flags().GetString("paraName")

	var config pt.ParaStageConfig
	config.Title = paraName
	config.Op = pt.ParaOpCancel
	config.Value = &pt.ParaStageConfig_Cancel{Cancel: &pt.ConfigCancelInfo{Id: id}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "selfConsStageConfig",
		Payload:    types.MustPBToJSON(&config),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func configCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "cancel for config cmd",
		Run:   stageCancelTx,
	}
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	cmd.MarkFlagRequired("id")
	return cmd
}

func paraStageConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stages",
		Short: "self consensus stages config cmd",
	}
	cmd.AddCommand(selfConsStageCmd())
	cmd.AddCommand(configVoteCmd())
	cmd.AddCommand(configCancelCmd())
	cmd.AddCommand(QuerySelfStagesCmd())
	cmd.AddCommand(GetSelfConsStagesCmd())
	cmd.AddCommand(GetSelfConsOneStageCmd())

	return cmd
}

func paraConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "parachain config cmd",
	}
	cmd.AddCommand(paraStageConfigCmd())

	return cmd
}

func nodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup",
		Short: "super node group manage cmd",
	}
	cmd.AddCommand(nodeGroupApplyCmd())
	cmd.AddCommand(nodeGroupApproveCmd())
	cmd.AddCommand(nodeGroupQuitCmd())
	cmd.AddCommand(nodeGroupModifyCmd())

	cmd.AddCommand(getNodeGroupAddrsCmd())
	cmd.AddCommand(nodeGroupStatusCmd())
	cmd.AddCommand(nodeGroupListCmd())

	return cmd
}

func addNodeGroupApplyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addrs", "a", "", "addrs apply for super node,split by ',' ")
	cmd.MarkFlagRequired("addrs")

	cmd.Flags().Float64P("coins", "c", 0, "coins amount to frozen, not less config")
	cmd.MarkFlagRequired("coins")

}

func nodeGroupApply(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addrs, _ := cmd.Flags().GetString("addrs")
	coins, _ := cmd.Flags().GetFloat64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 1, Addrs: addrs, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nodeGroupApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply for para chain's super node group",
		Run:   nodeGroupApply,
	}
	addNodeGroupApplyCmdFlags(cmd)
	return cmd
}

func addNodeGroupApproveCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply id for nodegroup ")
	cmd.MarkFlagRequired("id")

	cmd.Flags().Float64P("coins", "c", 0, "coins amount to frozen, not less config")
	cmd.MarkFlagRequired("coins")

}

func nodeGroupApprove(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	coins, _ := cmd.Flags().GetFloat64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 2, Id: id, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nodeGroupApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve for para chain's super node group application",
		Run:   nodeGroupApprove,
	}
	addNodeGroupApproveCmdFlags(cmd)
	return cmd
}

func addNodeGroupQuitCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply quit id for nodegroup ")
	cmd.MarkFlagRequired("id")

}

func nodeGroupQuit(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 3, Id: id}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nodeGroupQuitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "quit for para chain's super node group application",
		Run:   nodeGroupQuit,
	}
	addNodeGroupQuitCmdFlags(cmd)
	return cmd
}

func addNodeGroupModifyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("coins", "c", 0, "modify coins amount to frozen, not less config")
	cmd.MarkFlagRequired("coins")

}

func nodeGroupModify(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	coins, _ := cmd.Flags().GetFloat64("coins")
	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 4, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func nodeGroupModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "modify for para chain's super node group parameters",
		Run:   nodeGroupModify,
	}
	addNodeGroupModifyCmdFlags(cmd)
	return cmd
}

// IsSyncCmd query parachain is sync
func IsSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "is_sync",
		Short: "query parachain is sync",
		Run:   isSync,
	}
	return cmd
}

func isSync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res bool
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.IsSync", nil, &res)
	ctx.Run()
}

func consusHeight(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetHeight"
	req := types.ReqString{Data: paraName}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParacrossConsensusStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetHeightCmd get para chain's chain height and consensus height
func GetHeightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "height",
		Short: "query consensus height",
		Run:   consusHeight,
	}
	return cmd
}

func addBlockBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "s", 0, "block start height")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Int64P("end", "e", 0, "block end height")
	cmd.MarkFlagRequired("end")
}

func blockInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	startH, _ := cmd.Flags().GetInt64("start")
	endH, _ := cmd.Flags().GetInt64("end")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetBlock2MainInfo"
	req := types.ReqBlocks{
		Start: startH,
		End:   endH,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaBlock2MainInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetBlockInfoCmd get blocks hash with main chain hash map
func GetBlockInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "Get blocks with main chain hash map between [start, end], the same in main",
		Run:   blockInfo,
	}
	addBlockBodyCmdFlags(cmd)
	return cmd
}

func addLocalBlockBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "t", 0, "block height,-1:latest height")
	cmd.MarkFlagRequired("start")

}

func localBlockInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	startH, _ := cmd.Flags().GetInt64("start")

	params := types.ReqInt{
		Height: startH,
	}
	var res pt.ParaLocalDbBlockInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetParaLocalBlockInfo", params, &res)
	ctx.Run()

}

// GetLocalBlockInfoCmd get blocks hash with main chain hash map
func GetLocalBlockInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffer_block",
		Short: "Get para download-level block info",
		Run:   localBlockInfo,
	}
	addLocalBlockBodyCmdFlags(cmd)
	return cmd
}

func addParaBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "g", 0, "height to para chain")
	cmd.MarkFlagRequired("height")

}

func paraInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	height, _ := cmd.Flags().GetInt64("height")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetTitleHeight"
	req := pt.ReqParacrossTitleHeight{
		Title:  paraName,
		Height: height,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParacrossHeightStatusRsp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetParaInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consens_status",
		Short: "Get para chain heights' consensus status",
		Run:   paraInfo,
	}
	addParaBodyCmdFlags(cmd)
	return cmd
}

// GetParaListCmd get para chain info list
func GetParaListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "para_list",
		Short: "Get para chain info list by titles",
		Run:   paraList,
	}

	return cmd
}

func paraList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "ListTitles"
	req := types.ReqNil{}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.RespParacrossTitles
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// getNodeGroupCmd get node group addr
func getNodeGroupAddrsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrs",
		Short: "Query super node group's current addrs by title",
		Run:   nodeGroup,
	}
	return cmd
}

func addParaAssetTranCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "asset transfer tx hash")
	cmd.MarkFlagRequired("hash")

}

func paraAssetTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	hash, _ := cmd.Flags().GetString("hash")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetAssetTxResult"
	req := types.ReqString{
		Data: hash,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParacrossAssetRsp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaAssetTransCmd get para chain asset transfer info
func GetParaAssetTransCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asset_txinfo",
		Short: "Get para chain cross asset transfer info",
		Run:   paraAssetTransfer,
	}
	addParaAssetTranCmdFlags(cmd)
	return cmd
}

func nodeGroup(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetNodeGroupAddrs"
	req := pt.ReqParacrossNodeInfo{Title: paraName}
	params.Payload = types.MustPBToJSON(&req)

	var res types.ReplyConfig
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// nodeGroupStatusCmd get node group addr
func nodeGroupStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "query super node group apply status by title",
		Run:   nodeGroupStatus,
	}
	return cmd
}

func nodeGroupStatus(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetNodeGroupStatus"
	req := pt.ReqParacrossNodeInfo{
		Title: paraName,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaNodeGroupStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// nodeGroupListCmd get node group addr
func nodeGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "query super node group apply list by status",
		Run:   nodeGroupList,
	}
	getNodeGroupListCmdFlags(cmd)
	return cmd
}

func getNodeGroupListCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("status", "s", 0, "status:1:apply,2:approve,3:quit")
	cmd.MarkFlagRequired("status")
}

func nodeGroupList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	status, _ := cmd.Flags().GetInt32("status")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "ListNodeGroupStatus"
	req := pt.ReqParacrossNodeInfo{
		Status: status,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.RespParacrossNodeGroups
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

func stagesInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetSelfConsStages"
	req := types.ReqNil{}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.SelfConsensStages
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetSelfConsStagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "Get para chain self consensus stages",
		Run:   stagesInfo,
	}

	return cmd
}

func stageOneInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetInt64("height")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetSelfConsOneStage"
	req := types.Int64{Data: height}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.SelfConsensStage
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetSelfConsOneStageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "query para chain one self consensus stage",
		Run:   stageOneInfo,
	}
	cmd.Flags().Int64P("height", "g", 0, "height to para chain")
	cmd.MarkFlagRequired("height")
	return cmd
}

// QuerySelfStagesCmd 显示提案查询信息
func QuerySelfStagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "show self consensus stage apply info",
		Run:   showSelfStages,
	}
	addShowSelfStagesflags(cmd)
	return cmd
}

func addShowSelfStagesflags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "q", "", "stage apply ID")
	cmd.Flags().Uint32P("status", "s", 0, "status:1:applying,3:closed,4:canceled,5:voting")
	cmd.Flags().Int32P("count", "c", 1, "count, default is 1")
	cmd.Flags().Int32P("direction", "d", 0, "direction, default is reserve")
	cmd.Flags().Int64P("height", "t", -1, "height, default is -1")
	cmd.Flags().Int32P("index", "i", -1, "index, default is -1")
}

func showSelfStages(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	id, _ := cmd.Flags().GetString("id")
	status, _ := cmd.Flags().GetUint32("status")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	height, _ := cmd.Flags().GetInt64("height")
	index, _ := cmd.Flags().GetInt32("index")

	if id == "" && status == 0 {
		fmt.Fprintln(os.Stderr, "should fill id or status in")
		return
	}

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "ListSelfStages"
	req := pt.ReqQuerySelfStages{
		Status:    status,
		Id:        id,
		Count:     count,
		Direction: direction,
		Height:    height,
		Index:     index,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ReplyQuerySelfStages
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

func addConsensDoneCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "g", 0, "height to para chain")
	cmd.MarkFlagRequired("height")

}

func consensDoneInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	height, _ := cmd.Flags().GetInt64("height")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetDoneTitleHeight"
	req := pt.ReqParacrossTitleHeight{
		Title:  paraName,
		Height: height,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.RespParacrossDone
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// GetConsensDoneInfoCmd get para chain done height consens info
func GetConsensDoneInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consens_done",
		Short: "Get para chain done height consensus info",
		Run:   consensDoneInfo,
	}
	addConsensDoneCmdFlags(cmd)
	return cmd
}

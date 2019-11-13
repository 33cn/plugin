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
		GetParaInfoCmd(),
		GetParaListCmd(),
		IsSyncCmd(),
		GetHeightCmd(),
		GetBlockInfoCmd(),
		GetLocalBlockInfoCmd(),
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

	cmd.Flags().StringP("ptitle", "", "", "the title of para chain, like `user.p.guodun.`")
	cmd.MarkFlagRequired("ptitle")

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

	cmd.Flags().StringP("ptitle", "", "", "the title of para chain, like `user.p.guodun.`")
	cmd.MarkFlagRequired("ptitle")

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
	cfg := types.GetCliSysParam(title)

	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0 {
		return "", types.ErrAmount
	}
	amountInt64 := int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4

	toAddr, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	ptitle, _ := cmd.Flags().GetString("ptitle")
	if !strings.HasPrefix(ptitle, "user.p") {
		fmt.Fprintln(os.Stderr, "ptitle is not right, title format like `user.p.guodun.`")
		return "", types.ErrInvalidParam
	}
	execName := ptitle + pt.ParaX

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

	params := pt.ReqParacrossNodeInfo{
		Title: paraName,
		Addr:  addr,
	}
	var res pt.ParaNodeAddrIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeAddrStatus", params, &res)
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

	params := pt.ReqParacrossNodeInfo{
		Title: paraName,
		Id:    id,
	}
	var res pt.ParaNodeIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeIDStatus", params, &res)
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

	params := pt.ReqParacrossNodeInfo{
		Title:  paraName,
		Status: status,
	}
	var res pt.RespParacrossNodeAddrs
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.ListNodeStatus", params, &res)
	ctx.Run()
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

	var res pt.ParacrossConsensusStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetHeight", &types.ReqString{Data: paraName}, &res)
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

	params := types.ReqBlocks{
		Start: startH,
		End:   endH,
	}
	var res pt.ParaBlock2MainInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetBlock2MainInfo", params, &res)
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

	params := pt.ReqParacrossTitleHeight{
		Title:  paraName,
		Height: height,
	}
	var res pt.ParacrossHeightStatusRsp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetTitleHeight", params, &res)
	ctx.Run()
}

// GetParaInfoCmd get para chain status by height
func GetParaInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consens_status",
		Short: "Get para chain current consensus status",
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

	var res pt.RespParacrossTitles
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.ListTitles", types.ReqNil{}, &res)
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

func nodeGroup(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	var res types.ReplyConfig
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeGroupAddrs", pt.ReqParacrossNodeInfo{Title: paraName}, &res)
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

	params := pt.ReqParacrossNodeInfo{
		Title: paraName,
	}

	var res pt.ParaNodeGroupStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeGroupStatus", params, &res)
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

	params := pt.ReqParacrossNodeInfo{
		Status: status,
	}

	var res pt.RespParacrossNodeGroups
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.ListNodeGroupStatus", params, &res)
	ctx.Run()
}

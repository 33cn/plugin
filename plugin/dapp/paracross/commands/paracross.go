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
		CreateRawNodeManageCmd(),
		CreateNodeGroupApplyCmd(),
		GetParaInfoCmd(),
		GetParaListCmd(),
		GetNodeGroupCmd(),
		GetNodeInfoCmd(),
		GetNodeIDInfoCmd(),
		GetNodeListCmd(),
		NodeGroupStatusCmd(),
		NodeGroupListCmd(),
		IsSyncCmd(),
		GetHeightCmd(),
		GetBlockInfoCmd(),
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

	cmd.Flags().StringP("title", "", "", "the title of para chain, like `user.p.guodun.`")
	cmd.MarkFlagRequired("title")

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

	cmd.Flags().StringP("title", "", "", "the title of para chain, like `user.p.guodun.`")
	cmd.MarkFlagRequired("title")

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
	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0 {
		return "", types.ErrAmount
	}
	amountInt64 := int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4

	toAddr, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	title, _ := cmd.Flags().GetString("title")
	if !strings.HasPrefix(title, "user.p") {
		fmt.Fprintln(os.Stderr, "title is not right, title format like `user.p.guodun.`")
		return "", types.ErrInvalidParam
	}
	execName := title + pt.ParaX

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
	tx, err := pt.CreateRawAssetTransferTx(&param)
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

//CreateRawNodeManageCmd create super node mange tx
func CreateRawNodeManageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "Create a super node manage cmd",
		Run:   createNodeTx,
	}
	addNodeManageFlags(cmd)
	return cmd
}

func addNodeManageFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("operation", "o", 0, "operation:1:join,2:vote,3:quit,4:cancel")
	cmd.MarkFlagRequired("operation")

	cmd.Flags().StringP("addr", "a", "", "operating target addr[optional]")

	cmd.Flags().StringP("id", "i", "", "operating target id[optional]")

	cmd.Flags().Uint32P("value", "v", 1, "vote value: 1:yes,2:no")
	cmd.Flags().Float64P("coins_frozen", "c", 0, "frozen coins amount, should not less nodegroup's")
}

func createNodeTx(cmd *cobra.Command, args []string) {
	op, _ := cmd.Flags().GetUint32("operation")
	opAddr, _ := cmd.Flags().GetString("addr")
	id, _ := cmd.Flags().GetString("id")
	val, _ := cmd.Flags().GetUint32("value")
	coins, _ := cmd.Flags().GetFloat64("coins_frozen")

	payload := &pt.ParaNodeAddrConfig{Op: op, Id: id, Value: val, Addr: opAddr, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

// CreateNodeGroupApplyCmd get node group addr
func CreateNodeGroupApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup",
		Short: "apply for para chain's super node group",
		Run:   nodeGroupApply,
	}
	addNodeGroupApplyCmdFlags(cmd)
	return cmd
}

func addNodeGroupApplyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("operation", "o", 0, "operation:1:apply,2:approve,3:quit,4:modify")
	cmd.MarkFlagRequired("operation")

	cmd.Flags().StringP("id", "i", "", "apply id for nodegroup ")

	cmd.Flags().StringP("addrs", "a", "", "addrs apply for super node,split by ',' ")

	cmd.Flags().Float64P("coins_frozen", "c", 0, "coins amount to frozen, not less config")

}

func nodeGroupApply(cmd *cobra.Command, args []string) {
	op, _ := cmd.Flags().GetUint32("operation")
	addrs, _ := cmd.Flags().GetString("addrs")
	id, _ := cmd.Flags().GetString("id")
	coins, _ := cmd.Flags().GetFloat64("coins_frozen")

	payload := &pt.ParaNodeGroupConfig{Op: op, Id: id, Addrs: addrs, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
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

// GetHeightCmd get para chain's chain height and consensus height
func GetHeightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "height",
		Short: "query consensus height",
		Run:   consusHeight,
	}
	addTitleFlags(cmd)
	return cmd
}

func addTitleFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title, default null in para chain")
	cmd.MarkFlagRequired("title")
}

func consusHeight(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")

	var res pt.ParacrossConsensusStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetHeight", &types.ReqString{Data: title}, &res)
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

// GetParaInfoCmd get para chain status by height
func GetParaInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "para_status",
		Short: "Get para chain current status",
		Run:   paraInfo,
	}
	addParaBodyCmdFlags(cmd)
	return cmd
}

func addParaBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")

	cmd.Flags().Int64P("height", "g", 0, "height to para chain")
	cmd.MarkFlagRequired("height")

}

func paraInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")
	height, _ := cmd.Flags().GetInt64("height")

	params := pt.ReqParacrossTitleHeight{
		Title:  title,
		Height: height,
	}
	var res pt.ParacrossHeightStatusRsp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetTitleHeight", params, &res)
	ctx.Run()
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

// GetNodeInfoCmd get node current status
func GetNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node_addr_status",
		Short: "Get node current status:10:joined,11:quited from nodegroup",
		Run:   nodeInfo,
	}
	addNodeBodyCmdFlags(cmd)
	return cmd
}

func addNodeBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")

	cmd.Flags().StringP("addr", "a", "", "addr apply for super user")
	cmd.MarkFlagRequired("addr")

}

func nodeInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")
	addr, _ := cmd.Flags().GetString("addr")

	params := pt.ReqParacrossNodeInfo{
		Title: title,
		Addr:  addr,
	}
	var res pt.ParaNodeAddrIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeAddrStatus", params, &res)
	ctx.Run()
}

// GetNodeIDInfoCmd get node current status
func GetNodeIDInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node_id_status",
		Short: "Get node id current vote status:0:all,1:joining,2:quiting,3:closed,4:canceled",
		Run:   nodeIDInfo,
	}
	addNodeIDBodyCmdFlags(cmd)
	return cmd
}

func addNodeIDBodyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")

	cmd.Flags().StringP("id", "i", "", "id apply for super user")
	cmd.MarkFlagRequired("id")

}

func nodeIDInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")
	id, _ := cmd.Flags().GetString("id")

	params := pt.ReqParacrossNodeInfo{
		Title: title,
		Id:    id,
	}
	var res pt.ParaNodeIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeIDStatus", params, &res)
	ctx.Run()
}

// GetNodeListCmd get node list by status
func GetNodeListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node_list",
		Short: "Get node info list by status",
		Run:   nodeList,
	}
	addNodeListCmdFlags(cmd)
	return cmd
}

func addNodeListCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")

	cmd.Flags().Int32P("status", "s", 0, "status:1:adding,2:added,3:quiting,4:quited")
	cmd.MarkFlagRequired("status")

}

func nodeList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")
	status, _ := cmd.Flags().GetInt32("status")

	params := pt.ReqParacrossNodeInfo{
		Title:  title,
		Status: status,
	}
	var res pt.RespParacrossNodeAddrs
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.ListNodeStatus", params, &res)
	ctx.Run()
}

// GetNodeGroupCmd get node group addr
func GetNodeGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup_addrs",
		Short: "Get super node group's addrs by title",
		Run:   nodeGroup,
	}
	addNodeGroupCmdFlags(cmd)
	return cmd
}

func addNodeGroupCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")
}

func nodeGroup(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")

	var res types.ReplyConfig
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeGroupAddrs", pt.ReqParacrossNodeInfo{Title: title}, &res)
	ctx.Run()
}

// NodeGroupStatusCmd get node group addr
func NodeGroupStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup_status",
		Short: "query super node group apply status by title",
		Run:   nodeGroupStatus,
	}
	getNodeGroupStatusCmdFlags(cmd)
	return cmd
}

func getNodeGroupStatusCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("title", "t", "", "parallel chain's title")
	cmd.MarkFlagRequired("title")

}

func nodeGroupStatus(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	title, _ := cmd.Flags().GetString("title")

	params := pt.ReqParacrossNodeInfo{
		Title: title,
	}

	var res pt.ParaNodeGroupStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetNodeGroupStatus", params, &res)
	ctx.Run()
}

// NodeGroupListCmd get node group addr
func NodeGroupListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nodegroup_list",
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

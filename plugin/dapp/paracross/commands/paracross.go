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
	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/pkg/errors"
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
		CreateRawCrossAssetTransferCmd(),
		superNodeCmd(),
		nodeGroupCmd(),
		supervisionNodeCmd(),
		paraConfigCmd(),
		GetParaInfoCmd(),
		GetParaListCmd(),
		GetParaAssetTransCmd(),
		IsSyncCmd(),
		GetHeightCmd(),
		GetBlockInfoCmd(),
		GetLocalBlockInfoCmd(),
		GetConsensDoneInfoCmd(),
		blsCmd(),
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
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
}

func createAssetTransfer(cmd *cobra.Command, args []string) {
	txHex, err := createAssetTx(cmd, false)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
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
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("to", "t", "", "receiver account address")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
}

func createAssetWithdraw(cmd *cobra.Command, args []string) {
	txHex, err := createAssetTx(cmd, true)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(txHex)
}

func createAssetTx(cmd *cobra.Command, isWithdraw bool) (string, error) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return "", err
	}
	amount, _ := cmd.Flags().GetFloat64("amount")
	if amount < 0 {
		return "", types.ErrAmount
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.amount"))
		return "", err
	}

	toAddr, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	paraName, _ := cmd.Flags().GetString("paraName")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "title is not right, title format like `user.p.guodun.`")
		return "", types.ErrInvalidParam
	}

	param := types.CreateTx{
		To:          toAddr,
		Amount:      amountInt64,
		Fee:         0,
		Note:        []byte(note),
		IsWithdraw:  isWithdraw,
		IsToken:     false,
		TokenSymbol: symbol,
		ExecName:    types.GetExecName(pt.ParaX, paraName),
	}
	tx, err := pt.CreateRawAssetTransferTxExt(cfg.ChainID, cfg.MinTxFeeRate, &param)
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
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
	_ = cmd.MarkFlagRequired("symbol")
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
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "coins.bty", "default for bty, symbol for token")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	_ = cmd.MarkFlagRequired("exec")
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
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "default for bty, symbol for token")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "asset deposit exec")
	_ = cmd.MarkFlagRequired("exec")
}

func createWithdraw(cmd *cobra.Command, args []string) {
	commands.CreateAssetWithdraw(cmd, args, pt.ParaX)
}

// CreateRawCrossAssetTransferCmd create raw cross asset transfer tx
func CreateRawCrossAssetTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cross_transfer",
		Short: "Create a cross asset transfer transaction",
		Run:   createCrossAssetTransfer,
	}
	addCreateCrossAssetTransferFlags(cmd)
	return cmd
}

func addCreateCrossAssetTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "exec of asset resident")
	_ = cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("symbol", "s", "", "asset symbol like bty")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("to", "t", "", "transfer to account")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

}

func createCrossAssetTransfer(cmd *cobra.Command, args []string) {
	ty, _ := cmd.Flags().GetString("exec")
	toAddr, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetFloat64("amount")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	if amount < 0 {
		_, _ = fmt.Fprintln(os.Stderr, "amount < 0")
		return
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.amount"))
		return
	}

	paraName, _ := cmd.Flags().GetString("paraName")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	execName := paraName + pt.ParaX

	var config pt.CrossAssetTransfer
	config.AssetExec = ty
	config.AssetSymbol = symbol
	config.ToAddr = toAddr
	config.Note = note
	config.Amount = amountInt64

	params := &rpctypes.CreateTxIn{
		Execer:     execName,
		ActionName: "CrossAssetTransfer",
		Payload:    types.MustPBToJSON(&config),
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	_, err = ctx.RunResult()
	if err != nil {
		fmt.Println(err)
		return
	}
	//remove 0x
	fmt.Println(res[2:])

}

func superNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "super node manage cmd",
	}
	cmd.AddCommand(nodeJoinCmd())
	cmd.AddCommand(nodeVoteCmd())
	cmd.AddCommand(nodeQuitCmd())
	cmd.AddCommand(nodeCancelCmd())
	cmd.AddCommand(getNodeInfoCmd())
	cmd.AddCommand(getNodeIDInfoCmd())
	cmd.AddCommand(getNodeListCmd())
	cmd.AddCommand(nodeModifyCmd())

	cmd.AddCommand(nodeMinerCmd())
	return cmd
}

func nodeMinerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "miner",
		Short: "super node bind miner cmd",
	}

	cmd.AddCommand(nodeBindCmd())
	cmd.AddCommand(getNodeBindListCmd())
	cmd.AddCommand(getMinerBindListCmd())
	return cmd
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

func addNodeJoinFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "target join addr")
	_ = cmd.MarkFlagRequired("addr")

	cmd.Flags().Float64P("coins", "c", 0, "frozen coins amount, should not less nodegroup's setting")

}

func createNodeJoinTx(cmd *cobra.Command, args []string) {
	opAddr, _ := cmd.Flags().GetString("addr")
	coins, _ := cmd.Flags().GetFloat64("coins")
	paraName, _ := cmd.Flags().GetString("paraName")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	coinsInt64, err := types.FormatFloatDisplay2Value(coins, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.coins"))
		return
	}

	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: 1, Addr: opAddr, CoinsFrozen: coinsInt64}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

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

func addNodeVoteFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().Uint32P("value", "v", 1, "vote value: 1:yes,2:no")
	_ = cmd.MarkFlagRequired("value")
}

func createNodeVoteTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	val, _ := cmd.Flags().GetUint32("value")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
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

func nodeQuitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "super node apply for quit nodegroup cmd",
		Run:   createNodeQuitTx,
	}
	addNodeQuitFlags(cmd)
	return cmd
}

func addNodeQuitFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "target quit addr")
	_ = cmd.MarkFlagRequired("addr")

}

func createNodeQuitTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	opAddr, _ := cmd.Flags().GetString("addr")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
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

func nodeCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "super node cancel join or quit action by id cmd",
		Run:   createNodeCancelTx,
	}
	addNodeCancelFlags(cmd)
	return cmd
}

func addNodeCancelFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("id")

}

func createNodeCancelTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
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

func nodeModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "super node modify parameters",
		Run:   createNodeModifyTx,
	}
	addNodeModifyFlags(cmd)
	return cmd
}

func addNodeModifyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("addr")
	cmd.Flags().StringP("pubkey", "p", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("pubkey")

}

func createNodeModifyTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: pt.ParaOpModify, Addr: addr, BlsPubKey: pubkey}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func nodeBindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bind",
		Short: "bind miner for specific account",
		Run:   createNodeBindTx,
	}
	addNodeBindFlags(cmd)
	return cmd
}

func addNodeBindFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("action", "a", 1, "action bind:1, unbind:2, modify:3")
	_ = cmd.MarkFlagRequired("action")

	cmd.Flags().Uint64P("coins", "c", 0, "bind coins, unbind not needed")

	cmd.Flags().StringP("node", "n", "", "target node to bind/unbind miner")
	_ = cmd.MarkFlagRequired("node")

}

func createNodeBindTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	action, _ := cmd.Flags().GetUint32("action")
	node, _ := cmd.Flags().GetString("node")
	coins, _ := cmd.Flags().GetUint64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	if action == 1 && coins == 0 {
		_, _ = fmt.Fprintln(os.Stderr, "coins should bigger than 0")
	}

	payload := &pt.ParaBindMinerCmd{BindAction: int32(action), BindCoins: int64(coins), TargetNode: node}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "ParaBindMiner",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()

}

func getNodeBindListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "miner_list",
		Short: "Get node bind miner account list",
		Run:   minerBindInfo,
	}
	addMinerBindCmdFlags(cmd)
	return cmd
}

func addMinerBindCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("node", "n", "", "super node addr to bind miner")
	cmd.MarkFlagRequired("node")

	cmd.Flags().StringP("miner", "m", "", "bind miner addr")
	cmd.Flags().BoolP("unbind", "u", false, "query with unbinded miner,default false")

}

func minerBindInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	node, _ := cmd.Flags().GetString("node")
	miner, _ := cmd.Flags().GetString("miner")
	unbind, _ := cmd.Flags().GetBool("unbind")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetNodeBindMinerList"

	params.Payload = types.MustPBToJSON(&pt.ParaNodeMinerListReq{Node: node, Miner: miner, WithUnBind: unbind})

	var res pt.ParaBindMinerList
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

func getMinerBindListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node_list",
		Short: "Get miner bind consensus node account list",
		Run:   nodeBindInfo,
	}
	addNodeBindCmdFlags(cmd)
	return cmd
}

func addNodeBindCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("miner", "m", "", "bind miner addr")
	cmd.MarkFlagRequired("miner")

	cmd.Flags().BoolP("unbind", "u", false, "query with unbinded miner,default false")
}

func nodeBindInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	miner, _ := cmd.Flags().GetString("miner")
	unbind, _ := cmd.Flags().GetBool("unbind")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetMinerBindNodeList"

	params.Payload = types.MustPBToJSON(&pt.ParaNodeMinerListReq{Miner: miner, WithUnBind: unbind})

	var res types.ReplyStrings
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
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
	_ = cmd.MarkFlagRequired("addr")

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
	_ = cmd.MarkFlagRequired("id")

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
	_ = cmd.MarkFlagRequired("status")

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
	_ = cmd.MarkFlagRequired("height")

	cmd.Flags().Uint32P("enable", "e", 0, "if self consensus enable at height,1:enable,2:disable")
	_ = cmd.MarkFlagRequired("enable")

}

func selfConsStage(cmd *cobra.Command, args []string) {
	height, _ := cmd.Flags().GetInt64("height")
	enable, _ := cmd.Flags().GetUint32("enable")
	paraName, _ := cmd.Flags().GetString("paraName")

	var config pt.ParaStageConfig
	config.Title = paraName
	config.Ty = pt.ParaOpNewApply
	config.Value = &pt.ParaStageConfig_Stage{Stage: &pt.SelfConsensStage{StartHeight: height, Enable: enable}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SelfStageConfig",
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
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().Uint32P("value", "v", 1, "vote value: 1:yes,2:no")
	_ = cmd.MarkFlagRequired("value")
}

func createVoteTx(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	val, _ := cmd.Flags().GetUint32("value")
	paraName, _ := cmd.Flags().GetString("paraName")

	var config pt.ParaStageConfig
	config.Title = paraName
	config.Ty = pt.ParaOpVote
	config.Value = &pt.ParaStageConfig_Vote{Vote: &pt.ConfigVoteInfo{Id: id, Value: val}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SelfStageConfig",
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
	config.Ty = pt.ParaOpCancel
	config.Value = &pt.ParaStageConfig_Cancel{Cancel: &pt.ConfigCancelInfo{Id: id}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SelfStageConfig",
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
	_ = cmd.MarkFlagRequired("id")
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
	cmd.AddCommand(issueCoinsCmd())

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

func nodeGroupApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply for para chain's super node group",
		Run:   nodeGroupApply,
	}
	addNodeGroupApplyCmdFlags(cmd)
	return cmd
}

func addNodeGroupApplyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addrs", "a", "", "addrs apply for super node,split by ',' ")
	_ = cmd.MarkFlagRequired("addrs")

	cmd.Flags().StringP("blspubs", "p", "", "bls sign pub key for addr's private key,split by ',' (optional)")

	cmd.Flags().Float64P("coins", "c", 0, "coins amount to frozen, not less config")

}

func nodeGroupApply(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addrs, _ := cmd.Flags().GetString("addrs")
	blspubs, _ := cmd.Flags().GetString("blspubs")
	coins, _ := cmd.Flags().GetFloat64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	coinsInt64, err := types.FormatFloatDisplay2Value(coins, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.coins"))
		return
	}

	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 1, Addrs: addrs, BlsPubKeys: blspubs, CoinsFrozen: coinsInt64}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

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

func addNodeGroupApproveCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply id for nodegroup ")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("autonomyId", "a", "", "optional: autonomy approved id ")

	cmd.Flags().Float64P("coins", "c", 0, "optional: coins amount to frozen, not less config")

}

func nodeGroupApprove(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	coins, _ := cmd.Flags().GetFloat64("coins")
	autonomyId, _ := cmd.Flags().GetString("autonomyId")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	coinsInt64, err := types.FormatFloatDisplay2Value(coins, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.coins"))
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 2, Id: id, CoinsFrozen: coinsInt64, AutonomyItemID: autonomyId}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

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

func addNodeGroupQuitCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply quit id for nodegroup ")
	_ = cmd.MarkFlagRequired("id")

}

func nodeGroupQuit(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
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

func nodeGroupModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "modify for para chain's super node group parameters",
		Run:   nodeGroupModify,
	}
	addNodeGroupModifyCmdFlags(cmd)
	return cmd
}

func addNodeGroupModifyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("coins", "c", 0, "modify coins amount to frozen, not less config")
	_ = cmd.MarkFlagRequired("coins")

}

func nodeGroupModify(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	coins, _ := cmd.Flags().GetFloat64("coins")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	coinsInt64, err := types.FormatFloatDisplay2Value(coins, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.coins"))
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 4, CoinsFrozen: coinsInt64}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "NodeGroupConfig",
		Payload:    types.MustPBToJSON(payload),
	}

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

func blsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bls",
		Short: "bls sign manager cmd",
	}
	cmd.AddCommand(leaderCmd())
	cmd.AddCommand(cmtTxInfoCmd())
	cmd.AddCommand(blsPubKeyCmd())
	return cmd
}

// leaderCmd query parachain is sync
func leaderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "leader",
		Short: "current bls sign leader",
		Run:   leaderInfo,
	}
	return cmd
}

func leaderInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res pt.ElectionStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetParaNodeLeaderInfo", nil, &res)
	ctx.Run()
}

// cmtTxInfoCmd query parachain is sync
func cmtTxInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cmts",
		Short: "current bls sign commits info",
		Run:   cmtTxInfo,
	}
	return cmd
}

func cmtTxInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res pt.ParaBlsSignSumInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "paracross.GetParaCmtTxInfo", nil, &res)
	ctx.Run()
}

// cmtTxInfoCmd query parachain is sync
func blsPubKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pub",
		Short: "get bls pub key from secp256k1 prikey",
		Run:   blsPubKey,
	}
	cmd.Flags().StringP("prikey", "p", "", "secp256k1 private hex key")
	return cmd
}

func blsPubKey(cmd *cobra.Command, args []string) {

	prikey, _ := cmd.Flags().GetString("prikey")

	blsPub, err := getBlsPubFromSecp256Key(prikey)
	if prikey == "" {
		fmt.Fprintln(os.Stderr, "must input valid secp256k1 prikey, err:"+err.Error())
		return
	}
	fmt.Println("blsPub:", blsPub)
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
	_ = cmd.MarkFlagRequired("start")

	cmd.Flags().Int64P("end", "e", 0, "block end height")
	_ = cmd.MarkFlagRequired("end")
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
	_ = cmd.MarkFlagRequired("start")
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
	_ = cmd.MarkFlagRequired("height")
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
	_ = cmd.MarkFlagRequired("hash")

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

	var res pt.ParacrossAsset
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
	_ = cmd.MarkFlagRequired("status")
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

// GetSelfConsStagesCmd get para chain status by height
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

// GetSelfConsOneStageCmd get para chain status by height
func GetSelfConsOneStageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "one",
		Short: "query para chain one self consensus stage",
		Run:   stageOneInfo,
	}
	cmd.Flags().Int64P("height", "g", 0, "height to para chain")
	_ = cmd.MarkFlagRequired("height")
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
		_, _ = fmt.Fprintln(os.Stderr, "should fill id or status in")
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
	_ = cmd.MarkFlagRequired("height")

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

func issueCoinsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "issue new coins by super manager",
		Run:   createIssueCoinsTx,
	}
	addIssueCoinsFlags(cmd)
	return cmd
}

func addIssueCoinsFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("amount", "a", 0, "new issue amount")
	cmd.MarkFlagRequired("amount")
}

func createIssueCoinsTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	coins, _ := cmd.Flags().GetUint64("amount")

	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	if coins == 0 {
		fmt.Fprintln(os.Stderr, "coins should bigger than 0")
	}

	payload := &pt.ParacrossMinerAction{AddIssueCoins: int64(coins)}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "Miner",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func supervisionNodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "supervision_node",
		Short: "supervision node manage cmd",
	}
	cmd.AddCommand(supervisionNodeApplyCmd())
	cmd.AddCommand(supervisionNodeApproveCmd())
	cmd.AddCommand(supervisionNodeQuitCmd())
	cmd.AddCommand(supervisionNodeCancelCmd())

	cmd.AddCommand(getSupervisionNodeGroupAddrsCmd())
	cmd.AddCommand(supervisionNodeListInfoCmd())
	cmd.AddCommand(getSupervisionNodeInfoCmd())
	cmd.AddCommand(supervisionNodeModifyCmd())

	return cmd
}

func supervisionNodeApplyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "apply for para chain's supervision node",
		Run:   supervisionNodeApply,
	}
	addSupervisionNodeApplyCmdFlags(cmd)
	return cmd
}

func addSupervisionNodeApplyCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "addr apply for supervision node")
	_ = cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("blspub", "p", "", "bls sign pub key for addr's private key")

	cmd.Flags().Float64P("coins", "c", 0, "coins amount to frozen, not less config")
}

func supervisionNodeApply(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	blspub, _ := cmd.Flags().GetString("blspub")
	coins, _ := cmd.Flags().GetFloat64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 1, Addrs: addr, BlsPubKeys: blspub, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SupervisionNodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func supervisionNodeApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve for para chain's supervision node application",
		Run:   supervisionNodeApprove,
	}
	addSupervisionNodeApproveCmdFlags(cmd)
	return cmd
}

func addSupervisionNodeApproveCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply id for supervision node ")
	_ = cmd.MarkFlagRequired("id")

	cmd.Flags().StringP("autonomyId", "a", "", "autonomy approved id ")
	_ = cmd.MarkFlagRequired("autonomyId")

	cmd.Flags().Float64P("coins", "c", 0, "coins amount to frozen, not less config")
}

func supervisionNodeApprove(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	autonomyId, _ := cmd.Flags().GetString("autonomyId")
	coins, _ := cmd.Flags().GetFloat64("coins")

	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 2, Id: id, AutonomyItemID: autonomyId, CoinsFrozen: int64(math.Trunc((coins+0.0000001)*1e4)) * 1e4}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SupervisionNodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func supervisionNodeQuitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "quit",
		Short: "quit for para chain's supervision node application",
		Run:   supervisionNodeQuit,
	}
	addSupervisionNodeQuitCmdFlags(cmd)
	return cmd
}

func addSupervisionNodeQuitCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "apply quit id for supervision node")
	_ = cmd.MarkFlagRequired("addr")
}

func supervisionNodeQuit(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	opAddr, _ := cmd.Flags().GetString("addr")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 3, Addrs: opAddr}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SupervisionNodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func supervisionNodeCancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "cancel for para chain's supervision node application",
		Run:   supervisionNodeCancel,
	}
	addSupervisionNodeCancelCmdFlags(cmd)
	return cmd
}

func addSupervisionNodeCancelCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "apply cancel id for supervision node")
	_ = cmd.MarkFlagRequired("id")
}

func supervisionNodeCancel(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	id, _ := cmd.Flags().GetString("id")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeGroupConfig{Title: paraName, Op: 4, Id: id}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SupervisionNodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getSupervisionNodeGroupAddrsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addrs",
		Short: "Query supervision node group's current addrs by title",
		Run:   supervisionNodeGroup,
	}
	return cmd
}

func supervisionNodeGroup(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetSupervisionNodeGroupAddrs"
	req := pt.ReqParacrossNodeInfo{Title: paraName}
	params.Payload = types.MustPBToJSON(&req)

	var res types.ReplyConfig
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// supervisionNodeListInfoCmd get node list by status
func supervisionNodeListInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id_list",
		Short: "Get supervision node apply id list info by status",
		Run:   supervisionNodeListInfo,
	}
	getSupervisionNodeListInfoCmdFlags(cmd)
	return cmd
}

func getSupervisionNodeListInfoCmdFlags(cmd *cobra.Command) {
	cmd.Flags().Int32P("status", "s", 0, "status:0:all,1:joining,2:quiting,3:closed,4:canceled")
	_ = cmd.MarkFlagRequired("status")

}

func supervisionNodeListInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	status, _ := cmd.Flags().GetInt32("status")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "ListSupervisionNodeStatusInfo"
	req := pt.ReqParacrossNodeInfo{
		Title:  paraName,
		Status: status,
	}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.RespParacrossNodeGroups
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// getNodeInfoCmd get node current status
func getSupervisionNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addr_status",
		Short: "Get node current status:2:approve, 3:quit from supervision group",
		Run:   supervisionNodeInfo,
	}
	addSupervisionNodeInfoCmdFlags(cmd)
	return cmd
}

func addSupervisionNodeInfoCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "addr apply for super user")
	_ = cmd.MarkFlagRequired("addr")

}

func supervisionNodeInfo(cmd *cobra.Command, args []string) {
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

func supervisionNodeModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "supervision node modify parameters",
		Run:   createSupervisionNodeModifyTx,
	}
	addSupervisionNodeModifyFlags(cmd)
	return cmd
}

func addSupervisionNodeModifyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("addr")
	cmd.Flags().StringP("pubkey", "p", "", "operating target apply id")
	_ = cmd.MarkFlagRequired("pubkey")

}

func createSupervisionNodeModifyTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	pubkey, _ := cmd.Flags().GetString("pubkey")
	if !strings.HasPrefix(paraName, "user.p") {
		_, _ = fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}
	payload := &pt.ParaNodeAddrConfig{Title: paraName, Op: pt.ParacrossSupervisionNodeModify, Addr: addr, BlsPubKey: pubkey}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SupervisionNodeConfig",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

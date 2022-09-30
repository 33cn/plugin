// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpcTypes "github.com/33cn/chain33/rpc/types"
	commandtypes "github.com/33cn/chain33/system/dapp/commands/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/spf13/cobra"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/types"
)

// CreateAssetSendToExec 通用的创建 send_exec 交易， 额外指定资产合约
func CreateAssetSendToExec(cmd *cobra.Command, args []string, fromExec string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	exec, _ := cmd.Flags().GetString("exec")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	exec = GetRealExecName(paraName, exec)
	to, err := GetExecAddr(exec, cfg.DefaultAddressID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	amount, _ := cmd.Flags().GetFloat64("amount")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	payload := &types.AssetsTransferToExec{
		To:        to,
		Amount:    amountInt64,
		Note:      []byte(note),
		Cointoken: symbol,
		ExecName:  exec,
	}

	params := &rpcTypes.CreateTxIn{
		Execer:     types.GetExecName(fromExec, paraName),
		ActionName: "TransferToExec",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateAssetWithdraw 通用的创建 withdraw 交易， 额外指定资产合约
func CreateAssetWithdraw(cmd *cobra.Command, args []string, fromExec string) {
	exec, _ := cmd.Flags().GetString("exec")
	paraName, _ := cmd.Flags().GetString("paraName")
	exec = GetRealExecName(paraName, exec)
	amount, _ := cmd.Flags().GetFloat64("amount")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	exec = GetRealExecName(paraName, exec)
	execAddr, err := GetExecAddr(exec, cfg.DefaultAddressID)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	payload := &types.AssetsWithdraw{
		To:        execAddr,
		Amount:    amountInt64,
		Note:      []byte(note),
		Cointoken: symbol,
		ExecName:  exec,
	}
	params := &rpcTypes.CreateTxIn{
		Execer:     types.GetExecName(fromExec, paraName),
		ActionName: "Withdraw",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateAssetTransfer 通用的创建 transfer 交易， 额外指定资产合约
func CreateAssetTransfer(cmd *cobra.Command, args []string, fromExec string) {
	toAddr, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetFloat64("amount")
	note, _ := cmd.Flags().GetString("note")
	symbol, _ := cmd.Flags().GetString("symbol")
	paraName, _ := cmd.Flags().GetString("paraName")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	cfg, err := commandtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	payload := &types.AssetsTransfer{
		To:        toAddr,
		Amount:    amountInt64,
		Note:      []byte(note),
		Cointoken: symbol,
	}
	params := &rpcTypes.CreateTxIn{
		Execer:     types.GetExecName(fromExec, paraName),
		ActionName: "Transfer",
		Payload:    types.MustPBToJSON(payload),
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// GetExecAddr 获取执行器地址
func GetExecAddr(exec string, addressType int32) (string, error) {
	if ok := types.IsAllowExecName([]byte(exec), []byte(exec)); !ok {
		return "", types.ErrExecNameNotAllow
	}

	addrResult, err := address.GetExecAddress(exec, addressType)
	if err != nil {
		return "", err
	}
	return addrResult, nil
}

func GetRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, pt.ParaPrefix) {
		return name
	}
	return paraName + name
}

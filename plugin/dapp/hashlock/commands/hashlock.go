// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
	"github.com/spf13/cobra"
)

// HashlockCmd cmds
func HashlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hashlock",
		Short: "Hashlock operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		HashlockLockCmd(),
		HashlockUnlockCmd(),
		HashlockSendCmd(),
	)

	return cmd
}

// HashlockLockCmd construct lock tx
func HashlockLockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "Create hashlock lock transaction",
		Run:   hashlockLockCmd,
	}
	addHashlockLockCmdFlags(cmd)
	return cmd
}

func addHashlockLockCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("secret", "s", "", "secret information")
	cmd.MarkFlagRequired("secret")
	cmd.Flags().Float64P("amount", "a", 0.0, "locking amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().Int64P("delay", "d", 60, "delay period (minimum 60 seconds)")
	cmd.MarkFlagRequired("delay")
	cmd.Flags().StringP("to", "t", "", "to address")
	cmd.MarkFlagRequired("to")
	cmd.Flags().StringP("return", "r", "", "return address")
	cmd.MarkFlagRequired("return")

	cmd.Flags().Float64P("fee", "f", 0.0, "transaction fee")
}

func hashlockLockCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")
	toAddr, _ := cmd.Flags().GetString("to")
	returnAddr, _ := cmd.Flags().GetString("return")
	delay, _ := cmd.Flags().GetInt64("delay")
	amount, _ := cmd.Flags().GetFloat64("amount")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	fee, _ := cmd.Flags().GetFloat64("fee")

	if delay < 60 {
		fmt.Println("delay period changed to 60")
		delay = 60
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.amount"))
		return
	}
	feeInt64, err := types.FormatFloatDisplay2Value(fee, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.fee"))
		return
	}
	if feeInt64 < cfg.MinTxFeeRate {
		feeInt64 = cfg.MinTxFeeRate
	}
	params := pty.HashlockLockTx{
		Secret:     secret,
		Amount:     amountInt64,
		Time:       delay,
		ToAddr:     toAddr,
		ReturnAddr: returnAddr,
		Fee:        feeInt64,
	}

	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}

	paramWithExecAction := rpctypes.CreateTxIn{
		Execer:     "hashlock",
		ActionName: "HashlockLock",
		Payload:    payLoad,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", paramWithExecAction, nil)
	ctx.RunWithoutMarshal()
}

// HashlockUnlockCmd construct unlock tx
func HashlockUnlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "Create hashlock unlock transaction",
		Run:   hashlockUnlockCmd,
	}
	addHashlockCmdFlags(cmd)
	return cmd
}

func addHashlockCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("secret", "s", "", "secret information")
	cmd.MarkFlagRequired("secret")

	cmd.Flags().Float64P("fee", "f", 0.0, "transaction fee")
}

func hashlockUnlockCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64, err := types.FormatFloatDisplay2Value(fee, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.fee"))
		return
	}
	if feeInt64 < cfg.MinTxFeeRate {
		feeInt64 = cfg.MinTxFeeRate
	}

	params := pty.HashlockUnlockTx{
		Secret: secret,
		Fee:    feeInt64,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}

	paramWithExecAction := rpctypes.CreateTxIn{
		Execer:     "hashlock",
		ActionName: "HashlockUnlock",
		Payload:    payLoad,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", paramWithExecAction, nil)
	ctx.RunWithoutMarshal()
}

// HashlockSendCmd construct send tx
func HashlockSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Create hashlock send transaction",
		Run:   hashlockSendCmd,
	}
	addHashlockCmdFlags(cmd)
	return cmd
}

func hashlockSendCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	secret, _ := cmd.Flags().GetString("secret")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	fee, _ := cmd.Flags().GetFloat64("fee")
	feeInt64, err := types.FormatFloatDisplay2Value(fee, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.fee"))
		return
	}
	if feeInt64 < cfg.MinTxFeeRate {
		feeInt64 = cfg.MinTxFeeRate
	}
	params := pty.HashlockSendTx{
		Secret: secret,
		Fee:    feeInt64,
	}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return
	}

	paramWithExecAction := rpctypes.CreateTxIn{
		Execer:     "hashlock",
		ActionName: "HashlockSend",
		Payload:    payLoad,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", paramWithExecAction, nil)
	ctx.RunWithoutMarshal()
}

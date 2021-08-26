// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"os"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/retrieve/rpc"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
	"github.com/spf13/cobra"
)

// RetrieveResult response
type RetrieveResult struct {
	DelayPeriod int64 `json:"delayPeriod"`
	//RemainTime  int64  `json:"remainTime"`
	Status string `json:"status"`
}

// RetrieveCmd cmds
func RetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retrieve",
		Short: "Wallet retrieve operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		BackupCmd(),
		PrepareCmd(),
		PerformCmd(),
		CancelCmd(),
		RetrieveQueryCmd(),
	)

	return cmd
}

// BackupCmd construct backup tx
func BackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup the wallet",
		Run:   backupCmd,
	}
	addBakupCmdFlags(cmd)
	return cmd
}

func addBakupCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("backup", "b", "", "backup address")
	cmd.MarkFlagRequired("backup")
	cmd.Flags().StringP("default", "t", "", "default address")
	cmd.MarkFlagRequired("default")
	cmd.Flags().Int64P("delay", "d", 60, "delay period (minimum 60 seconds)")
	cmd.MarkFlagRequired("delay")

	cmd.Flags().Float64P("fee", "f", 0.0, "transaction fee")
}

func backupCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")
	delay, _ := cmd.Flags().GetInt64("delay")

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

	if delay < 60 {
		fmt.Println("delay period changed to 60")
		delay = 60
	}
	params := rpc.RetrieveBackupTx{
		BackupAddr:  backup,
		DefaultAddr: defaultAddr,
		DelayPeriod: delay,
		Fee:         feeInt64,
	}
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "retrieve.CreateRawRetrieveBackupTx", params, nil)
	ctx.RunWithoutMarshal()
}

// PrepareCmd construct prepare tx
func PrepareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Prepare the wallet",
		Run:   prepareCmd,
	}
	addRetrieveCmdFlags(cmd)
	return cmd
}

func addRetrieveCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("backup", "b", "", "backup address")
	cmd.MarkFlagRequired("backup")
	cmd.Flags().StringP("default", "t", "", "default address")
	cmd.MarkFlagRequired("default")

	cmd.Flags().Float64P("fee", "f", 0.0, "sign address")
}

func addPerformCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("backup", "b", "", "backup address")
	cmd.MarkFlagRequired("backup")
	cmd.Flags().StringP("default", "t", "", "default address")
	cmd.MarkFlagRequired("default")

	cmd.Flags().StringArrayP("exec", "e", []string{}, "asset exec")
	cmd.Flags().StringArrayP("symbol", "s", []string{}, "asset symbol")

	cmd.Flags().Float64P("fee", "f", 0.0, "sign address")
}

func prepareCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")

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

	params := rpc.RetrievePrepareTx{
		BackupAddr:  backup,
		DefaultAddr: defaultAddr,
		Fee:         feeInt64,
	}
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "retrieve.CreateRawRetrievePrepareTx", params, nil)
	ctx.RunWithoutMarshal()
}

// PerformCmd construct perform tx
func PerformCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "perform",
		Short: "Perform the retrieve",
		Run:   performCmd,
	}
	addPerformCmdFlags(cmd)
	return cmd
}

func performCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")

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

	execs, _ := cmd.Flags().GetStringArray("exec")
	symbols, _ := cmd.Flags().GetStringArray("symbol")

	params := rpc.RetrievePerformTx{
		BackupAddr:  backup,
		DefaultAddr: defaultAddr,
		Assets:      []rpc.Asset{},
		Fee:         feeInt64,
	}
	if len(execs) != len(symbols) {
		fmt.Printf("exec count must equal to symbol count\n")
		return
	}
	for i := 0; i < len(execs); i++ {
		params.Assets = append(params.Assets, rpc.Asset{Exec: execs[i], Symbol: symbols[i]})
	}

	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "retrieve.CreateRawRetrievePerformTx", params, nil)
	ctx.RunWithoutMarshal()
}

// CancelCmd construct cancel tx
func CancelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel the retrieve",
		Run:   cancelCmd,
	}
	addRetrieveCmdFlags(cmd)
	return cmd
}

func cancelCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")

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

	params := rpc.RetrieveCancelTx{
		BackupAddr:  backup,
		DefaultAddr: defaultAddr,
		Fee:         feeInt64,
	}
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "retrieve.CreateRawRetrieveCancelTx", params, nil)
	ctx.RunWithoutMarshal()
}

// RetrieveQueryCmd cmds
func RetrieveQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "show retrieve info",
		Run:   queryRetrieveCmd,
	}
	addQueryRetrieveCmdFlags(cmd)
	return cmd
}

func addQueryRetrieveCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("backup", "b", "", "backup address")
	cmd.MarkFlagRequired("backup")
	cmd.Flags().StringP("default", "t", "", "default address")
	cmd.MarkFlagRequired("default")

	cmd.Flags().StringP("asset_exec", "e", "", "asset exec")
	cmd.Flags().StringP("asset_symbol", "s", "", "asset symbol")
}

func parseRerieveDetail(arg interface{}) (interface{}, error) {
	res := arg.(*rt.RetrieveQuery)

	result := RetrieveResult{
		DelayPeriod: res.DelayPeriod,
	}
	switch res.Status {
	case rt.RetrieveBackup:
		result.Status = "backup"
	case rt.RetrievePreapre:
		result.Status = "prepared"
	case rt.RetrievePerform:
		result.Status = "performed"
	case rt.RetrieveCancel:
		result.Status = "canceled"
	default:
		result.Status = "unknown"
	}

	return result, nil
}

func queryRetrieveCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")
	exec, _ := cmd.Flags().GetString("asset_exec")
	symbol, _ := cmd.Flags().GetString("asset_symbol")

	req := &rt.ReqRetrieveInfo{
		BackupAddress:  backup,
		DefaultAddress: defaultAddr,
		AssetExec:      exec,
		AssetSymbol:    symbol,
	}

	var params rpctypes.Query4Jrpc
	params.Execer = "retrieve"
	params.FuncName = "GetRetrieveInfo"
	params.Payload = types.MustPBToJSON(req)

	var res rt.RetrieveQuery
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(parseRerieveDetail)
	ctx.Run()
}

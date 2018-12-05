// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"

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

	defaultFee := float64(types.GInt("MinFee")) / float64(types.Coin)
	cmd.Flags().Float64P("fee", "f", defaultFee, "transaction fee")
}

func backupCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")
	delay, _ := cmd.Flags().GetInt64("delay")
	fee, _ := cmd.Flags().GetFloat64("fee")

	if delay < 60 {
		fmt.Println("delay period changed to 60")
		delay = 60
	}
	feeInt64 := int64(fee*types.InputPrecision) * types.Multiple1E4
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

	defaultFee := float64(types.GInt("MinFee")) / float64(types.Coin)
	cmd.Flags().Float64P("fee", "f", defaultFee, "sign address")
}

func prepareCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee*types.InputPrecision) * types.Multiple1E4
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
	addRetrieveCmdFlags(cmd)
	return cmd
}

func performCmd(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	backup, _ := cmd.Flags().GetString("backup")
	defaultAddr, _ := cmd.Flags().GetString("default")
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee*types.InputPrecision) * types.Multiple1E4
	params := rpc.RetrievePerformTx{
		BackupAddr:  backup,
		DefaultAddr: defaultAddr,
		Fee:         feeInt64,
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
	fee, _ := cmd.Flags().GetFloat64("fee")

	feeInt64 := int64(fee*types.InputPrecision) * types.Multiple1E4
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
		Short: "Backup the wallet",
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

	req := &rt.ReqRetrieveInfo{
		BackupAddress:  backup,
		DefaultAddress: defaultAddr,
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

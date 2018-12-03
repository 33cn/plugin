// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/spf13/cobra"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze",
		Short: "Unfreeze construct management",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(createCmd())
	cmd.AddCommand(withdrawCmd())
	cmd.AddCommand(terminateCmd())
	cmd.AddCommand(showCmd())
	cmd.AddCommand(queryWithdrawCmd())
	return cmd
}

func createCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create unfreeze construct",
	}

	cmd.AddCommand(fixAmountCmd())
	cmd.AddCommand(leftCmd())
	return cmd
}

func createFlag(cmd *cobra.Command) *cobra.Command {
	cmd.PersistentFlags().StringP("beneficiary", "b", "", "address of beneficiary")
	cmd.MarkFlagRequired("beneficiary")

	cmd.PersistentFlags().StringP("asset_exec", "e", "", "asset exec")
	cmd.MarkFlagRequired("asset_exec")

	cmd.PersistentFlags().StringP("asset_symbol", "s", "", "asset symbol")
	cmd.MarkFlagRequired("asset_symbol")

	cmd.PersistentFlags().Float64P("total", "t", 0, "total count of asset")
	cmd.MarkFlagRequired("total")

	cmd.PersistentFlags().Int64P("start_ts", "", 0, "effect, UTC timestamp")
	//cmd.MarkFlagRequired("start_ts")

	return cmd
}

func getCreateFlags(cmd *cobra.Command) *pty.UnfreezeCreate {
	beneficiary, _ := cmd.Flags().GetString("beneficiary")
	exec, _ := cmd.Flags().GetString("asset_exec")
	symbol, _ := cmd.Flags().GetString("asset_symbol")
	total, _ := cmd.Flags().GetInt64("total")
	startTs, _ := cmd.Flags().GetInt64("start_ts")

	unfreeze := &pty.UnfreezeCreate{
		StartTime:   startTs,
		AssetExec:   exec,
		AssetSymbol: symbol,
		TotalCount:  total,
		Beneficiary: beneficiary,
		Means:       "",
	}
	return unfreeze
}

func fixAmountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix_amount",
		Short: "create fix amount means unfreeze construct",
		Run:   fixAmount,
	}
	cmd = createFlag(cmd)
	cmd.Flags().Int64P("amount", "a", 0, "amount every period")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Int64P("period", "p", 0, "period in second")
	cmd.MarkFlagRequired("period")
	return cmd
}

func fixAmount(cmd *cobra.Command, args []string) {
	create := getCreateFlags(cmd)

	amount, _ := cmd.Flags().GetInt64("amount")
	period, _ := cmd.Flags().GetInt64("period")
	create.Means = pty.FixAmountX
	create.MeansOpt = &pty.UnfreezeCreate_FixAmount{FixAmount: &pty.FixAmount{Period: period, Amount: amount}}

	paraName, _ := cmd.Flags().GetString("paraName")
	tx, err := pty.CreateUnfreezeCreateTx(paraName, create)
	if err != nil {
		fmt.Printf("Create Tx frailed: %s", err)
		return
	}
	outputTx(tx)
}

func outputTx(tx *types.Transaction) {
	txHex := types.Encode(tx)
	fmt.Println(hex.EncodeToString(txHex))
}

func leftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "left_proportion",
		Short: "create left proportion means unfreeze construct",
		Run:   left,
	}
	cmd = createFlag(cmd)
	cmd.Flags().Int64P("ten_thousandth", "", 0, "input/10000 of total")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Int64P("period", "p", 0, "period in second")
	cmd.MarkFlagRequired("period")
	return cmd
}

func left(cmd *cobra.Command, args []string) {
	create := getCreateFlags(cmd)

	tenThousandth, _ := cmd.Flags().GetInt64("ten_thousandth")
	period, _ := cmd.Flags().GetInt64("period")
	create.Means = pty.FixAmountX
	create.MeansOpt = &pty.UnfreezeCreate_LeftProportion{
		LeftProportion: &pty.LeftProportion{Period: period, TenThousandth: tenThousandth}}

	paraName, _ := cmd.Flags().GetString("paraName")
	tx, err := pty.CreateUnfreezeCreateTx(paraName, create)
	if err != nil {
		fmt.Printf("Create Tx frailed: %s", err)
		return
	}
	outputTx(tx)
}

func withdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "withdraw asset from construct",
		Run:   withdraw,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func terminateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terminate",
		Short: "terminate construct",
		Run:   terminate,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func showCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show construct",
		Run:   show,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func queryWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show available withdraw amount of one unfreeze construct",
		Run:   queryWithdraw,
	}
	cmd.Flags().StringP("id", "", "", "unfreeze construct id")
	cmd.MarkFlagRequired("id")

	return cmd
}

func withdraw(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	paraName, _ := cmd.Flags().GetString("paraName")
	tx, err := pty.CreateUnfreezeWithdrawTx(paraName, &pty.UnfreezeWithdraw{UnfreezeID: id})
	if err != nil {
		fmt.Printf("Create Tx frailed: %s", err)
		return
	}
	outputTx(tx)
}

func terminate(cmd *cobra.Command, args []string) {
	id, _ := cmd.Flags().GetString("id")
	paraName, _ := cmd.Flags().GetString("paraName")
	tx, err := pty.CreateUnfreezeTerminateTx(paraName, &pty.UnfreezeTerminate{UnfreezeID: id})
	if err != nil {
		fmt.Printf("Create Tx frailed: %s", err)
		return
	}
	outputTx(tx)
}

func queryWithdraw(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	id, _ := cmd.Flags().GetString("id")
	cli, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	param := &rpctypes.Query4Jrpc{
		Execer:   getRealExecName(paraName, pty.UnfreezeX),
		FuncName: "QueryWithdraw",
		Payload:  types.MustPBToJSON(&types.ReqString{Data: id}),
	}
	var resp pty.ReplyQueryUnfreezeWithdraw
	err = cli.Call("Chain33.Query", param, &resp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
}

func show(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	id, _ := cmd.Flags().GetString("id")
	cli, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	param := &rpctypes.Query4Jrpc{
		Execer:   getRealExecName(paraName, pty.UnfreezeX),
		FuncName: "GetUnfreeze",
		Payload:  types.MustPBToJSON(&types.ReqString{Data: id}),
	}
	var resp pty.Unfreeze
	err = cli.Call("Chain33.Query", param, &resp)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	data, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
}

func getRealExecName(paraName string, name string) string {
	if strings.HasPrefix(name, "user.p.") {
		return name
	}
	return paraName + name
}

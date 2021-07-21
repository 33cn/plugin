// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	coinsTy "github.com/33cn/plugin/plugin/dapp/coinsx/types"

	"github.com/spf13/cobra"
)

// CoinsxCmd coinsx 命令行
func CoinsxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coinsx",
		Short: "coins management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CoinsXConfigCmd(),
		coinsXQueryCmd(),
	)

	return cmd
}

// CoinsXConfigCmd create coinsx config cmd
func CoinsXConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Create coins config transaction",
	}
	cmd.AddCommand(transferFlagCmd())
	cmd.AddCommand(manageAccountsCmd())
	return cmd
}

func transferFlagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "config p2p transfer limit flag",
		Run:   createTransferFlagTx,
	}
	addTransferFlags(cmd)
	return cmd
}

func addTransferFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("flag", "f", 0, "set p2p transfer flag,1:enable,2:disable")
	cmd.MarkFlagRequired("flag")

}

func createTransferFlagTx(cmd *cobra.Command, args []string) {
	flag, _ := cmd.Flags().GetUint32("flag")
	paraName, _ := cmd.Flags().GetString("paraName")

	config := &coinsTy.CoinsConfig{}
	config.Ty = coinsTy.ConfigType_TRANSFER
	config.Value = &coinsTy.CoinsConfig_TransferFlag{TransferFlag: &coinsTy.TransferFlagConfig{Flag: coinsTy.TransferFlag(flag)}}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, coinsTy.CoinsxX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(config),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func manageAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "config manage accounts",
		Run:   createManageAccountsTx,
	}
	addManageAccountsFlags(cmd)
	return cmd
}

func addManageAccountsFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("op", "o", 0, "modify manager accounts,0:add,1:delete")
	cmd.MarkFlagRequired("op")

	cmd.Flags().StringP("accounts", "a", "", "accounts to modify, seperate with ',' ")
	cmd.MarkFlagRequired("accounts")
}

func createManageAccountsTx(cmd *cobra.Command, args []string) {
	op, _ := cmd.Flags().GetUint32("op")
	accounts, _ := cmd.Flags().GetString("accounts")
	paraName, _ := cmd.Flags().GetString("paraName")

	config := &coinsTy.CoinsConfig{}
	config.Ty = coinsTy.ConfigType_ACCOUNTS
	config.Value = &coinsTy.CoinsConfig_ManagerAccounts{
		ManagerAccounts: &coinsTy.ManagerAccountsConfig{
			Op:       coinsTy.AccountOp(op),
			Accounts: accounts,
		},
	}

	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, coinsTy.CoinsxX),
		ActionName: "Config",
		Payload:    types.MustPBToJSON(config),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// coinsXConfigCmd create coinsx query cmd
func coinsXQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query coinsx config status",
	}
	cmd.AddCommand(queryManagerAddrsCmd())
	return cmd
}

// queryManagerAddrsCmd get transfer flag cmd
func queryManagerAddrsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manager",
		Short: "Query manager status",
		Run:   queryManager,
	}
	return cmd
}

func queryManager(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = "coinsx"
	params.FuncName = "GetManagerStatus"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})

	var res coinsTy.ManagerStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

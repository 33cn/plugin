package main

import (
	"fmt"

	"github.com/33cn/chain33/rpc/jsonclient"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/spf13/cobra"
)

//StaticsCmd ...
func StaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statics",
		Short: "statics of lock/burn asset from or to Ethereum and chain33",
		Run:   ShowStatics,
	}

	ShowStaticsFlags(cmd)
	return cmd
}

//ShowLockStaticsFlags ...
func ShowStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().Int32P("from", "f", 0, "source chain, 0=ethereum, and 1=chain33")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().Int32P("operation", "o", 0, "operation type, 1=burn, and 2=lock")
	_ = cmd.MarkFlagRequired("operation")
	cmd.Flags().Int32P("status", "u", 0, "show with specified status, default to show all, 1=pending, 2=failed, 3=successful")
}

//ShowLockStatics ...
func ShowStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	from, _ := cmd.Flags().GetInt32("from")
	operation, _ := cmd.Flags().GetInt32("operation")
	status, _ := cmd.Flags().GetInt32("status")

	if from != 0 && 1 != from {
		fmt.Println("Pls set correct source chain flag, 0=ethereum, and 1=chain33")
		return
	}

	if operation != 2 && 1 != operation {
		fmt.Println("Pls set correct operation type, 1=burn, and 2=lock")
		return
	}

	if status < 0 || status > 3 {
		fmt.Println("Pls set correct status, default 0 to show all, 1=pending, 2=failed, 3=successful")
		return
	}

	para := ebTypes.TokenStaticsRequest{
		Symbol:    symbol,
		From:      from,
		Operation: operation,
		Status:    status,
	}
	var res ebTypes.TokenStaticsResponse
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowTokenStatics", para, &res)
	ctx.Run()
}

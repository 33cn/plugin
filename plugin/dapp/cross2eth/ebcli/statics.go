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
	cmd.Flags().StringP("symbol", "s", "", "token symbol(optional)")
	cmd.Flags().Int32P("from", "f", 0, "source chain, 0=ethereum, and 1=chain33")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("operation", "o", "b", "operation type, b=burn, l=lock, w=withdraw")
	_ = cmd.MarkFlagRequired("operation")
	cmd.Flags().Int32P("status", "u", 0, "show with specified status, default to show all, 1=pending, 2=successful, 3=failed")
	cmd.Flags().Int32P("count", "n", 0, "count to show, default to show all")
	cmd.Flags().Int32P("index", "i", 0, "tx index(optional, exclude, default from 0)")
}

//ShowLockStatics ...
func ShowStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	from, _ := cmd.Flags().GetInt32("from")
	operation, _ := cmd.Flags().GetString("operation")
	status, _ := cmd.Flags().GetInt32("status")
	count, _ := cmd.Flags().GetInt32("count")
	index, _ := cmd.Flags().GetInt32("index")

	if from != 0 && 1 != from {
		fmt.Println("Pls set correct source chain flag, 0=ethereum, and 1=chain33")
		return
	}

	if operation != "b" && "l" != operation && operation != "w" {
		fmt.Println("Pls set correct operation type, b=burn, l=lock, w=withdraw")
		return
	}

	if status < 0 || status > 3 {
		fmt.Println("Pls set correct status, default 0 to show all, 1=pending, 2=successful, 3=failed")
		return
	}

	var operationInt int32
	if operation == "b" {
		operationInt = 1
	} else if operation == "l" {
		operationInt = 2
	} else if operation == "w" {
		operationInt = 3
	}

	para := &ebTypes.TokenStaticsRequest{
		Symbol:    symbol,
		From:      from,
		Operation: operationInt,
		Status:    status,
		TxIndex:   int64(index),
		Count:     count,
	}
	var res ebTypes.TokenStaticsResponse
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowTokenStatics", para, &res)
	ctx.Run()
}

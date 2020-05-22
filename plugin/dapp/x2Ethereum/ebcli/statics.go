package main

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/types"
	"github.com/spf13/cobra"
)

func StaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "statics",
		Short: "statics of lock/unlock Eth or ERC20,or deposit/burn chain33 asset ",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		ShowLockStaticsCmd(),
		ShowDepositStaticsCmd(),
	)

	return cmd
}

func ShowLockStaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "show the lock statics of ETH or ERC20",
		Run:   ShowLockStatics,
	}
	ShowLockStaticsFlags(cmd)
	return cmd
}

func ShowLockStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
}

func ShowLockStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")

	para := ebTypes.TokenStatics{
		TokenAddr: tokenAddr,
	}
	var res ebTypes.StaticsLock
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowLockStatics", para, &res)
	ctx.Run()
}

func ShowDepositStaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "show the deposit statics of chain33 asset",
		Run:   ShowDepositStatics,
	}
	ShowDepositStaticsFlags(cmd)
	return cmd
}

func ShowDepositStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
}

func ShowDepositStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")

	para := ebTypes.TokenStatics{
		TokenAddr: tokenAddr,
	}
	var res ebTypes.StaticsDeposit
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowDepositStatics", para, &res)
	ctx.Run()
}

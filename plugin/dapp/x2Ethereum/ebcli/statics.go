package main

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
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
		//ShowUnlockStaticsCmd(),
		//ShowBurnStaticsCmd(),
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
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowLockStatics", para, &res)
	ctx.Run()
}

func ShowUnlockStaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unlock",
		Short: "show the unlock statics of ETH or ERC20",
		Run:   ShowUnlockStatics,
	}
	ShowUnlockStaticsFlags(cmd)
	return cmd
}

func ShowUnlockStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
	cmd.Flags().StringP("owner", "o", "", "owner address, optional, nil for all")
}

func ShowUnlockStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")
	owner, _ := cmd.Flags().GetString("owner")
	para := ebTypes.StaticsRequest{
		Owner:     owner,
		TokenAddr: tokenAddr,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowUnlockStatics", para, &res)
	ctx.Run()
}

func ShowBurnStaticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "show the burn statics of chain33 asset",
		Run:   ShowBurnStatics,
	}
	ShowBurnStaticsFlags(cmd)
	return cmd
}

func ShowBurnStaticsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
	cmd.Flags().StringP("owner", "o", "", "owner address, optional, nil for all")
}

func ShowBurnStatics(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")
	owner, _ := cmd.Flags().GetString("owner")

	para := ebTypes.StaticsRequest{
		Owner:     owner,
		TokenAddr: tokenAddr,
	}

	var res ebTypes.StaticsResponse
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowLockStaticsCmd", para, &res)
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
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowDepositStatics", para, &res)
	ctx.Run()
}

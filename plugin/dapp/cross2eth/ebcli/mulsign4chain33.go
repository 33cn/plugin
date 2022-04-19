package main

import (
	"fmt"
	"strings"

	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/spf13/cobra"
)

func MultiSignCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign",
		Short: "deploy,setup and trasfer multisign",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		SetupCmd(),
		TransferCmd(),
		ShowAddrCmd(),
		SetChain33MultiSignAddrCmd(),
		GetChain33MultiSignAddrCmd(),
	)
	return cmd
}

func ShowAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show address's hash160",
		Run:   ShowAddr,
	}
	ShowAddrCmdFlags(cmd)
	return cmd
}

func ShowAddrCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "address")
	_ = cmd.MarkFlagRequired("address")
}

func ShowAddr(cmd *cobra.Command, args []string) {
	addressstr, _ := cmd.Flags().GetString("address")

	addr, err := address.NewBtcAddress(addressstr)
	if nil != err {
		fmt.Println("Wrong address")
		return
	}
	fmt.Println(common.ToHex(addr.Hash160[:]))
	return
}

func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup owners to contract",
		Run:   SetupOwner,
	}
	SetupOwnerFlags(cmd)
	return cmd
}

func SetupOwnerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owners's address, separated by ','")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("operator", "k", "", "operator private key")
	_ = cmd.MarkFlagRequired("operator")
}

func SetupOwner(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ownersStr, _ := cmd.Flags().GetString("owner")
	operator, _ := cmd.Flags().GetString("operator")
	owners := strings.Split(ownersStr, ",")

	para := ebTypes.SetupMulSign{
		OperatorPrivateKey: operator,
		Owners:             owners,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SetupOwner4Chain33", para, &res)
	ctx.Run()
}

func TransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer via safe",
		Run:   Transfer,
	}
	TransferFlags(cmd)
	return cmd
}

func TransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("receiver", "r", "", "receive address")
	_ = cmd.MarkFlagRequired("receiver")

	cmd.Flags().Float64P("amount", "a", 0, "amount to transfer")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("keys", "k", "", "owners' private key, separated by ','")
	_ = cmd.MarkFlagRequired("keys")

	cmd.Flags().StringP("token", "t", "", "erc20 address,not need to set for BTY(optional)")
}

func Transfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	receiver, _ := cmd.Flags().GetString("receiver")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	keysStr, _ := cmd.Flags().GetString("keys")

	keys := strings.Split(keysStr, ",")

	para := ebTypes.SafeTransfer{
		To:               receiver,
		Token:            tokenAddr,
		Amount:           amount,
		OwnerPrivateKeys: keys,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SafeTransfer4Chain33", para, &res)
	ctx.Run()
}

func SetChain33MultiSignAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_multiSign",
		Short: "set multiSign address",
		Run:   SetChain33MultiSignAddr,
	}
	SetChain33MultiSignAddrCmdFlags(cmd)
	return cmd
}

func SetChain33MultiSignAddrCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "address")
	_ = cmd.MarkFlagRequired("address")
}

func SetChain33MultiSignAddr(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("address")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SetChain33MultiSignAddr", addr, &res)
	ctx.Run()
}

func GetChain33MultiSignAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_multiSign",
		Short: "get multiSign address",
		Run:   GetChain33MultiSignAddr,
	}
	return cmd
}

func GetChain33MultiSignAddr(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.GetChain33MultiSignAddr", "", &res)
	ctx.Run()
}

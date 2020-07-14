package main

import (
	"fmt"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/types"
	"github.com/spf13/cobra"
)

//Chain33RelayerCmd RelayerCmd command func
func Chain33RelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain33 ",
		Short: "Chain33 relayer ",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		ImportPrivateKeyCmd(),
		ShowValidatorAddrCmd(),
		ShowTxsHashCmd(),
	)

	return cmd
}

//ImportPrivateKeyCmd SetPwdCmd set password
func ImportPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import_privatekey",
		Short: "import ethereum private key to sign txs to be submitted to ethereum",
		Run:   importPrivatekey,
	}
	addImportPrivateKeyFlags(cmd)
	return cmd
}

func addImportPrivateKeyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "ethereum private key")
	cmd.MarkFlagRequired("key")
}

func importPrivatekey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKey, _ := cmd.Flags().GetString("key")
	importKeyReq := ebTypes.ImportKeyReq{
		PrivateKey: privateKey,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ImportChain33RelayerPrivateKey", importKeyReq, &res)
	ctx.Run()
}

//ShowValidatorAddrCmd ...
func ShowValidatorAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_validator",
		Short: "show me the validator",
		Run:   showValidatorAddr,
	}
	return cmd
}

func showValidatorAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowChain33RelayerValidator", nil, &res)
	ctx.Run()
}

//ShowTxsHashCmd ...
func ShowTxsHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_txhashes",
		Short: "show me the tx hashes",
		Run:   showChain33Relayer2EthTxs,
	}
	return cmd
}

func showChain33Relayer2EthTxs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var res ebTypes.Txhashes
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowChain33Relayer2EthTxs", nil, &res)
	if _, err := ctx.RunResult(); nil != err {
		errInfo := err.Error()
		fmt.Println("errinfo:" + errInfo)
		return
	}
	for _, hash := range res.Txhash {
		fmt.Println(hash)
	}
}

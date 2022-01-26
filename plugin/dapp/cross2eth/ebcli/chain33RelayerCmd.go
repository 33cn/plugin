package main

import (
	"fmt"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
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
		LockAsyncFromChain33Cmd(),
		BurnfromChain33Cmd(),
		simBurnFromEthCmd(),
		simLockFromEthCmd(),
		ShowBridgeRegistryAddr4chain33Cmd(),
		TokenAddressCmd(),
		MultiSignCmd(),
		ResendChain33EventCmd(),
		WithdrawFromChain33Cmd(),
		BurnWithIncreasefromChain33Cmd(),
	)

	return cmd
}

//TokenAddressCmd...
func TokenAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "show token address and it's corresponding symbol",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		ShowTokenAddressCmd(),
	)
	return cmd
}

func ShowTokenAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show token address",
		Run:   ShowTokenAddress,
	}
	ShowTokenFlags(cmd)
	return cmd
}

//SetTokenFlags ...
func ShowTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol(optional), if not set,show all the token")
}

func ShowTokenAddress(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")

	var res ebTypes.TokenAddressArray
	para := ebTypes.TokenAddress{
		Symbol:    symbol,
		ChainName: ebTypes.Chain33BlockChainName,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowTokenAddress", para, &res)
	ctx.Run()
}

func ShowBridgeRegistryAddr4chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridgeRegistry",
		Short: "show the address of Contract BridgeRegistry for chain33",
		Run:   ShowBridgeRegistryAddr4chain33,
	}
	return cmd
}

//ShowBridgeRegistryAddr ...
func ShowBridgeRegistryAddr4chain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowBridgeRegistryAddr4chain33", nil, &res)
	ctx.Run()
}

func simBurnFromEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim-burn",
		Short: "simulate burn bty assets from ethereum",
		Run:   simBurnFromEth,
	}
	SimBurnFlags(cmd)
	return cmd
}

//SimBurnFlags ...
func SimBurnFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "Ethereum sender address")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("receiver", "r", "", "receiver address on chain33")
	_ = cmd.MarkFlagRequired("receiver")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func simBurnFromEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	realAmount := utils.ToWei(amount, 8)

	para := ebTypes.Burn{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SimBurnFromEth", para, &res)
	ctx.Run()
}

func simLockFromEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sim-lock",
		Short: "simulate lock eth/erc20 assets from ethereum",
		Run:   simLockFromEth,
	}
	simLockEthErc20AssetFlags(cmd)
	return cmd
}

//LockEthErc20AssetFlags ...
func simLockEthErc20AssetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "Ethereum sender address")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("receiver", "r", "", "chain33 receiver address")
	_ = cmd.MarkFlagRequired("receiver")
}

func simLockFromEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	realAmount := utils.ToWei(amount, 8)

	para := ebTypes.LockEthErc20{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SimLockFromEth", para, &res)
	ctx.Run()
}

//LockAsyncCmd ...
func LockAsyncFromChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "async lock bty from chain33 and cross-chain transfer to ethereum",
		Run:   LockBTYAssetAsync,
	}
	LockBTYAssetFlags(cmd)
	return cmd
}

func LockBTYAssetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("receiver", "r", "", "etheruem receiver address")
	_ = cmd.MarkFlagRequired("receiver")
}

func LockBTYAssetAsync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	realAmount := utils.ToWei(amount, 8)

	para := ebTypes.LockBTY{
		OwnerKey:        key,
		Amount:          realAmount.String(),
		EtherumReceiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.LockBTYAssetAsync", para, &res)
	ctx.Run()
}

func BurnfromChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "async burn the asset from chain33 to make it unlocked on ethereum",
		Run:   BurnAsyncFromChain33,
	}
	BurnAsyncFromChain33Flags(cmd)
	return cmd
}

//BurnAsyncFromChain33Flags ...
func BurnAsyncFromChain33Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key for chain33")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("receiver", "r", "", "receiver address on Ethereum")
	_ = cmd.MarkFlagRequired("receiver")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func BurnAsyncFromChain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	d, err := utils.SimpleGetDecimals(tokenAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}
	para := ebTypes.BurnFromChain33{
		OwnerKey:         key,
		TokenAddr:        tokenAddr,
		Amount:           utils.ToWei(amount, d).String(),
		EthereumReceiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.BurnAsyncFromChain33", para, &res)
	ctx.Run()
}

func BurnWithIncreasefromChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn_increase",
		Short: "async burn the asset from chain33 to make it unlocked on ethereum",
		Run:   BurnWithIncreaseAsyncFromChain33,
	}
	BurnAsyncFromChain33Flags(cmd)
	return cmd
}

func BurnWithIncreaseAsyncFromChain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	d, err := utils.SimpleGetDecimals(tokenAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}
	para := ebTypes.BurnFromChain33{
		OwnerKey:         key,
		TokenAddr:        tokenAddr,
		Amount:           utils.ToWei(amount, d).String(),
		EthereumReceiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.BurnWithIncreaseAsyncFromChain33", para, &res)
	ctx.Run()
}

//ImportPrivateKeyCmd SetPwdCmd set password
func ImportPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import_privatekey",
		Short: "import chain33 private key to sign txs to be submitted to chain33 evm",
		Run:   importPrivatekey,
	}
	addImportPrivateKeyFlags(cmd)
	return cmd
}

func addImportPrivateKeyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "chain33 private key")
	_ = cmd.MarkFlagRequired("key")
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

func ResendChain33EventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resendChain33Event",
		Short: "resend Chain33Event to ethereum process goroutine",
		Run:   resendChain33Event,
	}
	addResendChain33EventFlags(cmd)
	return cmd
}

func addResendChain33EventFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "g", 0, "height begin to resend chain33 event ")
	_ = cmd.MarkFlagRequired("height")
}

func resendChain33Event(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetInt64("height")
	resendChain33EventReq := &ebTypes.ResendChain33EventReq{
		Height: height,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ResendChain33Event", resendChain33EventReq, &res)
	ctx.Run()
}

func WithdrawFromChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "async withdraw the asset from chain33 to make it unlocked on ethereum",
		Run:   WithdrawFromChain33,
	}
	addWithdrawFromChain33Flags(cmd)
	return cmd
}

//addWithdrawFromChain33CmdFlags ...
func addWithdrawFromChain33Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key for chain33")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("receiver", "r", "", "receiver address on Ethereum")
	_ = cmd.MarkFlagRequired("receiver")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func WithdrawFromChain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")

	d, err := utils.SimpleGetDecimals(tokenAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}
	para := ebTypes.WithdrawFromChain33{
		OwnerKey:         key,
		TokenAddr:        tokenAddr,
		Amount:           utils.ToWei(amount, d).String(),
		EthereumReceiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.WithdrawFromChain33", para, &res)
	ctx.Run()
}

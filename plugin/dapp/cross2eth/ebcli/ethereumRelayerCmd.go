package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/33cn/chain33/common"
	chain33Common "github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	relayerTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

// EthereumRelayerCmd command func
func EthereumRelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum",
		Short: "Ethereum relayer",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		ImportEthPrivateKeyCmd(),
		GenEthPrivateKeyCmd(),
		GetAddressFromPrivateKeyCmd(),
		ShowValidatorsAddrCmd(),
		ShowChain33TxsHashCmd(),
		IsValidatorActiveCmd(),
		ShowOperatorCmd(),
		ShowTxReceiptCmd(),
		//////auxiliary///////
		GetBalanceCmd(),
		ApproveCmd(),
		BurnCmd(),
		BurnAsyncCmd(),
		LockSyncCmd(),
		LockAsyncCmd(),
		ShowBridgeBankAddrCmd(),
		ShowBridgeRegistryAddrCmd(),
		ShowBalanceLockedCmd(),
		DeployERC20Cmd(),
		TokenCmd(),
		MultiSignEthCmd(),
		TransferEthCmd(),
		CfgWithdrawCmd(),
		GetCfgWithdrawCmd(),
		ResendEthLockEventCmd(),
		RegetEthLockEventCmd(),
		CreateLockEventCmd(),
		QueryCmd(),
	)

	return cmd
}

// TokenCmd TokenAddressCmd...
func TokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "create bridgeToken, ERC20 Token, show token address and it's corresponding symbol",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		ShowTokenAddress4EthCmd(),
		ShowTokenAddress4LockEthCmd(),
		TransferTokenCmd(),
	)
	return cmd
}

func ShowTokenAddress4EthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show token address",
		Run:   ShowTokenAddress4Eth,
	}
	ShowTokenFlags(cmd)
	return cmd
}

func ShowTokenAddress4Eth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")

	var res ebTypes.TokenAddressArray
	para := ebTypes.TokenAddress{
		Symbol:    symbol,
		ChainName: ebTypes.EthereumBlockChainName,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowTokenAddress", para, &res)
	ctx.Run()
}

func ShowTokenAddress4LockEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_lock",
		Short: "show lock token address",
		Run:   ShowTokenAddress4LockEth,
	}
	ShowTokenFlags(cmd)
	return cmd
}

func ShowTokenAddress4LockEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")

	var res ebTypes.TokenAddressArray
	para := ebTypes.TokenAddress{
		Symbol:    symbol,
		ChainName: ebTypes.EthereumBlockChainName,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowETHLockTokenAddress", para, &res)
	ctx.Run()
}

func ImportEthPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import_privatekey",
		Short: "import ethereum private key to sign txs to be submitted to ethereum",
		Run:   importEthereumPrivatekey,
	}
	addImportEthPrivateKeyFlags(cmd)
	return cmd
}

func addImportEthPrivateKeyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "ethereum private key")
	_ = cmd.MarkFlagRequired("key")
}

func importEthereumPrivatekey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKey, _ := cmd.Flags().GetString("key")
	params := privateKey

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ImportEthereumPrivateKey4EthRelayer", params, &res)
	ctx.Run()
}

func GetAddressFromPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_eth_addr",
		Short: "get addr from private key",
		Run:   getAddressFromPrivateKey,
	}
	cmd.Flags().StringP("key", "k", "", "ethereum private key")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func getAddressFromPrivateKey(cmd *cobra.Command, args []string) {
	key, _ := cmd.Flags().GetString("key")
	privateKeySlice, err := chain33Common.FromHex(key)
	if nil != err {
		fmt.Println("private key error: ", err)
		return
	}
	privateKey, err := crypto.ToECDSA(privateKeySlice)
	if nil != err {
		fmt.Println("private key error: ", err)
		return
	}

	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Println("addr: ", addr)
}

//GenEthPrivateKeyCmd ...
func GenEthPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_eth_key",
		Short: "create ethereum's private key to sign txs to be submitted to ethereum",
		Run:   generateEthereumPrivateKey,
	}
	return cmd
}

func generateEthereumPrivateKey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var res ebTypes.Account4Show
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.GenerateEthereumPrivateKey", nil, &res)
	ctx.Run()
}

//ShowValidatorsAddrCmd ...
func ShowValidatorsAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_validator",
		Short: "show me the validators including ethereum and chain33",
		Run:   showValidatorsAddr,
	}
	return cmd
}

func showValidatorsAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")

	var res ebTypes.ValidatorAddr4EthRelayer
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowEthRelayerValidator", ethChainName, &res)
	ctx.Run()
}

//ShowChain33TxsHashCmd ...
func ShowChain33TxsHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_chain33_tx",
		Short: "show me the chain33 tx hashes",
		Run:   showChain33Txs,
	}
	return cmd
}

func showChain33Txs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var res ebTypes.Txhashes
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowEthRelayer2Chain33Txs", nil, &res)
	if _, err := ctx.RunResult(); nil != err {
		errInfo := err.Error()
		fmt.Println("errinfo:" + errInfo)
		return
	}
	for _, hash := range res.Txhash {
		fmt.Println(hash)
	}
}

//IsValidatorActiveCmd ...
func IsValidatorActiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active",
		Short: "show whether the validator is active or not",
		Run:   IsValidatorActive,
	}
	IsValidatorActiveFlags(cmd)
	return cmd
}

//IsValidatorActiveFlags ...
func IsValidatorActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "validator address")
	_ = cmd.MarkFlagRequired("addr")
}

//IsValidatorActive ...
func IsValidatorActive(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := addr
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.IsValidatorActive", params, &res)
	ctx.Run()
}

//ShowOperatorCmd ...
func ShowOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "show me the operator",
		Run:   ShowOperator,
	}
	return cmd
}

//ShowOperator ...
func ShowOperator(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowOperator", nil, &res)
	ctx.Run()
}

// DeployERC20Cmd ...
func DeployERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy_erc20",
		Short: "deploy ERC20 contracts",
		Run:   DeployERC20,
	}
	DeployERC20Flags(cmd)
	return cmd
}

func DeployERC20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "c", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("name", "n", "", "erc20 name")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("symbol", "s", "", "erc20 symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("amount", "m", "0", "amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().Uint8P("decimals", "d", 8, "default set to 8, and can't be greater than 18")
}

func DeployERC20(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	owner, _ := cmd.Flags().GetString("owner")
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetString("amount")
	decimals, _ := cmd.Flags().GetUint8("decimals")

	if decimals > 18 {
		fmt.Println("decimals can't be greater than 18")
		return
	}

	para := ebTypes.ERC20Token{
		Owner:    owner,
		Name:     name,
		Symbol:   symbol,
		Amount:   amount,
		Decimals: int32(decimals),
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.DeployERC20", para, &res)
	ctx.Run()
}

//ShowTxReceiptCmd ...
func ShowTxReceiptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "receipt",
		Short: "show me the tx receipt for Ethereum",
		Run:   ShowTxReceipt,
	}
	ShowTxReceiptFlags(cmd)
	return cmd
}

//ShowTxReceiptFlags ...
func ShowTxReceiptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "tx hash")
	_ = cmd.MarkFlagRequired("hash")
}

//ShowTxReceipt ...
func ShowTxReceipt(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	txhash, _ := cmd.Flags().GetString("hash")

	para := &relayerTypes.TxReceiptReq{
		TxHash:    txhash,
		ChainName: ethChainName,
	}

	var res ethTypes.Receipt
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowTxReceipt", para, &res)
	ctx.Run()
}

//ApproveCmd ...
func ApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve the allowance to bridgebank by the owner",
		Run:   ApproveAllowance,
	}
	ApproveAllowanceFlags(cmd)
	return cmd
}

//ApproveAllowanceFlags ...
func ApproveAllowanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

//ApproveAllowance ...
func ApproveAllowance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error")
		return
	}

	realAmount := utils.ToWei(amount, d)
	para := ebTypes.ApproveAllowance{
		OwnerKey:  key,
		TokenAddr: tokenAddr,
		Amount:    realAmount.String(),
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ApproveAllowance", para, &res)
	ctx.Run()
}

//BurnCmd ...
func BurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "burn(including approve) the asset to make it unlocked on chain33",
		Run:   Burn,
	}
	BurnFlags(cmd)
	return cmd
}

//BurnFlags ...
func BurnFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("receiver", "r", "", "receiver address on chain33")
	_ = cmd.MarkFlagRequired("receiver")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

//BurnAsyncCmd ...
func BurnAsyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-async",
		Short: "async burn the asset to make it unlocked on chain33",
		Run:   BurnAsync,
	}
	BurnFlags(cmd)
	return cmd
}

// Burn ...
func Burn(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}
	para := ebTypes.Burn{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          utils.ToWei(amount, d).String(),
		Chain33Receiver: receiver,
		ChainName:       ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.Burn", para, &res)
	ctx.Run()
}

//BurnAsync ...
func BurnAsync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}
	para := ebTypes.Burn{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          utils.ToWei(amount, d).String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.BurnAsync", para, &res)
	ctx.Run()
}

//LockSyncCmd ...
func LockSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "lock(including approve) eth or erc20 and cross-chain transfer to chain33",
		Run:   LockEthErc20Asset,
	}
	LockEthErc20AssetFlags(cmd)
	return cmd
}

//LockAsyncCmd ...
func LockAsyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-async",
		Short: "async lock eth or erc20 and cross-chain transfer to chain33",
		Run:   LockEthErc20AssetAsync,
	}
	LockEthErc20AssetFlags(cmd)
	return cmd
}

//LockEthErc20AssetFlags ...
func LockEthErc20AssetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("receiver", "r", "", "chain33 receiver address")
	_ = cmd.MarkFlagRequired("receiver")
}

//LockEthErc20Asset ...
func LockEthErc20Asset(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}

	realAmount := utils.ToWei(amount, d)

	para := ebTypes.LockEthErc20{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
		ChainName:       ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.LockEthErc20Asset", para, &res)
	//ctx.Run
	for try := 0; try < 3; try++ {
		result, err := ctx.RunResult()
		if err != nil {
			time.Sleep(time.Millisecond * 500)
			continue
		}
		data, err := json.MarshalIndent(result, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		fmt.Println(string(data))
		return
	}
}

//LockEthErc20AssetAsync ...
func LockEthErc20AssetAsync(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	key, _ := cmd.Flags().GetString("key")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals err")
		return
	}

	realAmount := utils.ToWei(amount, d)

	para := ebTypes.LockEthErc20{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.LockEthErc20AssetAsync", para, &res)
	ctx.Run()
}

//ShowBridgeBankAddrCmd ...
func ShowBridgeBankAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridgeBankAddr",
		Short: "show the address of Contract BridgeBank",
		Run:   ShowBridgeBankAddr,
	}
	return cmd
}

//ShowBridgeBankAddr ...
func ShowBridgeBankAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowBridgeBankAddr", nil, &res)
	ctx.Run()
}

//ShowBridgeRegistryAddrCmd ...
func ShowBridgeRegistryAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridgeRegistry",
		Short: "show the address of Contract BridgeRegistry",
		Run:   ShowBridgeRegistryAddr,
	}
	return cmd
}

//ShowBridgeRegistryAddr ...
func ShowBridgeRegistryAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowBridgeRegistryAddr", nil, &res)
	ctx.Run()
}

//GetBalanceCmd ...
func ShowBalanceLockedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showBalanceLocked",
		Short: "show Balance Locked in BridgeBank",
		Run:   ShowBalanceLocked,
	}
	ShowBalanceLockedFlags(cmd)
	return cmd
}

func ShowBalanceLockedFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("bridgeBank", "b", "", "bridgeBank address")
	_ = cmd.MarkFlagRequired("bridgeBank")
	cmd.Flags().StringP("tokenAddr", "t", "", "token address, optional, nil for platform token(Eth)")
}

//GetBalance ...
func ShowBalanceLocked(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	bridgeBank, _ := cmd.Flags().GetString("bridgeBank")
	tokenAddr, _ := cmd.Flags().GetString("tokenAddr")

	if tokenAddr == "" {
		tokenAddr = "0x0000000000000000000000000000000000000000"
	}

	para := &ebTypes.BalanceLockedReq{
		BridgeBank: bridgeBank,
		TokenAddr:  tokenAddr,
	}
	var res ebTypes.ReplyBalance
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowBalanceLocked", para, &res)
	ctx.Run()
}

//GetBalanceCmd ...
func GetBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "get owner's balance for ETH or ERC20",
		Run:   GetBalance,
	}
	GetBalanceFlags(cmd)
	return cmd
}

//GetBalanceFlags ...
func GetBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("tokenAddr", "t", "", "token address, optional, nil for Eth")
}

//GetBalance ...
func GetBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	owner, _ := cmd.Flags().GetString("owner")
	tokenAddr, _ := cmd.Flags().GetString("tokenAddr")

	para := ebTypes.BalanceAddr{
		Owner:     owner,
		TokenAddr: tokenAddr,
		ChainName: ethChainName,
	}
	var res ebTypes.ReplyBalance
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.GetBalance", para, &res)
	ctx.Run()
}

//TransferTokenCmd ...
func TransferTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token_transfer",
		Short: "create a token transfer transaction",
		Run:   TransferToken,
	}
	TransferTokenFlags(cmd)
	return cmd
}

//TransferTokenFlags ...
func TransferTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("from", "k", "", "from private key")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("to", "r", "", "to address")
	_ = cmd.MarkFlagRequired("to")
	cmd.Flags().Float64P("amount", "m", 0, "amount")
	_ = cmd.MarkFlagRequired("amount")
}

//TransferToken ...
func TransferToken(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	tokenAddr, _ := cmd.Flags().GetString("token")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error", err.Error())
		return
	}

	realAmount := utils.ToWei(amount, d)
	para := ebTypes.TransferToken{
		TokenAddr: tokenAddr,
		FromKey:   from,
		ToAddr:    to,
		Amount:    realAmount.String(),
		ChainName: ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.TransferToken", para, &res)
	ctx.Run()
}

func MultiSignEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign",
		Short: "deploy,setup and trasfer multisign",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		DeployMultiSignEthCmd(),
		SetupEthCmd(),
		MultiSignTransferEthCmd(),
		ShowEthAddrCmd(),
		ConfigLockedTokenOfflineSaveCmd(),
		GetSelfBalanceCmd(),
		SetEthMultiSignAddrCmd(),
		GetEthMultiSignAddrCmd(),
	)
	return cmd
}

func GetSelfBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "selfbalance",
		Short: "get balance for multisign",
		Run:   showMultiBalance,
	}
	return cmd
}

//showMultiBalance ...
func showMultiBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")

	para := ebTypes.BalanceAddr{ChainName: ethChainName}
	var res ebTypes.ReplyBalance
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ShowMultiBalance", para, &res)
	ctx.Run()
}

func DeployMultiSignEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy mulsign to ethereum",
		Run:   DeployMultiSignEth,
	}
	return cmd
}

func DeployMultiSignEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.DeployMulsign2Eth", nil, &res)
	ctx.Run()
}

func ShowEthAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "show address's hash160",
		Run:   ShowEthAddr,
	}
	ShowEthAddrCmdFlags(cmd)
	return cmd
}

func ShowEthAddrCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "address")
	_ = cmd.MarkFlagRequired("address")
}

func ShowEthAddr(cmd *cobra.Command, args []string) {
	addressstr, _ := cmd.Flags().GetString("address")

	addr, err := address.NewBtcAddress(addressstr)
	if nil != err {
		fmt.Println("Wrong address")
		return
	}
	fmt.Println(common.ToHex(addr.Hash160[:]))
	return
}

func SetupEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup owners to contract",
		Run:   SetupEthOwner,
	}
	SetupEthOwnerFlags(cmd)
	return cmd
}

func SetupEthOwnerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owners's address, separated by ','")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("operator", "k", "", "operator private key")
	_ = cmd.MarkFlagRequired("operator")
}

func SetupEthOwner(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ownersStr, _ := cmd.Flags().GetString("owner")
	operator, _ := cmd.Flags().GetString("operator")
	owners := strings.Split(ownersStr, ",")

	para := ebTypes.SetupMulSign{
		OperatorPrivateKey: operator,
		Owners:             owners,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SetupOwner4Eth", para, &res)
	ctx.Run()
}

func MultiSignTransferEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer via safe",
		Run:   SafeTransferEth,
	}
	SafeTransferEthFlags(cmd)
	return cmd
}

func SafeTransferEthFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("receiver", "r", "", "receive address")
	_ = cmd.MarkFlagRequired("receiver")

	cmd.Flags().Float64P("amount", "a", 0, "amount to transfer")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("keys", "k", "", "owners' private key, separated by ','")
	_ = cmd.MarkFlagRequired("keys")

	cmd.Flags().StringP("operator", "o", "", "operator private key")
	_ = cmd.MarkFlagRequired("operator")

	cmd.Flags().StringP("token", "t", "", "erc20 address,not need to set for ETH(optional)")
}

func SafeTransferEth(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	receiver, _ := cmd.Flags().GetString("receiver")
	tokenAddr, _ := cmd.Flags().GetString("token")
	amount, _ := cmd.Flags().GetFloat64("amount")
	keysStr, _ := cmd.Flags().GetString("keys")
	operatorKey, _ := cmd.Flags().GetString("operator")

	keys := strings.Split(keysStr, ",")

	para := ebTypes.SafeTransfer{
		To:                 receiver,
		Token:              tokenAddr,
		Amount:             amount,
		OperatorPrivateKey: operatorKey,
		OwnerPrivateKeys:   keys,
		ChainName:          ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SafeTransfer4Eth", para, &res)
	ctx.Run()
}

//ConfigLockedTokenOfflineSaveCmd ...
func ConfigLockedTokenOfflineSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_offline_token",
		Short: "set config offline locked token",
		Run:   ConfigLockedTokenOfflineSave,
	}
	ConfigLockedTokenOfflineSaveFlags(cmd)
	return cmd
}

//ConfigLockedTokenOfflineSaveFlags ...
func ConfigLockedTokenOfflineSaveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token addr")
	//_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().Float64P("threshold", "m", 0, "threshold")
	_ = cmd.MarkFlagRequired("threshold")
	cmd.Flags().Uint32P("percents", "p", 50, "percents")
	//_ = cmd.MarkFlagRequired("percents")
}

//ConfigLockedTokenOfflineSave ...
func ConfigLockedTokenOfflineSave(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	symbol, _ := cmd.Flags().GetString("symbol")
	token, _ := cmd.Flags().GetString("token")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	percents, _ := cmd.Flags().GetUint32("percents")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(token, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error", err.Error())
		return
	}

	realAmount := utils.ToWei(threshold, d)
	para := ebTypes.ETHConfigLockedTokenOffline{
		Symbol:    symbol,
		Address:   token,
		Threshold: realAmount.String(),
		Percents:  percents,
		ChainName: ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ConfigLockedTokenOfflineSave", para, &res)
	ctx.Run()
}

func TransferEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "create a transfer transaction",
		Run:   TransferEth,
	}
	TransferEthFlags(cmd)
	return cmd
}

func TransferEthFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "k", "", "from private key")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("to", "r", "", "to address")
	_ = cmd.MarkFlagRequired("to")
	cmd.Flags().Float64P("amount", "m", 0, "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func TransferEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetFloat64("amount")

	d := int64(18)

	realAmount := utils.ToWei(amount, d)
	para := ebTypes.TransferToken{
		TokenAddr: "",
		FromKey:   from,
		ToAddr:    to,
		Amount:    realAmount.String(),
		ChainName: ethChainName,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.TransferEth", para, &res)
	ctx.Run()
}

func SetEthMultiSignAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_multiSign",
		Short: "set multiSign address",
		Run:   SetEthMultiSignAddr,
	}
	SetEthMultiSignAddrCmdFlags(cmd)
	return cmd
}

func SetEthMultiSignAddrCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "address")
	_ = cmd.MarkFlagRequired("address")
}

func SetEthMultiSignAddr(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	address, _ := cmd.Flags().GetString("address")
	para := &relayerTypes.CfgMultiSignAddr{
		ChainName:     ethChainName,
		MultiSignAddr: address,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.SetEthMultiSignAddr", para, &res)
	ctx.Run()
}

func GetEthMultiSignAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get_multiSign",
		Short: "get multiSign address",
		Run:   GetEthMultiSignAddr,
	}
	return cmd
}

func GetEthMultiSignAddr(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.GetEthMultiSignAddr", ethChainName, &res)
	ctx.Run()
}

func CfgWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cfgWithdraw",
		Short: "cfg withdraw fee",
		Run:   CfgWithdraw,
	}
	addCfgWithdrawFlags(cmd)
	return cmd
}

func addCfgWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().Float64P("fee", "f", 0, "fee amount")
	_ = cmd.MarkFlagRequired("fee")
	cmd.Flags().Float64P("amount", "a", 0, "accumulative amount allowed to be withdrew per day")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().Uint8P("decimal", "d", 0, "token decimal")
	_ = cmd.MarkFlagRequired("decimal")
}

func CfgWithdraw(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	symbol, _ := cmd.Flags().GetString("symbol")
	fee, _ := cmd.Flags().GetFloat64("fee")
	amount, _ := cmd.Flags().GetFloat64("amount")
	decimal, _ := cmd.Flags().GetUint8("decimal")

	req := &ebTypes.CfgWithdrawReq{
		Symbol:       symbol,
		FeeAmount:    utils.SmalToBig(fee, decimal).String(),
		AmountPerDay: utils.SmalToBig(amount, decimal).String(),
		ChainName:    ethChainName,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.CfgWithdraw", req, &res)
	ctx.Run()
}

func GetCfgWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "getCfgWithdraw",
		Short: "get cfg withdraw fee",
		Run:   GetCfgWithdraw,
	}
	addGetCfgWithdrawFlags(cmd)
	return cmd
}

func addGetCfgWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "symbol")
	_ = cmd.MarkFlagRequired("symbol")
}

func GetCfgWithdraw(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	symbol, _ := cmd.Flags().GetString("symbol")

	req := &ebTypes.CfgWithdrawReq{
		Symbol:    symbol,
		ChainName: ethChainName,
	}

	var res ebTypes.WithdrawPara
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.GetCfgWithdraw", req, &res)
	ctx.Run()
}

func ResendEthLockEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resendLockEvent",
		Short: "resend lock Event to chain33 process goroutine",
		Run:   resendLockEvent,
	}
	addResendLockEventFlags(cmd)
	return cmd
}

func addResendLockEventFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("height", "g", 0, "height begin to resend chain33 event ")
	_ = cmd.MarkFlagRequired("height")
	cmd.Flags().Int32P("index", "i", 0, "tx index")
	_ = cmd.MarkFlagRequired("index")
}

func resendLockEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetInt64("height")
	txIndex, _ := cmd.Flags().GetInt32("index")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	resendEthereumEventReq := &ebTypes.ResendEthereumEventReq{
		Height:    height,
		TxIndex:   txIndex,
		ChainName: ethChainName,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ResendEthereumLockEvent", resendEthereumEventReq, &res)
	ctx.Run()
}

func RegetEthLockEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regetLockEvent",
		Short: "reget lock Event to chain33 process goroutine",
		Run:   reGetEthereumEvent,
	}
	addRegetEventFlags(cmd)
	return cmd
}

func addRegetEventFlags(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "s", 0, "height begin")
	_ = cmd.MarkFlagRequired("start")
	cmd.Flags().Int64P("end", "e", 0, "stop height")
	_ = cmd.MarkFlagRequired("end")
}

func reGetEthereumEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetInt64("start")
	stop, _ := cmd.Flags().GetInt64("end")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	regetEthereumEventReq := &ebTypes.RegetEthereumEventReq{
		Start:     start,
		Stop:      stop,
		ChainName: ethChainName,
	}
	fmt.Println("start", regetEthereumEventReq.Start, "stop", regetEthereumEventReq.Stop)

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.ReGetEthereumEvent", regetEthereumEventReq, &res)
	ctx.Run()
}

func CreateLockEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createLockEvent",
		Short: "create lock Event as emitted from ethereum contract",
		Run:   createLockEvent,
	}
	addcreateLockEventFlags(cmd)
	return cmd
}

func addcreateLockEventFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "Ethereum from address")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("receiver", "r", "", "chain33 receiver address")
	_ = cmd.MarkFlagRequired("receiver")

	cmd.Flags().StringP("token", "t", "", "Ethereum token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("symbol", "s", "", "Ethereum token symbol")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("value", "v", "", "value")
	_ = cmd.MarkFlagRequired("value")
	cmd.Flags().StringP("nonce", "n", "", "nonce")
	_ = cmd.MarkFlagRequired("nonce")
}

func createLockEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("receiver")
	token, _ := cmd.Flags().GetString("token")
	symbol, _ := cmd.Flags().GetString("symbol")
	value, _ := cmd.Flags().GetString("value")
	nonce, _ := cmd.Flags().GetString("nonce")
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	createLockEventReq := &ebTypes.CreateLockEventReq{
		From:      from,
		To:        to,
		Token:     token,
		Symbol:    symbol,
		Value:     value,
		Nonce:     nonce,
		ChainName: ethChainName,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.CreateLockEventManually", createLockEventReq, &res)
	ctx.Run()
}

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query call",
		Run:   queryCall,
	}

	cmd.Flags().StringP("address", "a", "", "contract address")
	cmd.MarkFlagRequired("address")
	cmd.Flags().StringP("input", "b", "", "call params (abi format) like foobar(param1,param2)")
	cmd.MarkFlagRequired("input")
	cmd.Flags().StringP("caller", "c", "", "the owner address")
	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	return cmd
}

func queryCall(cmd *cobra.Command, args []string) {
	ethChainName, _ := cmd.Flags().GetString("eth_chain_name")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("address")
	input, _ := cmd.Flags().GetString("input")
	caller, _ := cmd.Flags().GetString("caller")
	path, _ := cmd.Flags().GetString("path")

	if caller == "" {
		caller = addr
	}

	abiFileName := path + addr + ".abi"
	abiStr, err := readFile(abiFileName)
	if nil != err {
		_, _ = fmt.Fprintln(os.Stderr, "Can't read abi info, Pls set correct abi path and provide abi file as", abiFileName)
		return
	}

	queryReq := &ebTypes.QueryReq{
		Param:        input,
		AbiData:      abiStr,
		ContractAddr: addr,
		Owner:        caller,
		ChainName:    ethChainName,
	}

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Manager.EthGeneralQuery", queryReq, &res)
	ctx.Run()
}

func readFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		return "", err
	}

	fileContent, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}

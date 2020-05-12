package main

import (
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	ebTypes "github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/ebrelayer/utils"
	"github.com/33cn/plugin/plugin/dapp/x2Ethereum/types"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
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
		ImportChain33PrivateKeyCmd(),
		ImportEthValidatorPrivateKeyCmd(),
		GenEthPrivateKeyCmd(),
		ShowValidatorsAddrCmd(),
		ShowChain33TxsHashCmd(),
		ShowEthereumTxsHashCmd(),
		ShowEthRelayerStatusCmd(),
		IsValidatorActiveCmd(),
		ShowOperatorCmd(),
		DeployContrctsCmd(),
		ShowTxReceiptCmd(),
		//////auxiliary///////
		CreateBridgeTokenCmd(),
		CreateEthereumTokenCmd(),
		MakeNewProphecyClaimCmd(),
		GetBalanceCmd(),
		IsProphecyPendingCmd(),
		MintErc20Cmd(),
		ApproveCmd(),
		BurnCmd(),
		BurnAsyncCmd(),
		LockSyncCmd(),
		LockAsyncCmd(),
		ShowBridgeBankAddrCmd(),
		ShowBridgeRegistryAddrCmd(),
		StaticsCmd(),
		TransferTokenCmd(),
	)

	return cmd
}

func ImportChain33PrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import_chain33privatekey",
		Short: "import chain33 private key to sign txs to be submitted to chain33",
		Run:   importChain33Privatekey,
	}
	addImportChain33PrivateKeyFlags(cmd)
	return cmd
}

func addImportChain33PrivateKeyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "chain33 private key")
	cmd.MarkFlagRequired("key")
}

func importChain33Privatekey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKey, _ := cmd.Flags().GetString("key")
	params := privateKey

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ImportChain33PrivateKey4EthRelayer", params, &res)
	ctx.Run()
}

func ImportEthValidatorPrivateKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import_ethprivatekey",
		Short: "import ethereum's validator private key ",
		Run:   importEthValidtorPrivatekey,
	}
	addImportPrivateKeyFlags(cmd)
	return cmd
}

func importEthValidtorPrivatekey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKey, _ := cmd.Flags().GetString("key")
	params := privateKey

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ImportEthValidatorPrivateKey", params, &res)
	ctx.Run()
}

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
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.GenerateEthereumPrivateKey", nil, &res)
	ctx.Run()
}

func ShowValidatorsAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_validators",
		Short: "show me the validators including ethereum and chain33",
		Run:   showValidatorsAddr,
	}
	return cmd
}

func showValidatorsAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ValidatorAddr4EthRelayer
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowEthRelayerValidator", nil, &res)
	ctx.Run()
}

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
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowEthRelayer2Chain33Txs", nil, &res)
	if _, err := ctx.RunResult(); nil != err {
		errInfo := err.Error()
		fmt.Println("errinfo:" + errInfo)
		return
	}
	for _, hash := range res.Txhash {
		fmt.Println(hash)
	}
}

func ShowEthereumTxsHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show_eth_tx",
		Short: "show me the ethereum tx hashes",
		Run:   showEthTxs,
	}
	return cmd
}

func showEthTxs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var res ebTypes.Txhashes
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowEthRelayer2Chain33Txs", nil, &res)
	if _, err := ctx.RunResult(); nil != err {
		errInfo := err.Error()
		fmt.Println("errinfo:" + errInfo)
		return
	}
	for _, hash := range res.Txhash {
		fmt.Println(hash)
	}
}

func ShowEthRelayerStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "show ethereum-relayer status",
		Run:   showEthRelayerStatus,
	}
	return cmd
}

func showEthRelayerStatus(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var res ebTypes.RelayerRunStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowEthRelayerStatus", nil, &res)
	ctx.Run()
}

func IsValidatorActiveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active",
		Short: "show whether the validator is active or not",
		Run:   IsValidatorActive,
	}
	IsValidatorActiveFlags(cmd)
	return cmd
}

func IsValidatorActiveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "validator address")
	_ = cmd.MarkFlagRequired("addr")
}

func IsValidatorActive(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := addr
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.IsValidatorActive", params, &res)
	ctx.Run()
}

func ShowOperatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "show me the operator",
		Run:   ShowOperator,
	}
	return cmd
}

func ShowOperator(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowOperator", nil, &res)
	ctx.Run()
}

func DeployContrctsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the corresponding Ethereum contracts",
		Run:   DeployContrcts,
	}
	return cmd
}

func DeployContrcts(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.DeployContrcts", nil, &res)
	ctx.Run()
}

func ShowTxReceiptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "receipt",
		Short: "show me the tx receipt for Ethereum",
		Run:   ShowTxReceipt,
	}
	ShowTxReceiptFlags(cmd)
	return cmd
}

func ShowTxReceiptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "tx hash")
	_ = cmd.MarkFlagRequired("hash")
}

func ShowTxReceipt(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	txhash, _ := cmd.Flags().GetString("hash")
	para := txhash
	var res ethTypes.Receipt
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowTxReceipt", para, &res)
	ctx.Run()
}

func CreateBridgeTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token4chain33",
		Short: "create new token as chain33 asset on Ethereum",
		Run:   CreateBridgeToken,
	}
	CreateBridgeTokenFlags(cmd)
	return cmd
}

func CreateBridgeTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
}

func CreateBridgeToken(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	token, _ := cmd.Flags().GetString("symbol")
	para := token
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.CreateBridgeToken", para, &res)
	ctx.Run()
}

func CreateEthereumTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token4erc20",
		Short: "create new erc20 asset on Ethereum",
		Run:   CreateEthereumTokenToken,
	}
	CreateEthereumTokenFlags(cmd)
	return cmd
}

func CreateEthereumTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
}

func CreateEthereumTokenToken(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	token, _ := cmd.Flags().GetString("symbol")
	para := token
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.CreateERC20Token", para, &res)
	ctx.Run()
}

func MintErc20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "mint erc20 asset on Ethereum, but only for operator",
		Run:   MintErc20,
	}
	MintErc20Flags(cmd)
	return cmd
}

func MintErc20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}
func MintErc20(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")
	owner, _ := cmd.Flags().GetString("owner")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error")
		return
	}

	realAmount := types.ToWei(amount, d)
	para := ebTypes.MintToken{
		Owner:     owner,
		TokenAddr: tokenAddr,
		Amount:    realAmount.String(),
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.MintErc20", para, &res)
	ctx.Run()
}

func ApproveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve",
		Short: "approve the allowance to bridgebank by the owner",
		Run:   ApproveAllowance,
	}
	ApproveAllowanceFlags(cmd)
	return cmd
}

func ApproveAllowanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
}

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

	realAmount := types.ToWei(amount, d)
	para := ebTypes.ApproveAllowance{
		OwnerKey:  key,
		TokenAddr: tokenAddr,
		Amount:    realAmount.String(),
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ApproveAllowance", para, &res)
	ctx.Run()
}

func BurnCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "burn(including approve) the asset to make it unlocked on chain33",
		Run:   Burn,
	}
	BurnFlags(cmd)
	return cmd
}

func BurnAsyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn-async",
		Short: "async burn the asset to make it unlocked on chain33",
		Run:   BurnAsync,
	}
	BurnFlags(cmd)
	return cmd
}

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

func Burn(cmd *cobra.Command, args []string) {
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
		Amount:          types.ToWei(amount, d).String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.Burn", para, &res)
	ctx.Run()
}

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
		Amount:          types.ToWei(amount, d).String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.BurnAsync", para, &res)
	ctx.Run()
}

func LockSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock",
		Short: "lock(including approve) eth or erc20 and cross-chain transfer to chain33",
		Run:   LockEthErc20Asset,
	}
	LockEthErc20AssetFlags(cmd)
	return cmd
}

func LockAsyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-async",
		Short: "async lock eth or erc20 and cross-chain transfer to chain33",
		Run:   LockEthErc20AssetAsync,
	}
	LockEthErc20AssetFlags(cmd)
	return cmd
}

func LockEthErc20AssetFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "owner private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("token", "t", "", "token address, optional, nil for ETH")
	cmd.Flags().Float64P("amount", "m", float64(0), "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("receiver", "r", "", "chain33 receiver address")
	_ = cmd.MarkFlagRequired("receiver")
}

func LockEthErc20Asset(cmd *cobra.Command, args []string) {
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

	realAmount := types.ToWei(amount, d)

	para := ebTypes.LockEthErc20{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.LockEthErc20Asset", para, &res)
	ctx.Run()
}

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

	realAmount := types.ToWei(amount, d)

	para := ebTypes.LockEthErc20{
		OwnerKey:        key,
		TokenAddr:       tokenAddr,
		Amount:          realAmount.String(),
		Chain33Receiver: receiver,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.LockEthErc20AssetAsync", para, &res)
	ctx.Run()
}

func ShowBridgeBankAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridgeBankAddr",
		Short: "show the address of Contract BridgeBank",
		Run:   ShowBridgeBankAddr,
	}
	return cmd
}

func ShowBridgeBankAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowBridgeBankAddr", nil, &res)
	ctx.Run()
}

func ShowBridgeRegistryAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bridgeRegistry",
		Short: "show the address of Contract BridgeRegistry",
		Run:   ShowBridgeRegistryAddr,
	}
	return cmd
}

func ShowBridgeRegistryAddr(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res ebTypes.ReplyAddr
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.ShowBridgeRegistryAddr", nil, &res)
	ctx.Run()
}

func MakeNewProphecyClaimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prophecy",
		Short: "Make New Prophecy Claim",
		Run:   MakeNewProphecyClaim,
	}
	MakeNewProphecyClaimFlags(cmd)
	return cmd
}

func MakeNewProphecyClaimFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32P("claim", "c", uint32(1), "claim type, 1 denote burn, and 2 denotes lock")
	_ = cmd.MarkFlagRequired("claim")
	cmd.Flags().StringP("chain33Sender", "a", "", "Chain33Sender")
	_ = cmd.MarkFlagRequired("chain33Sender")
	cmd.Flags().StringP("token", "t", "", "token address,optional, nil for ETH")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("ethReceiver", "r", "", "eth Receiver")
	_ = cmd.MarkFlagRequired("ethReceiver")
	cmd.Flags().Float64P("amount", "m", 0, "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("hash", "i", "", "chain33 tx hash")
	_ = cmd.MarkFlagRequired("hash")

}

func MakeNewProphecyClaim(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	claimType, _ := cmd.Flags().GetUint32("claim")
	if claimType != uint32(1) && claimType != uint32(2) {
		fmt.Println("Wrong claim type")
		return
	}
	chain33Sender, _ := cmd.Flags().GetString("chain33Sender")
	tokenAddr, _ := cmd.Flags().GetString("token")
	symbol, _ := cmd.Flags().GetString("symbol")
	ethReceiver, _ := cmd.Flags().GetString("ethReceiver")
	amount, _ := cmd.Flags().GetFloat64("amount")
	txhash, _ := cmd.Flags().GetString("hash")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error")
		return
	}

	realAmount := types.ToWei(amount, d)

	para := ebTypes.NewProphecyClaim{
		ClaimType:     claimType,
		Chain33Sender: chain33Sender,
		TokenAddr:     tokenAddr,
		Symbol:        symbol,
		EthReceiver:   ethReceiver,
		Amount:        realAmount.String(),
		TxHash:        txhash,
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.MakeNewProphecyClaim", para, &res)
	ctx.Run()
}

func GetBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "get owner's balance for ETH or ERC20",
		Run:   GetBalance,
	}
	GetBalanceFlags(cmd)
	return cmd
}

func GetBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("tokenAddr", "t", "", "token address, optional, nil for Eth")
}

func GetBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	owner, _ := cmd.Flags().GetString("owner")
	tokenAddr, _ := cmd.Flags().GetString("tokenAddr")

	para := ebTypes.BalanceAddr{
		Owner:     owner,
		TokenAddr: tokenAddr,
	}
	var res ebTypes.ReplyBalance
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.GetBalance", para, &res)
	ctx.Run()
}

func IsProphecyPendingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ispending",
		Short: "check whether the Prophecy is pending or not",
		Run:   IsProphecyPending,
	}
	IsProphecyPendingFlags(cmd)
	return cmd
}

func IsProphecyPendingFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("id", "i", "", "claim prophecy id")
	_ = cmd.MarkFlagRequired("id")
}

func IsProphecyPending(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	id, _ := cmd.Flags().GetString("id")
	para := common.HexToHash(id)

	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.IsProphecyPending", para, &res)
	ctx.Run()
}

func TransferTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "create a transfer transaction",
		Run:   TransferToken,
	}
	TransferTokenFlags(cmd)
	return cmd
}

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

func TransferToken(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenAddr, _ := cmd.Flags().GetString("token")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	amount, _ := cmd.Flags().GetFloat64("amount")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	d, err := utils.GetDecimalsFromNode(tokenAddr, nodeAddr)
	if err != nil {
		fmt.Println("get decimals error")
		return
	}

	realAmount := types.ToWei(amount, d)
	para := ebTypes.TransferToken{
		TokenAddr: tokenAddr,
		FromKey:   from,
		ToAddr:    to,
		Amount:    realAmount.String(),
	}
	var res rpctypes.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "RelayerManager.TransferToken", para, &res)
	ctx.Run()
}

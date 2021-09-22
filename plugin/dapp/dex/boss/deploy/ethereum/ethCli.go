package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/33cn/plugin/plugin/dapp/dex/boss/deploy/ethereum/offline"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

func EthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum",
		Short: "ethereumcake command",
	}
	cmd.AddCommand(
		CakeCmd(),
		FarmCmd(),
		GetBalanceCmd(),
		Erc20Cmd(),
	)
	return cmd
}

//Erc20Cmd

func Erc20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20",
		Short: "deploy erc20 contract",
		Run:   deplayErc20,
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
	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().Uint8P("decimals", "d", 18, "token decimals")
	_ = cmd.MarkFlagRequired("decimals")

}

func deplayErc20(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	owner, _ := cmd.Flags().GetString("owner")
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetString("amount")
	decimals, _ := cmd.Flags().GetUint8("decimals")
	key, _ := cmd.Flags().GetString("key")
	fmt.Println("owner", owner, "name", name, "symbol:", symbol, "amount", amount)
	priv, err := crypto.ToECDSA(common.FromHex(key))
	if nil != err {
		panic("Failed to recover private key")
	}
	deployFrom := crypto.PubkeyToAddress(priv.PublicKey)

	client, err := ethclient.Dial(rpcLaddr)
	if nil != err {
		panic(err)
	}
	ctx := context.Background()
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		panic(err)
	}

	nonce, err := client.PendingNonceAt(ctx, deployFrom)
	if nil != err {
		fmt.Println("err:", err)
	}
	contractAddr := crypto.CreateAddress(deployFrom, nonce)
	signedtx, hash, err := rewriteDeployErc20(owner, name, symbol, amount, decimals, nonce, gasPrice, priv)
	if err != nil {
		panic(err)
	}
	var tx = new(types.Transaction)
	err = tx.UnmarshalBinary(common.FromHex(signedtx))
	if err != nil {
		panic(err)
	}

	err = client.SendTransaction(ctx, tx)
	if nil != err {
		fmt.Println("err:", err)
	}

	fmt.Println("success deploy erc20:", symbol, "contract,ContractAddress", contractAddr, "txhash:", hash)
}

func rewriteDeployErc20(owner, name, symbol, amount string, decimals uint8, nonce uint64, gasPrice *big.Int, key *ecdsa.PrivateKey) (signedTx, hash string, err error) {
	erc20OwnerAddr := common.HexToAddress(owner)
	bn := big.NewInt(1)
	supply, ok := bn.SetString(utils.TrimZeroAndDot(amount), 10)
	if !ok {
		panic("amount format err")
	}
	parsed, err := abi.JSON(strings.NewReader(generated.ERC20ABI))
	if err != nil {
		return
	}

	erc20Bin := common.FromHex(generated.ERC20Bin)
	packdata, err := parsed.Pack("", name, symbol, supply, erc20OwnerAddr, decimals)
	if err != nil {
		panic(err)
	}
	input := append(erc20Bin, packdata...)
	var gasLimit = 100 * 10000
	tx := types.NewContractCreation(nonce, big.NewInt(0), uint64(gasLimit), gasPrice, input)
	return offline.SignTx(key, tx)
}

//GetBalanceCmd ...
func GetBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "get owner's balance for ETH or ERC20",
		Run:   ShowBalance,
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
func ShowBalance(cmd *cobra.Command, args []string) {
	owner, _ := cmd.Flags().GetString("owner")
	tokenAddr, _ := cmd.Flags().GetString("tokenAddr")
	ethNodeAddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")

	setupWebsocketEthClient(ethNodeAddr)
	balance, err := GetBalance(tokenAddr, owner)
	if nil != err {
		fmt.Println("err:", err.Error())
	}
	fmt.Println("balance =", balance)
}

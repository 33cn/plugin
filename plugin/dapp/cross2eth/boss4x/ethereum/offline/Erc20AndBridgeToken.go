package offline

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	bep20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/bep20/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	tetherUSDT "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/usdt/generated"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

/*
./boss4x ethereum offline create_erc20 -m 33000000000000000000 -s YCC -o 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a
./boss4x ethereum offline sign -f deployErc20YCC.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt

./boss4x ethereum offline create_add_lock_list -s YCC -t 0x20a32A5680EBf55740B0C98B54cDE8e6FD5a4FB0 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a
./boss4x ethereum offline sign -f create_add_lock_list.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt

./boss4x ethereum offline create_bridge_token -s YCC -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a
./boss4x ethereum offline sign -f create_bridge_token.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt
*/

func DeployBEP20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy_bep20",
		Short: "deploy BEP20 contracts",
		Run:   DeployBEP20,
	}
	DeployBEP20Flags(cmd)
	return cmd
}

func DeployBEP20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployAddr", "a", "", "addr to deploy contract ")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("symbol", "s", "", "BEP20 symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("totalSupply", "m", "0", "total supply")
	_ = cmd.MarkFlagRequired("totalSupply")
	cmd.Flags().IntP("decimal", "d", 8, "decimal")
	_ = cmd.MarkFlagRequired("decimal")

}

func DeployBEP20(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	deployerAddr, _ := cmd.Flags().GetString("deployAddr")
	owner, _ := cmd.Flags().GetString("owner")
	symbol, _ := cmd.Flags().GetString("symbol")
	totalSupply, _ := cmd.Flags().GetString("totalSupply")
	decimal, _ := cmd.Flags().GetInt("decimal")
	bnAmount := big.NewInt(1)
	bnAmount, _ = bnAmount.SetString(utils.TrimZeroAndDot(totalSupply), 10)
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("ethclient Dial error", err.Error())
		return
	}
	symbol = strings.ToUpper(symbol)

	ctx := context.Background()
	startNonce, err := client.PendingNonceAt(ctx, common.HexToAddress(deployerAddr))
	if nil != err {
		fmt.Println("PendingNonceAt error", err.Error())
		return
	}

	var infos []*DeployInfo

	parsed, err := abi.JSON(strings.NewReader(bep20.BEP20TokenABI))
	if err != nil {
		fmt.Println("abi.JSON(strings.NewReader(erc20.ERC20ABI)) error", err.Error())
		return
	}
	bin := common.FromHex(bep20.BEP20TokenBin)
	BEP20OwnerAddr := common.HexToAddress(owner)
	//constructor (string memory name_, string memory symbol_,uint256 totalSupply_, uint8 decimals_, address owner_) public {
	tokenName := symbol + " Token"
	packdata, err := parsed.Pack("", tokenName, symbol, bnAmount, uint8(decimal), BEP20OwnerAddr)
	if err != nil {
		fmt.Println("Pack error", err.Error())
		return
	}
	BEP20Addr := crypto.CreateAddress(common.HexToAddress(deployerAddr), startNonce)
	deployInfo := DeployInfo{
		PackData:       append(bin, packdata...),
		ContractorAddr: BEP20Addr,
		Name:           "BEP20: " + symbol,
		Nonce:          startNonce,
		To:             nil,
	}
	infos = append(infos, &deployInfo)
	fileName := fmt.Sprintf("deployBEP20%s.txt", symbol)
	err = NewTxWrite(infos, common.HexToAddress(deployerAddr), url, fileName)
	if err != nil {
		fmt.Println("NewTxWrite error", err.Error())
		return
	}
}

func DeployERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_erc20",
		Short: "create ERC20 contracts",
		Run:   DeployERC20,
	}
	DeployERC20Flags(cmd)
	return cmd
}

func DeployERC20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("symbol", "s", "", "erc20 symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("amount", "m", "0", "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func DeployERC20(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	deployerAddr, _ := cmd.Flags().GetString("deployAddr")
	owner, _ := cmd.Flags().GetString("owner")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetString("amount")
	bnAmount := big.NewInt(1)
	bnAmount, _ = bnAmount.SetString(utils.TrimZeroAndDot(amount), 10)
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("ethclient Dial error", err.Error())
		return
	}

	ctx := context.Background()
	startNonce, err := client.PendingNonceAt(ctx, common.HexToAddress(deployerAddr))
	if nil != err {
		fmt.Println("PendingNonceAt error", err.Error())
		return
	}

	var infos []*DeployInfo

	parsed, err := abi.JSON(strings.NewReader(erc20.ERC20ABI))
	if err != nil {
		fmt.Println("abi.JSON(strings.NewReader(erc20.ERC20ABI)) error", err.Error())
		return
	}
	bin := common.FromHex(erc20.ERC20Bin)
	Erc20OwnerAddr := common.HexToAddress(owner)
	packdata, err := parsed.Pack("", symbol, symbol, bnAmount, Erc20OwnerAddr, uint8(8))
	if err != nil {
		fmt.Println("Pack error", err.Error())
		return
	}
	Erc20Addr := crypto.CreateAddress(common.HexToAddress(deployerAddr), startNonce)
	deployInfo := DeployInfo{
		PackData:       append(bin, packdata...),
		ContractorAddr: Erc20Addr,
		Name:           "Erc20: " + symbol,
		Nonce:          startNonce,
		To:             nil,
	}
	infos = append(infos, &deployInfo)
	fileName := fmt.Sprintf("deployErc20%s.txt", symbol)
	err = NewTxWrite(infos, common.HexToAddress(deployerAddr), url, fileName)
	if err != nil {
		fmt.Println("NewTxWrite error", err.Error())
		return
	}
}

func DeployTetherUSDTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_tether_usdt",
		Short: "create tether usdt contracts",
		Run:   DeployTetherUSDT,
	}
	DeployTetherUSDTFlags(cmd)
	return cmd
}

func DeployTetherUSDTFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("symbol", "s", "", "erc20 symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("amount", "m", "0", "amount")
	_ = cmd.MarkFlagRequired("amount")
}

func DeployTetherUSDT(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	deployerAddr, _ := cmd.Flags().GetString("deployAddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetString("amount")
	bnAmount := big.NewInt(1e6 * 1e6)
	bnAmount, _ = bnAmount.SetString(utils.TrimZeroAndDot(amount), 10)
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("ethclient Dial error", err.Error())
		return
	}

	ctx := context.Background()
	startNonce, err := client.PendingNonceAt(ctx, common.HexToAddress(deployerAddr))
	if nil != err {
		fmt.Println("PendingNonceAt error", err.Error())
		return
	}

	var infos []*DeployInfo

	parsed, err := abi.JSON(strings.NewReader(tetherUSDT.TetherTokenABI))
	if err != nil {
		fmt.Println("abi.JSON(strings.NewReader(tetherUSDT.TetherTokenABI)) error", err.Error())
		return
	}
	bin := common.FromHex(tetherUSDT.TetherTokenBin)
	//function TetherToken(uint _initialSupply, string _name, string _symbol, uint _decimals)
	packdata, err := parsed.Pack("", bnAmount, symbol, symbol, big.NewInt(6))
	if err != nil {
		fmt.Println("Pack error", err.Error())
		return
	}
	Erc20Addr := crypto.CreateAddress(common.HexToAddress(deployerAddr), startNonce)
	deployInfo := DeployInfo{
		PackData:       append(bin, packdata...),
		ContractorAddr: Erc20Addr,
		Name:           "Erc20: " + symbol,
		Nonce:          startNonce,
		To:             nil,
	}
	infos = append(infos, &deployInfo)
	fileName := "deployTetherUSDT.txt"
	err = NewTxWrite(infos, common.HexToAddress(deployerAddr), url, fileName)
	if err != nil {
		fmt.Println("NewTxWrite error", err.Error())
		return
	}
}

func CreateAddToken2LockListTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_add_lock_list",
		Short: "add token to lock list",
		Run:   AddToken2LockListTx,
	}
	AddToken2LockListTxFlags(cmd)
	return cmd
}

func AddToken2LockListTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("token", "t", "", "token addr")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
}

func AddToken2LockListTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	symbol, _ := cmd.Flags().GetString("symbol")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	token, _ := cmd.Flags().GetString("token")
	contract, _ := cmd.Flags().GetString("contract")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")

	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		fmt.Println("JSON NewReader Err:", err)
		return
	}

	abiData, err := bridgeAbi.Pack("addToken2LockList", common.HexToAddress(token), symbol)
	if err != nil {
		fmt.Println("bridgeAbi.Pack addToken2LockList Err:", err)
		return
	}

	CreateTxInfoAndWrite(abiData, deployAddr, contract, "create_add_lock_list", url, chainEthId)
}

func CreateBridgeTokenTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_bridge_token",
		Short: "create new token as chain33 asset on Ethereum, and it's should be done by operator",
		Run:   CreateBridgeTokenTx,
	}
	CreateBridgeTokenTxFlags(cmd)
	return cmd
}

func CreateBridgeTokenTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
}

func CreateBridgeTokenTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	symbol, _ := cmd.Flags().GetString("symbol")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	contract, _ := cmd.Flags().GetString("contract")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")

	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		fmt.Println("JSON NewReader Err:", err)
		return
	}

	abiData, err := bridgeAbi.Pack("createNewBridgeToken", symbol)
	if err != nil {
		fmt.Println("bridgeAbi.Pack createNewBridgeToken Err:", err)
		return
	}
	CreateTxInfoAndWrite(abiData, deployAddr, contract, "create_bridge_token", url, chainEthId)
}

package offline

import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4chain33/generated"
	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline create_erc20 -s YCC -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -o 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ --chainID 33
./boss4x chain33 offline send -f deployErc20YCCChain33.txt

./boss4x chain33 offline approve_erc20 -a 330000000000 -s 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae  --chainID 33
./boss4x chain33 offline send -f approve_erc20.txt

./boss4x chain33 offline create_add_lock_list -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -t 1998HqVnt4JUirhC9KL5V71xYU8cFRn82c --chainID 33 -s YCC
./boss4x chain33 offline send -f create_add_lock_list.txt

./boss4x chain33 offline create_bridge_token -c 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -s YCC --chainID 33
./boss4x chain33 offline send -f create_bridge_token.txt
${Chain33Cli} evm abi call -a "${chain33BridgeBank}" -c "${chain33DeployAddr}" -b "getToken2address(YCC)"
./chain33-cli evm abi call -a 1JmWVu1GEdQYSN1opxS9C39aS4NvG57yTr -c 1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ -b 'getToken2address(YCC)'
*/

func CreateERC20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_erc20",
		Short: "create erc20 contracts and sign, default 3300*1e8 to be minted",
		Run:   CreateERC20,
	}
	CreateERC20Flags(cmd)
	return cmd
}

//CreateERC20Flags ...
func CreateERC20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("owner", "o", "", "owner address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().Float64P("amount", "a", 0, "amount to be minted(optional),default to 3300*1e8")
}

func CreateERC20(cmd *cobra.Command, _ []string) {
	symbol, _ := cmd.Flags().GetString("symbol")
	owner, _ := cmd.Flags().GetString("owner")
	amount, _ := cmd.Flags().GetFloat64("amount")
	amountInt64 := int64(3300 * 1e8)
	if 0 != int64(amount) {
		amountInt64 = int64(amount)
	}

	privateKeyStr, _ := cmd.Flags().GetString("key")
	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	fromAddr := address.PubKeyToAddress(privateKey.PubKey().Bytes())
	from := common.Address{
		Addr: fromAddr,
	}

	createPara := fmt.Sprintf("%s,%s,%s,%s", symbol, symbol, fmt.Sprintf("%d", amountInt64), owner)
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), erc20.ERC20Bin, erc20.ERC20ABI, createPara, "ERC20:"+symbol)
	if nil != err {
		fmt.Println("CreateContractAndSign erc20 fail")
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	Erc20Tx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy ERC20:" + symbol,
	}

	data, err := json.MarshalIndent(Erc20Tx, "", "    ")
	if err != nil {
		fmt.Println("MarshalIndent error", err.Error())
		return
	}
	fmt.Println(string(data))

	var txs []*utils.Chain33OfflineTx
	txs = append(txs, Erc20Tx)

	fileName := fmt.Sprintf("deployErc20%sChain33.txt", symbol)
	fmt.Printf("Write all the txs to file:   %s \n", fileName)
	utils.WriteToFileInJson(fileName, txs)
}

func ApproveErc20Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approve_erc20",
		Short: "approve erc20",
		Run:   ApproveErc20, //配置账户
	}
	addApproveErc20Flags(cmd)
	return cmd
}

func addApproveErc20Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("approve", "s", "", "approve addr")
	_ = cmd.MarkFlagRequired("approve")
	cmd.Flags().Float64P("amount", "a", 0, "approve amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("contract", "c", "", "Erc20 contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func ApproveErc20(cmd *cobra.Command, _ []string) {
	contract, _ := cmd.Flags().GetString("contract")
	approve, _ := cmd.Flags().GetString("approve")
	amount, _ := cmd.Flags().GetFloat64("amount")

	parameter := fmt.Sprintf("approve(%s,%d)", approve, int64(amount))
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeTokenABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, contract, "approve_erc20")
}

func AddToken2LockListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_add_lock_list",
		Short: "add token to lock list",
		Run:   AddToken2LockList, //配置账户
	}
	addAddToken2LockListFlags(cmd)
	return cmd
}

func addAddToken2LockListFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("token", "t", "", "token addr")
	_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func AddToken2LockList(cmd *cobra.Command, _ []string) {
	contract, _ := cmd.Flags().GetString("contract")
	symbol, _ := cmd.Flags().GetString("symbol")
	token, _ := cmd.Flags().GetString("token")

	parameter := fmt.Sprintf("addToken2LockList(%s,%s)", token, symbol)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, contract, "create_add_lock_list")
}

func CreateNewBridgeTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_bridge_token",
		Short: "create new token as ethereum asset on chain33, and it's should be done by operator",
		Run:   CreateNewBridgeToken, //配置账户
	}
	addCreateNewBridgeTokenFlags(cmd)
	return cmd
}

func addCreateNewBridgeTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func CreateNewBridgeToken(cmd *cobra.Command, _ []string) {
	contract, _ := cmd.Flags().GetString("contract")
	symbol, _ := cmd.Flags().GetString("symbol")

	parameter := fmt.Sprintf("createNewBridgeToken(%s)", symbol)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	callContractAndSignWrite(cmd, packData, contract, "create_bridge_token")
}

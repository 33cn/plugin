package offline

import (
	"fmt"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeFactory"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-swap-periphery/src/pancakeRouter"
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

	erc20 "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/erc20/generated"

	"github.com/spf13/cobra"
)

// 创建ERC20合约
func createERC20ContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "erc20",
		Short: "create ERC20 contract",
		Run:   createERC20Contract,
	}
	addCreateERC20ContractFlags(cmd)
	return cmd
}

func addCreateERC20ContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the caller's private key")
	cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("name", "a", "", "REC20 name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("symbol", "s", "", "REC20 symbol")
	cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("supply", "m", "", "REC20 supply")
	cmd.MarkFlagRequired("supply")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func createERC20Contract(cmd *cobra.Command, args []string) {
	privateKeyStr, _ := cmd.Flags().GetString("key")
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	supply, _ := cmd.Flags().GetString("supply")

	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4

	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())

	info := &utils.TxCreateInfo{
		PrivateKey: privateKeyStr,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	createPara := name + "," + symbol + "," + supply + "," + from.String()
	content, txHash, err := utils.CreateContractAndSign(info, erc20.ERC20Bin, erc20.ERC20ABI, createPara, "erc20")
	if nil != err {
		fmt.Println("Failed to create erc20 due to cause:", err.Error())
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	routerTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy router",
		Interval:      time.Second * 5,
	}

	utils.WriteToFileInJson("./erc20.txt", routerTx)
}

func createRouterContract(cmd *cobra.Command, args []string) {
	privateKeyStr, _ := cmd.Flags().GetString("key")

	var txs []*utils.Chain33OfflineTx
	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())
	i := 1
	fmt.Printf("%d: Going to create factory\n", i)
	i += 1
	factoryTx, err := createFactoryContract(cmd, from)
	if nil != err {
		fmt.Println("Failed to createValsetTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, factoryTx)

	fmt.Printf("%d: Going to create weth9\n", i)
	i += 1
	weth9Tx, err := createWeth9(cmd, from)
	if nil != err {
		fmt.Println("Failed to createWeth9 due to cause:", err.Error())
		return
	}
	txs = append(txs, weth9Tx)

	fmt.Printf("%d: Going to create router\n", i)
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKeyStr,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	//constructor(address _factory, address _WETH)
	createPara := factoryTx.ContractAddr + "," + weth9Tx.ContractAddr
	content, txHash, err := utils.CreateContractAndSign(info, pancakeRouter.PancakeRouterBin, pancakeRouter.PancakeRouterABI, createPara, "pancakeRouter")
	if nil != err {
		fmt.Println("Failed to create router due to cause:", err.Error())
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	routerTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy router",
		Interval:      time.Second * 5,
	}
	txs = append(txs, routerTx)
	utils.WriteToFileInJson("./router.txt", txs)
}

func createWeth9(cmd *cobra.Command, from common.Address) (*utils.Chain33OfflineTx, error) {
	privateKeyStr, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKeyStr,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	createPara := ""
	content, txHash, err := utils.CreateContractAndSign(info, pancakeRouter.WETH9Bin, pancakeRouter.WETH9ABI, createPara, "WETH9")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	weth9Tx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy weth9",
		Interval:      time.Second * 5,
	}
	return weth9Tx, nil
}

func createFactoryContract(cmd *cobra.Command, from common.Address) (*utils.Chain33OfflineTx, error) {
	feeToSetter, _ := cmd.Flags().GetString("feeToSetter")

	privateKeyStr, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKeyStr,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	//constructor(address _feeToSetter) public
	createPara := feeToSetter
	content, txHash, err := utils.CreateContractAndSign(info, pancakeFactory.PancakeFactoryBin, pancakeFactory.PancakeFactoryABI, createPara, "PancakeFactory")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	factoryTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy factory",
		Interval:      time.Second * 5,
	}
	return factoryTx, nil
}

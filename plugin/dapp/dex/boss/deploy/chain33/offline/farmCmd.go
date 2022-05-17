package offline

import (
	"fmt"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"

	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"

	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/cakeToken"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/masterChef"
	"github.com/33cn/plugin/plugin/dapp/dex/contracts/pancake-farm/src/syrupBar"
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	"github.com/spf13/cobra"
)

func farmofflineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "farm",
		Short: "create and sign tx to deploy farm and set lp, transfer ownership",
	}
	cmd.AddCommand(
		createMasterChefCmd(),
		AddPoolCmd(),
		updateAllocPointCmd(),
	)
	return cmd
}

func createCakeToken(cmd *cobra.Command, from common.Address) (*utils.Chain33OfflineTx, error) {
	privateKey, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKey,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	createPara := ""
	content, txHash, err := utils.CreateContractAndSign(info, cakeToken.CakeTokenBin, cakeToken.CakeTokenABI, createPara, "cakeToken")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	cakeTokenTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy cakeToken",
		Interval:      time.Second * 5,
	}
	return cakeTokenTx, nil
}

func createSyrupBar(cmd *cobra.Command, from common.Address, cakeToken string) (*utils.Chain33OfflineTx, error) {
	privateKey, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKey,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}
	//constructor(CakeToken _cake)
	createPara := cakeToken
	content, txHash, err := utils.CreateContractAndSign(info, syrupBar.SyrupBarBin, syrupBar.SyrupBarABI, createPara, "syrupBar")
	if nil != err {
		fmt.Println("Failed to create SyrupBar due to cause:", err.Error())
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	cakeTokenTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy SyrupBar",
		Interval:      time.Second * 5,
	}
	return cakeTokenTx, nil
}

func TransferOwnerShip(cmd *cobra.Command, from common.Address, masterChef, contractAddr, operationName string) (*utils.Chain33OfflineTx, error) {
	privateKey, _ := cmd.Flags().GetString("key")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	info := &utils.TxCreateInfo{
		PrivateKey: privateKey,
		Expire:     expire,
		Note:       note,
		Fee:        feeInt64,
		ParaName:   paraName,
		ChainID:    chainID,
	}

	//function transferOwnership(address newOwner) public onlyOwner
	parameter := fmt.Sprintf("transferOwnership(%s)", masterChef)
	_, packData, err := evmAbi.Pack(parameter, syrupBar.OwnableABI, false)
	if nil != err {
		fmt.Println("TransferOwnerShip", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}

	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: contractAddr}
	content, txHash, err := utils.CallContractAndSign(info, action, contractAddr)
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	transferOwnerShipTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: operationName,
		Interval:      time.Second * 5,
	}
	return transferOwnerShipTx, nil
}

func createMasterChefCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "masterChef",
		Short: "create masterChef contract",
		Run:   createMasterChef,
	}
	addCreateMasterChefFlags(cmd)
	return cmd
}

func addCreateMasterChefFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the caller's private key")
	cmd.MarkFlagRequired("key")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")

	cmd.Flags().StringP("devaddr", "d", "", "address of develop")
	cmd.MarkFlagRequired("devaddr")
	cmd.Flags().Int64P("cakePerBlock", "m", 0, "cake Per Block, should multiply 1e18")
	cmd.MarkFlagRequired("cakePerBlock")
	cmd.Flags().Int64P("startBlock", "s", 0, "start Block height")
	cmd.MarkFlagRequired("startBlock")
}

func createMasterChef(cmd *cobra.Command, args []string) {
	privateKeyStr, _ := cmd.Flags().GetString("key")

	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())

	var txs []*utils.Chain33OfflineTx
	i := 1
	fmt.Printf("%d: Going to create cake token\n", i)
	i += 1
	cakeTokenTx, err := createCakeToken(cmd, from)
	if nil != err {
		fmt.Println("Failed to create cake token due to cause:", err.Error())
		return
	}
	txs = append(txs, cakeTokenTx)

	fmt.Printf("%d: Going to create SyrupBar\n", i)
	i += 1
	syrupBarTx, err := createSyrupBar(cmd, from, cakeTokenTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to create syrupBar due to cause:", err.Error())
		return
	}
	txs = append(txs, syrupBarTx)

	fmt.Printf("%d: Going to create materchef\n", i)
	i += 1
	devaddr, _ := cmd.Flags().GetString("devaddr")
	cakePerBlock, _ := cmd.Flags().GetInt64("cakePerBlock")
	startBlock, _ := cmd.Flags().GetInt64("startBlock")

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
	createPara := fmt.Sprintf("%s,%s,%s,%d,%d", cakeTokenTx.ContractAddr, syrupBarTx.ContractAddr, devaddr, cakePerBlock, startBlock)
	content, txHash, err := utils.CreateContractAndSign(info, masterChef.MasterChefBin, masterChef.MasterChefABI, createPara, "masterChef")
	if nil != err {
		fmt.Println("Failed to create master chef due to cause:", err.Error())
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	masterChefTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy master Chef",
		Interval:      time.Second * 5,
	}
	txs = append(txs, masterChefTx)

	fmt.Printf("%d: Going to transfer OwnerShip from cake token to masterchef\n", i)
	i += 1
	cakeTokenOwnerShipTransferTx, err := TransferOwnerShip(cmd, from, masterChefTx.ContractAddr, cakeTokenTx.ContractAddr, "transfer cakeToken's ownership to masterchef")
	if nil != err {
		fmt.Println("Failed to Transfer OwnerShip from cake token:", err.Error())
		return
	}
	txs = append(txs, cakeTokenOwnerShipTransferTx)

	fmt.Printf("%d: Going to transfer OwnerShip from syrupBar to masterchef\n", i)
	//i += 1
	syrupBarOwnerShipTransferTx, err := TransferOwnerShip(cmd, from, masterChefTx.ContractAddr, syrupBarTx.ContractAddr, "transfer syrupBar's ownership to masterchef")
	if nil != err {
		fmt.Println("Failed to Transfer OwnerShip:", err.Error())
		return
	}
	txs = append(txs, syrupBarOwnerShipTransferTx)

	utils.WriteToFileInJson("./farm.txt", txs)
}

func AddPoolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addPool",
		Short: "add lp to pool",
		Run:   addPool,
	}
	addPoolFlags(cmd)
	return cmd
}

func addPoolFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().StringP("lptoken", "l", "", "lp Addr ")
	_ = cmd.MarkFlagRequired("lptoken")

	cmd.Flags().Int64P("alloc", "p", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "(Optional)with update")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.MarkFlagRequired("fee")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func addPool(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	lpToken, _ := cmd.Flags().GetString("lptoken")
	update, _ := cmd.Flags().GetBool("update")

	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	privateKeyStr, _ := cmd.Flags().GetString("key")

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
	parameter := fmt.Sprintf("add(%d, %s, %v)", allocPoint, lpToken, update)
	_, packData, err := evmAbi.Pack(parameter, masterChef.MasterChefABI, false)
	if nil != err {
		fmt.Println("AddPool2FarmHandle", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: masterChefAddrStr}
	content, txHash, err := utils.CallContractAndSign(info, action, masterChefAddrStr)
	if nil != err {
		fmt.Println("Failed to create master chef due to cause:", err.Error())
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	addPoolTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "add pool",
		Interval:      time.Second * 5,
	}
	file := fmt.Sprintf("./addPool_%s_%d.txt", lpToken, allocPoint)
	utils.WriteToFileInJson(file, addPoolTx)
}

func updateAllocPointCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "updateAllocPoint",
		Short: "update Alloc Point",
		Run:   updateAllocPoint,
	}
	addUpdateAllocPointFlags(cmd)
	return cmd
}

func addUpdateAllocPointFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("masterchef", "m", "", "master Chef Addr ")
	_ = cmd.MarkFlagRequired("masterchef")

	cmd.Flags().Int64P("pid", "d", 0, "id of pool")
	_ = cmd.MarkFlagRequired("pid")

	cmd.Flags().Int64P("alloc", "p", 0, "allocation point ")
	_ = cmd.MarkFlagRequired("alloc")

	cmd.Flags().BoolP("update", "u", true, "with update")
	_ = cmd.MarkFlagRequired("update")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func updateAllocPoint(cmd *cobra.Command, args []string) {
	masterChefAddrStr, _ := cmd.Flags().GetString("masterchef")
	pid, _ := cmd.Flags().GetInt64("pid")
	allocPoint, _ := cmd.Flags().GetInt64("alloc")
	update, _ := cmd.Flags().GetBool("update")

	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")
	feeInt64 := int64(fee*1e4) * 1e4
	privateKeyStr, _ := cmd.Flags().GetString("key")

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
	parameter := fmt.Sprintf("set(%d, %d, %v)", pid, allocPoint, update)
	_, packData, err := evmAbi.Pack(parameter, masterChef.MasterChefABI, false)
	if nil != err {
		fmt.Println("UpdateAllocPoint", "Failed to do abi.Pack due to:", err.Error())
		return
	}
	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: parameter, Para: packData, ContractAddr: masterChefAddrStr}
	content, txHash, err := utils.CallContractAndSign(info, action, masterChefAddrStr)
	if nil != err {
		fmt.Println("Failed to create master chef due to cause:", err.Error())
		return
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	addPoolTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "update alloc point",
		Interval:      time.Second * 5,
	}
	file := fmt.Sprintf("./addPool_pid%d_%d.txt", pid, allocPoint)
	utils.WriteToFileInJson(file, addPoolTx)
}

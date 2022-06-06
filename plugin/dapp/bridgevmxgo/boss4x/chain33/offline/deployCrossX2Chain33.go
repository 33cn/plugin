package offline

import (
	"fmt"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/contracts/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline create -f 1 -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -n 'deploy crossx to chain33' -r '1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, [1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, 155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6, 13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv, 113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG], [25, 25, 25, 25]' --chainID 33
./boss4x chain33 offline send -f deployCrossX2Chain33.txt
*/

func CreateBridgevmxgoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create and sign all the offline txs to deploy bridgevmxgo contracts to chain33 evm (include valset,goAssetBridge,bridgeBank,oracle,bridgeRegistry,mulSign)",
		Run:   createBridgevmxgo,
	}
	addCreateCrossBridgeFlags(cmd)
	return cmd
}

func addCreateCrossBridgeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.Flags().StringP("valset", "r", "", "contruct parameter for valset, as: 'addr, [addr, addr, addr, addr], [25, 25, 25, 25]'")
	_ = cmd.MarkFlagRequired("valset")
}

func createBridgevmxgo(cmd *cobra.Command, args []string) {
	_ = args
	var txs []*utils.Chain33OfflineTx
	privateKeyStr, _ := cmd.Flags().GetString("key")
	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())

	i := 1
	fmt.Printf("%d: Going to create Valset\n", i)
	i += 1
	valsetTx, err := createValsetTxAndSign(cmd, from)
	if nil != err {
		fmt.Println("Failed to createValsetTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, valsetTx)

	fmt.Printf("%d: Going to create EthereumBridge\n", i)
	i += 1
	goAssetBridgeTx, err := createGoAssetBridgeAndSign(cmd, from, valsetTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createGoAssetBridgeAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, goAssetBridgeTx)

	fmt.Printf("%d: Going to create Oracle\n", i)
	i += 1
	oracleTx, err := createOracleTxAndSign(cmd, from, valsetTx.ContractAddr, goAssetBridgeTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createOracleTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, oracleTx)

	fmt.Printf("%d: Going to create BridgeBank\n", i)
	i += 1
	bridgeBankTx, err := createBridgeBankTxAndSign(cmd, from, valsetTx.ContractAddr, goAssetBridgeTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createBridgeBankTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, bridgeBankTx)

	fmt.Printf("%d: Going to set BridgeBank to EthBridge \n", i)
	i += 1
	setBridgeBank2GoAssetBridgeTx, err := setBridgeBank2GoAssetBridgeTxAndSign(cmd, goAssetBridgeTx.ContractAddr, bridgeBankTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to setBridgeBank2EthBridgeTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, setBridgeBank2GoAssetBridgeTx)

	fmt.Printf("%d: Going to set Oracle to EthBridge \n", i)
	i += 1
	setOracle2EthBridgeTx, err := setOracle2GoAssetBridgeTxAndSign(cmd, goAssetBridgeTx.ContractAddr, oracleTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to setOracle2GoAssetBridgeTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, setOracle2EthBridgeTx)

	fmt.Printf("%d: Going to create BridgeRegistry \n", i)
	i += 1
	createBridgeRegistryTx, err := createBridgeRegistryTxAndSign(cmd, from, goAssetBridgeTx.ContractAddr, valsetTx.ContractAddr, bridgeBankTx.ContractAddr, oracleTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createBridgeRegistryTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, createBridgeRegistryTx)

	fmt.Printf("%d: Write all the txs to file:   %s \n", i, crossXfileName)
	utils.WriteToFileInJson(crossXfileName, txs)
}

func createBridgeRegistryTxAndSign(cmd *cobra.Command, from common.Address, ethereumBridge, valset, bridgeBank, oracle string) (*utils.Chain33OfflineTx, error) {
	createPara := fmt.Sprintf("%s,%s,%s,%s", ethereumBridge, bridgeBank, oracle, valset)
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), generated.BridgeRegistryBin, generated.BridgeRegistryABI, createPara, "BridgeRegistry")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	bridgeRegistryTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy BridgeRegistry",
		Interval:      time.Second * 5,
	}
	return bridgeRegistryTx, nil
}

func setOracle2GoAssetBridgeTxAndSign(cmd *cobra.Command, ethbridge, oracle string) (*utils.Chain33OfflineTx, error) {
	parameter := fmt.Sprintf("setOracle(%s)", oracle)
	_, packData, err := evmAbi.Pack(parameter, generated.GoAssetBridgeABI, false)
	if nil != err {
		fmt.Println("setOracle2GoAssetBridge", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}
	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: "setOracle2GoAssetBridge", Para: packData, ContractAddr: ethbridge}
	content, txHash, err := utils.CallContractAndSign(getTxInfo(cmd), action, ethbridge)
	if nil != err {
		return nil, err
	}

	setOracle2EthBridgeTx := &utils.Chain33OfflineTx{
		ContractAddr:  ethbridge,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "setOracle2GoAssetBridge",
		Interval:      time.Second * 5,
	}
	return setOracle2EthBridgeTx, nil
}

func setBridgeBank2GoAssetBridgeTxAndSign(cmd *cobra.Command, ethbridge, bridgebank string) (*utils.Chain33OfflineTx, error) {
	parameter := fmt.Sprintf("setBridgeBank(%s)", bridgebank)
	_, packData, err := evmAbi.Pack(parameter, generated.GoAssetBridgeABI, false)
	if nil != err {
		fmt.Println("setBridgeBank2GoAssetBridge", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}
	action := &evmtypes.EVMContractAction{Amount: 0, GasLimit: 0, GasPrice: 0, Note: "setBridgeBank2GoAssetBridge", Para: packData, ContractAddr: ethbridge}
	content, txHash, err := utils.CallContractAndSign(getTxInfo(cmd), action, ethbridge)
	if nil != err {
		return nil, err
	}

	setBridgeBank2EthBridgeTx := &utils.Chain33OfflineTx{
		ContractAddr:  ethbridge,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "setBridgeBank2GoAssetBridge",
		Interval:      time.Second * 5,
	}
	return setBridgeBank2EthBridgeTx, nil
}

func createBridgeBankTxAndSign(cmd *cobra.Command, from common.Address, oracle, ethereumBridge string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s,%s", operator, oracle, ethereumBridge)
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), generated.BridgeBankBin, generated.BridgeBankABI, createPara, "bridgeBank")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	bridgeBankTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy bridgeBank",
		Interval:      time.Second * 5,
	}
	return bridgeBankTx, nil
}

func createOracleTxAndSign(cmd *cobra.Command, from common.Address, valset, ethereumBridge string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s,%s", operator, valset, ethereumBridge)
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), generated.OracleBin, generated.OracleABI, createPara, "oralce")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	oracleTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy oracle",
		Interval:      time.Second * 5,
	}
	return oracleTx, nil
}

func createValsetTxAndSign(cmd *cobra.Command, from common.Address) (*utils.Chain33OfflineTx, error) {
	contructParameter, _ := cmd.Flags().GetString("valset")
	createPara := contructParameter
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), generated.ValsetBin, generated.ValsetABI, createPara, "valset")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	valsetTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy valset",
		Interval:      time.Second * 5,
	}
	return valsetTx, nil
}

func createGoAssetBridgeAndSign(cmd *cobra.Command, from common.Address, valset string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s", operator, valset)
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), generated.GoAssetBridgeBin, generated.GoAssetBridgeABI, createPara, "goAssetBridge")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	goAssetBridgeTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy goAssetBridge",
		Interval:      time.Second * 5,
	}
	return goAssetBridgeTx, nil
}

func createMulSignAndSign(cmd *cobra.Command, from common.Address) (*utils.Chain33OfflineTx, error) {
	content, txHash, err := utils.CreateContractAndSign(getTxInfo(cmd), gnosis.GnosisSafeBin, gnosis.GnosisSafeABI, "", "mulSign2chain33")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	mulSign2chain33Tx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy mulSign2chain33",
		Interval:      time.Second * 5,
	}
	return mulSign2chain33Tx, nil
}

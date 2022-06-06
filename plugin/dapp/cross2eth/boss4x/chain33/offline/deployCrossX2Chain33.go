package offline

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4chain33/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/spf13/cobra"
)

/*
./boss4x chain33 offline create -f 1 -k 0x027ca96466c71c7e7c5d73b7e1f43cb889b3bd65ebd2413eefd31c6709c262ae -n 'deploy crossx to chain33' -r '1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, [1N6HstkyLFS8QCeVfdvYxx1xoryXoJtvvZ, 155ooMPBTF8QQsGAknkK7ei5D78rwDEFe6, 13zBdQwuyDh7cKN79oT2odkxYuDbgQiXFv, 113ZzVamKfAtGt9dq45fX1mNsEoDiN95HG], [25, 25, 25, 25]' --chainID 33
./boss4x chain33 offline send -f deployCrossX2Chain33.txt
*/

type DeployChain33ConfigInfo struct {
	ValidatorsAddr []string `toml:"validatorsAddr"`
	InitPowers     []int64  `toml:"initPowers"`
	Symbol         string   `toml:"symbol"`
	MultisignAddrs []string `toml:"multisignAddrs"`
}

func CreateContractsWithFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_file",
		Short: "create and sign all the offline cross to ethereum contracts(inclue valset,ethereumBridge,bridgeBank,oracle,bridgeRegistry,mulSign)",
		Run:   createContractsWithFile,
	}
	addCreateContractsWithFileFlags(cmd)
	return cmd
}

func addCreateContractsWithFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the deployer private key")
	_ = cmd.MarkFlagRequired("key")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.Flags().StringP("conf", "c", "", "config file")
	_ = cmd.MarkFlagRequired("conf")
}

func createContractsWithFile(cmd *cobra.Command, _ []string) {
	cfgpath, _ := cmd.Flags().GetString("conf")
	var deployCfg DeployChain33ConfigInfo
	InitCfg(cfgpath, &deployCfg)
	if len(deployCfg.InitPowers) != len(deployCfg.ValidatorsAddr) {
		panic("not same number for validator address and power")
	}

	if len(deployCfg.ValidatorsAddr) < 3 {
		panic("the number of validator must be not less than 3")
	}
	validators := strings.Join(deployCfg.ValidatorsAddr, ",")
	var stringPowers []string
	for _, v := range deployCfg.InitPowers {
		stringPowers = append(stringPowers, strconv.FormatInt(v, 10))
	}
	initPowers := strings.Join(stringPowers, ",")

	privateKeyStr, _ := cmd.Flags().GetString("key")
	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())

	valsetParameter := from.String() + ",[" + validators + "],[" + initPowers + "]"
	multisignAddrs := strings.Join(deployCfg.MultisignAddrs, ",")
	txCreateInfo := getTxInfo(cmd)

	createChain33DeployTxs(txCreateInfo, valsetParameter, multisignAddrs, from)

}

func CreateCrossBridgeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create and sign all the offline cross to ethereum contracts(inclue valset,ethereumBridge,bridgeBank,oracle,bridgeRegistry,mulSign)",
		Run:   createCrossBridge,
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
	cmd.Flags().StringP("multisignAddrs", "m", "", "multisign address, as: 'addr, addr, addr, addr'")
	_ = cmd.MarkFlagRequired("multisignAddrs")
}

func createCrossBridge(cmd *cobra.Command, args []string) {
	_ = args
	privateKeyStr, _ := cmd.Flags().GetString("key")
	multisignAddrs, _ := cmd.Flags().GetString("multisignAddrs")
	valsetParameter, _ := cmd.Flags().GetString("valset")
	var driver secp256k1.Driver
	privateKeySli := common.FromHex(privateKeyStr)
	privateKey, err := driver.PrivKeyFromBytes(privateKeySli)
	if nil != err {
		fmt.Println("Failed to do PrivKeyFromBytes")
		return
	}
	from := common.PubKey2Address(privateKey.PubKey().Bytes())

	txCreateInfo := getTxInfo(cmd)

	createChain33DeployTxs(txCreateInfo, valsetParameter, multisignAddrs, from)
}

func createChain33DeployTxs(txCreateInfo *utils.TxCreateInfo, valsetParameter, multisignAddrs string, from common.Address) {
	var txs []*utils.Chain33OfflineTx
	i := 1
	fmt.Printf("%d: Going to create Valset\n", i)
	i += 1
	// 0
	valsetTx, err := createValsetTxAndSign(txCreateInfo, valsetParameter, from)
	if nil != err {
		fmt.Println("Failed to createValsetTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, valsetTx)

	fmt.Printf("%d: Going to create EthereumBridge\n", i)
	i += 1
	// 1
	ethereumBridgeTx, err := createEthereumBridgeAndSign(txCreateInfo, from, valsetTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createEthereumBridgeAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, ethereumBridgeTx)

	fmt.Printf("%d: Going to create Oracle\n", i)
	i += 1
	// 2
	oracleTx, err := createOracleTxAndSign(txCreateInfo, from, valsetTx.ContractAddr, ethereumBridgeTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createOracleTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, oracleTx)

	fmt.Printf("%d: Going to create BridgeBank\n", i)
	i += 1
	// 3
	bridgeBankTx, err := createBridgeBankTxAndSign(txCreateInfo, from, valsetTx.ContractAddr, ethereumBridgeTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createBridgeBankTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, bridgeBankTx)

	fmt.Printf("%d: Going to set BridgeBank to EthBridge \n", i)
	i += 1
	// 4
	setBridgeBank2EthBridgeTx, err := setBridgeBank2EthBridgeTxAndSign(txCreateInfo, ethereumBridgeTx.ContractAddr, bridgeBankTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to setBridgeBank2EthBridgeTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, setBridgeBank2EthBridgeTx)

	fmt.Printf("%d: Going to set Oracle to EthBridge \n", i)
	i += 1
	// 5
	setOracle2EthBridgeTx, err := setOracle2EthBridgeTxAndSign(txCreateInfo, ethereumBridgeTx.ContractAddr, oracleTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to setOracle2EthBridgeTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, setOracle2EthBridgeTx)

	fmt.Printf("%d: Going to create BridgeRegistry \n", i)
	i += 1
	// 6
	createBridgeRegistryTx, err := createBridgeRegistryTxAndSign(txCreateInfo, from, ethereumBridgeTx.ContractAddr, valsetTx.ContractAddr, bridgeBankTx.ContractAddr, oracleTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to createBridgeRegistryTxAndSign due to cause:", err.Error())
		return
	}
	txs = append(txs, createBridgeRegistryTx)

	fmt.Printf("%d: Going to create MulSign2chain33 \n", i)
	i += 1
	// 7
	createMulSign2chain33Tx, err := createMulSignAndSign(txCreateInfo, from)
	if nil != err {
		fmt.Println("Failed to createMulSign2chain33Tx due to cause:", err.Error())
		return
	}
	txs = append(txs, createMulSign2chain33Tx)

	fmt.Printf("%d: Going to save multisign contract config offline account \n", i)
	i += 1
	// 8
	configMultisignOfflineSaveAccountTx, err := configMultisignOfflineSaveAccount(txCreateInfo, createMulSign2chain33Tx.ContractAddr, bridgeBankTx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to configMultisignOfflineSaveAccountTx due to cause:", err.Error())
		return
	}
	txs = append(txs, configMultisignOfflineSaveAccountTx)

	fmt.Printf("%d: Going to Setup multisignAddrs to contract \n", i)
	i += 1
	// 9
	multisignSetupTx, err := multisignSetup(txCreateInfo, multisignAddrs, createMulSign2chain33Tx.ContractAddr)
	if nil != err {
		fmt.Println("Failed to configMultisignOfflineSaveAccountTx due to cause:", err.Error())
		return
	}
	txs = append(txs, multisignSetupTx)

	fmt.Printf("%d: Write all the txs to file:   %s \n", i, crossXfileName)
	utils.WriteToFileInJson(crossXfileName, txs)
}

func createBridgeRegistryTxAndSign(txCreateInfo *utils.TxCreateInfo, from common.Address, ethereumBridge, valset, bridgeBank, oracle string) (*utils.Chain33OfflineTx, error) {
	createPara := fmt.Sprintf("%s,%s,%s,%s", ethereumBridge, bridgeBank, oracle, valset)
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.BridgeRegistryBin, generated.BridgeRegistryABI, createPara, "BridgeRegistry")
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

func setOracle2EthBridgeTxAndSign(txCreateInfo *utils.TxCreateInfo, ethbridge, oracle string) (*utils.Chain33OfflineTx, error) {
	parameter := fmt.Sprintf("setOracle(%s)", oracle)
	_, packData, err := evmAbi.Pack(parameter, generated.EthereumBridgeABI, false)
	if nil != err {
		fmt.Println("setOracle2EthBridge", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}

	return createOfflineTx(txCreateInfo, packData, ethbridge, "setOracle2EthBridge", time.Second*5)
}

func setBridgeBank2EthBridgeTxAndSign(txCreateInfo *utils.TxCreateInfo, ethbridge, bridgebank string) (*utils.Chain33OfflineTx, error) {
	parameter := fmt.Sprintf("setBridgeBank(%s)", bridgebank)
	_, packData, err := evmAbi.Pack(parameter, generated.EthereumBridgeABI, false)
	if nil != err {
		fmt.Println("setBridgeBank2EthBridge", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}
	return createOfflineTx(txCreateInfo, packData, ethbridge, "setBridgeBank2EthBridge", time.Second*5)
}

func createBridgeBankTxAndSign(txCreateInfo *utils.TxCreateInfo, from common.Address, oracle, ethereumBridge string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s,%s", operator, oracle, ethereumBridge)
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.BridgeBankBin, generated.BridgeBankABI, createPara, "bridgeBank")
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

func createOracleTxAndSign(txCreateInfo *utils.TxCreateInfo, from common.Address, valset, ethereumBridge string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s,%s", operator, valset, ethereumBridge)
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.OracleBin, generated.OracleABI, createPara, "oralce")
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

func createValsetTxAndSign(txCreateInfo *utils.TxCreateInfo, valsetParameter string, from common.Address) (*utils.Chain33OfflineTx, error) {
	createPara := valsetParameter
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.ValsetBin, generated.ValsetABI, createPara, "valset")
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

func createEthereumBridgeAndSign(txCreateInfo *utils.TxCreateInfo, from common.Address, valset string) (*utils.Chain33OfflineTx, error) {
	operator := from.String()
	createPara := fmt.Sprintf("%s,%s", operator, valset)
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.EthereumBridgeBin, generated.EthereumBridgeABI, createPara, "EthereumBridge")
	if nil != err {
		return nil, err
	}

	newContractAddr := common.NewContractAddress(from, txHash).String()
	ethereumBridgeTx := &utils.Chain33OfflineTx{
		ContractAddr:  newContractAddr,
		TxHash:        common.Bytes2Hex(txHash),
		SignedRawTx:   content,
		OperationName: "deploy ethereumBridge",
		Interval:      time.Second * 5,
	}
	return ethereumBridgeTx, nil
}

func createMulSignAndSign(txCreateInfo *utils.TxCreateInfo, from common.Address) (*utils.Chain33OfflineTx, error) {
	content, txHash, err := utils.CreateContractAndSign(txCreateInfo, generated.GnosisSafeBin, generated.GnosisSafeABI, "", "mulSign2chain33")
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

func configMultisignOfflineSaveAccount(txCreateInfo *utils.TxCreateInfo, multisignContract, bridgebank string) (*utils.Chain33OfflineTx, error) {
	parameter := fmt.Sprintf("configOfflineSaveAccount(%s)", multisignContract)
	_, packData, err := evmAbi.Pack(parameter, generated.BridgeBankABI, false)
	if nil != err {
		fmt.Println("configOfflineSaveAccount", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}
	return createOfflineTx(txCreateInfo, packData, bridgebank, "configOfflineSaveAccount", time.Second*5)
}

func multisignSetup(txCreateInfo *utils.TxCreateInfo, multisignAddrs string, multisignContract string) (*utils.Chain33OfflineTx, error) {
	owners := strings.Split(multisignAddrs, ",")

	BTYAddrChain33 := ebTypes.BTYAddrChain33
	parameter := "setup(["
	parameter += fmt.Sprintf("%s", owners[0])
	for _, owner := range owners[1:] {
		parameter += fmt.Sprintf(",%s", owner)
	}
	parameter += "], "
	parameter += fmt.Sprintf("%d, %s, 0102, %s, %s, 0, %s)", len(owners), BTYAddrChain33, BTYAddrChain33, BTYAddrChain33, BTYAddrChain33)
	_, packData, err := evmAbi.Pack(parameter, generated.GnosisSafeABI, false)
	if nil != err {
		fmt.Println("multisign_setup", "Failed to do abi.Pack due to:", err.Error())
		return nil, err
	}

	return createOfflineTx(txCreateInfo, packData, multisignContract, "multisign_setup", time.Second*5)
}

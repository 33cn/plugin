package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

/*
./boss4x ethereum offline create -p 25,25,25,25 -o 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -v 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a,0x0df9a824699bc5878232c9e612fe1a5346a5a368,0xcb074cb21cdddf3ce9c3c0a7ac4497d633c9d9f1,0xd9dab021e74ecf475788ed7b61356056b2095830
./boss4x ethereum offline sign -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt
*/

func CreateEthBridgeBankRelatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "createEthBridgeBankRelated", //first step
		Short: "create and sign all the offline bridgeBank to ethereum contracts(inclue bridgeBank and bridgeRegistry)",
		Run:   CreateEthBridgeBankRelated, //对要部署的factory合约进行签名
	}
	addCreateEthBridgeBankRelatedFlags(cmd)
	return cmd
}

func addCreateEthBridgeBankRelatedFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "the deployer address")
	_ = cmd.MarkFlagRequired("owner")

	cmd.Flags().StringP("valSetAddr", "v", "", "valSet address")
	_ = cmd.MarkFlagRequired("valSetAddr")
}

func CreateEthBridgeBankRelated(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	owner, _ := cmd.Flags().GetString("owner")
	valSetAddrStr, _ := cmd.Flags().GetString("valSetAddr")

	deployerAddr := common.HexToAddress(owner)
	valSetAddr := common.HexToAddress(valSetAddrStr)

	client, err := ethclient.Dial(url)
	if err != nil {
		panic(err.Error())
	}

	ctx := context.Background()
	startNonce, err := client.PendingNonceAt(ctx, deployerAddr)
	if nil != err {
		panic(err.Error())
	}

	var infos []*DeployInfo

	//step1 chain33bridge
	packData, err := deploychain33BridgePackData(deployerAddr, valSetAddr)
	if err != nil {
		panic(err.Error())
	}
	chain33BridgeAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: chain33BridgeAddr, Name: "chain33Bridge", Nonce: startNonce, To: nil})
	startNonce = startNonce + 1

	//step2 oracle
	packData, err = deployOraclePackData(deployerAddr, valSetAddr, chain33BridgeAddr)
	if err != nil {
		panic(err.Error())
	}
	oracleAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: oracleAddr, Name: "oracle", Nonce: startNonce, To: nil})
	startNonce = startNonce + 1

	//step3 bridgebank
	packData, err = deployBridgeBankPackData(deployerAddr, chain33BridgeAddr, oracleAddr)
	if err != nil {
		panic(err.Error())
	}
	bridgeBankAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeBankAddr, Name: "bridgebank", Nonce: startNonce, To: nil})
	startNonce = startNonce + 1

	//step4
	packData, err = callSetBridgeBank(bridgeBankAddr)
	if err != nil {
		panic(err.Error())
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setbridgebank", Nonce: startNonce, To: &chain33BridgeAddr})
	startNonce = startNonce + 1

	//step5
	packData, err = callSetOracal(oracleAddr)
	if err != nil {
		panic(err.Error())
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setoracle", Nonce: startNonce, To: &chain33BridgeAddr})
	startNonce = startNonce + 1

	//step6 bridgeRegistry
	packData, err = deployBridgeRegistry(chain33BridgeAddr, bridgeBankAddr, oracleAddr, valSetAddr)
	if err != nil {
		panic(err.Error())
	}
	bridgeRegAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeRegAddr, Name: "bridgeRegistry", Nonce: startNonce, To: nil})

	err = NewTxWrite(infos, deployerAddr, url, "deployBridgeBank4Ethtxs.txt")
	if err != nil {
		panic(err.Error())
	}
}

// CreateCmd 查询deploy 私钥的nonce信息，并输出到文件中
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create", //first step
		Short: "create all the offline cross to ethereum contracts(inclue valset,ethereumBridge,bridgeBank,oracle,bridgeRegistry,mulSign), set symbol",
		Run:   createTx, //对要部署的factory合约进行签名
	}
	addCreateFlags(cmd)
	return cmd
}

func addCreateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("validatorsAddrs", "v", "", "validatorsAddrs, as: 'addr,addr,addr,addr'")
	_ = cmd.MarkFlagRequired("validatorsAddrs")
	cmd.Flags().StringP("initPowers", "p", "", "initPowers, as: '25,25,25,25'")
	_ = cmd.MarkFlagRequired("initPowers")
	cmd.Flags().StringP("owner", "o", "", "the deployer address")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("symbol", "s", "", "symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("multisignAddrs", "m", "", "multisignAddrs, as: 'addr,addr,addr,addr'")
	_ = cmd.MarkFlagRequired("multisignAddrs")
}

func createTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	validatorsAddrs, _ := cmd.Flags().GetString("validatorsAddrs")
	multisignAddrs, _ := cmd.Flags().GetString("multisignAddrs")
	multisignAddrsArray := strings.Split(multisignAddrs, ",")
	initpowers, _ := cmd.Flags().GetString("initPowers")
	owner, _ := cmd.Flags().GetString("owner")
	deployerAddr := common.HexToAddress(owner)
	validatorsAddrsArray := strings.Split(validatorsAddrs, ",")
	initPowersArray := strings.Split(initpowers, ",")
	symbol, _ := cmd.Flags().GetString("symbol")

	if len(validatorsAddrsArray) != len(initPowersArray) {
		fmt.Println("input validatorsAddrs initPowers error!")
		return
	}

	if len(validatorsAddrsArray) < 3 {
		fmt.Println("the number of validator must be not less than 3")
		return
	}

	var validators, multisigns []common.Address
	var initPowers []*big.Int
	for _, v := range validatorsAddrsArray {
		validators = append(validators, common.HexToAddress(v))
	}

	for _, v := range initPowersArray {
		vint64, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		initPowers = append(initPowers, big.NewInt(vint64))
	}

	for _, v := range multisignAddrsArray {
		multisigns = append(multisigns, common.HexToAddress(v))
	}

	err := createDeployTxs(url, deployerAddr, validators, multisigns, initPowers, symbol)
	if err != nil {
		panic(err)
	}
}

func createDeployTxs(url string, deployerAddr common.Address, validators, multisigns []common.Address, initPowers []*big.Int, symbol string) error {
	client, err := ethclient.Dial(url)
	if err != nil {
		return err
	}

	ctx := context.Background()
	startNonce, err := client.PendingNonceAt(ctx, deployerAddr)
	if nil != err {
		return err
	}

	var infos []*DeployInfo
	//step1 valSet
	packData, err := deployValSetPackData(validators, initPowers, deployerAddr)
	if err != nil {
		return err
	}
	valSetAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: valSetAddr, Name: "valSet", Nonce: startNonce, To: nil})
	startNonce += 1

	//step2 chain33bridge
	packData, err = deploychain33BridgePackData(deployerAddr, valSetAddr)
	if err != nil {
		return err
	}
	chain33BridgeAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: chain33BridgeAddr, Name: "chain33Bridge", Nonce: startNonce, To: nil})
	startNonce += 1

	//step3 oracle
	packData, err = deployOraclePackData(deployerAddr, valSetAddr, chain33BridgeAddr)
	if err != nil {
		return err
	}
	oracleAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: oracleAddr, Name: "oracle", Nonce: startNonce, To: nil})
	startNonce += 1

	//step4 bridgebank
	packData, err = deployBridgeBankPackData(deployerAddr, chain33BridgeAddr, oracleAddr)
	if err != nil {
		return err
	}
	bridgeBankAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeBankAddr, Name: "bridgebank", Nonce: startNonce, To: nil})
	startNonce += 1

	//step5
	packData, err = callSetBridgeBank(bridgeBankAddr)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setbridgebank", Nonce: startNonce, To: &chain33BridgeAddr})
	startNonce += 1

	//step6
	packData, err = callSetOracal(oracleAddr)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setoracle", Nonce: startNonce, To: &chain33BridgeAddr})
	startNonce += 1

	//step7 set symbol
	packData, err = setSymbol(symbol)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setsymbol", Nonce: startNonce, To: &bridgeBankAddr})
	startNonce += 1

	//step8 bridgeRegistry
	packData, err = deployBridgeRegistry(chain33BridgeAddr, bridgeBankAddr, oracleAddr, valSetAddr)
	if err != nil {
		return err
	}
	bridgeRegAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeRegAddr, Name: "bridgeRegistry", Nonce: startNonce, To: nil})
	startNonce += 1

	//step9 mulSign
	packData = common.FromHex(gnosis.GnosisSafeBin)
	mulSignAddr := crypto.CreateAddress(deployerAddr, startNonce)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: mulSignAddr, Name: "mulSignAddr", Nonce: startNonce, To: nil})
	startNonce += 1

	//step10 multisign configOfflineSaveAccount
	packData, err = offlineSaveAccount(mulSignAddr)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "configOfflineSaveAccount", Nonce: startNonce, To: &bridgeBankAddr})
	startNonce += 1

	//step11 multisignSetup
	packData, err = multisignSetup(multisigns)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "multisignSetup", Nonce: startNonce, To: &mulSignAddr})

	return NewTxWrite(infos, deployerAddr, url, "deploytxs.txt")
}

func NewTxWrite(infos []*DeployInfo, deployerAddr common.Address, url, fileName string) error {
	ctx := context.Background()
	client, err := ethclient.Dial(url)
	if err != nil {
		return err
	}
	price, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return err
	}
	//预估gas,批量构造交易
	for i, info := range infos {
		var msg ethereum.CallMsg
		msg.From = deployerAddr
		msg.To = info.To
		msg.Value = big.NewInt(0)
		msg.Data = info.PackData
		//估算gas
		gasLimit, err := client.EstimateGas(ctx, msg)
		if err != nil {
			return err
		}
		if gasLimit < 100*10000 {
			gasLimit = 100 * 10000
		}
		ntx := types.NewTx(&types.LegacyTx{
			//ntx := types.NewTx(&types.AccessListTx{
			//	ChainID:  big.NewInt(chainId),
			Nonce:    info.Nonce,
			Gas:      gasLimit,
			GasPrice: price,
			Data:     info.PackData,
			To:       info.To,
		})

		txBytes, err := ntx.MarshalBinary()
		if err != nil {
			return err
		}
		infos[i].RawTx = common.Bytes2Hex(txBytes)
		infos[i].Gas = gasLimit
	}

	jbytes, err := json.MarshalIndent(&infos, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(jbytes))
	writeToFile(fileName, &infos)

	return nil
}

//deploy contract step 1
func deployValSetPackData(validators []common.Address, powers []*big.Int, deployerAddr common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.ValsetABI))
	if err != nil {
		return nil, err
	}
	bin := common.FromHex(generated.ValsetBin)
	packdata, err := parsed.Pack("", deployerAddr, validators, powers)
	if err != nil {
		return nil, err
	}
	return append(bin, packdata...), nil
}

//deploy contract step 2
func deploychain33BridgePackData(deployerAddr, valSetAddr common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.Chain33BridgeABI))
	if err != nil {
		return nil, err
	}
	bin := common.FromHex(generated.Chain33BridgeBin)
	input, err := parsed.Pack("", deployerAddr, valSetAddr)
	if err != nil {
		return nil, err
	}
	return append(bin, input...), nil
}

//deploy contract step 3
func deployOraclePackData(deployerAddr, valSetAddr, bridgeAddr common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.OracleABI))
	if err != nil {
		return nil, err
	}
	bin := common.FromHex(generated.OracleBin)
	packData, err := parsed.Pack("", deployerAddr, valSetAddr, bridgeAddr)
	if err != nil {
		return nil, err
	}
	return append(bin, packData...), nil
}

//deploy contract step 4
func deployBridgeBankPackData(deployerAddr, bridgeAddr, oracalAddr common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		return nil, err
	}
	bin := common.FromHex(generated.BridgeBankBin)
	packData, err := parsed.Pack("", deployerAddr, oracalAddr, bridgeAddr)
	if err != nil {
		return nil, err
	}
	return append(bin, packData...), nil
}

////deploy contract step 5
func callSetBridgeBank(bridgeBankAddr common.Address) ([]byte, error) {
	method := "setBridgeBank"
	parsed, err := abi.JSON(strings.NewReader(generated.Chain33BridgeABI))
	if err != nil {
		return nil, err
	}
	packData, err := parsed.Pack(method, bridgeBankAddr)
	if err != nil {
		return nil, err
	}
	return packData, nil
}

//deploy contract step 6
func callSetOracal(oracalAddr common.Address) ([]byte, error) {
	method := "setOracle"
	parsed, err := abi.JSON(strings.NewReader(generated.Chain33BridgeABI))
	if err != nil {
		return nil, err
	}
	packData, err := parsed.Pack(method, oracalAddr)
	if err != nil {
		return nil, err
	}
	return packData, nil
}

//deploy contract step 7
func deployBridgeRegistry(chain33BridgeAddr, bridgeBankAddr, oracleAddr, valSetAddr common.Address) ([]byte, error) {
	parsed, err := abi.JSON(strings.NewReader(generated.BridgeRegistryABI))
	if err != nil {
		return nil, err
	}
	bin := common.FromHex(generated.BridgeRegistryBin)
	packData, err := parsed.Pack("", chain33BridgeAddr, bridgeBankAddr, oracleAddr, valSetAddr)
	if err != nil {
		return nil, err
	}
	return append(bin, packData...), nil
}

//step 8
func setSymbol(symbol string) ([]byte, error) {
	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		return nil, err
	}
	abiData, err := bridgeAbi.Pack("configplatformTokenSymbol", symbol)
	if err != nil {
		return nil, err
	}
	return abiData, nil
}

//step 10
func offlineSaveAccount(multisignContract common.Address) ([]byte, error) {
	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		return nil, err
	}

	abiData, err := bridgeAbi.Pack("configOfflineSaveAccount", multisignContract)
	if err != nil {
		return nil, err
	}
	return abiData, nil
}

//step 11
func multisignSetup(multisigns []common.Address) ([]byte, error) {
	AddressZero := common.HexToAddress(ebTypes.EthNilAddr)
	gnoAbi, err := abi.JSON(strings.NewReader(gnosis.GnosisSafeABI))
	if err != nil {
		return nil, err
	}

	abiData, err := gnoAbi.Pack("setup", multisigns, big.NewInt(int64(len(multisigns)/2+1)), AddressZero, []byte{'0', 'x'},
		AddressZero, AddressZero, big.NewInt(int64(0)), AddressZero)
	if err != nil {
		return nil, err
	}
	return abiData, nil
}

func CreateWithFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create_file", //first step
		Short: "create deploy tx with file",
		Run:   createWithFileTx, //对要部署的factory合约进行签名
	}
	addCreateWithFileFlags(cmd)
	return cmd
}

func addCreateWithFileFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("conf", "f", "", "config file")
	_ = cmd.MarkFlagRequired("conf")
}

func createWithFileTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	cfgpath, _ := cmd.Flags().GetString("conf")
	var deployCfg DeployConfigInfo
	InitCfg(cfgpath, &deployCfg)
	if len(deployCfg.InitPowers) != len(deployCfg.ValidatorsAddr) {
		panic("not same number for validator address and power")
	}

	if len(deployCfg.ValidatorsAddr) < 3 {
		panic("the number of validator must be not less than 3")
	}

	var validators, multisigns []common.Address
	var initPowers []*big.Int
	for i, addr := range deployCfg.ValidatorsAddr {
		validators = append(validators, common.HexToAddress(addr))
		initPowers = append(initPowers, big.NewInt(deployCfg.InitPowers[i]))
	}

	for _, addr := range deployCfg.MultisignAddrs {
		multisigns = append(multisigns, common.HexToAddress(addr))
	}

	err := createDeployTxs(url, common.HexToAddress(deployCfg.OperatorAddr), validators, multisigns, initPowers, deployCfg.Symbol)
	if err != nil {
		fmt.Println("createDeployTxs Err:", err)
		return
	}
}

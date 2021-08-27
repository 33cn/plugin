package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

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

// CreateCmd 查询deploy 私钥的nonce信息，并输出到文件中
func CreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create", //first step
		Short: "create and sign all the offline cross to ethereum contracts(inclue valset,ethereumBridge,bridgeBank,oracle,bridgeRegistry,mulSign)",
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
}

func createTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	validatorsAddrs, _ := cmd.Flags().GetString("validatorsAddrs")
	initpowers, _ := cmd.Flags().GetString("initPowers")
	owner, _ := cmd.Flags().GetString("owner")
	deployerAddr := common.HexToAddress(owner)
	validatorsAddrsArray := strings.Split(validatorsAddrs, ",")
	initPowersArray := strings.Split(initpowers, ",")

	if len(validatorsAddrsArray) != len(initPowersArray) {
		fmt.Println("input validatorsAddrs initPowers error!")
		return
	}

	if len(validatorsAddrsArray) < 3 {
		fmt.Println("the number of validator must be not less than 3")
		return
	}

	var validators []common.Address
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

	err := createDeployTxs(url, deployerAddr, validators, initPowers)
	if err != nil {
		panic(err)
	}
}

func createDeployTxs(url string, deployerAddr common.Address, validators []common.Address, initPowers []*big.Int) error {
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

	//step2 chain33bridge
	packData, err = deploychain33BridgePackData(deployerAddr, valSetAddr)
	if err != nil {
		return err
	}
	chain33BridgeAddr := crypto.CreateAddress(deployerAddr, startNonce+1)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: chain33BridgeAddr, Name: "chain33Bridge", Nonce: startNonce + 1, To: nil})

	//step3 oracle
	packData, err = deployOraclePackData(deployerAddr, valSetAddr, chain33BridgeAddr)
	if err != nil {
		return err
	}
	oracleAddr := crypto.CreateAddress(deployerAddr, startNonce+2)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: oracleAddr, Name: "oracle", Nonce: startNonce + 2, To: nil})

	//step4 bridgebank
	packData, err = deployBridgeBankPackData(deployerAddr, chain33BridgeAddr, oracleAddr)
	if err != nil {
		return err
	}
	bridgeBankAddr := crypto.CreateAddress(deployerAddr, startNonce+3)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeBankAddr, Name: "bridgebank", Nonce: startNonce + 3, To: nil})

	//step5
	packData, err = callSetBridgeBank(bridgeBankAddr)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setbridgebank", Nonce: startNonce + 4, To: &chain33BridgeAddr})

	//step6
	packData, err = callSetOracal(oracleAddr)
	if err != nil {
		return err
	}
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: common.Address{}, Name: "setoracle", Nonce: startNonce + 5, To: &chain33BridgeAddr})

	//step7 bridgeRegistry
	packData, err = deployBridgeRegistry(chain33BridgeAddr, bridgeBankAddr, oracleAddr, valSetAddr)
	if err != nil {
		return err
	}
	bridgeRegAddr := crypto.CreateAddress(deployerAddr, startNonce+6)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: bridgeRegAddr, Name: "bridgeRegistry", Nonce: startNonce + 6, To: nil})

	//step8 bridgeRegistry
	packData = common.FromHex(gnosis.GnosisSafeBin)
	mulSignAddr := crypto.CreateAddress(deployerAddr, startNonce+7)
	infos = append(infos, &DeployInfo{PackData: packData, ContractorAddr: mulSignAddr, Name: "mulSignAddr", Nonce: startNonce + 7, To: nil})

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
	cmd.Flags().StringP("conf", "c", "", "config file")
	_ = cmd.MarkFlagRequired("conf")
}

func createWithFileTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	cfgpath, _ := cmd.Flags().GetString("conf")
	var deployCfg DeployConfigInfo
	InitCfg(cfgpath, &deployCfg)
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(deployCfg.DeployerPrivateKey))
	if err != nil {
		fmt.Println("crypto.ToECDSA Err:", err)
		return
	}

	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
	if len(deployCfg.InitPowers) != len(deployCfg.ValidatorsAddr) {
		panic("not same number for validator address and power")
	}

	if len(deployCfg.ValidatorsAddr) < 3 {
		panic("the number of validator must be not less than 3")
	}

	var validators []common.Address
	var initPowers []*big.Int
	for i, addr := range deployCfg.ValidatorsAddr {
		validators = append(validators, common.HexToAddress(addr))
		initPowers = append(initPowers, big.NewInt(deployCfg.InitPowers[i]))
	}

	err = createDeployTxs(url, deployerAddr, validators, initPowers)
	if err != nil {
		fmt.Println("createDeployTxs Err:", err)
		return
	}
}

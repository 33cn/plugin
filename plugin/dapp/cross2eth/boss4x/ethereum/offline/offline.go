package offline

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"

	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
)

type DeployContractRet struct {
	ContractAddr string
	TxHash       string
	ContractName string
}

func DeployOfflineContractsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "offline",
		Short: "deploy the corresponding Ethereum contracts",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CreateCmd(), //构造交易
		CreateWithFileCmd(),
		DeployERC20Cmd(),
		DeployBEP20Cmd(),
		DeployTetherUSDTCmd(),
		//CreateCfgAccountTxCmd(), // set_offline_addr 设置离线多签地址
		//SetupCmd(),
		ConfigLockedTokenOfflineSaveCmd(),
		CreateAddToken2LockListTxCmd(),
		CreateBridgeTokenTxCmd(),
		PrepareCreateMultisignTransferTxCmd(),   // 预备创建一个多签转帐交易 在线
		PreliminarySignMultisignTransferTxCmd(), // 多签转帐交易 多签多个地址签名 离线
		CreateMultisignTransferTxCmd(),          // 创建多签转帐交易
		SignCmd(),                               // 签名交易 sign deploy contract tx
		SendTxsCmd(),                            // 发送交易 send all kinds of tx
		//ConfigplatformTokenSymbolCmd(),
		CreateEthBridgeBankRelatedCmd(), //构造交易
	)

	return cmd
}

type DeployInfo struct {
	Name           string
	PackData       []byte
	ContractorAddr common.Address
	Nonce          uint64
	To             *common.Address
	RawTx          string
	TxHash         string
	Gas            uint64
}

type DeployConfigInfo struct {
	OperatorAddr   string   `toml:"operatorAddr"`
	ValidatorsAddr []string `toml:"validatorsAddr"`
	InitPowers     []int64  `toml:"initPowers"`
	Symbol         string   `toml:"symbol"`
	MultisignAddrs []string `toml:"multisignAddrs"`
}

func CreateTxInfoAndWrite(abiData []byte, deployAddr, contract, name, url string, chainEthId int64) {
	client, err := ethclient.Dial(url)
	if err != nil {
		fmt.Println("Dial Err:", err)
		return
	}
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		fmt.Println("SuggestGasPrice Err:", err)
		return
	}

	nonce, err := client.NonceAt(context.Background(), common.HexToAddress(deployAddr), nil)
	if err != nil {
		fmt.Println("NonceAt Err:", err)
		return
	}

	contracAddr := common.HexToAddress(contract)
	var msg ethereum.CallMsg
	msg.Data = abiData
	msg.From = common.HexToAddress(deployAddr)
	msg.To = &contracAddr
	msg.Value = big.NewInt(0)
	//估算gas
	var gasLimit uint64
	// 模拟节点测试
	if chainEthId == 1337 {
		gasLimit = uint64(500 * 10000)
	} else {
		gasLimit, err = client.EstimateGas(context.Background(), msg)
		if err != nil {
			fmt.Println("EstimateGas Err:", err)
			return
		}
		gasLimit = uint64(1.2 * float64(gasLimit))
		if gasLimit < 100*10000 {
			gasLimit = 100 * 10000
		}
	}

	ntx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		Gas:      gasLimit,
		GasPrice: gasPrice,
		Data:     abiData,
		To:       &contracAddr,
	})

	txBytes, err := ntx.MarshalBinary()
	if err != nil {
		fmt.Println("MarshalBinary Err:", err)
		return
	}
	var info DeployInfo
	info.RawTx = common.Bytes2Hex(txBytes)
	info.ContractorAddr = crypto.CreateAddress(contracAddr, nonce)
	info.PackData = abiData
	info.Nonce = nonce
	info.Gas = gasLimit
	info.Name = name

	var infos []*DeployInfo
	infos = append(infos, &info)
	writeToFile(info.Name+".txt", infos)
}

func paraseFile(file string, result interface{}) error {
	_, err := os.Stat(file)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	return json.Unmarshal(b, result)
}

func writeToFile(fileName string, content interface{}) {
	jbytes, err := json.MarshalIndent(content, "", "\t")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(fileName, jbytes, 0666)
	if err != nil {
		fmt.Println("Failed to write to file:", fileName)
	}
	fmt.Println("tx is written to file: ", fileName)
}

func InitCfg(filepath string, cfg interface{}) {
	if _, err := tml.DecodeFile(filepath, cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return
}

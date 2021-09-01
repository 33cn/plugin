package ethereum

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/boss4x/ethereum/offline"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/relayer/ethereum/ethtxs"
	tml "github.com/BurntSushi/toml"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

type DepolyInfo struct {
	OperatorAddr       string   `toml:"operatorAddr"`
	DeployerPrivateKey string   `toml:"deployerPrivateKey"`
	ValidatorsAddr     []string `toml:"validatorsAddr"`
	InitPowers         []int64  `toml:"initPowers"`
}

func EthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ethereum",
		Short: "deploy to eth",
	}
	cmd.AddCommand(
		//DeployContrctsCmd(),
		offline.DeployOfflineContractsCmd(),
	)
	return cmd

}

//DeployContrctsCmd ...
func DeployContrctsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy the corresponding Ethereum contracts",
		Run:   DeployContrcts,
	}
	addDeployFlags(cmd)
	return cmd
}

func addDeployFlags(cmd *cobra.Command) {
	//私钥的优先权大于配置文件的
	cmd.Flags().StringP("privkey", "p", "", "deployer privatekey")
	_ = cmd.MarkFlagRequired("privkey")
	cmd.Flags().StringP("file", "f", "", "deploy config")
	_ = cmd.MarkFlagRequired("file")
}

//DeployContrcts ...
func DeployContrcts(cmd *cobra.Command, args []string) {
	var deployKeystr string
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	filepath, _ := cmd.Flags().GetString("file")
	privkey, _ := cmd.Flags().GetString("privkey")
	var deployCfg DepolyInfo
	InitCfg(filepath, &deployCfg)

	if privkey != "" {
		deployKeystr = privkey
	} else {
		deployKeystr = deployCfg.DeployerPrivateKey
	}
	deployPrivateKey, err := crypto.ToECDSA(common.FromHex(deployKeystr))
	if err != nil {
		panic(err)
	}
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

	deployerAddr := crypto.PubkeyToAddress(deployPrivateKey.PublicKey)
	para := &ethtxs.DeployPara{
		DeployPrivateKey: deployPrivateKey,
		Deployer:         deployerAddr,
		Operator:         deployerAddr,
		InitValidators:   validators,
		ValidatorPriKey:  []*ecdsa.PrivateKey{deployPrivateKey},
		InitPowers:       initPowers,
	}
	client, err := ethtxs.SetupWebsocketEthClient(rpcLaddr)
	_, x2EthDeployInfo, err := ethtxs.DeployAndInit(client, para)
	if err != nil {
		fmt.Println("DeployAndInit,err:", err.Error())
		return
	}
	bridgeRegistry := x2EthDeployInfo.BridgeRegistry.Address.String()

	fmt.Println("the BridgeRegistry address is:", bridgeRegistry)
}

func InitCfg(filepath string, cfg interface{}) {

	if _, err := tml.DecodeFile(filepath, cfg); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	return
}

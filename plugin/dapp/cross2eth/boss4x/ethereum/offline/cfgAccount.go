package offline

import (
	//"context"
	"fmt"
	"math/big"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	gnosis "github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/gnosis/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"

	//"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	//"github.com/ethereum/go-ethereum/core/types"
	//"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/ethclient"
	"github.com/spf13/cobra"
	//"math/big"
	"strings"
)

/*
./boss4x ethereum offline set_offline_addr -a 0xbf271b2B23DA4fA8Dc93Ce86D27dd09796a7Bf54 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a
./boss4x ethereum offline sign -f set_offline_addr.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt

./boss4x ethereum offline multisign_setup -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a -o 0x4c85848a7E2985B76f06a7Ed338FCB3aF94a7DCf,0x6F163E6daf0090D897AD7016484f10e0cE844994,0xbc333839E37bc7fAAD0137aBaE2275030555101f,0x495953A743ef169EC5D4aC7b5F786BF2Bd56aFd5 -m 0x871887bC6D9b21B82787a66145D38172cA816d09
tx is written to file:  multisign_setup.txt
./boss4x ethereum offline sign -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230 -f multisign_setup.txt
deployTxInfos size: 1
tx is written to file:  deploysigntxs.txt
./boss4x ethereum offline send -f deploysigntxs.txt
*/

func CreateCfgAccountTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_offline_addr",
		Short: "save config offline account",
		Run:   cfgAccountTx, //配置账户
	}
	addCfgTxFlags(cmd)
	return cmd
}

func addCfgTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "multisign address")
	_ = cmd.MarkFlagRequired("address")
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
}

func cfgAccountTx(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	address, _ := cmd.Flags().GetString("address")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	contract, _ := cmd.Flags().GetString("contract")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")

	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		fmt.Println("JSON NewReader Err:", err)
		return
	}

	abiData, err := bridgeAbi.Pack("configOfflineSaveAccount", common.HexToAddress(address))
	if err != nil {
		panic(err)
	}

	CreateTxInfoAndWrite(abiData, deployAddr, contract, "set_offline_addr", url, chainEthId)
}

func SetupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisign_setup",
		Short: "Setup owners to contract",
		Run:   SetupOwner,
	}
	SetupOwnerFlags(cmd)
	return cmd
}

func SetupOwnerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner", "o", "", "owners's address, separated by ','")
	_ = cmd.MarkFlagRequired("owner")
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("multisign", "m", "", "multisign contract address")
	_ = cmd.MarkFlagRequired("multisign")
}

func SetupOwner(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	multisign, _ := cmd.Flags().GetString("multisign")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")
	owner, _ := cmd.Flags().GetString("owner")
	owners := strings.Split(owner, ",")

	var _owners []common.Address
	for _, onwer := range owners {
		_owners = append(_owners, common.HexToAddress(onwer))
	}
	AddressZero := common.HexToAddress(ebTypes.EthNilAddr)

	gnoAbi, err := abi.JSON(strings.NewReader(gnosis.GnosisSafeABI))
	if err != nil {
		fmt.Println("JSON Err:", err)
		return
	}

	abiData, err := gnoAbi.Pack("setup", _owners, big.NewInt(int64(len(_owners))), AddressZero, []byte{'0', 'x'},
		AddressZero, AddressZero, big.NewInt(int64(0)), AddressZero)
	if err != nil {
		fmt.Println("Pack execTransaction Err:", err)
		return
	}

	CreateTxInfoAndWrite(abiData, deployAddr, multisign, "multisign_setup", url, chainEthId)

}

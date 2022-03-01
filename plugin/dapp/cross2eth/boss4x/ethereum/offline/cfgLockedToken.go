package offline

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	ebTypes "github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/types"
	"github.com/33cn/plugin/plugin/dapp/cross2eth/ebrelayer/utils"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"
)

/*
./boss4x ethereum offline set_offline_token -s ETH -m 20 -c 0xC65B02a22B714b55D708518E2426a22ffB79113d -d 0x8afdadfc88a1087c9a1d6c0f5dd04634b87f303a
./boss4x ethereum offline sign -f set_offline_token.txt -k 8656d2bc732a8a816a461ba5e2d8aac7c7f85c26a813df30d5327210465eb230
./boss4x ethereum offline send -f deploysigntxs.txt
*/

func ConfigLockedTokenOfflineSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_offline_token",
		Short: "set config offline locked token",
		Run:   ConfigLockedTokenOfflineSave,
	}
	ConfigLockedTokenOfflineSaveFlags(cmd)
	return cmd
}

//ConfigLockedTokenOfflineSaveFlags ...
func ConfigLockedTokenOfflineSaveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("token", "t", "", "token addr")
	//_ = cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().Float64P("threshold", "m", 0, "threshold")
	_ = cmd.MarkFlagRequired("threshold")
	cmd.Flags().Uint8P("percents", "p", 50, "percents")
	//_ = cmd.MarkFlagRequired("percents")
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
}

func ConfigLockedTokenOfflineSave(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	symbol, _ := cmd.Flags().GetString("symbol")
	token, _ := cmd.Flags().GetString("token")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	percents, _ := cmd.Flags().GetUint8("percents")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	contract, _ := cmd.Flags().GetString("contract")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")

	d, err := utils.GetDecimalsFromNode(token, url)
	if err != nil {
		fmt.Println("get decimals error", err.Error())
		return
	}

	realAmount := utils.ToWei(threshold, d)

	bn := big.NewInt(1)
	bn, _ = bn.SetString(utils.TrimZeroAndDot(realAmount.String()), 10)

	if percents > 100 || percents < 0 {
		fmt.Println("param percents err")
		return
	}

	if token == "" || symbol == "ETH" {
		token = ebTypes.EthNilAddr
	}
	tokenAddr := common.HexToAddress(token)

	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		fmt.Println("JSON NewReader Err:", err)
		return
	}

	abiData, err := bridgeAbi.Pack("configLockedTokenOfflineSave", tokenAddr, symbol, bn, percents)
	if err != nil {
		panic(err)
	}

	CreateTxInfoAndWrite(abiData, deployAddr, contract, "set_offline_token", url, chainEthId)
}

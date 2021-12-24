package offline

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"github.com/33cn/plugin/plugin/dapp/cross2eth/contracts/contracts4eth/generated"
	"github.com/spf13/cobra"
)

func ConfigplatformTokenSymbolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set_symbol",
		Short: "save config symbol",
		Run:   ConfigplatformTokenSymbol,
	}
	addConfigplatformTokenSymbolFlags(cmd)
	return cmd
}

func addConfigplatformTokenSymbolFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "ETH", "symbol")
	_ = cmd.MarkFlagRequired("symbol")
	cmd.Flags().StringP("deployAddr", "d", "", "deploy contract addr")
	_ = cmd.MarkFlagRequired("deployAddr")
	cmd.Flags().StringP("contract", "c", "", "bridgebank contract address")
	_ = cmd.MarkFlagRequired("contract")
}

func ConfigplatformTokenSymbol(cmd *cobra.Command, _ []string) {
	url, _ := cmd.Flags().GetString("rpc_laddr_ethereum")
	symbol, _ := cmd.Flags().GetString("symbol")
	deployAddr, _ := cmd.Flags().GetString("deployAddr")
	contract, _ := cmd.Flags().GetString("contract")
	chainEthId, _ := cmd.Flags().GetInt64("chainEthId")

	bridgeAbi, err := abi.JSON(strings.NewReader(generated.BridgeBankABI))
	if err != nil {
		fmt.Println("JSON NewReader Err:", err)
		return
	}

	abiData, err := bridgeAbi.Pack("configplatformTokenSymbol", symbol)
	if err != nil {
		panic(err)
	}

	CreateTxInfoAndWrite(abiData, deployAddr, contract, "set_symbol", url, chainEthId)
}

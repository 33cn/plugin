package ethereum

import (
	"github.com/33cn/plugin/plugin/dapp/cross2eth/boss4x/ethereum/offline"
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
		offline.DeployOfflineContractsCmd(),
	)
	return cmd

}

package chain33

import (
	"github.com/33cn/plugin/plugin/dapp/cross2eth/boss4x/chain33/offline"
	"github.com/spf13/cobra"
)

func Chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain33",
		Short: "deploy to chain33",
	}
	cmd.AddCommand(
		//deployCrossContractsCmd(),
		offline.Boss4xOfflineCmd(),
	)
	return cmd

}

func deployCrossContractsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "deploy all of the contracts for cross ",
		Run:   deployCrossContracts,
	}
	addDeployCrossContractsFlags(cmd)
	return cmd
}

func addDeployCrossContractsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
}

func deployCrossContracts(cmd *cobra.Command, args []string) {

}

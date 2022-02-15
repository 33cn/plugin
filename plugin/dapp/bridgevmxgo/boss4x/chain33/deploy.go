package chain33

import (
	"github.com/33cn/plugin/plugin/dapp/bridgevmxgo/boss4x/chain33/offline"
	"github.com/spf13/cobra"
)

func Chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain33",
		Short: "deploy to chain33",
	}
	cmd.AddCommand(
		offline.Boss4xOfflineCmd(),
		NewOracleClaimCmd(),
	)
	return cmd

}

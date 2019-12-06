/*Package commands implement dapp client commands*/
package commands

import (
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd exchange client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange",
		Short: "exchange command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
	//add sub command
	)
	return cmd
}

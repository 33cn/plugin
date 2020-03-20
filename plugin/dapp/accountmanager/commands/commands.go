/*Package commands implement dapp client commands*/
package commands

import (
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd accountmanager client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accountmanager",
		Short: "accountmanager command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
	//add sub command
	)
	return cmd
}

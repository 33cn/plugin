package offline

import (
	"github.com/33cn/plugin/plugin/dapp/dex/utils"
	"github.com/spf13/cobra"
)

func Chain33OfflineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chain33",
		Short: "create and sign offline tx to deploy and set dex contracts to chain33",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		createERC20ContractCmd(),
		createRouterCmd(),
		farmofflineCmd(),
		sendSignTxs2Chain33Cmd(),
	)
	return cmd
}

func createRouterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "router",
		Short: "create and sign offline weth9, factory and router contracts",
		Run:   createRouterContract,
	}
	addCreateRouterFlags(cmd)
	return cmd
}

func addCreateRouterFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "the private key")
	cmd.MarkFlagRequired("key")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")

	cmd.Flags().StringP("feeToSetter", "a", "", "address for fee to Setter")
	cmd.MarkFlagRequired("feeToSetter")

}

func sendSignTxs2Chain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "send one or serval dex txs to chain33 in serial",
		Run:   sendSignTxs2Chain33,
	}
	addSendSignTxs2Chain33Flags(cmd)
	return cmd
}

func addSendSignTxs2Chain33Flags(cmd *cobra.Command) {
	cmd.Flags().StringP("path", "p", "./", "(optional)path of txs file,default to current directroy")
	cmd.Flags().StringP("file", "f", "", "file name which contains the txs to be sent to chain33")
	_ = cmd.MarkFlagRequired("file")
}

func sendSignTxs2Chain33(cmd *cobra.Command, args []string) {
	filePath, _ := cmd.Flags().GetString("path")
	file, _ := cmd.Flags().GetString("file")
	url, _ := cmd.Flags().GetString("rpc_laddr")
	filePath += file
	utils.SendSignTxs2Chain33(filePath, url)
}

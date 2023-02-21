/*Package commands implement dapp client commands*/
package commands

import (
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd rollup client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollup",
		Short: "rollup command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		validatorCMD(),
		rollupStatusCMD(),
		roundInfoCMD(),
	)
	return cmd
}

func sendQueryRPC(cmd *cobra.Command, funcName string, req, reply types.Message) {
	rpcAddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	payLoad := types.MustPBToJSON(req)
	query := &rpctypes.Query4Jrpc{
		Execer:   types.GetExecName(rtypes.RollupX, paraName),
		FuncName: funcName,
		Payload:  payLoad,
	}

	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.Query", query, reply)
	ctx.Run()
}

func markRequired(cmd *cobra.Command, params ...string) {
	for _, param := range params {
		_ = cmd.MarkFlagRequired(param)
	}
}

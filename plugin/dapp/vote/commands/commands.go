/*Package commands implement dapp client commands*/
package commands

import (
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	vty "github.com/33cn/plugin/plugin/dapp/vote/types"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd vote client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		//create tx
		createGroupCMD(),
		updateGroupCMD(),
		createVoteCMD(),
		commitVoteCMD(),
		closeVoteCMD(),
		updateMemberCMD(),
		//query rpc
		groupInfoCMD(),
		voteInfoCMD(),
		memberInfoCMD(),
		listGroupCMD(),
		listVoteCMD(),
		listMemberCMD(),
	)
	return cmd
}

func markRequired(cmd *cobra.Command, params ...string) {
	for _, param := range params {
		_ = cmd.MarkFlagRequired(param)
	}
}

func sendCreateTxRPC(cmd *cobra.Command, actionName string, req types.Message) {
	rpcAddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	payLoad := types.MustPBToJSON(req)
	pm := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(vty.VoteX, paraName),
		ActionName: actionName,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

func sendQueryRPC(cmd *cobra.Command, funcName string, req, reply types.Message) {
	rpcAddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	payLoad := types.MustPBToJSON(req)
	query := &rpctypes.Query4Jrpc{
		Execer:   types.GetExecName(vty.VoteX, paraName),
		FuncName: funcName,
		Payload:  payLoad,
	}

	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.Query", query, reply)
	ctx.Run()
}

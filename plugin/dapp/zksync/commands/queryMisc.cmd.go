package commands

import (
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func getQueueIDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue_id",
		Short: "get queue id used from L1 Ethereum",
		Run:   getQueueID,
	}
	return cmd
}

func getQueueID(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	params.FuncName = "GetQueueID"

	var resp zt.EthPriorityQueueID
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}


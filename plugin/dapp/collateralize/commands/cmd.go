package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
)

// CollateralizeCmd 斗牛游戏命令行
func CollateralizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateralize",
		Short: "Collateralize command",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CollateralizeCreateRawTxCmd(),
		//CollateralizeBorrowRawTxCmd(),
		//CollateralizeAppendRawTxCmd(),
		//CollateralizeRepayRawTxCmd(),
		//CollateralizePriceFeedRawTxCmd(),
		//CollateralizeCloseRawTxCmd(),
		//CollateralizeManageRawTxCmd(),
	)

	return cmd
}

// CollateralizeCreateRawTxCmd 生成开始交易命令行
func CollateralizeCreateRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collateralize",
		Run:   CollateralizeCreate,
	}
	addCollateralizeCreateFlags(cmd)
	return cmd
}

func addCollateralizeCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("balance", "b", 0, "balance")
	cmd.MarkFlagRequired("balance")
}

func CollateralizeCreate(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	balance, _ := cmd.Flags().GetUint64("balance")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.CollateralizeX),
		ActionName: "Create",
		Payload:    []byte(fmt.Sprintf("{\"balance\":%d}", balance)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}
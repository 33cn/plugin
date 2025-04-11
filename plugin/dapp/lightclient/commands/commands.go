/*Package commands implement dapp client commands*/
package commands

import (
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ltypes "github.com/33cn/plugin/plugin/dapp/lightclient/lighttypes"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd lightclient client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lightcli",
		Short: "lightclient command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		btcCommand(),
	)
	return cmd
}

func btcCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "btc",
		Short: "btc lightclient command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		btcLastHeaderCMD(),
		btcHeaderCMD(),
	)
	return cmd
}

func btcLastHeaderCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last",
		Short: "show btc last header",
		Run:   getBtcLast,
	}

	return cmd
}

func getBtcLast(cmd *cobra.Command, args []string) {

	params := &types.ReqNil{}
	header := &ltypes.BtcHeader{}
	sendQueryRPC(cmd, "GetBtcLastHeader", params, header)
}

func btcHeaderCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "header",
		Short: "show btc header info",
		Run:   getBtcHeader,
	}
	cmd.Flags().Int64P("height", "h", 0, "btc height")
	markRequired(cmd, "height")

	return cmd
}

func getBtcHeader(cmd *cobra.Command, args []string) {

	height, _ := cmd.Flags().GetInt64("height")
	params := &types.ReqInt{Height: height}
	header := &ltypes.BtcHeader{}
	sendQueryRPC(cmd, "GetBtcHeader", params, header)
}

func sendQueryRPC(cmd *cobra.Command, funcName string, req, reply types.Message) {
	rpcAddr, _ := cmd.Flags().GetString("rpc_laddr")
	payLoad := types.MustPBToJSON(req)
	query := &rpctypes.Query4Jrpc{
		Execer:   ltypes.LightclientX,
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

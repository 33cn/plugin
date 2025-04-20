/*Package commands implement dapp client commands*/
package commands

import (
	"encoding/json"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

// Cmd rgbx client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rgbx",
		Short: "rgbx command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		mintAssetCMD(),
		transferAssetCMD(),
		getAssetCMD(),
		getPendingTxCMD(),
		listPendingTxCMD(),
		getConfirmedHeightCMD(),
	)
	return cmd
}

func sendQueryRPC(cmd *cobra.Command, funcName string, req, reply types.Message, runResult bool) {
	rpcAddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	payLoad := types.MustPBToJSON(req)
	query := &rpctypes.Query4Jrpc{
		Execer:   types.GetExecName(rtypes.RgbxX, paraName),
		FuncName: funcName,
		Payload:  payLoad,
	}

	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.Query", query, reply)
	if runResult {
		_, _ = ctx.RunResult()
		return
	}
	ctx.SetResultCb(func(res interface{}) (interface{}, error) {
		msg, ok := res.(types.Message)
		if !ok {
			return res, nil
		}
		resData, err := types.PBToJSONUTF8(msg)
		if err != nil {
			return res, nil
		}
		return json.RawMessage(resData), nil
	})
	ctx.Run()
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
		Execer:     types.GetExecName(rtypes.RgbxX, paraName),
		ActionName: actionName,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.CreateTransaction", pm, &res)
	ctx.RunWithoutMarshal()
}

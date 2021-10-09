package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/spf13/cobra"
)

func superNodeBindCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorizes_miner",
		Short: "super node authorizes mining",
		Run:   createSuperNodeBindTx,
	}
	addSuperNodeBindFlags(cmd)
	return cmd
}

func addSuperNodeBindFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("superNode", "s", "", "super node to bind/unbind miner")
	cmd.MarkFlagRequired("superNode")
	cmd.Flags().StringP("minerNode", "m", "", "authorizes miner node, empty is unbind")
	cmd.Flags().StringP("minerBlsPubKey", "b", "", "authorizes miner node bls pubkey")
}

func createSuperNodeBindTx(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	superNode, _ := cmd.Flags().GetString("superNode")
	minerNode, _ := cmd.Flags().GetString("minerNode")
	minerBlsPubKey, _ := cmd.Flags().GetString("minerBlsPubKey")

	if !strings.HasPrefix(paraName, "user.p") {
		fmt.Fprintln(os.Stderr, "paraName is not right, paraName format like `user.p.guodun.`")
		return
	}

	payload := &pt.ParaSuperNodeBindMiner{SuperAddress: superNode, MinerAddress: minerNode, MinerBlsPubKey: minerBlsPubKey}
	params := &rpctypes.CreateTxIn{
		Execer:     getRealExecName(paraName, pt.ParaX),
		ActionName: "SuperNodeBindMiner",
		Payload:    types.MustPBToJSON(payload),
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// getSuperNodeBindInfoCmd Get super node bind info
func getSuperNodeBindInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "super_bind_info",
		Short: "Get super node bind info",
		Run:   superNodeBindInfo,
	}
	addSuperNodeBindInfoCmdFlags(cmd)
	return cmd
}

func addSuperNodeBindInfoCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "super node")
	cmd.MarkFlagRequired("addr")
}

func superNodeBindInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetSuperNodeBindInfo"
	req := types.ReqString{Data: addr}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaNodeAddrIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// getAuthorizedNodeBindSuperNodeInfoCmd Get authorized node bind info
func getAuthorizedNodeBindSuperNodeInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorized_info",
		Short: "Get authorized node bind info",
		Run:   authorizedNodeBindSuperNodeInfo,
	}
	addAuthorizedNodeBindSuperNodeInfoCmdFlags(cmd)
	return cmd
}

func addAuthorizedNodeBindSuperNodeInfoCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "authorized node")
	cmd.MarkFlagRequired("addr")
}

func authorizedNodeBindSuperNodeInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc
	params.Execer = pt.ParaX
	params.FuncName = "GetAuthorizedNodeBindSuperNodeInfo"
	req := types.ReqString{Data: addr}
	params.Payload = types.MustPBToJSON(&req)

	var res pt.ParaNodeAddrIdStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

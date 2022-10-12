package l2txs

import (
	"fmt"
	"strconv"
	"strings"

	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func proxyManyExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxyexit_many",
		Short: "proxy exit many",
		Run:   proxyManyExit,
	}
	proxyManyExitFlag(cmd)
	return cmd
}

func proxyManyExitFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("proxyIDs", "p", "0", "L2 proxy account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("proxyIDs")
	cmd.Flags().StringP("targetIDs", "g", "0", "L2 target account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("targetIDs")
}

func proxyManyExit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKeys, _ := cmd.Flags().GetString("keys")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	proxyId, _ := cmd.Flags().GetString("proxyIDs")
	targetId, _ := cmd.Flags().GetString("targetIDs")
	paraName, _ := cmd.Flags().GetString("paraName")

	proxyIds := strings.Split(proxyId, ",")
	targetIds := strings.Split(targetId, ",")
	keys := strings.Split(privateKeys, ",")

	if len(proxyIds) != len(keys) || len(targetIds) != len(keys) {
		fmt.Println("err len(proxyIds) != len(keys) or len(targetIds) != len(keys)",
			len(proxyIds), "!=", len(keys), " or ", len(targetIds), "!=", len(keys))
		return
	}

	for i := 0; i < len(proxyIds); i++ {
		proxyId, _ := strconv.ParseInt(proxyIds[i], 10, 64)
		targetId, _ := strconv.ParseInt(targetIds[i], 10, 64)
		param := &zksyncTypes.ZkProxyExit{
			TokenId:  tokenId,
			ProxyId:  uint64(proxyId),
			TargetId: uint64(targetId),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyProxyExitAction,
			Value: &zksyncTypes.ZksyncAction_ProxyExit{
				ProxyExit: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("send ForceExit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

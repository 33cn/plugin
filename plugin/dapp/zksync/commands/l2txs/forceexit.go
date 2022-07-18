package l2txs

import (
	"fmt"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

func forceManyExitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forceexit_many",
		Short: "force exit many",
		Run:   forceManyExit,
	}
	forceManyExitFlag(cmd)
	return cmd
}

func forceManyExitFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("accountIDs", "a", "0", "L2 account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
}

func forceManyExit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKeys, _ := cmd.Flags().GetString("keys")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	accountIDs, _ := cmd.Flags().GetString("accountIDs")

	ids := strings.Split(accountIDs, ",")
	keys := strings.Split(privateKeys, ",")

	if len(ids) != len(keys) {
		fmt.Println("err len(ids) != len(keys)", len(ids), "!=", len(keys))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		param := &zksyncTypes.ZkForceExit{
			TokenId:   tokenId,
			AccountId: uint64(id),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyForceExitAction,
			Value: &zksyncTypes.ZksyncAction_ForceExit{
				ForceExit: param,
			},
		}

		tx, err := createChain33Tx(keys[i], action)
		if nil != err {
			fmt.Println("send ForceExit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

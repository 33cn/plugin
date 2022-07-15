package l2txs

import (
	"fmt"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func sendWithdrawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "send withdraw tx to chain33",
		Run:   sendWithdraw,
	}
	sendWithdrawFlags(cmd)
	return cmd
}

func sendWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")

	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("accountID", "a", 0, "L2 account id on chain33")
	_ = cmd.MarkFlagRequired("accountID")
	cmd.Flags().StringP("amount", "m", "0", "deposit amount")
	_ = cmd.MarkFlagRequired("amount")
}

func sendWithdraw(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	privateKey, _ := cmd.Flags().GetString("key")

	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	accountID, _ := cmd.Flags().GetUint64("accountID")
	amount, _ := cmd.Flags().GetString("amount")

	withdraw := &zksyncTypes.ZkWithdraw{
		TokenId:   tokenId,
		Amount:    amount,
		AccountId: accountID,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyWithdrawAction,
		Value: &zksyncTypes.ZksyncAction_Withdraw{
			Withdraw: withdraw,
		},
	}

	tx, err := createChain33Tx(privateKey, action)
	if nil != err {
		fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
		return
	}
	sendTx(rpcLaddr, tx)
}

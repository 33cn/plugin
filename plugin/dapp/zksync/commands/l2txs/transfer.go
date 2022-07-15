package l2txs

import (
	"fmt"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func SendTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "send transfer tx to chain33",
		Run:   tranfer,
	}
	sendTransferFlags(cmd)
	return cmd
}

func sendTransferFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("from", "f", 0, "from account id")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().Uint64P("to", "d", 0, "to account id")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("amount", "m", "0", "deposit amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func tranfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	fromAccountId, _ := cmd.Flags().GetUint64("from")
	toAccountId, _ := cmd.Flags().GetUint64("to")
	privateKey, _ := cmd.Flags().GetString("key")

	transfer := &zksyncTypes.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amount,
		FromAccountId: fromAccountId,
		ToAccountId:   toAccountId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferAction,
		Value: &zksyncTypes.ZksyncAction_Transfer{
			Transfer: transfer,
		},
	}

	tx, err := createChain33Tx(privateKey, action)
	if nil != err {
		fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
		return
	}
	sendTx(rpcLaddr, tx)
}

func BatchSendTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batchtransfer",
		Short: "send transfer tx to chain33 batch",
		Run:   batchSendTransfer,
	}
	batchSendTransferFlags(cmd)
	return cmd
}

func batchSendTransferFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("count", "c", 1, "count of txs to send in batch and max is 100")
	_ = cmd.MarkFlagRequired("count")

	cmd.Flags().Uint64P("from", "f", 0, "from account id")
	_ = cmd.MarkFlagRequired("from")
	cmd.Flags().Uint64P("to", "d", 0, "to account id")
	_ = cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("amount", "m", "1", "transfer amount")
	_ = cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func batchSendTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	count, _ := cmd.Flags().GetUint64("count")
	amount, _ := cmd.Flags().GetString("amount")
	fromAccountId, _ := cmd.Flags().GetUint64("from")
	toAccountId, _ := cmd.Flags().GetUint64("to")
	privateKey, _ := cmd.Flags().GetString("key")

	transfer := &zksyncTypes.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amount,
		FromAccountId: fromAccountId,
		ToAccountId:   toAccountId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferAction,
		Value: &zksyncTypes.ZksyncAction_Transfer{
			Transfer: transfer,
		},
	}

	if count > 100 {
		count = 100
	}

	for i := uint64(0); i < count; i++ {
		tx, err := createChain33Tx(privateKey, action)
		if nil != err {
			fmt.Println("send transfer failed to createChain33Tx due to err:", err.Error())
			return
		}
		fmt.Println("tx with index", i)
		sendTx(rpcLaddr, tx)
	}
}

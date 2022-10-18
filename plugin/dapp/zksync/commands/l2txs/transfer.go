package l2txs

import (
	"fmt"
	"strconv"
	"strings"

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
	paraName, _ := cmd.Flags().GetString("paraName")

	transfer := &zksyncTypes.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amount,
		FromAccountId: fromAccountId,
		ToAccountId:   toAccountId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferAction,
		Value: &zksyncTypes.ZksyncAction_ZkTransfer{
			ZkTransfer: transfer,
		},
	}

	tx, err := createChain33Tx(privateKey, getRealExecName(paraName, zksyncTypes.Zksync), action)
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
	paraName, _ := cmd.Flags().GetString("paraName")

	transfer := &zksyncTypes.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amount,
		FromAccountId: fromAccountId,
		ToAccountId:   toAccountId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferAction,
		Value: &zksyncTypes.ZksyncAction_ZkTransfer{
			ZkTransfer: transfer,
		},
	}

	if count > 100 {
		count = 100
	}

	for i := uint64(0); i < count; i++ {
		tx, err := createChain33Tx(privateKey, getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("send transfer failed to createChain33Tx due to err:", err.Error())
			return
		}
		fmt.Println("tx with index", i)
		sendTx(rpcLaddr, tx)
	}
}

func SendManyTransferTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_many",
		Short: "send many transfer tx to chain33",
		Run:   tranferMany,
	}
	sendManyTransferFlags(cmd)
	return cmd
}

func sendManyTransferFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "contractToTree amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("fromIDs", "f", "0", "from account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("fromIDs")
	cmd.Flags().StringP("toIDs", "d", "0", "to account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func tranferMany(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	fromIDs, _ := cmd.Flags().GetString("fromIDs")
	toIDs, _ := cmd.Flags().GetString("toIDs")
	privateKeys, _ := cmd.Flags().GetString("keys")
	paraName, _ := cmd.Flags().GetString("paraName")

	fids := strings.Split(fromIDs, ",")
	tids := strings.Split(toIDs, ",")
	keys := strings.Split(privateKeys, ",")

	if len(fids) != len(tids) || len(fids) != len(keys) {
		fmt.Println("err len(fids) != len(tids) != len(keys)", len(fids), "!=", len(tids), "!=", len(keys))
		return
	}

	for i := 0; i < len(fids); i++ {
		fid, _ := strconv.ParseInt(fids[i], 10, 64)
		tid, _ := strconv.ParseInt(tids[i], 10, 64)
		param := &zksyncTypes.ZkTransfer{
			TokenId:       tokenId,
			Amount:        amount,
			FromAccountId: uint64(fid),
			ToAccountId:   uint64(tid),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTransferAction,
			Value: &zksyncTypes.ZksyncAction_ZkTransfer{
				ZkTransfer: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

func SendManyTransferTxFromOneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_many_2",
		Short: "from one addr, send many transfer tx to chain33",
		Run:   tranferManyFromOne,
	}
	sendManyTransferFromOneFlags(cmd)
	return cmd
}

func sendManyTransferFromOneFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "contractToTree amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("fromID", "f", "0", "from account id")
	_ = cmd.MarkFlagRequired("fromID")
	cmd.Flags().StringP("toIDs", "d", "0", "to account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func tranferManyFromOne(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	fromID, _ := cmd.Flags().GetString("fromID")
	toIDs, _ := cmd.Flags().GetString("toIDs")
	key, _ := cmd.Flags().GetString("key")
	paraName, _ := cmd.Flags().GetString("paraName")

	tids := strings.Split(toIDs, ",")

	fid, _ := strconv.ParseInt(fromID, 10, 64)
	for i := 0; i < len(tids); i++ {
		tid, _ := strconv.ParseInt(tids[i], 10, 64)
		param := &zksyncTypes.ZkTransfer{
			TokenId:       tokenId,
			Amount:        amount,
			FromAccountId: uint64(fid),
			ToAccountId:   uint64(tid),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTransferAction,
			Value: &zksyncTypes.ZksyncAction_ZkTransfer{
				ZkTransfer: param,
			},
		}

		tx, err := createChain33Tx(key, getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

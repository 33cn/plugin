package l2txs

import (
	"fmt"
	"strings"

	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func sendDepositTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "send deposit tx to chain33",
		Run:   sendDeposit,
	}
	sendDepositFlags(cmd)
	return cmd
}

func sendDepositFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Int64P("queueId", "q", 0, "deposit priority queue id")
	_ = cmd.MarkFlagRequired("queueId")
	cmd.Flags().StringP("amount", "m", "0", "deposit amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddr", "e", "", "from eth addr")
	_ = cmd.MarkFlagRequired("ethAddr")
	cmd.Flags().StringP("chain33Addr", "a", "", "to chain33 addr")
	_ = cmd.MarkFlagRequired("chain33Addr")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func sendDeposit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	queueId, _ := cmd.Flags().GetInt64("queueId")
	amount, _ := cmd.Flags().GetString("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")
	privateKey, _ := cmd.Flags().GetString("key")
	paraName, _ := cmd.Flags().GetString("paraName")

	deposit := &zksyncTypes.ZkDeposit{
		TokenId:      tokenId,
		Amount:       amount,
		EthAddress:   ethAddress,
		Chain33Addr:  chain33Addr,
		L1PriorityId: queueId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyDepositAction,
		Value: &zksyncTypes.ZksyncAction_Deposit{
			Deposit: deposit,
		},
	}

	tx, err := createChain33Tx(privateKey, getRealExecName(paraName, zksyncTypes.Zksync), action)
	if nil != err {
		fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
		return
	}
	sendTx(rpcLaddr, tx)
}

func batchSendDepositTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batchdeposit",
		Short: "send deposit tx to chain33 batch",
		Run:   batchSendDeposit,
	}
	batchSendDepositFlags(cmd)
	return cmd
}

func batchSendDepositFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint64P("count", "c", 1, "count of txs to send in batch")
	_ = cmd.MarkFlagRequired("count")
	cmd.Flags().Int64P("queueId", "q", 0, "deposit queue id")
	_ = cmd.MarkFlagRequired("queueId")
	cmd.Flags().StringP("amount", "m", "0", "deposit amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddr", "e", "", "from eth addr")
	_ = cmd.MarkFlagRequired("ethAddr")
	cmd.Flags().StringP("chain33Addr", "a", "", "to chain33 addr")
	_ = cmd.MarkFlagRequired("chain33Addr")

	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func batchSendDeposit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	count, _ := cmd.Flags().GetUint64("count")
	queueId, _ := cmd.Flags().GetInt64("queueId")
	amount, _ := cmd.Flags().GetString("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")
	privateKey, _ := cmd.Flags().GetString("key")
	paraName, _ := cmd.Flags().GetString("paraName")

	deposit := &zksyncTypes.ZkDeposit{
		TokenId:      tokenId,
		Amount:       amount,
		EthAddress:   ethAddress,
		Chain33Addr:  chain33Addr,
		L1PriorityId: queueId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyDepositAction,
		Value: &zksyncTypes.ZksyncAction_Deposit{
			Deposit: deposit,
		},
	}

	for i := uint64(0); i < count; i++ {
		tx, err := createChain33Tx(privateKey, getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

func sendManyDepositTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit_many",
		Short: "send many deposit tx to chain33",
		Run:   sendManyDeposit,
	}
	sendManyDepositFlags(cmd)
	return cmd
}

func sendManyDepositFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "eth token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Int64P("queueId", "q", 0, "deposit queue id")
	_ = cmd.MarkFlagRequired("queueId")
	cmd.Flags().StringP("amount", "m", "0", "deposit amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddr", "e", "", "from eth addr")
	_ = cmd.MarkFlagRequired("ethAddr")
	cmd.Flags().StringP("chain33Addrs", "a", "", "to chain33 addrs, use ',' separate")
	_ = cmd.MarkFlagRequired("chain33Addrs")
	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func sendManyDeposit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	queueId, _ := cmd.Flags().GetInt64("queueId")
	amount, _ := cmd.Flags().GetString("amount")
	ethAddress, _ := cmd.Flags().GetString("ethAddr")
	chain33Addrs, _ := cmd.Flags().GetString("chain33Addrs")
	privateKey, _ := cmd.Flags().GetString("key")
	paraName, _ := cmd.Flags().GetString("paraName")

	toChain33Addrs := strings.Split(chain33Addrs, ",")

	for i := 0; i < len(toChain33Addrs); i++ {
		deposit := &zksyncTypes.ZkDeposit{
			TokenId:      tokenId,
			Amount:       amount,
			EthAddress:   ethAddress,
			Chain33Addr:  toChain33Addrs[i],
			L1PriorityId: queueId,
		}
		queueId++

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyDepositAction,
			Value: &zksyncTypes.ZksyncAction_Deposit{
				Deposit: deposit,
			},
		}

		tx, err := createChain33Tx(privateKey, getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

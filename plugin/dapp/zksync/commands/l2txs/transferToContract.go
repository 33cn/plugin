package l2txs

import (
	"fmt"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
	"strconv"
	"strings"
)

func treeManyToContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tree2contract_many",
		Short: "get treeToContract tx many",
		Run:   treeManyToContract,
	}
	treeManyToContractFlag(cmd)
	return cmd
}

func treeManyToContractFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "treeToContract tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "treeToContract amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("accountIDs", "a", "0", "L2 account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func treeManyToContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountIDs, _ := cmd.Flags().GetString("accountIDs")
	privateKeys, _ := cmd.Flags().GetString("keys")

	ids := strings.Split(accountIDs, ",")
	keys := strings.Split(privateKeys, ",")

	if len(ids) != len(keys) {
		fmt.Println("err len(ids) != len(keys)", len(ids), "!=", len(keys))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		param := &zksyncTypes.ZkTreeToContract{
			TokenId:   tokenId,
			Amount:    amount,
			AccountId: uint64(id),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTreeToContractAction,
			Value: &zksyncTypes.ZksyncAction_TreeToContract{
				TreeToContract: param,
			},
		}

		tx, err := createChain33Tx(keys[i], action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

func contractManyToTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contract2tree_many",
		Short: "get contractToTree tx many",
		Run:   contractManyToTree,
	}
	contractManyToTreeFlag(cmd)
	return cmd
}

func contractManyToTreeFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "contractToTree tokenId")
	cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "contractToTree amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("accountIDs", "a", "0", "L2 account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("accountIDs")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func contractManyToTree(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	accountIDs, _ := cmd.Flags().GetString("accountIDs")
	privateKeys, _ := cmd.Flags().GetString("keys")

	ids := strings.Split(accountIDs, ",")
	keys := strings.Split(privateKeys, ",")

	if len(ids) != len(keys) {
		fmt.Println("err len(ids) != len(keys)", len(ids), "!=", len(keys))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		param := &zksyncTypes.ZkContractToTree{
			TokenId:   tokenId,
			Amount:    amount,
			AccountId: uint64(id),
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyContractToTreeAction,
			Value: &zksyncTypes.ZksyncAction_ContractToTree{
				ContractToTree: param,
			},
		}

		tx, err := createChain33Tx(keys[i], action)
		if nil != err {
			fmt.Println("sendDeposit failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

package l2txs

import (
	"fmt"
	"strconv"
	"strings"

	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func transferManyToNewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer2new_many",
		Short: "get many transferToNew tx",
		Run:   transferManyToNew,
	}
	transferManyToNewFlag(cmd)
	return cmd
}

func transferManyToNewFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "transferToNew tokenId")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "transferToNew amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "transferToNew toEthAddress")
	_ = cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("fromIDs", "f", "0", "from account ids on chain33, use ',' separate")
	_ = cmd.MarkFlagRequired("fromIDs")
	cmd.Flags().StringP("chain33Addrs", "d", "0", "transferToNew toChain33Addrs, use ',' separate")
	_ = cmd.MarkFlagRequired("chain33Addrs")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func transferManyToNew(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	toEthAddress, _ := cmd.Flags().GetString("ethAddress")
	fromIDs, _ := cmd.Flags().GetString("fromIDs")
	chain33Addrs, _ := cmd.Flags().GetString("chain33Addrs")
	privateKeys, _ := cmd.Flags().GetString("keys")
	paraName, _ := cmd.Flags().GetString("paraName")

	fids := strings.Split(fromIDs, ",")
	addrs := strings.Split(chain33Addrs, ",")
	keys := strings.Split(privateKeys, ",")

	if len(fids) != len(addrs) || len(fids) != len(keys) {
		fmt.Println("err len(ids) != len(keys)", len(fids), "!=", len(addrs), "!=", len(keys))
		return
	}

	for i := 0; i < len(fids); i++ {
		fid, _ := strconv.ParseInt(fids[i], 10, 64)
		param := &zksyncTypes.ZkTransferToNew{
			TokenId:         tokenId,
			Amount:          amount,
			FromAccountId:   uint64(fid),
			ToEthAddress:    toEthAddress,
			ToLayer2Address: addrs[i],
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTransferToNewAction,
			Value: &zksyncTypes.ZksyncAction_TransferToNew{
				TransferToNew: param,
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

func transferToNewManyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer2new_to_many",
		Short: "from one id ,get many transferToNew tx",
		Run:   transferToNewMany,
	}
	transferToNewManyFlag(cmd)
	return cmd
}

func transferToNewManyFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("tokenId", "t", 0, "transferToNew tokenId")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().StringP("amount", "m", "0", "transferToNew amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("ethAddress", "e", "", "transferToNew toEthAddress")
	_ = cmd.MarkFlagRequired("ethAddress")
	cmd.Flags().StringP("fromID", "f", "0", "account id")
	_ = cmd.MarkFlagRequired("fromID")
	cmd.Flags().StringP("chain33Addrs", "d", "0", "transferToNew toChain33Addrs, use ',' separate")
	_ = cmd.MarkFlagRequired("chain33Addrs")
	cmd.Flags().StringP("key", "k", "", "private key")
	_ = cmd.MarkFlagRequired("key")
}

func transferToNewMany(cmd *cobra.Command, _ []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetString("amount")
	toEthAddress, _ := cmd.Flags().GetString("ethAddress")
	fromID, _ := cmd.Flags().GetString("fromID")
	chain33Addrs, _ := cmd.Flags().GetString("chain33Addrs")
	key, _ := cmd.Flags().GetString("key")
	paraName, _ := cmd.Flags().GetString("paraName")

	addrs := strings.Split(chain33Addrs, ",")
	fid, _ := strconv.ParseInt(fromID, 10, 64)

	for i := 0; i < len(addrs); i++ {
		param := &zksyncTypes.ZkTransferToNew{
			TokenId:         tokenId,
			Amount:          amount,
			FromAccountId:   uint64(fid),
			ToEthAddress:    toEthAddress,
			ToLayer2Address: addrs[i],
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTransferToNewAction,
			Value: &zksyncTypes.ZksyncAction_TransferToNew{
				TransferToNew: param,
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

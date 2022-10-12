package l2txs

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/33cn/chain33/types"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func nftManyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft",
		Short: "nft related cmd",
	}
	cmd.AddCommand(mintNFTCmd())
	cmd.AddCommand(transferNFTCmd())
	cmd.AddCommand(withdrawNFTCmd())

	return cmd
}

func mintNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "mint nft command",
		Run:   setMintNFT,
	}
	mintNFTFlag(cmd)
	return cmd
}

func mintNFTFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("creatorIds", "f", "0", "NFT creator ids, use ',' separate")
	_ = cmd.MarkFlagRequired("creatorIds")
	cmd.Flags().StringP("recipientIds", "t", "0", "NFT recipient ids, use ',' separate")
	_ = cmd.MarkFlagRequired("recipientIds")
	cmd.Flags().StringP("contentHashs", "e", "", "NFT content hash,must 64 hex chars, use ',' separate")
	_ = cmd.MarkFlagRequired("contentHashs")
	cmd.Flags().Uint64P("protocol", "p", 1, "NFT protocol, 1:ERC1155, 2: ERC721")
	_ = cmd.MarkFlagRequired("protocol")
	cmd.Flags().Uint64P("amount", "m", 1, "mint amount, only for ERC1155 case")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func setMintNFT(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	contentHashs, _ := cmd.Flags().GetString("contentHashs")
	protocol, _ := cmd.Flags().GetUint64("protocol")
	amount, _ := cmd.Flags().GetUint64("amount")
	accountIDs, _ := cmd.Flags().GetString("creatorIds")
	recipientIds, _ := cmd.Flags().GetString("recipientIds")
	privateKeys, _ := cmd.Flags().GetString("keys")
	paraName, _ := cmd.Flags().GetString("paraName")

	ids := strings.Split(accountIDs, ",")
	rids := strings.Split(recipientIds, ",")
	keys := strings.Split(privateKeys, ",")
	chashs := strings.Split(contentHashs, ",")

	if len(ids) != len(keys) || len(ids) != len(rids) || len(ids) != len(chashs) {
		fmt.Println("err len(ids) != len(keys) != len(rids) != len(chashs)", len(ids), "!=", len(keys), "!=", len(rids), "!=", len(chashs))
		return
	}

	if protocol == zksyncTypes.ZKERC721 && amount > 1 {
		_, _ = fmt.Fprintln(os.Stderr, errors.Wrapf(types.ErrInvalidParam, "NFT erc721 only allow 1 amount"))
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		rid, _ := strconv.ParseInt(rids[i], 10, 64)
		param := &zksyncTypes.ZkMintNFT{
			FromAccountId: uint64(id),
			RecipientId:   uint64(rid),
			ContentHash:   chashs[i],
			ErcProtocol:   protocol,
			Amount:        amount,
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyMintNFTAction,
			Value: &zksyncTypes.ZksyncAction_MintNFT{
				MintNFT: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("mint nft failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

func transferNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer nft command",
		Run:   transferNFT,
	}
	transferNFTFlag(cmd)
	return cmd
}

func transferNFTFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("fromIds", "a", "0", "NFT from ids")
	_ = cmd.MarkFlagRequired("fromIds")
	cmd.Flags().StringP("toIds", "r", "0", "NFT to ids")
	_ = cmd.MarkFlagRequired("toIds")
	cmd.Flags().StringP("tokenIds", "t", "0", "NFT token ids")
	_ = cmd.MarkFlagRequired("tokenIds")
	cmd.Flags().Uint64P("amount", "m", 1, "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func transferNFT(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenIds, _ := cmd.Flags().GetString("tokenIds")
	amount, _ := cmd.Flags().GetUint64("amount")
	fromIds, _ := cmd.Flags().GetString("fromIds")
	toIds, _ := cmd.Flags().GetString("toIds")
	privateKeys, _ := cmd.Flags().GetString("keys")
	paraName, _ := cmd.Flags().GetString("paraName")

	ids := strings.Split(fromIds, ",")
	tids := strings.Split(toIds, ",")
	keys := strings.Split(privateKeys, ",")
	tokenids := strings.Split(tokenIds, ",")
	if len(ids) != len(keys) || len(ids) != len(tids) || len(ids) != len(tokenids) {
		fmt.Println("err len(ids) != len(keys) != len(tids) != len(tokenids)", len(ids), "!=", len(keys), "!=", len(tids), "!=", len(tokenids))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		tid, _ := strconv.ParseInt(tids[i], 10, 64)
		tokenid, _ := strconv.ParseInt(tokenids[i], 10, 64)
		param := &zksyncTypes.ZkTransferNFT{
			FromAccountId: uint64(id),
			RecipientId:   uint64(tid),
			NFTTokenId:    uint64(tokenid),
			Amount:        amount,
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyTransferNFTAction,
			Value: &zksyncTypes.ZksyncAction_TransferNFT{
				TransferNFT: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("transfer nft failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

func withdrawNFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "withdraw to L1",
		Run:   withdrawNFT,
	}
	withdrawNFTFlag(cmd)
	return cmd
}

func withdrawNFTFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("fromIds", "a", "0", "NFT from ids")
	_ = cmd.MarkFlagRequired("fromIds")
	cmd.Flags().StringP("tokenIds", "t", "0", "NFT token ids")
	_ = cmd.MarkFlagRequired("tokenIds")
	cmd.Flags().Uint64P("amount", "m", 0, "amount")
	_ = cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("keys", "k", "", "private keys, use ',' separate")
	_ = cmd.MarkFlagRequired("keys")
}

func withdrawNFT(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenIds, _ := cmd.Flags().GetString("tokenIds")
	amount, _ := cmd.Flags().GetUint64("amount")
	fromIds, _ := cmd.Flags().GetString("fromIds")
	privateKeys, _ := cmd.Flags().GetString("keys")
	paraName, _ := cmd.Flags().GetString("paraName")

	ids := strings.Split(fromIds, ",")
	keys := strings.Split(privateKeys, ",")
	tokenids := strings.Split(tokenIds, ",")
	if len(ids) != len(keys) || len(ids) != len(tokenids) {
		fmt.Println("err len(ids) != len(keys) != len(tokenids)", len(ids), "!=", len(keys), "!=", len(tokenids))
		return
	}

	for i := 0; i < len(ids); i++ {
		id, _ := strconv.ParseInt(ids[i], 10, 64)
		tokenid, _ := strconv.ParseInt(tokenids[i], 10, 64)
		param := &zksyncTypes.ZkWithdrawNFT{
			FromAccountId: uint64(id),
			NFTTokenId:    uint64(tokenid),
			Amount:        amount,
		}

		action := &zksyncTypes.ZksyncAction{
			Ty: zksyncTypes.TyWithdrawNFTAction,
			Value: &zksyncTypes.ZksyncAction_WithdrawNFT{
				WithdrawNFT: param,
			},
		}

		tx, err := createChain33Tx(keys[i], getRealExecName(paraName, zksyncTypes.Zksync), action)
		if nil != err {
			fmt.Println("WithdrawNFT failed to createChain33Tx due to err:", err.Error())
			return
		}
		sendTx(rpcLaddr, tx)
	}
}

/*Package commands implement dapp client commands*/
package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func nftCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nft",
		Short: "nft related cmd",
	}
	cmd.AddCommand(mintNFTCmd())
	cmd.AddCommand(transferNFTCmd())
	cmd.AddCommand(withdrawNFTCmd())
	cmd.AddCommand(getNftByIdCmd())
	cmd.AddCommand(getNftByHashCmd())

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
	cmd.Flags().Uint64P("creatorId", "f", 0, "NFT creator id")
	cmd.MarkFlagRequired("creatorId")

	cmd.Flags().Uint64P("recipientId", "t", 0, "NFT recipient id")
	cmd.MarkFlagRequired("recipientId")

	cmd.Flags().StringP("contentHash", "c", "", "NFT content hash,must 64 hex char")
	cmd.MarkFlagRequired("contentHash")

	cmd.Flags().Uint64P("protocol", "e", 1, "NFT protocol, 1:ERC1155, 2: ERC721")
	cmd.MarkFlagRequired("protocol")

	cmd.Flags().Uint64P("amount", "a", 1, "mint amount, only for ERC1155 case")

}

func setMintNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("creatorId")
	toId, _ := cmd.Flags().GetUint64("recipientId")
	contentHash, _ := cmd.Flags().GetString("contentHash")
	protocol, _ := cmd.Flags().GetUint64("protocol")
	amount, _ := cmd.Flags().GetUint64("amount")

	if protocol == zt.ZKERC721 && amount > 1 {
		fmt.Fprintln(os.Stderr, errors.Wrapf(types.ErrInvalidParam, "NFT erc721 only allow 1 amount"))
	}

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkMintNFT{
		FromAccountId: accountId,
		RecipientId:   toId,
		ContentHash:   contentHash,
		ErcProtocol:   protocol,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     commands.GetRealExecName(paraName, zt.Zksync),
		ActionName: "MintNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
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
	cmd.Flags().Uint64P("fromId", "f", 0, "NFT from id")
	cmd.MarkFlagRequired("fromId")

	cmd.Flags().Uint64P("toId", "t", 0, "NFT to id")
	cmd.MarkFlagRequired("toId")

	cmd.Flags().Uint64P("tokenId", "i", 0, "NFT token id")
	cmd.MarkFlagRequired("tokenId")

	cmd.Flags().Uint64P("amount", "a", 1, "NFT token id")
	cmd.MarkFlagRequired("amount")

}

func transferNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("fromId")
	toId, _ := cmd.Flags().GetUint64("toId")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkTransferNFT{
		FromAccountId: accountId,
		RecipientId:   toId,
		NFTTokenId:    tokenId,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     commands.GetRealExecName(paraName, zt.Zksync),
		ActionName: "TransferNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
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
	cmd.Flags().Uint64P("fromId", "f", 0, "NFT from id")
	cmd.MarkFlagRequired("fromId")

	cmd.Flags().Uint64P("tokenId", "i", 0, "NFT token id")
	cmd.MarkFlagRequired("tokenId")

	cmd.Flags().Uint64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")

}

func withdrawNFT(cmd *cobra.Command, args []string) {
	accountId, _ := cmd.Flags().GetUint64("fromId")
	tokenId, _ := cmd.Flags().GetUint64("tokenId")
	amount, _ := cmd.Flags().GetUint64("amount")

	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nft := &zt.ZkWithdrawNFT{
		FromAccountId: accountId,
		NFTTokenId:    tokenId,
		Amount:        amount,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     commands.GetRealExecName(paraName, zt.Zksync),
		ActionName: "WithdrawNFT",
		Payload:    types.MustPBToJSON(nft),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func getNftByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get nft by id",
		Run:   getNftId,
	}
	getNftByIdFlag(cmd)
	return cmd
}

func getNftByIdFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("id", "i", 0, "nft token Id")
	cmd.MarkFlagRequired("id")
}

func getNftId(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	id, _ := cmd.Flags().GetUint64("id")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId: id,
	}

	params.FuncName = "GetNFTStatus"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkNFTTokenStatus
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getNftByHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hash",
		Short: "get nft by hash",
		Run:   getNftHash,
	}
	getNftByHashFlag(cmd)
	return cmd
}

func getNftByHashFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "nft content hash")
	cmd.MarkFlagRequired("hash")
}

func getNftHash(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	hash, _ := cmd.Flags().GetString("hash")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqString{
		Data: hash,
	}

	params.FuncName = "GetNFTId"
	params.Payload = types.MustPBToJSON(req)

	var id types.Int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &id)
	ctx.Run()
}

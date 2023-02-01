package commands

import (
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func queryProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proof",
		Short: "query proof related cmd",
	}
	cmd.AddCommand(getTxProofCmd())
	cmd.AddCommand(getTxProofByHeightCmd())
	cmd.AddCommand(getProofByHeightsCmd())
	cmd.AddCommand(getLastCommitProofCmd())
	cmd.AddCommand(getZkCommitProofCmd())
	cmd.AddCommand(getFirstRootHashCmd())
	cmd.AddCommand(getZkCommitProofListCmd())
	cmd.AddCommand(getProof2QueueInfoCmd())
	cmd.AddCommand(getExistProofCmd())
	cmd.AddCommand(getLastOnChainCommitProofCmd())
	cmd.AddCommand(buildTreeCmd())
	return cmd
}

func getProof2QueueInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proof2queue",
		Short: "get proof's l2 queue info",
		Run:   getProof2Queue,
	}
	getProof2QueueFlag(cmd)
	return cmd
}

func getProof2QueueFlag(cmd *cobra.Command) {
	cmd.Flags().Int64P("proofId", "i", 0, "proof id, id > 0")
	cmd.MarkFlagRequired("proofId")
}

func getProof2Queue(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetInt64("proofId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.Int64{Data: proofId}

	params.FuncName = "GetProofId2QueueId"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ProofId2QueueIdData
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTxProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "get tx proof",
		Run:   getTxProof,
	}
	getTxProofFlag(cmd)
	return cmd
}

func getTxProofFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("hash", "s", "", "zksync tx hash")
	cmd.MarkFlagRequired("hash")
}

func getTxProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	hash, _ := cmd.Flags().GetString("hash")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TxHash: hash,
	}

	params.FuncName = "GetProofByTxHash"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.OperationInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTxProofByHeightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "get block proofs by height",
		Run:   getTxProofByHeight,
	}
	getTxProofByHeightFlag(cmd)
	return cmd
}

func getTxProofByHeightFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("height", "g", 0, "zksync proof height")
	cmd.MarkFlagRequired("height")
}

func getTxProofByHeight(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	height, _ := cmd.Flags().GetUint64("height")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		BlockHeight: height,
	}

	params.FuncName = "GetTxProofByHeight"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getProofByHeightsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blocks",
		Short: "get proofs by height range",
		Run:   getProofByHeights,
	}
	getProofByHeightsFlag(cmd)
	return cmd
}

func getProofByHeightsFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("start", "s", 0, "start height")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Uint64P("end", "e", 0, "end height")
	cmd.MarkFlagRequired("end")

	cmd.Flags().Uint64P("index", "i", 0, "start index of block")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Uint32P("op", "o", 0, "op index of block")
	cmd.MarkFlagRequired("op")

	cmd.Flags().BoolP("detail", "d", false, "if need detail")
	cmd.MarkFlagRequired("detail")
}

func getProofByHeights(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetUint64("start")
	end, _ := cmd.Flags().GetUint64("end")
	index, _ := cmd.Flags().GetUint64("index")
	op, _ := cmd.Flags().GetUint32("op")
	detail, _ := cmd.Flags().GetBool("detail")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryProofReq{
		StartBlockHeight: start,
		EndBlockHeight:   end,
		StartIndex:       index,
		OpIndex:          op,
		NeedDetail:       detail,
	}

	params.FuncName = "GetTxProofByHeights"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryProofResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getLastCommitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last",
		Short: "get last committed proof",
		Run:   getLastCommitProof,
	}

	return cmd
}

func getLastCommitProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync

	params.FuncName = "GetLastCommitProof"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})

	var resp zt.CommitProofState
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

//
//func getLastOnChainCommitProofCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "onchain",
//		Short: "get last on chain committed proof",
//		Run:   getLastOnChainCommitProof,
//	}
//	getLastCommitProofFlag(cmd)
//	return cmd
//}
//
//func getLastOnChainCommitProof(cmd *cobra.Command, args []string) {
//	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
//	chainTitleId, _ := cmd.Flags().GetUint64("chainTitleId")
//
//	var params rpctypes.Query4Jrpc
//
//	params.Execer = zt.Zksync
//
//	params.FuncName = "GetLastOnChainProof"
//	params.Payload = types.MustPBToJSON(&zt.ZkChainTitle{ChainTitleId: chainTitleId})
//
//	var resp zt.LastOnChainProof
//	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
//	ctx.Run()
//}
//
//func getProofChainTitleListCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "chainlist",
//		Short: "get all chain list of committed proof",
//		Run:   getChainTitleList,
//	}
//	return cmd
//}
//
//func getChainTitleList(cmd *cobra.Command, args []string) {
//	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
//
//	var params rpctypes.Query4Jrpc
//
//	params.Execer = zt.Zksync
//
//	params.FuncName = "GetProofChainTitleList"
//	params.Payload = types.MustPBToJSON(&types.ReqNil{})
//
//	var resp zt.ZkChainTitleList
//	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
//	ctx.Run()
//}

func getZkCommitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get commit proof by id",
		Run:   getZkCommitProof,
	}
	getZkCommitProofFlag(cmd)
	return cmd
}

func getZkCommitProofFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("proofId", "i", 0, "commit proof id")
	cmd.MarkFlagRequired("proofId")

}

func getZkCommitProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetUint64("proofId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		ProofId: proofId,
	}

	params.FuncName = "GetCommitProofById"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.QueryProofInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getFirstRootHashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "initroot",
		Short: "get merkel tree init root, default from cfg fee",
		Run:   getFirstRootHash,
	}
	getFirstRootHashFlag(cmd)
	return cmd
}

func getFirstRootHashFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("ethAddr", "e", "", "optional eth fee addr, hex format default from config")
	cmd.Flags().StringP("chain33Addr", "c", "", "optional layer2 fee addr, hex format,default from config")
}

func getFirstRootHash(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eth, _ := cmd.Flags().GetString("ethAddr")
	chain33, _ := cmd.Flags().GetString("chain33Addr")

	var params rpctypes.Query4Jrpc
	params.Execer = zt.Zksync
	req := &types.ReqAddrs{}
	if (len(eth) == 0 && len(chain33) != 0) || (len(eth) == 0 && len(chain33) != 0) {
		fmt.Fprintln(os.Stderr, "eth or layer2 addr nil")
		return
	}
	if len(eth) > 0 && len(chain33) > 0 {
		req = &types.ReqAddrs{Addrs: []string{eth, chain33}}
	}

	params.FuncName = "GetTreeInitRoot"
	params.Payload = types.MustPBToJSON(req)

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getZkCommitProofListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plist",
		Short: "get committed proof list",
		Run:   getZkCommitProofList,
	}
	getZkCommitProofListFlag(cmd)
	return cmd
}

func getZkCommitProofListFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("proofId", "i", 0, "commit proof id")
	cmd.Flags().Uint64P("onChainProofId", "s", 0, "commit on chain proof id")

	cmd.Flags().BoolP("onChain", "o", false, "if req onChain proof by sub id")
	cmd.Flags().BoolP("latestProof", "l", false, "if req latest proof")
	cmd.Flags().Uint64P("endHeight", "e", 0, "latest proof before endHeight")
	cmd.Flags().Uint64P("chainTitleId", "n", 0, "chain  title id")
	cmd.MarkFlagRequired("chainTitleId")
}

func getZkCommitProofList(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetUint64("proofId")
	onChainProofId, _ := cmd.Flags().GetUint64("onChainProofId")
	onChain, _ := cmd.Flags().GetBool("onChain")
	latestProof, _ := cmd.Flags().GetBool("latestProof")
	end, _ := cmd.Flags().GetUint64("endHeight")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkFetchProofList{
		ProofId:         proofId,
		OnChainProofId:  onChainProofId,
		ReqOnChainProof: onChain,
		ReqLatestProof:  latestProof,
		EndHeight:       end,
	}

	params.FuncName = "GetProofList"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkCommitProof
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

//
//func getProofWitnessCmd() *cobra.Command {
//	cmd := &cobra.Command{
//		Use:   "witness",
//		Short: "get account's proof witness at current height for specific token",
//		Run:   getProofWitness,
//	}
//	getProofWitnessFlag(cmd)
//	return cmd
//}
//
//func getProofWitnessFlag(cmd *cobra.Command) {
//	cmd.Flags().Uint64P("account", "a", 0, "account id")
//	cmd.MarkFlagRequired("account")
//	cmd.Flags().Uint64P("token", "t", 0, "token id")
//	cmd.MarkFlagRequired("token")
//	cmd.Flags().Uint64P("chainTitleId", "n", 0, "chain title id")
//
//}
//
//func getProofWitness(cmd *cobra.Command, args []string) {
//	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
//	account, _ := cmd.Flags().GetUint64("account")
//	token, _ := cmd.Flags().GetUint64("token")
//	chainTitleId, _ := cmd.Flags().GetUint64("chainTitleId")
//
//	var params rpctypes.Query4Jrpc
//
//	params.Execer = zt.Zksync
//	req := &zt.ZkReqExistenceProof{
//		AccountId:    account,
//		TokenId:      token,
//		ChainTitleId: chainTitleId,
//	}
//
//	params.FuncName = "GetCurrentProof"
//	params.Payload = types.MustPBToJSON(req)
//
//	var resp zt.ZkProofWitness
//	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
//	ctx.Run()
//}
//
func getExistProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exist",
		Short: "get account's exist/exodus proof for specific token",
		Run:   getExist,
	}
	getExistFlag(cmd)
	return cmd
}

func getExistFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("account", "a", 0, "account id")
	cmd.MarkFlagRequired("account")
	cmd.Flags().Uint64P("token", "t", 0, "token id")
	cmd.MarkFlagRequired("token")
	cmd.Flags().StringP("rootHash", "r", "", "optional,target tree root hash")

}

func getExist(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	account, _ := cmd.Flags().GetUint64("account")
	token, _ := cmd.Flags().GetUint64("token")
	rootHash, _ := cmd.Flags().GetString("rootHash")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkReqExistenceProof{
		AccountId: account,
		TokenId:   token,
		RootHash:  rootHash,
	}

	params.FuncName = "GetExistenceProof"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkProofWitness
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getLastOnChainCommitProofCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "onchain",
		Short: "get last on chain committed proof",
		Run:   getLastOnChainCommitProof,
	}

	return cmd
}

func getLastOnChainCommitProof(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync

	params.FuncName = "GetLastOnChainProof"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})

	var resp zt.LastOnChainProof
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func buildTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build_tree",
		Short: "build db account tree for exodus proof,return tree roothash",
		Run:   buildTree,
	}
	buildTreeFlag(cmd)
	return cmd
}

func buildTreeFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("proofId", "i", 0, "target proof id to build tree")
	cmd.Flags().StringP("root", "r", "", "matched proof root to id")

}

func buildTree(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	proofId, _ := cmd.Flags().GetUint64("proofId")
	root, _ := cmd.Flags().GetString("root")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.CommitProofState{
		ProofId:     proofId,
		NewTreeRoot: root,
	}

	params.FuncName = "BuildHistoryAccounts"
	params.Payload = types.MustPBToJSON(req)

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

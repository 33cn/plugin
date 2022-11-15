package commands

import (
	"math/big"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query related cmd",
	}
	cmd.AddCommand(queryAccountCmd())
	cmd.AddCommand(queryProofCmd())
	cmd.AddCommand(getTokenSymbolCmd())
	cmd.AddCommand(getVerifiersCmd())
	cmd.AddCommand(getTokenFeeCmd())
	cmd.AddCommand(getEthPriorityInfoCmd())
	cmd.AddCommand(getEthLastPriorityCmd())
	return cmd
}

func queryAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "query account related cmd",
	}
	cmd.AddCommand(getAccountByIdCmd())
	cmd.AddCommand(getAccountByEthCmd())
	cmd.AddCommand(getAccountByChain33Cmd())
	cmd.AddCommand(getContractAccountCmd())
	cmd.AddCommand(getTokenBalanceCmd())
	cmd.AddCommand(getMaxAccountCmd())

	return cmd
}

func getMaxAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "max",
		Short: "get max account id",
		Run:   getMaxAccountId,
	}
	return cmd
}

func getMaxAccountId(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqNil{}

	params.FuncName = "GetMaxAccountId"
	params.Payload = types.MustPBToJSON(req)

	var resp types.Int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "id",
		Short: "get zksync account by id",
		Run:   getAccountById,
	}
	getAccountByIdFlag(cmd)
	return cmd
}

func getAccountByIdFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("accountId", "a", 0, "zksync accountId")
	cmd.MarkFlagRequired("accountId")
}

func getAccountById(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accountId, _ := cmd.Flags().GetUint64("accountId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		AccountId: accountId,
	}

	params.FuncName = "GetAccountById"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.Leaf
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByEthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accountE",
		Short: "get zksync account by ethAddress",
		Run:   getAccountByEth,
	}
	getAccountByEthFlag(cmd)
	return cmd
}

func getAccountByEthFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("ethAddress", "e", " ", "zksync account ethAddress")
	cmd.MarkFlagRequired("ethAddress")
}

func getAccountByEth(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ethAddress, _ := cmd.Flags().GetString("ethAddress")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		EthAddress: ethAddress,
	}

	params.FuncName = "GetAccountByEth"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getAccountByChain33Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accountC",
		Short: "get zksync account by chain33Addr",
		Run:   getAccountByChain33,
	}
	getAccountByChain33Flag(cmd)
	return cmd
}

func getAccountByChain33Flag(cmd *cobra.Command) {
	cmd.Flags().StringP("chain33Addr", "c", "", "zksync account chain33Addr")
	cmd.MarkFlagRequired("chain33Addr")
}

func getAccountByChain33(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		Chain33Addr: chain33Addr,
	}

	params.FuncName = "GetAccountByChain33"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getContractAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contractAccount",
		Short: "get zksync contractAccount by chain33WalletAddr and token symbol",
		Run:   getContractAccount,
	}
	getContractAccountFlag(cmd)
	return cmd
}

func getContractAccountFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("chain33Addr", "c", "", "chain33 wallet address")
	cmd.MarkFlagRequired("chain33Addr")
	cmd.Flags().StringP("token", "t", "", "token symbol")
	cmd.MarkFlagRequired("token")
}

func getContractAccount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	chain33Addr, _ := cmd.Flags().GetString("chain33Addr")
	token, _ := cmd.Flags().GetString("token")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenSymbol:       token,
		Chain33WalletAddr: chain33Addr,
	}

	params.FuncName = "GetZkContractAccount"
	params.Payload = types.MustPBToJSON(req)

	var resp types.Account
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTokenBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "get zksync tokenBalance by accountId and tokenId",
		Run:   getTokenBalance,
	}
	getTokenBalanceFlag(cmd)
	return cmd
}

func getTokenBalanceFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("accountId", "a", 0, "zksync account id")
	cmd.MarkFlagRequired("accountId")
	cmd.Flags().Uint64P("token", "t", 1, "zksync token id")
	cmd.MarkFlagRequired("token")
}

func getTokenBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	token, _ := cmd.Flags().GetUint64("token")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId:   token,
		AccountId: accountId,
	}

	params.FuncName = "GetTokenBalance"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getEthPriorityInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue_id",
		Short: "get eth deposit queue id info",
		Run:   getPriority,
	}
	getPriorityFlag(cmd)
	return cmd
}

func getPriorityFlag(cmd *cobra.Command) {
	cmd.Flags().Uint64P("priorityId", "i", 0, "eth priority id, id >= 0")
	cmd.MarkFlagRequired("priorityId")
}

func getPriority(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	priorityId, _ := cmd.Flags().GetUint64("priorityId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.EthPriorityQueueID{
		ID: new(big.Int).SetUint64(priorityId).String(),
	}

	params.FuncName = "GetPriorityOpInfo"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.OperationInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getEthLastPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last_queue_id",
		Short: "get last queue id from eth deposit",
		Run:   getLastPriority,
	}
	return cmd
}

func getLastPriority(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqNil{}

	params.FuncName = "GetLastPriorityQueueId"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.EthPriorityQueueID
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTokenFeeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fee",
		Short: "get config token fee",
		Run:   getTokenFee,
	}
	getTokenFeeFlag(cmd)
	return cmd
}

func getTokenFeeFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("action", "a", 0, "action ty,withdraw:2,transfer:3,transfer2new:4,proxyExit:5")
	cmd.MarkFlagRequired("action")
	cmd.Flags().Uint64P("token", "t", 0, "token id")
	cmd.MarkFlagRequired("token")
}

func getTokenFee(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	action, _ := cmd.Flags().GetInt32("action")
	token, _ := cmd.Flags().GetUint64("token")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkSetFee{
		ActionTy: action,
		TokenId:  token,
	}

	params.FuncName = "GetCfgTokenFee"
	params.Payload = types.MustPBToJSON(req)

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getTokenSymbolCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "symbol",
		Short: "get config token symbol",
		Run:   getTokenSymbol,
	}
	getTokenSymbolFlag(cmd)
	return cmd
}

func getTokenSymbolFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("id", "i", 0, "token id")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
}

func getTokenSymbol(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	id, _ := cmd.Flags().GetInt32("id")
	symbol, _ := cmd.Flags().GetString("symbol")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId:     uint64(id),
		TokenSymbol: symbol,
	}

	params.FuncName = "GetTokenSymbol"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkTokenSymbol
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getVerifiersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verifier",
		Short: "get verifiers",
		Run:   getVerifier,
	}
	return cmd
}

func getVerifier(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqNil{}

	params.FuncName = "GetVerifiers"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkVerifier
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

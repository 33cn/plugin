package commands

import (
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
	cmd.AddCommand(queryL2QueueCmd())
	cmd.AddCommand(queryL1PriorityCmd())

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
	cmd.Flags().Uint64P("accountId", "i", 0, "zksync accountId")
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
	cmd.Flags().Uint32P("decimal", "d", 0, "1:show with token's decimal, 0: real value")

}

func getTokenBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	accountId, _ := cmd.Flags().GetUint64("accountId")
	token, _ := cmd.Flags().GetUint64("token")
	decimal, _ := cmd.Flags().GetUint32("decimal")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId:   token,
		AccountId: accountId,
		Decimal:   decimal,
	}

	params.FuncName = "GetTokenBalance"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkQueryResp
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func queryL2QueueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l2",
		Short: "query l2 queue related cmd",
	}
	cmd.AddCommand(getL2QueueInfoCmd())
	cmd.AddCommand(getL2LastQueueIdCmd())
	cmd.AddCommand(getL2BatchQueueInfoCmd())
	cmd.AddCommand(getL2ExodusModeCmd())
	cmd.AddCommand(getL2TotalDepositCmd())

	return cmd
}

func getL2LastQueueIdCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last_queue_id",
		Short: "get l2 last queue id",
		Run:   getLastQueueId,
	}
	return cmd
}

func getLastQueueId(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqNil{}

	params.FuncName = "GetL2LastQueueId"
	params.Payload = types.MustPBToJSON(req)

	var resp types.Int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getL2QueueInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue_id",
		Short: "get l2 queue id operation info",
		Run:   getL2Queue,
	}
	getL2QueueFlag(cmd)
	return cmd
}

func getL2QueueFlag(cmd *cobra.Command) {
	cmd.Flags().Int64P("queueId", "i", 0, "l2 queue id, id >= 0")
	cmd.MarkFlagRequired("queueId")
}

func getL2Queue(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	priorityId, _ := cmd.Flags().GetInt64("queueId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.Int64{Data: priorityId}

	params.FuncName = "GetL2QueueOpInfo"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkOperation
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getL2BatchQueueInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queue_id_batch",
		Short: "batch get l2 queue id operation info",
		Run:   getL2BatchQueue,
	}
	getL2BatchQueueFlag(cmd)
	return cmd
}

func getL2BatchQueueFlag(cmd *cobra.Command) {
	cmd.Flags().Int64P("start", "s", 0, "start l2 queue id, id >= 0")
	cmd.MarkFlagRequired("start")
	cmd.Flags().Int64P("end", "e", 0, "end l2 queue id, id >= 0")
	cmd.MarkFlagRequired("end")
}

func getL2BatchQueue(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	start, _ := cmd.Flags().GetInt64("start")
	end, _ := cmd.Flags().GetInt64("end")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqBlocks{Start: start, End: end}

	params.FuncName = "GetL2BatchQueueOpInfo"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkBatchOperation
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getL2ExodusModeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exodus_mode",
		Short: "get l2 exodus mode,0:init,1:normal,2:pause,3:exodus_prepare,4:final",
		Run:   getExodusMode,
	}
	return cmd
}

func getExodusMode(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.ReqNil{}

	params.FuncName = "GetCurrentExodusMode"
	params.Payload = types.MustPBToJSON(req)

	var resp types.Int64
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func queryL1PriorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "l1",
		Short: "query l1 priority related cmd",
	}
	cmd.AddCommand(getEthPriorityInfoCmd())
	cmd.AddCommand(getEthLastPriorityCmd())

	return cmd
}

func getEthPriorityInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priority_id",
		Short: "get l1 deposit priority id info",
		Run:   getPriority,
	}
	getPriorityFlag(cmd)
	return cmd
}

func getPriorityFlag(cmd *cobra.Command) {
	cmd.Flags().Int64P("priorityId", "i", 0, "eth priority id, id >= 0")
	cmd.MarkFlagRequired("priorityId")
}

func getPriority(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	priorityId, _ := cmd.Flags().GetInt64("priorityId")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &types.Int64{Data: priorityId}

	params.FuncName = "GetPriorityOpInfo"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.ZkDepositWitnessInfo
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func getEthLastPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last_priority_id",
		Short: "get last priority id from L1 deposit",
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

	var resp zt.L1PriorityID
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

func getL2TotalDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total_deposit",
		Short: "get l2 total deposit",
		Run:   getL2TotalDeposit,
	}
	getL2TotalDepositFlag(cmd)
	return cmd
}

func getL2TotalDepositFlag(cmd *cobra.Command) {
	cmd.Flags().Int32P("tokenId", "t", 0, "token id")
	_ = cmd.MarkFlagRequired("tokenId")
	cmd.Flags().Uint32P("decimal", "d", 0, "1:show with token's decimal, 0: real value")
}

func getL2TotalDeposit(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	tokenId, _ := cmd.Flags().GetInt32("tokenId")
	decimal, _ := cmd.Flags().GetUint32("decimal")

	var params rpctypes.Query4Jrpc

	params.Execer = zt.Zksync
	req := &zt.ZkQueryReq{
		TokenId: uint64(tokenId),
		Decimal: decimal,
	}

	params.FuncName = "GetTotalDeposit"
	params.Payload = types.MustPBToJSON(req)

	var resp zt.TokenBalance
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

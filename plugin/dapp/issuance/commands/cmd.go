package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/issuance/types"
	"strconv"
)

// IssuanceCmd 斗牛游戏命令行
func IssuanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuance",
		Short: "Issuance command",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		IssuanceCreateRawTxCmd(),
		IssuanceDebtRawTxCmd(),
		IssuanceRepayRawTxCmd(),
		IssuancePriceFeedRawTxCmd(),
		IssuanceCloseRawTxCmd(),
		IssuanceManageRawTxCmd(),
		IssuanceQueryCmd(),
	)

	return cmd
}

// IssuanceCreateRawTxCmd 生成开始交易命令行
func IssuanceCreateRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a issuance",
		Run:   IssuanceCreate,
	}
	addIssuanceCreateFlags(cmd)
	return cmd
}

func addIssuanceCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Uint64P("balance", "b", 0, "balance")
	cmd.Flags().Uint64P("debtCeiling", "d", 0, "debtCeiling")
	cmd.Flags().Float32P("liquidationRatio", "l", 0, "liquidationRatio")
	cmd.Flags().Uint64P("period", "p", 0, "period")
}

func IssuanceCreate(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	balance, _ := cmd.Flags().GetUint64("balance")
	debtCeiling, _ := cmd.Flags().GetUint64("debtCeiling")
	liquidationRatio, _ := cmd.Flags().GetFloat32("liquidationRatio")
	period, _ := cmd.Flags().GetUint64("period")


	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceCreate",
		Payload:    []byte(fmt.Sprintf("{\"balance\":%d, \"debtCeiling\":%d, \"liquidationRatio\":%f, \"period\":%d,}",
			balance, debtCeiling, liquidationRatio, period)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuanceDebtRawTxCmd 生成开始交易命令行
func IssuanceDebtRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debt",
		Short: "Debt a issuance",
		Run:   IssuanceDebt,
	}
	addIssuanceDebtFlags(cmd)
	return cmd
}

func addIssuanceDebtFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("issuanceID", "g", "", "issuance ID")
	cmd.MarkFlagRequired("issuanceID")
	cmd.Flags().Uint64P("value", "v", 0, "value")
	cmd.MarkFlagRequired("value")
}

func IssuanceDebt(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")
	value, _ := cmd.Flags().GetUint64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceDebt",
		Payload:    []byte(fmt.Sprintf("{\"issuanceID\":%s,\"value\":%d}", issuanceID, value)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuanceRepayRawTxCmd 生成开始交易命令行
func IssuanceRepayRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay",
		Short: "Repay a issuance",
		Run:   IssuanceRepay,
	}
	addIssuanceRepayFlags(cmd)
	return cmd
}

func addIssuanceRepayFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("issuanceID", "g", "", "issuance ID")
	cmd.MarkFlagRequired("issuanceID")
}

func IssuanceRepay(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceRepay",
		Payload:    []byte(fmt.Sprintf("{\"issuanceID\":%s}", issuanceID)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuancePriceFeedRawTxCmd 生成开始交易命令行
func IssuancePriceFeedRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "price feed",
		Run:   IssuancePriceFeed,
	}
	addIssuancePriceFeedFlags(cmd)
	return cmd
}

func addIssuancePriceFeedFlags(cmd *cobra.Command) {
	cmd.Flags().Float32P("price", "p", 0, "price")
	cmd.MarkFlagRequired("price")
	cmd.Flags().Uint64P("volume", "v", 0, "volume")
	cmd.MarkFlagRequired("volume")
}

func IssuancePriceFeed(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	price, _ := cmd.Flags().GetFloat32("price")
	volume, _ := cmd.Flags().GetUint64("volume")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuancePriceFeed",
		Payload:    []byte(fmt.Sprintf("{[\"price\":%s],[\"volume\":%d]}", price, volume)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuanceCloseRawTxCmd 生成开始交易命令行
func IssuanceCloseRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "close a issuance",
		Run:   IssuanceClose,
	}
	addIssuanceCloseFlags(cmd)
	return cmd
}

func addIssuanceCloseFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("issuanceID", "g", "", "issuance ID")
	cmd.MarkFlagRequired("issuanceID")
}

func IssuanceClose(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceClose",
		Payload:    []byte(fmt.Sprintf("{\"issuanceID\":%s}", issuanceID)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuanceManageRawTxCmd 生成开始交易命令行
func IssuanceManageRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "manage a issuance",
		Run:   IssuanceManage,
	}
	addIssuanceManageFlags(cmd)
	return cmd
}

func addIssuanceManageFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "addr")
}

func IssuanceManage(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceManage",
		Payload:    []byte(fmt.Sprintf("{[\"addr\":%s]}", addr)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// IssuanceQueryCmd 查询命令行
func IssuanceQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query result",
		Run:   IssuanceQuery,
	}
	addIssuanceQueryFlags(cmd)
	return cmd
}

func addIssuanceQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("issuanceID", "g", "", "issuance ID")
	cmd.Flags().StringP("address", "a", "", "address")
	cmd.Flags().StringP("index", "i", "", "index")
	cmd.Flags().StringP("status", "s", "", "status")
	cmd.Flags().StringP("issuanceIDs", "d", "", "issuance IDs")
}

func IssuanceQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	// indexstr, _ := cmd.Flags().GetString("index")
	issuanceIDs, _ := cmd.Flags().GetString("issuanceIDs")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.IssuanceX
	//if indexstr != "" {
	//	index, err := strconv.ParseInt(indexstr, 10, 64)
	//	if err != nil {
	//		fmt.Println(err)
	//		cmd.Help()
	//		return
	//	}
	//	req.Index = index
	//}

	status, err := strconv.ParseInt(statusStr, 10, 32)
	if err != nil {
		fmt.Println(err)
		cmd.Help()
		return
	}

	if issuanceID != "" {
		if statusStr != "" {
			params.FuncName = "IssuanceDebtInfoByStatus"

			req := &pkt.ReqIssuanceDebtInfoByStatus{
				IssuanceId: issuanceID,
				Status: int32(status),
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceDebtInfos
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if address != "" {
			params.FuncName = "IssuanceDebtInfoByAddr"

			req := &pkt.ReqIssuanceDebtInfoByAddr{
				IssuanceId: issuanceID,
				Addr: address,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceDebtInfos
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else {
			params.FuncName = "IssuanceInfoByID"

			req := &pkt.ReqIssuanceInfo{
				IssuanceId: issuanceID,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceCurrentInfo
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		}
	} else if address != "" {
		params.FuncName = "IssuanceByAddr"

		req := &pkt.ReqIssuanceByAddr{Addr: address}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepIssuanceIDs
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if statusStr != "" {
		params.FuncName = "IssuanceByStatus"

		req := &pkt.ReqIssuanceByStatus{Status:int32(status)}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepIssuanceIDs
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if issuanceIDs != "" {
		params.FuncName = "IssuanceInfoByIDs"

		var issuanceIDsS []string
		issuanceIDsS = append(issuanceIDsS, issuanceIDs)
		issuanceIDsS = append(issuanceIDsS, issuanceIDs)
		req := &pkt.ReqIssuanceInfos{IssuanceIds: issuanceIDsS}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepIssuanceCurrentInfos
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		fmt.Println("Error: requeres at least one of gameID, address or status")
		cmd.Help()
	}
}

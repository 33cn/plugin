package commands

import (
	"fmt"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/issuance/types"
	"github.com/spf13/cobra"
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
	cmd.MarkFlagRequired("balance")
	cmd.Flags().Uint64P("debtCeiling", "d", 0, "debtCeiling")
	cmd.Flags().Float32P("liquidationRatio", "l", 0, "liquidationRatio")
	cmd.Flags().Uint64P("period", "p", 0, "period")
}

func IssuanceCreate(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	balance, _ := cmd.Flags().GetUint64("balance")
	debtCeiling, _ := cmd.Flags().GetUint64("debtCeiling")
	liquidationRatio, _ := cmd.Flags().GetFloat32("liquidationRatio")
	period, _ := cmd.Flags().GetUint64("period")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceCreate",
		Payload:    []byte(fmt.Sprintf("{\"totalBalance\":%d, \"debtCeiling\":%d, \"liquidationRatio\":%f, \"period\":%d}",
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
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")
	value, _ := cmd.Flags().GetUint64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceDebt",
		Payload:    []byte(fmt.Sprintf("{\"issuanceID\":\"%s\",\"value\":%d}", issuanceID, value)),
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
	cmd.Flags().StringP("debtID", "d", "", "debt ID")
	cmd.MarkFlagRequired("debtID")
}

func IssuanceRepay(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")
	debtID, _ := cmd.Flags().GetString("debtID")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceRepay",
		Payload:    []byte(fmt.Sprintf("{\"issuanceID\":\"%s\", \"debtID\":\"%s\"}", issuanceID, debtID)),
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
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	price, _ := cmd.Flags().GetFloat32("price")
	volume, _ := cmd.Flags().GetUint64("volume")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuancePriceFeed",
		Payload:    []byte(fmt.Sprintf("{\"price\":[ %f ], \"volume\":[ %d ]}", price, volume)),
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
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceClose",
		Payload:    []byte(fmt.Sprintf("{\"issuanceId\":\"%s\"}", issuanceID)),
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
	cmd.MarkFlagRequired("addr")
}

func IssuanceManage(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.IssuanceX),
		ActionName: "IssuanceManage",
		Payload:    []byte(fmt.Sprintf("{\"addr\":[\"%s\"]}", addr)),
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
	cmd.Flags().StringP("issuanceIDs", "e", "", "issuance IDs")
	cmd.Flags().StringP("debtID", "d", "", "debt ID")
}

func IssuanceQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	issuanceID, _ := cmd.Flags().GetString("issuanceID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	// indexstr, _ := cmd.Flags().GetString("index")
	issuanceIDs, _ := cmd.Flags().GetString("issuanceIDs")
	debtID, _ := cmd.Flags().GetString("debtID")

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

	var status int64
	var err error
	if statusStr != "" {
		status, err = strconv.ParseInt(statusStr, 10, 32)
		if err != nil {
			fmt.Println(err)
			cmd.Help()
			return
		}
	}

	if issuanceID != "" {
		if address != "" {
			params.FuncName = "IssuanceRecordsByAddr"

			req := &pkt.ReqIssuanceRecordsByAddr{
				IssuanceId: issuanceID,
				Status: int32(status),
				Addr: address,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if statusStr != "" {
			params.FuncName = "IssuanceRecordsByStatus"

			req := &pkt.ReqIssuanceRecordsByStatus{
				IssuanceId: issuanceID,
				Status: int32(status),
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if debtID != ""{
			params.FuncName = "IssuanceRecordByID"

			req := &pkt.ReqIssuanceDebtInfo{
				IssuanceId: issuanceID,
				DebtId: debtID,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepIssuanceDebtInfo
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
		fmt.Println(params.Payload)
		var res pkt.RepIssuanceCurrentInfos
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		cmd.Help()
	}
}

package commands

import (
	"fmt"
	"strconv"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
	"github.com/spf13/cobra"
)

// CollateralizeCmd 斗牛游戏命令行
func CollateralizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collateralize",
		Short: "Collateralize command",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CollateralizeCreateRawTxCmd(),
		CollateralizeBorrowRawTxCmd(),
		CollateralizeAppendRawTxCmd(),
		CollateralizeRepayRawTxCmd(),
		CollateralizePriceFeedRawTxCmd(),
		CollateralizeRetrieveRawTxCmd(),
		CollateralizeManageRawTxCmd(),
		CollateralizeQueryCmd(),
	)

	return cmd
}

// CollateralizeCreateRawTxCmd 生成开始交易命令行
func CollateralizeCreateRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a collateralize",
		Run:   CollateralizeCreate,
	}
	addCollateralizeCreateFlags(cmd)
	return cmd
}

func addCollateralizeCreateFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("balance", "b", 0, "balance")
	cmd.MarkFlagRequired("balance")
}

//CollateralizeCreate ...
func CollateralizeCreate(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	balance, _ := cmd.Flags().GetFloat64("balance")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeCreate",
		Payload:    []byte(fmt.Sprintf("{\"totalBalance\":%f}", balance)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeBorrowRawTxCmd 生成开始交易命令行
func CollateralizeBorrowRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "borrow",
		Short: "Borrow a collateralize",
		Run:   CollateralizeBorrow,
	}
	addCollateralizeBorrowFlags(cmd)
	return cmd
}

func addCollateralizeBorrowFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.MarkFlagRequired("collateralizeID")
	cmd.Flags().Float64P("value", "v", 0, "value")
	cmd.MarkFlagRequired("value")
}

//CollateralizeBorrow ...
func CollateralizeBorrow(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	value, _ := cmd.Flags().GetFloat64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeBorrow",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\",\"value\":%f}", collateralizeID, value)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeAppendRawTxCmd 生成开始交易命令行
func CollateralizeAppendRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "append",
		Short: "Append a collateralize",
		Run:   CollateralizeAppend,
	}
	addCollateralizeAppendFlags(cmd)
	return cmd
}

func addCollateralizeAppendFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.MarkFlagRequired("collateralizeID")
	cmd.Flags().StringP("recordID", "r", "", "recordID")
	cmd.MarkFlagRequired("recordID")
	cmd.Flags().Float64P("value", "v", 0, "value")
	cmd.MarkFlagRequired("value")
}

//CollateralizeAppend ...
func CollateralizeAppend(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	recordID, _ := cmd.Flags().GetString("recordID")
	value, _ := cmd.Flags().GetFloat64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeAppend",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\", \"recordID\":\"%s\", \"value\":%f}", collateralizeID, recordID, value)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeRepayRawTxCmd 生成开始交易命令行
func CollateralizeRepayRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repay",
		Short: "Repay a collateralize",
		Run:   CollateralizeRepay,
	}
	addCollateralizeRepayFlags(cmd)
	return cmd
}

func addCollateralizeRepayFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.MarkFlagRequired("collateralizeID")
	cmd.Flags().StringP("recordID", "r", "", "recordID")
	cmd.MarkFlagRequired("recordID")
}

//CollateralizeRepay ...
func CollateralizeRepay(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	recordID, _ := cmd.Flags().GetString("recordID")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeRepay",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\",\"recordID\":\"%s\"}", collateralizeID, recordID)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizePriceFeedRawTxCmd 生成开始交易命令行
func CollateralizePriceFeedRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feed",
		Short: "price feed",
		Run:   CollateralizePriceFeed,
	}
	addCollateralizePriceFeedFlags(cmd)
	return cmd
}

func addCollateralizePriceFeedFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("price", "p", 0, "price")
	cmd.MarkFlagRequired("price")
	cmd.Flags().Uint64P("volume", "v", 0, "volume")
	cmd.MarkFlagRequired("volume")
}

//CollateralizePriceFeed ...
func CollateralizePriceFeed(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	price, _ := cmd.Flags().GetFloat64("price")
	volume, _ := cmd.Flags().GetUint64("volume")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizePriceFeed",
		Payload:    []byte(fmt.Sprintf("{\"price\":[ %f ], \"volume\":[ %d ]}", price, volume)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeRetrieveRawTxCmd 生成开始交易命令行
func CollateralizeRetrieveRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retrieve",
		Short: "retrieve balance",
		Run:   CollateralizeRetrieve,
	}
	addCollateralizeRetrieveFlags(cmd)
	return cmd
}

func addCollateralizeRetrieveFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.MarkFlagRequired("collateralizeID")
	cmd.Flags().Float64P("balance", "b", 0, "retrieve balance")
	cmd.MarkFlagRequired("balance")
}

//CollateralizeRetrieve ...
func CollateralizeRetrieve(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	balance, _ := cmd.Flags().GetFloat64("balance")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeRetrieve",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\", \"balance\": %f}", collateralizeID, balance)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeManageRawTxCmd 生成开始交易命令行
func CollateralizeManageRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "manage",
		Short: "manage a collateralize",
		Run:   CollateralizeManage,
	}
	addCollateralizeManageFlags(cmd)
	return cmd
}

func addCollateralizeManageFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("debtCeiling", "d", 0, "debtCeiling")
	cmd.Flags().Float64P("liquidationRatio", "l", 0, "liquidationRatio")
	cmd.Flags().Float64P("stabilityFeeRatio", "s", 0, "stabilityFeeRatio")
	cmd.Flags().Uint64P("period", "p", 0, "period")
	cmd.Flags().Float64P("totalBalance", "t", 0, "totalBalance")
}

//CollateralizeManage ...
func CollateralizeManage(cmd *cobra.Command, args []string) {
	paraName, _ := cmd.Flags().GetString("paraName")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	debtCeiling, _ := cmd.Flags().GetFloat64("debtCeiling")
	liquidationRatio, _ := cmd.Flags().GetFloat64("liquidationRatio")
	stabilityFeeRatio, _ := cmd.Flags().GetFloat64("stabilityFeeRatio")
	period, _ := cmd.Flags().GetUint64("period")
	totalBalance, _ := cmd.Flags().GetFloat64("totalBalance")

	params := &rpctypes.CreateTxIn{
		Execer:     types.GetExecName(pkt.CollateralizeX, paraName),
		ActionName: "CollateralizeManage",
		Payload: []byte(fmt.Sprintf("{\"debtCeiling\":%f, \"liquidationRatio\":%f, \"stabilityFeeRatio\":%f, \"period\":%d, \"totalBalance\":%f}",
			debtCeiling, liquidationRatio, stabilityFeeRatio, period, totalBalance)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//CollateralizeQueryCfgCmd ...
func CollateralizeQueryCfgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Query config result",
		Run:   CollateralizeQueryConfig,
	}
	return cmd
}

//CollateralizeQueryConfig ...
func CollateralizeQueryConfig(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX

	params.FuncName = "CollateralizeConfig"
	var res pkt.RepCollateralizeConfig
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

//CollateralizeQueryPriceCmd ...
func CollateralizeQueryPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price",
		Short: "Query latest price",
		Run:   CollateralizeQueryPrice,
	}
	return cmd
}

//CollateralizeQueryPrice ...
func CollateralizeQueryPrice(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX

	params.FuncName = "CollateralizePrice"
	var res pkt.RepCollateralizePrice
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

//CollateralizeQueryUserBalanceCmd ...
func CollateralizeQueryUserBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Query user balance",
		Run:   CollateralizeQueryUserBalance,
	}
	addCollateralizeQueryBalanceFlags(cmd)
	return cmd
}

func addCollateralizeQueryBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "address")
	cmd.MarkFlagRequired("address")
}

//CollateralizeQueryUserBalance ...
func CollateralizeQueryUserBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("address")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX
	params.FuncName = "CollateralizeUserBalance"
	req := &pkt.ReqCollateralizeRecordByAddr{
		Addr: addr,
	}
	params.Payload = types.MustPBToJSON(req)

	var res pkt.RepCollateralizeUserBalance
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// CollateralizeQueryCmd 查询命令行
func CollateralizeQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query result",
		Run:   CollateralizeQuery,
	}
	addCollateralizeQueryFlags(cmd)
	cmd.AddCommand(
		CollateralizeQueryCfgCmd(),
		CollateralizeQueryPriceCmd(),
		CollateralizeQueryUserBalanceCmd(),
	)
	return cmd
}

func addCollateralizeQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.Flags().StringP("address", "a", "", "address")
	cmd.Flags().StringP("index", "i", "", "index")
	cmd.Flags().StringP("status", "s", "", "status")
	cmd.Flags().StringP("collateralizeIDs", "d", "", "collateralize IDs")
	cmd.Flags().StringP("borrowID", "b", "", "borrow ID")
}

//CollateralizeQuery ...
func CollateralizeQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	collateralizeIDs, _ := cmd.Flags().GetString("collateralizeIDs")
	borrowID, _ := cmd.Flags().GetString("borrowID")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX

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

	if collateralizeID != "" {
		if address != "" {
			params.FuncName = "CollateralizeRecordByAddr"

			req := &pkt.ReqCollateralizeRecordByAddr{
				CollateralizeId: collateralizeID,
				Status:          int32(status),
				Addr:            address,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if statusStr != "" {
			params.FuncName = "CollateralizeRecordByStatus"

			req := &pkt.ReqCollateralizeRecordByStatus{
				CollateralizeId: collateralizeID,
				Status:          int32(status),
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if borrowID != "" {
			params.FuncName = "CollateralizeRecordByID"

			req := &pkt.ReqCollateralizeRecord{
				CollateralizeId: collateralizeID,
				RecordId:        borrowID,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeRecord
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else {
			params.FuncName = "CollateralizeInfoByID"

			req := &pkt.ReqCollateralizeInfo{
				CollateralizeId: collateralizeID,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeCurrentInfo
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		}
	} else if address != "" {
		params.FuncName = "CollateralizeByAddr"

		req := &pkt.ReqCollateralizeByAddr{Addr: address}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepCollateralizeIDs
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if statusStr != "" {
		params.FuncName = "CollateralizeByStatus"

		req := &pkt.ReqCollateralizeByStatus{Status: int32(status)}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepCollateralizeIDs
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if collateralizeIDs != "" {
		params.FuncName = "CollateralizeInfoByIDs"

		var collateralizeIDsS []string
		collateralizeIDsS = append(collateralizeIDsS, collateralizeIDs)
		collateralizeIDsS = append(collateralizeIDsS, collateralizeIDs)
		req := &pkt.ReqCollateralizeInfos{CollateralizeIds: collateralizeIDsS}
		params.Payload = types.MustPBToJSON(req)
		var res pkt.RepCollateralizeCurrentInfos
		ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		fmt.Println("Error: requeres at least one of collId, address or status")
		cmd.Help()
	}
}

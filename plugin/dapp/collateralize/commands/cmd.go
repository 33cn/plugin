package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
	"strconv"
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
		CollateralizeCloseRawTxCmd(),
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
	cmd.Flags().Uint64P("balance", "b", 0, "balance")
	cmd.MarkFlagRequired("balance")
}

func CollateralizeCreate(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	balance, _ := cmd.Flags().GetUint64("balance")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizeCreate",
		Payload:    []byte(fmt.Sprintf("{\"totalBalance\":%d}", balance)),
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
	cmd.Flags().Uint64P("value", "v", 0, "value")
	cmd.MarkFlagRequired("value")
}

func CollateralizeBorrow(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	value, _ := cmd.Flags().GetUint64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizeBorrow",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\",\"value\":%d}", collateralizeID, value)),
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
	cmd.Flags().Uint64P("value", "v", 0, "value")
	cmd.MarkFlagRequired("value")
}

func CollateralizeAppend(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	recordID, _ := cmd.Flags().GetString("recordID")
	value, _ := cmd.Flags().GetUint64("value")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizeAppend",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\", \"recordID\":\"%s\", \"value\":%d}", collateralizeID, recordID, value)),
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

func CollateralizeRepay(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	recordID, _ := cmd.Flags().GetString("recordID")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
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
	cmd.Flags().Float32P("price", "p", 0, "price")
	cmd.MarkFlagRequired("price")
	cmd.Flags().Uint64P("volume", "v", 0, "volume")
	cmd.MarkFlagRequired("volume")
}

func CollateralizePriceFeed(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	price, _ := cmd.Flags().GetFloat32("price")
	volume, _ := cmd.Flags().GetUint64("volume")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizePriceFeed",
		Payload:    []byte(fmt.Sprintf("{\"price\":[ %f ], \"volume\":[ %d ]}", price, volume)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// CollateralizeCloseRawTxCmd 生成开始交易命令行
func CollateralizeCloseRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "close",
		Short: "close a collateralize",
		Run:   CollateralizeClose,
	}
	addCollateralizeCloseFlags(cmd)
	return cmd
}

func addCollateralizeCloseFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("collateralizeID", "g", "", "collateralize ID")
	cmd.MarkFlagRequired("collateralizeID")
}

func CollateralizeClose(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizeClose",
		Payload:    []byte(fmt.Sprintf("{\"collateralizeID\":\"%s\"}", collateralizeID)),
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
	cmd.Flags().Uint64P("debtCeiling", "d", 0, "debtCeiling")
	cmd.Flags().Float32P("liquidationRatio", "l", 0, "liquidationRatio")
	cmd.Flags().Float32P("stabilityFeeRatio", "s", 0, "stabilityFeeRatio")
	cmd.Flags().Uint64P("period", "p", 0, "period")
	cmd.Flags().Uint64P("totalBalance", "t", 0, "totalBalance")
}

func CollateralizeManage(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)
	if cfg == nil {
		panic(fmt.Sprintln("can not find CliSysParam title", title))
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	debtCeiling, _ := cmd.Flags().GetUint64("debtCeiling")
	liquidationRatio, _ := cmd.Flags().GetFloat32("liquidationRatio")
	stabilityFeeRatio, _ := cmd.Flags().GetFloat32("stabilityFeeRatio")
	period, _ := cmd.Flags().GetUint64("period")
	totalBalance, _ := cmd.Flags().GetUint64("totalBalance")

	params := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(pkt.CollateralizeX),
		ActionName: "CollateralizeManage",
		Payload:    []byte(fmt.Sprintf("{\"debtCeiling\":%d, \"liquidationRatio\":%f, \"stabilityFeeRatio\":%f, \"period\":%d, \"totalBalance\":%d}",
			debtCeiling, liquidationRatio, stabilityFeeRatio, period, totalBalance)),
	}

	var res string
	ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

func CollateralizeQueryCfgCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Query config result",
		Run:   CollateralizeQueryConfig,
	}
	return cmd
}

func CollateralizeQueryConfig(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX

	params.FuncName = "CollateralizeConfig"
	var res pkt.RepCollateralizeConfig
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

func CollateralizeQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	collateralizeID, _ := cmd.Flags().GetString("collateralizeID")
	address, _ := cmd.Flags().GetString("address")
	statusStr, _ := cmd.Flags().GetString("status")
	// indexstr, _ := cmd.Flags().GetString("index")
	collateralizeIDs, _ := cmd.Flags().GetString("collateralizeIDs")
	borrowID, _ := cmd.Flags().GetString("borrowID")

	var params rpctypes.Query4Jrpc
	params.Execer = pkt.CollateralizeX
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

	if collateralizeID != "" {
		 if address != "" {
			params.FuncName = "CollateralizeRecordByAddr"

			req := &pkt.ReqCollateralizeRecordByAddr{
				CollateralizeId: collateralizeID,
				Status: int32(status),
				Addr: address,
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if statusStr != "" {
			params.FuncName = "CollateralizeRecordByStatus"

			req := &pkt.ReqCollateralizeRecordByStatus{
				CollateralizeId: collateralizeID,
				Status: int32(status),
			}
			params.Payload = types.MustPBToJSON(req)
			var res pkt.RepCollateralizeRecords
			ctx := jsonrpc.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
			ctx.Run()
		} else if borrowID != ""{
			params.FuncName = "CollateralizeRecordByID"

			req := &pkt.ReqCollateralizeRecord{
				CollateralizeId: collateralizeID,
				RecordId: borrowID,
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

		req := &pkt.ReqCollateralizeByStatus{Status:int32(status)}
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

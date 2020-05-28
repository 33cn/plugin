package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/x2ethereum/ebrelayer/utils"
	types2 "github.com/33cn/plugin/plugin/dapp/x2ethereum/types"
	"github.com/spf13/cobra"
)

func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query x2ethereum",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		queryEthProphecyCmd(),
		queryValidatorsCmd(),
		queryConsensusCmd(),
		queryTotalPowerCmd(),
		querySymbolTotalAmountByTxTypeCmd(),
	)
	return cmd
}

func queryEthProphecyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prophecy",
		Short: "query prophecy",
		Run:   queryEthProphecy,
	}

	cmd.Flags().StringP("id", "i", "", "prophecy id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func queryEthProphecy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	id, _ := cmd.Flags().GetString("id")

	get := &types2.QueryEthProphecyParams{
		ID: id,
	}

	payLoad, err := types.PBToJSON(get)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}

	query := rpctypes.Query4Jrpc{
		Execer:   types2.X2ethereumX,
		FuncName: types2.FuncQueryEthProphecy,
		Payload:  payLoad,
	}

	channel := &types2.ReceiptEthProphecy{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func queryValidatorsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "query current validators",
		Run:   queryValidators,
	}
	cmd.Flags().StringP("validator", "v", "", "write if you want to check specific validator")
	//_ = cmd.MarkFlagRequired("validator")
	return cmd
}

func queryValidators(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	validator, _ := cmd.Flags().GetString("validator")

	get := &types2.QueryValidatorsParams{
		Validator: validator,
	}

	payLoad, err := types.PBToJSON(get)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}

	query := rpctypes.Query4Jrpc{
		Execer:   types2.X2ethereumX,
		FuncName: types2.FuncQueryValidators,
		Payload:  payLoad,
	}

	channel := &types2.ReceiptQueryValidator{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func queryConsensusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consensus",
		Short: "query current consensus need",
		Run:   queryConsensus,
	}
	return cmd
}

func queryConsensus(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	query := rpctypes.Query4Jrpc{
		Execer:   types2.X2ethereumX,
		FuncName: types2.FuncQueryConsensusThreshold,
	}

	channel := &types2.ReceiptQueryConsensusThreshold{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func queryTotalPowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "totalpower",
		Short: "query current total power",
		Run:   queryTotalPower,
	}
	return cmd
}

func queryTotalPower(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	query := rpctypes.Query4Jrpc{
		Execer:   types2.X2ethereumX,
		FuncName: types2.FuncQueryTotalPower,
	}

	channel := &types2.ReceiptQueryTotalPower{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

func querySymbolTotalAmountByTxTypeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lockburnasset",
		Short: "query current symbol total amount by tx type lock or withdraw",
		Run:   querySymbolTotalAmountByTxType,
	}

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	_ = cmd.MarkFlagRequired("symbol")

	cmd.Flags().Int64P("direction", "d", 0, "eth2chain33 = 1,chain33toeth = 2")
	_ = cmd.MarkFlagRequired("direction")

	cmd.Flags().Int64P("txtype", "t", 0, "lock = 1,burn = 2")
	_ = cmd.MarkFlagRequired("txtype")

	cmd.Flags().StringP("tokenaddress", "a", "", "token address,nil for all this token symbol,and if you want to query eth,you can also ignore it")
	return cmd
}

func querySymbolTotalAmountByTxType(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	direction, _ := cmd.Flags().GetInt64("direction")
	txType, _ := cmd.Flags().GetInt64("txtype")
	contract, _ := cmd.Flags().GetString("tokenaddress")
	nodeAddr, _ := cmd.Flags().GetString("node_addr")

	decimal, err := utils.GetDecimalsFromNode(contract, nodeAddr)
	if err != nil {
		fmt.Println("get decimal error")
		return
	}

	if strings.ToLower(symbol) == "eth" && contract == "" {
		contract = "0x0000000000000000000000000000000000000000"
	}

	var txTypeStr string
	if txType == 1 {
		txTypeStr = "lock"
	} else if txType == 2 {
		txTypeStr = "withdraw"
	}
	get := &types2.QuerySymbolAssetsByTxTypeParams{
		TokenSymbol: symbol,
		Direction:   direction,
		TxType:      txTypeStr,
		TokenAddr:   contract,
		Decimal:     decimal,
	}
	payLoad, err := types.PBToJSON(get)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "ErrPbToJson:"+err.Error())
		return
	}

	query := rpctypes.Query4Jrpc{
		Execer:   types2.X2ethereumX,
		FuncName: types2.FuncQuerySymbolTotalAmountByTxType,
		Payload:  payLoad,
	}

	channel := &types2.ReceiptQuerySymbolAssets{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", query, channel)
	ctx.Run()
}

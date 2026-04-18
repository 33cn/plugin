package commands

import (
	"fmt"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rgbx/types"
	"github.com/spf13/cobra"
	"os"
)

func listPendingTxCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listPend",
		Aliases: []string{"lp"},
		Short:   "list pending tx",
		Run:     listPending,
		Example: "listPend -s startHeight -i startIndex -c count",
	}
	listPendingFlags(cmd)
	return cmd
}

func listPendingFlags(cmd *cobra.Command) {

	cmd.Flags().Uint64P("startHeight", "s", 0, "list start Height")
	cmd.Flags().Uint64P("startIndex", "i", 0, "list start tx index")
	cmd.Flags().Uint32P("count", "c", 1, "list tx count")
}

func listPending(cmd *cobra.Command, _ []string) {

	startHeight, _ := cmd.Flags().GetUint64("startHeight")
	startIndex, _ := cmd.Flags().GetUint64("startIndex")
	count, _ := cmd.Flags().GetUint32("count")

	req := &rtypes.ReqListPendingTx{
		StartHeight: int64(startHeight),
		StartIndex:  int64(startIndex),
		Count:       int32(count),
	}
	reply := &rtypes.PendingTxs{}
	sendQueryRPC(cmd, "ListPendingTx", req, reply, false)
}

func getPendingTxCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getPend",
		Aliases: []string{"gp"},
		Short:   "get pending tx by height",
		Run:     getPending,
		Example: "getPend -h height -i index",
	}
	getPendingFlags(cmd)
	return cmd
}

func getPendingFlags(cmd *cobra.Command) {

	cmd.Flags().Uint64P("height", "h", 0, "block height")
	cmd.Flags().Uint64P("index", "i", 0, "tx index")
}

func getPending(cmd *cobra.Command, _ []string) {

	height, _ := cmd.Flags().GetUint64("height")
	index, _ := cmd.Flags().GetUint64("index")

	req := &rtypes.ReqGetPendingTx{
		Height: int64(height),
		Index:  int64(index),
	}
	reply := &rtypes.PendingTx{}
	sendQueryRPC(cmd, "GetPendingTx", req, reply, false)
}

func getAssetCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getAsset",
		Aliases: []string{"ga"},
		Short:   "get rgbx asset",
		Run:     getAsset,
		Example: "getAsset -s symbol",
	}
	getAssetFlags(cmd)
	return cmd
}

func getAssetFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("symbol", "s", "", "asset symbol")
	markRequired(cmd, "symbol")
}

func getAsset(cmd *cobra.Command, _ []string) {

	symbol, _ := cmd.Flags().GetString("symbol")
	if symbol == "" || len(symbol) > rtypes.MaxAssetSymbolLength {
		_, _ = fmt.Fprintf(os.Stderr, "invalid asset symbol: %s, "+
			"length must less than %d\n", symbol, rtypes.MaxAssetSymbolLength)
		return
	}

	req := &types.ReqString{
		Data: symbol,
	}
	reply := &rtypes.RgbxAsset{}
	sendQueryRPC(cmd, "GetAsset", req, reply, false)
}

func getConfirmedHeightCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getConfirmedHeight",
		Aliases: []string{"gch"},
		Short:   "get rgbx confirmed height",
		Run:     getConfirmedHeight,
		Example: "getConfirmedHeight",
	}
	return cmd
}

func getConfirmedHeight(cmd *cobra.Command, _ []string) {

	reply := &types.Int64{}
	sendQueryRPC(cmd, "GetConfirmedHeight", &types.ReqNil{}, reply, false)
}

func getCrossChainInfoCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "getCross",
		Aliases: []string{"gc"},
		Short:   "get cross-chain info",
		Run:     getCrossChainInfo,
		Example: "getCross -s BTC",
	}
	getCrossChainInfoFlags(cmd)
	return cmd
}

func getCrossChainInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "cross-chain asset symbol")
	markRequired(cmd, "symbol")
}

func getCrossChainInfo(cmd *cobra.Command, _ []string) {
	symbol, _ := cmd.Flags().GetString("symbol")
	if symbol == "" || len(symbol) > rtypes.MaxAssetSymbolLength {
		_, _ = fmt.Fprintf(os.Stderr, "invalid asset symbol: %s, length must less than %d\n", symbol, rtypes.MaxAssetSymbolLength)
		return
	}
	reply := &rtypes.CrossChainInfo{}
	sendQueryRPC(cmd, "GetCrossChainInfo", &types.ReqString{Data: symbol}, reply, false)
}

func listPendingTxByFromCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "listPendByFrom",
		Aliases: []string{"lpf"},
		Short:   "list pending tx by from address",
		Run:     listPendingByFrom,
		Example: "listPendByFrom -f 1xxxxxxxxxxxxxxxx",
	}
	listPendingTxByFromFlags(cmd)
	return cmd
}

func listPendingTxByFromFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("from", "f", "", "from address")
	markRequired(cmd, "from")
}

func listPendingByFrom(cmd *cobra.Command, _ []string) {
	from, _ := cmd.Flags().GetString("from")
	if from == "" {
		_, _ = fmt.Fprintln(os.Stderr, "from address is required")
		return
	}
	reply := &rtypes.PendingTxs{}
	sendQueryRPC(cmd, "ListPendingTxByFrom", &types.ReqString{Data: from}, reply, false)
}

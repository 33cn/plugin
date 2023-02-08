/*Package commands implement dapp client commands*/
package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

/*
 * 实现合约对应客户端
 */

var (
	tokenSymbol string
)

// Cmd evmxgo client command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evmxgo",
		Short: "evmxgo command",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		//add sub command
		CreateTokenTransferCmd(),
		CreateTokenWithdrawCmd(),
		GetTokensCreatedCmd(),
		GetTokenAssetsCmd(),
		CreateTokenTransferExecCmd(),
		CreateRawTokenMintTxCmd(),
		CreateRawTokenBurnTxCmd(),
		GetTokenLogsCmd(),
		GetTokenCmd(),
		QueryTxCmd(),
	)
	return cmd
}

// CreateTokenTransferCmd create raw transfer tx
func CreateTokenTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Create a evmxgo transfer transaction",
		Run:   createTokenTransfer,
	}
	addCreateTokenTransferFlags(cmd)
	return cmd
}

func addCreateTokenTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "receiver account address")
	cmd.MarkFlagRequired("to")
	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("note", "n", "", "transaction note info")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func createTokenTransfer(cmd *cobra.Command, args []string) {
	commands.CreateAssetTransfer(cmd, args, evmxgotypes.EvmxgoX)
}

// CreateTokenTransferExecCmd create raw transfer tx
func CreateTokenTransferExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send_exec",
		Short: "Create a evmxgo send to executor transaction",
		Run:   createTokenSendToExec,
	}
	addCreateTokenSendToExecFlags(cmd)
	return cmd
}

func addCreateTokenSendToExecFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "receiver executor address")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func createTokenSendToExec(cmd *cobra.Command, args []string) {
	commands.CreateAssetSendToExec(cmd, args, evmxgotypes.EvmxgoX)
}

// CreateTokenWithdrawCmd create raw withdraw tx
func CreateTokenWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Create a evmxgo withdraw transaction",
		Run:   createTokenWithdraw,
	}
	addCreateTokenWithdrawFlags(cmd)
	return cmd
}

func addCreateTokenWithdrawFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "execer withdrawn from")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().Float64P("amount", "a", 0, "withdraw amount")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func createTokenWithdraw(cmd *cobra.Command, args []string) {
	commands.CreateAssetWithdraw(cmd, args, evmxgotypes.EvmxgoX)
}

// GetTokensCreatedCmd get finish created tokens
func GetTokensCreatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "created",
		Short: "Get finish created tokens",
		Run:   getFinishCreatedTokens,
	}
	return cmd
}

func getFinishCreatedTokens(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	var reqtokens evmxgotypes.ReqEvmxgos
	reqtokens.QueryAll = true
	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "evmxgo")
	params.FuncName = "GetTokens"
	params.Payload = types.MustPBToJSON(&reqtokens)
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var res evmxgotypes.ReplyEvmxgos
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, createdToken := range res.Tokens {
		createdToken.Total = createdToken.Total / cfg.TokenPrecision

		//fmt.Printf("---The %dth Finish Created token is below--------------------\n", i)
		data, err := json.MarshalIndent(createdToken, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		fmt.Println(string(data))
	}
}

// GetTokenAssetsCmd get token assets
func GetTokenAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "Get evmxgo assets",
		Run:   tokenAssets,
	}
	addTokenAssetsFlags(cmd)
	return cmd
}

func addTokenAssetsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "execer name")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")
}

func tokenAssets(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	execer, _ := cmd.Flags().GetString("exec")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	execer = getRealExecName(paraName, execer)
	req := evmxgotypes.ReqAccountEvmxgoAssets{
		Address: addr,
		Execer:  execer,
	}
	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "evmxgo")
	params.FuncName = "GetAccountTokenAssets"
	params.Payload = types.MustPBToJSON(&req)

	var res evmxgotypes.ReplyAccountEvmxgoAssets
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCbExt(parseTokenAssetsRes)
	ctx.RunExt(cfg)
}

func parseTokenAssetsRes(arg ...interface{}) (interface{}, error) {
	res := arg[0].(*evmxgotypes.ReplyAccountEvmxgoAssets)
	cfg := arg[1].(*rpctypes.ChainConfigInfo)

	var result []*evmxgotypes.EvmxgoAccountResult
	for _, ta := range res.EvmxgoAssets {
		balanceResult := types.FormatAmount2FloatDisplay(ta.Account.Balance, cfg.TokenPrecision, true)
		frozenResult := types.FormatAmount2FloatDisplay(ta.Account.Frozen, cfg.TokenPrecision, true)
		tokenAccount := &evmxgotypes.EvmxgoAccountResult{
			Token:    ta.Symbol,
			Addr:     ta.Account.Addr,
			Currency: ta.Account.Currency,
			Balance:  balanceResult,
			Frozen:   frozenResult,
		}
		result = append(result, tokenAccount)
	}
	return result, nil
}

// CreateRawTokenMintTxCmd create raw token  mintage transaction
func CreateRawTokenMintTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Create a mint evmxgo transaction",
		Run:   tokenMint,
	}
	addTokenMintFlags(cmd)
	return cmd
}

func addTokenMintFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("amount", "a", 0, "amount of mintage")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Float64P("fee", "f", 0, "token transaction fee")
}

func tokenMint(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetFloat64("amount")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.amount"))
		return
	}
	params := &evmxgotypes.EvmxgoMint{
		Symbol: symbol,
		Amount: amountInt64,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.CreateRawTokenMintTx", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateRawTokenBurnTxCmd create raw token burn transaction
func CreateRawTokenBurnTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn",
		Short: "Create a burn token transaction",
		Run:   tokenBurn,
	}
	addTokenBurnFlags(cmd)
	return cmd
}

func addTokenBurnFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("amount", "a", 0, "amount of burn")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().Float64P("fee", "f", 0, "token transaction fee")
}

func tokenBurn(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	amount, _ := cmd.Flags().GetFloat64("amount")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	amountInt64, err := types.FormatFloatDisplay2Value(amount, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.amount"))
		return
	}
	params := &evmxgotypes.EvmxgoBurn{
		Symbol: symbol,
		Amount: amountInt64,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.CreateRawTokenBurnTx", params, nil)
	ctx.RunWithoutMarshal()
}

// GetTokenLogsCmd get logs of token
func GetTokenLogsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get logs of token",
		Run:   getTokenLogs,
	}
	getTokenLogsFlags(cmd)
	return cmd
}

func getTokenLogs(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	symbol, _ := cmd.Flags().GetString("symbol")

	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "evmxgo")
	params.FuncName = "GetTokenHistory"
	params.Payload = types.MustPBToJSON(&types.ReqString{Data: symbol})
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var res evmxgotypes.ReplyEvmxgoLogs
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(string(data))
}

func getTokenLogsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

// GetTokenCmd get token
func GetTokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get token info",
		Run:   getToken,
	}
	addGetTokenFlags(cmd)
	return cmd
}
func addGetTokenFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")
}

func getToken(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	symbol, _ := cmd.Flags().GetString("symbol")

	var reqtoken types.ReqString
	reqtoken.Data = symbol

	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetTokenInfo"
	params.Payload = types.MustPBToJSON(&reqtoken)
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	var res evmxgotypes.LocalEvmxgo
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	data, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(string(data))
}

// QueryTxCmd get tx by address
func QueryTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query_tx",
		Short: "Query transaction by token symbol",
		Run:   queryTx,
	}
	addQueryTxFlags(cmd)
	return cmd
}

func addQueryTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Int32P("flag", "f", 0, "transaction type(0: all txs relevant to addr, 1: addr as sender, 2: addr as receiver) (default 0)")
	cmd.Flags().Int32P("count", "c", 10, "maximum return number of transactions")
	cmd.Flags().Int32P("direction", "d", 0, "query direction from height:index(0: positive order -1:negative order) (default 0)")
	cmd.Flags().Int64P("height", "t", -1, "transaction's block height(-1: from latest txs, >=0: query from height)")
	cmd.Flags().Int64P("index", "i", 0, "query from index of tx in block height[0-100000] (default 0)")
}

func queryTx(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	addr, _ := cmd.Flags().GetString("addr")
	flag, _ := cmd.Flags().GetInt32("flag")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	height, _ := cmd.Flags().GetInt64("height")
	index, _ := cmd.Flags().GetInt64("index")
	symbol, _ := cmd.Flags().GetString("symbol")

	req := evmxgotypes.ReqEvmxgoTx{
		Symbol:    symbol,
		Addr:      addr,
		Flag:      flag,
		Count:     count,
		Direction: direction,
		Height:    height,
		Index:     index,
	}

	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "evmxgo")
	params.FuncName = "GetTxByToken"
	params.Payload = types.MustPBToJSON(&req)

	var res types.ReplyTxInfos
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/common/commands"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/spf13/cobra"
)

var (
	tokenSymbol string
)

// TokenCmd token 命令行
func TokenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Token management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CreateTokenTransferCmd(),
		CreateTokenWithdrawCmd(),
		GetTokensPreCreatedCmd(),
		GetTokensCreatedCmd(),
		GetTokenAssetsCmd(),
		GetTokenBalanceCmd(),
		CreateRawTokenPreCreateTxCmd(),
		CreateRawTokenFinishTxCmd(),
		CreateRawTokenRevokeTxCmd(),
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
		Short: "Create a token transfer transaction",
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
	commands.CreateAssetTransfer(cmd, args, tokenty.TokenX)
}

// CreateTokenTransferExecCmd create raw transfer tx
func CreateTokenTransferExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send_exec",
		Short: "Create a token send to executor transaction",
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
	commands.CreateAssetSendToExec(cmd, args, tokenty.TokenX)
}

// CreateTokenWithdrawCmd create raw withdraw tx
func CreateTokenWithdrawCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw",
		Short: "Create a token withdraw transaction",
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
	commands.CreateAssetWithdraw(cmd, args, tokenty.TokenX)
}

// GetTokensPreCreatedCmd get precreated tokens
func GetTokensPreCreatedCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precreated",
		Short: "Get precreated tokens",
		Run:   getPreCreatedTokens,
	}
	return cmd
}

func getPreCreatedTokens(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	var reqtokens tokenty.ReqTokens
	reqtokens.Status = tokenty.TokenStatusPreCreated
	reqtokens.QueryAll = true
	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetTokens"
	params.Payload = types.MustPBToJSON(&reqtokens)
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var res tokenty.ReplyTokens
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, preCreatedToken := range res.Tokens {
		preCreatedToken.Price = preCreatedToken.Price / cfg.CoinPrecision
		preCreatedToken.Total = preCreatedToken.Total / cfg.TokenPrecision

		data, err := json.MarshalIndent(preCreatedToken, "", "    ")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		fmt.Println(string(data))
	}
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

	var reqtokens tokenty.ReqTokens
	reqtokens.Status = tokenty.TokenStatusCreated
	reqtokens.QueryAll = true
	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetTokens"
	params.Payload = types.MustPBToJSON(&reqtokens)
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var res tokenty.ReplyTokens
	err = rpc.Call("Chain33.Query", params, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	for _, createdToken := range res.Tokens {
		createdToken.Price = createdToken.Price / cfg.CoinPrecision
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
		Short: "Get token assets",
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
	req := tokenty.ReqAccountTokenAssets{
		Address: addr,
		Execer:  execer,
	}
	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetAccountTokenAssets"
	params.Payload = types.MustPBToJSON(&req)

	var res tokenty.ReplyAccountTokenAssets
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCbExt(parseTokenAssetsRes)
	ctx.RunExt(cfg)
}

func parseTokenAssetsRes(arg ...interface{}) (interface{}, error) {
	res := arg[0].(*tokenty.ReplyAccountTokenAssets)
	cfg := arg[1].(*rpctypes.ChainConfigInfo)

	var result []*tokenty.TokenAccountResult
	for _, ta := range res.TokenAssets {
		balanceResult := types.FormatAmount2FloatDisplay(ta.Account.Balance, cfg.TokenPrecision, true)
		frozenResult := types.FormatAmount2FloatDisplay(ta.Account.Frozen, cfg.TokenPrecision, true)
		tokenAccount := &tokenty.TokenAccountResult{
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

// GetTokenBalanceCmd get token balance
func GetTokenBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get token balance of one or more addresses",
		Run:   tokenBalance,
	}
	addTokenBalanceFlags(cmd)
	return cmd
}

func addTokenBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&tokenSymbol, "symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("exec", "e", "", "execer name")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("address", "a", "", "account addresses, separated by space")
	cmd.MarkFlagRequired("address")
}

func tokenBalance(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("address")
	token, _ := cmd.Flags().GetString("symbol")
	execer, _ := cmd.Flags().GetString("exec")
	paraName, _ := cmd.Flags().GetString("paraName")
	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	execer = getRealExecName(paraName, execer)
	addresses := strings.Split(addr, " ")
	params := tokenty.ReqTokenBalance{
		Addresses:   addresses,
		TokenSymbol: token,
		Execer:      execer,
	}
	var res []*rpctypes.Account
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.GetTokenBalance", params, &res)
	ctx.SetResultCbExt(parseTokenBalanceRes)
	ctx.RunExt(cfg)
}

func parseTokenBalanceRes(arg ...interface{}) (interface{}, error) {
	res := arg[0].(*[]*rpctypes.Account)
	cfg := arg[1].(*rpctypes.ChainConfigInfo)

	var result []*tokenty.TokenAccountResult
	for _, one := range *res {
		balanceResult := types.FormatAmount2FloatDisplay(one.Balance, cfg.TokenPrecision, true)
		frozenResult := types.FormatAmount2FloatDisplay(one.Frozen, cfg.TokenPrecision, true)
		tokenAccount := &tokenty.TokenAccountResult{
			Token:    tokenSymbol,
			Addr:     one.Addr,
			Currency: one.Currency,
			Balance:  balanceResult,
			Frozen:   frozenResult,
		}
		result = append(result, tokenAccount)
	}
	return result, nil
}

// CreateRawTokenPreCreateTxCmd create raw token precreate transaction
func CreateRawTokenPreCreateTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "precreate",
		Short: "Create a precreated token transaction",
		Run:   tokenPrecreated,
	}
	addTokenPrecreatedFlags(cmd)
	return cmd
}

func addTokenPrecreatedFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "token name")
	cmd.MarkFlagRequired("name")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("introduction", "i", "", "token introduction")
	cmd.MarkFlagRequired("introduction")

	cmd.Flags().StringP("owner_addr", "a", "", "address of token owner")
	cmd.MarkFlagRequired("owner_addr")

	cmd.Flags().Float64P("price", "p", 0, "token price(mini: 0.0001)")
	cmd.MarkFlagRequired("price")

	cmd.Flags().Int64P("total", "t", 0, "total amount of the token")
	cmd.MarkFlagRequired("total")

	cmd.Flags().Int32P("category", "c", 0, "token category")

	cmd.Flags().Float64P("fee", "f", 0, "token transaction fee")
}

func tokenPrecreated(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	symbol, _ := cmd.Flags().GetString("symbol")
	introduction, _ := cmd.Flags().GetString("introduction")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")
	price, _ := cmd.Flags().GetFloat64("price")
	total, _ := cmd.Flags().GetInt64("total")
	category, _ := cmd.Flags().GetInt32("category")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	priceInt64, err := types.FormatFloatDisplay2Value(price, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value"))
		return
	}
	params := &tokenty.TokenPreCreate{
		Price:        priceInt64,
		Name:         name,
		Symbol:       symbol,
		Introduction: introduction,
		Owner:        ownerAddr,
		Total:        total * cfg.TokenPrecision,
		Category:     category,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.CreateRawTokenPreCreateTx", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateRawTokenFinishTxCmd create raw token finish create transaction
func CreateRawTokenFinishTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finish",
		Short: "Create a finish created token transaction",
		Run:   tokenFinish,
	}
	addTokenFinishFlags(cmd)
	return cmd
}

func addTokenFinishFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner_addr", "a", "", "address of token owner")
	cmd.MarkFlagRequired("owner_addr")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("fee", "f", 0, "token transaction fee")
}

func tokenFinish(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")
	params := &tokenty.TokenFinishCreate{
		Symbol: symbol,
		Owner:  ownerAddr,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.CreateRawTokenFinishTx", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateRawTokenRevokeTxCmd create raw token revoke transaction
func CreateRawTokenRevokeTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Create a token revoke transaction",
		Run:   tokenRevoke,
	}
	addTokenRevokeFlags(cmd)
	return cmd
}

func addTokenRevokeFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("owner_addr", "a", "", "address of token owner")
	cmd.MarkFlagRequired("owner_addr")

	cmd.Flags().StringP("symbol", "s", "", "token symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("fee", "f", 0, "token transaction fee")
}

func tokenRevoke(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	symbol, _ := cmd.Flags().GetString("symbol")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")

	params := &tokenty.TokenRevokeCreate{
		Symbol: symbol,
		Owner:  ownerAddr,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "token.CreateRawTokenRevokeTx", params, nil)
	ctx.RunWithoutMarshal()
}

// CreateRawTokenMintTxCmd create raw token  mintage transaction
func CreateRawTokenMintTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint",
		Short: "Create a mint token transaction",
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
	params := &tokenty.TokenMint{
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
	params := &tokenty.TokenBurn{
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
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetTokenHistory"
	params.Payload = types.MustPBToJSON(&types.ReqString{Data: symbol})
	rpc, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	var res tokenty.ReplyTokenLogs
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

	var res tokenty.LocalToken
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

	req := tokenty.ReqTokenTx{
		Symbol:    symbol,
		Addr:      addr,
		Flag:      flag,
		Count:     count,
		Direction: direction,
		Height:    height,
		Index:     index,
	}

	var params rpctypes.Query4Jrpc
	params.Execer = getRealExecName(paraName, "token")
	params.FuncName = "GetTxByToken"
	params.Payload = types.MustPBToJSON(&req)

	var res types.ReplyTxInfos
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.Run()
}

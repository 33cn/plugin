// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/spf13/cobra"
)

//MultiSigCmd :
func MultiSigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "multisig",
		Short: "multisig management",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		MultiSigAccountCmd(),
		MultiSigOwnerCmd(),
		MultiSigTxCmd(),
	)
	return cmd
}

//MultiSigAccountCmd :account相关的命令
func MultiSigAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "multisig account",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateMultiSigAccCreateCmd(),
		CreateMultiSigAccWeightModifyCmd(),
		CreateMultiSigAccDailyLimitModifyCmd(),
		GetMultiSigAccCountCmd(),
		GetMultiSigAccountsCmd(),
		GetMultiSigAccountInfoCmd(),
		GetMultiSigAccUnSpentTodayCmd(),
		GetMultiSigAccAssetsCmd(),
		GetMultiSigAccAllAddressCmd(),
		GetMultiSigAccByOwnerCmd(),
	)
	return cmd
}

//MultiSigOwnerCmd : owner相关的命令
func MultiSigOwnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owner",
		Short: "multisig owner",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateMultiSigAccOwnerAddCmd(),
		CreateMultiSigAccOwnerDelCmd(),
		CreateMultiSigAccOwnerModifyCmd(),
		CreateMultiSigAccOwnerReplaceCmd(),
	)
	return cmd
}

//MultiSigTxCmd : tx交易相关的命令
func MultiSigTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "multisig tx",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		CreateMultiSigConfirmTxCmd(),
		CreateMultiSigAccTransferInCmd(),
		CreateMultiSigAccTransferOutCmd(),
		GetMultiSigAccTxCountCmd(),
		GetMultiSigTxidsCmd(),
		GetMultiSigTxInfoCmd(),
		GetMultiSigTxConfirmedWeightCmd(),
	)
	return cmd
}

// CreateMultiSigAccCreateCmd create raw MultiSigAccCreate transaction
func CreateMultiSigAccCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a multisig account transaction",
		Run:   createMultiSigAccTransfer,
	}
	createMultiSigAccTransferFlags(cmd)
	return cmd
}

func createMultiSigAccTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("owners_addr", "a", "", "address of owners,separated by '-', addr0-addr1-addr2...")
	cmd.MarkFlagRequired("owners_addr")

	cmd.Flags().StringP("owners_weight", "w", "", "weight of owners,separated by '-', w0-w1-w2..., uint64 type")
	cmd.MarkFlagRequired("owners_weight")

	cmd.Flags().Uint64P("required_weight", "r", 0, "required weight of account execute tx")
	cmd.MarkFlagRequired("required_weight")

	cmd.Flags().StringP("execer", "e", "", "assets execer name")
	cmd.MarkFlagRequired("execer")

	cmd.Flags().StringP("symbol", "s", "", "assets symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("daily_limit", "d", 0, "daily_limit of assets ")
	cmd.MarkFlagRequired("daily_limit")
}

func createMultiSigAccTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	address, _ := cmd.Flags().GetString("owners_addr")
	addressArr := strings.Split(address, "-")

	weightstr, _ := cmd.Flags().GetString("owners_weight")
	weightsArr := strings.Split(weightstr, "-")

	//校验owner和权重数量要一致
	if len(addressArr) != len(weightsArr) {
		fmt.Fprintln(os.Stderr, "len of owners_addr mismatch len of owners_weight")
		return
	}

	//将字符转权重转换成uint64的值
	var weights []uint64
	var totalweight uint64
	var ownerCount int
	for _, weight := range weightsArr {
		ownerweight, err := strconv.ParseInt(weight, 10, 64)
		if err != nil || ownerweight <= 0 {
			fmt.Fprintln(os.Stderr, "weight invalid")
			return
		}
		weights = append(weights, uint64(ownerweight))
		totalweight += uint64(ownerweight)
		ownerCount = ownerCount + 1
	}
	var owners []*mty.Owner
	for index, addr := range addressArr {
		if addr != "" {
			owmer := &mty.Owner{OwnerAddr: addr, Weight: weights[index]}
			owners = append(owners, owmer)
		}
	}

	requiredweight, err := cmd.Flags().GetUint64("required_weight")
	if err != nil || requiredweight == 0 {
		fmt.Fprintln(os.Stderr, "required weight invalid")
		return
	}
	if requiredweight > totalweight {
		fmt.Fprintln(os.Stderr, "Requiredweight more than totalweight")
		return
	}

	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")

	dailylimit, _ := cmd.Flags().GetFloat64("daily_limit")
	err = isValidDailylimit(dailylimit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	symboldailylimit := &mty.SymbolDailyLimit{
		Symbol:     symbol,
		Execer:     execer,
		DailyLimit: uint64(math.Trunc((dailylimit+0.0000001)*1e4)) * 1e4,
	}

	params := &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: requiredweight,
		DailyLimit:     symboldailylimit,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAccCreateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccOwnerAddCmd create raw MultiSigAccOwnerAdd transaction
func CreateMultiSigAccOwnerAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Create a add owner  transaction",
		Run:   createOwnerAddTransfer,
	}
	createOwnerAddTransferFlags(cmd)
	return cmd
}

func createOwnerAddTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")

	cmd.Flags().StringP("owner_addr", "o", "", "address of owner")
	cmd.MarkFlagRequired("owner_addr")

	cmd.Flags().Uint64P("owner_weight", "w", 0, "weight of owner")
	cmd.MarkFlagRequired("owner_weight")

}

func createOwnerAddTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")
	ownerWeight, _ := cmd.Flags().GetUint64("owner_weight")

	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		NewOwner:        ownerAddr,
		NewWeight:       ownerWeight,
		OperateFlag:     mty.OwnerAdd,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigOwnerOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccOwnerDelCmd create raw MultiSigAccOwnerDel transaction
func CreateMultiSigAccOwnerDelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del",
		Short: "Create a del owner transaction",
		Run:   createOwnerDelTransfer,
	}
	createOwnerDelTransferFlags(cmd)
	return cmd
}

func createOwnerDelTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")

	cmd.Flags().StringP("owner_addr", "o", "", "address of owner")
	cmd.MarkFlagRequired("owner_addr")
}

func createOwnerDelTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")

	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        ownerAddr,
		OperateFlag:     mty.OwnerDel,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigOwnerOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccOwnerModifyCmd create raw MultiSigAccOwnerDel transaction
func CreateMultiSigAccOwnerModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "modify",
		Short: "Create a modify owner weight transaction",
		Run:   createOwnerModifyTransfer,
	}
	createOwnerModifyTransferFlags(cmd)
	return cmd
}

func createOwnerModifyTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")
	cmd.Flags().StringP("owner_addr", "o", "", "address of owner")
	cmd.MarkFlagRequired("owner_addr")
	cmd.Flags().Uint64P("owner_weight", "w", 0, "new weight of owner")
	cmd.MarkFlagRequired("owner_weight")
}

func createOwnerModifyTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")
	ownerWeight, _ := cmd.Flags().GetUint64("owner_weight")

	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        ownerAddr,
		NewWeight:       ownerWeight,
		OperateFlag:     mty.OwnerModify,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigOwnerOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccOwnerReplaceCmd create raw MultiSigAccOwnerReplace transaction
func CreateMultiSigAccOwnerReplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replace",
		Short: "Create a replace owner transaction",
		Run:   createOwnerReplaceTransfer,
	}
	createOwnerReplaceTransferFlags(cmd)
	return cmd
}

func createOwnerReplaceTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")
	cmd.Flags().StringP("owner_addr", "o", "", "address of old owner")
	cmd.MarkFlagRequired("owner_addr")
	cmd.Flags().StringP("new_owner", "n", "", "address of new owner")
	cmd.MarkFlagRequired("new_owner")
}

func createOwnerReplaceTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	ownerAddr, _ := cmd.Flags().GetString("owner_addr")
	newOwner, _ := cmd.Flags().GetString("new_owner")

	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        ownerAddr,
		NewOwner:        newOwner,
		OperateFlag:     mty.OwnerReplace,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigOwnerOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccWeightModifyCmd create raw CreateMultiSigAccWeightModifyCmd transaction
func CreateMultiSigAccWeightModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weight",
		Short: "Create a modify required weight transaction",
		Run:   createMultiSigAccWeightModifyTransfer,
	}
	createMultiSigAccWeightModifyTransferFlags(cmd)
	return cmd
}

func createMultiSigAccWeightModifyTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")
	cmd.Flags().Uint64P("weight", "w", 0, "new required weight of multisig account ")
}

func createMultiSigAccWeightModifyTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	weight, _ := cmd.Flags().GetUint64("weight")

	params := &mty.MultiSigAccOperate{
		MultiSigAccAddr:   multiSigAddr,
		NewRequiredWeight: weight,
		OperateFlag:       mty.AccWeightOp,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAccOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccDailyLimitModifyCmd create raw MultiSigAccDailyLimitModify transaction
func CreateMultiSigAccDailyLimitModifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dailylimit",
		Short: "Create a modify assets dailylimit transaction",
		Run:   createMultiSigAccDailyLimitModifyTransfer,
	}
	createMultiSigAccDailyLimitModifyTransferFlags(cmd)
	return cmd
}

func createMultiSigAccDailyLimitModifyTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")

	cmd.Flags().StringP("execer", "e", "", "assets execer name")
	cmd.MarkFlagRequired("execer")

	cmd.Flags().StringP("symbol", "s", "", "assets symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().Float64P("daily_limit", "d", 0, "daily_limit of assets ")
	cmd.MarkFlagRequired("daily_limit")
}

func createMultiSigAccDailyLimitModifyTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")
	dailylimit, _ := cmd.Flags().GetFloat64("daily_limit")

	err := isValidDailylimit(dailylimit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	assetsDailyLimit := &mty.SymbolDailyLimit{
		Symbol:     symbol,
		Execer:     execer,
		DailyLimit: uint64(math.Trunc((dailylimit+0.0000001)*1e4)) * 1e4,
	}
	params := &mty.MultiSigAccOperate{
		MultiSigAccAddr: multiSigAddr,
		DailyLimit:      assetsDailyLimit,
		OperateFlag:     mty.AccDailyLimitOp,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAccOperateTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigConfirmTxCmd create raw MultiSigConfirmTxCmd transaction
func CreateMultiSigConfirmTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirm",
		Short: "Create a confirm transaction",
		Run:   createMultiSigConfirmTransfer,
	}
	createMultiSigConfirmTransferFlags(cmd)
	return cmd
}

func createMultiSigConfirmTransferFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("multisig_addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("multisig_addr")

	cmd.Flags().Uint64P("txid", "i", 0, "txid of  multisig transaction")
	cmd.MarkFlagRequired("txid")

	cmd.Flags().StringP("confirm_or_revoke", "c", "t", "whether confirm or revoke tx (0/f/false for No; 1/t/true for Yes)")

}

func createMultiSigConfirmTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	multiSigAddr, _ := cmd.Flags().GetString("multisig_addr")
	txid, _ := cmd.Flags().GetUint64("txid")

	confirmOrRevoke, _ := cmd.Flags().GetString("confirm_or_revoke")
	confirmOrRevokeBool, err := strconv.ParseBool(confirmOrRevoke)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	params := &mty.MultiSigConfirmTx{
		MultiSigAccAddr: multiSigAddr,
		TxId:            txid,
		ConfirmOrRevoke: confirmOrRevokeBool,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigConfirmTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccTransferInCmd create raw MultiSigAccTransferInCmd transaction
func CreateMultiSigAccTransferInCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_in",
		Short: "Create a transfer to multisig account transaction",
		Run:   createMultiSigAccTransferIn,
	}
	createMultiSigAccTransferInFlags(cmd)
	return cmd
}

func createMultiSigAccTransferInFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("to", "t", "", "address of multisig account")
	cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("execer", "e", "", "assets  execer")
	cmd.MarkFlagRequired("execer")

	cmd.Flags().StringP("symbol", "s", "", "assets symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")
}

func createMultiSigAccTransferIn(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	to, _ := cmd.Flags().GetString("to")
	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")
	note, _ := cmd.Flags().GetString("note")
	amount, _ := cmd.Flags().GetFloat64("amount")

	if float64(types.MaxCoin/types.Coin) < amount {
		fmt.Fprintln(os.Stderr, types.ErrAmount)
		return
	}
	params := &mty.MultiSigExecTransferTo{
		Symbol:   symbol,
		Amount:   int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4,
		Note:     note,
		Execname: execer,
		To:       to,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAccTransferInTx", params, &res)
	ctx.RunWithoutMarshal()
}

// CreateMultiSigAccTransferOutCmd create raw MultiSigAccTransferOut transaction
func CreateMultiSigAccTransferOutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer_out",
		Short: "Create a transfer from multisig account transaction",
		Run:   createMultiSigAccTransferOut,
	}
	createMultiSigAccTransferOutFlags(cmd)
	return cmd
}

func createMultiSigAccTransferOutFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("from", "f", "", "address of multisig account")
	cmd.MarkFlagRequired("from")

	cmd.Flags().StringP("to", "t", "", "address of account")
	cmd.MarkFlagRequired("to")

	cmd.Flags().StringP("execer", "e", "", "assets execer")
	cmd.MarkFlagRequired("execer")

	cmd.Flags().StringP("symbol", "s", "", "assets symbol")
	cmd.MarkFlagRequired("symbol")

	cmd.Flags().StringP("note", "n", "", "transaction note info")

	cmd.Flags().Float64P("amount", "a", 0, "transaction amount")
	cmd.MarkFlagRequired("amount")
}

func createMultiSigAccTransferOut(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")
	note, _ := cmd.Flags().GetString("note")
	amount, _ := cmd.Flags().GetFloat64("amount")

	if float64(types.MaxCoin/types.Coin) < amount {
		fmt.Fprintln(os.Stderr, types.ErrAmount)
		return
	}
	params := &mty.MultiSigExecTransferFrom{
		Symbol:   symbol,
		Amount:   int64(math.Trunc((amount+0.0000001)*1e4)) * 1e4,
		Note:     note,
		Execname: execer,
		From:     from,
		To:       to,
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAccTransferOutTx", params, &res)
	ctx.RunWithoutMarshal()
}

//GetMultiSigAccCountCmd 获取已经创建的多重签名账户数量
func GetMultiSigAccCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "get multisig account count",
		Run:   getMultiSigAccCount,
	}
	return cmd
}

func getMultiSigAccCount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var params rpctypes.Query4Jrpc

	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccCount"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})
	rep = &types.Int64{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigAccountsCmd 获取已经创建的多重签名账户地址，通过转入的index
func GetMultiSigAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "get multisig account address",
		Run:   getMultiSigAccounts,
	}
	addgetMultiSigAccountsFlags(cmd)
	return cmd
}

func addgetMultiSigAccountsFlags(cmd *cobra.Command) {

	cmd.Flags().Int64P("start", "s", 0, "account start index")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Int64P("end", "e", 0, "account end index")
	cmd.MarkFlagRequired("end")
}

func getMultiSigAccounts(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	start, _ := cmd.Flags().GetInt64("start")
	end, _ := cmd.Flags().GetInt64("end")

	if start > end || start < 0 {
		fmt.Fprintln(os.Stderr, "input parameter invalid!")
		return
	}
	req := mty.ReqMultiSigAccs{
		Start: start,
		End:   end,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccounts"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.ReplyMultiSigAccs{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigAccountInfoCmd 获取已经创建的多重签名账户信息
func GetMultiSigAccountInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "get multisig account info",
		Run:   getMultiSigAccountInfo,
	}
	getMultiSigAccountInfoFlags(cmd)
	return cmd
}

func getMultiSigAccountInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")
}

func getMultiSigAccountInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: addr,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.MultiSig{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.SetResultCb(parseAccInfo)
	ctx.Run()
}

func parseAccInfo(view interface{}) (interface{}, error) {
	res := view.(*mty.MultiSig)
	var dailyLimitResults []*mty.DailyLimitResult

	for _, dailyLimit := range res.DailyLimits {
		dailyLimt := strconv.FormatFloat(float64(dailyLimit.DailyLimit)/float64(types.Coin), 'f', 4, 64)
		spentToday := strconv.FormatFloat(float64(dailyLimit.SpentToday)/float64(types.Coin), 'f', 4, 64)

		dailyLimitResult := &mty.DailyLimitResult{
			Symbol:     dailyLimit.Symbol,
			Execer:     dailyLimit.Execer,
			DailyLimit: dailyLimt,
			SpentToday: spentToday,
			LastDay:    time.Unix(dailyLimit.LastDay, 0).Format("2006-01-02 15:04:05"),
		}
		dailyLimitResults = append(dailyLimitResults, dailyLimitResult)
	}

	result := &mty.MultiSigResult{
		CreateAddr:     res.CreateAddr,
		MultiSigAddr:   res.MultiSigAddr,
		Owners:         res.Owners,
		DailyLimits:    dailyLimitResults,
		TxCount:        res.TxCount,
		RequiredWeight: res.RequiredWeight,
	}

	return result, nil
}

//GetMultiSigAccTxCountCmd 获取多重签名账户上的tx交易数量
func GetMultiSigAccTxCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "count",
		Short: "get multisig tx count",
		Run:   getMultiSigAccTxCount,
	}
	getMultiSigAccTxCountFlags(cmd)
	return cmd
}

func getMultiSigAccTxCountFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")
}

func getMultiSigAccTxCount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: addr,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccTxCount"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.Uint64{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigTxidsCmd 获取多重签名账户上的tx交易数量
func GetMultiSigTxidsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "txids",
		Short: "get multisig txids",
		Run:   getMultiSigTxids,
	}
	getMultiSigTxidsCmdFlags(cmd)
	return cmd
}

func getMultiSigTxidsCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().Uint64P("start", "s", 0, "tx start index")
	cmd.MarkFlagRequired("start")

	cmd.Flags().Uint64P("end", "e", 0, "tx end index")
	cmd.MarkFlagRequired("end")

	cmd.Flags().StringP("pending", "p", "t", "whether pending tx (0/f/false for No; 1/t/true for Yes)")

	cmd.Flags().StringP("executed", "x", "t", "whether executed tx (0/f/false for No; 1/t/true for Yes)")

}

func getMultiSigTxids(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")

	start, _ := cmd.Flags().GetUint64("start")
	end, _ := cmd.Flags().GetUint64("end")
	if start > end || start < 0 {
		fmt.Fprintln(os.Stderr, "input parameter invalid!")
		return
	}

	pending, _ := cmd.Flags().GetString("pending")
	pendingBool, err := strconv.ParseBool(pending)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	executed, _ := cmd.Flags().GetString("executed")
	executedBool, err := strconv.ParseBool(executed)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	req := mty.ReqMultiSigTxids{
		MultiSigAddr: addr,
		FromTxId:     start,
		ToTxId:       end,
		Pending:      pendingBool,
		Executed:     executedBool,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxids"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.ReplyMultiSigTxids{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigTxInfoCmd 获取已经创建的多重签名账户的交易信息
func GetMultiSigTxInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "get multisig account tx info",
		Run:   getMultiSigTxInfo,
	}
	getMultiSigTxInfoFlags(cmd)
	return cmd
}

func getMultiSigTxInfoFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().Uint64P("txid", "i", 0, "txid of  multisig transaction")
	cmd.MarkFlagRequired("txid")
}

func getMultiSigTxInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	txid, _ := cmd.Flags().GetUint64("txid")

	req := mty.ReqMultiSigTxInfo{
		MultiSigAddr: addr,
		TxId:         txid,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.MultiSigTx{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigTxConfirmedWeightCmd 获取交易已经被确认的总权重
func GetMultiSigTxConfirmedWeightCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "confirmed_weight",
		Short: "get the weight of the transaction confirmed.",
		Run:   getGetMultiSigTxConfirmedWeight,
	}
	getMultiSigTxConfirmedWeightFlags(cmd)
	return cmd
}

func getMultiSigTxConfirmedWeightFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().Uint64P("txid", "i", 0, "txid of  multisig transaction")
	cmd.MarkFlagRequired("txid")
}

func getGetMultiSigTxConfirmedWeight(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	txid, _ := cmd.Flags().GetUint64("txid")

	req := mty.ReqMultiSigTxInfo{
		MultiSigAddr: addr,
		TxId:         txid,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxConfirmedWeight"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.Uint64{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

//GetMultiSigAccUnSpentTodayCmd 获取多重签名账户今日免多重签名的余额
func GetMultiSigAccUnSpentTodayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unspent",
		Short: "get assets unspent today amount",
		Run:   getMultiSigAccUnSpentToday,
	}
	getMultiSigAccUnSpentTodayFlags(cmd)
	return cmd
}

func getMultiSigAccUnSpentTodayFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("execer", "e", "", "assets execer name")
	cmd.Flags().StringP("symbol", "s", "", "assets symbol")
}

func getMultiSigAccUnSpentToday(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")

	isallBool := true
	assets := &mty.Assets{}
	//获取指定资产信息时，execer和symbol不能为空
	if len(execer) != 0 && len(symbol) != 0 {
		err := mty.IsAssetsInvalid(execer, symbol)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		assets.Execer = execer
		assets.Symbol = symbol
		isallBool = false
	}

	req := mty.ReqAccAssets{
		MultiSigAddr: addr,
		Assets:       assets,
		IsAll:        isallBool,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccUnSpentToday"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.ReplyUnSpentAssets{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.SetResultCb(parseUnSpentToday)
	ctx.Run()
}

func parseUnSpentToday(view interface{}) (interface{}, error) {
	res := view.(*mty.ReplyUnSpentAssets)
	var result []*mty.UnSpentAssetsResult

	for _, unSpentAssets := range res.UnSpentAssets {
		amountResult := strconv.FormatFloat(float64(unSpentAssets.Amount)/float64(types.Coin), 'f', 4, 64)
		unSpentAssetsResult := &mty.UnSpentAssetsResult{
			Execer:  unSpentAssets.Assets.Execer,
			Symbol:  unSpentAssets.Assets.Symbol,
			UnSpent: amountResult,
		}
		result = append(result, unSpentAssetsResult)
	}
	return result, nil
}

//GetMultiSigAccAssetsCmd 获取多重签名账户上的资产信息
func GetMultiSigAccAssetsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assets",
		Short: "get assets of multisig account",
		Run:   getMultiSigAccAssets,
	}
	getMultiSigAccAssetsFlags(cmd)
	return cmd
}

func getMultiSigAccAssetsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of multisig account")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("execer", "e", "coins", "assets execer name ")
	cmd.Flags().StringP("symbol", "s", "BTY", "assets symbol")
}

func getMultiSigAccAssets(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	execer, _ := cmd.Flags().GetString("execer")
	symbol, _ := cmd.Flags().GetString("symbol")

	isallBool := true
	assets := &mty.Assets{}
	//获取指定资产信息时，execer和symbol不能为空
	if len(execer) != 0 && len(symbol) != 0 {
		err := mty.IsAssetsInvalid(execer, symbol)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		assets.Execer = execer
		assets.Symbol = symbol
		isallBool = false
	}

	req := mty.ReqAccAssets{
		MultiSigAddr: addr,
		Assets:       assets,
		IsAll:        isallBool,
	}

	var params rpctypes.Query4Jrpc
	var rep interface{}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.ReplyAccAssets{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.SetResultCb(parseAccAssets)
	ctx.Run()
}

func parseAccAssets(view interface{}) (interface{}, error) {
	res := view.(*mty.ReplyAccAssets)
	var result []*mty.AccAssetsResult

	for _, accAssets := range res.AccAssets {
		balanceResult := strconv.FormatFloat(float64(accAssets.Account.Balance)/float64(types.Coin), 'f', 4, 64)
		frozenResult := strconv.FormatFloat(float64(accAssets.Account.Frozen)/float64(types.Coin), 'f', 4, 64)
		receiverResult := strconv.FormatFloat(float64(accAssets.RecvAmount)/float64(types.Coin), 'f', 4, 64)

		accAssetsResult := &mty.AccAssetsResult{
			Execer:   accAssets.Assets.Execer,
			Symbol:   accAssets.Assets.Symbol,
			Addr:     accAssets.Account.Addr,
			Currency: accAssets.Account.Currency,
			Balance:  balanceResult,
			Frozen:   frozenResult,
			Receiver: receiverResult,
		}
		result = append(result, accAssetsResult)
	}
	return result, nil
}

//GetMultiSigAccAllAddressCmd 获取指定地址创建的所有多重签名账户
func GetMultiSigAccAllAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "creator",
		Short: "get all multisig accounts created by the address",
		Run:   getMultiSigAccAllAddress,
	}
	getMultiSigAccAllAddressFlags(cmd)
	return cmd
}

func getMultiSigAccAllAddressFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of created multisig account")
	cmd.MarkFlagRequired("addr")
}

func getMultiSigAccAllAddress(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	createAddr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc
	var rep interface{}
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: createAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAllAddress"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.AccAddress{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

func isValidDailylimit(dailylimit float64) error {
	if dailylimit < 0 || float64(types.MaxCoin/types.Coin) < dailylimit {
		return mty.ErrInvalidDailyLimit
	}
	return nil
}

//GetMultiSigAccByOwnerCmd 获取指定地址拥有的所有多重签名账户
func GetMultiSigAccByOwnerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owner",
		Short: "get multisig accounts by the owner",
		Run:   getMultiSigAccByOwner,
	}
	getMultiSigAccByOwnerFlags(cmd)
	return cmd
}

func getMultiSigAccByOwnerFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "address of owner")
}

func getMultiSigAccByOwner(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ownerAddr, _ := cmd.Flags().GetString("addr")

	params := &types.ReqString{
		Data: ownerAddr,
	}
	var res mty.OwnerAttrs
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "multisig.MultiSigAddresList", params, &res)
	ctx.Run()
}

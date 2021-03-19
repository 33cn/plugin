// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/spf13/cobra"
)

var (
	defMixCount int32 = 16
)

// PrivacyCmd 添加隐私交易的命令
func PrivacyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "privacy",
		Short: "Privacy transaction management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		showPrivacyKeyCmd(),
		showPrivacyAccountSpendCmd(),
		createPub2PrivTxCmd(),
		createPriv2PrivTxCmd(),
		createPriv2PubTxCmd(),
		showAmountsOfUTXOCmd(),
		showUTXOs4SpecifiedAmountCmd(),
		showPrivacyAccountInfoCmd(),
		listPrivacyTxsCmd(),
		rescanUtxosOptCmd(),
		enablePrivacyCmd(),
	)

	return cmd
}

// showPrivacyKeyCmd show privacy key by address
func showPrivacyKeyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpk",
		Short: "Show privacy key by address",
		Run:   showPrivacyKey,
	}
	showPrivacyKeyFlag(cmd)
	return cmd
}

func showPrivacyKeyFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")

}

func showPrivacyKey(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	params := types.ReqString{
		Data: addr,
	}
	var res pty.ReplyPrivacyPkPair
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.ShowPrivacyKey", params, &res)
	ctx.Run()
}

// CreatePub2PrivTxCmd create a public to privacy transaction
func createPub2PrivTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pub2priv",
		Short: "Create a public to privacy transaction",
		Run:   createPub2PrivTx,
	}
	createPub2PrivTxFlags(cmd)
	return cmd
}

func createPub2PrivTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkeypair", "p", "", "public key pair")
	cmd.MarkFlagRequired("pubkeypair")
	cmd.Flags().Float64P("amount", "a", 0.0, "transfer amount, at most 4 decimal places")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("note", "n", "", "note for transaction")
	cmd.Flags().Int64P("expire", "x", 0, "transfer expire, default one hour")
	cmd.Flags().IntP("expiretype", "", 1, "0: height  1: time default is 1")
}

func createPub2PrivTx(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkeypair, _ := cmd.Flags().GetString("pubkeypair")
	amount := cmdtypes.GetAmountValue(cmd, "amount")
	tokenname, _ := cmd.Flags().GetString("symbol")
	note, _ := cmd.Flags().GetString("note")
	expire, _ := cmd.Flags().GetInt64("expire")
	expiretype, _ := cmd.Flags().GetInt("expiretype")
	assetExec, _ := cmd.Flags().GetString("exec")
	if expiretype == 0 {
		if expire <= 0 {
			fmt.Println("Invalid expire. expire must large than 0 in expiretype==0, expire", expire)
			return
		}
	} else if expiretype == 1 {
		if expire <= 0 {
			expire = int64(time.Minute * 10)
		}
	} else {
		fmt.Println("Invalid expiretype", expiretype)
		return
	}

	params := pty.ReqCreatePrivacyTx{
		Tokenname:  tokenname,
		ActionType: pty.ActionPublic2Privacy,
		Amount:     amount,
		Note:       note,
		Pubkeypair: pubkeypair,
		Expire:     expire,
		AssetExec:  assetExec,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreatePriv2PrivTxCmd create a privacy to privacy transaction
func createPriv2PrivTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priv2priv",
		Short: "Create a privacy to privacy transaction",
		Run:   createPriv2PrivTx,
	}
	createPriv2PrivTxFlags(cmd)
	return cmd
}

func createPriv2PrivTxFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("pubkeypair", "p", "", "public key pair")
	cmd.MarkFlagRequired("pubkeypair")
	cmd.Flags().Float64P("amount", "a", 0.0, "transfer amount, at most 4 decimal places")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("from", "f", "", "from address")
	cmd.MarkFlagRequired("from")

	cmd.Flags().Int32P("mixcount", "m", defMixCount, "utxo mix count")
	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("note", "n", "", "note for transaction")
	cmd.Flags().Int64P("expire", "x", 0, "transfer expire, default one hour")
	cmd.Flags().IntP("expiretype", "", 1, "0: height  1: time default is 1")
}

func createPriv2PrivTx(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	pubkeypair, _ := cmd.Flags().GetString("pubkeypair")
	amount := cmdtypes.GetAmountValue(cmd, "amount")
	mixCount, _ := cmd.Flags().GetInt32("mixcount")
	tokenname, _ := cmd.Flags().GetString("symbol")
	note, _ := cmd.Flags().GetString("note")
	sender, _ := cmd.Flags().GetString("from")
	expire, _ := cmd.Flags().GetInt64("expire")
	expiretype, _ := cmd.Flags().GetInt("expiretype")
	assetExec, _ := cmd.Flags().GetString("exec")
	if expiretype == 0 {
		if expire <= 0 {
			fmt.Println("Invalid expire. expire must large than 0 in expiretype==0, expire", expire)
			return
		}
	} else if expiretype == 1 {
		if expire <= 0 {
			expire = int64(time.Minute * 10)
		}
	} else {
		fmt.Println("Invalid expiretype", expiretype)
		return
	}

	params := pty.ReqCreatePrivacyTx{
		Tokenname:  tokenname,
		ActionType: pty.ActionPrivacy2Privacy,
		Amount:     amount,
		Note:       note,
		Pubkeypair: pubkeypair,
		From:       sender,
		Mixcount:   mixCount,
		Expire:     expire,
		AssetExec:  assetExec,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

// CreatePriv2PubTxCmd create a privacy to public transaction
func createPriv2PubTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "priv2pub",
		Short: "Create a privacy to public transaction",
		Run:   createPriv2PubTx,
	}
	createPriv2PubTxFlags(cmd)
	return cmd
}

func createPriv2PubTxFlags(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0.0, "transfer amount, at most 4 decimal places")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("from", "f", "", "from address")
	cmd.MarkFlagRequired("from")
	cmd.Flags().StringP("to", "t", "", "to address")
	cmd.MarkFlagRequired("to")

	cmd.Flags().Int32P("mixcount", "m", defMixCount, "utxo mix count")
	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("note", "n", "", "note for transaction")
	cmd.Flags().Int64P("expire", "x", 0, "transfer expire, default one hour")
	cmd.Flags().IntP("expiretype", "", 1, "0: height  1: time default is 1")
}

func createPriv2PubTx(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	amount := cmdtypes.GetAmountValue(cmd, "amount")
	mixCount, _ := cmd.Flags().GetInt32("mixcount")
	tokenname, _ := cmd.Flags().GetString("symbol")
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	note, _ := cmd.Flags().GetString("note")
	expire, _ := cmd.Flags().GetInt64("expire")
	expiretype, _ := cmd.Flags().GetInt("expiretype")
	assetExec, _ := cmd.Flags().GetString("exec")
	if expiretype == 0 {
		if expire <= 0 {
			fmt.Println("Invalid expire. expire must large than 0 in expiretype==0, expire", expire)
			return
		}
	} else if expiretype == 1 {
		if expire <= 0 {
			expire = int64(time.Minute * 10)
		}
	} else {
		fmt.Println("Invalid expiretype", expiretype)
		return
	}

	params := pty.ReqCreatePrivacyTx{
		Tokenname:  tokenname,
		ActionType: pty.ActionPrivacy2Public,
		Amount:     amount,
		Note:       note,
		From:       from,
		To:         to,
		Mixcount:   mixCount,
		Expire:     expire,
		AssetExec:  assetExec,
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.CreateRawTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func showPrivacyAccountSpendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpas",
		Short: "Show privacy account spend command",
		Run:   showPrivacyAccountSpend,
	}
	showPrivacyAccountSpendFlag(cmd)
	return cmd
}

func showPrivacyAccountSpendFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.MarkFlagRequired("addr")
}

func showPrivacyAccountSpend(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	exec, _ := cmd.Flags().GetString("exec")
	symbol, _ := cmd.Flags().GetString("symbol")
	params := pty.ReqPrivBal4AddrToken{
		Addr:      addr,
		Token:     symbol,
		AssetExec: exec,
	}

	var res pty.UTXOHaveTxHashs
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.ShowPrivacyAccountSpend", params, &res)
	ctx.SetResultCb(parseShowPrivacyAccountSpendRes)
	ctx.Run()
}

func parseShowPrivacyAccountSpendRes(arg interface{}) (interface{}, error) {
	total := float64(0)
	res := arg.(*pty.UTXOHaveTxHashs)
	rets := make([]*PrivacyAccountSpendResult, 0)
	for _, utxo := range res.UtxoHaveTxHashs {
		amount := float64(utxo.Amount) / float64(types.Coin)
		total += amount

		var isSave bool
		for _, ret := range rets {
			if utxo.TxHash == ret.Txhash {
				result := &PrivacyAccountResult{
					Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
					OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
					Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
				}
				ret.Res = append(ret.Res, result)
				isSave = true
				break
			}
		}

		if !isSave {
			result := &PrivacyAccountResult{
				//Height:   utxo.UtxoBasic.UtxoGlobalIndex.Height,
				//TxIndex:  utxo.UtxoBasic.UtxoGlobalIndex.Txindex,
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			var SpendResult PrivacyAccountSpendResult
			SpendResult.Txhash = utxo.TxHash
			SpendResult.Res = append(SpendResult.Res, result)
			rets = append(rets, &SpendResult)
		}
	}

	fmt.Println(fmt.Sprintf("Total Privacy spend amount is %s", strconv.FormatFloat(total, 'f', 4, 64)))

	return rets, nil
}

func showAmountsOfUTXOCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showau",
		Short: "Show Amount of UTXO",
		Run:   showAmountOfUTXO,
	}
	showAmountOfUTXOFlag(cmd)
	return cmd
}

func showAmountOfUTXOFlag(cmd *cobra.Command) {

	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
}

func showAmountOfUTXO(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	exec, _ := cmd.Flags().GetString("exec")
	symbol, _ := cmd.Flags().GetString("symbol")
	reqPrivacyToken := pty.ReqPrivacyToken{AssetExec: exec, AssetSymbol: symbol}
	var params rpctypes.Query4Jrpc
	params.Execer = pty.PrivacyX
	params.FuncName = "ShowAmountsOfUTXO"
	params.Payload = types.MustPBToJSON(&reqPrivacyToken)

	var res pty.ReplyPrivacyAmounts
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(parseShowAmountOfUTXORes)
	ctx.Run()
}

func parseShowAmountOfUTXORes(arg interface{}) (interface{}, error) {
	res := arg.(*pty.ReplyPrivacyAmounts)
	for _, amount := range res.AmountDetail {
		amount.Amount = amount.Amount / types.Coin
	}
	return res, nil
}

func showUTXOs4SpecifiedAmountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showutxo4a",
		Short: "Show specified amount UTXOs",
		Run:   showUTXOs4SpecifiedAmount,
	}
	showUTXOs4SpecifiedAmountFlag(cmd)
	return cmd
}

func showUTXOs4SpecifiedAmountFlag(cmd *cobra.Command) {
	cmd.Flags().Float64P("amount", "a", 0, "amount")
	cmd.MarkFlagRequired("amount")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
}

func showUTXOs4SpecifiedAmount(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	amount, _ := cmd.Flags().GetFloat64("amount")
	amountInt64 := int64(amount*types.InputPrecision) * types.Multiple1E4
	exec, _ := cmd.Flags().GetString("exec")
	symbol, _ := cmd.Flags().GetString("symbol")
	reqPrivacyToken := pty.ReqPrivacyToken{
		AssetExec:   exec,
		AssetSymbol: symbol,
		Amount:      amountInt64,
	}
	var params rpctypes.Query4Jrpc
	params.Execer = pty.PrivacyX
	params.FuncName = "ShowUTXOs4SpecifiedAmount"
	params.Payload = types.MustPBToJSON(&reqPrivacyToken)

	var res pty.ReplyUTXOsOfAmount
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
	ctx.SetResultCb(parseShowUTXOs4SpecifiedAmountRes)
	ctx.Run()
}

func parseShowUTXOs4SpecifiedAmountRes(arg interface{}) (interface{}, error) {
	res := arg.(*pty.ReplyUTXOsOfAmount)
	ret := make([]*PrivacyAccountResult, 0)
	for _, item := range res.LocalUTXOItems {
		result := &PrivacyAccountResult{
			Txhash:        common.ToHex(item.Txhash),
			OutIndex:      item.Outindex,
			OnetimePubKey: hex.EncodeToString(item.Onetimepubkey),
		}
		ret = append(ret, result)
	}

	return ret, nil
}

// showPrivacyAccountInfoCmd
func showPrivacyAccountInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "showpai",
		Short: "Show privacy account information",
		Run:   showPrivacyAccountInfo,
	}
	showPrivacyAccountInfoFlag(cmd)
	return cmd
}

func showPrivacyAccountInfoFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("symbol", "s", "BTY", "asset symbol, default BTY")
	cmd.Flags().Int32P("displaymode", "d", 0, "display mode.(0: display collect. 1:display available detail. 2:display frozen detail. 3:display all")
}

func showPrivacyAccountInfo(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	exec, _ := cmd.Flags().GetString("exec")
	token, _ := cmd.Flags().GetString("symbol")
	mode, _ := cmd.Flags().GetInt32("displaymode")
	if mode < 0 || mode > 3 {
		fmt.Println("display mode only support 0-3")
		return
	}

	params := pty.ReqPrivacyAccount{
		Addr:        addr,
		Token:       token,
		Displaymode: mode,
		AssetExec:   exec,
	}

	var res pty.ReplyPrivacyAccount
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.ShowPrivacyAccountInfo", params, &res)
	ctx.SetResultCb(parseshowPrivacyAccountInfo)
	ctx.Run()
}

func parseshowPrivacyAccountInfo(arg interface{}) (interface{}, error) {
	total := float64(0)
	totalFrozen := float64(0)
	res := arg.(*pty.ReplyPrivacyAccount)

	var availableAmount, frozenAmount, totalAmount string

	utxos := make([]*PrivacyAccountResult, 0)
	for _, utxo := range res.Utxos.Utxos {
		amount := float64(utxo.Amount) / float64(types.Coin)
		total += amount

		if res.Displaymode == 1 || res.Displaymode == 3 {
			result := &PrivacyAccountResult{
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			utxos = append(utxos, result)
		}
	}
	availableAmount = strconv.FormatFloat(total, 'f', 4, 64)

	ftxos := make([]*PrivacyAccountResult, 0)
	for _, utxo := range res.Ftxos.Utxos {
		amount := float64(utxo.Amount) / float64(types.Coin)
		totalFrozen += amount

		if res.Displaymode == 2 || res.Displaymode == 3 {
			result := &PrivacyAccountResult{
				Txhash:   common.ToHex(utxo.UtxoBasic.UtxoGlobalIndex.Txhash),
				OutIndex: utxo.UtxoBasic.UtxoGlobalIndex.Outindex,
				Amount:   strconv.FormatFloat(amount, 'f', 4, 64),
			}
			ftxos = append(ftxos, result)
		}
	}
	frozenAmount = strconv.FormatFloat(totalFrozen, 'f', 4, 64)
	totalAmount = strconv.FormatFloat(total+totalFrozen, 'f', 4, 64)

	ret := &PrivacyAccountInfoResult{
		AvailableDetail: utxos,
		FrozenDetail:    ftxos,
		AvailableAmount: availableAmount,
		FrozenAmount:    frozenAmount,
		TotalAmount:     totalAmount,
	}
	return ret, nil
}

func listPrivacyTxsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list_txs",
		Short: "List privacy transactions in wallet",
		Run:   listPrivacyTxsFlags,
	}
	addListPrivacyTxsFlags(cmd)
	return cmd
}

func addListPrivacyTxsFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account address")
	cmd.MarkFlagRequired("addr")
	//
	cmd.Flags().Int32P("sendrecv", "f", 0, "send or recv flag (0: send, 1: recv), default 0")
	cmd.Flags().Int32P("count", "c", 10, "number of transactions, default 10")
	cmd.Flags().StringP("exec", "e", "coins", "asset executor(coins, token, paracross), default coins")
	cmd.Flags().StringP("symbol", "s", types.BTY, "asset sysmbol.(BTY default)")
	cmd.Flags().Int32P("direction", "d", 1, "query direction (0: pre page, 1: next page), valid with seedtxhash param")
	cmd.Flags().StringP("txHeightIndex", "x", "", "start tx height index string")
}

func listPrivacyTxsFlags(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	count, _ := cmd.Flags().GetInt32("count")
	direction, _ := cmd.Flags().GetInt32("direction")
	addr, _ := cmd.Flags().GetString("addr")
	sendRecvFlag, _ := cmd.Flags().GetInt32("sendrecv")
	symbol, _ := cmd.Flags().GetString("symbol")
	txHeightIndex, _ := cmd.Flags().GetString("txHeightIndex")
	exec, _ := cmd.Flags().GetString("exec")
	params := pty.ReqPrivacyTransactionList{
		AssetExec:          exec,
		AssetSymbol:        symbol,
		SendRecvFlag:       sendRecvFlag,
		Direction:          direction,
		Count:              count,
		Address:            addr,
		StartTxHeightIndex: txHeightIndex,
	}
	var res rpctypes.WalletTxDetails
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.GetPrivacyTxByAddr", params, &res)
	ctx.SetResultCb(parseWalletTxListRes)
	ctx.Run()
}

func parseWalletTxListRes(arg interface{}) (interface{}, error) {
	res := arg.(*rpctypes.WalletTxDetails)
	var result cmdtypes.WalletTxDetailsResult
	for _, v := range res.TxDetails {
		amountResult := strconv.FormatFloat(float64(v.Amount)/float64(types.Coin), 'f', 4, 64)
		wtxd := &cmdtypes.WalletTxDetailResult{
			Tx:         cmdtypes.DecodeTransaction(v.Tx),
			Receipt:    v.Receipt,
			Height:     v.Height,
			Index:      v.Index,
			Blocktime:  v.BlockTime,
			Amount:     amountResult,
			Fromaddr:   v.FromAddr,
			Txhash:     v.TxHash,
			ActionName: v.ActionName,
		}
		result.TxDetails = append(result.TxDetails, wtxd)
	}
	return result, nil
}

func rescanUtxosOptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rescanOpt",
		Short: "rescan Utxos in wallet and query rescan utxos status",
		Run:   rescanUtxosOpt,
	}
	rescanUtxosOptFlags(cmd)
	return cmd
}

func rescanUtxosOptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "privacy rescanOpt -a [all-addr0-addr1] (all indicate all wallet address)")
	cmd.MarkFlagRequired("addr")
	//
	cmd.Flags().Int32P("flag", "f", 0, "Rescan or query rescan flag (0: Rescan, 1: query rescan)")
}

func rescanUtxosOpt(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	address, _ := cmd.Flags().GetString("addr")
	flag, _ := cmd.Flags().GetInt32("flag")

	var params pty.ReqRescanUtxos

	params.Flag = flag
	if "all" != address {
		if len(address) > 0 {
			addrs := strings.Split(address, "-")
			params.Addrs = append(params.Addrs, addrs...)
		}
	}

	var res pty.RepRescanUtxos
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.RescanUtxos", params, &res)
	ctx.SetResultCb(parseRescanUtxosOpt)
	ctx.Run()
}

func parseRescanUtxosOpt(arg interface{}) (interface{}, error) {
	res := arg.(*pty.RepRescanUtxos)
	if 0 == res.Flag {
		str := "start rescan UTXO"
		return str, nil
	}

	var result showRescanResults
	for _, v := range res.RepRescanResults {
		str, ok := pty.RescanFlagMapint2string[v.Flag]
		if ok {
			showRescanResult := &ShowRescanResult{
				Addr:       v.Addr,
				FlagString: str,
			}
			result.RescanResults = append(result.RescanResults, showRescanResult)
		}
	}
	return &result, nil
}

func enablePrivacyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "enable",
		Short: "enable privacy address in wallet",
		Run:   enablePrivacy,
	}
	enablePrivacyFlags(cmd)
	return cmd
}

func enablePrivacyFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "privacy enable -a [all-addr0-addr1] (all indicate enable all wallet address)")
	cmd.MarkFlagRequired("addr")
}

func enablePrivacy(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	address, _ := cmd.Flags().GetString("addr")

	var params pty.ReqEnablePrivacy

	if "all" != address {
		if len(address) > 0 {
			addrs := strings.Split(address, "-")
			params.Addrs = append(params.Addrs, addrs...)
		}
	}

	var res pty.RepEnablePrivacy
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "privacy.EnablePrivacy", params, &res)
	ctx.SetResultCb(parseEnablePrivacy)
	ctx.Run()
}

func parseEnablePrivacy(arg interface{}) (interface{}, error) {
	res := arg.(*pty.RepEnablePrivacy)

	var result ShowEnablePrivacy
	for _, v := range res.Results {
		showPriAddrResult := &ShowPriAddrResult{
			Addr: v.Addr,
			IsOK: v.IsOK,
			Msg:  v.Msg,
		}
		result.Results = append(result.Results, showPriAddrResult)
	}
	return &result, nil
}

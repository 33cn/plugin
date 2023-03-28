// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	cmdtypes "github.com/33cn/chain33/system/dapp/commands/types"
	"github.com/pkg/errors"

	"encoding/json"
	"strings"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto/sha3"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	evmAbi "github.com/33cn/plugin/plugin/dapp/evm/executor/abi"
	evmCommon "github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	evmtypes "github.com/33cn/plugin/plugin/dapp/evm/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

//EvmCmd 是Evm命令行入口
func EvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evm",
		Short: "EVM contracts operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		calcNewContractAddrCmd(),
		createContractCmd(),
		callContractCmd(),
		queryCmd(),
		estimateGasCmd(),
		checkContractAddrCmd(),
		evmDebugCmd(),
		evmTransferCmd(),
		getEvmBalanceCmd(),
		evmToolsCmd(),
		showTimeNowCmd(),
	)
	cmd.PersistentFlags().Int32("chainID", 0, "chain ID")

	return cmd
}

// some tools for evm op
func evmToolsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tool",
		Short: "Some tools for evm op",
	}
	cmd.AddCommand(evmToolsAddressCmd())

	return cmd
}

// transfer address format between ethereum and chain33
func evmToolsAddressCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "address",
		Short: "Transfer address format between ethereum and local (you should input one address of them)",
		Run:   transferAddress,
	}
	addEvmAddressFlags(cmd)
	return cmd
}

func addEvmAddressFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("eth", "e", "", "ethereum address")

	cmd.Flags().StringP("local", "l", "", "evm contract address (like user.evm.xxx or plain address)")
}

func transferAddress(cmd *cobra.Command, args []string) {
	eth, _ := cmd.Flags().GetString("eth")
	local, _ := cmd.Flags().GetString("local")
	if len(eth) == 40 || len(eth) == 42 {
		data, err := common.FromHex(eth)
		if err != nil {
			fmt.Println(fmt.Errorf("ethereum address is invalid: %v", eth))
			return
		}
		fmt.Println(fmt.Sprintf("Ethereum Address: %v", eth))
		fmt.Println(fmt.Sprintf("Local Address: %v", evmCommon.BytesToAddress(data).String()))
		return
	}
	if len(local) >= 34 {
		var addr evmCommon.Address
		if strings.HasPrefix(local, evmtypes.EvmPrefix) {
			addr = evmCommon.ExecAddress(local)
			fmt.Println(fmt.Sprintf("Local Contract Name: %v", local))
			fmt.Println(fmt.Sprintf("Local Address: %v", addr.String()))
		} else {
			addrP := evmCommon.StringToAddress(local)
			if addrP == nil {
				fmt.Println(fmt.Errorf("Local address is invalid: %v", local))
				return
			}
			addr = *addrP
			fmt.Println(fmt.Sprintf("Local Address: %v", local))
		}
		fmt.Println(fmt.Sprintf("Ethereum Address: %v", checksumAddr(addr.Bytes())))

		return
	}
	fmt.Fprintln(os.Stderr, "address is invalid!")
}

// get balance of an execer
func getEvmBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Get balance of a evm contract address",
		Run:   evmBalance,
	}
	addEvmBalanceFlags(cmd)
	return cmd
}

func addEvmBalanceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account addr")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("exec", "e", "", "evm contract name (like user.evm.xxx)")
	cmd.MarkFlagRequired("exec")
}

func evmBalance(cmd *cobra.Command, args []string) {
	// 直接复用coins的查询余额命令
	//balance(cmd, args)

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	execer, _ := cmd.Flags().GetString("exec")
	err := address.CheckAddress(addr, -1)
	if err != nil {
		fmt.Fprintln(os.Stderr, types.ErrInvalidAddress)
		return
	}
	if ok := types.IsAllowExecName([]byte(execer), []byte(execer)); !ok {
		fmt.Fprintln(os.Stderr, types.ErrExecNameNotAllow)
		return
	}

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}

	var addrs []string
	addrs = append(addrs, addr)
	params := types.ReqBalance{
		Addresses: addrs,
		Execer:    execer,
		StateHash: "",
	}
	var res []*rpctypes.Account
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.GetBalance", params, &res)
	ctx.SetResultCbExt(parseGetBalanceRes)
	ctx.RunExt(cfg)
}

// AccountResult 账户余额查询出来之后进行单位转换
type AccountResult struct {
	// 货币
	Currency int32 `json:"currency,omitempty"`
	// 余额
	Balance string `json:"balance,omitempty"`
	// 冻结余额
	Frozen string `json:"frozen,omitempty"`
	// 账户地址
	Addr string `json:"addr,omitempty"`
}

func parseGetBalanceRes(arg ...interface{}) (interface{}, error) {
	res := *arg[0].(*[]*rpctypes.Account)
	cfg := arg[1].(*rpctypes.ChainConfigInfo)
	balanceResult := types.FormatAmount2FloatDisplay(res[0].Balance, cfg.CoinPrecision, true)
	frozenResult := types.FormatAmount2FloatDisplay(res[0].Frozen, cfg.CoinPrecision, true)
	result := &AccountResult{
		Addr:     res[0].Addr,
		Currency: res[0].Currency,
		Balance:  balanceResult,
		Frozen:   frozenResult,
	}
	return result, nil
}

func calcNewContractAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "calc",
		Short: "calulate new contract address",
		Run:   calcNewContractAddr,
	}
	addCalcNewContractAddrFlags(cmd)
	return cmd
}

func calcNewContractAddr(cmd *cobra.Command, args []string) {
	deployer, _ := cmd.Flags().GetString("deployer")
	hash, _ := cmd.Flags().GetString("hash")
	params := &evmtypes.EvmCalcNewContractAddrReq{
		Caller: deployer,
		Txhash: hash,
	}

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "evm.CalcNewContractAddr", params, &res)
	ctx.Run()
}

func addCalcNewContractAddrFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("deployer", "a", "", "deployer address")
	_ = cmd.MarkFlagRequired("deployer")

	cmd.Flags().StringP("hash", "s", "", "tx hash(exclude prefix '0x') which deployed the new contract ")
	_ = cmd.MarkFlagRequired("hash")
}

// 创建EVM合约
func createContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tx to deploy new EVM contract",
		Run:   createContract,
	}
	addCreateContractFlags(cmd)
	return cmd
}

func addCreateContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("code", "c", "", "contract binary code")
	_ = cmd.MarkFlagRequired("code")

	cmd.Flags().StringP("abi", "b", "", "abi string used for create constructor parameter(optional, not needed if no parameter for constructor)")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")

	cmd.Flags().StringP("alias", "s", "", "human readable contract alias name(optional)")

	cmd.Flags().StringP("parameter", "p", "", "(optional)parameter for constructor and should be input as constructor(xxx,xxx,xxx)")

}

func createContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	code, _ := cmd.Flags().GetString("code")
	note, _ := cmd.Flags().GetString("note")
	alias, _ := cmd.Flags().GetString("alias")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	abi, _ := cmd.Flags().GetString("abi")
	constructorPara, _ := cmd.Flags().GetString("parameter")
	chainID, _ := cmd.Flags().GetInt32("chainID")

	cfg, err := cmdtypes.GetChainConfig(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "GetChainConfig"))
		return
	}
	feeInt64, err := types.FormatFloatDisplay2Value(fee, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.fee"))
		return
	}

	var action evmtypes.EVMContractAction
	bCode, err := common.FromHex(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return
	}
	exector := types.GetExecName("evm", paraName)
	action = evmtypes.EVMContractAction{Amount: 0, Code: bCode, GasLimit: 0, GasPrice: 0, Note: note, Alias: alias, ContractAddr: address.ExecAddress(exector)}

	if "" != constructorPara {
		packData, err := evmAbi.PackContructorPara(constructorPara, abi)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Pack Contructor Para error:", err)
			return
		}

		action.Code = append(action.Code, packData...)
	}

	tx := &types.Transaction{Execer: []byte(exector), Payload: types.Encode(&action), Fee: 0, To: action.ContractAddr, ChainID: chainID}
	tx.Fee, _ = tx.GetRealFee(cfg.MinTxFeeRate)
	if tx.Fee < feeInt64 {
		tx.Fee += feeInt64
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)
	fmt.Println(rawTx)
}

func createEvmTx(cfg *rpctypes.ChainConfigInfo, action proto.Message, execer, caller, toAddr, expire, rpcLaddr string, fee int64, chainID int32) (string, error) {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: 0, To: toAddr, ChainID: chainID}

	tx.Fee, _ = tx.GetRealFee(cfg.MinTxFeeRate)
	if tx.Fee < fee {
		tx.Fee += fee
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	//tx.ChainID = cfg.GetChainID()
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)

	unsignedTx := &types.ReqSignRawTx{
		Addr:   caller,
		TxHex:  rawTx,
		Expire: expire,
		Fee:    tx.Fee,
	}

	var res string
	client, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}
	err = client.Call("Chain33.SignRawTx", unsignedTx, &res)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", err
	}

	return res, nil
}

// 调用EVM合约
func callContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "create tx to Call the EVM contract",
		Run:   callContract,
	}
	addCallContractFlags(cmd)
	return cmd
}

func callContract(cmd *cobra.Command, args []string) {
	note, _ := cmd.Flags().GetString("note")
	amount, _ := cmd.Flags().GetFloat64("amount")
	fee, _ := cmd.Flags().GetFloat64("fee")
	contractAddr, _ := cmd.Flags().GetString("exec")
	parameter, _ := cmd.Flags().GetString("parameter")
	path, _ := cmd.Flags().GetString("path")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")

	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
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
	feeInt64, err := types.FormatFloatDisplay2Value(fee, cfg.CoinPrecision)
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "FormatFloatDisplay2Value.fee"))
		return
	}

	abiFileName := path + contractAddr + ".abi"
	abiStr, err := readFile(abiFileName)
	if nil != err {
		_, _ = fmt.Fprintln(os.Stderr, "Can't read abi info, Pls set correct abi path and provide abi file as", abiFileName)
		return
	}

	_, packedParameter, err := evmAbi.Pack(parameter, abiStr, false)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to do para pack", err.Error())
		return
	}

	action := evmtypes.EVMContractAction{Amount: uint64(amountInt64), GasLimit: 0, GasPrice: 0, Note: note, Para: packedParameter, ContractAddr: contractAddr}

	exector := types.GetExecName("evm", paraName)
	toAddr := address.ExecAddress(exector)

	tx := &types.Transaction{Execer: []byte(exector), Payload: types.Encode(&action), Fee: 0, To: toAddr, ChainID: chainID}
	tx.Fee, _ = tx.GetRealFee(cfg.MinTxFeeRate)
	if tx.Fee < feeInt64 {
		tx.Fee += feeInt64
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	//tx.ChainID = cfg.GetChainID()
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)
	fmt.Println(rawTx)
}

func addCallContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")

	cmd.Flags().StringP("exec", "e", "", "evm contract address")
	_ = cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")

	cmd.Flags().StringP("parameter", "p", "", "tx input parameter as:approve(13nBqpmC4VaJpEZ6J6G9NUM1Y55FQvw558, 100000000)")
	_ = cmd.MarkFlagRequired("parameter")
}

// abi命令
func queryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "evm query call",
		Run:   evmQueryCall,
	}

	cmd.Flags().StringP("address", "a", "", "evm contract address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("input", "b", "", "call params (abi format) like foobar(param1,param2)")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("caller", "c", "", "the caller address")

	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	return cmd
}

func evmQueryCall(cmd *cobra.Command, args []string) {
	addr, _ := cmd.Flags().GetString("address")
	input, _ := cmd.Flags().GetString("input")
	caller, _ := cmd.Flags().GetString("caller")
	path, _ := cmd.Flags().GetString("path")

	abiFileName := path + addr + ".abi"
	abiStr, err := readFile(abiFileName)
	if nil != err {
		_, _ = fmt.Fprintln(os.Stderr, "Can't read abi info, Pls set correct abi path and provide abi file as", abiFileName)
		return
	}

	methodName, packData, err := evmAbi.Pack(input, abiStr, true)
	if nil != err {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to do evmAbi.Pack")
		return
	}
	packStr := common.ToHex(packData)
	var req = evmtypes.EvmQueryReq{Address: addr, Input: packStr, Caller: caller}
	var resp evmtypes.EvmQueryResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	query := sendQuery(rpcLaddr, "Query", &req, &resp)
	if !query {
		fmt.Println("Failed to send query")
		return

	}
	_, err = json.MarshalIndent(&resp, "", "  ")
	if err != nil {
		fmt.Println("MarshalIndent failed due to:", err.Error())
	}

	data, err := common.FromHex(resp.RawData)
	if nil != err {
		fmt.Println("common.FromHex failed due to:", err.Error())
	}

	outputs, err := evmAbi.Unpack(data, methodName, abiStr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "unpack evm return error", err)
	}

	for _, v := range outputs {
		fmt.Println(v.Value)
	}
}

func estimateGas(cmd *cobra.Command, args []string) {
	txStr, _ := cmd.Flags().GetString("tx")
	caller, _ := cmd.Flags().GetString("caller")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	txInfo := &evmtypes.EstimateEVMGasReq{
		Tx:   txStr,
		From: caller,
	}

	var estGasResp evmtypes.EstimateEVMGasResp
	query := sendQuery(rpcLaddr, "EstimateGas", txInfo, &estGasResp)
	if query {
		fmt.Fprintf(os.Stdout, "gas cost estimate %v\n", estGasResp.Gas)
	} else {
		fmt.Fprintln(os.Stderr, "gas cost estimate error")
	}
}

func addEstimateGasFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("tx", "x", "", "tx string(should be signatured)")
	_ = cmd.MarkFlagRequired("tx")

	cmd.Flags().StringP("caller", "c", "", "contract creator or caller")
	_ = cmd.MarkFlagRequired("caller")
}

// 估算合约消耗
func estimateGasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate the gas cost of calling or creating a contract",
		Run:   estimateGas,
	}
	addEstimateGasFlags(cmd)
	return cmd
}

// 检查地址是否为EVM合约
func checkContractAddrCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if the address is for a valid EVM contract",
		Run:   checkContractAddr,
	}
	addCheckContractAddrFlags(cmd)
	return cmd
}

func addCheckContractAddrFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "evm contract address")
	_ = cmd.MarkFlagRequired("addr")
}

func checkContractAddr(cmd *cobra.Command, args []string) {
	addr, _ := cmd.Flags().GetString("addr")

	var checkAddrReq = evmtypes.CheckEVMAddrReq{Addr: addr}
	var checkAddrResp evmtypes.CheckEVMAddrResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery(rpcLaddr, "CheckAddrExists", &checkAddrReq, &checkAddrResp)

	if query && checkAddrResp.Contract {
		proto.MarshalText(os.Stdout, &checkAddrResp)
	} else {
		fmt.Fprintln(os.Stderr, "not evm contract addr!")
	}
}

// 查询或设置EVM调试开关
func evmDebugCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Query or set evm debug status",
	}
	cmd.AddCommand(
		evmDebugQueryCmd(),
		evmDebugSetCmd(),
		evmDebugClearCmd())

	return cmd
}

func evmDebugQueryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query",
		Short: "Query evm debug status",
		Run:   evmDebugQuery,
	}
}
func evmDebugSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set",
		Short: "Set evm debug to ON",
		Run:   evmDebugSet,
	}
}
func evmDebugClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Set evm debug to OFF",
		Run:   evmDebugClear,
	}
}

func evmDebugQuery(cmd *cobra.Command, args []string) {
	evmDebugRPC(cmd, 0)
}

func evmDebugSet(cmd *cobra.Command, args []string) {
	evmDebugRPC(cmd, 1)
}

func evmDebugClear(cmd *cobra.Command, args []string) {
	evmDebugRPC(cmd, -1)
}
func evmDebugRPC(cmd *cobra.Command, flag int32) {
	var debugReq = evmtypes.EvmDebugReq{Optype: flag}
	var debugResp evmtypes.EvmDebugResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery(rpcLaddr, "EvmDebug", &debugReq, &debugResp)

	if query {
		proto.MarshalText(os.Stdout, &debugResp)
	} else {
		fmt.Fprintln(os.Stderr, "error")
	}
}

// 向EVM合约地址转账
func evmTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "transfer within evm contract platform, just same ETH transfer on Ethereum",
		Run:   evmTransfer,
	}
	addEvmTransferFlags(cmd)
	return cmd
}

func addEvmTransferFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("receiver", "r", "", "receiver address")
	cmd.MarkFlagRequired("receiver")

	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract, precision to 0.0001")
	cmd.MarkFlagRequired("amount")

	cmd.Flags().StringP("expire", "p", "120s", "transaction expire time (optional)")
}

func evmTransfer(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	paraName, _ := cmd.Flags().GetString("paraName")
	caller, _ := cmd.Flags().GetString("caller")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	expire, _ := cmd.Flags().GetString("expire")
	chainID, _ := cmd.Flags().GetInt32("chainID")

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

	r_addr, err := address.NewBtcAddress(receiver)
	if nil != err {
		_, _ = fmt.Println("Pls input correct address")
		return
	}

	exector := types.GetExecName("evm", paraName)
	toAddr := address.ExecAddress(exector)
	action := &evmtypes.EVMContractAction{
		Amount:       uint64(amountInt64),
		GasLimit:     0,
		GasPrice:     0,
		Code:         nil,
		Para:         r_addr.Hash160[:],
		Alias:        "",
		Note:         fmt.Sprintf("transfer from:"+caller+" to:"+receiver+" for amount: %s", amount),
		ContractAddr: toAddr,
	}
	data, err := createEvmTx(cfg, action, exector, caller, toAddr, expire, rpcLaddr, 0, chainID)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "create transfer tx error:", err)
		return
	}

	params := rpctypes.RawParm{
		Data: data,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func sendQuery(rpcAddr, funcName string, request types.Message, result proto.Message) bool {
	params := rpctypes.Query4Jrpc{
		Execer:   "evm",
		FuncName: funcName,
		Payload:  types.MustPBToJSON(request),
	}

	jsonrpc, err := jsonclient.NewJSONClient(rpcAddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = jsonrpc.Call("Chain33.Query", params, result)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}

// 这里实现 EIP55中提及的以太坊地址表示方式（增加Checksum）
func checksumAddr(address []byte) string {
	unchecksummed := hex.EncodeToString(address[:])
	sha := sha3.NewKeccak256()
	sha.Write([]byte(unchecksummed))
	hash := sha.Sum(nil)

	result := []byte(unchecksummed)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}
	return "0x" + string(result)
}

func showTimeNowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "now",
		Short: "show seconds from epoch",
		Run:   showNow,
	}
	return cmd
}

func showNow(cmd *cobra.Command, args []string) {
	fmt.Println(time.Now().Unix())
}

func readFile(fileName string) (string, error) {
	f, err := os.Open(fileName)
	defer f.Close()
	if err != nil {
		return "", err
	}

	fileContent, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return string(fileContent), nil
}

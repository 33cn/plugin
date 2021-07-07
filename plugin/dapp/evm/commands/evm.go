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

	"encoding/json"
	"strconv"
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
		abiCmd(),
		estimateContractCmd(),
		checkContractAddrCmd(),
		evmDebugCmd(),
		evmTransferCmd(),
		getEvmBalanceCmd(),
		evmToolsCmd(),
		getNonceCmd(),
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
	err := address.CheckAddress(addr)
	if err != nil {
		fmt.Fprintln(os.Stderr, types.ErrInvalidAddress)
		return
	}
	if ok := types.IsAllowExecName([]byte(execer), []byte(execer)); !ok {
		fmt.Fprintln(os.Stderr, types.ErrExecNameNotAllow)
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
	ctx.SetResultCb(parseGetBalanceRes)
	ctx.Run()
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

func parseGetBalanceRes(arg interface{}) (interface{}, error) {
	res := *arg.(*[]*rpctypes.Account)
	balanceResult := strconv.FormatFloat(float64(res[0].Balance)/float64(types.Coin), 'f', 4, 64)
	frozenResult := strconv.FormatFloat(float64(res[0].Frozen)/float64(types.Coin), 'f', 4, 64)
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
	newContractAddr := address.GetExecAddress(deployer + hash)
	fmt.Println(newContractAddr.String())
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

	cmd.Flags().StringP("abi", "b", "", "abi string used for create constructor parameter ")
	_ = cmd.MarkFlagRequired("abi")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")

	cmd.Flags().StringP("alias", "s", "", "human readable contract alias name")

	cmd.Flags().StringP("parameter", "p", "", "parameter for constructor and should be input as constructor(xxx,xxx,xxx)")
}

func createContract(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	code, _ := cmd.Flags().GetString("code")
	note, _ := cmd.Flags().GetString("note")
	alias, _ := cmd.Flags().GetString("alias")
	fee, _ := cmd.Flags().GetFloat64("fee")
	paraName, _ := cmd.Flags().GetString("paraName")
	abi, _ := cmd.Flags().GetString("abi")
	constructorPara, _ := cmd.Flags().GetString("parameter")
	chainID, _ := cmd.Flags().GetInt32("chainID")

	feeInt64 := uint64(fee*1e4) * 1e4

	var action evmtypes.EVMContractAction
	bCode, err := common.FromHex(code)
	if err != nil {
		fmt.Fprintln(os.Stderr, "parse evm code error", err)
		return
	}
	exector := cfg.ExecName(paraName + "evm")
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
	tx.Fee, _ = tx.GetRealFee(cfg.GetMinTxFeeRate())
	fmt.Println("feeInt64 is", feeInt64)
	if tx.Fee < int64(feeInt64) {
		tx.Fee += int64(feeInt64)
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()
	txHex := types.Encode(tx)
	rawTx := hex.EncodeToString(txHex)
	fmt.Println(rawTx)
}

func createEvmTx(cfg *types.Chain33Config, action proto.Message, execer, caller, toAddr, expire, rpcLaddr string, fee uint64, chainID int32) (string, error) {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: 0, To: toAddr, ChainID: chainID}

	tx.Fee, _ = tx.GetRealFee(cfg.GetMinTxFeeRate())
	if tx.Fee < int64(fee) {
		tx.Fee += int64(fee)
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
		Short: "Call the EVM contract",
		Run:   callContract,
	}
	addCallContractFlags(cmd)
	return cmd
}

func callContract(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	//code, _ := cmd.Flags().GetString("input")
	caller, _ := cmd.Flags().GetString("caller")
	expire, _ := cmd.Flags().GetString("expire")
	note, _ := cmd.Flags().GetString("note")
	amount, _ := cmd.Flags().GetFloat64("amount")
	fee, _ := cmd.Flags().GetFloat64("fee")
	contractAddr, _ := cmd.Flags().GetString("exec")
	parameter, _ := cmd.Flags().GetString("parameter")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	path, _ := cmd.Flags().GetString("path")
	paraName, _ := cmd.Flags().GetString("paraName")
	chainID, _ := cmd.Flags().GetInt32("chainID")

	amountInt64 := uint64(amount*1e4) * 1e4
	feeInt64 := uint64(fee*1e4) * 1e4

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

	action := evmtypes.EVMContractAction{Amount: amountInt64, GasLimit: 0, GasPrice: 0, Note: note, Para: packedParameter, ContractAddr: contractAddr}

	exector := cfg.ExecName(paraName + "evm")
	toAddr := address.ExecAddress(exector)
	//name表示发给哪个执行器
	data, err := createEvmTx(cfg, &action, exector, caller, toAddr, expire, rpcLaddr, feeInt64, chainID)

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "call contract error", err)
		return
	}

	params := rpctypes.RawParm{
		Data: data,
	}

	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.SendTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func addCallContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("caller", "c", "", "the caller address")
	cmd.MarkFlagRequired("caller")

	cmd.Flags().StringP("expire", "", "120s", "transaction expire time (optional)")

	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")

	cmd.Flags().Float64P("fee", "f", 0, "contract gas fee (optional)")
	cmd.MarkFlagRequired("fee")

	cmd.Flags().StringP("exec", "e", "", "evm contract address")
	cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")

	cmd.Flags().StringP("parameter", "p", "", "tx input parameter as:approve(13nBqpmC4VaJpEZ6J6G9NUM1Y55FQvw558, 100000000)")
}

// abi命令
func abiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abi",
		Short: "EVM ABI commands",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		callAbiCmd(),
	)
	return cmd
}

func callAbiCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "send query call by abi format",
		Run:   callAbi,
	}

	cmd.Flags().StringP("address", "a", "", "evm contract address")
	cmd.MarkFlagRequired("address")

	cmd.Flags().StringP("input", "b", "", "call params (abi format) like foobar(param1,param2)")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("caller", "c", "", "the caller address")

	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	return cmd
}

func callAbi(cmd *cobra.Command, args []string) {
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

func estimateContract(cmd *cobra.Command, args []string) {
	input, _ := cmd.Flags().GetString("input")
	name, _ := cmd.Flags().GetString("exec")
	caller, _ := cmd.Flags().GetString("caller")
	amount, _ := cmd.Flags().GetFloat64("amount")
	path, _ := cmd.Flags().GetString("path")

	toAddr := address.ExecAddress("evm")
	if len(name) > 0 {
		toAddr = address.ExecAddress(name)
	}

	amountInt64 := uint64(amount*1e4) * 1e4

	abiFileName := path + name + ".abi"
	abiStr, err := readFile(abiFileName)
	if nil != err {
		_, _ = fmt.Fprintln(os.Stderr, "Can't read abi info, Pls set correct abi path and provide abi file as", abiFileName)
		return
	}

	_, packedParameter, err := evmAbi.Pack(input, abiStr, false)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to do para pack", err.Error())
		return
	}

	var estGasReq = evmtypes.EstimateEVMGasReq{To: toAddr, Para: packedParameter, Caller: caller, Amount: amountInt64}
	var estGasResp evmtypes.EstimateEVMGasResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery(rpcLaddr, "EstimateGas", &estGasReq, &estGasResp)

	if query {
		fmt.Fprintf(os.Stdout, "gas cost estimate %v\n", estGasResp.Gas)
	} else {
		fmt.Fprintln(os.Stderr, "gas cost estimate error")
	}
}

func addEstimateFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("input", "i", "", "input contract binary code")
	cmd.MarkFlagRequired("input")

	cmd.Flags().StringP("exec", "e", "", "evm contract name (like user.evm.xxxxx)")

	cmd.Flags().StringP("caller", "c", "", "the caller address")

	cmd.Flags().StringP("path", "t", "./", "abi path(optional), default to .(current directory)")

	cmd.Flags().Float64P("amount", "a", 0, "the amount transfer to the contract (optional)")
}

// 估算合约消耗
func estimateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "estimate",
		Short: "Estimate the gas cost of calling or creating a contract",
		Run:   estimateContract,
	}
	addEstimateFlags(cmd)
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
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	paraName, _ := cmd.Flags().GetString("paraName")
	caller, _ := cmd.Flags().GetString("caller")
	amount, _ := cmd.Flags().GetFloat64("amount")
	receiver, _ := cmd.Flags().GetString("receiver")
	expire, _ := cmd.Flags().GetString("expire")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	chainID, _ := cmd.Flags().GetInt32("chainID")

	amountInt64 := int64(amount*1e4) * 1e4

	r_addr, err := address.NewAddrFromString(receiver)
	if nil != err {
		_, _ = fmt.Println("Pls input correct address")
		return
	}

	exector := cfg.ExecName(paraName + "evm")
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

func getNonceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nonce",
		Short: "Get user's nonce",
		Run:   getNonce,
	}
	getNonceFlags(cmd)
	return cmd
}

func getNonceFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("addr", "a", "", "account addr")
	cmd.MarkFlagRequired("addr")
}

func getNonce(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	addr, _ := cmd.Flags().GetString("addr")
	err := address.CheckAddress(addr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, types.ErrInvalidAddress)
		return
	}

	params := evmtypes.EvmGetNonceReq{
		Address: addr,
	}

	var resp evmtypes.EvmGetNonceRespose
	query := sendQuery(rpcLaddr, "GetNonce", &params, &resp)

	if query {
		fmt.Println("Nonce=", resp.Nonce)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "Failed to get nonce!")
	}
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

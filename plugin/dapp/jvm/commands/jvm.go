package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	jvmTypes "github.com/33cn/plugin/plugin/dapp/jvm/types"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cobra"
)

// JvmCmd jvm command
func JvmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jvm",
		Short: "java contracts operation",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		jvmCheckContractNameCmd(),
		jvmCreateContractCmd(),
		jvmCallContractCmd(),
		jvmUpdateContractCmd(),
		jvmQueryContractCmd(),
		jvmShowJarCodeCmd(),
	)

	return cmd
}

func jvmShowJarCodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "code",
		Short: "show me the jar code ",
		Run:   jvmShowJarCode,
	}
	jvmAddCreateUpdateContractFlags(cmd)
	return cmd
}

// 创建jvm合约
func jvmCreateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new jvm contract",
		Run:   jvmCreateContract,
	}
	jvmAddCreateUpdateContractFlags(cmd)
	return cmd
}

func jvmAddCreateUpdateContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("contract", "x", "", "contract name same with the jar file")
	_ = cmd.MarkFlagRequired("contract")

	cmd.Flags().StringP("path", "d", "", "path where stores jar file")
	_ = cmd.MarkFlagRequired("path")
}

func jvmCreateContract(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	contractName, _ := cmd.Flags().GetString("contract")
	path, _ := cmd.Flags().GetString("path")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nameReg, _ := regexp.Compile(jvmTypes.NameRegExp)
	if !nameReg.MatchString(contractName) {
		_, _ = fmt.Fprintln(os.Stderr, "Wrong jvm contract name format, which should be a-z and 0-9 ")
		return
	}

	if len(contractName) > 16 || len(contractName) < 4 {
		_, _ = fmt.Fprintln(os.Stderr, "jvm contract name's length should be within range [4-16]")
		return
	}

	codePath := path + "/" + contractName + ".jar"
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "read code error ", err)
		return
	}

	if !strings.Contains(contractName, jvmTypes.UserJvmX) {
		contractName = jvmTypes.UserJvmX + contractName
	}

	codeInstr := common.ToHex(code)
	var createReq = &jvmTypes.CreateJvmContract{
		Name: cfg.ExecName(contractName),
		Code: codeInstr,
	}

	payLoad, err := json.Marshal(createReq)
	if err != nil {
		return
	}
	parameter := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(jvmTypes.JvmX),
		ActionName: jvmTypes.CreateJvmContractStr,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", parameter, &res)
	ctx.RunWithoutMarshal()
}

func jvmShowJarCode(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("contract")
	path, _ := cmd.Flags().GetString("path")

	nameReg, _ := regexp.Compile(jvmTypes.NameRegExp)
	if !nameReg.MatchString(contractName) {
		_, _ = fmt.Fprintln(os.Stderr, "Wrong jvm contract name format, which should be a-z and 0-9 ")
		return
	}

	if len(contractName) > 16 || len(contractName) < 4 {
		_, _ = fmt.Fprintln(os.Stderr, "jvm contract name's length should be within range [4-16]")
		return
	}

	codePath := path + "/" + contractName + ".jar"
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "read code error ", err)
		return
	}

	if !strings.Contains(contractName, jvmTypes.UserJvmX) {
		contractName = jvmTypes.UserJvmX + contractName
	}

	codeInstr := common.ToHex(code)
	var createReq = &jvmTypes.CreateJvmContract{
		Name: contractName,
		Code: codeInstr,
	}
	data, err := json.MarshalIndent(createReq, "", "    ")
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
}

// 更新jvm合约
func jvmUpdateContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an old jvm contract",
		Run:   jvmUpdateContract,
	}
	jvmAddCreateUpdateContractFlags(cmd)
	return cmd
}

func jvmUpdateContract(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	contractName, _ := cmd.Flags().GetString("contract")
	path, _ := cmd.Flags().GetString("path")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	nameReg, _ := regexp.Compile(jvmTypes.NameRegExp)
	if !nameReg.MatchString(contractName) {
		_, _ = fmt.Fprintln(os.Stderr, "Wrong jvm contract name format, which should be a-z and 0-9 ")
		return
	}

	if len(contractName) > 16 || len(contractName) < 4 {
		_, _ = fmt.Fprintln(os.Stderr, "jvm contract name's length should be within range [4-16]")
		return
	}

	codePath := path + "/" + contractName + ".jar"
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "read code error ", err)
		return
	}

	codeInstr := common.ToHex(code)
	var updateReq = &jvmTypes.UpdateJvmContract{
		Name: contractName,
		Code: codeInstr,
	}

	payLoad, err := json.Marshal(updateReq)
	if err != nil {
		return
	}
	parameter := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(jvmTypes.JvmX),
		ActionName: jvmTypes.UpdateJvmContractStr,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", parameter, &res)
	ctx.RunWithoutMarshal()
}

//运行jvm合约的查询请求
func jvmQueryContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query the jvm contract",
		Run:   jvmQueryContract,
	}
	jvmAddQueryContractFlags(cmd)
	return cmd
}

func jvmQueryContract(cmd *cobra.Command, args []string) {
	contractName, _ := cmd.Flags().GetString("exec")
	paraOneStr, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var paraParsed []string
	paraOneStr = strings.TrimSpace(paraOneStr)
	paraParsed = strings.Split(paraOneStr, " ")

	queryReq := jvmTypes.JVMQueryReq{
		Contract: contractName,
		Para:     paraParsed,
	}

	var jvmQueryResponse jvmTypes.JVMQueryResponse
	query := sendQuery4jvm(rpcLaddr, jvmTypes.QueryJvmContract, &queryReq, &jvmQueryResponse)
	if !query {
		_, _ = fmt.Fprintln(os.Stderr, "get jvm query error")
		return
	}

	if !jvmQueryResponse.Success {
		fmt.Println("Exception occurred")
		return
	}

	for _, info := range jvmQueryResponse.Result {
		fmt.Println(info)
	}
	return
}

func jvmAddQueryContractFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "jvm contract name")
	_ = cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("para", "r", "", "multiple parameter splitted by space")
	_ = cmd.MarkFlagRequired("para")
}

// 调用jvm合约
func jvmCallContractCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "Call the jvm contract",
		Run:   jvmCallContract,
	}
	jvmAddCallContractFlags(cmd)
	return cmd
}

func jvmCallContract(cmd *cobra.Command, args []string) {
	title, _ := cmd.Flags().GetString("title")
	cfg := types.GetCliSysParam(title)

	contractName, _ := cmd.Flags().GetString("exec")
	actionName, _ := cmd.Flags().GetString("action")
	paraOneStr, _ := cmd.Flags().GetString("para")
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")

	var para = []string{actionName}
	if "" != paraOneStr {
		var paraParsed []string
		paraOneStr = strings.TrimSpace(paraOneStr)
		paraParsed = strings.Split(paraOneStr, " ")
		para = append(para, paraParsed...)
	}

	if !strings.Contains(contractName, jvmTypes.UserJvmX) {
		contractName = jvmTypes.UserJvmX + contractName
	}

	var callJvmContract = &jvmTypes.CallJvmContract{
		Name:       cfg.ExecName(contractName),
		ActionData: para,
	}

	payLoad, err := json.Marshal(callJvmContract)
	if err != nil {
		return
	}

	parameter := &rpctypes.CreateTxIn{
		Execer:     cfg.ExecName(contractName),
		ActionName: jvmTypes.CallJvmContractStr,
		Payload:    payLoad,
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", parameter, &res)
	ctx.RunWithoutMarshal()
}

func jvmAddCallContractFlags(cmd *cobra.Command) {
	jvmAddCommonFlags(cmd)
	cmd.Flags().StringP("exec", "e", "", "jvm contract name,like user.jvm.xxx")
	_ = cmd.MarkFlagRequired("exec")

	cmd.Flags().StringP("action", "x", "", "contract tx name")
	_ = cmd.MarkFlagRequired("action")

	cmd.Flags().StringP("para", "r", "", "multiple contract execution parameter splitted by space(optional)")
}

func jvmAddCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("note", "n", "", "transaction note info (optional)")
}

// 检查地址是否为Jvm合约
func jvmCheckContractNameCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if jvm contract used has been used already",
		Run:   jvmCheckContractAddr,
	}
	jvmAddCheckContractAddrFlags(cmd)
	return cmd
}

func jvmAddCheckContractAddrFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("exec", "e", "", "jvm contract name, like user.jvm.xxxxx(a-z0-9, within length [4-16])")
	_ = cmd.MarkFlagRequired("exec")
}

func jvmCheckContractAddr(cmd *cobra.Command, args []string) {
	name, _ := cmd.Flags().GetString("exec")
	if bytes.Contains([]byte(name), []byte(jvmTypes.UserJvmX)) {
		name = name[len(jvmTypes.UserJvmX):]
	}

	match, _ := regexp.MatchString(jvmTypes.NameRegExp, name)
	if !match {
		_, _ = fmt.Fprintln(os.Stderr, "Wrong jvm contract name format, which should be a-z and 0-9 ")
		return
	}

	var checkAddrReq = jvmTypes.CheckJVMContractNameReq{JvmContractName: name}
	var checkAddrResp jvmTypes.CheckJVMAddrResp
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	query := sendQuery4jvm(rpcLaddr, jvmTypes.CheckNameExistsFunc, &checkAddrReq, &checkAddrResp)
	if query {
		_, _ = fmt.Fprintln(os.Stdout, checkAddrResp.ExistAlready)
	} else {
		_, _ = fmt.Fprintln(os.Stderr, "error")
	}
}

func sendQuery4jvm(rpcAddr, funcName string, request types.Message, result proto.Message) bool {
	params := rpctypes.Query4Jrpc{
		Execer:   jvmTypes.JvmX,
		FuncName: funcName,
		Payload:  types.MustPBToJSON(request),
	}

	jsonrpc, err := jsonclient.NewJSONClient(rpcAddr)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return false
	}

	err = jsonrpc.Call("Chain33.Query", params, result)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		return false
	}
	return true
}

package commands

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	wasmtypes "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/spf13/cobra"
)

// Cmd wasm 命令行
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "Wasm management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		cmdCheckContract(),
		cmdCreateContract(),
		cmdUpdateContract(),
		cmdCallContract(),
		cmdQueryContract(),
	)

	return cmd
}

func cmdCheckContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check whether the contract with the given name exists or not.",
		Run:   checkContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func cmdCreateContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "publish a new contract on chain33",
		Run:   createContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	cmd.Flags().StringP("path", "p", "", "path of the wasm file, such as ./test.wasm")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func cmdUpdateContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update an existing contract on chain33",
		Run:   updateContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	cmd.Flags().StringP("path", "p", "", "path of the wasm file, such as ./test.wasm")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("path")
	return cmd
}

func cmdCallContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "call contract on chain33",
		Run:   callContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	cmd.Flags().StringP("method", "m", "", "method name")
	cmd.Flags().IntSliceP("parameters", "p", nil, "parameters of the method which should be num")
	cmd.Flags().StringSliceP("env", "v", nil, "string parameters set to environment")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("method")
	return cmd
}

func cmdQueryContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query wasm contract",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		cmdQueryStateDB(),
		cmdQueryLocalDB(),
	)

	return cmd
}

func cmdQueryStateDB() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state",
		Short: "query kv in state db",
		Run:   queryStateDB,
	}
	cmd.Flags().StringP("contract", "n", "", "contract name")
	cmd.Flags().StringP("key", "k", "", "key of state db")
	_ = cmd.MarkFlagRequired("contract")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func cmdQueryLocalDB() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "local",
		Short: "query kv in local db",
		Run:   queryLocalDB,
	}
	cmd.Flags().StringP("contract", "n", "", "contract name")
	cmd.Flags().StringP("key", "k", "", "key of local db")
	_ = cmd.MarkFlagRequired("contract")
	_ = cmd.MarkFlagRequired("key")
	return cmd
}

func createContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	path, _ := cmd.Flags().GetString("path")

	// Read WebAssembly *.wasm file.
	code, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	payload := wasmtypes.WasmCreate{
		Name: name,
		Code: code,
	}
	params := rpctypes.CreateTxIn{
		Execer:     wasmtypes.WasmX,
		ActionName: "Create",
		Payload:    types.MustPBToJSON(&payload),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func updateContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	path, _ := cmd.Flags().GetString("path")

	// Read WebAssembly *.wasm file.
	code, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	payload := wasmtypes.WasmUpdate{
		Name: name,
		Code: code,
	}
	params := rpctypes.CreateTxIn{
		Execer:     wasmtypes.WasmX,
		ActionName: "Update",
		Payload:    types.MustPBToJSON(&payload),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func callContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	method, _ := cmd.Flags().GetString("method")
	parameters, _ := cmd.Flags().GetIntSlice("parameters")
	env, _ := cmd.Flags().GetStringSlice("env")
	var parameters2 []int64
	for _, param := range parameters {
		parameters2 = append(parameters2, int64(param))
	}

	payload := wasmtypes.WasmCall{
		Contract:   name,
		Method:     method,
		Parameters: parameters2,
		Env:        env,
	}
	params := rpctypes.CreateTxIn{
		Execer:     wasmtypes.WasmX,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(&payload),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

func checkContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")

	params := rpctypes.Query4Jrpc{
		Execer:   wasmtypes.WasmX,
		FuncName: "Check",
		Payload: types.MustPBToJSON(&wasmtypes.QueryCheckContract{
			Name: name,
		}),
	}

	var resp types.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func queryStateDB(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	contract, _ := cmd.Flags().GetString("contract")
	key, _ := cmd.Flags().GetString("key")

	params := rpctypes.Query4Jrpc{
		Execer:   wasmtypes.WasmX,
		FuncName: "QueryStateDB",
		Payload: types.MustPBToJSON(&wasmtypes.QueryContractDB{
			Contract: contract,
			Key:      key,
		}),
	}

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

func queryLocalDB(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	contract, _ := cmd.Flags().GetString("contract")
	key, _ := cmd.Flags().GetString("key")

	params := rpctypes.Query4Jrpc{
		Execer:   wasmtypes.WasmX,
		FuncName: "QueryLocalDB",
		Payload: types.MustPBToJSON(&wasmtypes.QueryContractDB{
			Contract: contract,
			Key:      key,
		}),
	}

	var resp types.ReplyString
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
}

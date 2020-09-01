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

// TokenCmd token 命令行
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wasm",
		Short: "Wasm management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		CmdCheckContract(),
		CmdCreateContract(),
		CmdCallContract(),
	)

	return cmd
}

func CmdCheckContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check whether the contract with the given name exists or not.",
		Run:   checkContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func CmdCreateContract() *cobra.Command {
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

func CmdCallContract() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "call contract on chain33",
		Run:   callContract,
	}
	cmd.Flags().StringP("name", "n", "", "contract name")
	cmd.Flags().StringP("method", "m", "", "method name")
	cmd.Flags().IntSliceP("parameters", "p", nil, "parameters of the method which should be num")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("method")
	return cmd
}

func checkContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")

	params := rpctypes.Query4Jrpc{
		Execer:   wasmtypes.WasmX,
		FuncName: "Check",
		Payload: types.MustPBToJSON(&wasmtypes.QueryCheckConract{
			Name: name,
		}),
	}

	var resp types.Reply
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &resp)
	ctx.Run()
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

func callContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	method, _ := cmd.Flags().GetString("method")
	parameters, _ := cmd.Flags().GetIntSlice("parameters")
	var parameters2 []int64
	for _, param := range parameters {
		parameters2 = append(parameters2, int64(param))
	}

	payload := wasmtypes.WasmCall{
		Contract:   name,
		Method:     method,
		Parameters: parameters2,
	}
	params := rpctypes.CreateTxIn{
		Execer:     wasmtypes.WasmX,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(&payload),
	}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, nil)
	ctx.RunWithoutMarshal()
}

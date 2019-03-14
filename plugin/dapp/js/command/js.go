// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	jsty "github.com/33cn/plugin/plugin/dapp/js/types"
	"github.com/33cn/plugin/plugin/dapp/js/types/jsproto"
	//"github.com/gojson"
	"github.com/spf13/cobra"
)

//JavaScriptCmd :
func JavaScriptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jsvm",
		Short: "Java Script VM contract",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		JavaScriptCreateCmd(),
		JavaScriptCallCmd(),
		JavaScriptQueryCmd(),
	)
	return cmd
}

// JavaScriptCreateCmd :
func JavaScriptCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create java script contract",
		Run:   createJavaScriptContract,
	}
	createJavaScriptContractFlags(cmd)
	return cmd
}

func createJavaScriptContractFlags(cmd *cobra.Command) {

	cmd.Flags().StringP("code", "c", "", "path of js file,it must always be in utf-8.")
	cmd.MarkFlagRequired("code")

	cmd.Flags().StringP("name", "n", "", "contract name")
	cmd.MarkFlagRequired("name")
}

func createJavaScriptContract(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	patch, _ := cmd.Flags().GetString("code")
	name, _ := cmd.Flags().GetString("name")

	codestr, err := ioutil.ReadFile(patch)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	create := &jsproto.Create{
		Code: string(codestr),
		Name: name,
	}

	params := &rpctypes.CreateTxIn{
		Execer:     jsty.JsX,
		ActionName: "Create",
		Payload:    types.MustPBToJSON(create),
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// JavaScriptCallCmd :
func JavaScriptCallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call",
		Short: "call java script contract",
		Run:   callJavaScript,
	}
	callJavaScriptFlags(cmd)
	return cmd
}

func callJavaScriptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "java script contract name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("funcname", "f", "", "java script contract funcname")
	cmd.MarkFlagRequired("funcname")
	cmd.Flags().StringP("args", "a", "", "json str of args")
}

func callJavaScript(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	funcname, _ := cmd.Flags().GetString("funcname")
	input, _ := cmd.Flags().GetString("args")
	call := &jsproto.Call{
		Name:     name,
		Funcname: funcname,
		Args:     input,
	}
	params := &rpctypes.CreateTxIn{
		Execer:     "user." + jsty.JsX + "." + name,
		ActionName: "Call",
		Payload:    types.MustPBToJSON(call),
	}

	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

//JavaScriptQueryCmd :
func JavaScriptQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query java script contract",
		Run:   queryJavaScript,
	}
	queryJavaScriptFlags(cmd)
	return cmd
}

func queryJavaScriptFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("name", "n", "", "java script contract name")
	cmd.MarkFlagRequired("name")
	cmd.Flags().StringP("funcname", "f", "", "java script contract funcname")
	cmd.MarkFlagRequired("funcname")
	cmd.Flags().StringP("args", "a", "", "json str of args")
}

func queryJavaScript(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	name, _ := cmd.Flags().GetString("name")
	funcname, _ := cmd.Flags().GetString("funcname")
	input, _ := cmd.Flags().GetString("args")
	var params rpctypes.Query4Jrpc
	var rep interface{}
	req := &jsproto.Call{
		Name:     name,
		Funcname: funcname,
		Args:     input,
	}

	params.Execer = jsty.JsX
	params.FuncName = "Query"
	params.Payload = types.MustPBToJSON(req)
	rep = &jsproto.QueryResult{}
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, rep)
	ctx.Run()
}

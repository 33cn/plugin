package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/33cn/chain33/rpc/jsonclient"
	echotypes "github.com/33cn/plugin/plugin/dapp/echo/types/echo"
	"github.com/spf13/cobra"
)

// EchoCmd 本执行器的命令行初始化总入口
func EchoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "echo",
		Short: "echo commandline interface",
		Args:  cobra.MinimumNArgs(1),
	}
	cmd.AddCommand(
		QueryCmd(), // 查询消息记录
		// 如果有其它命令，在这里加入
	)
	return cmd
}

// QueryCmd query 命令
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query message history",
		Run:   queryMesage,
	}
	addPingPangFlags(cmd)
	return cmd
}

func addPingPangFlags(cmd *cobra.Command) {
	// type参数，指定查询的消息类型，为uint32类型，默认值为1，通过-t参数指定
	cmd.Flags().Uint32P("type", "t", 1, "message type, 1:ping  2:pang")
	//cmd.MarkFlagRequired("type")

	// message参数，执行消息内容，为string类型，默认值为空，通过-m参数制定
	cmd.Flags().StringP("message", "m", "", "message content")
	cmd.MarkFlagRequired("message")
}

func queryMesage(cmd *cobra.Command, args []string) {
	// 这个是命令行的默认参数，可以制定调用哪一个服务地址
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	echoType, _ := cmd.Flags().GetUint32("type")
	msg, _ := cmd.Flags().GetString("message")
	// 创建RPC客户端，调用我们实现的QueryPing服务接口
	client, err := jsonclient.NewJSONClient(rpcLaddr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// 初始化查询参数结构
	var action = &echotypes.Query{Msg: msg}
	if echoType != 1 {
		fmt.Fprintln(os.Stderr, "not support")
		return
	}

	var result echotypes.QueryResult
	err = client.Call("echo.QueryPing", action, &result)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	data, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println(string(data))
}

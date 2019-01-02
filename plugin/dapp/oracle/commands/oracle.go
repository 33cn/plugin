/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	oraclety "github.com/33cn/plugin/plugin/dapp/oracle/types"
	"github.com/spf13/cobra"
)

// OracleCmd 预言机命令行
func OracleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "oracle management",
		Args:  cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(
		OraclePublishEventRawTxCmd(),
		OracleAbortEventRawTxCmd(),
		OraclePrePublishResultRawTxCmd(),
		OracleAbortPrePubResultRawTxCmd(),
		OraclePublishResultRawTxCmd(),
		OracleQueryRawTxCmd(),
	)

	return cmd
}

// OraclePublishEventRawTxCmd 发布事件
func OraclePublishEventRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish_event",
		Short: "publish a new event",
		Run:   publishEvent,
	}
	addPublishEventFlags(cmd)
	return cmd
}

func addPublishEventFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("type", "t", "", "event type, such as \"football\"")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("subtype", "s", "", "event subtype, such as \"Premier League\"")
	cmd.MarkFlagRequired("subtype")

	cmd.Flags().StringP("time", "m", "", "time that event result may be shown, such as \"2019-01-21 15:30:00\"")
	cmd.MarkFlagRequired("time")

	cmd.Flags().StringP("content", "c", "", "event content, such as '{\"team1\":\"ChelSea\", \"team2\":\"Manchester\",\"resultType\":\"score\"}'")
	cmd.MarkFlagRequired("content")

	cmd.Flags().StringP("introduction", "i", "", "event introduction, such as \"guess the sore result of football game between ChelSea and Manchester in 2019-01-21 14:00:00\"")
	cmd.MarkFlagRequired("introduction")
}

func publishEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	ty, _ := cmd.Flags().GetString("type")
	subType, _ := cmd.Flags().GetString("subtype")
	introduction, _ := cmd.Flags().GetString("introduction")
	timeString, _ := cmd.Flags().GetString("time")
	content, _ := cmd.Flags().GetString("content")

	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, timeString)
	if err != nil {
		fmt.Printf("time error:%v\n", err.Error())
		return
	}

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(oraclety.OracleX),
		ActionName: oraclety.CreateEventPublishTx,
		Payload:    []byte(fmt.Sprintf("{\"type\":\"%s\",\"subType\":\"%s\",\"time\":%d, \"content\":\"%s\", \"introduction\":\"%s\"}", ty, subType, t.Unix(), content, introduction)),
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// OracleAbortEventRawTxCmd 取消发布事件
func OracleAbortEventRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort_publish_event",
		Short: "abort publish the event",
		Run:   abortPublishEvent,
	}
	addAbortPublishEventFlags(cmd)
	return cmd
}

func addAbortPublishEventFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("eventID", "e", "", "eventID")
	cmd.MarkFlagRequired("eventID")
}

func abortPublishEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eventID, _ := cmd.Flags().GetString("eventID")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(oraclety.OracleX),
		ActionName: oraclety.CreateAbortEventPublishTx,
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\"}", eventID)),
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// OraclePrePublishResultRawTxCmd 预发布结果
func OraclePrePublishResultRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepublish_result",
		Short: "pre publish result of a new event",
		Run:   prePublishResult,
	}
	addPrePublishResultFlags(cmd)
	return cmd
}

func addPrePublishResultFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("eventID", "e", "", "eventID")
	cmd.MarkFlagRequired("eventID")

	cmd.Flags().StringP("source", "s", "", "source where result from")
	cmd.MarkFlagRequired("source")

	cmd.Flags().StringP("result", "r", "", "result string")
	cmd.MarkFlagRequired("result")
}

func prePublishResult(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eventID, _ := cmd.Flags().GetString("eventID")
	source, _ := cmd.Flags().GetString("source")
	result, _ := cmd.Flags().GetString("result")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(oraclety.OracleX),
		ActionName: oraclety.CreatePrePublishResultTx,
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\", \"source\":\"%s\", \"result\":\"%s\"}", eventID, source, result)),
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// OracleAbortPrePubResultRawTxCmd 取消预发布的事件结果
func OracleAbortPrePubResultRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "abort_result",
		Short: "abort result pre-published before",
		Run:   abortPrePubResult,
	}
	addAbortPrePubResultFlags(cmd)
	return cmd
}

func addAbortPrePubResultFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("eventID", "e", "", "eventID")
	cmd.MarkFlagRequired("eventID")
}

func abortPrePubResult(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eventID, _ := cmd.Flags().GetString("eventID")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(oraclety.OracleX),
		ActionName: oraclety.CreateAbortResultPrePublishTx,
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\"}", eventID)),
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// OraclePublishResultRawTxCmd 发布事件结果
func OraclePublishResultRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "publish_result",
		Short: "publish final result event",
		Run:   publishResult,
	}
	addPublishResultFlags(cmd)
	return cmd
}

func addPublishResultFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("eventID", "e", "", "eventID")
	cmd.MarkFlagRequired("eventID")

	cmd.Flags().StringP("source", "s", "", "source where result from")
	cmd.MarkFlagRequired("source")

	cmd.Flags().StringP("result", "r", "", "result string, such as \"{\"team1\":3, \"team2\":2}\"")
	cmd.MarkFlagRequired("result")
}

func publishResult(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	eventID, _ := cmd.Flags().GetString("eventID")
	source, _ := cmd.Flags().GetString("source")
	result, _ := cmd.Flags().GetString("result")

	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(oraclety.OracleX),
		ActionName: oraclety.CreateResultPublishTx,
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\", \"source\":\"%s\", \"result\":\"%s\"}", eventID, source, result)),
	}
	var res string
	ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.CreateTransaction", params, &res)
	ctx.RunWithoutMarshal()
}

// OracleQueryRawTxCmd 查询事件
func OracleQueryRawTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query event and event status",
		Run:   oracleQuery,
	}
	addOracleQueryFlags(cmd)
	return cmd
}

func addOracleQueryFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("last_eventID", "l", "", "last eventID, to get next page data")
	cmd.MarkFlagRequired("last_eventID")

	cmd.Flags().StringP("type", "t", "", "event type, such as \"football\"")
	cmd.MarkFlagRequired("type")

	cmd.Flags().StringP("status", "s", "", "status, number 1-5")
	cmd.MarkFlagRequired("status")

	cmd.Flags().StringP("addr", "a", "", "address of event creator")
	cmd.MarkFlagRequired("addr")

	cmd.Flags().StringP("eventIDs", "d", "", "eventIDs, used for query eventInfo, use comma between many ids")
	cmd.MarkFlagRequired("eventIDs")
}

func oracleQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, _ := cmd.Flags().GetString("rpc_laddr")
	lastEventID, _ := cmd.Flags().GetString("last_eventID")
	eventIDs, _ := cmd.Flags().GetString("eventIDs")
	ty, _ := cmd.Flags().GetString("type")
	statusStr, _ := cmd.Flags().GetString("status")
	status, _ := strconv.ParseInt(statusStr, 10, 32)
	addr, _ := cmd.Flags().GetString("addr")

	var params rpctypes.Query4Jrpc
	params.Execer = oraclety.OracleX
	req := &oraclety.QueryEventID{
		Status:  int32(status),
		Addr:    addr,
		Type:    ty,
		EventID: lastEventID,
	}
	params.Payload = types.MustPBToJSON(req)
	if eventIDs != "" {
		params.FuncName = oraclety.FuncNameQueryOracleListByIDs
		var eIDs []string
		ids := strings.Split(eventIDs, ",")
		eIDs = append(eIDs, ids...)
		req := &oraclety.QueryOracleInfos{EventID: eIDs}
		params.Payload = types.MustPBToJSON(req)
		var res oraclety.ReplyOracleStatusList
		ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else if statusStr != "" {
		if status < 0 || status > 5 {
			fmt.Println("Error: status must be 1-5")
			cmd.Help()
			return
		} else if addr != "" {
			params.FuncName = oraclety.FuncNameQueryEventIDByAddrAndStatus
		} else if ty != "" {
			params.FuncName = oraclety.FuncNameQueryEventIDByTypeAndStatus
		} else {
			params.FuncName = oraclety.FuncNameQueryEventIDByStatus
		}
		var res oraclety.ReplyEventIDs
		ctx := jsonclient.NewRPCCtx(rpcLaddr, "Chain33.Query", params, &res)
		ctx.Run()
	} else {
		fmt.Println("Error: requeres at least one of eventID, eventIDs, status")
		cmd.Help()
	}
}

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
	err := cmd.MarkFlagRequired("type")
	if err != nil {
		fmt.Printf("MarkFlagRequired type Error: %v", err)
		return
	}

	cmd.Flags().StringP("subtype", "s", "", "event subtype, such as \"Premier League\"")
	err = cmd.MarkFlagRequired("subtype")
	if err != nil {
		fmt.Printf("MarkFlagRequired subtype Error: %v", err)
		return
	}

	cmd.Flags().StringP("time", "m", "", "time that event result may be shown, such as \"2019-01-21 15:30:00\"")
	err = cmd.MarkFlagRequired("time")
	if err != nil {
		fmt.Printf("MarkFlagRequired time Error: %v", err)
		return
	}

	cmd.Flags().StringP("content", "c", "", "event content, such as '{\"team1\":\"ChelSea\", \"team2\":\"Manchester\",\"resultType\":\"score\"}'")
	err = cmd.MarkFlagRequired("content")
	if err != nil {
		fmt.Printf("MarkFlagRequired content Error: %v", err)
		return
	}

	cmd.Flags().StringP("introduction", "i", "", "event introduction, such as \"guess the sore result of football game between ChelSea and Manchester in 2019-01-21 14:00:00\"")
	err = cmd.MarkFlagRequired("introduction")
	if err != nil {
		fmt.Printf("MarkFlagRequired introduction Error: %v", err)
		return
	}
}

func publishEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("publishEvent get rpc addr Error: %v", err)
		return
	}
	ty, err := cmd.Flags().GetString("type")
	if err != nil {
		fmt.Printf("publishEvent get type Error: %v", err)
		return
	}
	subType, err := cmd.Flags().GetString("subtype")
	if err != nil {
		fmt.Printf("publishEvent get subtype Error: %v", err)
		return
	}
	introduction, err := cmd.Flags().GetString("introduction")
	if err != nil {
		fmt.Printf("publishEvent get introduction Error: %v", err)
		return
	}
	timeString, err := cmd.Flags().GetString("time")
	if err != nil {
		fmt.Printf("publishEvent get time Error: %v", err)
		return
	}
	content, err := cmd.Flags().GetString("content")
	if err != nil {
		fmt.Printf("publishEvent get content Error: %v", err)
		return
	}

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
	err := cmd.MarkFlagRequired("eventID")
	if err != nil {
		fmt.Printf("MarkFlagRequired eventID Error: %v", err)
		return
	}
}

func abortPublishEvent(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("abortPublishEvent rpc_addr Error: %v", err)
		return
	}
	eventID, err := cmd.Flags().GetString("eventID")
	if err != nil {
		fmt.Printf("abortPublishEvent eventID Error: %v", err)
		return
	}

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
	err := cmd.MarkFlagRequired("eventID")
	if err != nil {
		fmt.Printf("addPrePublishResultFlags eventID Error: %v", err)
		return
	}

	cmd.Flags().StringP("source", "s", "", "source where result from")
	err = cmd.MarkFlagRequired("source")
	if err != nil {
		fmt.Printf("addPrePublishResultFlags source Error: %v", err)
		return
	}

	cmd.Flags().StringP("result", "r", "", "result string")
	err = cmd.MarkFlagRequired("result")
	if err != nil {
		fmt.Printf("addPrePublishResultFlags result Error: %v", err)
		return
	}
}

func prePublishResult(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("prePublishResult rpc_laddr Error: %v", err)
		return
	}
	eventID, err := cmd.Flags().GetString("eventID")
	if err != nil {
		fmt.Printf("prePublishResult eventID Error: %v", err)
		return
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		fmt.Printf("prePublishResult source Error: %v", err)
		return
	}
	result, err := cmd.Flags().GetString("result")
	if err != nil {
		fmt.Printf("prePublishResult result Error: %v", err)
		return
	}

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
	err := cmd.MarkFlagRequired("eventID")
	if err != nil {
		fmt.Printf("MarkFlagRequired eventID Error: %v", err)
		return
	}
}

func abortPrePubResult(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("abortPrePubResult rpc_laddr Error: %v", err)
		return
	}
	eventID, err := cmd.Flags().GetString("eventID")
	if err != nil {
		fmt.Printf("abortPrePubResult eventID Error: %v", err)
		return
	}

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
	err := cmd.MarkFlagRequired("eventID")
	if err != nil {
		fmt.Printf("addPublishResultFlags rpc_laddr Error: %v", err)
		return
	}

	cmd.Flags().StringP("source", "s", "", "source where result from")
	err = cmd.MarkFlagRequired("source")
	if err != nil {
		fmt.Printf("addPublishResultFlags source Error: %v", err)
		return
	}

	cmd.Flags().StringP("result", "r", "", "result string, such as \"{\"team1\":3, \"team2\":2}\"")
	err = cmd.MarkFlagRequired("result")
	if err != nil {
		fmt.Printf("addPublishResultFlags result Error: %v", err)
		return
	}
}

func publishResult(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("publishResult rpc_laddr Error: %v", err)
		return
	}
	eventID, err := cmd.Flags().GetString("eventID")
	if err != nil {
		fmt.Printf("publishResult eventID Error: %v", err)
		return
	}
	source, err := cmd.Flags().GetString("source")
	if err != nil {
		fmt.Printf("publishResult source Error: %v", err)
		return
	}
	result, err := cmd.Flags().GetString("result")
	if err != nil {
		fmt.Printf("publishResult result Error: %v", err)
		return
	}

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
	err := cmd.MarkFlagRequired("last_eventID")
	if err != nil {
		fmt.Printf("MarkFlagRequired last_eventID Error: %v", err)
		return
	}

	cmd.Flags().StringP("type", "t", "", "event type, such as \"football\"")
	err = cmd.MarkFlagRequired("type")
	if err != nil {
		fmt.Printf("MarkFlagRequired type Error: %v", err)
		return
	}

	cmd.Flags().StringP("status", "s", "", "status, number 1-5")
	err = cmd.MarkFlagRequired("status")
	if err != nil {
		fmt.Printf("MarkFlagRequired status Error: %v", err)
		return
	}

	cmd.Flags().StringP("addr", "a", "", "address of event creator")
	err = cmd.MarkFlagRequired("addr")
	if err != nil {
		fmt.Printf("MarkFlagRequired addr Error: %v", err)
		return
	}

	cmd.Flags().StringP("eventIDs", "d", "", "eventIDs, used for query eventInfo, use comma between many ids")
	err = cmd.MarkFlagRequired("eventIDs")
	if err != nil {
		fmt.Printf("MarkFlagRequired eventIDs Error: %v", err)
		return
	}
}

func oracleQuery(cmd *cobra.Command, args []string) {
	rpcLaddr, err := cmd.Flags().GetString("rpc_laddr")
	if err != nil {
		fmt.Printf("oracleQuery rpc_laddr Error: %v", err)
		return
	}
	lastEventID, err := cmd.Flags().GetString("last_eventID")
	if err != nil {
		fmt.Printf("oracleQuery last_eventID Error: %v", err)
		return
	}
	eventIDs, err := cmd.Flags().GetString("eventIDs")
	if err != nil {
		fmt.Printf("oracleQuery eventIDs Error: %v", err)
		return
	}
	ty, err := cmd.Flags().GetString("type")
	if err != nil {
		fmt.Printf("oracleQuery type Error: %v", err)
		return
	}
	statusStr, err := cmd.Flags().GetString("status")
	if err != nil {
		fmt.Printf("oracleQuery status Error: %v", err)
		return
	}
	status, err := strconv.ParseInt(statusStr, 10, 32)
	if err != nil {
		fmt.Printf("oracleQuery status Error: %v", err)
		return
	}
	addr, err := cmd.Flags().GetString("addr")
	if err != nil {
		fmt.Printf("oracleQuery addr Error: %v", err)
		return
	}

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

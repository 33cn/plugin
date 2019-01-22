/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

/*
操作说明：
1.sendAddPublisher 通过manage合约增加数据发布者地址
2.sendPublishEvent 发布一个事件
3.queryEventByeventID 通过事件ID查询事件状态
4.sendAbortPublishEvent 取消事件发布
5.sendPrePublishResult 预发布事件结果
6.sendAbortPublishResult 取消事件预发布结果
7.sendPublishResult 发布事件最终结果
测试步骤：
1.需要首先在配置文件中增加超级管理员账号，例如：
  [exec.sub.manage]
  superManager=["14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"]
2.TestPublishNomal 正常发布流程
3.TestAbortPublishEvent 取消事件发布
4.TestPrePublishResult  预发布结果
5.TestAbortPublishResult 取消结果预发布
6.TestPublishResult 发布结果
7.TestQueryEventIDByStatus 按状态查询
8.TestQueryEventIDByAddrAndStatus 按地址和状态查询
9.TestQueryEventIDByTypeAndStatus 按类型和状态查询
*/

package rpc_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	oty "github.com/33cn/plugin/plugin/dapp/oracle/types"
	"github.com/stretchr/testify/assert"
)

var (
	r *rand.Rand
)

func init() {
	r = rand.New(rand.NewSource(types.Now().UnixNano()))
}

func getRPCClient(t *testing.T, mocker *testnode.Chain33Mock) *jsonclient.JSONClient {
	jrpcClient := mocker.GetJSONC()
	assert.NotNil(t, jrpcClient)
	return jrpcClient
}

func getTx(t *testing.T, hex string) *types.Transaction {
	data, err := common.FromHex(hex)
	assert.Nil(t, err)
	var tx types.Transaction
	err = types.Decode(data, &tx)
	assert.Nil(t, err)
	return &tx
}

func TestPublishNomal(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)
	// publish event
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	//pre publish result
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)
	//publish result
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
}
func TestPublishEvent(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)
	// publish event
	// abort event
	// publish event
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)
	eventoldID := eventID
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	assert.NotEqual(t, eventID, eventoldID)

	// publish event
	// pre publish result
	// publish event
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)
	eventoldID = eventID
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	assert.NotEqual(t, eventID, eventoldID)

	// publish event
	// pre publish result
	// publilsh result
	// publish event
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
	eventoldID = eventID
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	assert.NotEqual(t, eventID, eventoldID)
}

func TestAbortPublishEvent(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	// publish event
	// abort event
	// abort event
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, oty.ErrEventAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)

	// publish event
	// pre publish result
	// abort event
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, oty.ErrEventAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)

	// publish event
	// pre publish result
	// abort pre publilsh result
	// abort event
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, oty.ErrEventAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)

	// publish event
	// pre publish result
	// publilsh result
	// abort event
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, oty.ErrEventAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
}

func TestPrePublishResult(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	// publish event
	// pre publish result
	// pre publish result
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPrePublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)

	// publish event
	// abort event
	// pre publish
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPrePublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)

	// publish event
	// pre publish result
	// abort pre publish result
	// pre publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)

	//publish result
	//pre publish result
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPrePublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
}

func TestAbortPublishResult(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	//publish event
	//abort prepublish result
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, oty.ErrPrePublishAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)

	// publish event
	// abort event
	// abort pre publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, oty.ErrPrePublishAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)

	// publish event
	// pre publish result
	// abort pre publish result
	// abort pre publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPrePublished)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, oty.ErrPrePublishAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)

	// publish event
	// pre publish result
	// publish result
	// abort pre publish result
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, oty.ErrPrePublishAbortNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
}

func TestPublishResult(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	//publish event
	//publish result
	eventID := sendPublishEvent(t, jrpcClient, mocker)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)
	sendPublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventPublished)

	// publish event
	// abort event
	// publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)
	sendPublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.EventAborted)

	// publish event
	// pre publish result
	// abort pre publish result
	// publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendAbortPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)
	sendPublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultAborted)

	// publish event
	// pre publish result
	// publish result
	// publish result
	eventID = sendPublishEvent(t, jrpcClient, mocker)
	sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
	sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
	sendPublishResult(eventID, t, jrpcClient, mocker, oty.ErrResultPublishNotAllowed)
	queryEventByeventID(eventID, t, jrpcClient, oty.ResultPublished)
}

func createAllStatusEvent(t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock) {
	//total loop*5
	loop := int(oty.DefaultCount + 10)
	for i := 0; i < loop; i++ {
		//EventPublished
		eventID := sendPublishEvent(t, jrpcClient, mocker)
		assert.NotEqual(t, "", eventID)

		//EventAborted
		eventID = sendPublishEvent(t, jrpcClient, mocker)
		sendAbortPublishEvent(eventID, t, jrpcClient, mocker, nil)

		//ResultPrePublished
		eventID = sendPublishEvent(t, jrpcClient, mocker)
		sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)

		//ResultAborted
		eventID = sendPublishEvent(t, jrpcClient, mocker)
		sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
		sendAbortPublishResult(eventID, t, jrpcClient, mocker, nil)

		//ResultPublished
		eventID = sendPublishEvent(t, jrpcClient, mocker)
		sendPrePublishResult(eventID, t, jrpcClient, mocker, nil)
		sendPublishResult(eventID, t, jrpcClient, mocker, nil)
	}
}

func TestQueryEventIDByStatus(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)
	createAllStatusEvent(t, jrpcClient, mocker)
	queryEventByStatus(t, jrpcClient)
}

func TestQueryEventIDByAddrAndStatus(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	createAllStatusEvent(t, jrpcClient, mocker)

	queryEventByStatusAndAddr(t, jrpcClient)

}

func TestQueryEventIDByTypeAndStatus(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	sendAddPublisher(t, jrpcClient, mocker)

	createAllStatusEvent(t, jrpcClient, mocker)

	queryEventByStatusAndType(t, jrpcClient)
}

func sendAddPublisher(t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock) {
	//1. 调用createrawtransaction 创建交易
	req := &rpctypes.CreateTxIn{
		Execer:     "manage",
		ActionName: "Modify",
		Payload:    []byte("{\"key\":\"oracle-publish-event\",\"op\":\"add\", \"value\":\"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv\"}"),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
}

func sendPublishEvent(t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock) (eventID string) {
	ti := time.Now().AddDate(0, 0, 1)
	//1. 调用createrawtransaction 创建交易
	req := &rpctypes.CreateTxIn{
		Execer:     oty.OracleX,
		ActionName: "EventPublish",
		Payload:    []byte(fmt.Sprintf("{\"type\":\"football\",\"subType\":\"Premier League\",\"time\":%d, \"content\":\"{\\\"team%d\\\":\\\"ChelSea\\\", \\\"team%d\\\":\\\"Manchester\\\",\\\"resultType\\\":\\\"score\\\"}\", \"introduction\":\"guess the sore result of football game between ChelSea and Manchester in 2019-01-21 14:00:00\"}", ti.Unix(), r.Int()%10, r.Int()%10)),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	result, err := mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	for _, log := range result.Receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			fmt.Println(log.TyName)
			fmt.Println(string(log.Log))
			status := oty.ReceiptOracle{}
			logData, err := common.FromHex(log.RawLog)
			assert.Nil(t, err)
			err = types.Decode(logData, &status)
			assert.Nil(t, err)
			eventID = status.EventID
		}
	}
	return eventID
}

func sendAbortPublishEvent(eventID string, t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock, expectErr error) {
	req := &rpctypes.CreateTxIn{
		Execer:     oty.OracleX,
		ActionName: "EventAbort",
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\"}", eventID)),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	result, err := mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	fmt.Println(string(result.Tx.Payload))
	for _, log := range result.Receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			fmt.Println(log.TyName)
			fmt.Println(string(log.Log))
		} else if log.Ty == 1 {
			logData, err := common.FromHex(log.RawLog)
			assert.Nil(t, err)
			assert.Equal(t, expectErr.Error(), string(logData))
		}
	}
}

func sendPrePublishResult(eventID string, t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock, expectErr error) {
	req := &rpctypes.CreateTxIn{
		Execer:     oty.OracleX,
		ActionName: "ResultPrePublish",
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\", \"source\":\"sina sport\", \"result\":\"%d:%d\"}", eventID, r.Int()%10, r.Int()%10)),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	result, err := mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	for _, log := range result.Receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			fmt.Println(log.TyName)
			fmt.Println(string(log.Log))
		} else if log.Ty == 1 {
			logData, err := common.FromHex(log.RawLog)
			assert.Nil(t, err)
			assert.Equal(t, expectErr.Error(), string(logData))
		}
	}
}

func sendAbortPublishResult(eventID string, t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock, expectErr error) {
	req := &rpctypes.CreateTxIn{
		Execer:     oty.OracleX,
		ActionName: "ResultAbort",
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\"}", eventID)),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	result, err := mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	for _, log := range result.Receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			fmt.Println(log.TyName)
			fmt.Println(string(log.Log))
		} else if log.Ty == 1 {
			logData, err := common.FromHex(log.RawLog)
			assert.Nil(t, err)
			assert.Equal(t, expectErr.Error(), string(logData))
		}
	}
}

func sendPublishResult(eventID string, t *testing.T, jrpcClient *jsonclient.JSONClient, mocker *testnode.Chain33Mock, expectErr error) {
	req := &rpctypes.CreateTxIn{
		Execer:     oty.OracleX,
		ActionName: "ResultPublish",
		Payload:    []byte(fmt.Sprintf("{\"eventID\":\"%s\", \"source\":\"sina sport\", \"result\":\"%d:%d\"}", eventID, r.Int()%10, r.Int()%10)),
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetHotKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	result, err := mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	for _, log := range result.Receipt.Logs {
		if log.Ty >= oty.TyLogEventPublish && log.Ty <= oty.TyLogResultPublish {
			fmt.Println(log.TyName)
			fmt.Println(string(log.Log))
		} else if log.Ty == 1 {
			logData, err := common.FromHex(log.RawLog)
			assert.Nil(t, err)
			assert.Equal(t, expectErr.Error(), string(logData))
		}
	}
}

func queryEventByeventID(eventID string, t *testing.T, jrpcClient *jsonclient.JSONClient, expectedStatus int32) {
	//按事件ID查询事件信息
	params := rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryOracleListByIDs,
		Payload:  []byte(fmt.Sprintf("{\"eventID\":[\"%s\"]}", eventID)),
	}
	var resStatus oty.ReplyOracleStatusList
	err := jrpcClient.Call("Chain33.Query", params, &resStatus)
	assert.Nil(t, err)
	assert.Equal(t, expectedStatus, resStatus.Status[0].Status.Status)
	fmt.Println(resStatus.Status[0])

}

func queryEventByStatus(t *testing.T, jrpcClient *jsonclient.JSONClient) {
	for i := 1; i <= 5; i++ {
		//查询第一页
		params := rpctypes.Query4Jrpc{
			Execer:   oty.OracleX,
			FuncName: oty.FuncNameQueryEventIDByStatus,
			Payload:  []byte(fmt.Sprintf("{\"status\":%d,\"addr\":\"\",\"type\":\"\",\"eventID\":\"\"}", i)),
		}
		var res oty.ReplyEventIDs
		err := jrpcClient.Call("Chain33.Query", params, &res)
		assert.Nil(t, err)
		assert.EqualValues(t, oty.DefaultCount, len(res.EventID))
		lastEventID := res.EventID[oty.DefaultCount-1]
		//查询下一页
		params = rpctypes.Query4Jrpc{
			Execer:   oty.OracleX,
			FuncName: oty.FuncNameQueryEventIDByStatus,
			Payload:  []byte(fmt.Sprintf("{\"status\":%d,\"addr\":\"\",\"type\":\"\",\"eventID\":\"%s\"}", i, lastEventID)),
		}
		err = jrpcClient.Call("Chain33.Query", params, &res)
		assert.Nil(t, err)
		assert.Equal(t, 10, len(res.EventID))
		lastEventID = res.EventID[9]
		//查询最后一条后面的,应该查不到
		params = rpctypes.Query4Jrpc{
			Execer:   oty.OracleX,
			FuncName: oty.FuncNameQueryEventIDByStatus,
			Payload:  []byte(fmt.Sprintf("{\"status\":%d,\"addr\":\"\",\"type\":\"\",\"eventID\":\"%s\"}", i, lastEventID)),
		}
		err = jrpcClient.Call("Chain33.Query", params, &res)
		assert.Equal(t, types.ErrNotFound, err)
	}
}

func queryEventByStatusAndAddr(t *testing.T, jrpcClient *jsonclient.JSONClient) {
	//查询处于事件发布状态的事件
	params := rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByAddrAndStatus,
		Payload:  []byte("{\"status\":1,\"addr\":\"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv\",\"type\":\"\",\"eventID\":\"\"}"),
	}
	var res oty.ReplyEventIDs
	err := jrpcClient.Call("Chain33.Query", params, &res)
	assert.Nil(t, err)
	assert.EqualValues(t, oty.DefaultCount, len(res.EventID))
	lastEventID := res.EventID[oty.DefaultCount-1]
	//第二页
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByAddrAndStatus,
		Payload:  []byte(fmt.Sprintf("{\"status\":1,\"addr\":\"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv\",\"type\":\"\",\"eventID\":\"%s\"}", lastEventID)),
	}
	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res.EventID))
	lastEventID = res.EventID[9]

	//最后一条以后查不到
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByAddrAndStatus,
		Payload:  []byte(fmt.Sprintf("{\"status\":1,\"addr\":\"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt\",\"type\":\"\",\"eventID\":\"%s\"}", lastEventID)),
	}

	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Equal(t, types.ErrNotFound, err)

	//查询另一个地址+状态，应该查不到
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByAddrAndStatus,
		Payload:  []byte("{\"status\":1,\"addr\":\"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt\",\"type\":\"\",\"eventID\":\"\"}"),
	}
	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Equal(t, types.ErrNotFound, err)
}

func queryEventByStatusAndType(t *testing.T, jrpcClient *jsonclient.JSONClient) {
	//查询处于事件发布状态的事件
	params := rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByTypeAndStatus,
		Payload:  []byte("{\"status\":1,\"addr\":\"\",\"type\":\"football\",\"eventID\":\"\"}"),
	}
	var res oty.ReplyEventIDs
	err := jrpcClient.Call("Chain33.Query", params, &res)
	assert.Nil(t, err)
	assert.EqualValues(t, oty.DefaultCount, len(res.EventID))
	lastEventID := res.EventID[oty.DefaultCount-1]
	//第二页
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByTypeAndStatus,
		Payload:  []byte(fmt.Sprintf("{\"status\":1,\"addr\":\"\",\"type\":\"football\",\"eventID\":\"%s\"}", lastEventID)),
	}
	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Nil(t, err)
	assert.Equal(t, 10, len(res.EventID))
	lastEventID = res.EventID[9]

	//最后一条以后查不到
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByTypeAndStatus,
		Payload:  []byte(fmt.Sprintf("{\"status\":1,\"addr\":\"\",\"type\":\"football\",\"eventID\":\"%s\"}", lastEventID)),
	}

	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Equal(t, types.ErrNotFound, err)

	//查询另一种类型+状态查不到
	params = rpctypes.Query4Jrpc{
		Execer:   oty.OracleX,
		FuncName: oty.FuncNameQueryEventIDByTypeAndStatus,
		Payload:  []byte("{\"status\":1,\"addr\":\"\",\"type\":\"gambling\",\"eventID\":\"\"}"),
	}
	err = jrpcClient.Call("Chain33.Query", params, &res)
	assert.Equal(t, types.ErrNotFound, err)

}

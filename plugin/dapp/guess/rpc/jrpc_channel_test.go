// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"strings"
	"testing"

	commonlog "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/jsonclient"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/pokerbull/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func init() {
	commonlog.SetLogLevel("error")
}

func TestJRPCChannel(t *testing.T) {
	// 启动RPCmocker
	mocker := testnode.New("--notset--", nil)
	defer func() {
		mocker.Close()
	}()
	mocker.Listen()

	jrpcClient := mocker.GetJsonC()
	assert.NotNil(t, jrpcClient)

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testStartRawTxCmd},
		{fn: testContinueRawTxCmd},
		{fn: testQuitRawTxCmd},
		{fn: testQueryGameById},
		{fn: testQueryGameByAddr},
	}
	for index, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		if err == nil {
			continue
		}
		assert.NotEqualf(t, err, types.ErrActionNotSupport, "test index %d", index)
		if strings.Contains(err.Error(), "rpc: can't find") {
			assert.FailNowf(t, err.Error(), "test index %d", index)
		}
	}
}

func testStartRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.PBStartTxReq{}
	var res string
	return jrpc.Call("pokerbull.PokerBullStartTx", params, &res)
}

func testContinueRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.PBContinueTxReq{}
	var res string
	return jrpc.Call("pokerbull.PokerBullContinueTx", params, &res)
}

func testQuitRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.PBContinueTxReq{}
	var res string
	return jrpc.Call("pokerbull.PokerBullQuitTx", params, &res)
}

func testQueryGameById(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params types.Query4Cli
	req := &pty.QueryPBGameInfo{}
	params.Execer = "pokerbull"
	params.FuncName = "QueryGameById"
	params.Payload = req
	rep = &pty.ReplyPBGame{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGameByAddr(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params types.Query4Cli
	req := &pty.QueryPBGameInfo{}
	params.Execer = "pokerbull"
	params.FuncName = "QueryGameByAddr"
	params.Payload = req
	rep = &pty.PBGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

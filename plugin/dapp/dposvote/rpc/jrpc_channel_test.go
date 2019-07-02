// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"fmt"
	"testing"

	commonlog "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/guess/types"
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

	jrpcClient := mocker.GetJSONC()
	assert.NotNil(t, jrpcClient)

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testStartRawTxCmd},
		{fn: testBetRawTxCmd},
		{fn: testStopBetRawTxCmd},
		{fn: testPublishRawTxCmd},
		{fn: testAbortRawTxCmd},
	}
	for _, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		assert.Nil(t, err)
	}

	testCases = []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testQueryGameByID},
		{fn: testQueryGamesByAddr},
		{fn: testQueryGamesByStatus},
		{fn: testQueryGamesByAdminAddr},
		{fn: testQueryGamesByAddrStatus},
		{fn: testQueryGamesByAdminStatus},
		{fn: testQueryGamesByCategoryStatus},
	}
	for index, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		assert.Equal(t, err, types.ErrNotFound, fmt.Sprint(index))
	}

	testCases = []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testQueryGamesByIDs},
	}
	for index, testCase := range testCases {
		err := testCase.fn(t, jrpcClient)
		assert.Equal(t, err, nil, fmt.Sprint(index))
	}
}

func testStartRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	payload := &pty.GuessGameStart{Topic: "WorldCup Final", Options: "A:France;B:Claodia", Category: "football", MaxBetsOneTime: 100e8, MaxBetsNumber: 1000e8, DevFeeFactor: 5, DevFeeAddr: "1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7", PlatFeeFactor: 5, PlatFeeAddr: "1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pty.GuessX),
		ActionName: pty.CreateStartTx,
		Payload:    types.MustPBToJSON(payload),
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", params, &res)
}

func testBetRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	payload := &pty.GuessGameBet{GameID: "0x76dae82fcbe554d4b8df5ed1460d71dcac86a50864649a0df43e0c50b245f004", Option: "A", BetsNum: 5e8}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pty.GuessX),
		ActionName: pty.CreateBetTx,
		Payload:    types.MustPBToJSON(payload),
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", params, &res)
}

func testStopBetRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	payload := &pty.GuessGameStopBet{GameID: "0x76dae82fcbe554d4b8df5ed1460d71dcac86a50864649a0df43e0c50b245f004"}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pty.GuessX),
		ActionName: pty.CreateStopBetTx,
		Payload:    types.MustPBToJSON(payload),
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", params, &res)
}

func testPublishRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	payload := &pty.GuessGamePublish{GameID: "0x76dae82fcbe554d4b8df5ed1460d71dcac86a50864649a0df43e0c50b245f004", Result: "A"}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pty.GuessX),
		ActionName: pty.CreatePublishTx,
		Payload:    types.MustPBToJSON(payload),
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", params, &res)
}

func testAbortRawTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	payload := &pty.GuessGameAbort{GameID: "0x76dae82fcbe554d4b8df5ed1460d71dcac86a50864649a0df43e0c50b245f004"}
	params := &rpctypes.CreateTxIn{
		Execer:     types.ExecName(pty.GuessX),
		ActionName: pty.CreateAbortTx,
		Payload:    types.MustPBToJSON(payload),
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", params, &res)
}

func testQueryGameByID(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByID
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyGuessGameInfo{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByAddr(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByAddr
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByIDs(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfos{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGamesByIDs
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyGuessGameInfos{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByStatus(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByStatus
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByAdminAddr(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByAdminAddr
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByAddrStatus(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByAddrStatus
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByAdminStatus(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByAdminStatus
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testQueryGamesByCategoryStatus(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.QueryGuessGameInfo{}
	params.Execer = pty.GuessX
	params.FuncName = pty.FuncNameQueryGameByCategoryStatus
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.GuessGameRecords{}
	return jrpc.Call("Chain33.Query", params, rep)
}

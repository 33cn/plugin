// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func TestJRPCChannel(t *testing.T) {
	// 启动RPCmocker
	mocker := testnode.New("--notset--", nil)
	defer func() {
		mocker.Close()
	}()
	mocker.Listen()

	jrpcClient := mocker.GetJSONC()

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testGetTokensPreCreatedCmd},
		{fn: testGetTokensFinishCreatedCmd},
		{fn: testGetTokenAssetsCmd},
		{fn: testGetTokenBalanceCmd},
		{fn: testCreateRawTokenPreCreateTxCmd},
		{fn: testCreateRawTokenFinishTxCmd},
		{fn: testCreateRawTokenRevokeTxCmd},
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

func testGetTokensPreCreatedCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqTokens{}
	params.Execer = "token"
	params.FuncName = "GetTokens"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyTokens{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testGetTokensFinishCreatedCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := rpctypes.Query4Jrpc{
		Execer:   "token",
		FuncName: "GetTokens",
		Payload:  types.MustPBToJSON(&pty.ReqTokens{}),
	}
	var res pty.ReplyTokens
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetTokenAssetsCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqAccountTokenAssets{}
	params.Execer = "token"
	params.FuncName = "GetAccountTokenAssets"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyAccountTokenAssets{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testGetTokenBalanceCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.ReqTokenBalance{}
	var res []*rpctypes.Account
	return jrpc.Call("token.GetTokenBalance", params, &res)
}

func testCreateRawTokenPreCreateTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.TokenPreCreate{}
	return jrpc.Call("token.CreateRawTokenPreCreateTx", params, nil)
}

func testCreateRawTokenFinishTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.TokenRevokeCreate{}
	return jrpc.Call("token.CreateRawTokenRevokeTx", params, nil)
}

func testCreateRawTokenRevokeTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.TokenFinishCreate{}
	return jrpc.Call("token.CreateRawTokenFinishTx", params, nil)
}

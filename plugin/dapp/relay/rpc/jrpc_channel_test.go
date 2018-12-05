// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"strings"
	"testing"

	commonlog "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/relay/types"
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
		{fn: testShowOnesCreateRelayOrdersCmd},
		{fn: testShowOnesAcceptRelayOrdersCmd},
		{fn: testShowOnesStatusOrdersCmd},
		{fn: testShowBTCHeadHeightListCmd},
		{fn: testCreateRawRelayOrderTxCmd},
		{fn: testCreateRawRelayAcceptTxCmd},
		{fn: testCreateRawRevokeTxCmd},
		{fn: testCreateRawRelayConfirmTxCmd},
		{fn: testCreateRawRelayVerifyBTCTxCmd},
		{fn: testCreateRawRelayBtcHeaderCmd},
		{fn: testGetBTCHeaderCurHeight},
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

func testShowOnesCreateRelayOrdersCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqRelayAddrCoins{}
	params.Execer = "relay"
	params.FuncName = "GetSellRelayOrder"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyRelayOrders{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testShowOnesAcceptRelayOrdersCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqRelayAddrCoins{}
	params.Execer = "relay"
	params.FuncName = "GetBuyRelayOrder"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyRelayOrders{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testShowOnesStatusOrdersCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqRelayAddrCoins{}
	params.Execer = "relay"
	params.FuncName = "GetRelayOrderByStatus"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyRelayOrders{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testShowBTCHeadHeightListCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &pty.ReqRelayBtcHeaderHeightList{}
	params.Execer = "relay"
	params.FuncName = "GetBTCHeaderList"
	params.Payload = types.MustPBToJSON(req)
	rep = &pty.ReplyRelayBtcHeadHeightList{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testCreateRawRelayOrderTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.RelayCreate{}
	var res string
	return jrpc.Call("relay.CreateRawRelayOrderTx", params, &res)
}

func testCreateRawRelayAcceptTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.RelayAccept{}
	var res string
	return jrpc.Call("relay.CreateRawRelayAcceptTx", params, &res)
}

func testCreateRawRevokeTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.RelayRevoke{}
	var res string
	return jrpc.Call("relay.CreateRawRelayRevokeTx", params, &res)
}

func testCreateRawRelayConfirmTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.RelayConfirmTx{}
	var res string
	return jrpc.Call("relay.CreateRawRelayConfirmTx", params, &res)
}

func testCreateRawRelayVerifyBTCTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.RelayVerifyCli{}
	var res string
	return jrpc.Call("relay.CreateRawRelayVerifyBTCTx", params, &res)
}

func testCreateRawRelayBtcHeaderCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.BtcHeader{}
	var res string
	return jrpc.Call("relay.CreateRawRelaySaveBTCHeadTx", params, &res)
}

func testGetBTCHeaderCurHeight(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var params rpctypes.Query4Jrpc
	req := &pty.ReqRelayBtcHeaderHeightList{}
	js, err := types.PBToJSON(req)
	assert.Nil(t, err)
	params.Execer = "relay"
	params.FuncName = "GetBTCHeaderCurHeight"
	params.Payload = js
	rep := &pty.ReplayRelayQryBTCHeadHeight{}
	err = jrpc.Call("Chain33.Query", params, rep)
	if err != nil {
		return err
	}
	assert.Equal(t, int64(-1), rep.CurHeight)
	return nil
}

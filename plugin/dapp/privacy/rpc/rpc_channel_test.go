// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	commonlog "github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func init() {
	commonlog.SetLogLevel("error")
}

func TestRPCChannel(t *testing.T) {
	// 启动RPCmocker
	mocker := testnode.New("--notset--", nil)
	defer func() {
		mocker.Close()
	}()
	mocker.Listen()

	rpcCfg := mocker.GetCfg().RPC
	jrpcClient, err := jsonclient.NewJSONClient(fmt.Sprintf("http://%s/", rpcCfg.JrpcBindAddr))
	assert.NoError(t, err)
	assert.NotNil(t, jrpcClient)

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testShowPrivacyKey},
		{fn: testShowPrivacyAccountInfo},
		{fn: testShowPrivacyAccountSpend},
		{fn: testShowAmountsOfUTXO},
		{fn: testShowUTXOs4SpecifiedAmount},
		{fn: testCreateUTXOs},
		{fn: testListPrivacyTxs},
		{fn: testRescanUtxosOpt},
		{fn: testEnablePrivacy},
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

func testShowPrivacyKey(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var res pty.ReplyPrivacyPkPair
	params := types.ReqString{
		Data: "1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
	}
	err := jrpc.Call("privacy.ShowPrivacykey", params, &res)
	return err
}

func testShowPrivacyAccountInfo(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.ReqPPrivacyAccount{
		Addr:        "1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
		Token:       types.BTY,
		Displaymode: 1,
	}
	var res pty.ReplyPrivacyAccount
	err := jrpc.Call("privacy.ShowPrivacyAccountInfo", params, &res)
	return err
}

func testShowPrivacyAccountSpend(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.ReqPrivBal4AddrToken{
		Addr:  "1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
		Token: types.BTY,
	}
	var res pty.UTXOHaveTxHashs
	err := jrpc.Call("privacy.ShowPrivacyAccountSpend", params, &res)
	return err
}

func testShowAmountsOfUTXO(t *testing.T, jrpc *jsonclient.JSONClient) error {
	reqPrivacyToken := pty.ReqPrivacyToken{Token: types.BTY}
	var params rpctypes.Query4Jrpc
	params.Execer = pty.PrivacyX
	params.FuncName = "ShowAmountsOfUTXO"
	params.Payload = types.MustPBToJSON(&reqPrivacyToken)

	var res pty.ReplyPrivacyAmounts
	err := jrpc.Call("Chain33.Query", params, &res)
	return err
}

func testShowUTXOs4SpecifiedAmount(t *testing.T, jrpc *jsonclient.JSONClient) error {
	reqPrivacyToken := pty.ReqPrivacyToken{
		Token:  types.BTY,
		Amount: 123456,
	}
	var params rpctypes.Query4Jrpc
	params.Execer = pty.PrivacyX
	params.FuncName = "ShowUTXOs4SpecifiedAmount"
	params.Payload = types.MustPBToJSON(&reqPrivacyToken)

	var res pty.ReplyUTXOsOfAmount
	err := jrpc.Call("Chain33.Query", params, &res)
	return err
}

func testCreateUTXOs(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &pty.ReqCreateUTXOs{
		Tokenname:  types.BTY,
		Sender:     "1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
		Pubkeypair: "92fe6cfec2e19cd15f203f83b5d440ddb63d0cb71559f96dc81208d819fea85886b08f6e874fca15108d244b40f9086d8c03260d4b954a40dfb3cbe41ebc7389",
		Amount:     123456,
		Count:      12,
		Note:       "for test",
		Expire:     int64(time.Hour),
	}

	var res rpctypes.ReplyHash
	err := jrpc.Call("privacy.CreateUTXOs", params, &res)
	return err
}

func testListPrivacyTxs(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := pty.ReqPrivacyTransactionList{
		Tokenname:    types.BTY,
		SendRecvFlag: 1,
		Direction:    1,
		Count:        16,
		Address:      "13cS5G1BDN2YfGudsxRxr7X25yu6ZdgxMU",
		Seedtxhash:   []byte("0xa64296792f90f364371e0b66fdac622080ceb7b2537ff9152e189aa9e88e61bd"),
	}
	var res rpctypes.WalletTxDetails
	err := jrpc.Call("privacy.CreateUTXOs", params, &res)
	return err
}

func testRescanUtxosOpt(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var params pty.ReqRescanUtxos
	var res pty.RepRescanUtxos
	err := jrpc.Call("privacy.RescanUtxos", params, &res)
	return err
}

func testEnablePrivacy(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var params pty.ReqEnablePrivacy
	var res pty.RepEnablePrivacy
	err := jrpc.Call("privacy.EnablePrivacy", params, &res)
	return err
}

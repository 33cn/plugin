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
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/stretchr/testify/assert"
	// 注册system和plugin 包
	rpctypes "github.com/33cn/chain33/rpc/types"
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

	testCases := []struct {
		fn func(*testing.T, *jsonclient.JSONClient) error
	}{
		{fn: testCreateMultiSigAccCreateCmd},
		{fn: testCreateMultiSigAccOwnerAddCmd},
		{fn: testCreateMultiSigAccOwnerDelCmd},
		{fn: testCreateMultiSigAccOwnerModifyCmd},
		{fn: testCreateMultiSigAccOwnerReplaceCmd},
		{fn: testCreateMultiSigAccWeightModifyCmd},
		{fn: testCreateMultiSigAccDailyLimitModifyCmd},
		{fn: testCreateMultiSigConfirmTxCmd},
		{fn: testCreateMultiSigAccTransferInCmd},
		{fn: testCreateMultiSigAccTransferOutCmd},

		{fn: testGetMultiSigAccCountCmd},
		{fn: testGetMultiSigAccountsCmd},
		{fn: testGetMultiSigAccountInfoCmd},
		{fn: testGetMultiSigAccTxCountCmd},
		{fn: testGetMultiSigTxidsCmd},
		{fn: testGetMultiSigTxInfoCmd},
		{fn: testGetGetMultiSigTxConfirmedWeightCmd},
		{fn: testGetGetMultiSigAccUnSpentTodayCmd},
		{fn: testGetMultiSigAccAssetsCmd},
		{fn: testGetMultiSigAccAllAddressCmd},
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

//创建交易
func testCreateMultiSigAccCreateCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigAccCreate{}
	return jrpc.Call("multisig.MultiSigAccCreateTx", params, nil)
}
func testCreateMultiSigAccOwnerAddCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigOwnerOperate{}
	return jrpc.Call("multisig.MultiSigOwnerOperateTx", params, nil)
}
func testCreateMultiSigAccOwnerDelCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigOwnerOperate{}
	return jrpc.Call("multisig.MultiSigOwnerOperateTx", params, nil)
}
func testCreateMultiSigAccOwnerModifyCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigOwnerOperate{}
	return jrpc.Call("multisig.MultiSigOwnerOperateTx", params, nil)
}
func testCreateMultiSigAccOwnerReplaceCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigOwnerOperate{}
	return jrpc.Call("multisig.MultiSigOwnerOperateTx", params, nil)
}
func testCreateMultiSigAccWeightModifyCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigAccOperate{}
	return jrpc.Call("multisig.MultiSigAccOperateTx", params, nil)
}

func testCreateMultiSigAccDailyLimitModifyCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigAccOperate{}
	return jrpc.Call("multisig.MultiSigAccOperateTx", params, nil)
}

func testCreateMultiSigConfirmTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigConfirmTx{}
	return jrpc.Call("multisig.MultiSigConfirmTx", params, nil)
}
func testCreateMultiSigAccTransferInCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigExecTransferTo{}
	return jrpc.Call("multisig.MultiSigAccTransferInTx", params, nil)
}
func testCreateMultiSigAccTransferOutCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &mty.MultiSigExecTransferFrom{}
	return jrpc.Call("multisig.MultiSigAccTransferOutTx", params, nil)
}

//get 多重签名账户信息
func testGetMultiSigAccCountCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &rpctypes.Query4Jrpc{
		Execer:   mty.MultiSigX,
		FuncName: "MultiSigAccCount",
		Payload:  types.MustPBToJSON(&types.ReqNil{}),
	}
	var res types.Int64
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetMultiSigAccountsCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &rpctypes.Query4Jrpc{
		Execer:   mty.MultiSigX,
		FuncName: "MultiSigAccounts",
		Payload:  types.MustPBToJSON(&mty.ReqMultiSigAccs{}),
	}
	var res mty.ReplyMultiSigAccs
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetMultiSigAccountInfoCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &rpctypes.Query4Jrpc{
		Execer:   mty.MultiSigX,
		FuncName: "MultiSigAccountInfo",
		Payload:  types.MustPBToJSON(&mty.ReqMultiSigAccInfo{}),
	}
	var res mty.MultiSig
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetMultiSigAccTxCountCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &rpctypes.Query4Jrpc{
		Execer:   mty.MultiSigX,
		FuncName: "MultiSigAccTxCount",
		Payload:  types.MustPBToJSON(&mty.ReqMultiSigAccInfo{}),
	}
	var res mty.Uint64
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetMultiSigTxidsCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &rpctypes.Query4Jrpc{
		Execer:   mty.MultiSigX,
		FuncName: "MultiSigTxids",
		Payload:  types.MustPBToJSON(&mty.ReqMultiSigTxids{}),
	}
	var res mty.ReplyMultiSigTxids
	return jrpc.Call("Chain33.Query", params, &res)
}

func testGetMultiSigTxInfoCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &mty.ReqMultiSigTxInfo{}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxInfo"
	params.Payload = types.MustPBToJSON(req)
	rep = &mty.MultiSigTx{}
	return jrpc.Call("Chain33.Query", &params, rep)
}

func testGetGetMultiSigTxConfirmedWeightCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &mty.ReqMultiSigTxInfo{}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxConfirmedWeight"
	params.Payload = types.MustPBToJSON(req)
	rep = &mty.Uint64{}
	return jrpc.Call("Chain33.Query", &params, rep)
}

func testGetGetMultiSigAccUnSpentTodayCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &mty.ReqAccAssets{}
	req.IsAll = true
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccUnSpentToday"
	params.Payload = types.MustPBToJSON(req)
	rep = &mty.ReplyUnSpentAssets{}
	return jrpc.Call("Chain33.Query", &params, rep)
}

func testGetMultiSigAccAssetsCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc

	req := &mty.ReqAccAssets{}
	req.IsAll = true
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(req)
	rep = &mty.ReplyAccAssets{}
	return jrpc.Call("Chain33.Query", &params, rep)
}

func testGetMultiSigAccAllAddressCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc

	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: "14jv8WB7CwNQSnh4qo9WDBgRPRBjM5LQo6",
	}

	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAllAddress"
	params.Payload = types.MustPBToJSON(&req)
	rep = &mty.AccAddress{}
	return jrpc.Call("Chain33.Query", &params, rep)
}

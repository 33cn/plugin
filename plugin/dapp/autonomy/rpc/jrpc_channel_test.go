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
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
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
		{fn: testPropBoardTxCmd},
		{fn: testRevokeProposalBoardTxCmd},
		{fn: testVoteProposalBoardTxCmd},
		{fn: testTerminateProposalBoardTxCmd},
		{fn: testGetProposalBoardCmd},
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

func testPropBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalBoard{}
	var res string
	return jrpc.Call("autonomy.PropBoardTx", params, &res)
}

func testRevokeProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalBoard{}
	var res string
	return jrpc.Call("autonomy.RevokeProposalBoardTx", params, &res)
}

func testVoteProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalBoard{}
	var res string
	return jrpc.Call("autonomy.VoteProposalBoardTx", params, &res)
}

func testTerminateProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalBoard{}
	var res string
	return jrpc.Call("autonomy.TerminateProposalBoardTx", params, &res)
}

func testGetProposalBoardCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &auty.ReqQueryProposalBoard{}
	params.FuncName = auty.GetProposalBoard
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalBoard{}
	return jrpc.Call("Chain33.Query", params, rep)
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"testing"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

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
	req := &types.ReqString{}
	params.FuncName = auty.GetProposalBoard
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalBoard{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testListProposalBoardCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &auty.ReqQueryProposalBoard{}
	params.FuncName = auty.ListProposalBoard
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalBoard{}
	return jrpc.Call("Chain33.Query", params, rep)
}

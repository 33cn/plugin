// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"testing"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin"
)

func testPropRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalRule{}
	var res string
	return jrpc.Call("autonomy.PropRuleTx", params, &res)
}

func testRevokeProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalRule{}
	var res string
	return jrpc.Call("autonomy.RevokeProposalRuleTx", params, &res)
}

func testVoteProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalRule{}
	var res string
	return jrpc.Call("autonomy.VoteProposalRuleTx", params, &res)
}

func testTerminateProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalRule{}
	var res string
	return jrpc.Call("autonomy.TerminateProposalRuleTx", params, &res)
}

func testGetProposalRuleCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &types.ReqString{}
	params.FuncName = auty.GetProposalRule
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalRule{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testListProposalRuleCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &auty.ReqQueryProposalRule{}
	params.FuncName = auty.ListProposalRule
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalRule{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testTransferFundTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TransferFund{}
	var res string
	return jrpc.Call("autonomy.TransferFundTx", params, &res)
}

func testCommentProposalTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.Comment{}
	var res string
	return jrpc.Call("autonomy.CommentProposalTx", params, &res)
}

func testListProposalCommentCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &auty.ReqQueryProposalComment{}
	params.FuncName = auty.ListProposalComment
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalComment{}
	return jrpc.Call("Chain33.Query", params, rep)
}
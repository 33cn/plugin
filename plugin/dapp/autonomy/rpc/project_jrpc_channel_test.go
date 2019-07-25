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

func testPropProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalProject{}
	var res string
	return jrpc.Call("autonomy.PropProjectTx", params, &res)
}

func testRevokeProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalProject{}
	var res string
	return jrpc.Call("autonomy.RevokeProposalProjectTx", params, &res)
}

func testVoteProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalProject{}
	var res string
	return jrpc.Call("autonomy.VoteProposalProjectTx", params, &res)
}

func testPubVoteProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.PubVoteProposalProject{}
	var res string
	return jrpc.Call("autonomy.PubVoteProposalProjectTx", params, &res)
}

func testTerminateProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalProject{}
	var res string
	return jrpc.Call("autonomy.TerminateProposalProjectTx", params, &res)
}

func testGetProposalProjectCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &types.ReqString{}
	params.FuncName = auty.GetProposalProject
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalProject{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testListProposalProjectCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	req := &auty.ReqQueryProposalProject{}
	params.FuncName = auty.ListProposalProject
	params.Payload = types.MustPBToJSON(req)
	rep = &auty.ReplyQueryProposalProject{}
	return jrpc.Call("Chain33.Query", params, rep)
}

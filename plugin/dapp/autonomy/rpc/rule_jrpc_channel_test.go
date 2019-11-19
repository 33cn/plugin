// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"testing"

	"encoding/json"

	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin"
	auty "github.com/33cn/plugin/plugin/dapp/autonomy/types"
)

func testPropRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalRule{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "PropRule",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testRevokeProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalRule{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "RvkPropRule",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testVoteProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalRule{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "VotePropRule",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testTerminateProposalRuleTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalRule{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "TmintPropRule",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
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

func testGetActiveRuleCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	params.FuncName = auty.GetActiveRule
	params.Payload = types.MustPBToJSON(&types.ReqString{})
	rep = &auty.RuleConfig{}
	return jrpc.Call("Chain33.Query", params, rep)
}

func testTransferFundTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TransferFund{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "Transfer",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testCommentProposalTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.Comment{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "CommentProp",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
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

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

func testPropProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalProject{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "PropProject",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testRevokeProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalProject{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "RvkPropProject",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testVoteProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalProject{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "VotePropProject",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testPubVoteProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.PubVoteProposalProject{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "PubVotePropProject",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testTerminateProposalProjectTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalProject{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "TmintPropProject",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
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

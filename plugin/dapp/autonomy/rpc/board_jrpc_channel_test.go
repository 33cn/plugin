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

var chainTestCfg = types.NewChain33Config(types.GetDefaultCfgstring())

func testPropBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.ProposalBoard{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "PropBoard",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testRevokeProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.RevokeProposalBoard{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "RvkPropBoard",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testVoteProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.VoteProposalBoard{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "VotePropBoard",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
}

func testTerminateProposalBoardTxCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	params := &auty.TerminateProposalBoard{}
	payLoad, err := json.Marshal(params)
	if err != nil {
		return err
	}
	pm := &rpctypes.CreateTxIn{
		Execer:     chainTestCfg.ExecName(auty.AutonomyX),
		ActionName: "TmintPropBoard",
		Payload:    payLoad,
	}
	var res string
	return jrpc.Call("Chain33.CreateTransaction", pm, &res)
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

func testGetActiveBoardCmd(t *testing.T, jrpc *jsonclient.JSONClient) error {
	var rep interface{}
	var params rpctypes.Query4Jrpc
	params.FuncName = auty.GetActiveBoard
	params.Payload = types.MustPBToJSON(&types.ReqString{})
	rep = &auty.ActiveBoard{}
	return jrpc.Call("Chain33.Query", params, rep)
}

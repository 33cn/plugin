/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package rpc

//only load all plugin and system
import (
	"encoding/hex"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	vt "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func newGrpc(api *mocks.QueueProtocolAPI) *channelClient {
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newJrpc(api *mocks.QueueProtocolAPI) *Jrpc {
	return &Jrpc{cli: newGrpc(api)}
}

func TestChannelClient_IsSync(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	client := newGrpc(api)
	client.Init("valnode", nil, nil, nil)
	req := &types.ReqNil{}
	api.On("QueryConsensusFunc", "tendermint", "IsHealthy", req).Return(&vt.IsHealthy{IsHealthy: true}, nil)
	result, err := client.IsSync(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsHealthy)
}

func TestJrpc_IsSync(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	J := newJrpc(api)
	req := &types.ReqNil{}
	var result interface{}
	api.On("QueryConsensusFunc", "tendermint", "IsHealthy", req).Return(&vt.IsHealthy{IsHealthy: true}, nil)
	err := J.IsSync(req, &result)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestChannelClient_GetNodeInfo(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	client := newGrpc(api)
	client.Init("valnode", nil, nil, nil)
	req := &types.ReqNil{}
	node := &vt.Validator{
		Address:     []byte("aaa"),
		PubKey:      []byte("bbb"),
		VotingPower: 10,
		Accum:       -1,
	}
	set := &vt.ValidatorSet{
		Validators: []*vt.Validator{node},
		Proposer:   node,
	}
	api.On("QueryConsensusFunc", "tendermint", "NodeInfo", req).Return(set, nil)
	result, err := client.GetNodeInfo(context.Background(), req)
	assert.Nil(t, err)
	assert.EqualValues(t, set, result)
}

func TestJrpc_GetNodeInfo(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	J := newJrpc(api)
	req := &types.ReqNil{}
	var result interface{}
	node := &vt.Validator{
		Address:     []byte("aaa"),
		PubKey:      []byte("bbb"),
		VotingPower: 10,
		Accum:       -1,
	}
	set := &vt.ValidatorSet{
		Validators: []*vt.Validator{node},
		Proposer:   node,
	}
	api.On("QueryConsensusFunc", "tendermint", "NodeInfo", req).Return(set, nil)
	err := J.GetNodeInfo(req, &result)
	assert.Nil(t, err)
	assert.EqualValues(t, hex.EncodeToString(types.Encode(set)), result)
}

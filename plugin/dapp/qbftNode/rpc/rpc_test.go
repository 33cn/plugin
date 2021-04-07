/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package rpc

//only load all plugin and system
import (
	"testing"

	"strings"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	vt "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func newGrpc(api client.QueueProtocolAPI) *channelClient {
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newJrpc(api client.QueueProtocolAPI) *Jrpc {
	return &Jrpc{cli: newGrpc(api)}
}

func TestChannelClient_IsSync(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	client := newGrpc(api)
	client.Init("qbftNode", nil, nil, nil)
	req := &types.ReqNil{}
	api.On("QueryConsensusFunc", "qbft", "IsHealthy", req).Return(&vt.QbftIsHealthy{IsHealthy: true}, nil)
	result, err := client.IsSync(context.Background(), req)
	assert.Nil(t, err)
	assert.Equal(t, true, result.IsHealthy)
}

func TestJrpc_IsSync(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	J := newJrpc(api)
	req := &types.ReqNil{}
	var result interface{}
	api.On("QueryConsensusFunc", "qbft", "IsHealthy", req).Return(&vt.QbftIsHealthy{IsHealthy: true}, nil)
	err := J.IsSync(req, &result)
	assert.Nil(t, err)
	assert.Equal(t, true, result)
}

func TestChannelClient_GetNodeInfo(t *testing.T) {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	client := newGrpc(api)
	client.Init("qbftNode", nil, nil, nil)
	req := &types.ReqNil{}
	node := &vt.QbftNodeInfo{
		NodeIP:      "127.0.0.1",
		NodeID:      "001",
		Address:     "aaa",
		PubKey:      "bbb",
		VotingPower: 10,
		Accum:       -1,
	}
	set := &vt.QbftNodeInfoSet{
		Nodes: []*vt.QbftNodeInfo{node},
	}
	api.On("QueryConsensusFunc", "qbft", "NodeInfo", req).Return(set, nil)
	result, err := client.GetNodeInfo(context.Background(), req)
	assert.Nil(t, err)
	assert.EqualValues(t, set, result)
}

func TestJrpc_GetNodeInfo(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	J := newJrpc(api)
	req := &types.ReqNil{}
	var result interface{}
	node := &vt.QbftNodeInfo{
		NodeIP:      "127.0.0.1",
		NodeID:      "001",
		Address:     "aaa",
		PubKey:      "bbb",
		VotingPower: 10,
		Accum:       -1,
	}
	set := &vt.QbftNodeInfoSet{
		Nodes: []*vt.QbftNodeInfo{node},
	}
	api.On("QueryConsensusFunc", "qbft", "NodeInfo", req).Return(set, nil)
	err := J.GetNodeInfo(req, &result)
	assert.Nil(t, err)
	assert.EqualValues(t, set, result)
}

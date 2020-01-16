/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package rpc

//only load all plugin and system
import (
	"testing"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	ptestNode "github.com/33cn/plugin/plugin/dapp/paracross/testnode"
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
	cfg := types.NewChain33Config(ptestNode.DefaultConfig)
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	client := newGrpc(api)
	client.Init("paracross", nil, nil, nil)
	req := &types.ReqNil{}
	api.On("QueryConsensusFunc", "para", "IsCaughtUp", req).Return(&types.IsCaughtUp{}, nil)
	_, err := client.IsSync(context.Background(), req)
	assert.Nil(t, err)
}

func TestJrpc_IsSync(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	J := newJrpc(api)
	req := &types.ReqNil{}
	var result interface{}
	api.On("QueryConsensusFunc", "para", "IsCaughtUp", req).Return(&types.IsCaughtUp{}, nil)
	err := J.IsSync(req, &result)
	assert.Nil(t, err)
}

//TODO wait finish
//func TestRPC_CallTestNode(t *testing.T) {
//	api := new(mocks.QueueProtocolAPI)
//	cfg, sub := testnode.GetDefaultConfig()
//	// para consensus
//	cfg.Consensus.Name = "para"
//	cfg.Title="user.p.test."
//	cfg.BlockChain.IsParaChain=true
//	cfg.Store.Name="mavl"
//	mock33 := testnode.NewWithConfig(cfg, sub, api)
//	defer func() {
//		mock33.Close()
//		mock.AssertExpectationsForObjects(t, api)
//	}()
//	g := newGrpc(api)
//	g.Init(pt.GetExecName(), mock33.GetRPC(), newJrpc(api), g)
//	time.Sleep(time.Millisecond)
//	mock33.Listen()
//	time.Sleep(time.Millisecond)
//	api.On("Query", pt.GetExecName(), "GetTitle", &types.ReqString{}).Return(&pt.ParacrossStatus{Title:"test"}, nil)
//	api.On("Query", pt.GetExecName(), "ListTitles", &types.ReqNil{}).Return(&pt.RespParacrossTitles{Titles:[]*pt.ReceiptParacrossDone{&pt.ReceiptParacrossDone{Title:"test1"},&pt.ReceiptParacrossDone{Title:"test2"}}}, nil)
//	api.On("Query", pt.GetExecName(), "GetTitleHeight", &pt.ReqParacrossTitleHeight{}).Return(&pt.ReceiptParacrossDone{Height:10}, nil)
//	api.On("Query", pt.GetExecName(), "GetAssetTxResult", &types.ReqHash{}).Return(&pt.ParacrossAsset{Symbol:"test"}, nil)
//	api.On("QueryConsensusFunc", "para", "IsCaughtUp",&types.ReqNil{}).Return(&types.IsCaughtUp{Iscaughtup:true}, nil)
//	//test  jrpc
//	rpcCfg := mock33.GetCfg().RPC
//	jsonClient, err := jsonclient.NewJSONClient("http://" + rpcCfg.JrpcBindAddr + "/")
//	assert.Nil(t, err)
//	assert.NotNil(t, jsonClient)
//	var result pt.ParacrossStatus
//	err = jsonClient.Call("paracross.GetHeight", nil, &result)
//	fmt.Println(err)
//	assert.Nil(t, err)
//	assert.Equal(t, "test", result.Title)
//
//	var reply types.IsCaughtUp
//	err = jsonClient.Call("paracross.IsSync", &types.ReqNil{}, &reply)
//	assert.Nil(t, err)
//	assert.Equal(t, true, reply.Iscaughtup)
//
//	var res pt.RespParacrossTitles
//	err = jsonClient.Call("paracross.ListTitles", &types.ReqNil{}, &res)
//	assert.Nil(t, err)
//	assert.Equal(t, 2, len(res.Titles))
//
//	//test  grpc
//
//	ctx := context.Background()
//	c, err := grpc.DialContext(ctx, rpcCfg.GrpcBindAddr, grpc.WithInsecure())
//	assert.Nil(t, err)
//	assert.NotNil(t, c)
//
//	client := pt.NewParacrossClient(c)
//	issync, err := client.IsSync(ctx, &types.ReqNil{})
//	assert.Nil(t, err)
//	assert.Equal(t, true, issync.Iscaughtup)
//}

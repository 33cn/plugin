// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/version"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func newGrpc(api *mocks.QueueProtocolAPI) *channelClient {
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newJrpc(api *mocks.QueueProtocolAPI) *Jrpc {
	return &Jrpc{cli: newGrpc(api)}
}

func TestChannelClient_BindMiner(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	client := newGrpc(api)
	client.Init("ticket", nil, nil, nil)
	head := &types.Header{StateHash: []byte("sdfadasds")}
	api.On("GetLastHeader").Return(head, nil)

	var acc = &types.Account{Addr: "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt", Balance: 100000 * types.Coin}
	accv := types.Encode(acc)
	storevalue := &types.StoreReplyValue{}
	storevalue.Values = append(storevalue.Values, accv)
	api.On("StoreGet", mock.Anything).Return(storevalue, nil)

	//var addrs = make([]string, 1)
	//addrs = append(addrs, "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt")
	var in = &ty.ReqBindMiner{
		BindAddr:     "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		OriginAddr:   "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		Amount:       10000 * types.Coin,
		CheckBalance: true,
	}
	_, err := client.CreateBindMiner(context.Background(), in)
	assert.Nil(t, err)
}

func testGetTicketCountOK(t *testing.T) {
	api := &mocks.QueueProtocolAPI{}
	g := newGrpc(api)
	api.On("QueryConsensusFunc", "ticket", "GetTicketCount", mock.Anything).Return(&types.Int64{}, nil)
	data, err := g.GetTicketCount(context.Background(), nil)
	assert.Nil(t, err, "the error should be nil")
	assert.Equal(t, data, &types.Int64{})
}

func TestGetTicketCount(t *testing.T) {
	//testGetTicketCountReject(t)
	testGetTicketCountOK(t)
}

func testSetAutoMiningOK(t *testing.T) {
	api := &mocks.QueueProtocolAPI{}
	g := newGrpc(api)
	in := &ty.MinerFlag{}
	api.On("ExecWalletFunc", "ticket", "WalletAutoMiner", in).Return(&types.Reply{}, nil)
	data, err := g.SetAutoMining(context.Background(), in)
	assert.Nil(t, err, "the error should be nil")
	assert.Equal(t, data, &types.Reply{})

}

func TestSetAutoMining(t *testing.T) {
	//testSetAutoMiningReject(t)
	testSetAutoMiningOK(t)
}

func testCloseTicketsOK(t *testing.T) {
	api := &mocks.QueueProtocolAPI{}
	g := newGrpc(api)
	var in = &ty.TicketClose{}
	api.On("ExecWalletFunc", "ticket", "CloseTickets", in).Return(&types.ReplyHashes{}, nil)
	data, err := g.CloseTickets(context.Background(), in)
	assert.Nil(t, err, "the error should be nil")
	assert.Equal(t, data, &types.ReplyHashes{})
}

func TestCloseTickets(t *testing.T) {
	//testCloseTicketsReject(t)
	testCloseTicketsOK(t)
}

func TestJrpc_SetAutoMining(t *testing.T) {
	api := &mocks.QueueProtocolAPI{}
	j := newJrpc(api)
	var mingResult rpctypes.Reply
	api.On("ExecWalletFunc", mock.Anything, mock.Anything, mock.Anything).Return(&types.Reply{IsOk: true, Msg: []byte("yes")}, nil)
	err := j.SetAutoMining(&ty.MinerFlag{}, &mingResult)
	assert.Nil(t, err)
	assert.True(t, mingResult.IsOk, "SetAutoMining")
}

func TestJrpc_GetTicketCount(t *testing.T) {
	api := &mocks.QueueProtocolAPI{}
	j := newJrpc(api)

	var ticketResult int64
	var expectRet = &types.Int64{Data: 100}
	api.On("QueryConsensusFunc", mock.Anything, mock.Anything, mock.Anything).Return(expectRet, nil)
	err := j.GetTicketCount(&types.ReqNil{}, &ticketResult)
	assert.Nil(t, err)
	assert.Equal(t, expectRet.GetData(), ticketResult, "GetTicketCount")
}

func TestRPC_CallTestNode(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	cfg, sub := testnode.GetDefaultConfig()
	// 测试环境下，默认配置的共识为solo，需要修改
	cfg.Consensus.Name = "ticket"
	mock33 := testnode.NewWithConfig(cfg, sub, api)
	defer func() {
		mock33.Close()
		mock.AssertExpectationsForObjects(t, api)
	}()
	g := newGrpc(api)
	g.Init("ticket", mock33.GetRPC(), newJrpc(api), g)
	time.Sleep(time.Millisecond)
	mock33.Listen()
	time.Sleep(time.Millisecond)
	ret := &types.Reply{
		IsOk: true,
		Msg:  []byte("123"),
	}
	api.On("IsSync").Return(ret, nil)
	api.On("Version").Return(&types.VersionInfo{Chain33: version.GetVersion()}, nil)
	api.On("Close").Return()
	rpcCfg := mock33.GetCfg().RPC
	jsonClient, err := jsonclient.NewJSONClient("http://" + rpcCfg.JrpcBindAddr + "/")
	assert.Nil(t, err)
	assert.NotNil(t, jsonClient)
	var result types.VersionInfo
	err = jsonClient.Call("Chain33.Version", nil, &result)
	fmt.Println(err)
	assert.Nil(t, err)
	assert.Equal(t, version.GetVersion(), result.Chain33)

	var isSnyc bool
	err = jsonClient.Call("Chain33.IsSync", &types.ReqNil{}, &isSnyc)
	assert.Nil(t, err)
	assert.Equal(t, ret.GetIsOk(), isSnyc)

	flag := &ty.MinerFlag{Flag: 1}
	//调用ticket.AutoMiner
	api.On("ExecWalletFunc", "ticket", "WalletAutoMiner", flag).Return(&types.Reply{IsOk: true}, nil)
	var res rpctypes.Reply
	err = jsonClient.Call("ticket.SetAutoMining", flag, &res)
	assert.Nil(t, err)
	assert.Equal(t, res.IsOk, true)

	//test  grpc

	ctx := context.Background()
	c, err := grpc.DialContext(ctx, rpcCfg.GrpcBindAddr, grpc.WithInsecure())
	assert.Nil(t, err)
	assert.NotNil(t, c)

	client := types.NewChain33Client(c)
	issync, err := client.IsSync(ctx, &types.ReqNil{})
	assert.Nil(t, err)
	assert.Equal(t, true, issync.IsOk)

	client2 := ty.NewTicketClient(c)
	r, err := client2.SetAutoMining(ctx, flag)
	assert.Nil(t, err)
	assert.Equal(t, r.IsOk, true)
}

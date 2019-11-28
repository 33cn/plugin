// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/version"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func newGrpc(api client.QueueProtocolAPI) *channelClient {
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newJrpc(api client.QueueProtocolAPI) *Jrpc {
	return &Jrpc{cli: newGrpc(api)}
}

func TestChannelClient_BindMiner(t *testing.T) {
	cfg := types.NewChain33Config(cfgstring)
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	client := newGrpc(api)
	client.Init("ticket", nil, nil, nil)
	head := &types.Header{Height: 2, StateHash: []byte("sdfadasds")}
	api.On("GetLastHeader").Return(head, nil).Times(4)

	var acc = &types.Account{Addr: "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt", Balance: 100000 * types.Coin}
	accv := types.Encode(acc)
	storevalue := &types.StoreReplyValue{}
	storevalue.Values = append(storevalue.Values, accv)
	api.On("StoreGet", mock.Anything).Return(storevalue, nil).Twice()

	//var addrs = make([]string, 1)
	//addrs = append(addrs, "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt")
	var in = &ty.ReqBindMiner{
		BindAddr:     "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		OriginAddr:   "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		Amount:       10000 * types.Coin,
		CheckBalance: false,
	}
	_, err := client.CreateBindMiner(context.Background(), in)
	assert.Nil(t, err)
}

func testGetTicketCountOK(t *testing.T) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	api := &mocks.QueueProtocolAPI{}
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
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
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	// 测试环境下，默认配置的共识为solo，需要修改
	cfg.GetModuleConfig().Consensus.Name = "ticket"

	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	mock33 := testnode.NewWithConfig(cfg, api)
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

var cfgstring = `
Title="test"

[mempool]
poolCacheSize=102400
minTxFee=100000
maxTxNumPerAccount=100

[exec]
isFree=false
minExecFee=100000
enableStat=false
enableMVCC=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
dbCache=16
signType="secp256k1"
minerdisable=false
minerwhitelist=["*"]

[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
powLimitBits = "0x1f00ffff"
maxTxNumber = 10000


[mver.consensus.ticket]
coinReward = 18
coinDevFund = 12
ticketPrice = 10000
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 5
ticketWithdrawTime = 10
ticketMinerWaitTime = 2
targetTimespan = 2304
targetTimePerBlock = 16

[mver.consensus.ticket.ForkChainParamV1]
ticketPrice = 3000

[mver.consensus.ticket.ForkChainParamV2]
ticketPrice = 6000

[fork.system]
ForkChainParamV1= 10
ForkChainParamV2= 20
ForkStateDBSet=-1
ForkCheckTxDup=0
ForkBlockHash= 1
ForkMinerTime= 10
ForkTransferExec= 100000
ForkExecKey= 200000
ForkTxGroup= 200000
ForkResetTx0= 200000
ForkWithdraw= 200000
ForkExecRollback= 450000
ForkTxHeight= -1
ForkTxGroupPara= -1
ForkCheckBlockTime=1200000
ForkMultiSignAddress=1298600
ForkBlockCheck=1
ForkLocalDBAccess=0
ForkBase58AddressCheck=1800000
ForkEnableParaRegExec=0
ForkCacheDriver=0
ForkTicketFundAddrV1=-1
[fork.sub.coins]
Enable=0

[fork.sub.manage]
Enable=0
ForkManageExec=100000

[fork.sub.store-kvmvccmavl]
ForkKvmvccmavl=1

`

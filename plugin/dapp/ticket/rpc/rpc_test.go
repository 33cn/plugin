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
	ty "github.com/33cn/plugin/plugin/dapp/ticket/types"
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

	var acc = &types.Account{Addr: "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt", Balance: 100000 * types.DefaultCoinPrecision}
	accv := types.Encode(acc)
	storevalue := &types.StoreReplyValue{}
	storevalue.Values = append(storevalue.Values, accv)
	api.On("StoreGet", mock.Anything).Return(storevalue, nil).Twice()

	//var addrs = make([]string, 1)
	//addrs = append(addrs, "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt")
	var in = &ty.ReqBindMiner{
		BindAddr:     "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		OriginAddr:   "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		Amount:       10000 * types.DefaultCoinPrecision,
		CheckBalance: false,
	}
	_, err := client.CreateBindMiner(context.Background(), in)
	assert.Nil(t, err)

	var in2 = &ty.ReqBindMiner{
		BindAddr:     "",
		OriginAddr:   "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt",
		Amount:       10000 * types.DefaultCoinPrecision,
		CheckBalance: false,
	}
	_, err = client.CreateBindMiner(context.Background(), in2)
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
DisableForkCheck=true

[mempool]
poolCacheSize=102400
minTxFeeRate=100000
maxTxNumPerAccount=100

[exec]
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
ForkRootHash=1
#地址key格式化, 主要针对eth地址
ForkFormatAddressKey=0
[fork.sub.coins]
Enable=0

[fork.sub.coinsx]
Enable=0

[fork.sub.ticket]
Enable=0
ForkTicketId =0
ForkTicketVrf =0

[fork.sub.retrieve]
Enable=0
ForkRetrive=0
ForkRetriveAsset=0

[fork.sub.hashlock]
Enable=0
ForkBadRepeatSecret=0

[fork.sub.manage]
Enable=0
ForkManageExec=100000
ForkManageAutonomyEnable=-1

[fork.sub.token]
Enable=0
ForkTokenBlackList= 0
ForkBadTokenSymbol= 0
ForkTokenPrice=0
ForkTokenSymbolWithNumber=0
ForkTokenCheck= 0

[fork.sub.trade]
Enable=0
ForkTradeBuyLimit= 0
ForkTradeAsset= 0
ForkTradeID = 0
ForkTradeFixAssetDB = 0
ForkTradePrice = 0

[fork.sub.paracross]
Enable=0
ForkParacrossWithdrawFromParachain=0
ForkParacrossCommitTx=0
ForkLoopCheckCommitTxDone=0
#仅平行链适用，自共识分阶段开启，缺省是0，若对应主链高度7200000之前开启过自共识，需要重新配置此分叉，并为之前自共识设置selfConsensEnablePreContract配置项
ForkParaSelfConsStages=0
ForkParaAssetTransferRbk=0
#仅平行链适用，开启挖矿交易的高度，已有代码版本可能未在0高度开启挖矿，需要设置这个高度，新版本默认从0开启挖矿，通过交易配置分阶段奖励
ForkParaFullMinerHeight=0
ForkParaSupervision=0
ForkParaRootHash=0
ForkParaAutonomySuperGroup=-1
ForkParaFreeRegister=0

[fork.sub.evm]
Enable=0
ForkEVMState=0
ForkEVMABI=0
ForkEVMFrozen=0
ForkEVMKVHash=0
ForkEVMYoloV1=0
ForkEVMTxGroup=0

[fork.sub.evmxgo]
Enable=0

[fork.sub.blackwhite]
Enable=0
ForkBlackWhiteV2=0

[fork.sub.cert]
Enable=0

[fork.sub.guess]
Enable=0

[fork.sub.lottery]
Enable=0

[fork.sub.oracle]
Enable=0

[fork.sub.relay]
Enable=0

[fork.sub.norm]
Enable=0

[fork.sub.pokerbull]
Enable=0

[fork.sub.privacy]
Enable=0

[fork.sub.game]
Enable=0

[fork.sub.vote]
Enable=0

[fork.sub.accountmanager]
Enable=0

[fork.sub.exchange]
Enable=0
ForkFix1=0
ForkParamV1 = 0
ForkParamV2 = 0
ForkParamV3 = 0
ForkParamV4 = 0
ForkParamV5 = 0
ForkParamV6 = 0
ForkParamV7 = 0
ForkParamV8 = 0
ForkParamV9 = 0
ForkParamV10 = 0
ForkParamV11 = 0
ForkParamV12 = 0
ForkParamV13 = 0
ForkParamV14 = 0
ForkParamV15 = 0
ForkParamV16 = 0
ForkParamV17 = 0
ForkParamV18 = 0
ForkParamV19 = 0
ForkParamV20 = 0
ForkParamV21 = 0
ForkParamV22 = 0
ForkParamV23 = 0
ForkParamV24 = 0
ForkParamV25 = 0
ForkParamV26 = 0
ForkParamV27 = 0
ForkParamV28 = 0
ForkParamV29 = 0

[fork.sub.wasm]
Enable=0

[fork.sub.valnode]
Enable=0
[fork.sub.dpos]
Enable=0
[fork.sub.echo]
Enable=0
[fork.sub.storage]
Enable=0
ForkStorageLocalDB=0


[fork.sub.multisig]
Enable=0

[fork.sub.mix]
Enable=0

[fork.sub.unfreeze]
Enable=0
ForkTerminatePart=0
ForkUnfreezeIDX= 0

[fork.sub.autonomy]
Enable=0
ForkAutonomyDelRule=0
ForkAutonomyEnableItem=0

[fork.sub.jsvm]
Enable=0

[fork.sub.issuance]
Enable=0
ForkIssuanceTableUpdate=0
ForkIssuancePrecision=0

[fork.sub.collateralize]
Enable=0
ForkCollateralizeTableUpdate=0
ForkCollateralizePrecision=0

[fork.sub.qbftNode]
Enable=0

#对已有的平行链如果不是从0开始同步数据，需要设置这个kvmvccmavl的对应平行链高度的fork，如果从0开始同步，statehash会跟以前mavl的不同
[fork.sub.store-kvmvccmavl]
ForkKvmvccmavl=1

[fork.sub.zksync]
Enable=0

`

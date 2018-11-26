// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"testing"

	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	context "golang.org/x/net/context"
)

func newTestChannelClient() *channelClient {
	api := &mocks.QueueProtocolAPI{}
	return &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}
}

func newTestJrpcClient() *Jrpc {
	return &Jrpc{cli: newTestChannelClient()}
}

func testChannelClientGetTokenBalanceToken(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)

	client := &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}

	head := &types.Header{StateHash: []byte("sdfadasds")}
	api.On("GetLastHeader").Return(head, nil)

	var acc = &types.Account{Addr: "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt", Balance: 100}
	accv := types.Encode(acc)
	storevalue := &types.StoreReplyValue{}
	storevalue.Values = append(storevalue.Values, accv)
	api.On("StoreGet", mock.Anything).Return(storevalue, nil)

	var addrs = make([]string, 1)
	addrs = append(addrs, "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt")
	var in = &tokenty.ReqTokenBalance{
		Execer:      types.ExecName(tokenty.TokenX),
		Addresses:   addrs,
		TokenSymbol: "xxx",
	}
	data, err := client.GetTokenBalance(context.Background(), in)
	assert.Nil(t, err)
	accounts := data.Acc
	assert.Equal(t, acc.Addr, accounts[0].Addr)

}

func testChannelClientGetTokenBalanceOther(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	client := &channelClient{
		ChannelClient: rpctypes.ChannelClient{QueueProtocolAPI: api},
	}

	head := &types.Header{StateHash: []byte("sdfadasds")}
	api.On("GetLastHeader").Return(head, nil)

	var acc = &types.Account{Addr: "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt", Balance: 100}
	accv := types.Encode(acc)
	storevalue := &types.StoreReplyValue{}
	storevalue.Values = append(storevalue.Values, accv)
	api.On("StoreGet", mock.Anything).Return(storevalue, nil)

	var addrs = make([]string, 1)
	addrs = append(addrs, "1Jn2qu84Z1SUUosWjySggBS9pKWdAP3tZt")
	var in = &tokenty.ReqTokenBalance{
		Execer:      types.ExecName("trade"),
		Addresses:   addrs,
		TokenSymbol: "xxx",
	}
	data, err := client.GetTokenBalance(context.Background(), in)
	assert.Nil(t, err)
	accounts := data.Acc
	assert.Equal(t, acc.Addr, accounts[0].Addr)

}

func TestChannelClientGetTokenBalance(t *testing.T) {
	testChannelClientGetTokenBalanceToken(t)
	testChannelClientGetTokenBalanceOther(t)

}

func TestChannelClientCreateRawTokenPreCreateTx(t *testing.T) {
	client := newTestJrpcClient()
	var data interface{}
	err := client.CreateRawTokenPreCreateTx(nil, &data)
	assert.NotNil(t, err)
	assert.Nil(t, data)

	token := &tokenty.TokenPreCreate{
		Owner:  "asdf134",
		Symbol: "CNY",
	}
	err = client.CreateRawTokenPreCreateTx(token, &data)
	assert.NotNil(t, data)
	assert.Nil(t, err)
}

func TestChannelClientCreateRawTokenRevokeTx(t *testing.T) {
	client := newTestJrpcClient()
	var data interface{}
	err := client.CreateRawTokenRevokeTx(nil, &data)
	assert.NotNil(t, err)
	assert.Nil(t, data)

	token := &tokenty.TokenRevokeCreate{
		Owner:  "asdf134",
		Symbol: "CNY",
	}
	err = client.CreateRawTokenRevokeTx(token, &data)
	assert.NotNil(t, data)
	assert.Nil(t, err)
}

func TestChannelClientCreateRawTokenFinishTx(t *testing.T) {
	client := newTestJrpcClient()
	var data interface{}
	err := client.CreateRawTokenFinishTx(nil, &data)
	assert.NotNil(t, err)
	assert.Nil(t, data)

	token := &tokenty.TokenFinishCreate{
		Owner:  "asdf134",
		Symbol: "CNY",
	}
	err = client.CreateRawTokenFinishTx(token, &data)
	assert.NotNil(t, data)
	assert.Nil(t, err)
}

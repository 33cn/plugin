package rpc

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var cfg *types.Chain33Config

func init() {
	cfg = types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
}

func TestJrpc_CheckContract(t *testing.T) {
	api := new(mocks.QueueProtocolAPI)
	api.On("Query", types2.WasmX, "Check", mock.Anything).Return(&types.Reply{}, nil)
	jrpc := &Jrpc{
		cli: &channelClient{
			rpctypes.ChannelClient{
				QueueProtocolAPI: api,
			},
		},
	}
	var result interface{}
	err := jrpc.CheckContract(&types2.QueryCheckContract{Name: "dice"}, &result)
	assert.Nil(t, err, "CheckContract error not nil")
	assert.Equal(t, false, result.(bool))
}

func TestJrpc_CreateContract(t *testing.T) {
	jrpc := &Jrpc{}
	code, err := ioutil.ReadFile("../contracts/dice/dice.wasm")
	assert.Nil(t, err, "read wasm file error")
	var result interface{}
	err = jrpc.CreateContract(&types2.WasmCreate{Name: "dice", Code: code}, &result)
	assert.Nil(t, err, "create contract error")
	t.Log(result)
}

func TestJrpc_CallContract(t *testing.T) {
	jrpc := &Jrpc{}
	var result interface{}
	err := jrpc.CallContract(&types2.WasmCall{Contract: "dice", Method: "play"}, &result)
	assert.Nil(t, err, "call contract error")
	t.Log(result)
}

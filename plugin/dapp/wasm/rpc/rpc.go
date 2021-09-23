package rpc

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	types2 "github.com/33cn/plugin/plugin/dapp/wasm/types"
)

func (c *channelClient) check(in *types2.QueryCheckContract) (*types.Reply, error) {
	if in == nil {
		return nil, types2.ErrInvalidParam
	}
	m, err := c.Query(types2.WasmX, "Check", in)
	if err != nil {
		return nil, err
	}
	if reply, ok := m.(*types.Reply); ok {
		return reply, nil
	}
	return nil, types2.ErrUnknown
}

func (j *Jrpc) CheckContract(param *types2.QueryCheckContract, result *interface{}) error {
	res, err := j.cli.check(param)
	if err != nil {
		return err
	}
	if res != nil {
		*result = res.IsOk
	} else {
		*result = false
	}
	return nil
}

func (j *Jrpc) CreateContract(param *types2.WasmCreate, result *interface{}) error {
	if param == nil {
		return types2.ErrInvalidParam
	}
	cfg := types.LoadExecutorType(types2.WasmX).GetConfig()
	data, err := types.CallCreateTx(cfg, cfg.ExecName(types2.WasmX), "Create", param)
	if err != nil {
		return err
	}
	*result = common.ToHex(data)
	return nil
}

func (j *Jrpc) CallContract(param *types2.WasmCall, result *interface{}) error {
	if param == nil {
		return types2.ErrInvalidParam
	}
	cfg := types.LoadExecutorType(types2.WasmX).GetConfig()
	data, err := types.CallCreateTx(cfg, cfg.ExecName(types2.WasmX), "Call", param)
	if err != nil {
		return err
	}

	*result = common.ToHex(data)
	return nil
}

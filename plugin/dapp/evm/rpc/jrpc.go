// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"context"
	"encoding/hex"

	"github.com/33cn/chain33/common/address"

	"github.com/33cn/chain33/types"
	evm "github.com/33cn/plugin/plugin/dapp/evm/types"
)

// EvmCreateTx 创建Evm合约接口
func (c *Jrpc) CreateDeployTx(parm *evm.EvmContractCreateReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.CreateDeployTx(context.Background(), *parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// CreateCallTx 创建调用EVM合约交易
func (c *Jrpc) CreateCallTx(parm *evm.EvmContractCallReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.CreateCallTx(context.Background(), *parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// CreateTransferOnlyTx 创建只进行evm内部转账的交易
func (c *Jrpc) CreateTransferOnlyTx(parm *evm.EvmTransferOnlyReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}

	reply, err := c.cli.CreateTransferOnlyTx(context.Background(), *parm)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply.Data)
	return nil
}

// CalcNewContractAddr Evm部署合约的地址
func (c *Jrpc) CalcNewContractAddr(parm *evm.EvmCalcNewContractAddrReq, result *interface{}) error {
	if parm == nil {
		return types.ErrInvalidParam
	}
	newContractAddr := address.GetExecAddress(parm.Caller + parm.Txhash)
	*result = newContractAddr.String()
	return nil
}

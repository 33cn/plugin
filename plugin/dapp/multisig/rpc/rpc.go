// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

// MultiSigAccCreateTx :构造创建多重签名账户的交易
func (c *Jrpc) MultiSigAccCreateTx(param *mty.MultiSigAccCreate, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}

	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigAccCreate", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigOwnerOperateTx :构造修改多重签名账户owner属性的交易
func (c *Jrpc) MultiSigOwnerOperateTx(param *mty.MultiSigOwnerOperate, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigOwnerOperate", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigAccOperateTx :构造修改多重签名账户属性的交易
func (c *Jrpc) MultiSigAccOperateTx(param *mty.MultiSigAccOperate, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigAccOperate", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigConfirmTx :构造确认多重签名账户的交易
func (c *Jrpc) MultiSigConfirmTx(param *mty.MultiSigConfirmTx, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigConfirmTx", param)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigAccTransferInTx :构造在多重签名合约中转账到多重签名账户的交易
func (c *Jrpc) MultiSigAccTransferInTx(param *mty.MultiSigExecTransferTo, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	v := *param
	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigExecTransferTo", &v)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigAccTransferOutTx :构造在多重签名合约中从多重签名账户转账的交易
func (c *Jrpc) MultiSigAccTransferOutTx(param *mty.MultiSigExecTransferFrom, result *interface{}) error {
	if param == nil {
		return types.ErrInvalidParam
	}
	v := *param
	data, err := types.CallCreateTx(types.ExecName(mty.MultiSigX), "MultiSigExecTransferFrom", &v)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(data)
	return nil
}

// MultiSigAddresList 获取owner地址上的多重签名账户列表{multiSigAddr，owneraddr，weight}
func (c *Jrpc) MultiSigAddresList(in *types.ReqString, result *interface{}) error {
	v := *in
	data, err := c.cli.ExecWalletFunc(mty.MultiSigX, "MultiSigAddresList", &v)
	if err != nil {
		return err
	}
	ownerAttrs := data.(*mty.OwnerAttrs)
	*result = ownerAttrs
	return nil
}

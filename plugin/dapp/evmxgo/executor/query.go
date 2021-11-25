// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	evmxgotypes "github.com/33cn/plugin/plugin/dapp/evmxgo/types"
)

// Query_GetTokens 获取evmxgo合约里面的币
func (e *evmxgo) Query_GetTokens(in *evmxgotypes.ReqEvmxgos) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return e.getTokens(in)
}

// Query_GetTokenInfo 获取evmxgo合约里指定的币
func (e *evmxgo) Query_GetTokenInfo(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return e.getTokenInfo(in.GetData())
}

// Query_GetBalance 获取evmxgo合约里指定地址的币
func (e *evmxgo) Query_GetBalance(in *types.ReqBalance) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	accounts, err := e.getBalance(in)
	if err != nil {
		return nil, err
	}
	reply := evmxgotypes.ReplyAccounts{}
	reply.Accounts = accounts
	return &reply, nil
}

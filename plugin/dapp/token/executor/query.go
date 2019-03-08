// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	tokenty "github.com/33cn/plugin/plugin/dapp/token/types"
)

// Query_GetTokens 获取token
func (t *token) Query_GetTokens(in *tokenty.ReqTokens) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return t.getTokens(in)
}

// Query_GetTokenInfo 获取token信息
func (t *token) Query_GetTokenInfo(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return t.getTokenInfo(in.GetData())
}

// Query_GetTotalAmount 获取token总量
func (t *token) Query_GetTotalAmount(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	ret, err := t.getTokenInfo(in.GetData())
	if err != nil {
		return nil, err
	}
	tokenInfo, ok := ret.(*tokenty.LocalToken)
	if !ok {
		return nil, types.ErrTypeAsset
	}
	return &types.TotalAmount{
		Total: tokenInfo.Total,
	}, nil
}

// Query_GetAddrReceiverforTokens 获取token接受人数据
func (t *token) Query_GetAddrReceiverforTokens(in *tokenty.ReqAddrTokens) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return t.getAddrReceiverforTokens(in)
}

// Query_GetAccountTokenAssets 获取账户的token资产
func (t *token) Query_GetAccountTokenAssets(in *tokenty.ReqAccountTokenAssets) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return t.getAccountTokenAssets(in)
}

// Query_GetTxByToken 获取token相关交易
func (t *token) Query_GetTxByToken(in *tokenty.ReqTokenTx) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	if !cfg.SaveTokenTxList {
		return nil, types.ErrActionNotSupport
	}
	return t.getTxByToken(in)
}

// Query_GetTokenHistory 获取token 的变更历史
func (t *token) Query_GetTokenHistory(in *types.ReqString) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	rows, err := list(t.GetLocalDB(), "symbol", &tokenty.LocalLogs{Symbol: in.Data}, -1, 0)
	if err != nil {
		tokenlog.Error("Query_GetTokenHistory", "err", err)
		return nil, err
	}
	var replys tokenty.ReplyTokenLogs
	for _, row := range rows {
		o, ok := row.Data.(*tokenty.LocalLogs)
		if !ok {
			tokenlog.Error("Query_GetTokenHistory", "err", "bad row type")
			return nil, types.ErrTypeAsset
		}
		replys.Logs = append(replys.Logs, o)
	}
	return &replys, nil
}

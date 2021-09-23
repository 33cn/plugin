// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"
	"encoding/json"

	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"golang.org/x/net/context"
)

// 显示指定地址的公钥对信息，可以作为后续交易参数
func (g *channelClient) ShowPrivacyKey(ctx context.Context, in *types.ReqString) (*pty.ReplyPrivacyPkPair, error) {
	data, err := g.ExecWalletFunc(pty.PrivacyX, "ShowPrivacyKey", in)
	if err != nil {
		return nil, err
	}
	return data.(*pty.ReplyPrivacyPkPair), nil
}

// 扫描UTXO以及获取扫描UTXO后的状态
func (g *channelClient) RescanUtxos(ctx context.Context, in *pty.ReqRescanUtxos) (*pty.RepRescanUtxos, error) {
	data, err := g.ExecWalletFunc(pty.PrivacyX, "RescanUtxos", in)
	if err != nil {
		return nil, err
	}
	return data.(*pty.RepRescanUtxos), nil
}

// 使能隐私账户
func (g *channelClient) EnablePrivacy(ctx context.Context, in *pty.ReqEnablePrivacy) (*pty.RepEnablePrivacy, error) {
	data, err := g.ExecWalletFunc(pty.PrivacyX, "EnablePrivacy", in)
	if err != nil {
		return nil, err
	}
	return data.(*pty.RepEnablePrivacy), nil
}

func (g *channelClient) CreateRawTransaction(ctx context.Context, in *pty.ReqCreatePrivacyTx) (*types.Transaction, error) {
	data, err := g.ExecWalletFunc(pty.PrivacyX, "CreateTransaction", in)
	if err != nil {
		return nil, err
	}
	return data.(*types.Transaction), nil
}

// ShowPrivacyAccountInfo display privacy account information for json rpc
func (c *Jrpc) ShowPrivacyAccountInfo(in *pty.ReqPrivacyAccount, result *json.RawMessage) error {
	reply, err := c.cli.ExecWalletFunc(pty.PrivacyX, "ShowPrivacyAccountInfo", in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

/////////////////privacy///////////////

// ShowPrivacyAccountSpend display spend privacy account for json rpc
func (c *Jrpc) ShowPrivacyAccountSpend(in *pty.ReqPrivBal4AddrToken, result *json.RawMessage) error {
	if 0 == len(in.Addr) {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.ExecWalletFunc(pty.PrivacyX, "ShowPrivacyAccountSpend", in)
	if err != nil {
		log.Info("ShowPrivacyAccountSpend", "return err info", err)
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// ShowPrivacyKey display privacy key for json rpc
func (c *Jrpc) ShowPrivacyKey(in *types.ReqString, result *json.RawMessage) error {
	reply, err := c.cli.ShowPrivacyKey(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// GetPrivacyTxByAddr get all privacy transaction list by param
func (c *Jrpc) GetPrivacyTxByAddr(in *pty.ReqPrivacyTransactionList, result *interface{}) error {
	if in.Direction != 0 && in.Direction != 1 {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.ExecWalletFunc(pty.PrivacyX, "PrivacyTransactionList", in)
	if err != nil {
		return err
	}
	var txdetails rpctypes.WalletTxDetails
	cfg := c.cli.GetConfig()
	err = rpctypes.ConvertWalletTxDetailToJSON(reply.(*types.WalletTxDetails), &txdetails, cfg.GetCoinExec(), cfg.GetCoinPrecision())
	if err != nil {
		return err
	}
	*result = &txdetails
	return nil
}

// RescanUtxos rescan utxosl for json rpc
func (c *Jrpc) RescanUtxos(in *pty.ReqRescanUtxos, result *json.RawMessage) error {
	reply, err := c.cli.RescanUtxos(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// EnablePrivacy enable privacy for json rpc
func (c *Jrpc) EnablePrivacy(in *pty.ReqEnablePrivacy, result *json.RawMessage) error {
	reply, err := c.cli.EnablePrivacy(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// CreateRawTransaction create raw trasaction for json rpc
func (c *Jrpc) CreateRawTransaction(in *pty.ReqCreatePrivacyTx, result *interface{}) error {
	reply, err := c.cli.CreateRawTransaction(context.Background(), in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(types.Encode(reply))
	return err
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/json"

	"github.com/33cn/chain33/types"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"golang.org/x/net/context"
)

// 显示指定地址的公钥对信息，可以作为后续交易参数
//func (g *channelClient) ShowPrivacyKey(ctx context.Context, in *types.ReqString) (*mixTy.AccountPrivacyKey, error) {
//	data, err := g.ExecWalletFunc(mixTy.MixX, "ShowPrivacyKey", in)
//	if err != nil {
//		return nil, err
//	}
//	return data.(*mixTy.AccountPrivacyKey), nil
//}

func (g *channelClient) GetRescanStatus(ctx context.Context, in *types.ReqNil) (*types.ReqString, error) {
	log.Info("GetRescanStatus in")
	data, err := g.ExecWalletFunc(mixTy.MixX, "GetRescanStatus", in)
	if err != nil {
		return nil, err
	}
	return data.(*types.ReqString), nil
}

//
//// 扫描UTXO以及获取扫描UTXO后的状态
func (g *channelClient) RescanNotes(ctx context.Context, in *types.ReqNil) (*types.ReqString, error) {
	log.Info("RescanNotes in")
	data, err := g.ExecWalletFunc(mixTy.MixX, "RescanNotes", in)
	if err != nil {
		return nil, err
	}
	return data.(*types.ReqString), nil
}

// 使能隐私账户
func (g *channelClient) EnablePrivacy(ctx context.Context, in *types.ReqAddrs) (*mixTy.ReqEnablePrivacyRst, error) {
	data, err := g.ExecWalletFunc(mixTy.MixX, "EnablePrivacy", in)
	if err != nil {
		return nil, err
	}
	return data.(*mixTy.ReqEnablePrivacyRst), nil
}

// ShowPrivacyAccountInfo display privacy account information for json rpc
func (c *Jrpc) ShowAccountPrivacyInfo(in *types.ReqString, result *json.RawMessage) error {
	log.Info("ShowAccountPrivacyInfo jrpc in")
	reply, err := c.cli.ExecWalletFunc(mixTy.MixX, "ShowAccountPrivacyInfo", in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

/////////////////privacy///////////////

// ShowPrivacyAccountSpend display spend privacy account for json rpc
func (c *Jrpc) ShowAccountNoteInfo(in *types.ReqAddrs, result *json.RawMessage) error {
	if 0 == len(in.Addrs) {
		return types.ErrInvalidParam
	}
	reply, err := c.cli.ExecWalletFunc(mixTy.MixX, "ShowAccountNoteInfo", in)
	if err != nil {
		log.Info("ShowPrivacyAccountSpend", "return err info", err)
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

func (c *Jrpc) GetRescanStatus(in *types.ReqNil, result *json.RawMessage) error {
	reply, err := c.cli.GetRescanStatus(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// RescanUtxos rescan utxosl for json rpc
func (c *Jrpc) RescanNotes(in *types.ReqNil, result *json.RawMessage) error {
	reply, err := c.cli.RescanNotes(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

// EnablePrivacy enable privacy for json rpc
func (c *Jrpc) EnablePrivacy(in *types.ReqAddrs, result *json.RawMessage) error {
	reply, err := c.cli.EnablePrivacy(context.Background(), in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

func (c *Jrpc) EncodeSecretData(in *mixTy.SecretData, result *json.RawMessage) error {
	reply, err := c.cli.ExecWalletFunc(mixTy.MixX, "EncodeSecretData", in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

func (c *Jrpc) EncryptSecretData(in *mixTy.EncryptSecretData, result *json.RawMessage) error {
	reply, err := c.cli.ExecWalletFunc(mixTy.MixX, "EncryptSecretData", in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

func (c *Jrpc) DecryptSecretData(in *mixTy.DecryptSecretData, result *json.RawMessage) error {
	reply, err := c.cli.ExecWalletFunc(mixTy.MixX, "DecryptSecretData", in)
	if err != nil {
		return err
	}
	*result, err = types.PBToJSON(reply)
	return err
}

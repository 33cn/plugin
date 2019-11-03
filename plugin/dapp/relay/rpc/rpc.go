// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

func createRawRelayOrderTx(cfg *types.Chain33Config, parm *ty.RelayCreate) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := *parm
	return types.CallCreateTx(cfg, cfg.ExecName(ty.RelayX), "Create", &v)
}

func createRawRelayAcceptTx(cfg *types.Chain33Config, parm *ty.RelayAccept) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	return types.CallCreateTx(cfg, cfg.ExecName(ty.RelayX), "Accept", parm)
}

func createRawRelayRevokeTx(cfg *types.Chain33Config, parm *ty.RelayRevoke) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	return types.CallCreateTx(cfg, cfg.ExecName(ty.RelayX), "Revoke", parm)
}

func createRawRelayConfirmTx(cfg *types.Chain33Config, parm *ty.RelayConfirmTx) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	return types.CallCreateTx(cfg, cfg.ExecName(ty.RelayX), "ConfirmTx", parm)
}

func createRawRelaySaveBTCHeadTx(cfg *types.Chain33Config, parm *ty.BtcHeader) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	head := &ty.BtcHeader{
		Hash:         parm.Hash,
		PreviousHash: parm.PreviousHash,
		MerkleRoot:   parm.MerkleRoot,
		Height:       parm.Height,
		Version:      parm.Version,
		Time:         parm.Time,
		Nonce:        parm.Nonce,
		Bits:         parm.Bits,
		IsReset:      parm.IsReset,
	}

	v := &ty.BtcHeaders{}
	v.BtcHeader = append(v.BtcHeader, head)
	return types.CallCreateTx(cfg, cfg.ExecName(ty.RelayX), "BtcHeaders", v)
}

//CreateRawRelayOrderTx jrpc create raw relay order
func (c *Jrpc) CreateRawRelayOrderTx(in *ty.RelayCreate, result *interface{}) error {
	cfg := c.cli.GetConfig()
	reply, err := createRawRelayOrderTx(cfg, in)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply)
	return nil
}

//CreateRawRelayAcceptTx jrpc creat relay accept tx
func (c *Jrpc) CreateRawRelayAcceptTx(in *ty.RelayAccept, result *interface{}) error {
	cfg := c.cli.GetConfig()
	reply, err := createRawRelayAcceptTx(cfg, in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

//CreateRawRelayRevokeTx jrpc create revoke tx
func (c *Jrpc) CreateRawRelayRevokeTx(in *ty.RelayRevoke, result *interface{}) error {
	cfg := c.cli.GetConfig()
	reply, err := createRawRelayRevokeTx(cfg, in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

//CreateRawRelayConfirmTx jrpc create confirm tx
func (c *Jrpc) CreateRawRelayConfirmTx(in *ty.RelayConfirmTx, result *interface{}) error {
	cfg := c.cli.GetConfig()
	reply, err := createRawRelayConfirmTx(cfg, in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

//CreateRawRelaySaveBTCHeadTx jrpc save btc header
func (c *Jrpc) CreateRawRelaySaveBTCHeadTx(in *ty.BtcHeader, result *interface{}) error {
	cfg := c.cli.GetConfig()
	reply, err := createRawRelaySaveBTCHeadTx(cfg, in)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply)
	return nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc

import (
	"encoding/hex"

	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

func createRawRelayOrderTx(parm *ty.RelayCreate) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := *parm
	return types.CallCreateTx(types.ExecName(ty.RelayX), "Create", &v)
}

func createRawRelayAcceptTx(parm *ty.RelayAccept) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	return types.CallCreateTx(types.ExecName(ty.RelayX), "Accept", parm)
}

func createRawRelayRevokeTx(parm *ty.RelayRevoke) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	return types.CallCreateTx(types.ExecName(ty.RelayX), "Revoke", parm)
}

func createRawRelayConfirmTx(parm *ty.RelayConfirmTx) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}

	return types.CallCreateTx(types.ExecName(ty.RelayX), "ConfirmTx", parm)
}

func createRawRelayVerifyBTCTx(parm *ty.RelayVerifyCli) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	v := *parm
	return types.CallCreateTx(types.ExecName(ty.RelayX), "VerifyCli", &v)
}

func createRawRelaySaveBTCHeadTx(parm *ty.BtcHeader) ([]byte, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	head := &ty.BtcHeader{
		Hash:         parm.Hash,
		PreviousHash: parm.PreviousHash,
		MerkleRoot:   parm.MerkleRoot,
		Height:       parm.Height,
		IsReset:      parm.IsReset,
	}

	v := &ty.BtcHeaders{}
	v.BtcHeader = append(v.BtcHeader, head)
	return types.CallCreateTx(types.ExecName(ty.RelayX), "BtcHeaders", v)
}

func (c *jrpc) CreateRawRelayOrderTx(in *ty.RelayCreate, result *interface{}) error {
	reply, err := createRawRelayOrderTx(in)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply)
	return nil
}

func (c *jrpc) CreateRawRelayAcceptTx(in *ty.RelayAccept, result *interface{}) error {
	reply, err := createRawRelayAcceptTx(in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

func (c *jrpc) CreateRawRelayRevokeTx(in *ty.RelayRevoke, result *interface{}) error {
	reply, err := createRawRelayRevokeTx(in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

func (c *jrpc) CreateRawRelayConfirmTx(in *ty.RelayConfirmTx, result *interface{}) error {
	reply, err := createRawRelayConfirmTx(in)
	if err != nil {
		return err
	}

	*result = hex.EncodeToString(reply)
	return nil
}

func (c *jrpc) CreateRawRelayVerifyBTCTx(in *ty.RelayVerifyCli, result *interface{}) error {
	reply, err := createRawRelayVerifyBTCTx(in)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply)
	return nil
}

func (c *jrpc) CreateRawRelaySaveBTCHeadTx(in *ty.BtcHeader, result *interface{}) error {
	reply, err := createRawRelaySaveBTCHeadTx(in)
	if err != nil {
		return err
	}
	*result = hex.EncodeToString(reply)
	return nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	rty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

func (r *relay) Exec_Create(payload *rty.RelayCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.create(payload)
}

func (r *relay) Exec_Accept(payload *rty.RelayAccept, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.accept(payload)
}

func (r *relay) Exec_Revoke(payload *rty.RelayRevoke, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.relayRevoke(payload)
}

func (r *relay) Exec_ConfirmTx(payload *rty.RelayConfirmTx, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.confirmTx(payload)
}

func (r *relay) Exec_Verify(payload *rty.RelayVerify, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.verifyTx(payload)
}

func (r *relay) Exec_VerifyCli(payload *rty.RelayVerifyCli, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.verifyCmdTx(payload)
}

func (r *relay) Exec_BtcHeaders(payload *rty.BtcHeaders, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := newRelayDB(r, tx)
	return action.saveBtcHeader(payload, r.GetLocalDB())
}

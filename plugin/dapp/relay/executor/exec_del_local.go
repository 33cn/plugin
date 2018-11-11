// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	rty "github.com/33cn/plugin/plugin/dapp/relay/types"
)

func (r *relay) execDelLocal(receipt *types.ReceiptData) ([]*types.KeyValue, error) {
	for _, item := range receipt.Logs {
		switch item.Ty {
		case rty.TyLogRelayCreate,
			rty.TyLogRelayRevokeCreate,
			rty.TyLogRelayAccept,
			rty.TyLogRelayRevokeAccept,
			rty.TyLogRelayConfirmTx,
			rty.TyLogRelayFinishTx:
			var receipt rty.ReceiptRelayLog
			err := types.Decode(item.Log, &receipt)
			if err != nil {
				return nil, err
			}
			return r.getDeleteOrderKv([]byte(receipt.OrderId), item.Ty), nil
		case rty.TyLogRelayRcvBTCHead:
			var kvSet []*types.KeyValue
			var receipt = &rty.ReceiptRelayRcvBTCHeaders{}
			err := types.Decode(item.Log, receipt)
			if err != nil {
				return nil, err
			}

			btc := newBtcStore(r.GetLocalDB())
			for _, head := range receipt.Headers {
				kv, err := btc.delBlockHead(head)
				if err != nil {
					return nil, err
				}
				kvSet = append(kvSet, kv...)
			}

			kv, err := btc.delBlockLastHead(receipt)
			if err != nil {
				return nil, err
			}
			kvSet = append(kvSet, kv...)
			return kvSet, nil
		default:
			break
		}
	}
	return nil, types.ErrNotSupport
}

func (r *relay) ExecDelLocal_Create(payload *rty.RelayCreate, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil

}

func (r *relay) ExecDelLocal_Accept(payload *rty.RelayAccept, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

func (r *relay) ExecDelLocal_Revoke(payload *rty.RelayRevoke, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

func (r *relay) ExecDelLocal_ConfirmTx(payload *rty.RelayConfirmTx, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

func (r *relay) ExecDelLocal_Verify(payload *rty.RelayVerify, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

func (r *relay) ExecDelLocal_VerifyCli(payload *rty.RelayVerifyCli, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

func (r *relay) ExecDelLocal_BtcHeaders(payload *rty.BtcHeaders, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	kv, err := r.execDelLocal(receiptData)
	if err != nil {
		return nil, err
	}
	return &types.LocalDBSet{KV: kv}, nil
}

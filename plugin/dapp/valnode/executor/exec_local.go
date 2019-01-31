// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"encoding/hex"
	"errors"

	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/valnode/types"
)

// ExecLocal_Node method
func (val *ValNode) ExecLocal_Node(node *pty.ValNode, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	if len(node.GetPubKey()) == 0 {
		return nil, errors.New("validator pubkey is empty")
	}
	if node.GetPower() < 0 {
		return nil, errors.New("validator power must not be negative")
	}
	clog.Info("update validator", "pubkey", hex.EncodeToString(node.GetPubKey()), "power", node.GetPower())
	key := CalcValNodeUpdateHeightIndexKey(val.GetHeight(), index)
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(node)})
	return set, nil
}

// ExecLocal_BlockInfo method
func (val *ValNode) ExecLocal_BlockInfo(blockInfo *pty.TendermintBlockInfo, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	key := CalcValNodeBlockInfoHeightKey(val.GetHeight())
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(blockInfo)})
	return set, nil
}

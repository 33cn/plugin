// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

// ExecLocal_Node method
func (val *QbftNode) ExecLocal_Node(node *pty.QbftNode, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	clog.Info("update validator", "pubkey", node.GetPubKey(), "power", node.GetPower())
	key := CalcQbftNodeUpdateHeightIndexKey(val.GetHeight(), index)
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(node)})
	return set, nil
}

// ExecLocal_BlockInfo method
func (val *QbftNode) ExecLocal_BlockInfo(blockInfo *pty.QbftBlockInfo, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	key := CalcQbftNodeBlockInfoHeightKey(val.GetHeight())
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: types.Encode(blockInfo)})
	return set, nil
}

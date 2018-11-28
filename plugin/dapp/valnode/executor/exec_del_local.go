// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/valnode/types"
)

// ExecDelLocal_Node method
func (val *ValNode) ExecDelLocal_Node(node *pty.ValNode, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	key := CalcValNodeUpdateHeightIndexKey(val.GetHeight(), index)
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: nil})
	return set, nil
}

// ExecDelLocal_BlockInfo method
func (val *ValNode) ExecDelLocal_BlockInfo(blockInfo *pty.TendermintBlockInfo, tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	set := &types.LocalDBSet{}
	key := CalcValNodeBlockInfoHeightKey(val.GetHeight())
	set.KV = append(set.KV, &types.KeyValue{Key: key, Value: nil})
	return set, nil
}

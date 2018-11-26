// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/valnode/types"
)

// Exec_Node method
func (val *ValNode) Exec_Node(node *pty.ValNode, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: nil}
	return receipt, nil
}

// Exec_BlockInfo method
func (val *ValNode) Exec_BlockInfo(blockInfo *pty.TendermintBlockInfo, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: nil}
	return receipt, nil
}

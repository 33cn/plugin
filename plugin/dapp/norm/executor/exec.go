// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
)

// Exec_Nput Action
func (n *Norm) Exec_Nput(nput *pty.NormPut, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: nil}
	normKV := &types.KeyValue{Key: Key(nput.Key), Value: nput.Value}
	receipt.KV = append(receipt.KV, normKV)
	return receipt, nil
}

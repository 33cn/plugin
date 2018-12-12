/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

func (f *f3d) Exec_Start(payload *pt.F3DStart, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(f, tx, index)
	return action.F3dStart(payload)
}

func (f *f3d) Exec_Draw(payload *pt.F3DLuckyDraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(f, tx, index)
	return action.F3dLuckyDraw(payload)
}

func (f *f3d) Exec_Buy(payload *pt.F3DBuyKey, tx *types.Transaction, index int) (*types.Receipt, error) {
	action := NewAction(f, tx, index)
	return action.F3dBuyKey(payload)
}

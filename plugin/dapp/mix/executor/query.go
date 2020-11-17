// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

// Query_GetTitle query paracross title
func (m *Mix) Query_GetTreePath(in *mixTy.TreePathReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return CalcTreeProve(m.GetStateDB(), in.RootHash, in.LeafHash)
}

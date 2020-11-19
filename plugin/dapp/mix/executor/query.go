// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
)

// Query_GetTitle query paracross title
func (m *Mix) Query_GetTreePath(in *mixTy.TreeInfoReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	return CalcTreeProve(m.GetStateDB(), in.RootHash, in.LeafHash)
}

// Query_GetTreeList query paracross title
func (m *Mix) Query_GetLeavesList(in *mixTy.TreeInfoReq) (types.Message, error) {
	if in == nil {
		return nil, types.ErrInvalidParam
	}
	var leaves *mixTy.CommitTreeLeaves
	var err error
	if len(in.RootHash) > 0 {
		leaves, err = getCommitLeaves(m.GetStateDB(), calcCommitTreeRootLeaves(in.RootHash))
	} else {
		leaves, err = getCommitLeaves(m.GetStateDB(), calcCurrentCommitLeavesKey())
	}
	if err != nil {
		return nil, err
	}
	var resp mixTy.TreeListResp
	for _, k := range leaves.Data {
		resp.Datas = append(resp.Datas, transferFr2String(k))
	}

	return &resp, nil

}

// Query_GetRootList query  title
func (m *Mix) Query_GetRootList(in *types.ReqNil) (types.Message, error) {
	roots, err := getArchiveCommitRoots(m.GetStateDB())
	if err != nil {
		return nil, err
	}
	var resp mixTy.TreeListResp
	for _, k := range roots.Data {
		resp.Datas = append(resp.Datas, transferFr2String(k))
	}

	return &resp, nil
}

// Query_ListMixTxs 批量查询
func (m *Mix) Query_ListMixTxs(in *mixTy.MixTxListReq) (types.Message, error) {
	return m.listMixInfos(in)
}

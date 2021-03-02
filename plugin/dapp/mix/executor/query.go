// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/pkg/errors"
)

// Query_GetTreePath 根据leaf获取path 证明和roothash
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

	if len(in.RootHash) > 0 {
		l, err := getCommitLeaves(m.GetStateDB(), calcCommitTreeRootLeaves(in.RootHash))
		if err != nil {
			return nil, err
		}
		leaves.Leaves = append(leaves.Leaves, l.Leaves...)
	} else {
		status, err := getCommitTreeStatus(m.GetStateDB())
		if err != nil {
			return nil, err
		}
		for i := int32(1); i <= status.SubLeavesSeq; i++ {
			l, err := getCommitLeaves(m.GetStateDB(), calcSubLeavesKey(i))
			if err != nil {
				return nil, errors.Wrapf(err, "get leaves of seq=%d", i)
			}
			leaves.Leaves = append(leaves.Leaves, l.Leaves...)
		}

	}

	var resp mixTy.TreeListResp
	for _, k := range leaves.Leaves {
		resp.Leaves = append(resp.Leaves, mixTy.Byte2Str(k))
	}

	return &resp, nil

}

// Query_GetRootList query  title
func (m *Mix) Query_GetRootList(in *types.ReqInt) (types.Message, error) {
	var roots mixTy.CommitTreeRoots
	if in.Height > 0 {
		r, err := getArchiveRoots(m.GetStateDB(), uint64(in.Height))
		if err != nil {
			return nil, err
		}
		roots.Roots = append(roots.Roots, r.Roots...)
	} else {
		status, err := getCommitTreeStatus(m.GetStateDB())
		if err != nil {
			return nil, err
		}
		for i := int32(1); i <= status.SubLeavesSeq; i++ {
			r, err := getSubRoots(m.GetStateDB(), i)
			if err != nil {
				return nil, errors.Wrapf(err, "get roots of seq=%d", i)
			}
			roots.Roots = append(roots.Roots, r.Roots...)
		}
	}

	var resp mixTy.RootListResp
	for _, k := range roots.Roots {
		resp.Roots = append(resp.Roots, mixTy.Byte2Str(k))
	}

	return &resp, nil
}

func (m *Mix) Query_GetTreeStatus(in *types.ReqNil) (types.Message, error) {
	status, err := getCommitTreeStatus(m.GetStateDB())
	if err != nil {
		return nil, err
	}

	var resp mixTy.TreeStatusResp
	resp.SubLeavesSeq = status.SubLeavesSeq
	resp.ArchiveRootsSeq = status.ArchiveRootsSeq
	for _, h := range status.SubTrees.SubTrees {
		resp.SubTrees = append(resp.SubTrees, &mixTy.SubTreeResp{Height: h.Height, Hash: mixTy.Byte2Str(h.Hash)})
	}
	return &resp, nil
}

// Query_ListMixTxs 批量查询
func (m *Mix) Query_ListMixTxs(in *mixTy.MixTxListReq) (types.Message, error) {
	return m.listMixInfos(in)
}

// Query_PaymentPubKey 批量查询
func (m *Mix) Query_PaymentPubKey(addr *types.ReqString) (types.Message, error) {
	return GetPaymentPubKey(m.GetStateDB(), addr.Data)

}

// Query_PaymentPubKey 批量查询
func (m *Mix) Query_VerifyProof(req *mixTy.VerifyProofInfo) (types.Message, error) {
	return &types.ReqNil{}, zkProofVerify(m.GetStateDB(), req.Proof, req.Ty)
}

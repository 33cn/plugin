// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark/crypto/hash/mimc/bn256"
	"github.com/pkg/errors"
)

func makeTreeLeavesReceipt(data *mixTy.CommitTreeLeaves) *types.Receipt {
	return makeReceipt(calcCurrentCommitLeavesKey(), mixTy.TyLogCurrentCommitTreeLeaves, data)
}

func makeTreeRootsReceipt(data *mixTy.CommitTreeRoots) *types.Receipt {
	return makeReceipt(calcCurrentCommitRootsKey(), mixTy.TyLogCurrentCommitTreeRoots, data)
}

func makeCurrentTreeReceipt(leaves *mixTy.CommitTreeLeaves, roots *mixTy.CommitTreeRoots) *types.Receipt {
	r1 := makeTreeLeavesReceipt(leaves)
	r2 := makeTreeRootsReceipt(roots)
	return mergeReceipt(r1, r2)
}

func makeTreeRootLeavesReceipt(root string, data *mixTy.CommitTreeLeaves) *types.Receipt {
	return makeReceipt(calcCommitTreeRootLeaves(root), mixTy.TyLogCommitTreeRootLeaves, data)
}

func makeTreeArchiveRootsReceipt(data *mixTy.CommitTreeRoots) *types.Receipt {
	return makeReceipt(calcCommitTreeArchiveRootsKey(), mixTy.TyLogCommitTreeArchiveRoots, data)
}

func getCommitLeaves(db dbm.KV, key []byte) (*mixTy.CommitTreeLeaves, error) {
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var leaves mixTy.CommitTreeLeaves
	err = types.Decode(v, &leaves)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db verify key")
	}

	return &leaves, nil
}

func getCurrentCommitTreeLeaves(db dbm.KV) (*mixTy.CommitTreeLeaves, error) {
	return getCommitLeaves(db, calcCurrentCommitLeavesKey())
}

func getCommitRootLeaves(db dbm.KV, rootHash string) (*mixTy.CommitTreeLeaves, error) {
	return getCommitLeaves(db, calcCommitTreeRootLeaves(rootHash))
}

func getCommitTreeRoots(db dbm.KV, key []byte) (*mixTy.CommitTreeRoots, error) {
	v, err := db.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "get db")
	}
	var roots mixTy.CommitTreeRoots
	err = types.Decode(v, &roots)
	if err != nil {
		return nil, errors.Wrapf(err, "decode db verify key")
	}

	return &roots, nil
}

func getCurrentCommitTreeRoots(db dbm.KV) (*mixTy.CommitTreeRoots, error) {
	return getCommitTreeRoots(db, calcCurrentCommitRootsKey())
}

func getArchiveCommitRoots(db dbm.KV) (*mixTy.CommitTreeRoots, error) {
	return getCommitTreeRoots(db, calcCommitTreeArchiveRootsKey())
}

func getNewTree() *merkletree.Tree {
	return merkletree.New(bn256.NewMiMC("seed"))
}

func calcTreeRoot(leaves *mixTy.CommitTreeLeaves) []byte {
	tree := getNewTree()
	for _, leaf := range leaves.Leaves {
		tree.Push(leaf)
	}
	return tree.Root()

}

func getNewCommitLeaves() (*mixTy.CommitTreeLeaves, *mixTy.CommitTreeRoots) {
	leaves := &mixTy.CommitTreeLeaves{}
	roots := &mixTy.CommitTreeRoots{}

	//第一个叶子节点都是固定的"00"字节
	leaf := []byte("00")
	leaves.Leaves = append(leaves.Leaves, leaf)
	roots.Roots = append(roots.Roots, calcTreeRoot(leaves))

	return leaves, roots
}

func initNewLeaves(leaf [][]byte) *types.Receipt {
	leaves, roots := getNewCommitLeaves()
	if len(leaf) > 0 {
		leaves.Leaves = append(leaves.Leaves, leaf...)
		roots.Roots = append(roots.Roots, calcTreeRoot(leaves))
	}

	return makeCurrentTreeReceipt(leaves, roots)
}

func archiveRoots(db dbm.KV, root []byte, leaves *mixTy.CommitTreeLeaves) (*types.Receipt, error) {
	receiptRootLeaves := makeTreeRootLeavesReceipt(transferFr2String(root), leaves)

	archiveRoots, err := getArchiveCommitRoots(db)
	if isNotFound(errors.Cause(err)) {
		archiveRoots = &mixTy.CommitTreeRoots{}
		err = nil
	}
	if err != nil {
		return nil, err
	}
	archiveRoots.Roots = append(archiveRoots.Roots, root)
	receiptArch := makeTreeArchiveRootsReceipt(archiveRoots)
	return mergeReceipt(receiptRootLeaves, receiptArch), nil
}

/*
1. 增加到当前leaves 和roots
2. 如果leaves 达到最大比如1024，则按root归档leaves，并归档相应root
3. 归档同时初始化新的current leaves 和roots

*/
func pushTree(db dbm.KV, leaf [][]byte) (*types.Receipt, error) {
	leaves, err := getCurrentCommitTreeLeaves(db)
	if isNotFound(errors.Cause(err)) {
		//系统初始状态
		return initNewLeaves(leaf), nil
	}
	if err != nil {
		return nil, err
	}

	//Roots应该和Leaves保持一致，如果err 是不应该的。
	roots, err := getCurrentCommitTreeRoots(db)
	if isNotFound(errors.Cause(err)) {
		roots = &mixTy.CommitTreeRoots{}
		err = nil
	}
	if err != nil {
		return nil, err
	}

	leaves.Leaves = append(leaves.Leaves, leaf...)
	currentRoot := calcTreeRoot(leaves)
	roots.Roots = append(roots.Roots, currentRoot)
	r := makeCurrentTreeReceipt(leaves, roots)

	//归档
	if len(leaves.Leaves) >= mixTy.MaxTreeLeaves {
		receiptArch, err := archiveRoots(db, currentRoot, leaves)
		if err != nil {
			return nil, err
		}
		mergeReceipt(r, receiptArch)

		//创建新的leaves
		receiptNew := initNewLeaves(nil)
		mergeReceipt(r, receiptNew)

	}

	return r, nil
}

func checkTreeRootHashExist(db dbm.KV, hash []byte) bool {
	var roots [][]byte
	currentRoots, err := getCurrentCommitTreeRoots(db)
	if err == nil {
		roots = append(roots, currentRoots.Roots...)
	}

	archiveRoots, err := getArchiveCommitRoots(db)
	if err == nil {
		roots = append(roots, archiveRoots.Roots...)
	}

	for _, k := range roots {
		if bytes.Equal(k, hash) {
			return true
		}
	}
	return false

}

func getProveData(targetLeaf []byte, leaves [][]byte) (*mixTy.CommitTreeProve, error) {
	index := 0
	found := false
	for i, key := range leaves {
		if bytes.Equal(key, targetLeaf) {
			index = i
			found = true
			break
		}
	}
	//index=0的leaf是占位"00"，不会和leaf相等
	if !found {
		return nil, mixTy.ErrLeafNotFound
	}

	tree := getNewTree()
	tree.SetIndex(uint64(index))
	for _, key := range leaves {
		tree.Push(key)
	}
	root, proofSet, proofIndex, num := tree.Prove()
	var prove mixTy.CommitTreeProve

	prove.RootHash = transferFr2String(root)
	prove.ProofIndex = uint32(proofIndex)
	prove.NumLeaves = uint32(num)
	//set[0] 是targetLeaf
	for _, s := range proofSet {
		prove.ProofSet = append(prove.ProofSet, transferFr2String(s))
	}

	helpers := merkletree.GenerateProofHelper(proofSet, proofIndex, num)

	for _, i := range helpers {
		prove.Helpers = append(prove.Helpers, uint32(i))
	}

	return &prove, nil

}

//1. 首先在当前tree查找
//2. 如果提供了rootHash,则根据roothash+leaf查找，否则全局遍历查找
func CalcTreeProve(db dbm.KV, rootHash, leaf string) (*mixTy.CommitTreeProve, error) {
	if len(leaf) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "leaf is null")
	}
	leaves, err := getCurrentCommitTreeLeaves(db)
	if err == nil {
		p, err := getProveData(transferFr2Bytes(leaf), leaves.Leaves)
		if err == nil {
			return p, nil
		}
	}

	if len(rootHash) > 0 {
		leaves, err := getCommitRootLeaves(db, rootHash)
		if err != nil {
			return nil, err
		}
		p, err := getProveData(transferFr2Bytes(leaf), leaves.Leaves)
		if err != nil {
			return nil, errors.Wrapf(err, "hash=%s,leaf=%s", rootHash, leaf)
		}
		return p, nil
	}

	roots, err := getArchiveCommitRoots(db)
	if err == nil {
		for _, root := range roots.Roots {
			leaves, err := getCommitRootLeaves(db, transferFr2String(root))
			if err == nil {
				p, err := getProveData(transferFr2Bytes(leaf), leaves.Leaves)
				if err == nil {
					return p, nil
				}
			}

		}
	}

	return nil, errors.Wrapf(err, "hash=%s,leaf=%s", rootHash, leaf)

}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/mix/executor/merkletree"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

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

func getCommitTreeStatus(db dbm.KV, assetExec, assetSymbol string) (*mixTy.CommitTreeStatus, error) {
	v, err := db.Get(calcCommitTreeCurrentStatusKey(assetExec, assetSymbol))
	if isNotFound(err) {
		//系统初始化开始，没有任何状态，初始seq设为1，如果一个merkle树完成后，status清空状态，seq也要初始为1，作为数据库占位，不然会往前查找
		return &mixTy.CommitTreeStatus{
			AssetExec:       assetExec,
			AssetSymbol:     assetSymbol,
			SubTrees:        &mixTy.CommitSubTrees{},
			ArchiveRootsSeq: 1}, nil
	}
	if err != nil {
		return nil, errors.Wrapf(err, "db.getCommitTreeStatus")
	}

	var status mixTy.CommitTreeStatus
	err = types.Decode(v, &status)
	if err != nil {
		return nil, errors.Wrapf(err, "db.decode CommitTreeStatus")
	}

	return &status, nil
}

func getSubLeaves(db dbm.KV, exec, symbol string, currentSeq int32) (*mixTy.CommitTreeLeaves, error) {
	var leaves mixTy.CommitTreeLeaves
	for i := int32(1); i <= currentSeq; i++ {
		l, err := getCommitLeaves(db, calcSubLeavesKey(exec, symbol, i))
		if err != nil {
			return nil, errors.Wrapf(err, "getSubLeaves seq=%d", i)
		}
		leaves.Leaves = append(leaves.Leaves, l.Leaves...)
	}

	return &leaves, nil
}

func getCommitRootLeaves(db dbm.KV, exec, symbol, rootHash string) (*mixTy.CommitTreeLeaves, error) {
	return getCommitLeaves(db, calcCommitTreeRootLeaves(exec, symbol, rootHash))
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

func getSubRoots(db dbm.KV, exec, symbol string, seq int32) (*mixTy.CommitTreeRoots, error) {
	return getCommitTreeRoots(db, calcSubRootsKey(exec, symbol, seq))
}

func getArchiveRoots(db dbm.KV, exec, symbol string, seq uint64) (*mixTy.CommitTreeRoots, error) {
	return getCommitTreeRoots(db, calcArchiveRootsKey(exec, symbol, seq))
}

//TODO seed config
func getNewTree() *merkletree.Tree {
	return merkletree.New(mimc.NewMiMC(mixTy.MimcHashSeed))
}

func calcTreeRoot(leaves *mixTy.CommitTreeLeaves) []byte {
	tree := getNewTree()
	for _, leaf := range leaves.Leaves {
		tree.Push(leaf)
	}
	return tree.Root()

}

func makeArchiveRootReceipt(exec, symbol string, seq uint64, root string) *types.Receipt {
	key := calcArchiveRootsKey(exec, symbol, seq)
	log := &mixTy.ReceiptArchiveTreeRoot{
		RootHash: root,
		Seq:      seq,
	}

	return &types.Receipt{
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&mixTy.CommitTreeRoots{Roots: [][]byte{mixTy.Str2Byte(root)}})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  mixTy.TyLogCommitTreeArchiveRoot,
				Log: types.Encode(log),
			},
		},
	}

}

func makeArchiveLeavesReceipt(exec, symbol, root string, leaves *mixTy.CommitTreeLeaves) *types.Receipt {

	key := calcCommitTreeRootLeaves(exec, symbol, root)
	log := &mixTy.ReceiptArchiveLeaves{
		RootHash: root,
		Count:    int32(len(leaves.Leaves)),
		LastLeaf: hex.EncodeToString(leaves.Leaves[len(leaves.Leaves)-1]),
	}
	return &types.Receipt{
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(leaves)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  mixTy.TyLogArchiveRootLeaves,
				Log: types.Encode(log),
			},
		},
	}

}

func makeSubRootsReceipt(exec, symbol string, seq int32, root []byte) *types.Receipt {
	key := calcSubRootsKey(exec, symbol, seq)
	log := &mixTy.ReceiptCommitSubRoots{
		Seq:  seq,
		Root: mixTy.Byte2Str(root),
	}
	return &types.Receipt{
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&mixTy.CommitTreeRoots{Roots: [][]byte{root}})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  mixTy.TyLogSubRoots,
				Log: types.Encode(log),
			},
		},
	}

}

func makeSubLeavesReceipt(exec, symbol string, seq int32, leaf []byte) *types.Receipt {
	key := calcSubLeavesKey(exec, symbol, seq)
	log := &mixTy.ReceiptCommitSubLeaves{
		Seq:  seq,
		Leaf: mixTy.Byte2Str(leaf),
	}
	return &types.Receipt{
		KV: []*types.KeyValue{
			{Key: key, Value: types.Encode(&mixTy.CommitTreeLeaves{Leaves: [][]byte{leaf}})},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  mixTy.TyLogSubLeaves,
				Log: types.Encode(log),
			},
		},
	}

}

func makeTreeStatusReceipt(exec, symbol string, prev, current *mixTy.CommitTreeStatus) *types.Receipt {
	keyStatus := calcCommitTreeCurrentStatusKey(exec, symbol)
	log := &mixTy.ReceiptCommitTreeStatus{
		Prev:    prev,
		Current: current,
	}

	return &types.Receipt{
		Ty: types.ExecOk,
		KV: []*types.KeyValue{
			{Key: keyStatus, Value: types.Encode(current)},
		},
		Logs: []*types.ReceiptLog{
			{
				Ty:  mixTy.TyLogCommitTreeStatus,
				Log: types.Encode(log),
			},
		},
	}

}

func getArchivedSubLeaves(db dbm.KV, exec, symbol string, maxTreeLeaves int32) (*mixTy.CommitTreeLeaves, error) {
	var leaves mixTy.CommitTreeLeaves
	//获取前1023个leaf
	for i := int32(1); i < maxTreeLeaves; i++ {
		r, err := getCommitLeaves(db, calcSubLeavesKey(exec, symbol, i))
		if err != nil {
			return nil, errors.Wrapf(err, "getArchivedSubTreeLeaves,i=%d", i)
		}
		leaves.Leaves = append(leaves.Leaves, r.Leaves...)
	}
	return &leaves, nil
}

func restoreSubTrees(subTrees *mixTy.CommitSubTrees) (*merkletree.Tree, error) {
	tree := getNewTree()
	for _, t := range subTrees.SubTrees {
		err := tree.PushSubTree(int(t.Height), t.Hash)
		if err != nil {
			return nil, errors.Wrapf(err, "restoreSubTrees.pushSubTree=%d", t.Height)
		}
	}

	return tree, nil

}

//merkle树由子树构成，任何一个节点都是一个子树，如果一个子树高度不为0，则是一个高阶子树，高阶子树可以直接参与root计算，不需要再统计其叶子节点
//高阶子树一定是一个完全子树，即左右叶子节点都存在，只有0阶子树不是完全子树，只有一个叶子
//为了计算1024个叶子节点任何一个新节点n的roothash，需要累计前n-1个叶子节点，由于1024个叶子存取耗时，这里只保存最多10个高阶子树即可计算roothash
func joinSubTrees(status *mixTy.CommitTreeStatus, leaf []byte) ([]byte, error) {
	tree, err := restoreSubTrees(status.SubTrees)
	if err != nil {
		return nil, errors.Wrapf(err, "joinSubTrees")
	}
	tree.Push(leaf)

	var newSubTrees mixTy.CommitSubTrees
	allSubTrees := tree.GetAllSubTrees()
	for _, s := range allSubTrees {
		newSubTrees.SubTrees = append(newSubTrees.SubTrees, &mixTy.CommitSubTree{Height: int32(s.GetHeight()), Hash: s.GetSum()})
	}

	status.SubTrees = &newSubTrees

	return tree.Root(), nil

}

func joinLeaves(db dbm.KV, status *mixTy.CommitTreeStatus, leaf []byte, maxTreeLeaves int32) (*types.Receipt, error) {
	receipts := &types.Receipt{Ty: types.ExecOk}

	//seq从1开始记录前1023个叶子和root
	if status.SubLeavesSeq < maxTreeLeaves {
		status.SubLeavesSeq++

		r := makeSubLeavesReceipt(status.AssetExec, status.AssetSymbol, status.SubLeavesSeq, leaf)
		mergeReceipt(receipts, r)

		//恢复并重新计算子树
		root, err := joinSubTrees(status, leaf)
		if err != nil {
			return nil, errors.Wrapf(err, "joinLeaves.joinSubTrees")
		}
		r = makeSubRootsReceipt(status.AssetExec, status.AssetSymbol, status.SubLeavesSeq, root)
		mergeReceipt(receipts, r)

		return receipts, nil
	}

	//累积到1024个叶子，需要归档
	sumLeaves, err := getArchivedSubLeaves(db, status.AssetExec, status.AssetSymbol, maxTreeLeaves)
	if err != nil {
		return nil, errors.Wrapf(err, "pushTree.joinLeaves")
	}
	//加第1024个leaf
	sumLeaves.Leaves = append(sumLeaves.Leaves, leaf)
	//重新计算1024个叶子root，确保正确
	root := mixTy.Byte2Str(calcTreeRoot(sumLeaves))
	//root-leaves保存leaves
	r := makeArchiveLeavesReceipt(status.AssetExec, status.AssetSymbol, root, sumLeaves)
	mergeReceipt(receipts, r)

	//1024叶子的root归档到相应archiveSeq
	r = makeArchiveRootReceipt(status.AssetExec, status.AssetSymbol, status.ArchiveRootsSeq, root)
	mergeReceipt(receipts, r)
	status.ArchiveRootsSeq++

	//reset 重新开始统计
	status.SubLeavesSeq = 0
	status.SubTrees = &mixTy.CommitSubTrees{}

	return receipts, nil
}

/*
1. 增加到当前leaves 和roots
2. 如果leaves 达到最大比如1024，则按root归档leaves，并归档相应root
3. 归档同时初始化新的current leaves 和roots
*/
func pushTree(db dbm.KV, exec, symbol string, leaves [][]byte, maxTreeLeaves int32) (*types.Receipt, error) {

	status, err := getCommitTreeStatus(db, exec, symbol)
	if err != nil {
		return nil, err
	}

	prev := proto.Clone(status).(*mixTy.CommitTreeStatus)

	receipts := &types.Receipt{Ty: types.ExecOk}
	for i, leaf := range leaves {
		if len(leaf) <= 0 {
			return nil, errors.Wrapf(types.ErrInvalidParam, "the %d leaf is null", i)
		}
		r, err := joinLeaves(db, status, leaf, maxTreeLeaves)
		if err != nil {
			return nil, errors.Wrapf(err, "pushTree.joinLeaves leaf=%s", mixTy.Byte2Str(leaf))
		}
		mergeReceipt(receipts, r)
	}
	r := makeTreeStatusReceipt(exec, symbol, prev, status)
	mergeReceipt(receipts, r)
	return receipts, nil

}

func checkExist(target []byte, list [][]byte) bool {
	for _, r := range list {
		if bytes.Equal(r, target) {
			return true
		}
	}
	return false
}

func checkTreeRootHashExist(db dbm.KV, exec, symbol string, hash []byte) (bool, error) {
	status, err := getCommitTreeStatus(db, exec, symbol)
	if err != nil {
		return false, errors.Wrapf(err, "checkTreeRootHashExist")
	}

	//查归档的subRoots,当前subSeq还未归档
	for i := int32(1); i <= status.SubLeavesSeq; i++ {
		subRoots, err := getSubRoots(db, exec, symbol, i)
		if err != nil {
			return false, errors.Wrapf(err, "checkTreeRootHashExist.getSubRoots seq=%d", i)
		}
		if checkExist(hash, subRoots.Roots) {
			return true, nil
		}
	}

	//再查归档的roots
	for i := status.ArchiveRootsSeq; i > 0; i-- {
		subRoots, err := getArchiveRoots(db, exec, symbol, i)
		if err != nil {
			return false, errors.Wrapf(err, "checkTreeRootHashExist.getArchiveRoots seq=%d", i)
		}
		if checkExist(hash, subRoots.Roots) {
			return true, nil
		}
	}

	return false, nil

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

	prove.RootHash = mixTy.Byte2Str(root)
	prove.ProofIndex = uint32(proofIndex)
	prove.NumLeaves = uint32(num)
	//set[0] 是targetLeaf
	for _, s := range proofSet {
		prove.ProofSet = append(prove.ProofSet, mixTy.Byte2Str(s))
	}

	helpers := merkletree.GenerateProofHelper(proofSet, proofIndex, num)

	for _, i := range helpers {
		prove.Helpers = append(prove.Helpers, uint32(i))
	}

	return &prove, nil

}

//1. 首先在当前tree查找
//2. 如果提供了rootHash,则根据roothash+leaf查找，否则全局遍历查找
func CalcTreeProve(db dbm.KV, exec, symbol, rootHash, leaf string) (*mixTy.CommitTreeProve, error) {
	if len(leaf) <= 0 {
		return nil, errors.Wrap(types.ErrInvalidParam, "leaf is null")
	}

	status, err := getCommitTreeStatus(db, exec, symbol)
	if err != nil {
		return nil, errors.Wrapf(err, "CalcTreeProve.getCommitTreeStatus")
	}

	leaves, err := getSubLeaves(db, exec, symbol, status.SubLeavesSeq)
	if err == nil {
		p, err := getProveData(mixTy.Str2Byte(leaf), leaves.Leaves)
		if err == nil {
			return p, nil
		}
	}

	if len(rootHash) > 0 {
		leaves, err := getCommitRootLeaves(db, exec, symbol, rootHash)
		if err != nil {
			return nil, errors.Wrapf(err, "getCommitRootLeaves rootHash=%s", rootHash)
		}
		//指定leaf 没有在指定的rootHash里面，返回错误，可以不指定rootHash 全局遍历
		p, err := getProveData(mixTy.Str2Byte(leaf), leaves.Leaves)
		if err != nil {
			return nil, errors.Wrapf(err, "hash=%s,leaf=%s", rootHash, leaf)
		}
		return p, nil
	}

	for i := status.ArchiveRootsSeq; i > 0; i-- {
		roots, err := getArchiveRoots(db, exec, symbol, i)
		if err != nil {
			return nil, errors.Wrapf(err, "getArchiveRoots,i=%d", i)
		}
		leaves, err := getCommitRootLeaves(db, exec, symbol, mixTy.Byte2Str(roots.Roots[0]))
		if err == nil {
			p, err := getProveData(mixTy.Str2Byte(leaf), leaves.Leaves)
			if err == nil {
				return p, nil
			}
		}
	}

	return nil, errors.Wrapf(mixTy.ErrLeafNotFound, "hash=%s,leaf=%s", rootHash, leaf)

}

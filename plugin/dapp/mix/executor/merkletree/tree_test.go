// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package merkletree

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	rand2 "math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

var h = mimc.NewMiMC("seed")

func TestLeafHash(t *testing.T) {
	leaves := []string{
		"16308793397024662832064523892418908145900866571524124093537199035808550255649",
	}

	s := leafSum(h, mixTy.Str2Byte(leaves[0]))
	assert.Equal(t, "4010939160279929375357088561050093294975728828994381439611270589357856115894", mixTy.Byte2Str(s))

	leaves = []string{
		"4062509694129705639089082212179777344962624935939361647381392834235969534831",
		"3656010751855274388516368747583374746848682779395325737100877017850943546836",
	}
	h.Reset()
	s = nodeSum(h, mixTy.Str2Byte(leaves[0]), mixTy.Str2Byte(leaves[1]))
	assert.Equal(t, "19203142125680456902129919467753534834062383756119332074615431762320316227830", mixTy.Byte2Str(s))

	proves := []string{
		"21467822781369104390668289189963532506973289112396605437823854946060028754354",
		"4010939160279929375357088561050093294975728828994381439611270589357856115894",
		"19203142125680456902129919467753534834062383756119332074615431762320316227830",
		"5949485921924623528448830799540295699445001672535082168235329176256394356669",
	}

	var sum []byte

	for i, l := range proves {
		if len(sum) == 0 {
			sum = leafSum(h, mixTy.Str2Byte(l))
			continue
		}
		//第0个leaf在第一个leaf的右侧，需要调整顺序
		if i < 2 {
			sum = nodeSum(h, mixTy.Str2Byte(l), sum)
			continue
		}
		sum = nodeSum(h, sum, mixTy.Str2Byte(l))
	}
	fmt.Println("sum", mixTy.Byte2Str(sum))
}

// TestBadInputs provides malicious inputs to the functions of the package,
// trying to trigger panics or unexpected behavior.
func TestBadInputs(t *testing.T) {
	// Get the root and proof of an empty tree.
	tree := New(h)
	if err := tree.SetIndex(0); err != nil {
		t.Fatal(err)
	}
	root := tree.Root()
	if root != nil {
		t.Error("root of empty tree should be nil")
	}
	_, proof, _, _ := tree.Prove()
	if proof != nil {
		t.Error("proof of empty tree should be nil")
	}

	// Get the proof of a tree that hasn't reached it's index.
	err := tree.SetIndex(3)
	if err != nil {
		t.Fatal(err)
	}
	tree.Push([]byte{1})
	_, proof, _, _ = tree.Prove()
	if proof != nil {
		t.Fatal(err)
	}
	err = tree.SetIndex(2)
	if err == nil {
		t.Error("expecting error, shouldn't be able to reset a tree after pushing")
	}

	// Try nil values in VerifyProof.
	mt := CreateMerkleTester(t)
	if VerifyProof(h, nil, mt.proofSets[1][0], 0, 1) {
		t.Error("VerifyProof should return false for nil merkle root")
	}
	if VerifyProof(h, []byte{1}, nil, 0, 1) {
		t.Error("VerifyProof should return false for nil proof set")
	}
	if VerifyProof(h, mt.roots[15], mt.proofSets[15][3][1:], 3, 15) {
		t.Error("VerifyProof should return false for too-short proof set")
	}
	if VerifyProof(h, mt.roots[15], mt.proofSets[15][10][1:], 10, 15) {
		t.Error("VerifyProof should return false for too-short proof set")
	}
	if VerifyProof(h, mt.roots[15], mt.proofSets[15][10], 15, 0) {
		t.Error("VerifyProof should return false when numLeaves is 0")
	}
}

// A MerkleTester contains data types that can be filled out manually to
// compare against function results.
type MerkleTester struct {
	// data is the raw data of the Merkle tree.
	data [][]byte

	// leaves is the hashes of the data, and should be the same length.
	leaves [][]byte

	// roots contains the root hashes of Merkle trees of various heights using
	// the data for input.
	roots map[int][]byte

	// proofSets contains proofs that certain data is in a Merkle tree. The
	// first map is the number of leaves in the tree that the proof is for. The
	// root of that tree can be found in roots. The second map is the
	// proofIndex that was used when building the proof.
	proofSets map[int]map[int][][]byte
	*testing.T
}

// CreateMerkleTester creates a Merkle tester and manually fills out many of
// the expected values for constructing Merkle tree roots and Merkle tree
// proofs. These manual values can then be compared against the values that the
// Tree creates.
func CreateMerkleTester(t *testing.T) (mt *MerkleTester) {
	mt = &MerkleTester{
		roots:     make(map[int][]byte),
		proofSets: make(map[int]map[int][][]byte),
	}
	mt.T = t

	// Fill out the data and leaves values.
	size := 16
	for i := 0; i < size; i++ {
		mt.data = append(mt.data, []byte{byte(i)})
	}
	for i := 0; i < size; i++ {
		mt.leaves = append(mt.leaves, sum(h, append([]byte{0}, mt.data[i]...)))
	}

	// Manually build out expected Merkle root values.
	mt.roots[0] = nil
	mt.roots[1] = mt.leaves[0]
	mt.roots[2] = mt.join(mt.leaves[0], mt.leaves[1])
	mt.roots[3] = mt.join(
		mt.roots[2],
		mt.leaves[2],
	)
	mt.roots[4] = mt.join(
		mt.roots[2],
		mt.join(mt.leaves[2], mt.leaves[3]),
	)
	mt.roots[5] = mt.join(
		mt.roots[4],
		mt.leaves[4],
	)

	mt.roots[6] = mt.join(
		mt.roots[4],
		mt.join(
			mt.leaves[4],
			mt.leaves[5],
		),
	)

	mt.roots[7] = mt.join(
		mt.roots[4],
		mt.join(
			mt.join(mt.leaves[4], mt.leaves[5]),
			mt.leaves[6],
		),
	)

	mt.roots[8] = mt.join(
		mt.roots[4],
		mt.join(
			mt.join(mt.leaves[4], mt.leaves[5]),
			mt.join(mt.leaves[6], mt.leaves[7]),
		),
	)

	mt.roots[15] = mt.join(
		mt.roots[8],
		mt.join(
			mt.join(
				mt.join(mt.leaves[8], mt.leaves[9]),
				mt.join(mt.leaves[10], mt.leaves[11]),
			),
			mt.join(
				mt.join(mt.leaves[12], mt.leaves[13]),
				mt.leaves[14],
			),
		),
	)

	// Manually build out some proof sets that should should match what the
	// Tree creates for the same values.
	mt.proofSets[1] = make(map[int][][]byte)
	mt.proofSets[1][0] = [][]byte{mt.data[0]}

	mt.proofSets[2] = make(map[int][][]byte)
	mt.proofSets[2][0] = [][]byte{
		mt.data[0],
		mt.leaves[1],
	}

	mt.proofSets[2][1] = [][]byte{
		mt.data[1],
		mt.leaves[0],
	}

	mt.proofSets[5] = make(map[int][][]byte)
	mt.proofSets[5][4] = [][]byte{
		mt.data[4],
		mt.roots[4],
	}

	mt.proofSets[6] = make(map[int][][]byte)
	mt.proofSets[6][0] = [][]byte{
		mt.data[0],
		mt.leaves[1],
		mt.join(
			mt.leaves[2],
			mt.leaves[3],
		),
		mt.join(
			mt.leaves[4],
			mt.leaves[5],
		),
	}

	mt.proofSets[6][2] = [][]byte{
		mt.data[2],
		mt.leaves[3],
		mt.roots[2],
		mt.join(
			mt.leaves[4],
			mt.leaves[5],
		),
	}

	mt.proofSets[6][4] = [][]byte{
		mt.data[4],
		mt.leaves[5],
		mt.roots[4],
	}

	mt.proofSets[6][5] = [][]byte{
		mt.data[5],
		mt.leaves[4],
		mt.roots[4],
	}

	mt.proofSets[7] = make(map[int][][]byte)
	mt.proofSets[7][5] = [][]byte{
		mt.data[5],
		mt.leaves[4],
		mt.leaves[6],
		mt.roots[4],
	}

	mt.proofSets[15] = make(map[int][][]byte)
	mt.proofSets[15][3] = [][]byte{
		mt.data[3],
		mt.leaves[2],
		mt.roots[2],
		mt.join(
			mt.join(mt.leaves[4], mt.leaves[5]),
			mt.join(mt.leaves[6], mt.leaves[7]),
		),
		mt.join(
			mt.join(
				mt.join(mt.leaves[8], mt.leaves[9]),
				mt.join(mt.leaves[10], mt.leaves[11]),
			),
			mt.join(
				mt.join(mt.leaves[12], mt.leaves[13]),
				mt.leaves[14],
			),
		),
	}

	mt.proofSets[15][10] = [][]byte{
		mt.data[10],
		mt.leaves[11],
		mt.join(
			mt.leaves[8],
			mt.leaves[9],
		),
		mt.join(
			mt.join(mt.leaves[12], mt.leaves[13]),
			mt.leaves[14],
		),
		mt.roots[8],
	}

	mt.proofSets[15][13] = [][]byte{
		mt.data[13],
		mt.leaves[12],
		mt.leaves[14],
		mt.join(
			mt.join(mt.leaves[8], mt.leaves[9]),
			mt.join(mt.leaves[10], mt.leaves[11]),
		),
		mt.roots[8],
	}

	return
}

// join returns the hash a  b.
func (mt *MerkleTester) join(a, b []byte) []byte {
	return sum(h, a, b)
}

// TestBuildRoot checks that the root returned by Tree matches the manually
// created roots for all of the manually created roots.
func TestBuildRoot(t *testing.T) {
	mt := CreateMerkleTester(t)

	// Compare the results of calling Root against all of the manually
	// constructed Merkle trees.
	var tree *Tree
	for i, root := range mt.roots {
		// Fill out the tree.
		tree = New(h)
		for j := 0; j < i; j++ {
			tree.Push(mt.data[j])
		}

		// Get the root and compare to the manually constructed root.
		treeRoot := tree.Root()
		if !bytes.Equal(root, treeRoot) {
			t.Error("tree root doesn't match manual root for index", i)
		}
	}
}

// TestBuildAndVerifyProof builds a proof using a tree for every single
// manually created proof in the MerkleTester. Then it checks that the proof
// matches the manually created proof, and that the proof is verified by
// VerifyProof. Then it checks that the proof fails for all other indices,
// which should happen if all of the leaves are unique.
func TestBuildAndVerifyProof(t *testing.T) {
	mt := CreateMerkleTester(t)

	// Compare the results of building a Merkle proof to all of the manually
	// constructed proofs.
	tree := New(h)
	for i, manualProveSets := range mt.proofSets {
		for j, expectedProveSet := range manualProveSets {
			// Build out the tree.
			tree = New(h)
			err := tree.SetIndex(uint64(j))
			if err != nil {
				t.Fatal(err)
			}
			for k := 0; k < i; k++ {
				tree.Push(mt.data[k])
			}

			// Get the proof and check all values.
			merkleRoot, proofSet, proofIndex, numSegments := tree.Prove()
			if !bytes.Equal(merkleRoot, mt.roots[i]) {
				t.Error("incorrect Merkle root returned by Tree for indices", i, j)
			}
			if len(proofSet) != len(expectedProveSet) {
				t.Error("proof set is wrong length for indices", i, j)
				continue
			}
			if proofIndex != uint64(j) {
				t.Error("incorrect proofIndex returned for indices", i, j)
			}
			if numSegments != uint64(i) {
				t.Error("incorrect numSegments returned for indices", i, j)
			}
			for k := range proofSet {
				if !bytes.Equal(proofSet[k], expectedProveSet[k]) {
					t.Error("proof set does not match expected proof set for indices", i, j, k)
				}
			}

			// Check that verification works on for the desired proof index but
			// fails for all other indices.
			if !VerifyProof(h, merkleRoot, proofSet, proofIndex, numSegments) {
				t.Error("proof set does not verify for indices", i, j)
			}
			for k := uint64(0); k < uint64(i); k++ {
				if k == proofIndex {
					continue
				}
				if VerifyProof(h, merkleRoot, proofSet, k, numSegments) {
					t.Error("proof set verifies for wrong index at indices", i, j, k)
				}
			}

			// Check that calling Prove a second time results in the same
			// values.
			merkleRoot2, proofSet2, proofIndex2, numSegments2 := tree.Prove()
			if !bytes.Equal(merkleRoot, merkleRoot2) {
				t.Error("tree returned different merkle roots after calling Prove twice for indices", i, j)
			}
			if len(proofSet) != len(proofSet2) {
				t.Error("tree returned different proof sets after calling Prove twice for indices", i, j)
			}
			for k := range proofSet {
				if !bytes.Equal(proofSet[k], proofSet2[k]) {
					t.Error("tree returned different proof sets after calling Prove twice for indices", i, j)
				}
			}
			if proofIndex != proofIndex2 {
				t.Error("tree returned different proof indexes after calling Prove twice for indices", i, j)
			}
			if numSegments != numSegments2 {
				t.Error("tree returned different segment count after calling Prove twice for indices", i, j)
			}
		}
	}
}

// TestCompatibility runs BuildProof for a large set of trees, and checks that
// verify affirms each proof, while rejecting for all other indexes (this
// second half requires that all input data be unique). The test checks that
// build and verify are internally consistent, but doesn't check for actual
// correctness.
func TestCompatibility(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	// Brute force all trees up to size 'max'. Running time for this test is max^3.
	max := uint64(17)
	tree := New(h)
	for i := uint64(1); i < max; i++ {
		// Try with proof at every possible index.
		for j := uint64(0); j < i; j++ {
			// Push unique data into the tree.
			tree = New(h)
			err := tree.SetIndex(j)
			if err != nil {
				t.Fatal(err)
			}
			for k := uint64(0); k < i; k++ {
				tree.Push([]byte{byte(k)})
			}

			// Build the proof for the tree and run it through verify.
			merkleRoot, proofSet, proofIndex, numLeaves := tree.Prove()
			if !VerifyProof(h, merkleRoot, proofSet, proofIndex, numLeaves) {
				t.Error("proof didn't verify for indices", i, j)
			}

			// Check that verification fails for all other indices.
			for k := uint64(0); k < i; k++ {
				if k == j {
					continue
				}
				if VerifyProof(h, merkleRoot, proofSet, k, numLeaves) {
					t.Error("proof verified for indices", i, j, k)
				}
			}
		}
	}

	// Check that proofs on larger trees are consistent.
	for i := 0; i < 5; i++ {
		// Determine a random size for the tree up to 64M elements.
		sizeI, err := rand.Int(rand.Reader, big.NewInt(256e3))
		if err != nil {
			t.Fatal(err)
		}
		size := uint64(sizeI.Int64())

		proofIndexI, err := rand.Int(rand.Reader, sizeI)
		if err != nil {
			t.Fatal(err)
		}
		proofIndex := uint64(proofIndexI.Int64())

		// Prepare the tree.
		tree = New(h)
		err = tree.SetIndex(proofIndex)
		if err != nil {
			t.Fatal(err)
		}

		// Insert 'size' unique elements.
		for j := 0; j < int(size); j++ {
			elem := []byte(strconv.Itoa(j))
			tree.Push(elem)
		}

		// Get the proof for the tree and run it through verify.
		merkleRoot, proofSet, proofIndex, numLeaves := tree.Prove()
		if !VerifyProof(h, merkleRoot, proofSet, proofIndex, numLeaves) {
			t.Error("proof didn't verify in long test", size, proofIndex)
		}
	}
}

// TestLeafCounts checks that the number of leaves in the tree are being
// reported correctly.
func TestLeafCounts(t *testing.T) {
	tree := New(h)
	err := tree.SetIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, leaves := tree.Prove()
	if leaves != 0 {
		t.Error("bad reporting of leaf count")
	}

	tree = New(h)
	err = tree.SetIndex(0)
	if err != nil {
		t.Fatal(err)
	}
	tree.Push([]byte{})
	_, _, _, leaves = tree.Prove()
	if leaves != 1 {
		t.Error("bad reporting on leaf count")
	}
}

// convert int to []byte
func IntToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

// TestPushSubTreeCorrectRoot creates data for 4 leaves, combines them in
// different ways and makes sure that the root is always the same.
func TestPushSubTreeCorrectRoot(t *testing.T) {
	hash := h
	r := rand2.New(rand2.NewSource(0))

	// Create the data for 4 leaves.
	leaf1Data := IntToBytes(r.Int())
	leaf2Data := IntToBytes(r.Int())
	leaf3Data := IntToBytes(r.Int())
	leaf4Data := IntToBytes(r.Int())

	// Push the leaves into a tree and get the root.
	tree := New(hash)
	tree.Push(leaf1Data)
	tree.Push(leaf2Data)
	tree.Push(leaf3Data)
	tree.Push(leaf4Data)
	expectedRoot := tree.Root()

	// Create 4 height 0 subtrees and combine them. The root should be the
	// same.
	tree2 := New(hash)
	leaf1Hash := leafSum(hash, leaf1Data)
	leaf2Hash := leafSum(hash, leaf2Data)
	leaf3Hash := leafSum(hash, leaf3Data)
	leaf4Hash := leafSum(hash, leaf4Data)
	tree2.PushSubTree(0, leaf1Hash)
	tree2.PushSubTree(0, leaf2Hash)
	tree2.PushSubTree(0, leaf3Hash)
	tree2.PushSubTree(0, leaf4Hash)

	if !bytes.Equal(tree2.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 2 height 1 subtrees and combine them. The root should be the
	// same.
	tree3 := New(hash)
	node12Hash := nodeSum(hash, leaf1Hash, leaf2Hash)
	node34Hash := nodeSum(hash, leaf3Hash, leaf4Hash)
	tree3.PushSubTree(1, node12Hash)
	tree3.PushSubTree(1, node34Hash)

	if !bytes.Equal(tree3.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 1 height 2 subtree and add it to the tree. The root should be the
	// same.
	tree4 := New(hash)
	node1234Hash := nodeSum(hash, node12Hash, node34Hash)
	if err := tree4.PushSubTree(2, node1234Hash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tree4.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 1 height 1 tree and add 2 height 0 trees. The root should be the
	// same.
	tree5 := New(hash)
	tree5.PushSubTree(1, node12Hash)
	tree5.PushSubTree(0, leaf3Hash)
	tree5.PushSubTree(0, leaf4Hash)

	if !bytes.Equal(tree5.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 1 height 1 tree and add 2 leaves. The root should be the same.
	tree6 := New(hash)
	if err := tree6.PushSubTree(1, node12Hash); err != nil {
		t.Fatal(err)
	}
	tree6.Push(leaf3Data)
	tree6.Push(leaf4Data)
	if !bytes.Equal(tree6.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 2 height 0 trees and add 1 height 1 tree. The root should be the
	// same.
	tree7 := New(hash)
	tree7.PushSubTree(0, leaf1Hash)
	tree7.PushSubTree(0, leaf2Hash)
	tree7.PushSubTree(1, node34Hash)

	if !bytes.Equal(tree7.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 2 leaves and add 1 height 1 tree. The root should be the same.
	tree8 := New(hash)
	tree8.Push(leaf1Data)
	tree8.Push(leaf2Data)
	if err := tree8.PushSubTree(1, node34Hash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tree8.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}
}

// TestPushSubTreeCorrectRootWithProof creates data for 4 leaves, combines them
// in different ways and makes sure that the root is always the same. It also
// creates a proof for them.
func TestPushSubTreeCorrectRootWithProof(t *testing.T) {
	hash := h
	r := rand2.New(rand2.NewSource(0))

	// Create the data for 4 leaves.
	leaf1Data := IntToBytes(r.Int())
	leaf2Data := IntToBytes(r.Int())
	leaf3Data := IntToBytes(r.Int())
	leaf4Data := IntToBytes(r.Int())

	// Push the leaves into a tree and get the root.
	tree := New(hash)
	proofIndex := uint64(r.Intn(4))
	if err := tree.SetIndex(proofIndex); err != nil {
		t.Fatal(err)
	}
	tree.Push(leaf1Data)
	tree.Push(leaf2Data)
	tree.Push(leaf3Data)
	tree.Push(leaf4Data)
	expectedRoot := tree.Root()

	// Create 1 height 1 tree and add 2 leaves. The root should be the same.
	tree2 := New(hash)
	proofIndex = uint64(2 + r.Intn(2))
	leaf1Hash := leafSum(hash, leaf1Data)
	leaf2Hash := leafSum(hash, leaf2Data)
	node12Hash := nodeSum(hash, leaf1Hash, leaf2Hash)
	if err := tree2.SetIndex(proofIndex); err != nil {
		t.Fatal(err)
	}
	if err := tree2.PushSubTree(1, node12Hash); err != nil {
		t.Fatal(err)
	}
	tree2.Push(leaf3Data)
	tree2.Push(leaf4Data)
	if !bytes.Equal(tree2.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Create 2 leaves and add 1 height 1 tree. The root should be the same.
	tree3 := New(hash)
	proofIndex = uint64(r.Intn(2))
	leaf3Hash := leafSum(hash, leaf3Data)
	leaf4Hash := leafSum(hash, leaf4Data)
	if err := tree3.SetIndex(proofIndex); err != nil {
		t.Fatal(err)
	}
	node34Hash := nodeSum(hash, leaf3Hash, leaf4Hash)
	tree3.Push(leaf1Data)
	tree3.Push(leaf2Data)
	if err := tree3.PushSubTree(1, node34Hash); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tree3.Root(), expectedRoot) {
		t.Fatal("root doesn't match expected root")
	}

	// Test the proofs for all the trees.
	merkleRoot, proofSet, index, numLeaves := tree.Prove()
	if !VerifyProof(h, merkleRoot, proofSet, index, numLeaves) {
		t.Fatal("failed to verify proof for tree")
	}
	merkleRoot, proofSet, index, numLeaves = tree2.Prove()
	if !VerifyProof(h, merkleRoot, proofSet, index, numLeaves) {
		t.Fatal("failed to verify proof for tree2")
	}
	merkleRoot, proofSet, index, numLeaves = tree3.Prove()
	if !VerifyProof(h, merkleRoot, proofSet, index, numLeaves) {
		t.Fatal("failed to verify proof for tree3")
	}
}

// TestPushSubTreeSimple tests pushing some valid and invalid subTrees to the
// tree.
func TestPushSubTreeSimple(t *testing.T) {
	tree := New(h)

	// Add a subTree of height 5 to the empty tree.
	if err := tree.PushSubTree(5, []byte{}); err != nil {
		t.Fatal(err)
	}
	if tree.Root() == nil {
		t.Fatal("root should not be nil after adding a subTree")
	}
	// Add a subTree of a height >5 to the tree. This should not be possible.
	if err := tree.PushSubTree(6, []byte{}); err == nil {
		t.Fatal("pushing a subTree with a larger height than the smallest subTree should fail")
	}
	// The current index should be 2^5
	expectedIndex := uint64(1 << 5)
	if tree.currentIndex != expectedIndex {
		t.Errorf("expected index %v but was %v", expectedIndex, tree.currentIndex)
	}
	// Add a subTree of the same height as the smallest subTree in the merkle
	// tree and check again.
	if err := tree.PushSubTree(5, []byte{}); err != nil {
		t.Fatal(err)
	}
	expectedIndex *= 2
	if tree.currentIndex != expectedIndex {
		t.Errorf("expected index %v but was %v", expectedIndex, tree.currentIndex)
	}
	// Push some data equal to height 2 and make sure the expectedIndex is correct.
	for i := 0; i < 4; i++ {
		tree.Push([]byte{})
		expectedIndex++
		if tree.currentIndex != expectedIndex {
			t.Errorf("expected index %v but was %v", expectedIndex, tree.currentIndex)
		}
	}
	// Add a subTree of height 2 and check the index again.
	if err := tree.PushSubTree(2, []byte{}); err != nil {
		t.Fatal(err)
	}
	expectedIndex += 4
	if tree.currentIndex != expectedIndex {
		t.Errorf("expected index %v but was %v", expectedIndex, tree.currentIndex)
	}

	// Create a new tree and set the proof index to 1. Afterwards we push twice
	// to create a subTree of height 1 that contains the proof index.
	tree2 := New(h)
	if err := tree2.SetIndex(1); err != nil {
		t.Fatal(err)
	}
	tree2.Push([]byte{})
	tree2.Push([]byte{})
	// Push a subTree of height 1. That should be fine.
	if err := tree2.PushSubTree(1, []byte{}); err != nil {
		t.Fatal(err)
	}
	// Create a new tree and set the proof index to 3. Afterwards we push twice
	// to create a subTree of height 1.
	tree3 := New(h)
	if err := tree3.SetIndex(2); err != nil {
		t.Fatal(err)
	}
	tree3.Push([]byte{})
	tree3.Push([]byte{})
	// Push a subTree of height 1. That shouldn't work since the subTree can't
	// contain the piece for the proof.
	if err := tree3.PushSubTree(1, []byte{}); err == nil {
		t.Fatal("we shouldn't be able to push a subTree that contains the proof index")
	}
	// Create a new tree and set the proof index to 4. Afterwards we push twice
	// to create a subTree of height 1.
	tree4 := New(h)
	if err := tree4.SetIndex(3); err != nil {
		t.Fatal(err)
	}
	tree4.Push([]byte{})
	tree4.Push([]byte{})
	// Push a subTree of height 1. That shouldn't work since the subTree can't
	// contain the piece for the proof.
	if err := tree4.PushSubTree(1, []byte{}); err == nil {
		t.Fatal("we shouldn't be able to push a subTree that contains the proof index")
	}
}

// BenchmarkTree64_4MB creates a Merkle tree out of 4MB using a segment size of
// 64 bytes, using sha256.
func BenchmarkTree64_4MB(b *testing.B) {
	data := make([]byte, 4*1024*1024)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatal(err)
	}
	segmentSize := 64

	b.ResetTimer()
	tree := New(h)
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data)/segmentSize; j++ {
			tree.Push(data[j*segmentSize : (j+1)*segmentSize])
		}
		tree.Root()
	}
}

// BenchmarkTree4k_4MB creates a Merkle tree out of 4MB using a segment size of
// 4096 bytes, using sha256.
func BenchmarkTree4k_4MB(b *testing.B) {
	data := make([]byte, 4*1024*1024)
	_, err := rand.Read(data)
	if err != nil {
		b.Fatal(err)
	}
	segmentSize := 4096

	b.ResetTimer()
	tree := New(h)
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(data)/segmentSize; j++ {
			tree.Push(data[j*segmentSize : (j+1)*segmentSize])
		}
		tree.Root()
	}
}

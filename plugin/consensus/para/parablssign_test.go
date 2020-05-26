// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/magiconair/properties/assert"
)

func TestSetAddrsBitMap(t *testing.T) {
	nodes := []string{"aa", "bb", "cc", "dd"}
	addrs := []string{}
	rst, rem := setAddrsBitMap(nodes, addrs)
	assert.Equal(t, len(rst), 0)
	assert.Equal(t, len(rem), 0)

	addrs = []string{"aa"}
	rst, rem = setAddrsBitMap(nodes, addrs)
	assert.Equal(t, rst, []byte{0x1})
	assert.Equal(t, len(rem), 0)

	addrs = []string{"aa", "cc"}
	rst, rem = setAddrsBitMap(nodes, addrs)
	assert.Equal(t, rst, []byte{0x5})
	assert.Equal(t, len(rem), 0)

	addrs = []string{"aa", "cc", "dd"}
	rst, rem = setAddrsBitMap(nodes, addrs)
	assert.Equal(t, rst, []byte{0xd})
	assert.Equal(t, len(rem), 0)
}

func TestIntegrateCommits(t *testing.T) {
	pool := make(map[int64]*pt.ParaBlsSignSumDetails)
	var commits []*pt.ParacrossCommitAction
	cmt1 := &pt.ParacrossCommitAction{
		Status: &pt.ParacrossNodeStatus{Height: 0},
		Bls:    &pt.ParacrossCommitBlsInfo{Addrs: []string{"aa"}, Sign: []byte{}},
	}
	cmt2 := &pt.ParacrossCommitAction{
		Status: &pt.ParacrossNodeStatus{Height: 0},
		Bls:    &pt.ParacrossCommitBlsInfo{Addrs: []string{"bb"}, Sign: []byte{}},
	}
	commits = []*pt.ParacrossCommitAction{cmt1, cmt1, cmt1, cmt2, cmt1}
	integrateCommits(pool, commits)
	assert.Equal(t, len(pool[0].Addrs), 2)
	assert.Equal(t, len(pool[0].Msgs), 2)
	assert.Equal(t, len(pool[0].Signs), 2)
	assert.Equal(t, pool[0].Addrs[0], "aa")
	assert.Equal(t, pool[0].Addrs[1], "bb")

}

func TestSecpPrikey2BlsPub(t *testing.T) {
	key := ""
	ret, _ := secpPrikey2BlsPub(key)
	assert.Equal(t, "", ret)

	key = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71"
	q := "0x87c58bb6cce41842462a0030335bb95948dcfba77e47e2d8ee893c0b2c34ac20d08c9e98a883ef2a6492d0ad808ace9a1730e8bae5d3b0861aaf743449df5de510073e2991c7274cab47f327e48d7eacf300e4b24174dae2e8603d1904b8a015"
	ret, _ = secpPrikey2BlsPub(key)
	assert.Equal(t, q, ret)
}

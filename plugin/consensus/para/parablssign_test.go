// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"testing"

	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	"github.com/herumi/bls-eth-go-binary/bls"
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

func TestBlsSignMain(t *testing.T) {
	//只初始化一次，多次初始化会并行产生冲突
	bls.Init(bls.BLS12_381)

	testSecpPrikey2BlsPub(t)
	testBlsSign(t)
	testVerifyBlsSign(t)
}

func testSecpPrikey2BlsPub(t *testing.T) {
	key := ""
	ret, _ := secp256Prikey2BlsPub(key)
	assert.Equal(t, "", ret)

	//real prikey="1626b254a75e5c44de9500a0c7897643e7736c09a7270b807546acb7cf7c94c9"
	key = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71"
	q := "0x980287e26d4d44f8c57944ffc096f7d98a460c97dadbffaed14ff0de901fa7f8afc59fcb1805a0b031e5eae5601df1c2"
	ret, _ = secp256Prikey2BlsPub(key)
	assert.Equal(t, q, ret)
}

func testBlsSign(t *testing.T) {
	status := &pt.ParacrossNodeStatus{}
	status.Height = 0
	status.Title = "user.p.para."

	KS := "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	PubKS := "a3d97d4186c80268fe6d3689dd574599e25df2dffdcff03f7d8ef64a3bd483241b7d0985958990de2d373d5604caf805"
	PriKS := "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"

	commit := &pt.ParacrossCommitAction{Status: status}
	client := &blsClient{}
	client.peersBlsPubKey = make(map[string]*bls.PublicKey)

	var prikey bls.SecretKey
	prikey.DeserializeHexStr(PriKS)
	t.Log("pri", prikey.SerializeToHexStr())
	client.blsPriKey = &prikey
	err := client.blsSign([]*pt.ParacrossCommitAction{commit})
	assert.Equal(t, err, nil)

	var pub bls.PublicKey
	pub.DeserializeHexStr(PubKS)
	client.peersBlsPubKey[KS] = &pub
	t.Log("pubks", pub.SerializeToHexStr())
	var sign bls.Sign
	sign.Deserialize(commit.Bls.Sign)
	msg := types.Encode(status)

	ret := sign.VerifyByte(&pub, msg)
	assert.Equal(t, ret, true)

	err = client.verifyBlsSign(KS, commit)
	assert.Equal(t, err, nil)
}

func testVerifyBlsSign(t *testing.T) {
	client := &blsClient{}
	client.peersBlsPubKey = make(map[string]*bls.PublicKey)
	KS := "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	PubKS := "a3d97d4186c80268fe6d3689dd574599e25df2dffdcff03f7d8ef64a3bd483241b7d0985958990de2d373d5604caf805"
	var pub bls.PublicKey
	pub.DeserializeHexStr(PubKS)
	client.peersBlsPubKey[KS] = &pub

	commit := &pt.ParacrossCommitAction{}
	blsInfo := &pt.ParacrossCommitBlsInfo{}
	signData := "0x82753675393576758571cbbaefada498614b4a0a967ca2dd5724eb46ecfd1c89f1e49792ebbe1866c1d6d6ceaf3054c7189751477a5b7312218eb77dcab1bfb6287c6fbf2e1c6cf8fe2ade7c17596b081dc98be785a34db5b45a5cca08e7e744"
	blsInfo.Sign = common.FromHex(signData)
	status := &pt.ParacrossNodeStatus{}
	data := "0x1a0c757365722e702e706172612e322097162f9d4a888121fdba2fb1ab402596acdbcb602121bd12284adb739d85f225"
	msg := common.FromHex(data)
	types.Decode(msg, status)
	commit.Status = status
	commit.Bls = blsInfo
	err := client.verifyBlsSign(KS, commit)
	assert.Equal(t, err, nil)
}

package executor

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/33cn/chain33/util"
	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/stretchr/testify/assert"
)

func TestAccountTree(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)
	info, err := generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	for i := 0; i < 3000; i++ {
		ethAddress := "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" + strconv.Itoa(i)
		chain33Addr := getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")
		_, localKvs,err := AddNewLeaf(statedb, localdb, info, ethAddress, 1, "1000", chain33Addr)
		assert.Equal(t, nil, err)
		for _, kv := range localKvs {
			localdb.Set(kv.GetKey(), kv.GetValue())
		}
	}
	tree, err := getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(3000), tree.GetTotalIndex())
	root, err := GetRootByStartIndex(statedb, 1, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, root)
	t.Log(root)
	for i := 0; i < 10; i++ {
		_, localKvs,err := UpdateLeaf(statedb, localdb, info, uint64(i+1), 2, "1000", zt.Add)
		assert.Equal(t, nil, err)
		for _, kv := range localKvs {
			localdb.Set(kv.GetKey(), kv.GetValue())
		}
		root, err = GetRootByStartIndex(statedb, 1, info)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, root)
		t.Log(root)
	}

	for i := 0; i < 10; i++ {
		_, localKvs,err := UpdateLeaf(statedb, localdb, info, uint64(i+2000), 1, "1000", zt.Sub)
		assert.Equal(t, nil, err)
		for _, kv := range localKvs {
			localdb.Set(kv.GetKey(), kv.GetValue())
		}
		root, err = GetRootByStartIndex(statedb, 1, info)
		assert.Equal(t, nil, err)
		assert.NotEqual(t, nil, root)
		t.Log(root)
	}
}

func getChain33Addr(privateKeyString string) string {
	privateKeyBytes, err := hex.DecodeString(privateKeyString)
	if err != nil {
		panic(err)
	}
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(privateKeyBytes))

	hash := mimc.NewMiMC(mixTy.MimcHashSeed)
	hash.Write(privateKey.PublicKey.Bytes())
	return hex.EncodeToString(hash.Sum(nil))
}

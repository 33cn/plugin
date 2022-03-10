package executor

import (
	"bytes"
	"encoding/hex"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"strconv"
	"testing"

	"github.com/33cn/chain33/util"
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
	for i := 0; i < 2000; i++ {
		ethAddress := "12345678901012345" + strconv.Itoa(i)
		chain33Addr := zt.HexAddr2Decimal(getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec"))
		_, localKvs, err := AddNewLeaf(statedb, localdb, info, ethAddress, 1, "1000", chain33Addr)
		tree, err := getAccountTree(statedb, info)
		t.Log("treeIndex", tree)
		assert.Equal(t, nil, err)
		for _, kv := range localKvs {
			localdb.Set(kv.GetKey(), kv.GetValue())
		}
	}
	tree, err := getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(2000), tree.GetTotalIndex())
	root, err := GetRootByStartIndex(statedb, 1, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, root)
	t.Log(root)
	for i := 0; i < 10; i++ {
		_, localKvs, err := UpdateLeaf(statedb, localdb, info, uint64(i+1), 2, "1000", zt.Add)
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
		_, localKvs, err := UpdateLeaf(statedb, localdb, info, uint64(i+2000), 1, "1000", zt.Sub)
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

	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash.Write(privateKey.PublicKey.Bytes())
	return hex.EncodeToString(hash.Sum(nil))
}

func TestAccountHash(t *testing.T) {
	var leaf zt.Leaf
	leaf.AccountId = 1
	leaf.EthAddress = "980818135352849559554652468538757099471386586455"
	leaf.Chain33Addr = "3415326846406104843498339737738292353412449296387254161761470177873504232418"

	leaf.TokenHash = "14633446003514262524099709640745596521508648778482661942408784061885334136010"
	var pubkey zt.ZkPubKey
	pubkey.X = "110829526890202442231796950896186450339098004198300292113013256946470504791"
	pubkey.Y = "12207062062295480868601430817261127111444831355336859496235449885847711361351"
	//leaf.PubKey = &pubkey

	hash := getLeafHash(&leaf)
	var f fr.Element
	f.SetBytes(hash)
	t.Log("hash", f.String())
}

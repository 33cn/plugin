package executor

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/33cn/chain33/util"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/stretchr/testify/assert"
)

func TestAccountTree(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)

	ethAddress1 := zt.HexAddr2Decimal("bbcd68033A72978C1084E2d44D1Fa06DdC4A2d5")
	chain33Addr1 := zt.HexAddr2Decimal(getChain33Addr("1266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec"))
	info, err := generateTreeUpdateInfo(statedb, localdb, ethAddress1, chain33Addr1)
	assert.Equal(t, nil, err)
	ethAddress := zt.HexAddr2Decimal("abcd68033A72978C1084E2d44D1Fa06DdC4A2d58")
	chain33Addr := zt.HexAddr2Decimal(getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec"))
	_, localKvs, err := AddNewLeaf(statedb, localdb, info, ethAddress, 1, "1000", chain33Addr)
	assert.Equal(t, nil, err)
	tree, err := getAccountTree(statedb, info)
	t.Log("treeIndex", tree)
	assert.Equal(t, nil, err)
	for _, kv := range localKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	tree, err = getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, uint64(2), tree.GetTotalIndex())

	_, localKvs, err = UpdateLeaf(statedb, localdb, info, 2, 2, "1000", zt.Add)
	assert.Equal(t, nil, err)
	for _, kv := range localKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

}

func getChain33Addr(privateKeyString string) string {
	privateKeyBytes, err := hex.DecodeString(privateKeyString)
	if err != nil {
		panic(err)
	}
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(privateKeyBytes))
	if err != nil {
		panic(err)
	}
	hash := mimc.NewMiMC(zt.ZkMimcHashSeed)
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.X.String()))
	hash.Write(zt.Str2Byte(privateKey.PublicKey.A.Y.String()))
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

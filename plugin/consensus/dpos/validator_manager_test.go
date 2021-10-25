package dpos

import (
	"bytes"
	"encoding/hex"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fmt"
	"os"
	"testing"
)

const (
	genesisContent = `{"genesis_time":"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFj","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""},{"pub_key":{"type":"secp256k1","data":"027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"},"name":""},{"pub_key":{"type":"secp256k1","data":"03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"},"name":""}],"app_hash":null}`
	pubkey11       = "03541AB9887951C038273648545072E5B6A46A639BFF535F3957E8150CBE2A70D7"
)

var (
	genDoc *ttypes.GenesisDoc
)

func init() {
	//为了使用VRF，需要使用SECP256K1体系的公私钥
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic("init ConsensusCrypto failed.")
	}

	ttypes.ConsensusCrypto = cr

	remove("./genesis.json")
	save("./genesis.json", genesisContent)
	genDoc, _ = ttypes.GenesisDocFromFile("./genesis.json")
}

func save(filename, filecontent string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("err = ", err)
		return
	}

	defer f.Close()

	n, err := f.WriteString(filecontent)
	if err != nil {
		fmt.Println("err = ", err)
		return
	}

	fmt.Println("n=", n, " contentlen=", len(filecontent))
}

func remove(filename string) {
	os.Remove(filename)

}

func TestMakeGenesisValidatorMgr(t *testing.T) {
	vMgr, err := MakeGenesisValidatorMgr(genDoc)
	require.Nil(t, err)
	assert.True(t, vMgr.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgr.AppHash) == 0)
	assert.True(t, len(vMgr.Validators.Validators) == 3)
	assert.True(t, vMgr.VrfValidators == nil)
	assert.True(t, vMgr.NoVrfValidators == nil)
	assert.True(t, vMgr.LastCycleBoundaryInfo == nil)

	vMgrCopy := vMgr.Copy()
	assert.True(t, vMgrCopy.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgrCopy.AppHash) == 0)
	assert.True(t, len(vMgrCopy.Validators.Validators) == 3)
	assert.True(t, vMgrCopy.VrfValidators == nil)
	assert.True(t, vMgrCopy.NoVrfValidators == nil)
	assert.True(t, vMgrCopy.LastCycleBoundaryInfo == nil)

	assert.True(t, vMgrCopy.Equals(vMgr))
	assert.True(t, vMgrCopy.IsEmpty() == false)
	assert.True(t, len(vMgrCopy.GetValidators().Validators) == 3)
}

func TestGetValidatorByIndex(t *testing.T) {
	vMgr, err := MakeGenesisValidatorMgr(genDoc)
	require.Nil(t, err)
	assert.True(t, vMgr.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgr.AppHash) == 0)
	assert.True(t, len(vMgr.Validators.Validators) == 3)
	assert.True(t, vMgr.VrfValidators == nil)
	assert.True(t, vMgr.NoVrfValidators == nil)
	assert.True(t, vMgr.LastCycleBoundaryInfo == nil)

	vMgr.ShuffleType = ShuffleTypeNoVrf
	addr, val := vMgr.GetValidatorByIndex(-1)
	require.Nil(t, addr)
	require.Nil(t, val)

	addr, val = vMgr.GetValidatorByIndex(1)
	assert.True(t, bytes.Equal(addr, vMgr.Validators.Validators[1].Address))
	assert.True(t, bytes.Equal(val.PubKey, vMgr.Validators.Validators[1].PubKey))

	vMgr.VrfValidators = ttypes.NewValidatorSet(vMgr.Validators.Validators)
	vMgr.ShuffleType = ShuffleTypeVrf

	addr, val = vMgr.GetValidatorByIndex(2)
	assert.True(t, bytes.Equal(addr, vMgr.VrfValidators.Validators[2].Address))
	assert.True(t, bytes.Equal(val.PubKey, vMgr.VrfValidators.Validators[2].PubKey))

	vMgr.ShuffleType = ShuffleTypePartVrf
	val, flag := vMgr.VrfValidators.Remove(addr)
	assert.True(t, flag)

	vMgr.NoVrfValidators = &ttypes.ValidatorSet{}
	vMgr.NoVrfValidators.Validators = append(vMgr.NoVrfValidators.Validators, val)
	addr, val = vMgr.GetValidatorByIndex(2)
	assert.True(t, bytes.Equal(addr, vMgr.NoVrfValidators.Validators[0].Address))
	assert.True(t, bytes.Equal(val.PubKey, vMgr.NoVrfValidators.Validators[0].PubKey))
}

func TestGetIndexByPubKey(t *testing.T) {
	vMgr, err := MakeGenesisValidatorMgr(genDoc)
	require.Nil(t, err)
	assert.True(t, vMgr.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgr.AppHash) == 0)
	assert.True(t, len(vMgr.Validators.Validators) == 3)
	assert.True(t, vMgr.VrfValidators == nil)
	assert.True(t, vMgr.NoVrfValidators == nil)
	assert.True(t, vMgr.LastCycleBoundaryInfo == nil)

	vMgr.ShuffleType = ShuffleTypeNoVrf
	index := vMgr.GetIndexByPubKey(nil)
	assert.True(t, index == -1)

	index = vMgr.GetIndexByPubKey(vMgr.Validators.Validators[1].PubKey)
	assert.True(t, index == 1)

	index = vMgr.GetIndexByPubKey(vMgr.Validators.Validators[0].PubKey)
	assert.True(t, index == 0)

	index = vMgr.GetIndexByPubKey(vMgr.Validators.Validators[2].PubKey)
	assert.True(t, index == 2)

	index = vMgr.GetIndexByPubKey([]byte("afdafafdfa"))
	assert.True(t, index == -1)

	vMgr.VrfValidators = ttypes.NewValidatorSet(vMgr.Validators.Validators)
	vMgr.ShuffleType = ShuffleTypeVrf

	index = vMgr.GetIndexByPubKey(vMgr.VrfValidators.Validators[1].PubKey)
	assert.True(t, index == 1)

	index = vMgr.GetIndexByPubKey(vMgr.VrfValidators.Validators[0].PubKey)
	assert.True(t, index == 0)

	index = vMgr.GetIndexByPubKey(vMgr.VrfValidators.Validators[2].PubKey)
	assert.True(t, index == 2)

	index = vMgr.GetIndexByPubKey([]byte("afdafafdfa"))
	assert.True(t, index == -1)

	vMgr.ShuffleType = ShuffleTypePartVrf
	val, flag := vMgr.VrfValidators.Remove(vMgr.VrfValidators.Validators[2].Address)
	assert.True(t, flag)

	vMgr.NoVrfValidators = &ttypes.ValidatorSet{}
	vMgr.NoVrfValidators.Validators = append(vMgr.NoVrfValidators.Validators, val)
	index = vMgr.GetIndexByPubKey(vMgr.NoVrfValidators.Validators[0].PubKey)
	assert.True(t, index == 2)
}

func TestFillVoteItem(t *testing.T) {

	vMgr, err := MakeGenesisValidatorMgr(genDoc)
	require.Nil(t, err)
	assert.True(t, vMgr.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgr.AppHash) == 0)
	assert.True(t, len(vMgr.Validators.Validators) == 3)
	assert.True(t, vMgr.VrfValidators == nil)
	assert.True(t, vMgr.NoVrfValidators == nil)
	assert.True(t, vMgr.LastCycleBoundaryInfo == nil)

	vMgr.ShuffleType = ShuffleTypeNoVrf

	voteItem := &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	assert.True(t, voteItem.LastCBInfo == nil)
	assert.True(t, voteItem.ShuffleType == ShuffleTypeNoVrf)
	assert.True(t, len(voteItem.Validators) == 3)
	assert.True(t, voteItem.VrfValidators == nil)
	assert.True(t, voteItem.NoVrfValidators == nil)

	vMgr.VrfValidators = ttypes.NewValidatorSet(vMgr.Validators.Validators)
	vMgr.ShuffleType = ShuffleTypeVrf
	vMgr.LastCycleBoundaryInfo = &dty.DposCBInfo{
		Cycle:      110,
		StopHeight: 1111,
		StopHash:   "abcdefg",
		Pubkey:     "xxxxxxxx",
	}

	voteItem = &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	assert.True(t, voteItem.LastCBInfo != nil)
	assert.True(t, voteItem.LastCBInfo.Cycle == 110)
	assert.True(t, voteItem.LastCBInfo.StopHeight == 1111)
	assert.True(t, voteItem.LastCBInfo.StopHash == "abcdefg")
	assert.True(t, voteItem.ShuffleType == ShuffleTypeVrf)
	assert.True(t, len(voteItem.Validators) == 3)
	assert.True(t, len(voteItem.VrfValidators) == 3)
	assert.True(t, voteItem.NoVrfValidators == nil)

	vMgr.ShuffleType = ShuffleTypePartVrf
	val, flag := vMgr.VrfValidators.Remove(vMgr.Validators.Validators[2].Address)
	assert.True(t, flag == true)

	vMgr.NoVrfValidators = &ttypes.ValidatorSet{}
	vMgr.NoVrfValidators.Validators = append(vMgr.NoVrfValidators.Validators, val)

	assert.True(t, len(vMgr.VrfValidators.Validators) == 2)
	assert.True(t, len(vMgr.NoVrfValidators.Validators) == 1)

	voteItem = &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	assert.True(t, voteItem.LastCBInfo != nil)
	assert.True(t, voteItem.ShuffleType == ShuffleTypePartVrf)
	fmt.Println(len(voteItem.Validators))
	assert.True(t, len(voteItem.Validators) == 3)
	assert.True(t, len(voteItem.VrfValidators) == 2)
	assert.True(t, len(voteItem.NoVrfValidators) == 1)
}

func TestUpdateFromVoteItem(t *testing.T) {

	vMgr, err := MakeGenesisValidatorMgr(genDoc)
	require.Nil(t, err)
	assert.True(t, vMgr.ChainID == "chain33-Z2cgFj")
	assert.True(t, len(vMgr.AppHash) == 0)
	assert.True(t, len(vMgr.Validators.Validators) == 3)
	assert.True(t, vMgr.VrfValidators == nil)
	assert.True(t, vMgr.NoVrfValidators == nil)
	assert.True(t, vMgr.LastCycleBoundaryInfo == nil)

	vMgr.ShuffleType = ShuffleTypeNoVrf

	voteItem := &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	assert.True(t, voteItem.LastCBInfo == nil)
	assert.True(t, voteItem.ShuffleType == ShuffleTypeNoVrf)
	assert.True(t, len(voteItem.Validators) == 3)
	assert.True(t, voteItem.VrfValidators == nil)
	assert.True(t, voteItem.NoVrfValidators == nil)

	/////
	newMgr := vMgr.Copy()
	newMgr.LastCycleBoundaryInfo = nil

	val, flag := newMgr.Validators.Remove(newMgr.Validators.Validators[0].Address)
	assert.True(t, flag)
	flag = newMgr.UpdateFromVoteItem(voteItem)
	assert.True(t, flag == false)

	/////
	pkbytes, _ := hex.DecodeString(pubkey11)
	pk11, _ := ttypes.ConsensusCrypto.PubKeyFromBytes(pkbytes)
	val.PubKey = pk11.Bytes()
	newMgr.Validators.Add(val)
	flag = newMgr.UpdateFromVoteItem(voteItem)
	assert.True(t, flag == false)

	/////
	vMgr.LastCycleBoundaryInfo = &dty.DposCBInfo{
		Cycle:      110,
		StopHeight: 1111,
		StopHash:   "abcdefg",
		Pubkey:     "xxxxxxxx",
	}
	voteItem = &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)

	newMgr = vMgr.Copy()
	newMgr.LastCycleBoundaryInfo = nil
	newMgr.UpdateFromVoteItem(voteItem)
	assert.True(t, newMgr.LastCycleBoundaryInfo != nil)
	assert.True(t, newMgr.LastCycleBoundaryInfo.Cycle == voteItem.LastCBInfo.Cycle)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHeight == voteItem.LastCBInfo.StopHeight)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHash == voteItem.LastCBInfo.StopHash)

	/////
	newMgr = vMgr.Copy()
	newMgr.LastCycleBoundaryInfo.Cycle = 111110
	newMgr.UpdateFromVoteItem(voteItem)
	assert.True(t, newMgr.LastCycleBoundaryInfo != nil)
	assert.True(t, newMgr.LastCycleBoundaryInfo.Cycle == voteItem.LastCBInfo.Cycle)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHeight == voteItem.LastCBInfo.StopHeight)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHash == voteItem.LastCBInfo.StopHash)

	/////
	vMgr.VrfValidators = ttypes.NewValidatorSet(vMgr.Validators.Validators)
	vMgr.ShuffleType = ShuffleTypeVrf

	voteItem = &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	assert.True(t, voteItem.LastCBInfo != nil)
	assert.True(t, voteItem.LastCBInfo.Cycle == 110)
	assert.True(t, voteItem.LastCBInfo.StopHeight == 1111)
	assert.True(t, voteItem.LastCBInfo.StopHash == "abcdefg")
	assert.True(t, voteItem.ShuffleType == ShuffleTypeVrf)
	assert.True(t, len(voteItem.Validators) == 3)
	assert.True(t, len(voteItem.VrfValidators) == 3)
	assert.True(t, voteItem.NoVrfValidators == nil)

	newMgr = vMgr.Copy()
	newMgr.ShuffleType = ShuffleTypeNoVrf
	newMgr.VrfValidators = nil
	newMgr.UpdateFromVoteItem(voteItem)
	assert.True(t, newMgr.LastCycleBoundaryInfo != nil)
	assert.True(t, newMgr.LastCycleBoundaryInfo.Cycle == voteItem.LastCBInfo.Cycle)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHeight == voteItem.LastCBInfo.StopHeight)
	assert.True(t, newMgr.LastCycleBoundaryInfo.StopHash == voteItem.LastCBInfo.StopHash)
	assert.True(t, newMgr.ShuffleType == ShuffleTypeVrf)
	assert.True(t, len(newMgr.Validators.Validators) == 3)
	assert.True(t, len(newMgr.VrfValidators.Validators) == 3)

	///
	vMgr.ShuffleType = ShuffleTypePartVrf
	val, flag = vMgr.VrfValidators.Remove(vMgr.Validators.Validators[2].Address)
	assert.True(t, flag == true)

	vMgr.NoVrfValidators = &ttypes.ValidatorSet{}
	vMgr.NoVrfValidators.Validators = append(vMgr.NoVrfValidators.Validators, val)

	assert.True(t, len(vMgr.VrfValidators.Validators) == 2)
	assert.True(t, len(vMgr.NoVrfValidators.Validators) == 1)

	voteItem = &ttypes.VoteItem{}
	vMgr.FillVoteItem(voteItem)
	newMgr = vMgr.Copy()
	newMgr.ShuffleType = ShuffleTypeNoVrf
	newMgr.VrfValidators = nil
	newMgr.NoVrfValidators = nil
	newMgr.UpdateFromVoteItem(voteItem)

	assert.True(t, newMgr.LastCycleBoundaryInfo != nil)
	assert.True(t, newMgr.ShuffleType == ShuffleTypePartVrf)
	assert.True(t, len(newMgr.Validators.Validators) == 3)
	assert.True(t, len(newMgr.VrfValidators.Validators) == 2)
	assert.True(t, len(newMgr.NoVrfValidators.Validators) == 1)
}

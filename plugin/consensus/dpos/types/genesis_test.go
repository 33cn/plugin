package types

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	errGenesisFile = `{"genesis_time:"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFX","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""},{"pub_key":{"type":"secp256k1","data":"027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"},"name":""},{"pub_key":{"type":"secp256k1","data":"03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"},"name":""}],"app_hash":null}`

	genesisFile = `{"genesis_time":"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFX","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""},{"pub_key":{"type":"secp256k1","data":"027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"},"name":""},{"pub_key":{"type":"secp256k1","data":"03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"},"name":""}],"app_hash":null}`
)

func init() {
	//为了使用VRF，需要使用SECP256K1体系的公私钥
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic("init ConsensusCrypto failed.")
	}

	ConsensusCrypto = cr
}

func TestGenesisDocFromFile(t *testing.T) {
	os.Remove("./genesis.json")
	os.Remove("../genesis.json")

	ioutil.WriteFile("genesis.json", []byte(genesisFile), 0664)

	genDoc, err := GenesisDocFromFile("../genesis.json")
	require.NotNil(t, err)
	require.Nil(t, genDoc)

	genDoc, err = GenesisDocFromFile("./genesis.json")
	require.NotNil(t, genDoc)
	require.Nil(t, err)
	os.Remove("./genesis.json")
}

func TestGenesisDocFromJSON(t *testing.T) {
	genDoc, err := GenesisDocFromJSON([]byte(genesisFile))
	require.NotNil(t, genDoc)
	require.Nil(t, err)
	assert.True(t, genDoc.ChainID == "chain33-Z2cgFX")
	assert.True(t, genDoc.AppHash == nil)
	assert.True(t, len(genDoc.Validators) == 3)

	genDoc, err = GenesisDocFromJSON([]byte(errGenesisFile))
	require.NotNil(t, err)
	require.Nil(t, genDoc)
}

func TestSaveAs(t *testing.T) {
	genDoc, err := GenesisDocFromJSON([]byte(genesisFile))
	require.NotNil(t, genDoc)
	require.Nil(t, err)
	assert.True(t, genDoc.ChainID == "chain33-Z2cgFX")
	assert.True(t, genDoc.AppHash == nil)
	assert.True(t, len(genDoc.Validators) == 3)

	err = genDoc.SaveAs("./tmp_genesis.json")
	require.Nil(t, err)

	genDoc2, err := GenesisDocFromFile("./tmp_genesis.json")
	require.NotNil(t, genDoc2)
	require.Nil(t, err)
	//assert.True(t, genDoc.ChainID == genDoc2.ChainID)
	//assert.True(t, genDoc.GenesisTime == genDoc2.GenesisTime)
	//assert.True(t, bytes.Equal(genDoc.AppHash, genDoc2.AppHash))

	assert.True(t, genDoc.Validators[0].Name == genDoc2.Validators[0].Name)
	assert.True(t, genDoc.Validators[0].PubKey.Data == genDoc2.Validators[0].PubKey.Data)
	assert.True(t, genDoc.Validators[0].PubKey.Kind == genDoc2.Validators[0].PubKey.Kind)

	assert.True(t, genDoc.Validators[1].Name == genDoc2.Validators[1].Name)
	assert.True(t, genDoc.Validators[1].PubKey.Data == genDoc2.Validators[1].PubKey.Data)
	assert.True(t, genDoc.Validators[1].PubKey.Kind == genDoc2.Validators[1].PubKey.Kind)

	assert.True(t, genDoc.Validators[2].Name == genDoc2.Validators[2].Name)
	assert.True(t, genDoc.Validators[2].PubKey.Data == genDoc2.Validators[2].PubKey.Data)
	assert.True(t, genDoc.Validators[2].PubKey.Kind == genDoc2.Validators[2].PubKey.Kind)

	err = os.Remove("./tmp_genesis.json")
	require.Nil(t, err)

}

func TestValidateAndComplete(t *testing.T) {
	genDoc, err := GenesisDocFromJSON([]byte(genesisFile))
	require.NotNil(t, genDoc)
	require.Nil(t, err)

	tt := genDoc.GenesisTime
	setSize := len(genDoc.Validators)
	err = genDoc.ValidateAndComplete()
	assert.True(t, tt == genDoc.GenesisTime)
	require.Nil(t, err)
	assert.True(t, setSize == len(genDoc.Validators))

	vals := genDoc.Validators
	genDoc.Validators = nil
	err = genDoc.ValidateAndComplete()
	require.NotNil(t, err)

	genDoc.Validators = vals
	genDoc.ChainID = ""
	err = genDoc.ValidateAndComplete()
	require.NotNil(t, err)
}

func TestValidatorHash(t *testing.T) {
	genDoc, err := GenesisDocFromJSON([]byte(genesisFile))
	require.NotNil(t, genDoc)
	require.Nil(t, err)

	hash := genDoc.ValidatorHash()
	assert.True(t, len(hash) > 0)
}

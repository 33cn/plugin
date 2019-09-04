package types

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	privValidatorFile = `{"address":"2FA286246F0222C4FF93210E91AECE0C66723F15","pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"last_height":1679,"last_round":0,"last_step":3,"last_signature":{"type":"secp256k1","data":"37892A916D6E487ADF90F9E88FE37024597677B6C6FED47444AD582F74144B3D6E4B364EAF16AF03A4E42827B6D3C86415D734A5A6CCA92E114B23EB9265AF09"},"last_signbytes":"7B22636861696E5F6964223A22636861696E33332D5A326367466A222C22766F7465223A7B22626C6F636B5F6964223A7B2268617368223A224F6A657975396B2B4149426A6E4859456739584765356A7A462B673D222C227061727473223A7B2268617368223A6E756C6C2C22746F74616C223A307D7D2C22686569676874223A313637392C22726F756E64223A302C2274696D657374616D70223A22323031382D30382D33315430373A35313A34332E3935395A222C2274797065223A327D7D","priv_key":{"type":"secp256k1","data":"5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"}}`
	strAddr           = "2FA286246F0222C4FF93210E91AECE0C66723F15"
	strPubkey         = "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"

	addr1 = "79F9608B6826762CACCA843E81AE86837ABFFB21"
	addr2 = "3480088E35099CBA75958DAE7A364A8AAD2C1BD0"
	addr3 = "9FF8678DBDA4EAE2F999CBFBCBD8F5F3FC47FBAE"
	addr4 = "70A51AD9777EF1F97250F7E4C156D8637BC7143C"
)

func init() {
	//为了使用VRF，需要使用SECP256K1体系的公私钥
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic("init ConsensusCrypto failed.")
	}

	ConsensusCrypto = cr
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

func read(filename string) bool {
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println("err=", err)
		return false
	}

	defer f.Close()
	buf := make([]byte, 1024*2)
	_, err1 := f.Read(buf)
	if err1 != nil && err1 != io.EOF {
		fmt.Println("err1=", err1)
		return false
	}

	//fmt.Println("buf=",string(buf[:n]))
	return true
}

func TestLoadOrGenPrivValidatorFS(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)
	require.NotNil(t, privValidator)

	assert.True(t, strings.EqualFold(strAddr, hex.EncodeToString(privValidator.GetAddress())))
	assert.True(t, strings.EqualFold(strPubkey, hex.EncodeToString(privValidator.GetPubKey().Bytes())))

	fmt.Println(privValidator.String())

	remove(filename)
}

func TestGenPrivValidatorImp(t *testing.T) {
	filename := "tmp_priv_validator2.json"
	//save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	require.NotNil(t, privValidator)

	assert.True(t, true == read(filename))
	remove(filename)
	assert.True(t, false == read(filename))
	privValidator.Reset()
	assert.True(t, true == read(filename))

	assert.True(t, len(privValidator.GetPubKey().Bytes()) > 0)
	assert.True(t, len(privValidator.GetAddress()) > 0)
	remove(filename)

}

func TestPrivValidatorImpSort(t *testing.T) {
	var arr []*PrivValidatorImp

	Addr1, _ := hex.DecodeString(addr1)
	Addr2, _ := hex.DecodeString(addr2)
	Addr3, _ := hex.DecodeString(addr3)
	Addr4, _ := hex.DecodeString(addr4)

	imp1 := &PrivValidatorImp{
		Address: Addr1,
	}
	arr = append(arr, imp1)

	imp2 := &PrivValidatorImp{
		Address: Addr2,
	}
	arr = append(arr, imp2)

	imp3 := &PrivValidatorImp{
		Address: Addr3,
	}
	arr = append(arr, imp3)

	imp4 := &PrivValidatorImp{
		Address: Addr4,
	}
	arr = append(arr, imp4)

	sort.Sort(PrivValidatorsByAddress(arr))

	assert.True(t, strings.EqualFold(addr2, hex.EncodeToString(arr[0].Address)))
	assert.True(t, strings.EqualFold(addr4, hex.EncodeToString(arr[1].Address)))
	assert.True(t, strings.EqualFold(addr1, hex.EncodeToString(arr[2].Address)))
	assert.True(t, strings.EqualFold(addr3, hex.EncodeToString(arr[3].Address)))
}

func TestSignAndVerifyVote(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	now := time.Now().Unix()
	//task := dpos.DecideTaskByTime(now)
	//生成vote， 对于vote进行签名
	voteItem := &VoteItem{
		VotedNodeAddress: privValidator.Address,
		VotedNodeIndex:   int32(0),
		Cycle:            100,
		CycleStart:       18888,
		CycleStop:        28888,
		PeriodStart:      20000,
		PeriodStop:       21000,
		Height:           100,
	}
	encode, err := json.Marshal(voteItem)
	if err != nil {
		panic("Marshal vote failed.")
	}

	voteItem.VoteID = crypto.Ripemd160(encode)

	vote := &Vote{
		DPosVote: &DPosVote{
			VoteItem:         voteItem,
			VoteTimestamp:    now,
			VoterNodeAddress: privValidator.GetAddress(),
			VoterNodeIndex:   int32(0),
		},
	}
	assert.True(t, 0 == len(vote.Signature))

	chainID := "test-chain-Ep9EcD"
	privValidator.SignVote(chainID, vote)

	assert.True(t, 0 < len(vote.Signature))

	vote2 := vote.Copy()
	err = vote2.Verify(chainID, privValidator.PubKey)
	require.Nil(t, err)
	remove(filename)

	privValidator2 := LoadOrGenPrivValidatorFS(filename)
	require.NotNil(t, privValidator2)

	err = vote2.Verify(chainID, privValidator2.PubKey)
	require.NotNil(t, err)

	remove(filename)
}

func TestSignAndVerifyNotify(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	now := time.Now().Unix()
	//task := dpos.DecideTaskByTime(now)
	//生成vote， 对于vote进行签名
	voteItem := &VoteItem{
		VotedNodeAddress: privValidator.Address,
		VotedNodeIndex:   int32(0),
		Cycle:            100,
		CycleStart:       18888,
		CycleStop:        28888,
		PeriodStart:      20000,
		PeriodStop:       21000,
		Height:           100,
	}
	encode, err := json.Marshal(voteItem)
	if err != nil {
		panic("Marshal vote failed.")
	}

	voteItem.VoteID = crypto.Ripemd160(encode)

	chainID := "test-chain-Ep9EcD"

	notify := &Notify{
		DPosNotify: &DPosNotify{
			Vote:              voteItem,
			HeightStop:        200,
			HashStop:          []byte("abcdef121212"),
			NotifyTimestamp:   now,
			NotifyNodeAddress: privValidator.GetAddress(),
			NotifyNodeIndex:   int32(0),
		},
	}

	err = privValidator.SignNotify(chainID, notify)
	require.Nil(t, err)

	notify2 := notify.Copy()
	err = notify2.Verify(chainID, privValidator.PubKey)
	require.Nil(t, err)
	remove(filename)

	privValidator2 := LoadOrGenPrivValidatorFS(filename)
	require.NotNil(t, privValidator2)

	err = notify2.Verify(chainID, privValidator2.PubKey)
	require.NotNil(t, err)

	remove(filename)
}

func TestSignMsg(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	byteCB := []byte("asdfadsasf")

	sig, err := privValidator.SignMsg(byteCB)
	require.Nil(t, err)
	assert.True(t, 0 < len(sig.Bytes()))

	remove(filename)
}

func TestVrf(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	input := []byte("abcdefghijklmn")
	hash, proof := privValidator.VrfEvaluate(input)
	assert.True(t, 32 == len(hash))
	assert.True(t, 0 < len(proof))

	result := privValidator.VrfProof(privValidator.PubKey.Bytes(), input, hash, proof)
	assert.True(t, result)

	remove(filename)
}

func TestSignTx(t *testing.T) {
	filename := "./tmp_priv_validator.json"
	save(filename, privValidatorFile)
	privValidator := LoadOrGenPrivValidatorFS(filename)

	tx := &types.Transaction{}
	privValidator.SignTx(tx)
	assert.True(t, types.SECP256K1 == tx.Signature.Ty)
	assert.True(t, bytes.Equal(privValidator.PubKey.Bytes(), tx.Signature.Pubkey))
	assert.True(t, 0 < len(tx.Signature.Signature))

	remove(filename)
}

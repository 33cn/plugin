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

	"github.com/33cn/chain33/common/address"
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
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
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

func TestPubkeyAndAddress(t *testing.T) {
	priv := "880D055D311827AB427031959F43238500D08F3EDF443EB903EC6D1A16A5783A"
	pub := "026BD23C69F9A1D7A50F185514EDCA25AFD00DD2A92485CC1E8CDA3EDD284CA838"
	addr := "22F4EA6D79D5AD0621900AB25A2BF01D3E288A7B"

	testOneKey(priv, pub, addr, t)

	priv = "2C1B7E0B53548209E6915D8C5EB3162E9AF43D051C1316784034C9D549DD5F61"
	pub = "02DE57B5427CE38E7D9811AC7CAE51A6FCFD251BCAD88DBDCF636EC9F7441B6AB6"
	addr = "F518C3964A84EF2E37FD3284B5C638D3C928C537"

	testOneKey(priv, pub, addr, t)

	priv = "2EA198F3F063F69B2DB860F41F8FAACB2EDEBCA1A74A75A5254B84E6F7D154B4"
	pub = "036A818C01FA49F455E3779D073BD5FBE551A3E60BD447D8CFEE87F95037C29E59"
	addr = "3FBB75FDC792E2618DA5D39B6C64B718AC0AAA5E"
	testOneKey(priv, pub, addr, t)

	priv = "EEDD4815535AD81B6191EF9883346702C5117C80F8D85EB36FBD77D6CC979C8F"
	pub = "023D3E5DDAD1F20C28E333FF05C16539ECB95CFAA881B66ADFCB2A5198E5BC0EFE"
	addr = "ACB850206A75F233DDBB56A4D7DA6A015C3FE2EB"
	testOneKey(priv, pub, addr, t)

	priv = "4EBB8B86F95DD7AB25CC27928D9AE04692DA9D7CA822CDA3FDE1DE778F01AFA7"
	pub = "03DC5D8106E9E19EDFF91687CE55C97F3F6FA7584678313CF14925F76D0DB09055"
	addr = "5A60D96560192EEE8EC9091FCC1AFC5511CA0BF0"
	testOneKey(priv, pub, addr, t)
}

func testOneKey(priv, pub, addr string, t *testing.T) {
	tmp, _ := hex.DecodeString(priv)
	bPriv, _ := ConsensusCrypto.PrivKeyFromBytes(tmp)
	fmt.Println(fmt.Sprintf("%x", bPriv))

	bPub, _ := PubKeyFromString(pub)
	assert.True(t, bytes.Equal(bPub.Bytes(), bPriv.PubKey().Bytes()))

	bAddr := address.BytesToBtcAddress(address.NormalVer, bPub.Bytes()).Hash160[:]
	fmt.Println("addr:", addr)
	fmt.Println(fmt.Sprintf("%x", bAddr))
	assert.True(t, addr == fmt.Sprintf("%X", bAddr))
}

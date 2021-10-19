package secp256k1

import (
	"fmt"
	"testing"

	"gotest.tools/assert"

	"github.com/33cn/chain33/common/crypto"

	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

//secp256k1签名DER编码格式
// 0x30 <length> 0x02
//<length r> r
//0x02 <length s>
//s

//30440220
//7C12FF568B6DA03EF2CD5681EE45EDF846172771AC6F9369B50FEBC95B1CF68F
//0220
//8B6916A3CC7423D9044E77ABECE410B69B7BD82C22ECCC76061B4BE79141D1A3
func Test_VerifySecp256k1SigFromTass_forChain33(t *testing.T) {
	require := require.New(t)

	//var pubKey secp256k1.PubKeySecp256k1

	pubBytes := common.FromHex("04C24FBA65F8CD81223D2935EDEA663048A1BEFB5A78BC67C80DCB5A1D601F898C35EA242D2E76CACE9EE5A61DBDA29A5076707325FE20B5A80DB0CA6D02C5D983")

	secpPubKey, err := ethCrypto.UnmarshalPubkey(pubBytes)
	require.Equal(nil, err)
	pub33Bytes := ethCrypto.CompressPubkey(secpPubKey)

	c := &secp256k1.Driver{}
	pubKey, err := c.PubKeyFromBytes(pub33Bytes)
	require.Equal(nil, err)

	//msg := []byte("12345678123456781234567812345678")
	msg := []byte("456789")

	hash := crypto.Sha256(msg)
	fmt.Println("hash = ", common.Bytes2Hex(hash))
	//0xfed9efbd5a8ef6820d639dbcb831daf9d6308312cc73d6188beb54a9a148e29a

	sig, err := c.SignatureFromBytes(common.FromHex("304502207C12FF568B6DA03EF2CD5681EE45EDF846172771AC6F9369B50FEBC95B1CF68F0221008B6916A3CC7423D9044E77ABECE410B69B7BD82C22ECCC76061B4BE79141D1A3"))
	//sig, err := c.SignatureFromBytes(common.FromHex("304402207C12FF568B6DA03EF2CD5681EE45EDF846172771AC6F9369B50FEBC95B1CF68F02208B6916A3CC7423D9044E77ABECE410B69B7BD82C22ECCC76061B4BE79141D1A3"))
	require.Equal(nil, err)

	result := pubKey.VerifyBytes(msg, sig)
	require.Equal(true, result)
}

//在以太坊上的验证签名的有效性
//注意：从加密中导出的签名信息中RS信息中的首字节必须大于０，否则签名验证失败
func Test_Verify4Eth(t *testing.T) {
	pub := common.FromHex("04C24FBA65F8CD81223D2935EDEA663048A1BEFB5A78BC67C80DCB5A1D601F898C35EA242D2E76CACE9EE5A61DBDA29A5076707325FE20B5A80DB0CA6D02C5D983")
	sig := common.FromHex("2F2F8EF10E6C9075CAB44DE3C4F904817220537C1E7DCFADD502C03F14F5B3974C405EA9BB189B85F15B91C82CE5D6191D66238ECCCE83FA8F8FF83173F1586F00")

	msg := []byte("456789")
	hash := crypto.Sha256(msg)
	fmt.Println("hash = ", common.Bytes2Hex(hash))

	pubRecoverd, err := ethCrypto.Ecrecover(hash[:], sig)
	require.Equal(t, nil, err)
	fmt.Println("pubRecoverd = ", common.Bytes2Hex(pubRecoverd))

	VerifyResult := ethCrypto.VerifySignature(pub, hash[:], sig[:64])
	assert.Equal(t, true, VerifyResult)
}

func Test_secp256k1(t *testing.T) {
	require := require.New(t)

	c := &secp256k1.Driver{}

	priv, err := c.PrivKeyFromBytes(common.FromHex("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))
	require.Nil(err)
	t.Logf("priv:%X, len:%d", priv.Bytes(), len(priv.Bytes()))

	pub := priv.PubKey()
	require.NotNil(pub)
	t.Logf("pub:%X, len:%d", pub.Bytes(), len(pub.Bytes()))

	//msg := []byte("12345678123456781234567812345678")
	//msg := []byte("hello world")
	msg := []byte("456789")
	signature := priv.Sign(msg)
	t.Logf("sign:%X, len:%d", signature.Bytes(), len(signature.Bytes()))
	t.Logf("signature in hex format:%s", common.Bytes2Hex(signature.Bytes()))
	//0x3045022100f4009ab47dc32880b3e0bfad47885e9cfd1fd2228e804b38fb7f0f5ea6c02405022061422eb681fdd5078aa3971770cf22ce4ef12e9116995e4a3e141e23f5403014
	ok := pub.VerifyBytes(msg, signature)
	require.Equal(true, ok)
}

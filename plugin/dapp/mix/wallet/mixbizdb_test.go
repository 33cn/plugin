package wallet

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	mixTy "github.com/33cn/plugin/plugin/dapp/mix/types"
	"github.com/stretchr/testify/assert"
)

func TestNewPrivacyWithPrivKey(t *testing.T) {
	prikey := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	keyByte, err := hex.DecodeString(prikey)
	assert.Equal(t, nil, err)

	pairs := newPrivacyKey(keyByte)

	t.Log("payPri", pairs.PaymentKey.SpendKey, "payPub", pairs.PaymentKey.ReceiveKey)
	t.Log("crytoPub", pairs.SecretKey.SecretPubKey, "crytoPri", pairs.SecretKey.SecretPrivKey)

	//prikey2 := "1257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	//keyByte2, err := hex.DecodeString(prikey2)
	//assert.Equal(t, nil, err)
	//pairs2, err := newPrivacyKey(keyByte2)
	//assert.Equal(t, nil, err)
	//t.Log("payPri2", pairs2.PaymentKey.SpendKey, "payPub", pairs2.PaymentKey.ReceiveKey, "crytoPub", pairs2.EncryptKey.PubKey, "crytoPri", pairs2.EncryptKey.PrivKey)

	secret1 := &mixTy.SecretData{
		ReceiverKey:  "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizeKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		NoteRandom:   "2824204835",
		Amount:       "28242048",
	}

	data, err := encryptData(pairs.SecretKey.SecretPubKey, types.Encode(secret1))
	assert.Nil(t, err)
	crypData, err := common.FromHex(data.Secret)
	assert.Nil(t, err)
	decryData1, err := decryptData(pairs.SecretKey.SecretPrivKey, data.OneTimePubKey, crypData)
	assert.Nil(t, err)
	var val mixTy.SecretData
	err = types.Decode(decryData1, &val)
	assert.Nil(t, err)
	assert.Equal(t, secret1.ReceiverKey, val.ReceiverKey)

}

func TestEncrypt(t *testing.T) {

	secret := &mixTy.SecretData{
		ReceiverKey:  "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizeKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		NoteRandom:   "2824204835",
		Amount:       "28242048",
	}

	password := "1314fuzamei"

	cryptData := encryptDataWithPadding([]byte(password), types.Encode(secret))
	decryptData, err := decryptDataWithPading([]byte(password), cryptData)
	assert.Nil(t, err)
	var raw mixTy.SecretData
	err = types.Decode(decryptData, &raw)
	assert.Nil(t, err)
	assert.Equal(t, raw.ReceiverKey, secret.ReceiverKey)

}

func TestEncodeSecretData(t *testing.T) {
	secret := &mixTy.SecretData{
		ReceiverKey:  "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizeKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		Amount:       "28242048",
	}

	//ret, err := encodeSecretData(secret)
	//assert.Nil(t, err)
	//t.Log(ret)

	//test encryp data
	prikey := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	keyByte, err := hex.DecodeString(prikey)
	assert.Equal(t, nil, err)
	privacy := newPrivacyKey(keyByte)

	ret := types.Encode(secret)
	hexRet := hex.EncodeToString(ret)
	//assert.Nil(t,err)

	req := &mixTy.EncryptSecretData{PeerSecretPubKey: privacy.SecretKey.SecretPubKey, Secret: hexRet}
	dhSecret, err := encryptSecretData(req)
	assert.Nil(t, err)
	//t.Log(dhSecret)

	data, err := common.FromHex(dhSecret.Secret)
	assert.Nil(t, err)
	rawData, err := decryptData(privacy.SecretKey.SecretPrivKey, dhSecret.OneTimePubKey, data)
	assert.Nil(t, err)
	var rawSecret mixTy.SecretData
	types.Decode(rawData, &rawSecret)
	assert.Equal(t, rawSecret.ReceiverKey, secret.ReceiverKey)
}

func TestMimcHashPriKey(t *testing.T) {
	//frModValue := "21888242871839275222246405745257275088548364400416034343698204186575808495617"
	a := "11888242871839275222246405745257275088548364400416034343698204186575808495617"
	b := "31888242871839275222246405745257275088548364400416034343698204186575808495617"

	var bigVal, lessVal big.Int
	lessVal.SetString(a, 10)
	bigVal.SetString(b, 10)

	//测试大值
	prikey := bigVal.Bytes()
	//prikey对modValue取了模，spendkey 不一定和Prikey相等了， lessval相等，bigVal不相等
	spendKey := mixTy.Byte2Str(prikey[:])
	t.Log("prikey ori", hex.EncodeToString(prikey), "prikey spendkey", hex.EncodeToString(mixTy.Str2Byte(spendKey)))

	//mimcHash 会对每个val取模，这样又相等了
	rcvSpenPri := mimcHashByte([][]byte{mixTy.Str2Byte(spendKey)})
	//payPrivKey 可能超出fr的模，spendKey是payPrivKey对fr取的模，有可能和payPrivKey不相等，这里需要用spendKey取hash
	rcvKeyPri := mimcHashByte([][]byte{prikey})
	receiveKeyBig := mimcHashByte([][]byte{mixTy.Str2Byte(b)})
	t.Log("spendPri", hex.EncodeToString(rcvSpenPri), "prikey", hex.EncodeToString(rcvKeyPri), "big", hex.EncodeToString(receiveKeyBig))

	assert.Equal(t, rcvSpenPri, rcvKeyPri)
}

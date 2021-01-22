package wallet

import (
	"encoding/hex"
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

	pairs, err := newPrivacyWithPrivKey(keyByte)
	assert.Equal(t, nil, err)
	t.Log("payPri", pairs.PaymentKey.SpendKey, "payPub", pairs.PaymentKey.PayKey, "crytoPub", pairs.ShareSecretKey.ReceivingPk, "crytoPri", pairs.ShareSecretKey.PrivKey.Data)

	prikey2 := "1257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	keyByte2, err := hex.DecodeString(prikey2)
	assert.Equal(t, nil, err)
	pairs2, err := newPrivacyWithPrivKey(keyByte2)
	assert.Equal(t, nil, err)
	t.Log("payPri", pairs2.PaymentKey.SpendKey, "payPub", pairs2.PaymentKey.PayKey, "crytoPub", pairs2.ShareSecretKey.ReceivingPk, "crytoPri", pairs2.ShareSecretKey.PrivKey.Data)

	secret1 := &mixTy.SecretData{
		PaymentPubKey:   "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnPubKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizePubKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		NoteRandom:      "2824204835",
		Amount:          "28242048",
	}
	//secret2 := &mixTy.CryptoData{
	//	SpendPubKey:"18829345085195922012068709111582461121107908772422825655963168999800303848486",
	//	ReturnPubKey:"16067249407809359746114321133992130903102335882983385972747813693681808870497",
	//	AuthorizePubKey:"13519883267141251871527102103999205179714486518503885909948192364772977661583",
	//	NoteRandom:"2824204835",
	//	Amount:"28242048",
	//}
	data := encryptData(pairs.ShareSecretKey.ReceivingPk, types.Encode(secret1))
	crypData, err := common.FromHex(data.Secret)
	assert.Nil(t, err)
	decryData1, err := decryptData(pairs.ShareSecretKey.PrivKey, data.Epk, crypData)
	assert.Nil(t, err)
	var val mixTy.SecretData
	err = types.Decode(decryData1, &val)
	assert.Nil(t, err)
	assert.Equal(t, secret1.PaymentPubKey, val.PaymentPubKey)

}

func TestEncrypt(t *testing.T) {

	secret1 := &mixTy.SecretData{
		PaymentPubKey:   "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnPubKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizePubKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		NoteRandom:      "2824204835",
		Amount:          "28242048",
	}

	password := "1314fuzamei"

	cryptData := encryptDataWithPadding([]byte(password), types.Encode(secret1))
	decryptData, err := decryptDataWithPading([]byte(password), cryptData)
	assert.Nil(t, err)
	var raw mixTy.SecretData
	err = types.Decode(decryptData, &raw)
	assert.Nil(t, err)
	assert.Equal(t, raw.PaymentPubKey, secret1.PaymentPubKey)

}

func TestEncodeSecretData(t *testing.T) {
	secret := &mixTy.SecretData{
		PaymentPubKey:   "13735985067536865723202617343666111332145536963656464451727087263423649028705",
		ReturnPubKey:    "16067249407809359746114321133992130903102335882983385972747813693681808870497",
		AuthorizePubKey: "13519883267141251871527102103999205179714486518503885909948192364772977661583",
		Amount:          "28242048",
	}

	ret, err := encodeSecretData(secret)
	assert.Nil(t, err)
	t.Log(ret)

	//test encryp data
	prikey := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	keyByte, err := hex.DecodeString(prikey)
	assert.Equal(t, nil, err)
	privacy, err := newPrivacyWithPrivKey(keyByte)
	assert.Equal(t, nil, err)

	req := &mixTy.EncryptSecretData{ReceivingPk: privacy.ShareSecretKey.ReceivingPk, Secret: ret.Encoded}
	dhSecret, err := encryptSecretData(req)
	assert.Nil(t, err)
	t.Log(dhSecret)

	data, err := common.FromHex(dhSecret.Secret)
	assert.Nil(t, err)
	rawData, err := decryptData(privacy.ShareSecretKey.PrivKey, dhSecret.Epk, data)
	assert.Nil(t, err)
	var rawSecret mixTy.SecretData
	types.Decode(rawData, &rawSecret)
	assert.Equal(t, rawSecret.PaymentPubKey, secret.PaymentPubKey)
}

package crypto

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

//DES 加解密测试
func TestAes(t *testing.T) {
	aes := NewAES(keys[2],ivs[0])
	result, err := aes.Encrypt(contents[1])
	if err != nil {
		t.Error(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(result))
	origData, err := aes.Decrypt(result)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, contents[1], origData)
}

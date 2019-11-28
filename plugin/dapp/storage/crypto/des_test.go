package crypto

import (
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	contents = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}

	keys = [][]byte{
		[]byte("123456ab"),
		[]byte("G2F4ED5m123456abx6vDrScs"),
		[]byte("G2F4ED5m123456abx6vDrScsHD3psX7k"),
	}
	ivs = [][]byte{
		[]byte("1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"),
		[]byte("1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"),
		[]byte("1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"),
		[]byte("1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"),
	}
)

//DES 加解密测试
func TestDes(t *testing.T) {
	des := NewDES(keys[0], ivs[0])
	result, err := des.Encrypt(contents[0])
	if err != nil {
		t.Error(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(result))
	origData, err := des.Decrypt(result)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, contents[0], origData)
}

//3DES 加解密测试
func Test3Des(t *testing.T) {
	des := NewTripleDES(keys[1], ivs[1])
	result, err := des.Encrypt(contents[0])
	if err != nil {
		t.Error(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(result))
	origData, err := des.Decrypt(result)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, contents[0], origData)
}

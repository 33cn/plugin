package types

import (
	"encoding/hex"
	fmt "fmt"
	"testing"

	"github.com/33cn/chain33/common/crypto"
)

func genKey(name string) (crypto.PrivKey, error) {
	c, err := crypto.New(name)
	if err != nil {
		return nil, err
	}
	priv, err := c.GenKey()
	if err != nil {
		return nil, err
	}
	return priv, nil
}

func testEd25519(t *testing.T) {
	priv, err := genKey("ed25519")
	if err != nil {
		t.Error(err)
	}

	hash := crypto.Sha256([]byte("1234567890"))
	for i := 0; i < 10; i++ {
		b1 := priv.Sign(hash).Bytes()
		b2 := priv.Sign(hash).Bytes()

		s1 := hex.EncodeToString(b1)
		s2 := hex.EncodeToString(b2)
		fmt.Println(s1, s2)

		if string(s1) != string(s2) {
			t.Error("error")
		}
	}
}

func testSecp256k1(t *testing.T) {
	priv, err := genKey("secp256k1")
	if err != nil {
		t.Error(err)
	}

	hash := crypto.Sha256([]byte("1234567890"))
	for i := 0; i < 10; i++ {
		b1 := priv.Sign(hash).Bytes()
		b2 := priv.Sign(hash).Bytes()

		s1 := hex.EncodeToString(b1)
		s2 := hex.EncodeToString(b2)
		fmt.Println(s1, s2)

		if string(s1) != string(s2) {
			t.Error("error")
		}
	}
}

func TestSignSame(t *testing.T) {
	testEd25519(t)
	testSecp256k1(t)
}

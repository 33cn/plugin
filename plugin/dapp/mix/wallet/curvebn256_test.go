package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSharedSecret(t *testing.T) {

	bn256 := NewCurveBn256ECDH()
	pri1, pub1 := bn256.GenerateKey(nil)
	pri2, pub2 := bn256.GenerateKey(nil)

	s1, _ := bn256.GenerateSharedSecret(pri1, pub2)
	s2, _ := bn256.GenerateSharedSecret(pri2, pub1)

	assert.Equal(t, s1, s2)

}

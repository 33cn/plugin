// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log"

	"github.com/stretchr/testify/assert"
)

type pubKeyMock struct {
}

func (*pubKeyMock) Bytes() []byte {
	return []byte("pubKeyMock")
}

func (*pubKeyMock) KeyString() string {
	return "pubKeyMock"
}

func (*pubKeyMock) VerifyBytes(msg []byte, sig crypto.Signature) bool {
	return true
}

func (*pubKeyMock) Equals(crypto.PubKey) bool {
	return true
}

type signatureMock struct {
}

func (*signatureMock) Bytes() []byte {
	return []byte("signatureMock")
}

func (*signatureMock) IsZero() bool {
	return true
}

func (*signatureMock) String() string {
	return "signatureMock"
}

func (*signatureMock) Equals(crypto.Signature) bool {
	return true
}

func formatByte32(b []byte) []byte {
	var b32 [32]byte
	copy(b32[:], b)
	return b32[:]
}

type privKeyMock struct {
}

func (mock *privKeyMock) Bytes() []byte {
	return formatByte32([]byte("1234"))
}

func (mock *privKeyMock) Sign(msg []byte) crypto.Signature {
	return &signatureMock{}
}

func (mock *privKeyMock) PubKey() crypto.PubKey {
	return &pubKeyMock{}
}

func (mock *privKeyMock) Equals(crypto.PrivKey) bool {
	return true
}

func init() {
	log.SetLogLevel("crit")
}

func TestNewPrivacy(t *testing.T) {
	test_NewPrivacy(t)
	test_NewPrivacyWithPrivKey(t)
	test_GenerateOneTimeAddr(t)
	test_RecoverOnetimePriKey(t)
}

func test_RecoverOnetimePriKey(t *testing.T) {
	R := formatByte32([]byte("1234"))
	pkm := privKeyMock{}
	privKey, err := RecoverOnetimePriKey(R, &pkm, &pkm, 0)
	assert.Nil(t, err)
	assert.NotNil(t, privKey)
}

func test_GenerateOneTimeAddr(t *testing.T) {
	bytes1 := [32]byte{}
	pkot, err := GenerateOneTimeAddr(&bytes1, &bytes1, &bytes1, 0)
	assert.Nil(t, err)
	assert.NotNil(t, pkot)
}

func test_NewPrivacy(t *testing.T) {
	p := NewPrivacy()
	assert.NotNil(t, p)
}

func test_NewPrivacyWithPrivKey(t *testing.T) {
	bytes1 := [KeyLen32]byte{}
	p, err := NewPrivacyWithPrivKey(&bytes1)
	assert.Nil(t, err)
	assert.NotNil(t, p)

}

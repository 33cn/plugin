// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bls

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/common"

	"github.com/33cn/chain33/common/crypto"
	"github.com/stretchr/testify/assert"
)

var blsDrv = &Driver{}

func TestGenKey(t *testing.T) {
	sk, err := blsDrv.GenKey()
	assert.NoError(t, err)
	assert.NotEmpty(t, sk)
	pk := sk.PubKey()
	assert.NotEmpty(t, pk)

	sk2, _ := blsDrv.GenKey()
	assert.NotEqual(t, sk.Bytes(), sk2.Bytes(), "should not generate two same key", sk, sk2)
}

func TestSignAndVerify(t *testing.T) {
	sk, _ := blsDrv.GenKey()
	pk := sk.PubKey()
	m1 := []byte("message to be signed. 将要做签名的消息")
	// sign and verify
	sig1 := sk.Sign(m1)
	ret := pk.VerifyBytes(m1, sig1)
	assert.Equal(t, true, ret)

	// different message should have different signature
	m2 := []byte("message to be signed. 将要做签名的消息.")
	sig2 := sk.Sign(m2)
	assert.NotEqual(t, sig1, sig2, "different message got the same signature", sig1, sig2)

	// different key should have different signature for a same message.
	sk2, _ := blsDrv.GenKey()
	sig12 := sk2.Sign(m1)
	ret = pk.VerifyBytes(m1, sig12)
	assert.Equal(t, false, ret)
}

func TestPrivKeyFromBytes(t *testing.T) {
	keybyte, _ := common.FromHex("0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71")
	p, err := blsDrv.PrivKeyFromBytes(keybyte)
	assert.Nil(t, p)
	assert.NotNil(t, err)

	keybyte, _ = common.FromHex("0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b")
	p, err = blsDrv.PrivKeyFromBytes(keybyte)
	assert.NotNil(t, p)
	assert.Nil(t, err)

}

func TestAggregate(t *testing.T) {
	m := []byte("message to be signed. 将要做签名的消息")
	n := 8
	pubs := make([]crypto.PubKey, 0, n)
	sigs := make([]crypto.Signature, 0, n) //signatures for the same message
	msgs := make([][]byte, 0, n)
	dsigs := make([]crypto.Signature, 0, n) //signatures for each (key,message) pair
	for i := 0; i < n; i++ {
		sk, _ := blsDrv.GenKey()
		pk := sk.PubKey()
		pubs = append(pubs, pk)
		sigs = append(sigs, sk.Sign(m))

		msgi := append(m, byte(i))
		msgs = append(msgs, msgi)
		dsigs = append(dsigs, sk.Sign(msgi))
	}

	asig, err := blsDrv.Aggregate(sigs)
	assert.NoError(t, err)
	// One
	err = blsDrv.VerifyAggregatedOne(pubs, m, asig)
	assert.NoError(t, err)

	apub, err := blsDrv.AggregatePublic(pubs)
	assert.NoError(t, err)

	ret := apub.VerifyBytes(m, asig)
	assert.Equal(t, true, ret)

	// N
	adsig, err := blsDrv.Aggregate(dsigs)
	assert.NoError(t, err)

	err = blsDrv.VerifyAggregatedN(pubs, msgs, adsig)
	assert.NoError(t, err)

	//lose some messages will cause an error
	err = blsDrv.VerifyAggregatedN(pubs, msgs[1:], adsig)
	assert.Error(t, err)

	//with out-of-order public keys, will has no effect on VerifyAggregatedOne, but DO effects VerifyAggregatedN
	pubs[0], pubs[1] = pubs[1], pubs[0]
	err = blsDrv.VerifyAggregatedOne(pubs, m, asig)
	assert.NoError(t, err)

	err = blsDrv.VerifyAggregatedN(pubs, msgs, adsig)
	assert.Error(t, err)

	//invalid length
	_, err = blsDrv.Aggregate(nil)
	assert.Error(t, err)
	_, err = blsDrv.AggregatePublic(make([]crypto.PubKey, 0))
	assert.Error(t, err)
}

//benchmark
func BenchmarkBLSAggregateSignature(b *testing.B) {
	msg := []byte(">16 character identical message")
	n := 200
	sigs := make([]crypto.Signature, 0, n) //signatures for the same message
	for i := 0; i < n; i++ {
		sk, _ := blsDrv.GenKey()
		sigs = append(sigs, sk.Sign(msg))

	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blsDrv.Aggregate(sigs) //nolint:errcheck
	}
}

func BenchmarkBLSSign(b *testing.B) {
	sks := make([]crypto.PrivKey, b.N)
	msgs := make([][]byte, 0, b.N)
	for i := range sks {
		sks[i], _ = blsDrv.GenKey()
		msgs = append(msgs, []byte(fmt.Sprintf("Hello world! 16 characters %d", i)))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sks[i].Sign(msgs[i])
	}
}

func BenchmarkBLSVerify(b *testing.B) {
	sk, _ := blsDrv.GenKey()
	pk := sk.PubKey()
	m := []byte(">16 character identical message")
	sig := sk.Sign(m)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pk.VerifyBytes(m, sig) //nolint:errcheck
	}
}

func BenchmarkBlsManager_VerifyAggregatedOne(b *testing.B) {
	m := []byte("message to be signed. 将要做签名的消息")
	n := 100
	pubs := make([]crypto.PubKey, 0, n)
	sigs := make([]crypto.Signature, 0, n) //signatures for the same message
	for i := 0; i < n; i++ {
		sk, _ := blsDrv.GenKey()
		pk := sk.PubKey()
		pubs = append(pubs, pk)
		sigs = append(sigs, sk.Sign(m))
	}
	asig, _ := blsDrv.Aggregate(sigs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blsDrv.VerifyAggregatedOne(pubs, m, asig) //nolint:errcheck
	}
}

func BenchmarkBlsManager_VerifyAggregatedN(b *testing.B) {
	m := []byte("message to be signed. 将要做签名的消息")
	n := 100
	pubs := make([]crypto.PubKey, 0, n)
	sigs := make([]crypto.Signature, 0, n)
	msgs := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		mi := append(m, byte(i))
		sk, _ := blsDrv.GenKey()
		pk := sk.PubKey()
		pubs = append(pubs, pk)
		sigs = append(sigs, sk.Sign(mi))
		msgs = append(msgs, mi)
	}
	asig, _ := blsDrv.Aggregate(sigs)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		blsDrv.VerifyAggregatedN(pubs, msgs, asig) //nolint:errcheck
	}
}

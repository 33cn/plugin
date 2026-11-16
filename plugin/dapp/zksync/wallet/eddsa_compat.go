// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"io"
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"golang.org/x/crypto/blake2b"
)

// sizeFr BN254 fr 元素字节数，与 gnark-crypto eddsa 包内 sizeFr 一致
const sizeFr = 32

// GenerateKeyCompat 生成与 gnark-crypto v0.5.3 兼容的 eddsa 密钥对。
//
// gnark-crypto v0.5.3 的 eddsa.GenerateKey 在标量字节反转时循环边界写为
// j=sizeFr（标准实现为 j=sizeFr-1），导致把 blake2b 输出的第 33 字节 h[32]
// 混入标量、丢失 h[0]。v0.12.1 修正为标准派生，同一 seed 生成的密钥对不再一致。
//
// zksync 的 layer2 地址（mimc(pubkey)）已按 v0.5.3 的派生固化在链上
// （leaf.Chain33Addr），升级后若改用新派生会导致历史用户 SetPubKey 校验失败、
// 资产不可操作。为保持历史兼容，此处复刻 v0.5.3 的标量构造逻辑，再通过
// v0.12.1 的 PrivateKey.SetBytes 构造密钥对象（Sign/Verify 跨版本一致）。
func GenerateKeyCompat(r io.Reader) (*eddsa.PrivateKey, error) {
	seed := make([]byte, 32)
	if _, err := io.ReadFull(r, seed); err != nil {
		return nil, err
	}
	h := blake2b.Sum512(seed)

	// randSrc = h[32..63]
	var randSrc [32]byte
	for i := 0; i < 32; i++ {
		randSrc[i] = h[i+32]
	}

	// prune the key (RFC 8032 §5.1.5)
	h[0] &= 0xF8
	h[31] &= 0x7F
	h[31] |= 0x40

	// v0.5.3 反转：j 从 sizeFr 开始，额外交换 h[sizeFr]（即 h[32]）
	for i, j := 0, sizeFr; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}
	var scalar [sizeFr]byte
	copy(scalar[:], h[:sizeFr])

	// pubkey = scalar * Base
	var bScalar big.Int
	bScalar.SetBytes(scalar[:])
	c := twistededwards.GetEdwardsCurve()
	var pubA twistededwards.PointAffine
	pubA.ScalarMultiplication(&c.Base, &bScalar)

	// PrivateKey.SetBytes 布局: publicKey(32) || scalar(32) || randSrc(32)
	pubBytes := pubA.Bytes()
	buf := make([]byte, 0, 3*sizeFr)
	buf = append(buf, pubBytes[:]...)
	buf = append(buf, scalar[:]...)
	buf = append(buf, randSrc[:]...)

	var priv eddsa.PrivateKey
	if _, err := priv.SetBytes(buf); err != nil {
		return nil, err
	}
	return &priv, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csp

import "crypto"

const (
	// ECDSA类型
	ECDSAP256KeyGen = 1
	SM2P256KygGen   = 2
)

// 通用key接口
type Key interface {
	Bytes() ([]byte, error)
	SKI() []byte
	Symmetric() bool
	Private() bool
	PublicKey() (Key, error)
}

// 签名器参数接口
type SignerOpts interface {
	crypto.SignerOpts
}

// 证书生成器接口
type CSP interface {
	KeyGen(opts int) (k Key, err error)
	Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error)
}

// key存储接口
type KeyStore interface {
	ReadOnly() bool

	StoreKey(k Key) (err error)
}

// 签名器接口
type Signer interface {
	Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error)
}

// key生成器接口
type KeyGenerator interface {
	KeyGen(opts int) (k Key, err error)
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package csp

import "crypto"

const (
	// ECDSAP256KeyGen ECDSA类型
	ECDSAP256KeyGen = 1
	// SM2P256KygGen SM2类型
	SM2P256KygGen = 2
)

// Key 通用key接口
type Key interface {
	Bytes() ([]byte, error)
	SKI() []byte
	Symmetric() bool
	Private() bool
	PublicKey() (Key, error)
}

// SignerOpts 签名器参数接口
type SignerOpts interface {
	crypto.SignerOpts
}

// CSP 证书生成器接口
type CSP interface {
	KeyGen(opts int) (k Key, err error)
	Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error)
}

// KeyStore key存储接口
type KeyStore interface {
	ReadOnly() bool

	StoreKey(k Key) (err error)
}

// Signer 签名器接口
type Signer interface {
	Sign(k Key, digest []byte, opts SignerOpts) (signature []byte, err error)
}

// KeyGenerator key生成器接口
type KeyGenerator interface {
	KeyGen(opts int) (k Key, err error)
}

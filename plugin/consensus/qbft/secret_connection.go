// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package qbftNode Uses nacl's secret_box to encrypt a net.Conn.
// It is (meant to be) an implementation of the STS protocol.
// Note we do not (yet) assume that a remote peer's pubkey
// is known ahead of time, and thus we are technically
// still vulnerable to MITM. (TODO!)
// See docs/sts-final.pdf for more info
package qbft

import (
	"bytes"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/33cn/chain33/common/crypto"
	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/crypto/nacl/secretbox"
	"golang.org/x/crypto/ripemd160"
)

// 2 + 1024 == 1026 total frame size
const (
	dataLenSize     = 2 // uint16 to describe the length, is <= dataMaxSize
	dataMaxSize     = 1024
	totalFrameSize  = dataMaxSize + dataLenSize
	sealedFrameSize = totalFrameSize + secretbox.Overhead
) // fixed size (length prefixed) byte arrays

// SecretConnection Implements net.Conn
type SecretConnection struct {
	conn       io.ReadWriteCloser
	recvBuffer []byte
	recvNonce  *[24]byte
	sendNonce  *[24]byte
	remPubKey  crypto.PubKey
	shrSecret  *[32]byte // shared secret
}

// MakeSecretConnection Performs handshake and returns a new authenticated SecretConnection.
// Returns nil if error in handshake.
// Caller should call conn.Close()
// See docs/sts-final.pdf for more information.
func MakeSecretConnection(conn io.ReadWriteCloser, locPrivKey crypto.PrivKey) (*SecretConnection, error) {

	locPubKey := locPrivKey.PubKey()

	// Generate ephemeral keys for perfect forward secrecy.
	locEphPub, locEphPriv := genEphKeys()

	// Write local ephemeral pubkey and receive one too.
	// NOTE: every 32-byte string is accepted as a Curve25519 public key
	// (see DJB's Curve25519 paper: http://cr.yp.to/ecdh/curve25519-20060209.pdf)
	remEphPub, err := shareEphPubKey(conn, locEphPub)
	if err != nil {
		return nil, fmt.Errorf("shareEphPubKey: %v", err)
	}

	// Compute common shared secret.
	shrSecret := computeSharedSecret(remEphPub, locEphPriv)

	// Sort by lexical order.
	loEphPub, hiEphPub := sort32(locEphPub, remEphPub)

	// Check if the local ephemeral public key
	// was the least, lexicographically sorted.
	locIsLeast := bytes.Equal(locEphPub[:], loEphPub[:])

	// Generate nonces to use for secretbox.
	recvNonce, sendNonce := genNonces(loEphPub, hiEphPub, locIsLeast)

	// Generate common challenge to sign.
	challenge := genChallenge(loEphPub, hiEphPub)

	// Construct SecretConnection.
	sc := &SecretConnection{
		conn:       conn,
		recvBuffer: nil,
		recvNonce:  recvNonce,
		sendNonce:  sendNonce,
		shrSecret:  shrSecret,
	}

	// Sign the challenge bytes for authentication.
	locSignature := signChallenge(challenge, locPrivKey)

	// Share (in secret) each other's pubkey & challenge signature
	authSigMsg, err := shareAuthSignature(sc, locPubKey, locSignature)
	if err != nil {
		return nil, fmt.Errorf("shareAuthSignature: %v", err)
	}
	remPubKey, remSignature := authSigMsg.Key, authSigMsg.Sig
	if !remPubKey.VerifyBytes(challenge[:], remSignature) {
		return nil, errors.New("Challenge verification failed")
	}

	// We've authorized.
	sc.remPubKey = remPubKey
	return sc, nil
}

// RemotePubKey Returns authenticated remote pubkey
func (sc *SecretConnection) RemotePubKey() crypto.PubKey {
	return sc.remPubKey
}

// Writes encrypted frames of `sealedFrameSize`
// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Write(data []byte) (n int, err error) {
	for 0 < len(data) {
		var frame = make([]byte, totalFrameSize)
		var chunk []byte
		if dataMaxSize < len(data) {
			chunk = data[:dataMaxSize]
			data = data[dataMaxSize:]
		} else {
			chunk = data
			data = nil
		}
		chunkLength := len(chunk)
		binary.BigEndian.PutUint16(frame, uint16(chunkLength))
		copy(frame[dataLenSize:], chunk)

		// encrypt the frame
		var sealedFrame = make([]byte, sealedFrameSize)
		secretbox.Seal(sealedFrame[:0], frame, sc.sendNonce, sc.shrSecret)
		incr2Nonce(sc.sendNonce)
		// end encryption

		_, err := sc.conn.Write(sealedFrame)
		if err != nil {
			return n, err
		}
		n += len(chunk)
	}
	return
}

// CONTRACT: data smaller than dataMaxSize is read atomically.
func (sc *SecretConnection) Read(data []byte) (n int, err error) {
	if 0 < len(sc.recvBuffer) {
		count := copy(data, sc.recvBuffer)
		sc.recvBuffer = sc.recvBuffer[count:]
		return
	}

	sealedFrame := make([]byte, sealedFrameSize)
	_, err = io.ReadFull(sc.conn, sealedFrame)
	if err != nil {
		return
	}

	// decrypt the frame
	var frame = make([]byte, totalFrameSize)
	_, ok := secretbox.Open(frame[:0], sealedFrame, sc.recvNonce, sc.shrSecret)
	if !ok {
		return n, errors.New("Failed to decrypt SecretConnection")
	}
	incr2Nonce(sc.recvNonce)
	// end decryption

	var chunkLength = binary.BigEndian.Uint16(frame) // read the first two bytes
	if chunkLength > dataMaxSize {
		return 0, errors.New("chunkLength is greater than dataMaxSize")
	}
	var chunk = frame[dataLenSize : dataLenSize+chunkLength]

	n = copy(data, chunk)
	sc.recvBuffer = chunk[n:]
	return
}

// Close Implements net.Conn
func (sc *SecretConnection) Close() error { return sc.conn.Close() }

// LocalAddr ...
func (sc *SecretConnection) LocalAddr() net.Addr { return sc.conn.(net.Conn).LocalAddr() }

// RemoteAddr ...
func (sc *SecretConnection) RemoteAddr() net.Addr { return sc.conn.(net.Conn).RemoteAddr() }

// SetDeadline ...
func (sc *SecretConnection) SetDeadline(t time.Time) error { return sc.conn.(net.Conn).SetDeadline(t) }

// SetReadDeadline ...
func (sc *SecretConnection) SetReadDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetReadDeadline(t)
}

// SetWriteDeadline ...
func (sc *SecretConnection) SetWriteDeadline(t time.Time) error {
	return sc.conn.(net.Conn).SetWriteDeadline(t)
}

func genEphKeys() (ephPub, ephPriv *[32]byte) {
	var err error
	ephPub, ephPriv, err = box.GenerateKey(crand.Reader)
	if err != nil {
		ttypes.PanicCrisis("Could not generate ephemeral keypairs")
	}
	return
}

func shareEphPubKey(conn io.ReadWriter, locEphPub *[32]byte) (remEphPub *[32]byte, err error) {
	var err1, err2 error

	Parallel(
		func() {
			_, err1 = conn.Write(locEphPub[:])
		},
		func() {
			remEphPub = new([32]byte)
			_, err2 = io.ReadFull(conn, remEphPub[:])
		},
	)

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	return remEphPub, nil
}

func computeSharedSecret(remPubKey, locPrivKey *[32]byte) (shrSecret *[32]byte) {
	shrSecret = new([32]byte)
	box.Precompute(shrSecret, remPubKey, locPrivKey)
	return
}

func sort32(foo, bar *[32]byte) (lo, hi *[32]byte) {
	if bytes.Compare(foo[:], bar[:]) < 0 {
		lo = foo
		hi = bar
	} else {
		lo = bar
		hi = foo
	}
	return
}

func genNonces(loPubKey, hiPubKey *[32]byte, locIsLo bool) (recvNonce, sendNonce *[24]byte) {
	nonce1 := hash24(append(loPubKey[:], hiPubKey[:]...))
	nonce2 := new([24]byte)
	copy(nonce2[:], nonce1[:])
	nonce2[len(nonce2)-1] ^= 0x01
	if locIsLo {
		recvNonce = nonce1
		sendNonce = nonce2
	} else {
		recvNonce = nonce2
		sendNonce = nonce1
	}
	return
}

func genChallenge(loPubKey, hiPubKey *[32]byte) (challenge *[32]byte) {
	return hash32(append(loPubKey[:], hiPubKey[:]...))
}

func signChallenge(challenge *[32]byte, locPrivKey crypto.PrivKey) (signature crypto.Signature) {
	signature = locPrivKey.Sign(challenge[:])
	return
}

type authSigMessage struct {
	Key crypto.PubKey
	Sig crypto.Signature
}

func shareAuthSignature(sc io.ReadWriter, pubKey crypto.PubKey, signature crypto.Signature) (*authSigMessage, error) {
	var recvMsg authSigMessage
	var err1, err2 error
	pubLen := len(pubKey.Bytes())
	sigLen := len(signature.Bytes())

	Parallel(
		func() {
			msgByte := make([]byte, pubLen+sigLen)
			copy(msgByte, pubKey.Bytes()[:pubLen])
			copy(msgByte[pubLen:], signature.Bytes())
			_, err1 = sc.Write(msgByte)
		},
		func() {
			readBuffer := make([]byte, pubLen+sigLen)
			_, err2 = io.ReadFull(sc, readBuffer)
			if err2 != nil {
				return
			}

			recvMsg.Key, err2 = ttypes.ConsensusCrypto.PubKeyFromBytes(readBuffer[:pubLen])
			if err2 != nil {
				return
			}
			recvMsg.Sig, err2 = ttypes.ConsensusCrypto.SignatureFromBytes(readBuffer[pubLen:])
			if err2 != nil {
				return
			}
		})

	if err1 != nil {
		return nil, err1
	}
	if err2 != nil {
		return nil, err2
	}

	return &recvMsg, nil
}

//--------------------------------------------------------------------------------

// sha256
func hash32(input []byte) (res *[32]byte) {
	hasher := sha256.New()
	_, err := hasher.Write(input) // nolint: errcheck, gas
	if err != nil {
		panic(err)
	}
	resSlice := hasher.Sum(nil)
	res = new([32]byte)
	copy(res[:], resSlice)
	return
}

// We only fill in the first 20 bytes with ripemd160
func hash24(input []byte) (res *[24]byte) {
	hasher := ripemd160.New()
	_, err := hasher.Write(input) // nolint: errcheck, gas
	if err != nil {
		panic(err)
	}
	resSlice := hasher.Sum(nil)
	res = new([24]byte)
	copy(res[:], resSlice)
	return
}

// increment nonce big-endian by 2 with wraparound.
func incr2Nonce(nonce *[24]byte) {
	incrNonce(nonce)
	incrNonce(nonce)
}

// increment nonce big-endian by 1 with wraparound.
func incrNonce(nonce *[24]byte) {
	for i := 23; 0 <= i; i-- {
		nonce[i]++
		if nonce[i] != 0 {
			return
		}
	}
}

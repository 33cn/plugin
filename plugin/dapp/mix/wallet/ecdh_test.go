// Copyright (c) 2016 Andreas Auernhammer. All rights reserved.
// Use of this source code is governed by a license that can be
// found in the LICENSE file.

package wallet

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"testing"
)

// An example for the ECDH key-exchange using Curve25519.
func ExampleX25519() {
	c25519 := X25519()

	privateAlice, publicAlice, err := c25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Failed to generate Alice's private/public key pair: %s\n", err)
	}
	privateBob, publicBob, err := c25519.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Printf("Failed to generate Bob's private/public key pair: %s\n", err)
	}

	if err := c25519.Check(publicBob); err != nil {
		fmt.Printf("Bob's public key is not on the curve: %s\n", err)
	}
	secretAlice := c25519.ComputeSecret(privateAlice, publicBob)

	if err := c25519.Check(publicAlice); err != nil {
		fmt.Printf("Alice's public key is not on the curve: %s\n", err)
	}
	secretBob := c25519.ComputeSecret(privateBob, publicAlice)

	if !bytes.Equal(secretAlice, secretBob) {
		fmt.Printf("key exchange failed - secret X coordinates not equal\n")
	}
	// Output:
}

func ExampleX25519_Params() {
	c25519 := X25519()
	p := c25519.Params()
	fmt.Printf("Name: %s BitSize: %d", p.Name, p.BitSize)
	// Output: Name: Curve25519 BitSize: 255
}

func TestX25519(t *testing.T) {
	dh := X25519()

	secret := make([]byte, 32)
	var priBob [32]byte
	for i := 0; i < 2; i++ {
		priAlice, pubAlice, err := dh.GenerateKey(nil)
		if err != nil {
			t.Fatalf("alice: key pair generation failed: %s", err)
		}

		if _, err := io.ReadFull(rand.Reader, priBob[:]); err != nil {
			t.Fatalf("carol: private key generation failed: %s", err)
		}
		pubBob := dh.PublicKey(&priBob)

		secAlice := dh.ComputeSecret(priAlice, pubBob)
		secBob := dh.ComputeSecret(&priBob, pubAlice)

		if !bytes.Equal(secAlice, secBob) {
			toStr := hex.EncodeToString
			t.Fatalf("DH failed: secrets are not equal:\nAlice got: %s\nBob   got: %s", toStr(secAlice), toStr(secBob))
		}
		if bytes.Equal(secret, secAlice) {
			t.Fatalf("DH generates the same secret all the time")
		}
		copy(secret, secAlice)
	}

}

// Benchmarks

func BenchmarkX25519(b *testing.B) {
	curve := X25519()
	privateAlice, _, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("Failed to generate Alice's private/public key pair: %s", err)
	}
	_, publicBob, err := curve.GenerateKey(rand.Reader)
	if err != nil {
		b.Fatalf("Failed to generate Bob's private/public key pair: %s", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		curve.ComputeSecret(privateAlice, publicBob)
	}
}

func BenchmarkKeyGenerateX25519(b *testing.B) {
	curve := X25519()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			b.Fatalf("Failed to generate Alice's private/public key pair: %s", err)
		}
	}
}

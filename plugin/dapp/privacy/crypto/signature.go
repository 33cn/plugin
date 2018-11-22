// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package privacy

import (
	"bytes"
	"fmt"

	"github.com/33cn/chain33/common/crypto"
)

// SignatureOnetime sinature data type
type SignatureOnetime [64]byte

// SignatureS signature data
type SignatureS struct {
	crypto.Signature
}

// Bytes get bytes
func (sig SignatureOnetime) Bytes() []byte {
	s := make([]byte, 64)
	copy(s, sig[:])
	return s
}

// IsZero check is zero
func (sig SignatureOnetime) IsZero() bool { return len(sig) == 0 }

// String format to string
func (sig SignatureOnetime) String() string {
	fingerprint := make([]byte, len(sig[:]))
	copy(fingerprint, sig[:])
	return fmt.Sprintf("/%X.../", fingerprint)
}

// Equals check signature equal
func (sig SignatureOnetime) Equals(other crypto.Signature) bool {
	if otherEd, ok := other.(SignatureOnetime); ok {
		return bytes.Equal(sig[:], otherEd[:])
	}
	return false
}

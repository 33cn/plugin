// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
)

var (
	verifyKeys    string
	authPubKeys   string
	paymentPubKey string

	commitTreeCurrentStatus string
	commitTreeSubRoots      string
	commitTreeSubLeaves     string
	commitTreeArchiveRoots  string
	commitTreeRootLeaves    string

	authorizeHash      string
	authorizeSpendHash string
	nullifierHash      string
)

func setPrefix() {
	verifyKeys = "mavl-mix-verify-keys-"
	authPubKeys = "mavl-mix-auth-pubkeys-"
	paymentPubKey = "mavl-mix-payment-pubkey-"

	commitTreeCurrentStatus = "mavl-mix-commitTree-current-status-"

	commitTreeSubRoots = "mavl-mix-commitTree-sub-roots-"
	commitTreeSubLeaves = "mavl-mix-commitTree-sub-leaves-"
	commitTreeRootLeaves = "mavl-mix-commitTree-root-Leaves-"
	commitTreeArchiveRoots = "mavl-mix-commitTree-archive-roots-"

	authorizeHash = "mavl-mix-authorizeHash"
	authorizeSpendHash = "mavl-mix-authorizeSpendHash-"
	nullifierHash = "mavl-mix-nullifierHash"

}

//support multi version verify parameter setting
func getVerifyKeysKey(ty int32) []byte {
	return []byte(fmt.Sprintf(verifyKeys+"%d", ty))
}

func getAuthPubKeysKey() []byte {
	return []byte(fmt.Sprintf(authPubKeys))
}

func getPaymentPubKey(addr string) []byte {
	return []byte(fmt.Sprintf(paymentPubKey+"%s", addr))
}

func calcCommitTreeCurrentStatusKey() []byte {
	return []byte(fmt.Sprintf(commitTreeCurrentStatus))
}

func calcArchiveRootsKey(seq uint64) []byte {
	return []byte(fmt.Sprintf(commitTreeArchiveRoots+"%022d", seq))
}

func calcSubRootsKey(seq int32) []byte {
	return []byte(fmt.Sprintf(commitTreeSubRoots+"%010d", seq))
}

func calcSubLeavesKey(seq int32) []byte {
	return []byte(fmt.Sprintf(commitTreeSubLeaves+"%010d", seq))
}

func calcCommitTreeRootLeaves(rootHash string) []byte {
	return []byte(fmt.Sprintf(commitTreeRootLeaves+"%s", rootHash))
}

func calcAuthorizeHashKey(hash string) []byte {
	return []byte(fmt.Sprintf(authorizeHash+"%s", hash))
}

func calcAuthorizeSpendHashKey(hash string) []byte {
	return []byte(fmt.Sprintf(authorizeSpendHash+"%s", hash))
}

func calcNullifierHashKey(hash string) []byte {
	return []byte(fmt.Sprintf(nullifierHash+"%s", hash))
}

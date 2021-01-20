// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
)

var (
	verifyKeys              string
	authPubKeys             string
	paymentPubKey           string
	commitTreeArchiveRoots  string
	commitTreeCurrentRoots  string
	commitTreeCurrentLeaves string
	commitTreeRootLeaves    string
	authorizeHash           string
	authorizeSpendHash      string
	nullifierHash           string
)

func setPrefix() {
	verifyKeys = "mavl-mix-verify-keys-"
	authPubKeys = "mavl-mix-auth-pubkeys-"
	paymentPubKey = "mavl-mix-payment-pubkey-"
	commitTreeArchiveRoots = "mavl-mix-commitTree-roots-archive-"
	commitTreeCurrentRoots = "mavl-mix-commitTree-current-roots"
	commitTreeCurrentLeaves = "mavl-mix-commitTree-current-leaves-"

	commitTreeRootLeaves = "mavl-mix-commitTree-rootLeaves-"
	authorizeHash = "mavl-mix-authorizeHash"
	authorizeSpendHash = "mavl-mix-authorizeHash-spend-"
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

func calcCommitTreeArchiveRootsKey() []byte {
	return []byte(fmt.Sprintf(commitTreeArchiveRoots))
}

func calcCurrentCommitRootsKey() []byte {
	return []byte(fmt.Sprintf(commitTreeCurrentRoots))
}

func calcCurrentCommitLeavesKey() []byte {
	return []byte(fmt.Sprintf(commitTreeCurrentLeaves))
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

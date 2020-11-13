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
	commitTreeArchiveRoots  string
	commitTreeCurrentRoots  string
	commitTreeCurrentLeaves string
	commitTreeRootLeaves    string
	authorizeHash           string
	authorizeSpendHash      string
	nullifierHash           string

	localTx    string
	localTitle string
)

func setPrefix() {
	verifyKeys = "mavl-mixcoin-verify-keys-"
	authPubKeys = "mavl-mixcoin-auth-pubkeys-"
	commitTreeArchiveRoots = "mavl-mixcoin-commitTree-roots-archive-"
	commitTreeCurrentRoots = "mavl-mixcoin-commitTree-current-roots"
	commitTreeCurrentLeaves = "mavl-mixcoin-commitTree-current-leaves-"

	commitTreeRootLeaves = "mavl-mixcoin-commitTree-rootLeaves-"
	authorizeHash = "mavl-mixcoin-authorizeHash"
	authorizeSpendHash = "mavl-mixcoin-authorizeHash-spend-"
	nullifierHash = "mavl-mixcoin-nullifierHash"

}

//support multi version verify parameter setting
func getVerifyKeysKey() []byte {
	return []byte(fmt.Sprintf(verifyKeys))
}

func getAuthPubKeysKey() []byte {
	return []byte(fmt.Sprintf(authPubKeys))
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
	return []byte(fmt.Sprintf(commitTreeRootLeaves+"s", rootHash))
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

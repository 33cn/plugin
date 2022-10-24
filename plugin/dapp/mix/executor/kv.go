// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
)

var (
	verifyKeys   string
	authPubKeys  string
	receivingKey string

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
	receivingKey = "mavl-mix-receiving-key-"

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

func calcReceivingKey(addr string) []byte {
	return []byte(fmt.Sprintf(receivingKey+"%s", address.FormatAddrKey(addr)))
}

func calcCommitTreeCurrentStatusKey(exec, symbol string) []byte {
	return []byte(fmt.Sprintf(commitTreeCurrentStatus+"%s-%s", exec, symbol))
}

func calcArchiveRootsKey(exec, symbol string, seq uint64) []byte {
	return []byte(fmt.Sprintf(commitTreeArchiveRoots+"%s-%s-%022d", exec, symbol, seq))
}

func calcSubRootsKey(exec, symbol string, seq int32) []byte {
	return []byte(fmt.Sprintf(commitTreeSubRoots+"%s-%s-%010d", exec, symbol, seq))
}

func calcSubLeavesKey(exec, symbol string, seq int32) []byte {
	return []byte(fmt.Sprintf(commitTreeSubLeaves+"%s-%s-%010d", exec, symbol, seq))
}

func calcCommitTreeRootLeaves(exec, symbol string, rootHash string) []byte {
	return []byte(fmt.Sprintf(commitTreeRootLeaves+"%s-%s-%s", exec, symbol, rootHash))
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

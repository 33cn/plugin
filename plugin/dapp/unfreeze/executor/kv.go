// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	pty "github.com/33cn/plugin/plugin/dapp/unfreeze/types"
)

var (
	id               = "mavl-" + pty.UnfreezeX + "-"
	initLocal        = "LODB-" + pty.UnfreezeX + "-init-"
	beneficiaryLocal = "LODB-" + pty.UnfreezeX + "-beneficiary-"
)

func unfreezeID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", id, txHash))
}

func initKey(init string) []byte {
	return []byte(fmt.Sprintf("%s%s", initLocal, init))
}

func beneficiaryKey(beneficiary string) []byte {
	return []byte(fmt.Sprintf("%s%s", beneficiaryLocal, beneficiary))
}

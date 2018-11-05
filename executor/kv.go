package executor

import (
	"fmt"

	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
)

var (
	id = "mavl-" + pty.UnfreezeX + "-"
	initLocal = "LODB-" + pty.UnfreezeX + "-init-"
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

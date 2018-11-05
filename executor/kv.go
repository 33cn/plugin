package executor

import (
	"fmt"

	pty "gitlab.33.cn/chain33/chain33/plugin/dapp/unfreeze/types"
)

var (
	id = "mavl-" + pty.UnfreezeX + "-"
)

func unfreezeID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", id, txHash))
}

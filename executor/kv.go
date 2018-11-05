package executor

import "fmt"

var (
	id = "mavl-unfreeze-"
)

func unfreezeID(txHash string) []byte {
	return []byte(fmt.Sprintf("%s%s", id, txHash))
}

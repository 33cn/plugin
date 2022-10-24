package executor

import (
	"testing"

	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
)

func TestGetOperationByChunk(t *testing.T) {
	chunks := []string{
		"105312291742116972659480692884871089622277708374248589119125140943",
		"4456014040586311680015918522164279193879411858150409902812907919443",
		"3410743365975142837943473808261655767383931994113549982714201899008",
	}
	op := getOperationByChunk(chunks, zt.TyDepositAction)
	t.Log("op", op)
	t.Log("op", op.Op.GetDeposit().EthAddress)
}

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
)

func (n *Norm) Exec_Nput(nput *pty.NormPut, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt := &types.Receipt{types.ExecOk, nil, nil}
	normKV := &types.KeyValue{Key(nput.Key), nput.Value}
	receipt.KV = append(receipt.KV, normKV)
	return receipt, nil
}

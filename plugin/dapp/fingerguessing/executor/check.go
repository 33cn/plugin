package executor

import "github.com/33cn/chain33/types"

// 本执行器不做任何校验
func (f *Fingerguessing) CheckTx(tx *types.Transaction, index int) error {
	return nil
}
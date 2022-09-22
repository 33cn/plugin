package executor

import (
	"encoding/hex"
	"github.com/33cn/chain33/types"
	rtypes "github.com/33cn/plugin/plugin/dapp/rollup/types"
)

// CheckTx 实现自定义检验交易接口，供框架调用
func (r *rollup) CheckTx(tx *types.Transaction, index int) error {
	txHash := hex.EncodeToString(tx.Hash())
	var action rtypes.RollupAction
	err := types.Decode(tx.Payload, &action)
	if err != nil {
		elog.Error("rollup CheckTx", "txHash", txHash, "Decode payload error", err)
		return types.ErrActionNotSupport
	}

	if action.Ty == rtypes.TyCommitBatchAction {

		err = r.checkCommitBatch(action.GetCommitBatch())
	}else {
		err = types.ErrActionNotSupport
	}
	if err != nil {
		elog.Error("rollup CheckTx", "txHash", txHash, "actionName", tx.ActionName(), "err", err, "actionData", action.String())
	}
	return err
}



func (r *rollup) checkCommitBatch(commit *rtypes.CommitBatch) error {


}

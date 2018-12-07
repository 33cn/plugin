package executor

import (
	 pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
	"github.com/33cn/chain33/types"
)

func (c *f3d) ExecLocal_Start(payload *pt.F3DStart, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *f3d) ExecLocal_Draw(payload *pt.F3DLuckyDraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *f3d) ExecLocal_Buy(payload *pt.F3DBuyKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

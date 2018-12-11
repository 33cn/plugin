package executor

import (
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

func (c *f3d) ExecDelLocal_Start(payload *pt.F3DStart, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *f3d) ExecDelLocal_Draw(payload *pt.F3DLuckyDraw, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

func (c *f3d) ExecDelLocal_Buy(payload *pt.F3DBuyKey, tx *types.Transaction, receiptData *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	return &types.LocalDBSet{}, nil
}

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

func (l *Lottery) Exec_Create(payload *pty.LotteryCreate, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryCreate(payload)
}

func (l *Lottery) Exec_Buy(payload *pty.LotteryBuy, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryBuy(payload)
}

func (l *Lottery) Exec_Draw(payload *pty.LotteryDraw, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryDraw(payload)
}

func (l *Lottery) Exec_Close(payload *pty.LotteryClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewLotteryAction(l, tx, index)
	return actiondb.LotteryClose(payload)
}

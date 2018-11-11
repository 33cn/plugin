package executor

import (
	"gitlab.33.cn/chain33/chain33/types"
	ty "gitlab.33.cn/chain33/plugin/plugin/dapp/ticket/types"
)

func (t *Ticket) Exec_Genesis(payload *ty.TicketGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.Count <= 0 {
		return nil, ty.ErrTicketCount
	}
	actiondb := NewAction(t, tx)
	return actiondb.GenesisInit(payload)
}

func (t *Ticket) Exec_Topen(payload *ty.TicketOpen, tx *types.Transaction, index int) (*types.Receipt, error) {
	if payload.Count <= 0 {
		tlog.Error("topen ", "value", payload)
		return nil, ty.ErrTicketCount
	}
	actiondb := NewAction(t, tx)
	return actiondb.TicketOpen(payload)
}

func (t *Ticket) Exec_Tbind(payload *ty.TicketBind, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.TicketBind(payload)
}

func (t *Ticket) Exec_Tclose(payload *ty.TicketClose, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.TicketClose(payload)
}

func (t *Ticket) Exec_Miner(payload *ty.TicketMiner, tx *types.Transaction, index int) (*types.Receipt, error) {
	actiondb := NewAction(t, tx)
	return actiondb.TicketMiner(payload, index)
}

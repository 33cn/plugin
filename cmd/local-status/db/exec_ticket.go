package db

import "github.com/33cn/chain33/types"
import (
	"encoding/json"

	rpcTypes "github.com/33cn/chain33/rpc/types"
	logType "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// ticket/ticket/Id
type ticket struct {
	Id      string `json:"id"`
	Address string `json:"address"`
	Height  int64  `json:"height"`
	Ts      int64  `json:"ts"`
}

type bind struct {
	oldMiner      string `json:"old_miner"`
	newMiner      string `json:"new_miner"`
	returnAddress string `json:"return_address"`
	height        int64  `json:"height"`
	ts            int64  `json:"ts"`
}

const (
	//log for ticket

	//TyLogNewTicket new ticket log type
	TyLogNewTicket = 111
	// TyLogCloseTicket close ticket log type
	TyLogCloseTicket = 112
	// TyLogMinerTicket miner ticket log type
	TyLogMinerTicket = 113
	// TyLogTicketBind bind ticket log type
	TyLogTicketBind = 114
)

type ticketConvert struct {
	block *rpcTypes.BlockDetail
}

func (t *ticketConvert) Convert(ty int64, jsonString string) (key []string, prev, current []byte, err error) {
	if ty == types.TyLogFee {
		return LogFeeConvert([]byte(jsonString))
	} else if ty == TyLogNewTicket {
		return t.LogNewTicketConvert([]byte(jsonString))
	} else if ty == TyLogCloseTicket {
		return t.LogCloseTicketConvert([]byte(jsonString))
	} else if ty == TyLogTicketBind {
		return t.LogBindTicketConvert([]byte(jsonString))
	} else if ty == TyLogMinerTicket {
		return t.LogMinerTicketConvert([]byte(jsonString))
	}
	return CommonConverts(ty, []byte(jsonString))
}

func (t *ticketConvert) LogNewTicketConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l logType.ReceiptTicket
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		key = []string{"ticket-new", "ticket", l.TicketId}
		prev, _ = json.Marshal("")
		current, _ = json.Marshal(ticket{
			Id:      l.TicketId,
			Address: l.Addr,
			Height:  t.block.Block.Height,
			Ts:      t.block.Block.BlockTime,
		})
		return
	}
	return convertFailed()
}

func (t *ticketConvert) LogCloseTicketConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l logType.ReceiptTicket
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		key = []string{"ticket-close", "ticket", l.TicketId}
		prev, _ = json.Marshal("")
		current, _ = json.Marshal(ticket{
			Id:      l.TicketId,
			Address: l.Addr,
			Height:  t.block.Block.Height,
			Ts:      t.block.Block.BlockTime,
		})
		return
	}
	return convertFailed()
}

func (t *ticketConvert) LogMinerTicketConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l logType.ReceiptTicket
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		key = []string{"ticket-miner", "ticket", l.TicketId}
		prev, _ = json.Marshal("")
		current, _ = json.Marshal(ticket{
			Id:      l.TicketId,
			Address: l.Addr,
			Height:  t.block.Block.Height,
			Ts:      t.block.Block.BlockTime,
		})
		return
	}
	return convertFailed()
}

func (t *ticketConvert) LogBindTicketConvert(v []byte) (key []string, prev, current []byte, err error) {
	var l logType.ReceiptTicketBind
	err = types.JSONToPB([]byte(v), &l)
	if err == nil {
		key = []string{"ticket-bind", "ticket", l.ReturnAddress}
		prev, _ = json.Marshal("")
		current, _ = json.Marshal(bind{
			oldMiner:      l.OldMinerAddress,
			newMiner:      l.NewMinerAddress,
			returnAddress: l.ReturnAddress,
			height:        t.block.Block.Height,
			ts:            t.block.Block.BlockTime,
		})
		return
	}
	return convertFailed()
}

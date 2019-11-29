// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

// Query_Pos33TicketInfos query tick info
func (ticket *Pos33Ticket) Query_Pos33TicketInfos(param *pty.Pos33TicketInfos) (types.Message, error) {
	return Infos(ticket.GetStateDB(), param)
}

// Query_Pos33TicketList query tick list
func (ticket *Pos33Ticket) Query_Pos33TicketList(param *pty.Pos33TicketList) (types.Message, error) {
	return List(ticket.GetLocalDB(), ticket.GetStateDB(), param)
}

// Query_MinerAddress query miner addr
func (ticket *Pos33Ticket) Query_MinerAddress(param *types.ReqString) (types.Message, error) {
	value, err := ticket.GetLocalDB().Get(calcBindReturnKey(param.Data))
	if value == nil || err != nil {
		return nil, types.ErrNotFound
	}
	return &types.ReplyString{Data: string(value)}, nil
}

// Query_MinerSourceList query miner src list
func (ticket *Pos33Ticket) Query_MinerSourceList(param *types.ReqString) (types.Message, error) {
	key := calcBindMinerKeyPrefix(param.Data)
	values, err := ticket.GetLocalDB().List(key, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}
	reply := &types.ReplyStrings{}
	for _, value := range values {
		reply.Datas = append(reply.Datas, string(value))
	}
	return reply, nil
}

// Query_RandNumHash query randnumhash
func (ticket *Pos33Ticket) Query_RandNumHash(param *types.ReqRandHash) (types.Message, error) {
	return ticket.GetRandNum(param.Hash, param.BlockNum)
}

// Query_Pos33AllPos33TicketCount query all ticket count
func (ticket *Pos33Ticket) Query_Pos33AllPos33TicketCount(param *pty.Pos33AllPos33TicketCount) (types.Message, error) {
	count, err := ticket.getAllPos33TicketCount(param.Height)
	if err != nil {
		return nil, err
	}
	return &pty.ReplyPos33AllPos33TicketCount{Count: int64(count)}, nil
}

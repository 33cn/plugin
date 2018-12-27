// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

// Query_TicketInfos query tick info
func (ticket *Ticket) Query_TicketInfos(param *pty.TicketInfos) (types.Message, error) {
	return Infos(ticket.GetStateDB(), param)
}

// Query_TicketList query tick list
func (ticket *Ticket) Query_TicketList(param *pty.TicketList) (types.Message, error) {
	return List(ticket.GetLocalDB(), ticket.GetStateDB(), param)
}

// Query_MinerAddress query miner addr
func (ticket *Ticket) Query_MinerAddress(param *types.ReqString) (types.Message, error) {
	value, err := ticket.GetLocalDB().Get(calcBindReturnKey(param.Data))
	if value == nil || err != nil {
		return nil, types.ErrNotFound
	}
	return &types.ReplyString{Data: string(value)}, nil
}

// Query_MinerSourceList query miner src list
func (ticket *Ticket) Query_MinerSourceList(param *types.ReqString) (types.Message, error) {
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
func (ticket *Ticket) Query_RandNumHash(param *types.ReqRandHash) (types.Message, error) {
	return ticket.GetRandNum(param.Hash, param.BlockNum)
}

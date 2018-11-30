// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"
	tickettypes "github.com/33cn/plugin/plugin/dapp/ticket/types"
)

const (
	minBlockNum = 3
	maxBlockNum = 10
)

// GetRandNum for ticket executor
func (ticket *Ticket) GetRandNum(height int64, blockNum int64) (types.Message, error) {
	tlog.Debug("GetRandNum", "height", height, "blockNum", blockNum)
	if blockNum < minBlockNum {
		blockNum = minBlockNum
	} else if blockNum > maxBlockNum {
		blockNum = maxBlockNum
	}

	txActions, err := ticket.getTxActions(height, blockNum)
	if err != nil {
		return nil, err
	}
	var modifies []byte
	var bits uint32
	var ticketIds string
	var privHashs []byte

	for _, ticketAction := range txActions {
		//tlog.Debug("GetRandNum", "modify", ticketAction.GetMiner().GetModify(), "bits", ticketAction.GetMiner().GetBits(), "ticketId", ticketAction.GetMiner().GetTicketId(), "PrivHash", ticketAction.GetMiner().GetPrivHash())
		modifies = append(modifies, ticketAction.GetMiner().GetModify()...)
		bits += ticketAction.GetMiner().GetBits()
		ticketIds += ticketAction.GetMiner().GetTicketId()
		privHashs = append(privHashs, ticketAction.GetMiner().GetPrivHash()...)
	}

	newmodify := fmt.Sprintf("%s:%s:%d:%s", string(modifies), ticketIds, bits, string(privHashs))

	modify := common.Sha256([]byte(newmodify))

	return &types.ReplyHash{Hash: modify}, nil
}

func (ticket *Ticket) getTxActions(height int64, blockNum int64) ([]*tickettypes.TicketAction, error) {
	var txActions []*tickettypes.TicketAction

	tlog.Debug("getTxActions", "height", height, "blockNum", blockNum)

	// take attention
	// the user query the height, but maybe the block is not accept yet
	// so the height--
	height--

	req := &types.ReqBlocks{Start: height - blockNum + 1, End: height, IsDetail: false, Pid: []string{""}}

	blockDetails, err := ticket.GetAPI().GetBlocks(req)
	if err != nil {
		tlog.Error("getTxActions", "height", height, "blockNum", blockNum, "err", err)
		return txActions, err
	}
	for _, block := range blockDetails.Items {
		//tlog.Debug("getTxActions", "blockHeight", block.Block.Height, "blockhash", block.Block.Hash())
		ticketAction, err := ticket.getMinerTx(block.Block)
		if err != nil {
			return txActions, err
		}
		txActions = append(txActions, ticketAction)
	}
	return txActions, nil

}

func (ticket *Ticket) getMinerTx(current *types.Block) (*tickettypes.TicketAction, error) {
	//检查第一个笔交易的execs, 以及执行状态
	if len(current.Txs) == 0 {
		return nil, types.ErrEmptyTx
	}
	baseTx := current.Txs[0]
	//判断交易类型和执行情况
	var ticketAction tickettypes.TicketAction
	err := types.Decode(baseTx.GetPayload(), &ticketAction)
	if err != nil {
		return nil, err
	}
	if ticketAction.GetTy() != tickettypes.TicketActionMiner {
		return nil, types.ErrCoinBaseTxType
	}
	//判断交易执行是否OK
	if ticketAction.GetMiner() == nil {
		return nil, tickettypes.ErrEmptyMinerTx
	}
	return &ticketAction, nil
}

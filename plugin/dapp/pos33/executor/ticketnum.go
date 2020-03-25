// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/types"

	tickettypes "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

func (ticket *Pos33Ticket) getTxActions(blockHash []byte, blockNum int64) ([]*tickettypes.Pos33TicketAction, error) {
	var txActions []*tickettypes.Pos33TicketAction
	var reqHashes types.ReqHashes
	currHash := blockHash
	tlog.Debug("getTxActions", "blockHash", common.ToHex(blockHash), "blockNum", blockNum)

	//根据blockHash，查询block，循环blockNum
	for blockNum > 0 {
		req := types.ReqHash{Hash: currHash}

		tempBlock, err := ticket.GetAPI().GetBlockOverview(&req)
		if err != nil {
			return txActions, err
		}
		if tempBlock.Head.Height <= 0 {
			return nil, nil
		}
		reqHashes.Hashes = append(reqHashes.Hashes, currHash)
		currHash = tempBlock.Head.ParentHash
		if tempBlock.Head.Height < 0 && blockNum > 1 {
			return txActions, types.ErrBlockNotFound
		}
		if tempBlock.Head.Height <= 1 {
			break
		}
		blockNum--
	}
	blockDetails, err := ticket.GetAPI().GetBlockByHashes(&reqHashes)
	if err != nil {
		tlog.Error("getTxActions", "blockHash", blockHash, "blockNum", blockNum, "err", err)
		return txActions, err
	}
	cfg := ticket.GetAPI().GetConfig()
	for _, block := range blockDetails.Items {
		tlog.Debug("getTxActions", "blockHeight", block.Block.Height, "blockhash", common.ToHex(block.Block.Hash(cfg)))
		ticketAction, err := ticket.getMinerTx(block.Block)
		if err != nil {
			return txActions, err
		}
		txActions = append(txActions, ticketAction)
	}
	return txActions, nil
}

func (ticket *Pos33Ticket) getMinerTx(current *types.Block) (*tickettypes.Pos33TicketAction, error) {
	//检查第一个笔交易的execs, 以及执行状态
	if len(current.Txs) == 0 {
		return nil, types.ErrEmptyTx
	}
	baseTx := current.Txs[0]
	//判断交易类型和执行情况
	var ticketAction tickettypes.Pos33TicketAction
	err := types.Decode(baseTx.GetPayload(), &ticketAction)
	if err != nil {
		return nil, err
	}
	if ticketAction.GetTy() != tickettypes.Pos33TicketActionMiner {
		return nil, types.ErrCoinBaseTxType
	}
	//判断交易执行是否OK
	if ticketAction.GetMiner() == nil {
		return nil, tickettypes.ErrEmptyMinerTx
	}
	return &ticketAction, nil
}

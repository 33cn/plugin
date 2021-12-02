// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
)

// Query_GetQbftNodeByHeight method
func (val *QbftNode) Query_GetQbftNodeByHeight(in *pty.ReqQbftNodes) (types.Message, error) {
	height := in.GetHeight()

	if height <= 0 {
		return nil, types.ErrInvalidParam
	}
	key := CalcQbftNodeUpdateHeightKey(height)
	values, err := val.GetLocalDB().List(key, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	reply := &pty.QbftNodes{}
	for _, qbftNodeByte := range values {
		var qbftNode pty.QbftNode
		err := types.Decode(qbftNodeByte, &qbftNode)
		if err != nil {
			return nil, err
		}
		reply.Nodes = append(reply.Nodes, &qbftNode)
	}
	return reply, nil
}

// Query_GetBlockInfoByHeight method
func (val *QbftNode) Query_GetBlockInfoByHeight(in *pty.ReqQbftBlockInfo) (types.Message, error) {
	height := in.GetHeight()

	if height <= 0 {
		return nil, types.ErrInvalidParam
	}
	key := CalcQbftNodeBlockInfoHeightKey(height)
	value, err := val.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, types.ErrNotFound
	}

	reply := &pty.QbftBlockInfo{}
	err = types.Decode(value, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// Query_GetCurrentState method
func (val *QbftNode) Query_GetCurrentState(in *types.ReqNil) (types.Message, error) {
	return val.GetAPI().QueryConsensusFunc("qbft", "CurrentState", &types.ReqNil{})
}

// Query_GetPerfState method
func (val *QbftNode) Query_GetPerfStat(in *pty.ReqQbftPerfStat) (types.Message, error) {
	start := in.GetStart()
	end := in.GetEnd()

	if start < 0 || end < 0 || start > end || end > val.GetHeight() {
		return nil, types.ErrInvalidParam
	}
	if start == 0 {
		start = 1
	}
	if end == 0 {
		end = val.GetHeight()
	}

	startKey := CalcQbftNodeBlockInfoHeightKey(start)
	startValue, err := val.GetLocalDB().Get(startKey)
	if err != nil {
		return nil, err
	}
	if len(startValue) == 0 {
		return nil, types.ErrNotFound
	}
	startInfo := &pty.QbftBlockInfo{}
	err = types.Decode(startValue, startInfo)
	if err != nil {
		return nil, err
	}

	endKey := CalcQbftNodeBlockInfoHeightKey(end)
	endValue, err := val.GetLocalDB().Get(endKey)
	if err != nil {
		return nil, err
	}
	if len(endValue) == 0 {
		return nil, types.ErrNotFound
	}
	endInfo := &pty.QbftBlockInfo{}
	err = types.Decode(endValue, endInfo)
	if err != nil {
		return nil, err
	}

	startHeader := startInfo.Block.Header
	endHeader := endInfo.Block.Header
	totalTx := endHeader.TotalTxs - startHeader.TotalTxs + startHeader.NumTxs
	totalBlock := endHeader.Height - startHeader.Height + 1
	totalSecond := endHeader.Time - startHeader.Time + 1
	return &pty.QbftPerfStat{
		TotalTx:     totalTx,
		TotalBlock:  totalBlock,
		TxPerBlock:  totalTx / totalBlock,
		TotalSecond: totalSecond,
		TxPerSecond: totalTx / totalSecond,
	}, nil
}

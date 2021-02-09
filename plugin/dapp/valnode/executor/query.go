// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/valnode/types"
)

// Query_GetValNodeByHeight method
func (val *ValNode) Query_GetValNodeByHeight(in *pty.ReqValNodes) (types.Message, error) {
	height := in.GetHeight()

	if height <= 0 {
		return nil, types.ErrInvalidParam
	}
	key := CalcValNodeUpdateHeightKey(height)
	values, err := val.GetLocalDB().List(key, nil, 0, 1)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	reply := &pty.ValNodes{}
	for _, valnodeByte := range values {
		var valnode pty.ValNode
		err := types.Decode(valnodeByte, &valnode)
		if err != nil {
			return nil, err
		}
		reply.Nodes = append(reply.Nodes, &valnode)
	}
	return reply, nil
}

// Query_GetBlockInfoByHeight method
func (val *ValNode) Query_GetBlockInfoByHeight(in *pty.ReqBlockInfo) (types.Message, error) {
	height := in.GetHeight()

	if height <= 0 {
		return nil, types.ErrInvalidParam
	}
	key := CalcValNodeBlockInfoHeightKey(height)
	value, err := val.GetLocalDB().Get(key)
	if err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, types.ErrNotFound
	}

	reply := &pty.TendermintBlockInfo{}
	err = types.Decode(value, reply)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

// Query_GetPerfState method
func (val *ValNode) Query_GetPerfState(in *pty.ReqPerfStat) (types.Message, error) {
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

	startKey := CalcValNodeBlockInfoHeightKey(start)
	startValue, err := val.GetLocalDB().Get(startKey)
	if err != nil {
		return nil, err
	}
	if len(startValue) == 0 {
		return nil, types.ErrNotFound
	}
	startInfo := &pty.TendermintBlockInfo{}
	err = types.Decode(startValue, startInfo)
	if err != nil {
		return nil, err
	}

	endKey := CalcValNodeBlockInfoHeightKey(end)
	endValue, err := val.GetLocalDB().Get(endKey)
	if err != nil {
		return nil, err
	}
	if len(endValue) == 0 {
		return nil, types.ErrNotFound
	}
	endInfo := &pty.TendermintBlockInfo{}
	err = types.Decode(endValue, endInfo)
	if err != nil {
		return nil, err
	}

	startHeader := startInfo.Block.Header
	endHeader := endInfo.Block.Header
	totalTx := endHeader.TotalTxs - startHeader.TotalTxs
	totalBlock := endHeader.Height - startHeader.Height + 1
	totalSecond := endHeader.Time - startHeader.Time + 1
	return &pty.PerfStat{
		TotalTx:     totalTx,
		TotalBlock:  totalBlock,
		TxPerBlock:  totalTx / totalBlock,
		TotalSecond: totalSecond,
		TxPerSecond: totalTx / totalSecond,
	}, nil
}

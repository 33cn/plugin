// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"

	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/blackwhite/types"
)

var clog = log.New("module", "execs.blackwhite")

var blackwhiteAddr = address.ExecAddress(gt.BlackwhiteX)

var driverName = gt.BlackwhiteX

func init() {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&Blackwhite{}))
}

// Init 重命名执行器名称
func Init(name string, sub []byte) {
	driverName = name
	gt.BlackwhiteX = driverName
	gt.ExecerBlackwhite = []byte(driverName)
	drivers.Register(name, newBlackwhite, types.GetDappFork(driverName, "Enable"))
}

// Blackwhite 几类执行器结构体
type Blackwhite struct {
	drivers.DriverBase
}

func newBlackwhite() drivers.Driver {
	c := &Blackwhite{}
	c.SetChild(c)
	c.SetExecutorType(types.LoadExecutorType(driverName))
	return c
}

// GetName 获取执行器别名
func GetName() string {
	return newBlackwhite().GetName()
}

// GetDriverName 获取执行器名字
func (c *Blackwhite) GetDriverName() string {
	return driverName
}

func (c *Blackwhite) saveLoopResult(res *gt.ReplyLoopResults) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcRoundKey4LoopResult(res.GetGameID())
	kv.Value = types.Encode(res)
	kvs = append(kvs, kv)
	return kvs
}

func (c *Blackwhite) delLoopResult(res *gt.ReplyLoopResults) (kvs []*types.KeyValue) {
	kv := &types.KeyValue{}
	kv.Key = calcRoundKey4LoopResult(res.GetGameID())
	kv.Value = nil
	kvs = append(kvs, kv)
	return kvs
}

func (c *Blackwhite) saveHeightIndex(res *gt.ReceiptBlackwhiteStatus) (kvs []*types.KeyValue) {
	heightstr := genHeightIndexStr(res.GetIndex())
	kv := &types.KeyValue{}
	kv.Key = calcRoundKey4AddrHeight(res.GetAddr(), heightstr)
	kv.Value = []byte(res.GetGameID())
	kvs = append(kvs, kv)

	kv1 := &types.KeyValue{}
	kv1.Key = calcRoundKey4StatusAddrHeight(res.GetStatus(), res.GetAddr(), heightstr)
	kv1.Value = []byte(res.GetGameID())
	kvs = append(kvs, kv1)

	if res.GetStatus() >= 1 {
		kv := &types.KeyValue{}
		kv.Key = calcRoundKey4StatusAddrHeight(res.GetPrevStatus(), res.GetAddr(), heightstr)
		kv.Value = nil
		kvs = append(kvs, kv)
	}
	return kvs
}

func (c *Blackwhite) saveRollHeightIndex(res *gt.ReceiptBlackwhiteStatus) (kvs []*types.KeyValue) {
	heightstr := genHeightIndexStr(res.GetIndex())
	kv := &types.KeyValue{}
	kv.Key = calcRoundKey4AddrHeight(res.GetAddr(), heightstr)
	kv.Value = []byte(res.GetGameID())
	kvs = append(kvs, kv)

	kv1 := &types.KeyValue{}
	kv1.Key = calcRoundKey4StatusAddrHeight(res.GetPrevStatus(), res.GetAddr(), heightstr)
	kv1.Value = []byte(res.GetGameID())
	kvs = append(kvs, kv1)

	return kvs
}

func (c *Blackwhite) delHeightIndex(res *gt.ReceiptBlackwhiteStatus) (kvs []*types.KeyValue) {
	heightstr := genHeightIndexStr(res.GetIndex())
	kv := &types.KeyValue{}
	kv.Key = calcRoundKey4AddrHeight(res.GetAddr(), heightstr)
	kv.Value = nil
	kvs = append(kvs, kv)

	kv1 := &types.KeyValue{}
	kv1.Key = calcRoundKey4StatusAddrHeight(res.GetStatus(), res.GetAddr(), heightstr)
	kv1.Value = nil
	kvs = append(kvs, kv1)
	return kvs
}

// GetBlackwhiteRoundInfo 获取当前游戏信息
func (c *Blackwhite) GetBlackwhiteRoundInfo(req *gt.ReqBlackwhiteRoundInfo) (types.Message, error) {
	gameID := req.GameID
	key := calcMavlRoundKey(gameID)
	values, err := c.GetStateDB().Get(key)
	if err != nil {
		return nil, err
	}

	var round gt.BlackwhiteRound
	err = types.Decode(values, &round)
	if err != nil {
		return nil, err
	}
	//密钥不显示
	for _, addRes := range round.AddrResult {
		addRes.ShowSecret = ""
	}
	roundRes := &gt.BlackwhiteRoundResult{
		GameID:         round.GameID,
		Status:         round.Status,
		PlayAmount:     round.PlayAmount,
		PlayerCount:    round.PlayerCount,
		CurPlayerCount: round.CurPlayerCount,
		Loop:           round.Loop,
		CurShowCount:   round.CurShowCount,
		CreateTime:     round.CreateTime,
		ShowTime:       round.ShowTime,
		Timeout:        round.Timeout,
		CreateAddr:     round.CreateAddr,
		GameName:       round.GameName,
		AddrResult:     round.AddrResult,
		Winner:         round.Winner,
		Index:          round.Index,
	}
	var rep gt.ReplyBlackwhiteRoundInfo
	rep.Round = roundRes
	return &rep, nil
}

// GetBwRoundListInfo 根据要求获取游戏信息，包括游戏所处状态，或者参与者地址
func (c *Blackwhite) GetBwRoundListInfo(req *gt.ReqBlackwhiteRoundList) (types.Message, error) {
	var key []byte
	var values [][]byte
	var err error
	var prefix []byte

	if 0 == req.Status {
		prefix = calcRoundKey4AddrHeight(req.Address, "")
	} else {
		prefix = calcRoundKey4StatusAddrHeight(req.Status, req.Address, "")
	}
	localDb := c.GetLocalDB()
	if req.GetIndex() == -1 {
		values, err = localDb.List(prefix, nil, req.Count, req.GetDirection())
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			return nil, types.ErrNotFound
		}
	} else { //翻页查找指定的txhash列表
		heightstr := genHeightIndexStr(req.GetIndex())
		if 0 == req.Status {
			key = calcRoundKey4AddrHeight(req.Address, heightstr)
		} else {
			key = calcRoundKey4StatusAddrHeight(req.Status, req.Address, heightstr)
		}
		values, err = localDb.List(prefix, key, req.Count, req.Direction)
		if err != nil {
			return nil, err
		}
		if len(values) == 0 {
			return nil, types.ErrNotFound
		}
	}

	if len(values) == 0 {
		return nil, types.ErrNotFound
	}
	storeDb := c.GetStateDB()
	var rep gt.ReplyBlackwhiteRoundList
	for _, value := range values {
		v, err := storeDb.Get(calcMavlRoundKey(string(value)))
		if nil != err {
			return nil, err
		}
		var round gt.BlackwhiteRound
		err = types.Decode(v, &round)
		if err != nil {
			return nil, err
		}
		//密钥不显示
		for _, addRes := range round.AddrResult {
			addRes.ShowSecret = ""
		}
		roundRes := &gt.BlackwhiteRoundResult{
			GameID:         round.GameID,
			Status:         round.Status,
			PlayAmount:     round.PlayAmount,
			PlayerCount:    round.PlayerCount,
			CurPlayerCount: round.CurPlayerCount,
			Loop:           round.Loop,
			CurShowCount:   round.CurShowCount,
			CreateTime:     round.CreateTime,
			ShowTime:       round.ShowTime,
			Timeout:        round.Timeout,
			CreateAddr:     round.CreateAddr,
			GameName:       round.GameName,
			AddrResult:     round.AddrResult,
			Winner:         round.Winner,
			Index:          round.Index,
		}
		rep.Round = append(rep.Round, roundRes)
	}
	return &rep, nil
}

// GetBwRoundLoopResult 获取游戏中每轮的胜负结果
func (c *Blackwhite) GetBwRoundLoopResult(req *gt.ReqLoopResult) (types.Message, error) {
	localDb := c.GetLocalDB()
	values, err := localDb.Get(calcRoundKey4LoopResult(req.GameID))
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, types.ErrNotFound
	}

	var result gt.ReplyLoopResults
	err = types.Decode(values, &result)
	if err != nil {
		return nil, err
	}

	if req.LoopSeq > 0 { //取出具体一轮
		if len(result.Results) < int(req.LoopSeq) {
			return nil, gt.ErrNoLoopSeq
		}
		res := &gt.ReplyLoopResults{
			GameID: result.GameID,
		}
		index := int(req.LoopSeq)
		perRes := &gt.PerLoopResult{}
		perRes.Winers = append(perRes.Winers, res.Results[index-1].Winers...)
		perRes.Losers = append(perRes.Losers, res.Results[index-1].Losers...)
		res.Results = append(res.Results, perRes)
		return res, nil
	}
	return &result, nil //将所有轮数取出
}

func genHeightIndexStr(index int64) string {
	return fmt.Sprintf("%018d", index)
}

func heightIndexToIndex(height int64, index int32) int64 {
	return height*types.MaxTxsPerBlock + int64(index)
}

// GetPayloadValue 获取执行器action结构体
func (c *Blackwhite) GetPayloadValue() types.Message {
	return &gt.BlackwhiteAction{}
}

// CheckReceiptExecOk return true to check if receipt ty is ok
func (c *Blackwhite) CheckReceiptExecOk() bool {
	return true
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

// Query_GetLotteryNormalInfo not changed info
func (l *Lottery) Query_GetLotteryNormalInfo(param *pty.ReqLotteryInfo) (types.Message, error) {
	lottery, err := findLottery(l.GetStateDB(), param.GetLotteryId())
	if err != nil {
		return nil, err
	}
	return &pty.ReplyLotteryNormalInfo{CreateHeight: lottery.CreateHeight,
		PurBlockNum:    lottery.PurBlockNum,
		DrawBlockNum:   lottery.DrawBlockNum,
		CreateAddr:     lottery.CreateAddr,
		OpRewardRatio:  lottery.OpRewardRatio,
		DevRewardRatio: lottery.DevRewardRatio}, nil
}

// Query_GetLotteryPurchaseAddr for current round
func (l *Lottery) Query_GetLotteryPurchaseAddr(param *pty.ReqLotteryInfo) (types.Message, error) {
	lottery, err := findLottery(l.GetStateDB(), param.GetLotteryId())
	if err != nil {
		return nil, err
	}
	reply := &pty.ReplyLotteryPurchaseAddr{}
	for _, recs := range lottery.PurRecords {
		reply.Address = append(reply.Address, recs.Addr)
	}
	//lottery.Records
	return reply, nil
}

// Query_GetLotteryCurrentInfo state
func (l *Lottery) Query_GetLotteryCurrentInfo(param *pty.ReqLotteryInfo) (types.Message, error) {
	lottery, err := findLottery(l.GetStateDB(), param.GetLotteryId())
	if err != nil {
		return nil, err
	}
	reply := &pty.ReplyLotteryCurrentInfo{Status: lottery.Status,
		Fund:                       lottery.Fund,
		LastTransToPurState:        lottery.LastTransToPurState,
		LastTransToDrawState:       lottery.LastTransToDrawState,
		TotalPurchasedTxNum:        lottery.TotalPurchasedTxNum,
		Round:                      lottery.Round,
		LuckyNumber:                lottery.LuckyNumber,
		LastTransToPurStateOnMain:  lottery.LastTransToPurStateOnMain,
		LastTransToDrawStateOnMain: lottery.LastTransToDrawStateOnMain,
		PurBlockNum:                lottery.PurBlockNum,
		DrawBlockNum:               lottery.DrawBlockNum,
		MissingRecords:             lottery.MissingRecords,
		TotalAddrNum:               lottery.TotalAddrNum,
		BuyAmount:                  lottery.BuyAmount}
	return reply, nil
}

// Query_GetLotteryHistoryLuckyNumber for all history
func (l *Lottery) Query_GetLotteryHistoryLuckyNumber(param *pty.ReqLotteryLuckyHistory) (types.Message, error) {
	return ListLotteryLuckyHistory(l.GetLocalDB(), l.GetStateDB(), param)
}

// Query_GetLotteryRoundLuckyNumber for each round
func (l *Lottery) Query_GetLotteryRoundLuckyNumber(param *pty.ReqLotteryLuckyInfo) (types.Message, error) {
	//	var req pty.ReqLotteryLuckyInfo
	var records []*pty.LotteryDrawRecord
	//	err := types.Decode(param, &req)
	//if err != nil {
	//	return nil, err
	//}
	for _, round := range param.Round {
		key := calcLotteryDrawKey(param.LotteryId, round)
		record, err := l.findLotteryDrawRecord(key)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return &pty.LotteryDrawRecords{Records: records}, nil
}

// Query_GetLotteryHistoryBuyInfo for all history
func (l *Lottery) Query_GetLotteryHistoryBuyInfo(param *pty.ReqLotteryBuyHistory) (types.Message, error) {
	return ListLotteryBuyRecords(l.GetLocalDB(), l.GetStateDB(), param)
}

// Query_GetLotteryBuyRoundInfo for each round
func (l *Lottery) Query_GetLotteryBuyRoundInfo(param *pty.ReqLotteryBuyInfo) (types.Message, error) {
	key := calcLotteryBuyRoundPrefix(param.LotteryId, param.Addr, param.Round)
	record, err := l.findLotteryBuyRecords(key)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Query_GetLotteryHistoryGainInfo for all history
func (l *Lottery) Query_GetLotteryHistoryGainInfo(param *pty.ReqLotteryGainHistory) (types.Message, error) {
	return ListLotteryGainRecords(l.GetLocalDB(), l.GetStateDB(), param)
}

// Query_GetLotteryRoundGainInfo for each round
func (l *Lottery) Query_GetLotteryRoundGainInfo(param *pty.ReqLotteryGainInfo) (types.Message, error) {
	key := calcLotteryGainKey(param.LotteryId, param.Addr, param.Round)
	record, err := l.findLotteryGainRecord(key)
	if err != nil {
		return nil, err
	}
	return record, nil
}

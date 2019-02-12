// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strconv"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

const (
	exciting = 100000
	lucky    = 1000
	happy    = 100
	notbad   = 10
)

const (
	minPurBlockNum  = 30
	minDrawBlockNum = 40
)

const (
	creatorKey = "lottery-creator"
)

// List control
const (
	ListDESC    = int32(0)
	ListASC     = int32(1)
	DefultCount = int32(20)  //默认一次取多少条记录
	MaxCount    = int32(100) //最多取100条
)

// Star level
const (
	FiveStar  = 5
	ThreeStar = 3
	TwoStar   = 2
	OneStar   = 1
)

const (
	luckyNumMol = 100000
	decimal     = types.Coin //1e8
	blockNum    = 5
)

const (
	maxRatio      = 100
	rewardBase    = 1000
	devRewardAddr = "1D6RFZNp2rh6QdbcZ1d7RWuBUz61We6SD7"
	opRewardAddr  = "1PHtChNt3UcfssR7v7trKSk3WJtAWjKjjX"
)

// LotteryDB def
type LotteryDB struct {
	pty.Lottery
}

// NewLotteryDB New management instance
func NewLotteryDB(lotteryID string, purBlock int64, drawBlock int64,
	blockHeight int64, addr string) *LotteryDB {
	lott := &LotteryDB{}
	lott.LotteryId = lotteryID
	lott.PurBlockNum = purBlock
	lott.DrawBlockNum = drawBlock
	lott.CreateHeight = blockHeight
	lott.Fund = 0
	lott.Status = pty.LotteryCreated
	lott.TotalPurchasedTxNum = 0
	lott.CreateAddr = addr
	lott.Round = 0
	lott.MissingRecords = make([]*pty.MissingRecord, 5)
	for index := range lott.MissingRecords {
		tempTimes := make([]int32, 10)
		lott.MissingRecords[index] = &pty.MissingRecord{Times: tempTimes}
	}
	return lott
}

// GetKVSet for LotteryDB
func (lott *LotteryDB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&lott.Lottery)
	kvset = append(kvset, &types.KeyValue{Key: Key(lott.LotteryId), Value: value})
	return kvset
}

// Save for LotteryDB
func (lott *LotteryDB) Save(db dbm.KV) {
	set := lott.GetKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

// Key for lottery
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+pty.LotteryX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

// Action struct
type Action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	difficulty   uint64
	index        int
	lottery      *Lottery
}

// NewLotteryAction generate New Action
func NewLotteryAction(l *Lottery, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{
		coinsAccount: l.GetCoinsAccount(), db: l.GetStateDB(),
		txhash: hash, fromaddr: fromaddr, blocktime: l.GetBlockTime(),
		height: l.GetHeight(), execaddr: dapp.ExecAddress(string(tx.Execer)),
		difficulty: l.GetDifficulty(), index: index, lottery: l}
}

// GetLottCommonRecipt generate logs for lottery common action
func (action *Action) GetLottCommonRecipt(lottery *pty.Lottery, preStatus int32) *pty.ReceiptLottery {
	l := &pty.ReceiptLottery{}
	l.LotteryId = lottery.LotteryId
	l.Status = lottery.Status
	l.PrevStatus = preStatus
	return l
}

// GetCreateReceiptLog generate logs for lottery create action
func (action *Action) GetCreateReceiptLog(lottery *pty.Lottery, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogLotteryCreate

	l := action.GetLottCommonRecipt(lottery, preStatus)

	log.Log = types.Encode(l)

	return log
}

// GetBuyReceiptLog generate logs for lottery buy action
func (action *Action) GetBuyReceiptLog(lottery *pty.Lottery, preStatus int32, round int64, buyNumber int64, amount int64, way int64) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogLotteryBuy

	l := action.GetLottCommonRecipt(lottery, preStatus)

	l.Round = round
	l.Number = buyNumber
	l.Amount = amount
	l.Addr = action.fromaddr
	l.Way = way
	l.Index = action.GetIndex()
	l.Time = action.blocktime
	l.TxHash = common.ToHex(action.txhash)

	log.Log = types.Encode(l)

	return log
}

// GetDrawReceiptLog generate logs for lottery draw action
func (action *Action) GetDrawReceiptLog(lottery *pty.Lottery, preStatus int32, round int64, luckyNum int64, updateInfo *pty.LotteryUpdateBuyInfo, addrNumThisRound int64, buyAmountThisRound int64, gainInfos *pty.LotteryGainInfos,
	luckyAddrNum int64, totalFund int64, factor int64) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogLotteryDraw

	l := action.GetLottCommonRecipt(lottery, preStatus)

	l.Round = round
	l.LuckyNumber = luckyNum
	l.Time = action.blocktime
	l.TxHash = common.ToHex(action.txhash)
	l.TotalAddrNum = addrNumThisRound
	l.BuyAmount = buyAmountThisRound
	l.LuckyAddrNum = luckyAddrNum
	l.TotalFund = totalFund
	l.Factor = factor
	if len(updateInfo.BuyInfo) > 0 {
		l.UpdateInfo = updateInfo
	}

	l.GainInfos = gainInfos

	log.Log = types.Encode(l)

	return log
}

// GetCloseReceiptLog generate logs for lottery close action
func (action *Action) GetCloseReceiptLog(lottery *pty.Lottery, preStatus int32) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pty.TyLogLotteryClose

	l := action.GetLottCommonRecipt(lottery, preStatus)

	log.Log = types.Encode(l)

	return log
}

// GetIndex returns index in block
func (action *Action) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

// LotteryCreate Action
// creator should be valid
func (action *Action) LotteryCreate(create *pty.LotteryCreate) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	lotteryID := common.ToHex(action.txhash)

	if create.OpRewardRatio > maxRatio || create.DevRewardRatio > maxRatio || create.OpRewardRatio < 0 || create.DevRewardRatio < 0 {
		return nil, pty.ErrRewardFactor
	}
	if !isRightCreator(action.fromaddr, action.db, false) {
		return nil, pty.ErrNoPrivilege
	}

	if create.GetPurBlockNum() < minPurBlockNum {
		return nil, pty.ErrLotteryPurBlockLimit
	}

	if create.GetDrawBlockNum() < minDrawBlockNum {
		return nil, pty.ErrLotteryDrawBlockLimit
	}

	if create.GetPurBlockNum() > create.GetDrawBlockNum() {
		return nil, pty.ErrLotteryDrawBlockLimit
	}

	_, err := findLottery(action.db, lotteryID)
	if err != types.ErrNotFound {
		llog.Error("LotteryCreate", "LotteryCreate repeated", lotteryID)
		return nil, pty.ErrLotteryRepeatHash
	}

	lott := NewLotteryDB(lotteryID, create.GetPurBlockNum(),
		create.GetDrawBlockNum(), action.height, action.fromaddr)

	lott.OpRewardRatio = create.OpRewardRatio
	lott.DevRewardRatio = create.DevRewardRatio
	lott.TotalAddrNum = 0
	lott.BuyAmount = 0
	llog.Debug("LotteryCreate", "OpRewardRatio", lott.OpRewardRatio, "DevRewardRatio", lott.DevRewardRatio)
	if types.IsPara() {
		lott.CreateOnMain = action.lottery.GetMainHeight()
	}

	llog.Debug("LotteryCreate created", "lotteryID", lotteryID)

	lott.Save(action.db)
	kv = append(kv, lott.GetKVSet()...)

	receiptLog := action.GetCreateReceiptLog(&lott.Lottery, 0)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// LotteryBuy Action
// One coin for one ticket
func (action *Action) LotteryBuy(buy *pty.LotteryBuy) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//var receipt *types.Receipt

	lottery, err := findLottery(action.db, buy.LotteryId)
	if err != nil {
		llog.Error("LotteryBuy", "LotteryId", buy.LotteryId)
		return nil, err
	}

	lott := &LotteryDB{*lottery}
	preStatus := lott.Status

	if lott.Status == pty.LotteryClosed {
		llog.Error("LotteryBuy", "status", lott.Status)
		return nil, pty.ErrLotteryStatus
	}

	if lott.Status == pty.LotteryDrawed {
		//no problem both on main and para
		if action.height <= lott.LastTransToDrawState {
			llog.Error("LotteryBuy", "action.heigt", action.height, "lastTransToDrawState", lott.LastTransToDrawState)
			return nil, pty.ErrLotteryStatus
		}
	}

	if lott.Status == pty.LotteryCreated || lott.Status == pty.LotteryDrawed {
		llog.Debug("LotteryBuy switch to purchasestate")
		lott.LastTransToPurState = action.height
		lott.Status = pty.LotteryPurchase
		lott.Round++
		if types.IsPara() {
			lott.LastTransToPurStateOnMain = action.lottery.GetMainHeight()
		}
	}

	if lott.Status == pty.LotteryPurchase {
		if types.IsPara() {
			mainHeight := action.lottery.GetMainHeight()
			if mainHeight-lott.LastTransToPurStateOnMain > lott.GetPurBlockNum() {
				llog.Error("LotteryBuy", "action.height", action.height, "mainHeight", mainHeight, "LastTransToPurStateOnMain", lott.LastTransToPurStateOnMain)
				return nil, pty.ErrLotteryStatus
			}
		} else {
			if action.height-lott.LastTransToPurState > lott.GetPurBlockNum() {
				llog.Error("LotteryBuy", "action.height", action.height, "LastTransToPurState", lott.LastTransToPurState)
				return nil, pty.ErrLotteryStatus
			}
		}
	}

	if lott.CreateAddr == action.fromaddr {
		return nil, pty.ErrLotteryCreatorBuy
	}

	if buy.GetAmount() <= 0 {
		llog.Error("LotteryBuy", "buyAmount", buy.GetAmount())
		return nil, pty.ErrLotteryBuyAmount
	}

	if buy.GetNumber() < 0 || buy.GetNumber() >= luckyNumMol {
		llog.Error("LotteryBuy", "buyNumber", buy.GetNumber())
		return nil, pty.ErrLotteryBuyNumber
	}

	newRecord := &pty.PurchaseRecord{Amount: buy.GetAmount(), Number: buy.GetNumber(), Index: action.GetIndex(), Way: buy.GetWay()}
	llog.Debug("LotteryBuy", "amount", buy.GetAmount(), "number", buy.GetNumber())

	/**********
	Once ExecTransfer succeed, ExecFrozen succeed, no roolback needed
	**********/

	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, lott.CreateAddr, action.execaddr, buy.GetAmount()*decimal)
	if err != nil {
		llog.Error("LotteryBuy.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", buy.GetAmount())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	receipt, err = action.coinsAccount.ExecFrozen(lott.CreateAddr, action.execaddr, buy.GetAmount()*decimal)

	if err != nil {
		llog.Error("LotteryBuy.Frozen", "addr", lott.CreateAddr, "execaddr", action.execaddr, "amount", buy.GetAmount())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	lott.Fund += buy.GetAmount()
	lott.BuyAmount += buy.GetAmount()
	lott.TotalPurchasedTxNum++

	newAddr := true
	for i := range lott.PurRecords {
		if action.fromaddr == lott.PurRecords[i].Addr {
			lott.PurRecords[i].Record = append(lott.PurRecords[i].Record, newRecord)
			lott.PurRecords[i].AmountOneRound += buy.Amount
			newAddr = false
			break
		}
	}
	if newAddr {
		initrecord := &pty.PurchaseRecords{}
		initrecord.Record = append(initrecord.Record, newRecord)
		initrecord.FundWin = 0
		initrecord.AmountOneRound = buy.Amount
		initrecord.Addr = action.fromaddr
		lott.PurRecords = append(lott.PurRecords, initrecord)
		lott.TotalAddrNum++
	}

	lott.Save(action.db)
	kv = append(kv, lott.GetKVSet()...)

	receiptLog := action.GetBuyReceiptLog(&lott.Lottery, preStatus, lott.Round, buy.GetNumber(), buy.GetAmount(), buy.GetWay())
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// LotteryDraw Action
func (action *Action) LotteryDraw(draw *pty.LotteryDraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var receipt *types.Receipt

	lottery, err := findLottery(action.db, draw.LotteryId)
	if err != nil {
		llog.Error("LotteryBuy", "LotteryId", draw.LotteryId)
		return nil, err
	}

	lott := &LotteryDB{*lottery}

	preStatus := lott.Status

	if lott.Status != pty.LotteryPurchase {
		llog.Error("LotteryDraw", "lott.Status", lott.Status)
		return nil, pty.ErrLotteryStatus
	}

	if types.IsPara() {
		mainHeight := action.lottery.GetMainHeight()
		if mainHeight-lott.GetLastTransToPurStateOnMain() < lott.GetDrawBlockNum() {
			llog.Error("LotteryDraw", "action.height", action.height, "mainHeight", mainHeight, "GetLastTransToPurStateOnMain", lott.GetLastTransToPurState())
			return nil, pty.ErrLotteryStatus
		}
	} else {
		if action.height-lott.GetLastTransToPurState() < lott.GetDrawBlockNum() {
			llog.Error("LotteryDraw", "action.height", action.height, "GetLastTransToPurState", lott.GetLastTransToPurState())
			return nil, pty.ErrLotteryStatus
		}
	}

	if action.fromaddr != lott.GetCreateAddr() {
		//if _, ok := lott.Records[action.fromaddr]; !ok {
		llog.Error("LotteryDraw", "action.fromaddr", action.fromaddr)
		return nil, pty.ErrLotteryDrawActionInvalid
		//}
	}

	//record addr and amount this round
	addrNumThisRound := lott.TotalAddrNum
	buyAmountThisRound := lott.BuyAmount

	rec, updateInfo, gainInfos, luckyAddrNum, totalFund, factor, err := action.checkDraw(lott)
	if err != nil {
		return nil, err
	}
	kv = append(kv, rec.KV...)
	logs = append(logs, rec.Logs...)

	lott.Save(action.db)
	kv = append(kv, lott.GetKVSet()...)

	receiptLog := action.GetDrawReceiptLog(&lott.Lottery, preStatus, lott.Round, lott.LuckyNumber, updateInfo, addrNumThisRound, buyAmountThisRound, gainInfos, luckyAddrNum, totalFund, factor)
	logs = append(logs, receiptLog)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// LotteryClose Action
func (action *Action) LotteryClose(draw *pty.LotteryClose) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//var receipt *types.Receipt

	if !isEableToClose() {
		return nil, pty.ErrLotteryErrUnableClose
	}

	lottery, err := findLottery(action.db, draw.LotteryId)
	if err != nil {
		llog.Error("LotteryBuy", "LotteryId", draw.LotteryId)
		return nil, err
	}

	lott := &LotteryDB{*lottery}
	preStatus := lott.Status

	if action.fromaddr != lott.CreateAddr {
		return nil, pty.ErrLotteryErrCloser
	}

	if lott.Status == pty.LotteryClosed {
		return nil, pty.ErrLotteryStatus
	}

	var totalReturn int64
	for _, recs := range lott.PurRecords {
		totalReturn += recs.AmountOneRound
	}
	llog.Debug("LotteryClose", "totalReturn", totalReturn)

	if totalReturn > 0 {

		if !action.CheckExecAccount(lott.CreateAddr, decimal*totalReturn, true) {
			return nil, pty.ErrLotteryFundNotEnough
		}

		for _, recs := range lott.PurRecords {
			if recs.AmountOneRound > 0 {
				receipt, err := action.coinsAccount.ExecTransferFrozen(lott.CreateAddr, recs.Addr, action.execaddr,
					decimal*recs.AmountOneRound)
				if err != nil {
					return nil, err
				}

				kv = append(kv, receipt.KV...)
				logs = append(logs, receipt.Logs...)
			}
		}
	}

	for i := range lott.PurRecords {
		lott.PurRecords[i].Record = lott.PurRecords[i].Record[0:0]
	}
	lott.PurRecords = lott.PurRecords[0:0]
	lott.TotalPurchasedTxNum = 0
	llog.Debug("LotteryClose switch to closestate")
	lott.Status = pty.LotteryClosed

	lott.Save(action.db)
	kv = append(kv, lott.GetKVSet()...)

	receiptLog := action.GetCloseReceiptLog(&lott.Lottery, preStatus)
	logs = append(logs, receiptLog)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//random used for verification in solo
func (action *Action) findLuckyNum(isSolo bool, lott *LotteryDB) (int64, error) {
	var num int64
	if isSolo {
		//used for internal verification
		num = 12345
	} else {
		//发消息给randnum模块
		//在主链上，当前高度查询不到，如果要保证区块个数，高度传入action.height-1
		llog.Debug("findLuckyNum on randnum module")
		param := &types.ReqRandHash{
			ExecName: "ticket",
			BlockNum: blockNum,
			Hash:     action.lottery.GetLastHash(),
		}
		hash, err := action.lottery.GetExecutorAPI().GetRandNum(param)
		if err != nil {
			return -1, err
		}
		baseNum, err := strconv.ParseUint(common.ToHex(hash[0:4]), 0, 64)
		llog.Debug("findLuckyNum", "baseNum", baseNum)
		if err != nil {
			return -1, err
		}
		num = int64(baseNum) % luckyNumMol
	}
	return num, nil
}

func checkFundAmount(luckynum int64, guessnum int64, way int64) (int64, int64) {
	if way == FiveStar && luckynum == guessnum {
		return exciting, FiveStar
	} else if way == ThreeStar && luckynum%1000 == guessnum%1000 {
		return lucky, ThreeStar
	} else if way == TwoStar && luckynum%100 == guessnum%100 {
		return happy, TwoStar
	} else if way == OneStar && luckynum%10 == guessnum%10 {
		return notbad, OneStar
	} else {
		return 0, 0
	}
}

func (action *Action) checkDraw(lott *LotteryDB) (*types.Receipt, *pty.LotteryUpdateBuyInfo, *pty.LotteryGainInfos, int64, int64, int64, error) {
	luckynum, err := action.findLuckyNum(false, lott)
	if luckynum < 0 || luckynum >= luckyNumMol {
		return nil, nil, nil, 0, 0, 0, err
	}

	llog.Debug("checkDraw", "luckynum", luckynum)

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var updateInfo pty.LotteryUpdateBuyInfo
	var gainInfos pty.LotteryGainInfos
	var tempFund int64
	var totalFund int64
	var luckyAddrNum int64

	updateInfo.BuyInfo = make(map[string]*pty.LotteryUpdateRecs)

	for i := range lott.PurRecords {
		for _, rec := range lott.PurRecords[i].Record {
			fund, fundType := checkFundAmount(luckynum, rec.Number, rec.Way)
			if fund != 0 {
				newUpdateRec := &pty.LotteryUpdateRec{Index: rec.Index, Type: fundType}
				if update, ok := updateInfo.BuyInfo[lott.PurRecords[i].Addr]; ok {
					update.Records = append(update.Records, newUpdateRec)
				} else {
					initrecord := &pty.LotteryUpdateRecs{}
					initrecord.Records = append(initrecord.Records, newUpdateRec)
					updateInfo.BuyInfo[lott.PurRecords[i].Addr] = initrecord
				}
			}
			tempFund = fund * rec.Amount
			lott.PurRecords[i].FundWin += tempFund
			totalFund += tempFund
		}
	}
	llog.Debug("checkDraw", "lenofupdate", len(updateInfo.BuyInfo), "update", updateInfo.BuyInfo)

	var factor = decimal
	if totalFund > 0 {
		if totalFund > lott.GetFund()/2 {
			llog.Debug("checkDraw ajust fund", "lott.Fund", lott.Fund, "totalFund", totalFund)
			factor = decimal * (lott.GetFund()) / 2 / (totalFund)
			lott.Fund = lott.Fund / 2
		} else {
			lott.Fund -= totalFund
		}

		llog.Debug("checkDraw", "factor", factor, "totalFund", totalFund)

		if !action.CheckExecAccount(lott.CreateAddr, factor*totalFund+1, true) {
			return nil, nil, nil, 0, 0, 0, pty.ErrLotteryFundNotEnough
		}

		for _, recs := range lott.PurRecords {
			if recs.FundWin > 0 {
				fund := (recs.FundWin * factor) * (rewardBase - lott.OpRewardRatio - lott.DevRewardRatio) / rewardBase //any problem when too little?
				llog.Debug("checkDraw", "fund", fund)
				gain := &pty.LotteryGainInfo{Addr: recs.Addr, BuyAmount: recs.AmountOneRound, FundAmount: fund}
				gainInfos.Gains = append(gainInfos.Gains, gain)
				receipt, err := action.coinsAccount.ExecTransferFrozen(lott.CreateAddr, recs.Addr, action.execaddr, fund)
				if err != nil {
					return nil, nil, nil, 0, 0, 0, err
				}
				luckyAddrNum++
				kv = append(kv, receipt.KV...)
				logs = append(logs, receipt.Logs...)
			} else {
				gain := &pty.LotteryGainInfo{Addr: recs.Addr, BuyAmount: recs.AmountOneRound, FundAmount: 0}
				gainInfos.Gains = append(gainInfos.Gains, gain)
			}
		}

		//op reward
		fundOp := factor * totalFund * lott.OpRewardRatio / rewardBase
		receipt, err := action.coinsAccount.ExecTransferFrozen(lott.CreateAddr, opRewardAddr, action.execaddr, fundOp)
		if err != nil {
			return nil, nil, nil, 0, 0, 0, err
		}
		kv = append(kv, receipt.KV...)
		logs = append(logs, receipt.Logs...)
		//dev reward
		fundDev := factor * totalFund * lott.DevRewardRatio / rewardBase
		receipt, err = action.coinsAccount.ExecTransferFrozen(lott.CreateAddr, devRewardAddr, action.execaddr, fundDev)
		if err != nil {
			return nil, nil, nil, 0, 0, 0, err
		}
		kv = append(kv, receipt.KV...)
		logs = append(logs, receipt.Logs...)
	} else {
		for _, recs := range lott.PurRecords {
			gain := &pty.LotteryGainInfo{Addr: recs.Addr, BuyAmount: recs.AmountOneRound, FundAmount: 0}
			gainInfos.Gains = append(gainInfos.Gains, gain)
		}
	}

	for i := range lott.PurRecords {
		lott.PurRecords[i].Record = lott.PurRecords[i].Record[0:0]
	}
	lott.PurRecords = lott.PurRecords[0:0]

	llog.Debug("checkDraw lottery switch to drawed")
	lott.LastTransToDrawState = action.height
	lott.Status = pty.LotteryDrawed
	lott.TotalPurchasedTxNum = 0
	lott.LuckyNumber = luckynum
	lott.TotalAddrNum = 0
	lott.BuyAmount = 0
	action.recordMissing(lott)

	if types.IsPara() {
		lott.LastTransToDrawStateOnMain = action.lottery.GetMainHeight()
	}
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, &updateInfo, &gainInfos, luckyAddrNum, totalFund, factor, nil
}
func (action *Action) recordMissing(lott *LotteryDB) {
	temp := int32(lott.LuckyNumber)
	initNum := int32(10000)
	sample := [10]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	var eachNum [5]int32
	for i := 0; i < 5; i++ {
		eachNum[i] = temp / initNum
		temp -= eachNum[i] * initNum
		initNum = initNum / 10
	}
	for i := 0; i < 5; i++ {
		for j := 0; j < 10; j++ {
			if eachNum[i] != sample[j] {
				lott.MissingRecords[i].Times[j]++
			}
		}
	}
}

func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		return nil, err
	}
	return value, nil
}

func isRightCreator(addr string, db dbm.KV, isSolo bool) bool {
	if isSolo {
		return true
	}
	value, err := getManageKey(creatorKey, db)
	if err != nil {
		llog.Error("LotteryCreate", "creatorKey", creatorKey)
		return false
	}
	if value == nil {
		llog.Error("LotteryCreate found nil value")
		return false
	}

	var item types.ConfigItem
	err = types.Decode(value, &item)
	if err != nil {
		llog.Error("LotteryCreate", "Decode", value)
		return false
	}

	for _, op := range item.GetArr().Value {
		if op == addr {
			return true
		}
	}
	return false

}

func isEableToClose() bool {
	return true
}

func findLottery(db dbm.KV, lotteryID string) (*pty.Lottery, error) {
	data, err := db.Get(Key(lotteryID))
	if err != nil {
		llog.Debug("findLottery", "get", err)
		return nil, err
	}
	var lott pty.Lottery
	//decode
	err = types.Decode(data, &lott)
	if err != nil {
		llog.Debug("findLottery", "decode", err)
		return nil, err
	}
	return &lott, nil
}

// CheckExecAccount check the account avoiding rollback
func (action *Action) CheckExecAccount(addr string, amount int64, isFrozen bool) bool {
	acc := action.coinsAccount.LoadExecAccount(addr, action.execaddr)
	if isFrozen {
		if acc.GetFrozen() >= amount {
			return true
		}
	} else {
		if acc.GetBalance() >= amount {
			return true
		}
	}

	return false
}

// ListLotteryLuckyHistory returns all the luckynum in history
func ListLotteryLuckyHistory(db dbm.Lister, stateDB dbm.KV, param *pty.ReqLotteryLuckyHistory) (types.Message, error) {
	direction := ListDESC
	if param.GetDirection() == ListASC {
		direction = ListASC
	}
	count := DefultCount
	if 0 < param.GetCount() && param.GetCount() <= MaxCount {
		count = param.GetCount()
	}
	var prefix []byte
	var key []byte
	var values [][]byte
	var err error

	prefix = calcLotteryDrawPrefix(param.LotteryId)
	key = calcLotteryDrawKey(param.LotteryId, param.GetRound())

	if param.GetRound() == 0 { //第一次查询
		values, err = db.List(prefix, nil, count, direction)
	} else {
		values, err = db.List(prefix, key, count, direction)
	}
	if err != nil {
		return nil, err
	}

	var records pty.LotteryDrawRecords
	for _, value := range values {
		var record pty.LotteryDrawRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records.Records = append(records.Records, &record)
	}

	return &records, nil
}

// ListLotteryBuyRecords for addr
func ListLotteryBuyRecords(db dbm.Lister, stateDB dbm.KV, param *pty.ReqLotteryBuyHistory) (types.Message, error) {
	direction := ListDESC
	if param.GetDirection() == ListASC {
		direction = ListASC
	}
	count := DefultCount
	if 0 < param.GetCount() && param.GetCount() <= MaxCount {
		count = param.GetCount()
	}
	var prefix []byte
	var key []byte
	var values [][]byte
	var err error

	prefix = calcLotteryBuyPrefix(param.LotteryId, param.Addr)
	key = calcLotteryBuyKey(param.LotteryId, param.Addr, param.GetRound(), param.GetIndex())

	if param.GetRound() == 0 { //第一次查询
		values, err = db.List(prefix, nil, count, direction)
	} else {
		values, err = db.List(prefix, key, count, direction)
	}

	if err != nil {
		return nil, err
	}

	var records pty.LotteryBuyRecords
	for _, value := range values {
		var record pty.LotteryBuyRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records.Records = append(records.Records, &record)
	}

	return &records, nil

}

// ListLotteryGainRecords for addr
func ListLotteryGainRecords(db dbm.Lister, stateDB dbm.KV, param *pty.ReqLotteryGainHistory) (types.Message, error) {
	direction := ListDESC
	if param.GetDirection() == ListASC {
		direction = ListASC
	}
	count := DefultCount
	if 0 < param.GetCount() && param.GetCount() <= MaxCount {
		count = param.GetCount()
	}
	var prefix []byte
	var key []byte
	var values [][]byte
	var err error

	prefix = calcLotteryGainPrefix(param.LotteryId, param.Addr)
	key = calcLotteryGainKey(param.LotteryId, param.Addr, param.GetRound())

	if param.GetRound() == 0 { //第一次查询
		values, err = db.List(prefix, nil, count, direction)
	} else {
		values, err = db.List(prefix, key, count, direction)
	}

	if err != nil {
		return nil, err
	}

	var records pty.LotteryGainRecords
	for _, value := range values {
		var record pty.LotteryGainRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records.Records = append(records.Records, &record)
	}

	return &records, nil

}

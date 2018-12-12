/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package executor

import (
	"fmt"
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pt "github.com/33cn/plugin/plugin/dapp/f3d/ptypes"
)

// round status
const (
	decimal = 100000000 //1e8
	// ListDESC  desc query
	ListDESC = int32(0)
	// ListASC  asc query
	ListASC      = int32(1)
	F3dRoundLast = "round-last"
	// DefaultCount 默认一次取多少条记录
	DefaultCount = int32(20)

	// MaxCount 最多取100条
	MaxCount = int32(100)
)

// GetReceiptLog get receipt log
func (action *Action) GetStartReceiptLog(roundInfo *pt.RoundInfo) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogf3dStart
	r := &pt.ReceiptF3D{
		Addr:   action.fromaddr,
		Round:  roundInfo.Round,
		Index:  action.GetIndex(),
		Action: pt.TyLogf3dStart,
	}
	log.Log = types.Encode(r)

	return log
}
func (action *Action) GetBuyReceiptLog(addrInfo *pt.AddrInfo) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogf3dBuy
	r := &pt.ReceiptF3D{
		Addr:     action.fromaddr,
		Round:    addrInfo.Round,
		Index:    action.GetIndex(),
		BuyCount: addrInfo.BuyCount,
		Action:   pt.F3dActionBuy,
	}
	log.Log = types.Encode(r)

	return log
}
func (action *Action) GetDrawReceiptLog(roundInfo *pt.RoundInfo) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogf3dDraw
	r := &pt.ReceiptF3D{
		Addr:   action.fromaddr,
		Round:  roundInfo.Round,
		Index:  action.GetIndex(),
		Action: pt.TyLogf3dDraw,
	}
	log.Log = types.Encode(r)

	return log
}

//GetIndex get index
func (action *Action) GetIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

func calcF3dByRound(round int64) string {
	key := fmt.Sprintf("roundInfo-%010d", round)
	return key
}

func calcF3dUserAddrs(round int64, addr string) string {
	key := fmt.Sprintf("user-Addrs-%010d-%s", round, addr)
	return key
}
func calcF3dUserKeys(round int64, addr string, index int64) string {
	key := fmt.Sprintf("user-keys-%010d-%s-%018d", round, addr, index)
	return key
}

//GetKVSet get kv set
func (action *Action) GetKVSet(param interface{}) (kvset []*types.KeyValue, result interface{}) {
	if roundInfo, ok := param.(*pt.RoundInfo); ok {
		value := types.Encode(roundInfo)
		//更新stateDB缓存
		action.db.Set(Key(calcF3dByRound(roundInfo.Round)), value)
		action.db.Set(Key(F3dRoundLast), value)
		kvset = append(kvset, &types.KeyValue{Key: Key(calcF3dByRound(roundInfo.Round)), Value: value})
		kvset = append(kvset, &types.KeyValue{Key: Key(F3dRoundLast), Value: value})
	}
	if keyInfo, ok := param.(*pt.KeyInfo); ok {
		value := types.Encode(keyInfo)
		action.db.Set(Key(calcF3dUserKeys(keyInfo.Round, keyInfo.Addr, action.GetIndex())), value)
		kvset = append(kvset, &types.KeyValue{Key: Key(calcF3dUserKeys(keyInfo.Round, keyInfo.Addr, action.GetIndex())), Value: value})
		addrInfo, err := getF3dAddrInfo(action.db, Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)))
		if err != nil {
			flog.Warn("F3D db getF3dAddrInfo", "can't get value from db,key:", calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr))
			var addr pt.AddrInfo
			addr.Addr = action.fromaddr
			addr.KeyNum = keyInfo.KeyNum
			addr.BuyCount = 1
			addr.Round = keyInfo.Round
			value := types.Encode(&addr)
			action.db.Set(Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), value)
			kvset = append(kvset, &types.KeyValue{Key: Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), Value: value})
			return kvset, &addr
		} else {
			addrInfo.Addr = action.fromaddr
			addrInfo.BuyCount = addrInfo.BuyCount + 1
			addrInfo.Round = keyInfo.Round
			addrInfo.KeyNum = addrInfo.KeyNum + keyInfo.KeyNum
			value := types.Encode(addrInfo)
			action.db.Set(Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), value)
			kvset = append(kvset, &types.KeyValue{Key: Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), Value: value})
			return kvset, addrInfo
		}

	}
	return kvset, nil
}

func (action *Action) updateCount(status int32, addr string) (kvset []*types.KeyValue) {

	return kvset
}

//func (action *Action) updateStateDBCache(param interface{}){
//	if roundInfo, ok := param.(*pt.RoundInfo); ok {
//		action.db.Set(Key(calcF3dByRound(roundInfo.Round)), types.Encode(roundInfo))
//		action.db.Set(Key(F3dRoundLast), types.Encode(roundInfo))
//	}
//	if keyInfo, ok := param.(*pt.KeyInfo); ok {
//		addrInfo,err:=getF3dAddrInfo(action.db,Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)))
//		action.db.Set(Key(calcF3dUserKeys(keyInfo.Round, keyInfo.Addr,action.GetIndex())), types.Encode(keyInfo))
//		if err !=nil {
//			flog.Warn("F3D db getF3dAddrInfo", "can't get value from db,key:", calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr))
//			var addr pt.AddrInfo
//			addr.Addr=action.fromaddr
//			addr.KeyNum=keyInfo.KeyNum
//			addr.IsFirstBuy=true
//			action.db.Set(Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), types.Encode(&addr))
//		}else{
//			addrInfo.Addr=action.fromaddr
//			addrInfo.IsFirstBuy=false
//			addrInfo.KeyNum=addrInfo.KeyNum+keyInfo.KeyNum
//			action.db.Set(Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), types.Encode(addrInfo))
//		}
//
//
//	}
//
//}

// Key gameId to save key
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+pt.F3DX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

// Action action struct
type Action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	localDB      dbm.Lister
	index        int
}

// NewAction new action
func NewAction(f *f3d, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{f.GetCoinsAccount(), f.GetStateDB(), hash, fromaddr,
		f.GetBlockTime(), f.GetHeight(), dapp.ExecAddress(string(tx.Execer)), f.GetLocalDB(), index}
}

func (action *Action) checkExecAccountBalance(fromAddr string, active, frozen int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= active && acc.GetFrozen() >= frozen {
		return true
	}
	return false
}

//F3d start game
func (action *Action) F3dStart(f3d *pt.F3DStart) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	var startRound int64
	//addr check
	if action.fromaddr != pt.GetF3dManagerAddr() {
		flog.Error("F3dStart", "manager addr not match.", "err", pt.ErrF3dManageAddr.Error())
		return nil, pt.ErrF3dManageAddr
	}
	lastRound, err := getF3dRoundInfo(action.db, Key(F3dRoundLast))
	if err == nil && lastRound != nil {
		if lastRound.EndTime == 0 {
			flog.Error("F3dStart", "start round", startRound)
			return nil, pt.ErrF3dStartRound
		}
		startRound = lastRound.Round
	}

	account := action.coinsAccount.LoadExecAccount(action.fromaddr, action.execaddr)
	roundInfo := &pt.RoundInfo{
		Round:        startRound + 1,
		BeginTime:    action.blocktime,
		LastKeyPrice: pt.GetF3dKeyPriceStart(),
		LastKeyTime:  action.blocktime,
		RemainTime:   pt.GetF3dTimeLife(),
		//TODO is the floating-point precision here accurate?
		BonusPool:  float32(account.Frozen) / decimal,
		UpdateTime: action.blocktime,
	}

	receiptLog := action.GetStartReceiptLog(roundInfo)
	logs = append(logs, receiptLog)
	kvset, _ := action.GetKVSet(roundInfo)
	kv = append(kv, kvset...)
	receipt := &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

//F3d start game
func (action *Action) F3dBuyKey(buy *pt.F3DBuyKey) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//addr check
	if action.fromaddr == pt.GetF3dManagerAddr() {
		flog.Error("F3dBuyKey", "manager can't buy key.", "err", pt.ErrF3dManageBuyKey.Error())
		return nil, pt.ErrF3dManageBuyKey
	}

	lastRound, err := getF3dRoundInfo(action.db, Key(F3dRoundLast))
	if err != nil || lastRound == nil {
		flog.Error("F3dBuyKey", "last round", lastRound.Round, "err", fmt.Errorf("not found the last round info!"))
		return nil, fmt.Errorf("not found the last round info!")
	}
	//round game status check
	if lastRound.EndTime != 0 {
		flog.Error("F3dBuyKey", "last round", lastRound.Round, "err", pt.ErrF3dBuyKey.Error())
		return nil, pt.ErrF3dBuyKey
	}
	// remainTime check
	if lastRound.UpdateTime+lastRound.RemainTime < action.blocktime {
		flog.Error("F3dBuyKey", "time out", "err", pt.ErrF3dBuyKeyTimeOut.Error())
		return nil, pt.ErrF3dBuyKeyTimeOut
	}
	// balance check
	if !action.checkExecAccountBalance(action.fromaddr, buy.GetKeyNum()*int64(lastRound.GetLastKeyPrice()*decimal), 0) {
		flog.Error("F3dBuyKey", "checkExecAccountBalance", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}

	receipt, err := action.coinsAccount.ExecTransfer(action.fromaddr, pt.GetF3dManagerAddr(), action.execaddr, buy.GetKeyNum()*int64(lastRound.GetLastKeyPrice()*decimal))
	if err != nil {
		flog.Error("F3dBuyKey.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", buy.GetKeyNum()*int64(lastRound.GetLastKeyPrice()*decimal))
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	receipt, err = action.coinsAccount.ExecFrozen(pt.GetF3dManagerAddr(), action.execaddr, buy.GetKeyNum()*int64(lastRound.GetLastKeyPrice()*decimal))
	if err != nil {
		flog.Error("F3dBuyKey.Frozen", "addr", pt.GetF3dManagerAddr(), "execaddr", action.execaddr, "amount", buy.GetKeyNum()*int64(lastRound.GetLastKeyPrice()*decimal))
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	keyInfo := &pt.KeyInfo{}
	keyInfo.KeyNum = buy.KeyNum
	keyInfo.Round = lastRound.Round
	keyInfo.Addr = action.fromaddr
	keyInfo.KeyPrice = lastRound.LastKeyPrice
	keyInfo.BuyKeyTime = action.blocktime
	keyInfo.BuyKeyTxHash = common.ToHex(action.txhash)
	keyInfo.Index = action.GetIndex()
	kvset, v := action.GetKVSet(keyInfo)
	kv = append(kv, kvset...)
	if addrInfo, ok := v.(*pt.AddrInfo); ok {
		// first buy
		if addrInfo.BuyCount == 1 {
			lastRound.UserCount = lastRound.UserCount + 1
		}
		receiptLog := action.GetBuyReceiptLog(addrInfo)
		logs = append(logs, receiptLog)
	}
	lastRound.BonusPool = lastRound.BonusPool + float32(buy.GetKeyNum())*lastRound.LastKeyPrice
	lastRound.KeyCount = lastRound.KeyCount + buy.KeyNum
	lastRound.LastKeyPrice = lastRound.LastKeyPrice + lastRound.LastKeyPrice*pt.GetF3dKeyPriceIncr()
	lastRound.LastKeyTime = action.blocktime
	lastRound.UpdateTime = action.blocktime
	lastRound.LastOwner = action.fromaddr
	addTime := 30 * buy.KeyNum
	if addTime >= pt.GetF3dTimeMaxkey() {
		addTime = pt.GetF3dTimeMaxkey()
	}
	if lastRound.RemainTime+addTime >= pt.GetF3dTimeLife() {
		lastRound.RemainTime = pt.GetF3dTimeLife()
	} else {
		lastRound.RemainTime = lastRound.RemainTime + addTime
	}
	//Todo  add addr and nums
	kvset, _ = action.GetKVSet(lastRound)
	kv = append(kv, kvset...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//F3d lucky draws
func (action *Action) F3dLuckyDraw(buy *pt.F3DLuckyDraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	//addr check
	if action.fromaddr != pt.GetF3dManagerAddr() {
		flog.Error("F3dLuckyDraw", "manager addr not match.", "err", pt.ErrF3dManageAddr.Error())
		return nil, pt.ErrF3dManageAddr
	}

	lastRound, err := getF3dRoundInfo(action.db, Key(F3dRoundLast))

	if err != nil || lastRound == nil {
		flog.Error("F3dLuckyDraw", "last round", lastRound.Round, "err", fmt.Errorf("not found the last round info!"))
		return nil, fmt.Errorf("not found the last round info!")
	}
	// remainTime check
	if lastRound.UpdateTime+lastRound.RemainTime > action.blocktime {
		flog.Error("F3dLuckyDraw", "remain time not be zerio", "err", pt.ErrF3dDrawRemainTime.Error())
		return nil, pt.ErrF3dDrawRemainTime
	}
	//round game status check
	if lastRound.EndTime != 0 {
		flog.Error("F3dLuckyDraw", "last round", lastRound.Round, "err", pt.ErrF3dDrawRepeat.Error())
		return nil, pt.ErrF3dDrawRepeat
	}

	//round info check,when no one buy keys,just finish the game
	if lastRound.KeyCount == 0 {
		lastRound.RemainTime = 0
		lastRound.EndTime = action.blocktime
		lastRound.UpdateTime = action.blocktime
		receiptLog := action.GetDrawReceiptLog(lastRound)
		logs = append(logs, receiptLog)
		kvset, _ := action.GetKVSet(lastRound)
		kv = append(kv, kvset...)
		return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
	}

	bonus := int64(lastRound.BonusPool * (pt.GetF3dBonusDeveloper() + pt.GetF3dBonusKey() + pt.GetF3dBonusWinner()) * decimal)
	winner := int64(lastRound.BonusPool * pt.GetF3dBonusWinner() * decimal)
	developer := int64(lastRound.BonusPool * pt.GetF3dBonusDeveloper() * decimal)
	Keys := float32(lastRound.BonusPool * pt.GetF3dBonusKey() * decimal)
	//balance check
	// balance check
	if !action.checkExecAccountBalance(action.fromaddr, 0, bonus) {
		flog.Error("F3dLuckyDraw", "checkExecAccountBalance", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	receipt, err := action.coinsAccount.ExecActive(action.fromaddr, action.execaddr, bonus)
	if err != nil {
		flog.Error("F3dLuckyDraw.ExecActive", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", bonus/decimal)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	//pay bonus for winner
	receipt, err = action.coinsAccount.ExecTransfer(action.fromaddr, lastRound.LastOwner, action.execaddr, winner)
	if err != nil {
		flog.Error("F3dLuckyDraw.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", winner/decimal)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	//pay bonus for developer
	receipt, err = action.coinsAccount.ExecTransfer(action.fromaddr, pt.GetF3dDeveloperAddr(), action.execaddr, developer)
	if err != nil {
		flog.Error("F3dLuckyDraw.ExecTransfer", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", developer/decimal)
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	//pay bonus for key owner
	query := &pt.QueryAddrInfo{
		Round: lastRound.Round,
		Count: DefaultCount,
	}
HERE:
	reply, err := queryList(action.localDB, action.db, query)
	if err != nil {
		flog.Error("F3dLuckyDraw.queryList", "err", err)
	}
	if replyAddr, ok := reply.(*pt.ReplyAddrInfoList); ok {
		for _, addr := range replyAddr.AddrInfoList {
			info, err := getF3dAddrInfo(action.db, Key(calcF3dUserAddrs(lastRound.Round, addr.Addr)))
			if err != nil {
				continue
			}
			var keyBonus int64
			if info.Addr == lastRound.LastOwner {
				keyBonus = int64(Keys * (float32(info.KeyNum-1) / float32(lastRound.KeyCount)))
			} else {
				keyBonus = int64(Keys * (float32(info.KeyNum) / float32(lastRound.KeyCount)))
			}
			if keyBonus <= 0 {
				continue
			}
			receipt, err = action.coinsAccount.ExecTransfer(action.fromaddr, info.Addr, action.execaddr, keyBonus)
			if err != nil {
				flog.Error("F3dLuckyDraw.ExecTransfer", "addr", info.Addr, "execaddr", action.execaddr, "amount", keyBonus/decimal)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		if int32(len(replyAddr.AddrInfoList)) == DefaultCount {
			query.Addr = replyAddr.AddrInfoList[int(DefaultCount-1)].Addr
			goto HERE
		}
	}
	lastRound.RemainTime = 0
	lastRound.EndTime = action.blocktime
	lastRound.UpdateTime = action.blocktime
	receiptLog := action.GetDrawReceiptLog(lastRound)
	logs = append(logs, receiptLog)
	kvset, _ := action.GetKVSet(lastRound)
	kv = append(kv, kvset...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}
func getF3dRoundInfo(db dbm.KV, key []byte) (*pt.RoundInfo, error) {
	value, err := db.Get(key)
	if err != nil {
		flog.Error("F3D db getF3dRoundInfo", "can't get value from db,key:", key, "err", err.Error())
		return nil, err
	}

	var roundInfo pt.RoundInfo
	err = types.Decode(value, &roundInfo)
	if err != nil {
		return nil, err
	}
	return &roundInfo, nil

}

func getF3dAddrInfo(db dbm.KV, key []byte) (*pt.AddrInfo, error) {
	value, err := db.Get(key)
	if err != nil {
		flog.Error("F3D db getF3dAddrInfo", "can't get value from db,key:", key, "err", err.Error())
		return nil, err
	}

	var info pt.AddrInfo
	err = types.Decode(value, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func getF3dBuyRecord(db dbm.KV, key []byte) (*pt.KeyInfo, error) {
	value, err := db.Get(key)
	if err != nil {
		flog.Error("F3D db getF3dBuyRecord", "can't get value from db,key:", key, "err", err.Error())
		return nil, err
	}

	var info pt.KeyInfo
	err = types.Decode(value, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func queryList(db dbm.Lister, stateDB dbm.KV, param interface{}) (types.Message, error) {
	direction := ListDESC
	count := DefaultCount
	if query, ok := param.(*pt.QueryAddrInfo); ok {
		direction = query.GetDirection()
		if 0 < query.GetCount() && query.GetCount() <= MaxCount {
			count = query.GetCount()
		}
		if query.Round == 0 {
			return nil, fmt.Errorf("round can't be zero!")
		}

		var values [][]byte
		var err error
		if query.GetAddr() == "" { //第一次查询
			values, err = db.List(calcF3dAddrPrefix(query.Round), nil, count, direction)
		} else {
			values, err = db.List(calcF3dAddrPrefix(query.Round), calcF3dAddrRound(query.Round, query.Addr), count, direction)
		}
		if err != nil {
			return nil, err
		}
		var addrList []*pt.AddrInfo
		for _, value := range values {
			var addrInfo pt.AddrInfo
			err := types.Decode(value, &addrInfo)
			if err != nil {
				continue
			}
			addrList = append(addrList, &addrInfo)
		}
		return &pt.ReplyAddrInfoList{AddrInfoList: addrList}, nil
	}
	//TODO open or not open as needed ?
	//if query, ok := param.(*pt.QueryF3DListByRound); ok {
	//	direction = query.GetDirection()
	//	if 0 < query.GetCount() && query.GetCount() <= MaxCount {
	//		count = query.GetCount()
	//	}
	//	if query.StartRound == 0 {
	//		return nil, fmt.Errorf("round can't be zero!")
	//	}
	//
	//
	//}
	//query last round info
	if _, ok := param.(*pt.QueryF3DLastRound); ok {
		lastRound, err := getF3dRoundInfo(stateDB, Key(F3dRoundLast))
		if err != nil {
			flog.Error("F3D db queryList", "can't get lastRound:err", err.Error())
			return nil, err
		}
		return lastRound, nil
	}
	//query round info by round
	if query, ok := param.(*pt.QueryF3DByRound); ok {
		if query.Round == 0 {
			return nil, fmt.Errorf("round can't be zero!")
		}
		round, err := getF3dRoundInfo(stateDB, Key(calcF3dByRound(query.Round)))
		if err != nil {
			flog.Error("F3D db queryList", "can't get lastRound:err", err.Error())
			return nil, err
		}
		return round, nil
	}

	//query addr info
	if query, ok := param.(*pt.QueryKeyCountByRoundAndAddr); ok {
		if query.Round == 0 || query.Addr == "" {
			return nil, fmt.Errorf("round can't be zero,addr can't be empty!")
		}
		addrInfo, err := getF3dAddrInfo(stateDB, Key(calcF3dUserAddrs(query.Round, query.Addr)))
		if err != nil {
			flog.Error("F3D db queryList", "can't get addr Info,err", err.Error())
			return nil, err
		}
		return addrInfo, nil
	}
	//query buy record
	if query, ok := param.(*pt.QueryBuyRecordByRoundAndAddr); ok {
		if query.Round == 0 || query.Addr == "" {
			return nil, fmt.Errorf("round can't be zero,addr can't be empty!")
		}

		var values [][]byte
		var err error
		if query.Index == 0 { //第一次查询
			values, err = db.List(calcF3dBuyPrefix(query.Round, query.Addr), nil, count, direction)
		} else {
			values, err = db.List(calcF3dBuyPrefix(query.Round, query.Addr), calcF3dBuyRound(query.Round, query.Addr, query.Index), count, direction)
		}
		if err != nil {
			return nil, err
		}
		var recordList []*pt.KeyInfo
		for _, value := range values {
			var r pt.F3DBuyRecord
			err := types.Decode(value, &r)
			if err != nil {
				continue
			}
			record, err := getF3dBuyRecord(stateDB, Key(calcF3dUserKeys(r.Round, r.Addr, r.Index)))
			if err != nil {
				flog.Error("F3D db queryList", "can't get buy record,err", err.Error())
				continue
			}
			recordList = append(recordList, record)
		}
		return &pt.ReplyBuyRecord{RecordList: recordList}, nil
	}
	return nil, fmt.Errorf("this query can't be supported!")
}

//func getConfValue(db dbm.KV, key string, defaultValue int64) int64 {
//	value, err := getConfigKey(key, db)
//	if err != nil {
//		return defaultValue
//	}
//	if value != nil {
//		v, err := strconv.ParseInt(string(value), 10, 64)
//		if err != nil {
//			flog.Error("gamedb getConfValue", "Type conversion error:", err.Error())
//			return defaultValue
//		}
//		return v
//	}
//	return defaultValue
//}

//func getManageKey(key string, db dbm.KV) ([]byte, error) {
//	manageKey := types.ManageKey(key)
//	value, err := db.Get([]byte(manageKey))
//	if err != nil {
//		if types.IsPara() { //平行链只有一种存储方式
//			flog.Error("gamedb getManage", "can't get value from db,key:", key, "err", err.Error())
//			return nil, err
//		}
//		flog.Debug("gamedb getManageKey", "get db key", "not found")
//		return getConfigKey(key, db)
//	}
//	return value, nil
//}

//func getConfigKey(key string, db dbm.KV) ([]byte, error) {
//	configKey := types.ConfigKey(key)
//	value, err := db.Get([]byte(configKey))
//	if err != nil {
//		flog.Error("gamedb getConfigKey", "can't get value from db,key:", key, "err", err.Error())
//		return nil, err
//	}
//	return value, nil
//}

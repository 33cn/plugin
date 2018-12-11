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
		Addr:  action.fromaddr,
		Round: roundInfo.Round,
		Index: action.GetIndex(),
	}
	log.Log = types.Encode(r)

	return log
}
func (action *Action) GetBuyReceiptLog(addrInfo *pt.AddrInfo) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogf3dBuy
	r := &pt.ReceiptF3D{
		Addr:       action.fromaddr,
		Round:      addrInfo.Round,
		Index:      action.GetIndex(),
		IsFirstBuy: addrInfo.IsFirstBuy,
	}
	log.Log = types.Encode(r)

	return log
}
func (action *Action) GetDrawReceiptLog(roundInfo *pt.RoundInfo) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	log.Ty = pt.TyLogf3dDraw
	r := &pt.ReceiptF3D{
		Addr:  action.fromaddr,
		Round: roundInfo.Round,
		Index: action.GetIndex(),
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
			addr.IsFirstBuy = true
			addr.Round = keyInfo.Round
			value := types.Encode(&addr)
			action.db.Set(Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), value)
			kvset = append(kvset, &types.KeyValue{Key: Key(calcF3dUserAddrs(keyInfo.Round, keyInfo.Addr)), Value: value})
			return kvset, &addr
		} else {
			addrInfo.Addr = action.fromaddr
			addrInfo.IsFirstBuy = false
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
	//add check
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
		RemainTime:   pt.GetF3dTimeKey(),
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
	//round game status check
	lastRound, err := getF3dRoundInfo(action.db, Key(F3dRoundLast))
	if err != nil || lastRound.EndTime != 0 {
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
	kvset, v := action.GetKVSet(keyInfo)
	kv = append(kv, kvset...)
	if addrInfo, ok := v.(*pt.AddrInfo); ok {
		if addrInfo.IsFirstBuy {
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

//F3d luck draws
func (action *Action) F3dLuckyDraw(buy *pt.F3DLuckyDraw) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
   //TODO:

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

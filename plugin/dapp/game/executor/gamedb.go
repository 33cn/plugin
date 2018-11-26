// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

//database opeartion for executor game
import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	gt "github.com/33cn/plugin/plugin/dapp/game/types"
)

const (

	// Scissor 剪刀
	Scissor = int32(1)

	// Rock 石头
	Rock = int32(2)
	// Paper 布
	Paper = int32(3)

	// Unknown 未知结果
	Unknown = int32(4)

	// IsDraw 平局
	IsDraw = int32(1)

	// IsCreatorWin creator win
	IsCreatorWin = int32(2)

	// IsMatcherWin matcher win
	IsMatcherWin = int32(3)

	// IsTimeOut 开奖超时
	IsTimeOut = int32(4)

	// ListDESC  desc query
	ListDESC = int32(0)

	// ListASC  asc query
	ListASC = int32(1)

	// GameCount 根据状态，地址统计整个合约目前总共成功执行了多少场游戏
	GameCount = "GameCount"

	// MaxGameAmount max game amount.单位为types.Coin  1e8
	MaxGameAmount = int64(100)

	// MinGameAmount min game amount
	MinGameAmount = int64(2)

	// DefaultCount 默认一次取多少条记录
	DefaultCount = int64(20)

	// MaxCount 最多取100条
	MaxCount = int64(100)

	//ActiveTime 从有matcher参与游戏开始计算本局游戏开奖的有效时间，单位为天
	ActiveTime = int64(24)
)

// name configured in manager
var (
	ConfNameActiveTime    = gt.GameX + ":" + "activeTime"
	ConfNameDefaultCount  = gt.GameX + ":" + "defaultCount"
	ConfNameMaxCount      = gt.GameX + ":" + "maxCount"
	ConfNameMaxGameAmount = gt.GameX + ":" + "maxGameAmount"
	ConfNameMinGameAmount = gt.GameX + ":" + "minGameAmount"
)

// GetReceiptLog get receipt log
func (action *Action) GetReceiptLog(game *gt.Game) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &gt.ReceiptGame{}
	//TODO 记录这个action由哪个地址触发的
	r.Addr = action.fromaddr
	if game.Status == gt.GameActionCreate {
		// TODO: 这里需要将游戏相关的类型从系统核心代码中移除
		log.Ty = gt.TyLogCreateGame
		r.PrevStatus = -1
	} else if game.Status == gt.GameActionCancel {
		log.Ty = gt.TyLogCancleGame
		r.PrevStatus = gt.GameActionCreate
	} else if game.Status == gt.GameActionMatch {
		log.Ty = gt.TyLogMatchGame
		r.PrevStatus = gt.GameActionCreate
	} else if game.Status == gt.GameActionClose {
		log.Ty = gt.TyLogCloseGame
		r.PrevStatus = gt.GameActionMatch
		r.Addr = game.GetCreateAddress()
	}
	r.GameId = game.GameId
	r.Status = game.Status
	r.CreateAddr = game.GetCreateAddress()
	r.MatchAddr = game.GetMatchAddress()
	r.Index = game.GetIndex()
	r.PrevIndex = game.GetPrevIndex()
	log.Log = types.Encode(r)
	return log
}

//GetIndex get index
func (action *Action) GetIndex(game *gt.Game) int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

//GetKVSet get kv set
func (action *Action) GetKVSet(game *gt.Game) (kvset []*types.KeyValue) {
	value := types.Encode(game)
	kvset = append(kvset, &types.KeyValue{Key: Key(game.GameId), Value: value})
	return kvset
}

func (action *Action) updateCount(status int32, addr string) (kvset []*types.KeyValue) {
	count, err := queryCountByStatusAndAddr(action.db, status, addr)
	if err != nil {
		glog.Error("updateCount", "Query count have err:", err.Error())
	}
	kvset = append(kvset, &types.KeyValue{Key: calcCountKey(status, addr), Value: []byte(strconv.FormatInt(count+1, 10))})
	return kvset
}

func (action *Action) updateStateDBCache(status int32, addr string) {
	count, err := queryCountByStatusAndAddr(action.db, status, addr)
	if err != nil {
		glog.Error("updateStateDBCache", "Query count have err:", err.Error())
	}
	action.db.Set(calcCountKey(status, addr), []byte(strconv.FormatInt(count+1, 10)))
}

func (action *Action) saveStateDB(game *gt.Game) {
	action.db.Set(Key(game.GetGameId()), types.Encode(game))
}

func calcCountKey(status int32, addr string) (key []byte) {
	key = append(key, []byte("mavl-"+gt.GameX+"-")...)
	key = append(key, []byte(fmt.Sprintf("%s:%d:%s", GameCount, status, addr))...)
	return key
}

// Key gameId to save key
func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+gt.GameX+"-")...)
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
func NewAction(g *Game, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{g.GetCoinsAccount(), g.GetStateDB(), hash, fromaddr,
		g.GetBlockTime(), g.GetHeight(), dapp.ExecAddress(string(tx.Execer)), g.GetLocalDB(), index}
}

func (action *Action) checkExecAccountBalance(fromAddr string, ToFrozen, ToActive int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= ToFrozen && acc.GetFrozen() >= ToActive {
		return true
	}
	return false
}

// GameCreate create game
func (action *Action) GameCreate(create *gt.GameCreate) (*types.Receipt, error) {
	gameID := common.ToHex(action.txhash)
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	maxGameAmount := getConfValue(action.db, ConfNameMaxGameAmount, MaxGameAmount)
	if create.GetValue() > maxGameAmount*types.Coin {
		glog.Error("Create the game, the deposit is too big  ", "value", create.GetValue(), "err", gt.ErrGameCreateAmount.Error())
		return nil, gt.ErrGameCreateAmount
	}
	minGameAmount := getConfValue(action.db, ConfNameMinGameAmount, MinGameAmount)
	if create.GetValue() < minGameAmount*types.Coin || math.Remainder(float64(create.GetValue()), 2) != 0 {
		return nil, fmt.Errorf("%s", "The amount you participate in cannot be less than 2 and must be an even number!")
	}
	if !action.checkExecAccountBalance(action.fromaddr, create.GetValue(), 0) {
		glog.Error("GameCreate", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			gameID, "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	//冻结子账户资金
	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, create.GetValue())
	if err != nil {
		glog.Error("GameCreate.ExecFrozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", create.GetValue(), "err", err.Error())
		return nil, err
	}
	game := &gt.Game{
		GameId:        gameID,
		Value:         create.GetValue(),
		HashType:      create.GetHashType(),
		HashValue:     create.GetHashValue(),
		CreateTime:    action.blocktime,
		CreateAddress: action.fromaddr,
		Status:        gt.GameActionCreate,
		CreateTxHash:  gameID,
	}
	//更新stateDB缓存，用于计数
	action.updateStateDBCache(game.GetStatus(), "")
	action.updateStateDBCache(game.GetStatus(), game.GetCreateAddress())
	game.Index = action.GetIndex(game)
	action.saveStateDB(game)
	receiptLog := action.GetReceiptLog(game)
	logs = append(logs, receiptLog)
	kv = append(kv, action.GetKVSet(game)...)
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	kv = append(kv, action.updateCount(game.GetStatus(), "")...)
	kv = append(kv, action.updateCount(game.GetStatus(), game.GetCreateAddress())...)
	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// GameMatch match game
func (action *Action) GameMatch(match *gt.GameMatch) (*types.Receipt, error) {
	game, err := action.readGame(match.GetGameId())
	if err != nil {
		glog.Error("GameMatch", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			match.GetGameId(), "err", err.Error())
		return nil, err
	}
	if game.GetStatus() != gt.GameActionCreate {
		glog.Error("GameMatch", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			match.GetGameId(), "err", gt.ErrGameMatchStatus.Error())
		return nil, gt.ErrGameMatchStatus
	}
	if game.GetCreateAddress() == action.fromaddr {
		glog.Error("GameMatch", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			match.GetGameId(), "err", gt.ErrGameMatch.Error())
		return nil, gt.ErrGameMatch
	}
	if !action.checkExecAccountBalance(action.fromaddr, game.GetValue()/2, 0) {
		glog.Error("GameMatch", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			match.GetGameId(), "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	//冻结 game value 中资金的一半
	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, game.GetValue()/2)
	if err != nil {
		glog.Error("GameMatch.ExecFrozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", game.GetValue()/2, "err", err.Error())
		return nil, err
	}
	game.Status = gt.GameActionMatch
	game.Value = game.GetValue()/2 + game.GetValue()
	game.MatchAddress = action.fromaddr
	game.MatchTime = action.blocktime
	game.MatcherGuess = match.GetGuess()
	game.MatchTxHash = common.ToHex(action.txhash)
	game.PrevIndex = game.GetIndex()
	game.Index = action.GetIndex(game)
	action.saveStateDB(game)
	action.updateStateDBCache(game.GetStatus(), "")
	action.updateStateDBCache(game.GetStatus(), game.GetCreateAddress())
	action.updateStateDBCache(game.GetStatus(), game.GetMatchAddress())
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	receiptLog := action.GetReceiptLog(game)
	logs = append(logs, receiptLog)
	kvs = append(kvs, action.GetKVSet(game)...)
	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)
	kvs = append(kvs, action.updateCount(game.GetStatus(), "")...)
	kvs = append(kvs, action.updateCount(game.GetStatus(), game.GetCreateAddress())...)
	kvs = append(kvs, action.updateCount(game.GetStatus(), game.GetMatchAddress())...)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

// GameCancel cancel game
func (action *Action) GameCancel(cancel *gt.GameCancel) (*types.Receipt, error) {
	game, err := action.readGame(cancel.GetGameId())
	if err != nil {
		glog.Error("GameCancel ", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			cancel.GetGameId(), "err", err.Error())
		return nil, err
	}
	if game.GetCreateAddress() != action.fromaddr {
		glog.Error("GameCancel ", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			cancel.GetGameId(), "err", gt.ErrGameCancleAddr.Error())
		return nil, gt.ErrGameCancleAddr
	}
	if game.GetStatus() != gt.GameActionCreate {
		glog.Error("GameCancel ", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			cancel.GetGameId(), "err", gt.ErrGameCancleStatus.Error())
		return nil, gt.ErrGameCancleStatus
	}
	if !action.checkExecAccountBalance(action.fromaddr, 0, game.GetValue()) {
		glog.Error("GameCancel", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			game.GetGameId(), "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	receipt, err := action.coinsAccount.ExecActive(game.GetCreateAddress(), action.execaddr, game.GetValue())
	if err != nil {
		glog.Error("GameCancel ", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			cancel.GetGameId(), "amount", game.GetValue(), "err", err.Error())
		return nil, err
	}
	game.Closetime = action.blocktime
	game.Status = gt.GameActionCancel
	game.CancelTxHash = common.ToHex(action.txhash)
	game.PrevIndex = game.GetIndex()
	game.Index = action.GetIndex(game)
	action.saveStateDB(game)
	action.updateStateDBCache(game.GetStatus(), "")
	action.updateStateDBCache(game.GetStatus(), game.GetCreateAddress())
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	logs = append(logs, receipt.Logs...)
	receiptLog := action.GetReceiptLog(game)
	logs = append(logs, receiptLog)
	kvs := action.GetKVSet(game)
	kv = append(kv, receipt.KV...)
	kv = append(kv, kvs...)
	kv = append(kv, action.updateCount(game.GetStatus(), "")...)
	kv = append(kv, action.updateCount(game.GetStatus(), game.GetCreateAddress())...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// GameClose close game
func (action *Action) GameClose(close *gt.GameClose) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	game, err := action.readGame(close.GetGameId())
	if err != nil {
		glog.Error("GameClose ", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			close.GetGameId(), "err", err.Error())
		return nil, err
	}
	//开奖时间控制
	if action.fromaddr != game.GetCreateAddress() && !action.checkGameIsTimeOut(game) {
		//如果不是游戏创建者开奖，则检查是否超时
		glog.Error(gt.ErrGameCloseAddr.Error())
		return nil, gt.ErrGameCloseAddr
	}
	if game.GetStatus() != gt.GameActionMatch {
		glog.Error(gt.ErrGameCloseStatus.Error())
		return nil, gt.ErrGameCloseStatus
	}
	//各自冻结余额检查
	if !action.checkExecAccountBalance(game.GetCreateAddress(), 0, 2*game.GetValue()/3) {
		glog.Error("GameClose", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "id",
			game.GetGameId(), "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	if !action.checkExecAccountBalance(game.GetMatchAddress(), 0, game.GetValue()/3) {
		glog.Error("GameClose", "addr", game.GetMatchAddress(), "execaddr", action.execaddr, "id",
			game.GetGameId(), "err", types.ErrNoBalance.Error())
		return nil, types.ErrNoBalance
	}
	result, creatorGuess := action.checkGameResult(game, close)
	if result == IsCreatorWin {
		//如果是庄家赢了，则解冻所有钱,并将对赌者相应冻结的钱转移到庄家的合约账户中
		receipt, err := action.coinsAccount.ExecActive(game.GetCreateAddress(), action.execaddr, 2*game.GetValue()/3)
		if err != nil {
			glog.Error("GameClose.execActive", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", 2*game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		receipt, err = action.coinsAccount.ExecTransferFrozen(game.GetMatchAddress(), game.GetCreateAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.GetCreateAddress(), action.execaddr, 2*game.GetValue()/3) // rollback
			glog.Error("GameClose.ExecTransferFrozen", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", 2*game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	} else if result == IsMatcherWin {
		//如果是庄家输了，则反向操作
		receipt, err := action.coinsAccount.ExecActive(game.GetCreateAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			glog.Error("GameClose.ExecActive", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		receipt, err = action.coinsAccount.ExecActive(game.GetMatchAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.GetCreateAddress(), action.execaddr, game.GetValue()/3) // rollback
			glog.Error("GameClose.ExecActive", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		receipt, err = action.coinsAccount.ExecTransferFrozen(game.GetCreateAddress(), game.GetMatchAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.GetCreateAddress(), action.execaddr, game.GetValue()/3) // rollback
			action.coinsAccount.ExecFrozen(game.GetMatchAddress(), action.execaddr, game.GetValue()/3)  // rollback
			glog.Error("GameClose.ExecTransferFrozen", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)

	} else if result == IsDraw {
		//平局是解冻各自的押注即可
		receipt, err := action.coinsAccount.ExecActive(game.GetCreateAddress(), action.execaddr, 2*game.GetValue()/3)
		if err != nil {
			glog.Error("GameClose.ExecActive", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", 2*game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		receipt, err = action.coinsAccount.ExecActive(game.GetMatchAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.GetCreateAddress(), action.execaddr, 2*game.GetValue()/3) // rollback
			glog.Error("GameClose.ExecActive", "addr", game.GetMatchAddress(), "execaddr", action.execaddr, "amount", game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	} else if result == IsTimeOut {
		//开奖超时，庄家输掉所有筹码
		receipt, err := action.coinsAccount.ExecActive(game.GetMatchAddress(), action.execaddr, game.GetValue()/3)
		if err != nil {
			glog.Error("GameClose.ExecActive", "addr", game.GetCreateAddress(), "execaddr", action.execaddr, "amount", 2*game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
		receipt, err = action.coinsAccount.ExecTransferFrozen(game.GetCreateAddress(), game.GetMatchAddress(), action.execaddr, 2*game.GetValue()/3)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.GetMatchAddress(), action.execaddr, game.GetValue()/3) // rollback
			glog.Error("GameClose.ExecTransferFrozen", "addr", game.GetMatchAddress(), "execaddr", action.execaddr, "amount", game.GetValue()/3,
				"err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}
	game.Closetime = action.blocktime
	game.Status = gt.GameActionClose
	game.Secret = close.GetSecret()
	game.Result = result
	game.CloseTxHash = common.ToHex(action.txhash)
	game.PrevIndex = game.GetIndex()
	game.Index = action.GetIndex(game)
	game.CreatorGuess = creatorGuess
	action.saveStateDB(game)
	action.updateStateDBCache(game.GetStatus(), "")
	action.updateStateDBCache(game.GetStatus(), game.GetCreateAddress())
	action.updateStateDBCache(game.GetStatus(), game.GetMatchAddress())
	receiptLog := action.GetReceiptLog(game)
	logs = append(logs, receiptLog)
	kvs := action.GetKVSet(game)
	kv = append(kv, kvs...)
	kv = append(kv, action.updateCount(game.GetStatus(), "")...)
	kv = append(kv, action.updateCount(game.GetStatus(), game.GetCreateAddress())...)
	kv = append(kv, action.updateCount(game.GetStatus(), game.GetMatchAddress())...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// 检查开奖是否超时，若超过一天，则不让庄家开奖，但其他人可以开奖，
// 若没有一天，则其他人没有开奖权限，只有庄家有开奖权限
func (action *Action) checkGameIsTimeOut(game *gt.Game) bool {
	activeTime := getConfValue(action.db, ConfNameActiveTime, ActiveTime)
	DurTime := 60 * 60 * activeTime
	return action.blocktime > (game.GetMatchTime() + DurTime)
}

//根据传入密钥，揭晓游戏结果
func (action *Action) checkGameResult(game *gt.Game, close *gt.GameClose) (int32, int32) {
	//如果超时，直接走超时开奖逻辑
	if action.checkGameIsTimeOut(game) {
		return IsTimeOut, Unknown
	}
	if bytes.Equal(common.Sha256([]byte(close.GetSecret()+string(Rock))), game.GetHashValue()) {
		//此刻庄家出的是石头
		if game.GetMatcherGuess() == Rock {
			return IsDraw, Rock
		} else if game.GetMatcherGuess() == Scissor {
			return IsCreatorWin, Rock
		} else if game.GetMatcherGuess() == Paper {
			return IsMatcherWin, Rock
		}
		//其他情况说明matcher 使坏，填了其他值，当做作弊处理
		return IsCreatorWin, Rock

	} else if bytes.Equal(common.Sha256([]byte(close.GetSecret()+string(Scissor))), game.GetHashValue()) {
		//此刻庄家出的剪刀
		if game.GetMatcherGuess() == Rock {
			return IsMatcherWin, Scissor
		} else if game.GetMatcherGuess() == Scissor {
			return IsDraw, Scissor
		} else if game.GetMatcherGuess() == Paper {
			return IsCreatorWin, Scissor
		}
		return IsCreatorWin, Scissor

	} else if bytes.Equal(common.Sha256([]byte(close.GetSecret()+string(Paper))), game.GetHashValue()) {
		//此刻庄家出的是布
		if game.GetMatcherGuess() == Rock {
			return IsCreatorWin, Paper
		} else if game.GetMatcherGuess() == Scissor {
			return IsMatcherWin, Paper
		} else if game.GetMatcherGuess() == Paper {
			return IsDraw, Paper
		}
		return IsCreatorWin, Paper

	}
	//其他情况默认是matcher win
	return IsMatcherWin, Unknown
}

func (action *Action) readGame(id string) (*gt.Game, error) {
	data, err := action.db.Get(Key(id))
	if err != nil {
		return nil, err
	}
	var game gt.Game
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

// List query game list
func List(db dbm.Lister, stateDB dbm.KV, param *gt.QueryGameListByStatusAndAddr) (types.Message, error) {
	return QueryGameListByPage(db, stateDB, param)
}

// QueryGameListByPage 分页查询
func QueryGameListByPage(db dbm.Lister, stateDB dbm.KV, param *gt.QueryGameListByStatusAndAddr) (types.Message, error) {
	switch param.GetStatus() {
	case gt.GameActionCreate, gt.GameActionMatch, gt.GameActionClose, gt.GameActionCancel:
		return queryGameListByStatusAndAddr(db, stateDB, param)
	}
	return nil, fmt.Errorf("%s", "the status only fill in 1,2,3,4!")
}

func queryGameListByStatusAndAddr(db dbm.Lister, stateDB dbm.KV, param *gt.QueryGameListByStatusAndAddr) (types.Message, error) {
	direction := ListDESC
	if param.GetDirection() == ListASC {
		direction = ListASC
	}
	count := int32(getConfValue(stateDB, ConfNameDefaultCount, DefaultCount))
	maxCount := int32(getConfValue(stateDB, ConfNameMaxCount, MaxCount))
	if 0 < param.GetCount() && param.GetCount() <= maxCount {
		count = param.GetCount()
	}
	var prefix []byte
	var key []byte
	if param.GetAddress() == "" {
		prefix = calcGameStatusIndexPrefix(param.Status)
		key = calcGameStatusIndexKey(param.Status, param.GetIndex())
	} else {
		prefix = calcGameAddrIndexPrefix(param.Status, param.GetAddress())
		key = calcGameAddrIndexKey(param.Status, param.GetAddress(), param.GetIndex())
	}
	var values [][]byte
	var err error
	if param.GetIndex() == 0 { //第一次查询
		values, err = db.List(prefix, nil, count, direction)
	} else {
		values, err = db.List(prefix, key, count, direction)
	}
	if err != nil {
		return nil, err
	}
	var gameIds []string
	for _, value := range values {
		var record gt.GameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		gameIds = append(gameIds, record.GetGameId())
	}
	return &gt.ReplyGameList{Games: GetGameList(stateDB, gameIds)}, nil
}

// QueryGameListCount count数查询
func QueryGameListCount(stateDB dbm.KV, param *gt.QueryGameListCount) (types.Message, error) {
	if param.Status < 1 || param.Status > 4 {
		return nil, fmt.Errorf("%s", "the status only fill in 1,2,3,4!")
	}
	return &gt.ReplyGameListCount{Count: QueryCountByStatusAndAddr(stateDB, param.GetStatus(), param.GetAddress())}, nil
}

// QueryCountByStatusAndAddr query game count by status and addr
func QueryCountByStatusAndAddr(stateDB dbm.KV, status int32, addr string) int64 {
	switch status {
	case gt.GameActionCreate, gt.GameActionMatch, gt.GameActionCancel, gt.GameActionClose:
		count, _ := queryCountByStatusAndAddr(stateDB, status, addr)
		return count
	}
	glog.Error("the status only fill in 1,2,3,4!")
	return 0
}
func queryCountByStatusAndAddr(stateDB dbm.KV, status int32, addr string) (int64, error) {
	data, err := stateDB.Get(calcCountKey(status, addr))
	if err != nil {
		glog.Error("queryCountByStatusAndAddr", "query count have err:", err.Error())
		return 0, err
	}
	count, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		glog.Error("queryCountByStatusAndAddr", "Type conversion error:", err.Error())
		return 0, err
	}
	return count, nil
}

func readGame(db dbm.KV, id string) (*gt.Game, error) {
	data, err := db.Get(Key(id))
	if err != nil {
		glog.Error("readGame", "query data have err:", err.Error())
		return nil, err
	}
	var game gt.Game
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		glog.Error("readGame", "decode game have err:", err.Error())
		return nil, err
	}
	return &game, nil
}

// QueryGameListByIds query game list by gameIds
func QueryGameListByIds(db dbm.KV, infos *gt.QueryGameInfos) (types.Message, error) {
	var games []*gt.Game
	for i := 0; i < len(infos.GameIds); i++ {
		id := infos.GameIds[i]
		game, err := readGame(db, id)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return &gt.ReplyGameList{Games: games}, nil
}

// GetGameList 安全批量查询方式,防止因为脏数据导致查询接口奔溃
func GetGameList(db dbm.KV, values []string) []*gt.Game {
	var games []*gt.Game
	for _, value := range values {
		game, err := readGame(db, value)
		if err != nil {
			continue
		}
		games = append(games, game)
	}
	return games
}
func getConfValue(db dbm.KV, key string, defaultValue int64) int64 {
	var item types.ConfigItem
	value, err := getManageKey(key, db)
	if err != nil {
		return defaultValue
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			glog.Error("gamedb getConfValue", "decode db key:", key, "err", err.Error())
			return defaultValue
		}
	}
	values := item.GetArr().GetValue()
	if len(values) == 0 {
		glog.Error("gamedb getConfValue", "can't get value from values arr. key:", key)
		return defaultValue
	}
	//取数组最后一位，作为最新配置项的值
	v, err := strconv.ParseInt(values[len(values)-1], 10, 64)
	if err != nil {
		glog.Error("gamedb getConfValue", "Type conversion error:", err.Error())
		return defaultValue
	}
	return v
}
func getManageKey(key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		if types.IsPara() { //平行链只有一种存储方式
			glog.Error("gamedb getManage", "can't get value from db,key:", key, "err", err.Error())
			return nil, err
		}
		glog.Debug("gamedb getManageKey", "get db key", "not found")
		return getConfigKey(key, db)
	}
	return value, nil
}

func getConfigKey(key string, db dbm.KV) ([]byte, error) {
	configKey := types.ConfigKey(key)
	value, err := db.Get([]byte(configKey))
	if err != nil {
		glog.Error("gamedb getConfigKey", "can't get value from db,key:", key, "err", err.Error())
		return nil, err
	}
	return value, nil
}

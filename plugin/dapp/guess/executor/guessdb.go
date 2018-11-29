// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"errors"
	"fmt"
	"github.com/33cn/chain33/client"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	pkt "github.com/33cn/plugin/plugin/dapp/guess/types"
)

const (
	ListDESC = int32(0)
	ListASC  = int32(1)

	DefaultCount   = int32(20) //默认一次取多少条记录
	MaxBets = 10000            //一次最多下多少注
	MaxBetsNumber = 100000     //一局游戏最多接受多少注
	MaxHeight = 10000000000    //最大区块高度

	MinBetBlockNum = 720          //从创建游戏开始，一局游戏最少的可下注区块数量
	MinBetTimeInterval = "2h"     //从创建游戏开始，一局游戏最短的可下注时间
	MinBetTimeoutNum = 8640       //从游戏结束下注开始，一局游戏最少的超时块数
	MinBetTimeoutInterval = "24h" //从游戏结束下注开始，一局游戏最短的超时时间

	MIN_PLAY_VALUE = 10 * types.Coin
	//DefaultStyle   = pkt.PlayStyleDefault
	MinOneBet = 1
)

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
	api          client.QueueProtocolAPI
}

func NewAction(guess *Guess, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromAddr := tx.From()

	return &Action{
		coinsAccount: guess.GetCoinsAccount(),
		db: guess.GetStateDB(),
		txhash: hash,
		fromaddr: fromAddr,
		blocktime: guess.GetBlockTime(),
		height: guess.GetHeight(),
		execaddr: dapp.ExecAddress(string(tx.Execer)),
		localDB: guess.GetLocalDB(),
		index: index,
		api: guess.GetApi(),
	}
}

func (action *Action) CheckExecAccountBalance(fromAddr string, ToFrozen, ToActive int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= ToFrozen && acc.GetFrozen() >= ToActive {
		return true
	}
	return false
}

func Key(id string) (key []byte) {
	key = append(key, []byte("mavl-"+types.ExecName(pkt.GuessX)+"-")...)
	key = append(key, []byte(id)...)
	return key
}

func readGame(db dbm.KV, id string) (*pkt.GuessGame, error) {
	data, err := db.Get(Key(id))
	if err != nil {
		logger.Error("query data have err:", err.Error())
		return nil, err
	}
	var game pkt.GuessGame
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		logger.Error("decode game have err:", err.Error())
		return nil, err
	}
	return &game, nil
}

//安全批量查询方式,防止因为脏数据导致查询接口奔溃
func GetGameList(db dbm.KV, values []string) []*pkt.GuessGame {
	var games []*pkt.GuessGame
	for _, value := range values {
		game, err := readGame(db, value)
		if err != nil {
			continue
		}
		games = append(games, game)
	}
	return games
}

func Infos(db dbm.KV, infos *pkt.QueryGuessGameInfos) (types.Message, error) {
	var games []*pkt.GuessGame
	for i := 0; i < len(infos.GameIds); i++ {
		id := infos.GameIds[i]
		game, err := readGame(db, id)
		if err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return &pkt.ReplyGuessGameInfos{Games: games}, nil
}

func getGameListByAddr(db dbm.Lister, addr string, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameAddrPrefix(addr), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameAddrPrefix(addr), calcGuessGameAddrKey(addr, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var gameIds []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		gameIds = append(gameIds, &record)
	}

	return &pkt.GuessGameRecords{Records: gameIds}, nil
}

func getGameListByStatus(db dbm.Lister, status int32, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameStatusPrefix(status), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameStatusPrefix(status), calcGuessGameStatusKey(status, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var gameIds []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		gameIds = append(gameIds, &record)
	}

	return &pkt.GuessGameRecords{Records: gameIds}, nil
}

func queryGameListByStatusAndPlayer(db dbm.Lister, stat int32, player int32, value int64) ([]string, error) {
	values, err := db.List(calcPBGameStatusAndPlayerPrefix(stat, player, value), nil, DefaultCount, ListDESC)
	if err != nil {
		return nil, err
	}

	var gameIds []string
	for _, value := range values {
		var record pkt.PBGameIndexRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		gameIds = append(gameIds, record.GetGameId())
	}

	return gameIds, nil
}

func (action *Action) saveGame(game *pkt.GuessGame) (kvset []*types.KeyValue) {
	value := types.Encode(game)
	action.db.Set(Key(game.GetGameId()), value)
	kvset = append(kvset, &types.KeyValue{Key: Key(game.GameId), value})
	return kvset
}

func (action *Action) getIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

func (action *Action) GetReceiptLog(game *pkt.GuessGame, statusChange bool) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &pkt.ReceiptGuessGame{}
	r.Addr = action.fromaddr
	if game.Status == pkt.GuessGameStatusStart {
		log.Ty = pkt.TyLogGuessGameStart
	} else if game.Status == pkt.GuessGameStatusBet {
		log.Ty = pkt.TyLogGuessGameBet
	} else if game.Status == pkt.GuessGameStatusStopBet {
		log.Ty = pkt.TyLogGuessGameStopBet
	} else if game.Status == pkt.GuessGameStatusAbort {
		log.Ty = pkt.TyLogGuessGameAbort
	} else if game.Status == pkt.GuessGameStatusPublish {
		log.Ty = pkt.TyLogGuessGamePublish
	} else if game.Status == pkt.GuessGameStatusTimeOut {
		log.Ty = pkt.TyLogGuessGameTimeout
	}

	r.Index = game.Index
	r.GameId = game.GameId
	r.Status = game.Status
	r.AdminAddr = game.AdminAddr
	r.PreStatus = game.PreStatus
	r.StatusChange = statusChange
	r.PreIndex = game.PreIndex
	log.Log = types.Encode(r)
	return log
}

func (action *Action) readGame(id string) (*pkt.GuessGame, error) {
	data, err := action.db.Get(Key(id))
	if err != nil {
		return nil, err
	}
	var game pkt.PokerBull
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
}

func (action *Action) calculate(game *pkt.PokerBull) *pkt.PBResult {
	var handS HandSlice
	for _, player := range game.Players {
		hand := &pkt.PBHand{}
		hand.Cards = Deal(game.Poker, player.TxHash) //发牌
		hand.Result = Result(hand.Cards)             //计算结果
		hand.Address = player.Address

		//存入玩家数组
		player.Hands = append(player.Hands, hand)

		//存入临时切片待比大小排序
		handS = append(handS, hand)

		//为下一个continue状态初始化player
		player.Ready = false
	}

	// 升序排列
	if !sort.IsSorted(handS) {
		sort.Sort(handS)
	}
	winner := handS[len(handS)-1]

	// 将有序的临时切片加入到结果数组
	result := &pkt.PBResult{}
	result.Winner = winner.Address
	//TODO Dealer:暂时不支持倍数
	//result.Leverage = Leverage(winner)
	result.Hands = make([]*pkt.PBHand, len(handS))
	copy(result.Hands, handS)

	game.Results = append(game.Results, result)
	return result
}

func (action *Action) calculateDealer(game *pkt.PokerBull) *pkt.PBResult {
	var handS HandSlice
	var dealer *pkt.PBHand
	for _, player := range game.Players {
		hand := &pkt.PBHand{}
		hand.Cards = Deal(game.Poker, player.TxHash) //发牌
		hand.Result = Result(hand.Cards)             //计算结果
		hand.Address = player.Address

		//存入玩家数组
		player.Hands = append(player.Hands, hand)

		//存入临时切片待比大小排序
		handS = append(handS, hand)

		//为下一个continue状态初始化player
		player.Ready = false

		//记录庄家
		if player.Address == game.DealerAddr {
			dealer = hand
		}
	}

	for _, hand := range handS {
		if hand.Address == game.DealerAddr {
			continue
		}

		if CompareResult(hand, dealer) {
			hand.IsWin = false
		} else {
			hand.IsWin = true
			hand.Leverage = Leverage(hand)
		}
	}

	// 将有序的临时切片加入到结果数组
	result := &pkt.PBResult{}
	result.Dealer = game.DealerAddr
	result.DealerLeverage = Leverage(dealer)
	result.Hands = make([]*pkt.PBHand, len(handS))
	copy(result.Hands, handS)

	game.Results = append(game.Results, result)
	return result
}

func (action *Action) nextDealer(game *pkt.PokerBull) string {
	var flag = -1
	for i, player := range game.Players {
		if player.Address == game.DealerAddr {
			flag = i
		}
	}
	if flag == -1 {
		logger.Error("Get next dealer failed.")
		return game.DealerAddr
	}

	if flag == len(game.Players)-1 {
		return game.Players[0].Address
	}

	return game.Players[flag+1].Address
}

func (action *Action) settleDealerAccount(lastAddress string, game *pkt.PokerBull) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	result := action.calculateDealer(game)
	for _, hand := range result.Hands {
		// 最后一名玩家没有冻结
		if hand.Address != lastAddress {
			receipt, err := action.coinsAccount.ExecActive(hand.Address, action.execaddr, game.GetValue()*POKERBULL_LEVERAGE_MAX)
			if err != nil {
				logger.Error("GameSettleDealer.ExecActive", "addr", hand.Address, "execaddr", action.execaddr, "amount", game.GetValue(),
					"err", err)
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		//给赢家转账
		var receipt *types.Receipt
		var err error
		if hand.Address != result.Dealer {
			if hand.IsWin {
				receipt, err = action.coinsAccount.ExecTransfer(result.Dealer, hand.Address, action.execaddr, game.GetValue()*int64(hand.Leverage))
				if err != nil {
					action.coinsAccount.ExecFrozen(hand.Address, action.execaddr, game.GetValue()) // rollback
					logger.Error("GameSettleDealer.ExecTransfer", "addr", hand.Address, "execaddr", action.execaddr,
						"amount", game.GetValue()*int64(hand.Leverage), "err", err)
					return nil, nil, err
				}
			} else {
				receipt, err = action.coinsAccount.ExecTransfer(hand.Address, result.Dealer, action.execaddr, game.GetValue()*int64(result.DealerLeverage))
				if err != nil {
					action.coinsAccount.ExecFrozen(hand.Address, action.execaddr, game.GetValue()) // rollback
					logger.Error("GameSettleDealer.ExecTransfer", "addr", hand.Address, "execaddr", action.execaddr,
						"amount", game.GetValue()*int64(result.DealerLeverage), "err", err)
					return nil, nil, err
				}
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}
	game.DealerAddr = action.nextDealer(game)

	return logs, kv, nil
}

func (action *Action) settleDefaultAccount(lastAddress string, game *pkt.PokerBull) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue
	result := action.calculate(game)

	for _, player := range game.Players {
		// 最后一名玩家没有冻结
		if player.Address != lastAddress {
			receipt, err := action.coinsAccount.ExecActive(player.GetAddress(), action.execaddr, game.GetValue()*POKERBULL_LEVERAGE_MAX)
			if err != nil {
				logger.Error("GameSettleDefault.ExecActive", "addr", player.GetAddress(), "execaddr", action.execaddr,
					"amount", game.GetValue()*POKERBULL_LEVERAGE_MAX, "err", err)
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}

		//给赢家转账
		if player.Address != result.Winner {
			receipt, err := action.coinsAccount.ExecTransfer(player.Address, result.Winner, action.execaddr, game.GetValue() /**int64(result.Leverage)*/) //TODO Dealer:暂时不支持倍数
			if err != nil {
				action.coinsAccount.ExecFrozen(result.Winner, action.execaddr, game.GetValue()) // rollback
				logger.Error("GameSettleDefault.ExecTransfer", "addr", result.Winner, "execaddr", action.execaddr,
					"amount", game.GetValue() /**int64(result.Leverage)*/, "err", err) //TODO Dealer:暂时不支持倍数
				return nil, nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}

	return logs, kv, nil
}

func (action *Action) settleAccount(lastAddress string, game *pkt.PokerBull) ([]*types.ReceiptLog, []*types.KeyValue, error) {
	if DefaultStyle == pkt.PlayStyleDealer {
		return action.settleDealerAccount(lastAddress, game)
	} else {
		return action.settleDefaultAccount(lastAddress, game)
	}
}

func (action *Action) genTxRnd(txhash []byte) (int64, error) {
	randbyte := make([]byte, 7)
	for i := 0; i < 7; i++ {
		randbyte[i] = txhash[i]
	}

	randstr := common.ToHex(randbyte)
	randint, err := strconv.ParseInt(randstr, 0, 64)
	if err != nil {
		return 0, err
	}

	return randint, nil
}

// 新建一局游戏
func (action *Action) newGame(gameId string, start *pkt.GuessGameStart) (*pkt.GuessGame, error) {
	game := &pkt.GuessGame{
		GameId:      gameId,
		Status:      pkt.GuessGameActionStart,
		//StartTime:   action.blocktime,
		StartTxHash: gameId,
		Topic:       start.Topic,
		Category:    start.Category,
		Options:     start.Options,
		MaxTime:     start.MaxTime,
		MaxHeight:   start.MaxHeight,
		Symbol:      start.Symbol,
		Exec:        start.Exec,
		OneBet:      start.OneBet,
		MaxBets:     start.MaxBets,
		MaxBetsNumber: start.MaxBetsNumber,
		Fee: start.Fee,
		FeeAddr: start.FeeAddr,
		Expire: start.Expire,
		ExpireHeight: start.ExpireHeight,
		//AdminAddr: action.fromaddr,
		BetsNumber: 0,
		//Index:       action.getIndex(game),
	}

	return game, nil
}


func (action *Action) GameStart(start *pkt.GuessGameStart) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if start.MaxHeight >= MaxHeight {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum height number is %d which is less thanstart.MaxHeight %d", MaxHeight, start.MaxHeight))
		return nil, types.ErrInvalidParam
	}

	if start.MaxBetsNumber >= MaxBetsNumber {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum bets number is %d which is less than start.MaxBetsNumber %d", MaxBetsNumber, start.MaxBetsNumber))
		return nil, types.ErrInvalidParam
	}

	if len(start.Topic) == 0 || len(start.Options) == 0 {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("Illegal parameters,Topic:%s | options: %s | category: %s", start.Topic, start.Options, start.Category))
		return nil, types.ErrInvalidParam
	}

	options, ok := GetOptions(start.Options)
	if !ok {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The options is illegal:%s", start.Options))
		return nil, types.ErrInvalidParam
	}

	if !action.CheckTime() {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The height and time parameters are illegal:MaxTime %s MaxHeight %d Expire %s, ExpireHeight %d", start.MaxTime, start.MaxHeight, start.Expire, start.ExpireHeight))
		return nil, types.ErrInvalidParam
	}

	if len(start.Symbol) == 0 {
		start.Symbol = "bty"
	}

	if len(start.Exec) == 0 {
		start.Exec = "coins"
	}

	if start.OneBet < MinOneBet {
		start.OneBet = MinOneBet
	}

	if start.MaxBets >= MaxBets {
		start.MaxBets = MaxBets
	}

	gameId := common.ToHex(action.txhash)
	game, err := action.newGame(gameId, start)
	if err != nil {
		return nil, err
	}
	game.StartTime = action.blocktime
	game.AdminAddr = action.fromaddr
	game.PreIndex = 0
	game.Index = action.getIndex()
	game.Status = pkt.GuessGameStatusStart
	game.BetStat.TotalBetTimes = 0
	game.BetStat.TotalBetsNumber = 0
    for i := 0; i < len(options); i++ {
		item := &pkt.GuessBetStatItem{Option: options[i], BetsNumber: 0, BetsTimes: 0}
		game.BetStat.Items = append(game.BetStat.Items, item)
	}

	receiptLog := action.GetReceiptLog(game, false)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func getReadyPlayerNum(players []*pkt.PBPlayer) int {
	var readyC = 0
	for _, player := range players {
		if player.Ready {
			readyC++
		}
	}
	return readyC
}

func getPlayerFromAddress(players []*pkt.PBPlayer, addr string) *pkt.PBPlayer {
	for _, player := range players {
		if player.Address == addr {
			return player
		}
	}
	return nil
}

func (action *Action) GameBet(pbBet *pkt.GuessGameBet) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbBet.GetGameId())
	if err != nil {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbBet.GetGameId(), "err", err)
		return nil, err
	}

	prevStatus := game.Status
	if game.Status != pkt.GuessGameStatusStart && game.Status != pkt.GuessGameStatusBet && game.Status != pkt.GuessGameStatusStopBet{
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, errors.New("ErrGameStatus")
	}

	canBet := action.RefreshStatusByTime(game)

	if canBet == false {
		var receiptLog *types.ReceiptLog
		if prevStatus != game.Status {
			receiptLog = action.GetReceiptLog(game, true)
		} else {
			receiptLog = action.GetReceiptLog(game, false)
		}

		logs = append(logs, receiptLog)
		kv = append(kv, action.saveGame(game)...)

		return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
	}

	//检查竞猜选项是否合法
	options, legal := GetOptions(game.GetOptions())
	if !legal || len(options) == 0{
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Game Options illegal",
			game.GetOptions())
		return nil, types.ErrInvalidParam
	}

	if !IsLegalOption(options, pbBet.GetOption()) {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Option illegal",
			pbBet.GetOption())
		return nil, types.ErrInvalidParam
	}

	//检查下注金额是否超限，如果超限，按最大值
	if pbBet.GetBetsNum() > game.GetMaxBets() {
		pbBet.BetsNum = game.GetMaxBets()
	}

	if game.BetsNumber + pbBet.GetBetsNum() > game.MaxBetsNumber {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "MaxBetsNumber over limit",
			game.MaxBetsNumber, "current Bets Number", game.BetsNumber)
		return nil, types.ErrInvalidParam
	}

	// 检查余额账户余额
	checkValue := int64(game.GetOneBet() * pbBet.BetsNum + game.Fee)
	if !action.CheckExecAccountBalance(action.fromaddr, checkValue, 0) {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			pbBet.GetGameId(), "err", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, checkValue)
	if err != nil {
		logger.Error("GameCreate.ExecFrozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", checkValue, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	//如果当前游戏状态可以下注，统一设定游戏状态为GuessGameStatusBet
	action.ChangeStatus(game, pkt.GuessGameStatusBet)
	action.AddGuessBet(game, pbBet)

	var receiptLog *types.ReceiptLog
	if prevStatus != game.Status {
		receiptLog = action.GetReceiptLog(game, true)
	} else {
		receiptLog = action.GetReceiptLog(game, false)
	}
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *Action) AddGuessBet(game *pkt.GuessGame, pbBet *pkt.GuessGameBet) {
	bet := &pkt.GuessBet{ Option: pbBet.GetOption(), BetsNumber: pbBet.BetsNum, Index: game.Index}
	player := &pkt.GuessPlayer{ Addr: action.fromaddr, Bet: bet}
	game.Plays = append(game.Plays, player)

	for i := 0; i < len(game.BetStat.Items); i ++ {
		if game.BetStat.Items[i].Option == pbBet.GetOption() {
			//针对具体选项更新统计项
			game.BetStat.Items[i].BetsNumber += pbBet.GetBetsNum()
			game.BetStat.Items[i].BetsTimes++

			//更新整体统计
			game.BetStat.TotalBetsNumber += pbBet.GetBetsNum()
			game.BetStat.TotalBetTimes++
			break
		}
	}

	game.BetsNumber += pbBet.GetBetsNum()
}

func (action *Action) GamePublish(publish *pkt.GuessGamePublish) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(publish.GetGameId())
	if err != nil {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			publish.GetGameId(), "err", err)
		return nil, err
	}

	//只有adminAddr可以发起publish
	if game.AdminAddr != action.fromaddr {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "fromAddr is not adminAddr",
			action.fromaddr, "adminAddr", game.AdminAddr)
		return nil, types.ErrInvalidParam
	}

	prevStatus := game.Status
	if game.Status != pkt.GuessGameStatusStart && game.Status != pkt.GuessGameStatusBet && game.Status != pkt.GuessGameStatusStopBet{
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, errors.New("ErrGameStatus")
	}

	//检查竞猜选项是否合法
	options, legal := GetOptions(game.GetOptions())
	if !legal || len(options) == 0{
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Game Options illegal",
			game.GetOptions())
		return nil, types.ErrInvalidParam
	}

	if !IsLegalOption(options, publish.GetResult()) {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Option illegal",
			publish.GetResult())
		return nil, types.ErrInvalidParam
	}

	game.Result = publish.Result

	//先遍历所有下注数据，对于输家，转移资金到Admin账户合约地址；
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		value := int64(player.Bet.BetsNumber * game.OneBet + game.Fee)
		receipt, err := action.coinsAccount.ExecTransfer(player.Addr, game.AdminAddr, action.execaddr, value)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, value) // rollback
			logger.Error("GamePublish", "addr", game.AdminAddr, "execaddr", action.execaddr,
				"amount", value, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	action.ChangeStatus(game, pkt.GuessGameStatusPublish)
	//计算竞猜正确的筹码总数
	totalBetsNumber := game.BetStat.TotalBetsNumber
	winBetsNumber := uint32(0)
	for j := 0; j < len(game.BetStat.Items); j++ {
		if game.BetStat.Items[j].Option == game.Result {
			winBetsNumber = game.BetStat.Items[j].BetsNumber
		}
	}

	//再遍历赢家，按照投注占比分配所有筹码
	for j := 0; j < len(game.Plays); j++ {
		player := game.Plays[j]
		if player.Bet.Option == game.Result {
			value := int64(player.Bet.BetsNumber * totalBetsNumber * game.OneBet/ winBetsNumber)
			receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, player.Addr, action.execaddr, value)
			if err != nil {
				action.coinsAccount.ExecFrozen(player.Addr, action.execaddr, value) // rollback
				logger.Error("GamePublish", "addr", player.Addr, "execaddr", action.execaddr,
					"amount", value, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}

	//如果设置了手续费专用地址，则将本局游戏收取的手续费转移到专用地址
	if game.Fee > 0 && len(game.FeeAddr) != 0 && game.FeeAddr != game.AdminAddr {
		value := int64(game.Fee * uint32(len(game.Plays))
		receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, game.FeeAddr, action.execaddr, value)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.FeeAddr, action.execaddr, value) // rollback
			logger.Error("GamePublish", "addr", game.FeeAddr, "execaddr", action.execaddr, "amount", value, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	var receiptLog *types.ReceiptLog
	if prevStatus != game.Status {
		receiptLog = action.GetReceiptLog(game, true)
	} else {
		receiptLog = action.GetReceiptLog(game, false)
	}
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *Action) GameAbort(pbend *pkt.GuessGameAbort) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbend.GetGameId())
	if err != nil {
		logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbend.GetGameId(), "err", err)
		return nil, err
	}

	if game.Status == pkt.GuessGameStatusPublish ||  game.Status == pkt.GuessGameStatusAbort{

		logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "game status not allow abort",
			game.Status)
		return nil, err
	}

	preStatus := game.Status
	//根据区块链高度或时间刷新游戏状态。
	action.RefreshStatusByTime(game)

	//如果游戏超时，则任何地址都可以Abort，否则只有创建游戏的地址可以Abort
	if game.Status != pkt.GuessGameStatusTimeOut {
		if game.AdminAddr != action.fromaddr {
			logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "Only admin can abort",
				action.fromaddr, "status", game.Status)
			return nil, err
		}
	}

	//激活冻结账户
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		value := int64(player.Bet.BetsNumber * game.OneBet + game.Fee)
		receipt, err := action.coinsAccount.ExecActive(player.Addr, action.execaddr, value)
		if err != nil {
			logger.Error("GameAbort", "addr", player.Addr, "execaddr", action.execaddr, "amount", value, "err", err)
			continue
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if game.Status != preStatus {
		//说明action.RefreshStatusByTime(game)调用时已经更新过状态和index了，这里直接再改状态就行了。
		game.Status = pkt.GuessGameStatusAbort
	} else {
		action.ChangeStatus(game, pkt.GuessGameStatusAbort)
	}

	receiptLog := action.GetReceiptLog(game, true)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func GetOptions(strOptions string) (options []string, legal bool){
	legal = true
	items := strings.Split(strOptions, ";")
	for i := 0 ; i < len(items); i++ {
		item := strings.Split(items[i],":")
		for j := 0; j < len(options); j++ {
			if item[0] == options[j] {
				legal = false
				return
			}
		}

		options = append(options, item[0])
	}

	return options, legal
}

func IsLegalOption(options []string, option string) bool {
	for i := 0; i < len(options); i++ {
		if options[i] == option {
			return true
		}
	}

	return false
}

func (action *Action) ChangeStatus(game *pkt.GuessGame, destStatus uint32) {
	if game.Status != destStatus {
		game.PreStatus = game.Status
		game.PreIndex = game.Index
		game.Status = destStatus
		game.Index = action.getIndex()
	}

	return
}
func (action *Action) RefreshStatusByTime(game *pkt.GuessGame) (canBet bool) {
	// 检查区块高度是否超过最大下注高度限制，看是否可以下注
	if game.GetMaxHeight() <= action.height {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Height over limit",
			action.height, "MaxHeight", game.GetMaxHeight())
		if game.GetExpireHeight() > action.height {
			action.ChangeStatus(game, pkt.GuessGameStatusStopBet)
		} else {
			action.ChangeStatus(game, pkt.GuessGameStatusTimeOut)
		}

		canBet = false
        return  canBet
	}

	// 检查区块高度是否超过下注时间限制，看是否可以下注
	if len(game.GetMaxTime()) > 0 {
		tMax, err := time.Parse("2006-01-02 15:04:05", game.GetMaxTime())
		if err != nil {
			logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse MaxTime failed",
				game.GetMaxTime())
			canBet = true
			return canBet
		}

		tExpire, err := time.Parse("2006-01-02 15:04:05", game.GetExpire())
		if err != nil {
			logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse Expire failed",
				game.GetExpire())
			canBet = true
			return canBet
		}

		tNow := time.Now()
		if tNow.After(tMax) {
			logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Time over MaxTime",
				game.GetMaxTime())

			if tNow.After(tExpire) {
				action.ChangeStatus(game, pkt.GuessGameStatusTimeOut)
			} else {
				action.ChangeStatus(game, pkt.GuessGameStatusStopBet)
			}

			canBet = false
			return canBet
		}
	}

	canBet = true
	return canBet
}

func (action *Action) CheckTime(start *pkt.GuessGameStart) bool {
	if action.height + MinBetBlockNum > start.MaxHeight || start.MaxHeight + MinBetTimeoutNum > start.ExpireHeight {
		return false
	}

	tNow := time.Now()
	d1, _ := time.ParseDuration(MinBetTimeInterval)      //最短开奖时间
	d2, _ := time.ParseDuration(MinBetTimeoutInterval)   //最短游戏过期时间
	if len(start.GetMaxTime()) == 0 {
		tNow.Add(d1)
		start.MaxTime = tNow.Format("2006-01-02 15:04:05")
	}

	if len(start.GetExpire()) == 0 {
		tMax, _ := time.Parse("2006-01-02 15:04:05", start.GetMaxTime())
		tMax.Add(d2)
		start.Expire = tMax.Format("2006-01-02 15:04:05")
	}

	tMax, err := time.Parse("2006-01-02 15:04:05", start.GetMaxTime())
	if err != nil {
		logger.Error("CheckTime", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse MaxTime failed",
			start.GetMaxTime())
		return false
	}

	tExpire, err := time.Parse("2006-01-02 15:04:05", start.GetExpire())
	if err != nil {
		logger.Error("CheckTime", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse Expire failed",
			start.GetExpire())
		return false
	}

	if tMax.After(tNow.Add(d1)) && tExpire.After(tMax.Add(d2)){
		return true
	}

	return false
}

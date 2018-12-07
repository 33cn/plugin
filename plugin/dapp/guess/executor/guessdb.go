// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"context"
	"fmt"
	"github.com/33cn/chain33/client"
	"google.golang.org/grpc"
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
	MaxBetsOneTime = 10000            //一次最多下多少注
	MaxBetsNumber = 1000000     //一局游戏最多接受多少注
	MaxBetHeight = 10000000000    //最大区块高度

	MinBetBlockNum = 720          //从创建游戏开始，一局游戏最少的可下注区块数量
	MinBetTimeInterval = "2h"     //从创建游戏开始，一局游戏最短的可下注时间
	MinBetTimeoutNum = 8640       //从游戏结束下注开始，一局游戏最少的超时块数
	MinBetTimeoutInterval = "24h" //从游戏结束下注开始，一局游戏最短的超时时间

    grpcRecSize int = 5 * 30 * 1024 * 1024

    retryNum = 10
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
	conn         *grpc.ClientConn
	grpcClient   types.Chain33Client
}

func NewAction(guess *Guess, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromAddr := tx.From()

	msgRecvOp := grpc.WithMaxMsgSize(grpcRecSize)
	conn, err := grpc.Dial(cfg.ParaRemoteGrpcClient, grpc.WithInsecure(), msgRecvOp)

	if err != nil {
		panic(err)
	}
	grpcClient := types.NewChain33Client(conn)

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
		api: guess.GetAPI(),
		conn: conn,
		grpcClient: grpcClient,
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
	//key = append(key, []byte("mavl-"+types.ExecName(pkt.GuessX)+"-")...)
	key = append(key, []byte("mavl-"+pkt.GuessX+"-")...)
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

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
}

func getGameListByAdminAddr(db dbm.Lister, addr string, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameAdminPrefix(addr), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameAdminPrefix(addr), calcGuessGameAdminKey(addr, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
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

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
}

func getGameListByAddrStatus(db dbm.Lister, addr string, status int32, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameAddrStatusPrefix(addr, status), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameAddrStatusPrefix(addr, status), calcGuessGameAddrStatusKey(addr, status, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
}

func getGameListByAdminStatus(db dbm.Lister, admin string, status int32, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameAdminStatusPrefix(admin, status), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameAdminStatusPrefix(admin, status), calcGuessGameAdminStatusKey(admin, status, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
}

func getGameListByCategoryStatus(db dbm.Lister, category string, status int32, index int64) (types.Message, error) {
	var values [][]byte
	var err error
	if index == 0 {
		values, err = db.List(calcGuessGameCategoryStatusPrefix(category, status), nil, DefaultCount, ListDESC)
	} else {
		values, err = db.List(calcGuessGameCategoryStatusPrefix(category, status), calcGuessGameCategoryStatusKey(category, status, index), DefaultCount, ListDESC)
	}
	if err != nil {
		return nil, err
	}

	var records []*pkt.GuessGameRecord
	for _, value := range values {
		var record pkt.GuessGameRecord
		err := types.Decode(value, &record)
		if err != nil {
			continue
		}
		records = append(records, &record)
	}

	return &pkt.GuessGameRecords{Records: records}, nil
}

func (action *Action) saveGame(game *pkt.GuessGame) (kvset []*types.KeyValue) {
	value := types.Encode(game)
	action.db.Set(Key(game.GetGameId()), value)
	kvset = append(kvset, &types.KeyValue{Key: Key(game.GameId), Value: value})
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
	var game pkt.GuessGame
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		return nil, err
	}
	return &game, nil
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
		MaxBetTime:     start.MaxBetTime,
		MaxBetHeight:   start.MaxBetHeight,
		Symbol:      start.Symbol,
		Exec:        start.Exec,
		MaxBetsOneTime:     start.MaxBetsOneTime,
		MaxBetsNumber: start.MaxBetsNumber,
		DevFeeFactor: start.DevFeeFactor,
		DevFeeAddr: start.DevFeeAddr,
		PlatFeeFactor: start.PlatFeeFactor,
		PlatFeeAddr: start.PlatFeeAddr,
		Expire: start.Expire,
		ExpireHeight: start.ExpireHeight,
		//AdminAddr: action.fromaddr,
		BetsNumber: 0,
		//Index:       action.getIndex(game),
		DrivenByAdmin: start.DrivenByAdmin,
	}

	return game, nil
}


func (action *Action) GameStart(start *pkt.GuessGameStart) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if start.MaxBetHeight >= MaxBetHeight {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum height number is %d which is less thanstart.MaxHeight %d", MaxBetHeight, start.MaxBetHeight))
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

	if !action.CheckTime(start) {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The height and time parameters are illegal:MaxTime %s MaxHeight %d Expire %s, ExpireHeight %d", start.MaxBetTime, start.MaxBetHeight, start.Expire, start.ExpireHeight))
		return nil, types.ErrInvalidParam
	}

	if len(start.Symbol) == 0 {
		start.Symbol = "bty"
	}

	if len(start.Exec) == 0 {
		start.Exec = "coins"
	}

	if start.MaxBetsOneTime >= MaxBetsOneTime {
		start.MaxBetsOneTime = MaxBetsOneTime
	}

	gameId := common.ToHex(action.txhash)
	game, _ := action.newGame(gameId, start)
	game.StartTime = action.blocktime
	game.AdminAddr = action.fromaddr
	game.PreIndex = 0
	game.Index = action.getIndex()
	game.Status = pkt.GuessGameStatusStart
	game.BetStat = &pkt.GuessBetStat{TotalBetTimes:0, TotalBetsNumber:0}
    for i := 0; i < len(options); i++ {
		item := &pkt.GuessBetStatItem{Option: options[i], BetsNumber: 0, BetsTimes: 0}
		game.BetStat.Items = append(game.BetStat.Items, item)
	}

	receiptLog := action.GetReceiptLog(game, false)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
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
		return nil, pkt.ErrGuessStatus
	}

	canBet := action.RefreshStatusByTime(game)

	if canBet == false {
		var receiptLog *types.ReceiptLog
		if prevStatus != game.Status {
			//状态发生了变化，且是变到了不可下注的状态，那么对于所有下注的addr来说，其addr:status主键的数据都需要更新
			action.ChangeAllAddrIndex(game)
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
	if pbBet.GetBetsNum() > game.GetMaxBetsOneTime() {
		pbBet.BetsNum = game.GetMaxBetsOneTime()
	}

	if game.BetsNumber + pbBet.GetBetsNum() > game.MaxBetsNumber {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "MaxBetsNumber over limit",
			game.MaxBetsNumber, "current Bets Number", game.BetsNumber)
		return nil, types.ErrInvalidParam
	}

	// 检查余额账户余额
	checkValue := int64(pbBet.BetsNum)
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
		//状态发生变化，更新所有addr对应记录的index
		action.ChangeAllAddrIndex(game)
		receiptLog = action.GetReceiptLog(game, true)
	} else {
		receiptLog = action.GetReceiptLog(game, false)
	}
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

func (action *Action) GameStopBet(pbBet *pkt.GuessGameStopBet) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbBet.GetGameId())
	if err != nil {
		logger.Error("GameStopBet", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbBet.GetGameId(), "err", err)
		return nil, err
	}

	if game.Status != pkt.GuessGameStatusStart && game.Status != pkt.GuessGameStatusBet{
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, pkt.ErrGuessStatus
	}

	//只有adminAddr可以发起stopBet
	if game.AdminAddr != action.fromaddr {
		logger.Error("GameStopBet", "addr", action.fromaddr, "execaddr", action.execaddr, "fromAddr is not adminAddr",
			action.fromaddr, "adminAddr", game.AdminAddr)
		return nil, types.ErrInvalidParam
	}

	action.ChangeStatus(game, pkt.GuessGameStatusStopBet)

	var receiptLog *types.ReceiptLog
	//状态发生变化，更新所有addr对应记录的index
	action.ChangeAllAddrIndex(game)
	receiptLog = action.GetReceiptLog(game, true)

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

	if game.Status != pkt.GuessGameStatusStart && game.Status != pkt.GuessGameStatusBet && game.Status != pkt.GuessGameStatusStopBet{
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, pkt.ErrGuessStatus
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

	//先遍历所有下注数据，转移资金到Admin账户合约地址；
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		value := int64(player.Bet.BetsNumber)
		receipt, err := action.coinsAccount.ExecTransfer(player.Addr, game.AdminAddr, action.execaddr, value)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, value) // rollback
			logger.Error("GamePublish", "addr", player.Addr, "execaddr", action.execaddr,
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

	//按创建游戏时设定的比例，转移佣金到开发者账户和平台账户
	devAddr := pkt.DevShareAddr
	platAddr := pkt.PlatformShareAddr
	devFee := int64(0)
	platFee := int64(0)
	if len(game.DevFeeAddr) > 0 {
		devAddr = game.DevFeeAddr
	}

	if len(game.PlatFeeAddr) > 0 {
		platAddr = game.PlatFeeAddr
	}

	if game.DevFeeFactor > 0 {
		devFee = int64(totalBetsNumber) * game.DevFeeFactor / 1000
		receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, devAddr, action.execaddr, devFee)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, devFee) // rollback
			logger.Error("GamePublish", "adminAddr", game.AdminAddr, "execaddr", action.execaddr,
				"amount", devFee, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if game.PlatFeeFactor > 0 {
		platFee = int64(totalBetsNumber) * game.PlatFeeFactor / 1000
		receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, platAddr, action.execaddr, platFee)
		if err != nil {
			action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, platFee) // rollback
			logger.Error("GamePublish", "adminAddr", game.AdminAddr, "execaddr", action.execaddr,
				"amount", platFee, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	//再遍历赢家，按照投注占比分配所有筹码
	winValue := int64(totalBetsNumber) - devFee - platFee
	for j := 0; j < len(game.Plays); j++ {
		player := game.Plays[j]
		if player.Bet.Option == game.Result {
			value := int64(player.Bet.BetsNumber * uint32(winValue) / winBetsNumber)
			receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, player.Addr, action.execaddr, value)
			if err != nil {
				action.coinsAccount.ExecFrozen(player.Addr, action.execaddr, value) // rollback
				logger.Error("GamePublish", "addr", player.Addr, "execaddr", action.execaddr,
					"amount", value, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
			player.Bet.IsWinner = true
			player.Bet.Profit = value
		}
	}

	var receiptLog *types.ReceiptLog
    action.ChangeAllAddrIndex(game)
	receiptLog = action.GetReceiptLog(game, true)

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
		return nil, pkt.ErrGuessStatus
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
		value := int64(player.Bet.BetsNumber)
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

	//状态发生变化，统一更新所有addr记录的index
	action.ChangeAllAddrIndex(game)

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

func (action *Action) ChangeStatus(game *pkt.GuessGame, destStatus int32) {
	if game.Status != destStatus {
		game.PreStatus = game.Status
		game.PreIndex = game.Index
		game.Status = destStatus
		game.Index = action.getIndex()
	}

	return
}

func (action *Action) ChangeAllAddrIndex(game *pkt.GuessGame) {
	for i := 0; i < len(game.Plays) ; i++ {
		player := game.Plays[i]
		player.Bet.PreIndex = player.Bet.Index
		player.Bet.Index = game.Index
	}
}

func (action *Action) RefreshStatusByTime(game *pkt.GuessGame) (canBet bool) {
	//如果完全由管理员驱动状态变化，则不需要做如下判断保护。
	if game.DrivenByAdmin {
		return true
	}

	// 检查区块高度是否超过最大下注高度限制，看是否可以下注
	if game.GetMaxBetHeight() <= action.height {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Height over limit",
			action.height, "MaxHeight", game.GetMaxBetHeight())
		if game.GetExpireHeight() > action.height {
			action.ChangeStatus(game, pkt.GuessGameStatusStopBet)
		} else {
			action.ChangeStatus(game, pkt.GuessGameStatusTimeOut)
		}

		canBet = false
        return  canBet
	}

	// 检查区块高度是否超过下注时间限制，看是否可以下注
	if len(game.GetMaxBetTime()) > 0 {
		tMax, err := time.Parse("2006-01-02 15:04:05", game.GetMaxBetTime())
		if err != nil {
			logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse MaxTime failed",
				game.GetMaxBetTime())
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
				game.GetMaxBetTime())

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
	if len(start.MaxBetTime) == 0 && len(start.Expire) == 0 && start.MaxBetHeight == 0 && start.ExpireHeight == 0 {
		//如果上述字段都不携带，则认为完全由admin的指令驱动。
		start.DrivenByAdmin = true
		return true
	}

	if action.height + MinBetBlockNum > start.MaxBetHeight || start.MaxBetHeight + MinBetTimeoutNum > start.ExpireHeight {
		return false
	}

	tNow := time.Now()
	d1, _ := time.ParseDuration(MinBetTimeInterval)      //最短开奖时间
	d2, _ := time.ParseDuration(MinBetTimeoutInterval)   //最短游戏过期时间
	if len(start.GetMaxBetTime()) == 0 {
		tNow.Add(d1)
		start.MaxBetTime = tNow.Format("2006-01-02 15:04:05")
	}

	if len(start.GetExpire()) == 0 {
		tMax, _ := time.Parse("2006-01-02 15:04:05", start.GetMaxBetTime())
		tMax.Add(d2)
		start.Expire = tMax.Format("2006-01-02 15:04:05")
	}

	tMax, err := time.Parse("2006-01-02 15:04:05", start.GetMaxBetTime())
	if err != nil {
		logger.Error("CheckTime", "addr", action.fromaddr, "execaddr", action.execaddr, "Parse MaxTime failed",
			start.GetMaxBetTime())
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

// GetMainHeightByTxHash get Block height
func (action *Action) GetMainHeightByTxHash(txHash []byte) int64 {
	for i := 0; i < retryNum; i++ {
		req := &types.ReqHash{Hash: txHash}
		txDetail, err := action.grpcClient.QueryTransaction(context.Background(), req)
		if err != nil {
			time.Sleep(time.Second)
		} else {
			return txDetail.GetHeight()
		}
	}

	return -1
}

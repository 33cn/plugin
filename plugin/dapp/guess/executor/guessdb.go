// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	gty "github.com/33cn/plugin/plugin/dapp/guess/types"
)

const (
	//ListDESC 表示记录降序排列
	ListDESC = int32(0)

	//ListASC 表示记录升序排列
	ListASC = int32(1)

	//DefaultCount 默认一次获取的记录数
	DefaultCount = int32(10)

	//DefaultCategory 默认分类
	DefaultCategory = "default"

	//MaxBetsOneTime 一次最多下多少注
	MaxBetsOneTime = 10000e8

	//MaxBetsNumber 一局游戏最多接受多少注
	MaxBetsNumber = 10000000e8

	//MaxBetHeight 距离游戏创建区块的最大可下注高度差
	MaxBetHeight = 1000000

	//MaxExpireHeight 距离游戏创建区块的最大过期高度差
	MaxExpireHeight = 1000000
)

//Action 具体动作执行
type Action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
	localDB      dbm.KVDB
	index        int
	mainHeight   int64
}

//NewAction 生成Action对象
func NewAction(guess *Guess, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromAddr := tx.From()

	return &Action{
		coinsAccount: guess.GetCoinsAccount(),
		db:           guess.GetStateDB(),
		txhash:       hash,
		fromaddr:     fromAddr,
		blocktime:    guess.GetBlockTime(),
		height:       guess.GetHeight(),
		execaddr:     dapp.ExecAddress(string(tx.Execer)),
		localDB:      guess.GetLocalDB(),
		index:        index,
		mainHeight:   guess.GetMainHeight(),
	}
}

//CheckExecAccountBalance 检查地址在Guess合约中的余额是否足够
func (action *Action) CheckExecAccountBalance(fromAddr string, ToFrozen, ToActive int64) bool {
	acc := action.coinsAccount.LoadExecAccount(fromAddr, action.execaddr)
	if acc.GetBalance() >= ToFrozen && acc.GetFrozen() >= ToActive {
		return true
	}
	return false
}

//Key State数据库中存储记录的Key值格式转换
func Key(id string) (key []byte) {
	//key = append(key, []byte("mavl-"+types.ExecName(pkt.GuessX)+"-")...)
	key = append(key, []byte("mavl-"+gty.GuessX+"-")...)
	key = append(key, []byte(id)...)
	return key
}

//queryGameInfos 根据游戏id列表查询多个游戏详情信息
func queryGameInfos(kvdb db.KVDB, infos *gty.QueryGuessGameInfos) (types.Message, error) {
	var games []*gty.GuessGame
	gameTable := gty.NewGuessGameTable(kvdb)
	query := gameTable.GetQuery(kvdb)

	for i := 0; i < len(infos.GameIDs); i++ {
		rows, err := query.ListIndex("gameid", []byte(infos.GameIDs[i]), nil, 1, 0)
		if err != nil {
			return nil, err
		}

		game := rows[0].Data.(*gty.GuessGame)
		games = append(games, game)
	}
	return &gty.ReplyGuessGameInfos{Games: games}, nil
}

//queryGameInfo 根据gameid查询game信息
func queryGameInfo(kvdb db.KVDB, gameID []byte) (*gty.GuessGame, error) {
	gameTable := gty.NewGuessGameTable(kvdb)
	query := gameTable.GetQuery(kvdb)
	rows, err := query.ListIndex("gameid", gameID, nil, 1, 0)
	if err != nil {
		return nil, err
	}

	game := rows[0].Data.(*gty.GuessGame)

	return game, nil
}

//queryUserTableData 查询user表数据
func queryUserTableData(query *table.Query, indexName string, prefix, primaryKey []byte) (types.Message, error) {
	rows, err := query.ListIndex(indexName, prefix, primaryKey, DefaultCount, 0)
	if err != nil {
		return nil, err
	}

	var records []*gty.GuessGameRecord

	for i := 0; i < len(rows); i++ {
		userBet := rows[i].Data.(*gty.UserBet)
		var record gty.GuessGameRecord
		record.GameID = userBet.GameID
		record.StartIndex = userBet.StartIndex
		records = append(records, &record)
	}

	var primary string
	if len(rows) == int(DefaultCount) {
		primary = string(rows[len(rows)-1].Primary)
	}

	return &gty.GuessGameRecords{Records: records, PrimaryKey: primary}, nil
}

//queryGameTableData 查询game表数据
func queryGameTableData(query *table.Query, indexName string, prefix, primaryKey []byte) (types.Message, error) {
	rows, err := query.ListIndex(indexName, prefix, primaryKey, DefaultCount, 0)
	if err != nil {
		return nil, err
	}

	var records []*gty.GuessGameRecord

	for i := 0; i < len(rows); i++ {
		game := rows[i].Data.(*gty.GuessGame)
		var record gty.GuessGameRecord
		record.GameID = game.GameID
		record.StartIndex = game.StartIndex
		records = append(records, &record)
	}

	var primary string
	if len(rows) == int(DefaultCount) {
		primary = string(rows[len(rows)-1].Primary)
	}

	return &gty.GuessGameRecords{Records: records, PrimaryKey: primary}, nil
}

//queryJoinTableData 查询join表数据
func queryJoinTableData(talbeJoin *table.JoinTable, indexName string, prefix, primaryKey []byte) (types.Message, error) {
	rows, err := talbeJoin.ListIndex(indexName, prefix, primaryKey, DefaultCount, 0)
	if err != nil {
		return nil, err
	}

	var records []*gty.GuessGameRecord

	for i := 0; i < len(rows); i++ {
		game := rows[i].Data.(*table.JoinData).Right.(*gty.GuessGame)
		var record gty.GuessGameRecord
		record.GameID = game.GameID
		record.StartIndex = game.StartIndex
		records = append(records, &record)
	}

	var primary string
	if len(rows) == int(DefaultCount) {
		primary = fmt.Sprintf("%018d", rows[len(rows)-1].Data.(*table.JoinData).Left.(*gty.UserBet).Index)
	}

	return &gty.GuessGameRecords{Records: records, PrimaryKey: primary}, nil
}

func (action *Action) saveGame(game *gty.GuessGame) (kvset []*types.KeyValue) {
	value := types.Encode(game)
	err := action.db.Set(Key(game.GetGameID()), value)
	if err != nil {
		logger.Error("saveGame have err:", err.Error())
	}
	kvset = append(kvset, &types.KeyValue{Key: Key(game.GameID), Value: value})
	return kvset
}

func (action *Action) getIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

//getReceiptLog 根据游戏信息生成收据记录
func (action *Action) getReceiptLog(game *gty.GuessGame, statusChange bool, bet *gty.GuessGameBet) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &gty.ReceiptGuessGame{}
	r.Addr = action.fromaddr
	if game.Status == gty.GuessGameStatusStart {
		log.Ty = gty.TyLogGuessGameStart
	} else if game.Status == gty.GuessGameStatusBet {
		log.Ty = gty.TyLogGuessGameBet
	} else if game.Status == gty.GuessGameStatusStopBet {
		log.Ty = gty.TyLogGuessGameStopBet
	} else if game.Status == gty.GuessGameStatusAbort {
		log.Ty = gty.TyLogGuessGameAbort
	} else if game.Status == gty.GuessGameStatusPublish {
		log.Ty = gty.TyLogGuessGamePublish
	} else if game.Status == gty.GuessGameStatusTimeOut {
		log.Ty = gty.TyLogGuessGameTimeout
	}

	r.StartIndex = game.StartIndex
	r.Index = action.getIndex()
	r.GameID = game.GameID
	r.Status = game.Status
	r.AdminAddr = game.AdminAddr
	r.PreStatus = game.PreStatus
	r.StatusChange = statusChange
	r.PreIndex = game.PreIndex
	r.Category = game.Category
	if nil != bet {
		r.Bet = true
		r.Option = bet.Option
		r.BetsNumber = bet.BetsNum
	}
	r.Game = game
	log.Log = types.Encode(r)
	return log
}

func (action *Action) readGame(id string) (*gty.GuessGame, error) {
	data, err := action.db.Get(Key(id))
	if err != nil {
		logger.Error("readGame have err:", err.Error())
		return nil, err
	}
	var game gty.GuessGame
	//decode
	err = types.Decode(data, &game)
	if err != nil {
		logger.Error("decode game have err:", err.Error())
		return nil, err
	}
	return &game, nil
}

// 新建一局游戏
func (action *Action) newGame(gameID string, start *gty.GuessGameStart) *gty.GuessGame {
	game := &gty.GuessGame{
		GameID: gameID,
		Status: gty.GuessGameStatusStart,
		//StartTime:   action.blocktime,
		StartTxHash:    gameID,
		Topic:          start.Topic,
		Category:       start.Category,
		Options:        start.Options,
		MaxBetHeight:   start.MaxBetHeight,
		MaxBetsOneTime: start.MaxBetsOneTime,
		MaxBetsNumber:  start.MaxBetsNumber,
		DevFeeFactor:   start.DevFeeFactor,
		DevFeeAddr:     start.DevFeeAddr,
		PlatFeeFactor:  start.PlatFeeFactor,
		PlatFeeAddr:    start.PlatFeeAddr,
		ExpireHeight:   start.ExpireHeight,
		//AdminAddr: action.fromaddr,
		BetsNumber: 0,
		//Index:       action.getIndex(game),
		DrivenByAdmin: start.DrivenByAdmin,
	}

	return game
}

//GameStart 创建游戏动作执行
func (action *Action) GameStart(start *gty.GuessGameStart) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	if start.MaxBetHeight >= MaxBetHeight {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum height diff number is %d which is less than start.MaxBetHeight %d", MaxBetHeight, start.MaxBetHeight))
		return nil, types.ErrInvalidParam
	}

	if start.ExpireHeight >= MaxExpireHeight {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum height diff number is %d which is less than start.MaxBetHeight %d", MaxBetHeight, start.MaxBetHeight))
		return nil, types.ErrInvalidParam
	}

	if start.MaxBetsNumber >= MaxBetsNumber {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The maximum bets number is %d which is less than start.MaxBetsNumber %d", int64(MaxBetsNumber), start.MaxBetsNumber))
		return nil, gty.ErrOverBetsLimit
	}

	if len(start.Topic) == 0 || len(start.Options) == 0 {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("Illegal parameters,Topic:%s | options: %s | category: %s", start.Topic, start.Options, start.Category))
		return nil, types.ErrInvalidParam
	}

	options, ok := getOptions(start.Options)
	if !ok {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The options is illegal:%s", start.Options))
		return nil, types.ErrInvalidParam
	}

	if !action.checkTime(start) {
		logger.Error("GameStart", "addr", action.fromaddr, "execaddr", action.execaddr,
			"err", fmt.Sprintf("The height and time parameters are illegal:MaxHeight %d ,ExpireHeight %d", start.MaxBetHeight, start.ExpireHeight))
		return nil, types.ErrInvalidParam
	}

	if len(start.Category) == 0 {
		start.Category = DefaultCategory
	}

	if start.MaxBetsOneTime >= MaxBetsOneTime {
		start.MaxBetsOneTime = MaxBetsOneTime
	}

	gameID := common.ToHex(action.txhash)
	game := action.newGame(gameID, start)
	game.StartTime = action.blocktime
	game.StartHeight = action.mainHeight
	game.AdminAddr = action.fromaddr
	game.PreIndex = 0
	game.Index = action.getIndex()
	game.StartIndex = game.Index
	game.Status = gty.GuessGameStatusStart
	game.BetStat = &gty.GuessBetStat{TotalBetTimes: 0, TotalBetsNumber: 0}
	for i := 0; i < len(options); i++ {
		item := &gty.GuessBetStatItem{Option: options[i], BetsNumber: 0, BetsTimes: 0}
		game.BetStat.Items = append(game.BetStat.Items, item)
	}

	receiptLog := action.getReceiptLog(game, false, nil)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//GameBet 参与游戏动作执行
func (action *Action) GameBet(pbBet *gty.GuessGameBet) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbBet.GetGameID())
	if err != nil || game == nil {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbBet.GetGameID(), "err", err)
		return nil, err
	}

	prevStatus := game.Status
	if game.Status != gty.GuessGameStatusStart && game.Status != gty.GuessGameStatusBet {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, gty.ErrGuessStatus
	}

	canBet := action.refreshStatusByTime(game)

	if !canBet {
		var receiptLog *types.ReceiptLog
		if prevStatus != game.Status {
			//状态发生了变化，且是变到了不可下注的状态，那么对于所有下注的addr来说，其addr:status主键的数据都需要更新
			action.changeAllAddrIndex(game)
			receiptLog = action.getReceiptLog(game, true, nil)
		} else {
			receiptLog = action.getReceiptLog(game, false, nil)
		}

		logs = append(logs, receiptLog)
		kv = append(kv, action.saveGame(game)...)

		return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
	}

	//检查竞猜选项是否合法
	options, legal := getOptions(game.GetOptions())
	if !legal || len(options) == 0 {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Game Options illegal",
			game.GetOptions())
		return nil, types.ErrInvalidParam
	}

	if !isLegalOption(options, pbBet.GetOption()) {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Option illegal",
			pbBet.GetOption())
		return nil, types.ErrInvalidParam
	}

	//检查下注金额是否超限，如果超限，按最大值
	if pbBet.GetBetsNum() > game.GetMaxBetsOneTime() {
		pbBet.BetsNum = game.GetMaxBetsOneTime()
	}

	if game.BetsNumber+pbBet.GetBetsNum() > game.MaxBetsNumber {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "MaxBetsNumber over limit",
			game.MaxBetsNumber, "current Bets Number", game.BetsNumber)
		return nil, types.ErrInvalidParam
	}

	// 检查账户余额
	checkValue := pbBet.BetsNum
	if !action.CheckExecAccountBalance(action.fromaddr, checkValue, 0) {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "id",
			pbBet.GetGameID(), "err", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, checkValue)
	if err != nil {
		logger.Error("GameCreate.ExecFrozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", checkValue, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	var receiptLog *types.ReceiptLog
	if prevStatus != gty.GuessGameStatusBet {
		action.changeStatus(game, gty.GuessGameStatusBet)
		action.addGuessBet(game, pbBet)
		receiptLog = action.getReceiptLog(game, true, pbBet)
	} else {
		action.addGuessBet(game, pbBet)
		receiptLog = action.getReceiptLog(game, false, pbBet)
	}

	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//GameStopBet 停止游戏下注动作执行
func (action *Action) GameStopBet(pbBet *gty.GuessGameStopBet) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbBet.GetGameID())
	if err != nil || game == nil {
		logger.Error("GameStopBet", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbBet.GetGameID(), "err", err)
		return nil, err
	}

	if game.Status != gty.GuessGameStatusStart && game.Status != gty.GuessGameStatusBet {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, gty.ErrGuessStatus
	}

	//只有adminAddr可以发起stopBet
	if game.AdminAddr != action.fromaddr {
		logger.Error("GameStopBet", "addr", action.fromaddr, "execaddr", action.execaddr, "fromAddr is not adminAddr",
			action.fromaddr, "adminAddr", game.AdminAddr)
		return nil, gty.ErrNoPrivilege
	}

	action.changeStatus(game, gty.GuessGameStatusStopBet)

	var receiptLog *types.ReceiptLog
	//状态发生变化，更新所有addr对应记录的index
	action.changeAllAddrIndex(game)
	receiptLog = action.getReceiptLog(game, true, nil)

	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//addGuessBet 向游戏结构中加入下注信息
func (action *Action) addGuessBet(game *gty.GuessGame, pbBet *gty.GuessGameBet) {
	bet := &gty.GuessBet{Option: pbBet.GetOption(), BetsNumber: pbBet.BetsNum, Index: action.getIndex()}
	player := &gty.GuessPlayer{Addr: action.fromaddr, Bet: bet}
	game.Plays = append(game.Plays, player)

	for i := 0; i < len(game.BetStat.Items); i++ {
		if game.BetStat.Items[i].Option == trimStr(pbBet.GetOption()) {
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

//GamePublish 公布竞猜游戏结果动作执行
func (action *Action) GamePublish(publish *gty.GuessGamePublish) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(publish.GetGameID())
	if err != nil || game == nil {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			publish.GetGameID(), "err", err)
		return nil, err
	}

	//只有adminAddr可以发起publish
	if game.AdminAddr != action.fromaddr {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "fromAddr is not adminAddr",
			action.fromaddr, "adminAddr", game.AdminAddr)
		return nil, gty.ErrNoPrivilege
	}

	if game.Status != gty.GuessGameStatusStart && game.Status != gty.GuessGameStatusBet && game.Status != gty.GuessGameStatusStopBet {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "Status error",
			game.GetStatus())
		return nil, gty.ErrGuessStatus
	}

	//检查竞猜选项是否合法
	options, legal := getOptions(game.GetOptions())
	if !legal || len(options) == 0 {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "Game Options illegal",
			game.GetOptions())
		return nil, types.ErrInvalidParam
	}

	if !isLegalOption(options, publish.GetResult()) {
		logger.Error("GamePublish", "addr", action.fromaddr, "execaddr", action.execaddr, "Option illegal",
			publish.GetResult())
		return nil, types.ErrInvalidParam
	}

	game.Result = trimStr(publish.Result)

	//先遍历所有下注数据，转移资金到Admin账户合约地址；
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		value := player.Bet.BetsNumber
		receipt, err := action.coinsAccount.ExecActive(player.Addr, action.execaddr, value)
		if err != nil {
			logger.Error("GamePublish.ExecActive", "addr", player.Addr, "execaddr", action.execaddr, "amount", value,
				"err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)

		receipt, err = action.coinsAccount.ExecTransfer(player.Addr, game.AdminAddr, action.execaddr, value)
		if err != nil {
			//action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, value) // rollback
			logger.Error("GamePublish", "addr", player.Addr, "execaddr", action.execaddr,
				"amount", value, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	action.changeStatus(game, gty.GuessGameStatusPublish)
	//计算竞猜正确的筹码总数
	totalBetsNumber := game.BetStat.TotalBetsNumber
	winBetsNumber := int64(0)
	for j := 0; j < len(game.BetStat.Items); j++ {
		if game.BetStat.Items[j].Option == game.Result {
			winBetsNumber = game.BetStat.Items[j].BetsNumber
		}
	}

	//按创建游戏时设定的比例，转移佣金到开发者账户和平台账户
	devAddr := gty.DevShareAddr
	platAddr := gty.PlatformShareAddr
	devFee := int64(0)
	platFee := int64(0)
	if len(game.DevFeeAddr) > 0 {
		devAddr = game.DevFeeAddr
	}

	if len(game.PlatFeeAddr) > 0 {
		platAddr = game.PlatFeeAddr
	}

	if game.DevFeeFactor > 0 {
		fee := big.NewInt(totalBetsNumber)
		factor := big.NewInt(game.DevFeeFactor)
		thousand := big.NewInt(1000)
		devFee = fee.Mul(fee, factor).Div(fee, thousand).Int64()
		receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, devAddr, action.execaddr, devFee)
		if err != nil {
			//action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, devFee) // rollback
			logger.Error("GamePublish", "adminAddr", game.AdminAddr, "execaddr", action.execaddr,
				"amount", devFee, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if game.PlatFeeFactor > 0 {
		fee := big.NewInt(totalBetsNumber)
		factor := big.NewInt(game.PlatFeeFactor)
		thousand := big.NewInt(1000)
		platFee = fee.Mul(fee, factor).Div(fee, thousand).Int64()
		receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, platAddr, action.execaddr, platFee)
		if err != nil {
			//action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, platFee) // rollback
			logger.Error("GamePublish", "adminAddr", game.AdminAddr, "execaddr", action.execaddr,
				"amount", platFee, "err", err)
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	//再遍历赢家，按照投注占比分配所有筹码
	winValue := totalBetsNumber - devFee - platFee
	for j := 0; j < len(game.Plays); j++ {
		player := game.Plays[j]
		if trimStr(player.Bet.Option) == trimStr(game.Result) {
			betsNumber := big.NewInt(player.Bet.BetsNumber)
			totalWinBetsNumber := big.NewInt(winBetsNumber)
			leftWinBetsNumber := big.NewInt(winValue)

			value := betsNumber.Mul(betsNumber, leftWinBetsNumber).Div(betsNumber, totalWinBetsNumber).Int64()
			receipt, err := action.coinsAccount.ExecTransfer(game.AdminAddr, player.Addr, action.execaddr, value)
			if err != nil {
				//action.coinsAccount.ExecFrozen(player.Addr, action.execaddr, value) // rollback
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
	action.changeAllAddrIndex(game)
	receiptLog = action.getReceiptLog(game, true, nil)

	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//GameAbort 撤销游戏动作执行
func (action *Action) GameAbort(pbend *gty.GuessGameAbort) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	game, err := action.readGame(pbend.GetGameID())
	if err != nil || game == nil {
		logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "get game failed",
			pbend.GetGameID(), "err", err)
		return nil, err
	}

	if game.Status == gty.GuessGameStatusPublish || game.Status == gty.GuessGameStatusAbort {

		logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "game status not allow abort",
			game.Status)
		return nil, gty.ErrGuessStatus
	}

	preStatus := game.Status
	//根据区块链高度或时间刷新游戏状态。
	action.refreshStatusByTime(game)

	//如果游戏超时，则任何地址都可以Abort，否则只有创建游戏的地址可以Abort
	if game.Status != gty.GuessGameStatusTimeOut {
		if game.AdminAddr != action.fromaddr {
			logger.Error("GameAbort", "addr", action.fromaddr, "execaddr", action.execaddr, "Only admin can abort",
				action.fromaddr, "status", game.Status)
			return nil, err
		}
	}

	//激活冻结账户
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		value := player.Bet.BetsNumber
		receipt, err := action.coinsAccount.ExecActive(player.Addr, action.execaddr, value)
		if err != nil {
			logger.Error("GameAbort", "addr", player.Addr, "execaddr", action.execaddr, "amount", value, "err", err)
			continue
		}

		player.Bet.IsWinner = true
		player.Bet.Profit = value

		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	if game.Status != preStatus {
		//说明action.RefreshStatusByTime(game)调用时已经更新过状态和index了，这里直接再改状态就行了。
		game.Status = gty.GuessGameStatusAbort
	} else {
		action.changeStatus(game, gty.GuessGameStatusAbort)
	}

	//状态发生变化，统一更新所有addr记录的index
	action.changeAllAddrIndex(game)

	receiptLog := action.getReceiptLog(game, true, nil)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveGame(game)...)
	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//getOptions 获得竞猜选项，并判断是否符合约定格式，类似"A:xxxx;B:xxxx;C:xxx"，“：”前为选项名称，不能重复，":"后为选项说明。
func getOptions(strOptions string) (options []string, legal bool) {
	if len(strOptions) == 0 {
		return nil, false
	}

	legal = true
	items := strings.Split(strOptions, ";")
	for i := 0; i < len(items); i++ {
		item := strings.Split(items[i], ":")
		for j := 0; j < len(options); j++ {
			if item[0] == options[j] {
				legal = false
				return
			}
		}

		options = append(options, trimStr(item[0]))
	}

	return options, legal
}

//trimStr 去除字符串中的空格、制表符、换行符
func trimStr(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	return str
}

//isLegalOption 判断选项是否为合法选项
func isLegalOption(options []string, option string) bool {
	option = trimStr(option)
	for i := 0; i < len(options); i++ {
		if options[i] == option {
			return true
		}
	}

	return false
}

//changeStatus 修改游戏状态，同步更新历史记录
func (action *Action) changeStatus(game *gty.GuessGame, destStatus int32) {
	if game.Status != destStatus {
		game.PreStatus = game.Status
		game.PreIndex = game.Index
		game.Status = destStatus
		game.Index = action.getIndex()
	}
}

//changeAllAddrIndex 状态更新时，更新下注记录的历史信息
func (action *Action) changeAllAddrIndex(game *gty.GuessGame) {
	for i := 0; i < len(game.Plays); i++ {
		player := game.Plays[i]
		player.Bet.PreIndex = player.Bet.Index
		player.Bet.Index = action.getIndex()
	}
}

//refreshStatusByTime 检测游戏是否过期，是否可以下注
func (action *Action) refreshStatusByTime(game *gty.GuessGame) (canBet bool) {
	mainHeight := action.mainHeight
	//如果完全由管理员驱动状态变化，则除了保护性过期判断外，不需要做其他判断。
	if game.DrivenByAdmin {

		if (mainHeight - game.StartHeight) >= game.ExpireHeight {
			action.changeStatus(game, gty.GuessGameStatusTimeOut)
			canBet = false
			return canBet
		}

		return true
	}

	// 检查区块高度是否超过最大可下注高度限制，看是否可以下注
	heightDiff := mainHeight - game.StartHeight
	if heightDiff >= game.MaxBetHeight {
		logger.Error("GameBet", "addr", action.fromaddr, "execaddr", action.execaddr, "Height over limit",
			mainHeight, "startHeight", game.StartHeight, "MaxHeightDiff", game.GetMaxBetHeight())
		if game.ExpireHeight > heightDiff {
			action.changeStatus(game, gty.GuessGameStatusStopBet)
		} else {
			action.changeStatus(game, gty.GuessGameStatusTimeOut)
		}

		canBet = false
		return canBet
	}

	canBet = true
	return canBet
}

//checkTime 检测游戏的过期设置。
func (action *Action) checkTime(start *gty.GuessGameStart) bool {
	if start.MaxBetHeight == 0 && start.ExpireHeight == 0 {
		//如果上述字段都不携带，则认为完全由admin的指令驱动。
		start.DrivenByAdmin = true

		//依然设定最大过期高度差，作为最后的保护
		start.ExpireHeight = MaxExpireHeight
		return true
	}

	if start.MaxBetHeight == 0 {
		start.MaxBetHeight = MaxBetHeight
	}

	if start.ExpireHeight == 0 {
		start.ExpireHeight = MaxExpireHeight
	}

	if start.MaxBetHeight <= start.ExpireHeight {
		return true
	}

	return false
}

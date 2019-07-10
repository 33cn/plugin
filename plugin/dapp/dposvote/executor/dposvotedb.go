// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	"sort"
	"strings"
)

const (
	//ListDESC 表示记录降序排列
	ListDESC = int32(0)

	//ListASC 表示记录升序排列
	ListASC = int32(1)

	//DefaultCount 默认一次获取的记录数
	DefaultCount = int32(10)
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
func NewAction(dpos *DPos, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromAddr := tx.From()

	return &Action{
		coinsAccount: dpos.GetCoinsAccount(),
		db:           dpos.GetStateDB(),
		txhash:       hash,
		fromaddr:     fromAddr,
		blocktime:    dpos.GetBlockTime(),
		height:       dpos.GetHeight(),
		execaddr:     dapp.ExecAddress(string(tx.Execer)),
		localDB:      dpos.GetLocalDB(),
		index:        index,
		mainHeight:   dpos.GetMainHeight(),
	}
}

//CheckExecAccountBalance 检查地址在Dpos合约中的余额是否足够
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
	key = append(key, []byte("mavl-"+dty.DPosX+"-")...)
	key = append(key, []byte(id)...)
	return key
}
//queryVrfByTime 根据时间信息，查询TopN的受托节点的VRF信息
func queryVrfByTime(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByTime{
		return nil, types.ErrInvalidParam
	}

	cycleInfo := calcCycleByTime(req.Timestamp)
	req.Ty = dty.QueryVrfByTime
	req.Cycle = cycleInfo.cycle
	return queryVrfByCycle(kvdb, req)
}
//queryVrfByCycle 根据Cycle信息，查询TopN的受托节点的VRF信息
func queryVrfByCycle(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByCycle {
		return nil, types.ErrInvalidParam
	}

	topNReq := &dty.CandidatorQuery{
		TopN: int32(dposDelegateNum),
	}

	reply, err := queryTopNCands(kvdb, topNReq)
	res := reply.(*dty.CandidatorReply)
	if err != nil || len(res.Candidators) < int(dposDelegateNum){
		logger.Error("queryVrf failed", "Candidators", len(res.Candidators), "need Candidators", dposDelegateNum)
		return nil, dty.ErrCandidatorNotEnough
	}

	VrfRPTable := dty.NewDposVrfRPTable(kvdb)
	query := VrfRPTable.GetQuery(kvdb)

	var tempCands [] *dty.Candidator
	var vrfs [] *dty.VrfInfo
	for i := 0; i < len(res.Candidators); i ++ {
		rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", res.Candidators[i].Pubkey, req.Cycle)), nil, 1, 0)
		if err != nil {
			logger.Error("queryVrf RP failed", "pubkey", fmt.Sprintf("%X", res.Candidators[i].Pubkey), "cycle", req.Cycle)
			tempCands = append(tempCands, res.Candidators[i])
			continue
		}

		vrfRP := rows[0].Data.(*dty.DposVrfRP)
		vrf := &dty.VrfInfo{
			Index: vrfRP.Index,
			Pubkey: vrfRP.Pubkey,
			Cycle: vrfRP.Cycle,
			Height: vrfRP.Height,
			M: vrfRP.M,
			R: vrfRP.R,
			P: vrfRP.P,
			Time: vrfRP.Time,
		}
		vrfs = append(vrfs, vrf)
	}

	if tempCands == nil || len(tempCands) == 0 {
		return &dty.DposVrfReply{Vrf: vrfs}, nil
	}

	vrfMTable := dty.NewDposVrfMTable(kvdb)
	query = vrfMTable.GetQuery(kvdb)
	for i := 0; i < len(tempCands); i++ {
		rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", tempCands[i].Pubkey, req.Cycle)), nil, 1, 0)
		if err != nil {
			logger.Error("queryVrf M failed", "pubkey", fmt.Sprintf("%X", res.Candidators[i].Pubkey), "cycle", req.Cycle)
			continue
		}

		vrfM := rows[0].Data.(*dty.DposVrfM)
		vrf := &dty.VrfInfo{
			Index: vrfM.Index,
			Pubkey: vrfM.Pubkey,
			Cycle: vrfM.Cycle,
			Height: vrfM.Height,
			M: vrfM.M,
			Time: vrfM.Time,
		}
		vrfs = append(vrfs, vrf)
	}

	return &dty.DposVrfReply{Vrf: vrfs}, nil
}

//queryCands 根据候选节点的Pubkey下旬候选节点信息,得票数、状态等
func queryCands(kvdb db.KVDB, req *dty.CandidatorQuery) (types.Message, error) {
	var cands []*dty.Candidator
	candTable := dty.NewDposCandidatorTable(kvdb)
	query := candTable.GetQuery(kvdb)

	for i := 0; i < len(req.Pubkeys); i++ {
		bPubkey, _ := hex.DecodeString(req.Pubkeys[i])
		rows, err := query.ListIndex("pubkey", bPubkey, nil, 1, 0)
		if err != nil {
			continue
		}

		candInfo := rows[0].Data.(*dty.CandidatorInfo)
		cand := &dty.Candidator{
			Pubkey: candInfo.Pubkey,
			Address: candInfo.Address,
			Ip: candInfo.Ip,
			Votes: candInfo.Votes,
			Status: candInfo.Status,
		}
		cands = append(cands, cand)
	}
	return &dty.CandidatorReply{Candidators: cands}, nil
}

//queryTopNCands 查询得票数TopN的候选节点信息，包括得票数，状态等
func queryTopNCands(kvdb db.KVDB, req *dty.CandidatorQuery) (types.Message, error) {
	var cands []*dty.Candidator
	candTable := dty.NewDposCandidatorTable(kvdb)
	query := candTable.GetQuery(kvdb)

	number := int32(0)
	rows, err := query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusVoted)), nil, 0, 0)
	if err == nil {
		for index := 0; index < len(rows); index++ {
			candInfo := rows[index].Data.(*dty.CandidatorInfo)
			cand := &dty.Candidator{
				Pubkey:  candInfo.Pubkey,
				Address: candInfo.Address,
				Ip:      candInfo.Ip,
				Votes:   candInfo.Votes,
				Status:  candInfo.Status,
			}
			cands = append(cands, cand)
			number ++
		}

		sort.Slice(cands, func(i, j int) bool {
			return cands[i].Votes > cands[j].Votes
		})
	}

	if number < req.TopN {
		rows, err = query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusRegist)), nil, req.TopN - number, 0)
		if err == nil {
			for index := 0; index < len(rows); index++ {
				candInfo := rows[index].Data.(*dty.CandidatorInfo)
				cand := &dty.Candidator{
					Pubkey:  candInfo.Pubkey,
					Address: candInfo.Address,
					Ip:      candInfo.Ip,
					Votes:   candInfo.Votes,
					Status:  candInfo.Status,
				}
				cands = append(cands, cand)
				number ++
				if number == req.TopN {
					break
				}
			}
		}

		rows, err = query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusReRegist)), nil, req.TopN - number, 0)
		if err == nil {
			for index := 0; index < len(rows); index++ {
				candInfo := rows[index].Data.(*dty.CandidatorInfo)
				cand := &dty.Candidator{
					Pubkey:  candInfo.Pubkey,
					Address: candInfo.Address,
					Ip:      candInfo.Ip,
					Votes:   candInfo.Votes,
					Status:  candInfo.Status,
				}
				cands = append(cands, cand)
				number ++
				if number == req.TopN {
					break
				}
			}
		}

	} else {
		cands = cands[0:req.TopN]
	}

	return &dty.CandidatorReply{Candidators: cands}, nil
}

//isValidPubkey 判断一个公钥是否属于一个公钥集合
func isValidPubkey(pubkeys []string, pubkey string) bool{
	if len(pubkeys) == 0 || len(pubkey) == 0 {
		return false
	}

	for i := 0 ; i < len(pubkeys); i++ {
		if strings.EqualFold(pubkeys[i], pubkey) {
			return true
		}
	}

	return false
}

//queryVote 根据用户地址信息查询用户的投票情况
func queryVote(kvdb db.KVDB, req *dty.DposVoteQuery) (types.Message, error) {
	var voters []*dty.DposVoter
	voteTable := dty.NewDposVoteTable(kvdb)
	query := voteTable.GetQuery(kvdb)

	rows, err := query.ListIndex("addr", []byte(req.Addr), nil, 0, 0)
	if err != nil {
		return nil, err
	}

	for index := 0; index < len(rows); index++ {
		voter := rows[index].Data.(*dty.DposVoter)
		voters = append(voters, voter)
	}

	//如果不指定pubkeys，则返回所有；否则，需要判断pubkey是否为指定的值之一。
	if len(req.Pubkeys) == 0 {
		return &dty.DposVoteReply{Votes: voters}, nil
	}

	reply := &dty.DposVoteReply{}
	for index := 0; index < len(voters); index ++ {
		strPubkey := hex.EncodeToString(voters[index].Pubkey)
		if isValidPubkey(req.Pubkeys, strPubkey) {
			reply.Votes = append(reply.Votes, voters[index])
		}
	}

	return reply, nil
}

func (action *Action) saveCandicator(candInfo *dty.CandidatorInfo) (kvset []*types.KeyValue) {
	value := types.Encode(candInfo)
	pubkey := hex.EncodeToString(candInfo.GetPubkey())
	err := action.db.Set(Key(pubkey), value)
	if err != nil {
		logger.Error("saveCandicator have err:", err.Error())
	}
	kvset = append(kvset, &types.KeyValue{Key: Key(pubkey), Value: value})
	return kvset
}

func (action *Action) getIndex() int64 {
	return action.height*types.MaxTxsPerBlock + int64(action.index)
}

//getReceiptLog 根据候选节点信息及投票信息生成收据信息
func (action *Action) getReceiptLog(candInfo *dty.CandidatorInfo, statusChange bool,  voted bool, vote *dty.DposVote) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &dty.ReceiptCandicator{}
	//r.StartIndex = can.StartIndex

	if candInfo.Status == dty.CandidatorStatusRegist {
		log.Ty = dty.TyLogCandicatorRegist
	} else if candInfo.Status == dty.CandidatorStatusVoted {
		log.Ty = dty.TyLogCandicatorVoted
	} else if candInfo.Status == dty.CandidatorStatusCancelRegist {
		log.Ty = dty.TyLogCandicatorCancelRegist
	} else if candInfo.Status == dty.CandidatorStatusReRegist{
		log.Ty = dty.TyLogCandicatorReRegist
	}

	r.Index = action.getIndex()
	r.Time = action.blocktime
	r.StatusChange = statusChange

	r.Status = candInfo.Status
	r.PreStatus = candInfo.PreStatus
	r.Address = candInfo.Address
	r.Pubkey = candInfo.Pubkey
	r.Voted = voted
	if voted {
		r.Votes = vote.Votes
		r.FromAddr = vote.FromAddr
		if r.Votes < 0 {
			log.Ty = dty.TyLogCandicatorCancelVoted
		}
	}
	r.CandInfo = candInfo
	log.Log = types.Encode(r)
	return log
}

//readCandicatorInfo 根据候选节点的公钥查询候选节点信息
func (action *Action) readCandicatorInfo(pubkey []byte) (*dty.CandidatorInfo, error) {
	strPubkey := hex.EncodeToString(pubkey)
	data, err := action.db.Get(Key(strPubkey))
	if err != nil {
		logger.Error("readCandicator have err:", err.Error())
		return nil, err
	}
	var cand dty.CandidatorInfo
	//decode
	err = types.Decode(data, &cand)
	if err != nil {
		logger.Error("decode candicator have err:", err.Error())
		return nil, err
	}
	return &cand, nil
}

// newCandicatorInfo 新建候选节点信息对象
func (action *Action) newCandicatorInfo(regist *dty.DposCandidatorRegist) *dty.CandidatorInfo {
	bPubkey, _ := hex.DecodeString(regist.Pubkey)
	candInfo := &dty.CandidatorInfo{
		Pubkey: bPubkey,
		Address: regist.Address,
		Ip: regist.Ip,
	}
	return candInfo
}

//Regist 注册候选节点
func (action *Action) Regist(regist *dty.DposCandidatorRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(regist.Pubkey)
	if err != nil {
		logger.Info("Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			regist.Pubkey)
		return nil, types.ErrInvalidParam
	}

	candInfo, err := action.readCandicatorInfo(bPubkey)
	if err == nil && candInfo != nil {
		logger.Info("Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is exist",
			candInfo.String())
		return nil, dty.ErrCandidatorExist
	}

	acc := action.coinsAccount.LoadExecAccount(action.fromaddr, action.execaddr)
	if acc.GetFrozen() < dty.RegistFrozenCoins {
		if acc.GetBalance() + acc.GetFrozen() < dty.RegistFrozenCoins {
			logger.Error("Regist failed", "addr", action.fromaddr, "execaddr", action.execaddr, "Balance", acc.GetBalance(), "Frozen", acc.GetFrozen(),"err", types.ErrNoBalance)
			return nil, types.ErrNoBalance
		}

		receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, dty.RegistFrozenCoins)
		if err != nil {
			logger.Error("ExecFrozen failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", dty.RegistFrozenCoins, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	logger.Info("Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "new candicator", regist.String())

	candInfo = action.newCandicatorInfo(regist)
	candInfo.Status = dty.CandidatorStatusRegist
	candInfo.StartTime = action.blocktime
	candInfo.StartHeight = action.mainHeight
	candInfo.StartIndex = action.getIndex()
	candInfo.Index = candInfo.StartIndex
	candInfo.StartTxHash = common.ToHex(action.txhash)
	candInfo.PreIndex = 0

	receiptLog := action.getReceiptLog(candInfo, false, false, nil)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveCandicator(candInfo)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//ReRegist 重新注册一个注销的候选节点
func (action *Action) ReRegist(regist *dty.DposCandidatorRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(regist.Pubkey)
	if err != nil {
		logger.Info("ReRegist", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			regist.Pubkey)
		return nil, types.ErrInvalidParam
	}

	candInfo, err := action.readCandicatorInfo(bPubkey)
	if err != nil || candInfo == nil {
		logger.Info("ReRegist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status != dty.CandidatorStatusCancelRegist {
		logger.Info("ReRegist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator status is not correct",
			candInfo.String())
		return nil, dty.ErrCandidatorInvalidStatus
	}

	acc := action.coinsAccount.LoadExecAccount(action.fromaddr, action.execaddr)
	if acc.GetFrozen() < dty.RegistFrozenCoins {
		if acc.GetBalance() + acc.GetFrozen() < dty.RegistFrozenCoins {
			logger.Error("Regist failed", "addr", action.fromaddr, "execaddr", action.execaddr, "Balance", acc.GetBalance(), "Frozen", acc.GetFrozen(),"err", types.ErrNoBalance)
			return nil, types.ErrNoBalance
		}

		receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, dty.RegistFrozenCoins)
		if err != nil {
			logger.Error("ExecFrozen failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", dty.RegistFrozenCoins, "err", err.Error())
			return nil, err
		}
		logs = append(logs, receipt.Logs...)
		kv = append(kv, receipt.KV...)
	}

	logger.Info("Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "new candicator", regist.String())

	candInfo = action.newCandicatorInfo(regist)
	candInfo.Status = dty.CandidatorStatusReRegist
	candInfo.PreStatus = dty.CandidatorStatusCancelRegist
	candInfo.StartTime = action.blocktime
	candInfo.StartHeight = action.mainHeight
	candInfo.StartIndex = action.getIndex()
	candInfo.Index = candInfo.StartIndex
	candInfo.StartTxHash = common.ToHex(action.txhash)
	candInfo.PreIndex = 0

	receiptLog := action.getReceiptLog(candInfo, false, false, nil)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveCandicator(candInfo)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}


//CancelRegist 撤销一个候选节点的注册
func (action *Action) CancelRegist(req *dty.DposCandidatorCancelRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(req.Pubkey)
	if err != nil {
		logger.Info("CancelRegist", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			req.Pubkey)
		return nil, types.ErrInvalidParam
	}

	candInfo, err := action.readCandicatorInfo(bPubkey)
	if err != nil  || candInfo == nil{
		logger.Error("Cancel Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status == dty.CandidatorStatusCancelRegist {
		logger.Error("Cancel Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is already canceled.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	if action.fromaddr != candInfo.GetAddress() {
		logger.Error("Cancel Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "from addr is not candicator address.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	logger.Info("Cancel Regist", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator",
		candInfo.String())

	if candInfo.Status == dty.CandidatorStatusVoted {
		for _, voter := range candInfo.Voters{
			receipt, err := action.coinsAccount.ExecActive(voter.FromAddr, action.execaddr, voter.Votes)
			if err != nil {
				//action.coinsAccount.ExecFrozen(game.AdminAddr, action.execaddr, devFee) // rollback
				logger.Error("Cancel Regist active votes", "addr", voter.FromAddr, "execaddr", action.execaddr,
					"amount", voter.Votes, "err", err)
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kv = append(kv, receipt.KV...)
		}
	}

	receipt, err := action.coinsAccount.ExecActive(action.fromaddr, action.execaddr, dty.RegistFrozenCoins)
	if err != nil {
		logger.Error("ExecActive failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", dty.RegistFrozenCoins, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	candInfo.PreStatus = candInfo.Status
	candInfo.Status = dty.CandidatorStatusCancelRegist
	candInfo.PreIndex = candInfo.Index
	candInfo.Index = action.getIndex()
	candInfo.Voters = nil
	receiptLog := action.getReceiptLog(candInfo, true, false, nil)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveCandicator(candInfo)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//Vote 为某一个候选节点投票
func (action *Action) Vote(vote *dty.DposVote) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(vote.Pubkey)
	if err != nil {
		logger.Info("Vote", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			vote.Pubkey)
		return nil, types.ErrInvalidParam
	}

	candInfo, err := action.readCandicatorInfo(bPubkey)
	if err != nil  || candInfo == nil{
		logger.Error("Vote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status == dty.CandidatorStatusCancelRegist {
		logger.Error("Vote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is already canceled.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	logger.Info("vote", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator",
		candInfo.String())

	statusChange := false
	if candInfo.Status == dty.CandidatorStatusRegist || candInfo.Status == dty.CandidatorStatusReRegist{
		candInfo.PreStatus = candInfo.Status
		candInfo.Status = dty.CandidatorStatusVoted
		statusChange = true
	}

	checkValue := vote.Votes
	if !action.CheckExecAccountBalance(action.fromaddr, checkValue, 0) {
		logger.Error("Vote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "err", types.ErrNoBalance)
		return nil, types.ErrNoBalance
	}

	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, checkValue)
	if err != nil {
		logger.Error("ExecFrozen failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", checkValue, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	voter := &dty.DposVoter{
		FromAddr: vote.FromAddr,
		Pubkey: bPubkey,
		Votes: vote.Votes,
		Index: action.getIndex(),
		Time: action.blocktime,
	}
	candInfo.Voters = append(candInfo.Voters, voter)
	candInfo.Votes += vote.Votes
	candInfo.PreIndex = candInfo.Index
	candInfo.Index = action.getIndex()

	receiptLog := action.getReceiptLog(candInfo, statusChange, true, vote)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveCandicator(candInfo)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//CancelVote 撤销对某个候选节点的投票
func (action *Action) CancelVote(vote *dty.DposCancelVote) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(vote.Pubkey)
	if err != nil {
		logger.Info("CancelVote", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			vote.Pubkey)
		return nil, types.ErrInvalidParam
	}
	candInfo, err := action.readCandicatorInfo(bPubkey)
	if err != nil  || candInfo == nil{
		logger.Error("CancelVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status != dty.CandidatorStatusVoted {
		logger.Error("CancelVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is already canceled.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	logger.Info("CancelVote", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator",
		candInfo.String())


	votes := vote.Votes
	availVotes := int64(0)
	enoughVotes := false
	for _, voter := range candInfo.Voters {
		if voter.FromAddr == action.fromaddr && bytes.Equal(voter.Pubkey, bPubkey){
			//if action.blocktime - voter.Time >= dty.VoteFrozenTime {
				availVotes += voter.Votes
				if availVotes >= votes {
					enoughVotes = true
					break
				}
			//}
		}
	}
	if !enoughVotes {
		logger.Error("RevokeVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "not enough avail votes",
			availVotes, "revoke votes", vote.Votes)
		return nil, dty.ErrNotEnoughVotes
	}

	for index, voter := range candInfo.Voters {
		if voter.FromAddr == action.fromaddr && bytes.Equal(voter.Pubkey, bPubkey){
			//if action.blocktime - voter.Time >= dty.VoteFrozenTime {
				if voter.Votes > votes {
					voter.Votes -= votes
					break
				} else if voter.Votes == votes {
					candInfo.Voters = append(candInfo.Voters[:index], candInfo.Voters[index+1:]...)
					break
				} else {
					candInfo.Voters = append(candInfo.Voters[:index], candInfo.Voters[index+1:]...)
					votes = votes - voter.Votes
				}
			//}
		}
	}

	checkValue := vote.Votes
	receipt, err := action.coinsAccount.ExecActive(action.fromaddr, action.execaddr, checkValue)
	if err != nil {
		logger.Error("ExecActive failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", checkValue, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	candInfo.Votes -= vote.Votes

	vote2 := &dty.DposVote{
		FromAddr: action.fromaddr,
		Pubkey: vote.Pubkey,
		Votes: (-1) * vote.Votes,
	}
	candInfo.PreIndex = candInfo.Index
	candInfo.Index = action.getIndex()

	receiptLog := action.getReceiptLog(candInfo, false, true, vote2)
	logs = append(logs, receiptLog)
	kv = append(kv, action.saveCandicator(candInfo)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}


//RegistVrfM 注册受托节点的Vrf M信息（输入信息）
func (action *Action) RegistVrfM(vrfMReg *dty.DposVrfMRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(vrfMReg.Pubkey)
	if err != nil {
		logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			vrfMReg.Pubkey)
		return nil, types.ErrInvalidParam
	}

	bM, err := hex.DecodeString(vrfMReg.M)
	if err != nil {
		logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "M is not correct",
			vrfMReg.M)
		return nil, types.ErrInvalidParam
	}


	req := &dty.CandidatorQuery{
		TopN: int32(dposDelegateNum),
	}

	reply, err := queryTopNCands(action.localDB, req)
	res := reply.(*dty.CandidatorReply)
	if err != nil || len(res.Candidators) < int(dposDelegateNum){
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "not enough Candidators",
			dposDelegateNum)
		return nil, dty.ErrCandidatorNotEnough
	}

	legalCand := false
    for i := 0; i < len(res.Candidators); i++ {
    	if bytes.Equal(bPubkey, res.Candidators[i].Pubkey) && action.fromaddr == res.Candidators[i].Address {
			legalCand = true
		}
	}

    if !legalCand {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "not legal Candidators",
			res.String())
		return nil, dty.ErrCandidatorNotLegal
	}

    cycleInfo := calcCycleByTime(action.blocktime)
    if vrfMReg.Cycle != cycleInfo.cycle {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "cycle is not the same with current blocktime",
			vrfMReg.String())
		return nil, types.ErrInvalidParam
	}

    //todo 还需要检查是否针对这个cycle已经有注册过M了，如果注册过了，也需要提示失败
	vrfMTable := dty.NewDposVrfMTable(action.localDB)
	query := vrfMTable.GetQuery(action.localDB)
	_, err = query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfMReg.Cycle)), nil, 1, 0)
	if err == nil {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "VrfM already is registed",
			vrfMReg.String())
		return nil, dty.ErrVrfMAlreadyRegisted
	}

	log := &types.ReceiptLog{}
	r := &dty.ReceiptVrf{}
	r.Index = action.getIndex()
	r.Pubkey = bPubkey
	r.Status = dty.VrfStatusMRegist
	r.Cycle = cycleInfo.cycle
	r.Height = action.mainHeight
	r.M = bM
	r.Time = action.blocktime

	log.Ty = dty.TyLogVrfMRegist
	log.Log = types.Encode(r)

	logs = append(logs, log)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//RegistVrfRP 注册受托节点的Vrf R/P信息
func (action *Action) RegistVrfRP(vrfRPReg *dty.DposVrfRPRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	bPubkey, err := hex.DecodeString(vrfRPReg.Pubkey)
	if err != nil {
		logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			vrfRPReg.Pubkey)
		return nil, types.ErrInvalidParam
	}

	bR, err := hex.DecodeString(vrfRPReg.R)
	if err != nil {
		logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "M is not correct",
			vrfRPReg.R)
		return nil, types.ErrInvalidParam
	}

	bP, err := hex.DecodeString(vrfRPReg.P)
	if err != nil {
		logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "M is not correct",
			vrfRPReg.P)
		return nil, types.ErrInvalidParam
	}

	//todo 从localdb中查找对应的pubkey:cycle的信息，如果没找到，说明对应的M没有发布出来，则也不允许发布R,P。
	vrfMTable := dty.NewDposVrfMTable(action.localDB)
	query := vrfMTable.GetQuery(action.localDB)
	rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfRPReg.Cycle)), nil, 1, 0)
	if err != nil {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "VrfM is not exist",
			vrfRPReg.String())
		return nil, dty.ErrVrfMNotRegisted
	}
	//对于可以注册的R、P，则允许。

	cycleInfo := calcCycleByTime(action.blocktime)
	//对于cycle不一致的情况，则不允许注册
	if vrfRPReg.Cycle != cycleInfo.cycle {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "cycle is not the same with current blocktime",
			vrfRPReg.String())
		return nil, types.ErrInvalidParam
	}

	//todo 还需要检查是否针对这个cycle已经有注册过R、P了，如果注册过了，也需要提示失败
	VrfRPTable := dty.NewDposVrfRPTable(action.localDB)
	query = VrfRPTable.GetQuery(action.localDB)
	_,err = query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfRPReg.Cycle)), nil, 1, 0)
	if err == nil {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "RegistVrfRP is already resisted.",
			vrfRPReg.String())
		return nil, dty.ErrVrfRPAlreadyRegisted
	}

	log := &types.ReceiptLog{}
	r := &dty.ReceiptVrf{}
	r.Index = action.getIndex()
	r.Pubkey = bPubkey
	r.Status = dty.VrfStatusRPRegist
	r.Cycle = cycleInfo.cycle
	r.Height = action.mainHeight
	r.R = bR
	r.P = bP
	r.M = rows[0].Data.(*dty.DposVrfM).M
	r.Time = action.blocktime

	log.Ty = dty.TyLogVrfRPRegist
	log.Log = types.Encode(r)

	logs = append(logs, log)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/db"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
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

//TopNKey State数据库中存储记录的Key值格式转换
func TopNKey(id string) (key []byte) {
	key = append(key, []byte("mavl-"+dty.DPosX+"-"+"topn"+"-")...)
	key = append(key, []byte(id)...)
	return key
}

//queryVrfByTime 根据时间信息，查询TopN的受托节点的VRF信息
func queryVrfByTime(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByTime {
		return nil, types.ErrInvalidParam
	}

	cycleInfo := calcCycleByTime(req.Timestamp)
	req.Ty = dty.QueryVrfByCycle
	req.Cycle = cycleInfo.cycle
	return queryVrfByCycle(kvdb, req)
}

func getJSONVrfs(vrfs []*dty.VrfInfo) []*dty.JSONVrfInfo {
	var jsonVrfs []*dty.JSONVrfInfo
	for i := 0; i < len(vrfs); i++ {
		jsonVrf := &dty.JSONVrfInfo{
			Index:  vrfs[i].Index,
			Pubkey: hex.EncodeToString(vrfs[i].Pubkey),
			Cycle:  vrfs[i].Cycle,
			Height: vrfs[i].Height,
			M:      string(vrfs[i].M),
			R:      hex.EncodeToString(vrfs[i].R),
			P:      hex.EncodeToString(vrfs[i].P),
			Time:   vrfs[i].Time,
		}

		jsonVrfs = append(jsonVrfs, jsonVrf)
	}

	return jsonVrfs
}

func getVrfInfoFromVrfRP(vrfRP *dty.DposVrfRP) *dty.VrfInfo {
	if nil == vrfRP {
		return nil
	}

	vrf := &dty.VrfInfo{
		Index:  vrfRP.Index,
		Pubkey: vrfRP.Pubkey,
		Cycle:  vrfRP.Cycle,
		Height: vrfRP.Height,
		M:      vrfRP.M,
		R:      vrfRP.R,
		P:      vrfRP.P,
		Time:   vrfRP.Time,
	}

	return vrf
}

func isVrfMRecordExist(vrfM *dty.DposVrfM, vrfs []*dty.VrfInfo) bool {
	if nil == vrfM || nil == vrfs || 0 == len(vrfs) {
		return false
	}

	for i := 0; i < len(vrfs); i++ {
		if vrfM.Cycle == vrfs[i].Cycle && bytes.Equal(vrfM.Pubkey, vrfs[i].Pubkey) {
			return true
		}
	}

	return false
}

func getVrfInfoFromVrfM(vrfM *dty.DposVrfM) *dty.VrfInfo {
	if nil == vrfM {
		return nil
	}

	vrf := &dty.VrfInfo{
		Index:  vrfM.Index,
		Pubkey: vrfM.Pubkey,
		Cycle:  vrfM.Cycle,
		Height: vrfM.Height,
		M:      vrfM.M,
		Time:   vrfM.Time,
	}

	return vrf
}

//queryVrfByCycleAndPubkeys 根据Cycle、Pubkeys信息，查询受托节点的VRF信息
func queryVrfByCycleAndPubkeys(kvdb db.KVDB, pubkeys []string, cycle int64) []*dty.VrfInfo {
	VrfRPTable := dty.NewDposVrfRPTable(kvdb)
	query := VrfRPTable.GetQuery(kvdb)

	var tempPubkeys []string
	var vrfs []*dty.VrfInfo
	for i := 0; i < len(pubkeys); i++ {
		rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%s:%018d", strings.ToUpper(pubkeys[i]), cycle)), nil, 1, 0)
		if err != nil {
			logger.Error("queryVrf RP failed", "pubkey", pubkeys[i], "cycle", cycle)
			tempPubkeys = append(tempPubkeys, pubkeys[i])
			continue
		}

		vrfRP := rows[0].Data.(*dty.DposVrfRP)
		vrf := getVrfInfoFromVrfRP(vrfRP)
		vrfs = append(vrfs, vrf)
	}

	if len(tempPubkeys) == 0 {
		return vrfs
	}

	vrfMTable := dty.NewDposVrfMTable(kvdb)
	query = vrfMTable.GetQuery(kvdb)
	for i := 0; i < len(tempPubkeys); i++ {
		rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%s:%018d", strings.ToUpper(tempPubkeys[i]), cycle)), nil, 1, 0)
		if err != nil {
			logger.Error("queryVrf M failed", "pubkey", tempPubkeys[i], "cycle", cycle)
			continue
		}

		vrfM := rows[0].Data.(*dty.DposVrfM)
		vrf := getVrfInfoFromVrfM(vrfM)
		vrfs = append(vrfs, vrf)
	}

	return vrfs
}

//queryVrfByCycleForPubkeys 根据Cycle、Pubkeys信息，查询受托节点的VRF信息
func queryVrfByCycleForPubkeys(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByCycleForPubkeys {
		return nil, types.ErrInvalidParam
	}

	vrfs := queryVrfByCycleAndPubkeys(kvdb, req.Pubkeys, req.Cycle)

	return &dty.DposVrfReply{Vrf: getJSONVrfs(vrfs)}, nil
}

//queryVrfByCycleForTopN 根据Cycle信息，查询TopN的受托节点的VRF信息
func queryVrfByCycleForTopN(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByCycleForTopN {
		return nil, types.ErrInvalidParam
	}

	topNReq := &dty.CandidatorQuery{
		TopN: int32(dposDelegateNum),
	}

	reply, err := queryTopNCands(kvdb, topNReq)
	res := reply.(*dty.CandidatorReply)
	if err != nil || len(res.Candidators) < int(dposDelegateNum) {
		logger.Error("queryVrf failed", "Candidators", len(res.Candidators), "need Candidators", dposDelegateNum)
		return nil, dty.ErrCandidatorNotEnough
	}

	var pubkeys []string
	for i := 0; i < len(res.Candidators); i++ {
		pubkeys = append(pubkeys, res.Candidators[i].Pubkey)
		//zzh
		logger.Info("queryVrfByCycleForTopN", "pubkey", pubkeys[i])
	}

	vrfs := queryVrfByCycleAndPubkeys(kvdb, pubkeys, req.Cycle)

	return &dty.DposVrfReply{Vrf: getJSONVrfs(vrfs)}, nil
}

//queryVrfByCycle 根据Cycle信息，查询所有受托节点的VRF信息
func queryVrfByCycle(kvdb db.KVDB, req *dty.DposVrfQuery) (types.Message, error) {
	if req.Ty != dty.QueryVrfByCycle {
		return nil, types.ErrInvalidParam
	}

	VrfRPTable := dty.NewDposVrfRPTable(kvdb)
	query := VrfRPTable.GetQuery(kvdb)

	var vrfs []*dty.VrfInfo
	rows, err := query.ListIndex("cycle", []byte(fmt.Sprintf("%018d", req.Cycle)), nil, 0, 0)
	if err != nil {
		logger.Error("queryVrf RP failed", "cycle", req.Cycle)
	} else {
		for i := 0; i < len(rows); i++ {
			vrfRP := rows[i].Data.(*dty.DposVrfRP)
			vrf := getVrfInfoFromVrfRP(vrfRP)
			vrfs = append(vrfs, vrf)
		}
	}

	vrfMTable := dty.NewDposVrfMTable(kvdb)
	query = vrfMTable.GetQuery(kvdb)
	rows, err = query.ListIndex("cycle", []byte(fmt.Sprintf("%018d", req.Cycle)), nil, 1, 0)
	if err != nil {
		logger.Error("queryVrf M failed", "cycle", req.Cycle)
	} else {
		for i := 0; i < len(rows); i++ {
			vrfM := rows[i].Data.(*dty.DposVrfM)
			if !isVrfMRecordExist(vrfM, vrfs) {
				vrf := getVrfInfoFromVrfM(vrfM)
				vrfs = append(vrfs, vrf)
			}
		}
	}

	return &dty.DposVrfReply{Vrf: getJSONVrfs(vrfs)}, nil
}

//queryCands 根据候选节点的Pubkey下旬候选节点信息,得票数、状态等
func queryCands(kvdb db.KVDB, req *dty.CandidatorQuery) (types.Message, error) {
	var cands []*dty.JSONCandidator
	candTable := dty.NewDposCandidatorTable(kvdb)
	query := candTable.GetQuery(kvdb)

	for i := 0; i < len(req.Pubkeys); i++ {
		bPubkey, _ := hex.DecodeString(req.Pubkeys[i])
		rows, err := query.ListIndex("pubkey", bPubkey, nil, 1, 0)
		if err != nil {
			continue
		}

		candInfo := rows[0].Data.(*dty.CandidatorInfo)
		cand := &dty.JSONCandidator{
			Pubkey:  strings.ToUpper(hex.EncodeToString(candInfo.Pubkey)),
			Address: candInfo.Address,
			IP:      candInfo.IP,
			Votes:   candInfo.Votes,
			Status:  candInfo.Status,
		}
		cands = append(cands, cand)
	}
	return &dty.CandidatorReply{Candidators: cands}, nil
}

//queryTopNCands 查询得票数TopN的候选节点信息，包括得票数，状态等
func queryTopNCands(kvdb db.KVDB, req *dty.CandidatorQuery) (types.Message, error) {
	var cands []*dty.JSONCandidator
	candTable := dty.NewDposCandidatorTable(kvdb)
	query := candTable.GetQuery(kvdb)

	number := int32(0)
	rows, err := query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusVoted)), nil, 0, 0)
	if err == nil {
		for index := 0; index < len(rows); index++ {
			candInfo := rows[index].Data.(*dty.CandidatorInfo)
			cand := &dty.JSONCandidator{
				Pubkey:  strings.ToUpper(hex.EncodeToString(candInfo.Pubkey)),
				Address: candInfo.Address,
				IP:      candInfo.IP,
				Votes:   candInfo.Votes,
				Status:  candInfo.Status,
			}
			cands = append(cands, cand)
			number++
		}

		sort.Slice(cands, func(i, j int) bool {
			return cands[i].Votes > cands[j].Votes
		})
	}

	if number < req.TopN {
		rows, err = query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusRegist)), nil, req.TopN-number, 0)
		if err == nil {
			for index := 0; index < len(rows); index++ {
				candInfo := rows[index].Data.(*dty.CandidatorInfo)
				cand := &dty.JSONCandidator{
					Pubkey:  strings.ToUpper(hex.EncodeToString(candInfo.Pubkey)),
					Address: candInfo.Address,
					IP:      candInfo.IP,
					Votes:   candInfo.Votes,
					Status:  candInfo.Status,
				}
				cands = append(cands, cand)
				number++
				if number == req.TopN {
					break
				}
			}
		}

		rows, err = query.ListIndex("status", []byte(fmt.Sprintf("%2d", dty.CandidatorStatusReRegist)), nil, req.TopN-number, 0)
		if err == nil {
			for index := 0; index < len(rows); index++ {
				candInfo := rows[index].Data.(*dty.CandidatorInfo)
				cand := &dty.JSONCandidator{
					Pubkey:  strings.ToUpper(hex.EncodeToString(candInfo.Pubkey)),
					Address: candInfo.Address,
					IP:      candInfo.IP,
					Votes:   candInfo.Votes,
					Status:  candInfo.Status,
				}
				cands = append(cands, cand)
				number++
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
func isValidPubkey(pubkeys []string, pubkey string) bool {
	if len(pubkeys) == 0 || len(pubkey) == 0 {
		return false
	}

	for i := 0; i < len(pubkeys); i++ {
		if strings.EqualFold(pubkeys[i], pubkey) {
			return true
		}
	}

	return false
}

//queryVote 根据用户地址信息查询用户的投票情况
func queryVote(kvdb db.KVDB, req *dty.DposVoteQuery) (types.Message, error) {
	var voters []*dty.JSONDposVoter
	voteTable := dty.NewDposVoteTable(kvdb)
	query := voteTable.GetQuery(kvdb)

	rows, err := query.ListIndex("addr", []byte(req.Addr), nil, 0, 0)
	if err != nil {
		return nil, err
	}

	for index := 0; index < len(rows); index++ {
		voter := rows[index].Data.(*dty.DposVoter)
		jsonVoter := &dty.JSONDposVoter{
			FromAddr: voter.FromAddr,
			Pubkey:   strings.ToUpper(hex.EncodeToString(voter.Pubkey)),
			Votes:    voter.Votes,
			Index:    voter.Index,
			Time:     voter.Time,
		}
		voters = append(voters, jsonVoter)
	}

	//如果不指定pubkeys，则返回所有；否则，需要判断pubkey是否为指定的值之一。
	if len(req.Pubkeys) == 0 {
		return &dty.DposVoteReply{Votes: voters}, nil
	}

	reply := &dty.DposVoteReply{}
	for index := 0; index < len(voters); index++ {
		strPubkey := voters[index].Pubkey
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
func (action *Action) getReceiptLog(candInfo *dty.CandidatorInfo, statusChange bool, voteType int32, vote *dty.DposVoter) *types.ReceiptLog {
	log := &types.ReceiptLog{}
	r := &dty.ReceiptCandicator{}
	//r.StartIndex = can.StartIndex

	if candInfo.Status == dty.CandidatorStatusRegist {
		log.Ty = dty.TyLogCandicatorRegist
	} else if candInfo.Status == dty.CandidatorStatusVoted {
		log.Ty = dty.TyLogCandicatorVoted
	} else if candInfo.Status == dty.CandidatorStatusCancelRegist {
		log.Ty = dty.TyLogCandicatorCancelRegist
	} else if candInfo.Status == dty.CandidatorStatusReRegist {
		log.Ty = dty.TyLogCandicatorReRegist
	}

	r.Index = action.getIndex()
	r.Time = action.blocktime
	r.StatusChange = statusChange

	r.Status = candInfo.Status
	r.PreStatus = candInfo.PreStatus
	r.Address = candInfo.Address
	r.Pubkey = candInfo.Pubkey
	r.VoteType = voteType
	switch voteType {
	case dty.VoteTypeNone:
	case dty.VoteTypeVote:
		r.Vote = vote
	case dty.VoteTypeCancelVote:
		log.Ty = dty.TyLogCandicatorCancelVoted
		r.Vote = vote
	case dty.VoteTypeCancelAllVote:
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
		logger.Error("readCandicator have err:", "err", err.Error())
		return nil, err
	}
	var cand dty.CandidatorInfo
	//decode
	err = types.Decode(data, &cand)
	if err != nil {
		logger.Error("decode candicator have err:", "err", err.Error())
		return nil, err
	}
	return &cand, nil
}

// newCandicatorInfo 新建候选节点信息对象
func (action *Action) newCandicatorInfo(regist *dty.DposCandidatorRegist) *dty.CandidatorInfo {
	bPubkey, _ := hex.DecodeString(regist.Pubkey)
	candInfo := &dty.CandidatorInfo{
		Pubkey:  bPubkey,
		Address: regist.Address,
		IP:      regist.IP,
	}
	return candInfo
}

//readTopNCandicators 根据版本信息查询特定高度区间的TOPN候选节点信息
func (action *Action) readTopNCandicators(version int64) (*dty.TopNCandidators, error) {
	strVersion := fmt.Sprintf("%018d", version)
	data, err := action.db.Get(TopNKey(strVersion))
	if err != nil {
		logger.Error("readTopNCandicators have err:", "err", err.Error())
		return nil, err
	}
	var cands dty.TopNCandidators
	//decode
	err = types.Decode(data, &cands)
	if err != nil {
		logger.Error("decode TopNCandidators have err:", err.Error())
		return nil, err
	}
	return &cands, nil
}

func (action *Action) saveTopNCandicators(topCands *dty.TopNCandidators) (kvset []*types.KeyValue) {
	value := types.Encode(topCands)
	strVersion := fmt.Sprintf("%018d", topCands.Version)
	err := action.db.Set(TopNKey(strVersion), value)
	if err != nil {
		logger.Error("saveCandicator have err:", err.Error())
	}
	kvset = append(kvset, &types.KeyValue{Key: TopNKey(strVersion), Value: value})
	return kvset
}

//queryCBInfoByCycle 根据cycle查询stopHeight及stopHash等CBInfo信息，用于VRF计算
func queryCBInfoByCycle(kvdb db.KVDB, req *dty.DposCBQuery) (types.Message, error) {
	cbTable := dty.NewDposCBTable(kvdb)
	query := cbTable.GetQuery(kvdb)

	rows, err := query.ListIndex("cycle", []byte(fmt.Sprintf("%018d", req.Cycle)), nil, 1, 0)
	if err != nil {
		logger.Error("queryCBInfoByCycle have err", "cycle", req.Cycle, "err", err.Error())
		return nil, err
	}

	cbInfo := rows[0].Data.(*dty.DposCycleBoundaryInfo)
	info := &dty.DposCBInfo{
		Cycle:      cbInfo.Cycle,
		StopHeight: cbInfo.StopHeight,
		StopHash:   hex.EncodeToString(cbInfo.StopHash),
		Pubkey:     strings.ToUpper(hex.EncodeToString(cbInfo.Pubkey)),
		Signature:  hex.EncodeToString(cbInfo.StopHash),
	}
	logger.Info("queryCBInfoByCycle ok", "cycle", req.Cycle, "info", info.String())

	return &dty.DposCBReply{CbInfo: info}, nil
}

//queryCBInfoByHeight 根据stopHeight查询stopHash等CBInfo信息，用于VRF计算
func queryCBInfoByHeight(kvdb db.KVDB, req *dty.DposCBQuery) (types.Message, error) {
	cbTable := dty.NewDposCBTable(kvdb)
	query := cbTable.GetQuery(kvdb)

	rows, err := query.ListIndex("height", []byte(fmt.Sprintf("%018d", req.StopHeight)), nil, 1, 0)
	if err != nil {
		logger.Error("queryCBInfoByHeight have err", "height", req.StopHeight, "err", err.Error())
		return nil, err
	}

	cbInfo := rows[0].Data.(*dty.DposCycleBoundaryInfo)
	info := &dty.DposCBInfo{
		Cycle:      cbInfo.Cycle,
		StopHeight: cbInfo.StopHeight,
		StopHash:   hex.EncodeToString(cbInfo.StopHash),
		Pubkey:     strings.ToUpper(hex.EncodeToString(cbInfo.Pubkey)),
		Signature:  hex.EncodeToString(cbInfo.StopHash),
	}
	logger.Info("queryCBInfoByHeight ok", "height", req.StopHeight, "info", info.String())

	return &dty.DposCBReply{CbInfo: info}, nil
}

//queryCBInfoByHash 根据stopHash查询CBInfo信息，用于VRF计算
func queryCBInfoByHash(kvdb db.KVDB, req *dty.DposCBQuery) (types.Message, error) {
	cbTable := dty.NewDposCBTable(kvdb)
	query := cbTable.GetQuery(kvdb)

	hash, err := hex.DecodeString(req.StopHash)
	if err != nil {
		logger.Error("queryCBInfoByHash failed for decoding hash failed", "hash", req.StopHash, "err", err.Error())

		return nil, err
	}
	rows, err := query.ListIndex("hash", hash, nil, 1, 0)
	if err != nil {
		logger.Error("queryCBInfoByHash have err", "hash", req.StopHash, "err", err.Error())

		return nil, err
	}

	cbInfo := rows[0].Data.(*dty.DposCycleBoundaryInfo)
	info := &dty.DposCBInfo{
		Cycle:      cbInfo.Cycle,
		StopHeight: cbInfo.StopHeight,
		StopHash:   hex.EncodeToString(cbInfo.StopHash),
		Pubkey:     strings.ToUpper(hex.EncodeToString(cbInfo.Pubkey)),
		Signature:  hex.EncodeToString(cbInfo.StopHash),
	}
	logger.Info("queryCBInfoByHash ok", "hash", req.StopHash, "info", info.String())

	return &dty.DposCBReply{CbInfo: info}, nil
}

//queryTopNByVersion 根据version查询具体周期使用的TopN超级节点信息
func queryTopNByVersion(db dbm.KV, req *dty.TopNCandidatorsQuery) (types.Message, error) {
	strVersion := fmt.Sprintf("%018d", req.Version)
	data, err := db.Get(TopNKey(strVersion))
	if err != nil || data == nil {
		logger.Error("queryTopNByVersion have err", "err", err.Error())
		return nil, err
	}
	var cands dty.TopNCandidators
	//decode
	err = types.Decode(data, &cands)
	if err != nil {
		logger.Error("decode TopNCandidators have err:", err.Error())
		return nil, err
	}

	reply := &dty.TopNCandidatorsReply{
		TopN: &cands,
	}

	return reply, nil
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
		if acc.GetBalance()+acc.GetFrozen() < dty.RegistFrozenCoins {
			logger.Error("Regist failed", "addr", action.fromaddr, "execaddr", action.execaddr, "Balance", acc.GetBalance(), "Frozen", acc.GetFrozen(), "err", types.ErrNoBalance)
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

	receiptLog := action.getReceiptLog(candInfo, false, dty.VoteTypeNone, nil)
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
		if acc.GetBalance()+acc.GetFrozen() < dty.RegistFrozenCoins {
			logger.Error("Regist failed", "addr", action.fromaddr, "execaddr", action.execaddr, "Balance", acc.GetBalance(), "Frozen", acc.GetFrozen(), "err", types.ErrNoBalance)
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
	candInfo.Votes = 0
	candInfo.Voters = nil

	receiptLog := action.getReceiptLog(candInfo, false, dty.VoteTypeNone, nil)

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
	if err != nil || candInfo == nil {
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
		for _, voter := range candInfo.Voters {
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

	receiptLog := action.getReceiptLog(candInfo, true, dty.VoteTypeCancelAllVote, nil)

	candInfo.Votes = 0
	candInfo.Voters = nil

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
	if err != nil || candInfo == nil {
		logger.Error("Vote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status == dty.CandidatorStatusCancelRegist {
		logger.Error("Vote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is already canceled.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	logger.Info("vote", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator", candInfo.String())

	statusChange := false
	if candInfo.Status == dty.CandidatorStatusRegist || candInfo.Status == dty.CandidatorStatusReRegist {
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
		Pubkey:   bPubkey,
		Votes:    vote.Votes,
		Index:    action.getIndex(),
		Time:     action.blocktime,
	}
	candInfo.Voters = append(candInfo.Voters, voter)
	candInfo.Votes += vote.Votes
	candInfo.PreIndex = candInfo.Index
	candInfo.Index = action.getIndex()

	receiptLog := action.getReceiptLog(candInfo, statusChange, dty.VoteTypeVote, voter)
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
	if err != nil || candInfo == nil {
		logger.Error("CancelVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is not exist",
			candInfo.String())
		return nil, dty.ErrCandidatorNotExist
	}

	if candInfo.Status != dty.CandidatorStatusVoted {
		logger.Error("CancelVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, "candicator is already canceled.",
			candInfo.String())
		return nil, types.ErrInvalidParam
	}

	logger.Info("CancelVote", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey", vote.Pubkey, "index", vote.Index)

	findVote := false
	var oriVote *dty.DposVoter
	for index, voter := range candInfo.Voters {
		if voter.FromAddr == action.fromaddr && bytes.Equal(voter.Pubkey, bPubkey) && voter.Index == vote.Index {
			oriVote = voter
			findVote = true
			candInfo.Voters = append(candInfo.Voters[0:index], candInfo.Voters[index+1:]...)
			break
		}
	}

	if !findVote {
		logger.Error("CancelVote failed", "addr", action.fromaddr, "execaddr", action.execaddr, vote.Pubkey, "index", vote.Index)
		return nil, dty.ErrNoSuchVote
	}

	receipt, err := action.coinsAccount.ExecActive(action.fromaddr, action.execaddr, oriVote.Votes)
	if err != nil {
		logger.Error("ExecActive failed", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", oriVote.Votes, "err", err.Error())
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)

	candInfo.Votes -= oriVote.Votes
	candInfo.PreIndex = candInfo.Index
	candInfo.Index = action.getIndex()

	receiptLog := action.getReceiptLog(candInfo, false, dty.VoteTypeCancelVote, oriVote)
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

	bM := []byte(vrfMReg.M)

	req := &dty.CandidatorQuery{}
	req.Pubkeys = append(req.Pubkeys, vrfMReg.Pubkey)

	reply, err := queryCands(action.localDB, req)
	res := reply.(*dty.CandidatorReply)
	if err != nil || len(res.Candidators) != 1 || res.Candidators[0].Status == dty.CandidatorStatusCancelRegist {
		logger.Error("RegistVrfM failed for no valid Candidators", "addr", action.fromaddr, "execaddr", action.execaddr)
		return nil, dty.ErrCandidatorNotExist
	}

	legalCand := false
	if (strings.ToUpper(vrfMReg.Pubkey) == res.Candidators[0].Pubkey) && (action.fromaddr == res.Candidators[0].Address) {
		legalCand = true
	}

	if !legalCand {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "not legal Candidator",
			res.Candidators[0].String())
		return nil, dty.ErrCandidatorNotLegal
	}

	cycleInfo := calcCycleByTime(action.blocktime)
	middleTime := cycleInfo.cycleStart + (cycleInfo.cycleStop-cycleInfo.cycleStart)/2
	if vrfMReg.Cycle != cycleInfo.cycle {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "cycle is not correct",
			vrfMReg.String(), "current cycle info", fmt.Sprintf("cycle:%d,start:%d,stop:%d,time:%d", cycleInfo.cycle, cycleInfo.cycleStart, cycleInfo.cycleStop, action.blocktime))
		return nil, types.ErrInvalidParam
	} else if action.blocktime > middleTime {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "time is not allowed, over the middle of this cycle",
			action.blocktime, "allow time", fmt.Sprintf("cycle:%d,start:%d,middle:%d,stop:%d", cycleInfo.cycle, cycleInfo.cycleStart, middleTime, cycleInfo.cycleStop))
		return nil, types.ErrInvalidParam
	}

	logger.Info("RegistVrfM", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey", vrfMReg.Pubkey, "cycle", vrfMReg.Cycle, "M", vrfMReg.M, "now", action.blocktime,
		"info", fmt.Sprintf("cycle:%d,start:%d,middle:%d,stop:%d", cycleInfo.cycle, cycleInfo.cycleStart, middleTime, cycleInfo.cycleStop))

	//todo 还需要检查是否针对这个cycle已经有注册过M了，如果注册过了，也需要提示失败
	vrfMTable := dty.NewDposVrfMTable(action.localDB)
	query := vrfMTable.GetQuery(action.localDB)
	_, err = query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfMReg.Cycle)), nil, 1, 0)
	if err == nil {
		logger.Error("RegistVrfM failed", "addr", action.fromaddr, "execaddr", action.execaddr, "VrfM already is registered",
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
	r.CycleStart = cycleInfo.cycleStart
	r.CycleStop = cycleInfo.cycleStop
	r.CycleMiddle = middleTime

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
		logger.Info("RegistVrfRP", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey is not correct",
			vrfRPReg.Pubkey)
		return nil, types.ErrInvalidParam
	}

	bR, err := hex.DecodeString(vrfRPReg.R)
	if err != nil {
		logger.Info("RegistVrfRP", "addr", action.fromaddr, "execaddr", action.execaddr, "R is not correct",
			vrfRPReg.R)
		return nil, types.ErrInvalidParam
	}

	bP, err := hex.DecodeString(vrfRPReg.P)
	if err != nil {
		logger.Info("RegistVrfRP", "addr", action.fromaddr, "execaddr", action.execaddr, "P is not correct",
			vrfRPReg.P)
		return nil, types.ErrInvalidParam
	}

	cycleInfo := calcCycleByTime(action.blocktime)
	middleTime := cycleInfo.cycleStart + (cycleInfo.cycleStop-cycleInfo.cycleStart)/2
	//对于cycle不一致的情况，则不允许注册
	if vrfRPReg.Cycle != cycleInfo.cycle {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "cycle is not the same with current blocktime",
			vrfRPReg.String())
		return nil, types.ErrInvalidParam
	} else if action.blocktime < middleTime {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "time is not allowed, not over the middle of this cycle",
			action.blocktime, "allow time", fmt.Sprintf("cycle:%d,start:%d,middle:%d,stop:%d", cycleInfo.cycle, cycleInfo.cycleStart, middleTime, cycleInfo.cycleStop))
		return nil, types.ErrInvalidParam
	}

	logger.Info("RegistVrfRP", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey", vrfRPReg.Pubkey, "cycle", vrfRPReg.Cycle, "R", vrfRPReg.R, "P", vrfRPReg.P,
		"now", action.blocktime, "info", fmt.Sprintf("cycle:%d,start:%d,middle:%d,stop:%d", cycleInfo.cycle, cycleInfo.cycleStart, middleTime, cycleInfo.cycleStop))

	//从localdb中查找对应的pubkey:cycle的信息，如果没找到，说明对应的M没有发布出来，则也不允许发布R,P。
	vrfMTable := dty.NewDposVrfMTable(action.localDB)
	query := vrfMTable.GetQuery(action.localDB)
	rows, err := query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfRPReg.Cycle)), nil, 1, 0)
	if err != nil {
		logger.Error("RegistVrfRP failed", "addr", action.fromaddr, "execaddr", action.execaddr, "VrfM is not registered",
			vrfRPReg.String())
		return nil, dty.ErrVrfMNotRegisted
	}
	//对于可以注册的R、P，则允许。

	//todo 还需要检查是否针对这个cycle已经有注册过R、P了，如果注册过了，也需要提示失败
	VrfRPTable := dty.NewDposVrfRPTable(action.localDB)
	query = VrfRPTable.GetQuery(action.localDB)
	_, err = query.ListIndex("pubkey_cycle", []byte(fmt.Sprintf("%X:%018d", bPubkey, vrfRPReg.Cycle)), nil, 1, 0)
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
	r.CycleStart = cycleInfo.cycleStart
	r.CycleStop = cycleInfo.cycleStop
	r.CycleMiddle = middleTime

	log.Ty = dty.TyLogVrfRPRegist
	log.Log = types.Encode(r)

	logs = append(logs, log)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//RecordCB 记录cycle boundary info
func (action *Action) RecordCB(cbInfo *dty.DposCBInfo) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	hash, err := hex.DecodeString(cbInfo.StopHash)
	if err != nil {
		logger.Info("RecordCB", "addr", action.fromaddr, "execaddr", action.execaddr, "StopHash is not correct", cbInfo.StopHash)
		return nil, types.ErrInvalidParam
	}

	pubkey, err := hex.DecodeString(cbInfo.Pubkey)
	if err != nil {
		logger.Info("RecordCB", "addr", action.fromaddr, "execaddr", action.execaddr, "Pubkey is not correct", cbInfo.Pubkey)
		return nil, types.ErrInvalidParam
	}

	sig, err := hex.DecodeString(cbInfo.Signature)
	if err != nil {
		logger.Info("RecordCB", "addr", action.fromaddr, "execaddr", action.execaddr, "Sig is not correct", cbInfo.Signature)
		return nil, types.ErrInvalidParam
	}

	logger.Info("RecordCB", "addr", action.fromaddr, "execaddr", action.execaddr, "info", fmt.Sprintf("cycle:%d,stopHeight:%d,stopHash:%s,pubkey:%s", cbInfo.Cycle, cbInfo.StopHeight, cbInfo.StopHash, cbInfo.Pubkey))

	cb := &dty.DposCycleBoundaryInfo{
		Cycle:      cbInfo.Cycle,
		StopHeight: cbInfo.StopHeight,
		StopHash:   hash,
		Pubkey:     pubkey,
		Signature:  sig,
	}

	cbTable := dty.NewDposCBTable(action.localDB)
	query := cbTable.GetQuery(action.localDB)
	rows, err := query.ListIndex("cycle", []byte(fmt.Sprintf("%018d", cbInfo.Cycle)), nil, 1, 0)
	if err == nil && rows[0] != nil {
		logger.Error("RecordCB failed", "addr", action.fromaddr, "execaddr", action.execaddr, "CB info is already recorded.", cbInfo.String())
		return nil, dty.ErrCBRecordExist
	}

	cycleInfo := calcCycleByTime(action.blocktime)

	if cbInfo.Cycle > cycleInfo.cycle+1 || cbInfo.Cycle < cycleInfo.cycle-2 {
		logger.Error("RecordCB failed for cycle over range", "addr", action.fromaddr, "execaddr", action.execaddr, "CB info cycle", cbInfo.Cycle, "current cycle", cycleInfo.cycle)
		return nil, dty.ErrCycleNotAllowed
	}

	middleTime := cycleInfo.cycleStart + (cycleInfo.cycleStop-cycleInfo.cycleStart)/2
	log := &types.ReceiptLog{}
	r := &dty.ReceiptCB{}
	r.Index = action.getIndex()
	r.Pubkey = pubkey
	r.Status = dty.CBStatusRecord
	r.Cycle = cycleInfo.cycle
	r.Height = action.mainHeight
	r.Time = action.blocktime
	r.CycleStart = cycleInfo.cycleStart
	r.CycleStop = cycleInfo.cycleStop
	r.CycleMiddle = middleTime
	r.CbInfo = cb

	log.Ty = dty.TyLogCBInfoRecord
	log.Log = types.Encode(r)

	logs = append(logs, log)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

//RegistTopN 注册TopN节点
func (action *Action) RegistTopN(regist *dty.TopNCandidatorRegist) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	err := regist.Cand.Verify()
	if err != nil {
		logger.Error("RegistTopN failed for signature verify failed.", "addr", action.fromaddr, "execaddr", action.execaddr)
		return nil, types.ErrInvalidParam
	}

	currentVersion, left := calcTopNVersion(action.mainHeight)
	topNVersion, _ := calcTopNVersion(regist.Cand.Height)
	if currentVersion != topNVersion {
		logger.Error("RegistTopN failed for wrong version.", "addr", action.fromaddr, "execaddr", action.execaddr,
			"regist height", regist.Cand.Height, "regist version", topNVersion, "current height", action.mainHeight, "current version", currentVersion)
		return nil, types.ErrInvalidParam
	}

	if left >= registTopNHeightLimit {
		logger.Error("RegistTopN failed for height limit.", "addr", action.fromaddr, "execaddr", action.execaddr,
			"current height", action.mainHeight, "registTopNHeightLimit", registTopNHeightLimit, "height in new circle", left)
		return nil, types.ErrInvalidParam
	}

	version := topNVersion - 1
	for version >= 0 {
		lastTopN, err := action.readTopNCandicators(version)
		if err != nil {
			logger.Error("read old TopN failed.", "addr", action.fromaddr, "execaddr", action.execaddr, "version", version)

			if version == 0 {
				//如果从没有注册过，认为是创世阶段，可信环境，只有可信的节点来注册，可以不做过多的判断。
				break
			} else {
				version--
				continue
			}
		}

		if lastTopN.Status != dty.TopNCandidatorsVoteMajorOK {
			logger.Error("Not legal topN exist.", "addr", action.fromaddr, "execaddr", action.execaddr, "version", version)
			if version > 0 {
				version--
				continue
			} else {
				break
			}
		}

		isLegalVoter := false
		for i := 0; i < len(lastTopN.FinalCands); i++ {
			if bytes.Equal(regist.Cand.SignerPubkey, lastTopN.FinalCands[i].Pubkey) {
				isLegalVoter = true
			}
		}

		if !isLegalVoter {
			logger.Error("RegistTopN failed for the voter is not legal topN.", "addr", action.fromaddr, "execaddr", action.execaddr, "voter pubkey", hex.EncodeToString(regist.Cand.SignerPubkey))
			return nil, dty.ErrNotLegalTopN
		}

		break
	}

	topNCands, err := action.readTopNCandicators(topNVersion)
	if err != nil {
		logger.Error("RegistTopN failed readTopNCandicators", "addr", action.fromaddr, "execaddr", action.execaddr, "version", topNVersion)
		return nil, types.ErrInvalidParam
	}
	if topNCands == nil {
		topNCands = &dty.TopNCandidators{
			Version: topNVersion,
			Status:  dty.TopNCandidatorsVoteInit,
		}
		topNCands.CandsVotes = append(topNCands.CandsVotes, regist.Cand)
	} else {
		for i := 0; i < len(topNCands.CandsVotes); i++ {
			if bytes.Equal(topNCands.CandsVotes[i].SignerPubkey, regist.Cand.SignerPubkey) {
				logger.Error("RegistTopN failed for vote exist.", "addr", action.fromaddr, "execaddr", action.execaddr, "pubkey", hex.EncodeToString(regist.Cand.SignerPubkey))
				return nil, types.ErrInvalidParam
			}
		}
		topNCands.CandsVotes = append(topNCands.CandsVotes, regist.Cand)
	}

	topNCands.CheckVoteStauts(dposDelegateNum)

	logger.Info("RegistTopN add one vote", "addr", action.fromaddr, "execaddr", action.execaddr, "version", topNVersion, "voter pubkey", hex.EncodeToString(regist.Cand.SignerPubkey))

	log := &types.ReceiptLog{}
	r := &dty.ReceiptTopN{}
	log.Ty = dty.TyLogTopNCandidatorRegist

	r.Index = action.getIndex()
	r.Time = action.blocktime
	r.Height = action.mainHeight
	r.Version = topNVersion
	r.Status = dty.TopNCandidatorStatusRegist
	r.Pubkey = regist.Cand.SignerPubkey
	r.TopN = regist.Cand
	log.Log = types.Encode(r)

	logs = append(logs, log)
	kv = append(kv, action.saveTopNCandicators(topNCands)...)

	return &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}, nil
}

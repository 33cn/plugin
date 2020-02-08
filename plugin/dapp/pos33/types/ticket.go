// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"encoding/json"
	"errors"
	"reflect"
	"time"

	//log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/types"
)

const (
	//log for ticket

	//TyLogNewPos33Ticket new ticket log type
	TyLogNewPos33Ticket = 111
	// TyLogClosePos33Ticket close ticket log type
	TyLogClosePos33Ticket = 112
	// TyLogMinerPos33Ticket miner ticket log type
	TyLogMinerPos33Ticket = 113
	// TyLogPos33TicketBind bind ticket log type
	TyLogPos33TicketBind = 114
)

//ticket
const (
	// Pos33TicketActionGenesis action type
	Pos33TicketActionGenesis = 11
	// Pos33TicketActionOpen action type
	Pos33TicketActionOpen = 12
	// Pos33TicketActionClose action type
	Pos33TicketActionClose = 13
	// Pos33TicketActionList  action type
	Pos33TicketActionList = 14 //读的接口不直接经过transaction
	// Pos33TicketActionInfos action type
	Pos33TicketActionInfos = 15 //读的接口不直接经过transaction
	// Pos33TicketActionMiner action miner
	Pos33TicketActionMiner = 16
	// Pos33TicketActionBind action bind
	Pos33TicketActionBind = 17
)

// Pos33TicketOldParts old tick type
const Pos33TicketOldParts = 3

// Pos33TicketCountOpenOnce count open once
const Pos33TicketCountOpenOnce = 1000

// ErrOpenPos33TicketPubHash err type
var ErrOpenPos33TicketPubHash = errors.New("ErrOpenPos33TicketPubHash")

// Pos33TicketX dapp name
var Pos33TicketX = "pos33"

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(Pos33TicketX))
	types.RegFork(Pos33TicketX, InitFork)
	types.RegExec(Pos33TicketX, InitExecutor)

}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(Pos33TicketX, "Enable", 0)
	//cfg.RegisterDappFork(Pos33TicketX, "ForkTicketId", 1062000)
	//cfg.RegisterDappFork(Pos33TicketX, "ForkPos33TicketVrf", 1770000)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(Pos33TicketX, NewType(cfg))
}

// Pos33TicketType ticket exec type
type Pos33TicketType struct {
	types.ExecTypeBase
}

// NewType new type
func NewType(cfg *types.Chain33Config) *Pos33TicketType {
	c := &Pos33TicketType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload get payload
func (ticket *Pos33TicketType) GetPayload() types.Message {
	return &Pos33TicketAction{}
}

// GetLogMap get log map
func (ticket *Pos33TicketType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogNewPos33Ticket:   {Ty: reflect.TypeOf(ReceiptPos33Ticket{}), Name: "LogNewPos33Ticket"},
		TyLogClosePos33Ticket: {Ty: reflect.TypeOf(ReceiptPos33Ticket{}), Name: "LogClosePos33Ticket"},
		TyLogMinerPos33Ticket: {Ty: reflect.TypeOf(ReceiptPos33Ticket{}), Name: "LogMinerPos33Ticket"},
		TyLogPos33TicketBind:  {Ty: reflect.TypeOf(ReceiptPos33TicketBind{}), Name: "LogPos33TicketBind"},
	}
}

// // Amount get amount
// func (ticket Pos33TicketType) Amount(tx *types.Transaction) (int64, error) {
// 	var action Pos33TicketAction
// 	err := types.Decode(tx.GetPayload(), &action)
// 	if err != nil {
// 		return 0, types.ErrDecode
// 	}
// 	if action.Ty == Pos33TicketActionMiner && action.GetMiner() != nil {
// 		ticketMiner := action.GetMiner()
// 		return ticketMiner.Reward, nil
// 	}
// 	return 0, nil
// }

// GetName get name
func (ticket *Pos33TicketType) GetName() string {
	return Pos33TicketX
}

// GetTypeMap get type map
func (ticket *Pos33TicketType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Genesis": Pos33TicketActionGenesis,
		"Topen":   Pos33TicketActionOpen,
		"Tbind":   Pos33TicketActionBind,
		"Tclose":  Pos33TicketActionClose,
		"Miner":   Pos33TicketActionMiner,
	}
}

// Pos33TicketMinerParam is ...
type Pos33TicketMinerParam struct {
	CoinDevFund              int64
	CoinReward               int64
	FutureBlockTime          int64
	Pos33TicketPrice         int64
	Pos33TicketFrozenTime    int64
	Pos33TicketWithdrawTime  int64
	Pos33TicketMinerWaitTime int64
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	RetargetAdjustmentFactor int64
}

// GetPos33TicketMinerParam 获取ticket miner config params
func GetPos33TicketMinerParam(cfg *types.Chain33Config, height int64) *Pos33TicketMinerParam {
	conf := types.Conf(cfg, "mver.consensus.pos33")
	c := &Pos33TicketMinerParam{}
	c.CoinDevFund = conf.MGInt("coinDevFund", height) * types.Coin
	c.CoinReward = conf.MGInt("coinReward", height) * types.Coin
	c.FutureBlockTime = conf.MGInt("futureBlockTime", height)
	c.Pos33TicketPrice = conf.MGInt("ticketPrice", height) * types.Coin
	c.Pos33TicketFrozenTime = conf.MGInt("ticketFrozenTime", height)
	c.Pos33TicketWithdrawTime = conf.MGInt("ticketWithdrawTime", height)
	c.Pos33TicketMinerWaitTime = conf.MGInt("ticketMinerWaitTime", height)
	c.TargetTimespan = time.Duration(conf.MGInt("targetTimespan", height)) * time.Second
	c.TargetTimePerBlock = time.Duration(conf.MGInt("targetTimePerBlock", height)) * time.Second
	c.RetargetAdjustmentFactor = conf.MGInt("retargetAdjustmentFactor", height)
	return c
}

// Pos33AllTicketCountKeyPrefix for query all ticket count
const Pos33AllTicketCountKeyPrefix = "LODB-pos33-all:"

const (
	// Pos33BlockReward 区块奖励
	Pos33BlockReward = types.Coin * 15
	// Pos33SortitionSize 多少区块做一次抽签
	Pos33SortitionSize = 10
	// Pos33VoteReward 每ticket区块voter奖励
	Pos33VoteReward = types.Coin * 7 / 10
	// Pos33BpReward 每ticket区块bp奖励
	Pos33BpReward = types.Coin * 2 / 10
	// Pos33ProposerSize 候选区块Proposer数量
	Pos33ProposerSize = 7
	// Pos33VoterSize  候选区块Voter数量
	Pos33VoterSize = 15
	// Pos33MustVotes 必须达到的票数
	Pos33MustVotes = 8 //15*2/3 + 1
	// Pos33MinVotes 必须达到的票数
	Pos33MinVotes = 5
)

// Verify is verify vote msg
func (v *Pos33VoteMsg) Verify() bool {
	s := v.Sig
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	v.Sig = s
	return types.CheckSign(b, "", s)
}

// Equal is ...
func (v *Pos33VoteMsg) Equal(other *Pos33VoteMsg) bool {
	return v.Sort.Proof.Input.Height == other.Sort.Proof.Input.Height &&
		v.Sort.Proof.Input.Round == other.Sort.Proof.Input.Round &&
		v.Sort.SortHash.Tid == other.Sort.SortHash.Tid &&
		v.Tid == other.Tid
}

// Sign is sign vote msg
func (v *Pos33VoteMsg) Sign(priv crypto.PrivKey) {
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	sig := priv.Sign(b)
	v.Sig = &types.Signature{Ty: types.SECP256K1, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
}

func (m *Pos33BlockMsg) Sign(priv crypto.PrivKey) {
	m.Sig = nil
	b := crypto.Sha256(types.Encode(m))
	sig := priv.Sign(b)
	m.Sig = &types.Signature{Ty: types.SECP256K1, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
}

func (m *Pos33BlockMsg) Verify() bool {
	s := m.Sig
	m.Sig = nil
	b := crypto.Sha256(types.Encode(m))
	m.Sig = s
	return types.CheckSign(b, "", s)
}

/*
// ToString is rands to string
func (m *Pos33Rands) ToString() string {
	s := ""
	for _, r := range m.Rands {
		s += hex.EncodeToString(r.Hash) + " "
	}
	return s
}
*/

// ToString is reword to string
func (act *Pos33Miner) ToString() string {
	b, err := json.MarshalIndent(act, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// Sorts is for sort []*Pos33SortitionMsg
type Sorts []*Pos33SortMsg

func (m Sorts) Len() int { return len(m) }
func (m Sorts) Less(i, j int) bool {
	return string(m[i].SortHash.Hash) < string(m[j].SortHash.Hash)
}
func (m Sorts) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

// Votes for sort
type Votes []*Pos33VoteMsg

func (v Votes) Len() int { return len(v) }
func (v Votes) Less(i, j int) bool {
	return string(v[i].Sort.SortHash.Hash) < string(v[i].Sort.SortHash.Hash)
}
func (v Votes) Swap(i, j int) { v[i], v[j] = v[j], v[i] }

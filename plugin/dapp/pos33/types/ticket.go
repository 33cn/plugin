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

	//TyLogNewTicket new ticket log type
	TyLogNewTicket = 111
	// TyLogCloseTicket close ticket log type
	TyLogCloseTicket = 112
	// TyLogMinerTicket miner ticket log type
	TyLogMinerTicket = 113
	// TyLogTicketBind bind ticket log type
	TyLogTicketBind = 114
)

//ticket
const (
	// TicketActionGenesis action type
	TicketActionGenesis = 11
	// TicketActionOpen action type
	TicketActionOpen = 12
	// TicketActionClose action type
	TicketActionClose = 13
	// TicketActionList  action type
	TicketActionList = 14 //读的接口不直接经过transaction
	// TicketActionInfos action type
	TicketActionInfos = 15 //读的接口不直接经过transaction
	// TicketActionMiner action miner
	TicketActionMiner = 16
	// TicketActionBind action bind
	TicketActionBind = 17
)

// TicketOldParts old tick type
const TicketOldParts = 3

// TicketCountOpenOnce count open once
const TicketCountOpenOnce = 1000

// ErrOpenTicketPubHash err type
var ErrOpenTicketPubHash = errors.New("ErrOpenTicketPubHash")

// TicketX dapp name
var TicketX = "pos33"

func init() {
	types.AllowUserExec = append(types.AllowUserExec, []byte(TicketX))
	types.RegFork(TicketX, InitFork)
	types.RegExec(TicketX, InitExecutor)

}

func InitFork(cfg *types.Chain33Config) {
	cfg.RegisterDappFork(TicketX, "Enable", 0)
	//cfg.RegisterDappFork(TicketX, "ForkTicketId", 1062000)
	//cfg.RegisterDappFork(TicketX, "ForkTicketVrf", 1770000)
}

func InitExecutor(cfg *types.Chain33Config) {
	types.RegistorExecutor(TicketX, NewType(cfg))
}

// TicketType ticket exec type
type TicketType struct {
	types.ExecTypeBase
}

// NewType new type
func NewType(cfg *types.Chain33Config) *TicketType {
	c := &TicketType{}
	c.SetChild(c)
	c.SetConfig(cfg)
	return c
}

// GetPayload get payload
func (ticket *TicketType) GetPayload() types.Message {
	return &TicketAction{}
}

// GetLogMap get log map
func (ticket *TicketType) GetLogMap() map[int64]*types.LogInfo {
	return map[int64]*types.LogInfo{
		TyLogNewTicket:   {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogNewTicket"},
		TyLogCloseTicket: {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogCloseTicket"},
		TyLogMinerTicket: {Ty: reflect.TypeOf(ReceiptTicket{}), Name: "LogMinerTicket"},
		TyLogTicketBind:  {Ty: reflect.TypeOf(ReceiptTicketBind{}), Name: "LogTicketBind"},
	}
}

// Amount get amount
func (ticket TicketType) Amount(tx *types.Transaction) (int64, error) {
	var action TicketAction
	err := types.Decode(tx.GetPayload(), &action)
	if err != nil {
		return 0, types.ErrDecode
	}
	if action.Ty == TicketActionMiner && action.GetMiner() != nil {
		ticketMiner := action.GetMiner()
		return ticketMiner.Reward, nil
	}
	return 0, nil
}

// GetName get name
func (ticket *TicketType) GetName() string {
	return TicketX
}

// GetTypeMap get type map
func (ticket *TicketType) GetTypeMap() map[string]int32 {
	return map[string]int32{
		"Genesis": TicketActionGenesis,
		"Topen":   TicketActionOpen,
		"Tbind":   TicketActionBind,
		"Tclose":  TicketActionClose,
		"Miner":   TicketActionMiner,
	}
}

// TicketMinerParam is ...
type TicketMinerParam struct {
	CoinDevFund              int64
	CoinReward               int64
	FutureBlockTime          int64
	TicketPrice              int64
	TicketFrozenTime         int64
	TicketWithdrawTime       int64
	TicketMinerWaitTime      int64
	TargetTimespan           time.Duration
	TargetTimePerBlock       time.Duration
	RetargetAdjustmentFactor int64
}

// GetTicketMinerParam 获取ticket miner config params
func GetTicketMinerParam(cfg *types.Chain33Config, height int64) *TicketMinerParam {
	conf := types.Conf(cfg, "mver.consensus.ticket")
	c := &TicketMinerParam{}
	c.CoinDevFund = conf.MGInt("coinDevFund", height) * types.Coin
	c.CoinReward = conf.MGInt("coinReward", height) * types.Coin
	c.FutureBlockTime = conf.MGInt("futureBlockTime", height)
	c.TicketPrice = conf.MGInt("ticketPrice", height) * types.Coin
	c.TicketFrozenTime = conf.MGInt("ticketFrozenTime", height)
	c.TicketWithdrawTime = conf.MGInt("ticketWithdrawTime", height)
	c.TicketMinerWaitTime = conf.MGInt("ticketMinerWaitTime", height)
	c.TargetTimespan = time.Duration(conf.MGInt("targetTimespan", height)) * time.Second
	c.TargetTimePerBlock = time.Duration(conf.MGInt("targetTimePerBlock", height)) * time.Second
	c.RetargetAdjustmentFactor = conf.MGInt("retargetAdjustmentFactor", height)
	return c
}

// Pos33AllTicketCountKeyPrefix for query all ticket count
const Pos33AllTicketCountKeyPrefix = "LODB-ticket-all:"

const (
	// Pos33MinDeposit 抵押的最小单位
	Pos33MinDeposit = types.Coin * 10000
	// Pos33BlockReward 区块奖励
	Pos33BlockReward = types.Coin * 15
	// Pos33SortitionSize 多少区块做一次抽签
	Pos33SortitionSize = 10
	// Pos33VoteReward 每个区块的奖励
	Pos33VoteReward = types.Coin / 2
	// Pos33ProposerSize 候选区块Proposer数量
	Pos33ProposerSize = 7
	// Pos33VoterSize  候选区块Voter数量
	Pos33VoterSize = 10
	// Pos33DepositPeriod 抵押周期
	Pos33DepositPeriod = 40320
	// Pos33FundKeyAddr ycc开发基金地址
	Pos33FundKeyAddr = "1DvAFGqS26Recz22yeoHcovzxN7dUh92ZY"
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
	h1 := crypto.Sha256(types.Encode(v))
	h2 := crypto.Sha256(types.Encode(other))
	return string(h1) == string(h2)
}

// Sign is sign vote msg
func (v *Pos33VoteMsg) Sign(priv crypto.PrivKey) {
	v.Sig = nil
	b := crypto.Sha256(types.Encode(v))
	sig := priv.Sign(b)
	v.Sig = &types.Signature{Ty: types.ED25519, Pubkey: priv.PubKey().Bytes(), Signature: sig.Bytes()}
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
type Sorts []*Pos33SortitionMsg

func (m Sorts) Len() int { return len(m) }
func (m Sorts) Less(i, j int) bool {
	return string(m[i].Hash) < string(m[j].Hash)
}
func (m Sorts) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

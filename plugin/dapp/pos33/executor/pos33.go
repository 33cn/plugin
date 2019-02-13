package executor

import (
	"errors"
	"strconv"

	"github.com/33cn/chain33/common/address"
	log "github.com/33cn/chain33/common/log/log15"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
)

var plog = log.New("module", "exec.pos33")

// Init to register pos33 driver
func Init(name string, sub []byte) {
	drivers.Register(newPos33().GetDriverName(), newPos33, 0)
}

// Pos33 is the pos33 executor
type Pos33 struct {
	drivers.DriverBase
}

func newPos33() drivers.Driver {
	p := &Pos33{}
	p.SetChild(p)
	return p
}

// GetDriverName return pos33 name
func (p *Pos33) GetDriverName() string {
	return "pos33"
}

func (p *Pos33) handleDepositAction(tx *types.Transaction, index int, act *ty.Pos33DepositAction) (*types.Receipt, error) {
	r, err := p.GetCoinsAccount().ExecDepositFrozen(tx.From(), drivers.ExecAddress(p.GetDriverName()), ty.Pos33Miner*act.W)
	if err != nil {
		panic(err)
	}
	return r, err
}

func (p *Pos33) handleWithdrawAction(tx *types.Transaction, index int, act *ty.Pos33WithdrawAction) (*types.Receipt, error) {
	return p.GetCoinsAccount().ExecActive(tx.From(), drivers.ExecAddress(p.GetDriverName()), ty.Pos33Miner*act.W)
}

func (p *Pos33) handleDelegateAction(tx *types.Transaction, index int, act *ty.Pos33DelegateAction) (*types.Receipt, error) {
	return nil, nil
}

//
func (p *Pos33) handleRewordAction(tx *types.Transaction, index int, act *ty.Pos33RewordAction) (*types.Receipt, error) {
	sumw := 0
	for i, v := range act.Votes {
		w := int(v.Weight)
		sumw += w
		if sumw > int(ty.Pos33BlockReword/types.Coin) {
			act.Votes = act.Votes[:i]
			sumw -= w
			break
		}
	}

	db := p.GetCoinsAccount()
	var kvs []*types.KeyValue

	const vr = ty.Pos33VoteReword
	bpReword := vr * int64(sumw)
	bp := address.PubKeyToAddress(tx.Signature.Pubkey).String()
	bpAcc := db.LoadAccount(bp)
	bpAcc.Balance += int64(bpReword)

	for _, v := range act.Votes {
		addr := address.PubKeyToAddress(v.Sig.Pubkey).String()
		acc := db.LoadAccount(addr)
		acc.Balance += vr * int64(v.Weight)
		kvs = append(kvs, db.GetKVSet(acc)...)
		plog.Info("block reword", "voter", addr, "voter reword", vr*int64(v.Weight))
	}
	facc := db.LoadAccount(ty.Pos33FundKeyAddr)
	fr := ty.Pos33BlockReword - types.Coin*int64(sumw)
	facc.Balance += fr
	kvs = append(kvs, db.GetKVSet(facc)...)

	plog.Info("block reword", "bp", bp, "bp reword", bpReword, "fund reword", fr)
	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}

func (p *Pos33) handlePunishAction(tx *types.Transaction, index int, act *ty.Pos33PunishAction) (*types.Receipt, error) {
	db := p.GetCoinsAccount()
	var kvs []*types.KeyValue
	for who := range act.Punishs {
		frozen := db.LoadAccount(who).Frozen
		_, err := db.ExecActive(who, drivers.ExecAddress(p.GetDriverName()), frozen)
		if err != nil {
			return nil, err
		}
		acc := db.LoadAccount(who)
		acc.Balance -= frozen
		kvs = append(kvs, db.GetKVSet(acc)...)
	}

	return &types.Receipt{Ty: types.ExecOk, KV: kvs}, nil
}

// Exec execute the tx and modify the state of the block chain
func (p *Pos33) Exec(tx *types.Transaction, index int) (*types.Receipt, error) {
	var pa ty.Pos33Action
	err := types.Decode(tx.Payload, &pa)
	if err != nil {
		return nil, err
	}
	var rt *types.Receipt
	switch pa.Value.(type) {
	case *ty.Pos33Action_Deposit:
		rt, err = p.handleDepositAction(tx, index, pa.GetDeposit())
	case *ty.Pos33Action_Withdraw:
		rt, err = p.handleWithdrawAction(tx, index, pa.GetWithdraw())
	case *ty.Pos33Action_Delegate:
		rt, err = p.handleDelegateAction(tx, index, pa.GetDelegate())
	case *ty.Pos33Action_Reword:
		rt, err = p.handleRewordAction(tx, index, pa.GetReword())
	case *ty.Pos33Action_Punish:
		rt, err = p.handlePunishAction(tx, index, pa.GetPunish())
	default:
		err = errors.New("action type NOT support")
	}

	if err != nil {
		return nil, err
	}

	return rt, nil
}

// ExecLocal execute the tx and modify local state for block chain
func (p *Pos33) ExecLocal(tx *types.Transaction, receipt *types.ReceiptData, index int) (*types.LocalDBSet, error) {
	var pa ty.Pos33Action
	err := types.Decode(tx.Payload, &pa)
	if err != nil {
		return nil, err
	}
	allw := p.getAllWeight()
	var kvs []*types.KeyValue
	switch pa.Value.(type) {
	case *ty.Pos33Action_Deposit:
		w := int(pa.GetDeposit().W)
		kvs = append(kvs, p.addWeight(tx.From(), w))
		kvs = append(kvs, p.setAllWeight(allw+w))
	case *ty.Pos33Action_Withdraw:
		w := int(pa.GetDeposit().W)
		kvs = append(kvs, p.addWeight(tx.From(), -w))
		kvs = append(kvs, p.setAllWeight(allw-w))
	default:
		return nil, nil
	}
	return &types.LocalDBSet{KV: kvs}, nil
}

func (p *Pos33) setAllWeight(w int) *types.KeyValue {
	k := []byte(ty.Pos33AllWeight)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getAllWeight() int {
	k := []byte(ty.Pos33AllWeight)
	val, err := p.GetLocalDB().Get(k)
	if err != nil {
		return 0
	}

	w, err := strconv.Atoi(string(val))
	if err != nil {
		panic(err)
	}
	return w
}

func (p *Pos33) addWeight(addr string, w int) *types.KeyValue {
	w += p.getWeight(addr)
	k := []byte(ty.Pos33Weight + addr)
	v := []byte(strconv.Itoa(w))
	p.GetLocalDB().Set(k, v)
	return &types.KeyValue{Key: k, Value: v}
}

func (p *Pos33) getWeight(addr string) int {
	val, err := p.GetLocalDB().Get([]byte(ty.Pos33Weight + addr))
	if err != nil {
		return 0
	}

	w, err := strconv.Atoi(string(val))
	if err != nil {
		panic(err)
	}
	return w
}

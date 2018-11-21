// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

//database opeartion for execs hashlock
import (
	"encoding/json"
	"fmt"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

var hlog = log.New("module", "hashlock.db")

const (
	hashlockLocked   = 1
	hashlockUnlocked = 2
	hashlockSent     = 3
)

// DB struct
type DB struct {
	pty.Hashlock
}

// NewDB instance
func NewDB(id []byte, returnWallet string, toAddress string, blocktime int64, amount int64, time int64) *DB {
	h := &DB{}
	h.HashlockId = id
	h.ReturnAddress = returnWallet
	h.ToAddress = toAddress
	h.CreateTime = blocktime
	h.Status = hashlockLocked
	h.Amount = amount
	h.Frozentime = time
	return h
}

// GetKVSet for hashlock
func (h *DB) GetKVSet() (kvset []*types.KeyValue) {
	value := types.Encode(&h.Hashlock)

	kvset = append(kvset, &types.KeyValue{Key: Key(h.HashlockId), Value: value})
	return kvset
}

// Save KV
func (h *DB) Save(db dbm.KV) {
	set := h.GetKVSet()
	for i := 0; i < len(set); i++ {
		db.Set(set[i].GetKey(), set[i].Value)
	}
}

// Key for hashlock
func Key(id []byte) (key []byte) {
	key = append(key, []byte("mavl-hashlock-")...)
	key = append(key, id...)
	return key
}

// Action def
type Action struct {
	coinsAccount *account.DB
	db           dbm.KV
	txhash       []byte
	fromaddr     string
	blocktime    int64
	height       int64
	execaddr     string
}

// NewAction gen action instance
func NewAction(h *Hashlock, tx *types.Transaction, execaddr string) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{h.GetCoinsAccount(), h.GetStateDB(), hash, fromaddr, h.GetBlockTime(), h.GetHeight(), execaddr}
}

// Hashlocklock Action
func (action *Action) Hashlocklock(hlock *pty.HashlockLock) (*types.Receipt, error) {

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	//不存在相同的hashlock，假定采用sha256
	//_, err := readHashlock(action.db, hlock.Hash)
	_, err := readHashlock(action.db, common.Sha256(hlock.Hash))
	if err != types.ErrNotFound {
		hlog.Error("Hashlocklock", "hlock.Hash repeated", hlock.Hash)
		return nil, pty.ErrHashlockReapeathash
	}

	h := NewDB(hlock.Hash, action.fromaddr, hlock.ToAddress, action.blocktime, hlock.Amount, hlock.Time)
	//冻结子账户资金
	receipt, err := action.coinsAccount.ExecFrozen(action.fromaddr, action.execaddr, hlock.Amount)

	if err != nil {
		hlog.Error("Hashlocklock.Frozen", "addr", action.fromaddr, "execaddr", action.execaddr, "amount", hlock.Amount)
		return nil, err
	}

	h.Save(action.db)
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	//logs = append(logs, h.GetReceiptLog())
	kv = append(kv, h.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// Hashlockunlock Action
func (action *Action) Hashlockunlock(unlock *pty.HashlockUnlock) (*types.Receipt, error) {

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	hash, err := readHashlock(action.db, common.Sha256(unlock.Secret))
	if err != nil {
		hlog.Error("Hashlockunlock", "unlock.Secret", unlock.Secret)
		return nil, err
	}

	if hash.ReturnAddress != action.fromaddr {
		hlog.Error("Hashlockunlock.Frozen", "action.fromaddr", action.fromaddr)
		return nil, pty.ErrHashlockReturnAddrss
	}

	if hash.Status != hashlockLocked {
		hlog.Error("Hashlockunlock", "hash.Status", hash.Status)
		return nil, pty.ErrHashlockStatus
	}

	if action.blocktime-hash.GetCreateTime() < hash.Frozentime {
		hlog.Error("Hashlockunlock", "action.blocktime-hash.GetCreateTime", action.blocktime-hash.GetCreateTime())
		return nil, pty.ErrTime
	}

	//different with typedef in C
	h := &DB{*hash}
	receipt, errR := action.coinsAccount.ExecActive(h.ReturnAddress, action.execaddr, h.Amount)
	if errR != nil {
		hlog.Error("ExecActive error", "ReturnAddress", h.ReturnAddress, "execaddr", action.execaddr, "amount", h.Amount)
		return nil, errR
	}

	h.Status = hashlockUnlocked
	h.Save(action.db)
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	//logs = append(logs, t.GetReceiptLog())
	kv = append(kv, h.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

// Hashlocksend Action
func (action *Action) Hashlocksend(send *pty.HashlockSend) (*types.Receipt, error) {

	var logs []*types.ReceiptLog
	var kv []*types.KeyValue

	hash, err := readHashlock(action.db, common.Sha256(send.Secret))
	if err != nil {
		hlog.Error("Hashlocksend", "send.Secret", send.Secret)
		return nil, err
	}

	if hash.Status != hashlockLocked {
		hlog.Error("Hashlocksend", "hash.Status", hash.Status)
		return nil, pty.ErrHashlockStatus
	}

	if action.fromaddr != hash.ToAddress {
		hlog.Error("Hashlocksend", "action.fromaddr", action.fromaddr, "hash.ToAddress", hash.ToAddress)
		return nil, pty.ErrHashlockSendAddress
	}

	if action.blocktime-hash.GetCreateTime() > hash.Frozentime {
		hlog.Error("Hashlocksend", "action.blocktime-hash.GetCreateTime", action.blocktime-hash.GetCreateTime())
		return nil, pty.ErrTime
	}

	//different with typedef in C
	h := &DB{*hash}
	receipt, errR := action.coinsAccount.ExecTransferFrozen(h.ReturnAddress, h.ToAddress, action.execaddr, h.Amount)
	if errR != nil {
		hlog.Error("ExecTransferFrozen error", "ReturnAddress", h.ReturnAddress, "ToAddress", h.ToAddress, "execaddr", action.execaddr, "amount", h.Amount)
		return nil, errR
	}
	h.Status = hashlockSent
	h.Save(action.db)
	logs = append(logs, receipt.Logs...)
	kv = append(kv, receipt.KV...)
	//logs = append(logs, t.GetReceiptLog())
	kv = append(kv, h.GetKVSet()...)

	receipt = &types.Receipt{Ty: types.ExecOk, KV: kv, Logs: logs}
	return receipt, nil
}

func readHashlock(db dbm.KV, id []byte) (*pty.Hashlock, error) {
	data, err := db.Get(Key(id))
	if err != nil {
		return nil, err
	}
	var hashlock pty.Hashlock
	//decode
	err = types.Decode(data, &hashlock)
	if err != nil {
		return nil, err
	}
	return &hashlock, nil
}

// NewHashlockquery gen query instance
func NewHashlockquery() *pty.Hashlockquery {
	q := pty.Hashlockquery{}
	return &q
}

func calcHashlockIDKey(id []byte) []byte {
	return append([]byte("LODB-hashlock-"), id...)
}

// GeHashReciverKV gen KV
func GeHashReciverKV(hashlockID []byte, information *pty.Hashlockquery) *types.KeyValue {
	clog.Error("GeHashReciverKV action")
	infor := pty.Hashlockquery{Time: information.Time, Status: information.Status, Amount: information.Amount, CreateTime: information.CreateTime, CurrentTime: information.CurrentTime}
	clog.Error("GeHashReciverKV action", "Status", information.Status)
	reciver, err := json.Marshal(infor)
	if err == nil {
		fmt.Println("成功转换为json格式")
	} else {
		fmt.Println(err)
	}
	clog.Error("GeHashReciverKV action", "reciver", reciver)
	kv := &types.KeyValue{Key: calcHashlockIDKey(hashlockID), Value: reciver}
	clog.Error("GeHashReciverKV action", "kv", kv)
	return kv
}

// GetHashReciver get hashlock
func GetHashReciver(db dbm.KVDB, hashlockID []byte) (*pty.Hashlockquery, error) {
	//reciver := types.Int64{}
	clog.Error("GetHashReciver action", "hashlockID", hashlockID)
	reciver := NewHashlockquery()
	hashReciver, err := db.Get(hashlockID)
	if err != nil {
		clog.Error("Get err")
		return reciver, err
	}
	fmt.Println(hashReciver)
	if hashReciver == nil {
		clog.Error("nilnilnilllllllllll")

	}
	clog.Error("hashReciver", "len", len(hashReciver))
	clog.Error("GetHashReciver", "hashReciver", hashReciver)
	err = json.Unmarshal(hashReciver, reciver)
	if err != nil {
		clog.Error("hashReciver Unmarshal")
		return nil, err
	}
	clog.Error("GetHashReciver", "reciver", reciver)
	return reciver, nil
}

// SetHashReciver save hashlock
func SetHashReciver(db dbm.KVDB, hashlockID []byte, information *pty.Hashlockquery) error {
	clog.Error("SetHashReciver action")
	kv := GeHashReciverKV(hashlockID, information)
	return db.Set(kv.Key, kv.Value)
}

// UpdateHashReciver update status for hashlock
func UpdateHashReciver(cachedb dbm.KVDB, hashlockID []byte, information pty.Hashlockquery) (*types.KeyValue, error) {
	clog.Error("UpdateHashReciver", "hashlockId", hashlockID)
	recv, err := GetHashReciver(cachedb, hashlockID)
	if err != nil && err != types.ErrNotFound {
		clog.Error("UpdateHashReciver", "err", err)
		return nil, err
	}
	fmt.Println(recv)
	clog.Error("UpdateHashReciver", "recv", recv)
	//	clog.Error("UpdateHashReciver", "Status", information.Status)
	//	var action types.Action
	//当处于lock状态时，在db中是找不到的，此时需要创建并存储于db中，其他状态则能从db中找到
	if information.Status == hashlockLocked {
		clog.Error("UpdateHashReciver", "Hashlock_Locked", hashlockLocked)
		if err == types.ErrNotFound {
			clog.Error("UpdateHashReciver", "Hashlock_Locked")
			recv.Time = information.Time
			recv.Status = hashlockLocked //1
			recv.Amount = information.Amount
			recv.CreateTime = information.CreateTime
			//			clog.Error("UpdateHashReciver", "Statuslock", recv.Status)
			clog.Error("UpdateHashReciver", "recv", recv)
		}
	} else if information.Status == hashlockUnlocked {
		clog.Error("UpdateHashReciver", "Hashlock_Unlocked", hashlockUnlocked)
		if err == nil {
			recv.Status = hashlockUnlocked //2
			//			clog.Error("UpdateHashReciver", "Statusunlock", recv.Status)
			clog.Error("UpdateHashReciver", "recv", recv)
		}
	} else if information.Status == hashlockSent {
		clog.Error("UpdateHashReciver", "Hashlock_Sent", hashlockSent)
		if err == nil {
			recv.Status = hashlockSent //3
			//			clog.Error("UpdateHashReciver", "Statussend", recv.Status)
			clog.Error("UpdateHashReciver", "recv", recv)
		}
	}
	SetHashReciver(cachedb, hashlockID, recv)
	//keyvalue
	return GeHashReciverKV(hashlockID, recv), nil
}

// GetTxsByHashlockID get hashlock record
func (n *Hashlock) GetTxsByHashlockID(hashlockID []byte, differTime int64) (types.Message, error) {
	clog.Error("GetTxsByHashlockId action")
	db := n.GetLocalDB()
	query, err := GetHashReciver(db, hashlockID)
	if err != nil {
		return nil, err
	}
	query.CurrentTime = differTime
	//	qresult := types.Hashlockquery{query.Time, query.Status, query.Amount, query.CreateTime, currentTime}
	return query, nil
}

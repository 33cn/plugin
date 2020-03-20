package executor

import (
	"fmt"
	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

// Action action struct
type Action struct {
	statedb   dbm.KV
	txhash    []byte
	fromaddr  string
	blocktime int64
	height    int64
	execaddr  string
	localDB   dbm.KVDB
	index     int
	api       client.QueueProtocolAPI
}

func NewAction(e accountmanager, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{e.GetStateDB(), hash, fromaddr,
		e.GetBlockTime(), e.GetHeight(), dapp.ExecAddress(string(tx.Execer)), e.GetLocalDB(), index, e.GetAPI()}
}

//GetIndex get index
func (a *Action) GetIndex() int64 {
	return (a.height*types.MaxTxsPerBlock + int64(a.index)) * 1e4
}

//GetKVSet get kv set
func (a *Action) GetKVSet(account *et.Account) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: calcAccountKey(account.AccountID), Value: types.Encode(account)})
	return kvset
}

func (a *Action) Register(payload *et.Register) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	account, err := queryMarketDepth(a.localDB, payload.AccountID)
	if err == nil && account != nil {
		return nil, et.ErrAccountNameExist
	}
	//TODO 有效期后面统一配置目前暂定五年时间

	re := &et.Receipt{
		AccountID:  payload.AccountID,
		Addr:       a.fromaddr,
		Index:      a.GetIndex(),
		Status:     et.Normal,
		CreateTime: a.blocktime,
		ExpireTime: a.blocktime + 5*360*24*3600,
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyRegisterLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: logs}
	return receipts, nil
}

//为了避免别人恶意重置别人的帐号,这个操作仅有系统管理员有权限去操作
func (a *Action) ReSet(payload *et.Reset) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	account, err := queryMarketDepth(a.localDB, payload.AccountID)
	if err != nil {
		return nil, et.ErrAccountNameNotExist
	}
	//TODO 重置公钥锁定期暂定15天，后面可以由管理员去配置
	re := &et.Receipt{
		AccountID:  account.AccountID,
		PrevAddr:   account.Addr,
		Addr:       payload.Addr,
		Index:      account.Index,
		Status:     et.Locked,
		CreateTime: account.CreateTime,
		ExpireTime: account.ExpireTime,
		LockTime:   a.blocktime + 15*24*3600,
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyRegisterLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: logs}
	return receipts, nil
}
func(a *Action) Transfer(payload *et.Transfer)(*types.Receipt, error){
	cfg := a.api.GetConfig()
	acc, err := account.NewAccountDB(cfg, payload.Asset.GetExecer(), payload.Asset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
}
func queryMarketDepth(localdb dbm.KV, accountName string) (*et.Account, error) {
	table := NewAccountTable(localdb)
	primaryKey := []byte(fmt.Sprintf("%s", accountName))
	row, err := table.GetData(primaryKey)
	if err != nil {
		return nil, err
	}
	return row.Data.(*et.Account), nil
}

package executor

import (
	"fmt"
	"strconv"
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	dbm "github.com/33cn/chain33/common/db"
	tab "github.com/33cn/chain33/common/db/table"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
)

var (
	// ConfNameActiveTime 有效期
	ConfNameActiveTime = et.AccountmanagerX + "-" + "activeTime"
	// ConfNameLockTime 密钥重置锁定期
	ConfNameLockTime = et.AccountmanagerX + "-" + "lockTime"
	// ConfNameManagerAddr 管理员地址
	ConfNameManagerAddr = et.AccountmanagerX + "-" + "managerAddr"
	// DefaultActiveTime 默认有效期
	DefaultActiveTime = int64(5 * 360 * 24 * 3600)
	// DefaultLockTime 默认密钥重置锁定期
	DefaultLockTime = int64(15 * 24 * 3600)
	// DefaultManagerAddr 默认管理员地址
	DefaultManagerAddr = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
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

//NewAction ...
func NewAction(e *Accountmanager, tx *types.Transaction, index int) *Action {
	hash := tx.Hash()
	fromaddr := tx.From()
	return &Action{e.GetStateDB(), hash, fromaddr,
		e.GetBlockTime(), e.GetHeight(), dapp.ExecAddress(string(tx.Execer)), e.GetLocalDB(), index, e.GetAPI()}
}

//GetIndex get index 主键索引,实际上是以过期时间为主键
func (a *Action) GetIndex() int64 {
	return a.blocktime*types.MaxTxsPerBlock + int64(a.index)
}

//GetKVSet get kv set
func (a *Action) GetKVSet(account *et.Account) (kvset []*types.KeyValue) {
	kvset = append(kvset, &types.KeyValue{Key: calcAccountKey(account.AccountID), Value: types.Encode(account)})
	return kvset
}

//Register ...
func (a *Action) Register(payload *et.Register) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	account1, err := findAccountByID(a.localDB, payload.AccountID)
	if err == nil && account1 != nil {
		return nil, et.ErrAccountIDExist
	}

	//默认有效期时五年
	cfg := a.api.GetConfig()
	defaultActiveTime := getConfValue(cfg, a.statedb, ConfNameActiveTime, DefaultActiveTime)
	account := &et.Account{
		AccountID:  payload.AccountID,
		Addr:       a.fromaddr,
		PrevAddr:   "",
		Status:     et.Normal,
		Level:      et.Normal,
		CreateTime: a.blocktime,
		ExpireTime: a.blocktime + defaultActiveTime,
		LockTime:   0,
		Index:      a.GetIndex(),
	}
	re := &et.AccountReceipt{
		Account: account,
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyRegisterLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: logs}
	return receipts, nil
}

//Reset 为了避免别人恶意重置别人的帐号,这个操作仅有系统管理员有权限去操作
func (a *Action) Reset(payload *et.ResetKey) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	cfg := a.api.GetConfig()
	managerAddr := getManagerAddr(cfg, a.statedb, ConfNameManagerAddr, DefaultManagerAddr)
	if managerAddr != a.fromaddr {
		return nil, et.ErrNotAdmin
	}
	account, err := findAccountByID(a.localDB, payload.AccountID)
	if err != nil {
		return nil, et.ErrAccountIDNotExist
	}
	//重置公钥锁定期暂定15天,可以由管理员去配置
	defaultLockTime := getConfValue(cfg, a.statedb, ConfNameLockTime, DefaultLockTime)
	account.Status = et.Locked
	account.LockTime = a.blocktime + defaultLockTime
	account.PrevAddr = account.Addr
	account.Addr = payload.Addr
	re := &et.AccountReceipt{
		Account: account,
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyResetLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: nil, Logs: logs}
	return receipts, nil
}

//Transfer ...
func (a *Action) Transfer(payload *et.Transfer) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	cfg := a.api.GetConfig()
	account1, err := findAccountByID(a.localDB, payload.FromAccountID)
	if err != nil {
		elog.Error("Transfer", "fromAccountID", payload.FromAccountID, "err", et.ErrAccountIDNotExist)
		return nil, et.ErrAccountIDNotExist
	}
	if account1.Status != et.Normal || account1.Addr != a.fromaddr || account1.ExpireTime <= a.blocktime {
		elog.Error("Transfer", "fromaddr", a.fromaddr, "err", et.ErrAccountIDNotPermiss)
		return nil, et.ErrAccountIDNotPermiss
	}
	//如果prevAddr地址不为空，先查看余额，将该地址下面得资产划转到新得公钥地址下
	if account1.PrevAddr != "" {
		assetDB, err := account.NewAccountDB(cfg, payload.Asset.GetExec(), payload.Asset.GetSymbol(), a.statedb)
		if err != nil {
			return nil, err
		}
		prevAccount := assetDB.LoadExecAccount(account1.PrevAddr, a.execaddr)
		if prevAccount.Balance > 0 {
			receipt, err := assetDB.ExecTransfer(account1.PrevAddr, account1.Addr, a.execaddr, prevAccount.Balance)
			if err != nil {
				return nil, err
			}
			logs = append(logs, receipt.Logs...)
			kvs = append(kvs, receipt.KV...)
		}
	}
	if payload.FromAccountID == payload.ToAccountID {
		re := &et.TransferReceipt{
			FromAccount: account1,
			ToAccount:   account1,
			Index:       a.GetIndex(),
		}
		receiptlog := &types.ReceiptLog{Ty: et.TyTransferLog, Log: types.Encode(re)}
		logs = append(logs, receiptlog)
		receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
		return receipts, nil
	}
	account2, err := findAccountByID(a.localDB, payload.ToAccountID)
	if err != nil {
		elog.Error("Transfer,check to accountID", "toAccountID", payload.ToAccountID, "err", et.ErrAccountIDNotExist)
		return nil, et.ErrAccountIDNotExist
	}
	if account2.Status != et.Normal || account2.ExpireTime <= a.blocktime {
		elog.Error("Transfer", "ToAccountID", account2.AccountID, "err", et.ErrAccountIDNotPermiss)
		return nil, et.ErrAccountIDNotPermiss
	}

	assetDB, err := account.NewAccountDB(cfg, payload.Asset.GetExec(), payload.Asset.GetSymbol(), a.statedb)
	if err != nil {
		return nil, err
	}
	fromAccount := assetDB.LoadExecAccount(a.fromaddr, a.execaddr)
	if fromAccount.Balance < payload.Asset.Amount {
		elog.Error("Transfer, check  balance", "addr", a.fromaddr, "avail", fromAccount.Balance, "need", payload.Asset.Amount)
		return nil, et.ErrAssetBalance
	}
	receipt, err := assetDB.ExecTransfer(account1.Addr, account2.Addr, a.execaddr, payload.Asset.Amount)
	if err != nil {
		return nil, err
	}
	logs = append(logs, receipt.Logs...)
	kvs = append(kvs, receipt.KV...)

	re := &et.TransferReceipt{
		FromAccount: account1,
		ToAccount:   account2,
		Index:       a.GetIndex(),
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyTransferLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//Supervise ...
func (a *Action) Supervise(payload *et.Supervise) (*types.Receipt, error) {
	//鉴权，看一下地址是否时管理员地址
	cfg := a.api.GetConfig()
	managerAddr := getManagerAddr(cfg, a.statedb, ConfNameManagerAddr, DefaultManagerAddr)
	if managerAddr != a.fromaddr {
		return nil, et.ErrNotAdmin
	}
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue
	var re et.SuperviseReceipt
	for _, ID := range payload.AccountIDs {
		accountM, err := findAccountByID(a.localDB, ID)
		if err != nil {
			elog.Error("Supervise", "AccountID", ID, "err", et.ErrAccountIDNotExist)
			return nil, et.ErrAccountIDNotExist
		}
		switch payload.Op {
		case et.Freeze:
			//TODO 冻结操作交给外部其他执行器去控制,处于freeze状态的地址禁止操作
			accountM.Status = et.Frozen

		case et.UnFreeze:
			accountM.Status = et.Normal

		case et.AddExpire:
			cfg := a.api.GetConfig()
			defaultActiveTime := getConfValue(cfg, a.statedb, ConfNameActiveTime, DefaultActiveTime)
			accountM.Status = et.Normal
			accountM.ExpireTime = a.blocktime + defaultActiveTime
		case et.Authorize:
			accountM.Level = payload.Level
		}
		re.Accounts = append(re.Accounts, accountM)
	}
	re.Op = payload.Op
	re.Index = a.GetIndex()
	receiptlog := &types.ReceiptLog{Ty: et.TySuperviseLog, Log: types.Encode(&re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

//Apply ...
func (a *Action) Apply(payload *et.Apply) (*types.Receipt, error) {
	var logs []*types.ReceiptLog
	var kvs []*types.KeyValue

	accountM, err := findAccountByID(a.localDB, payload.AccountID)
	if err != nil {
		elog.Error("Apply", "AccountID", payload.AccountID, "err", et.ErrAccountIDNotExist)
		return nil, et.ErrAccountIDNotExist
	}
	switch payload.Op {
	case et.RevokeReset:
		if accountM.Status != et.Locked || accountM.PrevAddr != a.fromaddr {
			elog.Error("Apply", "fromaddr", a.fromaddr, "err", et.ErrAccountIDNotPermiss)
			return nil, et.ErrAccountIDNotPermiss
		}
		accountM.LockTime = 0
		accountM.Status = et.Normal
		accountM.Addr = a.fromaddr

	case et.EnforceReset:
		if accountM.Status != et.Locked || accountM.Addr != a.fromaddr {
			elog.Error("Apply", "fromaddr", a.fromaddr, "err", et.ErrAccountIDNotPermiss)
			return nil, et.ErrAccountIDNotPermiss
		}
		accountM.LockTime = 0
		accountM.Status = et.Normal
		//TODO 这里只做coins主笔资产得自动划转，token资产转移,放在转transfer中执行 fromAccountID == toAccountID
		cfg := a.api.GetConfig()
		coinsAssetDB, err := account.NewAccountDB(cfg, cfg.GetCoinExec(), cfg.GetCoinSymbol(), a.statedb)
		if err != nil {
			return nil, err
		}
		coinsAccount := coinsAssetDB.LoadExecAccount(accountM.PrevAddr, a.execaddr)

		receipt, err := coinsAssetDB.ExecTransfer(accountM.PrevAddr, accountM.Addr, a.execaddr, coinsAccount.Balance)
		if err != nil {
			elog.Error("Apply ExecTransfer", "AccountID", accountM.AccountID, "err", err)
		}
		logs = append(logs, receipt.Logs...)
		kvs = append(kvs, receipt.KV...)
	}

	re := &et.AccountReceipt{
		Account: accountM,
	}
	receiptlog := &types.ReceiptLog{Ty: et.TyApplyLog, Log: types.Encode(re)}
	logs = append(logs, receiptlog)
	receipts := &types.Receipt{Ty: types.ExecOk, KV: kvs, Logs: logs}
	return receipts, nil
}

func getConfValue(cfg *types.Chain33Config, db dbm.KV, key string, defaultValue int64) int64 {
	var item types.ConfigItem
	value, err := getManageKey(cfg, key, db)
	if err != nil {
		return defaultValue
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			elog.Debug("accountmanager getConfValue", "decode db key:", key, "err", err.Error())
			return defaultValue
		}
	}
	values := item.GetArr().GetValue()
	if len(values) == 0 {
		elog.Debug("accountmanager getConfValue", "can't get value from values arr. key:", key)
		return defaultValue
	}
	//取数组最后一位，作为最新配置项的值
	v, err := strconv.ParseInt(values[len(values)-1], 10, 64)
	if err != nil {
		elog.Debug("accountmanager getConfValue", "Type conversion error:", err.Error())
		return defaultValue
	}
	return v
}
func getManagerAddr(cfg *types.Chain33Config, db dbm.KV, key, defaultValue string) string {
	var item types.ConfigItem
	value, err := getManageKey(cfg, key, db)
	if err != nil {
		return defaultValue
	}
	if value != nil {
		err = types.Decode(value, &item)
		if err != nil {
			elog.Debug("accountmanager getConfValue", "decode db key:", key, "err", err.Error())
			return defaultValue
		}
	}
	values := item.GetArr().GetValue()
	if len(values) == 0 {
		elog.Debug("accountmanager getConfValue", "can't get value from values arr. key:", key)
		return defaultValue
	}
	return values[len(values)-1]
}

func getManageKey(cfg *types.Chain33Config, key string, db dbm.KV) ([]byte, error) {
	manageKey := types.ManageKey(key)
	value, err := db.Get([]byte(manageKey))
	if err != nil {
		if cfg.IsPara() { //平行链只有一种存储方式
			elog.Debug("accountmanager getManage", "can't get value from db,key:", key, "err", err.Error())
			return nil, err
		}
		elog.Debug("accountmanager getManageKey", "get db key", "not found")
		return getConfigKey(key, db)
	}
	return value, nil
}

func getConfigKey(key string, db dbm.KV) ([]byte, error) {
	configKey := types.ConfigKey(key)
	value, err := db.Get([]byte(configKey))
	if err != nil {
		elog.Debug("accountmanager getConfigKey", "can't get value from db,key:", key, "err", err.Error())
		return nil, err
	}
	return value, nil
}

//正序遍历数据，与传入时间进行对比，看是否逾期
func findAccountListByIndex(localdb dbm.KV, expireTime int64, primaryKey string) (*et.ReplyAccountList, error) {
	table := NewAccountTable(localdb)
	var rows []*tab.Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex("index", nil, nil, et.Count, et.ListASC)
	} else {
		rows, err = table.ListIndex("index", nil, []byte(primaryKey), et.Count, et.ListASC)
	}
	if err != nil {
		elog.Error("findAccountListByIndex.", "index", primaryKey, "err", err.Error())
		return nil, err
	}
	var reply et.ReplyAccountList
	for _, row := range rows {
		account := row.Data.(*et.Account)
		if account.ExpireTime > expireTime {
			break
		}
		//状态变成逾期状态
		account.Status = et.Expired
		reply.Accounts = append(reply.Accounts, account)
	}
	//设置主键索引
	if len(rows) == int(et.Count) {
		reply.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &reply, nil
}

func findAccountByID(localdb dbm.KV, accountID string) (*et.Account, error) {
	table := NewAccountTable(localdb)
	prefix := []byte(accountID)
	//第一次查询,默认展示最新得成交记录
	rows, err := table.ListIndex("accountID", prefix, nil, 1, et.ListDESC)
	if err != nil {
		elog.Debug("findAccountByID.", "accountID", accountID, "err", err.Error())
		return nil, err
	}
	for _, row := range rows {
		account := row.Data.(*et.Account)
		return account, nil
	}
	return nil, types.ErrNotFound
}

func findAccountByAddr(localdb dbm.KV, addr string) (*et.Account, error) {
	table := NewAccountTable(localdb)
	prefix := []byte(addr)
	//第一次查询,默认展示最新得成交记录
	rows, err := table.ListIndex("addr", prefix, nil, 1, et.ListDESC)
	if err != nil {
		elog.Error("findAccountByAddr.", "addr", addr, "err", err.Error())
		return nil, err
	}
	for _, row := range rows {
		account := row.Data.(*et.Account)
		return account, nil
	}
	return nil, types.ErrNotFound
}

func findAccountListByStatus(localdb dbm.KV, status, direction int32, primaryKey string) (*et.ReplyAccountList, error) {
	if status == et.Expired {
		return findAccountListByIndex(localdb, time.Now().Unix(), primaryKey)
	}
	table := NewAccountTable(localdb)
	prefix := []byte(fmt.Sprintf("%d", status))

	var rows []*tab.Row
	var err error
	if primaryKey == "" { //第一次查询,默认展示最新得成交记录
		rows, err = table.ListIndex("status", prefix, nil, et.Count, direction)
	} else {
		rows, err = table.ListIndex("status", prefix, []byte(primaryKey), et.Count, direction)
	}
	if err != nil {
		elog.Error("findAccountListByStatus.", "status", status, "err", err.Error())
		return nil, err
	}
	var reply et.ReplyAccountList
	for _, row := range rows {
		account := row.Data.(*et.Account)
		reply.Accounts = append(reply.Accounts, account)
	}
	//设置主键索引
	if len(rows) == int(et.Count) {
		reply.PrimaryKey = string(rows[len(rows)-1].Primary)
	}
	return &reply, nil
}

func queryBalanceByID(statedb, localdb dbm.KV, cfg *types.Chain33Config, execName string, in *et.QueryBalanceByID) (*et.Balance, error) {
	acc, err := findAccountByID(localdb, in.AccountID)
	if err != nil {
		return nil, err
	}
	assetDB, err := account.NewAccountDB(cfg, in.Asset.GetExec(), in.Asset.GetSymbol(), statedb)
	if err != nil {
		return nil, err
	}
	var balance et.Balance
	if acc.PrevAddr != "" {
		prevAccount := assetDB.LoadExecAccount(acc.PrevAddr, dapp.ExecAddress(execName))
		balance.Balance += prevAccount.Balance
		balance.Frozen += prevAccount.Frozen
	}
	currAccount := assetDB.LoadExecAccount(acc.Addr, dapp.ExecAddress(execName))
	balance.Balance += currAccount.Balance
	balance.Frozen += currAccount.Frozen
	return &balance, nil
}

package executor

import (
	"time"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"

	"testing"

	et "github.com/33cn/plugin/plugin/dapp/accountmanager/types"
	"github.com/stretchr/testify/assert"
)

type execEnv struct {
	blockTime   int64
	blockHeight int64
	difficulty  uint64
}

var (
	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	Nodes    = []string{
		"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
		"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
		"1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k",
		"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs",
	}
)

func TestAccountManager(t *testing.T) {
	//环境准备
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	Init(et.AccountmanagerX, cfg, nil)
	total := 100 * types.DefaultCoinPrecision
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[0],
	}
	accountB := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[1],
	}

	accountC := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[2],
	}
	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    Nodes[3],
	}
	_, stateDB, kvdb := util.CreateTestDB()
	//defer util.CloseTestDB(dir, stateDB)
	execAddr := address.ExecAddress(et.AccountmanagerX)

	accA, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accA.SaveExecAccount(execAddr, &accountA)

	accB, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accB.SaveExecAccount(execAddr, &accountB)

	accC, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accC.SaveExecAccount(execAddr, &accountC)

	accD, _ := account.NewAccountDB(cfg, "coins", "bty", stateDB)
	accD.SaveExecAccount(execAddr, &accountD)
	env := &execEnv{
		time.Now().Unix(),
		1,
		1539918074,
	}
	// set config key
	item := &types.ConfigItem{
		Key: "mavl-manage-" + ConfNameActiveTime,
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{"10"}},
		},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	item2 := &types.ConfigItem{
		Key: "mavl-manage-" + ConfNameManagerAddr,
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{Nodes[0]}},
		},
	}
	stateDB.Set([]byte(item2.Key), types.Encode(item2))

	item3 := &types.ConfigItem{
		Key: "mavl-manage-" + ConfNameLockTime,
		Value: &types.ConfigItem_Arr{
			Arr: &types.ArrayConfig{Value: []string{"2"}},
		},
	}
	stateDB.Set([]byte(item3.Key), types.Encode(item3))

	//注册
	tx1, err := CreateRegister(&et.Register{AccountID: "harrylee2015"}, PrivKeyB)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx1)
	assert.Equal(t, err, nil)
	_, err = Exec_QueryAccountByID("harrylee2015", stateDB, kvdb)
	assert.Equal(t, err, nil)
	_, err = Exec_QueryAccountByAddr(Nodes[1], stateDB, kvdb)
	assert.Equal(t, err, nil)
	tx2, err := CreateRegister(&et.Register{AccountID: "harrylee2015"}, PrivKeyC)
	assert.Nil(t, err)
	err = Exec_Block(t, stateDB, kvdb, env, tx2)
	assert.Equal(t, err, et.ErrAccountIDExist)
	tx3, err := CreateRegister(&et.Register{AccountID: "harrylee2020"}, PrivKeyC)
	assert.Nil(t, err)
	Exec_Block(t, stateDB, kvdb, env, tx3)
	//转账
	tx4, err := CreateTransfer(&et.Transfer{FromAccountID: "harrylee2015", ToAccountID: "harrylee2020", Asset: &types.Asset{Exec: "coins", Symbol: "bty", Amount: 1e8}}, PrivKeyB)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx4)
	assert.Equal(t, err, nil)
	//重置公钥
	tx5, err := CreateReset(&et.ResetKey{Addr: "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs", AccountID: "harrylee2015"}, PrivKeyA)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx5)
	assert.Nil(t, err)
	//在锁定期内撤回请求
	tx6, err := CreateApply(&et.Apply{Op: et.RevokeReset, AccountID: "harrylee2015"}, PrivKeyB)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx6)
	assert.Equal(t, err, nil)
	//重置公钥
	tx5, err = CreateReset(&et.ResetKey{Addr: "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs", AccountID: "harrylee2015"}, PrivKeyA)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx5)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	//过了锁定期，申请生效
	tx6, err = CreateApply(&et.Apply{Op: et.EnforceReset, AccountID: "harrylee2015"}, PrivKeyD)
	assert.Equal(t, err, nil)
	err = Exec_Block(t, stateDB, kvdb, env, tx6)
	assert.Nil(t, err)
	tx7, _ := CreateTransfer(&et.Transfer{FromAccountID: "harrylee2015", ToAccountID: "harrylee2015", Asset: &types.Asset{Exec: "coins", Symbol: "bty", Amount: 1e8}}, PrivKeyD)

	err = Exec_Block(t, stateDB, kvdb, env, tx7)
	assert.Equal(t, err, nil)
	balance, err := Exec_QueryBalanceByID(&et.QueryBalanceByID{AccountID: "harrylee2015", Asset: &types.Asset{Symbol: "bty", Exec: "coins"}}, stateDB, kvdb)
	assert.Equal(t, err, nil)
	assert.Equal(t, balance.Balance, 199*types.DefaultCoinPrecision)

	//将某个账户冻结
	tx8, _ := CreateSupervise(&et.Supervise{
		AccountIDs: []string{"harrylee2015"},
		Op:         et.Freeze,
	}, PrivKeyA)
	err = Exec_Block(t, stateDB, kvdb, env, tx8)
	assert.Equal(t, err, nil)
	//根据状态查询
	accounts, err := Exec_QueryAccountsByStatus(et.Frozen, stateDB, kvdb)
	assert.Equal(t, err, nil)
	assert.Equal(t, accounts.Accounts[0].Status, et.Frozen)

	//解冻账户
	tx9, _ := CreateSupervise(&et.Supervise{
		AccountIDs: []string{"harrylee2015"},
		Op:         et.UnFreeze,
	}, PrivKeyA)
	err = Exec_Block(t, stateDB, kvdb, env, tx9)
	assert.Equal(t, err, nil)
	//根据状态查询
	accounts, err = Exec_QueryAccountsByStatus(et.Frozen, stateDB, kvdb)
	assert.NotEqual(t, err, nil)

	//过期账户查询
	time.Sleep(11 * time.Second)
	accounts, err = Exec_QueryExpiredAccounts(time.Now().Unix(), stateDB, kvdb)
	assert.Equal(t, err, nil)
	assert.Equal(t, 2, len(accounts.Accounts))
	//账户延期
	tx10, _ := CreateSupervise(&et.Supervise{
		AccountIDs: []string{"harrylee2015"},
		Op:         et.AddExpire,
	}, PrivKeyA)
	err = Exec_Block(t, stateDB, kvdb, env, tx10)
	assert.Equal(t, err, nil)
	accounts, err = Exec_QueryExpiredAccounts(time.Now().Unix(), stateDB, kvdb)
	assert.Equal(t, err, nil)
	assert.Equal(t, 1, len(accounts.Accounts))
	//账户授权
	tx11, _ := CreateSupervise(&et.Supervise{
		AccountIDs: []string{"harrylee2015"},
		Op:         et.Authorize,
		Level:      2,
	}, PrivKeyA)
	err = Exec_Block(t, stateDB, kvdb, env, tx11)
	assert.Equal(t, err, nil)
	acc, err := Exec_QueryAccountByID("harrylee2015", stateDB, kvdb)
	assert.Equal(t, err, nil)
	assert.Equal(t, acc.Level, int32(2))

}

func CreateRegister(register *et.Register, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.AccountmanagerX)
	tx, err = ety.Create(et.NameRegisterAction, register)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.AccountmanagerX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateReset(reset *et.ResetKey, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.AccountmanagerX)
	tx, err = ety.Create(et.NameResetAction, reset)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.AccountmanagerX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateTransfer(tranfer *et.Transfer, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.AccountmanagerX)
	tx, err = ety.Create(et.NameTransferAction, tranfer)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.AccountmanagerX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateSupervise(supervise *et.Supervise, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.AccountmanagerX)
	tx, err = ety.Create(et.NameSuperviseAction, supervise)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.AccountmanagerX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func CreateApply(apply *et.Apply, privKey string) (tx *types.Transaction, err error) {
	ety := types.LoadExecutorType(et.AccountmanagerX)
	tx, err = ety.Create(et.NameApplyAction, apply)
	if err != nil {
		return nil, err
	}
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	tx, err = types.FormatTx(cfg, et.AccountmanagerX, tx)
	if err != nil {
		return nil, err
	}
	tx, err = signTx(tx, privKey)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

//模拟区块中交易得执行过程
func Exec_Block(t *testing.T, stateDB db.DB, kvdb db.KVDB, env *execEnv, txs ...*types.Transaction) error {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	e := exec.(*Accountmanager)
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	for index, tx := range txs {
		err := e.CheckTx(tx, index)
		if err != nil {
			return err
		}

	}
	env.blockHeight = env.blockHeight + 1
	env.blockTime = env.blockTime + 1
	env.difficulty = env.difficulty + 1
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	for index, tx := range txs {
		receipt, err := exec.Exec(tx, index)
		if err != nil {
			return err
		}
		for _, kv := range receipt.KV {
			stateDB.Set(kv.Key, kv.Value)
		}
		receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
		set, err := exec.ExecLocal(tx, receiptData, index)
		if err != nil {
			return err
		}
		for _, kv := range set.KV {
			kvdb.Set(kv.Key, kv.Value)
		}
		//save to database
		util.SaveKVList(stateDB, set.KV)
		assert.Equal(t, types.ExecOk, int(receipt.Ty))
	}
	return nil
}

func Exec_QueryAccountByID(accountID string, stateDB db.KV, kvdb db.KVDB) (*et.Account, error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	msg, err := exec.Query(et.FuncNameQueryAccountByID, types.Encode(&et.QueryAccountByID{AccountID: accountID}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.Account), err
}

func Exec_QueryAccountByAddr(addr string, stateDB db.KV, kvdb db.KVDB) (*et.Account, error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	msg, err := exec.Query(et.FuncNameQueryAccountByAddr, types.Encode(&et.QueryAccountByAddr{Addr: addr}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.Account), err
}

func Exec_QueryAccountsByStatus(status int32, stateDB db.KV, kvdb db.KVDB) (*et.ReplyAccountList, error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	msg, err := exec.Query(et.FuncNameQueryAccountsByStatus, types.Encode(&et.QueryAccountsByStatus{Status: status}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.ReplyAccountList), err
}

func Exec_QueryBalanceByID(in *et.QueryBalanceByID, stateDB db.KV, kvdb db.KVDB) (*et.Balance, error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	msg, err := exec.Query(et.FuncNameQueryBalanceByID, types.Encode(in))
	if err != nil {
		return nil, err
	}
	return msg.(*et.Balance), err
}

func Exec_QueryExpiredAccounts(expiredtime int64, stateDB db.KV, kvdb db.KVDB) (*et.ReplyAccountList, error) {
	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	exec := newAccountmanager()
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	msg, err := exec.Query(et.FuncNameQueryExpiredAccounts, types.Encode(&et.QueryExpiredAccounts{ExpiredTime: expiredtime, Direction: 0}))
	if err != nil {
		return nil, err
	}
	return msg.(*et.ReplyAccountList), err
}
func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		return tx, err
	}
	bytes, err := common.FromHex(hexPrivKey[:])
	if err != nil {
		return tx, err
	}
	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return tx, err
	}
	tx.Sign(int32(signType), privKey)
	return tx, nil
}

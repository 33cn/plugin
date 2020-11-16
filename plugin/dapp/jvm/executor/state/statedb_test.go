package state

import (
	"fmt"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/db"
	dbm "github.com/33cn/chain33/common/db"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type statedb_test struct {
	stateDB dbm.KV
	localDB dbm.KVDB
	base    *drivers.DriverBase
}

var (
	opener       = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
	player       = "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
	chainTestCfg = types.NewChain33Config(types.GetDefaultCfgstring())
)

func setupTestEnv() *statedb_test {
	sdb, _ := db.NewGoMemDB("JvmTestDb", "test", 128)
	_, _, kvdb := util.CreateTestDB()

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chainTestCfg, nil)

	statedb_test_env := &statedb_test{
		stateDB: sdb,
		localDB: kvdb,
		base:    &drivers.DriverBase{},
	}
	statedb_test_env.base.SetAPI(api)
	statedb_test_env.base.SetStateDB(sdb)
	return statedb_test_env
}

func deposit2contract(t *testing.T, acc *account.DB, contractName, addr string) {
	account2operate := &types.Account{
		Balance: 1000 * 1e8,
		Addr:    addr,
	}
	contractAddr := address.ExecAddress(contractName)
	acc.SaveAccount(account2operate)
	account2operate = acc.LoadAccount(addr)
	assert.Equal(t, int64(1000*1e8), account2operate.Balance)
	_, err := acc.TransferToExec(addr, contractAddr, 200*1e8)
	assert.Nil(t, err)
	account2operate = acc.LoadExecAccount(addr, contractAddr)
	assert.Equal(t, int64(200*1e8), account2operate.Balance)
}

func Test_CreateAccount(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)
	account := memoryStateDB.GetAccount(addr)
	assert.Equal(t, account.Addr, addr)
}

func Test_GetBalance(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)

	deposit2contract(t, memoryStateDB.CoinsAccount, exectorName, opener)
	openerBalance := memoryStateDB.GetBalance(opener)
	assert.Equal(t, int64(openerBalance), int64(800*1e8))

	openerBalanceInContract := memoryStateDB.GetBalance(addr)
	assert.Equal(t, int64(openerBalanceInContract), int64(200*1e8))
}

func Test_SetCodeAndGet(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)

	code := []byte{1, 2, 3, 4}
	memoryStateDB.SetCodeAndAbi(addr, code, nil)
	codeGet := memoryStateDB.GetCode(addr)
	assert.Equal(t, code, codeGet)

	contractName := memoryStateDB.GetName(addr)
	assert.Equal(t, exectorName, contractName)
}

func Test_LocalDBList(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)

	var keys [][]byte
	var values [][]byte
	txHash := "0x1b8b0520b9a5d5325f2bcf900b7ad1d9d638d06899e1d968e47cea14315d6154"
	for i := 0; i < 10; i++ {
		keys = append(keys, []byte(fmt.Sprintf("key:%d", i)))
		values = append(values, []byte(fmt.Sprintf("value:%d", i)))
		memoryStateDB.SetValue2Local(addr, string(keys[i]), values[i], txHash)
	}

	keyValues := GetAllLocalKeyValues(txHash)

	for _, keyvalue := range keyValues {
		_ = memoryStateDB.LocalDB.Set(keyvalue.Key, keyvalue.Value)
	}

	valuesListed := memoryStateDB.List([]byte("LODB"), 20)
	fmt.Println(string(valuesListed[0]), string(valuesListed[1]))
	assert.Equal(t, 10, len(valuesListed))
}

func Test_ExecTransferFrozen(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)

	deposit2contract(t, memoryStateDB.CoinsAccount, exectorName, opener)
	openerBalance := memoryStateDB.GetBalance(opener)
	assert.Equal(t, int64(openerBalance), int64(800*1e8))

	openerBalanceInContract := memoryStateDB.GetBalance(addr)
	assert.Equal(t, int64(openerBalanceInContract), int64(200*1e8))

	tx := &types.Transaction{
		Execer: []byte(exectorName),
	}
	//在合约内部的opener冻结 200
	memoryStateDB.ExecFrozen(tx, opener, int64(200*1e8))
	openerAccount := memoryStateDB.CoinsAccount.LoadExecAccount(opener, addr)
	assert.Equal(t, openerAccount.Frozen, int64(200*1e8))
	playerAccount := memoryStateDB.CoinsAccount.LoadExecAccount(player, addr)
	assert.Equal(t, playerAccount.Balance, int64(0))

	memoryStateDB.ExecTransferFrozen(tx, opener, player, int64(100*1e8))
	openerAccount = memoryStateDB.CoinsAccount.LoadExecAccount(opener, addr)
	assert.Equal(t, openerAccount.Frozen, int64(100*1e8))
	playerAccount = memoryStateDB.CoinsAccount.LoadExecAccount(player, addr)
	assert.Equal(t, playerAccount.Balance, int64(100*1e8))
}

func Test_AccountOpErrorBranch(t *testing.T) {
	env := setupTestEnv()
	exectorName := "user.jvm.Dice"
	memoryStateDB := NewMemoryStateDB(exectorName, env.stateDB, env.localDB, env.base.GetCoinsAccount(), 10)
	addr := address.ExecAddress(exectorName)
	memoryStateDB.CreateAccount(addr, opener, exectorName)

	tx := &types.Transaction{
		Execer: []byte(exectorName),
	}

	result := memoryStateDB.ExecFrozen(nil, opener, int64(100*1e8))
	assert.Equal(t, false, result)
	result = memoryStateDB.ExecFrozen(tx, player, int64(100*1e8))
	assert.Equal(t, false, result)

	result = memoryStateDB.ExecActive(nil, opener, int64(100*1e8))
	assert.Equal(t, false, result)
	result = memoryStateDB.ExecActive(tx, player, int64(100*1e8))
	assert.Equal(t, false, result)

	result = memoryStateDB.ExecTransfer(nil, opener, player, int64(100*1e8))
	assert.Equal(t, false, result)
	result = memoryStateDB.ExecTransfer(tx, player, player, int64(100*1e8))
	assert.Equal(t, false, result)

	result = memoryStateDB.ExecTransferFrozen(nil, opener, player, int64(100*1e8))
	assert.Equal(t, false, result)
	result = memoryStateDB.ExecTransferFrozen(tx, player, player, int64(100*1e8))
	assert.Equal(t, false, result)

}

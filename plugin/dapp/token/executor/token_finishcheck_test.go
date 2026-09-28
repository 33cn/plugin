// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"
	"sync/atomic"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pty "github.com/33cn/plugin/plugin/dapp/token/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// regression tests for fork ForkTokenFinishCheck:
// finishCreate must check global token symbol uniqueness, otherwise the same
// symbol precreated by N different owners can be finished N times and each
// finish triggers GenesisInit, over-issuing N*Total of the same token.

var dupFinishNonce int64 = 2000000

func dupFinishCfg(forkHeight int64) *types.Chain33Config {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1))
	cfg.SetDappFork(pty.TokenX, pty.ForkTokenCheckX, 1600000)
	cfg.SetDappFork(pty.TokenX, pty.ForkTokenFinishCheckX, forkHeight)
	return cfg
}

// newDupFinishExec builds a token executor with fresh state/local dbs at height 1600000.
// It does not call Init(), to avoid conflicting with the driver registration in token_test.go.
func newDupFinishExec(cfg *types.Chain33Config) (*token, dbm.KVDB, *dbm.GoMemDB) {
	ety := types.LoadExecutorType(driverName)
	ety.InitFuncList(types.ListMethod(&token{}))

	stateDB, _ := dbm.NewGoMemDB("dupfinish", "dupfinish", 100)
	_, _, kvdb := util.CreateTestDB()

	// token-blacklist manage config, needed since ForkTokenBlackListX is active
	item := &types.ConfigItem{
		Key:   "mavl-manage-token-blacklist",
		Value: &types.ConfigItem_Arr{Arr: &types.ArrayConfig{Value: []string{"bty"}}},
	}
	stateDB.Set([]byte(item.Key), types.Encode(item))

	// token-finisher manage config: Nodes[0] is a valid finisher
	item2 := &types.ConfigItem{
		Key:   "mavl-manage-token-finisher",
		Value: &types.ConfigItem_Arr{Arr: &types.ArrayConfig{Value: []string{string(Nodes[0])}}},
	}
	stateDB.Set([]byte(item2.Key), types.Encode(item2))

	exec := newToken()
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec.SetAPI(api)
	exec.SetStateDB(stateDB)
	exec.SetLocalDB(kvdb)
	exec.SetEnv(1600000, 1539918074, 1539918074)
	return exec.(*token), kvdb, stateDB
}

func dupFinishApplyReceipt(stateDB *dbm.GoMemDB, receipt *types.Receipt) {
	for _, kv := range receipt.KV {
		stateDB.Set(kv.Key, kv.Value)
	}
}

// dupFinishPreCreate executes one preCreate with price=0, owner and signer priv match.
func dupFinishPreCreate(t *testing.T, exec *token, stateDB *dbm.GoMemDB,
	symbol, owner, priv string, total int64) error {
	p := &pty.TokenPreCreate{
		Name: symbol, Symbol: symbol, Introduction: symbol,
		Total: total, Price: 0, Owner: owner,
	}
	tx, err := types.CallCreateTransaction(pty.TokenX, "TokenPreCreate", p)
	assert.Nil(t, err)
	tx.Nonce = atomic.AddInt64(&dupFinishNonce, 1)
	tx, err = signTx(tx, priv)
	assert.Nil(t, err)
	receipt, err := exec.Exec(tx, 1)
	if err == nil {
		dupFinishApplyReceipt(stateDB, receipt)
	}
	return err
}

// dupFinishCreate finishes the token creation, signed by the finisher (Nodes[0], PrivKeyA).
func dupFinishCreate(t *testing.T, exec *token, stateDB *dbm.GoMemDB, symbol, owner string) error {
	p := &pty.TokenFinishCreate{Symbol: symbol, Owner: owner}
	tx, err := types.CallCreateTransaction(pty.TokenX, "TokenFinishCreate", p)
	assert.Nil(t, err)
	tx.Nonce = atomic.AddInt64(&dupFinishNonce, 1)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	receipt, err := exec.Exec(tx, 1)
	if err == nil {
		dupFinishApplyReceipt(stateDB, receipt)
	}
	return err
}

// TestTokenDupFinishCreateRejected verifies with ForkTokenFinishCheck active:
// two owners precreate the same symbol, the first finish succeeds, the second
// finish is rejected and no over-issuance happens. The normal
// precreate/finish flow is not affected.
func TestTokenDupFinishCreateRejected(t *testing.T) {
	cfg := dupFinishCfg(0)
	exec, _, stateDB := newDupFinishExec(cfg)
	symbol := "DUPCHECK"
	total := int64(10000 * 1e8)
	ownerA := string(Nodes[0])
	ownerB := string(Nodes[1])

	// both owners can still precreate the same symbol (per-owner records)
	err := dupFinishPreCreate(t, exec, stateDB, symbol, ownerA, PrivKeyA, total)
	assert.Nil(t, err, "owner A precreate should succeed")
	err = dupFinishPreCreate(t, exec, stateDB, symbol, ownerB, PrivKeyB, total)
	assert.Nil(t, err, "owner B precreate same symbol should succeed")

	// first finish succeeds and issues total to owner A
	err = dupFinishCreate(t, exec, stateDB, symbol, ownerA)
	assert.Nil(t, err, "first finishCreate should succeed")

	// second finish of the same symbol by another owner must be rejected
	err = dupFinishCreate(t, exec, stateDB, symbol, ownerB)
	assert.Equal(t, pty.ErrTokenExist, err, "duplicate finishCreate must be rejected")

	// same-owner repeated finish is still rejected by the status check
	err = dupFinishCreate(t, exec, stateDB, symbol, ownerA)
	assert.Equal(t, pty.ErrTokenNotPrecreated, err, "repeated finish of created token rejected")

	// only owner A got the genesis amount, total issued equals declared Total
	accDB, err := account.NewAccountDB(cfg, "token", symbol, stateDB)
	assert.Nil(t, err)
	assert.Equal(t, total, accDB.LoadAccount(ownerA).Balance)
	assert.Equal(t, int64(0), accDB.LoadAccount(ownerB).Balance)

	tk, err := loadTokenDB(stateDB, symbol)
	assert.Nil(t, err)
	assert.Equal(t, total, tk.token.Total)
}

// TestTokenDupFinishCreatePreFork verifies the behavior before the fork height
// stays unchanged: duplicate finishCreate of the same symbol still succeeds.
func TestTokenDupFinishCreatePreFork(t *testing.T) {
	cfg := dupFinishCfg(2000000) // fork active above the test height 1600000
	exec, _, stateDB := newDupFinishExec(cfg)
	symbol := "PREFORK"
	total := int64(10000 * 1e8)
	ownerA := string(Nodes[0])
	ownerB := string(Nodes[1])

	err := dupFinishPreCreate(t, exec, stateDB, symbol, ownerA, PrivKeyA, total)
	assert.Nil(t, err)
	err = dupFinishPreCreate(t, exec, stateDB, symbol, ownerB, PrivKeyB, total)
	assert.Nil(t, err)

	err = dupFinishCreate(t, exec, stateDB, symbol, ownerA)
	assert.Nil(t, err)
	// pre-fork behavior: no global symbol uniqueness check in finishCreate
	err = dupFinishCreate(t, exec, stateDB, symbol, ownerB)
	assert.Nil(t, err, "pre-fork duplicate finishCreate keeps the old behavior")

	accDB, err := account.NewAccountDB(cfg, "token", symbol, stateDB)
	assert.Nil(t, err)
	assert.Equal(t, total, accDB.LoadAccount(ownerA).Balance)
	assert.Equal(t, total, accDB.LoadAccount(ownerB).Balance)
}

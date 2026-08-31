package executor

// Regression tests for the IssuanceRepay collateral-theft vulnerability.
// Before ForkIssuanceRepayOwner, a third party repaying someone's debt record
// received the debtor's collateral. After the fork the collateral is returned
// to the debt record owner (record.AccountAddr) instead of the caller.

import (
	"sync"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pkt "github.com/33cn/plugin/plugin/dapp/issuance/types"
	tokenE "github.com/33cn/plugin/plugin/dapp/token/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// repayOwnerIssInitOnce guards driver registration: initEnv in issuance_test.go
// may already have registered the driver, and registering twice panics.
var repayOwnerIssInitOnce sync.Once

// newRepayOwnerIssEnv builds a test env with ForkIssuanceRepayOwner registered
// at the given height, so both pre-fork and post-fork behavior can be tested.
func newRepayOwnerIssEnv(repayForkHeight int64) *execEnv {
	repayOwnerIssInitOnce.Do(func() {
		// ignore duplicated driver registration when run together with other tests
		defer func() { recover() }()
		cfg0 := types.NewChain33Config(types.GetDefaultCfgstring())
		cfg0.SetTitleOnlyForTest("chain33")
		cfg0.RegisterDappFork(pkt.IssuanceX, pkt.ForkIssuanceTableUpdate, 0)
		cfg0.RegisterDappFork(pkt.IssuanceX, pkt.ForkIssuanceRepayOwner, 0)
		Init(pkt.IssuanceX, cfg0, nil)
	})

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	cfg.RegisterDappFork(pkt.IssuanceX, pkt.ForkIssuanceTableUpdate, 0)
	cfg.RegisterDappFork(pkt.IssuanceX, pkt.ForkIssuanceRepayOwner, repayForkHeight)
	_, ldb, kvdb := util.CreateTestDB()

	accountA := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[0])}
	accountAToken := types.Account{Balance: totalToken, Frozen: 0, Addr: string(Nodes[0])}
	accountB := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[1])}
	accountC := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[2])}

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	execAddr := dapp.ExecAddress(pkt.IssuanceX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)
	manageKeySet("issuance-manage", accountA.Addr, stateDB)
	manageKeySet("issuance-fund", accountA.Addr, stateDB)
	tokenAccA, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccA.SaveExecAccount(execAddr, &accountAToken)

	accB := account.NewCoinsAccount(cfg)
	accB.SetDB(stateDB)
	accB.SaveExecAccount(execAddr, &accountB)
	manageKeySet("issuance-price-feed", accountB.Addr, stateDB)

	accC := account.NewCoinsAccount(cfg)
	accC.SetDB(stateDB)
	accC.SaveExecAccount(execAddr, &accountC)
	manageKeySet("issuance-guarantor", accountC.Addr, stateDB)

	return &execEnv{
		blockTime:   time.Now().Unix(),
		blockHeight: cfg.GetDappFork(pkt.IssuanceX, "Enable"),
		difficulty:  1539918074,
		kvdb:        kvdb,
		api:         api,
		db:          stateDB,
		execAddr:    execAddr,
		cfg:         cfg,
		ldb:         ldb,
	}
}

func repayOwnerIssNewExec(env *execEnv) *Issuance {
	exec := newIssuance().(*Issuance)
	exec.SetAPI(env.api)
	exec.SetStateDB(env.db)
	exec.SetLocalDB(env.kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	return exec
}

func repayOwnerIssExecTx(t *testing.T, env *execEnv, exec *Issuance, tx *types.Transaction, index int) (*types.Receipt, error) {
	receipt, err := exec.Exec(tx, index)
	if err != nil {
		return receipt, err
	}
	for _, kv := range receipt.KV {
		env.db.Set(kv.Key, kv.Value)
	}
	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	set, err := exec.ExecLocal(tx, receiptData, index)
	assert.Nil(t, err)
	util.SaveKVList(env.ldb, set.KV)
	return receipt, nil
}

func repayOwnerIssGiveToken(env *execEnv, addr string, amount int64) {
	tokenAcc, _ := account.NewAccountDB(env.cfg, tokenE.GetName(), pkt.CCNYTokenName, env.db)
	tokenAcc.SaveExecAccount(env.execAddr, &types.Account{Balance: amount, Addr: addr})
}

func repayOwnerIssTokenBalance(env *execEnv, addr string) int64 {
	tokenAcc, _ := account.NewAccountDB(env.cfg, tokenE.GetName(), pkt.CCNYTokenName, env.db)
	return tokenAcc.LoadExecAccount(addr, env.execAddr).GetBalance()
}

func repayOwnerIssCoinBalance(env *execEnv, addr string) int64 {
	acc := account.NewCoinsAccount(env.cfg)
	acc.SetDB(env.db)
	return acc.LoadExecAccount(addr, env.execAddr).GetBalance()
}

// repayOwnerIssPrepare: create(A, 1000 CCNY) + feed(B, price=1) + manage(A, super=B)
func repayOwnerIssPrepare(t *testing.T, env *execEnv, exec *Issuance) string {
	p1 := &pkt.IssuanceCreateTx{
		TotalBalance:     1000,
		DebtCeiling:      200,
		LiquidationRatio: 0.25,
		Period:           3600 * 24 * 365,
	}
	tx, err := pkt.CreateRawIssuanceCreateTx(env.cfg, p1)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)
	issuanceID := common.ToHex(tx.Hash())

	p2 := &pkt.IssuanceFeedTx{}
	p2.Price = append(p2.Price, 1)
	p2.Volume = append(p2.Volume, 100)
	tx, err = pkt.CreateRawIssuanceFeedTx(env.cfg, p2)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	p3 := &pkt.IssuanceManageTx{}
	p3.Addr = append(p3.Addr, string(Nodes[1]))
	tx, err = pkt.CreateRawIssuanceManageTx(env.cfg, p3)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	return issuanceID
}

// repayOwnerIssDebt: debtor B (super address) borrows value CCNY
func repayOwnerIssDebt(t *testing.T, env *execEnv, exec *Issuance, issuanceID string, value float64) string {
	p := &pkt.IssuanceDebtTx{IssuanceID: issuanceID, Value: value}
	tx, err := pkt.CreateRawIssuanceDebtTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)
	return common.ToHex(tx.Hash())
}

// After ForkIssuanceRepayOwner: a third party may still repay on behalf of the
// debtor, but the collateral must go back to the debt record owner.
func TestRepayOwnerForkRepayByNonDebtor(t *testing.T) {
	env := newRepayOwnerIssEnv(0)
	exec := repayOwnerIssNewExec(env)

	// repayer C funds the repayment, C is unrelated to the debt record
	repayFund := 1000 * types.DefaultCoinPrecision
	repayOwnerIssGiveToken(env, string(Nodes[2]), repayFund)

	issuanceID := repayOwnerIssPrepare(t, env, exec)
	// debtor B borrows 100 CCNY with 400 BTY collateral (price=1, ratio=0.25)
	debtID := repayOwnerIssDebt(t, env, exec, issuanceID, 100)

	record, err := queryIssuanceRecordByID(env.db, issuanceID, debtID)
	assert.Nil(t, err)
	assert.Equal(t, string(Nodes[1]), record.AccountAddr)

	coinBBefore := repayOwnerIssCoinBalance(env, string(Nodes[1]))
	coinCBefore := repayOwnerIssCoinBalance(env, string(Nodes[2]))
	tokenCBefore := repayOwnerIssTokenBalance(env, string(Nodes[2]))

	// C repays B's debt record
	p := &pkt.IssuanceRepayTx{IssuanceID: issuanceID, DebtID: debtID}
	tx, err := pkt.CreateRawIssuanceRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyC)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	coinBAfter := repayOwnerIssCoinBalance(env, string(Nodes[1]))
	coinCAfter := repayOwnerIssCoinBalance(env, string(Nodes[2]))
	tokenCAfter := repayOwnerIssTokenBalance(env, string(Nodes[2]))

	// collateral returns to the debtor B, not to the repayer C
	assert.Equal(t, record.CollateralValue, coinBAfter-coinBBefore, "collateral must return to the debtor")
	assert.Equal(t, coinCBefore, coinCAfter, "repayer must not receive the collateral")
	// the repayer only loses the repayment funds (the exact debt value)
	assert.Equal(t, record.DebtValue, tokenCBefore-tokenCAfter, "repayer cost must equal the debt value")
}

// After ForkIssuanceRepayOwner: the debtor repaying his own record keeps working
// as before and gets the collateral back.
func TestRepayOwnerForkRepayByDebtor(t *testing.T) {
	env := newRepayOwnerIssEnv(0)
	exec := repayOwnerIssNewExec(env)

	issuanceID := repayOwnerIssPrepare(t, env, exec)
	debtID := repayOwnerIssDebt(t, env, exec, issuanceID, 100)

	record, err := queryIssuanceRecordByID(env.db, issuanceID, debtID)
	assert.Nil(t, err)

	coinBBefore := repayOwnerIssCoinBalance(env, string(Nodes[1]))
	tokenBBefore := repayOwnerIssTokenBalance(env, string(Nodes[1]))

	p := &pkt.IssuanceRepayTx{IssuanceID: issuanceID, DebtID: debtID}
	tx, err := pkt.CreateRawIssuanceRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	assert.Equal(t, record.CollateralValue, repayOwnerIssCoinBalance(env, string(Nodes[1]))-coinBBefore,
		"debtor must get the collateral back")
	assert.Equal(t, record.DebtValue, tokenBBefore-repayOwnerIssTokenBalance(env, string(Nodes[1])),
		"debtor cost must equal the debt value")
}

// Before ForkIssuanceRepayOwner the old behavior is preserved: the repayer
// receives the collateral (documenting the vulnerable pre-fork semantics).
func TestRepayOwnerPreForkRepayByNonDebtor(t *testing.T) {
	env := newRepayOwnerIssEnv(100000000)
	exec := repayOwnerIssNewExec(env)

	repayFund := 1000 * types.DefaultCoinPrecision
	repayOwnerIssGiveToken(env, string(Nodes[2]), repayFund)

	issuanceID := repayOwnerIssPrepare(t, env, exec)
	debtID := repayOwnerIssDebt(t, env, exec, issuanceID, 100)

	record, err := queryIssuanceRecordByID(env.db, issuanceID, debtID)
	assert.Nil(t, err)

	coinBBefore := repayOwnerIssCoinBalance(env, string(Nodes[1]))
	coinCBefore := repayOwnerIssCoinBalance(env, string(Nodes[2]))

	p := &pkt.IssuanceRepayTx{IssuanceID: issuanceID, DebtID: debtID}
	tx, err := pkt.CreateRawIssuanceRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.IssuanceX)
	tx, err = signTx(tx, PrivKeyC)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerIssExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	// pre-fork behavior unchanged: collateral goes to the caller (repayer C)
	assert.Equal(t, record.CollateralValue, repayOwnerIssCoinBalance(env, string(Nodes[2]))-coinCBefore,
		"pre-fork: collateral goes to the repayer")
	assert.Equal(t, coinBBefore, repayOwnerIssCoinBalance(env, string(Nodes[1])),
		"pre-fork: debtor gets nothing back")
}

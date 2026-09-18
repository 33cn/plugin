package executor

// Regression tests for the CollateralizeRepay collateral-theft vulnerability.
// Before ForkCollateralizeRepayOwner, a third party repaying someone's borrow
// record received the borrower's collateral. After the fork the collateral is
// returned to the borrow record owner (record.AccountAddr) instead of the caller.

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
	pkt "github.com/33cn/plugin/plugin/dapp/collateralize/types"
	tokenE "github.com/33cn/plugin/plugin/dapp/token/executor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// repayOwnerInitOnce guards driver registration: initEnv in collateralize_test.go
// may already have registered the driver, and registering twice panics.
var repayOwnerInitOnce sync.Once

// newRepayOwnerEnv builds a test env with ForkCollateralizeRepayOwner registered
// at the given height, so both pre-fork and post-fork behavior can be tested.
func newRepayOwnerEnv(repayForkHeight int64) *execEnv {
	repayOwnerInitOnce.Do(func() {
		// ignore duplicated driver registration when run together with other tests
		defer func() { recover() }()
		cfg0 := types.NewChain33Config(types.GetDefaultCfgstring())
		cfg0.SetTitleOnlyForTest("chain33")
		cfg0.RegisterDappFork(pkt.CollateralizeX, pkt.ForkCollateralizeTableUpdate, 0)
		cfg0.RegisterDappFork(pkt.CollateralizeX, pkt.ForkCollateralizeRepayOwner, 0)
		Init(pkt.CollateralizeX, cfg0, nil)
	})

	cfg := types.NewChain33Config(types.GetDefaultCfgstring())
	cfg.SetTitleOnlyForTest("chain33")
	cfg.RegisterDappFork(pkt.CollateralizeX, pkt.ForkCollateralizeTableUpdate, 0)
	cfg.RegisterDappFork(pkt.CollateralizeX, pkt.ForkCollateralizeRepayOwner, repayForkHeight)
	_, ldb, kvdb := util.CreateTestDB()

	accountA := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[0])}
	accountAToken := types.Account{Balance: totalToken, Frozen: 0, Addr: string(Nodes[0])}
	accountB := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[1])}
	accountBToken := types.Account{Balance: types.DefaultCoinPrecision / 10, Frozen: 0, Addr: string(Nodes[1])}
	accountC := types.Account{Balance: total, Frozen: 0, Addr: string(Nodes[2])}

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	execAddr := dapp.ExecAddress(pkt.CollateralizeX)
	stateDB, _ := dbm.NewGoMemDB("1", "2", 100)

	accA := account.NewCoinsAccount(cfg)
	accA.SetDB(stateDB)
	accA.SaveExecAccount(execAddr, &accountA)
	manageKeySet("issuance-manage", accountA.Addr, stateDB)
	addrKeySet(accountA.Addr, stateDB)
	tokenAccA, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccA.SaveExecAccount(execAddr, &accountAToken)

	accB := account.NewCoinsAccount(cfg)
	accB.SetDB(stateDB)
	accB.SaveExecAccount(execAddr, &accountB)
	manageKeySet("issuance-price-feed", accountB.Addr, stateDB)
	tokenAccB, _ := account.NewAccountDB(cfg, tokenE.GetName(), pkt.CCNYTokenName, stateDB)
	tokenAccB.SaveExecAccount(execAddr, &accountBToken)

	accC := account.NewCoinsAccount(cfg)
	accC.SetDB(stateDB)
	accC.SaveExecAccount(execAddr, &accountC)
	manageKeySet("issuance-guarantor", accountC.Addr, stateDB)

	return &execEnv{
		blockTime:   time.Now().Unix(),
		blockHeight: cfg.GetDappFork(pkt.CollateralizeX, "Enable"),
		difficulty:  1539918074,
		kvdb:        kvdb,
		api:         api,
		db:          stateDB,
		execAddr:    execAddr,
		cfg:         cfg,
		ldb:         ldb,
	}
}

func repayOwnerNewExec(env *execEnv) *Collateralize {
	exec := newCollateralize().(*Collateralize)
	exec.SetAPI(env.api)
	exec.SetStateDB(env.db)
	exec.SetLocalDB(env.kvdb)
	exec.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	return exec
}

func repayOwnerExecTx(t *testing.T, env *execEnv, exec *Collateralize, tx *types.Transaction, index int) (*types.Receipt, error) {
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

func repayOwnerGiveToken(env *execEnv, addr string, amount int64) {
	tokenAcc, _ := account.NewAccountDB(env.cfg, tokenE.GetName(), pkt.CCNYTokenName, env.db)
	tokenAcc.SaveExecAccount(env.execAddr, &types.Account{Balance: amount, Addr: addr})
}

func repayOwnerTokenBalance(env *execEnv, addr string) int64 {
	tokenAcc, _ := account.NewAccountDB(env.cfg, tokenE.GetName(), pkt.CCNYTokenName, env.db)
	return tokenAcc.LoadExecAccount(addr, env.execAddr).GetBalance()
}

func repayOwnerCoinBalance(env *execEnv, addr string) int64 {
	acc := account.NewCoinsAccount(env.cfg)
	acc.SetDB(env.db)
	return acc.LoadExecAccount(addr, env.execAddr).GetBalance()
}

// repayOwnerPreparePool: manage(A) + create(A, 1000 CCNY) + feed(B, price=1)
func repayOwnerPreparePool(t *testing.T, env *execEnv, exec *Collateralize) string {
	p := &pkt.CollateralizeManageTx{}
	p.Period = 3600 * 24 * 365
	p.LiquidationRatio = 0.25
	p.DebtCeiling = 100
	p.StabilityFeeRatio = 0.0001
	p.TotalBalance = 10000
	tx, err := pkt.CreateRawCollateralizeManageTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	p1 := &pkt.CollateralizeCreateTx{TotalBalance: 1000}
	tx, err = pkt.CreateRawCollateralizeCreateTx(env.cfg, p1)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyA)
	assert.Nil(t, err)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)
	collID := common.ToHex(tx.Hash())

	p2 := &pkt.CollateralizeFeedTx{}
	p2.Price = append(p2.Price, 1)
	p2.Volume = append(p2.Volume, 100)
	tx, err = pkt.CreateRawCollateralizeFeedTx(env.cfg, p2)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)
	return collID
}

// repayOwnerBorrow: borrower B borrows value CCNY
func repayOwnerBorrow(t *testing.T, env *execEnv, exec *Collateralize, collID string, value float64) string {
	p := &pkt.CollateralizeBorrowTx{CollateralizeID: collID, Value: value}
	tx, err := pkt.CreateRawCollateralizeBorrowTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+1, env.blockTime+1, env.difficulty)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)
	return common.ToHex(tx.Hash())
}

// After ForkCollateralizeRepayOwner: a third party may still repay on behalf of
// the borrower, but the collateral must go back to the borrow record owner.
func TestRepayOwnerForkRepayByNonBorrower(t *testing.T) {
	env := newRepayOwnerEnv(0)
	exec := repayOwnerNewExec(env)

	// repayer C funds the repayment, C is unrelated to the borrow record
	repayFund := 1000 * types.DefaultCoinPrecision
	repayOwnerGiveToken(env, string(Nodes[2]), repayFund)

	collID := repayOwnerPreparePool(t, env, exec)
	// borrower B borrows 100 CCNY with 400 BTY collateral (price=1, ratio=0.25)
	borrowID := repayOwnerBorrow(t, env, exec, collID, 100)

	record, err := queryCollateralizeRecordByID(env.db, collID, borrowID)
	assert.Nil(t, err)
	assert.Equal(t, string(Nodes[1]), record.AccountAddr)

	coinBBefore := repayOwnerCoinBalance(env, string(Nodes[1]))
	coinCBefore := repayOwnerCoinBalance(env, string(Nodes[2]))
	tokenCBefore := repayOwnerTokenBalance(env, string(Nodes[2]))

	// C repays B's borrow record
	p := &pkt.CollateralizeRepayTx{CollateralizeID: collID, RecordID: borrowID}
	tx, err := pkt.CreateRawCollateralizeRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyC)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	coinBAfter := repayOwnerCoinBalance(env, string(Nodes[1]))
	coinCAfter := repayOwnerCoinBalance(env, string(Nodes[2]))
	tokenCAfter := repayOwnerTokenBalance(env, string(Nodes[2]))

	// collateral returns to the borrower B, not to the repayer C
	assert.Equal(t, record.CollateralValue, coinBAfter-coinBBefore, "collateral must return to the borrower")
	assert.Equal(t, coinCBefore, coinCAfter, "repayer must not receive the collateral")
	// the repayer only loses the repayment funds (debt + stability fee)
	tokenCost := tokenCBefore - tokenCAfter
	assert.True(t, tokenCost >= record.DebtValue, "repayer cost must cover the debt")
	assert.True(t, tokenCost < record.CollateralValue, "repayer must not profit from the collateral")
}

// After ForkCollateralizeRepayOwner: the borrower repaying his own record keeps
// working as before and gets the collateral back.
func TestRepayOwnerForkRepayByBorrower(t *testing.T) {
	env := newRepayOwnerEnv(0)
	exec := repayOwnerNewExec(env)

	collID := repayOwnerPreparePool(t, env, exec)
	borrowID := repayOwnerBorrow(t, env, exec, collID, 100)

	record, err := queryCollateralizeRecordByID(env.db, collID, borrowID)
	assert.Nil(t, err)

	coinBBefore := repayOwnerCoinBalance(env, string(Nodes[1]))
	tokenBBefore := repayOwnerTokenBalance(env, string(Nodes[1]))

	p := &pkt.CollateralizeRepayTx{CollateralizeID: collID, RecordID: borrowID}
	tx, err := pkt.CreateRawCollateralizeRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyB)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	assert.Equal(t, record.CollateralValue, repayOwnerCoinBalance(env, string(Nodes[1]))-coinBBefore,
		"borrower must get the collateral back")
	assert.True(t, tokenBBefore-repayOwnerTokenBalance(env, string(Nodes[1])) >= record.DebtValue,
		"borrower cost must cover the debt")
}

// Before ForkCollateralizeRepayOwner the old behavior is preserved: the repayer
// receives the collateral (documenting the vulnerable pre-fork semantics).
func TestRepayOwnerPreForkRepayByNonBorrower(t *testing.T) {
	env := newRepayOwnerEnv(100000000)
	exec := repayOwnerNewExec(env)

	repayFund := 1000 * types.DefaultCoinPrecision
	repayOwnerGiveToken(env, string(Nodes[2]), repayFund)

	collID := repayOwnerPreparePool(t, env, exec)
	borrowID := repayOwnerBorrow(t, env, exec, collID, 100)

	record, err := queryCollateralizeRecordByID(env.db, collID, borrowID)
	assert.Nil(t, err)

	coinBBefore := repayOwnerCoinBalance(env, string(Nodes[1]))
	coinCBefore := repayOwnerCoinBalance(env, string(Nodes[2]))

	p := &pkt.CollateralizeRepayTx{CollateralizeID: collID, RecordID: borrowID}
	tx, err := pkt.CreateRawCollateralizeRepayTx(env.cfg, p)
	assert.Nil(t, err)
	tx.Execer = []byte(pkt.CollateralizeX)
	tx, err = signTx(tx, PrivKeyC)
	assert.Nil(t, err)
	exec.SetEnv(env.blockHeight+2, env.blockTime+2, env.difficulty)
	_, err = repayOwnerExecTx(t, env, exec, tx, 1)
	assert.Nil(t, err)

	// pre-fork behavior unchanged: collateral goes to the caller (repayer C)
	assert.Equal(t, record.CollateralValue, repayOwnerCoinBalance(env, string(Nodes[2]))-coinCBefore,
		"pre-fork: collateral goes to the repayer")
	assert.Equal(t, coinBBefore, repayOwnerCoinBalance(env, string(Nodes[1])),
		"pre-fork: borrower gets nothing back")
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// Regression tests for the LotteryBuy amount overflow fix (ForkLotteryFixOverflow):
//  1. buy.Amount * CoinPrecision int64 overflow is rejected after the fork
//  2. pre-fork behavior is unchanged (old bookkeeping kept for consensus compatibility)
//  3. normal buy/draw/close paths are not affected by the fix

import (
	"encoding/binary"
	"testing"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/dapp"
	manageTy "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rt "github.com/33cn/plugin/plugin/dapp/lottery/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixOverflowBuyAmount makes amount*1e8 wrap around to a small positive value 90448384 (~0.9 coin),
// while Fund/AmountOneRound would be recorded as 184467440738 before the fix
const fixOverflowBuyAmount = int64(184467440738)

type lotteryFixOverflowEnv struct {
	exec      drivers.Driver
	db        dbm.KV
	q         queue.Queue
	cfg       *types.Chain33Config
	creator   string
	creatorPk crypto.PrivKey
	buyer     string
	buyerPk   crypto.PrivKey
	lotteryID string
	nonce     int64
}

// newLotteryFixOverflowEnv builds an isolated lottery executor env.
// The driver is registered once by the existing init() in lottery_test.go.
func newLotteryFixOverflowEnv(t *testing.T, cfg *types.Chain33Config) *lotteryFixOverflowEnv {
	env := &lotteryFixOverflowEnv{}
	env.creator, env.creatorPk = genaddress()
	env.buyer, env.buyerPk = genaddress()

	if cfg == nil {
		cfg = types.NewChain33Config(types.GetDefaultCfgstring())
	}
	env.cfg = cfg
	lott := newLottery()
	_, _, kvdb := util.CreateTestDB()
	env.db = kvdb
	q := queue.New("channel")
	q.SetConfig(cfg)
	env.q = q
	api, _ := client.New(q.Client(), nil)
	lott.SetAPI(api)
	lott.SetExecutorAPI(api, nil)
	lott.SetStateDB(kvdb)
	lott.SetLocalDB(kvdb)
	env.exec = lott

	// configure lottery-creator manage privilege
	var item types.ConfigItem
	item.Key = creatorKey
	item.Ty = manageTy.ConfigItemArrayConfig
	arr := types.ConfigItem_Arr{Arr: &types.ArrayConfig{Value: []string{env.creator}}}
	item.Value = &arr
	err := env.db.Set([]byte(types.ManageKey(creatorKey)), types.Encode(&item))
	require.Nil(t, err)
	return env
}

func (env *lotteryFixOverflowEnv) nextNonce() int64 {
	env.nonce++
	return env.nonce
}

func (env *lotteryFixOverflowEnv) createLottery(t *testing.T) {
	vc := &rt.LotteryAction_Create{Create: &rt.LotteryCreate{
		PurBlockNum: minPurBlockNum, DrawBlockNum: minDrawBlockNum, OpRewardRatio: 10, DevRewardRatio: 10}}
	action := &rt.LotteryAction{Value: vc, Ty: rt.LotteryActionCreate}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(action), Fee: 1e6,
		To: address.ExecAddress(env.cfg.ExecName(rt.LotteryX)), Nonce: env.nextNonce()}
	tx.Sign(types.SECP256K1, env.creatorPk)
	env.exec.SetEnv(1, 1000, 0)
	_, err := env.exec.Exec(tx, 0)
	require.Nil(t, err)
	env.lotteryID = common.ToHex(tx.Hash())
}

func (env *lotteryFixOverflowEnv) buyTx(amount, number, way int64) *types.Transaction {
	vb := &rt.LotteryAction_Buy{Buy: &rt.LotteryBuy{LotteryId: env.lotteryID, Amount: amount, Number: number, Way: way}}
	action := &rt.LotteryAction{Value: vb, Ty: rt.LotteryActionBuy}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(action), Fee: 1e6,
		To: address.ExecAddress(env.cfg.ExecName(rt.LotteryX)), Nonce: env.nextNonce()}
	tx.Sign(types.SECP256K1, env.buyerPk)
	return tx
}

func (env *lotteryFixOverflowEnv) drawTx() *types.Transaction {
	vd := &rt.LotteryAction_Draw{Draw: &rt.LotteryDraw{LotteryId: env.lotteryID}}
	action := &rt.LotteryAction{Value: vd, Ty: rt.LotteryActionDraw}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(action), Fee: 1e6,
		To: address.ExecAddress(env.cfg.ExecName(rt.LotteryX)), Nonce: env.nextNonce()}
	tx.Sign(types.SECP256K1, env.creatorPk)
	return tx
}

func (env *lotteryFixOverflowEnv) closeTx() *types.Transaction {
	vc := &rt.LotteryAction_Close{Close: &rt.LotteryClose{LotteryId: env.lotteryID}}
	action := &rt.LotteryAction{Value: vc, Ty: rt.LotteryActionClose}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(action), Fee: 1e6,
		To: address.ExecAddress(env.cfg.ExecName(rt.LotteryX)), Nonce: env.nextNonce()}
	tx.Sign(types.SECP256K1, env.creatorPk)
	return tx
}

func (env *lotteryFixOverflowEnv) fundBuyer(balance int64) {
	execaddr := address.ExecAddress("lottery")
	acc := env.exec.GetCoinsAccount().LoadExecAccount(env.buyer, execaddr)
	acc.Balance = balance
	env.exec.GetCoinsAccount().SaveExecAccount(execaddr, acc)
}

func (env *lotteryFixOverflowEnv) buyerBalance() int64 {
	return env.exec.GetCoinsAccount().LoadExecAccount(env.buyer, address.ExecAddress("lottery")).Balance
}

func (env *lotteryFixOverflowEnv) creatorFrozen() int64 {
	return env.exec.GetCoinsAccount().LoadExecAccount(env.creator, address.ExecAddress("lottery")).Frozen
}

func (env *lotteryFixOverflowEnv) loadLottery(t *testing.T) *rt.Lottery {
	lott, err := findLottery(env.db, env.lotteryID)
	require.Nil(t, err)
	return lott
}

// mockRandNum mocks the ticket module RandNumHash query to control the lucky number
func (env *lotteryFixOverflowEnv) mockRandNum(luckyNum uint32) {
	luckyHash := make([]byte, 32)
	binary.BigEndian.PutUint32(luckyHash[0:4], luckyNum)
	c := env.q.Client()
	c.Sub("execs")
	go func() {
		for msg := range c.Recv() {
			if msg.Ty == types.EventBlockChainQuery {
				data, ok := msg.GetData().(*types.ChainExecutor)
				if ok && data.Driver == "ticket" && data.FuncName == "RandNumHash" {
					msg.Reply(c.NewMessage("", types.EventBlockChainQuery, &types.ReplyHash{Hash: luckyHash}))
					continue
				}
				msg.Reply(c.NewMessage("", types.EventBlockChainQuery, types.ErrNotSupport))
			}
		}
	}()
}

// TestLotteryFixBuyOverflowRejected verifies that an overflowing buy amount is rejected after ForkLotteryFixOverflow
func TestLotteryFixBuyOverflowRejected(t *testing.T) {
	env := newLotteryFixOverflowEnv(t, nil)
	require.True(t, env.cfg.IsDappFork(0, rt.LotteryX, rt.ForkLotteryFixOverflow))
	env.createLottery(t)
	env.fundBuyer(1e9)

	// amount*CoinPrecision wraps to a small positive value, must be rejected
	env.exec.SetEnv(2, 1001, 0)
	_, err := env.exec.Exec(env.buyTx(fixOverflowBuyAmount, 12345, FiveStar), 0)
	require.Equal(t, rt.ErrLotteryBuyAmount, err)

	// amount exceeding the max coin supply must be rejected as well
	_, err = env.exec.Exec(env.buyTx(types.MaxCoin, 12345, FiveStar), 0)
	require.Equal(t, rt.ErrLotteryBuyAmount, err)

	// no funds moved, no bookkeeping changed
	assert.Equal(t, int64(1e9), env.buyerBalance())
	assert.Equal(t, int64(0), env.creatorFrozen())
	lott := env.loadLottery(t)
	assert.Equal(t, int64(0), lott.Fund)
	assert.Equal(t, int64(0), lott.BuyAmount)
}

// TestLotteryFixNormalBuyDraw verifies the normal buy/draw payout path is not affected
func TestLotteryFixNormalBuyDraw(t *testing.T) {
	env := newLotteryFixOverflowEnv(t, nil)
	env.createLottery(t)
	env.fundBuyer(100 * 1e8)
	env.mockRandNum(12345)

	env.exec.SetEnv(2, 1001, 0)
	// buy 10 coins, FiveStar, number 12345, always wins with odds 100000
	_, err := env.exec.Exec(env.buyTx(10, 12345, FiveStar), 0)
	require.Nil(t, err)
	assert.Equal(t, int64(10)*1e8, env.creatorFrozen())

	env.exec.SetEnv(2+minDrawBlockNum, 2000, 0)
	_, err = env.exec.Exec(env.drawTx(), 0)
	require.Nil(t, err)

	lott := env.loadLottery(t)
	assert.EqualValues(t, rt.LotteryDrawed, lott.Status)
	assert.Equal(t, int64(12345), lott.LuckyNumber)
	// totalFund=10*100000=1e6 > Fund/2=5 => factor=1e8*10/2/1e6=500
	// gain = 1e6*500*(1000-10-10)/1000 = 4.9e8; op=dev=500*1e6*10/1000=5e6
	assert.Equal(t, int64(5), lott.Fund)
	assert.Equal(t, int64(100)*1e8-int64(10)*1e8+int64(49)*1e7, env.buyerBalance())
	assert.Equal(t, int64(10)*1e8-int64(49)*1e7-2*int64(5)*1e6, env.creatorFrozen())
}

// TestLotteryFixNormalBuyClose verifies the normal close refund path is not affected
func TestLotteryFixNormalBuyClose(t *testing.T) {
	env := newLotteryFixOverflowEnv(t, nil)
	env.createLottery(t)
	env.fundBuyer(100 * 1e8)

	env.exec.SetEnv(2, 1001, 0)
	_, err := env.exec.Exec(env.buyTx(10, 12345, FiveStar), 0)
	require.Nil(t, err)

	env.exec.SetEnv(3, 1002, 0)
	_, err = env.exec.Exec(env.closeTx(), 0)
	require.Nil(t, err)
	assert.Equal(t, int64(100)*1e8, env.buyerBalance())
	assert.Equal(t, int64(0), env.creatorFrozen())

	lott := env.loadLottery(t)
	assert.EqualValues(t, rt.LotteryClosed, lott.Status)
}

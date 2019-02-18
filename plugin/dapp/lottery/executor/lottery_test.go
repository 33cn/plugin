// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/dapp"
	pty "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rt "github.com/33cn/plugin/plugin/dapp/lottery/types"
)

var (
	creatorAddr string
	buyAddr     string
	buyPriv     crypto.PrivKey
	creatorPriv crypto.PrivKey
	testNormErr error
	lottery     drivers.Driver
	r           *rand.Rand
	mydb        db.KV
	lotteryID   string
)

func init() {
	creatorAddr, creatorPriv = genaddress()
	buyAddr, buyPriv = genaddress()
	testNormErr = errors.New("Err")
	lottery = constructLotteryInstance()
	r = rand.New(rand.NewSource(types.Now().UnixNano()))
}

func TestExecCreateLottery(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr = rt.ErrNoPrivilege
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructCreateTx()
	receipt, err = lottery.Exec(tx, 0)

	// ErrNoPrivilege case
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}

	var item types.ConfigItem
	item.Key = "lottery-creator"
	item.Addr = creatorAddr
	item.Ty = pty.ConfigItemArrayConfig
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, creatorAddr)
	item.Addr = creatorAddr

	key := types.ManageKey("lottery-creator")
	valueSave := types.Encode(&item)
	mydb.Set([]byte(key), valueSave)

	// success case
	targetErr = nil
	receipt, err = lottery.Exec(tx, 0)
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}

	lotteryID = common.ToHex(tx.Hash())
	fmt.Println(lotteryID)
}

func TestExecBuyLottery(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr = types.ErrNoBalance
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructBuyTx()
	receipt, err = lottery.Exec(tx, 0)

	// ErrNoBalance case
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}

	acc1 := lottery.GetCoinsAccount().LoadExecAccount(buyAddr, address.ExecAddress("lottery"))
	acc1.Balance = 10000000000
	lottery.GetCoinsAccount().SaveExecAccount(address.ExecAddress("lottery"), acc1)

	targetErr = nil
	receipt, err = lottery.Exec(tx, 0)
	// success case
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecDrawLottery(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr = rt.ErrLotteryStatus
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructDrawTx()
	receipt, err = lottery.Exec(tx, 0)

	// ErrLotteryStatus case
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
	lottery.SetEnv(100, 0, 0)
	receipt, err = lottery.Exec(tx, 0)
	targetErr = types.ErrActionNotSupport
	// ErrActionNotSupport case
	if !CompareLotteryExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}

	// mock message between randnum nextstep
}

func genaddress() (string, crypto.PrivKey) {
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddress(privto.PubKey().Bytes())
	return addrto.String(), privto
}

func NewTestDB() db.KV {
	return executor.NewStateDB(nil, nil, nil, &executor.StateDBOption{Height: types.GetFork("ForkExecRollback")})
}

func ConstructCreateTx() *types.Transaction {

	var purBlockNum int64 = 30
	var drawBlockNum int64 = 40
	var opRatio int64 = 10
	var devRatio int64 = 10
	var fee int64 = 1e6

	vcreate := &rt.LotteryAction_Create{Create: &rt.LotteryCreate{PurBlockNum: purBlockNum, DrawBlockNum: drawBlockNum, OpRewardRatio: opRatio, DevRewardRatio: devRatio}}

	transfer := &rt.LotteryAction{Value: vcreate, Ty: rt.LotteryActionCreate}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(transfer), Fee: fee, To: address.ExecAddress(types.ExecName(rt.LotteryX))}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, creatorPriv)
	return tx
}

func ConstructBuyTx() *types.Transaction {

	var amount int64 = 1
	var number int64 = 12345
	var way int64 = 1
	var fee int64 = 1e6

	vbuy := &rt.LotteryAction_Buy{Buy: &rt.LotteryBuy{LotteryId: lotteryID, Amount: amount, Number: number, Way: way}}

	transfer := &rt.LotteryAction{Value: vbuy, Ty: rt.LotteryActionBuy}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(transfer), Fee: fee, To: address.ExecAddress(types.ExecName(rt.LotteryX))}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, buyPriv)
	return tx
}

func ConstructDrawTx() *types.Transaction {

	var fee int64 = 1e6

	vdraw := &rt.LotteryAction_Draw{Draw: &rt.LotteryDraw{LotteryId: lotteryID}}

	transfer := &rt.LotteryAction{Value: vdraw, Ty: rt.LotteryActionDraw}
	tx := &types.Transaction{Execer: []byte("lottery"), Payload: types.Encode(transfer), Fee: fee, To: address.ExecAddress(types.ExecName(rt.LotteryX))}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, creatorPriv)
	return tx
}

func constructLotteryInstance() drivers.Driver {
	lottery := newLottery()
	//lottery.SetStateDB(NewTestDB())
	_, _, kvdb := util.CreateTestDB()
	mydb = kvdb
	lottery.SetStateDB(mydb)
	lottery.SetLocalDB(kvdb)
	q := queue.New("channel")
	client.New(q.Client(), nil)
	//lottery.SetAPI(qclient)
	return lottery
}

func CompareLotteryExecLocalRes(dbset1 *types.LocalDBSet, err1 error, dbset2 *types.LocalDBSet, err2 error) bool {
	//fmt.Println(err1, err2, dbset1, dbset2)
	if err1 != err2 {
		fmt.Println(err1, err2)
		return false
	}

	if dbset1 == nil && dbset2 == nil {
		return true
	}

	if (dbset1 == nil) != (dbset2 == nil) {
		return false
	}

	if dbset1.KV == nil && dbset2.KV == nil {
		return true
	}

	if (dbset1.KV == nil) != (dbset2.KV == nil) {
		return false
	}
	if len(dbset1.KV) != len(dbset2.KV) {
		return false
	}

	for i := range dbset1.KV {
		if !bytes.Equal(dbset1.KV[i].Key, dbset2.KV[i].Key) {
			return false
		}
		if !bytes.Equal(dbset1.KV[i].Value, dbset2.KV[i].Value) {
			return false
		}
	}
	return true
}

func CompareLotteryExecResult(rec1 *types.Receipt, err1 error, rec2 *types.Receipt, err2 error) bool {
	//fmt.Println(err1, err2, rec1, rec2)
	if err1 != err2 {
		fmt.Println(err1, err2)
		return false
	}
	// err, so receipt not concerned
	if err1 != nil && err1 == err2 {
		return true
	}
	if (rec1 == nil) != (rec2 == nil) {
		return false
	}
	if rec1.Ty != rec2.Ty {
		fmt.Println(rec1.Ty, rec2.Ty)
		return false
	}
	return true
}

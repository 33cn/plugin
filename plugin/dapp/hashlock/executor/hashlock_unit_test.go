// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	crand "crypto/rand"
	"errors"
	"fmt"
	"testing"

	"math/rand"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	pty "github.com/33cn/plugin/plugin/dapp/hashlock/types"
)

var (
	toAddr      string
	returnAddr  string
	toPriv      crypto.PrivKey
	returnPriv  crypto.PrivKey
	testNormErr error
	hashlock    drivers.Driver
	secret      []byte
	addrexec    string
)

const secretLen = 32

func genaddress() (string, crypto.PrivKey) {
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		panic(err)
	}
	privto, err := cr.GenKey()
	if err != nil {
		panic(err)
	}
	addrto := address.PubKeyToAddr(address.DefaultID, privto.PubKey().Bytes())
	return addrto, privto
}

func TestInit(t *testing.T) {
	toAddr, toPriv = genaddress()
	returnAddr, returnPriv = genaddress()
	testNormErr = errors.New("Err")
	hashlock = constructHashlockInstance()
	secret = make([]byte, secretLen)
	crand.Read(secret)
	addrexec = address.ExecAddress("hashlock")
}

func TestExecHashlock(t *testing.T) {

	var targetReceipt types.Receipt
	var targetErr error
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructLockTx()

	acc1 := hashlock.GetCoinsAccount().LoadExecAccount(returnAddr, addrexec)
	acc1.Balance = 100
	hashlock.GetCoinsAccount().SaveExecAccount(addrexec, acc1)

	receipt, err = hashlock.Exec(tx, 0)

	if !CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

//timelimit
func TestExecHashunlock(t *testing.T) {

	var targetReceipt types.Receipt
	var targetErr = pty.ErrTime
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructUnlockTx()

	receipt, err = hashlock.Exec(tx, 0)

	if CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecHashsend(t *testing.T) {

	var targetReceipt types.Receipt
	var targetErr error
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructSendTx()

	receipt, err = hashlock.Exec(tx, 0)

	if !CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

func constructHashlockInstance() drivers.Driver {
	chainTestCfg := types.NewChain33Config(types.GetDefaultCfgstring())
	Init(pty.HashlockX, chainTestCfg, nil)
	h := newHashlock()
	q := queue.New("channel")
	q.SetConfig(chainTestCfg)
	api, _ := client.New(q.Client(), nil)
	h.SetAPI(api)
	_, _, kvdb := util.CreateTestDB()
	h.SetStateDB(kvdb)
	return h
}

func ConstructLockTx() *types.Transaction {

	var lockAmount int64 = 90
	var locktime int64 = 70
	var fee int64 = 1e6

	vlock := &pty.HashlockAction_Hlock{Hlock: &pty.HashlockLock{Amount: lockAmount, Time: locktime, Hash: common.Sha256(secret), ToAddress: toAddr, ReturnAddress: returnAddr}}
	transfer := &pty.HashlockAction{Value: vlock, Ty: pty.HashlockActionLock}
	tx := &types.Transaction{Execer: []byte("hashlock"), Payload: types.Encode(transfer), Fee: fee, To: toAddr}
	tx.Nonce = rand.Int63()
	tx.Sign(types.SECP256K1, returnPriv)

	return tx
}

func ConstructUnlockTx() *types.Transaction {

	var fee int64 = 1e6

	vunlock := &pty.HashlockAction_Hunlock{Hunlock: &pty.HashlockUnlock{Secret: secret}}
	transfer := &pty.HashlockAction{Value: vunlock, Ty: pty.HashlockActionUnlock}
	tx := &types.Transaction{Execer: []byte("hashlock"), Payload: types.Encode(transfer), Fee: fee, To: toAddr}
	tx.Nonce = rand.Int63()
	tx.Sign(types.SECP256K1, returnPriv)
	return tx
}

func ConstructSendTx() *types.Transaction {

	var fee int64 = 1e6

	vsend := &pty.HashlockAction_Hsend{Hsend: &pty.HashlockSend{Secret: secret}}
	transfer := &pty.HashlockAction{Value: vsend, Ty: pty.HashlockActionSend}
	tx := &types.Transaction{Execer: []byte("hashlock"), Payload: types.Encode(transfer), Fee: fee, To: toAddr}
	tx.Nonce = rand.Int63()
	tx.Sign(types.SECP256K1, toPriv)
	return tx
}

func CompareRetrieveExecResult(rec1 *types.Receipt, err1 error, rec2 *types.Receipt, err2 error) bool {
	if err1 != err2 {
		fmt.Println(err1, err2)
		return false
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

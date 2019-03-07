// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/executor"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
)

var (
	backupAddr  string
	defaultAddr string
	backupPriv  crypto.PrivKey
	defaultPriv crypto.PrivKey
	testNormErr error
	retrieve    drivers.Driver
)

func init() {
	backupAddr, backupPriv = genaddress()
	defaultAddr, defaultPriv = genaddress()
	testNormErr = errors.New("Err")
	retrieve = constructRetrieveInstance()
}

func NewTestDB() db.KV {
	return executor.NewStateDB(nil, nil, nil, &executor.StateDBOption{Height: types.GetFork("ForkExecRollback")})
}
func TestExecBackup(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr error
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructBackupTx()
	receipt, err = retrieve.Exec(tx, 0)

	if !CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecPrepare(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr error
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructPrepareTx()
	receipt, err = retrieve.Exec(tx, 0)

	if !CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

//timelimit
func TestExecPerform(t *testing.T) {
	var targetReceipt types.Receipt
	var targetErr = rt.ErrRetrievePeriodLimit
	var receipt *types.Receipt
	var err error
	targetReceipt.Ty = 2
	tx := ConstructPerformTx()
	receipt, err = retrieve.Exec(tx, 0)

	if CompareRetrieveExecResult(receipt, err, &targetReceipt, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecLocalBackup(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	info := rt.RetrieveQuery{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: 70, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrieveBackup}
	value := types.Encode(&info)

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: value}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructBackupTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecLocalPrepare(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	info := rt.RetrieveQuery{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: 70, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrievePrepare}
	value := types.Encode(&info)

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: value}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructPrepareTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecLocalPerform(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	info := rt.RetrieveQuery{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: 70, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrievePerform}
	value := types.Encode(&info)

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: value}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructPerformTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecDelLocalPerform(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	info := rt.RetrieveQuery{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: 70, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrievePrepare}
	value := types.Encode(&info)

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: value}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructPerformTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecDelLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecDelLocalPrepare(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	info := rt.RetrieveQuery{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: 70, PrepareTime: zeroPrepareTime, RemainTime: zeroRemainTime, Status: retrieveBackup}
	value := types.Encode(&info)

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: value}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructPrepareTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecDelLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func TestExecDelLocalBackup(t *testing.T) {
	var targetDBSet types.LocalDBSet
	var targetErr error
	var dbset *types.LocalDBSet
	var err error

	kv := &types.KeyValue{Key: calcRetrieveKey(backupAddr, defaultAddr), Value: nil}
	targetDBSet.KV = append(targetDBSet.KV, kv)

	tx := ConstructBackupTx()
	var receiptData types.ReceiptData
	receiptData.Ty = types.ExecOk

	dbset, err = retrieve.ExecDelLocal(tx, &receiptData, 0)
	if err != nil {
		t.Error(testNormErr)
	}

	if !CompareRetrieveExecLocalRes(&targetDBSet, err, dbset, targetErr) {
		t.Error(testNormErr)
	}
}

func constructRetrieveInstance() drivers.Driver {
	r := newRetrieve()
	_, _, kvdb := util.CreateTestDB()
	r.SetStateDB(kvdb)
	r.SetLocalDB(kvdb)
	return r
}

func ConstructBackupTx() *types.Transaction {

	var delayPeriod int64 = 70
	var fee int64 = 1e6

	vbackup := &rt.RetrieveAction_Backup{Backup: &rt.BackupRetrieve{BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: delayPeriod}}
	//fmt.Println(vlock)
	transfer := &rt.RetrieveAction{Value: vbackup, Ty: rt.RetrieveActionBackup}
	tx := &types.Transaction{Execer: []byte("retrieve"), Payload: types.Encode(transfer), Fee: fee, To: backupAddr}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, defaultPriv)
	return tx
}

func ConstructPrepareTx() *types.Transaction {
	var fee int64 = 1e6
	vprepare := &rt.RetrieveAction_Prepare{Prepare: &rt.PrepareRetrieve{BackupAddress: backupAddr, DefaultAddress: defaultAddr}}
	transfer := &rt.RetrieveAction{Value: vprepare, Ty: rt.RetrieveActionPrepare}
	tx := &types.Transaction{Execer: []byte("retrieve"), Payload: types.Encode(transfer), Fee: fee, To: backupAddr}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, backupPriv)
	//tx.Sign(types.SECP256K1, defaultPriv)
	return tx
}

func ConstructPerformTx() *types.Transaction {
	var fee int64 = 1e6

	vperform := &rt.RetrieveAction_Perform{Perform: &rt.PerformRetrieve{BackupAddress: backupAddr, DefaultAddress: defaultAddr}}
	transfer := &rt.RetrieveAction{Value: vperform, Ty: rt.RetrieveActionPerform}
	tx := &types.Transaction{Execer: []byte("retrieve"), Payload: types.Encode(transfer), Fee: fee, To: backupAddr}
	tx.Nonce = r.Int63()
	tx.Sign(types.SECP256K1, backupPriv)

	return tx
}

func CompareRetrieveExecLocalRes(dbset1 *types.LocalDBSet, err1 error, dbset2 *types.LocalDBSet, err2 error) bool {
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

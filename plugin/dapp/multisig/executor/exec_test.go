// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/common"
	commonlog "github.com/33cn/chain33/common/log"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/stretchr/testify/assert"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
)

type execEnv struct {
	blockTime   int64 // 1539918074
	blockHeight int64
	index       int
	difficulty  uint64
	txHash      string
}

var (
	Symbol = "BTY"
	Asset  = "coins"

	PrivKeyA = "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b" // 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4
	PrivKeyB = "0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4" // 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
	PrivKeyC = "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115" // 1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k
	PrivKeyD = "0xcacb1f5d51700aea07fca2246ab43b0917d70405c65edea9b5063d72eb5c6b71" // 1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs
	AddrA    = "1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4"
	AddrB    = "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR"
	AddrC    = "1NLHPEcbTWWxxU3dGUZBhayjrCHD3psX7k"
	AddrD    = "1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs"

	AddrBWeight           uint64 = 1
	AddrCWeight           uint64 = 4
	AddrDWeight           uint64 = 10
	NewWeight             uint64 = 2
	Requiredweight        uint64 = 5
	NewRequiredweight     uint64 = 4
	CoinsBtyDailylimit    uint64 = 100
	NewCoinsBtyDailylimit uint64 = 10
	PrintFlag                    = false
	InAmount              int64  = 10
	OutAmount             int64  = 5
)

func init() {
	commonlog.SetLogLevel("debug")
	types.AllowUserExec = append(types.AllowUserExec, []byte("coins"))

}

//创建一个多重签名的账户
func TestMultiSigAccCreate(t *testing.T) {

	total := int64(100000)
	accountA := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrA,
	}

	accountD := types.Account{
		Balance: total,
		Frozen:  0,
		Addr:    AddrD,
	}

	env := execEnv{
		1539918074,
		types.GetDappFork("multisig", "ForkMultiSigV1"),
		2,
		1539918074,
		"hash",
	}

	stateDB, _ := dbm.NewGoMemDB("state", "state", 100)
	localDB := new(dbmock.KVDB)
	api := new(apimock.QueueProtocolAPI)

	// 给账户accountB在multisig合约中写入coins-bty资产
	accB := account.NewCoinsAccount()
	accB.SetDB(stateDB)
	accB.SaveExecAccount(address.ExecAddress("multisig"), &accountA)

	// 给账户accountD在multisig合约中写入coins-bty资产
	accD := account.NewCoinsAccount()
	accD.SetDB(stateDB)
	accD.SaveExecAccount(address.ExecAddress("multisig"), &accountD)

	//accA, _ := account.NewAccountDB(AssetExecToken, Symbol, stateDB)
	//accA.SaveExecAccount(address.ExecAddress("multisig"), &accountA)

	driver := newMultiSig()
	driver.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	driver.SetStateDB(stateDB)
	driver.SetLocalDB(localDB)
	driver.SetAPI(api)

	//add create MultiSigAcc
	multiSigAddr, err := testMultiSigAccCreate(t, driver, env, localDB)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}
	//add addrB
	//t.Log("------testMultiSigOwnerAdd------")
	testMultiSigOwnerAdd(t, driver, env, multiSigAddr, AddrB, AddrBWeight)
	//del addrB
	//t.Log("------testMultiSigOwnerDel------")
	testMultiSigOwnerDel(t, driver, env, multiSigAddr)
	//modify addrC
	//t.Log("------testMultiSigOwnerModify------")
	testMultiSigOwnerModify(t, driver, env, multiSigAddr)
	//replace addrC with addrB
	//t.Log("------testMultiSigOwnerReplace------")
	testMultiSigOwnerReplace(t, driver, env, multiSigAddr)
	//modify NewRequiredweight
	//t.Log("------testMultiSigAccWeightModify------")
	testMultiSigAccWeightModify(t, driver, env, multiSigAddr)

	//modify assets DailyLimit NewCoinsBtyDailylimit
	//t.Log("------testMultiSigAccDailyLimitModify------")
	testMultiSigAccDailyLimitModify(t, driver, env, multiSigAddr)

	//confirmtx  assets DailyLimit NewCoinsBtyDailylimit
	//t.Log("------testMultiSigAccConfirmTx------")
	testMultiSigAccConfirmTx(t, driver, env, api, multiSigAddr)

	//从外部账户转账到多重签名账户
	//t.Log("------testMultiSigAccExecTransferTo------")
	testMultiSigAccExecTransferTo(t, driver, env, multiSigAddr)

	//从多重签名账户转账到外部账户
	//t.Log("------testMultiSigAccExecTransferFrom------")
	testMultiSigAccExecTransferFrom(t, driver, env, multiSigAddr)

}

func testMultiSigAccCreate(t *testing.T, driver drivers.Driver, env execEnv, localDB *dbmock.KVDB) (string, error) {
	//---------测试账户创建--------------------
	var owners []*mty.Owner
	owmer1 := &mty.Owner{OwnerAddr: AddrC, Weight: AddrCWeight}
	owners = append(owners, owmer1)
	owmer2 := &mty.Owner{OwnerAddr: AddrD, Weight: AddrDWeight}
	owners = append(owners, owmer2)

	symboldailylimit := &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     "coins",
		DailyLimit: CoinsBtyDailylimit,
	}

	param := &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: Requiredweight,
		DailyLimit:     symboldailylimit,
	}

	tx, _ := multiSigAccCreate(param)
	tx, _ = signTx(tx, PrivKeyA)

	addr := address.MultiSignAddress(tx.Hash())
	localDB.On("Get", calcMultiSigAcc(addr)).Return(nil, types.ErrNotFound)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return "", err
	}

	//解析kv
	var multiSigAccount mty.MultiSig
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	////t.Log(multiSigAccount)

	//解析log
	var receiptMultiSig mty.MultiSig
	ty := receipt.Logs[0].Ty
	////t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSig)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSig)
	assert.Equal(t, int(ty), mty.TyLogMultiSigAccCreate)

	return receiptMultiSig.MultiSigAddr, nil

}

func testMultiSigOwnerAdd(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr, addr string, weight uint64) {
	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		NewOwner:        addr,
		NewWeight:       weight,
		OperateFlag:     mty.OwnerAdd,
	}

	tx, _ := multiSigOwnerOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigOwnerAdd kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log0
	var receiptMultiSigOwnerAddOrDel mty.ReceiptOwnerAddOrDel
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigOwnerAddOrDel)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigOwnerAddOrDel)
	assert.Equal(t, int(ty), mty.TyLogMultiSigOwnerAdd)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)

	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
}

func testMultiSigOwnerDel(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        AddrB,
		OperateFlag:     mty.OwnerDel,
	}

	tx, _ := multiSigOwnerOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigOwnerDel kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log0
	var receiptMultiSigOwnerAddOrDel mty.ReceiptOwnerAddOrDel
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigOwnerAddOrDel)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigOwnerAddOrDel)
	assert.Equal(t, int(ty), mty.TyLogMultiSigOwnerDel)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)

	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
}
func testMultiSigOwnerModify(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        AddrC,
		NewWeight:       NewWeight,
		OperateFlag:     mty.OwnerModify,
	}
	tx, _ := multiSigOwnerOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigOwnerModify kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log0
	var receiptMultiSigOwnerModOrRep mty.ReceiptOwnerModOrRep
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigOwnerModOrRep)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigOwnerModOrRep)
	assert.Equal(t, int(ty), mty.TyLogMultiSigOwnerModify)

	assert.Equal(t, receiptMultiSigOwnerModOrRep.PrevOwner.OwnerAddr, AddrC)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.CurrentOwner.OwnerAddr, AddrC)

	//assert.Equal(t, receiptMultiSigOwnerModOrRep.PrevOwner.Weight, addrC)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.CurrentOwner.Weight, NewWeight)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.ModOrRep, true)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)
	assert.Equal(t, uint64(3), receiptTxCountUpdate.CurTxCount)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)

	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
}
func testMultiSigOwnerReplace(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAddr,
		OldOwner:        AddrC,
		NewWeight:       AddrBWeight,
		NewOwner:        AddrB,
		OperateFlag:     mty.OwnerReplace,
	}
	tx, _ := multiSigOwnerOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigOwnerReplace kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log0
	var receiptMultiSigOwnerModOrRep mty.ReceiptOwnerModOrRep
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigOwnerModOrRep)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigOwnerModOrRep)
	assert.Equal(t, int(ty), mty.TyLogMultiSigOwnerReplace)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.PrevOwner.OwnerAddr, AddrC)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.CurrentOwner.OwnerAddr, AddrB)
	assert.Equal(t, receiptMultiSigOwnerModOrRep.ModOrRep, false)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)
	assert.Equal(t, uint64(4), receiptTxCountUpdate.CurTxCount)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)
	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
}

//account 操作测试
func testMultiSigAccWeightModify(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigAccOperate{
		MultiSigAccAddr:   multiSigAddr,
		NewRequiredWeight: NewRequiredweight,
		OperateFlag:       mty.AccWeightOp,
	}

	tx, _ := multiSigAccOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigAccWeightModify kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log0
	var receiptMultiSigWeightModify mty.ReceiptWeightModify
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigWeightModify)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigWeightModify)

	assert.Equal(t, int(ty), mty.TyLogMultiSigAccWeightModify)

	assert.Equal(t, Requiredweight, receiptMultiSigWeightModify.PrevWeight)
	assert.Equal(t, NewRequiredweight, receiptMultiSigWeightModify.CurrentWeight)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)
	assert.Equal(t, uint64(5), receiptTxCountUpdate.CurTxCount)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)
	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
}
func testMultiSigAccDailyLimitModify(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	assetsDailyLimit := &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     Asset,
		DailyLimit: NewCoinsBtyDailylimit,
	}
	params := &mty.MultiSigAccOperate{
		MultiSigAccAddr: multiSigAddr,
		DailyLimit:      assetsDailyLimit,
		OperateFlag:     mty.AccDailyLimitOp,
	}

	tx, _ := multiSigAccOperate(params)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig
	//解析kv0
	//t.Log("TyLogMultiSigAccDailyLimitModify kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log0
	var receiptMultiSigDailyLimitOperate mty.ReceiptDailyLimitOperate
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigDailyLimitOperate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigDailyLimitOperate)

	assert.Equal(t, int(ty), mty.TyLogMultiSigAccDailyLimitModify)

	assert.Equal(t, receiptMultiSigDailyLimitOperate.PrevDailyLimit.Symbol, Symbol)
	assert.Equal(t, receiptMultiSigDailyLimitOperate.PrevDailyLimit.Execer, Asset)
	assert.Equal(t, receiptMultiSigDailyLimitOperate.PrevDailyLimit.DailyLimit, CoinsBtyDailylimit)

	assert.Equal(t, receiptMultiSigDailyLimitOperate.PrevDailyLimit.Symbol, Symbol)
	assert.Equal(t, receiptMultiSigDailyLimitOperate.PrevDailyLimit.Execer, Asset)
	assert.Equal(t, receiptMultiSigDailyLimitOperate.CurDailyLimit.DailyLimit, NewCoinsBtyDailylimit)

	//解析kv1
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log1
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)
	assert.Equal(t, uint64(6), receiptTxCountUpdate.CurTxCount)

	//解析kv2
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[2].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)
	//解析log2
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)

}

//account 操作测试
//当前账户的信息，txcount：6 Requiredweight:4
//ownerAddr:"1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR" weight:2
//ownerAddr:"1MCftFynyvG2F4ED5mdHYgziDxx6vDrScs" weight:10 ]
//[symbol:"bty" execer:"coins" dailyLimit:10 ]
//6
// 4}
func testMultiSigAccConfirmTx(t *testing.T, driver drivers.Driver, env execEnv, api *apimock.QueueProtocolAPI, multiSigAddr string) {
	//首先创建一个修改Requiredweight权重的交易。由AddrB提交此交易，权限不够交易不被执行。
	//然后由权重比较高的AddrD owner再次提交确认此交易，交易被执行完成

	params := &mty.MultiSigAccOperate{
		MultiSigAccAddr:   multiSigAddr,
		NewRequiredWeight: Requiredweight,
		OperateFlag:       mty.AccWeightOp,
	}

	tx, _ := multiSigAccOperate(params)
	tx, _ = signTx(tx, PrivKeyB)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	var multiSigAccount mty.MultiSig

	//解析kv&log0
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)

	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)
	assert.Equal(t, uint64(7), receiptTxCountUpdate.CurTxCount)

	//解析kv&log[1]
	//t.Log("TyLogMultiSigTx kv & log ")
	var multiSigTx mty.MultiSigTx
	err = types.Decode(receipt.KV[1].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)

	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
	assert.Equal(t, false, receiptAccExecTransferTx.CurExecuted)

	txid := receiptAccExecTransferTx.MultiSigTxOwner.Txid

	//构造api接口
	txDetails := &types.TransactionDetails{}
	txDetail := &types.TransactionDetail{
		Tx: tx,
	}
	txDetails.Txs = append(txDetails.Txs, txDetail)
	api.On("GetTransactionByHash", &types.ReqHashes{Hashes: [][]byte{tx.Hash()}}).Return(txDetails, nil)

	//让addrB owner撤销此确认交易
	//t.Log("-----MultiSigConfirmTx Revoke  -----")
	param := &mty.MultiSigConfirmTx{
		MultiSigAccAddr: multiSigAddr,
		TxId:            txid,
		ConfirmOrRevoke: false,
	}

	txConfirm, _ := multiSigConfirmTx(param)
	txConfirm, _ = signTx(txConfirm, PrivKeyB)

	receipt, err = driver.Exec(txConfirm, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	//解析kv0
	//t.Log("ReceiptConfirmTx kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)
	//解析log0
	var receiptMultiSigConfirmTx mty.ReceiptConfirmTx
	ty = receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigConfirmTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigConfirmTx)

	assert.Equal(t, int(ty), mty.TyLogMultiSigConfirmTxRevoke)
	assert.Equal(t, false, receiptMultiSigConfirmTx.ConfirmeOrRevoke)
	assert.Equal(t, AddrB, receiptMultiSigConfirmTx.MultiSigTxOwner.ConfirmedOwner.OwnerAddr)

	//t.Log("-----MultiSigConfirmTx Confirm  -----")
	//让addrD owner来确认此交易
	para := &mty.MultiSigConfirmTx{
		MultiSigAccAddr: multiSigAddr,
		TxId:            txid,
		ConfirmOrRevoke: true,
	}

	txConfirm, _ = multiSigConfirmTx(para)
	txConfirm, _ = signTx(txConfirm, PrivKeyD)

	receipt, err = driver.Exec(txConfirm, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}

	//解析kv0
	//t.Log("TyLogMultiSigAccWeightModify kv & log ")
	err = types.Decode(receipt.KV[0].Value, &multiSigAccount)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAccount)
	//解析log0
	var receiptMultiSigWeightModify mty.ReceiptWeightModify
	ty = receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptMultiSigWeightModify)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptMultiSigWeightModify)

	assert.Equal(t, int(ty), mty.TyLogMultiSigAccWeightModify)

	assert.Equal(t, NewRequiredweight, receiptMultiSigWeightModify.PrevWeight)
	assert.Equal(t, Requiredweight, receiptMultiSigWeightModify.CurrentWeight)

	//解析kv&log[1]
	//t.Log("TyLogMultiSigTx kv & log ")
	err = types.Decode(receipt.KV[1].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)

	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)
	assert.Equal(t, true, receiptAccExecTransferTx.CurExecuted)

}

//合约内转账到多重签名账户
func testMultiSigAccExecTransferTo(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigExecTransferTo{
		Symbol:   Symbol,
		Amount:   InAmount,
		Note:     "testMultiSigAccExecTransferTo",
		Execname: Asset,
		To:       multiSigAddr,
	}

	tx, _ := multiSigExecTransferTo(params, false)
	tx, _ = signTx(tx, PrivKeyD)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}
	var acc types.Account
	var receiptExecAccountTransfer types.ReceiptExecAccountTransfer

	//解析kv0
	//t.Log("TyLogExecTransfer kv & log ")
	err = types.Decode(receipt.KV[0].Value, &acc)
	assert.Nil(t, err, "decode account")
	//t.Log(acc)
	//解析log0
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptExecAccountTransfer)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptExecAccountTransfer)

	assert.Equal(t, int(ty), types.TyLogExecTransfer)

	assert.Equal(t, receiptExecAccountTransfer.Current.Balance+InAmount, receiptExecAccountTransfer.Prev.Balance)
	//assert.Equal(t, NewRequiredweight, receiptMultiSigWeightModify.CurrentWeight)

	//解析kv1
	//t.Log("TyLogExecTransfer kv & log ")
	err = types.Decode(receipt.KV[1].Value, &acc)
	assert.Nil(t, err, "decode account")
	//t.Log(acc)
	//解析log0
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptExecAccountTransfer)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptExecAccountTransfer)

	assert.Equal(t, int(ty), types.TyLogExecTransfer)
	assert.Equal(t, receiptExecAccountTransfer.Current.Balance, receiptExecAccountTransfer.Prev.Balance+InAmount)

	//解析kv2
	//t.Log("TyLogExecFrozen kv & log ")
	err = types.Decode(receipt.KV[2].Value, &acc)
	assert.Nil(t, err, "decode account")
	//t.Log(acc)
	//解析log0
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptExecAccountTransfer)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptExecAccountTransfer)

	assert.Equal(t, int(ty), types.TyLogExecFrozen)
	assert.Equal(t, receiptExecAccountTransfer.Prev.Balance, InAmount)
	assert.Equal(t, receiptExecAccountTransfer.Current.Frozen, InAmount)
	assert.Equal(t, receiptExecAccountTransfer.Current.Balance, receiptExecAccountTransfer.Prev.Balance-InAmount)

}

//合约内转账到多重签名账户
func testMultiSigAccExecTransferFrom(t *testing.T, driver drivers.Driver, env execEnv, multiSigAddr string) {
	params := &mty.MultiSigExecTransferFrom{
		Symbol:   Symbol,
		Amount:   OutAmount,
		Note:     "testMultiSigAccExecTransferFrom",
		Execname: Asset,
		From:     multiSigAddr,
		To:       AddrD,
	}

	tx, _ := multiSigExecTransferFrom(params, true)
	tx, _ = signTx(tx, PrivKeyB)

	receipt, err := driver.Exec(tx, env.index)
	if err != nil {
		assert.Nil(t, err, "exec failed")
		return
	}
	var multiSigAcc mty.MultiSig
	var acc types.Account
	var receiptExecAccountTransfer types.ReceiptExecAccountTransfer

	//解析kv0
	//t.Log("TyLogExecTransfer kv & log ")
	err = types.Decode(receipt.KV[0].Value, &acc)
	assert.Nil(t, err, "decode account")
	//t.Log(acc)
	//解析log0
	ty := receipt.Logs[0].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[0].Log, &receiptExecAccountTransfer)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptExecAccountTransfer)
	assert.Equal(t, int(ty), types.TyLogExecTransfer)
	assert.Equal(t, receiptExecAccountTransfer.Current.Frozen, receiptExecAccountTransfer.Prev.Frozen-OutAmount)

	//解析kv1
	//t.Log("TyLogExecTransfer kv & log ")
	err = types.Decode(receipt.KV[1].Value, &acc)
	assert.Nil(t, err, "decode account")
	//t.Log(acc)
	ty = receipt.Logs[1].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[1].Log, &receiptExecAccountTransfer)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptExecAccountTransfer)

	assert.Equal(t, int(ty), types.TyLogExecTransfer)
	assert.Equal(t, receiptExecAccountTransfer.Current.Balance, receiptExecAccountTransfer.Prev.Balance+OutAmount)

	//解析kv2
	var receiptTxCountUpdate mty.ReceiptTxCountUpdate
	//t.Log("TyLogTxCountUpdate kv & log ")
	err = types.Decode(receipt.KV[2].Value, &multiSigAcc)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAcc)
	ty = receipt.Logs[2].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[2].Log, &receiptTxCountUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptTxCountUpdate)
	assert.Equal(t, int(ty), mty.TyLogTxCountUpdate)

	//解析kv3
	var receiptAccDailyLimitUpdate mty.ReceiptAccDailyLimitUpdate
	//t.Log("TyLogDailyLimitUpdate kv & log ")
	err = types.Decode(receipt.KV[3].Value, &multiSigAcc)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigAcc)
	ty = receipt.Logs[3].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[3].Log, &receiptAccDailyLimitUpdate)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccDailyLimitUpdate)
	assert.Equal(t, int(ty), mty.TyLogDailyLimitUpdate)

	//解析kv4
	var receiptAccExecTransferTx mty.ReceiptMultiSigTx
	var multiSigTx mty.MultiSigTx
	//t.Log("TyLogExecTransfer kv & log ")
	err = types.Decode(receipt.KV[4].Value, &multiSigTx)
	assert.Nil(t, err, "decode account")
	//t.Log(multiSigTx)
	//解析log0
	ty = receipt.Logs[4].Ty
	//t.Log(ty)
	err = types.Decode(receipt.Logs[4].Log, &receiptAccExecTransferTx)
	assert.Nil(t, err, "decode Logs")
	//t.Log(receiptAccExecTransferTx)
	assert.Equal(t, int(ty), mty.TyLogMultiSigTx)

}

func signTx(tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.New(types.GetSignName(mty.MultiSigX, signType))
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

func multiSigAccCreate(parm *mty.MultiSigAccCreate) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigAccCreate,
		Value: &mty.MultiSigAction_MultiSigAccCreate{MultiSigAccCreate: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}

func multiSigOwnerOperate(parm *mty.MultiSigOwnerOperate) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigOwnerOperate,
		Value: &mty.MultiSigAction_MultiSigOwnerOperate{MultiSigOwnerOperate: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}

func multiSigAccOperate(parm *mty.MultiSigAccOperate) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigAccOperate,
		Value: &mty.MultiSigAction_MultiSigAccOperate{MultiSigAccOperate: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}

func multiSigConfirmTx(parm *mty.MultiSigConfirmTx) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigConfirmTx,
		Value: &mty.MultiSigAction_MultiSigConfirmTx{MultiSigConfirmTx: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}
func multiSigExecTransferTo(parm *mty.MultiSigExecTransferTo, fromOrTo bool) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigExecTransferTo,
		Value: &mty.MultiSigAction_MultiSigExecTransferTo{MultiSigExecTransferTo: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}

func multiSigExecTransferFrom(parm *mty.MultiSigExecTransferFrom, fromOrTo bool) (*types.Transaction, error) {
	if parm == nil {
		return nil, types.ErrInvalidParam
	}
	multiSig := &mty.MultiSigAction{
		Ty:    mty.ActionMultiSigExecTransferFrom,
		Value: &mty.MultiSigAction_MultiSigExecTransferFrom{MultiSigExecTransferFrom: parm},
	}
	return types.CreateFormatTx(types.ExecName(mty.MultiSigX), types.Encode(multiSig))
}

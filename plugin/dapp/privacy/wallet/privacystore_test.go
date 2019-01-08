/*
 * Copyright Fuzamei Corp. 2018 All Rights Reserved.
 * Use of this source code is governed by a BSD-style
 * license that can be found in the LICENSE file.
 */

package wallet

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"

	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/util/testnode"
	pt "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/stretchr/testify/assert"
)

func createStore(t *testing.T) *privacyStore {
	cfg, _ := testnode.GetDefaultConfig()
	util.ResetDatadir(cfg, "$TEMP/")
	cfgWallet := cfg.Wallet
	walletStoreDB := dbm.NewDB("wallet", cfgWallet.Driver, cfgWallet.DbPath, cfgWallet.DbCache)
	store := newStore(walletStoreDB)
	assert.NotNil(t, store)
	return store
}

func TestPrivacyStore(t *testing.T) {
	testStore_getVersion(t)
	testStore_setVersion(t)
	testStore_getAccountByPrefix(t)
	testStore_getAccountByAddr(t)
	testStore_setWalletAccountPrivacy(t)
	testStore_listAvailableUTXOs(t)
	testStore_listFrozenUTXOs(t)
	testStore_getWalletPrivacyTxDetails(t)
	testStore_getPrivacyTokenUTXOs(t)
	testStore_moveUTXO2FTXO(t)
	testStore_getRescanUtxosFlag4Addr(t)
	testStore_saveREscanUTXOsAddresses(t)
	testStore_setScanPrivacyInputUTXO(t)
	testStore_isUTXOExist(t)
	testStore_updateScanInputUTXOs(t)
	testStore_moveUTXO2STXO(t)
	testStore_selectPrivacyTransactionToWallet(t)
	testStore_setUTXO(t)
	testStore_storeScanPrivacyInputUTXO(t)
	testStore_listSpendUTXOs(t)
	testStore_getWalletFtxoStxo(t)
	testStore_getFTXOlist(t)
	testStore_moveFTXO2STXO(t)
	testStore_moveFTXO2UTXO(t)
	testStore_unsetUTXO(t)
	testStore_moveSTXO2FTXO(t)
}

func testStore_moveSTXO2FTXO(t *testing.T) {
	store := createStore(t)
	batch := store.NewBatch(true)
	err := store.moveSTXO2FTXO(nil, "moveSTXO2FTXO", batch)
	assert.NotNil(t, err)
}

func testStore_unsetUTXO(t *testing.T) {
	store := createStore(t)
	addr := ""
	txhash := ""
	batch := store.NewBatch(true)
	err := store.unsetUTXO(&addr, &txhash, 0, "", batch)
	assert.NotNil(t, err)

	addr = "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	txhash = "TXHASH"
	err = store.unsetUTXO(&addr, &txhash, 0, "BTY", batch)
	assert.NoError(t, err)
}

func testStore_moveFTXO2UTXO(t *testing.T) {

}

func testStore_moveFTXO2STXO(t *testing.T) {
	store := createStore(t)
	batch := store.NewBatch(true)
	err := store.moveFTXO2STXO(nil, "TXHASH", batch)
	assert.NotNil(t, err)

}

func testStore_getFTXOlist(t *testing.T) {
	store := createStore(t)
	txs, bts := store.getFTXOlist()
	assert.Equal(t, 0, len(bts))
	assert.Equal(t, 0, len(txs))

}

func testStore_getWalletFtxoStxo(t *testing.T) {
	store := createStore(t)
	_, _, err := store.getWalletFtxoStxo("")
	assert.Nil(t, err)
}

func testStore_listSpendUTXOs(t *testing.T) {

}

func testStore_storeScanPrivacyInputUTXO(t *testing.T) {

}

func testStore_setUTXO(t *testing.T) {
	var addr, txhash string
	store := createStore(t)
	dbbatch := store.NewBatch(true)
	err := store.setUTXO(&addr, &txhash, 0, nil, dbbatch)
	assert.NotNil(t, err)

	addr = "setUTXO"
	txhash = "TXHASH"
	err = store.setUTXO(&addr, &txhash, 0, nil, dbbatch)
	assert.NotNil(t, err)
}

func testStore_selectPrivacyTransactionToWallet(t *testing.T) {

}

func testStore_moveUTXO2STXO(t *testing.T) {

}

func testStore_updateScanInputUTXOs(t *testing.T) {

}

func testStore_isUTXOExist(t *testing.T) {
	store := createStore(t)
	pdbs, err := store.isUTXOExist("", 0)
	assert.Nil(t, pdbs)
	assert.NotNil(t, err)

}

func testStore_setScanPrivacyInputUTXO(t *testing.T) {
	store := createStore(t)
	utxogls := store.setScanPrivacyInputUTXO(0)
	assert.Nil(t, utxogls)

}

func testStore_saveREscanUTXOsAddresses(t *testing.T) {

}

func testStore_getRescanUtxosFlag4Addr(t *testing.T) {
	store := createStore(t)
	utxos, err := store.getRescanUtxosFlag4Addr(&pt.ReqRescanUtxos{})
	assert.Nil(t, utxos)
	assert.NotNil(t, err)

}

func testStore_moveUTXO2FTXO(t *testing.T) {

}

func testStore_getPrivacyTokenUTXOs(t *testing.T) {
	store := createStore(t)
	utxos, err := store.getPrivacyTokenUTXOs("", "")
	assert.Nil(t, err)
	assert.NotNil(t, utxos)

	token := "BTY"
	addr := "getPrivacyTokenUTXOs"

	for n := 0; n < 5; n++ {
		data := &pt.PrivacyDBStore{Txindex: int32(n)}
		bt, err := proto.Marshal(data)
		assert.NoError(t, err)
		key := fmt.Sprintf("Key%d", n)
		err = store.Set(calcUTXOKey4TokenAddr(token, addr, "txhash", n), []byte(key))
		assert.NoError(t, err)
		err = store.Set([]byte(key), bt)
		assert.NoError(t, err)
	}
	utxos, err = store.getPrivacyTokenUTXOs(token, addr)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(utxos.utxos))
}

func testStore_getWalletPrivacyTxDetails(t *testing.T) {
	store := createStore(t)
	wtds, err := store.getWalletPrivacyTxDetails(nil)
	assert.Nil(t, wtds)
	assert.NotNil(t, err)

	wtds, err = store.getWalletPrivacyTxDetails(&pt.ReqPrivacyTransactionList{})
	assert.Nil(t, wtds)
	assert.NotNil(t, err)
}

func testStore_listFrozenUTXOs(t *testing.T) {
	store := createStore(t)
	token := "BTY"
	addr := "26htvcBNSEA7fZhAdLJphDwQRQJaHpyHTq"
	txs, err := store.listFrozenUTXOs("", "")
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	txs, err = store.listFrozenUTXOs(token, addr)
	assert.Nil(t, txs)
	assert.Nil(t, err)
	tx := &pt.FTXOsSTXOsInOneTx{Tokenname: "BTY"}
	bt, err := proto.Marshal(tx)
	assert.NoError(t, err)
	err = store.Set(calcKey4FTXOsInTx(token, addr, "TXHASH"), bt)
	assert.NoError(t, err)
	txs, err = store.listFrozenUTXOs(token, addr)
	assert.Nil(t, txs)
	assert.NotNil(t, err)
	err = store.Set(calcKey4FTXOsInTx(token, addr, "TXHASH"), []byte("DataKey"))
	assert.NoError(t, err)
	err = store.Set([]byte("DataKey"), bt)
	assert.NoError(t, err)
	txs, err = store.listFrozenUTXOs(token, addr)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, true, proto.Equal(tx, txs[0]))
}

func testStore_listAvailableUTXOs(t *testing.T) {
	store := createStore(t)
	utxos, err := store.listAvailableUTXOs("", "")
	assert.Nil(t, utxos)
	assert.Equal(t, err, types.ErrInvalidParam)

	addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTq"
	token := "BTY"
	txhash := "123456"
	utxo := &pt.PrivacyDBStore{
		Tokenname: "BTY",
	}
	key := calcUTXOKey4TokenAddr(token, addr, txhash, 0)
	bt, err := proto.Marshal(utxo)
	assert.NoError(t, err)
	err = store.Set(key, []byte("AccKey"))
	assert.NoError(t, err)
	utxos, err = store.listAvailableUTXOs(token, addr)
	assert.Nil(t, utxos)
	assert.NotNil(t, err)
	err = store.Set([]byte("AccKey"), bt)
	assert.NoError(t, err)
	utxos, err = store.listAvailableUTXOs(token, addr)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(utxos))
	assert.Equal(t, true, proto.Equal(utxo, utxos[0]))
}

func testStore_setWalletAccountPrivacy(t *testing.T) {
	store := createStore(t)
	addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	err := store.setWalletAccountPrivacy("", nil)
	assert.Equal(t, err, types.ErrInvalidParam)
	err = store.setWalletAccountPrivacy(addr, nil)
	assert.Equal(t, err, types.ErrInvalidParam)
	err = store.setWalletAccountPrivacy(addr, &pt.WalletAccountPrivacy{})
	assert.NoError(t, err)
}

func testStore_getAccountByAddr(t *testing.T) {
	store := createStore(t)
	addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTq"
	was, err := store.getAccountByAddr("")
	assert.Nil(t, was)
	assert.Equal(t, err, types.ErrInvalidParam)
	was, err = store.getAccountByAddr(addr)
	assert.Nil(t, was)
	assert.Equal(t, err, types.ErrAddrNotExist)

	account := &types.WalletAccountStore{
		Label: "Label1",
	}
	bt, err := proto.Marshal(account)
	assert.NoError(t, err)
	err = store.Set(calcAddrKey(addr), bt)
	assert.NoError(t, err)
	was, err = store.getAccountByAddr(addr)
	assert.Equal(t, true, proto.Equal(was, account))
	assert.NoError(t, err)
}

func testStore_getAccountByPrefix(t *testing.T) {
	store := createStore(t)
	addr := "16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
	was, err := store.getAccountByAddr("")
	assert.Nil(t, was)
	assert.Equal(t, err, types.ErrInvalidParam)

	was, err = store.getAccountByAddr(addr)
	assert.Nil(t, was)
	assert.Equal(t, err, types.ErrAddrNotExist)

	// 这里始终是成功的，所以不能建立测试
	//other := &types.ReqSignRawTx{Expire:"Ex"}
	//bt, err := proto.Marshal(other)
	//assert.NoError(t, err)
	//err = store.Set(calcAddrKey(addr), bt)
	//assert.NoError(t, err)
	//was, err = store.getAccountByAddr(addr)
	//assert.Nil(t, was)
	//assert.Equal(t, err, types.ErrUnmarshal)

	account := &types.WalletAccountStore{
		Label: "Label1",
	}
	bt, err := proto.Marshal(account)
	assert.NoError(t, err)
	err = store.Set(calcAddrKey(addr), bt)
	assert.NoError(t, err)
	was, err = store.getAccountByAddr(addr)
	assert.NoError(t, err)
	assert.Equal(t, true, proto.Equal(was, account))
}

func testStore_setVersion(t *testing.T) {
	store := createStore(t)
	err := store.setVersion()
	assert.NoError(t, err)
}

func testStore_getVersion(t *testing.T) {
	store := createStore(t)
	bt, err := json.Marshal("this is a string")
	assert.NoError(t, err)
	err = store.Set(calcPrivacyDBVersion(), bt)
	assert.NoError(t, err)
	version := store.getVersion()
	assert.Equal(t, int64(0), version)
	bt, err = json.Marshal(PRIVACYDBVERSION)
	assert.NoError(t, err)
	err = store.Set(calcPrivacyDBVersion(), bt)
	assert.NoError(t, err)
	version = store.getVersion()
	assert.Equal(t, PRIVACYDBVERSION, version)
}

package executor

import (
	"bytes"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/stretchr/testify/assert"
	"testing"
)

//func TestZksyncOption(t *testing.T) {
//	dir, statedb, localdb := util.CreateTestDB()
//	defer util.CloseTestDB(dir, statedb)
//	NewAccountTree(localdb)
//	tree, err := getAccountTree(localdb)
//	assert.Equal(t, nil, err)
//	assert.NotEqual(t, nil, tree)
//	action := &Action{localDB: localdb}
//	deposit := &zt.Deposit{
//		ChainType: "ETH",
//		TokenId: 1,
//		Amount: 10000,
//		EthAddress: "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b",
//	}
//	receipt,err := action.Deposit(deposit)
//	t.Log(receipt)
//	for _, log := range receipt.GetLogs() {
//		detail := &zt.ReceiptLeaf{}
//		types.Decode(log.GetLog(), detail)
//		t.Log(detail)
//	}
//
//	assert.Equal(t, nil, err)
//	leaf , err:= GetLeafByEthAddress(localdb, "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b")
//	assert.Equal(t, nil, err)
//	assert.NotEqual(t, nil, leaf)
//	t.Log(leaf)
//	tree, err = getAccountTree(localdb)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int32(1), tree.GetTotalIndex())
//
//	withdraw := &zt.Withdraw{
//		AccountId: 1,
//		ChainType: "ETH",
//		TokenId: 1,
//		Amount: 500,
//	}
//	receipt,err = action.Withdraw(withdraw)
//	assert.Equal(t, nil, err)
//	t.Log(receipt)
//	for _, log := range receipt.GetLogs() {
//		detail := &zt.ReceiptLeaf{}
//		types.Decode(log.GetLog(), detail)
//		t.Log(detail)
//	}
//	leaf , err= GetLeafByAccountId(localdb, 1)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(9500), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	assert.NotEqual(t, nil, leaf)
//	t.Log(leaf)
//
//	transferToNew := &zt.TransferToNew{
//		FromAccountId: 1,
//		ChainType: "ETH",
//		TokenId: 1,
//		Amount: 500,
//		ToEthAddress: "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81v",
//	}
//	receipt,err = action.TransferToNew(transferToNew)
//	assert.Equal(t, nil, err)
//	t.Log(receipt)
//	for _, log := range receipt.GetLogs() {
//		detail := &zt.ReceiptLeaf{}
//		types.Decode(log.GetLog(), detail)
//		t.Log(detail)
//	}
//	leaf , err= GetLeafByAccountId(localdb, 1)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(9000), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	leaf , err= GetLeafByAccountId(localdb, 2)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(500), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	assert.NotEqual(t, nil, leaf)
//	t.Log(leaf)
//
//
//	transfer := &zt.Transfer{
//		FromAccountId: 1,
//		ChainType: "ETH",
//		TokenId: 1,
//		Amount: 500,
//		ToAccountId: 2,
//	}
//	receipt,err = action.Transfer(transfer)
//	assert.Equal(t, nil, err)
//	t.Log(receipt)
//	for _, log := range receipt.GetLogs() {
//		detail := &zt.ReceiptLeaf{}
//		types.Decode(log.GetLog(), detail)
//		t.Log(detail)
//	}
//	leaf , err= GetLeafByAccountId(localdb, 1)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(8500), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	leaf , err= GetLeafByAccountId(localdb, 2)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(1000), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	assert.NotEqual(t, nil, leaf)
//	t.Log(leaf)
//
//	forceQuit := &zt.ForceQuit{
//		ChainType: "ETH",
//		TokenId: 1,
//		EthAddress: "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b",
//	}
//	receipt,err = action.ForceQuit(forceQuit)
//	assert.Equal(t, nil, err)
//	t.Log(receipt)
//	for _, log := range receipt.GetLogs() {
//		detail := &zt.ReceiptLeaf{}
//		types.Decode(log.GetLog(), detail)
//		t.Log(detail)
//	}
//	leaf , err= GetLeafByAccountId(localdb, 1)
//	assert.Equal(t, nil, err)
//	assert.Equal(t, int64(0), leaf.ChainBalances[0].TokenBalances[0].Balance)
//	t.Log(leaf)
//}

func TestEddsa(t *testing.T) {
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)
	ans := privateKey.PublicKey.Bytes()
	t.Log(privateKey.PublicKey.A.X)
	t.Log(privateKey.PublicKey.A.Y)
	t.Log(ans)
	t.Log(len(ans))
}

package executor

import (
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestZksyncOption(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)
	NewAccountTree(localdb)
	tree, err := getAccountTree(localdb)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, tree)
	action := &Action{localDB: localdb}
	deposit := &zt.Deposit{
		ChainType: "ETH",
		TokenId: 1,
		Amount: 1000,
		EthAddress: "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b",
	}
	receipt,err := action.Deposit(deposit)
	t.Log(receipt)
	for _, log := range receipt.GetLogs() {
		detail := &zt.ReceiptLeaf{}
		types.Decode(log.GetLog(), detail)
		t.Log(detail)
	}

	assert.Equal(t, nil, err)
	leaf , err:= GetLeafByEthAddress(localdb, "0x6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b")
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)
	tree, err = getAccountTree(localdb)
	assert.Equal(t, nil, err)
	assert.Equal(t, int32(1), tree.GetTotalIndex())

	withdraw := &zt.Withdraw{
		AccountId: 1,
		ChainType: "ETH",
		TokenId: 1,
		Amount: 500,
	}
	receipt,err = action.Withdraw(withdraw)
	t.Log(receipt)
	for _, log := range receipt.GetLogs() {
		detail := &zt.ReceiptLeaf{}
		types.Decode(log.GetLog(), detail)
		t.Log(detail)
	}
	leaf , err= GetLeafByAccountId(localdb, 1)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)
}

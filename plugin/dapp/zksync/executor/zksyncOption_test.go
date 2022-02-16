package executor

import (
	"bytes"
	"encoding/hex"
	"github.com/33cn/chain33/types"
	"math/big"
	"strings"
	"testing"

	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/stretchr/testify/assert"
)

func TestZksyncOption(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)
	/*************************deposit*************************/
	info, err := generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	action := &Action{localDB: localdb, statedb: statedb, height: 1, index: 0, fromaddr: "operator"}
	deposit := &zt.ZkDeposit{
		TokenId:     1,
		Amount:      "10000",
		EthAddress:  "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57",
		Chain33Addr: getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec"),
	}
	receipt, err := action.Deposit(deposit)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	var zklog zt.ZkReceiptLog
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	assert.Equal(t, nil, err)
	leaf, err := GetLeafByAccountId(statedb, 1, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)

	/*************************setPubKey*************************/
	info, err = generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)
	setPubKey := &zt.ZkSetPubKey{
		AccountId: 1,
		PubKey: &zt.ZkPubKey{
			X: privateKey.PublicKey.A.X.String(),
			Y: privateKey.PublicKey.A.Y.String(),
		},
	}

	receipt, err = action.SetPubKey(setPubKey)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	/*************************withdraw*************************/
	info, err = generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	withdraw := &zt.ZkWithdraw{
		AccountId: 1,
		TokenId:   1,
		Amount:    "5000",
	}
	msg := wallet.GetWithdrawMsg(withdraw)
	privateKey, err = eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)

	signInfo, err := wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)

	withdraw.Signature = signInfo
	receipt, err = action.Withdraw(withdraw)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	tree, err := getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	t.Log(tree)

	leaf, err = GetLeafByAccountId(statedb, 1, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)

	token, err := GetTokenByAccountIdAndTokenId(statedb, 1, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "5000", token.Balance)

	/*************************transferToNew*************************/
	info, err = generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	transferToNew := &zt.ZkTransferToNew{
		FromAccountId:    1,
		TokenId:          1,
		Amount:           "500",
		ToEthAddress:     "abcd68033A72978C1084E2d44D1Fa06DdC4A2d58",
		ToChain33Address: getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aed"),
	}
	t.Log(strings.ToLower(transferToNew.ToEthAddress))
	msg = wallet.GetTransferToNewMsg(transferToNew)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	transferToNew.Signature = signInfo
	receipt, err = action.TransferToNew(transferToNew)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 1, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "4500", token.Balance)
	token, err = GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "500", token.Balance)

	/*************************transfer*************************/
	info, err = generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	transfer := &zt.ZkTransfer{
		FromAccountId: 1,
		TokenId:       1,
		Amount:        "500",
		ToAccountId:   2,
	}
	msg = wallet.GetTransferMsg(transfer)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	transfer.Signature = signInfo

	receipt, err = action.Transfer(transfer)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 1, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "4000", token.Balance)
	token, err = GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "1000", token.Balance)

	/*************************forceQuit*************************/
	info, err = generateTreeUpdateInfo(statedb)
	assert.Equal(t, nil, err)
	forceQuit := &zt.ZkForceExit{
		AccountId: 1,
		TokenId:   1,
	}
	msg = wallet.GetForceExitMsg(forceQuit)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	forceQuit.Signature = signInfo
	receipt, err = action.ForceExit(forceQuit)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 1, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0", token.Balance)

	tree, err = getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	t.Log(tree)
}

func TestEddsa(t *testing.T) {
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)
	ans := privateKey.PublicKey.Bytes()
	t.Log(privateKey.PublicKey.A.X)
	t.Log(privateKey.PublicKey.A.Y)
	t.Log(ans)
	t.Log(len(ans))
}

func TestBigInt(t *testing.T) {
	byteVal :=  big.NewInt(0).Bytes()
	stringVal := hex.EncodeToString(byteVal)
	t.Log("bigInt 0 byteVal", byteVal)
	t.Log("bigInt 0 stringVal", stringVal)
	t.Log("0 stringVal", "0")
	t.Log("0 byteVal", []byte("0"))
	t.Log("is equal", stringVal == "0")
}

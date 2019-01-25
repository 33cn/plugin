// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rpc_test

import (
	"testing"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util/testnode"
	mty "github.com/33cn/plugin/plugin/dapp/multisig/types"
	"github.com/stretchr/testify/assert"
)

var (
	Symbol     = "BTY"
	Asset      = "coins"
	PrivKeyA   = "0x06c0fa653c719275d1baa365c7bc0b9306447287499a715b541b930482eaa504" // 1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK
	PrivKeyB   = "0x4c8663cded61093af20339ae038b3c6bfa58a33e65874a655022f82eaf3f2fa0" // 1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj
	PrivKeyC   = "0x9abcf378b397682109c174b37a45bfc8a459c9514dd2ef719e22a9815373047d" // 1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd
	PrivKeyD   = "0xbf8f865a03fec64f30d2243847807e88d2dbc8104e77925e4fc11c4d4380f3da" // 166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf
	PrivKeyE   = "0x5b8ca316cf073aa94f1056a9e3f6e0b9a9ec11ae45862d58c7a09640b4d55302" // 1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo
	PrivKeyGen = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	AddrA      = "1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK"
	AddrB      = "1LDGrokrZjo1HtSmSnw8ef3oy5Vm1nctbj"
	AddrC      = "1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd"
	AddrD      = "166po3ghRbRu53hu8jBBQzddp7kUJ9Ynyf"
	AddrE      = "1KHwX7ZadNeQDjBGpnweb4k2dqj2CWtAYo"

	GenAddr = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
)

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

func getRPCClient(t *testing.T, mocker *testnode.Chain33Mock) *jsonclient.JSONClient {
	jrpcClient := mocker.GetJSONC()
	assert.NotNil(t, jrpcClient)
	return jrpcClient
}

func getTx(t *testing.T, hex string) *types.Transaction {
	data, err := common.FromHex(hex)
	assert.Nil(t, err)
	var tx types.Transaction
	err = types.Decode(data, &tx)
	assert.Nil(t, err)
	return &tx
}

func TestMultiSigAccount(t *testing.T) {
	mocker := testnode.New("--free--", nil)
	defer mocker.Close()
	mocker.Listen()
	jrpcClient := getRPCClient(t, mocker)
	//创建多重签名账户,owner:AddrA,AddrB,GenAddr,weight:20,10,30;coins:BTY 1000000000 RequestWeight:15
	multiSigAccAddr := testAccCreateTx(t, mocker, jrpcClient)
	//多重签名地址转入操作:4000000000
	testTransferInTx(t, mocker, jrpcClient, multiSigAccAddr)
	//多重签名地址转出操作 AddrB  2000000000
	testTransferOutTx(t, mocker, jrpcClient, multiSigAccAddr)
	//owner add AddrE
	testAddOwner(t, mocker, jrpcClient, multiSigAccAddr)
	//owner del AddrE
	testDelOwner(t, mocker, jrpcClient, multiSigAccAddr)
	//owner AddrA modify weight to 30
	testModifyOwnerWeight(t, mocker, jrpcClient, multiSigAccAddr)
	//owner AddrA replace by  AddrE
	testReplaceOwner(t, mocker, jrpcClient, multiSigAccAddr)
	//modify dailylimit coins:BTY  1200000000
	testModifyDailyLimit(t, mocker, jrpcClient, multiSigAccAddr)
	//add dailylimit token:HYB  1000000000
	testAddDailyLimit(t, mocker, jrpcClient, multiSigAccAddr)
	//Modify RequestWeight 16
	testModifyRequestWeight(t, mocker, jrpcClient, multiSigAccAddr)
	//AddrE ConfirmTx
	testConfirmTx(t, mocker, jrpcClient, multiSigAccAddr)

	//AddrE ConfirmTx
	testAbnormal(t, mocker, jrpcClient)
}

//创建多重签名账户
func testAccCreateTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient) string {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc
	//1. MultiSigAccCreateTx 创建交易
	var owners []*mty.Owner
	owmer1 := &mty.Owner{OwnerAddr: AddrA, Weight: 20}
	owmer2 := &mty.Owner{OwnerAddr: AddrB, Weight: 10}
	owmer3 := &mty.Owner{OwnerAddr: GenAddr, Weight: 30}
	owners = append(owners, owmer1)
	owners = append(owners, owmer2)
	owners = append(owners, owmer3)

	symboldailylimit := &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     Asset,
		DailyLimit: 1000000000,
	}

	req := &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}

	var res string
	err := jrpcClient.Call("multisig.MultiSigAccCreateTx", req, &res)
	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account 计数
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccCount"
	params.Payload = types.MustPBToJSON(&types.ReqNil{})
	rep := &types.Int64{}
	err = jrpcClient.Call("Chain33.Query", &params, rep)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), rep.Data)

	//查询account addr
	//t.Log("MultiSigAccounts ")
	req1 := mty.ReqMultiSigAccs{
		Start: 0,
		End:   0,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccounts"
	params.Payload = types.MustPBToJSON(&req1)
	rep1 := &mty.ReplyMultiSigAccs{}
	err = jrpcClient.Call("Chain33.Query", params, rep1)
	assert.Nil(t, err)
	//t.Log(rep1)

	multiSigAccAddr := rep1.Address[0]
	//查询account info
	//t.Log("MultiSigAccountInfo ")
	req2 := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req2)
	rep2 := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep2)
	assert.Nil(t, err)
	assert.Equal(t, AddrA, rep2.Owners[0].OwnerAddr)
	assert.Equal(t, AddrB, rep2.Owners[1].OwnerAddr)
	assert.Equal(t, uint64(15), rep2.RequiredWeight)
	//t.Log(rep2)

	//查询account addr 通过创建者地址creator
	//t.Log("MultiSigAccounts ")
	req3 := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt",
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAllAddress"
	params.Payload = types.MustPBToJSON(&req3)
	rep3 := &mty.AccAddress{}
	err = jrpcClient.Call("Chain33.Query", params, rep3)
	assert.Nil(t, err)
	assert.Equal(t, rep3.Address[0], multiSigAccAddr)
	//t.Log(rep3)

	//获取owner拥有的多重签名账户地址
	req4 := &types.ReqString{
		Data: GenAddr,
	}
	var res4 mty.OwnerAttrs
	err = jrpcClient.Call("multisig.MultiSigAddresList", req4, &res4)
	assert.Nil(t, err)
	assert.Equal(t, res4.Items[0].OwnerAddr, GenAddr)
	return multiSigAccAddr
}

//多重签名地址转入操作
func testTransferInTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc
	//send to exec
	//t.Log("CreateRawTransaction ")
	req3 := &rpctypes.CreateTx{
		To:          address.ExecAddress(mty.MultiSigX),
		Amount:      5000000000,
		Fee:         1,
		Note:        "12312",
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    mty.MultiSigX,
	}
	var res4 string
	err := jrpcClient.Call("Chain33.CreateRawTransaction", req3, &res4)
	assert.Nil(t, err)
	tx := getTx(t, res4)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	block := mocker.GetLastBlock()
	balance := mocker.GetExecAccount(block.StateHash, mty.MultiSigX, mocker.GetGenesisAddress()).Balance
	assert.Equal(t, int64(5000000000), balance)
	//t.Log(balance)

	//转账到多重签名账户中4000000000
	//t.Log("MultiSigAccTransferInTx")
	params1 := &mty.MultiSigExecTransferTo{
		Symbol:   Symbol,
		Amount:   4000000000,
		Note:     "test ",
		Execname: Asset,
		To:       multiSigAccAddr,
	}
	var res5 string
	err = jrpcClient.Call("multisig.MultiSigAccTransferInTx", params1, &res5)
	assert.Nil(t, err)
	tx = getTx(t, res5)
	tx.Sign(types.SECP256K1, gen)
	reply, err = mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	block = mocker.GetLastBlock()
	balance = mocker.GetExecAccount(block.StateHash, mty.MultiSigX, mocker.GetGenesisAddress()).Balance
	assert.Equal(t, int64(1000000000), balance)
	//t.Log(balance)

	//查询多重签名地址上的余额
	//t.Log("MultiSigAccAssets")
	assets := &mty.Assets{
		Symbol: Symbol,
		Execer: Asset,
	}
	req4 := mty.ReqAccAssets{
		MultiSigAddr: multiSigAccAddr,
		Assets:       assets,
		IsAll:        false,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(&req4)
	rep4 := &mty.ReplyAccAssets{}
	err = jrpcClient.Call("Chain33.Query", params, rep4)
	assert.Nil(t, err)
	assert.Equal(t, int64(4000000000), rep4.AccAssets[0].Account.Frozen)
	assert.Equal(t, int64(4000000000), rep4.AccAssets[0].RecvAmount)
	//t.Log(rep4)
}

//多重签名地址转出操作 AddrB
func testTransferOutTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	//t.Log("MultiSigAccTransferOutTx")
	params3 := &mty.MultiSigExecTransferFrom{
		Symbol:   Symbol,
		Amount:   2000000000,
		Note:     "test ",
		Execname: Asset,
		From:     multiSigAccAddr,
		To:       AddrB,
	}
	var res6 string
	err := jrpcClient.Call("multisig.MultiSigAccTransferOutTx", params3, &res6)
	assert.Nil(t, err)
	tx := getTx(t, res6)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询多重签名地址上的余额
	//t.Log("MultiSigAccAssets")
	assets := &mty.Assets{
		Symbol: Symbol,
		Execer: Asset,
	}
	req7 := mty.ReqAccAssets{
		MultiSigAddr: multiSigAccAddr,
		Assets:       assets,
		IsAll:        false,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(&req7)
	rep7 := &mty.ReplyAccAssets{}
	err = jrpcClient.Call("Chain33.Query", params, rep7)
	assert.Nil(t, err)
	assert.Equal(t, int64(2000000000), rep7.AccAssets[0].Account.Frozen)
	assert.Equal(t, int64(4000000000), rep7.AccAssets[0].RecvAmount)
	//t.Log(rep7)

	//查询AddrB在多重签名合约中的余额
	//t.Log("MultiSigAccAssets")
	assets = &mty.Assets{
		Symbol: Symbol,
		Execer: Asset,
	}
	req8 := mty.ReqAccAssets{
		MultiSigAddr: AddrB,
		Assets:       assets,
		IsAll:        false,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(&req8)
	rep8 := &mty.ReplyAccAssets{}
	err = jrpcClient.Call("Chain33.Query", params, rep8)
	assert.Nil(t, err)
	assert.Equal(t, int64(2000000000), rep8.AccAssets[0].Account.Balance)
	assert.Equal(t, int64(2000000000), rep8.AccAssets[0].RecvAmount)
	//t.Log(rep8)
}

//owner add AddrE
func testAddOwner(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	params9 := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAccAddr,
		NewOwner:        AddrE,
		NewWeight:       8,
		OperateFlag:     mty.OwnerAdd,
	}
	var res9 string
	err := jrpcClient.Call("multisig.MultiSigOwnerOperateTx", params9, &res9)
	assert.Nil(t, err)
	tx := getTx(t, res9)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req10 := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req10)
	rep10 := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep10)
	assert.Nil(t, err)
	find := false
	for _, tempowner := range rep10.Owners {
		if tempowner.OwnerAddr == AddrE && tempowner.Weight == uint64(8) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep10)
}

//owner del AddrE
func testDelOwner(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	param := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAccAddr,
		OldOwner:        AddrE,
		OperateFlag:     mty.OwnerDel,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigOwnerOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	find := false
	for _, tempowner := range rep.Owners {
		if tempowner.OwnerAddr == AddrE {
			find = true
			break
		}
	}
	assert.Equal(t, find, false)
	//t.Log(rep)
}

//ModifyOwnerWeight
func testModifyOwnerWeight(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	param := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAccAddr,
		OldOwner:        AddrA,
		NewWeight:       30,
		OperateFlag:     mty.OwnerModify,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigOwnerOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	find := false
	for _, tempowner := range rep.Owners {
		if tempowner.OwnerAddr == AddrA && tempowner.Weight == uint64(30) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep)
}

//testReplaceOwner owner AddrA replace by  AddrE
func testReplaceOwner(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	param := &mty.MultiSigOwnerOperate{
		MultiSigAccAddr: multiSigAccAddr,
		OldOwner:        AddrA,
		NewOwner:        AddrE,
		OperateFlag:     mty.OwnerReplace,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigOwnerOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	find := false
	for _, tempowner := range rep.Owners {
		if tempowner.OwnerAddr == AddrE && tempowner.Weight == uint64(30) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep)
}

//testModifyDailyLimit modify dailylimit coins:BTY  1200000000
func testModifyDailyLimit(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	assetsDailyLimit := &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     Asset,
		DailyLimit: 1200000000,
	}
	param := &mty.MultiSigAccOperate{
		MultiSigAccAddr: multiSigAccAddr,
		DailyLimit:      assetsDailyLimit,
		OperateFlag:     mty.AccDailyLimitOp,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigAccOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	find := false
	for _, dailyLimit := range rep.DailyLimits {
		if dailyLimit.Symbol == Symbol && dailyLimit.Execer == Asset && dailyLimit.DailyLimit == uint64(1200000000) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep)
}

//testAddDailyLimit add dailylimit token:HYB  1000000000
func testAddDailyLimit(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	assetsDailyLimit := &mty.SymbolDailyLimit{
		Symbol:     "HYB",
		Execer:     "token",
		DailyLimit: 1000000000,
	}
	param := &mty.MultiSigAccOperate{
		MultiSigAccAddr: multiSigAccAddr,
		DailyLimit:      assetsDailyLimit,
		OperateFlag:     mty.AccDailyLimitOp,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigAccOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info owner AddrE add 成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	find := false
	for _, dailyLimit := range rep.DailyLimits {
		if dailyLimit.Symbol == "HYB" && dailyLimit.Execer == "token" && dailyLimit.DailyLimit == uint64(1000000000) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep)
}

//testModifyRequestWeight Modify RequestWeight 16
func testModifyRequestWeight(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	gen := mocker.GetGenesisKey()
	var params rpctypes.Query4Jrpc

	param := &mty.MultiSigAccOperate{
		MultiSigAccAddr:   multiSigAccAddr,
		NewRequiredWeight: 16,
		OperateFlag:       mty.AccWeightOp,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigAccOperateTx", param, &res)

	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询account info RequestWeight 16成功
	//t.Log("MultiSigAccountInfo ")
	req := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	assert.Equal(t, uint64(16), rep.RequiredWeight)
	//t.Log(rep)
}
func testConfirmTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string) {
	var params rpctypes.Query4Jrpc

	//1. 转账到AddrB地址，
	//t.Log("CreateRawTransaction ")
	req := &rpctypes.CreateTx{
		To:          AddrB,
		Amount:      1000000000,
		Fee:         1,
		Note:        "12312",
		IsWithdraw:  false,
		IsToken:     false,
		TokenSymbol: "",
		ExecName:    "",
	}
	var res string
	err := jrpcClient.Call("Chain33.CreateRawTransaction", req, &res)
	assert.Nil(t, err)
	gen := mocker.GetGenesisKey()
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
	block := mocker.GetLastBlock()
	balance := mocker.GetAccount(block.StateHash, AddrB).Balance
	assert.Equal(t, int64(1000000000), balance)
	//t.Log(balance)

	//2.owner AddrB从多重签名地址转账100000000到1C5xK2ytuoFqxmVGMcyz9XFKFWcDA8T3rK地址
	//t.Log("MultiSigAccTransferOutTx")
	params3 := &mty.MultiSigExecTransferFrom{
		Symbol:   Symbol,
		Amount:   100000000,
		Note:     "test ",
		Execname: Asset,
		From:     multiSigAccAddr,
		To:       AddrA,
	}
	var res6 string
	err = jrpcClient.Call("multisig.MultiSigAccTransferOutTx", params3, &res6)
	assert.Nil(t, err)
	tx = getTx(t, res6)
	//tx.Sign(types.SECP256K1, gen)
	tx, _ = signTx(tx, PrivKeyB)

	reply, err = mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询多重签名地址上的余额
	amount := int64(1900000000)
	recvAmount := int64(4000000000)
	checkMultiSigAccAssets(t, jrpcClient, multiSigAccAddr, amount, recvAmount, true)

	//查询AddrA在多重签名合约中的余额
	amount = int64(100000000)
	recvAmount = int64(100000000)
	checkMultiSigAccAssets(t, jrpcClient, AddrA, amount, recvAmount, false)
	//查询account info coins:BTY  SpentToday=100000000
	//t.Log("MultiSigAccountInfo ")
	req9 := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccountInfo"
	params.Payload = types.MustPBToJSON(&req9)
	rep9 := &mty.MultiSig{}
	err = jrpcClient.Call("Chain33.Query", params, rep9)
	assert.Nil(t, err)
	find := false
	for _, dailyLimit := range rep9.DailyLimits {
		if dailyLimit.Symbol == Symbol && dailyLimit.Execer == Asset && dailyLimit.SpentToday == uint64(100000000) {
			find = true
			break
		}
	}
	assert.Equal(t, find, true)
	//t.Log(rep9)

	//2.owner AddrB从多重签名地址转账120000000到AddrD地址
	//t.Log("MultiSigAccTransferOutTx")
	params10 := &mty.MultiSigExecTransferFrom{
		Symbol:   Symbol,
		Amount:   1200000000,
		Note:     "test ",
		Execname: Asset,
		From:     multiSigAccAddr,
		To:       AddrD,
	}
	var res10 string
	err = jrpcClient.Call("multisig.MultiSigAccTransferOutTx", params10, &res10)
	assert.Nil(t, err)
	tx = getTx(t, res10)
	tx, _ = signTx(tx, PrivKeyB)

	reply, err = mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)

	//查询多重签名地址上的余额没有变化
	amount = int64(1900000000)
	recvAmount = int64(4000000000)
	checkMultiSigAccAssets(t, jrpcClient, multiSigAccAddr, amount, recvAmount, true)

	//获取此交易的txid
	req11 := mty.ReqMultiSigAccInfo{
		MultiSigAccAddr: multiSigAccAddr,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccTxCount"
	params.Payload = types.MustPBToJSON(&req11)
	rep11 := &mty.Uint64{}
	err = jrpcClient.Call("Chain33.Query", params, rep11)
	assert.Nil(t, err)
	//t.Log(rep11)
	txid := rep11.Data - 1

	//获取txid对应的tx信息，并且执行状态是false
	checkTxInfo(t, jrpcClient, multiSigAccAddr, txid, false, AddrB)
	//撤销此交易
	confirmTx(t, mocker, jrpcClient, multiSigAccAddr, PrivKeyB, txid, false)
	//查询txid的交易Confirm信息已经被撤销
	checkTxInfo(t, jrpcClient, multiSigAccAddr, txid, false, "")
	//通过高权重的owner确认这笔交易
	confirmTx(t, mocker, jrpcClient, multiSigAccAddr, PrivKeyGen, txid, true)
	//查询交易信息已经执行成功，被owner GenAddr
	checkTxInfo(t, jrpcClient, multiSigAccAddr, txid, true, GenAddr)

	//查询当前账户的余额
	amount = int64(700000000)
	recvAmount = int64(4000000000)
	checkMultiSigAccAssets(t, jrpcClient, multiSigAccAddr, amount, recvAmount, true)

	//AddrD   在多重签名合约的月是1200000000
	amount = int64(1200000000)
	recvAmount = int64(1200000000)
	checkMultiSigAccAssets(t, jrpcClient, AddrD, amount, recvAmount, false)

}

//查询txid的交易Confirm信息
func checkTxInfo(t *testing.T, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string, txid uint64, executed bool, ownerAddr string) {

	var params rpctypes.Query4Jrpc

	req := mty.ReqMultiSigTxInfo{
		MultiSigAddr: multiSigAccAddr,
		TxId:         txid,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigTxInfo"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.MultiSigTx{}
	err := jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	//t.Log(rep)

	assert.Equal(t, executed, rep.Executed)
	if ownerAddr == "" {
		assert.Equal(t, 0, len(rep.ConfirmedOwner))
		return
	}
	find := false
	for _, confirmed := range rep.ConfirmedOwner {
		if confirmed.OwnerAddr == ownerAddr {
			find = true
			break
		}
	}
	assert.Equal(t, true, find)
}
func confirmTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, multiSigAccAddr string, privKey string, txid uint64, confirmOrRevoke bool) {
	//撤销这个交易的确认信息
	req := &mty.MultiSigConfirmTx{
		MultiSigAccAddr: multiSigAccAddr,
		TxId:            txid,
		ConfirmOrRevoke: confirmOrRevoke,
	}
	var res string
	err := jrpcClient.Call("multisig.MultiSigConfirmTx", req, &res)
	assert.Nil(t, err)
	tx := getTx(t, res)
	tx, _ = signTx(tx, privKey)

	reply, err := mocker.GetAPI().SendTx(tx)
	assert.Nil(t, err)
	_, err = mocker.WaitTx(reply.GetMsg())
	assert.Nil(t, err)
}
func checkMultiSigAccAssets(t *testing.T, jrpcClient *jsonclient.JSONClient, addr string, amount, recvAmount int64, isMultiSigAddr bool) {
	//t.Log("MultiSigAccAssets")
	var params rpctypes.Query4Jrpc

	assets := &mty.Assets{
		Symbol: Symbol,
		Execer: Asset,
	}
	req := mty.ReqAccAssets{
		MultiSigAddr: addr,
		Assets:       assets,
		IsAll:        false,
	}
	params.Execer = mty.MultiSigX
	params.FuncName = "MultiSigAccAssets"
	params.Payload = types.MustPBToJSON(&req)
	rep := &mty.ReplyAccAssets{}
	err := jrpcClient.Call("Chain33.Query", params, rep)
	assert.Nil(t, err)
	if isMultiSigAddr {
		assert.Equal(t, amount, rep.AccAssets[0].Account.Frozen)
	} else {
		assert.Equal(t, amount, rep.AccAssets[0].Account.Balance)
	}
	assert.Equal(t, recvAmount, rep.AccAssets[0].RecvAmount)
	//t.Log(rep)
}

//异常测试，主要是参数的合法性校验
func testAbnormal(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient) {
	//1. MultiSigAccCreateTx owner重复
	var owners []*mty.Owner
	owmer1 := &mty.Owner{OwnerAddr: AddrA, Weight: 20}
	owmer2 := &mty.Owner{OwnerAddr: AddrB, Weight: 10}
	owmer3 := &mty.Owner{OwnerAddr: AddrA, Weight: 30}
	owmer4 := &mty.Owner{OwnerAddr: AddrC, Weight: 30}

	owners = append(owners, owmer1)
	owners = append(owners, owmer2)
	owners = append(owners, owmer3)

	symboldailylimit := &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     Asset,
		DailyLimit: 1000000000,
	}

	req := &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, mty.ErrOwnerExist)

	// owner addr 错误
	owmer5 := &mty.Owner{OwnerAddr: "34W6mMVYquzGAwY62TwrqnhnhM82VtbGDJ", Weight: 30}
	owners[2] = owmer5
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, types.ErrInvalidAddress)

	// owner weight==0 错误
	owmer5 = &mty.Owner{OwnerAddr: "1DkrXbz2bK6XMpY4v9z2YUnhwWTXT6V5jd", Weight: 0}
	owners[2] = owmer5
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, mty.ErrInvalidWeight)

	// Symbol 错误
	owners[2] = owmer4
	symboldailylimit = &mty.SymbolDailyLimit{
		Symbol:     "hyb",
		Execer:     Asset,
		DailyLimit: 1000000000,
	}
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, nil)

	// Execer 错误
	symboldailylimit = &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     "hyb",
		DailyLimit: 1000000000,
	}
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 15,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, mty.ErrInvalidExec)

	// RequiredWeight > totalownerweight错误
	symboldailylimit = &mty.SymbolDailyLimit{
		Symbol:     Symbol,
		Execer:     Asset,
		DailyLimit: 1000000000,
	}
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 1000,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, mty.ErrRequiredweight)
	//RequiredWeight ==0
	req = &mty.MultiSigAccCreate{
		Owners:         owners,
		RequiredWeight: 0,
		DailyLimit:     symboldailylimit,
	}
	testAbnormalCreateTx(t, mocker, jrpcClient, req, mty.ErrInvalidWeight)
}
func testAbnormalCreateTx(t *testing.T, mocker *testnode.Chain33Mock, jrpcClient *jsonclient.JSONClient, req *mty.MultiSigAccCreate, expecterr error) {
	gen := mocker.GetGenesisKey()
	var res string
	err := jrpcClient.Call("multisig.MultiSigAccCreateTx", req, &res)
	assert.Nil(t, err)
	tx := getTx(t, res)
	tx.Sign(types.SECP256K1, gen)
	_, err = mocker.GetAPI().SendTx(tx)
	assert.Equal(t, err, expecterr)
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"errors"
	"fmt"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	"github.com/33cn/chain33/wallet"
	wcom "github.com/33cn/chain33/wallet/common"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"

	pwallet "github.com/33cn/plugin/plugin/dapp/privacy/wallet"
)

var (
	initBalance = types.DefaultCoinPrecision * 10000
	initHeight  = int64(100)

	// 测试的私钥
	testPrivateKeys = []string{
		"0x8dea7332c7bb3e3b0ce542db41161fd021e3cfda9d7dabacf24f98f2dfd69558",
		"0x920976ffe83b5a98f603b999681a0bc790d97e22ffc4e578a707c2234d55cc8a",
		"0xb59f2b02781678356c231ad565f73699753a28fd3226f1082b513ebf6756c15c",
	}
	// 测试的地址
	testAddrs = []string{
		"1EDDghAtgBsamrNEtNmYdQzC1QEhLkr87t",
		"13cS5G1BDN2YfGudsxRxr7X25yu6ZdgxMU",
		"1JSRSwp16NvXiTjYBYK9iUQ9wqp3sCxz2p",
	}
	// 测试的隐私公钥对
	testPubkeyPairs = []string{
		"92fe6cfec2e19cd15f203f83b5d440ddb63d0cb71559f96dc81208d819fea85886b08f6e874fca15108d244b40f9086d8c03260d4b954a40dfb3cbe41ebc7389",
		"6326126c968a93a546d8f67d623ad9729da0e3e4b47c328a273dfea6930ffdc87bcc365822b80b90c72d30e955e7870a7a9725e9a946b9e89aec6db9455557eb",
		"44bf54abcbae297baf3dec4dd998b313eafb01166760f0c3a4b36509b33d3b50239de0a5f2f47c2fc98a98a382dcd95a2c5bf1f4910467418a3c2595b853338e",
	}

	// exec privacy addr
	execAddr = "1FeyE6VDZ4FYgpK1n2okWMDAtPkwBuooQd"

	testPolicy     = pwallet.New()
	testPolicyName = pty.PrivacyX + "test"

	testCfg = types.NewChain33Config(types.GetDefaultCfgstring())
)

func init() {
	log.SetLogLevel("error")
	Init(pty.PrivacyX, testCfg, nil)
	wcom.RegisterPolicy(testPolicyName, testPolicy)
}

type testExecMock struct {
	dbDir   string
	localDB dbm.KVDB
	stateDB dbm.DB
	exec    dapp.Driver
	wallet  *walletMock
	policy  wcom.WalletBizPolicy
	cfg     *types.Chain33Config
	q       queue.Queue
	qapi    client.QueueProtocolAPI
}

type testcase struct {
	payload            types.Message
	expectExecErr      error
	expectCheckErr     error
	expectExecLocalErr error
	expectExecDelErr   error
	priv               string
	index              int
	systemCreate       bool
	testState          int
	testSign           []byte
	testFee            int64
}

// InitEnv init env
func (mock *testExecMock) InitEnv() {

	mock.cfg = testCfg
	util.ResetDatadir(mock.cfg.GetModuleConfig(), "$TEMP/")
	mock.q = queue.New("channel")
	mock.q.SetConfig(mock.cfg)
	mock.qapi, _ = client.New(mock.q.Client(), nil)
	mock.initExec()
	mock.initWallet()

}

func (mock *testExecMock) FreeEnv() {
	util.CloseTestDB(mock.dbDir, mock.stateDB)
}

func (mock *testExecMock) initExec() {
	mock.dbDir, mock.stateDB, mock.localDB = util.CreateTestDB()
	exec := newPrivacy()
	exec.SetAPI(mock.qapi)
	exec.SetStateDB(mock.stateDB)
	exec.SetLocalDB(mock.localDB)
	exec.SetEnv(100, 1539918074, 1539918074)
	mock.exec = exec
}

func (mock *testExecMock) initWallet() {
	mock.wallet = &walletMock{}
	mock.wallet.Wallet = wallet.New(mock.cfg)
	mock.policy = testPolicy
	mock.wallet.SetQueueClient(mock.q.Client())
	mock.policy.Init(mock.wallet, nil)
	seed, _ := mock.wallet.GenSeed(1)
	mock.wallet.SaveSeed("abcd1234", seed.Seed)
	mock.wallet.ProcWalletUnLock(&types.WalletUnLock{Passwd: "abcd1234"})

	accCoin := account.NewCoinsAccount(mock.cfg)
	accCoin.SetDB(mock.stateDB)
	for index, addr := range testAddrs {
		account := &types.Account{
			Balance: initBalance,
			Addr:    addr,
		}
		accCoin.SaveAccount(account)
		accCoin.SaveExecAccount(execAddr, account)
		privBytes, _ := common.FromHex(testPrivateKeys[index])
		bpriv := wcom.CBCEncrypterPrivkey([]byte(mock.wallet.Password), privBytes)
		was := &types.WalletAccountStore{
			Privkey:   common.ToHex(bpriv),
			Label:     fmt.Sprintf("label%d", index),
			Addr:      addr,
			TimeStamp: types.Now().String(),
		}

		mock.wallet.SetWalletAccount(false, addr, was)
	}

	mock.wallet.GetAPI().ExecWalletFunc(testPolicyName, "EnablePrivacy", &pty.ReqEnablePrivacy{Addrs: testAddrs})
}

func (mock *testExecMock) addBlockTx(tx *types.Transaction, receipt *types.ReceiptData) {

	block := &types.BlockDetail{
		Block: &types.Block{
			Height: initHeight,
		},
		Receipts: []*types.ReceiptData{receipt},
	}

	batch := mock.wallet.GetDBStore().NewBatch(true)
	defer batch.Write()
	mock.policy.OnAddBlockTx(block, tx, 0, batch)
}

func createTx(mock *testExecMock, payload types.Message, priv string, systemCreate bool) (*types.Transaction, error) {

	c, err := crypto.Load(crypto.GetName(types.SECP256K1), -1)
	if err != nil {
		return nil, err
	}
	bytes, err := common.FromHex(priv[:])
	if err != nil {
		return nil, err
	}
	privKey, err := c.PrivKeyFromBytes(bytes)
	if err != nil {
		return nil, err
	}
	if systemCreate {
		action, _ := buildAction(payload)
		tx, err := types.CreateFormatTx(mock.cfg, mock.cfg.ExecName(pty.PrivacyX), types.Encode(action))
		if err != nil {
			return nil, err
		}
		tx.Sign(int32(types.SECP256K1), privKey)
		return tx, nil
	}
	req := payload.(*pty.ReqCreatePrivacyTx)
	if req.GetAssetExec() == "" {
		req.AssetExec = "coins"
	}
	reply, err := mock.wallet.GetAPI().ExecWalletFunc(testPolicyName, "CreateTransaction", payload)
	if err != nil {
		return nil, errors.New("createTxErr:" + err.Error())
	}
	signTxReq := &types.ReqSignRawTx{
		TxHex: common.ToHex(types.Encode(reply)),
	}
	_, signTx, err := mock.policy.SignTransaction(privKey, signTxReq)
	if err != nil {
		return nil, errors.New("signPrivacyTxErr:" + err.Error())
	}
	signTxBytes, _ := common.FromHex(signTx)
	tx := &types.Transaction{}
	err = types.Decode(signTxBytes, tx)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func buildAction(param types.Message) (types.Message, error) {

	action := &pty.PrivacyAction{
		Value: nil,
		Ty:    0,
	}
	if val, ok := param.(*pty.Public2Privacy); ok {
		action.Value = &pty.PrivacyAction_Public2Privacy{Public2Privacy: val}
		action.Ty = pty.ActionPublic2Privacy
	} else if val, ok := param.(*pty.Privacy2Privacy); ok {
		action.Value = &pty.PrivacyAction_Privacy2Privacy{Privacy2Privacy: val}
		action.Ty = pty.ActionPrivacy2Privacy
	} else if val, ok := param.(*pty.Privacy2Public); ok {
		action.Value = &pty.PrivacyAction_Privacy2Public{Privacy2Public: val}
		action.Ty = pty.ActionPrivacy2Public
	} else {
		return nil, types.ErrActionNotSupport
	}
	return action, nil
}

type walletMock struct {
	*wallet.Wallet
}

func (w *walletMock) GetBlockHeight() int64 {
	return initHeight + pty.UtxoMaturityDegree
}

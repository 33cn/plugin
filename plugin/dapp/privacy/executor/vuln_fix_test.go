// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

// 回归测试：隐私交易金额守恒漏洞修复验证(ForkPrivacyAmountCheck)
// 1. 公转私(ActionPublic2Privacy)校验 output 总额与 payload.Amount 一致，
//    存 1 币伪造任意面额 UTXO 的攻击交易必须被 CheckTx 拒绝(ErrAmount)；
// 2. 所有 KeyOutput/KeyInput 金额必须为正且不超过最大币量，
//    负数输出抵消手续费守恒检查的攻击交易必须被拒绝(ErrAmount)；
// 3. 私转私/私转公对所有资产类型(含 token、平行链)强制金额守恒，
//    输出大于输入的交易必须被拒绝(ErrPrivacyTxFeeNotEnough)；
//    utxo手续费燃烧语义与钱包构造一致：仅主链coins燃烧1 coin，token/平行链不燃烧；
// 4. 合法的公转私/私转私/私转公流程不受影响。

import (
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	privacycrypto "github.com/33cn/plugin/plugin/dapp/privacy/crypto"
	pty "github.com/33cn/plugin/plugin/dapp/privacy/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newPrivacyFixMock 构造轻量执行器测试环境(不调用 util.ResetDatadir)，并直接为测试地址注资
func newPrivacyFixMock(t *testing.T) *testExecMock {
	return newPrivacyFixMockWithCfg(testCfg)
}

// newPrivacyFixMockWithCfg 使用指定配置构造轻量执行器测试环境
func newPrivacyFixMockWithCfg(cfg *types.Chain33Config) *testExecMock {
	mock := &testExecMock{cfg: cfg}
	mock.q = queue.New("channel")
	mock.q.SetConfig(mock.cfg)
	mock.qapi, _ = client.New(mock.q.Client(), nil)
	mock.initExec()

	accCoin := account.NewCoinsAccount(cfg)
	accCoin.SetDB(mock.stateDB)
	for _, addr := range testAddrs {
		acc := &types.Account{Balance: initBalance, Addr: addr}
		accCoin.SaveAccount(acc)
		accCoin.SaveExecAccount(execAddr, acc)
	}
	return mock
}

// genFixOnetimeKeyPair 生成测试用的一次性密钥对
func genFixOnetimeKeyPair(t *testing.T) (privacycrypto.PrivKeyPrivacy, privacycrypto.PubKeyPrivacy) {
	var priv privacycrypto.PrivKeyPrivacy
	var pub privacycrypto.PubKeyPrivacy
	privacycrypto.GenerateKeyPair(&priv, &pub)
	return priv, pub
}

// signFixPrivacyTx 以与钱包 signatureTx 相同的方式为私转私/私转公交易生成环签名
func signFixPrivacyTx(t *testing.T, tx *types.Transaction, privs []privacycrypto.PrivKeyPrivacy, pubs []privacycrypto.PubKeyPrivacy, keyImages []*privacycrypto.KeyImage) {
	tx.Signature = nil
	tx.Fee = pty.PrivacyTxFee * testCfg.GetCoinPrecision()
	h := common.BytesToHash(types.Encode(tx))
	ringSign := &types.RingSignature{Items: make([]*types.RingSignatureItem, len(privs))}
	for i := range privs {
		item, err := privacycrypto.GenerateRingSignature(h.Bytes(),
			[]*pty.UTXOBasic{{OnetimePubkey: pubs[i][:]}}, privs[i].Bytes(), 0, keyImages[i][:])
		require.NoError(t, err)
		ringSign.Items[i] = item
	}
	tx.Signature = &types.Signature{
		Ty:        pty.RingBaseonED25519,
		Signature: types.Encode(ringSign),
		Pubkey:    address.ExecPubKey(testCfg.ExecName(pty.PrivacyX)),
	}
}

// depositFixUtxo 执行一笔诚实的公转私，生成指定金额的 UTXO 并落账
func depositFixUtxo(t *testing.T, mock *testExecMock, amount int64, pub privacycrypto.PubKeyPrivacy) *types.Transaction {
	tx, err := createTx(mock, &pty.Public2Privacy{
		Tokenname: "bty",
		Amount:    amount,
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: amount, Onetimepubkey: pub[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	require.NoError(t, mock.exec.CheckTx(tx, 0))
	receipt, err := mock.exec.Exec(tx, 0)
	require.NoError(t, err)
	util.SaveKVList(mock.stateDB, receipt.KV)
	return tx
}

// TestFixVuln_Public2Privacy_AmountInflation 漏洞一回归：
// 公转私声明的 UTXO 总额(1001 coin)远大于实际扣减的 payload.Amount(1 coin)，
// 修复后 CheckTx 必须拒绝该交易
func TestFixVuln_Public2Privacy_AmountInflation(t *testing.T) {
	mock := newPrivacyFixMock(t)
	require.NotNil(t, mock.exec)
	defer util.CloseTestDB(mock.dbDir, mock.stateDB)
	precision := testCfg.GetCoinPrecision()

	_, pub1 := genFixOnetimeKeyPair(t)
	_, pub2 := genFixOnetimeKeyPair(t)

	tx, err := createTx(mock, &pty.Public2Privacy{
		Tokenname: "bty",
		Amount:    1 * precision,
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: 1000 * precision, Onetimepubkey: pub1[:]},
			{Amount: 1 * precision, Onetimepubkey: pub2[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)

	err = mock.exec.CheckTx(tx, 0)
	assert.Equal(t, types.ErrAmount, err, "public2privacy with outputs(1001) != amount(1) must be rejected")

	// 被拒绝的交易不能产生任何 UTXO
	forgedKey := CalcPrivacyOutputKey("", "bty", 1000*precision, common.ToHex(tx.Hash()), 0)
	v, _ := mock.stateDB.Get(forgedKey)
	assert.Nil(t, v, "forged UTXO must not exist in stateDB")
}

// TestFixVuln_Privacy2Privacy_NegativeOutput 漏洞二回归：
// 私转私使用负数输出抵消手续费守恒检查，修复后 CheckTx 必须拒绝;
// 输出总额大于输入总额(无负数抵消)的私转私同样必须被拒绝
func TestFixVuln_Privacy2Privacy_NegativeOutput(t *testing.T) {
	mock := newPrivacyFixMock(t)
	require.NotNil(t, mock.exec)
	defer util.CloseTestDB(mock.dbDir, mock.stateDB)
	precision := testCfg.GetCoinPrecision()

	priv1, pub1 := genFixOnetimeKeyPair(t)
	_, pub2 := genFixOnetimeKeyPair(t)
	_, pub3 := genFixOnetimeKeyPair(t)

	// 诚实的公转私：10 coin
	tx1 := depositFixUtxo(t, mock, 10*precision, pub1)

	ki1, err := privacycrypto.GenerateKeyImage(priv1, pub1[:])
	require.NoError(t, err)

	// 攻击交易：输入 10 coin，输出 1,000,000 coin 与负数两个 UTXO
	huge := int64(1000000) * precision
	neg := -(huge - 10*precision + pty.PrivacyTxFee*precision)
	tx2, err := createTx(mock, &pty.Privacy2Privacy{
		Tokenname: "bty",
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 10 * precision, KeyImage: ki1[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: tx1.Hash(), Outindex: 0}}},
		}},
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: huge, Onetimepubkey: pub2[:]},
			{Amount: neg, Onetimepubkey: pub3[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, tx2,
		[]privacycrypto.PrivKeyPrivacy{priv1},
		[]privacycrypto.PubKeyPrivacy{pub1},
		[]*privacycrypto.KeyImage{ki1})

	err = mock.exec.CheckTx(tx2, 1)
	assert.Equal(t, types.ErrAmount, err, "privacy2privacy with negative output must be rejected")

	// 无负数抵消但输出远超输入的私转私，违反金额守恒，必须被拒绝
	tx3, err := createTx(mock, &pty.Privacy2Privacy{
		Tokenname: "bty",
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 10 * precision, KeyImage: ki1[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: tx1.Hash(), Outindex: 0}}},
		}},
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: huge, Onetimepubkey: pub2[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, tx3,
		[]privacycrypto.PrivKeyPrivacy{priv1},
		[]privacycrypto.PubKeyPrivacy{pub1},
		[]*privacycrypto.KeyImage{ki1})

	err = mock.exec.CheckTx(tx3, 1)
	assert.Equal(t, pty.ErrPrivacyTxFeeNotEnough, err, "privacy2privacy with output > input must be rejected")
}

// TestFixVuln_TokenPrivacy2Privacy_NoConservation 漏洞三回归：
// token 资产的私转私此前不进入 coins 主链手续费分支、完全无守恒校验，
// 修复后输出大于输入的 token 私转私必须被拒绝
func TestFixVuln_TokenPrivacy2Privacy_NoConservation(t *testing.T) {
	mock := newPrivacyFixMock(t)
	require.NotNil(t, mock.exec)
	defer util.CloseTestDB(mock.dbDir, mock.stateDB)
	precision := testCfg.GetCoinPrecision()

	priv1, pub1 := genFixOnetimeKeyPair(t)
	_, pub2 := genFixOnetimeKeyPair(t)

	// 直接在 stateDB 中放置一个 1 token 的 UTXO(模拟此前正常的公转私结果)
	fakePreTxHash := []byte("fakeTokenUtxoPreTxHash00000000001")
	utxoKey := CalcPrivacyOutputKey("token", "TEST", 1*precision, common.ToHex(fakePreTxHash), 0)
	err := mock.stateDB.Set(utxoKey, types.Encode(&pty.KeyOutput{
		Amount:        1 * precision,
		Onetimepubkey: pub1[:],
	}))
	require.NoError(t, err)

	ki1, err := privacycrypto.GenerateKeyImage(priv1, pub1[:])
	require.NoError(t, err)

	// 输入 1 token，输出 1,000,000 token
	tx, err := createTx(mock, &pty.Privacy2Privacy{
		AssetExec: "token",
		Tokenname: "TEST",
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 1 * precision, KeyImage: ki1[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: fakePreTxHash, Outindex: 0}}},
		}},
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: 1000000 * precision, Onetimepubkey: pub2[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, tx,
		[]privacycrypto.PrivKeyPrivacy{priv1},
		[]privacycrypto.PubKeyPrivacy{pub1},
		[]*privacycrypto.KeyImage{ki1})

	err = mock.exec.CheckTx(tx, 0)
	assert.Equal(t, pty.ErrPrivacyTxFeeNotEnough, err, "token privacy2privacy with output > input must be rejected")

	mintedKey := CalcPrivacyOutputKey("token", "TEST", 1000000*precision, common.ToHex(tx.Hash()), 0)
	v, _ := mock.stateDB.Get(mintedKey)
	assert.Nil(t, v, "minted token UTXO must not exist in stateDB")

	// 合法的 token 私转私：token/平行链场景不燃烧utxo手续费(与钱包构造逻辑一致)，
	// 输入总额等于输出总额即可通过，参照 ci_paracross 的 token(GD) priv2priv 用例
	_, pub3 := genFixOnetimeKeyPair(t)
	legitTx, err := createTx(mock, &pty.Privacy2Privacy{
		AssetExec: "token",
		Tokenname: "TEST",
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 1 * precision, KeyImage: ki1[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: fakePreTxHash, Outindex: 0}}},
		}},
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: 1 * precision, Onetimepubkey: pub3[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, legitTx,
		[]privacycrypto.PrivKeyPrivacy{priv1},
		[]privacycrypto.PubKeyPrivacy{pub1},
		[]*privacycrypto.KeyImage{ki1})
	require.NoError(t, mock.exec.CheckTx(legitTx, 0), "legit token privacy2privacy with zero utxo fee must pass CheckTx")
	receipt, err := mock.exec.Exec(legitTx, 0)
	require.NoError(t, err)
	util.SaveKVList(mock.stateDB, receipt.KV)
}

// TestFix_LegitPrivacyFlows 合法流程回归：
// 金额守恒的公转私/私转私/私转公交易在修复后不受影响
func TestFix_LegitPrivacyFlows(t *testing.T) {
	mock := newPrivacyFixMock(t)
	require.NotNil(t, mock.exec)
	defer util.CloseTestDB(mock.dbDir, mock.stateDB)
	precision := testCfg.GetCoinPrecision()
	execAddr := address.ExecAddress(testCfg.ExecName(pty.PrivacyX))

	accCoin := account.NewCoinsAccount(testCfg)
	accCoin.SetDB(mock.stateDB)
	balB0 := accCoin.LoadExecAccount(testAddrs[1], execAddr).GetBalance()

	priv1, pub1 := genFixOnetimeKeyPair(t)
	priv2, pub2 := genFixOnetimeKeyPair(t)

	// 公转私：存入 10 coin，输出 10 coin UTXO
	tx1 := depositFixUtxo(t, mock, 10*precision, pub1)

	ki1, err := privacycrypto.GenerateKeyImage(priv1, pub1[:])
	require.NoError(t, err)
	ki2, err := privacycrypto.GenerateKeyImage(priv2, pub2[:])
	require.NoError(t, err)

	// 私转私：输入 10 coin，输出 9 coin，差额 1 coin 作为手续费
	tx2, err := createTx(mock, &pty.Privacy2Privacy{
		Tokenname: "bty",
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 10 * precision, KeyImage: ki1[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: tx1.Hash(), Outindex: 0}}},
		}},
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: 9 * precision, Onetimepubkey: pub2[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, tx2,
		[]privacycrypto.PrivKeyPrivacy{priv1},
		[]privacycrypto.PubKeyPrivacy{pub1},
		[]*privacycrypto.KeyImage{ki1})
	require.NoError(t, mock.exec.CheckTx(tx2, 1), "legit privacy2privacy must pass CheckTx")
	receipt2, err := mock.exec.Exec(tx2, 1)
	require.NoError(t, err)
	util.SaveKVList(mock.stateDB, receipt2.KV)

	// 私转公：输入 9 coin，提取 8 coin 到公开地址，差额 1 coin 作为手续费
	tx3, err := createTx(mock, &pty.Privacy2Public{
		Tokenname: "bty",
		Amount:    8 * precision,
		To:        testAddrs[1],
		Output:    &pty.PrivacyOutput{},
		Input: &pty.PrivacyInput{Keyinput: []*pty.KeyInput{
			{Amount: 9 * precision, KeyImage: ki2[:],
				UtxoGlobalIndex: []*pty.UTXOGlobalIndex{{Txhash: tx2.Hash(), Outindex: 0}}},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)
	signFixPrivacyTx(t, tx3,
		[]privacycrypto.PrivKeyPrivacy{priv2},
		[]privacycrypto.PubKeyPrivacy{pub2},
		[]*privacycrypto.KeyImage{ki2})
	require.NoError(t, mock.exec.CheckTx(tx3, 2), "legit privacy2public must pass CheckTx")
	receipt3, err := mock.exec.Exec(tx3, 2)
	require.NoError(t, err)
	util.SaveKVList(mock.stateDB, receipt3.KV)

	balB1 := accCoin.LoadExecAccount(testAddrs[1], execAddr).GetBalance()
	assert.Equal(t, balB0+8*precision, balB1, "legit privacy2public must deposit exactly 8 coins")
}

// TestFix_ForkGate 分叉门控回归：ForkPrivacyAmountCheck 激活前保持旧行为，
// 公转私金额不守恒的交易按旧逻辑直接放行，以兼容在运链的历史共识
func TestFix_ForkGate(t *testing.T) {
	// Title=local 时所有分叉强制为高度0, 替换为非local标题后才能设置自定义分叉高度
	cfgStr := strings.Replace(types.GetDefaultCfgstring(), `Title="local"`, `Title="privacy-fork-test"`, 1)
	cfg := types.NewChain33Config(cfgStr)
	cfg.SetDappFork(pty.PrivacyX, pty.ForkPrivacyAmountCheck, 10000)
	mock := newPrivacyFixMockWithCfg(cfg)
	require.NotNil(t, mock.exec)
	defer util.CloseTestDB(mock.dbDir, mock.stateDB)
	precision := cfg.GetCoinPrecision()

	_, pub1 := genFixOnetimeKeyPair(t)

	// 执行器高度为100(initExec SetEnv), 分叉在10000激活, 当前处于分叉前
	tx, err := createTx(mock, &pty.Public2Privacy{
		Tokenname: "bty",
		Amount:    1 * precision,
		Output: &pty.PrivacyOutput{Keyoutput: []*pty.KeyOutput{
			{Amount: 1000 * precision, Onetimepubkey: pub1[:]},
		}},
	}, testPrivateKeys[0], true)
	require.NoError(t, err)

	err = mock.exec.CheckTx(tx, 0)
	assert.NoError(t, err, "pre-fork behavior must be preserved: no conservation check before fork height")
}

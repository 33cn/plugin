// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"strings"
	"testing"

	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
	rexec "github.com/33cn/plugin/plugin/dapp/rollup/executor"
)

// 复现并回归 paracross 节点组「标题前缀」接管漏洞(参考 docs/security/paracross-nodegroup-title-takeover-analysis.md):
// 攻击者用真实平行链的execer(user.p.superx.paracross)携带伪造的前缀title(user.p.super)自建节点组,
// 再借助跨链交易的前缀匹配(types.IsSpecificParaExecName)接管真实平行链资产。
// 本文件校验 ForkParaTitleExecMatch 分叉后的防护逻辑。
const (
	fakeParaTitle = "user.p.super"            //攻击入口使用的伪造前缀title，缺少结尾'.'
	fakeParaExec  = "user.p.super.paracross"  //伪造title自身对应的execer
	realParaTitle = "user.p.superx."          //真实平行链title
	realParaExec  = "user.p.superx.paracross" //真实平行链execer
	newParaTitle  = "user.p.newchain."        //未注册的新平行链title(自由注册场景)
)

// createNodeGroupConfigTx 构造指定execer与payload title的节点组配置交易
func createNodeGroupConfigTx(execer string, config *pt.ParaNodeGroupConfig) *types.Transaction {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionNodeGroupApply,
		Value: &pt.ParacrossAction_NodeGroupConfig{NodeGroupConfig: config},
	}
	return &types.Transaction{
		Payload: types.Encode(action),
		Execer:  []byte(execer),
	}
}

type TitleTakeoverTestSuite struct {
	suite.Suite
	stateDB dbm.KV
	localDB *dbmock.KVDB
	api     *apimock.QueueProtocolAPI
	exec    *Paracross
}

func (suite *TitleTakeoverTestSuite) SetupSuite() {
	suite.stateDB, _ = dbm.NewGoMemDB("state", "state", 1024)
	suite.localDB = new(dbmock.KVDB)
	suite.api = new(apimock.QueueProtocolAPI)
	//主链场景，主链只校验 len(title) > 0
	suite.api.On("GetConfig", mock.Anything).Return(chain33TestMainCfg, nil)
	suite.api.On("GetBlockHash", mock.Anything).Return(&types.ReplyHash{Hash: []byte("main-block-hash-10")}, nil)

	suite.exec = newParacross().(*Paracross)
	suite.exec.SetAPI(suite.api)
	suite.exec.SetLocalDB(suite.localDB)
	suite.exec.SetStateDB(suite.stateDB)
	suite.exec.SetEnv(0, 0, 0)
	suite.exec.SetBlockInfo([]byte(""), []byte(""), 0)
	enableParacrossTransfer = false
}

// setNodeGroupAddrs 直接写入节点组，模拟分叉前已经创建的(伪造)节点组
func (suite *TitleTakeoverTestSuite) setNodeGroupAddrs(title string, addrs ...string) {
	var item types.ConfigItem
	key := calcParaNodeGroupAddrsKey(title)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, addrs...)
	item.Addr = addrs[0]
	suite.Nil(suite.stateDB.Set(key, types.Encode(&item)))
}

// signTitleTx 与signTx逻辑一致，但接收*testing.T避免按值拷贝suite.Suite(go vet copylocks)
func signTitleTx(t *testing.T, tx *types.Transaction, hexPrivKey string) (*types.Transaction, error) {
	signType := types.SECP256K1
	c, err := crypto.Load(types.GetSignName("", signType), -1)
	if err != nil {
		return tx, err
	}
	bytesData, err := common.FromHex(hexPrivKey)
	if err != nil {
		return tx, err
	}
	privKey, err := c.PrivKeyFromBytes(bytesData)
	if err != nil {
		return tx, err
	}
	tx.Sign(int32(signType), privKey)
	return tx, nil
}

func (suite *TitleTakeoverTestSuite) execNodeGroupTx(privKey string, tx *types.Transaction) (*types.Receipt, error) {
	tx, err := signTitleTx(suite.T(), tx, privKey)
	suite.Nil(err)
	return suite.exec.Exec(tx, 0)
}

// TestApplyWithRealChainExecerAndFakeTitleRejected 攻击入口交易:
// execer=user.p.superx.paracross(真实链) + title=user.p.super(伪造前缀title) 必须被拒绝
func (suite *TitleTakeoverTestSuite) TestApplyWithRealChainExecerAndFakeTitleRejected() {
	config := &pt.ParaNodeGroupConfig{
		Title: fakeParaTitle,
		Addrs: Account14K,
		Op:    pt.ParacrossNodeGroupApply,
	}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeGroupConfigTx(realParaExec, config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
	//确认没有写入节点组
	_, err = suite.stateDB.Get(calcParaNodeGroupAddrsKey(fakeParaTitle))
	suite.Error(err)
}

// TestApplyNonCanonicalTitleRejected title缺少结尾'.'的非规范形式一律拒绝
func (suite *TitleTakeoverTestSuite) TestApplyNonCanonicalTitleRejected() {
	config := &pt.ParaNodeGroupConfig{
		Title: fakeParaTitle,
		Addrs: Account14K,
		Op:    pt.ParacrossNodeGroupApply,
	}
	//即使execer与title同名(前缀相同)，非规范title也不能写入
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeGroupConfigTx(fakeParaExec, config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestApplyEmptyNameTitleRejected user.p. 会匹配所有平行链
func (suite *TitleTakeoverTestSuite) TestApplyEmptyNameTitleRejected() {
	config := &pt.ParaNodeGroupConfig{
		Title: pt.ParaPrefix,
		Addrs: Account14K,
		Op:    pt.ParacrossNodeGroupApply,
	}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeGroupConfigTx("user.p.paracross", config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestApplyLegitTitleAccepted 合法的自由注册(execer与title精确对应)不受影响
func (suite *TitleTakeoverTestSuite) TestApplyLegitTitleAccepted() {
	config := &pt.ParaNodeGroupConfig{
		Title: newParaTitle,
		Addrs: Account14K,
		Op:    pt.ParacrossNodeGroupApply,
	}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeGroupConfigTx(newParaTitle+pt.ParaX, config))
	suite.Nil(err)
	suite.NotNil(receipt)
	suite.Equal(int32(types.ExecOk), receipt.Ty)
}

// TestApproveFakeTitleRejected 自审批同样受title/execer精确匹配约束
func (suite *TitleTakeoverTestSuite) TestApproveFakeTitleRejected() {
	config := &pt.ParaNodeGroupConfig{
		Title: fakeParaTitle,
		Id:    "mavl-paracross-title-nodegroupid-user.p.super-0xdeadbeef",
		Op:    pt.ParacrossNodeGroupApprove,
	}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeGroupConfigTx(realParaExec, config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestCommitFakePrefixTitleRejected 即使伪造前缀title的节点组已经存在(分叉前创建),
// 分叉后使用该title提交commit也必须被拒绝，不能再触发真实平行链的跨链交易
func (suite *TitleTakeoverTestSuite) TestCommitFakePrefixTitleRejected() {
	suite.setNodeGroupAddrs(fakeParaTitle, Account14K)

	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   []byte("main-block-hash-10"),
		MainBlockHeight: 10,
		Title:           fakeParaTitle,
		Height:          10,
		PreBlockHash:    []byte("pre-block-hash"),
		BlockHash:       []byte("fake-block-hash"),
		PreStateHash:    []byte("pre-state-hash"),
		StateHash:       []byte("state-hash"),
		TxCounts:        1,
		TxResult:        []byte("01"),
	}
	act := &pt.ParacrossCommitAction{Status: status}
	tx, err := pt.CreateRawCommitTx4MainChain(chain33TestMainCfg, act, pt.GetExecName(chain33TestMainCfg), 0)
	suite.Nil(err)
	tx, err = signTitleTx(suite.T(), tx, PrivKey14K)
	suite.Nil(err)

	receipt, err := suite.exec.Exec(tx, 0)
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestCheckParaTitleExecMatch title与execer的精确匹配关系
func (suite *TitleTakeoverTestSuite) TestCheckParaTitleExecMatch() {
	cases := []struct {
		title string
		exec  string
		ok    bool
	}{
		{realParaTitle, realParaExec, true},
		{fakeParaTitle, realParaExec, false},              //漏洞攻击入口
		{fakeParaTitle, fakeParaTitle + pt.ParaX, false},  //非规范title
		{pt.ParaPrefix, realParaExec, false},              //空名称title
		{"u", realParaExec, false},                        //匹配所有平行链
		{"USER.P.SUPERX.", realParaExec, false},           //大小写变体
		{realParaTitle, "user.p.SuperX.paracross", false}, //execer大小写变体
		{realParaTitle, "user.p.super.paracross", false},  //另一个平行链
		{realParaTitle, "user.p.superx.coins", false},     //非paracross执行器
	}
	for _, c := range cases {
		tx := &types.Transaction{Execer: []byte(c.exec)}
		a := &action{tx: tx, exec: suite.exec, api: suite.api}
		err := a.checkParaTitleExecMatch(c.title)
		if c.ok {
			suite.NoError(err, "title=%s exec=%s should match", c.title, c.exec)
		} else {
			suite.Equal(pt.ErrInvalidTitle, errors.Cause(err), "title=%s exec=%s should be rejected", c.title, c.exec)
		}
	}
}

// TestCheckCrossTxTitleMatch 跨链交易执行入口(title/execer精确对应)的兜底校验，
// 覆盖rollup等其他触发跨链交易执行的路径
func (suite *TitleTakeoverTestSuite) TestCheckCrossTxTitleMatch() {
	a := &action{exec: suite.exec, api: suite.api}
	//攻击: 伪造前缀title执行真实平行链的跨链交易
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(
		a.checkCrossTxTitleMatch(fakeParaTitle, &types.Transaction{Execer: []byte(realParaExec)})))
	//rollup攻击变体: chainTitle为"真实链title去掉结尾'.'"(user.p.superx)时，
	//虽然与真实链user.p.superx.是不同的节点组key，但归一化后会指向真实链的execer，必须拒绝
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(
		a.checkCrossTxTitleMatch("user.p.superx", &types.Transaction{Execer: []byte(realParaExec)})))
	//合法平行链跨链交易(规范title)
	suite.NoError(a.checkCrossTxTitleMatch(realParaTitle, &types.Transaction{Execer: []byte(realParaExec)}))
	suite.NoError(a.checkCrossTxTitleMatch(fakeParaTitle+".", &types.Transaction{Execer: []byte(fakeParaExec)}))
	//空交易
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(a.checkCrossTxTitleMatch(realParaTitle, nil)))
}

// TestCheckCommitParaTitle commit title必须为规范平行链title
func (suite *TitleTakeoverTestSuite) TestCheckCommitParaTitle() {
	a := &action{tx: &types.Transaction{Execer: []byte(pt.ParaX)}, exec: suite.exec, api: suite.api}
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(a.checkCommitParaTitle(fakeParaTitle)))
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(a.checkCommitParaTitle(pt.ParaPrefix)))
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(a.checkCommitParaTitle("u")))
	suite.NoError(a.checkCommitParaTitle(realParaTitle))
	suite.NoError(a.checkCommitParaTitle(Title))
}

// TestFilterCrossTxsByExactTitle 分叉后跨链交易只按精确title匹配，
// 伪造前缀title不能再匹配到真实平行链的跨链交易
func TestFilterCrossTxsByExactTitle(t *testing.T) {
	txs := []*types.Transaction{
		{Execer: []byte(realParaExec)},
		{Execer: []byte(fakeParaExec)},
		{Execer: []byte("user.p.other.paracross")},
	}

	//历史前缀匹配: user.p.super 会匹配真实链 user.p.superx.paracross
	oldMatched := 0
	for _, tx := range txs {
		if types.IsSpecificParaExecName(fakeParaTitle, string(tx.Execer)) {
			oldMatched++
		}
	}
	assert.Equal(t, 2, oldMatched)

	//分叉后精确匹配: 仅自身(规范title)
	got := filterCrossTxsByExactTitle(chain33TestMainCfg, 1, fakeParaTitle+".", txs)
	assert.Len(t, got, 1)
	assert.Equal(t, fakeParaExec, string(got[0].Execer))

	//非规范title(真实链title去掉结尾'.')不能匹配到真实平行链的跨链交易
	assert.Empty(t, filterCrossTxsByExactTitle(chain33TestMainCfg, 1, "user.p.superx", txs))
	assert.Empty(t, filterCrossTxsByExactTitle(chain33TestMainCfg, 1, fakeParaTitle, txs))
}

// TestExecCrossTxsTitleGuard 跨链交易执行入口的兜底校验(rollup路径同样经过该入口):
// 非规范title即使伪造了节点组/rollup轮次，也不能执行真实平行链的跨链交易
func (suite *TitleTakeoverTestSuite) TestExecCrossTxsTitleGuard() {
	a := &action{exec: suite.exec, api: suite.api}
	realCrossTx := &types.Transaction{Execer: []byte(realParaExec)}
	//前缀title
	receipt, err := a.execCrossTxs(fakeParaTitle, 10, []*types.Transaction{realCrossTx}, []byte{1})
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
	//"真实链title去掉结尾'.'"
	receipt, err = a.execCrossTxs("user.p.superx", 10, []*types.Transaction{realCrossTx}, []byte{1})
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestFilterParaMainCrossTxHashesExact 平行链侧跨链交易hash计算同样按分叉精确匹配
func TestFilterParaMainCrossTxHashesExact(t *testing.T) {
	txs := []*types.Transaction{
		{Execer: []byte(realParaExec)},
		{Execer: []byte(fakeParaExec)},
	}
	//历史前缀匹配会带上真实链交易
	assert.Len(t, FilterParaMainCrossTxHashes(fakeParaTitle, txs), 2)
	//分叉后精确匹配: 只取自己的跨链交易
	assert.Len(t, FilterParaMainCrossTxHashesFork(fakeParaTitle+".", txs), 1)
	//分叉后非规范title(前缀title/去掉结尾'.'的真实链title)什么都匹配不到
	assert.Empty(t, FilterParaMainCrossTxHashesFork(fakeParaTitle, txs))
	assert.Empty(t, FilterParaMainCrossTxHashesFork("user.p.superx", txs))
}

// newTitleForkTestCfg 返回一份独立的主链配置，可单独设置 ForkParaTitleExecMatch 高度，
// 不影响其它用例共享的 chain33TestMainCfg（用于分叉前/分叉后两种行为的对照用例）
func newTitleForkTestCfg(forkHeight int64) *types.Chain33Config {
	cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"test\"", 1))
	cfg.SetDappFork(pt.ParaX, pt.ForkParaTitleExecMatch, forkHeight)
	return cfg
}

func createNodeAddrConfigTx(execer string, config *pt.ParaNodeAddrConfig) *types.Transaction {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionNodeConfig,
		Value: &pt.ParacrossAction_NodeConfig{NodeConfig: config},
	}
	return &types.Transaction{Payload: types.Encode(action), Execer: []byte(execer)}
}

func createSupervisionNodeConfigTx(execer string, config *pt.ParaNodeGroupConfig) *types.Transaction {
	action := &pt.ParacrossAction{
		Ty:    pt.ParacrossActionSupervisionNodeConfig,
		Value: &pt.ParacrossAction_SupervisionNodeConfig{SupervisionNodeConfig: config},
	}
	return &types.Transaction{Payload: types.Encode(action), Execer: []byte(execer)}
}

// newTitleOnlyExec 构造一个只绑定指定配置的执行器(独立stateDB)，用于分叉前后行为对照
func newTitleOnlyExec(cfg *types.Chain33Config) (*Paracross, dbm.KV, error) {
	stateDB, err := dbm.NewGoMemDB("state", "state", 1024)
	if err != nil {
		return nil, nil, err
	}
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	exec := newParacross().(*Paracross)
	exec.SetAPI(api)
	exec.SetLocalDB(new(dbmock.KVDB))
	exec.SetStateDB(stateDB)
	exec.SetEnv(0, 0, 0)
	exec.SetBlockInfo([]byte(""), []byte(""), 0)
	return exec, stateDB, nil
}

// buildCrossTxBlock 构造一个主链区块：一条真实平行链(user.p.superx.)的跨链交易
// 加一条伪造平行链(user.p.super.)的跨链交易
func buildCrossTxBlock() *types.BlockDetail {
	realTx := &types.Transaction{Execer: []byte(realParaExec), Payload: []byte("real-cross-tx")}
	fakeTx := &types.Transaction{Execer: []byte(fakeParaExec), Payload: []byte("fake-cross-tx")}
	return &types.BlockDetail{
		Block: &types.Block{
			Height: 10,
			Txs:    []*types.Transaction{realTx, fakeTx},
		},
		Receipts: []*types.ReceiptData{{Ty: types.ExecOk}, {Ty: types.ExecOk}},
	}
}

func hashsOf(txs []*types.Transaction) [][]byte {
	hashs := make([][]byte, 0, len(txs))
	for _, tx := range txs {
		hashs = append(hashs, tx.Hash())
	}
	return hashs
}

func newCrossTxTestAPI(cfg *types.Chain33Config) *apimock.QueueProtocolAPI {
	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	api.On("GetBlockByHashes", mock.Anything).Return(
		&types.BlockDetails{Items: []*types.BlockDetail{buildCrossTxBlock()}}, nil)
	return api
}

// TestGetCrossTxsByRstExactTitle 覆盖本次漏洞的主路径：主链按title抽取跨链交易时的精确过滤。
// 把 getCrossTxsByRst 中的 filterCrossTxsByExactTitle 退回旧的前缀实现时，该用例必须失败。
func (suite *TitleTakeoverTestSuite) TestGetCrossTxsByRstExactTitle() {
	api := newCrossTxTestAPI(chain33TestMainCfg)
	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   []byte("main-block-hash-10"),
		MainBlockHeight: 10,
		//前缀title：前缀匹配会命中真实链 user.p.superx.paracross
		Title:    fakeParaTitle,
		Height:   10,
		TxResult: []byte("01"),
	}
	//分叉后非规范title不能匹配到任何跨链交易
	crossTxs, _, err := getCrossTxsByRst(api, status)
	suite.NoError(err)
	suite.Empty(crossTxs)

	//规范title只返回自身链的跨链交易
	status.Title = fakeParaTitle + "."
	crossTxs, _, err = getCrossTxsByRst(api, status)
	suite.NoError(err)
	suite.Len(crossTxs, 1)
	suite.Equal(fakeParaExec, string(crossTxs[0].Execer))
}

// TestGetCrossTxsLegacyBranchExactTitle 覆盖 getCrossTxs 的另一条分支(ForkLoopCheckCommitTxDone之前)
func (suite *TitleTakeoverTestSuite) TestGetCrossTxsLegacyBranchExactTitle() {
	cfg := newTitleForkTestCfg(0)
	cfg.SetDappFork(pt.ParaX, pt.ForkLoopCheckCommitTxDone, types.MaxHeight)
	api := newCrossTxTestAPI(cfg)

	//前缀匹配会同时命中真实链与伪造链的交易
	detail := buildCrossTxBlock()
	baseTxs := FilterTxsForPara(cfg, detail.FilterParaTxsByTitle(cfg, fakeParaTitle))
	suite.Len(baseTxs, 2)

	status := &pt.ParacrossNodeStatus{
		MainBlockHash:   []byte("main-block-hash-10"),
		MainBlockHeight: 10,
		Title:           fakeParaTitle,
		Height:          10,
		TxResult:        []byte("01"),
		CrossTxResult:   []byte("01"),
		TxHashs:         [][]byte{CalcTxHashsHash(hashsOf(baseTxs))},
		//精确过滤后跨链交易集合为空，因此校验hash为nil
		CrossTxHashs: [][]byte{nil},
	}
	crossTxs, _, err := getCrossTxs(api, status)
	suite.NoError(err)
	suite.Empty(crossTxs)
}

// TestRollupCrossTxRejectsNonCanonicalChainTitle rollup入口的chainTitle必须是规范平行链title：
// "真实链title去掉结尾'.'"(user.p.superx)与真实链user.p.superx.是两个不同的节点组key，必须拒绝
func (suite *TitleTakeoverTestSuite) TestRollupCrossTxRejectsNonCanonicalChainTitle() {
	a := &action{exec: suite.exec, api: suite.api, txhash: []byte("rollup-cross-tx")}
	_, err := a.rollupCrossTx(&pt.RollupCrossTx{ChainTitle: "user.p.superx", CommitRound: 1})
	suite.Equal(rexec.ErrChainTitle, errors.Cause(err))

	_, err = a.rollupCrossTx(&pt.RollupCrossTx{ChainTitle: fakeParaTitle, CommitRound: 1})
	suite.Equal(rexec.ErrChainTitle, errors.Cause(err))
}

// TestNodeConfigRejectsFakeTitle 节点加入/退出(NodeConfig)同样受title↔execer精确绑定约束
func (suite *TitleTakeoverTestSuite) TestNodeConfigRejectsFakeTitle() {
	config := &pt.ParaNodeAddrConfig{Title: fakeParaTitle, Addr: Account14K, Op: pt.ParaOpNewApply}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createNodeAddrConfigTx(realParaExec, config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestSupervisionNodeConfigRejectsFakeTitle 监督节点配置同样受title↔execer精确绑定约束
func (suite *TitleTakeoverTestSuite) TestSupervisionNodeConfigRejectsFakeTitle() {
	config := &pt.ParaNodeGroupConfig{
		Title: fakeParaTitle,
		Addrs: Account14K,
		Op:    pt.ParacrossSupervisionNodeApply,
	}
	receipt, err := suite.execNodeGroupTx(PrivKey14K, createSupervisionNodeConfigTx(realParaExec, config))
	suite.Nil(receipt)
	suite.Equal(pt.ErrInvalidTitle, errors.Cause(err))
}

// TestPreForkCompatibility 分叉前必须保持历史行为(不绑定title、保持前缀匹配)，用于历史高度重放
func (suite *TitleTakeoverTestSuite) TestPreForkCompatibility() {
	cfg := newTitleForkTestCfg(types.MaxHeight)
	exec, _, err := newTitleOnlyExec(cfg)
	suite.Nil(err)

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)
	a := &action{exec: exec, api: api, tx: &types.Transaction{Execer: []byte(realParaExec)}}
	//分叉前不校验 title 与 execer 的绑定
	suite.NoError(a.checkParaTitleExecMatch(fakeParaTitle))
	suite.NoError(a.checkCommitParaTitle(fakeParaTitle))

	//分叉前保持前缀匹配(旧实现)语义
	txs := []*types.Transaction{{Execer: []byte(realParaExec)}, {Execer: []byte(fakeParaExec)}}
	suite.Len(filterCrossTxsByExactTitle(cfg, 10, fakeParaTitle, txs), 2)

	//分叉前该入口交易(真实链execer + 伪造前缀title)按历史逻辑仍会成功——这正是漏洞当时可利用的原因
	config := &pt.ParaNodeGroupConfig{Title: fakeParaTitle, Addrs: Account14K, Op: pt.ParacrossNodeGroupApply}
	tx, err := signTitleTx(suite.T(), createNodeGroupConfigTx(realParaExec, config), PrivKey14K)
	suite.Nil(err)
	receipt, err := exec.Exec(tx, 0)
	suite.NoError(err)
	suite.NotNil(receipt)
	suite.Equal(int32(types.ExecOk), receipt.Ty)
}

// TestIsHaveCrossTxsEmptyHashs 空的 CrossTxHashs 不能让 isHaveCrossTxs 越界 panic(历史缺陷)
func TestIsHaveCrossTxsEmptyHashs(t *testing.T) {
	cfg := newTitleForkTestCfg(0)
	//关闭 ForkLoopCheckCommitTxDone，走到检查 CrossTxHashs[0] 的分支
	cfg.SetDappFork(pt.ParaX, pt.ForkLoopCheckCommitTxDone, types.MaxHeight)

	//空切片：修复前这里会 index out of range
	assert.False(t, isHaveCrossTxs(cfg, &pt.ParacrossNodeStatus{Height: 10, MainBlockHeight: 10}))

	//[0]为nil：ForkCommitTx后视为没有跨链交易
	status := &pt.ParacrossNodeStatus{Height: 10, MainBlockHeight: 10, CrossTxHashs: [][]byte{nil}}
	assert.False(t, isHaveCrossTxs(cfg, status))

	//有跨链交易
	status.CrossTxHashs = [][]byte{[]byte("cross-tx-hash")}
	assert.True(t, isHaveCrossTxs(cfg, status))
}

func localTitleEntry(title string) []byte {
	return types.Encode(&pt.ReceiptParacrossDone{Title: title, Height: 10})
}

func localNodeGroupEntry(title string) []byte {
	return types.Encode(&pt.ParaNodeGroupStatus{Title: title, Status: pt.ParacrossNodeGroupApprove})
}

// TestListTitlesFiltersInvalidTitle 分叉后 Query_ListTitles 不再返回非规范title(评审建议)
func (suite *TitleTakeoverTestSuite) TestListTitlesFiltersInvalidTitle() {
	suite.localDB.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[][]byte{localTitleEntry(realParaTitle), localTitleEntry(fakeParaTitle), localTitleEntry("u")}, nil).Once()

	resp, err := suite.exec.paracrossListTitles()
	suite.Nil(err)
	titles := resp.(*pt.RespParacrossTitles)
	suite.Len(titles.Titles, 1)
	suite.Equal(realParaTitle, titles.Titles[0].Title)
}

// TestListTitlesPreForkKeepsInvalidTitle 分叉前保持历史行为(不过滤)
func (suite *TitleTakeoverTestSuite) TestListTitlesPreForkKeepsInvalidTitle() {
	cfg := newTitleForkTestCfg(types.MaxHeight)
	exec, _, err := newTitleOnlyExec(cfg)
	suite.Nil(err)
	localDB := new(dbmock.KVDB)
	localDB.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[][]byte{localTitleEntry(realParaTitle), localTitleEntry(fakeParaTitle)}, nil).Once()
	exec.SetLocalDB(localDB)

	resp, err := exec.paracrossListTitles()
	suite.Nil(err)
	suite.Len(resp.(*pt.RespParacrossTitles).Titles, 2)
}

// TestListNodeGroupStatusFiltersInvalidTitle 按状态遍历节点组时同样过滤非规范title
func (suite *TitleTakeoverTestSuite) TestListNodeGroupStatusFiltersInvalidTitle() {
	exec, _, err := newTitleOnlyExec(chain33TestMainCfg)
	suite.Nil(err)
	localDB := new(dbmock.KVDB)
	localDB.On("List", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(
		[][]byte{localNodeGroupEntry(realParaTitle), localNodeGroupEntry(fakeParaTitle)}, nil).Once()
	exec.SetLocalDB(localDB)

	resp, err := exec.Query_ListNodeGroupStatus(&pt.ReqParacrossNodeInfo{Status: 0})
	suite.Nil(err)
	groups := resp.(*pt.RespParacrossNodeGroups)
	suite.Len(groups.Ids, 1)
	suite.Equal(realParaTitle, groups.Ids[0].Title)
}

func TestTitleTakeoverTestSuite(t *testing.T) {
	suite.Run(t, new(TitleTakeoverTestSuite))
}

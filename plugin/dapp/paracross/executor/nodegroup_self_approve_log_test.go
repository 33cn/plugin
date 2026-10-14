// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	dbm "github.com/33cn/chain33/common/db"
	dbmock "github.com/33cn/chain33/common/db/mocks"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

// 覆盖 checkApproveOpNew 里新增的自审批建组告警(logSelfApproveNodeGroup):
//
// 自由注册(ForkParaFreeRegister)之后, 对一个尚无节点组的 title, 任何人都能 Apply 后用同一地址
// Approve, 成为该 title 唯一的超级节点 —— title 维度的资产处置权限就此被接管, 而链上不留下除该
// 告警之外的任何痕迹。本文件断言的是这条日志的"级别选择":
//  1. 托管池有余额(有资产却在裸奔建组, 接管特征) -> crit;
//  2. 托管池为空(正常自助建组)                   -> eror, 不能刷成 crit;
//  3. 非自审批(超级管理员审批别人的申请)          -> 两条都不打。
//
// 同时把托管池的 key 形状钉死: 告警判级别读的就是它, key 拼错就等于哑告警。
const (
	selfApproveTitle = "user.p.selftake."
	selfApproveExec  = "user.p.selftake.paracross"

	selfApproveCritMsg  = "paracross node group created by self approval with balance in escrow pool"
	selfApproveErrorMsg = "paracross node group created by self approval"
)

// logSink 采集 clog(log15)的输出, 用原始 logfmt 行使断言贴近运维实际看到的日志
type logSink struct {
	mu    sync.Mutex
	lines []string
}

func (s *logSink) log(r *log15.Record) error {
	line := strings.TrimRight(string(log15.LogfmtFormat().Format(r)), "\n")
	s.mu.Lock()
	s.lines = append(s.lines, line)
	s.mu.Unlock()
	return nil
}

func (s *logSink) record(msg string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, line := range s.lines {
		if lineMsg(line) == msg {
			out = append(out, line)
		}
	}
	return out
}

// lineMsg 取出 logfmt 行里的 msg 字段(本包告警文案不含引号, 取到第一个收尾引号即可)
func lineMsg(line string) string {
	const key = `msg="`
	i := strings.Index(line, key)
	if i < 0 {
		return ""
	}
	rest := line[i+len(key):]
	j := strings.Index(rest, `"`)
	if j < 0 {
		return ""
	}
	return rest[:j]
}

// captureLogs 把根 logger 的输出接到内存 sink, 测试结束还原
func captureLogs(t *testing.T) *logSink {
	sink := new(logSink)
	old := log15.Root().GetHandler()
	log15.Root().SetHandler(log15.FuncHandler(int(log15.LvlDebug), sink.log))
	t.Cleanup(func() { log15.Root().SetHandler(old) })
	return sink
}

// selfApproveEnv 一条 title 的主链执行环境 + 日志采集
type selfApproveEnv struct {
	t    *testing.T
	cfg  *types.Chain33Config
	db   dbm.KV
	exec *Paracross
	acc  *account.DB
	logs *logSink

	poolAddr string //该 title 的托管池地址
	execAddr string //paracross 执行器地址
}

// newSelfApproveEnv 构造环境(主链 "test" 配置: 不能用 cfg.S 改 config.* 项, 会 panic)。
// 默认配置的 [exec.sub.manage] superManager 里已含 Account12Q, 非自审批用例直接用。
func newSelfApproveEnv(t *testing.T) *selfApproveEnv {
	cfgString := strings.Replace(types.GetDefaultCfgstring(), `Title="local"`, `Title="test"`, 1)
	cfg := types.NewChain33Config(cfgString)
	require.False(t, cfg.IsPara(), "precondition: 主链场景")
	require.True(t, cfg.IsDappFork(0, pt.ParaX, pt.ForkParaFreeRegister),
		"precondition: 自由注册分叉必须在高度 0 生效")

	stateDB, err := dbm.NewGoMemDB("state", "state", 1024)
	require.NoError(t, err)

	env := &selfApproveEnv{
		t:        t,
		cfg:      cfg,
		db:       stateDB,
		logs:     captureLogs(t),
		poolAddr: address.ExecAddress(selfApproveTitle + pt.ParaX),
		execAddr: address.ExecAddress(pt.ParaX),
	}

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	env.exec = newParacross().(*Paracross)
	env.exec.SetAPI(api)
	env.exec.SetLocalDB(new(dbmock.KVDB))
	env.exec.SetStateDB(stateDB)
	env.exec.SetEnv(0, 0, 0)
	env.exec.SetBlockInfo([]byte(""), []byte(""), 0)

	env.acc, err = account.NewAccountDB(cfg, cfg.GetCoinExec(), cfg.GetCoinSymbol(), stateDB)
	require.NoError(t, err)
	return env
}

// escrowKey 该 title 托管池的 state key: coins 账户库里 paracross 执行器地址下,
// 以 <title>paracross 的执行器地址为用户地址的 exec 子账户
func (e *selfApproveEnv) escrowKey() string {
	return fmt.Sprintf("mavl-%s-%s-exec-%s:%s", e.cfg.GetCoinExec(), e.cfg.GetCoinSymbol(), e.execAddr, e.poolAddr)
}

// seedEscrow 给该 title 的托管池预置余额
func (e *selfApproveEnv) seedEscrow(amount int64) {
	_, err := e.acc.ExecDeposit(e.poolAddr, e.execAddr, amount)
	require.NoError(e.t, err)
}

// applyNodeGroup 走一遍节点组申请, 返回已签名的申请交易(approve 需要它的 hash 作为 id)
func (e *selfApproveEnv) applyNodeGroup(privKey string) *types.Transaction {
	tx, err := signTitleTx(e.t, createNodeGroupConfigTx(selfApproveExec, &pt.ParaNodeGroupConfig{
		Title: selfApproveTitle,
		Addrs: Account14K,
		Op:    pt.ParacrossNodeGroupApply,
	}), privKey)
	require.NoError(e.t, err)

	receipt, err := e.exec.Exec(tx, 0)
	require.NoError(e.t, err)
	require.Equal(e.t, int32(types.ExecOk), receipt.Ty)
	e.commit(receipt)
	return tx
}

// approveNodeGroup 审批上面那笔申请。分叉后 getNodeGroupID 只接受去掉 key 前缀的交易id
func (e *selfApproveEnv) approveNodeGroup(privKey string, applyTx *types.Transaction) (*types.Receipt, error) {
	tx, err := signTitleTx(e.t, createNodeGroupConfigTx(selfApproveExec, &pt.ParaNodeGroupConfig{
		Title: selfApproveTitle,
		Id:    common.ToHex(applyTx.Hash()),
		Op:    pt.ParacrossNodeGroupApprove,
	}), privKey)
	require.NoError(e.t, err)

	receipt, err := e.exec.Exec(tx, 0)
	if err == nil && receipt != nil {
		e.commit(receipt)
	}
	return receipt, err
}

// commit 把 receipt 的 KV 落库(框架在区块执行阶段做这一步, 直接调 Exec 不会落)
func (e *selfApproveEnv) commit(receipt *types.Receipt) {
	for _, kv := range receipt.KV {
		require.NoError(e.t, e.db.Set(kv.Key, kv.Value))
	}
}

func (e *selfApproveEnv) groupCreated() bool {
	_, err := e.db.Get(calcParaNodeGroupAddrsKey(selfApproveTitle))
	return err == nil
}

// setGroupAddrs 预置一个已存在的节点组(模拟该 title 已经建过组)
func (e *selfApproveEnv) setGroupAddrs(addrs ...string) {
	var item types.ConfigItem
	key := calcParaNodeGroupAddrsKey(selfApproveTitle)
	item.Key = string(key)
	emptyValue := &types.ArrayConfig{Value: make([]string, 0)}
	arr := types.ConfigItem_Arr{Arr: emptyValue}
	item.Value = &arr
	item.GetArr().Value = append(item.GetArr().Value, addrs...)
	item.Addr = addrs[0]
	require.NoError(e.t, e.db.Set(key, types.Encode(&item)))
}

// TestSelfApproveNodeGroupWithAssetsLogsCrit 场景1: 托管池有余额 -> 必须打出 crit
func TestSelfApproveNodeGroupWithAssetsLogsCrit(t *testing.T) {
	env := newSelfApproveEnv(t)
	const initial = int64(10)
	env.seedEscrow(initial)

	//托管池的 key 形状必须与告警读的是同一个: 断言这条原始 key 真的落在状态库里
	raw, err := env.db.Get([]byte(env.escrowKey()))
	require.NoError(t, err, "托管池 key 形状不符: %s", env.escrowKey())
	require.NotNil(t, raw)

	applyTx := env.applyNodeGroup(PrivKey14K)
	receipt, err := env.approveNodeGroup(PrivKey14K, applyTx)
	require.NoError(t, err)
	require.Equal(t, int32(types.ExecOk), receipt.Ty)
	require.True(t, env.groupCreated(), "自审批应真的建出节点组")

	crit := env.logs.record(selfApproveCritMsg)
	require.Len(t, crit, 1, "托管池有资产时必须打出 crit")
	require.Empty(t, env.logs.record(selfApproveErrorMsg), "有资产时不应降级为 eror")

	t.Logf("托管池余额=%d, key=%s", initial, env.escrowKey())
	t.Logf("crit 原始日志行: %s", crit[0])
}

// TestSelfApproveNodeGroupEmptyPoolLogsError 场景2: 托管池为空 -> 只出 eror, 不能刷成 crit
func TestSelfApproveNodeGroupEmptyPoolLogsError(t *testing.T) {
	env := newSelfApproveEnv(t)
	_, err := env.db.Get([]byte(env.escrowKey()))
	require.Error(t, err, "前置条件: 托管池为空")

	applyTx := env.applyNodeGroup(PrivKey14K)
	receipt, err := env.approveNodeGroup(PrivKey14K, applyTx)
	require.NoError(t, err)
	require.Equal(t, int32(types.ExecOk), receipt.Ty)
	require.True(t, env.groupCreated())

	eror := env.logs.record(selfApproveErrorMsg)
	require.Len(t, eror, 1, "池子为空的自助建组只记 eror")
	require.Empty(t, env.logs.record(selfApproveCritMsg), "池子为空不能刷成 crit")

	t.Logf("eror 原始日志行: %s", eror[0])
}

// TestNonSelfApproveNodeGroupNoLog 场景3: 超级管理员审批别人的申请 -> 两条都不打
func TestNonSelfApproveNodeGroupNoLog(t *testing.T) {
	env := newSelfApproveEnv(t)
	require.True(t, isSuperManager(env.cfg, Account12Q), "前置条件: 12Q 是超级管理员")

	//池子给上余额: 若告警漏判"非自审批", 这里就会打出一条 crit, 用例必须失败
	env.seedEscrow(10)

	applyTx := env.applyNodeGroup(PrivKey14K) //14K 申请
	receipt, err := env.approveNodeGroup(PrivKey12Q, applyTx)
	require.NoError(t, err)
	require.Equal(t, int32(types.ExecOk), receipt.Ty)
	require.True(t, env.groupCreated())

	require.Empty(t, env.logs.record(selfApproveCritMsg), "非自审批不应打 crit")
	require.Empty(t, env.logs.record(selfApproveErrorMsg), "非自审批不应打 eror")
}

// TestSelfApproveNodeGroupExistingGroupNoLog 场景4: 该 title 已经建过组, 本次审批不是"建组"
// (例如申请后组被别的申请抢先建出来), 自审批也不再打这两条日志 —— 否则告警文案("created by
// self approval")与实际发生的事情不符。
func TestSelfApproveNodeGroupExistingGroupNoLog(t *testing.T) {
	env := newSelfApproveEnv(t)
	env.seedEscrow(10)
	env.setGroupAddrs(Account14K)

	a := &action{exec: env.exec, api: env.exec.GetAPI(), db: env.db, fromaddr: Account14K}
	a.logSelfApproveNodeGroup(
		&pt.ParaNodeGroupConfig{Title: selfApproveTitle, Id: "some-proposal-id"},
		&pt.ParaNodeGroupStatus{Title: selfApproveTitle, FromAddr: Account14K})

	require.Empty(t, env.logs.record(selfApproveCritMsg), "已有节点组时不应打 crit")
	require.Empty(t, env.logs.record(selfApproveErrorMsg), "已有节点组时不应打 eror")
}

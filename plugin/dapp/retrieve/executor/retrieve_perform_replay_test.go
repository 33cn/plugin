// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package executor

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/33cn/chain33/account"
	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/dapp"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	rt "github.com/33cn/plugin/plugin/dapp/retrieve/types"
	"github.com/stretchr/testify/assert"
)

// newPerformRelayRetrieve 构造独立 retrieve 执行器实例, 各分叉高度可控
func newPerformRelayRetrieve(forkRetriveH, forkAssetH, forkPerformRelayH int64) (drivers.Driver, *types.Chain33Config, dbm.KVDB) {
	cfgstring := strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"chain33\"", 1)
	cfg := types.NewChain33Config(cfgstring)
	cfg.SetDappFork(rt.RetrieveX, rt.ForkRetriveX, forkRetriveH)
	cfg.SetDappFork(rt.RetrieveX, rt.ForkRetriveAssetX, forkAssetH)
	cfg.SetDappFork(rt.RetrieveX, rt.ForkRetrivePerformRelayX, forkPerformRelayH)
	q := queue.New("channel")
	q.SetConfig(cfg)
	api, _ := client.New(q.Client(), nil)
	r := newRetrieve()
	_, _, kvdb := util.CreateTestDB()
	r.SetAPI(api)
	r.SetStateDB(kvdb)
	r.SetLocalDB(kvdb)
	return r, cfg, kvdb
}

func signPerformRelayTx(priv crypto.PrivKey, action *rt.RetrieveAction) *types.Transaction {
	tx := &types.Transaction{Execer: []byte("retrieve"), Payload: types.Encode(action), Fee: 1e6, To: address.ExecAddress("retrieve")}
	tx.Nonce = rand.Int63()
	tx.Sign(types.SECP256K1, priv)
	return tx
}

func performRelayBackupAction(backupAddr, defaultAddr string, delayPeriod int64) *rt.RetrieveAction {
	return &rt.RetrieveAction{Value: &rt.RetrieveAction_Backup{Backup: &rt.BackupRetrieve{
		BackupAddress: backupAddr, DefaultAddress: defaultAddr, DelayPeriod: delayPeriod}}, Ty: rt.RetrieveActionBackup}
}

func performRelayPrepareAction(backupAddr, defaultAddr string) *rt.RetrieveAction {
	return &rt.RetrieveAction{Value: &rt.RetrieveAction_Prepare{Prepare: &rt.PrepareRetrieve{
		BackupAddress: backupAddr, DefaultAddress: defaultAddr}}, Ty: rt.RetrieveActionPrepare}
}

func performRelayPerformAction(backupAddr, defaultAddr string) *rt.RetrieveAction {
	return &rt.RetrieveAction{Value: &rt.RetrieveAction_Perform{Perform: &rt.PerformRetrieve{
		BackupAddress: backupAddr, DefaultAddress: defaultAddr}}, Ty: rt.RetrieveActionPerform}
}

// TestRetrievePerformClearRelationPostFork 验证:
// ForkRetrivePerformRelay 之后, ForkRetriveAsset 分支下 perform 成功后清除 backup<->default 找回关系,
// backup 无法在不重新 prepare 的情况下再次 perform 转走 default 地址后续存入的资产。
func TestRetrievePerformClearRelationPostFork(t *testing.T) {
	r, cfg, kvdb := newPerformRelayRetrieve(0, 0, 0)
	execAddr := address.ExecAddress("retrieve")

	victimAddr, victimPriv := genaddress()
	backupAddr2, backupPriv2 := genaddress()

	accdb, err := account.NewAccountDB(cfg, cfg.GetCoinExec(), cfg.GetCoinSymbol(), kvdb)
	assert.Nil(t, err)
	accdb.SaveExecAccount(execAddr, &types.Account{Balance: 1000, Addr: victimAddr})

	// backup + prepare + perform 正常流程
	r.SetEnv(100, 1000, 0)
	receipt, err := r.Exec(signPerformRelayTx(victimPriv, performRelayBackupAction(backupAddr2, victimAddr, 70)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	receipt, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPrepareAction(backupAddr2, victimAddr)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	r.SetEnv(100, 1000+70, 0)
	receipt, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPerformAction(backupAddr2, victimAddr)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, int64(1000), accdb.LoadExecAccount(backupAddr2, execAddr).Balance)
	assert.Equal(t, int64(0), accdb.LoadExecAccount(victimAddr, execAddr).Balance)

	// perform 成功后找回关系必须被清除
	rel, err := readRetrieve(kvdb, backupAddr2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rel.RetPara), "perform 成功后找回关系应被清除")

	// 受害者地址后续又存入 500, backup 再次 perform 必须被拒绝
	accdb.SaveExecAccount(execAddr, &types.Account{Balance: 500, Addr: victimAddr})
	receipt, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPerformAction(backupAddr2, victimAddr)), 0)
	assert.Equal(t, rt.ErrRetrieveRelation, err, "关系已清除, 二次 perform 应被拒绝")
	assert.Nil(t, receipt)
	assert.Equal(t, int64(1000), accdb.LoadExecAccount(backupAddr2, execAddr).Balance)
	assert.Equal(t, int64(500), accdb.LoadExecAccount(victimAddr, execAddr).Balance)
}

// TestRetrievePerformReplayPreFork 验证:
// ForkRetrivePerformRelay 之前保持旧行为(ForkRetriveAsset 分支 perform 成功后不清除关系), 不破坏在运链共识。
func TestRetrievePerformReplayPreFork(t *testing.T) {
	// ForkRetrivePerformRelay 高度设为极大值, 测试高度远低于它
	r, cfg, kvdb := newPerformRelayRetrieve(0, 0, 1e15)
	execAddr := address.ExecAddress("retrieve")

	victimAddr, victimPriv := genaddress()
	backupAddr2, backupPriv2 := genaddress()

	accdb, err := account.NewAccountDB(cfg, cfg.GetCoinExec(), cfg.GetCoinSymbol(), kvdb)
	assert.Nil(t, err)
	accdb.SaveExecAccount(execAddr, &types.Account{Balance: 1000, Addr: victimAddr})

	r.SetEnv(100, 1000, 0)
	_, err = r.Exec(signPerformRelayTx(victimPriv, performRelayBackupAction(backupAddr2, victimAddr, 70)), 0)
	assert.Nil(t, err)
	_, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPrepareAction(backupAddr2, victimAddr)), 0)
	assert.Nil(t, err)

	r.SetEnv(100, 1000+70, 0)
	receipt, err := r.Exec(signPerformRelayTx(backupPriv2, performRelayPerformAction(backupAddr2, victimAddr)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, int64(1000), accdb.LoadExecAccount(backupAddr2, execAddr).Balance)

	// 分叉前旧行为: perform 成功后关系残留, 状态停留在 retrievePrepare
	rel, err := readRetrieve(kvdb, backupAddr2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rel.RetPara), "分叉前 perform 成功后关系保持不变")
	assert.Equal(t, int32(retrievePrepare), rel.RetPara[0].Status)
}

// TestRetrieveNormalFlowPostFork 验证:
// ForkRetrivePerformRelay 之后, 正常 backup/prepare/perform/cancel 流程不受影响。
func TestRetrieveNormalFlowPostFork(t *testing.T) {
	r, cfg, kvdb := newPerformRelayRetrieve(0, 0, 0)
	execAddr := address.ExecAddress("retrieve")

	defaultAddr2, defaultPriv2 := genaddress()
	backupAddr2, backupPriv2 := genaddress()

	accdb, err := account.NewAccountDB(cfg, cfg.GetCoinExec(), cfg.GetCoinSymbol(), kvdb)
	assert.Nil(t, err)
	accdb.SaveExecAccount(execAddr, &types.Account{Balance: 1000, Addr: defaultAddr2})

	// backup
	r.SetEnv(100, 1000, 0)
	receipt, err := r.Exec(signPerformRelayTx(defaultPriv2, performRelayBackupAction(backupAddr2, defaultAddr2, 70)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	// delay 期未到, perform 被拒绝
	_, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPrepareAction(backupAddr2, defaultAddr2)), 0)
	assert.Nil(t, err)
	receipt, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPerformAction(backupAddr2, defaultAddr2)), 0)
	assert.Equal(t, rt.ErrRetrievePeriodLimit, err)
	assert.Nil(t, receipt)

	// cancel: default 地址持有人取消找回, 关系被清除
	cancelAction := &rt.RetrieveAction{Value: &rt.RetrieveAction_Cancel{Cancel: &rt.CancelRetrieve{
		BackupAddress: backupAddr2, DefaultAddress: defaultAddr2}}, Ty: rt.RetrieveActionCancel}
	receipt, err = r.Exec(signPerformRelayTx(defaultPriv2, cancelAction), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)

	rel, err := readRetrieve(kvdb, backupAddr2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rel.RetPara), "cancel 后找回关系应被清除")

	// 重新走一遍完整 backup/prepare/perform 流程, perform 成功且关系清除
	_, err = r.Exec(signPerformRelayTx(defaultPriv2, performRelayBackupAction(backupAddr2, defaultAddr2, 70)), 0)
	assert.Nil(t, err)
	_, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPrepareAction(backupAddr2, defaultAddr2)), 0)
	assert.Nil(t, err)

	r.SetEnv(100, 1000+70, 0)
	receipt, err = r.Exec(signPerformRelayTx(backupPriv2, performRelayPerformAction(backupAddr2, defaultAddr2)), 0)
	assert.Nil(t, err)
	assert.NotNil(t, receipt)
	assert.Equal(t, int64(1000), accdb.LoadExecAccount(backupAddr2, execAddr).Balance)
	assert.Equal(t, int64(0), accdb.LoadExecAccount(defaultAddr2, execAddr).Balance)

	rel, err = readRetrieve(kvdb, backupAddr2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(rel.RetPara), "perform 成功后找回关系应被清除")
}

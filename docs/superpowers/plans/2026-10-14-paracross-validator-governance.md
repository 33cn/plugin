# paracross 验证人治理化 + 跨链急停（Phase 1）实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development (recommended) or executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 把平行链验证人集合的更换从"内部 2/3 投票即生效"改为"平行链提案 + 主链分档治理/时间锁 + 显式激活"，并新增治理可控的双向跨链急停。

**Architecture:** 复用现有 `ParaNodeGroupConfig` action（不新增 protobuf）新增 op；新增一个 title+组别 维度的状态机（Propose/Approve/Cancel/Activate），延迟按主链区块高度计算，`Activate` 到点后由任意人发送并改写现有节点组 KV；commit 校验读路径保持不变，主链与平行链对同一交易写入相同状态。急停是独立的 KV 开关，在跨链提交与执行两个入口生效。

**Tech Stack:** Go 1.22、chain33 1.72.x、gogo/protobuf（仅读取，不生成）、testify、go test。

**Spec:** `docs/superpowers/specs/2026-10-14-paracross-validator-governance-design.md`

## Global Constraints

- 所有新逻辑由 `pt.ForkParaValidatorSet` 开启；未达分叉高度时行为与现状逐字节一致。
- 分叉判定用 `cfg.IsDappFork(mainHeight, pt.ParaX, pt.ForkParaValidatorSet)`，**禁用** `pt.IsParaForkHeight`（未映射 key 在平行链上会 panic）。平行链在 `[fork.sub.paracross]` 配置与主链相同的主链高度。
- title 必须为规范形式 `user.p.<name>.` 且与交易 execer 精确对应：复用 PR #1326 的 `pt.IsValidParaTitle` 与 `a.checkParaTitleExecMatch`（该 PR 可能尚未合入 master；若不在基线中，先 cherry-pick 或复制这两个函数）。
- **不新增 protobuf 生成代码**：本环境无 `protoc`；状态用确定性自定义编码（大端、定长、版本字节）。
- 延迟一律按**主链区块高度**，禁止墙钟时间；参数从链配置读取（`types.ConfSub(cfg, pt.ParaX).GInt(...)`），默认 `validatorSetHighValueThreshold = 100000 * cfg.GetCoinPrecision()`、`validatorSetDelayBlocksHigh = 8640`、`validatorSetDelayBlocksLow = 60480`。
- "锁定资产"口径：`a.coinsAccount.LoadExecAccount(address.ExecAddress(title+pt.ParaX), address.ExecAddress(pt.ParaX)).Balance`（主链视角，与 `executor/asset.go` 的入金账户一致）。
- 主链判定治理、平行链跟随（`if !cfg.IsPara()`），与现有 `nodeGroupApprove` 模式一致。
- 所有含 `[fork.sub.paracross]` 的配置文件都必须补 `ForkParaValidatorSet`，否则 chain33 `initForkConfig` 直接 panic。
- 每个任务结束提交一次，提交信息用 `feat(paracross): ...` / `test(paracross): ...` / `docs(paracross): ...`。

## Review Focus

1. **高价值档未获治理批准就 Activate** —— 必须失败；放行等于绕过治理闸门。
2. **`EffectiveMainHeight` 边界（−1 / 恰好到点）** —— 差一个区块就切换集合会让主链与平行链对签名人判断不一致。
3. **同一 title 同时存在两个待生效提案** —— 必须拒绝第二个，否则 Activate 写入的集合与提案内容不符。
4. **急停期间的挂起提现（ExecPack）** —— 必须不执行也不改状态，且**不能**让 commit 报错停链。
5. **未升级节点 / 配置文件缺 fork 键** —— 节点启动 panic 或分叉后静默分叉。
6. **重复 Activate** —— 第二次必须失败，不能重复写集合或回退版本号。

## File Structure

| 文件 | 职责 |
|---|---|
| `plugin/dapp/paracross/types/paracross.go`（改） | 新 op 常量、提案状态常量、分档常量 |
| `plugin/dapp/paracross/types/type.go`（改） | `ForkParaValidatorSet` 常量与注册 |
| `plugin/dapp/paracross/executor/kv.go`（改） | 新增 KV key 构造函数 |
| `plugin/dapp/paracross/executor/valsetparam.go`（新） | 阈值/延迟参数读取 |
| `plugin/dapp/paracross/executor/valsetstore.go`（新） | 提案状态编解码、读写、版本号、历史 |
| `plugin/dapp/paracross/executor/valset.go`（新） | 状态机：Propose/Approve/Cancel/Activate、提案产出与集合写入 |
| `plugin/dapp/paracross/executor/circuitbreaker.go`（新） | 急停开关读写与检查 |
| `plugin/dapp/paracross/executor/superaccount.go`（改） | `NodeGroupConfig`/`NodeConfig` 接线 |
| `plugin/dapp/paracross/executor/supervisionaccount.go`（改） | 监督节点组接线 |
| `plugin/dapp/paracross/executor/asset.go`、`action.go`（改） | 急停在提交与挂起执行两个入口 |
| `plugin/dapp/paracross/executor/query.go`（改） | 查询接口 |
| `plugin/dapp/paracross/commands/paracross.go`（改） | CLI |
| `plugin/dapp/paracross/executor/{valsetstore,valset,circuitbreaker,valsetparam}_test.go`（新） | 单元测试 |
| `chain33.fork.toml` 等 11 个含 `[fork.sub.paracross]` 的配置（改） | 分叉高度 |

---

### Task 1: 常量、参数与分叉注册

**Files:**
- Modify: `plugin/dapp/paracross/types/paracross.go`（`ParacrossNodeGroup*` 常量块之后）
- Modify: `plugin/dapp/paracross/types/type.go`
- Create: `plugin/dapp/paracross/executor/valsetparam.go`
- Test: `plugin/dapp/paracross/executor/valsetparam_test.go`

**Interfaces:**
- Consumes: 无
- Produces:
  - `pt.ForkParaValidatorSet = "ForkParaValidatorSet"`（`InitFork` 中 `RegisterDappFork(ParaX, ForkParaValidatorSet, 0)`）
  - `pt.ParacrossValidatorSetPropose=5`、`ParacrossValidatorSetApprove=6`、`ParacrossValidatorSetCancel=7`、`ParacrossValidatorSetActivate=8`
  - `pt.ParacrossCrossChainDisable=9`、`ParacrossCrossChainEnable=10`
  - `pt.ParacrossSupValidatorSetPropose=11`、`...Approve=12`、`...Cancel=13`、`...Activate=14`
  - `pt.ParaValidatorSetPending=0`、`ParaValidatorSetApproved=1`、`ParaValidatorSetCancelled=2`、`ParaValidatorSetActivated=3`
  - `pt.ParaValidatorSetTierLow=0`、`ParaValidatorSetTierHigh=1`
  - `pt.TyLogParaValidatorSetConfig = 684`（新日志类型，并加入 `ParacrossType.GetLogMap`）
  - 错误常量：`pt.ErrValidatorSetNotEffective`、`pt.ErrValidatorSetPendingExists`、`pt.ErrValidatorSetNoActive`、`pt.ErrCrossChainDisabled`（加入 `types/errors.go`）
  - `executor.validatorSetParams(cfg *types.Chain33Config) (threshold, delayHigh, delayLow int64)`

- [ ] **Step 1: 写失败测试** `plugin/dapp/paracross/executor/valsetparam_test.go`

```go
func TestValidatorSetParamsDefault(t *testing.T) {
    threshold, delayHigh, delayLow := validatorSetParams(chain33TestMainCfg)
    assert.Equal(t, int64(100000)*chain33TestMainCfg.GetCoinPrecision(), threshold)
    assert.Equal(t, int64(8640), delayHigh)
    assert.Equal(t, int64(60480), delayLow)
}

func TestValidatorSetParamsFromConf(t *testing.T) {
    cfg := types.NewChain33Config(strings.Replace(types.GetDefaultCfgstring(), "Title=\"local\"", "Title=\"test\"", 1))
    cfg.S("config.exec.sub.paracross.validatorSetHighValueThreshold", int64(500))
    cfg.S("config.exec.sub.paracross.validatorSetDelayBlocksHigh", int64(10))
    cfg.S("config.exec.sub.paracross.validatorSetDelayBlocksLow", int64(20))
    threshold, delayHigh, delayLow := validatorSetParams(cfg)
    assert.Equal(t, int64(500), threshold)
    assert.Equal(t, int64(10), delayHigh)
    assert.Equal(t, int64(20), delayLow)
}
```

- [ ] **Step 2: 运行确认失败** — `go test ./plugin/dapp/paracross/executor/ -run TestValidatorSetParams -v` → 编译错误 `undefined: validatorSetParams`。
- [ ] **Step 3: 加常量与实现** — 按上面 Interfaces 加常量；`valsetparam.go` 用 `types.ConfSub(cfg, pt.ParaX).GInt(key)` 读取，key 缺省或 ≤0 时回落到默认值（阈值默认 `100000 * cfg.GetCoinPrecision()`）。
- [ ] **Step 4: 运行确认通过** — `go test ./plugin/dapp/paracross/executor/ -run TestValidatorSetParams -v` → PASS。
- [ ] **Step 5: 提交** — `git add -A && git commit -m "feat(paracross): add validator set governance fork, ops and params"`

---

### Task 2: 提案状态编解码与 KV helper

**Files:**
- Modify: `plugin/dapp/paracross/executor/kv.go`
- Create: `plugin/dapp/paracross/executor/valsetstore.go`
- Test: `plugin/dapp/paracross/executor/valsetstore_test.go`

**Interfaces:**
- Consumes: Task 1 的常量
- Produces:
  - `type paraValidatorSetChange struct { Id, Title, FromAddr string; NewAddrs, NewBlsPubKeys []string; ProposeMainHeight, LockedAssets int64; Tier int32; Approved bool; ApprovalMainHeight int64; AutonomyItemID string; EffectiveMainHeight int64; Status int32; Version uint64 }`
  - `const validatorGroupSuper = "super"`、`const validatorGroupSupervision = "sup"`
  - `encodeValidatorSetChange(c *paraValidatorSetChange) []byte` / `decodeValidatorSetChange(b []byte) (*paraValidatorSetChange, error)`
  - `calcValidatorSetChangeKey(kind, title string) []byte` → `mavl-paracross-valset-<kind>-change-<title>`
  - `calcValidatorSetVersionKey(kind, title string) []byte` → `mavl-paracross-valset-<kind>-version-<title>`
  - `calcValidatorSetHistoryKey(kind, title string, version uint64) []byte` → `mavl-paracross-valset-<kind>-history-<title>-<version>`
  - `(a *action) getValidatorSetChange(kind, title string) (*paraValidatorSetChange, error)`（不存在返回 `pt.ErrKeyNotExist`）
  - `(a *action) saveValidatorSetChange(kind string, c *paraValidatorSetChange) *types.KeyValue`
  - `(a *action) validatorSetVersion(kind, title string) uint64`（不存在返回 0）
  - 编码格式 v1（全部大端）：`u8 version=1`，随后依次为长度前缀字符串 `Id`、`FromAddr`、`AutonomyItemID`，`u32 count` + 逐项长度前缀 `NewAddrs`，`u32 count` + 逐项长度前缀 `NewBlsPubKeys`，`i64` × 5（`ProposeMainHeight`、`LockedAssets`、`ApprovalMainHeight`、`EffectiveMainHeight`、版本号），`u8 tier`、`u8 status`、`u8 approved`。`Title` 不进编码（由 key 推导，读取时回填）。

- [ ] **Step 1: 写失败测试** — `TestValidatorSetChangeEncodeDecodeRoundTrip`（构造含 2 个 addr/blspub 的完整结构，编码→解码，`assert.Equal` 全字段）、`TestValidatorSetChangeDecodeRejectsBadVersion`（首字节改 9 → 返回 error）、`TestValidatorSetChangeDecodeRejectsTruncated`（截断一半 → error）、`TestValidatorSetChangeKeyLayout`（断言三个 key 的字符串前缀为 `mavl-paracross-valset-super-`）。
- [ ] **Step 2: 运行确认失败** — `go test ./plugin/dapp/paracross/executor/ -run TestValidatorSetChange -v` → FAIL（undefined）。
- [ ] **Step 3: 实现编解码与 key/KV helper** — 编码用 `bytes.Buffer` + `encoding/binary`；解码逐字段校验剩余长度，长度上限设 64KB 防异常输入。
- [ ] **Step 4: 运行确认通过** — 同上命令 → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): add validator set change state encoding and kv helpers"`

---

### Task 3: Propose

**Files:**
- Create: `plugin/dapp/paracross/executor/valset.go`
- Test: `plugin/dapp/paracross/executor/valset_test.go`

**Interfaces:**
- Consumes: Task 1/2；`a.checkParaTitleExecMatch(title string) error`、`a.getNodesGroup(title) (map[string]struct{}, []string, error)`（现状）、`getSupervisionNodeGroupAddrs`（现状）
- Produces:
  - `(a *action) validatorSetChangePropose(kind string, config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) proposeValidatorSetChange(kind, title, id string, addrs, blsPubKeys []string) (*types.Receipt, error)`（Task 6 复用）
  - `(a *action) lockedAssets(title string) (int64, error)`
  - `(a *action) activeValidatorSet(kind, title string) (map[string]struct{}, []string, error)`

- [ ] **Step 1: 写失败测试** — `valset_test.go` 用 `chain33TestMainCfg` + memdb + `apimock`（参考 `superaccount_test.go` 的 `NodeManageTestSuite` 装配），先 `setNodeGroupAddrs(Title, Account14K,...)` 造出生效集合：
  - `TestValidatorSetProposeOK`：addrs 乱序 + 重复 → 断言写入的提案 `NewAddrs` 已去重排序、`Status==ParaValidatorSetPending`、`Tier==ParaValidatorSetTierLow`（锁定资产 0）、`EffectiveMainHeight==ProposeMainHeight+60480`。
  - `TestValidatorSetProposeRejectsSecondPending`：连续两次 Propose → 第二次 `errors.Cause(err)==pt.ErrValidatorSetPendingExists` 且第一次的提案未变。
  - `TestValidatorSetProposeRejectsNonCanonicalTitle`：`Title="user.p.test"`（无结尾点）→ `errors.Cause(err)==pt.ErrInvalidTitle`。
  - `TestValidatorSetProposeRejectsEmptyChange`：新集合与当前集合相同 → error。
  - `TestValidatorSetProposeRejectsWhenNoActiveSet`：该 title 无生效集合 → `errors.Cause(err)==pt.ErrValidatorSetNoActive`。
  - `TestValidatorSetProposeHighTier`：`setExecAccountBalance(Title+pt.ParaX, 100000*precision+1)` → `Tier==ParaValidatorSetTierHigh` 且 `EffectiveMainHeight==0`（未批准前为 0）。
- [ ] **Step 2: 运行确认失败** — `go test ./plugin/dapp/paracross/executor/ -run TestValidatorSetPropose -v` → FAIL。
- [ ] **Step 3: 实现** — 前置校验顺序：fork → `checkParaTitleExecMatch` → 已有生效集合（`activeValidatorSet` 非空）→ 无 pending/approved 提案 → addrs/blspubs 数量相等非空 → 去重排序（addr 与 bls 配对后按 addr 排序）→ 与当前集合不同；随后 `lockedAssets` 定档、写提案 KV、发 log `pt.TyLogParaValidatorSetConfig`（log 结构沿用现有 `pt.ReceiptParaNodeGroupConfig`）。
- [ ] **Step 4: 运行确认通过** — 同上 → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): add validator set change proposal"`

---

### Task 4: Approve / Cancel

**Files:**
- Modify: `plugin/dapp/paracross/executor/valset.go`
- Test: `plugin/dapp/paracross/executor/valset_test.go`

**Interfaces:**
- Consumes: Task 3
- Produces:
  - `(a *action) validatorSetChangeApprove(kind string, config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) validatorSetChangeCancel(kind string, config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) checkValidatorSetGovernance(config *pt.ParaNodeGroupConfig) error`（主链才执行；`IsAutonomyApprovedItem(AutonomyItemID, Id)` 或 `isSuperManager`，否则 `types.ErrNotAllow`）

- [ ] **Step 1: 写失败测试**
  - `TestValidatorSetApproveBySuperManager`：`chain33TestMainCfg.S("config.exec.sub.manage.superManager", []interface{}{Account12Q})`，用 `PrivKey12Q` 签 Approve → `Approved==true`、`ApprovalMainHeight` 正确、`EffectiveMainHeight==ApprovalMainHeight+8640`。
  - `TestValidatorSetApproveRejectsNonManager`：用 `PrivKey14K` → `errors.Cause(err)==types.ErrNotAllow`。
  - `TestValidatorSetApproveRejectsUnknownProposal` → error。
  - `TestValidatorSetCancelByProposer` / `TestValidatorSetCancelByGovernance`：`Status==ParaValidatorSetCancelled`。
  - `TestValidatorSetCancelRejectsAfterActivate` → error。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — 平行链（`cfg.IsPara()`）跳过治理查询直接记录；`ApprovalMainHeight = a.exec.GetMainHeight()`；Cancel 授权：治理通过 或 `a.fromaddr == c.FromAddr`（`FromAddr` 已在 Task 2 的结构与编码中定义）。
- [ ] **Step 4: 运行确认通过** → PASS（含 Task 2/3 全部用例）。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): add validator set approval and cancel"`

---

### Task 5: Activate

**Files:**
- Modify: `plugin/dapp/paracross/executor/valset.go`
- Test: `plugin/dapp/paracross/executor/valset_test.go`

**Interfaces:**
- Consumes: Task 2/3/4；`nodeGroupCreate`（super 写集合+每节点 BLS 状态的现有实现）、`supervisionNodeGroupCreate`/`updateSupervisionNodeAddrStatus`
- Produces:
  - `(a *action) validatorSetChangeActivate(kind string, config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) writeActiveValidatorSet(kind string, c *paraValidatorSetChange) (*types.Receipt, error)`

- [ ] **Step 1: 写失败测试**
  - `TestValidatorSetActivateBeforeEffectiveFails`：把 `a.exec` 高度设为 `EffectiveMainHeight-1` → `errors.Cause(err)==pt.ErrValidatorSetNotEffective`（对应 Review Focus 2 的下边界）。
  - `TestValidatorSetActivateAtEffectiveOK`：高度 = `EffectiveMainHeight` → 集合 KV 变为新集合、`validatorSetVersion==1`、`Status==ParaValidatorSetActivated`、每节点 BLS 公钥已写入（对应 Review Focus 2 的上边界）。
  - `TestValidatorSetActivateHighTierWithoutApprovalFails`（写死在断言里，对应 Review Focus 1）。
  - `TestValidatorSetActivateTwiceFails`（Review Focus 6）。
  - `TestValidatorSetActivateRejectsVersionRegression`：手工把版本号设为 2 后再 Activate 版本 1 的提案 → error。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — 校验顺序：fork → title/execer → 提案存在且 `Status∈{Pending,Approved}` → `mainHeight ≥ EffectiveMainHeight` → `Tier==Low || Approved` → `c.Version > validatorSetVersion()`；通过后 `writeActiveValidatorSet`（super 复用 `nodeGroupCreate`，supervision 复用监督组写入）+ 版本+1 + 写 history + 状态置 Activated。
- [ ] **Step 4: 运行确认通过** → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): activate validator set changes after timelock"`

---

### Task 6: 入口接线（NodeGroupConfig / NodeConfig / SupervisionNodeConfig）

**Files:**
- Modify: `plugin/dapp/paracross/executor/superaccount.go`（`NodeGroupConfig`、`NodeConfig` 投票完成处）
- Modify: `plugin/dapp/paracross/executor/supervisionaccount.go`
- Test: `plugin/dapp/paracross/executor/valset_test.go`

**Interfaces:**
- Consumes: Task 3 `proposeValidatorSetChange`
- Produces:
  - `NodeGroupConfig` 新增 op 分发：5/6/7/8 → super 状态机；9/10 → 急停；11/12/13/14 → supervision 状态机
  - `NodeConfig` 成员变更在 2/3 达成后，**分叉后**改为调用 `proposeValidatorSetChange`（新集合 = 当前集合 ± 成员，proposal id = 当前 tx hash），不再直接 `updateNodeGroup`
  - `SupervisionNodeConfig` 同理（kind=`validatorGroupSupervision`）

- [ ] **Step 1: 写失败测试**
  - `TestNodeGroupConfigDispatchesValidatorSetOps`：op=5 走状态机（断言写入提案），op=1 仍走旧流程。
  - `TestNodeVoteCompletionCreatesProposal`：用现有 `nodeVote` 测试辅助把票数推到 2/3 → 断言 `mavl-paracross-nodes-title-<title>` **未变**、而 `valset-super-change-<title>` 已写入新集合（Review Focus 3 的前置条件）。
  - `TestBootstrapStillWritesSetDirectly`：op=1/2 首次建组仍直接写集合（回归，确保没被接管）。
  - `TestSupervisionChangeCreatesProposal`：监督组变更同样只产生提案。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — 在 `NodeGroupConfig` 的 op switch 末尾加分支；把 `nodeJoin`/`nodeQuit`/`nodeVote` 中调用 `updateNodeGroup` 的位置改为"若 fork 生效 → 计算新集合后 `proposeValidatorSetChange`，否则保持原逻辑"；`SupervisionNodeConfig` 同构。
- [ ] **Step 4: 运行确认通过** → PASS（并跑 `go test ./plugin/dapp/paracross/executor/` 全量）。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): route member changes through validator set state machine"`

---

### Task 7: 查询接口

**Files:**
- Modify: `plugin/dapp/paracross/executor/query.go`、`plugin/dapp/paracross/types/type.go`（如需注册查询名）
- Test: `plugin/dapp/paracross/executor/valset_test.go`

**Interfaces:**
- Consumes: Task 2/5
- Produces: `Query_GetValidatorSetStatus(in *pt.ReqParacrossNodeInfo) (types.Message, error)` 返回 `*types.ReplyConfig`，`Value` 为该 title 的 JSON：`{"activeVersion":1,"activeAddrs":[...],"pending":{"id":"...","status":0,"tier":0,"effectiveMainHeight":123,"approved":false}}`（无 pending 时 `pending` 为 null）。

- [ ] **Step 1: 写失败测试** `TestQueryValidatorSetStatus`：无提案、有提案、已激活三种状态各断言 JSON 字段。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — `encoding/json` 只用于查询返回值，不用于链上状态。
- [ ] **Step 4: 运行确认通过** → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): query validator set status"`

---

### Task 8: 急停开关

**Files:**
- Create: `plugin/dapp/paracross/executor/circuitbreaker.go`
- Modify: `plugin/dapp/paracross/executor/query.go`（可选：开关状态并入 Task 7 的 JSON）
- Test: `plugin/dapp/paracross/executor/circuitbreaker_test.go`

**Interfaces:**
- Consumes: Task 1/4 的治理校验
- Produces:
  - `(a *action) crossChainDisable(config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) crossChainEnable(config *pt.ParaNodeGroupConfig) (*types.Receipt, error)`
  - `(a *action) isCrossChainDisabled(title string) bool`
  - `calcCrossChainSwitchKey(title string) []byte` → `mavl-paracross-crosschain-switch-<title>`（值 = `types.Encode(&types.Int64{Data: setMainHeight})`，**key 存在即禁用**）
  - `calcCrossChainSwitchInfoKey(title string) []byte` → `mavl-paracross-crosschain-switchinfo-<title>`（值 = `types.Encode(&types.ReqMultiStrings{Datas: []string{reason, autonomyItemID}})`)
  - `Enable` 用 `Set(key, nil)` 删除两个 key。

- [ ] **Step 1: 写失败测试**
  - `TestCrossChainDisableRequiresGovernance`：非治理地址 → `types.ErrNotAllow`。
  - `TestCrossChainDisableSetsSwitch`：治理地址 → `isCrossChainDisabled==true`、info 含 reason。
  - `TestCrossChainEnableClearsSwitch`：`Enable` 后 `isCrossChainDisabled==false`。
  - `TestCrossChainDisableTwiceFails` / `TestCrossChainEnableWhenEnabledFails`（幂等边界）。
  - `TestCrossChainDisableRejectsNonCanonicalTitle`。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — 复用 Task 4 的 `checkValidatorSetGovernance` 逻辑（抽成 `checkValidatorSetGovernance(config *pt.ParaNodeGroupConfig) error` 供两者共用）。
- [ ] **Step 4: 运行确认通过** → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): governance controlled cross chain circuit breaker"`

---

### Task 9: 急停生效点（提交 + 挂起执行）

**Files:**
- Modify: `plugin/dapp/paracross/executor/asset.go`（提交入口）、`plugin/dapp/paracross/executor/action.go`（`execCrossTxs`）
- Test: `plugin/dapp/paracross/executor/circuitbreaker_test.go`

**Interfaces:**
- Consumes: Task 8
- Produces:
  - `(a *action) checkCrossChainEnabled(title string) error`（禁用时返回 `pt.ErrCrossChainDisabled`，Task 1 定义）
  - 提交入口：`AssetTransfer`、`AssetWithdraw`、`CrossAssetTransfer` 三个 action 在 fork 生效时先调 `checkCrossChainEnabled(title)`；title 用 chain33 的 `types.GetParaExecTitleName(string(a.tx.Execer))` 推导；若解析失败（非 `user.p.<name>.paracross` execer）则不做拦截并保持现状（这三个 action 在主链本来就只能以该形态提交）
  - 执行入口：`execCrossTxs` 对禁用 title **跳过**（返回空 receipt，不报错）

- [ ] **Step 1: 写失败测试**
  - `TestDisabledChainRejectsCrossAssetTransferSubmit`：设置 switch 后提交 `CrossAssetTransfer` → 失败且无 KV 变化。
  - `TestDisabledChainRejectsAssetTransferAndWithdrawSubmit`（双向：入金与出金都拒）。
  - `TestDisabledChainSkipsPendingCrossTxExecution`：构造 `crossTxs` 调 `a.execCrossTxs(title, ...)` → 返回空 receipt、无错误、账户余额不变（Review Focus 4）。
  - `TestDisabledChainCommitDoesNotFail`：走 `procCrossTxs` 路径断言 commit 仍成功。
  - `TestEnabledChainUnaffected`（回归）。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现** — 提交入口三个 action 加检查；`execCrossTxs` 循环前判断 `isCrossChainDisabled(title)`，为真直接返回 `&types.Receipt{}, nil`。
- [ ] **Step 4: 运行确认通过** → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): enforce cross chain circuit breaker on submit and execution"`

---

### Task 10: 配置文件补分叉键

**Files:**
- Modify: `chain33.fork.toml`、`chain33.para.toml`、`chain33.para.toml.readme`、`plugin/dapp/evm/cmd/ci2/chain33.proxyminer.toml`、`plugin/dapp/ticket/executor/testdata/chain33.cfg.toml`、`plugin/dapp/ticket/rpc/rpc_test.go`、`build/ci/{cross2eth,bridgevmxgo,rollup,rgbx}/chain33.para.toml`、`build/ci/evm2/chain33.proxyminer.toml`

**Interfaces:**
- Consumes: Task 1 的 `ForkParaValidatorSet`
- Produces: 每个含 `[fork.sub.paracross]` 的文件都有 `ForkParaValidatorSet=<height>`（主链模板 `chain33.fork.toml` 用一个大于当前主链高度的占位值；平行链/测试配置 `0`）

- [ ] **Step 1: 用一次性命令确认当前缺键**（不落库）

```bash
for f in $(grep -rl "fork.sub.paracross" --include=*.toml --include=*.go --include=*.readme .); do
  grep -q "ForkParaValidatorSet" "$f" || echo "MISSING $f"
done
```
Expected: 列出全部文件为 MISSING。

- [ ] **Step 2: 修改配置** — 在每个 `[fork.sub.paracross]` 段的 `Enable=` 行后插入该键（沿用 PR #1326 的写法与取值风格）。
- [ ] **Step 3: 验证脚本通过** — 同上命令，无输出。
- [ ] **Step 4: 验证配置可被解析** — `go test ./plugin/dapp/ticket/executor/ ./plugin/dapp/ticket/rpc/ -count=1`（这两个测试会加载含 `[fork.sub.paracross]` 的配置；缺键会 panic）。
- [ ] **Step 5: 提交** — `git commit -m "chore(paracross): ship ForkParaValidatorSet in every paracross fork config"`

---

### Task 11: 集成与一致性测试

**Files:**
- Create: `plugin/dapp/paracross/executor/valset_integration_test.go`

**Interfaces:**
- Consumes: Task 1–9 全部
- Produces: 端到端用例，证明主链与平行链执行同一批交易后状态一致

- [ ] **Step 1: 写失败测试**
  - `TestValidatorSetChangeMainParaConsistency`：同一批交易（Propose→Approve→Activate）分别在 `chain33TestMainCfg`（主链）与 `chain33TestCfg`（平行链，`Title="user.p.test."`）构造的 executor 上按相同顺序执行，断言两侧 `mavl-paracross-nodes-title-<title>` 内容与每节点 BLS 公钥**逐字节相同**。
  - `TestValidatorSetRollbackRestoresActiveSet`：Activate 产生 receipt 后丢弃 receipt KV（模拟回滚），断言版本号与集合仍是旧值。
  - `TestCircuitBreakerEndToEnd`：建组 → 禁用 → 提交被拒 → 挂起提现不执行 → 恢复 → 提现可执行。
- [ ] **Step 2: 运行确认失败** → FAIL（用例先写，实现已就绪时应只剩断言问题；若直接通过，说明用例无效，必须重写为能捕获上述断言的形态）。
- [ ] **Step 3: 修正实现中的不一致**（若测试暴露主链/平行链差异）。
- [ ] **Step 4: 运行确认通过** — `go test ./plugin/dapp/paracross/... -count=1` 全绿。
- [ ] **Step 5: 提交** — `git commit -m "test(paracross): validator set governance main/para consistency and breaker e2e"`

---

### Task 12: CLI 与运维文档

**Files:**
- Modify: `plugin/dapp/paracross/commands/paracross.go`
- Modify: `plugin/dapp/paracross/executor/superaccount.md`（或新增 `docs/` 运维章节）

**Interfaces:**
- Consumes: Task 7/8 的查询与 op
- Produces:
  - `paracross valset propose|approve|cancel|activate`（参数：paraName、addrs、blspubs、id、autonomyItemID）
  - `paracross crosschain disable|enable --paraName --reason`
  - `paracross valset status --paraName`
  - 运维文档：变更生命周期、急停/恢复 SOP、监控项（`valset pending` 年龄、`tier`、紧停开关状态）

- [ ] **Step 1: 写失败测试** — `plugin/dapp/paracross/commands/paracross_test.go`：`TestValidatorSetCmdBuildsPayload` 断言 `CreateTx` 的 `Execer=="user.p.test.paracross"`、`Payload` 解码后 `Op==pt.ParacrossValidatorSetPropose`。
- [ ] **Step 2: 运行确认失败** → FAIL。
- [ ] **Step 3: 实现命令与文档**（复用现有 `nodeGroupApply` 的 flag/`rpctypes.CreateTxIn` 写法）。
- [ ] **Step 4: 运行确认通过** — `go test ./plugin/dapp/paracross/commands/ -count=1` → PASS。
- [ ] **Step 5: 提交** — `git commit -m "feat(paracross): cli and ops doc for validator set governance"`

---

## 自检记录

- **Spec 覆盖**：第 5 节→Task 3-6；第 6 节→Task 1-2；第 7 节→Task 3-5；第 8 节→Task 6；第 9 节→Task 4；第 10 节→Task 8-9；第 11 节→Task 10-11；第 12 节→不在本期（仅路线）；第 13 节→Task 6（bootstrap 不变）+ Task 11；第 14 节→Task 10-12；第 15 节→各任务测试 + Task 11。
- **类型一致性**：`paraValidatorSetChange` 字段、`validatorGroupSuper/Supervision`、`validatorSetParams`、`checkValidatorSetGovernance`、`checkCrossChainEnabled` 在 Task 1–12 中命名一致。
- **Review Focus**：6 项分别落在 Task 4（治理缺失）、Task 5（生效边界/幂等）、Task 3（重复提案）、Task 9（挂起提现与 commit 不停链）、Task 10（配置缺键）、Task 5（重复 Activate）的测试步骤中。

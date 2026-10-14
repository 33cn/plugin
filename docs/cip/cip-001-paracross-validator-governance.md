---
cip: 001
title: 平行链验证人来源治理化与主链跨链急停
author: 33cn/plugin maintainers
status: Draft
type: Standards Track
category: Consensus / Cross-chain
created: 2026-10-14
requires: PR #1326（paracross title 与 execer 精确绑定）
discussions-to: https://github.com/33cn/plugin/issues
---

## 摘要

本提案把 paracross 平行链**验证人集合的更换**从"节点组内部 2/3 投票达成即生效"改为"**平行链提案 → 主链分档治理/时间锁 → 显式激活**"，并新增**主链治理可控的双向跨链急停**。

分档规则：平行链锁定资产高于阈值（默认 10 万 BTY）时，更换**必须**经主链治理批准，批准后仍有 1 天锁定期且治理可在窗口内撤销；低于或等于阈值且未获治理批准时，更换需经历 7 天锁定。生效一律由显式 `Activate` 交易在到点后完成，commit 校验的读路径保持不变。

本提案**不**声称解决"多数验证人联合作恶"。主链对平行链状态的独立验证（挑战期、欺诈证明、罚没、zk）列为 Phase 2 路线，见"实施路线"。

## 动机

paracross 的跨链权限模型把「节点组（super node group）」当作唯一权威。主链只校验 commit 是否由该组 2/3 以上 BLS 聚合签名、高度与哈希是否连续，**不判断平行链状态转移本身是否合法**。由此产生两个结构性问题：

1. **状态只由平行链自证**：主链无法独立验证"新状态是否由其交易正确推导"。
2. **验证人身份等价于"持有私钥 + 自建节点组"**：`ForkParaFreeRegister` 之后首次建组为自由注册 + 自审批；已有集合的成员变更由组内 2/3 投票决定。控制或买通一组私钥，即可左右跨链结果。

2026-10-13 的"标题前缀接管"事件即为实例：攻击者自建节点组，用伪造 title 的 commit 触发真实平行链的挂起提现，盗走约 161.8 万 BTY（分析见 `docs/security/paracross-nodegroup-title-takeover-analysis.md`）。该事件的入口缺陷已由 PR #1326 修复；本提案处理其**权限模型**层面的根因：验证人集合可以被少数私钥快速更换，且主链没有任何反应窗口与止血手段。

## 规范

### 术语与现状

- **节点组**：某一平行链 title 的授权验证人集合，存储于 `mavl-paracross-nodes-title-<title>`；每个节点的 BLS 公钥与状态存储于 `mavl-paracross-nodes-titleAddr-<title>-<addr>`。
- **监督节点组**：与节点组同构的第二组验证人，存储于 `mavl-paracross-supervision-nodes-title-<title>`，同样需要 2/3 才能达成共识。
- 节点组不是节点本地配置：**主链与平行链各存一份**，平行链共识模块本地查询自己那份，主链验证 commit 时查询主链那份。任何变更必须在两侧确定性同时生效。
- 首次建组（`NodeGroupConfig` op=1/2）维持现状，不在本提案范围内（见"安全考虑·残留风险"）。

### 分叉与参数

- 新分叉开关：`paracross.ForkParaValidatorSet`。所有新增逻辑**必须**由该分叉开启；未达分叉高度时行为与现状逐字节一致。
- 分叉判定**必须**使用 `cfg.IsDappFork(<主链高度>, "paracross", ForkParaValidatorSet)`。**不得**使用 `IsParaForkHeight`/`GetDappForkHeight`：该函数对未映射的 key 在平行链上会 panic。
- 平行链**必须**在 `[fork.sub.paracross]` 配置与主链相同的主链高度；所有包含 `[fork.sub.paracross]` 的配置文件**必须**补上该键，否则 chain33 `initForkConfig` 会使节点启动失败。
- 参数（从链配置 `config.exec.sub.paracross.*` 读取，缺省回落默认值）：

| 参数 | 默认值 | 含义 |
|---|---|---|
| `validatorSetHighValueThreshold` | `100000 × coinPrecision` | 高/低价值分档阈值 |
| `validatorSetDelayBlocksHigh` | `8640` | 治理批准后的锁定期（主链区块数，约 1 天） |
| `validatorSetDelayBlocksLow` | `60480` | 未获治理批准的低价值档锁定期（约 7 天） |

延迟**必须**按主链区块高度计算，**不得**使用墙钟时间。

### 数据结构

- 消息复用：本提案**不得**新增 protobuf 生成代码。沿用现有 `ParaNodeGroupConfig`（字段 `Title`/`Op`/`Id`/`Addrs`/`BlsPubKeys`/`AutonomyItemID`），新增以下 op：

| op | 名称 | 作用 |
|---|---|---|
| 5 | `ParacrossValidatorSetPropose` | 提交节点组变更提案 |
| 6 | `ParacrossValidatorSetApprove` | 治理批准（校验并记录） |
| 7 | `ParacrossValidatorSetCancel` | 撤销提案 |
| 8 | `ParacrossValidatorSetActivate` | 到点后激活（任何人可发） |
| 9 | `ParacrossCrossChainDisable` | 急停：切断该链跨链资金来往 |
| 10 | `ParacrossCrossChainEnable` | 解除急停 |
| 11–14 | `ParacrossSupValidatorSet{Propose,Approve,Cancel,Activate}` | 监督节点组的同一状态机 |

- 提案状态常量：`ParaValidatorSetPending=0`、`Approved=1`、`Cancelled=2`、`Activated=3`；分档常量：`ParaValidatorSetTierLow=0`、`TierHigh=1`。
- 新增错误：`ErrValidatorSetNotEffective`、`ErrValidatorSetPendingExists`、`ErrValidatorSetNoActive`、`ErrCrossChainDisabled`。
- 状态对象（组别维度，`kind ∈ {super, sup}`）：

```go
type paraValidatorSetChange struct {
    Id, Title, FromAddr string
    NewAddrs, NewBlsPubKeys []string
    ProposeMainHeight, LockedAssets int64
    Tier int32
    Approved bool
    ApprovalMainHeight int64
    AutonomyItemID string
    EffectiveMainHeight int64
    Status int32
    Version uint64
}
```

- KV 布局：

| Key | 值 | 写入者 |
|---|---|---|
| `mavl-paracross-valset-<kind>-change-<title>` | 提案状态（确定性自定义编码，见下） | Propose/Approve/Cancel/Activate |
| `mavl-paracross-valset-<kind>-version-<title>` | 当前生效版本号 | Activate |
| `mavl-paracross-valset-<kind>-history-<title>-<version>` | 该版本的集合与生效高度（供 Phase 2 追责） | Activate |
| `mavl-paracross-nodes-title-<title>` | 现有 key，**仅 Activate 改写** | Activate（节点组） |
| `mavl-paracross-supervision-nodes-title-<title>` | 现有 key，**仅 Activate 改写** | Activate（监督组） |
| `mavl-paracross-crosschain-switch-<title>` | 急停高度；**key 存在即禁用** | Disable/Enable（Enable 时删除） |
| `mavl-paracross-crosschain-switchinfo-<title>` | `{reason, autonomyItemID}` | Disable |

- 提案状态的编码**必须**确定性：版本字节 `=1`，其后依次为大端长度前缀字符串 `Id`、`FromAddr`、`AutonomyItemID`，`u32` 计数 + 逐项长度前缀的 `NewAddrs` 与 `NewBlsPubKeys`，`i64 × 5`（`ProposeMainHeight`/`LockedAssets`/`ApprovalMainHeight`/`EffectiveMainHeight`/版本号），`u8 tier`、`u8 status`、`u8 approved`；`Title` 不进编码，由 key 推导。解码**必须**校验版本字节与长度上限（64KB），异常输入返回错误而非 panic。
- 新增日志类型 `TyLogParaValidatorSetConfig`，用于索引提案与激活事件。

### 规则

**锁定资产口径**：某 title 的锁定资产 = `LoadExecAccount(ExecAddress(title + "paracross"), ExecAddress("paracross")).Balance`（主链视角，与 `executor/asset.go` 的入金账户一致），在 `Propose` 时快照定档。

**Propose** 必须满足：

1. 分叉已生效；
2. `Title` 为规范形式 `user.p.<name>.` 且与交易 execer 精确对应（PR #1326 的 `IsValidParaTitle` + `checkParaTitleExecMatch`）；
3. 该 title 已有生效集合，否则返回 `ErrValidatorSetNoActive`（首次建组走既有 op=1/2）；
4. 同一 title **不得**同时存在 `Pending`/`Approved` 提案，否则返回 `ErrValidatorSetPendingExists`；
5. `Addrs` 与 `BlsPubKeys` 数量相等且非空；地址去重并按地址排序（BLS 公钥随地址同步重排），保证主链与平行链写入字节一致；
6. 新集合与当前生效集合不得完全相同。

提案 `Id` = 提案交易哈希，`FromAddr` = 交易发起地址。`Tier` 按锁定资产是否**严格大于**阈值判定；未获治理批准前 `EffectiveMainHeight = 0`。

**Approve**：提案必须为 `Pending`；治理判定**仅在主链执行**（`if !cfg.IsPara()`），平行链跟随主链结果。治理通过 = autonomy `IsAutonomyApprovedItem(AutonomyItemID, Id)` 成功，或 `isSuperManager(fromaddr)`；否则返回 `ErrNotAllow`。成功后记录 `Approved=true`、`ApprovalMainHeight = 当前主链高度`、`EffectiveMainHeight = ApprovalMainHeight + validatorSetDelayBlocksHigh`。

**Cancel**：`Activate` 之前任意时刻，治理（autonomy 或超级管理员）或提案发起地址（`FromAddr`）可以撤销；状态置 `Cancelled`，不影响当前生效集合。

**Activate**（任何人可发）必须同时满足：

1. 提案状态为 `Pending` 或 `Approved`；
2. 当前主链高度 ≥ `EffectiveMainHeight`；未获批准的 `TierHigh` 提案 `EffectiveMainHeight` 恒为 0，但**不得**因此被激活——`Tier == TierHigh && !Approved` 时**必须**拒绝并返回 `ErrValidatorSetNotEffective`；
3. 提案版本号 > 当前生效版本号。

执行内容：写入该组别的集合 KV 与每节点状态（含 BLS 公钥，复用现有 `nodeGroupCreate` / 监督组写入路径）、版本号加一、写 history、状态置 `Activated`。**重复 Activate 必须失败**（幂等）。

### 入口接管

| 现有入口 | 现状 | 分叉后 |
|---|---|---|
| `NodeGroupConfig` op=1/2 | 首次建立集合（自由注册 + 自审批） | **保持不变** |
| `NodeGroupConfig` op=3/4 | 撤销未批准申请 / 修改冻结参数 | 不变（不涉及成员集合） |
| `NodeConfig` op=NewApply/Vote/Quit/Cancel/Modify | 成员变更，内部 2/3 投票达成后直接改写集合 | **改为产出 `Pending` 提案**（新集合 = 当前集合 ± 成员，`Id` = 该交易哈希） |
| `SupervisionNodeConfig` | 监督节点组变更 | 同上（`kind=sup`） |

即：**内部 2/3 投票达成不等于生效**。平行链侧的提案与投票仍由平行链自己完成，主链新增的是治理闸门与时间锁。

### 跨链急停

- 粒度：per 平行链 title。触发与解除**必须**经治理批准（autonomy 项或超级管理员），并链上记录原因与设置高度。
- 生效范围（**双向**）：
  1. 新提交：该链的 `AssetTransfer`、`AssetWithdraw`、`CrossAssetTransfer`（execer 为 `user.p.<title>.paracross`）**必须**失败并返回 `ErrCrossChainDisabled`；
  2. 已挂起：`execCrossTxs` 执行挂起提现（ExecPack）时，被禁用 title 的交易**必须**跳过——不执行、不改状态，且**不得**让 commit 交易失败（避免停链）；
  3. 解除后挂起提现可正常执行，无需重新提交；
  4. 不影响已验证历史与验证人变更流程（两者独立）。
- 被禁用的 parallel chain 状态**不得**影响其它 title。

### 查询

新增 `GetValidatorSetStatus`（入参 `ReqParacrossNodeInfo`，返回 `ReplyConfig`，值为 JSON）：包含 `activeVersion`、`activeAddrs`、`pending`（`id`/`status`/`tier`/`approved`/`effectiveMainHeight`）与急停开关状态。JSON 仅用于查询返回，**不得**用于链上状态编码。

## 理由

- **显式 `Activate` 而非按高度动态解析**：动态解析会改动 commit 校验的读路径（`getNodesGroup`），主链、平行链、历史查询与回滚都必须采用同一高度口径；任一未升级节点会在同一高度算出不同集合。显式激活把切换收敛为一个确定的状态写入，读路径不变，共识风险最小。
- **复用 `ParaNodeGroupConfig`**：避免新增 protobuf 生成代码与 action/oneof 变更；`Addrs`/`BlsPubKeys`/`Id`/`AutonomyItemID` 已足够表达提案与治理项。
- **延迟用主链高度而非时间**：区块高度在所有节点上确定，墙钟时间不是共识输入。
- **高价值链 1 天 + 治理批准、低价值链 7 天免批准**：高价值链的额外风险由治理承担，因此给更短的窗口但必须有批准；低价值链用更长的窗口替代治理介入，兼顾开链效率。
- **首次建组不收紧**：主链不存在平行链注册表，无法在合约内判定 title 真伪；收紧会改变既定共识语义与开链流程。该缺口作为已接受风险记录，并由急停与监控缓解。

## 向后兼容

- 未达分叉高度：行为与现状完全一致，无需迁移。
- 分叉高度后：**必须**先升级主链与**全部**平行链节点。未升级的平行链节点会与主链对"有效集合"的判断不一致，导致共识停摆或分叉——与既有 paracross 分叉的升级要求相同。
- 已存在的节点组视为 version 1，无需迁移；其后续成员变更走本状态机。
- 现有查询与 CLI 保持可用；新增查询与命令为增量。

## 安全考虑

**本提案缓解**

1. 私钥被控后可进行的"集合更换"被强制延迟 1/7 天，并在高价值链上必须经治理批准；
2. 治理可在窗口内撤销变更，并可用急停立即切断跨链资金来往；
3. 变更全过程链上可审计（提案、批准、激活、急停均有事件与原因字段）。

**本提案明确不缓解**

- 现有集合多数联合作恶仍可按旧规则签名并提交伪造 commit 盗取资产（无欺诈证明与罚没）——需 Phase 2；
- 治理被控的风险：被控治理可以批准恶意变更、滥用急停（只能"停"不能"取"，但可造成跨链拒绝服务）。

**残留风险**

- 首次建组仍为自由注册：无生效集合的平行链仍可能被他人抢先建立集合，此后才受时间锁保护。缓解：急停止血；监控"已有平行链 title 出现新节点组/自审批"。
- 参数（阈值、延迟区块数）若各节点配置不一致，会造成两侧判断不一致。缓解：纳入升级检查清单，后续可迁移为治理参数/mver 固化。
- `Activate` 需要有人发送（低价值链不会自动生效）。缓解：任何人可发；CLI 与监控提示。

## 测试与参考实现

参考实现：`docs/superpowers/plans/2026-10-14-paracross-validator-governance.md`（12 个 TDD 任务），设计工作稿：`docs/superpowers/specs/2026-10-14-paracross-validator-governance-design.md`。

必须覆盖的用例：

- 状态机全部合法/非法转移：重复 Propose、缺治理批准的高价值 Activate、批准后 Cancel、已激活提案再操作、重复 Activate、版本回退；
- 分档边界：锁定资产 = 阈值、阈值 + 1；
- 延迟边界：`EffectiveMainHeight - 1` 失败、`EffectiveMainHeight` 成功；
- 主链/平行链一致性：同一批交易在两侧执行后集合 KV 与 BLS 公钥逐字节相同；
- 回滚：丢弃 Activate 的 receipt KV 后版本号与集合回到旧值；
- 急停：双向提交被拒、挂起提现不执行且 commit 不失败、解除后恢复；
- 回归：现有 nodegroup/node/supervision/asset/rollup 测试全绿；
- 配置：所有含 `[fork.sub.paracross]` 的配置文件补键后可正常解析（缺键会 panic）。

## 实施路线

- **Phase 1（本提案范围）**：验证人集合变更状态机 + 治理闸门 + 跨链急停。交付后形成"急停（止血）→ 变更延迟（反应窗口）"两层防御。
- **Phase 2（后续 CIP）**：主链可验证平行链状态。
  1. 跨链转出进入挑战期，watcher 提交**见证式欺诈证明**（引用 `{paraHeight, PreStateHash, StateHash, 被质疑交易}`，用 chain33 现有 mavl 证明验证被触碰 key），挑战成立则撤销提现；
  2. 随机抽样审计 + 罚没（前置项：先引入质押，本提案未引入）；
  3. zk 有效性证明（长期）；
  4. 数据可用性补齐（当前平行链交易已通过 tx group 上主链，暂不缺）。

## 版权

CIP 文本按 [CC0-1.0](https://creativecommons.org/publicdomain/zero/1.0/) 放弃版权；参考实现代码遵循本仓库 BSD 3-Clause License。

# 平行链验证人来源治理化 + 主链跨链急停（Phase 1 设计）

- 日期：2026-10-14
- 状态：设计已逐节确认（第 1–6 节），待编写实施计划
- 相关文档：`docs/security/paracross-nodegroup-title-takeover-analysis.md`（"标题前缀接管"漏洞分析与修复，PR #1326）
- 相关代码：`plugin/dapp/paracross`（执行器）、`plugin/consensus/para`（平行链共识）、`plugin/dapp/rollup`、autonomy `plugin/dapp/autonomy`

## 1. 背景与动机

paracross 的跨链权限模型把「节点组（super node group）」当作唯一权威：主链只校验 commit 是否由该组 2/3 以上 BLS 聚合签名、高度与哈希是否连续，**从不判断平行链的状态转移本身是否合法**。

这带来两个结构性问题：

1. **状态只由平行链自证**：主链无法独立判断"这条平行链新产生的状态是否由其交易正确推导而来"，只能相信签名集合。
2. **验证人身份等价于"持有私钥 + 自建节点组"**：`ForkParaFreeRegister` 之后，首次建组是自由注册 + 自审批；已有集合的成员变更由组内 2/3 投票决定。因此控制（或买通）一组私钥，或抢先注册一个无人认领的 title，就能左右跨链结果。

2026-10-13 的"标题前缀接管"事件（`docs/security/`）正是该模型被利用的实例：攻击者自建节点组后，用伪造 title 的 commit 触发真实平行链的提现，盗走约 161.8 万 BTY。

本设计不追求一次性彻底去信任化。Phase 1 的目标是把"更换验证人"从**内部投票即生效**改成**平行链提案 + 主链分档治理/时间锁 + 显式激活**，并给治理一个**跨链急停**开关；主链对状态的独立验证（挑战期/欺诈证明/罚没/zk）作为 Phase 2 在同一文档中给出路线与接口。

## 2. 威胁模型与目标

**威胁模型**

- 攻击者可以取得或买通某条平行链当前验证人集合的私钥（含 BLS 私钥）；
- 攻击者可以自建/抢先注册无人认领的平行链 title 节点组；
- 攻击者可能提交伪造的平行链状态与 commit（占用主链算力与状态空间）；
- 治理（autonomy board）被部分或完全控制属于需要评估的风险，但不是"可以随意盗币"的前提。

**Phase 1 目标（本期交付）**

1. 验证人集合的**任何更换**都有最短反应窗口（1 天或 7 天），窗口内治理与提案人都能撤销；
2. 高价值平行链（锁定资产 > 阈值）的更换**必须**经治理批准；
3. 治理可以**切断**指定平行链的跨链资金来往（双向），作为止血开关；
4. 以上机制不改变 commit 校验、BLS 阈值与提现的最终性语义。

**Phase 1 不承诺（明确不解决）**

- 不承诺"多数验证人联合作恶也无法盗取资产"——这需要 Phase 2 的主链独立验证；
- 不改变"首次建组自由注册"的现状（见第 13 节残留缺口）；
- 不引入质押与罚没（Phase 2 罚没的前置项）。

## 3. 范围

| 项 | Phase 1 | Phase 2（仅路线，本期不实现） |
|---|---|---|
| 验证人集合变更 | 治理闸门 + 时间锁 + 显式激活 | 与罚没挂钩 |
| 跨链急停 | 治理可控的双向开关 | 与挑战期/自动风控联动 |
| 主链状态验证 | 不做 | 挑战期 + 见证式欺诈证明；抽样审计；zk 有效性证明 |
| commit/BLS/提现最终性 | 不变 | 提现引入挑战期（最终性延后） |

**非目标**：把平行链共识替换为主链共识；主链全量重执行平行链；在 Phase 1 引入 zk 电路。

## 4. 代码现状核对（设计依据）

| 事实 | 位置 |
|---|---|
| 节点组地址集合存在 `mavl-paracross-nodes-title-<title>` | `executor/kv.go:calcParaNodeGroupAddrsKey` |
| 每个节点的 BLS 公钥与状态存在 `mavl-paracross-nodes-titleAddr-<title>-<addr>` | `executor/superaccount.go:updateNodeAddrStatus`、`action.go:getAddrBlsPubKey` |
| 首次建组：`NodeGroupConfig` op=1/2（Apply/Approve），`ForkParaFreeRegister` 后自审批 | `executor/superaccount.go:nodeGroupApply/nodeGroupApprove/checkApproveOpNew` |
| 成员变更：`NodeConfig`（NewApply/Vote/Quit/Cancel/Modify），2/3 内部投票达成后改写集合 | `executor/superaccount.go:nodeJoin/nodeVote/nodeQuit/updateNodeGroup` |
| 监督节点组同构，也是必须共同达成 2/3 的一类验证人 | `executor/supervisionaccount.go` |
| 平行链共识模块**本地查询**自己的节点组状态（主链与平行链各存一份） | `plugin/consensus/para/paracommitmsg.go:getNodeGroupAddrs`（注释："node group 会在主链和平行链都同时配置，只本地查询就可以"） |
| 主链 commit 校验读主链那份集合 | `executor/action.go:getNodesGroup` |
| 治理判定复用 autonomy 与超级管理员 | `executor/superaccount.go:checkApproveOpOld/checkApproveOpNew`（`IsAutonomyApprovedItem`） |
| 链配置参数参与共识判定的先例 | `executor/superaccount.go:nodeGroupCoinsFrozen`（`nodeGroupFrozenCoins`） |
| 首次建组仅按"完全同名 key"判重 | `executor/superaccount.go:checkNodeGroupExist` |

**结论**：节点组不是本地配置，而是**主链与平行链各一份、必须保持一致**的链上状态。任何验证人变更都必须在两侧确定性同时生效，否则主链与平行链对"谁有权签名"的判断分叉，共识停摆。

## 5. 总体架构

### 5.1 组件

1. **`ValidatorSetChange` 状态机**（paracross 执行器内，主链与平行链各执行一份）：
   提案 →（可选治理批准）→ 时间锁 → 激活；激活时改写节点组地址集合与每节点状态（含 BLS 公钥），并递增集合版本号。
2. **治理闸门**：复用 autonomy（`IsAutonomyApprovedItem`）与超级管理员；治理判定**只在主链做**，平行链按主链执行结果跟随（与现有 `NodeGroupApprove` 完全相同的模式）。
3. **跨链急停（circuit breaker）**：title 维度的双向跨链开关，治理可切断/恢复。
4. **分叉与升级**：主链新分叉高度；平行链按既有 `mainFork*` 约定配置同一主链高度。

### 5.2 核心不变式

> **任一时刻，"当前有效验证人集合"只由已被 `Activate` 的版本决定，绝不按高度动态推导；主链与平行链读到的是同一个集合。**

这是选择"延迟到点后由显式激活交易改写"（而非"按高度动态解析"）的根本原因：commit 校验的读路径（`getNodesGroup`）完全不变，共识风险最小。

### 5.3 数据流

```
平行链治理（现有节点组内部 2/3 投票，或平行链链上治理合约）
   └─> 主链交易 op=Propose(execer = user.p.<title>.paracross, 新集合 + BLS 公钥)
         ├─ 快照该链锁定资产 → 定档
         ├─ 高价值档(>阈值): 必须 op=Approve(autonomy 项) → 生效高度 = 批准高度 + 1 天
         ├─ 低价值档(≤阈值): 无需批准      → 生效高度 = 提案高度 + 7 天
         └─ 窗口内: op=Cancel(治理或提案人) 可撤销
   └─> 到点后 op=Activate(任何人可发)
         └─ 改写 mavl-paracross-nodes-title-<title> 与每节点状态, 版本号 +1
```

## 6. 数据结构与接口

### 6.1 复用现有消息，不新增 protobuf/action

`ParaNodeGroupConfig` 已包含 `Title / Op / Id / Addrs / BlsPubKeys / AutonomyItemID`，足够表达新集合与治理项。因此**在同一个 action 上新增 op**，不新增 `ParacrossAction` oneof 字段、不重新生成 `.pb.go`、不改 `GetTypeMap` 与 `allow()`：

```go
// plugin/dapp/paracross/types/paracross.go 的 ParacrossNodeGroup* 常量后追加
ParacrossValidatorSetPropose  = 5
ParacrossValidatorSetApprove  = 6
ParacrossValidatorSetCancel   = 7
ParacrossValidatorSetActivate = 8
ParacrossCrossChainDisable    = 9
ParacrossCrossChainEnable     = 10
// 监督节点组：Propose=11 / Approve=12 / Cancel=13 / Activate=14
// 分档阈值治理：ParamPropose=15 / ParamApply=16
ParacrossValidatorSetParamPropose = 15
ParacrossValidatorSetParamApply   = 16
```

字段映射：`Title`=平行链 title；`Addrs`=新集合地址（逗号分隔）；`BlsPubKeys`=与 `Addrs` 一一对应的 BLS 公钥；`Id`=提案 id；`AutonomyItemID`=治理批准项 id。

### 6.2 新增 KV（沿用 `mavl-paracross-` 前缀）

| Key | 值 | 写入者 |
|---|---|---|
| `mavl-paracross-valset-change-<title>` | 当前待生效提案（第 6.3 节结构） | Propose/Approve/Cancel |
| `mavl-paracross-valset-version-<title>` | 当前生效版本号 `uint64` | Activate |
| `mavl-paracross-valset-history-<title>-<version>` | 该版本的集合与 `ActiveFromMainHeight` | Activate（供 Phase 2 追责） |
| `mavl-paracross-valset-threshold` | 全局分档阈值（int64，coin 精度） | ParamApply |
| `mavl-paracross-valset-paramchange` | 待应用阈值提案 `{id, value}` | ParamPropose（Apply 后删除） |
| `mavl-paracross-crosschain-switch-<title>` | 急停状态结构 | Disable/Enable |
| `mavl-paracross-nodes-title-<title>` | **现有 key，仅 Activate 改写** | Activate |
| `mavl-paracross-nodes-titleAddr-<title>-<addr>` | **现有 key**，每节点状态与 BLS 公钥 | Activate（复用 `nodeGroupCreate` 的写入逻辑） |

### 6.3 状态对象

```go
type ParaValidatorSetChange struct {
    Id                  string // 提案 id（Propose 交易 hash）
    Title               string
    NewAddrs            string // "addr1,addr2,..."
    NewBlsPubKeys       string // "pub1,pub2,..."
    ProposeMainHeight   int64
    LockedAssets        int64  // Propose 时快照
    Tier                int32  // 0=低价值(≤阈值) 1=高价值(>阈值)
    Approved            bool
    ApprovalMainHeight  int64
    AutonomyItemID      string
    EffectiveMainHeight int64  // 见 6.4
    Status              int32  // 0=Pending 1=Approved 2=Cancelled 3=Activated
    Version             uint64 // 激活后版本号
}

type ParaCrossChainSwitch struct {
    Disabled        bool
    Reason          string
    SetMainHeight   int64
    AutonomyItemID  string
}
```

### 6.4 参数与决定方式

| 参数 | 初始/默认值 | 决定方式 |
|---|---|---|
| 分档阈值（`mavl-paracross-valset-threshold`） | `100000` × coinPrecision（10 万 BTY） | **主链治理投票**（见 6.5）；链上无该 key 时用代码默认值 |
| `validatorSetDelayBlocksHigh` | `8640` | 分叉参数（链配置，按 10s/块 ≈ 1 天） |
| `validatorSetDelayBlocksLow` | `60480` | 分叉参数（链配置，≈ 7 天） |

阈值**不得**以代码常量或节点配置作为运行值：它由主链治理投票决定，代码默认值只在链上尚无该 key 时生效。延迟仍走链配置（口径同现有 `nodeGroupFrozenCoins`）。

延迟一律按**主链区块高度**计算，不使用墙钟时间（共识确定性）：

- 已批准（高或低价值档）：`EffectiveMainHeight = ApprovalMainHeight + validatorSetDelayBlocksHigh`
- 未批准且低价值档：`EffectiveMainHeight = ProposeMainHeight + validatorSetDelayBlocksLow`
- 高价值档未获批准：**不可激活**（`Activate` 失败）

### 6.5 阈值治理化（两段式）

分档阈值由主链治理投票决定，流程复用现有 autonomy `PropItem`（董事会对某笔交易 hash 投票），与现有 `NodeGroupApply → autonomy item → NodeGroupApprove` 完全同构：

1. `ParamPropose`（op=15）：新阈值承载于 `ParaNodeGroupConfig.CoinsFrozen`，写入 `{id = 本交易 hash, value}`；**不改生效阈值**；
2. 董事会针对该提案交易 hash 发起 autonomy item 并投票通过；
3. `ParamApply`（op=16）：携带 `AutonomyItemID` 与 `Id`（= 提案交易 hash），校验 `IsAutonomyApprovedItem` 通过后写入 `mavl-paracross-valset-threshold`，并删除待应用提案。

**必须两段式的原因**：autonomy item 只能绑定一个交易 hash；若只用一笔交易（数值写在 item 描述里），链上无法证明董事会批准的就是该数值，任何持有该 itemID 的人都能提交任意阈值。两段式让「提案交易 hash」同时绑定「新数值」。

阈值变更**立即生效**，但**只影响此后新提交的提案**——提案在 `Propose` 时已快照 `Tier`，阈值变化不得改变已存在提案的档位（否则等于绕过时间锁）。阈值变更不影响任何链的锁定资产，因此不设额外时间锁。

## 7. 状态机

### 7.1 状态与转移

```
                 Propose                      Approve(治理)
   (无提案) ──────────────> PENDING ──────────────────────────> APPROVED
                              │  ▲                                  │
                              │  └────────── Cancel ────────────────┤
                              │      (治理或提案人, Activate 前任意时刻)
                              ▼                                     ▼
                          CANCELLED                            Activate
                    (mainHeight ≥ Effective 且 (低价值 或 已批准))
                                                                 │
                                                                 ▼
                                                             ACTIVATED
                                        (改写集合 + 每节点状态, 版本号 +1)
```

### 7.2 规则与幂等

1. **Propose 前置**：`title` 必须已有生效集合（首次建组仍走既有 op=1/2 流程，见第 8 节）；同一 title 只允许一个 `PENDING/APPROVED` 提案，重复 Propose 直接失败；`Addrs` 与 `BlsPubKeys` 数量必须相等且非空；地址去重后排序，保证两侧写入字节一致；新集合与当前生效集合完全相同则 Propose 失败（空变更）。
   提案携带的是**绝对新集合**（不是增量）。由于同一时刻只有一个待生效提案、且生效集合只能由 `Activate` 改写，提案期间当前集合不会变化，`Activate` 写入的集合与提案内容必然一致。
2. **Approve 前置**：提案存在且为 `PENDING`；`IsAutonomyApprovedItem(AutonomyItemID, Id)` 通过（仅主链判定）；记录 `Approved=true` 与 `ApprovalMainHeight`。
3. **Cancel**：治理（autonomy 或超级管理员）可在任意时刻撤销未激活提案；提案人（`Id` 对应的原始提案地址）同样可撤销。Cancel 后状态置 `CANCELLED`，不影响当前生效集合。
4. **Activate**：任何人可发；必须同时满足 —— 提案状态为 `PENDING/APPROVED`、`mainHeight ≥ EffectiveMainHeight`、`tier==0 || Approved`、目标版本号 > 当前版本号。执行内容：写入集合与每节点状态（复用 `nodeGroupCreate` 的写入路径）、`valset-version+1`、`valset-history` 记录、提案状态置 `ACTIVATED`。
5. **幂等**：同版本重复 Activate 失败；`Activated/Cancelled` 状态的提案不可再 Approve/Activate。
6. 所有状态变更都产生独立的 receipt log（含 title、提案 id、新版本号、生效高度），便于索引与监控。
7. **阈值状态机**：`ParamPropose` 要求新阈值 > 0、与当前生效阈值不同、且当前无待应用提案；`ParamApply` 要求待应用提案存在、`Id` 与其绑定、治理校验通过；应用后阈值立即对**此后**的 `Propose` 生效，已存在提案的 `Tier` 不变。

## 8. 变更入口接管

| 现有入口 | 现状作用 | Phase 1 之后 |
|---|---|---|
| `NodeGroupConfig` op=1/2（Apply/Approve） | **首次建立集合**（自由注册 + 自审批） | **保持不变**（本期决策：不收紧开链；缺口见第 13 节） |
| `NodeGroupConfig` op=3/4（Quit/Modify） | 撤销未批准申请 / 修改冻结参数 | 不涉及成员集合，保持不变 |
| `NodeConfig` op=NewApply/Vote/Quit/Cancel/Modify | **已有集合的成员变更**（内部 2/3 投票达成后直接改写集合） | **被状态机接管**：投票达成后不再改写集合，而是生成/更新 `PENDING` 提案（新集合 = 当前集合 ± 成员） |
| `SupervisionNodeConfig` | 监督节点组（第二组验证人，同样需要 2/3） | 同样接管 |

即：**内部 2/3 投票达成 ≠ 生效**。平行链侧投票仍由平行链自己完成（保留"平行链自身治理发起"的自主性），主链侧新增的是治理闸门与时间锁。

## 9. 治理闸门

- 治理主体：autonomy 自治 board（`types.AutonomyCfgKey` + `IsAutonomyApprovedItem`），超级管理员（`isSuperManager`）作为兜底；
- 治理批准项（autonomy item）的发起方式不在本设计范围内改变：沿用现有流程——由平行链治理在主链 autonomy 发起提案、board 成员投票通过后 `IsAutonomyApprovedItem(itemID, proposalId)` 返回成功。本设计的 `Approve` 交易只是**校验并记录**该批准结果，不改变 autonomy 自身的投票机制；
- 判定只在主链执行（`if !cfg.IsPara()`），平行链跟随主链执行结果，与现有 `nodeGroupApprove` 的模式一致；
- 高价值档必须批准；低价值档可选择批准以把等待期从 7 天缩短到 1 天；
- 批准后 1 天窗口内治理可撤销；`Activate` 之前任意时刻治理与提案人均可撤销；
- 分档阈值同样由治理投票决定（6.5），但走独立的 `ParamPropose/ParamApply` 两段式：阈值是全局参数、影响此后所有提案的档位，因此不复用单个 title 的提案对象。

## 10. 跨链急停

- 状态：`mavl-paracross-crosschain-switch-<title>`，per 平行链；
- 操作：op=9 `CrossChainDisable` / op=10 `CrossChainEnable`，均需治理批准（autonomy 项，超级管理员兜底），链上记录原因与高度；
- 生效范围（**双向**，已确认）：
  1. 新提交：`CrossAssetTransfer` / `AssetWithdraw` / 存款类跨链动作（execer = `user.p.<title>.paracross`）直接失败；
  2. 已挂起：commit 触发 `execCrossTxs` 时，被禁用链的挂起提现（ExecPack）不执行、不改状态；
  3. 恢复：`Enable` 后挂起提现可正常执行，不额外延迟；
  4. 不影响：历史已验证状态、验证人变更流程（两者独立，急停期间仍可走变更流程）。
- 安全性：急停只能"停"不能"取"，滥用最多造成跨链暂停（拒绝服务），不会造成资金损失；开关变化全程有事件与原因字段。

## 11. 共识安全

- **两侧一致**：`Activate`、`Disable/Enable` 都是普通 paracross 交易，主链与平行链各自执行同一笔交易，写入 KV 完全确定；不使用墙钟时间，不用随区块变化的本地状态。
- **不改读路径**：commit 校验仍只读 `mavl-paracross-nodes-title-<title>`，因此验证人切换不会引入"同一高度不同节点算出不同集合"的风险。
- **分叉**：主链注册 `ForkParaValidatorSet`；平行链在 `[fork.sub.paracross]` 配置同一主链高度（与既有 `ForkParaAssetTransferRbk`/`ForkParaSupervision` 等分叉一致）。使用 `cfg.IsDappFork(mainHeight, pt.ParaX, fork)` 判定，不使用 `IsParaForkHeight`（后者对未映射 key 在平行链上会 panic）。
- **升级要求**：分叉高度前必须完成主链与**全部平行链节点**的升级，否则未升级节点在分叉高度后与本设计判断不一致。
- **回滚**：Activate/Disable 只写 receipt KV，框架自动回滚，无需额外逻辑。
- **参数一致性**：第 6.4 节参数必须全网一致（进链配置并纳入升级检查清单）。

## 12. Phase 2 演进路线（主链也能验证平行链状态）

1. **挑战期 + 见证式欺诈证明（推荐第一步）**
   - 跨链转出进入 `PendingWithdrawal` 队列，等待 `challengeWindow`（建议与 Phase 1 的 1 天/7 天口径统一）；
   - 任何 watcher 可提交 `FraudProof`，引用 `{paraHeight, PreStateHash, StateHash, 被质疑交易}`，用 chain33 现成的 mavl 证明（`Tree.ConstructProof` / `Proof.Verify(key,value,root)`，见 `system/store/mavl/db/proof.go`）给出被触碰 key 的状态见证；
   - 主链只需执行被质疑交易涉及的账户/合约子集即可判定，无需全量重执行、无需 zk；
   - 挑战成立：撤销该提现、回滚相关状态；有质押则罚没。
2. **抽样审计 + 罚没**：随机抽取平行链高度，要求验证人在时限内提供见证，超时/错误即罚没。**前置项：Phase 1 未引入质押，罚没需要先补质押机制。**
3. **有效性证明（zk）**：平行链随块提交状态转移证明，主链验证；无需挑战期与诚实方，但需 zkVM/电路与验证器，列为长期目标。
4. **数据可用性**：目前平行链交易已通过 tx group 上主链（`FilterParaTxsByTitle` 的数据来源），暂不缺；若未来改为只传哈希，需要 DA 采样或强制上链。

**分层防御**：急停（止血）→ 变更延迟（反应窗口）→ 挑战期（纠错）→ zk（免信任）。Phase 1 交付前两层。

## 13. 风险登记与残留缺口

| 风险 | 说明 | 缓解 |
|---|---|---|
| 首次建组仍自由注册（本期决策 C） | 无生效集合的平行链仍可能被他人抢先建立集合，此后才受 1/7 天保护 | 治理急停可止血；监控"已有平行链 title 出现新节点组/自审批"；后续可选：对"无有效集合"的链要求一次性治理批准开链 |
| 治理被控 | board 被控可批准恶意变更、可滥用急停 | 变更仍有最短窗口且可撤销；急停只能停不能取；board 门槛 + 超级管理员兜底 + 全程事件审计 |
| 现有集合私钥被控 | 本设计不改变签名验证，被控私钥仍能按旧规则出块/commit | 变更被强制延迟 1/7 天，治理可在窗口内撤销并急停；彻底解决依赖 Phase 2 的"可证明作恶 + 罚没" |
| 参数/配置不一致 | 阈值与延迟区块数若各节点不同 → 两侧判断分叉 | 参数进链配置；纳入升级检查清单；后续用治理参数/mver 固化 |
| 未升级节点 | 分叉后两侧集合判断不一致 | 强制升级窗口 + 分叉高度留足；启动校验 |
| `Activate` 无人发送 | 低价值档不会自动生效 | 任何人可发（提案人/节点/浏览器/脚本）；CLI 与监控提示；**不**改为 commit 时顺带激活（会动读路径，违背方案取舍） |

## 14. 迁移步骤

1. 发布新版本（主链 + 全部平行链节点），`ForkParaValidatorSet` 设在部署后若干区块；
2. 平行链按 `mainFork*` 约定配置同一主链高度；
3. 分叉后已有节点组自动视为 version 1，无需迁移；此后成员变更走状态机；
4. 配套交付：查询接口（提案状态、生效高度、当前版本、急停状态）、CLI 命令、监控指标与告警；
5. 运维文档：急停与恢复的操作流程、变更提案的生命周期说明。

## 15. 测试矩阵

**单元测试**

- 状态机全部合法/非法转移：重复 Propose、缺治理批准的高价值 Activate、批准后 Cancel、Cancel 后 Activate、已完成提案再操作；
- 分档边界：锁定资产 = 阈值、阈值 + 1；
- 延迟边界：`EffectiveMainHeight - 1`（失败）、`EffectiveMainHeight`（成功）；
- 幂等：同版本重复 Activate 失败；
- 急停：新提交被拒（双向）、挂起提现不执行、恢复后可执行、无治理批准不可切换；
- 阈值治理：未批准 `ParamApply` 失败、`Id` 与提案不匹配失败、重复 `ParamPropose` 失败、应用后新提案按新阈值分档且已有提案档位不变、链上无 key 时回落到代码默认值。

**集成测试（testnode，主链 + 平行链）**

- Activate 后两侧 `GetNodeGroupAddrs` 与 BLS 公钥完全一致；
- 变更期间旧集合仍可正常 commit，激活后旧集合签名失效；
- 区块回滚后状态正确恢复；
- 未升级节点在分叉高度后的行为（应被升级检查拦住，而不是静默分叉）。

**回归测试**

- 现有 `nodegroup` / `node` / `supervision` / `asset transfer` / `rollup` 测试全绿；
- 标题/execer 精确绑定校验（PR #1326）在新流程下仍生效。

**混沌/边界**

- 延迟窗口内连续提案（第二个 Propose 必须失败）；
- 撤销后重新提案；
- 急停与验证人变更并发；
- 治理批准与 Cancel 在同一窗口内竞争。

## 16. 决策记录

| # | 决策 | 选择 |
|---|---|---|
| 1 | 本期产出 | 设计方案/可行性文档，暂不写共识代码 |
| 2 | 第一阶段范围 | 只改验证人来源（方向二）；主链验证状态留 Phase 2 |
| 3 | 验证人变更权限 | 平行链自身治理发起，主链按锁定资产分档处理 |
| 4 | 分档规则 | 锁定资产 > 10 万：必须治理批准，批准后锁 1 天且窗口内可取消；≤ 10 万且无批准：锁 7 天 |
| 5 | 实现方案 | 验证人集合状态机 + 延迟到点后显式激活（commit 读路径不变） |
| 6 | 首次建集合 | 保持现有自由注册（状态机只管"已有集合的更换"） |
| 7 | 急停 | 治理可控、**双向**切断跨链（含入金），恢复后挂起提现可执行 |
| 8 | 分档阈值 | 由**主链治理投票**决定（链上参数，初始 10 万 BTY），两段式 ParamPropose/ParamApply 绑定数值；批准后立即生效、只影响此后新提案 |

## 17. 假设与默认口径（如需变更请在实施前提出）

1. **治理主体** = autonomy 自治 board（`IsAutonomyApprovedItem`），超级管理员兜底；
2. **锁定资产口径** = `dapp.ExecAddress("user.p.<name>.paracross")` 的 coin 余额（token 暂不计入），在 `Propose` 时快照定档；
3. **同一时刻仅一个待生效提案**，新提案必须等前一个 Activate/Cancel 之后；
4. **提案 id** = `Propose` 交易的 hash；`Cancel` 权限 = 治理或提案发起地址；
5. **延迟参数**（1 天/7 天）走链配置；**分档阈值**由主链治理投票决定（第 6.5 节），代码默认值仅在链上无该 key 时生效；
6. **监督节点组** 与超级节点组同样纳入本状态机；
7. **版本号** 从 1 开始，已有节点组视为 version 1。

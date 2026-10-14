# Paracross 节点组「标题前缀」接管攻击分析

> 事件地址：`1KLQq9Y1CRbPMvqLwycG9L7rWTsMRVia5p`
> 入口交易：`0xfee96eb6d13f2d417671e8f09727fecf0dddd29d046e4b022f1d56de931be2f3`（高度 47805850）
> 分析时间：2026-10-13

> **披露策略**：本文档区分「根因/修复」（公开）与「操作细节/第三方归因」（暂不公开）。
> 第 2、5 节的攻击交易解读与根因表、第 6 节修复建议、第 9 节修复实现与验证保持公开；
> 第 3、7、8 节的可操作细节与第三方归因已收束，待主链 `ForkParaTitleExecMatch` 分叉激活后再补充。
> 未收束的完整版本保存在内部文档，不随仓库发布。

## 1. 结论（黑客是如何拿到平行链管理权限的）

黑客**没有盗取任何人的私钥**，而是利用了 paracross「超级节点组（node group）」权限模型中的三个缺陷，自己给自己造了一个节点组，并把它的管理权“套”到了真实平行链上：

1. **title 与 execer 不绑定、且主链不校验 title 合法性**
   `NodeGroupConfig`（`plugin/dapp/paracross/executor/superaccount.go:1108`）只检查 `validTitle`，而主链分支的 `validTitle` 仅要求 `len(title) > 0`（`executor/action.go:90`）；`Paracross.allowIsParaTx`（`executor/paracross.go:397`）只校验 execer 形状（`user.p.` 前缀 + 一个点 + 以 `paracross` 结尾）。
   → 黑客可以用**真实平行链的 execer**（`user.p.superx.paracross`）携带**任意伪造 title**（`user.p.super`）。

2. **节点组创建无权限校验 + 自由注册自审批**
   `op=1`（Apply）无任何身份校验，`checkNodeGroupExist`（`superaccount.go:715`）只拒绝“完全同名”的已存在节点组；伪造/前缀 title 没有节点组，申请直接通过。
   `op=2`（Approve）在 `ForkParaFreeRegister` 生效后走 `checkApproveOpNew`（`superaccount.go:988`），其中 `status.FromAddr == a.fromaddr` 即返回成功——**申请者可以自己批准自己**，无需超级管理员或 autonomy 投票。
   → `nodeGroupCreate`（`superaccount.go:1062`）把 `mavl-paracross-nodes-title-<title>` 写成黑客地址，黑客即成为该 title 的**唯一超级节点**。

3. **跨链交易过滤使用前缀匹配，使伪造 title 覆盖真实平行链**
   `types.IsSpecificParaExecName(title, exec)` = `strings.HasPrefix(exec, title)`（chain33 `types/config.go:613`），在 `executor/filtertxs.go:194/218` 被用于筛选某个 title 的跨链交易。
   → 节点组 title `user.p.super` 会前缀匹配真实链的 execer `user.p.superx.paracross`；`user.p.bscde` 匹配 `user.p.bscdex.paracross`；`user.p.proo` 匹配 `user.p.proof.paracross`；`user.p.` 则匹配**所有**平行链。

**权限 → 资产**：commit 只要求发送者在 `getNodesGroup(commit.Status.Title)` 中（`action.go:555`），且法定票数为 `3*mostSame > 2*nodes`（`action.go:143`）——1/1 必然通过。黑客用伪造 title 提交**伪造平行链区块**的 commit，同时之前已提交的 `crossAssetTransfer` 提现请求（`ExecPack` 挂起）被 commit 触发执行，把真实平行链在 paracross 合约中的锁定资产转入黑客地址在 paracross 下的子账户，最后用 `coins withdraw` 一次性提走。

## 2. 攻击交易解读（0xfee96eb6…）

```json
{
  "execer": "user.p.superx.paracross",
  "payload": {
    "nodeGroupConfig": {
      "title": "user.p.super",              // 伪造 title，是真实链 user.p.superx 的前缀
      "op": 1,                               // ParacrossNodeGroupApply
      "id": "",
      "addrs": "1KLQq9Y1CRbPMvqLwycG9L7rWTsMRVia5p",  // 黑客自己
      "coinsFrozen": "0"
    },
    "ty": 10003
  },
  "from": "1KLQq9Y1CRbPMvqLwycG9L7rWTsMRVia5p",
  "to": "115k3diaLTPnprZkDR844667iE814ueMKc",
  "hash": "0xfee96eb6d13f2d417671e8f09727fecf0dddd29d046e4b022f1d56de931be2f3"
}
```

- 回执 `ExecOk`（ty=2），日志 `LogParaNodeGroupConfig`（ty=660），`status=1`（Apply），
  `id="mavl-paracross-title-nodegroupid-user.p.super-0xfee96eb6…"`。
- **这只是第一步**：同一地址在 11 个区块后自己批准了它。

| 高度 | 交易 | 动作 | 说明 |
|------|------|------|------|
| 47805838 | `0x2aabd57f…` | nodeGroupConfig op=1 title=`user.p.superx` | 失败（rety=1）：真实链已有节点组，被 `checkNodeGroupExist` 拒绝 |
| **47805850** | **`0xfee96eb6…`（入口交易）** | nodeGroupConfig op=1 title=`user.p.super` | 成功：伪造前缀 title，无同名节点组 |
| 47805861 | `0x3524808093b503645f8b89c6bd0d8dd4095cb5cc0cbf82fb25551b726f2097af` | nodeGroupConfig op=2 title=`user.p.super`，`id=0xfee96eb6…` | 成功：**自审批**（`status.FromAddr == from`）→ 节点组建立 |

链上现状（查询 paracross）：

```
GetNodeGroupAddrs("user.p.super")  = 1KLQq9Y1CRbPMvqLwycG9L7rWTsMRVia5p   # 黑客唯一节点
GetNodeGroupAddrs("user.p.superx") = 1Ph4ULn2Ksmz3se5uEuuBnuq1GCtZNcMgy  # 真实链原有节点组
```

## 3. 权限如何变成资产（操作细节暂不公开）

> 本节原有内容（各条挂起提现的交易哈希/高度/金额/子账户，伪造 commit 的具体构造，以及执行日志片段）
> 属于可直接照做的操作细节，按安全披露策略**暂不随本仓库公开**，待主链 `ForkParaTitleExecMatch` 分叉激活后再补充。
> 完整版本保存在内部文档 `paracross-nodegroup-title-takeover-analysis.full.md`（不随仓库发布）。

可公开的机制层面描述：

1. 攻击者先用**真实平行链的 execer** 提交 `crossAssetTransfer` 提现，主链返回 `ExecPack`，即把提现挂起等待该平行链共识；
2. 再用**伪造前缀 title 的节点组**提交 commit（攻击者是该组唯一节点，1/1 满足 `3*mostSame > 2*nodes`），
   主链按 title 前缀把真实平行链的挂起提现匹配出来并执行，资金记入攻击者在 paracross 合约下的子账户；
3. 最后用 `coins withdraw`（`execName=paracross`）一次性提走。

资金影响量级：paracross 合约锁仓资产被抽走约 161.8 万 coins，合约余额从约 166.8 万 coins 降至约 4.9 万 coins。

## 4. 攻击时间线（关键节点）

| 高度 | 事件 |
|------|------|
| 47805455–47805518 | 尝试 `manage apply/modify/approve`、`autonomy propItem` 夺取管理权 → 失败 |
| 47805595 | 首次节点组尝试：`user.p.i8731x.` 成功；`user.p.superx.` 失败（已有节点组） |
| **47805850** | **入口交易**：Apply `user.p.super`（前缀 title） |
| 47805861 | Self-Approve，节点组建立 |
| 47805883–47806165 | 批量 `crossAssetTransfer` 挂起提现 + 伪造 commit 触发执行 |
| 47806035–47806072 | 大批 title 变体轰炸：`user.p.superx.` / `user.p.SuperX.` / `USER.P.I8847A` / `user.p.zzzdecouple` / `user.p.i8847a.` 等 |
| 47806177–47806180 | 创建 title=`user.p.`（匹配**所有**平行链）与 `u` 的节点组；commit 再次转移 2×5 coins |
| **47806194** | `coins withdraw` 提走 1,618,556.11765638 coins |
| 47806593–47806830 | 大量归集转账（约 2.6 万 coins），后续 `rollup commit` + `paracross rollupCrossTx` 继续注入跨链交易 |
| 47822766 | 转出 1,644,901.33365968 coins |

## 5. 根因归纳

| # | 位置 | 缺陷 |
|---|------|------|
| 1 | `executor/action.go:90` `validTitle` | 主链只判断 `len(title)>0`，title 与 execer/已注册平行链无任何绑定 |
| 2 | `executor/paracross.go:397` `allowIsParaTx` | 只校验 execer 形状，不校验 payload title 是否等于 execer 的 title |
| 3 | `executor/superaccount.go:1108/715/781` `NodeGroupConfig`/`checkNodeGroupExist`/`nodeGroupApply` | 节点组申请无权限校验，仅按“完全同名 key”判重 |
| 4 | `executor/superaccount.go:988` `checkApproveOpNew` | `ForkParaFreeRegister` 下允许**自审批**，申请人即成为节点组 |
| 5 | `executor/filtertxs.go:194/218` + chain33 `types/config.go:613` | 跨链交易按 `strings.HasPrefix(execer, title)` 前缀匹配，伪造前缀 title 即接管真实链 |
| 6 | `executor/action.go:555/143` | commit 仅按 payload title 查节点组，不校验 title 是否为已注册平行链；1/1 即满足 2/3 |
| 7 | `executor/rollupcrosstx.go:90` `rollupCrossTx` | 不校验调用者是否是该 `chainTitle` 的授权节点，配合 rollup 可继续注入跨链交易 |
| 8 | `types/type.go:77` | `ForkParaFreeRegister` 默认高度 0，自审批在链上默认生效 |

**一句话根因**：paracross 把「节点组 title」当作唯一权限键，却既没有把 title 与真实平行链/execer 做精确绑定，又允许任何人自建自批节点组；再叠加跨链交易的前缀匹配，使一个伪造的前缀 title 直接获得对真实平行链资产的管理与处置权。

## 6. 修复建议

### 6.1 精确绑定 title 与 execer（第一优先级）

在 `NodeGroupConfig` / `NodeConfig` / `SupervisionNodeConfig` / `CrossAssetTransfer` 等所有 paracross 主链入口，强制：

```go
// execer 必须是 user.p.<title>.paracross，且 payload.Title 必须等于该 title 的规范形式
execTitle, ok := types.GetParaExecTitleName(string(a.tx.Execer)) // "user.p.superx."
if !ok || (config.Title != execTitle && config.Title != strings.TrimSuffix(execTitle, ".")) {
    return nil, types.ErrInvalidParam
}
```

并拒绝不存在/未注册的平行链 title。

### 6.2 消除前缀匹配

`filtertxs.go`、`FilterParaTxsByTitle` 等处将 `strings.HasPrefix(execer, title)` 改为**精确相等**（`types.GetParaExecTitleName(execer) == title` 或 `execer == title+"paracross"`）。同时禁止创建“是已有 title 前缀”的节点组 title。

### 6.3 节点组创建/审批加权限

- 自由注册仅允许作用于**已注册且尚无节点组**的平行链 title；
- 仍需超级管理员 / autonomy 授权，或引入抵押 + 挑战期；
- `checkNodeGroupExist` 改成“任一前缀/同名冲突即拒绝”。

### 6.4 commit / 跨链校验

- `commit.Status.Title` 必须是已注册平行链，且 `blockHash` 必须能解析到真实平行链区块；
- 拒绝占位哈希（`0xaaaa…`、`0xbbbb…` 等）与 `txCounts/txResult` 自洽性不成立的 commit；
- `rollupCrossTx` 增加 `chainTitle` 授权校验（调用者必须是该 title 的节点或 rollup 验证人）；
- 跨链提现应校验提现额度不超过该 title 的实际锁仓、且来源交易确实出自该平行链。

### 6.5 分叉控制与应急处置

- 所有逻辑修改以新 `ForkParaXxx` 高度开关；
- 立即冻结/标记黑客地址与其归集地址，评估 paracross 合约剩余余额与回滚/追偿方案；
- 全网排查是否还存在其他前缀 title 节点组（`user.p.`、`u`、`user.p.zzz*`、大小写变体等），并清理。

> 代码侧的修复实现与验证见文末 **第 9 节**，链上排查/冻结/追偿仍属运维处置范围。

## 7. 附：主要链上对象（暂不公开）

> 本节原有内容（攻击者地址、paracross 合约地址，以及入口 Apply / 自审批 / 兑现 / 外转四笔交易的哈希与高度）
> 是定位整个攻击的索引清单，同样按安全披露策略**暂不随本仓库公开**，见内部完整版文档。

## 8. 资金归集路径（归因与推断暂不公开）

> 本节原有内容（65 个沉睡地址的批量扫款清单、归集/中转地址、进入 EVM 后的分层路径，
> 以及对某个服务方充值地址与热钱包的**第三方归因**）按安全披露策略**暂不随本仓库公开**。
> 其中部分结论本身也属推断、未经证实；即便在内部文档中也需按"未证实"对待。

可公开的结论层面描述：

- 漏洞所得与另一条**独立**的资金线（一批多年未动的地址被脚本化全额扫款）汇入同一归集地址后单跳转出；
- 归集地址随后经 EVM 转入某个服务方的充值/热钱包体系，即"通过服务方出货"；
- 该第二条资金线需要持有那批地址的私钥，与 paracross 漏洞本身无关，属独立的资金归集行为。

## 9. 修复实现（`ForkParaTitleExecMatch`）

代码侧修复已完成，全部新校验由新分叉 `paracross.ForkParaTitleExecMatch` 开启，分叉前行为保持不变。

### 9.1 核心思路：按「状态 key 精确」而非「归一化后相等」

把「节点组 title」的权限严格限定在**该 title 自身的平行链执行器**上：

```
title 必须为规范形式 user.p.<name>.  且  execer 必须精确等于 title + "paracross"
```

关键点是**不做隐式归一化**。`user.p.superx` 与 `user.p.superx.` 在节点组/commit 中是两个**不同的状态 key**
（`mavl-paracross-nodes-title-<title>`），但归一化后会指向同一个 execer；一旦在执行入口容忍归一化，
攻击者就可以给"真实链 title 去掉结尾 `.`"（如 `user.p.superx`）建立自己的节点组，再借 rollup 执行真实链
（`user.p.superx.`）的跨链交易 —— 这与入口交易的前缀变体是同一类问题。因此：

- 规范化只作为**输入校验**使用（要求 title 必须已带结尾 `.`），不参与权限判定；权限判定一律按原样字符串比较；
- `user.p.super`（缺少 `.`）、`user.p.superx`（真实 title 去掉 `.`）、`user.p.`、`u`、`USER.P.X.`、`user.p.a.b.` 全部拒绝；
- 前缀双方（`user.p.super` 与 `user.p.superx.`）在字符串上不再可能互相匹配：规范 title 一定以 `.` 结尾，
  而真实平行链 execer 在 title 之后紧跟链名，`HasPrefix(user.p.superx.paracross, user.p.super.) == false`；
- 跨链交易**执行入口**（`execCrossTxs`，commit 与 rollup 两条路径共用）再按同一规则兜底校验，
  即使伪造 title 的节点组在分叉前已经存在于链上，分叉后也无法执行任何其他平行链的跨链交易。

### 9.2 代码变更

| 位置 | 变更 |
|------|------|
| `plugin/dapp/paracross/types/type.go` | 新增分叉 `ForkParaTitleExecMatch`，`InitFork` 中 `RegisterDappFork(ParaX, ForkParaTitleExecMatch, 0)` |
| `plugin/dapp/paracross/types/title.go`（新增） | `NormalizeParaTitle` / `IsValidParaTitle` / `IsExactParaTitleExec`；`IsValidParaTitle` 对 `user.p.` 做长度保护，避免切片越界（该 title 原本会导致节点 panic） |
| `plugin/dapp/paracross/executor/action.go` | `checkParaTitleExecMatch`（入口绑定）、`checkCommitParaTitle`（commit title）、`checkCrossTxTitleMatch`（跨链执行兜底）、`getCrossTxs`/`getCrossTxsByRst` 精确过滤 |
| `plugin/dapp/paracross/executor/filtertxs.go` | `filterCrossTxsByExactTitle`、`FilterParaMainCrossTxHashesFork`；`crossTxGroupProc`/`isSpecificParaTxOfTitle` 支持精确/前缀两种模式 |
| `plugin/dapp/paracross/executor/exec_local.go` | 平行链侧 `setMinerTxResult` 分叉后使用 `FilterParaMainCrossTxHashesFork` |
| `plugin/dapp/paracross/executor/superaccount.go` | `NodeGroupConfig` / `NodeConfig` 增加 title↔execer 校验 |
| `plugin/dapp/paracross/executor/supervisionaccount.go` | `SupervisionNodeConfig` 增加 title↔execer 校验 |
| `plugin/dapp/paracross/executor/rollupcrosstx.go` | `rollupCrossTx` 分叉后要求 `chainTitle` 为规范 title；`getRollupCrossTxs` 的索引列表同样按精确 title 过滤 |
| `plugin/dapp/paracross/executor/query.go` | 分叉后查询接口不再返回非规范 title：`Query_ListTitles`、按状态遍历节点组 `Query_ListNodeGroupStatus` 过滤掉 `IsValidParaTitle` 不通过的记录（仅记日志保留可发现性）；按 title 精确查询保留，便于排查/清理历史数据 |

对各根因的覆盖：

| 根因 | 处置 |
|------|------|
| 1 `validTitle` 只查 `len>0` | `checkParaTitleExecMatch`：title 必须规范且与 execer 精确对应 |
| 2 `allowIsParaTx` 只校验 execer 形状 | 入口改为校验 title 与 execer 的**一致性**（形状校验之后仍保留） |
| 3 节点组申请无权限、仅同名判重 | 申请/审批/修改/退出四类 op 统一走 `checkParaTitleExecMatch`；非规范 title 无法建立节点组 |
| 4 `ForkParaFreeRegister` 自审批 | 保留自由注册语义；自审批后节点组只能作用于自己的 title（权限已被限定），不再构成越权 |
| 5 跨链交易前缀匹配 | 主链 `getCrossTxs*`、平行链 `FilterParaMainCrossTxHashesFork`、rollup 索引列表均按精确 title 匹配；执行入口再加 `checkCrossTxTitleMatch` 兜底（覆盖 commit 与 rollup） |
| 6 commit 仅按 title 查节点组 | `checkCommitParaTitle` 拒绝非规范 title；跨链交易本身的归属由精确匹配保证 |
| 7 `rollupCrossTx` 无调用者校验 | 分叉后 rollup `chainTitle` 必须为规范 title；轮次数据仍由 `CommitRound`+`CrossTxCheckHash` 约束；执行入口兜底校验使非规范 title 无法执行其他链的跨链交易 |
| 8 分叉默认高度 | 新分叉在所有配置文件中显式配置（见下） |

### 9.3 分叉与配置

- `chain33.fork.toml`：`ForkParaTitleExecMatch=48100000`。
  取值依据：主链当前目标出块时间 **5 秒**（`chain33.toml` 的 `[mver.consensus.ticket.ForkChainParamV2] targetTimePerBlock=5`，
  该参数在 `chain33.fork.toml` 的 `ForkChainParamV2=1692674` 生效，当前高度已远超该分叉）；
  **2 周 = 14×86400/5 = 241920 个区块**。以 2026-10-14 主链高度约 4783 万为锚（漏洞入口交易高度 47805850，2026-10-13），
  最小分叉高度 ≈ 47807800，取整并留约 1.5 天余量后定为 **48100000**。
  公式：`分叉高度 = 公告日主链高度 + 241920`（若上线推迟，按当日实际高度重算；务必大于部署时的主链高度，
  否则未升级节点会在分叉高度处与升级节点行为不一致）。
- 平行链/测试/CI 配置同键值：`chain33.para.toml`、`chain33.para.toml.readme`、`build/ci/{cross2eth,bridgevmxgo,rollup,rgbx}/chain33.para.toml`、
  `plugin/dapp/ticket/executor/testdata/chain33.cfg.toml`、`plugin/dapp/ticket/rpc/rpc_test.go` 置 `0`；
  `plugin/dapp/evm/cmd/ci2/chain33.proxyminer.toml`、`build/ci/evm2/chain33.proxyminer.toml` 置 `48100000`。
- ⚠️ chain33 的 `initForkConfig` 会对「代码已注册但配置文件中缺失」的 fork 直接 panic，
  因此**所有**含 `[fork.sub.paracross]` 的配置文件都必须补上该键，否则节点无法启动。
- ⚠️ 激活前必须处置/排查两类存量数据：
  1. 链上已存在的**非规范节点组**（`user.p.super`、`user.p.superx`、`user.p.`、`u` 等）。分叉后它们不能再执行任何跨链交易，
     但建议清理或标记，避免后续 rollup 等路径反复报错。分叉后这类 title 不再出现在 `Query_ListTitles`/`Query_ListNodeGroupStatus`
     的结果里（会打 Error 日志），按 title 精确查询仍可读到，便于定位与清理；
  2. 任何使用**非规范 `chainTitle`**（缺少结尾 `.`）的 rollup 链。分叉后其 `RollupCrossTx` 会被拒绝，
     需在分叉高度前把 rollup 配置的 title 改为规范形式（`user.p.<name>.`，与平行链配置一致）。

### 9.4 已知残留与未采纳项

- **未注册平行链仍可被自由注册接管（已接受的剩余风险）**：`ForkParaFreeRegister` 之后，
  任意规范化 title 只要尚无节点组，任何人都能申请并自审批，随后该链自身的跨链交易就会经正常 commit 路径执行。
  对"尚无节点组的真实平行链"，其影响与原漏洞同级。主链不存在平行链注册表，无法在合约内判定 title 真伪，
  因此不在本次改动中收紧（会改变既定共识语义）。**运维缓解**：把 `[exec.sub.paracross] nodeGroupFrozenCoins`
  设为非 0（提高接管成本）、监控并告警"已有平行链 title 出现新节点组/自审批"，必要时冻结可疑地址。
- **`loopCommitTxDone` 会吞掉执行期错误**：该路径在记录 `CommitDone` 后才调用跨链执行，出错仅记日志
  （历史行为）。分叉后兜底校验失败时不会有任何资产变动，但状态可能被静默推进；跨链交易的**执行**不会发生。

- **「拒绝占位哈希 / txCounts 自洽」**：主链无法验证平行链区块的真实性（`BlockHash/StateHash` 只有平行链自己知道）。
  精确匹配后，伪造 title 的 commit 只能触发自己链的跨链交易，占位哈希不再产生越权，保留为监控项。
- **「自由注册加权限/抵押挑战期」**：属于权限模型重构，会改变既定共识语义，本次不引入。
- **「rollupCrossTx 调用者白名单」**：需同时改动 rollup 共识与执行器，风险较高；现有 `CommitRound`+`CrossTxCheckHash`
  已要求轮次数据来自该 title 的验证人集合（验证人集合即该 title 的节点组），再叠加"chainTitle 必须规范 + 执行入口兜底"，
  非规范 title 已无法越权执行。
- **主链/平行链分叉高度映射**：本分叉与既有的 `ForkParaAssetTransferRbk` / `ForkParaSupervision` /
  `ForkParaFreeRegister` / `ForkParaCheckTx` 一样使用 `cfg.IsDappFork(<主链高度>, …)`，因此平行链配置需与主链填同一高度
  （这是这几个分叉一贯的部署约定）。未改用 `IsParaForkHeight`/`GetDappForkHeight`：该函数对未映射的 key 在平行链上会
  直接 panic，需要新增 `mainForkParaTitleExecMatch` 配置项与映射分支，收益不足而部署风险更高。

### 9.5 测试与验证

- `plugin/dapp/paracross/types/title_test.go`：title 规范性/精确匹配的边界用例（含 `user.p.`、大小写、多级名称）。
- `plugin/dapp/paracross/executor/title_takeover_test.go`：复现漏洞入口交易
  （`execer=user.p.superx.paracross` + `title=user.p.super`）被拒、非规范 title 被拒、空名称 title 被拒、
  自审批被拒、伪造前缀 title 的 commit 被拒、跨链交易执行入口校验，以及合法自由注册不受影响。
- 查询侧：`Query_ListTitles` / `Query_ListNodeGroupStatus` 分叉后过滤非规范 title、分叉前保持历史行为，均有对应用例。
- 覆盖 rollup 变体（安全评审发现）：`execer=user.p.superx.paracross` 的跨链交易在
  `title=user.p.superx`（真实 title 去掉结尾 `.`）或 `title=user.p.super` 下经 `execCrossTxs` 执行时必须失败。
- 变异验证（两轮，均已实测）：① 把 `ForkParaTitleExecMatch` 临时注册为 `types.MaxHeight`（关闭分叉）后 11 个用例失败；
  ② 把 `IsExactParaTitleExec` 改回"归一化后比较"（即评审指出的洞）后，`TestCheckCrossTxTitleMatch`、
  `TestExecCrossTxsTitleGuard`、`TestFilterCrossTxsByExactTitle`、`TestFilterParaMainCrossTxHashesExact`、
  `TestIsExactParaTitleExec` 全部失败。说明用例覆盖真实漏洞路径，而非恒真断言。
- 回归：`go test ./plugin/dapp/paracross/...`、`./plugin/consensus/{para,rollup}/...`、`./plugin/dapp/{ticket,rollup}/...` 全部通过。

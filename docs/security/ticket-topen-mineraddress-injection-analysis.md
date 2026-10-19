# Ticket 票据归属缺陷（审计项 H-2）与 Genesis 重入分析

> 代码版本：`plugin` HEAD `92b68766e`，依赖 `chain33 v1.72.0`
> 分析时间：2026-10-15（含独立复核代理的交叉验证与修正）
> 可执行证据：
> ① 执行器级 PoC：`plugin/dapp/ticket/executor/ticket_miner_inject_poc_test.go`（6 个用例，全部通过；其中演示“注入”的用例显式把 `ForkTicketMinerAuth` 配到未来高度，固定修复前的旧行为）
> ② 真实测试节点端到端 PoC：`plugin/dapp/ticket/testnode/genesis_reentry_attack_test.go`（mempool → 出块 → 执行 → 票共识挖票 → 奖励结算，通过）
> 复现：`go test -run 'Test_TicketOpen_Injects|Test_InjectedTicket|Test_TicketGenesisReRun|Test_TicketOpen_CountExceeds' -v ./plugin/dapp/ticket/executor/`
> 　　　`go test -run 'TestGenesisReentryAttackOnTicketNode' -v ./plugin/dapp/ticket/testnode/`
>
> 修复状态（本 PR，见 §8）：
> 1. **执行器层**：新增 dapp 分叉 `ticket.ForkTicketMinerAuth`，**默认关闭**（`InitFork`
>    注册为 `types.MaxHeight`，必须由链配置显式指定激活高度），拒绝 `MinerAddress` 既非
>    签名者、也非其绑定矿工的开票；主链配置为 `48100000`。
> 2. **矿工侧（无需分叉）**：选票入口按 `wallet.sub.ticket.minerwhitelist`（ReturnAddress
>    白名单）过滤，只挖自挖票与名单内的票。这是矿工本地的“批准名单”：执行器校验挡不住
>    自签 `Tbind` 的挂票（`Tbind` 只由 ReturnAddress 签名，见 §8.3），最终是否消耗算力
>    由矿工节点自己决定。
> 3. **§3 的 `Exec_Genesis` 重入由独立提交修复**（`Exec_Genesis` 增加 `height != 0` 校验），
>    与本 PR 互不依赖。

## 1. 结论

| 问题 | 结论 |
| --- | --- |
| “向任意矿工注入票据”能否实现 | **可以**。攻击者无需受害矿工私钥或任何授权，只需以自己地址作 `ReturnAddress` 签一笔 `Topen`（或 `Genesis`） |
| 报告“盗取奖励”的危害 | **成立，且分叉生效后同样成立**——但分叉后的路径不是 `Topen`（会被 `pubHash` 挡住），而是 **`Exec_Genesis` 缺少 `height != 0` 保护导致的重入**（§3，严重/高危） |
| `Topen` 注入在 `ForkTicketId` 生效后的单独危害 | **中危**：票据集合/本地索引污染、替受害矿工出废块、链上归因伪造、清理受限；不盗币、不实质掉速 |
| `Count > len(PubHashes)` panic | **低–中危**（修正）：panic 被 `DriverBase.Exec` 的 recover 兜住 → `ErrActionNotSupport`，自挖块会丢弃该笔交易，节点继续出块，**不会停摆** |
| 修复 | **`MinerAddress` 归属校验与 PubHashes 越界见 §8（本 PR）**；`Exec_Genesis` 重入见独立提交 |

一句话：**H-2 的根因（不校验 `MinerAddress` 归属）真实存在；它所声称的盗币危害在现网（`ForkTicketId` 已生效）依然成立，但要借道另一个缺陷——`Ticket.Exec_Genesis` 没有像 coins/coinsx/token 那样拒绝高度非 0 的重入。**

## 2. 根因 A：`TicketOpen` 不校验 `MinerAddress` 归属

`plugin/dapp/ticket/executor/ticketdb.go:243-251`：

```go
//addr from
if action.fromaddr != topen.ReturnAddress {
    mineraddr := action.getBind(topen.ReturnAddress)
    if mineraddr != action.fromaddr {
        return nil, ty.ErrMinerNotPermit
    }
    if topen.MinerAddress != mineraddr {
        return nil, ty.ErrMinerAddr
    }
}
```

归属校验整体被放进 `fromaddr != ReturnAddress` 的分支。调用方只要把自己的地址填进 `ReturnAddress`（最常见的自挖写法），`MinerAddress` 就完全不受约束——不必等于 `fromaddr`，也不必等于 `ReturnAddress` 绑定的矿工。`Exec_Topen`（`exec.go:26-35`）也没有 `address.CheckAddress`（对比 `TicketBind` 有：`ticketdb.go:218-222`）。冻结仍发生在调用方自己的 `ReturnAddress` 上（`ticketdb.go:266`），链上没有任何威慑。

放大危害的是索引与授权都绑定 `MinerAddress`：

- 回执地址标签 `ReceiptTicket.Addr = t.MinerAddress`（`ticketdb.go:92`）→ 本地索引 `LODB-ticket-tl:<FormatAddrKey(MinerAddress)>:<status>:<ticketId>`（`exec_local.go:16-23`、`ticket.go:116-137`）；
- 钱包按账户地址（= `MinerAddress`）取票（`wallet/exec.go:30-34`、`wallet/ticket.go:475-510`）；
- 共识 `searchTargetTicket` 要求 `privmap[ticket.MinerAddress]` 存在（`consensus/ticket/ticket.go:582`，`getPrivMap` 见 `:250-274`）。

因此“把 `MinerAddress` 写成受害矿工”= 让受害节点的钱包与共识把票当成自己的票。

## 3. 根因 B（关键）：`Exec_Genesis` 没有高度保护，使分叉后仍可造出“免检”票据

`plugin/dapp/ticket/executor/exec.go:14-23`：

```go
func (t *Ticket) Exec_Genesis(payload *ty.TicketGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
    if payload.Count <= 0 {
        return nil, ty.ErrTicketCount
    }
    payload.MinerAddress = common.FmtEthAddressWithFork(...)
    payload.ReturnAddress = common.FmtEthAddressWithFork(...)
    actiondb := NewAction(t, tx)
    return actiondb.GenesisInit(payload)   // 没有任何 height/fork 检查
}
```

对照同类执行器，都有重入保护并返回 `types.ErrReRunGenesis`：

- chain33 `system/dapp/coins/executor/exec.go:58-71`（`if c.GetHeight() == 0 { ... } return nil, types.ErrReRunGenesis`）
- `plugin/dapp/coinsx/executor/exec.go:71-79`
- `plugin/dapp/token/executor/transwithdraw.go:39-46`

chain33 执行器/驱动层也没有针对 Genesis action 的全局防护（`types`/`executor`/`system/dapp` 中不存在 `GenesisAction`/`IsGenesisTx` 之类的校验）。票据 `CheckTx` 在 mempool 只拒绝 `TicketActionMiner`（`executor/ticket.go:182-195`），所以一笔 `TicketAction{Ty=11(Genesis)}` 的普通用户交易可以正常进块并执行。

`GenesisInit`（`ticketdb.go:147-171`）生成的是**创世格式票据**：

```go
id := prefix + fmt.Sprintf("%010d", i)                       // 3 段 miner:txhash:index，无 pubHash
t := NewDB(..., genesis.MinerAddress, genesis.ReturnAddress, ..., /*isGenesis*/ true)
```

这带来三重“免检”：

1. **免 `pubHash` 校验**：3 段 `ticketId` 下 `genPrivHash` 返回空（`consensus/ticket/ticket.go:654-670`），`TicketMiner` 的校验条件 `len(miner.PrivHash) != 0` 不成立而被整段跳过（`ticketdb.go:328-337`）；
2. **免成熟期**：`TicketMiner` 只对 `!ticket.IsGenesis` 校验 `TicketFrozenTime`（`ticketdb.go:318-322`），创世票立刻可挖；
3. **免提现等待**：`TicketClose` 的时间校验同样只对 `!ticket.IsGenesis` 生效（`ticketdb.go:396-408`），攻击者可当场 `Tclose` 收回票面价 + 奖励（`ticketdb.go:424-425`）。

于是分叉生效后仍存在完整的盗币链：

```
攻击者签 Topen/Genesis 交易（MinerAddress=受害矿工, ReturnAddress=攻击者, Count=N）
  → 受害节点按 MinerAddress 把 N 张票纳入 ticketsMap（§2）
  → 受害节点正常出块，抽中这些票时 Exec_Miner 直接通过（免检 1、2）
  → ExecDepositFrozen(ReturnAddress, MinerValue) 增发奖励到攻击者冻结余额（ticketdb.go:340,348）
  → 攻击者立刻 Tclose（免检 3）拿走票面价 + 奖励，本金零锁仓、逐块滚动复用
```

攻击者成本仅手续费；受害矿工损失的是自己算力本应获得的出块奖励。奖励流向说明：`ExecDepositFrozen` 的语义是**增发新币并冻结给 `addr`**（chain33 `account/execaccount.go:232-268`，增发见 `account/account.go:160-175`），`addr` 是票据 `ReturnAddress`——**奖励从来不给 `MinerAddress`**。

## 4. 危害分析

### 4.1 `ForkTicketId` 已生效：借道 Genesis 重入盗取奖励（严重）

**执行器级 PoC**（修复前的 `Test_TicketGenesisReRun_PostFork_RewardGoesToAttacker`，随 `Exec_Genesis` 修复已改为回归断言 `Test_TicketGenesisReRunRejected_PostFork`）在 `height=1,700,000`（> `ForkTicketId=1,600,000`，断言分叉已生效）当时实测：

```
post-fork(height=1700000) genesis re-run: victim 1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR
  mined, reward 500000000 credited to attacker 1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4, closed immediately
```

**真实测试节点端到端 PoC**（`plugin/dapp/ticket/testnode/genesis_reentry_attack_test.go`，默认配置下 `ForkTicketId=0` 创世即生效，全程走钱包/mempool → 出块 → 执行器 → 票共识选票挖票 → 奖励结算）：

```
node height=3 ForkTicketId=true ticketPrice=1000000000000 coinReward=1800000000
attacker=1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF victim=12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
harvested: attacker ticket balance=1000003600000000 frozen=0 -> forged count=1000
victim node height=77 minedForged=1 attacker delta=1800000000 (coinReward=1800000000) minerValue=[1800000000]
attacker closed a mined forged genesis ticket immediately at height=81
--- PASS: TestGenesisReentryAttackOnTicketNode (9.02s)
```

要点：一笔 `Genesis` 交易在高度 >0 被执行（`ExecOk`），造出 1000 张 `IsGenesis=true`、3 段 `ticketId` 的票据挂在受害矿工 `12qyocay` 名下；受害节点在高度 77 把它挖出（`minerValue=1800000000`）；攻击者地址的余额增量恰好等于 1 × `coinReward`（该地址已无其它收入来源，故可精确归因）；随后攻击者在高度 81 立即 `Tclose` 成功（创世票跳过 `TicketWithdrawTime`）。

### 4.2 `ForkTicketId` 未生效（或分叉前预埋 3 段票）：`Topen` 注入直接盗取奖励（高）

分叉未生效时 `ticketId` 只有 3 段（`ticketdb.go:254-262`），`pubHash` 校验整段跳过（`ticketdb.go:328-330`；`genPrivHash` 对 3 段 id 返回空）。攻击者签 `MinerAddress=受害矿工` 的 `Topen` 即可；票据无过期机制（`searchTargetTicket` 只校验成熟时间，`consensus/ticket/ticket.go:577-580`），**分叉前预埋的票在分叉后依然被放行**（这就是 §3 的第 1 条“免检”在时间维度上的表现）。PoC 4 实测奖励 5 币记到攻击者地址。

现网配置：`chain33.para.toml:358`、chain33 默认注册（`plugin/dapp/ticket/types/ticket.go:81`）均为 `ForkTicketId=0`（创世即生效），主链配置无该项亦为 0；`chain33.fork.toml:212` 为 `1062000`、执行器测试配置为 `1600000`——**凡尚未到达该高度的链都处于 4.1/4.2 的直接暴露状态**。

### 4.3 `ForkTicketId` 已生效且不使用 Genesis：污染 / 废块 / 归因（中）

只看 `Topen` 注入时，`pubHash` 由开票人私钥派生（`wallet/ticket.go:662-669`），受害矿工用自己的私钥算出的 `privHash` 必然对不上 → `ErrCheckPubHash`（`ticketdb.go:331-337`），PoC 2 已验证。剩余危害：

1. **票据集合/本地索引污染**：注入票进入 `ticketsMap` 与 `LODB-ticket-tl:<矿工>:1`，每次挖矿尝试遍历并哈希全部票（O(N) 次 sha256）；共识侧 `flushTicket`（默认 3600s，`consensus/ticket/ticket.go:42,99-111,234-248`）与钱包事件会整表拉取重建；
2. **替受害矿工出废块**：注入票先满足难度时，节点组装并提交区块，coinbase 执行失败 → 自挖路径 `PreExecBlock` 丢弃失败的 coinbase（chain33 `util/util.go:381-396`，`blockchain/process.go:296` 中 `errReturn=false`），随后 `CheckBlock` 因 `Txs[0]` 不再是矿工交易而报 `ErrCoinBaseTxType`（`consensus/ticket/ticket.go:288-290,343`）→ 区块被拒 → `WriteBlock` 失败 → `CreateBlock` 用同一个 `block` 无限重试（`consensus/ticket/ticket.go:614-647,808-819`）。多矿工链会被别人出的块打断恢复；单矿工链若注入票长期占据首个合格位则可能停滞；
3. **链上归因伪造**：回执与 `TicketList(受害矿工)` 显示受害矿工名下有自己没开过的票，可用于甩锅/合规误判/监控误报；
4. **清理受限**：自动 close 只处理 `status=2`（`wallet/ticket.go:520-549`），`status=1` 的注入票不被清理；只能等 `TicketWithdrawTime`（主网 172800s ≈ 2 天，`chain33.toml:196`）后由运营方手动 force-close（`wallet/exec.go:13-27`、`wallet/ticket.go:359-397,436-450`），期间攻击者可滚动续投。

**定量说明（为什么不按“停摆”定级）**：设单票单次尝试满足难度的概率为 `q`、受害自有票数 `M`、注入票数 `N`，则 `P(单次尝试成功) ≈ M(1-e^{-Nq})/N`；`Nq ≪ 1` 时与未注入相同（注入票只是“只输不赢”的抽签券，且攻击者无法定向提高难度——难度哈希含受害矿工的 `privHash`）。显著掉速需要 `N ≳ 1/q`（量级≈全网票池），资本门槛极高。注意这与 §4.1 不同：Genesis 票一旦被抽中是**成功出块并付钱给攻击者**，所以是小额即可持续收割，而非掉速。

### 4.4 与报告定级的对照

报告的“高”应当保留，但需要把攻击链写清楚：**根因是 H-2（`MinerAddress` 无归属校验），在 `ForkTicketId` 未生效时用 `Topen` 直接利用；在 `ForkTicketId` 已生效时用 `Exec_Genesis` 重入（缺 height 保护）+ 同一个 `MinerAddress` 无校验来利用。** 若只按“分叉后 `Topen` 注入能否偷钱”评估，会得出“不能”的错误结论。

## 5. 相邻缺陷：`Count > len(PubHashes)` panic（修正为低–中危）

`ticketdb.go:257-262` 只拒绝 `len(topen.PubHashes) == 0`，随后按 `Count` 下标访问 `topen.PubHashes[i]`；`Count=1000, PubHashes=[1 项]` → `runtime error: index out of range [1] with length 1`（修复前由 PoC 5 验证，现已改为回归断言）。

链上后果（PoC 5 + 代码 + 独立复核实测）：

- mempool 不拦：票据 `CheckTx` 只拒绝 `TicketActionMiner`（`executor/ticket.go:182-195`），chain33 mempool 检查只调用 `CheckTx`（`system/mempool/check.go:189-209`），手续费/基本检查通过；
- panic 被 **`DriverBase.Exec` 的 recover** 兜住并返回 `types.ErrActionNotSupport`（chain33 `system/dapp/driver.go:296-307`）——不是 `ErrExecPanic`（`executor.Exec` 的 recover 在 `execenv.go:251-261`，层级更高，看不到这个 panic）。修复前的 PoC 5 断言的正是 `ErrActionNotSupport`；该修复不设分叉保护，修复后 `Test_TicketOpen_CountExceedsPubHashesRejected` 断言返回 `ErrOpenTicketPubHash` 且不再 panic；
- 自挖路径 `PreExecBlock` 会把执行失败（`ExecErr`）的交易从区块中剔除并重算状态根（chain33 `util/util.go:381-396`），`WriteBlock` 的 diff 逻辑再把它从 mempool 清除（`system/consensus/base.go:479-490`）；因此节点会继续正常出块（复核实测高度 435→600 约 20s），**不会停摆**；非法区块也不会被广播（广播发生在 `connectBlock` 成功之后）；
- 残余风险：这属于健壮性缺陷——目前唯一阻止节点 panic 的就是驱动层的 recover，一旦该 recover 被绕过或调整（例如执行器直接调用 action），就会变成进程级崩溃。

修复：`if len(topen.PubHashes) < int(topen.Count) { return nil, ty.ErrOpenTicketPubHash }`。

## 6. 修复建议（按优先级）

1. **Critical — `Exec_Genesis` 补高度保护**（`plugin/dapp/ticket/executor/exec.go:14`），与 coins/coinsx/token 对齐：

```go
func (t *Ticket) Exec_Genesis(payload *ty.TicketGenesis, tx *types.Transaction, index int) (*types.Receipt, error) {
    if t.GetHeight() != 0 {
        return nil, types.ErrReRunGenesis
    }
    ...
}
```

2. **High — `Exec_Topen` 补归属与地址校验**（保留自挖与冷地址绑定两条合法路径）：

```go
if err := address.CheckAddress(topen.MinerAddress, action.height); err != nil {
    return nil, err
}
bindMiner := action.getBind(topen.ReturnAddress)              // 可能为空
if topen.MinerAddress != action.fromaddr && topen.MinerAddress != bindMiner {
    return nil, ty.ErrMinerAddr
}
if action.fromaddr != topen.ReturnAddress && bindMiner != action.fromaddr {
    return nil, ty.ErrMinerNotPermit
}
```

兼容性：`buyTicketOne`（`wallet/ticket.go:717`，miner == return == 自己）与 `buyMinerAddrTicketOne`（`:819`，miner = 自己且已与冷地址绑定）都满足新条件；被挡掉的只有“用任意 `MinerAddress` 开票”。

3. **Medium — 消除 panic**：`len(PubHashes) < Count` 直接返回 `ErrOpenTicketPubHash`。

4. **纵深防御**：
   - 分叉生效后，`TicketMiner` 对 3 段（旧格式/创世格式）`ticketId` 应直接拒绝，而不是跳过 `pubHash` 校验——这能同时封堵“分叉前预埋票”和“Genesis 重入”两类绕过；
   - `searchTargetTicket` 跳过 `ReturnAddress` 未与本节点绑定的票据，或至少告警；
   - 票据增加最大未挖高度/过期机制，避免陈票长期占用票池与索引；
   - `Topen` 的 `Count` 加合理上限（当前仅 `> 0`）。

## 7. 验证范围与残余不确定性

**已由可执行 PoC 验证**：

执行器级（`plugin/dapp/ticket/executor/ticket_miner_inject_poc_test.go`，6/6 通过；注入类用例显式固定 `ForkTicketMinerAuth` 未生效的配置）：

| 用例 | 结论 |
| --- | --- |
| `Test_TicketOpen_InjectsTicketsToArbitraryMiner` | 第三方可把票据挂到任意矿工名下；`LODB-ticket-tl:<受害矿工>:1` 生成；官方 `TicketList(受害矿工)` 返回这些票 |
| `Test_InjectedTicket_NotMinableByVictim_PostFork` | 分叉后受害矿工挖 `Topen` 注入票 → `ErrCheckPubHash`（不盗币） |
| `Test_TicketGenesisReRunRejected_PostFork` / `Test_TicketGenesisAllowedAtHeightZero` | 修复后回归：高度 1.7M 的 `Exec_Genesis` 被框架层以 `ErrReRunGenesis` 拒绝；高度 0 的正常创世路径不受影响 |
| `Test_InjectedTicket_MinedByVictim_PreFork_RewardGoesToAttacker` | 分叉前 `Topen` 注入可被受害矿工挖出，奖励记到攻击者地址 |
| `Test_TicketOpen_CountExceedsPubHashesRejected` | 修复后回归：`Count > len(PubHashes)` 返回 `ErrOpenTicketPubHash`，不再 panic（该修复不设分叉保护） |

真实测试节点端到端（`plugin/dapp/ticket/testnode/genesis_reentry_attack_test.go`，通过）：默认配置（`ForkTicketId=0`）下发送一笔 `Genesis` 交易 → 上链执行成功 → 1000 张 `IsGenesis` 3 段票据挂在受害矿工名下 → 受害节点在高度 77 挖出 → 攻击者地址余额增量 = 1 × `coinReward` → 攻击者在高度 81 立即关闭提款。

**H-2 修复回归**（`plugin/dapp/ticket/executor/ticket_miner_auth_test.go`，本 PR，8/8 通过）：第三方开票被 `ErrMinerAddr` 拒绝且不动用票面资金；自挖与 `Tbind` 绑定代挖路径不受影响；非法 `MinerAddress` 被 `CheckAddress` 拒绝；`Count > len(PubHashes)` 返回错误而非 panic；分叉高度之前保持旧行为；分叉默认关闭；并固定已知限制“自签 Tbind 仍可挂票”（由矿工侧白名单兜底）。

**矿工侧白名单回归**（`plugin/dapp/ticket/wallet/ticket_test.go`，`TestFilterTrustedTickets`）：默认 `minerwhitelist=["*"]` 时不过滤（保持原有挖矿行为）；显式白名单时只保留自挖票与白名单 ReturnAddress 的票，注入票被过滤并打 `skipUntrustedTicket` 告警。详见 §8.2。

**由代码推导 + 独立复核实测、但未纳入本仓库测试**：灌入票据后节点重试/停滞的时序（多矿工 vs 单矿工）；panic 交易被 `PreExecBlock` 丢弃后节点继续出块（复核在临时节点上实测 435→600）。**未做大规模仿真**：§4.3 的抽签模型为解析推导。

本次分析未改动业务代码；修复代码与回归用例见 §8，另附 PoC 测试文件与本文档。

## 8. 修复实现（H-2，本 PR）

修复分两层：执行器层用**默认关闭**的分叉收紧开票校验；矿工侧用**无需分叉**的 ReturnAddress
白名单过滤选票。前者阻止“裸挂”上链，后者让矿工不再替别人出块；两层互相补充。

### 8.1 执行器层：新增 dapp 分叉 `ticket.ForkTicketMinerAuth`（默认关闭）

`InitFork` 把该分叉注册为 **`types.MaxHeight`（关闭）**，只有链配置
`[fork.sub.ticket] ForkTicketMinerAuth` 显式指定激活高度后才启用。这一点很关键：链配置里
没有 `[fork]` 段时（例如部署仓的 `bityuan.toml`）`initForkConfig` 会被跳过，分叉静默按注册
默认值生效；若默认注册为 0，升级二进制就会立刻在所有这类链上启用本规则，既可能让历史区块
重放失败，也会改变当下的开票/挖矿行为。回归：`TestForkTicketMinerAuthDefaultDisabled`。

1. `Action.TicketOpen`（`plugin/dapp/ticket/executor/ticketdb.go`）新增强制校验：
   - `address.CheckAddress(MinerAddress, height)`：非法地址直接拒绝；
   - `MinerAddress ∈ {fromaddr, getBind(ReturnAddress)}`：只能挂在自己名下，或挂在
     `ReturnAddress` 已通过 `Tbind` 明确授权的矿工名下，否则返回 `ty.ErrMinerAddr`。
   原有“非 `ReturnAddress` 签名者必须是该 `ReturnAddress` 绑定矿工”的逻辑保持不变，
   因此 `buyTicketOne`（自挖）与 `buyMinerAddrTicketOne`（冷地址绑定矿工）两条合法路径不受影响。
   注意：这条规则只能挡住“未绑定”的裸挂，挡不住自签 `Tbind`，见 §8.3。
2. 同一函数修掉 `Count > len(PubHashes)` 的越界 panic：`len(topen.PubHashes) == 0`
   改为 `len(topen.PubHashes) < int(topen.Count)`，返回 `ErrOpenTicketPubHash`。
   该修复无需分叉保护：panic 交易会被 `PreExecBlock` 从区块中丢弃，从未写入任何区块，
   因此不影响历史区块重放。
3. 分叉高度配置：
   - 主链 `chain33.fork.toml` = `48100000`、镜像主链的 `plugin/dapp/evm/cmd/ci2/chain33.proxyminer.toml`
     同为 `48100000`（按仓库既有做法取 `tip + 241920` ≈ 两周升级窗口，主链 2026-10-15 高度约 47.85M）；
   - `chain33.para.toml`、`chain33.para.toml.readme`、`plugin/dapp/ticket/rpc/rpc_test.go` 内嵌配置、
     `plugin/dapp/ticket/executor/testdata/chain33.cfg.toml` = `0`（平行链/测试配置直接生效）；
   - **未显式配置该键的链保持关闭**（注册默认值 `types.MaxHeight`），升级二进制本身不改变行为；
   - 分叉高度必须留在升级窗口之后：`Test_TicketOpen_LegacyPathBeforeFork` 断言分叉高度之前仍按旧逻辑放行。

### 8.2 矿工侧（无需分叉）：ReturnAddress 白名单过滤选票

执行器规则无法挡住“自签 `Tbind`”的挂票（§8.3），而这类票的危害发生在**选票/出块**环节：
本节点挖出票后，`ExecDepositFrozen` 会把出块奖励增发给票的 `ReturnAddress`。因此本 PR 在
选票入口增加本地过滤，不改变任何链上状态，**不需要分叉**：

- 位置：`plugin/dapp/ticket/wallet/exec.go` 的 `On_WalletGetTickets`（共识 `flushTicket`
  通过 `WalletGetTickets` 取票的唯一入口）→ `filterTrustedTickets`；
- 规则：只保留 `MinerAddress == ReturnAddress`（自挖）或 `ReturnAddress` 命中
  `wallet.sub.ticket.minerwhitelist` 的票；被过滤的票打 `skipUntrustedTicket` 告警；
- `minerwhitelist` 就是矿工本地的**批准名单**：只有名单内的 ReturnAddress 才能使用本矿工的
  挖矿身份。默认 `["*"]` 表示不过滤，保持原有挖矿行为；链上不新增任何批准动作；
- 局限：只保护升级并配置了白名单的节点；不阻止攻击者把票发上链，也不清除链上票据与本地
  索引（`TicketList` 仍能查到），即“链上归因伪造”仍在。它解决的是算力被占用、奖励被
  拿走这一核心危害。

回归：`TestFilterTrustedTickets`（默认 `*` 不过滤；显式白名单时注入票被过滤）。

### 8.3 为什么 `Tbind` 不能当作“矿工侧批准”

`Tbind`（`Action.TicketBind`）要求 `fromaddr == ReturnAddress`，即**票主单方签名**，矿工本人
不签名；它表达的是“票主授权某矿工动用我的票面资金”，不是“矿工同意别人把票挂到我名下”。
因此攻击者可以自签 `Tbind{ReturnAddress=自己, MinerAddress=受害矿工}`，随后
`Topen{MinerAddress=受害矿工, ReturnAddress=自己}` 就能通过 8.1 的绑定分支（回归用例
`Test_TicketOpen_SelfTbindStillInjects` 固定这一已知限制，票据仍会以受害矿工为 `MinerAddress`）。
要链上强制“矿工批准”，需要新增矿工签名的批准动作（新 action + 分叉 + 钱包支持，历史绑定还需
迁移），不在本 PR 范围；本 PR 用 §8.2 的本地白名单承担矿工侧批准。

### 8.4 合并上游的适配

已入库的 H-2 PoC 文件（`plugin/dapp/ticket/executor/ticket_miner_inject_poc_test.go`）中演示
“注入成功”的用例显式改用 `ForkTicketMinerAuth` 未生效的配置（`pocCfgBeforeMinerAuth`），并断言
该前提，从而继续固定修复前行为；`Count > len(PubHashes)` 用例改为断言 `ErrOpenTicketPubHash`
且不再 panic。

### 8.5 回归用例

执行器层（`plugin/dapp/ticket/executor/ticket_miner_auth_test.go`）：

| 用例 | 断言 |
| --- | --- |
| `TestForkTicketMinerAuthDefaultDisabled` | 没有 `[fork]` 段的链配置下分叉默认关闭（`types.MaxHeight`），升级不改变旧行为 |
| `Test_TicketOpen_RejectsForeignMinerAddress` | 第三方把票挂到别的矿工 → `ErrMinerAddr`，且校验发生在冻结之前（票面资金未被动用） |
| `Test_TicketOpen_AllowsOwnMinerAddress` | 自挖路径照常 `ExecOk`，票据以自己为 `MinerAddress` 落地 |
| `Test_TicketOpen_AllowsBoundMinerAddress` | 票主开票给已绑定矿工、绑定矿工代开票，两条路径都放行 |
| `Test_TicketOpen_RejectsInvalidMinerAddress` | 非法 `MinerAddress` 被地址校验拒绝 |
| `Test_TicketOpen_RejectsPubHashesCountMismatch` | `Count > len(PubHashes)` 返回 `ErrOpenTicketPubHash`，不再 panic |
| `Test_TicketOpen_LegacyPathBeforeFork` | 分叉高度之前保持旧行为，说明必须留足升级窗口 |
| `Test_TicketOpen_SelfTbindStillInjects` | 已知限制：自签 `Tbind` 仍可把票挂到任意矿工（由 §8.2 兜底） |

矿工侧（`plugin/dapp/ticket/wallet/ticket_test.go`）：

| 用例 | 断言 |
| --- | --- |
| `TestFilterTrustedTickets` | 默认 `["*"]` 不过滤；显式白名单时只保留自挖票与白名单 ReturnAddress 的票 |

仍建议后续处理（不在本 PR 范围）：分叉生效后拒绝非创世的 3 段 `ticketId`（清除分叉前预埋的旧格式票，
这类票的 `pubHash` 校验恒被跳过）；若需要链上强制“矿工批准”，新增矿工签名的批准动作；
已用创世重入伪造出来的 `IsGenesis` 票据需由运维用 force-close 清理。

# TicketClose 对重复 ticketId 重复付款分析

- **日期**：2026-10-15
- **组件**：`plugin/dapp/ticket`（ticket 合约，`TicketClose`）
- **等级**：中（同一张票被多次付款 → 该地址押金被提前解冻；**不盗币、不增发、主链无跨地址伤害**）
- **状态**：已修复，`ForkTicketCloseDup` 门控（**当前取 0**；§9 从"无第三方伤害"这一前提重新评估后，
  建议改取大于当前链尖的高度）
- **影响边界**：**不能盗走他人资产**（入账方恒为票自身 `ReturnAddress`，提取额受押金池封顶），
  **无增发**，且**不能跨地址伤及无关第三方** —— 唯一那条路径（抽干全链开发基金池）被系统分叉
  `ForkTicketFundAddrV1` 门控，而该分叉在主链早已生效，主链上根本不存在（§4.3、§4.7）。
  主链上真实可达的是**同一个返还地址内部**的影响：绕过提现时间锁（4.2）、抽干共享押金池致
  该地址其余票暂时关不掉（4.4）、以及随后的状态/回执不一致（4.5）与治理票权副作用（4.6）。
  跨交易的重复关闭本就由票状态挡住（§8）。详见第 4 节。

## 1. 结论

`TicketClose` 分两段：**第一段只校验、不写库**，把 payload 里每个 ticketId 读出来建成
`tickets` 切片；**第二段才逐个付款**。由于第一段的 `readTicket` 永远读到的是同一份未变更的
票状态，同一个 ticketId 在 payload 里出现 N 次，第二段就会**对同一张票支付 N 次**，
而库里只记一次关闭。

```go
for i := 0; i < len(tclose.TicketId); i++ {
    ticket, err := readTicket(action.db, tclose.TicketId[i])  // 只校验，不写库
    ...
    tickets[i] = &DB{*ticket, prevstatus}                     // 每个下标各存一份
}
for i := 0; i < len(tickets); i++ {
    retValue := t.GetRealPrice(chain33Cfg) + t.MinerValue
    receipt1, err := action.coinsAccount.ExecActive(t.ReturnAddress, action.execaddr, retValue)  // 每个下标各付一次
    ...
    t.Save(action.db)                                         // 写的是同一个 key
}
```

付款来源是 `(ReturnAddress, ticket 执行地址)` 下的**冻结池**，而该池同时是这个返还地址
**所有未关闭票**的押金。因此重复一次就多掏一份**同一地址其他未关闭票**的押金（该地址内部，
不跨地址）。

## 2. 根因

`plugin/dapp/ticket/executor/ticketdb.go` 的 `TicketClose`（修复前）问题有二：

1. **构造与副作用分离**：校验循环里没有任何写操作，重复的 id 不会互相看见。注意 `readTicket`
   走的是 `action.db`（StateDB），即便循环中写库，同一笔交易内的后续读也只在
   `Access`/`GetSetKeys` 语义下才可见 —— 该处原本就没有任何去重。
2. **既未做输入去重，也未校验「已关闭」**：`Status` 只在第二段开始时改内存对象，
   对第一段后续的读取不产生任何影响。

对比同类实现可以看到缺失的防线：`hashlock` 的 `Hashlocklock` 在入口就
`readHashlock` 判重并返回 `ErrHashlockReapeathash`（`hashlockdb.go:97`），
而 ticket 这里连最基本的「payload 内 id 唯一」都没有校验。

## 3. 攻击构造（已实测复现）

一笔 `Tclose`，payload 里把同一个 ticketId 写两次：

```json
{"execer": "ticket", "payload": {"ty": 13, "tclose": {"ticketId": ["<id>", "<id>"]}}}
```

前置条件：

1. 攻击者持有该票（`MinerAddress` 或 `ReturnAddress` 为 `tx.From`），票处于
   `TicketOpened`/`TicketMined` 且已过 `TicketWithdrawTime`/`TicketMinerWaitTime`；
2. 该返还地址的冻结池里至少有 `2 × TicketPrice`（即该地址还持有其他未关闭的票）。

`executor/ticketclose_dup_test.go` 的 `TestTicketCloseDupTicketIDPreFork` 即为该 PoC：

```
池初始 Frozen = 2 × 100，payload 内 id 重复两次
→ 交易 ExecOk
→ Balance = 200，Frozen = 0，票状态 = TicketClosed
```

一张票取走了**两张票的押金**，池被掏空。

## 4. 影响

### 4.1 不能盗走他人资产（已实测确认）

**付款的入账方恒为票自身的 `ReturnAddress`**，攻击者无法把钱转给自己或任意第三方：

```go
receipt1, err := action.coinsAccount.ExecActive(t.ReturnAddress, action.execaddr, retValue)
//                                              ^^^^^^^^^^^^^^^^ 固定为票的返还地址
```

而 `TicketClose` 的 `fromaddr` 校验只放行两类票：自己是 `ReturnAddress` 的票，或自己是其
**绑定矿工**的票（`action.fromaddr != ticket.MinerAddress && action.fromaddr != ticket.ReturnAddress`
→ `ErrFromAddr`）。因此：

| 攻击者身份 | 多付款项的归属 | 攻击者净收益 |
|---|---|---|
| 票主本人（`fromaddr == ReturnAddress`） | 自己的子账户 | 0（只是把自己冻结的钱提前解冻，且受自己押金总额封顶） |
| 绑定矿工（`fromaddr == MinerAddress`） | **票主**的子账户 | **0（分文未得，钱全进票主口袋）** |

`TestTicketCloseDupPaysOnlyReturnAddress` 实测该边界：绑定矿工发起 3 次重复关闭后，
票主余额 `+3×Price`、冻结归零，而**攻击者子账户余额与冻结均为 0**。

也无法凭重复取出超过该地址已存入的押金：`ExecActive` 在 `Frozen - amount < 0` 时返回
`ErrNoBalance`，提取额以池内余额为硬上限。`TestTicketCloseDupExtractionCappedByPool`
实测：池内只有 2×Price 却重复 5 次，第 3 次付款即失败（`ErrNoBalance`）。

另外，攻击者也无法为他人建票再据为己有：`TicketOpen` 的冻结资金取自 `topen.ReturnAddress`，
而 `TicketBind` 要求签名者就是 `ReturnAddress` 本人，故无法「用别人的钱开票、把返还地址写成自己」。

### 4.2 绕过提现时间锁（对系统的实际违背）

提现时间锁 `ticketWithdrawTime` 是**按单张票**校验的：

```go
if ticket.Status == ty.TicketOpened && action.blocktime-ticket.GetCreateTime() < cfg.TicketWithdrawTime {
    return nil, ty.ErrTime
}
```

但押金池是**按返还地址共享**的。于是只需要一张**已到期**的票，靠重复它就能把同一地址下
**尚未到期**的票的押金一并解冻出来 —— 时间锁对整池资金的约束被绕过。
`TestTicketCloseDupBypassesWithdrawLock` 实测：未到期的票 B 单独关闭被拒（`ErrTime`），
而重复已到期的票 A 一次即把整个池（含 B 的押金）解冻，B 自身从未到期。

### 4.3 消耗全链共享的开发基金池（**仅限 `ForkTicketFundAddrV1` 未生效的链**）

> **可达性结论（先看这条）。** 本节的路径在**主链上不存在**。它被 chain33 的**系统分叉**
> `ForkTicketFundAddrV1` 门控（`plugin/dapp/ticket/executor/ticketdb.go` 付款循环内）：
>
> ```go
> if t.prevstatus == 2 {
>     if !chain33Cfg.IsFork(action.height, "ForkTicketFundAddrV1") {
>         receipt2, err := action.coinsAccount.ExecActive(chain33Cfg.GetFundAddr(), action.execaddr, cfg.CoinDevFund)
> ```
>
> | 配置 | 取值 | 旧解冻分支 |
> |---|---|---|
> | `chain33.fork.toml:24`（仓内大数据 fork 配置，与代码默认一致） | `3350000` | **死**（早已生效） |
> | chain33 `types/fork.go:147` `RegisterSystemFork`（代码注册默认值，偏测试网口径） | `3350000` | **死** |
> | **部署仓 `bityuan/bityuan` 的 `bityuan.go:354`（比特元主链）** | **`4320000`** | **死**（当前主链高度约 4788 万 ≫ 432 万） |
> | `chain33.para.toml:331`、`chain33.para.toml.readme:265`、`build/ci/*/chain33.para.toml` | `-1` → `MaxHeight`（== 关闭） | **活** |
> | `plugin/dapp/ticket/executor/testdata/chain33.cfg.toml:219`、`rpc_test.go:274` | `-1` | **活**（本节"实测"跑在这里） |
>
> 主链高度是按主链 EVM（eth）最新块号取的（`plugin/dapp/evm/executor/evm.go` 把 chain33 高度直接
> 作为 `BlockNumber`），2026-10-16 约 47,878,986；即便不依赖该推断，`chain33.fork.toml` 与代码默认
> 的 `3350000` 也已远低于主链高度，结论相同。
>
> 即：该路径只存在于**分叉未生效**的链（平行链）与**执行器测试配置**上。平行链上它确实可达：
> `[fork.sub.ticket] Enable=0`（不是 `-1`，不触发转换）、ticket 驱动按 `GetDappFork(name,"Enable")=0`
> 注册，二者都意味着票合约从创世就能用，而 `ForkTicketFundAddrV1=-1 → MaxHeight` 让
> `!IsFork` 恒为真。因此本节既不能算主链
> 影响，也不能作为「能伤到无关第三方」的依据。相应用例已更名带 `PreFork` 后缀、并在运行时断言
> `requireLegacyDevFundPath`（fork 必须未生效，否则用例空转）；主链状态则由
> `TestTicketCloseMinedPostFundAddrV1KeepsFundPool` 固定（fork 生效时基金池分文不动）。
> 本条是复核意见 1/2/3 的直接产物，方法论教训见 §4.7。

**这个池是什么、钱从哪来。** 票的押金由 `TicketOpen` 从**用户余额**冻结而来，属于开票人自己；
而开发基金池不同：它是历次挖矿时由 `TicketMiner` 增发并冻结的币 ——

```go
// TicketMiner（同一分叉的 else 分支）
receipt2, err = action.coinsAccount.ExecDepositFrozen(chain33Cfg.GetFundAddr(), action.execaddr, cfg.CoinDevFund)
//                                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^ 记账在 fundKeyAddr 名下
// ExecDepositFrozen 内部：ExecIssueCoins(execaddr, amount) → depositBalance → 再冻结到 (fundKeyAddr, execAddr)
```

所以它是**全链唯一一个、由增发累积而成的池**，不属于任何用户，而是所有待关闭票共享的
「解冻储备」。`fundKeyAddr` 在链配置里固定（`[mver.consensus.ForkTicketFundAddrV1]`）。

**为什么能被动到（在分叉未生效的链上）。** 关闭一张**已挖矿**的票时，付款循环内除了付给票主，
还会从这个池解冻一份 `CoinDevFund`（上面那段 `if !IsFork(...)`）。

这一步在循环内，所以重复 N 次就解冻 N 份，与本地址押金多少**无关**。
`TestTicketCloseDupDrainsDevFundPoolPreFork` 实测：池内 10 份、重复 3 次 → 被消耗 3 份（应为 1 份）。

**在分叉未生效的链上，这才是能伤到第三方的路径（4.6）。** `ExecActive` 在 `Frozen < amount` 时
返回 `ErrNoBalance`，而该错误会让**整笔 tclose 失败**。攻击者把基金池抽干后，**其他毫不相干的
持票人**关闭自己已挖矿的票时就会卡在这一步：`ErrNoBalance`，整笔失败，押金与挖矿奖励都取不回。
`TestTicketCloseDupFundDrainBlocksOtherHoldersPreFork` 实测该跨地址影响：攻击者抽干基金池后，
受害者（另一私钥、另一地址）的 tclose 失败，票卡在 `TicketMined`；基金池回填 1 份后受害者即可
正常关闭。注意该用例只到执行器层、证明不了区块层的回滚（§7 覆盖面）。

注意：`TicketClose` 在此**只做解冻（`ExecActive`），不增发新币**。增发路径
（`ExecIssueCoins`）只存在于 `TicketMiner`，不在本漏洞路径上。

**主链上这条路径关闭后剩下什么。** 分叉生效后 `TicketClose` 完全不再触碰 `GetFundAddr()` 池
（`TestTicketCloseMinedPostFundAddrV1KeepsFundPool` 实测：票主照常拿回 `Price + MinerValue`，
基金池 `Frozen`/`Balance` 分文未动）。主链上真实可达的只有 4.2/4.4（同地址押金池）。

### 4.4 同地址其余票被锁住（可恢复）

押金池被抽干后，同地址其余票因 `ErrNoBalance` 无法关闭 —— 而关闭是唯一的解冻路径。

**程度限定（已实测）：不是永久锁死。** 挖矿奖励走
`ExecDepositFrozen(t.ReturnAddress, action.execaddr, MinerValue)`（见 `TicketMiner`），
存入的**正是同一个池**，因此池会随该地址继续挖到块而回填。
`TestTicketCloseDupDrainsSharedPool` 实测该恢复路径：抽干 → B 关闭失败（`ErrNoBalance`）
→ 回填一张票的 `Price` → B 关闭成功。危害因此是**持续到池被挖矿奖励填满为止**，而非永久损失。

另外已解冻的部分可通过 `withdrawFromTicketOne`（`wallet/ticket.go:645`）由账户私钥提取，
不会凭空消失。

### 4.5 状态与回执不一致

回执里会出现 N 条 `TyLogCloseTicket` 指向同一个票 id，而 Mavl 里只落一次关闭 ——
破坏「一笔票对应一次关闭」这一假设。

### 4.6 谁能被伤到（主链 vs 分叉未生效的链）

**主链（`ForkTicketFundAddrV1` 已生效）—— 伤害不跨地址。**

| 受害方 | 机制 | 能否恢复 |
|---|---|---|
| 攻击者自己 | 重复付款只把自己冻结的钱提前解冻，受自己押金封顶 | — （无净损失） |
| 同地址的其他票 | 共享押金池被抽干 → 暂时关不掉 | 能：挖矿奖励回填该池（4.4） |
| 同地址未到期的票 | 时间锁被绕过，押金被提前解冻 | 资金已解冻，属既成事实（4.2） |
| 该地址的治理票权 | `autonomy.getAddressVotes` 按 `voteAccount.Frozen / (TicketPrice×precision)` 计票（`plugin/dapp/autonomy/executor/boardaction.go:629`）：池被抽干 → 票权下降；若该链 `UseBalance=true` 则反向（可用余额被抬高 → 票权上升） | 视回填与治理规则 |

**「无增发」不等于「后果仅限于提前解冻」（复核意见 2，已在代码中逐条确认）。** 钱包选票
（`wallet/ticket.go` 的 `getTicketsByStatus`）、ticket 共识选票（`plugin/consensus/ticket/ticket.go`
的 `searchTargetTicket` / `CheckBlock`）与执行器的 `TicketMiner` 都**不校验「冻结余额是否仍覆盖
票价」**：`ExecActive` / `ExecFrozen` 只做聚合额校验，仓库里不存在任何 per-ticket 的
「票记录 ↔ 押金」对应检查。因此票 B 只要还是 `TicketOpened`/`TicketMined` 就能继续被选中挖矿，
而它背后的押金已被解冻；解冻出来的可用余额也可由该地址私钥经 coins `Withdraw`
（`wallet/ticket.go` 的 `withdrawFromTicketOne`）转出。
这不是盗币，但说明「共享押金池 ≥ 该地址所有未关闭票的押金」这一不变量已被打破。

**分叉未生效的链（平行链等）—— 才有一条跨地址路径。**

| 受害方 | 机制 | 能否恢复 |
|---|---|---|
| **任意第三方持票人** | **全链基金池被抽干 → 他们的 tclose 因 `ErrNoBalance` 失败，票卡住、押金与奖励取不回** | 能：其他票继续挖矿会为基金池补充新份额（4.3） |

⚠️ 上一行是本分析早先版本"唯一跨地址伤害无关第三方"的论据，但**主链上它不成立**
（4.3 可达性表：该 fork 主链早已生效）。它只在 `ForkTicketFundAddrV1` 未生效的链上成立。

**现实受害场景是矿池**：矿池作为 `ReturnAddress` 持有大量押金并绑定矿工 ——
被绑定的矿工即可把矿池的票批量锁住（消耗的是矿池自己的池）；在 fork 未生效的链上还会
同时消耗全链基金池、间接卡住其他持票人的提现。

### 4.7 可达性与方法论（为什么早先漏掉了分叉门）

第一版分析**引用了**那道分叉门（4.3 的代码块），却只把它当成"这段代码长什么样"的上下文，
没有把**它的值**当成可达性前提，于是把「fork 未生效的链 + 执行器测试配置」上的实测结论写成了
现网影响，连标题级结论（"跨地址伤到无关第三方"）都受污染。按链拆开后是这样：

| 影响路径 | 门控分叉 | 主链 | 平行链 | 执行器 testdata |
|---|---|---|---|---|
| 同地址共享押金池被抽干（4.2/4.4） | 无 | 可达 | 可达 | 可达 |
| 全链开发基金池被抽干（4.3） | `ForkTicketFundAddrV1` | **不可达**（已生效） | 可达（`-1`） | 可达（`-1`） |

四个具体成因，供后续分析复用：

1. **「代码里有分支」≠「链上可达」**：分析路径上出现的每个 `IsFork`/`IsDappFork` 都必须先解析成
   具体数字，再写影响结论 —— 不允许只引用门、不引用值。
2. **测试配置被当成生产配置**：`testdata/chain33.cfg.toml` 刻意设 `ForkTicketFundAddrV1=-1`
   （保留旧协议路径，供修复前的行为回归），测试文件里甚至已注明"该分支在 `ForkTicketFundAddrV1`
   之前生效"——**事实在手，却没有传导到 4.3/4.6 的结论**。规则：任何"实测"都要声明宿主配置的
   分叉取值，否则只能算"代码路径存在"。
3. **系统分叉 ≠ dapp 分叉**：`ForkTicketFundAddrV1` 是裸字符串系统分叉（注册在 chain33
   `RegisterSystemFork`，无导出常量），不在 `pty.Fork*` 常量表里；而且它的 `-1` 读作"fork 关闭"，
   恰恰意味着**旧代码路径打开** —— 扫配置时最容易被当成无关行。检索必须同时覆盖两个命名空间。
4. **分叉枚举应是固定检查轴，而不是"主题相关才做"**：同仓库的 H-2 分析
   （`docs/security/ticket-topen-mineraddress-injection-analysis.md`）因为漏洞本身围绕
   `ForkTicketId`，做了完整的按链分叉普查；H-1 与分叉正交，这条轴就被整条丢掉了 ——
   尽管另一个分叉恰好卡着它的一条影响路径。可达性要进流程，不能靠"想起来"。

## 5. 修复

单一收口点：在 `TicketClose` 的校验循环里对 payload 的 ticketId 去重，
发现重复立即返回，**在任何付款发生之前**：

```go
checkDup := chain33Cfg.IsDappFork(action.height, ty.TicketX, ty.ForkTicketCloseDup)
var seenIDs map[string]struct{}
if checkDup {
    seenIDs = make(map[string]struct{}, len(tclose.TicketId))
}
for i := 0; i < len(tclose.TicketId); i++ {
    if checkDup {
        if _, ok := seenIDs[tclose.TicketId[i]]; ok {
            tlog.Error("TicketClose repeated ticketId", "id", tclose.TicketId[i], "tx", common.ToHex(action.txhash))
            return nil, ty.ErrDupTicketID
        }
        seenIDs[tclose.TicketId[i]] = struct{}{}
    }
    ticket, err := readTicket(action.db, tclose.TicketId[i])
    ...
}
```

拒绝发生在付款循环之前，因此失败时余额、冻结池与票状态全部原样保留
（与「成功才推进水位」的写法不同：这里根本不需要额外 KV）。

## 6. 分叉与配置

新增 dapp fork `ticket.ForkTicketCloseDup`，按用户决定**取 0（立即生效）**。

含 `[fork.sub.ticket]` 的配置都必须有该条目，否则 chain33 `initForkConfig` 会 panic：

- `chain33.fork.toml`
- `chain33.para.toml` 与 `chain33.para.toml.readme`
- `plugin/dapp/evm/cmd/ci2/chain33.proxyminer.toml`
- `plugin/dapp/ticket/executor/testdata/chain33.cfg.toml`
- `plugin/dapp/ticket/rpc/rpc_test.go`（内嵌 cfgstring）

`Title="local"` 的配置走 `setLocalFork()`，全部 fork 归 0，因此修复在创世即生效。

**关于「补条目」的因果（复核意见 2 的更正）。** `initForkConfig` 的赋值循环只遍历 toml 里**写了**
的键（`for k, v := range forks.System` / `for dapp, forklist := range forks.Sub`），所以缺失的键不会被
`SetFork`，保留 `RegisterDappFork` 的注册默认值 —— 本 fork 注册的是 `0`
（`plugin/dapp/ticket/types/ticket.go` 的 `InitFork`），**即创世生效**。它**不会"落回 `MaxHeight`"**：
`MaxHeight` 只来自 toml 里的 `-1`，由同一个函数显式转换（`types/fork.go`）。要点：

- **缺键的两种后果**：正常配置（`disableCheckFork=false`）会在同一个函数的校验里
  `panic("exec ticket name ForkTicketCloseDup not config in config file")`；只有
  `DisableForkCheck=true` 的配置（如 `testdata/chain33.cfg.toml`）才会走到"静默按 0 生效"。
- 所以补条目是为了与仓库惯例一致、把生效高度显式写在配置里，**不是"条目在保证生效"**；
  反过来，**在 `DisableForkCheck=true` 的配置里省掉条目反而是静默生效（0）**，并不更保守。
- 「生效」不能靠"文本里存在键名"验证：`TestTicketForkConfigDeclared` 现在解析**段内**的值
  （要求取 0）、并断言本包实际加载的配置 `IsDappFork(0, ...)` 为真 —— 写成 `-1`
  （"配了但没生效"）会直接让用例失败。

**部署侧注意（新增）**：主链部署仓 `bityuan/bityuan` 当前 `[fork.sub.ticket]`
（`bityuan.go:395` 起，`ForkTicketId=1600000`、`ForkTicketVrf=2270000`）**没有** `ForkTicketCloseDup`
条目；本 PR 合并发布后若不同步补上，该链节点会在 `initForkConfig` 直接 panic。

## 7. 测试与验证

| 测试 | 覆盖 |
|---|---|
| `TestTicketCloseDupTicketIDPreFork` | 记录漏洞本身：分叉前重复 id 被接受，付 2×、池被掏空、票只关闭一次 |
| `TestTicketCloseDupTicketIDPostFork` | 分叉后重复 id 被拒（`ErrDupTicketID`），余额/冻结池/票状态全部未变；相邻重复与非相邻重复都覆盖 |
| `TestTicketCloseNormalPostFork` | 正常路径不受影响：单张关闭、两张不同票的批量关闭均照常付款并落库 |
| `TestTicketCloseDupPaysOnlyReturnAddress` | 影响边界（4.1）：绑定矿工发起重复关闭，钱只进票主口袋，攻击者分文未得 |
| `TestTicketCloseDupExtractionCappedByPool` | 影响边界（4.1）：重复 5 次但池内只有 2×Price 时，超额付款以 `ErrNoBalance` 失败 |
| `TestTicketCloseDupBypassesWithdrawLock` | 影响（4.2）：未到期票单独关闭被拒（`ErrTime`），但重复已到期票可解冻整池 |
| `TestTicketCloseDupDrainsDevFundPoolPreFork` | 影响（4.3，**仅 fork 未生效的链**）：全链共享的开发基金池被重复消耗 3 份（应为 1 份）；运行时断言 `ForkTicketFundAddrV1` 未生效 |
| `TestTicketCloseDupFundDrainBlocksOtherHoldersPreFork` | 影响（4.3/4.6，**仅 fork 未生效的链**）：**跨地址**——攻击者抽干基金池后，另一持票人的 tclose 失败、票卡住；断言受害者账户前后状态；基金池回填 1 份后恢复可关闭 |
| `TestTicketCloseMinedPostFundAddrV1KeepsFundPool` | **主链可达性**（4.3/4.7）：`ForkTicketFundAddrV1` 生效时关闭已挖矿票，票主照常拿回押金+奖励，基金池分文未动 |
| `TestTicketCloseDupDrainsSharedPool` | 影响边界（4.4）：抽干共享池 → 受害票关闭失败 → 挖矿奖励回填后恢复可关闭 |
| `TestTicketForkConfigDeclared` | 遍历 git 跟踪的、含 `[fork.sub.ticket]` 的配置：条目必须位于**该段内**、值必须为 0（`-1` 会让修复永不生效）、并断言本包实际加载的配置 `IsDappFork(0,...)` 为真；忽略 gitignore 下的本地生成配置 |

**变异验证**（均按预期失败后回滚，确保断言非空转）：

| 突变 | 应失败的用例 | 实测结果 |
|---|---|---|
| `ticketdb.go` 把 `checkDup` 守卫短路（`if false && checkDup`，回到修复前） | `TestTicketCloseDupTicketIDPostFork` | 失败（`expected 0 / actual 200`，重复付款重现） |
| `chain33.para.toml` 把 `ForkTicketCloseDup=0` 改成 `-1` | `TestTicketForkConfigDeclared` | 失败：`ForkTicketCloseDup=-1 无法保证修复生效` |
| 把该条目移出 `[fork.sub.ticket]` 段（注释掉） | `TestTicketForkConfigDeclared` | 失败：`[fork.sub.ticket] 段内缺少 ForkTicketCloseDup 条目` |
| testdata 把 `ForkTicketFundAddrV1=-1` 改成 `0` | 两个 `...PreFork` 用例 | 失败：`ForkTicketFundAddrV1 在 height=1000 已生效，开发基金池解冻路径不存在` |
| `ticketdb.go` 把 `if !IsFork(action.height, "ForkTicketFundAddrV1")` 短路为 `if true` | `TestTicketCloseMinedPostFundAddrV1KeepsFundPool` | 失败：`fork 生效后关闭已挖矿票不得解冻基金池` |

**覆盖面边界（复核意见 3）。** `TestTicketCloseDupFundDrainBlocksOtherHoldersPreFork` 直接调
`driver.Exec`，只到**执行器层**：`DriverBase.Exec`（chain33 `system/dapp/driver.go`）在动作返回
error 时既不写 statedb 也不回滚，而 `TicketClose` 是**直接写 statedb** 的，所以该用例里的失败交易
会留下中间态（受害者 `Frozen→0`、`Balance→2×Price`、票仍 `TicketMined`）。用例现在把这个
driver 级中间态**显式断言**出来（避免被误读为链上终态），并在一个全新环境里模拟"回滚已发生"，
验证"基金池补回 1 份即可恢复关闭"。链上终态由区块执行层的 `ForkExecRollback` 保证
（失败交易照常进块、扣手续费、状态改动全部丢弃），本 PR **没有**区块层端到端用例。

全量回归：`plugin/dapp/ticket/...`、`plugin/consensus/ticket/...`、`plugin/dapp/paracross/executor/...`
在 `-race` 下全部通过。

## 8. 残留与后续

- 只拒绝**同一笔交易内**的重复。跨交易的重复关闭本来就由票状态
  （`Status != TicketOpened && != TicketMined` → `ErrTicketClosed`）挡住，无需额外处理。
- 分叉取 0 意味着**历史区块若含重复 ticketId 的 tclose，将无法原样重放**。已按用户决定接受
  该风险；§9 从"无第三方伤害"这一新前提下重新评估了高度选择，结论是**建议改为大于当前链尖
  高度**（或先在目标链上跑通重放检查再取 0）。
- 本修复不涉及 `tickets` 切片与付款循环的先后结构（该结构本身是「先全部校验、再全部执行」
  的常规写法），只补上缺失的去重校验。
- 未做：区块层（`ForkExecRollback`）端到端的失败交易回滚用例；见 §7 覆盖面边界。

## 9. 修复必要性评估（"既然不伤害第三方，还要不要修"）

**结论：要修；但理由变了，因此分叉高度不建议维持 0。**

主链上的第三人伤害确实不成立（4.3 / 4.6 / 4.7）。于是这条修复的收益不再是"挡住对别人的攻击"，
而是下面三类 —— 它们都只落在**同一个返还地址内部**（攻击者必须是该票的 `ReturnAddress` 本人，
或该票的绑定矿工 `MinerAddress`）：

1. **提现时间锁被绕过（4.2）**：`ticketWithdrawTime` 是票一级的冷静期，重复投喂同一张已到期票
   即可把整池押金提前解冻，且解冻**不可逆**（进可用余额后可转出）。地址本人可借此提前解锁
   自己本该锁定的押金；若是绑定矿工发起，则是**半可信方越权**：他本只能按票逐张、按时间锁
   关闭，重复投喂让他一次性解除该地址全部票的资金锁定。
2. **协议不变量被打破（4.6）**：「共享押金池 ≥ 该地址所有未关闭票的押金」不再成立，票却仍能
   被选中挖矿、可用余额可被转走。今天没有任何代码依赖这个不变量（钱包选票 / 共识选票 /
   `TicketMiner` 都不查），**正因如此它没有第二道防线**：一旦将来有人在这个不变量上写逻辑
   （自治委员会票权已经间接依赖 `Frozen`，见 4.6），错误会被无声放大。
3. **回执 / 索引不一致（4.5）**：一笔交易里 N 条 `TyLogCloseTicket` 指向同一张票，而 Mavl 只落
   一次关闭；索引器、审计与对账都会被误导。

修复成本极低（入口一处 map 去重、拒绝发生在任何付款之前、不需要额外 KV），与上述收益相比
不对称。所以唯一要权衡的是**分叉高度**：

| 方案 | 收益 | 风险 |
|---|---|---|
| `0`（创世生效，本 PR 现状） | 立即覆盖所有高度 | 若**任何**目标链的历史里存在含重复 ticketId 的 tclose，全量重放会卡在该区块（该 tx 在新规则下变成失败交易，状态根与历史不一致） |
| 大于当前链尖高度（推荐） | 未来的重复 payload 一律被拒，上述三类收益照常拿到 | 生效前的老行为继续存在（只影响同一地址内部，且发起方必须是票主或绑定矿工） |

支持"改高度"的事实依据有两个，方向一致：

- **重复 id 只可能来自手工构造**：钱包 `closeTickets`（`wallet/ticket.go` 无去重）的 id 来自索引
  列表，而索引键 `LODB-ticket-tl:<addr>:<status>:<ticketID>` 中一张票只有一个状态、`List`
  每个键只返回一次，1-list 与 2-list 不会重叠；RPC `CloseTickets` 与 CLI `close` 根本不接受
  外部 ticketId。因此**诚实历史里出现重复的概率极低** —— 但无法从本仓库证明为 0。
- **取 0 换来的额外保护区间只是"创世至今的历史"**，而那段区间内的重复只可能来自蓄意构造；
  代价却是把"历史可原样重放"这条硬约束押在一个无法在本仓库内证伪的假设上。

因此建议：**分叉高度取一个大于当前链尖的值**（与仓库既有做法一致，参考 `ForkParaTitleExecMatch`）；
若部署方坚持立即生效，至少应先在目标链上完成 §8 提到的重放检查，再取 0。最终取值由部署方按链
的实际情况决定，本节只提供取舍依据。

**决策记录（2026-10-16）**：维护者选择**保持 0（立即生效）**，即接受"历史若含重复 ticketId 的
tclose 则无法原样重放"这一风险；本节关于改取大于链尖高度的建议仅作为后续参考保留。
`TestTicketForkConfigDeclared` 的值断言（要求段内取 0）与该决策一致。

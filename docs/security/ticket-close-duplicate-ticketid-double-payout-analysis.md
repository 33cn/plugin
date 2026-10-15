# TicketClose 对重复 ticketId 重复付款分析

- **日期**：2026-10-15
- **组件**：`plugin/dapp/ticket`（ticket 合约，`TicketClose`）
- **等级**：高（同一张票被多次付款，可抽干共享押金池）
- **状态**：已修复，`ForkTicketCloseDup` 门控（按用户决定取 0，立即生效）
- **影响边界**：**不能盗走他人资产**（入账方恒为票自身 `ReturnAddress`，提取额受押金池封顶），
  **无增发**；但能**跨地址伤到无关第三方**：抽干全链共享的开发基金池后，
  其他持票人的 tclose 会因 `ErrNoBalance` 失败，票卡在已挖矿状态、押金与奖励暂时取不回。
  另有两条同地址影响：绕过提现时间锁、抽干共享押金池致其余票暂时无法关闭。详见第 4 节。

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
**所有未关闭票**的押金。因此重复一次就多掏一份别人的押金。

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

### 4.3 消耗全链共享的开发基金池，并可阻断他人提现

**这个池是什么、钱从哪来。** 票的押金由 `TicketOpen` 从**用户余额**冻结而来，属于开票人自己；
而开发基金池不同：它是历次挖矿时由 `TicketMiner` 增发并冻结的币 ——

```go
// TicketMiner
receipt2, err = action.coinsAccount.ExecDepositFrozen(chain33Cfg.GetFundAddr(), action.execaddr, cfg.CoinDevFund)
//                                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^ 记账在 fundKeyAddr 名下
// ExecDepositFrozen 内部：ExecIssueCoins(execaddr, amount) → depositBalance → 再冻结到 (fundKeyAddr, execAddr)
```

所以它是**全链唯一一个、由增发累积而成的池**，不属于任何用户，而是所有待关闭票共享的
「解冻储备」。`fundKeyAddr` 在链配置里固定（主链 `chain33.fork.toml`、
`mver.consensus.ForkTicketFundAddrV1`）。

**为什么能被动到。** 关闭一张**已挖矿**的票时，付款循环内除了付给票主，还会从这个池解冻
一份 `CoinDevFund`：

```go
if t.prevstatus == 2 {
    if !chain33Cfg.IsFork(action.height, "ForkTicketFundAddrV1") {
        receipt2, err := action.coinsAccount.ExecActive(chain33Cfg.GetFundAddr(), action.execaddr, cfg.CoinDevFund)
        ...
    }
}
```

这一步在循环内，所以重复 N 次就解冻 N 份，与本地址押金多少**无关**。
`TestTicketCloseDupDrainsDevFundPool` 实测：池内 10 份、重复 3 次 → 被消耗 3 份（应为 1 份）。

**这才是能伤到第三方的路径（4.6）。** `ExecActive` 在 `Frozen < amount` 时返回 `ErrNoBalance`，
而该错误会让**整笔 tclose 失败**。攻击者把基金池抽干后，**其他毫不相干的持票人**关闭自己
已挖矿的票时就会卡在这一步：`ErrNoBalance`，整笔失败，押金与挖矿奖励都取不回。
`TestTicketCloseDupFundDrainBlocksOtherHolders` 实测该跨地址影响：攻击者抽干基金池后，
受害者（另一私钥、另一地址）的 tclose 失败，票卡在 `TicketMined`；
把基金池多留 1 份后受害者即可正常关闭（变异验证）。

注意：`TicketClose` 在此**只做解冻（`ExecActive`），不增发新币**。增发路径
（`ExecIssueCoins`）只存在于 `TicketMiner`，不在本漏洞路径上。

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

### 4.6 谁能被伤到

| 受害方 | 机制 | 能否恢复 |
|---|---|---|
| 攻击者自己 | 重复付款只把自己冻结的钱提前解冻，受自己押金封顶 | — （无净损失） |
| 同地址的其他票 | 共享押金池被抽干 → 暂时关不掉 | 能：挖矿奖励回填该池（4.4） |
| 同地址未到期的票 | 时间锁被绕过，押金被提前解冻 | 资金已解冻，属既成事实（4.2） |
| **任意第三方持票人** | **全链基金池被抽干 → 他们的 tclose 因 `ErrNoBalance` 失败，票卡住、押金与奖励取不回** | 能：其他票继续挖矿会为基金池补充新份额（4.3） |

最后一行是本漏洞唯一**跨地址伤害无关第三方**的路径，且不需要受害者做错任何事。

**现实受害场景是矿池**：矿池作为 `ReturnAddress` 持有大量押金并绑定矿工 ——
被绑定的矿工即可把矿池的票批量锁住（消耗的是矿池自己的池），同时消耗全链基金池、
间接卡住其他持票人的提现。

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
`testdata/chain33.cfg.toml` 虽设 `DisableForkCheck=true`，仍补上条目以免落回 `MaxHeight`。

## 7. 测试与验证

| 测试 | 覆盖 |
|---|---|
| `TestTicketCloseDupTicketIDPreFork` | 记录漏洞本身：分叉前重复 id 被接受，付 2×、池被掏空、票只关闭一次 |
| `TestTicketCloseDupTicketIDPostFork` | 分叉后重复 id 被拒（`ErrDupTicketID`），余额/冻结池/票状态全部未变；相邻重复与非相邻重复都覆盖 |
| `TestTicketCloseNormalPostFork` | 正常路径不受影响：单张关闭、两张不同票的批量关闭均照常付款并落库 |
| `TestTicketCloseDupPaysOnlyReturnAddress` | 影响边界（4.1）：绑定矿工发起重复关闭，钱只进票主口袋，攻击者分文未得 |
| `TestTicketCloseDupExtractionCappedByPool` | 影响边界（4.1）：重复 5 次但池内只有 2×Price 时，超额付款以 `ErrNoBalance` 失败 |
| `TestTicketCloseDupBypassesWithdrawLock` | 影响（4.2）：未到期票单独关闭被拒（`ErrTime`），但重复已到期票可解冻整池 |
| `TestTicketCloseDupDrainsDevFundPool` | 影响（4.3）：全链共享的开发基金池被重复消耗 3 份（应为 1 份） |
| `TestTicketCloseDupFundDrainBlocksOtherHolders` | 影响（4.3/4.6）：**跨地址**——攻击者抽干基金池后，另一持票人的 tclose 失败、票卡住；补回基金后即可关闭（变异验证） |
| `TestTicketCloseDupDrainsSharedPool` | 影响边界（4.4）：抽干共享池 → 受害票关闭失败 → 挖矿奖励回填后恢复可关闭 |
| `TestTicketForkConfigDeclared` | 遍历仓库内所有含 `[fork.sub.ticket]` 的配置，断言新 fork 条目齐备，防止漏配导致节点 panic |

**变异验证**：临时把 `ticketdb.go` 的守卫短路（`if false && checkDup`，保留符号定义；
等价于回到修复前），`TestTicketCloseDupTicketIDPostFork` 立即失败
（实测 `expected 0 / actual 200`，即重复付款重新出现），确认测试确实在验证修复而非空转。

全量回归：`plugin/dapp/ticket/...`、`plugin/consensus/ticket/...`、`plugin/dapp/paracross/executor/...`
在 `-race` 下全部通过。

## 8. 残留与后续

- 只拒绝**同一笔交易内**的重复。跨交易的重复关闭本来就由票状态
  （`Status != TicketOpened && != TicketMined` → `ErrTicketClosed`）挡住，无需额外处理。
- 分叉取 0 意味着**历史区块若含重复 ticketId 的 tclose，将无法原样重放**。
  已按用户决定接受该风险；如需保守，可把高度改为大于当前主链高度的值
  （参考 `ForkParaTitleExecMatch` 的处理）。
- 本修复不涉及 `tickets` 切片与付款循环的先后结构（该结构本身是「先全部校验、再全部执行」
  的常规写法），只补上缺失的去重校验。

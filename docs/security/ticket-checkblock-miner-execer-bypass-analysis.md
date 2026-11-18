# Ticket 共识 CheckBlock 不校验 tx[0] 执行器导致无票出块分析

- **日期**：2026-10-15
- **组件**：`plugin/consensus/ticket`（ticket 共识）、`plugin/dapp/ticket`（ticket 合约）
- **等级**：严重（可无票出块，且可赢得分叉选择）
- **状态**：已修复，`ForkMinerExecerCheck` 门控

## 1. 结论

ticket 共识的 `CheckBlock` 对挖矿交易（`tx[0]`）只校验 **payload 的形状**，从不校验
**`tx[0].Execer`**。而执行层是按 `tx.Execer` 路由的，于是攻击者可以把 `tx[0]` 指向
其他执行器，使 ticket 合约的 `Exec_Miner` —— 唯一校验「票归属、票状态、冻结时间」的地方
—— 根本不执行，同时仍能取得 `ExecOk` 回执满足 `CheckBlock` 的判定，从而**在没有票的情况下出块**。

`CheckBlock` 其余条件都是攻击者可自算的：`Reward`/`Bits`/`Modify` 均为自报字段，
PoW 是对攻击者自选的 `blocktime/ticketId/modify/privHash` 求哈希，VRF 又是用 `tx[0]`
自己的密钥验证的。实测仅枚举 `blocktime` 就能把 target 压到比 powLimit 难约 2^128，
因此伪造区块不只是被接受，还能在 `CmpBestBlock` 中取胜。

## 2. 根因

`plugin/consensus/ticket/ticket.go` 的 `getMinerTx`（修复前）：

```go
baseTx := current.Txs[0]
var ticketAction ty.TicketAction
err := types.Decode(baseTx.GetPayload(), &ticketAction)   // 只看 payload
if ticketAction.GetTy() != ty.TicketActionMiner {
    return nil, types.ErrCoinBaseTxType
}
if ticketAction.GetMiner() == nil {
    return nil, ty.ErrEmptyMinerTx
}
```

`CheckBlock` 在此基础上只补一条回执判定：

```go
if current.Receipts[0].Ty != types.ExecOk {
    return types.ErrCoinBaseExecErr
}
```

两条防线都只看「payload 是不是一个 Miner 动作」和「第 0 笔回执成功没有」，
**没有一条**把 `tx[0]` 绑定到 ticket 执行器。而 `Exec_Miner`（`ticketdb.go` 的
`Action.TicketMiner`）才是真正做校验的地方：

| `Exec_Miner` 的校验 | 被绕过后的后果 |
|---|---|
| `index != 0` → `ErrCoinBaseIndex` | — |
| 读取 `ticketId` 对应票，不存在则报错 | **可凭空捏造票 id** |
| `ticket.Status == TicketOpened` | 无票也无需状态 |
| `blocktime - CreateTime >= TicketFrozenTime` | 不受冻结期约束 |
| `fromaddr` 必须等于票的 miner/return 地址 | **不需要拥有任何票** |
| `privHash` 与票的 pubHash 匹配 | 可自选 privHash |
| 冻结/增发奖励到票的地址 | 奖励流向被改写 |

即：这七条校验全部失效。

## 3. 攻击构造（PoC 已验证）

`plugin/consensus/ticket/c1_checkblock_execer_poc_test.go` 在真实 PoW ticket 测试链上
构造并提交了这样一个区块，修复前被 `CheckBlock` 接受。

### 3.1 前置条件

一笔普通的用户交易：用 `coins` 的 `TransferToExec` 把少量币锁进 ticket 执行器账户。
这是让下面那笔「冒名」的 coins 动作能成功执行的唯一前提，而该动作会把币原路取回。
不需要任何票。

### 3.2 双解码 payload

关键在于构造**同一段字节**，使它同时是：

- `TicketAction{ Ty = 16 (Miner), Miner != nil }` —— `getMinerTx` 检查的形状
- `CoinsAction{ Ty = 3 (Withdraw), Withdraw != nil }` —— 实际执行的动作

成立的依据有两点：

1. **protobuf-go 容忍 wire-type 不匹配**。某字段声明为 message（bytes），遇到 varint
   时该值被当作未知字段跳过，不报错。
2. **`oneof` 后写的 tag 生效**（last-wins）。

`TicketAction` 与 `CoinsAction` 的顶层 tag 恰好重叠（1、4，以及 `ty` 在 coins 是 3、
ticket 是 10），`TicketMiner` 又与 `AssetsWithdraw` 在字段 2/3/4/5 上重叠，于是同一段
字节对两个执行器都是合法消息。顶层按 `3 → 4 → 10` 的顺序写入，使 **两个 oneof 都落在
tag 4**。

一个实现细节：ticket 的 `privHash` 在 tag 5，对 coins 而言是 `withdraw.to`（proto
`string`，会被 UTF-8 校验），所以 `privHash` 必须以 hex 文本而非裸字节写入，否则解码报
`string field contains invalid UTF-8`。`CheckBlock` 从不校验 `privHash`，任意编码均可。

### 3.3 其余条件的自洽

- **Reward**：`cfg.CoinReward + calcTotalFee(block)`，而 `calcTotalFee` 恒返回 0，
  所以填 `CoinReward` 即可。
- **Bits**：等于 `current.Block.Difficulty`，自报且与 header 一致即可。
- **Modify**：取自父块的 `miner.Modify`（`defaultModify`），公开可读。
- **PoW**：`getCurrentTarget(blocktime, ticketId, modify, privHash)` 的输入全部由攻击者
  掌握，且只要求 `<= powLimit`。
- **VRF**：`vrfVerify(minerTx.Signature.Pubkey, ...)` —— 用 `tx[0]` 自己的密钥验证，
  攻击者当然能算对。

## 4. 影响

1. **无票出块**：不持有任何 ticket 即可让区块通过 `CheckBlock`。
2. **奖励流向被改写**：`Exec_Miner` 不执行，`Reward` 不再按票分发，`coins.IsFriend`
   那条「挖矿交易可改 coins 状态」的授权路径也随执行器改变而不再适用。
3. **赢得分叉选择**：`CmpBestBlock` 复用同一个未校验的 `getMinerTx`，按
   `getCurrentTarget` 比难度。实测枚举 `blocktime`（20 万次采样）即可得到比 powLimit
   难约 2^128 的 target，因此伪造块能被选为主链。
4. **回执/状态不一致**：链上出现「第 0 笔不是 ticket 执行」的区块，破坏了
   `Receipts[i] ⇄ Txs[i]` 之外更根本的「挖矿交易即 ticket 交易」这一假设。

## 5. 修复

单一收口点：所有调用方（`CheckBlock`、`CmpBestBlock`、`addMinerTx`）都经过
`getMinerTx`，因此在那里加校验即可全覆盖。

```go
func (client *Client) checkMinerExecer(height int64, tx *types.Transaction) error {
    cfg := client.GetAPI().GetConfig()
    if !cfg.IsDappFork(height, ty.TicketX, ty.ForkMinerExecerCheck) {
        return nil
    }
    if want := cfg.ExecName(ty.TicketX); string(tx.GetExecer()) != want {
        return ty.ErrMinerTxExecer
    }
    return nil
}
```

在 `getMinerTx` 中、解码 payload 之前调用。

用 `cfg.ExecName(ty.TicketX)` 而非字面量 `"ticket"`，与框架自身对「挖矿交易执行器」的
定义保持一致：`coins.IsFriend`（chain33 `system/dapp/coins/executor/coins.go`）用
`cfg.ExecName(exec) == string(othertx.Execer)` 识别挖矿交易，
`autonomy` 也用 `ticketName = cfg.ExecName(auty.TicketX)`；同时它能正确处理平行链前缀。
`createMinerTx` 经 `types.CreateFormatTx(cfg, "ticket", ...)` 产生诚实 `tx[0]`，
其 `Execer` 正是该值，因此诚实出块不受影响。

## 6. 分叉与配置

新增 dapp fork `ticket.ForkMinerExecerCheck`，按用户决定**取 0（立即生效）**。

含 `[fork.sub.ticket]` 的配置都必须有该条目，否则 chain33 `initForkConfig` 会 panic：

- `chain33.fork.toml`
- `chain33.para.toml` 与 `chain33.para.toml.readme`
- `plugin/dapp/evm/cmd/ci2/chain33.proxyminer.toml`
- `plugin/dapp/ticket/executor/testdata/chain33.cfg.toml`
- `plugin/dapp/ticket/rpc/rpc_test.go`（内嵌 cfgstring）

`Title="local"` 的配置（共识测试用）走 `setLocalFork()`，全部 fork 归 0，因此修复在
创世即生效，可直接被 PoC 验证。未显式配置 `[fork]` 的配置（如 `chain33.toml`）沿用
代码默认值 0。

## 7. 测试与验证

| 测试 | 覆盖 |
|---|---|
| `TestC1_CheckBlockRejectsNonTicketMinerTx` | 端到端：真实 PoW 链上构造 `Execer="coins"` 的 `tx[0]`，断言被拒且错误恰为 `ErrMinerTxExecer` |
| `TestGetMinerTx_RejectsForeignExecerAfterFork` | fork 生效后，`coins`/`trade`/`none`/`ticket2`/空 等执行器一律被拒；诚实 `ticket` 仍通过 |
| `TestGetMinerTx_AcceptsForeignExecerBeforeFork` | 门控可关闭时保持旧行为；且 payload 形状规则仍独立生效 |

PoC 内置**对照组**：用诚实 `ticket` 执行器 + 不存在的票 id，必须被原有的回执防线
`ErrCoinBaseExecErr` 拒绝。这证明测试确实走到了 `CheckBlock`，而不是空洞通过
——（该对照组不能用 reward 规则：reward 判定位于回执门之后，非 ticket 执行器的区块根本到不了那里）。

**变异验证**：临时把强制逻辑短路（保留符号定义）后，上述 PoC 与全部 5 个执行器用例
均失败，确认测试真的在验证修复而非空转。

全量回归：`plugin/consensus/...`、`plugin/dapp/ticket/...` 全部通过。

## 8. 残留与后续

- 本次只校验执行器**名字**。`tx[0]` 的 payload 仍可与 `TicketMiner` 之外的字段组合，
  但 `Exec_Miner` 恢复执行后，第 3 节表格中的票归属/状态/冻结校验重新生效，
  伪造票 id 会在执行层被 `readTicket` 拒绝。
- PoC 依赖「攻击者自持 `tx[0]` 签名密钥 + powLimit 目标可枚举」这一测试链参数；
  在主网参数下 PoW 难度更高，但该构造**不依赖难度**——难度只影响出块成本，
  不改变校验缺失这一事实。

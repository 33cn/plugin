# EVM 账户黑名单：为什么框架层拦截之外还需要 EVM 内部检查

## 背景

7.20 uint64 溢出攻击（见 [evm-uint64-overflow-attack-analysis.md](./evm-uint64-overflow-attack-analysis.md)）之后，
chain33 在框架层引入 `ForkAccountBlacklist` 与 `types.CheckTxBlockedAccount`，
在 mempool、出块验块、executor.checkTx 三处统一拦截涉及黑名单地址的交易。

本 PR（33cn/plugin#1304）在 EVM 插件内部再加了若干检查。评审中的核心问题是：
**框架层既然已经完整拦截，EVM 层的检查是冗余还是必要？**

本文按「拦截从黑名单把资产打出」这一判定标准逐条回答，并说明代码取舍与回放约束。

## 判定标准

黑名单的目标是**冻结**攻击者已到手的资产，而不是阻止别人给它打钱。
资产打入黑名单等同于加深冻结，本就不必拦；
真正要封死的是每一条能把黑名单地址名下余额转移出去的路径。

## 1. chain33 框架层覆盖了什么

三处入口最终都走 `types.checkTxBlockedAccountCore`（chain33 `types/account_blacklist.go`）：

| 层 | 位置 | fork 门控 |
|---|---|---|
| mempool 入口 | `system/mempool/check.go:76`、`eventprocess.go:216/339` | 无，随升级立即生效 |
| 出块 / 验块 | `system/consensus/base.go:606` `AddTxsToBlock` | 有 |
| executor.checkTx | `executor/execenv.go:173` | 有 |

判定四个维度：`tx.From()`、`tx.GetTo()`、`tx.GetRealToAddr()`、
EVM payload 中的 `ContractAddr` 与**恰好 20 字节**的 `Para`。

因为任何"打出"都必须由黑名单地址自己签名一笔交易，`from` 维度把主路径全部封死：

| 案例 | 命中维度 |
|---|---|
| 黑名单发起 EVM 合约调用 / 纯转账 | `from` |
| 黑名单发起 coins / token 转账 | `from` |
| 黑名单提现 evm 子账户余额（`coins.Withdraw` 受益人恒为 `tx.From`） | `from` |
| 外部直接调用黑名单合约 | `to` / `ContractAddr` |
| 交易组任一笔命中 | 整组拒绝 |
| 黑名单为矿工地址（挖矿交易由矿工签名） | `from` |

框架 checkTx 在 `execTx` 中先于驱动的 `CheckTx` / `Exec` 执行，
所以真实节点上命中名单的交易**根本进不到 EVM 驱动**。

## 2. 框架层看不见、只有 EVM 内部能拦的路径

框架只解析交易**信封**。合约跑起来之后内部 CALL 了谁、以谁的名义付款，
只存在于运行时栈和 calldata 里。以下每条路径在 `blacklist_gap_test.go` 都有对应用例，
用真实 coins 账户断言余额，并以 fork 关闭为对照证明路径确实是活的。

### B1. 黑名单合约被内部 CALL 唤醒后转出自身余额

```
干净用户 U → 干净合约 B → CALL → 黑名单合约 A → A 把自己的余额转给 U
```

信封：`from=U, to=B`，框架放行。A 只出现在 B 的运行时栈上。
**拦截点：`runtime/evm.go` `Call` 的 target 检查**（`checkBlockedAccount(evm, caller, addr)` 中的 `addr`）。

用例：`TestGapB1_BlockedContractWokenByInnerCall`。

### B2. token 预编译 `transfer(from, to, amount)` 第三方代打 —— 价值最高

`vm/runtime/token.go` 的 `transfer` 分支：

```go
from := common.BytesToAddress(input[4:36])   // 取自 calldata，与 caller 无绑定
...
evm.StateDB.TransferToToken(from.String(), to.String(), tokenName, amount)
```

任何 `manage` 名单内创建者部署的合约都能把 `from` 指定为黑名单地址。
chain33 侧看到的 `Para` 是 100 字节 ABI calldata，`IsBlockedAccountRaw` 要求精确 20 字节，**完全看不见**。
**拦截点：`statedb.go` `TransferToToken` 的 `from` 检查**，这条链路没有任何替代防线。

用例：`TestGapB2_TokenPrecompileThirdPartyFrom`（走真实 `tokenPrecompile.Run`）。

### B3. 黑名单合约 SELFDESTRUCT 把余额打给受益人

`opSuicide` → `AddBalance(beneficiary, CodeAddr, balance)` → `statedb.Transfer(CodeAddr → beneficiary)`。
**拦截点：`statedb.go` `Transfer` 的 `sender` 检查**。

用例：`TestGapB3_SelfdestructBeneficiary`。

### B3b. DELEGATECALL 到黑名单代码执行 SELFDESTRUCT —— Call 检查覆盖不到

```
干净用户 U → 干净合约 P → DELEGATECALL → 黑名单合约 A 的代码执行 SELFDESTRUCT
```

`runtime.DelegateCall` 没有黑名单检查（借代码在调用者上下文执行，一般不构成打出）。
但 `opSuicide` 的**付款方取 `contract.CodeAddr`（= A）**，金额取 `contract.Address()`（= P）的余额：
攻击者给 P 充值 X，就能让 A 向受益人付出 X。

这条路径证明 `statedb.Transfer` 的 `sender` 检查不是对 `Call` 检查的重复，
而是唯一能拦下它的位置。

用例：`TestGapB3b_DelegatecallSelfdestructDrainsCodeAddr`。

### B4. 黑名单合约内部带 value 的 CALL

与 B1 部分重叠；`Transfer` 的 `sender` 检查是最后一道闸，任何未来新增的调用路径都会落到这里。

用例：`TestGapB4_BlockedContractInnerValueCall`。

## 3. 谁都拦不住的盲区

**ERC20 合约 storage 记账的 `transferFrom(黑名单, x, n)`**：黑名单事先 `approve` 一个干净地址，
余额是 keccak storage 的读写，EVM 层没有任何钩子可挂。
唯一解法是把该代币合约地址一并列入黑名单（冻结整个合约），或由代币合约自身加检查。

`ERC20.transfer(黑名单, x)` 则是打入，按本文标准不拦。

## 4. 代码取舍

### 实质防线（3 处，不可替代）

| 位置 | 覆盖 |
|---|---|
| `runtime/evm.go` `Call` 的 **target** 检查 | B1 |
| `statedb.go` `TransferToToken` 的 **from** 检查 | B2 |
| `statedb.go` `Transfer` 的 **sender** 检查 | B3 / B3b / B4 |

### 纵深防御（对真实交易冗余，源码中已逐处注明）

| 位置 | 为什么对真实交易不生效 |
|---|---|
| `evm.go` `CheckTx` 的 from 检查 | 框架 checkTx 已在其之前拦截；且函数开头 `IsPara()` 直接返回，平行链上永不执行 |
| `exec.go` `innerExec` 的 from / contractAddr 检查 | `msg.From() == tx.From()`，框架 `from` / `ContractAddr` 维度已覆盖 |
| `statedb.go` `CanTransfer` 的 sender 检查 | 与 `Transfer` 重复；上层把 `false` 翻译成 `ErrNoBalance`，排查时以日志 `blocked account` 为准 |
| `runtime/evm.go` `Call` / `Create` 的 **caller** 检查 | 顶层 caller 即 `tx.From`；内部 caller 若是黑名单合约，必先通过某次 `Call` 的 target 检查 |
| `Transfer` / `TransferToToken` 的 **recipient** 检查 | 拦的是打入。保留以符合 chain33"禁止收发"的原始设计，删除不影响冻结目标 |

`innerExec` 的 receiver 检查有一处实质覆盖：`isTransferOnly` 路径下 `Para` 超过 20 字节时，
chain33 的 `Para` 维度看不见，而 `BytesToAddress` 取后 20 字节仍能解析出黑名单地址
（用例 `TestGap_ParaLengthAsymmetry`）。这只能打入，不紧急，但两侧行为已由测试固定。

## 5. 回放兼容约束

上表"纵深防御"的分支**已随本分支在主网执行过**。它们对真实交易开不了火，
但一旦升级窗口内出现过一笔"信封干净、内部碰到黑名单"的交易，其 receipt / 状态根就依赖这些分支。
用去掉检查的二进制回放旧块会对不上。

因此撤除任何一处检查必须满足其一：

- 新开 fork：旧高度仍走现逻辑，新高度跳过，老代码留在二进制里供回放；或
- 先证明启用本分支且 `ForkAccountBlacklist` 生效至今，没有任何一笔交易的执行结果依赖这些分支
  （检索日志 `blocked account`，或新旧二进制回放同段区块比对 state hash）。

在此之前，"纵深防御"只作为评审结论记录在源码注释与本文中，不在本 PR 内删除。

## 6. fork 门控

statedb / runtime 两层的检查都以 `cfg.IsFork(height, ForkAccountBlacklist)` 门控。
未到分叉高度时不得改变执行结果，否则与未升级节点分链（用例 `TestGap_ForkGateBlocksNothingBeforeHeight`）。
mempool 入口无门控是故意的：它是节点本地行为、不进状态计算，随二进制升级立即止血。

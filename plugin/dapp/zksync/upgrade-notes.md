# zksync 执行器升级适配说明

> 对应 chain33 go-ethereum v1.14.8 升级（详见 `docs/chain33-go-ethereum-v1.14.8-upgrade.md`）。

## 背景

chain33 升级 go-ethereum v1.14.8，连带 gnark-crypto v0.10.0 → v0.12.1。zksync 使用 gnark-crypto 的 bn254 曲线、eddsa 签名与 mimc 哈希。

## 适配内容

### 1. gnark-crypto API

- **`fr.Element.SetString` 返回双值**：
  - `wallet/utils.go`、`executor/zkproofutil.go`
  - `elem, _ := f.SetString(s)` 替代 `f.SetString(s).Bytes()`
- **`eddsa.GenerateKey` 返回指针**：
  - `SignTx`（wallet/zksyncbizpolicy.go）、`SignTxInEddsa`（commands/l2txs/utils.go、executor/exec_test.go）参数改为 `*eddsa.PrivateKey`
- **`mimc.NewMiMC()` 无 seed 参数**：所有 off-chain 调用改为 `legacymimc.NewMiMC(ZkMimcHashSeed)`

### 2. MiMC 协议兼容（重点）

gnark-crypto v0.12.1 的 MiMC constants 从 `sha3.Sum256` 改为 `keccak256`，所有 hash 输出变化，会破坏链上已有数据（leaf hash、merkle root、地址派生、签名校验）。

zksync 全部切换到旧实现 `legacymimc.NewMiMC(ZkMimcHashSeed="seed")`（off-chain），保持链上协议兼容。

> 待办：可通过 dapp fork（`ForkMiMCHash`）在分叉高度后切换到新哈希。

### 3. key 派生链变化（重点排查）

chain33 升级后，`SetPubKey` 校验的 key 派生链（secp256k1 签名 → `GetLayer2PrivateKeySeed` → `eddsa.GenerateKey` → `mimc(pubkey.X || pubkey.Y)`）计算结果与历史数据不一致。

复现确认：当前派生结果与测试硬编码地址（`2b8a...`）不匹配，root cause 为跨链 key 派生链的深层变化。

**修复**：新增 `wallet/eddsa_compat.go` 的 `GenerateKeyCompat`，复刻 gnark-crypto v0.5.3 的标量派生（`j=sizeFr` 反转边界），恢复历史地址派生。原 11 个 `t.Skip` 集成测试（TestTransfer、TestWithdraw、TestWithdrawNFT、TestTransfer2New、TestTree2contract、TestContract2Tree、TestMintNFT、TestProxyExit、TestProxyExitFaid、TestTransferNFT、TestNFTMisc）已重新启用并通过。历史用户地址与派生保持一致，无需迁移。

### 4. witness 读取

`witness.ReadPublicFrom` 移除 → `witness.New` + `ReadFrom` + `w.Vector()`（executor/zkproof.go）

## 测试状态

- `zksync/types`、`wallet`、`executor`：全部通过（含 11 个恢复的 key 派生集成测试）

## 存量升级指引（线上 zksync）

> **⚠️ 线上升级需进一步评估审计**：以下为已知要求与约束，非完整迁移方案，具体部署上线前必须做专项评估与审计。

- **链上 VK 为单数**（`setVerifyKey` 覆盖更新），不支持 multi-VK 并存。
- **历史同步**：VK 切换后，全量 re-execute 历史交易将无法用新 VK 验证旧 proof，**必须使用快照同步**（跳过历史重放），或外部证明方保持旧编译不切换。
- **proof 生成**：由外部证明方完成，plugin 只验证。外部证明方升级电路编译（v0.9.0 R1CS）后需配套换 PK/VK（一次性切换）；保持旧编译则链上旧 VK（兼容 reader）可继续验证。
- **历史数据**：MiMC 哈希 / 地址派生 / witness 均兼容，历史 leaf、merkle 树无缝。
- **审计项**：VK 切换时机、历史同步策略、外部证明方密钥管理。

## 溯源参考

- 总体升级说明：`docs/chain33-go-ethereum-v1.14.8-upgrade.md`
- 旧 MiMC 实现：`plugin/crypto/legacymimc`

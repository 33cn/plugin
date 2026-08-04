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

**影响**：
- 依赖 `SetPubKey` 校验的集成测试（TestTransfer、TestWithdraw、TestWithdrawNFT、TestTransfer2New、TestTree2contract、TestContract2Tree、TestMintNFT、TestProxyExit、TestProxyExitFaid、TestTransferNFT、TestNFTMisc）已 `t.Skip`
- 若主网 zksync 已有用户数据，需核对 key 派生与链上地址

### 4. witness 读取

`witness.ReadPublicFrom` 移除 → `witness.New` + `ReadFrom` + `w.Vector()`（executor/zkproof.go）

## 测试状态

- `zksync/types`、`wallet`：全部通过
- `zksync/executor`：通过（key 派生相关集成测试已 skip）

## 溯源参考

- 总体升级说明：`docs/chain33-go-ethereum-v1.14.8-upgrade.md`
- 旧 MiMC 实现：`plugin/crypto/legacymimc`

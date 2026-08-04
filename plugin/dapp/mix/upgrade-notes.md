# mix 执行器升级适配说明

> 对应 chain33 go-ethereum v1.14.8 升级（详见 `docs/chain33-go-ethereum-v1.14.8-upgrade.md`）。

## 背景

chain33 升级 go-ethereum v1.14.8，连带 gnark v0.5.2 → v0.9.0、gnark-crypto v0.10.0 → v0.12.1。mix 深度依赖 gnark（电路）与 gnark-crypto（哈希/曲线运算），受影响较大。

## 适配内容

### 1. 电路 API（gnark v0.9.0）

- **`Define` 签名**：`Define(curveID ecc.ID, api frontend.API)` → `Define(api frontend.API)`
  - `types/deposit.go`、`withdraw.go`、`transferInput.go`、`transferOutput.go`、`authorize.go`
- **`frontend.Variable` 变为 interface{}**：
  - `Assign()` 移除 → 直接赋值
  - `GetWitnessValue()` 移除 → `mixTy.VariableToElement()`（types/util.go 新增）
- **电路内 mimc**：`gnark/std/hash/mimc` → `legacymimc.NewCircuitMiMC(api, seed)`

### 2. 曲线运算

- **路径迁移**：`gnark/std/algebra/twistededwards` → `gnark/std/algebra/native/twistededwards`
- **`NewEdCurve` 签名**：`NewEdCurve(id)` → `NewEdCurve(api, twistededwards.ID)`
- **`ScalarMul` 改名**：`PointAffine.ScalarMul` → `ScalarMultiplication`
- **`frontend.Compile`**：改用 `r1cs.NewBuilder` + `ecc.BN254.ScalarField()`
- **groth16**：`Prove`/`Verify` 需要 `witness.Witness`（`frontend.NewWitness`），`ReadAndVerify` 移除

### 3. MiMC 协议兼容（重点）

gnark-crypto v0.12.1 将 MiMC constants 从 `sha3.Sum256` 改为 `keccak256`，**所有 MiMC hash 输出变化**，会破坏链上已有数据（note hash、merkle root、zk proof）。

mix 全部切换到旧实现：

- off-chain：`legacymimc.NewMiMC(MimcHashSeed)`（wallet/cryptokey.go、executor/committree.go）
- in-circuit：`legacymimc.NewCircuitMiMC(api, MimcHashSeed)`（5 个电路）

`legacymimc.CircuitMiMC` 保持 gnark **v0.5.2 的 Miyaguchi-Preneel 算法**（`E(m,key)+m`），与 v0.9.0（`E(h+m)`）不同，必须保留旧算法才能匹配链上旧 proof 语义。

> 待办：可通过 dapp fork（`ForkMiMCHash`）在分叉高度后切换到新哈希。

### 4. CBC 随机 IV 适配

chain33 的 `CBCEncrypterPrivkey` 改为随机 IV，返回 `IV(16)+ciphertext` 格式。但其 `CBCDecrypterPrivkey` 新格式仅支持 32 字节明文（钱包私钥场景），mix 加密数据更大。

`wallet/cryptokey.go:decryptDataWithPading` 自行按新格式解密，并回退兼容旧格式。

### 5. groth16 序列化格式（密钥重新生成）

gnark v0.9.0 的 VK/PK/proof 二进制格式与 v0.5.2 不兼容，且密钥绑定 R1CS 布局，**必须用 v0.9.0 重新生成**。旧 `chain33key.tar.gz`（2022，v0.5.2）无法被 v0.9.0 读取（`read pk: invalid fr.Element encoding`），导致 `mix deposit` 生成 proof 失败 → 无 note。

**处理**：
- 新增 `mix/cmd/genzkkey/`：用 gnark v0.9.0 编译 5 个电路并生成 PK/VK hex 文件（与 `createZkKeyFile` 输出一致）
- `ci_mix` 改为 CI 内实时生成密钥，不再下载旧 tarball
- `testcase.sh` 的 `config vk` 改为 **运行时从 `./gnark/circuit_*.vk` 读取**，不再硬编码 —— 因为 groth16.Setup 是随机的，每次生成 PK/VK 都不同，硬编码 VK 与 CI 新生成 PK 必然不匹配（`pairing doesn't match`）
- 已本地验证：新 PK 可被 v0.9.0 读取，deposit prove+verify 往返通过（off-chain mimc 与旧协议 hash 一致）

**影响**：链上已部署的 mix VK 需重新生成部署（`setVerifyKey`）。`executor/zksnark` 测试预置的旧格式 VK 仍无法读取，相关测试保持 `t.Skip`。

## 测试状态

- `mix/executor`、`merkletree`、`types`、`wallet`：全部通过
- `mix/executor/zksnark`：6 个测试 skip（旧 groth16 VK 格式）

## 溯源参考

- 总体升级说明：`docs/chain33-go-ethereum-v1.14.8-upgrade.md`
- 旧 MiMC 实现：`plugin/crypto/legacymimc`

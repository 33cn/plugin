# RGBX 配置说明手册（运维/测试）

本文档仅说明配置项及含义，覆盖：

- Chain33 主链配置
- Chain33 平行链配置
- Neutrino 子配置（`rpc.sub.light.neutrino`）
- Bitcoin 节点关键配置

## 1. 配置文件范围

建议按职责拆分为三类配置：

- 主链配置（示例：`chain33.toml`）
- 平行链配置（示例：`chain33.para.toml`）
- BTC 节点配置（示例：`bitcoin.conf` 或等价启动参数）

其中跨链相关核心段落在主链和平行链里分别是：

- 主链：`[exec.sub.lightclient]`、`[exec.sub.rgbx]`
- 平行链：`[rpc.sub.light]`、`[rpc.sub.light.neutrino]`、`[rpc.sub.light.neutrino.btcRPC]`、`[rpc.sub.light.neutrino.tss]`

## 2. Chain33 主链配置项

### 2.1 `[exec.sub.lightclient]`

- `btcNetName`
  - 含义：BTC 网络类型
  - 常用值：`regtest`、`testnet`、`mainnet`
- `commitAddress`
  - 含义：提交 lightclient 相关交易的授权地址
  - 要求：应为受控地址，且与运维密钥管理策略一致
- `allowRegtestTimeWarp`
  - 含义：仅用于 regtest 测试场景的时间容错开关
  - 建议：仅在 regtest 打开，生产网络关闭

### 2.2 `[exec.sub.rgbx]`

- `commitAddress`
  - 含义：提交 RGBX 关键交易（如确认交易）的授权地址
  - 要求：应与对应签名私钥匹配
- `crossChainAssetPrefix`
  - 含义：跨链映射资产前缀
  - 示例：`X`，则 BTC 映射资产为 `XBTC`
- `guardianParachainTitle`
  - 含义：受信平行链标题（para title）
  - 要求：与平行链 `Title` 完全一致

## 3. Chain33 平行链配置项

### 3.1 平行链基础配置（按模块归属）

- 根级：`Title`
  - 含义：平行链标题
  - 要求：与主链 `guardianParachainTitle` 一致
- `[rpc.parachain]`：`mainChainGrpcAddr`
  - 含义：主链 gRPC 地址
  - 用途：平行链访问主链查询/提交接口
- `[consensus.sub.para]`：`authAccount`
  - 含义：平行链共识节点账户
  - 用途：节点身份与权限校验（跨链节点需与钱包私钥对应）
- `[crypto]`：`enableTSS`
  - 含义：是否开启 TSS 功能
  - 建议：跨链节点必须开启
- `[p2p]`：`types`、`enable`、`waitPid`
  - 含义：P2P 网络与 DHT 开关配置
  - 建议：开启 DHT（`types=["dht"]`、`enable=true`），保证 TSS 节点发现与消息分发
- `[p2p.sub.dht]`：`DHTDataPath`
  - 含义：DHT 数据目录
  - 建议：配置稳定持久化路径，避免重启后频繁重建邻居关系

### 3.2 `[rpc.sub.light]`

- `clients`
  - 含义：light client 插件列表
  - RGBX 场景固定为：`["neutrino"]`
- `commitAddr`
  - 含义：light RPC 使用的提交地址
  - 要求：与对应提交私钥匹配

## 4. Neutrino 配置项（平行链）

以下字段定义来源于 `plugin/dapp/lightclient/rpc/lightclient/neutrino/config.go`。

### 4.1 `[rpc.sub.light.neutrino]`

- `isOfficialNode` (bool)
  - `true`：启动官方流程（BTC 同步、充值监听、提现处理、确认提交）
  - `false`：不执行官方提交流程，通常用于验证/协作节点
- `maxPeer` (int)
  - 含义：Neutrino 最大对等节点数
  - 默认行为：小于 1 时使用默认值（实现内为 8）
- `blockCacheSize` (uint64)
  - 含义：区块缓存大小（字节）
  - 默认行为：小于 1MB 时使用默认值（约 20MB）
- `netName` (string)
  - 含义：BTC 网络名
  - 要求：与主链 lightclient 的 `btcNetName` 一致
- `addPeers` ([]string)
  - 含义：启动后主动增加的 P2P 节点列表
- `connectPeers` ([]string)
  - 含义：固定连接节点列表
  - 说明：配置后通常只连该列表，不再自动找出站节点
- `btcBlockInterval` (uint32)
  - 含义：BTC 区块同步轮询间隔基准（秒）
  - 默认行为：未设置时采用实现默认值
- `blockConfirmations` (uint32)
  - 含义：BTC 确认数门限
  - 影响：区块头提交、pending 拉取、充值/提现确认时机
- `btcHeaderStartHeight` (uint64)
  - 含义：首次提交 BTC header 的起始高度
  - 默认行为：未设置时从 1 开始
- `maxUtxoRescanTime` (int64)
  - 含义：UTXO 重扫超时（单位：小时）
  - 特性：0 表示不超时；内部会转为秒

### 4.2 `[rpc.sub.light.neutrino.btcRPC]`

- `host`
  - 含义：BTC RPC 地址（host:port）
- `user`
  - 含义：BTC RPC 用户名
- `pass`
  - 含义：BTC RPC 密码
- `mode`
  - 含义：RPC 连接模式，支持 `ws`（默认）或 `http`
- `disableTLS` (bool)
  - 含义：是否禁用 TLS
  - 建议：生产环境使用 TLS
- `certFile`
  - 含义：TLS 证书路径（可选）
  - 要求：启用 TLS 时路径可读且证书与 `host` 匹配

### 4.3 `[rpc.sub.light.neutrino.tss]`

- `peers` ([]string)
  - 含义：TSS 参与节点地址列表
  - 要求：所有参与节点配置一致
- `threshold` (uint32)
  - 含义：阈值签名门限（t-of-n 中的 t）
  - 建议：按容错策略设置，且不大于节点总数
- `rank` (uint32)
  - 含义：节点角色标识（用于区分官方/验证角色）
  - 要求：同一节点在全网配置必须稳定一致

## 5. Bitcoin 节点关键配置项

以下示例使用 `bitcoin.conf` 格式说明关键配置（btcd/bitcoin-core 参数名有差异时，以节点实现文档为准）：

```ini
# 网络（与 Chain33 的 btcNetName/netName 保持一致）
regtest=1
# testnet=1
# mainnet 默认不需要显式开启

# RPC 监听地址（需保证 Neutrino 的 btcRPC.host 可达）
rpcbind=0.0.0.0
rpcport=18443

# RPC 认证（需与 rpc.sub.light.neutrino.btcRPC.user/pass 一致）
rpcuser=root
rpcpassword=1314

# TLS / 证书（若启用 TLS，证书路径需与 Neutrino certFile 对应）
# 常见做法是由节点自动生成证书；若禁用 TLS 则需与 disableTLS=true 匹配
# rpcssl=1
# rpcsslcertificatechainfile=/path/to/rpc.cert

# 交易与地址索引（建议开启）
txindex=1
addrindex=1

# 过滤能力（支持neutrino轻节点）
blockfilterindex=1
peerblockfilters=1
```

字段对照关系：

- `rpc.sub.light.neutrino.btcRPC.host` <-> `rpcbind/rpcport`
- `rpc.sub.light.neutrino.btcRPC.user` <-> `rpcuser`
- `rpc.sub.light.neutrino.btcRPC.pass` <-> `rpcpassword`
- `rpc.sub.light.neutrino.btcRPC.disableTLS/certFile` <-> TLS/证书配置
- `rpc.sub.light.neutrino.netName` <-> `regtest/testnet/mainnet` 网络选择

补充说明（参考 Bitcoin Core Neutrino 模式文档）：

- Bitcoin Core 需支持 BIP157/BIP158（常见要求为 0.21.0+）
- 首次开启 `blockfilterindex=1` 后，节点会重建过滤索引，过程可能较慢
- 如果不希望节点对外自动发现，可在 `bitcoin.conf` 配置 `discover=0`

## 6. TSS 组网说明（当前实现约束）

当前实现建议采用：**1 个官方节点 + N 个第三方节点**（N >= 2，且满足阈值）。

- 官方节点：
  - `rpc.sub.light.neutrino.isOfficialNode=true`
  - 负责发起业务主流程（BTC 同步、充值监听、提现处理、确认提交）
- 第三方节点：
  - `rpc.sub.light.neutrino.isOfficialNode=false`
  - 参与 DKG 和签名协作，不发起官方提交流程
- 角色约定（实践中）：
  - 官方节点 `rank=0`
  - 第三方节点 `rank=1`
- 全体节点必须保持一致：
  - `rpc.sub.light.neutrino.tss.peers`
  - `rpc.sub.light.neutrino.tss.threshold`
- 组网依赖：
  - 平行链需启用 DHT P2P（`[p2p] types=["dht"]` 且 `enable=true`）

## 7. 配置一致性检查清单

上线/联调前建议逐项核对：

- `btcNetName == netName == BTC 节点 network`
- 主链 `guardianParachainTitle == 平行链 Title`
- `commitAddress/commitAddr/authAccount` 与私钥管理匹配
- `blockConfirmations` 符合环境安全要求（测试可低，生产应高）
- `btcRPC.host/user/pass/TLS` 与 BTC 节点一致
- 全部 TSS 节点的 `peers/threshold` 一致，且 `rank` 分配无冲突

## 8. 最小示例（仅配置片段）

```toml
# main
[exec.sub.lightclient]
btcNetName="regtest"
commitAddress="1xxxxxxxxxxxxxxxx"
allowRegtestTimeWarp=true

[exec.sub.rgbx]
commitAddress="1xxxxxxxxxxxxxxxx"
crossChainAssetPrefix="X"
guardianParachainTitle="user.p.rgbx."

# para root-level
Title="user.p.rgbx."

[p2p]
types=["dht"]
enable=true
waitPid=false

[p2p.sub.dht]
DHTDataPath="paradatadir/p2pstore"

[rpc.parachain]
mainChainGrpcAddr="127.0.0.1:8802"

[consensus.sub.para]
authAccount="1xxxxxxxxxxxxxxxx"

[crypto]
enableTSS=true

[rpc.sub.light]
clients=["neutrino"]
commitAddr="1xxxxxxxxxxxxxxxx"

[rpc.sub.light.neutrino]
isOfficialNode=true
netName="regtest"
connectPeers=["127.0.0.1:18444"]
btcBlockInterval=2
blockConfirmations=1
maxUtxoRescanTime=60

[rpc.sub.light.neutrino.btcRPC]
host="127.0.0.1:18443"
user="root"
pass="1314"
disableTLS=false
certFile="/path/to/rpc.cert"

[rpc.sub.light.neutrino.tss]
peers=["1addrA","1addrB","1addrC","1addrD"]
threshold=3
rank=0 # 官方节点；第三方节点配置为 rank=1，且 isOfficialNode=false
```

第三方节点相对官方节点的最小差异：

- `[rpc.sub.light.neutrino] isOfficialNode=false`
- `[rpc.sub.light.neutrino.tss] rank=1`
- 节点自身 `authAccount` / `commitAddr` 使用本节点受控地址
- `peers` 与 `threshold` 必须与官方节点保持一致

## 9. 常见配置错误

- 网络不一致：`netName` 与 BTC 实际网络不一致导致校验失败
- TLS 不匹配：启用 TLS 但 `certFile` 错误或证书主机名不匹配
- TSS 不一致：节点间 `peers/threshold` 不一致导致 DKG 或签名异常
- 多官方节点：多个 `isOfficialNode=true` 节点并行处理，导致重复提交/状态竞争
- 确认数过低：测试通过但生产抗重组能力不足
- 授权地址不匹配：`commitAddress/commitAddr/authAccount` 与私钥不对应

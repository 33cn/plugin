# 平行链扩容

即平行链自主打包出块, 并通过rollup只将状态数据提交至主链存证


## 功能
rollup 涉及功能介绍

### 交易打包
- 平行链需要配置挖矿共识, 并自主打包交易, 产生区块
- 交易需要直接发送至平行链端, 否则无法被打包

### 状态提交
状态提交包括全量提交和精简提交模式, 同时包含超时提交及分段提交策略

#### 精简提交
- 验证者节点对本地的区块头数据共识, 并提交至主链rollup合约存证
- 默认每32个区块触发一次提交


#### 全量提交
- 验证者节点对本地的区块及交易数据共识, 并提交至主链rollup合约存证
- 默认至少收集128笔交易触发一次提交, 该种模式下交易必须采用bls算法签名
- 建议crypto配置中限制签名算法, 避免误发送非bls签名的交易

#### 超时提交
- 本地一直不生成新区块, 且存在未提交区块数据时, 触发超时提交
- 配置项maxCommitInterval指定

#### 分段提交
- 仅在全量提交中会触发
- 单个区块交易过多, 单次提交数据超过最大交易容量, 触发分段

### 资产跨链
- rollup模式兼容已有的资产跨链转账, 但跨链结算需要等到状态提交后
- 跨链交易会被转发到主链优先执行, 主链执行后由平行链自动拉取到本地执行
- 如果交易组中包含有跨链交易, 则交易组判定为跨链交易

 
### 区块同步
- 支持节点间基于p2p区块同步
- 配置项isParaChain需要置为false


## 配置

### 配置文件

基于已有平行链配置文件做调整
- 自主挖矿打包出块, 为已有配置改动, 参照联盟链节点, 包括mempool, 共识, 以及p2p等
- 状态提交, 为新增配置, 即rollup关联配置

```toml
[blockchain]
# 配置为false或删除, 该配置用于经典平行链模式, 会限制节点间相互同步
isParaChain = false


[mempool]
# 配置para以外的mempool插件
name="timeline"
poolCacheSize=10240
# 最低交易费率, 根据需要配置
minTxFeeRate=0


# p2p功能需开启, 建议使用dht插件
[p2p]
types=["dht"]
enable=true

[p2p.sub.dht]
port=13803
DHTDataPath="paradatadir/p2pstore"


# 平行链相关的rpc配置
[rpc.parachain]
#主链节点的grpc服务器ip，当前可以支持多ip负载均衡，如“118.31.177.1:8802,39.97.2.127:8802”
mainChainGrpcAddr="localhost:8802"
# 平行链跨链交易需要转发到主链
forwardExecs=["paracross"]
forwardActionNames=["crossAssetTransfer"]


[consensus]
# 挖矿共识, 平行链需要自主打包, 需配置联盟链类型共识, 如tendermint, pbft等, 测试可用solo/raft
name="solo"
# 提交共识, 验证节点需要配置, 删除该配置即可关闭rollup功能
committer="rollup"

[consensus.sub.rollup]

# 配置节点账户私钥, 隐秘性不好, 用于测试
authKey=""
# 当authKey未配置时, 支持配置节点账户地址 
# 即从钱包中获取对应的私钥, 节点需要创建钱包并解锁导入对应的私钥 
authAccount=""

# 全量数据即提交所有交易源数据, 此时交易签名必须为bls类型
# 默认关闭, 即仅提交区块header数据
fullDataCommit=false
# 最大状态提交间隔时长
# 最低60s, 默认5min
maxCommitInterval=60 #seconds
# 设置平行链启动时对应的主链高度
startHeight=0
# 同步主链区块头, 预留高度, 减少回滚概率 
# 默认12, 最低设为1
reservedMainHeight=12

[fork.system]
ForkBlockHash= 0
ForkRootHash=0
```
                                                                  

      

### 部署配置
- 主链开启rollup合约
- 验证节点主链质押, 对应平行链的NodeGroup配置
- 节点共识提交采用bls签名, 在配置NodeGroup时需要指定节点的bls公钥信息

```
# 命令行基于验证节点私钥生成对应的bls公钥信息
./cli para bls pub -p <nodeAuthKey>
 
# 主链 apply node group构建交易
./cli para nodegroup apply --paraName=<paraTitle> -a <nodeAuthAddr> -p <nodeBlsPub> -c <frozenAmount>

# 主链 approve node group构建交易, applyID即apply交易的哈希
./cli para nodegroup approve --paraName=<paraTitle> -c <frozenAmount> -i <applyID> 

```

[NodeGroup相关文档](https://chain.33.cn/document/134)

   
## 命令行


```
# 查看平行链rollup状态
./cli rollup status -t <paraTitle>

# 查看单轮提交的汇总信息
./cli rollup round -t <paraTitle> -r <roundNum>

# 查看验证者bls公钥信息
./cli rollup validator -t <paraTitle>

```


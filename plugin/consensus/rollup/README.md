## 平行链扩容

即平行链自主打包出块, 并通过rollup只将状态数据提交至主链存证


### 配置

#### 平行链配置文件

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


[consensus]
# 挖矿共识, 平行链需要自主打包, 需配置联盟链类型共识, 如tendermint, pbft等, 测试可用solo/raft
name="solo"
# 提交共识, 区别于挖矿共识, 用于平行链向主链进行状态提交, 目前仅支持rollup插件
committer="rollup"

[consensus.sub.rollup]
# secp256k1私钥, 用于签名commit交易
commitTxKey=""
# bls私钥, 用于验证者共识签名
validatorBlsKey=""
# 全量数据即提交所有交易源数据, 此时交易签名必须为bls类型
# 默认关闭, 即仅提交区块header数据
fullDataCommit=false
# 最大状态提交间隔
# 最低60s, 默认5min
maxCommitInterval=60 #seconds
# 设置平行链启动时对应的主链高度
startHeight=0
```
                                                                  

      

#### 部署配置

rollup模式下平行链依赖主链质押授权, 对应已有平行链的NodeGroup配置

在申请nodegroup时需要设置已配置的共识节点bls公钥

可以参考[相关文档](https://chain.33.cn/document/134)




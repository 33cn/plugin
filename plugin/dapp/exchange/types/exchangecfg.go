package types

var cfgstring = `
# Title为local，表示此配置文件为本地单节点的配置。此时本地节点所在的链上只有这一个节点，共识模块一般采用solo模式。
Title="local"
TestNet=true
FixTime=false
[crypto]
[log]
# 日志级别，支持debug(dbug)/info/warn/error(eror)/crit
loglevel = "info"
logConsoleLevel = "info"
# 日志文件名，可带目录，所有生成的日志文件都放到此目录下
logFile = "logs/chain33.log"
# 单个日志文件的最大值（单位：兆）
maxFileSize = 300
# 最多保存的历史日志文件个数
maxBackups = 100
# 最多保存的历史日志消息（单位：天）
maxAge = 28
# 日志文件名是否使用本地时间（否则使用UTC时间）
localTime = true
# 历史日志文件是否压缩（压缩格式为gz）
compress = true
# 是否打印调用源文件和行号
callerFile = false
# 是否打印调用方法
callerFunction = false
[blockchain]
# 缓存区块的个数
defCacheSize=128
# 同步区块时一次最多申请获取的区块个数
maxFetchBlockNum=128
# 向对端节点请求同步区块的时间间隔
timeoutSeconds=5
# 使用的数据库类型
driver="leveldb"
# 数据库文件目录
dbPath="datadir"
# 数据库缓存大小
dbCache=64
# 是否为单节点
singleMode=true
# 同步区块批量写数据库时，是否需要立即写磁盘，非固态硬盘的电脑可以设置为false，以提高性能
batchsync=false
# 是否记录添加或者删除区块的序列，若节点作为主链节点，为平行链节点提供服务，需要设置为true
isRecordBlockSequence=true
# 是否为平行链节点
isParaChain=false
# 是否开启交易快速查询索引
enableTxQuickIndex=false
[p2p]
types=["dht"]
msgCacheSize=10240
driver="leveldb"
dbPath="datadir/addrbook"
dbCache=4
grpcLogFile="grpc33.log"
[rpc]
# jrpc绑定地址
jrpcBindAddr="localhost:8801"
# grpc绑定地址
grpcBindAddr="localhost:8802"
# 白名单列表，允许访问的IP地址，默认是“*”，允许所有IP访问
whitelist=["127.0.0.1"]
# jrpc方法请求白名单，默认是“*”，允许访问所有RPC方法
jrpcFuncWhitelist=["*"]
# jrpc方法请求黑名单，禁止调用黑名单里配置的rpc方法，一般和白名单配合使用，默认是空
# jrpcFuncBlacklist=["xxxx"]
# grpc方法请求白名单，默认是“*”，允许访问所有RPC方法
grpcFuncWhitelist=["*"]
# grpc方法请求黑名单，禁止调用黑名单里配置的rpc方法，一般和白名单配合使用，默认是空
# grpcFuncBlacklist=["xxx"]
# 是否开启https
enableTLS=false
# 证书文件，证书和私钥文件可以用cli工具生成
certFile="cert.pem"
# 私钥文件
keyFile="key.pem"
[mempool]
# mempool队列名称，可配，timeline，score，price
name="timeline"
# mempool缓存容量大小，默认10240
poolCacheSize=10240
# 最小得交易手续费用，这个没有默认值，必填，一般是100000
minTxFeeRate=100000
minTxFee=100000
# 每个账户在mempool中得最大交易数量，默认100
maxTxNumPerAccount=10000
# timeline 是默认的先来先进的按时间排序
[mempool.sub.timeline]
# mempool缓存容量大小，默认10240
poolCacheSize=10240
# 最小得交易手续费用，这个没有默认值，必填，一般是100000
minTxFee=100000
# 每个账户在mempool中得最大交易数量，默认100
maxTxNumPerAccount=10000
# score是分数队列模式(分数=常量a*手续费/交易字节数-常量b*时间*定量c，按分数排队，高的优先，常量a，b和定量c可配置)，按分数来排序
[mempool.sub.score]
# mempool缓存容量大小，默认10240
poolCacheSize=10240
# 最小得交易手续费用，这个没有默认值，必填，一般是100000
minTxFee=100000
# 每个账户在mempool中得最大交易数量，默认100
maxTxNumPerAccount=10000
# 时间占价格比例
timeParam=1
# 手续费相对于时间的一个合适的常量，取当前unix时间戳前四位数，排队时手续费高1e-5的分数~=快1s的分数
priceConstant=1544
# 常量比例
pricePower=1
# price是价格队列模式(价格=手续费/交易字节数，价格高者优先，同价则时间早优先)
[mempool.sub.price]
# mempool缓存容量大小，默认10240
poolCacheSize=10240
# 最小得交易手续费用，这个没有默认值，必填，一般是100000
minTxFee=100000
# 每个账户在mempool中得最大交易数量，默认100
maxTxNumPerAccount=10000
[consensus]
#共识名,可选项有solo,ticket,raft,tendermint,para
name="solo"
#是否开启挖矿,开启挖矿才能创建区块
minerstart=true
#创世区块时间(UTC时间)
genesisBlockTime=1514533394
#创世交易地址
genesis="1McRuWgLN7daKTFfZyibb5eK769NVgNH22"
[mver.consensus]
#基金账户地址
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
#用户回报
coinReward = 18
#发展基金回报
coinDevFund = 12
#ticket价格
ticketPrice = 10000
#挖矿难度
powLimitBits = "0x1f00ffff"
#每次调整难度的最大的范围，如果设置成 4 的话，范围是 (1/4 - 4)，一次不能增加 4倍以上的难度，或者难度减少为 原来的 1/4 ，这个参数，是为了难度不会突然爆增加或者减少
retargetAdjustmentFactor = 4
#表示如果区块时间大于当前时间 16s ，那么区块就会判定为无效区块。
futureBlockTime = 16
#ticket冻结时长
ticketFrozenTime = 5    #5s only for test
ticketWithdrawTime = 10 #10s only for test
ticketMinerWaitTime = 2 #2s only for test
#区块包含最多交易数
maxTxNumber = 1600      #160
#调整挖矿难度的间隔，(ps:难度不是每个区块都调整的，而是每隔 targetTimespan / targetTimePerBlock 块调整一次)
targetTimespan = 2304
#每个区块打包的目标时间
targetTimePerBlock = 16
# 仅保留这一项，其他consensus相关的配置全部删除
[consensus.sub.solo]
#创世交易地址
genesis="1McRuWgLN7daKTFfZyibb5eK769NVgNH22"
#创世区块时间(UTC时间)
genesisBlockTime=1514533394
#获取交易间隔时长,单位纳秒
waitTxMs=10
[store]
# 数据存储格式名称，目前支持mavl,kvdb,kvmvcc,mpt
name="mavl"
# 数据存储驱动类别，目前支持leveldb,goleveldb,memdb,gobadgerdb,ssdb,pegasus
driver="leveldb"
# 数据文件存储路径
dbPath="datadir/mavltree"
# Cache大小
dbCache=128
# 数据库版本
localdbVersion="1.0.0"
[store.sub.mavl]
# 是否使能mavl加前缀
enableMavlPrefix=false
# 是否使能MVCC,如果mavl中enableMVCC为true此处必须为true
enableMVCC=false
# 是否使能mavl数据裁剪
enableMavlPrune=false
# 裁剪高度间隔
pruneHeight=10000
[wallet]
# 交易发送最低手续费，单位0.00000001BTY(1e-8),默认100000，即0.001BTY
minFee=100000
# walletdb驱动名，支持leveldb/memdb/gobadgerdb/ssdb/pegasus
driver="leveldb"
# walletdb路径
dbPath="wallet"
# walletdb缓存大小
dbCache=16
# 钱包发送交易签名方式
signType="secp256k1"
[wallet.sub.ticket]
# 是否关闭ticket自动挖矿，默认false
minerdisable=false
# 允许购买ticket挖矿的白名单地址，默认配置“*”，允许所有地址购买
minerwhitelist=["*"]
[exec]
#执行器执行是否免费
isFree=false
#执行器执行所需最小费用,低于Mempool和Wallet设置的MinFee,在minExecFee = 0 的情况下，isFree = true才会生效
minExecFee=100000
#是否开启stat插件
enableStat=false
#是否开启MVCC插件
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]
[exec.sub.token]
#是否保存token交易信息
saveTokenTxList=true
#token审批人地址
tokenApprs = [
    "1McRuWgLN7daKTFfZyibb5eK769NVgNH22",
    "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK",
    "1LY8GFia5EiyoTodMLfkB5PHNNpXRqxhyB",
    "1GCzJDS6HbgTQ2emade7mEJGGWFfA15pS9",
    "1JYB8sxi4He5pZWHCd3Zi2nypQ4JMB6AxN",
    "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
]
[exec.sub.cert]
# 是否启用证书验证和签名
enable=false
# 加密文件路径
cryptoPath="authdir/crypto"
# 带证书签名类型，支持"auth_ecdsa", "auth_sm2"
signType="auth_ecdsa"
[exec.sub.relay]
#relay执行器保存BTC头执行权限地址
genesis="1McRuWgLN7daKTFfZyibb5eK769NVgNH22"
[exec.sub.manage]
#manage执行器超级管理员地址
superManager=[
    "1McRuWgLN7daKTFfZyibb5eK769NVgNH22",
    "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
    "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
]
[metrics]
#是否使能发送metrics数据的发送
enableMetrics=false
#数据保存模式
dataEmitMode="influxdb"
[metrics.sub.influxdb]
#以纳秒为单位的发送间隔
duration=1000000000
url="http://influxdb:8086"
database="chain33metrics"
username=""
password=""
namespace=""

#exchange合约相关配置
[mver.exec.sub.exchange]
#银行帐户列表（现第一个地址用来收取手续费）
banks = [
    "1PTGVR7TUm1MJUH7M1UNcKBGMvfJ7nCrnN"
]
#机器人帐户列表
robots = [
    "1Nq8MDDVqSsS2zQeEZZa7bH53b9vBuUmEW"
]
#币种配置，
#coin   转入exchange合约的币种名称
#execer 转入exchange合约的币种执行器名称
#name   执行器币种的别称
coins = [
    {coin = "bty", execer = "coins", name = "BTY"},
    {coin = "CCNY", execer = "token", name = "CCNY"},

    {coin = "coins.bty", execer = "paracross", name = "BTY"},
    {coin = "YCC", execer = "evmxgo", name = "YCC"},
    {coin = "ETH", execer = "evmxgo", name = "ETH"},
    {coin = "USDT", execer = "evmxgo", name = "USDT"}
]
#现货交易配置
#symbol 币种对；priceDigits 价格最小位数；amountDigits 数量最小位数； minFee  最小手续费
#taker  吃单手续费率,配置时需*1e8(如：收取每笔交易百分之一的手续费，maker=1000000)未配置交易对默认为100000
#maker  挂单手续费,配置时需*1e8(如：收取每笔交易千分之一的手续费，taker=100000)未配置交易对默认为100000
exchanges = [
    {symbol = "BTY_CCNY", priceDigits = 4, amountDigits = 1, taker = 1000000, maker = 100000,  minFee = 0},

    {symbol = "BTY_USDT", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0},
    {symbol = "YCC_USDT", priceDigits = 5, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0},
    {symbol = "ETH_USDT", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0}
]


[fork.sub.exchange]
Enable=0
ForkFix1=0
ForkParamV1 = 0
ForkParamV2 = 0
ForkParamV3 = 0
ForkParamV4 = 0
ForkParamV5 = 0
ForkParamV6 = 0
ForkParamV7 = 0
ForkParamV8 = 0
ForkParamV9 = 0
ForkParamV10 = 0
ForkParamV11 = 0
ForkParamV12 = 0
ForkParamV13 = 0
ForkParamV14 = 0
ForkParamV15 = 0
ForkParamV16 = 0
ForkParamV17 = 0
ForkParamV18 = 0
ForkParamV19 = 0
ForkParamV20 = 0
ForkParamV21 = 0
ForkParamV22 = 0
ForkParamV23 = 0
ForkParamV24 = 0
ForkParamV25 = 0
ForkParamV26 = 0
ForkParamV27 = 0
ForkParamV28 = 0
ForkParamV29 = 0


`

var defaultCfgstring = `
Title="local"
TestNet=true
FixTime=false
TxHeight=true
CoinSymbol="bty"
ChainID=33

# crypto模块配置
[crypto]
enableTypes=[]    #设置启用的加密插件名称，不配置启用所有
[crypto.enableHeight]  #配置已启用插件的启用高度，不配置采用默认高度0， 负数表示不启用
secp256k1=0
[crypto.sub.secp256k1] #支持插件子配置

[log]
# 日志级别，支持debug(dbug)/info/warn/error(eror)/crit
loglevel = "debug"
logConsoleLevel = "info"
# 日志文件名，可带目录，所有生成的日志文件都放到此目录下
logFile = "logs/chain33.log"
# 单个日志文件的最大值（单位：兆）
maxFileSize = 300
# 最多保存的历史日志文件个数
maxBackups = 100
# 最多保存的历史日志消息（单位：天）
maxAge = 28
# 日志文件名是否使用本地事件（否则使用UTC时间）
localTime = true
# 历史日志文件是否压缩（压缩格式为gz）
compress = true
# 是否打印调用源文件和行号
callerFile = false
# 是否打印调用方法
callerFunction = false

[blockchain]
defCacheSize=128
maxFetchBlockNum=128
timeoutSeconds=5
batchBlockNum=128
driver="leveldb"
dbPath="datadir"
dbCache=64
isStrongConsistency=false
singleMode=true
batchsync=false
isRecordBlockSequence=true
isParaChain=false
enableTxQuickIndex=true
txHeight=true

# 使能精简localdb
enableReduceLocaldb=false
# 关闭分片存储,默认false为开启分片存储;平行链不需要分片需要修改此默认参数为true
disableShard=false
# 分片存储中每个大块包含的区块数
chunkblockNum=1000
# 使能从P2pStore中获取数据
enableFetchP2pstore=false
# 使能假设已删除已归档数据后,获取数据情况
enableIfDelLocalChunk=false

enablePushSubscribe=true
maxActiveBlockNum=1024
maxActiveBlockSize=100


[p2p]
enable=false
driver="leveldb"
dbPath="datadir/addrbook"
dbCache=4
grpcLogFile="grpc33.log"

[rpc]
jrpcBindAddr="localhost:0"
grpcBindAddr="localhost:0"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]

[rpc.parachain]
mainChainGrpcAddr="localhost:8802"

[mempool]
name="timeline"
poolCacheSize=102400
# 最小得交易手续费率，这个没有默认值，必填，一般是0.001 coins
minTxFeeRate=100000
maxTxNumPerAccount=100000

[consensus]
name="solo"
minerstart=true
genesisBlockTime=1514533394
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
minerExecs=["ticket", "autonomy"]

[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
powLimitBits = "0x1f00ffff"
maxTxNumber = 10000

[mver.consensus.ForkChainParamV1]
maxTxNumber = 10000

[mver.consensus.ForkChainParamV2]
powLimitBits = "0x1f2fffff"

[mver.consensus.ForkTicketFundAddrV1]
fundKeyAddr = "1Ji3W12KGScCM7C2p8bg635sNkayDM8MGY"

[mver.consensus.ticket]
coinReward = 18
coinDevFund = 12
ticketPrice = 10000
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 5
ticketWithdrawTime = 10
ticketMinerWaitTime = 2
targetTimespan = 2304
targetTimePerBlock = 16

[mver.consensus.ticket.ForkChainParamV1]
targetTimespan = 288 #only for test
targetTimePerBlock = 2

[consensus.sub.para]
#主链指定高度的区块开始同步
startHeight=345850
#打包时间间隔，单位秒
writeBlockSeconds=2
#主链每隔几个没有相关交易的区块，平行链上打包空区块
emptyBlockInterval=50
#验证账户，验证节点需要配置自己的账户，并且钱包导入对应种子，非验证节点留空
authAccount=""
#等待平行链共识消息在主链上链并成功的块数，超出会重发共识消息，最小是2
waitBlocks4CommitMsg=2
searchHashMatchedBlockDepth=100

[consensus.sub.solo]
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
genesisBlockTime=1514533394
waitTxMs=1

[consensus.sub.ticket]
genesisBlockTime=1514533394
[[consensus.sub.ticket.genesis]]
minerAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
returnAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
count=10000

[[consensus.sub.ticket.genesis]]
minerAddr="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
returnAddr="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
count=1000

[[consensus.sub.ticket.genesis]]
minerAddr="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
returnAddr="1KcCVZLSQYRUwE5EXTsAoQs9LuJW6xwfQa"
count=1000

[store]
name="mavl"
driver="leveldb"
dbPath="datadir/mavltree"
dbCache=128

[store.sub.mavl]
enableMavlPrefix=false
enableMVCC=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
dbCache=16
signType="secp256k1"
coinType="bty"

[wallet.sub.ticket]
minerdisable=false
minerwhitelist=["*"]

[exec]
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]
[exec.sub.coins]

[exec.sub.token]
saveTokenTxList=true
tokenApprs = [
	"1Bsg9j6gW83sShoee1fZAt9TkUjcrCgA9S",
	"1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK",
	"1LY8GFia5EiyoTodMLfkB5PHNNpXRqxhyB",
	"1GCzJDS6HbgTQ2emade7mEJGGWFfA15pS9",
	"1JYB8sxi4He5pZWHCd3Zi2nypQ4JMB6AxN",
	"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
]

[exec.sub.relay]
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

[exec.sub.cert]
# 是否启用证书验证和签名
enable=false
# 加密文件路径
cryptoPath="authdir/crypto"
# 带证书签名类型，支持"auth_ecdsa", "auth_sm2"
signType="auth_ecdsa"

[exec.sub.manage]
superManager=[
    "1Bsg9j6gW83sShoee1fZAt9TkUjcrCgA9S", 
    "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", 
    "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
]
autonomyExec=""

[exec.sub.autonomy]
total="16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
useBalance=false

[exec.sub.jvm]
jdkPath="../../../../build/j2sdk-image"
`

//GetDefaultCfgstring ...
func GetDefaultCfgstring() string {
	return cfgstring
	//return defaultCfgstring
}

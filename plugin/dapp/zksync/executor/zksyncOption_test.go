package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/types"
	"github.com/stretchr/testify/mock"

	"github.com/33cn/chain33/util"
	"github.com/33cn/plugin/plugin/dapp/evm/executor/vm/common"
	zt "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/stretchr/testify/assert"
)

func TestZksyncOption(t *testing.T) {
	dir, statedb, localdb := util.CreateTestDB()
	defer util.CloseTestDB(dir, statedb)
	/*************************deposit*************************/

	cfg := types.NewChain33Config(cfgstring)
	api := new(mocks.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(cfg, nil)

	action := &Action{localDB: localdb, statedb: statedb, height: 1, index: 0, fromaddr: "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv", api: api}
	deposit := &zt.ZkDeposit{
		TokenId:     1,
		Amount:      "100000000",
		EthAddress:  "abcd68033A72978C1084E2d44D1Fa06DdC4A2d58",
		Chain33Addr: getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec"),
	}
	receipt, err := action.Deposit(deposit)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	var zklog zt.ZkReceiptLog
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	assert.Equal(t, nil, err)
	ethFeeAddr, chain33FeeAddr := getCfgFeeAddr(cfg)
	info, err := generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	leaf, err := GetLeafByAccountId(statedb, 2, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)

	/*************************setPubKey*************************/
	_, err = generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)
	setPubKey := &zt.ZkSetPubKey{
		AccountId: 2,
		PubKey: &zt.ZkPubKey{
			X: privateKey.PublicKey.A.X.String(),
			Y: privateKey.PublicKey.A.Y.String(),
		},
	}

	receipt, err = action.SetPubKey(setPubKey)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	/*************************withdraw*************************/
	info, err = generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	withdraw := &zt.ZkWithdraw{
		AccountId: 2,
		TokenId:   1,
		Amount:    "5000",
	}
	msg := wallet.GetWithdrawMsg(withdraw)
	privateKey, err = eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)

	signInfo, err := wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)

	withdraw.Signature = signInfo
	receipt, err = action.Withdraw(withdraw)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}

	tree, err := getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	t.Log(tree)

	leaf, err = GetLeafByAccountId(statedb, 1, info)
	assert.Equal(t, nil, err)
	assert.NotEqual(t, nil, leaf)
	t.Log(leaf)

	token, err := GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "98995000", token.Balance)

	/*************************transferToNew*************************/
	info, err = generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	transferToNew := &zt.ZkTransferToNew{
		FromAccountId:    2,
		TokenId:          1,
		Amount:           "5000",
		ToEthAddress:     "abcd68033A72978C1084E2d44D1Fa06DdC4A2d59",
		ToChain33Address: getChain33Addr("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aed"),
	}
	t.Log(strings.ToLower(transferToNew.ToEthAddress))
	msg = wallet.GetTransferToNewMsg(transferToNew)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	transferToNew.Signature = signInfo
	receipt, err = action.TransferToNew(transferToNew)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "98890000", token.Balance)
	token, err = GetTokenByAccountIdAndTokenId(statedb, 3, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "5000", token.Balance)

	/*************************transfer*************************/
	info, err = generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	transfer := &zt.ZkTransfer{
		FromAccountId: 2,
		TokenId:       1,
		Amount:        "5000",
		ToAccountId:   3,
	}
	msg = wallet.GetTransferMsg(transfer)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	transfer.Signature = signInfo

	receipt, err = action.Transfer(transfer)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "98785000", token.Balance)
	token, err = GetTokenByAccountIdAndTokenId(statedb, 3, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "10000", token.Balance)

	/*************************forceQuit*************************/
	info, err = generateTreeUpdateInfo(statedb, localdb, ethFeeAddr, chain33FeeAddr)
	assert.Equal(t, nil, err)
	forceQuit := &zt.ZkForceExit{
		AccountId: 2,
		TokenId:   1,
	}
	msg = wallet.GetForceExitMsg(forceQuit)
	signInfo, err = wallet.SignTx(msg, privateKey)
	assert.Equal(t, nil, err)
	forceQuit.Signature = signInfo
	receipt, err = action.ForceExit(forceQuit)
	assert.Equal(t, nil, err)
	t.Log(receipt)
	for _, kv := range receipt.GetKV() {
		statedb.Set(kv.GetKey(), kv.GetValue())
	}
	err = types.Decode(receipt.Logs[0].GetLog(), &zklog)
	assert.Equal(t, nil, err)
	for _, kv := range zklog.LocalKvs {
		localdb.Set(kv.GetKey(), kv.GetValue())
	}
	token, err = GetTokenByAccountIdAndTokenId(statedb, 2, 1, info)
	assert.Equal(t, nil, err)
	assert.Equal(t, "0", token.Balance)

	tree, err = getAccountTree(statedb, info)
	assert.Equal(t, nil, err)
	t.Log(tree)
}

func TestEddsa(t *testing.T) {
	privateKey, err := eddsa.GenerateKey(bytes.NewReader(common.FromHex("7266444b7e6408a9ee603de7b73cc8fc168ebf570c7fd482f7fa6b968b6a5aec")))
	assert.Equal(t, nil, err)
	ans := privateKey.PublicKey.Bytes()
	t.Log(privateKey.PublicKey.A.X)
	t.Log(privateKey.PublicKey.A.Y)
	t.Log(ans)
	t.Log(len(ans))
}

func TestBigInt(t *testing.T) {
	byteVal := big.NewInt(0).Bytes()
	stringVal := hex.EncodeToString(byteVal)
	t.Log("bigInt 0 byteVal", byteVal)
	t.Log("bigInt 0 stringVal", stringVal)
	t.Log("0 stringVal", "0")
	t.Log("0 byteVal", []byte("0"))
	t.Log("is equal", stringVal == "0")
}

func TestInitTreeRoot(t *testing.T) {
	eth := zt.HexAddr2Decimal("832367164346888E248bd58b9A5f480299F1e88d")
	chain33 := zt.HexAddr2Decimal("2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a")
	leafs := getInitAccountLeaf(eth, chain33)
	merkleTree := getNewTree()
	for _, l := range leafs {
		merkleTree.Push(getLeafHash(l))
	}
	tree := &zt.AccountTree{
		Index:           2,
		TotalIndex:      2,
		MaxCurrentIndex: 1024,
		SubTrees:        make([]*zt.SubTree, 0),
	}
	for _, subtree := range merkleTree.GetAllSubTrees() {
		tree.SubTrees = append(tree.SubTrees, &zt.SubTree{
			RootHash: subtree.GetSum(),
			Height:   int32(subtree.GetHeight()),
		})
	}
	fmt.Println("len", len(tree.SubTrees), "dec", zt.Byte2Str(tree.SubTrees[len(tree.SubTrees)-1].RootHash))
}

var cfgstring = `
Title="chain33"
TestNet=true
FixTime=false
version="6.3.0"

[crypto]

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
singleMode=false
batchsync=false
isRecordBlockSequence=true
isParaChain=false
enableTxQuickIndex=true
enableReExecLocal=true
# 使能精简localdb
enableReduceLocaldb=true
enablePushSubscribe=false

# 关闭分片存储,默认false为开启分片存储;平行链不需要分片需要修改此默认参数为true
disableShard=false
# 分片存储中每个大块包含的区块数
chunkblockNum=100
# 使能从P2pStore中获取数据
enableFetchP2pstore=true
# 使能假设已删除已归档数据后,获取数据情况
enableIfDelLocalChunk=false

[p2p]
# p2p类型
types=["dht"]
# 是否启动P2P服务
enable=true
# 使用的数据库类型
driver="leveldb"
# 使用的数据库类型
dbPath="datadir/addrbook"
# 数据库缓存大小
dbCache=4
# GRPC请求日志文件
grpcLogFile="grpc33.log"
#waitPid 等待seed导入
waitPid=false


[p2p.sub.gossip]
seeds=[]
isSeed=false
serverStart=true
innerSeedEnable=true
useGithub=true
innerBounds=300
#是否启用ssl/tls 通信，默认不开启
enableTLS=false
#如果需要CA配合认证，则需要配置caCert,caServer
caCert=""
certFile=""
keyFile=""
# ca服务端接口http://ip:port
caServer=""

[p2p.sub.dht]
seeds=[]
port=13803
maxConnectNum=100
# 禁止通过局域网发现节点
disableFindLANPeers=false
# 配置为全节点模式，全节点保存所有分片数据
isFullNode=false
# 分片数据默认保存比例，最低可配置为10
percentage=30

[rpc]
jrpcBindAddr="localhost:8801"
grpcBindAddr="localhost:8802"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]


[mempool]
name="price"
poolCacheSize=10240
maxTxNumPerAccount=100
# 最小得交易手续费率，这个没有默认值，必填，一般是0.001 coins
minTxFeeRate=100000
# 最大的交易手续费率, 0.1 coins
maxTxFeeRate=10000000
# 单笔交易最大的手续费, 10 coins
maxTxFee=1000000000
isLevelFee=true
[mempool.sub.timeline]
poolCacheSize=10240

[mempool.sub.score]
poolCacheSize=10240
timeParam=1      #时间占价格比例
priceConstant=10  #手续费相对于时间的一个的常量,排队时手续费高1e3的分数~=快1h的分数
pricePower=1     #常量比例

[mempool.sub.price]
poolCacheSize=10240

[consensus]
name="ticket"
minerstart=true
genesisBlockTime=1514533394
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
minerExecs=["ticket", "autonomy"]
enableBestBlockCmp=true


[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
powLimitBits="0x1f00ffff"
maxTxNumber = 1600      #160

[mver.consensus.ForkChainParamV1]
maxTxNumber = 1500

[mver.consensus.ForkTicketFundAddrV1]
fundKeyAddr = "1Ji3W12KGScCM7C2p8bg635sNkayDM8MGY"

[mver.consensus.ticket]
coinReward = 18
coinDevFund = 12
ticketPrice = 10000
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 5    #5s only for test
ticketWithdrawTime = 10 #10s only for test
ticketMinerWaitTime = 2 #2s only for test
targetTimespan=2304
targetTimePerBlock=16

[mver.consensus.ticket.ForkChainParamV1]
futureBlockTime = 15
ticketFrozenTime = 43200
ticketWithdrawTime = 172800
ticketMinerWaitTime = 7200
targetTimespan=2160
targetTimePerBlock=15

[mver.consensus.ticket.ForkChainParamV2]
coinReward = 5
coinDevFund = 3
targetTimespan=720
targetTimePerBlock=5
ticketPrice = 3000



[consensus.sub.ticket]
genesisBlockTime=1514533394
[[consensus.sub.ticket.genesis]]
minerAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
returnAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
count=10000

[[consensus.sub.ticket.genesis]]
minerAddr="1PUiGcbsccfxW3zuvHXZBJfznziph5miAo"
returnAddr="1EbDHAXpoiewjPLX9uqoz38HsKqMXayZrF"
count=10000

[[consensus.sub.ticket.genesis]]
minerAddr="1EDnnePAZN48aC2hiTDzhkczfF39g1pZZX"
returnAddr="1KcCVZLSQYRUwE5EXTsAoQs9LuJW6xwfQa"
count=10000

[store]
name="kvmvccmavl"
driver="leveldb"
dbPath="datadir/mavltree"
dbCache=128
# store数据库版本
storedbVersion="2.0.0"

[store.sub.mavl]
enableMavlPrefix=false
enableMVCC=false
enableMavlPrune=false
pruneHeight=10000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000

[store.sub.kvmvccmavl]
# 开启该配置可以方便遍历最新的状态数据，节省磁盘空间可以关闭该配置项
enableMVCCIter=true
enableMavlPrefix=false
enableMVCC=false
enableMavlPrune=false
pruneMavlHeight=10000
# 开启该配置项会精简mvcc历史高度的数据，默认不精简
enableMVCCPrune=false
# 每次精简mvcc的间隔高度，默认每100w高度精简一次
pruneMVCCHeight=1000000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000
# 该参数针对平行链，主链无需开启此功能
enableEmptyBlockHandle=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
dbCache=16
signType="secp256k1"

[wallet.sub.ticket]
minerdisable=false
minerwhitelist=["*"]

[wallet.sub.multisig]
rescanMultisigAddr=false

[exec]
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]
# 记录地址相关的交易列表，便于按地址查询交易
disableAddrIndex=false
# 记录每个高度总的手续费消耗量
disableFeeIndex=false
# 开启后会进一步精简localdb，用户查询合约功能会受影响，纯挖矿节点可以开启节省磁盘空间
disableExecLocal=false

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

[exec.sub.cert]
# 是否启用证书验证和签名
enable=false
# 加密文件路径
cryptoPath="authdir/crypto"
# 带证书签名类型，支持"auth_ecdsa", "auth_sm2"
signType="auth_ecdsa"

[exec.sub.relay]
genesis="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"

[exec.sub.manage]
superManager=[
    "1Bsg9j6gW83sShoee1fZAt9TkUjcrCgA9S",
    "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv",
    "1Q8hGLfoGe63efeWa8fJ4Pnukhkngt6poK"
]
#autonomy执行器名字
autonomyExec="autonomy"

[exec.sub.paracross]
nodeGroupFrozenCoins=0
#平行链共识停止后主链等待的高度
paraConsensusStopBlocks=30000
#配置平行链资产跨链交易的高度列表，title省略user.p,不同title使用,分割，不同hit高度使用"."分割，
#不同ignore范围高度使用"-"分割，hit高度在ignore范围内，为平行链自身的高度，不是主链高度
#para.hit.10.50.250, para.ignore.1-100.200-300
paraCrossAssetTxHeightList=[]

[exec.sub.autonomy]
total="16htvcBNSEA7fZhAdLJphDwQRQJaHpyHTp"
useBalance=false

[exec.sub.evm]
#免交易费模式联盟链允许的最大gas，该配置只对不收取交易费部署方式有效，其他部署方式下该配置不会产生作用
#当前最大为200万
evmGasLimit=2000000
#evm内部调试输出，指令级的，默认关闭,0：关闭；1：打开
evmDebugEnable=0

[exec.sub.mix]
#私对私的交易费,交易比较大，需要多的手续费
txFee=100000000
#私对私token转账，花费token(true)还是BTY(false),
tokenFee=false
#curve H point
pointHX="19172955941344617222923168298456110557655645809646772800021167670156933290312"
pointHY="21116962883761739586121793871108889864627195706475546685847911817475098399811"
#电路最大支持1024个叶子hash，10 level， 配置可以小于1024,但不能大于
maxTreeLeaves=1024
mixApprs=["12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"]

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
#币种配置，
#coin   转入exchange合约的币种名称
#execer 转入exchange合约的币种执行器名称
#name   执行器币种的别称
#minFee最小手续费,配置时需*1e8(如：最小手续费收取1个，minFee=100000000)
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
	{symbol = "YCC_USDT", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0},
	{symbol = "ETH_USDT", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0}
]


[exec.sub.zksync]
manager=[
    "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt",
    "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
]

`

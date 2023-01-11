package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strings"
	"testing"
	"time"

	"github.com/33cn/chain33/account"
	apimock "github.com/33cn/chain33/client/mocks"
	chain33Common "github.com/33cn/chain33/common"
	chain33Crypto "github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/system/crypto/secp256k1"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	zksyncTypes "github.com/33cn/plugin/plugin/dapp/zksync/types"
	"github.com/33cn/plugin/plugin/dapp/zksync/wallet"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type execEnv4perf struct {
	blockTime   int64 // 1539918074
	blockHeight int64
	index       int
	difficulty  uint64

	txHash string
}

var (
	cfgstring4execTest = `
Title="local"
TestNet=true
FixTime=false
TxHeight=true
CoinSymbol="bty"
ChainID=33

[address]
defaultDriver="btc"
[address.enableHeight]
eth=-2

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
[rpc.sub.eth]
enable=false

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
ParaRemoteGrpcClient="localhost:8802"
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

[exec.sub.zkspot]
manager=[
"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt",
"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
]

[mver.exec.sub.zksync]
#银行帐户列表（现第一个地址用来收取手续费）
banks = [
	"1KSBd17H7ZK8iT37aJztFB22XGwsPTdwE4",
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
# USDT = 1, BTY = 2, YCC = 3
exchanges = [
	{symbol = "2_1", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0},
	{symbol = "3_1", priceDigits = 4, amountDigits = 4, taker = 1000000, maker = 100000,  minFee = 0},
]

[exec.sub.zksync]
manager=[
"14KEKbYtKKQm4wMthSK9J4La4nAiidGozt",
"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
]

#运营方配置收交易费地址
#可把二层交易费提取到ETH的地址
ethFeeAddr="832367164346888E248bd58b9A5f480299F1e88d"
#二层的基于zk的chain33地址，注意:非基于sep256k1的普通的chain33地址，而是基于私钥产生的可用于二层的地址
layer2FeeAddr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"
`
	chain33TestCfg    = types.NewChain33Config(strings.Replace(cfgstring4execTest, "Title=\"local\"", "Title=\"chain33\"", 1))
	zksyncHandle      *zksync
	dbHanleGlobal     db.DB
	index             = 0
	dbDir             = ""
	firstUserAccoutID = zksyncTypes.SystemTree2ContractAcctId + 1
)

func initSetup() {
	env := execEnv4perf{
		1539918074,
		0,
		2,
		1539918074,
		"hash",
	}

	dir, dbHanle, kvdb := util.CreateTestDB()
	accB := account.NewCoinsAccount(chain33TestCfg)
	accB.SetDB(kvdb)

	api := new(apimock.QueueProtocolAPI)
	api.On("GetConfig", mock.Anything).Return(chain33TestCfg, nil)
	driver := NewZksync()
	driver.SetAPI(api)
	driver.SetEnv(env.blockHeight, env.blockTime, env.difficulty)
	driver.SetStateDB(kvdb)
	driver.SetLocalDB(kvdb)
	zksyncHandle = driver.(*zksync)
	dbHanleGlobal = dbHanle
	dbDir = dir
}

//++ docker exec build_chain33_1 /root/chain33-cli zksync l2addr -k 19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4
//+ chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1

func TestOutputChain33L2AddrDecimal(t *testing.T) {
	chain33AddrL2 := "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1"
	chain33AddrL2Decimal, _ := zksyncTypes.HexAddr2Decimal(chain33AddrL2)
	fmt.Println("chain33AddrL2Decimal=", chain33AddrL2Decimal)
	//2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a --->
	//20033148478649779061292402960935477249437023394422514689332944628159941947226

	//2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c  --->
	//19449356208766688579807449875624267384186019758574787579222132129615224099980
	//2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1    --->
	//19694183066356799104974294716313078444659172842638956126168373945465009608401
}

func TestDeposit(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	//zksync_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zksync_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, localReceipt, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
}

func TestWithdraw(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	//zksync_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zksync_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, localReceipt, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyWithdrawAction], zksyncTypes.TyWithdrawAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	//测试提币
	receipt, _, err = withdraw(zksyncHandle, acc1privkey, accountID, tokenId, "200")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	withdrawFee := 1000000
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(200)-int64(withdrawFee))
	fmt.Println("Balance is", balance)
	assert.Equal(t, balance, acc4token1Balance.Balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
}

func TestTransfer(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferAction], zksyncTypes.TyTransferAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferToNewAction], zksyncTypes.TyTransferToNewAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//测试向新账户进行转币操作
	toEthAddr := "12a0e25e62c1dbd32e505446062b26aecb65f028"
	toL2Chain33Addr := "2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c"
	//toAddrprivkey := "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
	receipt, _, err = transfer2New(zksyncHandle, acc1privkey, tokenId, accountID, "200", toEthAddr, toL2Chain33Addr)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//继续发送交易
	fromAccountId := accountID
	toAccountId := accountID + 1
	receipt, _, err = transfer(zksyncHandle, acc1privkey, fromAccountId, toAccountId, tokenId, "200")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//确认发送者的balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	tranferFee := 100000
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(200*2)-int64(tranferFee*2))
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//确认接收者的balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), toAccountId, tokenId)
	assert.Nil(t, err)
	toBalance := fmt.Sprintf("%d", 200*2)
	fmt.Println("Balance is", toBalance)
	assert.Equal(t, acc4token1Balance.Balance, toBalance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
}

func TestTransfer2New(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferToNewAction], zksyncTypes.TyTransferToNewAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//测试向新账户进行转币操作
	toEthAddr := "12a0e25e62c1dbd32e505446062b26aecb65f028"
	toL2Chain33Addr := "2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c"
	//toAddrprivkey := "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
	receipt, _, err = transfer2New(zksyncHandle, acc1privkey, tokenId, accountID, "200", toEthAddr, toL2Chain33Addr)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//确认发送者的balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	tranferFee := 100000
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(200)-int64(tranferFee))
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//确认接收者的balance
	toAccountID := accountID + 1
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), toAccountID, tokenId)
	assert.Nil(t, err)
	toBalance := fmt.Sprintf("%d", 200)
	fmt.Println("Balance is", toBalance)
	assert.Equal(t, acc4token1Balance.Balance, toBalance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
}

func TestTree2contract(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	//zksync_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zksync_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	tokenId := uint64(0)

	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	symbol := "ETH"
	receipt4SetToken, _, err := setTokenSymbol(zksyncHandle, mpriKey, symbol, "0", 18)
	assert.Nil(t, err)
	assert.Equal(t, receipt4SetToken.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyContractToTreeAction], zksyncTypes.TyContractToTreeAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTreeToContractAction], zksyncTypes.TyTreeToContractAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//测试将L2账户余额转入到合约
	receipt, _, err = tree2contract(zksyncHandle, acc1privkey, accountID, tokenId, "10000000000")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(10000000000)-int64(10000))
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//确认合约余额
	zkQueryReq := &zksyncTypes.ZkQueryReq{
		TokenSymbol:       symbol,
		Chain33WalletAddr: "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
	}
	msg, err := zksyncHandle.Query_GetZkContractAccount(zkQueryReq)
	assert.Nil(t, err)
	accountInfo, ok := msg.(*types.Account)
	assert.Equal(t, ok, true)
	assert.Equal(t, int64(1), accountInfo.Balance)
	fmt.Println("accountInfo =", accountInfo)
}

func TestContract2Tree(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	//zksync_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zksync_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	tokenId := uint64(0)

	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "2000000000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "2000000000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	symbol := "ETH"
	receipt4SetToken, _, err := setTokenSymbol(zksyncHandle, mpriKey, symbol, "0", 18)
	assert.Nil(t, err)
	assert.Equal(t, receipt4SetToken.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyContractToTreeAction], zksyncTypes.TyContractToTreeAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTreeToContractAction], zksyncTypes.TyTreeToContractAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferAction], zksyncTypes.TyTransferAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//测试将L2账户余额转入到合约 1 00 0000 0000 0000 0000
	receipt, _, err = tree2contract(zksyncHandle, acc1privkey, accountID, tokenId, "1000000000000000000")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	balance := fmt.Sprintf("%d", int64(2000000000000000000)-int64(1000000000000000000)-int64(10000))
	fmt.Println("Balance is", balance)
	assert.Equal(t, balance, acc4token1Balance.Balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//确认合约余额
	zkQueryReq := &zksyncTypes.ZkQueryReq{
		TokenSymbol:       symbol,
		Chain33WalletAddr: "1JRNjdEqp4LJ5fqycUBm9ayCKSeeskgMKR",
	}
	msg, err := zksyncHandle.Query_GetZkContractAccount(zkQueryReq)
	assert.Nil(t, err)
	accountInfo, ok := msg.(*types.Account)
	assert.Equal(t, ok, true)
	assert.Equal(t, int64(100000000), accountInfo.Balance)
	fmt.Println("accountInfo =", accountInfo)

	//测试将合约余额转回到L2账户余额
	receipt, _, err = contract2tree(zksyncHandle, acc1privkey, accountID, symbol, "90")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//确认L2账户balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	balance = fmt.Sprintf("%d", int64(2000000000000000000)-int64(1000000000000000000)-int64(10000)+int64(90*1e10))
	fmt.Println("Balance is", balance)
	assert.Equal(t, balance, acc4token1Balance.Balance)

	//确认L1账户余额
	msg, err = zksyncHandle.Query_GetZkContractAccount(zkQueryReq)
	assert.Nil(t, err)
	accountInfo, ok = msg.(*types.Account)
	assert.Equal(t, ok, true)
	assert.Equal(t, int64(100000000-90-10000), accountInfo.Balance)
	fmt.Println("accountInfo =", accountInfo)
}

//通过proxyExit模式进行提币时，需要确保未设置公钥，否则提币失败
func TestProxyExitFaid(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, localReceipt, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	targetAccountID := uint64(4)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), targetAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, targetAccountID)
	assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	// 再次铸币
	// zkAddr0: 12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d *** l2addr0: 27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97 *** key: 0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4
	queueId = uint64(1)
	tokenId = uint64(0)
	receipt, localReceipt, err = deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	proxyAccountID := uint64(5)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), proxyAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyProxyExitAction], zksyncTypes.TyProxyExitAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err = chain33Common.FromHex("0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4")
	assert.Nil(t, err)
	acc1privkey, err = driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, proxyAccountID)
	assert.Nil(t, err)

	//测试提币
	receipt, localReceipt, err = proxyExit(zksyncHandle, acc1privkey, targetAccountID, proxyAccountID, tokenId)
	assert.NotNil(t, err)
}

func TestProxyExit(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, localReceipt, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	targetAccountID := uint64(4)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), targetAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	//acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	//assert.Nil(t, err)
	//acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	//assert.Nil(t, err)
	//err = setPubKey(zksyncHandle, acc1privkey, targetAccountID)
	//assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	// 再次铸币
	// zkAddr0: 12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d *** l2addr0: 27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97 *** key: 0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4
	queueId = uint64(1)
	tokenId = uint64(0)
	receipt, localReceipt, err = deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	proxyAccountID := uint64(5)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), proxyAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyProxyExitAction], zksyncTypes.TyProxyExitAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, proxyAccountID)
	assert.Nil(t, err)

	//测试提币
	receipt, _, err = proxyExit(zksyncHandle, acc1privkey, targetAccountID, proxyAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), targetAccountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "0")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//因为未设置交易费，所以余额不变动
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), proxyAccountID, tokenId)
	assert.Nil(t, err)
	balance := int64(1000000000000) - int64(1000000)
	balanceActual, _ := big.NewInt(0).SetString(acc4token1Balance.Balance, 10)

	assert.Equal(t, uint64(balance), balanceActual.Uint64())
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//检查交易费的账户余额
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemFeeAccountId, tokenId)
	assert.Nil(t, err)
	proxyExitFee := "1000000"
	assert.Equal(t, acc4token1Balance.Balance, proxyExitFee)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
}

func TestMintNFT(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyMintNFTAction], zksyncTypes.TyMintNFTAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//NFT 铸币
	contentHash := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	receipt, _, err = mintNFT(zksyncHandle, acc1privkey, accountID, accountID, contentHash)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//1.检查交易费变动，确认铸币者的balance,扣除了交易费
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	mintFee := 100
	balance := fmt.Sprintf("%d", int64(1000000000000)-int64(mintFee))
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
	//1.2系统交易费账户增加相应的数量
	systemFeeAccountBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemFeeAccountId, tokenId)
	assert.Nil(t, err)
	mintFeeStr := "100"
	assert.Equal(t, systemFeeAccountBalance.Balance, mintFeeStr)
	assert.Equal(t, systemFeeAccountBalance.TokenId, uint64(0))
	fmt.Println("systemFeeAccountBalance is", systemFeeAccountBalance.Balance)

	//2　铸币账户的次数为1
	SystemNFTTokenBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, zksyncTypes.SystemNFTTokenId)
	assert.Nil(t, err)
	systemNFTTokenBalanceStr := "1"
	assert.Equal(t, SystemNFTTokenBalance.Balance, systemNFTTokenBalanceStr)
	assert.Equal(t, SystemNFTTokenBalance.TokenId, uint64(zksyncTypes.SystemNFTTokenId))
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", zksyncTypes.SystemNFTTokenId, SystemNFTTokenBalance.Balance)

	//3.SystemNFTAccountId's SystemNFTTokenId+1, 产生新的NFT的id
	SystemNFTTokenBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemNFTAccountId, zksyncTypes.SystemNFTTokenId)
	assert.Nil(t, err)
	systemNFTTokenBalanceStr = "258"
	assert.Equal(t, SystemNFTTokenBalance.Balance, systemNFTTokenBalanceStr)
	assert.Equal(t, SystemNFTTokenBalance.TokenId, uint64(zksyncTypes.SystemNFTTokenId))
	fmt.Println("TokenBalance for account ID", zksyncTypes.SystemNFTAccountId, "tokenId", zksyncTypes.SystemNFTTokenId, SystemNFTTokenBalance.Balance)

	//4. SystemNFTAccountId set new NFT id to balance by NFT contentHash
	newNFTid := uint64(258)
	newNFTIdBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemNFTAccountId, newNFTid)
	assert.Nil(t, err)
	creatorSerialId := "0"
	contentPart1, contentPart2, _, err := zksyncTypes.SplitNFTContent(contentHash)
	assert.Nil(t, err)
	newNFTTokenBalanceStr, err := getNewNFTTokenBalance(accountID, creatorSerialId, zksyncTypes.ZKERC721, 1, contentPart1.String(), contentPart2.String())
	assert.Nil(t, err)
	assert.Equal(t, newNFTIdBalance.Balance, newNFTTokenBalanceStr)
	assert.Equal(t, newNFTIdBalance.TokenId, newNFTid)
	fmt.Println("TokenBalance for account ID", zksyncTypes.SystemNFTAccountId, "tokenId", newNFTid, newNFTIdBalance.Balance)

	//5. recipientAddr new NFT id balance+amount
	tokenBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, newNFTid)
	assert.Nil(t, err)
	balanceStr := "1"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", newNFTid, tokenBalance.Balance)

	// 再次铸币
	// zkAddr0: 12HKLEn6g4FH39yUbHh4EVJWcFo5CXg22d *** l2addr0: 27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97 *** key: 0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4
	queueId = uint64(1)
	tokenId = uint64(0)
	receipt, _, err = deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "27f272f1adf1c12e0ea7c48d8ace0370610952f17666bdb11ea5a8d7ab980d97")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID = uint64(5)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	acc1privkeySli, err = chain33Common.FromHex("0x9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4")
	assert.Nil(t, err)
	acc1privkey, err = driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//NFT 铸币
	contentHash = "9d4f8ab11361be596468b265cb66946c87873d4a119713fd0c3d8302eae0a8e4"
	receipt, _, err = mintNFT(zksyncHandle, acc1privkey, accountID, accountID, contentHash)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	//1.检查交易费变动，确认铸币者的balance,扣除了交易费
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	mintFee = 100
	balance = fmt.Sprintf("%d", int64(1000000000000)-int64(mintFee))
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))
	//1.2系统交易费账户增加相应的数量
	systemFeeAccountBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemFeeAccountId, tokenId)
	assert.Nil(t, err)
	mintFeeStr = "200"
	assert.Equal(t, systemFeeAccountBalance.Balance, mintFeeStr)
	assert.Equal(t, systemFeeAccountBalance.TokenId, uint64(0))
	fmt.Println("systemFeeAccountBalance is", systemFeeAccountBalance.Balance)

	//2　铸币账户的次数为1
	SystemNFTTokenBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, zksyncTypes.SystemNFTTokenId)
	assert.Nil(t, err)
	systemNFTTokenBalanceStr = "1"
	assert.Equal(t, SystemNFTTokenBalance.Balance, systemNFTTokenBalanceStr)
	assert.Equal(t, SystemNFTTokenBalance.TokenId, uint64(zksyncTypes.SystemNFTTokenId))
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", zksyncTypes.SystemNFTTokenId, SystemNFTTokenBalance.Balance)

	//3.SystemNFTAccountId's SystemNFTTokenId+1, 产生新的NFT的id
	SystemNFTTokenBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemNFTAccountId, zksyncTypes.SystemNFTTokenId)
	assert.Nil(t, err)
	systemNFTTokenBalanceStr = "259"
	assert.Equal(t, SystemNFTTokenBalance.Balance, systemNFTTokenBalanceStr)
	assert.Equal(t, SystemNFTTokenBalance.TokenId, uint64(zksyncTypes.SystemNFTTokenId))
	fmt.Println("TokenBalance for account ID", zksyncTypes.SystemNFTAccountId, "tokenId", zksyncTypes.SystemNFTTokenId, SystemNFTTokenBalance.Balance)

	//4. SystemNFTAccountId set new NFT id to balance by NFT contentHash
	newNFTid = uint64(259)
	newNFTIdBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), zksyncTypes.SystemNFTAccountId, newNFTid)
	assert.Nil(t, err)
	creatorSerialId = "0"
	contentPart1, contentPart2, _, err = zksyncTypes.SplitNFTContent(contentHash)
	assert.Nil(t, err)
	newNFTTokenBalanceStr, err = getNewNFTTokenBalance(accountID, creatorSerialId, zksyncTypes.ZKERC721, 1, contentPart1.String(), contentPart2.String())
	assert.Nil(t, err)
	assert.Equal(t, newNFTIdBalance.Balance, newNFTTokenBalanceStr)
	assert.Equal(t, newNFTIdBalance.TokenId, newNFTid)
	fmt.Println("TokenBalance for account ID", zksyncTypes.SystemNFTAccountId, "tokenId", newNFTid, newNFTIdBalance.Balance)

	//5. recipientAddr new NFT id balance+amount
	tokenBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, newNFTid)
	assert.Nil(t, err)
	balanceStr = "1"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", newNFTid, tokenBalance.Balance)
}

func TestWithdrawNFT(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyWithdrawNFTAction], zksyncTypes.TyWithdrawNFTAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//NFT 铸币
	contentHash := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	receipt, _, err = mintNFT(zksyncHandle, acc1privkey, accountID, accountID, contentHash)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	nftTokenId := uint64(258)
	tokenBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr := "1"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", nftTokenId, tokenBalance.Balance)

	//NFT 提币

	receipt, _, err = withdrawNFT(zksyncHandle, acc1privkey, accountID, nftTokenId, 1)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	tokenBalance, err = GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr = "0"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", nftTokenId, tokenBalance.Balance)
}

func TestTransferNFT(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	receipt, _, err = setTxFee(zksyncHandle, mpriKey, tokenId, zksyncTypes.FeeMap[zksyncTypes.TyTransferNFTAction], zksyncTypes.TyTransferNFTAction)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//测试向新账户进行转币操作
	toEthAddr := "12a0e25e62c1dbd32e505446062b26aecb65f028"
	toL2Chain33Addr := "2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c"
	//toAddrprivkey := "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
	receipt, _, err = transfer2New(zksyncHandle, acc1privkey, tokenId, accountID, "200", toEthAddr, toL2Chain33Addr)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//NFT 铸币
	contentHash := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	receipt, _, err = mintNFT(zksyncHandle, acc1privkey, accountID, accountID, contentHash)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//NFT 转币操作
	nftTokenId := uint64(258)
	fromAccountID := accountID
	toAccountID := accountID + 1
	receipt, _, err = transferNFT(zksyncHandle, acc1privkey, fromAccountID, toAccountID, nftTokenId, 1)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	tokenBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr := "0"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", nftTokenId, tokenBalance.Balance)

	toBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), toAccountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr = "1"
	assert.Equal(t, toBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", toAccountID, "tokenId", nftTokenId, toBalance.Balance)
}

func TestNFTMisc(t *testing.T) {
	initSetup()
	defer util.CloseTestDB(dbDir, dbHanleGlobal)

	fmt.Println("Going to do TestSpot")

	var driver secp256k1.Driver

	//12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv
	managerPrivateKeySli, err := chain33Common.FromHex("4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01")
	assert.Nil(t, err)
	mpriKey, err := driver.PrivKeyFromBytes(managerPrivateKeySli)
	assert.Nil(t, err)

	queueId := uint64(0)
	tokenId := uint64(0)
	receipt, _, err := deposit(zksyncHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)
	accountID := uint64(firstUserAccoutID)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(0))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc1privkey, accountID)
	assert.Nil(t, err)

	//测试向新账户进行转币操作
	toEthAddr := "12a0e25e62c1dbd32e505446062b26aecb65f028"
	toL2Chain33Addr := "2afff20cc3c20f9def369626463fb027ebeba0bd976025f68316bb8eab55d48c"
	//toAddrprivkey := "0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115"
	receipt, _, err = transfer2New(zksyncHandle, acc1privkey, tokenId, accountID, "200", toEthAddr, toL2Chain33Addr)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//设置公钥,给账户４
	acc4privkeySli, err := chain33Common.FromHex("0x7a80a1f75d7360c6123c32a78ecf978c1ac55636f87892df38d8b85a9aeff115")
	assert.Nil(t, err)
	acc4privkey, err := driver.PrivKeyFromBytes(acc4privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zksyncHandle, acc4privkey, accountID+1)
	assert.Nil(t, err)

	//NFT 铸币
	nftTokenId := uint64(258)
	fromAccountID := accountID
	toAccountID := accountID + 1
	contentHash := "4257D8692EF7FE13C68B65D6A52F03933DB2FA5CE8FAF210B5B8B80C721CED01"
	receipt, _, err = mintNFT(zksyncHandle, acc1privkey, accountID, toAccountID, contentHash)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	//NFT 转币操作
	receipt, _, err = transferNFT(zksyncHandle, acc4privkey, toAccountID, fromAccountID, nftTokenId, 1)
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(receipt.KV), 0)

	tokenBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), accountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr := "1"
	assert.Equal(t, tokenBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", accountID, "tokenId", nftTokenId, tokenBalance.Balance)

	toBalance, err := GetTokenByAccountIdAndTokenIdInDB(zksyncHandle.GetStateDB(), toAccountID, nftTokenId)
	assert.Nil(t, err)
	balanceStr = "0"
	assert.Equal(t, toBalance.Balance, balanceStr)
	fmt.Println("TokenBalance for account ID", toAccountID, "tokenId", nftTokenId, toBalance.Balance)
}

func deposit(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, tokenId, queueId uint64, amount, ethAddress, chain33Addr string) (*types.Receipt, *types.LocalDBSet, error) {
	deposit := &zksyncTypes.ZkDeposit{
		TokenId:      tokenId,
		Amount:       amount,
		EthAddress:   ethAddress,
		Chain33Addr:  chain33Addr,
		L1PriorityId: int64(queueId),
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyDepositAction,
		Value: &zksyncTypes.ZksyncAction_Deposit{
			Deposit: deposit,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_Deposit(action.GetDeposit(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec deposit cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_Deposit(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("deposit cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func setPubKey(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, accountId uint64) error {
	setPubKey := &zksyncTypes.ZkSetPubKey{
		AccountId: accountId,
		PubKeyTy:  0,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TySetPubKeyAction,
		Value: &zksyncTypes.ZksyncAction_SetPubKey{
			SetPubKey: setPubKey,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return err
	}

	receipt, err := zksyncHandle.Exec_SetPubKey(action.GetSetPubKey(), tx, index)
	if nil != err {
		return err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_SetPubKey(nil, tx, receiptData, index)
	if nil != err {
		return err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}

	index++
	return err
}

func withdraw(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, accountID, tokenId uint64, amount string) (*types.Receipt, *types.LocalDBSet, error) {
	withdraw := &zksyncTypes.ZkWithdraw{
		TokenId:   tokenId,
		Amount:    amount,
		AccountId: accountID,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyWithdrawAction,
		Value: &zksyncTypes.ZksyncAction_ZkWithdraw{
			ZkWithdraw: withdraw,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_ZkWithdraw(action.GetZkWithdraw(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec withdraw cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_Withdraw(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("withdraw cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func setTxFee(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, tokenId uint64, amount string, actionTy int32) (*types.Receipt, *types.LocalDBSet, error) {
	//zksyncTypes.TyWithdrawAction
	setFee := &zksyncTypes.ZkSetFee{
		TokenId:  tokenId,
		Amount:   amount,
		ActionTy: actionTy,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TySetFeeAction,
		Value: &zksyncTypes.ZksyncAction_SetFee{
			SetFee: setFee,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_SetFee(action.GetSetFee(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec withdraw cost time = ", time.Since(t1))

	index++
	fmt.Println("withdraw cost time = ", time.Since(t1))

	return receipt, nil, nil
}

func transfer(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, fromAccountId, toAccountId, tokenId uint64, amount string) (*types.Receipt, *types.LocalDBSet, error) {
	transfer := &zksyncTypes.ZkTransfer{
		TokenId:       tokenId,
		Amount:        amount,
		FromAccountId: fromAccountId,
		ToAccountId:   toAccountId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferAction,
		Value: &zksyncTypes.ZksyncAction_ZkTransfer{
			ZkTransfer: transfer,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_ZkTransfer(action.GetZkTransfer(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec transfer cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_Transfer(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("transfer cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func transfer2New(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, tokenId, fromAccountId uint64, amount, toEthAddress, toChain33Address string) (*types.Receipt, *types.LocalDBSet, error) {
	transfer2New := &zksyncTypes.ZkTransferToNew{
		TokenId:         tokenId,
		Amount:          amount,
		FromAccountId:   fromAccountId,
		ToEthAddress:    toEthAddress,
		ToLayer2Address: toChain33Address,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferToNewAction,
		Value: &zksyncTypes.ZksyncAction_TransferToNew{
			TransferToNew: transfer2New,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_TransferToNew(action.GetTransferToNew(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec transfer2New cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_TransferToNew(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("transfer2New cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func tree2contract(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, accountID, tokenId uint64, amount string) (*types.Receipt, *types.LocalDBSet, error) {
	tree2contract := &zksyncTypes.ZkTreeToContract{
		TokenId:   tokenId,
		Amount:    amount,
		AccountId: accountID,
		ToAcctId:  zksyncTypes.SystemTree2ContractAcctId,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTreeToContractAction,
		Value: &zksyncTypes.ZksyncAction_TreeToContract{
			TreeToContract: tree2contract,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_TreeToContract(action.GetTreeToContract(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec tree2contract cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_TreeToContract(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("tree2contract cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func setTokenSymbol(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, symbol, tokenID string, decimal uint32) (*types.Receipt, *types.LocalDBSet, error) {
	contract2tree := &zksyncTypes.ZkTokenSymbol{
		Id:      tokenID,
		Symbol:  symbol,
		Decimal: decimal,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TySetTokenSymbolAction,
		Value: &zksyncTypes.ZksyncAction_SetTokenSymbol{
			SetTokenSymbol: contract2tree,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_SetTokenSymbol(action.GetSetTokenSymbol(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec contract2tree cost time = ", time.Since(t1))
	index++
	fmt.Println("contract2tree cost time = ", time.Since(t1))

	return receipt, nil, nil
}

func contract2tree(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, toAccountID uint64, symbol, amount string) (*types.Receipt, *types.LocalDBSet, error) {
	contract2tree := &zksyncTypes.ZkContractToTree{
		TokenSymbol: symbol,
		Amount:      amount,
		ToAccountId: toAccountID,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyContractToTreeAction,
		Value: &zksyncTypes.ZksyncAction_ContractToTree{
			ContractToTree: contract2tree,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_ContractToTree(action.GetContractToTree(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec contract2tree cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_ContractToTree(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("contract2tree cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func proxyExit(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, targetAccountID, proxyAccountID, tokenId uint64) (*types.Receipt, *types.LocalDBSet, error) {
	proxyExit := &zksyncTypes.ZkProxyExit{
		TokenId:  tokenId,
		ProxyId:  proxyAccountID,
		TargetId: targetAccountID,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyProxyExitAction,
		Value: &zksyncTypes.ZksyncAction_ProxyExit{
			ProxyExit: proxyExit,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_ProxyExit(action.GetProxyExit(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec forceExit cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_ProxyExit(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("forceExit cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func fullExit(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, accountID, tokenId, priorityQueueId uint64) (*types.Receipt, *types.LocalDBSet, error) {
	fullExit := &zksyncTypes.ZkFullExit{
		TokenId:            tokenId,
		AccountId:          accountID,
		EthPriorityQueueId: int64(priorityQueueId),
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyFullExitAction,
		Value: &zksyncTypes.ZksyncAction_FullExit{
			FullExit: fullExit,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_FullExit(action.GetFullExit(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec fullExit cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_FullExit(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("fullExit cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func mintNFT(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, fromAccountId, recipientAccountId uint64, contentHash string) (*types.Receipt, *types.LocalDBSet, error) {
	mintNFT := &zksyncTypes.ZkMintNFT{
		FromAccountId: fromAccountId,
		RecipientId:   recipientAccountId,
		ContentHash:   contentHash,
		ErcProtocol:   zksyncTypes.ZKERC721,
		Amount:        1,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyMintNFTAction,
		Value: &zksyncTypes.ZksyncAction_MintNFT{
			MintNFT: mintNFT,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_MintNFT(action.GetMintNFT(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec mintNFT cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_MintNFT(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++

	fmt.Println("mintNFT cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func withdrawNFT(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, fromAccountId, nftTokenId, amount uint64) (*types.Receipt, *types.LocalDBSet, error) {
	withdrawNFT := &zksyncTypes.ZkWithdrawNFT{
		FromAccountId: fromAccountId,
		NFTTokenId:    nftTokenId,
		Amount:        amount,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyWithdrawNFTAction,
		Value: &zksyncTypes.ZksyncAction_WithdrawNFT{
			WithdrawNFT: withdrawNFT,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_WithdrawNFT(action.GetWithdrawNFT(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec withdrawNFT cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_WithdrawNFT(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("withdrawNFT cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func transferNFT(zksyncHandle *zksync, privateKey chain33Crypto.PrivKey, fromAccountId, recipientAccountId, ntTokenId, amount uint64) (*types.Receipt, *types.LocalDBSet, error) {
	transferNFT := &zksyncTypes.ZkTransferNFT{
		FromAccountId: fromAccountId,
		RecipientId:   recipientAccountId,
		NFTTokenId:    ntTokenId,
		Amount:        amount,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyTransferNFTAction,
		Value: &zksyncTypes.ZksyncAction_TransferNFT{
			TransferNFT: transferNFT,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	t1 := time.Now()
	receipt, err := zksyncHandle.Exec_TransferNFT(action.GetTransferNFT(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zksyncHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}
	fmt.Println("exec transferNFT cost time = ", time.Since(t1))

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zksyncHandle.ExecLocal_TransferNFT(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zksyncHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++
	fmt.Println("transferNFT cost time = ", time.Since(t1))

	return receipt, localDBSet, nil
}

func createChain33Tx(privateKey chain33Crypto.PrivKey, action proto.Message, execer string, fee int64) *types.Transaction {
	tx := &types.Transaction{Execer: []byte(execer), Payload: types.Encode(action), Fee: fee}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	tx.Nonce = random.Int63()

	err := SignTransaction(privateKey, tx)
	if nil != err {
		return nil
	}

	return tx
}

func SignTransaction(key chain33Crypto.PrivKey, tx *types.Transaction) (err error) {
	action := new(zksyncTypes.ZksyncAction)
	if err = types.Decode(tx.Payload, action); err != nil {
		return
	}

	privateKey, err := eddsa.GenerateKey(bytes.NewReader(key.Bytes()))
	if err != nil {
		return
	}

	var msg *zksyncTypes.ZkMsg
	var signInfo *zksyncTypes.ZkSignature
	switch action.GetTy() {
	case zksyncTypes.TyDepositAction:
		deposit := action.GetDeposit()
		msg = wallet.GetDepositMsg(deposit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		deposit.Signature = signInfo
	case zksyncTypes.TyWithdrawAction:
		withDraw := action.GetZkWithdraw()
		msg = wallet.GetWithdrawMsg(withDraw)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		withDraw.Signature = signInfo
	case zksyncTypes.TyContractToTreeAction:
		contractToLeaf := action.GetContractToTree()
		msg = wallet.GetContractToTreeMsg(contractToLeaf)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		contractToLeaf.Signature = signInfo
	case zksyncTypes.TyTreeToContractAction:
		leafToContract := action.GetTreeToContract()
		msg = wallet.GetTreeToContractMsg(leafToContract)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		leafToContract.Signature = signInfo
	case zksyncTypes.TyTransferAction:
		transfer := action.GetZkTransfer()
		msg = wallet.GetTransferMsg(transfer)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		transfer.Signature = signInfo
	case zksyncTypes.TyTransferToNewAction:
		transferToNew := action.GetTransferToNew()
		msg = wallet.GetTransferToNewMsg(transferToNew)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		transferToNew.Signature = signInfo
	case zksyncTypes.TyProxyExitAction:
		forceQuit := action.GetProxyExit()
		msg = wallet.GetProxyExitMsg(forceQuit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		forceQuit.Signature = signInfo
	case zksyncTypes.TySetPubKeyAction:
		setPubKey := action.GetSetPubKey()
		//如果是添加公钥的操作，则默认设置这里生成的公钥 todo:要是未来修改可以自定义公钥，这里需要删除
		//如果是添加公钥的操作，则默认设置这里生成的公钥
		if setPubKey.PubKeyTy == 0 {
			pubKey := &zksyncTypes.ZkPubKey{
				X: privateKey.PublicKey.A.X.String(),
				Y: privateKey.PublicKey.A.Y.String(),
			}
			setPubKey.PubKey = pubKey
		}

		msg = wallet.GetSetPubKeyMsg(setPubKey)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		setPubKey.Signature = signInfo
	case zksyncTypes.TyFullExitAction:
		forceQuit := action.GetFullExit()
		msg = wallet.GetFullExitMsg(forceQuit)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		forceQuit.Signature = signInfo

	case zksyncTypes.TyMintNFTAction:
		nft := action.GetMintNFT()
		msg := wallet.GetMintNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	case zksyncTypes.TyTransferNFTAction:
		nft := action.GetTransferNFT()
		msg := wallet.GetTransferNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	case zksyncTypes.TyWithdrawNFTAction:
		nft := action.GetWithdrawNFT()
		msg := wallet.GetWithdrawNFTMsg(nft)
		signInfo, err = SignTxInEddsa(msg, privateKey)
		if err != nil {
			return
		}
		nft.Signature = signInfo
	}

	tx.Payload = types.Encode(action)
	tx.Sign(types.SECP256K1, key)
	return
}

func SignTxInEddsa(msg *zksyncTypes.ZkMsg, privateKey eddsa.PrivateKey) (*zksyncTypes.ZkSignature, error) {
	signInfo, err := privateKey.Sign(wallet.GetMsgHash(msg), mimc.NewMiMC(zksyncTypes.ZkMimcHashSeed))
	if err != nil {
		return nil, err
	}
	pubKey := &zksyncTypes.ZkPubKey{
		X: privateKey.PublicKey.A.X.String(),
		Y: privateKey.PublicKey.A.Y.String(),
	}
	sign := &zksyncTypes.ZkSignature{
		PubKey:   pubKey,
		SignInfo: hex.EncodeToString(signInfo),
		Msg:      msg,
	}
	return sign, nil
}

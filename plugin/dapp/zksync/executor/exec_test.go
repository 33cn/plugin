package executor

import (
	"bytes"
	"encoding/hex"
	"fmt"
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

[mver.exec.sub.zkspot]
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
zkChain33FeeAddr="2c4a5c378be2424fa7585320630eceba764833f1ec1ffb2fafc1af97f27baf5a"
`
	chain33TestCfg = types.NewChain33Config(strings.Replace(cfgstring4execTest, "Title=\"local\"", "Title=\"chain33\"", 1))
	zkspotHandle   *zksync
	dbHanleGlobal  db.DB
	index          = 0
	dbDir          = ""
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
	zkspotHandle = driver.(*zksync)
	dbHanleGlobal = dbHanle
	dbDir = dir
}

//zkspot_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
//zkspot_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
//zkspot_setPubKey ${acc2id} ${acc2privkey}
//zkspot_setPubKey ${acc3id} ${acc3privkey}
//zkspot_limitorder buy  2 1 150000000 2000000000 ${acc2id} ${acc2eth} ${acc2privkey}
//zkspot_limitorder sell 2 1 150000000 2000000000 ${acc3id} ${acc3eth} ${acc3privkey}
//zkspot_account2token ${acc2id} 1
//zkspot_account2token ${acc2id} 2
//zkspot_account2token ${acc3id} 1
//zkspot_account2token ${acc3id} 2

//go test -v -test.bench BenchmarkSpot -test.run BenchmarkSpot -count 1
//go test --run=TestSpot

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

	//zkspot_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zkspot_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	receipt, localReceipt, err := deposit(zkspotHandle, mpriKey, 1, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	accountID := uint64(3)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zkspotHandle.GetStateDB(), accountID, 1)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(1))
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

	//zkspot_deposit 1 1000000000000 ${acc2privkey} ${acc2eth} 87
	//zkspot_deposit 2 1000000000000 ${acc3privkey} ${acc3eth} 88
	queueId := uint64(0)
	tokenId := uint64(1)
	receipt, localReceipt, err := deposit(zkspotHandle, mpriKey, tokenId, queueId, "1000000000000", "abcd68033A72978C1084E2d44D1Fa06DdC4A2d57", "2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	accountID := uint64(3)
	//确认balance
	acc4token1Balance, err := GetTokenByAccountIdAndTokenIdInDB(zkspotHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
	assert.Equal(t, acc4token1Balance.Balance, "1000000000000")
	assert.Equal(t, acc4token1Balance.TokenId, uint64(1))

	//设置公钥
	acc1privkeySli, err := chain33Common.FromHex("0x19c069234f9d3e61135fefbeb7791b149cdf6af536f26bebb310d4cd22c3fee4")
	assert.Nil(t, err)
	acc1privkey, err := driver.PrivKeyFromBytes(acc1privkeySli)
	assert.Nil(t, err)
	err = setPubKey(zkspotHandle, acc1privkey, accountID)
	assert.Nil(t, err)
	//leaf.Chain33Addr=2b8a83399ffc86cc88f0493f17c9698878dcf7caf0bf04a3a5321542a7a416d1
	//calcChain33Addr= 19694183066356799104974294716313078444659172842638956126168373945465009608401

	//测试提币
	receipt, localReceipt, err = withdraw(zkspotHandle, acc1privkey, accountID, tokenId, "200")
	assert.Nil(t, err)
	assert.Equal(t, receipt.Ty, int32(types.ExecOk))
	assert.Greater(t, len(localReceipt.KV), 0)
	//确认balance
	acc4token1Balance, err = GetTokenByAccountIdAndTokenIdInDB(zkspotHandle.GetStateDB(), accountID, tokenId)
	assert.Nil(t, err)
    withdrawFee := 1000000
	balance := fmt.Sprintf("%d", 1000000000000 - 200 - withdrawFee)
	fmt.Println("Balance is", balance)
	assert.Equal(t, acc4token1Balance.Balance, balance)
	assert.Equal(t, acc4token1Balance.TokenId, uint64(1))
}

func deposit(zkspotHandle *zksync, privateKey chain33Crypto.PrivKey, tokenId, queueId uint64, amount, ethAddress, chain33Addr string) (*types.Receipt, *types.LocalDBSet, error) {
	deposit := &zksyncTypes.ZkDeposit{
		TokenId:            tokenId,
		Amount:             amount,
		EthAddress:         ethAddress,
		Chain33Addr:        chain33Addr,
		EthPriorityQueueId: int64(queueId),
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

	receipt, err := zkspotHandle.Exec_Deposit(action.GetDeposit(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zkspotHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zkspotHandle.ExecLocal_Deposit(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zkspotHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++

	return receipt, localDBSet, nil
}

func setPubKey(zkspotHandle *zksync, privateKey chain33Crypto.PrivKey, accountId uint64) error {
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

	receipt, err := zkspotHandle.Exec_SetPubKey(action.GetSetPubKey(), tx, index)
	if nil != err {
		return err
	}

	for _, kv := range receipt.KV {
		_ = zkspotHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zkspotHandle.ExecLocal_SetPubKey(nil, tx, receiptData, index)
	if nil != err {
		return err
	}
	for _, kv := range localDBSet.KV {
		_ = zkspotHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}

	index++
	return err
}

func withdraw(zkspotHandle *zksync, privateKey chain33Crypto.PrivKey, accountID, tokenId uint64, amount string) (*types.Receipt, *types.LocalDBSet, error) {
	withdraw := &zksyncTypes.ZkWithdraw{
		TokenId:            tokenId,
		Amount:             amount,
		AccountId:         accountID,
	}

	action := &zksyncTypes.ZksyncAction{
		Ty: zksyncTypes.TyWithdrawAction,
		Value: &zksyncTypes.ZksyncAction_Withdraw{
			Withdraw: withdraw,
		},
	}

	tx := createChain33Tx(privateKey, action, zksyncTypes.Zksync, int64(1e8))
	if err := types.Decode(tx.Payload, action); nil != err {
		return nil, nil, err
	}

	receipt, err := zkspotHandle.Exec_Withdraw(action.GetWithdraw(), tx, index)
	if nil != err {
		return nil, nil, err
	}

	for _, kv := range receipt.KV {
		_ = zkspotHandle.GetStateDB().Set(kv.GetKey(), kv.GetValue())
	}

	receiptData := &types.ReceiptData{Ty: receipt.Ty, Logs: receipt.Logs}
	localDBSet, err := zkspotHandle.ExecLocal_Withdraw(nil, tx, receiptData, index)
	if nil != err {
		return nil, nil, err
	}
	for _, kv := range localDBSet.KV {
		_ = zkspotHandle.GetLocalDB().Set(kv.GetKey(), kv.GetValue())
	}
	index++

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
		withDraw := action.GetWithdraw()
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
		transfer := action.GetTransfer()
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
	case zksyncTypes.TyForceExitAction:
		forceQuit := action.GetForceExit()
		msg = wallet.GetForceExitMsg(forceQuit)
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
// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wallet

import (
	"encoding/hex"
	"math/rand"
	"sync"
	"testing"

	"github.com/stretchr/testify/mock"

	wcom "github.com/33cn/chain33/wallet/common"
	ty "github.com/33cn/plugin/plugin/dapp/pos33/types"
	"github.com/stretchr/testify/assert"

	_ "github.com/33cn/chain33/system"

	"github.com/33cn/chain33/types"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/client/mocks"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/db"
)

const (
	sendhash = "sendhash"
)

func TestForceClosePos33TicketList(t *testing.T) {
	cfg := types.NewChain33ConfigNoInit(cfgstring)
	cfg.GetModuleConfig().Consensus.Name = "pos33"

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi

	ticket.walletOperate = wallet
	t1 := &ty.Pos33Ticket{Status: 1, IsGenesis: false}
	t2 := &ty.Pos33Ticket{Status: 2, IsGenesis: false}
	t3 := &ty.Pos33Ticket{Status: 3, IsGenesis: false}
	tlist := []*ty.Pos33Ticket{t1, t2, t3}

	r1, r2 := ticket.forceClosePos33TicketList(0, nil, tlist)
	assert.Equal(t, []byte(sendhash), r1)
	assert.Nil(t, r2)

}

func TestClosePos33TicketsByAddr(t *testing.T) {
	cfg := types.NewChain33ConfigNoInit(cfgstring)
	cfg.GetModuleConfig().Consensus.Name = "pos33"

	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet

	t1 := &ty.Pos33Ticket{Status: 1, IsGenesis: false}
	t2 := &ty.Pos33Ticket{Status: 2, IsGenesis: false}
	t3 := &ty.Pos33Ticket{Status: 3, IsGenesis: false}

	tlist := &ty.ReplyPos33TicketList{Tickets: []*ty.Pos33Ticket{t1, t2, t3}}
	qapi.On("Query", ty.Pos33TicketX, "Pos33TicketList", mock.Anything).Return(tlist, nil)

	r1, r2 := ticket.closePos33TicketsByAddr(0, priKey)
	assert.Equal(t, []byte(sendhash), r1)
	assert.Nil(t, r2)

}

func TestBuyPos33TicketOne(t *testing.T) {
	cfg := types.NewChain33ConfigNoInit(cfgstring)
	cfg.GetModuleConfig().Consensus.Name = "pos33"

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet
	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)
	hash, r1, r2 := ticket.buyPos33TicketOne(0, priKey)
	assert.Equal(t, []byte(sendhash), hash)
	assert.Equal(t, 10, r1)
	assert.Nil(t, r2)

}

func TestBuyMinerAddrPos33TicketOne(t *testing.T) {
	cfg := types.NewChain33ConfigNoInit(cfgstring)
	cfg.GetModuleConfig().Consensus.Name = "pos33"

	pk, err := hex.DecodeString("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944")
	assert.Nil(t, err)
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	assert.Nil(t, err)
	priKey, err := secp.PrivKeyFromBytes(pk)
	assert.Nil(t, err)

	ticket := &ticketPolicy{mtx: &sync.Mutex{}}
	ticket.cfg = &subConfig{}
	ticket.initMinerWhiteList(nil)
	wallet := new(walletOperateMock)
	qapi := new(mocks.QueueProtocolAPI)
	qapi.On("GetConfig", mock.Anything).Return(cfg, nil)
	wallet.api = qapi
	ticket.walletOperate = wallet

	tlist := &types.ReplyStrings{Datas: []string{"12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}}
	qapi.On("Query", ty.Pos33TicketX, "MinerSourceList", mock.Anything).Return(tlist, nil)

	hashs, r2, r3 := ticket.buyMinerAddrPos33TicketOne(0, priKey)
	assert.Equal(t, [][]byte{[]byte(sendhash)}, hashs)
	assert.Equal(t, 10, r2)
	assert.Nil(t, r3)

}

type walletOperateMock struct {
	api client.QueueProtocolAPI
}

func (_m *walletOperateMock) AddrInWallet(addr string) bool {
	return false
}

// CheckWalletStatus provides a mock function with given fields:
func (_m *walletOperateMock) CheckWalletStatus() (bool, error) {
	return false, nil
}

// GetAPI provides a mock function with given fields:
func (_m *walletOperateMock) GetAPI() client.QueueProtocolAPI {
	return _m.api
}

// GetAllPrivKeys provides a mock function with given fields:
func (_m *walletOperateMock) GetAllPrivKeys() ([]crypto.PrivKey, error) {
	return nil, nil
}

// GetBalance provides a mock function with given fields: addr, execer
func (_m *walletOperateMock) GetBalance(addr string, execer string) (*types.Account, error) {

	return &types.Account{Balance: 10000000000000, Addr: "12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"}, nil
}

// GetBlockHeight provides a mock function with given fields:
func (_m *walletOperateMock) GetBlockHeight() int64 {
	return 0
}

// GetConfig provides a mock function with given fields:
func (_m *walletOperateMock) GetConfig() *types.Wallet {
	return nil
}

// GetDBStore provides a mock function with given fields:
func (_m *walletOperateMock) GetDBStore() db.DB {
	return nil
}

// GetLastHeader provides a mock function with given fields:
func (_m *walletOperateMock) GetLastHeader() *types.Header {
	return nil
}

// GetMutex provides a mock function with given fields:
func (_m *walletOperateMock) GetMutex() *sync.Mutex {
	return nil
}

// GetPassword provides a mock function with given fields:
func (_m *walletOperateMock) GetPassword() string {
	return ""
}

// GetPrivKeyByAddr provides a mock function with given fields: addr
func (_m *walletOperateMock) GetPrivKeyByAddr(addr string) (crypto.PrivKey, error) {
	return nil, nil
}

// GetRandom provides a mock function with given fields:
func (_m *walletOperateMock) GetRandom() *rand.Rand {
	return nil
}

// GetSignType provides a mock function with given fields:
func (_m *walletOperateMock) GetSignType() int {
	return 0
}

// GetTxDetailByHashs provides a mock function with given fields: ReqHashes
func (_m *walletOperateMock) GetTxDetailByHashs(ReqHashes *types.ReqHashes) {
}

// GetWaitGroup provides a mock function with given fields:
func (_m *walletOperateMock) GetWaitGroup() *sync.WaitGroup {
	return nil
}

// GetWalletAccounts provides a mock function with given fields:
func (_m *walletOperateMock) GetWalletAccounts() ([]*types.WalletAccountStore, error) {
	return nil, nil
}

// GetWalletDone provides a mock function with given fields:
func (_m *walletOperateMock) GetWalletDone() chan struct{} {
	return nil
}

// IsCaughtUp provides a mock function with given fields:
func (_m *walletOperateMock) IsCaughtUp() bool {
	return false
}

// IsClose provides a mock function with given fields:
func (_m *walletOperateMock) IsClose() bool {
	return false
}

// IsWalletLocked provides a mock function with given fields:
func (_m *walletOperateMock) IsWalletLocked() bool {
	return true
}

// Nonce provides a mock function with given fields:
func (_m *walletOperateMock) Nonce() int64 {
	return 0
}

// RegisterMineStatusReporter provides a mock function with given fields: reporter
func (_m *walletOperateMock) RegisterMineStatusReporter(reporter wcom.MineStatusReport) error {
	return nil
}

// SendToAddress provides a mock function with given fields: priv, addrto, amount, note, Istoken, tokenSymbol
func (_m *walletOperateMock) SendToAddress(priv crypto.PrivKey, addrto string, amount int64, note string, Istoken bool, tokenSymbol string) (*types.ReplyHash, error) {
	return &types.ReplyHash{Hash: []byte(sendhash)}, nil
}

// SendTransaction provides a mock function with given fields: payload, execer, priv, to
func (_m *walletOperateMock) SendTransaction(payload types.Message, execer []byte, priv crypto.PrivKey, to string) ([]byte, error) {
	return []byte(sendhash), nil
}

// WaitTx provides a mock function with given fields: hash
func (_m *walletOperateMock) WaitTx(hash []byte) *types.TransactionDetail {
	return nil
}

// WaitTxs provides a mock function with given fields: hashes
func (_m *walletOperateMock) WaitTxs(hashes [][]byte) []*types.TransactionDetail {
	return nil
}

var cfgstring = `
Title="YCC"
coinSymbol="YCC"
TestNet=true
FixTime=false
version="6.3.0"


[log]
# 日志级别，支持debug(dbug)/info/warn/error(eror)/crit
loglevel = "dbug"
logConsoleLevel = "error"
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
callerFile = true
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
enableReExecLocal=true
txHeight=true

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
channel=2
seeds=[]
isSeed=true
serverStart=true
innerSeedEnable=false
useGithub=false
innerBounds=300

[p2p.sub.dht]
seeds=[]
port=13801
channel=2

[rpc]
jrpcBindAddr="localhost:8801"
grpcBindAddr="localhost:8802"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]

[mempool]
name="price"
poolCacheSize=10240
minTxFee=100000
maxTxNumPerAccount=10000
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
name="pos33"
minerstart=true
genesisBlockTime=1514533394
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
minerExecs=["pos33"]

[consensus.sub.pos33]
genesisBlockTime=1514533394
listenPort="10901"

[[consensus.sub.pos33.genesis]]
minerAddr="12qyocayNF7Lv6C9qW4avxs2E7U41fKSfv"
returnAddr="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
count=10000

[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
powLimitBits="0x1f00ffff"
maxTxNumber = 5000

[mver.consensus.pos33]
coinReward = 15
coinDevFund = 5
ticketPrice = 10000
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 43200
ticketWithdrawTime = 172800
ticketMinerWaitTime = 7200
targetTimespan=2304
targetTimePerBlock=16

[store]
name="kvmvccmavl"
driver="leveldb"
storedbVersion="2.0.0"
dbPath="paradatadir/mavltree"
dbCache=128

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
enableMVCCIter=true
enableMavlPrefix=false
enableMVCC=false
enableMavlPrune=false
pruneMavlHeight=10000
enableMVCCPrune=false
pruneMVCCHeight=10000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000
# 该参数针对平行链，如果平行链的ForkKvmvccmavl高度不为0,需要开启此功能,开启此功能需要从0开始执行区块
enableEmptyBlockHandle=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
dbCache=16
signType="secp256k1"

[wallet.sub.pos33]
minerdisable=false
minerwhitelist=["*"]

[wallet.sub.multisig]
rescanMultisigAddr=false

#系统中所有的fork,默认用chain33的测试网络的
#但是我们可以替换
[fork.system]
ForkChainParamV1= 0
ForkCheckTxDup=0
ForkBlockHash=10000000000000000
ForkMinerTime= 0
ForkTransferExec=0
ForkExecKey=0
ForkTxGroup=0
ForkResetTx0=0
ForkWithdraw=0
ForkExecRollback=0
ForkCheckBlockTime=0
ForkTxHeight=0
ForkTxGroupPara=0
ForkChainParamV2=0
ForkMultiSignAddress=0
ForkStateDBSet=0
ForkLocalDBAccess=0
ForkBlockCheck=0
ForkBase58AddressCheck=0
#平行链上使能平行链执行器如user.p.x.coins执行器的注册，缺省为0，对已有的平行链需要设置一个fork高度
ForkEnableParaRegExec=0
ForkCacheDriver=0
ForkTicketFundAddrV1=-1 #fork6.3
#主链和平行链都使用同一个fork高度
ForkRootHash=0 

[fork.sub.pos33]
Enable=0

[fork.sub.coins]
Enable=0

[fork.sub.ticket]
Enable=0
ForkTicketId =0
ForkTicketVrf =0

[fork.sub.retrieve]
Enable=0
ForkRetrive=0
ForkRetriveAsset=0

[fork.sub.hashlock]
Enable=0
ForkBadRepeatSecret=0

[fork.sub.manage]
Enable=0
ForkManageExec=0

[fork.sub.token]
Enable=0
ForkTokenBlackList= 0
ForkBadTokenSymbol= 0
ForkTokenPrice=0
ForkTokenSymbolWithNumber=0
ForkTokenCheck= 0

[fork.sub.trade]
Enable=0
ForkTradeBuyLimit= 0
ForkTradeAsset= 0
ForkTradeID = 0
ForkTradeFixAssetDB = 0
ForkTradePrice = 0

[fork.sub.paracross]
Enable=0
ForkParacrossWithdrawFromParachain=0
ForkParacrossCommitTx=0
ForkLoopCheckCommitTxDone=0
#仅平行链适用，自共识分阶段开启，缺省是0，若对应主链高度7000000之前开启过自共识，需要重新配置此分叉，并为之前自共识设置selfConsensEnablePreContract配置项
ForkParaSelfConsStages=0
ForkParaAssetTransferRbk=0

[fork.sub.evm]
Enable=0
ForkEVMState=0
ForkEVMABI=0
ForkEVMFrozen=0
ForkEVMKVHash=0

[fork.sub.blackwhite]
Enable=0
ForkBlackWhiteV2=0

[fork.sub.cert]
Enable=0

[fork.sub.guess]
Enable=0

[fork.sub.lottery]
Enable=0

[fork.sub.oracle]
Enable=0

[fork.sub.relay]
Enable=0

[fork.sub.norm]
Enable=0

[fork.sub.pokerbull]
Enable=0

[fork.sub.privacy]
Enable=0

[fork.sub.game]
Enable=0

[fork.sub.multisig]
Enable=0

[fork.sub.unfreeze]
Enable=0
ForkTerminatePart=0
ForkUnfreezeIDX= 0

[fork.sub.autonomy]
Enable=0

[fork.sub.jsvm]
Enable=0

[fork.sub.issuance]
Enable=0
ForkIssuanceTableUpdate=0

[fork.sub.collateralize]
Enable=0
ForkCollateralizeTableUpdate=0

#对已有的平行链如果不是从0开始同步数据，需要设置这个kvmvccmavl的对应平行链高度的fork，如果从0开始同步，statehash会跟以前mavl的不同
[fork.sub.store-kvmvccmavl]
ForkKvmvccmavl=0

[exec]
isFree=false
minExecFee=100000
maxExecFee=1000000000
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]

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

`

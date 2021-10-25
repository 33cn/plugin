package dpos

import (
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/33cn/chain33/system/p2p/dht/protocol"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	secureConnCrypto crypto.Crypto
	securePriv       crypto.PrivKey
	sum              = 0
	mutx             sync.Mutex
	privKey          = "B3DC4C0725884EBB7264B92F1D8D37584A64ADE1799D997EC64B4FE3973E08DE220ACBE680DF2473A0CB48987A00FCC1812F106A7390BE6B8E2D31122C992A19"
	expectAddress    = "02A13174B92727C4902DB099E51A3339F48BD45E"

	//localGenesis = `{"genesis_time":"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFj","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""},{"pub_key":{"type":"secp256k1","data":"027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"},"name":""},{"pub_key":{"type":"secp256k1","data":"03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"},"name":""}],"app_hash":null}`

	localGenesis = `{"genesis_time":"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFj","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""}],"app_hash":null}`
	localPriv    = `{"address":"2FA286246F0222C4FF93210E91AECE0C66723F15","pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"last_height":1679,"last_round":0,"last_step":3,"last_signature":{"type":"secp256k1","data":"37892A916D6E487ADF90F9E88FE37024597677B6C6FED47444AD582F74144B3D6E4B364EAF16AF03A4E42827B6D3C86415D734A5A6CCA92E114B23EB9265AF09"},"last_signbytes":"7B22636861696E5F6964223A22636861696E33332D5A326367466A222C22766F7465223A7B22626C6F636B5F6964223A7B2268617368223A224F6A657975396B2B4149426A6E4859456739584765356A7A462B673D222C227061727473223A7B2268617368223A6E756C6C2C22746F74616C223A307D7D2C22686569676874223A313637392C22726F756E64223A302C2274696D657374616D70223A22323031382D30382D33315430373A35313A34332E3935395A222C2274797065223A327D7D","priv_key":{"type":"secp256k1","data":"5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"}}`

	config1 = `Title="local"
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
defCacheSize=512
maxFetchBlockNum=128
timeoutSeconds=5
batchBlockNum=128
driver="leveldb"
dbPath="datadir1"
dbCache=64
isStrongConsistency=true
singleMode=true
batchsync=false
enableTxQuickIndex=true

[p2p]
types=["dht"]
enable=true
msgCacheSize=10240
driver="leveldb"
dbPath="datadir/addrbook"
dbCache=4
grpcLogFile="grpc33.log"


[rpc]
jrpcBindAddr="localhost:8801"
grpcBindAddr="localhost:8802"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]

[mempool]
name="timeline"
poolCacheSize=10240
minTxFeeRate=100000

[consensus]
name="dpos"
minerstart=false

[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
coinReward = 18
coinDevFund = 12
ticketPrice = 10000
powLimitBits = "0x1f00ffff"
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 5    #5s only for test
ticketWithdrawTime = 10 #10s only for test
ticketMinerWaitTime = 2 #2s only for test
maxTxNumber = 1600      #160
targetTimespan = 2304
targetTimePerBlock = 16

[mver.consensus.ForkChainParamV1]
maxTxNumber = 10000
targetTimespan = 288 #only for test
targetTimePerBlock = 2

[mver.consensus.ForkChainParamV2]
powLimitBits = "0x1f2fffff"

[consensus.sub.dpos]
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
genesisBlockTime=1514533394
timeoutCheckConnections=1000
timeoutVoting=3000
timeoutWaitNotify=2000
createEmptyBlocks=false
createEmptyBlocksInterval=0
validatorNodes=["127.0.0.1:36656"]
delegateNum=1
blockInterval=2
continueBlockNum=12
isValidator=true
port="36656"
#shuffleType为1表示使用固定出块顺序，为2表示使用vrf信息进行出块顺序洗牌
shuffleType=1
#是否更新topN，如果为true，根据下面几个配置项定期更新topN节点;如果为false，则一直使用初始配置的节点，不关注投票结果
whetherUpdateTopN=false
blockNumToUpdateDelegate=20000
registTopNHeightLimit=100
updateTopNHeightLimit=200

[store]
name="kvdb"
driver="leveldb"
dbPath="datadir1/mavltree"
dbCache=128

[store.sub.kvdb]
enableMavlPrefix=false
enableMVCC=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet1"
dbCache=16
signType="secp256k1"

[wallet.sub.ticket]
minerdisable=false
minerwhitelist=["*"]

[exec]
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]
saveTokenTxList=false

[exec.sub.cert]
# 是否启用证书验证和签名
enable=false
# 加密文件路径
cryptoPath="authdir/crypto"
# 带证书签名类型，支持"auth_ecdsa", "auth_sm2"
signType="auth_ecdsa"
`
	///////////////////////////////////////////////////////////////////////////
	config2 = `Title="local"
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
defCacheSize=512
maxFetchBlockNum=128
timeoutSeconds=5
batchBlockNum=128
driver="leveldb"
dbPath="datadir2"
dbCache=64
isStrongConsistency=true
singleMode=true
batchsync=false
enableTxQuickIndex=true

[p2p]
types=["dht"]
enable=true
msgCacheSize=10240
driver="leveldb"
dbPath="datadir2/addrbook"
dbCache=4
grpcLogFile="grpc33.log"



[rpc]
jrpcBindAddr="localhost:8803"
grpcBindAddr="localhost:8804"
whitelist=["127.0.0.1"]
jrpcFuncWhitelist=["*"]
grpcFuncWhitelist=["*"]

[mempool]
name="timeline"
poolCacheSize=10240
minTxFeeRate=100000

[consensus]
name="dpos"
minerstart=false

[mver.consensus]
fundKeyAddr = "1BQXS6TxaYYG5mADaWij4AxhZZUTpw95a5"
coinReward = 18
coinDevFund = 12
ticketPrice = 10000
powLimitBits = "0x1f00ffff"
retargetAdjustmentFactor = 4
futureBlockTime = 16
ticketFrozenTime = 5    #5s only for test
ticketWithdrawTime = 10 #10s only for test
ticketMinerWaitTime = 2 #2s only for test
maxTxNumber = 1600      #160
targetTimespan = 2304
targetTimePerBlock = 16

[mver.consensus.ForkChainParamV1]
maxTxNumber = 10000
targetTimespan = 288 #only for test
targetTimePerBlock = 2

[mver.consensus.ForkChainParamV2]
powLimitBits = "0x1f2fffff"

[consensus.sub.dpos]
genesis="14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"
genesisBlockTime=1514533394
timeoutCheckConnections=1000
timeoutVoting=3000
timeoutWaitNotify=2000
createEmptyBlocks=false
createEmptyBlocksInterval=0
validatorNodes=["127.0.0.1:36656"]
delegateNum=1
blockInterval=2
continueBlockNum=12
isValidator=true
port="36657"
#shuffleType为1表示使用固定出块顺序，为2表示使用vrf信息进行出块顺序洗牌
shuffleType=1
#是否更新topN，如果为true，根据下面几个配置项定期更新topN节点;如果为false，则一直使用初始配置的节点，不关注投票结果
whetherUpdateTopN=false
blockNumToUpdateDelegate=20000
registTopNHeightLimit=100
updateTopNHeightLimit=200

[store]
name="kvmvccmavl"
driver="leveldb"
dbPath="datadir2/mavltree"
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
pruneHeight=10000
# 是否使能mavl数据载入内存
enableMemTree=true
# 是否使能mavl叶子节点数据载入内存
enableMemVal=true
# 缓存close ticket数目，该缓存越大同步速度越快，最大设置到1500000
tkCloseCacheLen=100000

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet2"
dbCache=16
signType="secp256k1"

[wallet.sub.ticket]
minerdisable=false
minerwhitelist=["*"]

[exec]
enableStat=false
enableMVCC=false
alias=["token1:token","token2:token","token3:token"]
saveTokenTxList=false

[exec.sub.cert]
# 是否启用证书验证和签名
enable=false
# 加密文件路径
cryptoPath="authdir/crypto"
# 带证书签名类型，支持"auth_ecdsa", "auth_sm2"
signType="auth_ecdsa"
`
)

func init() {
	cr2, err := crypto.Load(types.GetSignName("", types.ED25519), -1)
	if err != nil {
		fmt.Println("crypto.Load failed for types.ED25519")
		return
	}
	secureConnCrypto = cr2

	securePriv, err = cr2.GenKey()
	if err != nil {
		fmt.Println("GenKey err", err)
		return
	}
}
func Init() {
	fmt.Println("=======Init Data1!=======")
	os.RemoveAll("datadir1")
	os.RemoveAll("datadir2")
	os.RemoveAll("wallet1")
	os.RemoveAll("wallet2")

	os.Remove("genesis.json")
	os.Remove("priv_validator.json")

	os.Remove("chain33.test1.toml")
	os.Remove("chain33.test2.toml")

	ioutil.WriteFile("genesis.json", []byte(localGenesis), 0664)
	ioutil.WriteFile("priv_validator.json", []byte(localPriv), 0664)
	ioutil.WriteFile("chain33.test1.toml", []byte(config1), 0664)
	ioutil.WriteFile("chain33.test2.toml", []byte(config2), 0664)
}

func TestParallel(t *testing.T) {
	Parallel(func() {
		mutx.Lock()
		sum++
		mutx.Unlock()
	},
		func() {
			mutx.Lock()
			sum += 2
			mutx.Unlock()
		},
		func() {
			mutx.Lock()
			sum += 3
			mutx.Unlock()
		},
	)

	fmt.Println("TestParallel ok")
	assert.Equal(t, 6, sum)
}

func TestGenAddressByPubKey(t *testing.T) {
	tmp, err := hex.DecodeString(privKey)
	require.Nil(t, err)

	priv, err := secureConnCrypto.PrivKeyFromBytes(tmp)
	require.Nil(t, err)

	addr := GenAddressByPubKey(priv.PubKey())
	strAddr := fmt.Sprintf("%X", addr)
	assert.Equal(t, expectAddress, strAddr)
	fmt.Println("TestGenAddressByPubKey ok")
}

func TestIP2IPPort(t *testing.T) {
	testMap := NewMutexMap()
	assert.Equal(t, false, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.1", "1.1.1.1:80")
	assert.Equal(t, true, testMap.Has("1.1.1.1"))

	testMap.Set("1.1.1.2", "1.1.1.2:80")
	assert.Equal(t, true, testMap.Has("1.1.1.2"))

	testMap.Delete("1.1.1.1")
	assert.Equal(t, false, testMap.Has("1.1.1.1"))
	fmt.Println("TestIP2IPPort ok")

}

func TestNode(t *testing.T) {
	fmt.Println("=======start TestNode!=======")
	Init()
	protocol.ClearEventHandler()
	q1, chain1, s1, mem1, exec1, cs1, p2p1 := initEnvDpos1("chain33.test1.toml")

	defer clearTestData1()
	defer chain1.Close()
	defer mem1.Close()
	defer exec1.Close()
	defer s1.Close()
	defer q1.Close()
	defer cs1.Close()
	defer p2p1.Close()

	time.Sleep(2 * time.Second)

	_, _, err := createConn("127.0.0.1:8802")
	for err != nil {
		_, _, err = createConn("127.0.0.1:8802")
	}

	fmt.Println("node1 ip:", cs1.(*Client).GetNode().IP)
	fmt.Println("node1 id:", cs1.(*Client).GetNode().ID)
	fmt.Println("node1 network:", cs1.(*Client).GetNode().Network)
	fmt.Println("node1 version:", cs1.(*Client).GetNode().Version)

	nodeinfo := NodeInfo{
		ID:      cs1.(*Client).GetNode().ID,
		IP:      cs1.(*Client).GetNode().IP,
		Network: cs1.(*Client).GetNode().Network,
		Version: cs1.(*Client).GetNode().Version,
	}

	require.Nil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))
	cs1.(*Client).GetNode().Version = "0.1"
	require.NotNil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))
	cs1.(*Client).GetNode().Version = nodeinfo.Version
	nodeinfo.Version = "0.1"
	require.NotNil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))
	nodeinfo.Version = "1.1.0"
	require.NotNil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))
	nodeinfo.Version = "0.0.0"
	require.Nil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))
	nodeinfo.Version = cs1.(*Client).GetNode().Version
	nodeinfo.Network = "chain33-Z2cgFi"
	require.NotNil(t, cs1.(*Client).GetNode().CompatibleWith(nodeinfo))

	fmt.Println("TestNodeCompatibleWith ok")

	fmt.Println(q1.Name())
	fmt.Println(cs1.(*Client).testFlag)
	fmt.Println(cs1.(*Client).GetConsensusState() != nil)
	fmt.Println(cs1.(*Client).GetConsensusState().String())
	fmt.Println(len(cs1.(*Client).GetConsensusState().GetValidators()) == 1)
	// fix dpos testcase err --- 写只会在共识初始化，正常运行不会有datarace
	//cs1.(*Client).GetConsensusState().SetPrivValidator(cs1.(*Client).GetConsensusState().GetPrivValidator(), cs1.(*Client).GetConsensusState().privValidatorIndex)
	fmt.Println(cs1.(*Client).GetConsensusState().GetValidatorMgr().ChainID)
	fmt.Println(cs1.(*Client).GetConsensusState().GetPrivValidator().GetAddress() != nil)
	fmt.Println(cs1.(*Client).GetConsensusState().IsProposer())
	fmt.Println(cs1.(*Client).PrivValidator().GetAddress() != nil)
	fmt.Println(cs1.(*Client).GenesisDoc().ChainID)
	fmt.Println("Validator index: ", cs1.(*Client).ValidatorIndex())

	time.Sleep(1 * time.Second)

	if cs1.(*Client).GetNode().IsRunning() {
		fmt.Println("=======cs1 is running=======")
	} else {
		fmt.Println("======= cs1 is not running=======")
	}

	fmt.Println("=======test state machine=======")
	vote := &ttypes.DPosVote{}
	InitStateObj.sendVote(cs1.(*Client).GetConsensusState(), vote)
	voteReply := &ttypes.DPosVoteReply{}
	InitStateObj.sendVoteReply(cs1.(*Client).GetConsensusState(), voteReply)
	InitStateObj.recvVoteReply(cs1.(*Client).GetConsensusState(), voteReply)
	notify := &ttypes.DPosNotify{}
	InitStateObj.sendNotify(cs1.(*Client).GetConsensusState(), notify)

	VotingStateObj.sendVoteReply(cs1.(*Client).GetConsensusState(), voteReply)
	VotingStateObj.sendNotify(cs1.(*Client).GetConsensusState(), notify)
	VotingStateObj.recvNotify(cs1.(*Client).GetConsensusState(), notify)

	VotedStateObj.sendVote(cs1.(*Client).GetConsensusState(), vote)

	WaitNotifyStateObj.sendVote(cs1.(*Client).GetConsensusState(), vote)
	WaitNotifyStateObj.sendVoteReply(cs1.(*Client).GetConsensusState(), voteReply)
	WaitNotifyStateObj.recvVoteReply(cs1.(*Client).GetConsensusState(), voteReply)
	WaitNotifyStateObj.sendNotify(cs1.(*Client).GetConsensusState(), notify)

	fmt.Println("=======testNode ok=======")
}

func clearTestData1() {
	fmt.Println("=======start clear test data1!=======")

	os.Remove("chain33.test1.toml")
	os.Remove("chain33.test2.toml")
	os.Remove("genesis.json")
	os.Remove("priv_validator.json")

	os.RemoveAll("wallet1")
	os.RemoveAll("wallet2")

	err := os.RemoveAll("datadir1")
	if err != nil {
		fmt.Println("delete datadir1 have a err:", err.Error())
	}
	err = os.RemoveAll("datadir2")
	if err != nil {
		fmt.Println("delete datadir2 have a err:", err.Error())
	}

	fmt.Println("test data clear successfully!")
}

func initEnvDpos1(configName string) (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module) {
	flag.Parse()
	chain33Cfg := types.NewChain33Config(types.ReadFile("chain33.test.toml"))
	var q = queue.New("channel")
	q.SetConfig(chain33Cfg)
	cfg := chain33Cfg.GetModuleConfig()
	cfg.Log.LogFile = ""
	sub := chain33Cfg.GetSubConfig()
	rpc.InitCfg(cfg.RPC)

	chain := blockchain.New(chain33Cfg)
	chain.SetQueueClient(q.Client())

	exec := executor.New(chain33Cfg)
	exec.SetQueueClient(q.Client())
	chain33Cfg.SetMinFee(0)
	s := store.New(chain33Cfg)
	s.SetQueueClient(q.Client())

	cs := New(cfg.Consensus, sub.Consensus["dpos"])
	cs.(*Client).SetTestFlag()
	cs.SetQueueClient(q.Client())

	mem := mempool.New(chain33Cfg)
	mem.SetQueueClient(q.Client())
	network := p2p.NewP2PMgr(chain33Cfg)

	network.SetQueueClient(q.Client())

	gapi := rpc.NewGRpcServer(q.Client(), nil)
	go gapi.Listen()

	return q, chain, s, mem, exec, cs, network
}

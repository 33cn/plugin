package commands

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	"github.com/33cn/plugin/plugin/consensus/dpos"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
	_ "github.com/33cn/plugin/plugin/store/init"
	"google.golang.org/grpc"
)

var (
	random    *rand.Rand
	loopCount = 10
	conn      *grpc.ClientConn
	c         types.Chain33Client
	strPubkey = "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"
	//pubkey    []byte

	genesisKey   = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	validatorKey = "5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"

	//genesisAddr   = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

	validatorAddr = "15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b"

	genesis = `{"genesis_time":"2018-08-16T15:38:56.951569432+08:00","chain_id":"chain33-Z2cgFj","validators":[{"pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"name":""},{"pub_key":{"type":"secp256k1","data":"027848E7FA630B759DB406940B5506B666A344B1060794BBF314EB459D40881BB3"},"name":""},{"pub_key":{"type":"secp256k1","data":"03F4AB6659E61E8512C9A24AC385CC1AC4D52B87D10ADBDF060086EA82BE62CDDE"},"name":""}],"app_hash":null}`
	priv    = `{"address":"2B226E6603E52C94715BA4E92080EEF236292E33","pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"last_height":1679,"last_round":0,"last_step":3,"last_signature":{"type":"secp256k1","data":"37892A916D6E487ADF90F9E88FE37024597677B6C6FED47444AD582F74144B3D6E4B364EAF16AF03A4E42827B6D3C86415D734A5A6CCA92E114B23EB9265AF09"},"last_signbytes":"7B22636861696E5F6964223A22636861696E33332D5A326367466A222C22766F7465223A7B22626C6F636B5F6964223A7B2268617368223A224F6A657975396B2B4149426A6E4859456739584765356A7A462B673D222C227061727473223A7B2268617368223A6E756C6C2C22746F74616C223A307D7D2C22686569676874223A313637392C22726F756E64223A302C2274696D657374616D70223A22323031382D30382D33315430373A35313A34332E3935395A222C2274797065223A327D7D","priv_key":{"type":"secp256k1","data":"5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"}}`
	config  = `Title="local"
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
dbPath="datadir"
dbCache=64
isStrongConsistency=true
singleMode=true
batchsync=false
enableTxQuickIndex=true


[p2p]
types=["dht"]
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
name="tendermint"
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
validatorNodes=["127.0.0.1:46656"]
delegateNum=1
port="36686"
blockInterval=2
continueBlockNum=12
isValidator=true
rpcAddr="http://localhost:9801"
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
dbPath="datadir/mavltree"
dbCache=128

[store.sub.kvdb]
enableMavlPrefix=false
enableMVCC=false

[wallet]
minFee=100000
driver="leveldb"
dbPath="wallet"
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

const fee = 1e6

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
	//pubkey, _ = hex.DecodeString(strPubkey)

	os.Remove("genesis.json")
	os.Remove("priv_validator.json")
	os.Remove("genesis_file.json")
	os.Remove("priv_validator_0.json")
	os.Remove("chain33.test.toml")

	ioutil.WriteFile("genesis.json", []byte(genesis), 0664)
	ioutil.WriteFile("priv_validator.json", []byte(priv), 0664)
	ioutil.WriteFile("chain33.test.toml", []byte(config), 0664)
}
func TestDposPerf(t *testing.T) {
	DposPerf()
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func DposPerf() {
	q, chain, s, mem, exec, cs, p2p, cmd := initEnvDpos()
	defer chain.Close()
	defer mem.Close()
	defer exec.Close()
	defer s.Close()
	defer q.Close()
	defer cs.Close()
	defer p2p.Close()
	err := createConn()
	for err != nil {
		err = createConn()
	}
	time.Sleep(2 * time.Second)
	for i := 0; i < loopCount; i++ {
		NormPut(q.GetConfig())
		time.Sleep(time.Second)
	}
	time.Sleep(2 * time.Second)

	testCmd(cmd)
	time.Sleep(2 * time.Second)
	os.Remove("genesis.json")
	os.Remove("priv_validator.json")
	os.Remove("genesis_file.json")
	os.Remove("priv_validator_0.json")
	os.Remove("chain33.test.toml")

}

func initEnvDpos() (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module, *cobra.Command) {
	flag.Parse()
	chain33Cfg := types.NewChain33Config(types.ReadFile("chain33.test.toml"))
	var q = queue.New("channel")
	q.SetConfig(chain33Cfg)
	cfg := chain33Cfg.GetModuleConfig()
	cfg.Log.LogFile = ""
	sub := chain33Cfg.GetSubConfig()

	chain := blockchain.New(chain33Cfg)
	chain.SetQueueClient(q.Client())

	exec := executor.New(chain33Cfg)
	exec.SetQueueClient(q.Client())
	chain33Cfg.SetMinFee(0)
	s := store.New(chain33Cfg)
	s.SetQueueClient(q.Client())

	cs := dpos.New(cfg.Consensus, sub.Consensus["dpos"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(chain33Cfg)
	mem.SetQueueClient(q.Client())
	network := p2p.NewP2PMgr(chain33Cfg)

	network.SetQueueClient(q.Client())

	rpc.InitCfg(cfg.RPC)
	gapi := rpc.NewGRpcServer(q.Client(), nil)
	go gapi.Listen()

	japi := rpc.NewJSONRPCServer(q.Client(), nil)
	go japi.Listen()

	cmd := DPosCmd()
	return q, chain, s, mem, exec, cs, network, cmd
}

func createConn() error {
	var err error
	url := "127.0.0.1:8802"
	fmt.Println("grpc url:", url)
	conn, err = grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	c = types.NewChain33Client(conn)
	//r = rand.New(rand.NewSource(types.Now().UnixNano()))
	return nil
}

func generateKey(i, valI int) string {
	key := make([]byte, valI)
	binary.PutUvarint(key[:10], uint64(valI))
	binary.PutUvarint(key[12:24], uint64(i))
	if _, err := rand.Read(key[24:]); err != nil {
		os.Exit(1)
	}
	return string(key)
}

func generateValue(i, valI int) string {
	value := make([]byte, valI)
	binary.PutUvarint(value[:16], uint64(i))
	binary.PutUvarint(value[32:128], uint64(i))
	if _, err := rand.Read(value[128:]); err != nil {
		os.Exit(1)
	}
	return string(value)
}

func getprivkey(key string) crypto.PrivKey {
	bkey, err := hex.DecodeString(key)
	if err != nil {
		panic(err)
	}
	priv, err := ttypes.ConsensusCrypto.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}

func NormPut(cfg *types.Chain33Config) {
	tx := prepareTxList(cfg)

	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		return
	}
}

func prepareTxList(cfg *types.Chain33Config) *types.Transaction {

	var key string
	var value string
	var i int

	key = generateKey(i, 32)
	value = generateValue(i, 180)

	nput := &pty.NormAction_Nput{Nput: &pty.NormPut{Key: []byte(key), Value: []byte(value)}}
	action := &pty.NormAction{Value: nput, Ty: pty.NormActionPut}
	tx := &types.Transaction{Execer: []byte("norm"), Payload: types.Encode(action), Fee: fee}
	tx.To = address.ExecAddress("norm")
	tx.Nonce = random.Int63()
	tx.ChainID = cfg.GetChainID()
	tx.Sign(types.SECP256K1, getprivkey(genesisKey))
	return tx
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	fmt.Println("test data clear successfully!")
}

func testCmd(cmd *cobra.Command) {
	var rootCmd = &cobra.Command{
		Use:   "chain33-cli",
		Short: "chain33 client tools",
	}

	chain33Cfg := types.NewChain33Config(types.ReadFile("chain33.test.toml"))
	types.SetCliSysParam(chain33Cfg.GetTitle(), chain33Cfg)

	rootCmd.PersistentFlags().String("title", chain33Cfg.GetTitle(), "get title name")
	rootCmd.PersistentFlags().String("rpc_laddr", "http://127.0.0.1:8802", "http url")
	rootCmd.AddCommand(cmd)

	rootCmd.SetArgs([]string{"dpos", "regist", "--address", validatorAddr, "--pubkey", strPubkey, "--ip", "127.0.0.1", "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cancelRegist", "--address", validatorAddr, "--pubkey", strPubkey, "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "reRegist", "--address", validatorAddr, "--pubkey", strPubkey, "--ip", "127.0.0.1", "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "candidatorQuery", "--type", "topN", "--top", "1"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "candidatorQuery", "--type", "pubkeys", "--pubkeys", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "voteQuery", "--address", validatorAddr, "--pubkeys", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vote", "--addr", validatorAddr, "--pubkey", strPubkey, "--votes", "60"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "init_keyfile", "--num", "1"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbRecord", "--cycle", "1000", "--hash", strPubkey, "--height", "60", "--privKey", validatorKey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "cycle", "--cycle", "1000"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "height", "--height", "1000"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "hash", "--hash", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfMRegist", "--cycle", "1000", "--m", "data1", "--pubkey", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfRPRegist", "--cycle", "1000", "--hash", "22a58fbbe8002939b7818184e663e6c57447f4354adba31ad3c7f556e153353c", "--proof", "5ed22d8c1cc0ad131c1c9f82daec7b99ff25ae5e717624b4a8cf60e0f3dca2c97096680cd8df0d9ed8662ce6513edf5d1676ad8d72b7e4f0e0de687bd38623f404eb085d28f5631207cf97a02c55f835bd3733241c7e068b80cf75e2afd12fd4c4cb8e6f630afa2b7b2918dff3d279e50acab59da1b25b3ff920b69c443da67320", "--pubkey", "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfQuery", "--type", "dtime", "--time", "2006-01-02 15:04:05"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfQuery", "--type", "timestamp", "--timestamp", "121211212"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfQuery", "--type", "cycle", "--cycle", "1000"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfQuery", "--type", "topN"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfQuery", "--type", "pubkeys", "--pubkeys", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfEvaluate", "--privKey", validatorKey, "--m", "input"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfVerify", "--pubkey", strPubkey, "--m", "input", "--hash", "3975b20c89894a3961dbab6cefb07ce7736761b4105931f268e47a99511eb635", "--proof", "6fe5e2d8a5de203da8f487459c5af24b1e56bf69848f4ca2f786eac5d4ad60bab35d83fc15b903b3007f570e8766942031ffed84d42e9bb3314d408fec557fd5043e72a99cf64ae29c89282367c473e0925e8bd841063d508264af5c6320faecb61692f8fbde47cd3b82d0e9804e30d89d13f2fafd1769fe32d9bb9750d943ddb4"})
	rootCmd.Execute()
}

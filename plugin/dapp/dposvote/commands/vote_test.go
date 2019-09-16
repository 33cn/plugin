package commands

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"

	"flag"
	"fmt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"



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
	priv = `{"address":"2B226E6603E52C94715BA4E92080EEF236292E33","pub_key":{"type":"secp256k1","data":"03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"},"last_height":1679,"last_round":0,"last_step":3,"last_signature":{"type":"secp256k1","data":"37892A916D6E487ADF90F9E88FE37024597677B6C6FED47444AD582F74144B3D6E4B364EAF16AF03A4E42827B6D3C86415D734A5A6CCA92E114B23EB9265AF09"},"last_signbytes":"7B22636861696E5F6964223A22636861696E33332D5A326367466A222C22766F7465223A7B22626C6F636B5F6964223A7B2268617368223A224F6A657975396B2B4149426A6E4859456739584765356A7A462B673D222C227061727473223A7B2268617368223A6E756C6C2C22746F74616C223A307D7D2C22686569676874223A313637392C22726F756E64223A302C2274696D657374616D70223A22323031382D30382D33315430373A35313A34332E3935395A222C2274797065223A327D7D","priv_key":{"type":"secp256k1","data":"5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"}}`
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


	ioutil.WriteFile("genesis.json", []byte(genesis), 0664)
	ioutil.WriteFile("priv_validator.json", []byte(priv), 0664)
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
		NormPut()
		time.Sleep(time.Second)
	}
	time.Sleep(2 * time.Second)

	testCmd(cmd)
	time.Sleep(2 * time.Second)
	os.Remove("genesis.json")
	os.Remove("priv_validator.json")
	os.Remove("genesis_file.json")
	os.Remove("priv_validator_0.json")

}

func initEnvDpos() (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module, *cobra.Command) {
	var q = queue.New("channel")
	flag.Parse()
	cfg, sub := types.InitCfg("chain33.test.toml")
	types.Init(cfg.Title, cfg)
	chain := blockchain.New(cfg.BlockChain)
	chain.SetQueueClient(q.Client())

	exec := executor.New(cfg.Exec, sub.Exec)
	exec.SetQueueClient(q.Client())
	types.SetMinFee(0)
	s := store.New(cfg.Store, sub.Store)
	s.SetQueueClient(q.Client())

	cs := dpos.New(cfg.Consensus, sub.Consensus["dpos"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(cfg.Mempool, nil)
	mem.SetQueueClient(q.Client())
	network := p2p.New(cfg.P2P)

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

func NormPut() {
	tx := prepareTxList()

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

func prepareTxList() *types.Transaction {
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

	rootCmd.PersistentFlags().String("rpc_laddr", "http://127.0.0.1:8802", "http url")
	rootCmd.AddCommand(cmd)

	rootCmd.SetArgs([]string{"dpos", "regist", "--address", validatorAddr, "--pubkey",strPubkey, "--ip", "127.0.0.1", "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cancelRegist", "--address", validatorAddr, "--pubkey",strPubkey, "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "reRegist", "--address", validatorAddr, "--pubkey",strPubkey, "--ip", "127.0.0.1", "--rpc_laddr", "http://127.0.0.1:8801"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "candidatorQuery", "--type", "topN", "--top","1"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "candidatorQuery", "--type", "pubkeys", "--pubkeys",strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "voteQuery", "--address", validatorAddr, "--pubkeys",strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vote", "--addr", validatorAddr, "--pubkey",strPubkey, "--votes", "60"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "init_keyfile", "--num", "1"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbRecord", "--cycle", "1000", "--hash",strPubkey, "--height", "60", "--privKey", validatorKey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "cycle", "--cycle", "1000"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "height", "--height", "1000"})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "cbQuery", "--type", "hash", "--hash", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfMRegist", "--cycle", "1000", "--m", "data1", "--pubkey", strPubkey})
	rootCmd.Execute()

	rootCmd.SetArgs([]string{"dpos", "vrfRPRegist", "--cycle", "1000", "--hash", "22a58fbbe8002939b7818184e663e6c57447f4354adba31ad3c7f556e153353c","--proof", "5ed22d8c1cc0ad131c1c9f82daec7b99ff25ae5e717624b4a8cf60e0f3dca2c97096680cd8df0d9ed8662ce6513edf5d1676ad8d72b7e4f0e0de687bd38623f404eb085d28f5631207cf97a02c55f835bd3733241c7e068b80cf75e2afd12fd4c4cb8e6f630afa2b7b2918dff3d279e50acab59da1b25b3ff920b69c443da67320",  "--pubkey", "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"})
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
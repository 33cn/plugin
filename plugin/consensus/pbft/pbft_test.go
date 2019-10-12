// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import (
	"flag"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"encoding/binary"
	"os"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/store"
	"github.com/33cn/chain33/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	random    *rand.Rand
	txNumber  = 5
	loopCount = 20
)

var zeroHash [32]byte

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
}

func TestWindows(t *testing.T) {
	OnePbft("Windows.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func TestClient(t *testing.T) {
	OnePbft("Client.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func TestPbft1(t *testing.T) {
	OnePbft("Replica_1.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}
func TestPbft2(t *testing.T) {
	OnePbft("Replica_2.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}
func TestPbft3(t *testing.T) {
	OnePbft("Replica_3.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}
func TestPbft4(t *testing.T) {
	OnePbft("Replica_4.test.toml")
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

// OnePbft 用于测试节点动态
func OnePbft(cfgName string) {
	q, chain, s, mem, exec, cs, p2p := initEnvPbft(cfgName)
	defer chain.Close()
	defer mem.Close()
	defer exec.Close()
	defer s.Close()
	defer cs.Close()
	defer q.Close()
	defer p2p.Close()
	time.Sleep(1 * time.Second)

	if cfgName == "Client.test.toml" || cfgName == "Windows.test.toml" {
		sendReplyList(q)
	} else {
		for {
			time.Sleep(12000 * time.Second)
			break
		}
	}
}

// 初始化pbft环境
func initEnvPbft(cfgName string) (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module) {
	var q = queue.New("channel")
	flag.Parse()

	cfg, sub := types.InitCfg(cfgName)
	types.Init(cfg.Title, cfg)

	chain := blockchain.New(cfg.BlockChain)
	chain.SetQueueClient(q.Client())

	exec := executor.New(cfg.Exec, sub.Exec)
	exec.SetQueueClient(q.Client())
	types.SetMinFee(0)

	s := store.New(cfg.Store, sub.Store)
	s.SetQueueClient(q.Client())

	cs := NewPbftNode(cfg.Consensus, sub.Consensus["pbft"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(cfg.Mempool, nil)
	mem.SetQueueClient(q.Client())

	network := p2p.New(cfg.P2P)
	network.SetQueueClient(q.Client())

	return q, chain, s, mem, exec, cs, network
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
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	bkey, err := common.FromHex(key)
	if err != nil {
		panic(err)
	}
	priv, err := cr.PrivKeyFromBytes(bkey)
	if err != nil {
		panic(err)
	}
	return priv
}

func sendReplyList(q queue.Queue) {
	client := q.Client()
	client.Sub("mempool")
	var count int
	for msg := range client.Recv() {
		if msg.Ty == types.EventTxList {
			count++
			msg.Reply(client.NewMessage("consensus", types.EventReplyTxList,
				&types.ReplyTxList{Txs: getReplyList(txNumber)}))
			if count >= loopCount {
				time.Sleep(4 * time.Second)
				break
			}
		}
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
	tx := &types.Transaction{Execer: []byte("norm"), Payload: types.Encode(action), Fee: 0}
	tx.To = address.ExecAddress("norm")
	tx.Nonce = random.Int63()
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))
	return tx
}

func getReplyList(n int) (txs []*types.Transaction) {

	for i := 0; i < int(n); i++ {
		txs = append(txs, prepareTxList())
	}
	return txs
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	err = os.RemoveAll("chain33_pbft-1")
	if err != nil {
		fmt.Println("delete chain33_pbft dir have a err:", err.Error())
	}
	fmt.Println("test data clear sucessfully!")
}

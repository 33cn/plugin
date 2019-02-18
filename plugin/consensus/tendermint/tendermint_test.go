// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tendermint

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/store"
	_ "github.com/33cn/chain33/system"
	"github.com/33cn/chain33/types"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	random    *rand.Rand
	txNumber  = 10
	loopCount = 10
)

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
}

func TestTendermintPerf(t *testing.T) {
	TendermintPerf()
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func TendermintPerf() {
	q, chain, s, exec, cs, p2p := initEnvTendermint()
	defer chain.Close()
	defer exec.Close()
	defer s.Close()
	defer q.Close()
	defer cs.Close()
	defer p2p.Close()
	sendReplyList(q)
}

func initEnvTendermint() (queue.Queue, *blockchain.BlockChain, queue.Module, *executor.Executor, queue.Module, queue.Module) {
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

	cs := New(cfg.Consensus, sub.Consensus["tendermint"])
	cs.SetQueueClient(q.Client())

	network := p2p.New(cfg.P2P)
	network.SetQueueClient(q.Client())

	return q, chain, s, exec, cs, network
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

func sendReplyList(q queue.Queue) {
	client := q.Client()
	client.Sub("mempool")
	var count int
	for msg := range client.Recv() {
		if msg.Ty == types.EventTxList {
			count++
			msg.Reply(client.NewMessage("consensus", types.EventReplyTxList,
				&types.ReplyTxList{Txs: getReplyList(txNumber)}))
			//tendermint call RequestTx twice
			if count >= loopCount*2 {
				time.Sleep(4 * time.Second)
				break
			}
		}
	}
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	fmt.Println("test data clear successfully!")
}

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/33cn/chain33/common/crypto"

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/p2p"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc"
	"github.com/33cn/chain33/store"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
	"google.golang.org/grpc"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	random    *rand.Rand
	loopCount = 10
	conn      *grpc.ClientConn
	c         types.Chain33Client
	strPubkey = "03EF0E1D3112CF571743A3318125EDE2E52A4EB904BCBAA4B1F75020C2846A7EB4"
	pubkey    []byte
)

const fee = 1e6

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	random = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
	pubkey, _ = hex.DecodeString(strPubkey)
}
func TestDposPerf(t *testing.T) {
	DposPerf()
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func DposPerf() {
	q, chain, s, mem, exec, cs, p2p := initEnvDpos()
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
	time.Sleep(10 * time.Second)
	for i := 0; i < loopCount; i++ {
		NormPut()
		time.Sleep(time.Second)
	}
	time.Sleep(10 * time.Second)
	now := time.Now().Unix()
	task := DecideTaskByTime(now)

	dposClient := cs.(*Client)
	dposClient.csState.QueryCycleBoundaryInfo(task.Cycle)
	dposClient.csState.GetCBInfoByCircle(task.Cycle)
	dposClient.csState.QueryVrf(pubkey, task.Cycle)
	dposClient.csState.QueryVrfs(dposClient.csState.validatorMgr.Validators, task.Cycle)
	dposClient.csState.GetVrfInfoByCircle(task.Cycle, VrfQueryTypeM)
	dposClient.csState.GetVrfInfoByCircle(task.Cycle, VrfQueryTypeRP)
	dposClient.csState.GetVrfInfosByCircle(task.Cycle)
	input := []byte("data1")
	hash, proof := dposClient.csState.VrfEvaluate(input)
	if dposClient.csState.VrfProof(pubkey, input, hash, proof) {
		fmt.Println("VrfProof ok")
	}
	dposClient.QueryTopNCandidators(1)

	time.Sleep(1 * time.Second)
	info := &dty.DposCBInfo{
		Cycle:      task.Cycle,
		StopHeight: 10,
		StopHash:   "absadfafa",
		Pubkey:     strPubkey,
	}
	if dposClient.csState.SendCBTx(info) {
		fmt.Println("sendCBTx ok")
	} else {
		fmt.Println("sendCBTx failed")
	}
	time.Sleep(2 * time.Second)
	info2 := dposClient.csState.GetCBInfoByCircle(task.Cycle)
	if info2 != nil && info2.StopHeight == info.StopHeight {
		fmt.Println("GetCBInfoByCircle ok")
	} else {
		fmt.Println("GetCBInfoByCircle failed")
	}
	time.Sleep(1 * time.Second)
	for {
		now = time.Now().Unix()
		task = DecideTaskByTime(now)
		if now < task.CycleStart+(task.CycleStop-task.CycleStart)/2 {
			break
		}
		time.Sleep(1 * time.Second)
	}

	vrfM := &dty.DposVrfMRegist{
		Pubkey: strPubkey,
		Cycle:  task.Cycle,
		M:      "absadfafa",
	}
	if dposClient.csState.SendRegistVrfMTx(vrfM) {
		fmt.Println("SendRegistVrfMTx ok")
	} else {
		fmt.Println("SendRegistVrfMTx failed")
	}
	time.Sleep(2 * time.Second)
	vrfInfo, err := dposClient.csState.QueryVrf(pubkey, task.Cycle)
	if err != nil || vrfInfo == nil {
		fmt.Println("QueryVrf failed")
	} else {
		fmt.Println("QueryVrf ok,", vrfInfo.Cycle, "|", len(vrfInfo.M))
	}

	for {
		now = time.Now().Unix()
		task = DecideTaskByTime(now)
		if now > task.CycleStart+(task.CycleStop-task.CycleStart)/2 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	vrfRP := &dty.DposVrfRPRegist{
		Pubkey: strPubkey,
		Cycle:  task.Cycle,
		R:      "Rabsadfafa",
		P:      "Pabsadfafa",
	}
	if dposClient.csState.SendRegistVrfRPTx(vrfRP) {
		fmt.Println("SendRegistVrfRPTx ok")
	} else {
		fmt.Println("SendRegistVrfRPTx failed")
	}
	time.Sleep(2 * time.Second)
	vrfInfo, err = dposClient.csState.QueryVrf(pubkey, task.Cycle)
	if err != nil || vrfInfo == nil {
		fmt.Println("QueryVrf failed")
	} else {
		fmt.Println("QueryVrf ok,", vrfInfo.Cycle, "|", len(vrfInfo.M), "|", len(vrfInfo.R), "|", len(vrfInfo.P))
	}
	time.Sleep(2 * time.Second)
}

func initEnvDpos() (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module) {
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

	cs := New(cfg.Consensus, sub.Consensus["dpos"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(cfg.Mempool, nil)
	mem.SetQueueClient(q.Client())
	network := p2p.New(cfg.P2P)

	network.SetQueueClient(q.Client())

	rpc.InitCfg(cfg.RPC)
	gapi := rpc.NewGRpcServer(q.Client(), nil)
	go gapi.Listen()
	return q, chain, s, mem, exec, cs, network
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
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))
	return tx
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	fmt.Println("test data clear successfully!")
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

// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tendermint

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

	"github.com/33cn/chain33/blockchain"
	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/common/limits"
	"github.com/33cn/chain33/common/log"
	"github.com/33cn/chain33/executor"
	"github.com/33cn/chain33/mempool"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc"
	"github.com/33cn/chain33/store"
	mty "github.com/33cn/chain33/system/dapp/manage/types"
	"github.com/33cn/chain33/types"
	ty "github.com/33cn/plugin/plugin/consensus/tendermint/types"
	pty "github.com/33cn/plugin/plugin/dapp/norm/types"
	vty "github.com/33cn/plugin/plugin/dapp/valnode/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	_ "github.com/33cn/chain33/system"
	_ "github.com/33cn/plugin/plugin/dapp/init"
	_ "github.com/33cn/plugin/plugin/store/init"
)

var (
	r         *rand.Rand
	loopCount = 3
	conn      *grpc.ClientConn
	c         types.Chain33Client
)

func init() {
	err := limits.SetLimits()
	if err != nil {
		panic(err)
	}
	r = rand.New(rand.NewSource(types.Now().UnixNano()))
	log.SetLogLevel("info")
}
func TestTendermintPerf(t *testing.T) {
	TendermintPerf(t)
	fmt.Println("=======start clear test data=======")
	clearTestData()
}

func TendermintPerf(t *testing.T) {
	q, chain, s, mem, exec, cs := initEnvTendermint()
	defer chain.Close()
	defer mem.Close()
	defer exec.Close()
	defer s.Close()
	defer q.Close()
	defer cs.Close()
	err := createConn()
	for err != nil {
		err = createConn()
	}
	time.Sleep(2 * time.Second)
	ConfigManager()
	for i := 0; i < loopCount; i++ {
		NormPut(q.GetConfig().GetChainID())
		time.Sleep(time.Second)
	}
	CheckState(t, cs.(*Client))
	AddNode()
	for i := 0; i < loopCount*3; i++ {
		NormPut(q.GetConfig().GetChainID())
		time.Sleep(time.Second)
	}
	time.Sleep(2 * time.Second)
}

func initEnvTendermint() (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module) {
	flag.Parse()
	chain33Cfg := types.NewChain33Config(types.ReadFile("chain33.test.toml"))
	var q = queue.New("channel")
	q.SetConfig(chain33Cfg)
	cfg := chain33Cfg.GetModuleConfig()
	sub := chain33Cfg.GetSubConfig()

	chain := blockchain.New(chain33Cfg)
	chain.SetQueueClient(q.Client())

	exec := executor.New(chain33Cfg)
	exec.SetQueueClient(q.Client())
	chain33Cfg.SetMinFee(0)
	s := store.New(chain33Cfg)
	s.SetQueueClient(q.Client())

	cs := New(cfg.Consensus, sub.Consensus["tendermint"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(chain33Cfg)
	mem.SetQueueClient(q.Client())

	rpc.InitCfg(cfg.RPC)
	gapi := rpc.NewGRpcServer(q.Client(), nil)
	go gapi.Listen()
	return q, chain, s, mem, exec, cs
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

func prepareTxList(chainid int32) *types.Transaction {
	var key string
	var value string
	var i int

	key = generateKey(i, 32)
	value = generateValue(i, 180)

	nput := &pty.NormAction_Nput{Nput: &pty.NormPut{Key: []byte(key), Value: []byte(value)}}
	action := &pty.NormAction{Value: nput, Ty: pty.NormActionPut}
	tx := &types.Transaction{Execer: []byte("norm"), Payload: types.Encode(action), Fee: fee}
	tx.To = address.ExecAddress("norm")
	tx.Nonce = r.Int63()
	tx.ChainID = chainid
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))
	return tx
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir err:", err.Error())
	}
	fmt.Println("clear test data successfully!")
}

func NormPut(chainid int32) {
	tx := prepareTxList(chainid)

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

func AddNode() {
	pubkey := "788657125A5A547B499F8B74239092EBB6466E8A205348D9EA645D510235A671"
	pubkeybyte, err := hex.DecodeString(pubkey)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	nput := &vty.ValNodeAction_Node{Node: &vty.ValNode{PubKey: pubkeybyte, Power: int64(2)}}
	action := &vty.ValNodeAction{Value: nput, Ty: vty.ValNodeActionUpdate}
	tx := &types.Transaction{Execer: []byte("valnode"), Payload: types.Encode(action), Fee: fee}
	tx.To = address.ExecAddress("valnode")
	tx.Nonce = r.Int63()
	version, _ := c.Version(context.Background(), nil)
	if version != nil {
		tx.ChainID = version.ChainID
	}
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))

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

func ConfigManager() {
	v := &types.ModifyConfig{Key: "tendermint-manager", Op: "add", Value: "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt", Addr: ""}
	modify := &mty.ManageAction{
		Ty:    mty.ManageActionModifyConfig,
		Value: &mty.ManageAction_Modify{Modify: v},
	}
	tx := &types.Transaction{Execer: []byte("manage"), Payload: types.Encode(modify), Fee: fee}
	tx.To = address.ExecAddress("manage")
	tx.Nonce = r.Int63()
	version, _ := c.Version(context.Background(), nil)
	if version != nil {
		tx.ChainID = version.ChainID

	}
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))

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

func CheckState(t *testing.T, client *Client) {
	state := client.csState.GetState()
	assert.NotEmpty(t, state)
	_, curVals := state.GetValidators()
	assert.NotEmpty(t, curVals)
	assert.True(t, state.Equals(state.Copy()))

	_, vals := client.csState.GetValidators()
	assert.Len(t, vals, 1)

	storeHeight := client.csStore.LoadStateHeight()
	assert.True(t, storeHeight > 0)

	sc := client.csState.LoadCommit(storeHeight)
	assert.NotEmpty(t, sc)
	bc := client.csState.LoadCommit(storeHeight - 1)
	assert.NotEmpty(t, bc)

	assert.NotEmpty(t, client.LoadBlockState(storeHeight))
	assert.NotEmpty(t, client.LoadProposalBlock(storeHeight))

	assert.Nil(t, client.LoadBlockCommit(0))
	assert.Nil(t, client.LoadBlockState(0))
	assert.Nil(t, client.LoadProposalBlock(0))

	csdb := client.csState.blockExec.db
	assert.NotEmpty(t, csdb)
	assert.NotEmpty(t, csdb.LoadState())
	if storeHeight > 1 {
		valset, err := csdb.LoadValidators(storeHeight - 1)
		assert.Nil(t, err)
		assert.NotEmpty(t, valset)
	}

	genState, err := MakeGenesisStateFromFile("genesis.json")
	assert.Nil(t, err)
	assert.Equal(t, genState.LastBlockHeight, int64(0))

	assert.Equal(t, client.csState.Prevote(0), 1000*time.Millisecond)
	assert.Equal(t, client.csState.Precommit(0), 1000*time.Millisecond)
	assert.Equal(t, client.csState.PeerGossipSleep(), 100*time.Millisecond)
	assert.Equal(t, client.csState.PeerQueryMaj23Sleep(), 2000*time.Millisecond)
	assert.Equal(t, client.csState.IsProposer(), true)
	assert.Nil(t, client.csState.GetPrevotesState(state.LastBlockHeight, 0, nil))
	assert.Nil(t, client.csState.GetPrecommitsState(state.LastBlockHeight, 0, nil))
	assert.Len(t, client.GenesisDoc().Validators, 1)

	msg1, err := client.Query_IsHealthy(&types.ReqNil{})
	assert.Nil(t, err)
	flag := msg1.(*vty.IsHealthy).IsHealthy
	assert.Equal(t, true, flag)

	msg2, err := client.Query_NodeInfo(&types.ReqNil{})
	assert.Nil(t, err)
	tvals := msg2.(*vty.ValNodeInfoSet).Nodes
	assert.Len(t, tvals, 1)

	err = client.CommitBlock(client.GetCurrentBlock())
	assert.Nil(t, err)
}

func TestCompareHRS(t *testing.T) {
	assert.Equal(t, CompareHRS(1, 1, ty.RoundStepNewHeight, 1, 1, ty.RoundStepNewHeight), 0)

	assert.Equal(t, CompareHRS(1, 1, ty.RoundStepPrevote, 2, 1, ty.RoundStepNewHeight), -1)
	assert.Equal(t, CompareHRS(1, 1, ty.RoundStepPrevote, 1, 2, ty.RoundStepNewHeight), -1)
	assert.Equal(t, CompareHRS(1, 1, ty.RoundStepPrevote, 1, 1, ty.RoundStepPrecommit), -1)

	assert.Equal(t, CompareHRS(2, 1, ty.RoundStepNewHeight, 1, 1, ty.RoundStepPrevote), 1)
	assert.Equal(t, CompareHRS(1, 2, ty.RoundStepNewHeight, 1, 1, ty.RoundStepPrevote), 1)
	assert.Equal(t, CompareHRS(1, 1, ty.RoundStepPrecommit, 1, 1, ty.RoundStepPrevote), 1)
	fmt.Println("TestCompareHRS ok")
}

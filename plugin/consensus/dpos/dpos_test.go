// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/33cn/chain33/util"

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
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	_ "github.com/33cn/plugin/plugin/dapp/init"
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
	pubkey    []byte

	genesisKey   = "CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"
	validatorKey = "5A6A14DA6F5A42835E529D75D87CC8904544F59EEE5387A37D87EEAD194D7EB2"

	//genesisAddr   = "14KEKbYtKKQm4wMthSK9J4La4nAiidGozt"

	validatorAddr = "15LsTP6tkYGZcN7tc1Xo2iYifQfowxot3b"
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

func TestPrint(t *testing.T) {
	fmt.Println(printNodeIPs([]string{"127.0.0.1:35566", "192.168.0.1:8080"}))

	var cands []*dty.Candidator
	cand := &dty.Candidator{
		Pubkey:  pubkey,
		Address: validatorAddr,
		IP:      "127.0.0.1",
		Votes:   60,
		Status:  0,
	}

	cands = append(cands, cand)
	cand = &dty.Candidator{
		Pubkey:  pubkey,
		Address: validatorAddr,
		IP:      "127.0.0.2",
		Votes:   60,
		Status:  0,
	}
	cands = append(cands, cand)
	fmt.Println(printCandidators(cands))
}
func TestDposPerf(t *testing.T) {

	os.Remove("genesis.json")
	os.Remove("priv_validator.json")
	ioutil.WriteFile("genesis.json", []byte(localGenesis), 0664)
	ioutil.WriteFile("priv_validator.json", []byte(localPriv), 0664)

	DposPerf()
	fmt.Println("=======start clear test data!=======")
	clearTestData()
}

func DposPerf() {
	fmt.Println("=======start dpos test!=======")
	q, chain, s, mem, exec, cs, p2p := initEnvDpos()
	cfg := q.GetConfig()
	defer chain.Close()
	defer mem.Close()
	defer exec.Close()
	defer s.Close()
	defer q.Close()
	defer cs.Close()
	defer p2p.Close()
	err := createConn2()
	for err != nil {
		err = createConn2()
	}
	time.Sleep(10 * time.Second)
	fmt.Println("=======start NormPut!=======")

	for i := 0; i < loopCount; i++ {
		NormPut(cfg)
		time.Sleep(time.Second)
	}

	fmt.Println("=======start sendTransferTx sendTransferToExecTx!=======")
	//从创世地址向测试地址转入代币
	sendTransferTx(cfg, genesisKey, validatorAddr, 2000000000000)
	time.Sleep(3 * time.Second)
	in := &types.ReqBalance{}
	in.Addresses = append(in.Addresses, validatorAddr)
	acct, err := c.GetBalance(context.Background(), in)
	if err != nil || len(acct.Acc) == 0 {
		fmt.Println("no balance for ", validatorAddr)
	} else {
		fmt.Println(validatorAddr, " balance:", acct.Acc[0].Balance, "frozen:", acct.Acc[0].Frozen)
	}
	//从测试地址向dos合约转入代币
	sendTransferToExecTx(cfg, validatorKey, "dpos", 1600000000000)

	time.Sleep(3 * time.Second)

	fmt.Println("=======start GetBalance!=======")

	in2 := &types.ReqBalance{}
	in2.Addresses = append(in2.Addresses, validatorAddr)
	acct, err = c.GetBalance(context.Background(), in2)
	if err != nil || len(acct.Acc) == 0 {
		fmt.Println("no balance for ", validatorAddr)
	} else {
		fmt.Println(validatorAddr, " balance:", acct.Acc[0].Balance, "frozen:", acct.Acc[0].Frozen)
	}

	fmt.Println("=======start sendRegistCandidatorTx!=======")

	sendRegistCandidatorTx(cfg, strPubkey, validatorAddr, "127.0.0.1", validatorKey)

	time.Sleep(3 * time.Second)

	fmt.Println("=======start query many things!=======")

	now := time.Now().Unix()
	task := DecideTaskByTime(now)

	dposClient := cs.(*Client)
	dposClient.csState.QueryCycleBoundaryInfo(task.Cycle)
	//dposClient.csState.GetCBInfoByCircle(task.Cycle)
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
	fmt.Println("=======start SendCBTx!=======")

	info := &dty.DposCBInfo{
		Cycle:      task.Cycle,
		StopHeight: dposClient.GetCurrentHeight(),
		StopHash:   hex.EncodeToString(dposClient.GetCurrentBlock().Hash(cfg)),
		Pubkey:     strPubkey,
	}
	dposClient.csState.SendCBTx(info)

	time.Sleep(4 * time.Second)
	fmt.Println("=======start verifyCB!=======")
	if verifyCB(dposClient.csState, info) {
		fmt.Println("Verify CB ok.")
	} else {
		fmt.Println("Verify CB failed.")
	}

	fmt.Println("=======start VoteVerify!=======")
	vote := generateVote(dposClient.csState)
	if nil == vote {
		fmt.Println("generateVote failed.")
	} else {
		fmt.Println("Vote:\n", vote.String())
		if err := dposClient.csState.privValidator.SignVote(dposClient.csState.validatorMgr.ChainID, vote); err != nil {
			fmt.Println("SignVote failed")
		} else {
			if dposClient.csState.VerifyVote(vote.DPosVote) {
				fmt.Println("Verify Vote ok.")
			} else {
				fmt.Println("Verify Vote failed.")
			}
		}
	}

	fmt.Println("=======start NotifyVerify!=======")
	block := dposClient.GetCurrentBlock()
	notify := &ttypes.Notify{
		DPosNotify: &ttypes.DPosNotify{
			Vote: &ttypes.VoteItem{
				VotedNodeIndex: vote.VoteItem.VotedNodeIndex,
				Cycle:          vote.VoteItem.Cycle,
				CycleStart:     vote.VoteItem.CycleStart,
				CycleStop:      vote.VoteItem.CycleStop,
				PeriodStart:    vote.VoteItem.PeriodStart,
				PeriodStop:     vote.VoteItem.PeriodStop,
				Height:         vote.VoteItem.Height,
				ShuffleType:    vote.VoteItem.ShuffleType,
			},
			HeightStop:        block.Height,
			HashStop:          block.Hash(cfg),
			NotifyTimestamp:   now,
			NotifyNodeAddress: vote.VoteItem.VotedNodeAddress,
			NotifyNodeIndex:   vote.VoteItem.VotedNodeIndex,
		},
	}

	notify.DPosNotify.Vote.VotedNodeAddress = make([]byte, len(vote.VoteItem.VotedNodeAddress))
	copy(notify.DPosNotify.Vote.VotedNodeAddress, vote.VoteItem.VotedNodeAddress)
	notify.DPosNotify.Vote.VoteID = make([]byte, len(vote.VoteItem.VoteID))
	copy(notify.DPosNotify.Vote.VoteID, vote.VoteItem.VoteID)

	if err := dposClient.csState.privValidator.SignNotify(dposClient.csState.validatorMgr.ChainID, notify); err != nil {
		fmt.Println("SignNotify failed.")
	} else {
		fmt.Println("Notify:\n", notify.String())
		if dposClient.csState.VerifyNotify(notify.DPosNotify) {
			fmt.Println("Verify Notify ok.")
		} else {
			fmt.Println("Verify Notify failed.")
		}
	}

	fmt.Println("=======start SendRegistVrfMTx!=======")

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
		M:      "data1",
	}
	if dposClient.csState.SendRegistVrfMTx(vrfM) {
		fmt.Println("SendRegistVrfMTx ok")
	} else {
		fmt.Println("SendRegistVrfMTx failed")
	}
	sendRegistVrfMTx(dposClient.csState, vrfM)

	time.Sleep(2 * time.Second)

	fmt.Println("=======start QueryVrf!=======")

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

	fmt.Println("=======start SendRegistVrfRPTx!=======")

	vrfRP := &dty.DposVrfRPRegist{
		Pubkey: strPubkey,
		Cycle:  task.Cycle,
		R:      "22a58fbbe8002939b7818184e663e6c57447f4354adba31ad3c7f556e153353c",
		P:      "5ed22d8c1cc0ad131c1c9f82daec7b99ff25ae5e717624b4a8cf60e0f3dca2c97096680cd8df0d9ed8662ce6513edf5d1676ad8d72b7e4f0e0de687bd38623f404eb085d28f5631207cf97a02c55f835bd3733241c7e068b80cf75e2afd12fd4c4cb8e6f630afa2b7b2918dff3d279e50acab59da1b25b3ff920b69c443da67320",
	}
	if dposClient.csState.SendRegistVrfRPTx(vrfRP) {
		fmt.Println("SendRegistVrfRPTx ok")
	} else {
		fmt.Println("SendRegistVrfRPTx failed")
	}

	sendRegistVrfRPTx(dposClient.csState, vrfRP)

	time.Sleep(2 * time.Second)
	fmt.Println("=======start QueryVrf2!=======")

	vrfInfo, err = dposClient.csState.QueryVrf(pubkey, task.Cycle)
	if err != nil || vrfInfo == nil {
		fmt.Println("QueryVrf failed")
	} else {
		fmt.Println("QueryVrf ok,", vrfInfo.Cycle, "|", len(vrfInfo.M), "|", len(vrfInfo.R), "|", len(vrfInfo.P))
	}

	fmt.Println("=======start QueryVrfInfos!=======")
	var pubkeys [][]byte
	pubkeys = append(pubkeys, pubkey)
	vrfInfos, err := dposClient.QueryVrfInfos(pubkeys, task.Cycle)
	if err != nil || vrfInfos == nil {
		fmt.Println("QueryVrf failed")
	} else {
		fmt.Println("QueryVrf ok,", vrfInfos[0].Cycle, "|", len(vrfInfos[0].M), "|", len(vrfInfos[0].R), "|", len(vrfInfos[0].P))
	}

	time.Sleep(2 * time.Second)
	fmt.Println("=======start SendTopNRegistTx!=======")

	var cands []*dty.Candidator
	cand := &dty.Candidator{
		Pubkey:  pubkey,
		Address: hex.EncodeToString(dposClient.csState.privValidator.GetAddress()),
		IP:      "127.0.0.1",
		Votes:   100,
		Status:  0,
	}
	cands = append(cands, cand)
	topNCand := &dty.TopNCandidator{
		Cands:        cands,
		Hash:         []byte("abafasfda"),
		Height:       dposClient.GetCurrentHeight(),
		SignerPubkey: pubkey,
	}
	reg := &dty.TopNCandidatorRegist{
		Cand: topNCand,
	}

	if dposClient.csState.SendTopNRegistTx(reg) {
		fmt.Println("SendTopNRegistTx ok")
	} else {
		fmt.Println("SendTopNRegistTx failed")
	}

	time.Sleep(2 * time.Second)
	fmt.Println("=======start QueryTopNCandidators!=======")

	dposClient.QueryTopNCandidators(dposClient.GetCurrentHeight() / blockNumToUpdateDelegate)

	time.Sleep(2 * time.Second)
}

func initEnvDpos() (queue.Queue, *blockchain.BlockChain, queue.Module, queue.Module, *executor.Executor, queue.Module, queue.Module) {
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

	cs := New(cfg.Consensus, sub.Consensus["dpos"])
	cs.SetQueueClient(q.Client())

	mem := mempool.New(chain33Cfg)
	mem.SetQueueClient(q.Client())
	network := p2p.NewP2PMgr(chain33Cfg)

	network.SetQueueClient(q.Client())

	rpc.InitCfg(cfg.RPC)
	gapi := rpc.NewGRpcServer(q.Client(), nil)
	go gapi.Listen()
	return q, chain, s, mem, exec, cs, network
}

func createConn(url string) (*grpc.ClientConn, types.Chain33Client, error) {
	var err error
	//url := "127.0.0.1:8802"
	fmt.Println("grpc url:", url)
	conn1, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return conn1, nil, err
	}
	c1 := types.NewChain33Client(conn)
	//r = rand.New(rand.NewSource(types.Now().UnixNano()))
	return conn1, c1, nil
}

func createConn2() error {
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
	tx.Sign(types.SECP256K1, getprivkey("CC38546E9E659D15E6B4893F0AB32A06D103931A8230B0BDE71459D2B27D6944"))
	return tx
}

func clearTestData() {
	err := os.RemoveAll("datadir")
	if err != nil {
		fmt.Println("delete datadir have a err:", err.Error())
	}
	os.Remove("genesis.json")
	os.Remove("priv_validator.json")
	fmt.Println("test data clear successfully!")
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

// SendCBTx method
func verifyCB(cs *ConsensusState, info *dty.DposCBInfo) bool {
	canonical := dty.CanonicalOnceCBInfo{
		Cycle:      info.Cycle,
		StopHeight: info.StopHeight,
		StopHash:   info.StopHash,
		Pubkey:     info.Pubkey,
	}

	byteCB, err := json.Marshal(&canonical)
	if err != nil {
		dposlog.Error("marshal CanonicalOnceCBInfo failed", "err", err)
	}

	sig, err := cs.privValidator.SignMsg(byteCB)
	if err != nil {
		dposlog.Error("SignCBInfo failed.", "err", err)
		return false
	}

	info.Signature = hex.EncodeToString(sig.Bytes())

	return cs.VerifyCBInfo(info)
}

func sendRegistVrfMTx(cs *ConsensusState, info *dty.DposVrfMRegist) bool {
	tx, err := cs.client.CreateRegVrfMTx(info)
	if err != nil {
		dposlog.Error("CreateRegVrfMTx failed.", "err", err)
		return false
	}
	tx.Fee = fee

	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign RegistVrfMTx ok.")
	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		return false
	}

	return true
}

// SendRegistVrfRPTx method
func sendRegistVrfRPTx(cs *ConsensusState, info *dty.DposVrfRPRegist) bool {
	tx, err := cs.client.CreateRegVrfRPTx(info)
	if err != nil {
		dposlog.Error("CreateRegVrfRPTx failed.", "err", err)
		return false
	}

	tx.Fee = fee

	cs.privValidator.SignTx(tx)
	dposlog.Info("Sign RegVrfRPTx ok.")
	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return false
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		return false
	}

	return true
}

func sendTransferTx(cfg *types.Chain33Config, fromKey, to string, amount int64) bool {
	signer := util.HexToPrivkey(fromKey)
	var tx *types.Transaction
	transfer := &cty.CoinsAction{}
	v := &cty.CoinsAction_Transfer{Transfer: &types.AssetsTransfer{Amount: amount, Note: []byte(""), To: to}}
	transfer.Value = v
	transfer.Ty = cty.CoinsActionTransfer
	execer := []byte(cfg.GetCoinExec())
	tx = &types.Transaction{Execer: execer, Payload: types.Encode(transfer), To: to, Fee: fee}
	tx, err := types.FormatTx(cfg, string(execer), tx)
	if err != nil {
		fmt.Println("in sendTransferTx formatTx failed")
		return false
	}

	tx.Sign(types.SECP256K1, signer)
	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("in sendTransferTx SendTransaction failed")

		return false
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		fmt.Println("in sendTransferTx SendTransaction failed,reply not ok.")

		return false
	}
	fmt.Println("sendTransferTx ok")

	return true
}

func sendTransferToExecTx(cfg *types.Chain33Config, fromKey, execName string, amount int64) bool {
	signer := util.HexToPrivkey(fromKey)
	var tx *types.Transaction
	transfer := &cty.CoinsAction{}
	execAddr := address.ExecAddress(execName)
	v := &cty.CoinsAction_TransferToExec{TransferToExec: &types.AssetsTransferToExec{Amount: amount, Note: []byte(""), ExecName: execName, To: execAddr}}
	transfer.Value = v
	transfer.Ty = cty.CoinsActionTransferToExec
	execer := []byte("coins")
	tx = &types.Transaction{Execer: execer, Payload: types.Encode(transfer), To: address.ExecAddress("dpos"), Fee: fee}
	tx, err := types.FormatTx(cfg, string(execer), tx)
	if err != nil {
		fmt.Println("sendTransferToExecTx formatTx failed.")

		return false
	}

	tx.Sign(types.SECP256K1, signer)
	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("in sendTransferToExecTx SendTransaction failed")

		return false
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		fmt.Println("in sendTransferToExecTx SendTransaction failed,reply not ok.")

		return false
	}

	fmt.Println("sendTransferToExecTx ok")

	return true
}

func sendRegistCandidatorTx(cfg *types.Chain33Config, ppubkey, addr, ip, privKey string) bool {
	signer := util.HexToPrivkey(privKey)
	var tx *types.Transaction
	action := &dty.DposVoteAction{}

	v := &dty.DposVoteAction_Regist{
		Regist: &dty.DposCandidatorRegist{
			Pubkey:  ppubkey,
			Address: addr,
			IP:      ip,
		},
	}

	action.Value = v
	action.Ty = dty.DposVoteActionRegist
	execer := []byte("dpos")
	tx = &types.Transaction{Execer: execer, Payload: types.Encode(action), To: address.ExecAddress(string(execer)), Fee: fee}
	tx, err := types.FormatTx(cfg, string(execer), tx)
	if err != nil {
		fmt.Println("sendRegistCandidatorTx formatTx failed.")

		return false
	}

	tx.Sign(types.SECP256K1, signer)
	reply, err := c.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("in sendRegistCandidatorTx SendTransaction failed")

		return false
	}
	if !reply.IsOk {
		fmt.Fprintln(os.Stderr, errors.New(string(reply.GetMsg())))
		fmt.Println("in sendTransferToExecTx SendTransaction failed,reply not ok.")

		return false
	}

	fmt.Println("sendRegistCandidatorTx ok")

	return true
}

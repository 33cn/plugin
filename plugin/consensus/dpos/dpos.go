// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/33cn/chain33/common/address"
	"github.com/33cn/chain33/util"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"

	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
	"github.com/golang/protobuf/proto"
)

const dposVersion = "0.1.0"
const dposShuffleTypeFixOrderByAddr = 1
const dposShuffleTypeOrderByVrfInfo = 2

var (
	dposlog                   = log15.New("module", "dpos")
	genesis                   string
	genesisBlockTime          int64
	timeoutCheckConnections   int32 = 1000
	timeoutVoting             int32 = 3000
	timeoutWaitNotify         int32 = 2000
	createEmptyBlocks               = false
	createEmptyBlocksInterval int32 // second
	validatorNodes                  = []string{"127.0.0.1:46656"}
	isValidator                     = false

	dposDelegateNum          int64 = 3 //委托节点个数，从配置读取，以后可以根据投票结果来定
	dposBlockInterval        int64 = 3 //出块间隔，当前按3s
	dposContinueBlockNum     int64 = 6 //一个委托节点当选后，一次性持续出块数量
	dposCycle                      = dposDelegateNum * dposBlockInterval * dposContinueBlockNum
	dposPeriod                     = dposBlockInterval * dposContinueBlockNum
	zeroHash                 [32]byte
	dposPort                       = "36656"
	shuffleType              int32 = dposShuffleTypeOrderByVrfInfo //shuffleType为1表示使用固定出块顺序，为2表示使用vrf信息进行出块顺序洗牌
	whetherUpdateTopN              = false                         //是否更新topN，如果为true，根据下面几个配置项定期更新topN节点;如果为false，则一直使用初始配置的节点，不关注投票结果
	blockNumToUpdateDelegate int64 = 20000
	registTopNHeightLimit    int64 = 100
	updateTopNHeightLimit    int64 = 200
)

func init() {
	drivers.Reg("dpos", New)
	drivers.QueryData.Register("dpos", &Client{})
}

// Client Tendermint implementation
type Client struct {
	//config
	*drivers.BaseClient
	genesisDoc    *ttypes.GenesisDoc // initial validator set
	privValidator ttypes.PrivValidator
	privKey       crypto.PrivKey // local node's p2p key
	pubKey        string
	csState       *ConsensusState
	crypto        crypto.Crypto
	node          *Node
	stopC         chan struct{}
	isDelegator   bool
	blockTime     int64
	testFlag      bool
}

type subConfig struct {
	Genesis                   string   `json:"genesis"`
	GenesisBlockTime          int64    `json:"genesisBlockTime"`
	TimeoutCheckConnections   int32    `json:"timeoutCheckConnections"`
	TimeoutVoting             int32    `json:"timeoutVoting"`
	TimeoutWaitNotify         int32    `json:"timeoutWaitNotify"`
	CreateEmptyBlocks         bool     `json:"createEmptyBlocks"`
	CreateEmptyBlocksInterval int32    `json:"createEmptyBlocksInterval"`
	ValidatorNodes            []string `json:"validatorNodes"`
	DelegateNum               int64    `json:"delegateNum"`
	BlockInterval             int64    `json:"blockInterval"`
	ContinueBlockNum          int64    `json:"continueBlockNum"`
	IsValidator               bool     `json:"isValidator"`
	Port                      string   `json:"port"`
	ShuffleType               int32    `json:"shuffleType"`
	WhetherUpdateTopN         bool     `json:"whetherUpdateTopN"`
	BlockNumToUpdateDelegate  int64    `json:"blockNumToUpdateDelegate"`
	RegistTopNHeightLimit     int64    `json:"registTopNHeightLimit"`
	UpdateTopNHeightLimit     int64    `json:"updateTopNHeightLimit"`
}

func (client *Client) applyConfig(sub []byte) {
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	if subcfg.Genesis != "" {
		genesis = subcfg.Genesis
	}
	if subcfg.GenesisBlockTime > 0 {
		genesisBlockTime = subcfg.GenesisBlockTime
	}
	if subcfg.TimeoutCheckConnections > 0 {
		timeoutCheckConnections = subcfg.TimeoutCheckConnections
	}
	if subcfg.TimeoutVoting > 0 {
		timeoutVoting = subcfg.TimeoutVoting
	}
	if subcfg.TimeoutWaitNotify > 0 {
		timeoutWaitNotify = subcfg.TimeoutWaitNotify
	}
	createEmptyBlocks = subcfg.CreateEmptyBlocks
	if subcfg.CreateEmptyBlocksInterval > 0 {
		createEmptyBlocksInterval = subcfg.CreateEmptyBlocksInterval
	}

	if subcfg.DelegateNum > 0 {
		dposDelegateNum = subcfg.DelegateNum
	}

	if len(subcfg.ValidatorNodes) > 0 {
		validatorNodes = subcfg.ValidatorNodes
		//dposDelegateNum = int64(len(subcfg.ValidatorNodes))
	}

	if subcfg.BlockInterval > 0 {
		dposBlockInterval = subcfg.BlockInterval
	}

	if subcfg.ContinueBlockNum > 0 {
		dposContinueBlockNum = subcfg.ContinueBlockNum
	}

	if subcfg.Port != "" {
		dposPort = subcfg.Port
	}
	dposCycle = dposDelegateNum * dposBlockInterval * dposContinueBlockNum
	dposPeriod = dposBlockInterval * dposContinueBlockNum

	if subcfg.CreateEmptyBlocks {
		createEmptyBlocks = true
	}

	if subcfg.IsValidator {
		isValidator = true
	}

	if subcfg.ShuffleType > 0 {
		shuffleType = subcfg.ShuffleType
	}

	if subcfg.WhetherUpdateTopN {
		whetherUpdateTopN = subcfg.WhetherUpdateTopN
	}

	if subcfg.BlockNumToUpdateDelegate > 0 {
		blockNumToUpdateDelegate = subcfg.BlockNumToUpdateDelegate
	}

	if subcfg.RegistTopNHeightLimit > 0 {
		registTopNHeightLimit = subcfg.RegistTopNHeightLimit
	}

	if subcfg.UpdateTopNHeightLimit > 0 {
		updateTopNHeightLimit = subcfg.UpdateTopNHeightLimit
	}
}

// New ...
func New(cfg *types.Consensus, sub []byte) queue.Module {
	dposlog.Info("Start to create dpos client")
	//init rand
	ttypes.Init()

	genDoc, err := ttypes.GenesisDocFromFile("./genesis.json")
	if err != nil {
		dposlog.Error("NewDPosClient", "msg", "GenesisDocFromFile failded", "error", err)
		//return nil
	}

	//为了使用VRF，需要使用SECP256K1体系的公私钥
	cr, err := crypto.Load(types.GetSignName("", types.SECP256K1), -1)
	if err != nil {
		dposlog.Error("NewDPosClient", "err", err)
		return nil
	}

	ttypes.ConsensusCrypto = cr

	//安全连接仍然要使用ed25519
	cr2, err := crypto.Load(types.GetSignName("", types.ED25519), -1)
	if err != nil {
		dposlog.Error("NewDPosClient", "err", err)
		return nil
	}
	ttypes.SecureConnCrypto = cr2

	priv, err := cr2.GenKey()
	if err != nil {
		dposlog.Error("NewDPosClient", "GenKey err", err)
		return nil
	}

	privValidator := ttypes.LoadOrGenPrivValidatorFS("./priv_validator.json")
	if privValidator == nil {
		dposlog.Error("NewDPosClient create priv_validator file failed")
		//return nil
	}

	ttypes.InitMessageMap()

	pubkey := privValidator.GetPubKey().KeyString()
	c := drivers.NewBaseClient(cfg)
	client := &Client{
		BaseClient:    c,
		genesisDoc:    genDoc,
		privValidator: privValidator,
		privKey:       priv,
		pubKey:        pubkey,
		crypto:        cr,
		stopC:         make(chan struct{}, 1),
		isDelegator:   false,
		testFlag:      false,
	}
	c.SetChild(client)

	client.applyConfig(sub)
	return client
}

// PrivValidator returns the Node's PrivValidator.
func (client *Client) PrivValidator() ttypes.PrivValidator {
	return client.privValidator
}

// GenesisDoc returns the Node's GenesisDoc.
func (client *Client) GenesisDoc() *ttypes.GenesisDoc {
	return client.genesisDoc
}

// Close TODO:may need optimize
func (client *Client) Close() {
	client.node.Stop()
	client.stopC <- struct{}{}
	dposlog.Info("consensus dpos closed")
}

// SetQueueClient ...
func (client *Client) SetQueueClient(q queue.Client) {
	client.InitClient(q, func() {
		//call init block
		//client.InitBlock()
	})

	go client.EventLoop()
	go client.StartConsensus()
}

// DebugCatchup define whether catch up now
const DebugCatchup = false

// StartConsensus a routine that make the consensus start
func (client *Client) StartConsensus() {
	//进入共识前先同步到最大高度
	hint := time.NewTicker(5 * time.Second)
	beg := time.Now()
	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}
OuterLoop:
	for !DebugCatchup && block != nil {
		select {
		case <-hint.C:
			dposlog.Info("Still catching up max height......", "cost", time.Since(beg))
		default:
			if client.IsCaughtUp() {
				dposlog.Info("This node has caught up max height")
				break OuterLoop
			}
			time.Sleep(time.Second)
		}
	}
	hint.Stop()

	//如果非候选节点，直接返回，接受同步区块数据，不做任何共识相关的事情。
	if !isValidator {
		dposlog.Info("This node is not a validator,does not join the consensus, just syncs blocks from validators")
		client.InitBlock()
		return
	}

	var valMgr ValidatorMgr
	valMgrTmp, err := MakeGenesisValidatorMgr(client.genesisDoc)
	if err != nil {
		dposlog.Error("StartConsensus", "msg", "MakeGenesisValidatorMgr failded", "error", err)
		return
	}
	valMgr = valMgrTmp.Copy()
	dposlog.Debug("Load Validator Manager finish", "state", valMgr)
	block, err = client.RequestLastBlock()
	if err != nil {
		panic(err)
	}
	if block != nil && whetherUpdateTopN {
		//time.Sleep(time.Second * 5)
		//cands, err := client.QueryCandidators()
		info := CalcTopNVersion(block.Height)
		version := info.Version
		var topN *dty.TopNCandidators
		for version >= 0 {
			topN, err = client.QueryTopNCandidators(version)
			if err != nil || topN == nil {
				version--
			} else {
				break
			}
		}

		if topN == nil {
			dposlog.Info("QueryTopNCandidators failed, no candidators")
		} else if len(topN.FinalCands) != int(dposDelegateNum) {
			dposlog.Info("QueryTopNCandidators success but no enough candidators", "dposDelegateNum", dposDelegateNum, "candidatorNum", len(topN.FinalCands))
		} else {
			validators := make([]*ttypes.Validator, dposDelegateNum)
			nodes := make([]string, dposDelegateNum)
			for i, val := range topN.FinalCands {
				// Make validator
				validators[i] = &ttypes.Validator{
					Address: address.BytesToBtcAddress(address.NormalVer, val.Pubkey).Hash160[:],
					PubKey:  val.Pubkey,
				}
				nodes[i] = val.IP + ":" + dposPort
			}
			valMgr.Validators = ttypes.NewValidatorSet(validators)
			dposlog.Info("QueryCandidators success and update validator set", "old validators", printValidators(valMgrTmp.Validators), "new validators", printValidators(valMgr.Validators))
			dposlog.Info("QueryCandidators success and update validator node ips", "old validator ips", printNodeIPs(validatorNodes), "new validators ips", printNodeIPs(nodes))
			validatorNodes = nodes
		}
	}

	dposlog.Info("StartConsensus", "validators", printValidators(valMgr.Validators))
	// Log whether this node is a delegator or an observer
	if valMgr.Validators.HasAddress(client.privValidator.GetAddress()) {
		dposlog.Info("This node is a delegator")
		client.isDelegator = true
	} else {
		dposlog.Info("This node is not a delegator")
		dposlog.Info("StartConsensus", "privValidator addr", hex.EncodeToString(client.privValidator.GetAddress()))
	}

	// Make ConsensusReactor
	csState := NewConsensusState(client, valMgr)

	client.csState = csState

	csState.SetPrivValidator(client.privValidator, client.ValidatorIndex())

	// Create & add listener
	protocol, listeningAddress := "tcp", "0.0.0.0:"+dposPort
	node := NewNode(validatorNodes, protocol, listeningAddress, client.privKey, valMgr.ChainID, dposVersion, csState)

	client.node = node

	// 对于受托节点，才需要初始化区块，启动共识相关程序等,后续支持投票要做成动态切换的。
	if client.isDelegator {
		client.InitBlock()
		time.Sleep(time.Second * 2)
		client.csState.Init()
		node.Start(client.testFlag)
	}

	//go client.MonitorCandidators()
}

func printValidators(set *ttypes.ValidatorSet) string {
	result := "Validators:["
	for _, v := range set.Validators {
		result = fmt.Sprintf("%s%s|%s,", result, hex.EncodeToString(v.PubKey), hex.EncodeToString(v.Address))
	}

	result += "]"
	return result
}

func printNodeIPs(ips []string) string {
	result := "nodeIPs:["
	for _, v := range ips {
		result = fmt.Sprintf("%s%s,", result, v)
	}

	result += "]"
	return result
}

// GetGenesisBlockTime ...
func (client *Client) GetGenesisBlockTime() int64 {
	return genesisBlockTime
}

// CreateGenesisTx ...
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte(client.GetAPI().GetConfig().GetCoinExec())
	tx.To = genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = 1e8 * client.GetAPI().GetConfig().GetCoinPrecision()
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

// CheckBlock 暂不检查任何的交易
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	return nil
}

// ProcEvent ...
func (client *Client) ProcEvent(msg *queue.Message) bool {
	return false
}

// CreateBlock a routine monitor whether some transactions available and tell client by available channel
func (client *Client) CreateBlock() {
	lastBlock := client.GetCurrentBlock()
	cfg := client.GetAPI().GetConfig()
	txs := client.RequestTx(int(cfg.GetP(lastBlock.Height+1).MaxTxNumber), nil)
	if len(txs) == 0 {
		block := client.GetCurrentBlock()
		if createEmptyBlocks {
			emptyBlock := &types.Block{}
			emptyBlock.StateHash = block.StateHash
			emptyBlock.ParentHash = block.Hash(cfg)
			emptyBlock.Height = block.Height + 1
			emptyBlock.Txs = nil
			emptyBlock.TxHash = zeroHash[:]
			emptyBlock.BlockTime = client.blockTime
			err := client.WriteBlock(lastBlock.StateHash, emptyBlock)
			//判断有没有交易是被删除的，这类交易要从mempool 中删除
			if err != nil {
				return
			}
		} else {
			dposlog.Info("Ignore to create new Block for no tx in mempool", "Height", block.Height+1)
		}

		return
	}
	//check dup
	txs = client.CheckTxDup(txs, client.GetCurrentHeight())
	var newblock types.Block
	newblock.ParentHash = lastBlock.Hash(cfg)
	newblock.Height = lastBlock.Height + 1
	client.AddTxsToBlock(&newblock, txs)
	newblock.Difficulty = cfg.GetP(0).PowLimitBits

	//需要首先对交易进行排序然后再计算TxHash
	if cfg.IsFork(newblock.Height, "ForkRootHash") {
		newblock.Txs = types.TransactionSort(newblock.Txs)
	}
	newblock.TxHash = merkle.CalcMerkleRoot(cfg, newblock.Height, newblock.Txs)
	newblock.BlockTime = client.blockTime

	err := client.WriteBlock(lastBlock.StateHash, &newblock)
	//判断有没有交易是被删除的，这类交易要从mempool 中删除
	if err != nil {
		return
	}
}

// StopC stop client
func (client *Client) StopC() <-chan struct{} {
	return client.stopC
}

// CheckTxDup check transactions that duplicate
func (client *Client) CheckTxDup(txs []*types.Transaction, height int64) (transactions []*types.Transaction) {
	cacheTxs := types.TxsToCache(txs)
	var err error
	cacheTxs, err = util.CheckTxDup(client.GetQueueClient(), cacheTxs, height)
	if err != nil {
		return txs
	}
	return types.CacheToTxs(cacheTxs)
}

// SetBlockTime set current block time to generate new block
func (client *Client) SetBlockTime(blockTime int64) {
	client.blockTime = blockTime
}

// ValidatorIndex get the index of local this validator if it's
func (client *Client) ValidatorIndex() int {
	if client.isDelegator {
		index, _ := client.csState.validatorMgr.Validators.GetByAddress(client.privValidator.GetAddress())
		return index
	}

	return -1
}

// QueryCandidators query the topN candidators from blockchain
func (client *Client) QueryCandidators() ([]*dty.Candidator, error) {
	req := &dty.CandidatorQuery{
		TopN: int32(dposDelegateNum),
	}
	param, err := proto.Marshal(req)
	if err != nil {
		dposlog.Error("Marshal CandidatorQuery failed", "err", err)
		return nil, err
	}
	msg := client.GetQueueClient().NewMessage("execs", types.EventBlockChainQuery,
		&types.ChainExecutor{
			Driver:    dty.DPosX,
			FuncName:  dty.FuncNameQueryCandidatorByTopN,
			StateHash: zeroHash[:],
			Param:     param,
		})

	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		dposlog.Error("send CandidatorQuery to dpos exec failed", "err", err)
		return nil, err
	}

	msg, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		dposlog.Error("send CandidatorQuery wait failed", "err", err)
		return nil, err
	}

	res := msg.GetData().(types.Message).(*dty.CandidatorReply)

	var cands []*dty.Candidator
	for _, val := range res.GetCandidators() {
		bPubkey, err := hex.DecodeString(val.Pubkey)
		if err != nil {
			return nil, err
		}

		cand := &dty.Candidator{
			Pubkey:  bPubkey,
			Address: val.Address,
			IP:      val.IP,
			Votes:   val.Votes,
			Status:  val.Status,
		}

		cands = append(cands, cand)
	}
	return cands, nil
}

func (client *Client) isValidatorSetSame(v1, v2 *ttypes.ValidatorSet) bool {
	if v1 == nil || v2 == nil || len(v1.Validators) != len(v2.Validators) {
		return false
	}

	for i := 0; i < len(v1.Validators); i++ {
		if !bytes.Equal(v1.Validators[i].PubKey, v2.Validators[i].PubKey) {
			return false
		}
	}

	return true
}

// CreateRecordCBTx create the tx to record cb
func (client *Client) CreateRecordCBTx(info *dty.DposCBInfo) (tx *types.Transaction, err error) {
	var action dty.DposVoteAction
	action.Value = &dty.DposVoteAction_RecordCB{
		RecordCB: info,
	}
	action.Ty = dty.DposVoteActionRecordCB
	cfg := client.GetAPI().GetConfig()
	tx, err = types.CreateFormatTx(cfg, "dpos", types.Encode(&action))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CreateRegVrfMTx create the tx to regist Vrf M
func (client *Client) CreateRegVrfMTx(info *dty.DposVrfMRegist) (tx *types.Transaction, err error) {
	var action dty.DposVoteAction
	action.Value = &dty.DposVoteAction_RegistVrfM{
		RegistVrfM: info,
	}
	action.Ty = dty.DposVoteActionRegistVrfM
	cfg := client.GetAPI().GetConfig()
	tx, err = types.CreateFormatTx(cfg, "dpos", types.Encode(&action))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// CreateRegVrfRPTx create the tx to regist Vrf RP
func (client *Client) CreateRegVrfRPTx(info *dty.DposVrfRPRegist) (tx *types.Transaction, err error) {
	var action dty.DposVoteAction
	action.Value = &dty.DposVoteAction_RegistVrfRP{
		RegistVrfRP: info,
	}
	action.Ty = dty.DposVoteActionRegistVrfRP
	cfg := client.GetAPI().GetConfig()
	tx, err = types.CreateFormatTx(cfg, "dpos", types.Encode(&action))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// QueryVrfInfos query the vrf infos by pubkeys
func (client *Client) QueryVrfInfos(pubkeys [][]byte, cycle int64) ([]*dty.VrfInfo, error) {
	req := &dty.DposVrfQuery{
		Cycle: cycle,
		Ty:    dty.QueryVrfByCycleForPubkeys,
	}

	for i := 0; i < len(pubkeys); i++ {
		req.Pubkeys = append(req.Pubkeys, strings.ToUpper(hex.EncodeToString(pubkeys[i])))
	}

	param, err := proto.Marshal(req)
	if err != nil {
		dposlog.Error("Marshal DposVrfQuery failed", "err", err)
		return nil, err
	}
	msg := client.GetQueueClient().NewMessage("execs", types.EventBlockChainQuery,
		&types.ChainExecutor{
			Driver:    dty.DPosX,
			FuncName:  dty.FuncNameQueryVrfByCycleForPubkeys,
			StateHash: zeroHash[:],
			Param:     param,
		})

	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		dposlog.Error("send DposVrfQuery to dpos exec failed", "err", err)
		return nil, err
	}

	msg, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		dposlog.Error("send DposVrfQuery wait failed", "err", err)
		return nil, err
	}

	res := msg.GetData().(types.Message).(*dty.DposVrfReply)
	if len(res.Vrf) > 0 {
		dposlog.Info("DposVrfQuerys ok")
	} else {
		dposlog.Info("DposVrfQuerys ok,but no info")
	}

	var infos []*dty.VrfInfo
	for _, val := range res.Vrf {
		bPubkey, err := hex.DecodeString(val.Pubkey)
		if err != nil {
			bPubkey = nil
		}

		bM, err := hex.DecodeString(val.M)
		if err != nil {
			bM = nil
		}

		bR, err := hex.DecodeString(val.R)
		if err != nil {
			bR = nil
		}

		bP, err := hex.DecodeString(val.P)
		if err != nil {
			bP = nil
		}
		info := &dty.VrfInfo{
			Index:  val.Index,
			Pubkey: bPubkey,
			Cycle:  val.Cycle,
			Height: val.Height,
			Time:   val.Time,
			M:      bM,
			R:      bR,
			P:      bP,
		}

		infos = append(infos, info)
		dposlog.Info("VrfInfos", "info", fmt.Sprintf("Cycle:%d,pubkey:%s,Height:%d,M:%s,R:%s,P:%s", val.Cycle, val.Pubkey, val.Height, val.M, val.R, val.P))

	}

	return infos, nil
}

// CreateTopNRegistTx create tx to regist topN
func (client *Client) CreateTopNRegistTx(reg *dty.TopNCandidatorRegist) (tx *types.Transaction, err error) {
	var action dty.DposVoteAction
	action.Value = &dty.DposVoteAction_RegistTopN{
		RegistTopN: reg,
	}
	action.Ty = dty.DPosVoteActionRegistTopNCandidator
	cfg := client.GetAPI().GetConfig()
	tx, err = types.CreateFormatTx(cfg, "dpos", types.Encode(&action))
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// QueryTopNCandidators method
func (client *Client) QueryTopNCandidators(version int64) (*dty.TopNCandidators, error) {
	req := &dty.TopNCandidatorsQuery{Version: version}
	param, err := proto.Marshal(req)
	if err != nil {
		dposlog.Error("Marshal TopNCandidatorsQuery failed", "version", version, "err", err)
		return nil, err
	}
	msg := client.GetQueueClient().NewMessage("execs", types.EventBlockChainQuery,
		&types.ChainExecutor{
			Driver:    dty.DPosX,
			FuncName:  dty.FuncNameQueryTopNByVersion,
			StateHash: zeroHash[:],
			Param:     param,
		})

	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		dposlog.Error("send TopNCandidatorsQuery to dpos exec failed", "version", version, "err", err)
		return nil, err
	}

	msg, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		dposlog.Error("send TopNCandidatorsQuery wait failed", "version", version, "err", err)
		return nil, err
	}

	res := msg.GetData().(types.Message).(*dty.TopNCandidatorsReply)
	info := res.TopN
	dposlog.Info("TopNCandidatorsQuery get reply", "version", info.Version, "status", info.Status, "final candidators", printCandidators(info.FinalCands))

	return info, nil
}

func printCandidators(cands []*dty.Candidator) string {
	result := "["
	for i := 0; i < len(cands); i++ {
		result = fmt.Sprintf("%spubkey:%s,ip:%s;", result, hex.EncodeToString(cands[i].Pubkey), cands[i].IP)
	}
	result += "]"

	return result
}

// GetConsensusState return the pointer to ConsensusState
func (client *Client) GetConsensusState() *ConsensusState {
	return client.csState
}

// SetTestFlag set the test flag
func (client *Client) SetTestFlag() {
	client.testFlag = true
}

// GetNode return the pointer to Node
func (client *Client) GetNode() *Node {
	return client.node
}

//CmpBestBlock 比较newBlock是不是最优区块
func (client *Client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	return false
}

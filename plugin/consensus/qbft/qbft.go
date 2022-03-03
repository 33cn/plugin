// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package qbft

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	"github.com/33cn/chain33/common/crypto"
	dbm "github.com/33cn/chain33/common/db"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ttypes "github.com/33cn/plugin/plugin/consensus/qbft/types"
	tmtypes "github.com/33cn/plugin/plugin/dapp/qbftNode/types"
	"github.com/golang/protobuf/proto"
)

const (
	qbftVersion = "0.2.1"

	// DefaultQbftPort 默认端口
	DefaultQbftPort = 33001
)

var (
	qbftlog = log15.New("module", "qbft")

	genesis               string
	genesisAmount         int64 = 1e8
	genesisBlockTime      int64
	timeoutTxAvail        atomic.Value // 1000  millisecond
	timeoutPropose        atomic.Value // 3000
	timeoutProposeDelta   atomic.Value // 500
	timeoutPrevote        atomic.Value // 1000
	timeoutPrevoteDelta   atomic.Value // 500
	timeoutPrecommit      atomic.Value // 1000
	timeoutPrecommitDelta atomic.Value // 500
	timeoutCommit         atomic.Value // 1000
	skipTimeoutCommit     atomic.Value // false
	emptyBlockInterval    atomic.Value // 0  second
	genesisFile                        = "genesis.json"
	privFile                           = "priv_validator.json"
	dbPath                             = fmt.Sprintf("datadir%sqbft", string(os.PathSeparator))
	port                  int32        = DefaultQbftPort
	validatorNodes                     = []string{"127.0.0.1:33001"}
	fastSync                           = false
	preExec               atomic.Value // false
	signName              atomic.Value // "ed25519"
	useAggSig             atomic.Value // false
	multiBlocks           atomic.Value // 1
	gossipVotes           atomic.Value
	detachExec            atomic.Value // false
	sameBlocktime         atomic.Value // false

	zeroHash                    [32]byte
	random                      *rand.Rand
	peerGossipSleepDuration     atomic.Value
	peerQueryMaj23SleepDuration int32 = 2000
)

func init() {
	drivers.Reg("qbft", New)
	drivers.QueryData.Register("qbft", &Client{})
	random = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Client qbft implementation
type Client struct {
	//config
	*drivers.BaseClient
	genesisDoc    *ttypes.GenesisDoc // initial validator set
	privValidator ttypes.PrivValidator
	privKey       crypto.PrivKey // node private key
	csState       *ConsensusState
	csStore       *ConsensusStore // save consensus state
	node          *Node
	txsAvailable  chan int64
	ctx           context.Context
	cancel        context.CancelFunc
}

type subConfig struct {
	Genesis               string   `json:"genesis"`
	GenesisAmount         int64    `json:"genesisAmount"`
	GenesisBlockTime      int64    `json:"genesisBlockTime"`
	TimeoutTxAvail        int32    `json:"timeoutTxAvail"`
	TimeoutPropose        int32    `json:"timeoutPropose"`
	TimeoutProposeDelta   int32    `json:"timeoutProposeDelta"`
	TimeoutPrevote        int32    `json:"timeoutPrevote"`
	TimeoutPrevoteDelta   int32    `json:"timeoutPrevoteDelta"`
	TimeoutPrecommit      int32    `json:"timeoutPrecommit"`
	TimeoutPrecommitDelta int32    `json:"timeoutPrecommitDelta"`
	TimeoutCommit         int32    `json:"timeoutCommit"`
	SkipTimeoutCommit     bool     `json:"skipTimeoutCommit"`
	EmptyBlockInterval    int32    `json:"emptyBlockInterval"`
	GenesisFile           string   `json:"genesisFile"`
	PrivFile              string   `json:"privFile"`
	DbPath                string   `json:"dbPath"`
	Port                  int32    `json:"port"`
	ValidatorNodes        []string `json:"validatorNodes"`
	FastSync              bool     `json:"fastSync"`
	PreExec               bool     `json:"preExec"`
	SignName              string   `json:"signName"`
	UseAggregateSignature bool     `json:"useAggregateSignature"`
	MultiBlocks           int64    `json:"multiBlocks"`
	MessageInterval       int32    `json:"messageInterval"`
	DetachExecution       bool     `json:"detachExecution"`
	SameBlocktime         bool     `json:"sameBlocktime"`
}

func applyConfig(cfg *types.Consensus, sub []byte) {
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	genesis = subcfg.Genesis
	genesisAmount = subcfg.GenesisAmount
	genesisBlockTime = subcfg.GenesisBlockTime
	if !cfg.Minerstart {
		return
	}

	timeoutTxAvail.Store(subcfg.TimeoutTxAvail)
	timeoutPropose.Store(subcfg.TimeoutPropose)
	timeoutProposeDelta.Store(subcfg.TimeoutProposeDelta)
	timeoutPrevote.Store(subcfg.TimeoutPrevote)
	timeoutPrevoteDelta.Store(subcfg.TimeoutPrevoteDelta)
	timeoutPrecommit.Store(subcfg.TimeoutPrecommit)
	timeoutPrecommitDelta.Store(subcfg.TimeoutPrecommitDelta)
	timeoutCommit.Store(subcfg.TimeoutCommit)
	skipTimeoutCommit.Store(subcfg.SkipTimeoutCommit)
	emptyBlockInterval.Store(subcfg.EmptyBlockInterval)
	if subcfg.GenesisFile != "" {
		genesisFile = subcfg.GenesisFile
	}
	if subcfg.PrivFile != "" {
		privFile = subcfg.PrivFile
	}
	if subcfg.DbPath != "" {
		dbPath = subcfg.DbPath
	}
	if subcfg.Port > 0 {
		port = subcfg.Port
	}
	if len(subcfg.ValidatorNodes) > 0 {
		validatorNodes = subcfg.ValidatorNodes
	}
	fastSync = subcfg.FastSync
	preExec.Store(subcfg.PreExec)
	signName.Store("ed25519")
	if subcfg.SignName != "" {
		signName.Store(subcfg.SignName)
	}
	useAggSig.Store(subcfg.UseAggregateSignature)
	multiBlocks.Store(int64(1))
	if subcfg.MultiBlocks > 0 {
		multiBlocks.Store(subcfg.MultiBlocks)
	}
	peerGossipSleepDuration.Store(int32(100))
	if subcfg.MessageInterval > 0 {
		peerGossipSleepDuration.Store(subcfg.MessageInterval)
	}
	detachExec.Store(subcfg.DetachExecution)
	sameBlocktime.Store(subcfg.SameBlocktime)

	gossipVotes.Store(true)
}

// UseAggSig returns whether use aggregate signature
func UseAggSig() bool {
	return useAggSig.Load().(bool)
}

// DetachExec returns whether detach Execution from Consensus
func DetachExec() bool {
	return detachExec.Load().(bool)
}

// DefaultDBProvider returns a database
func DefaultDBProvider(name string) dbm.DB {
	return dbm.NewDB(name, "leveldb", dbPath, 0)
}

// New ...
func New(cfg *types.Consensus, sub []byte) queue.Module {
	qbftlog.Info("create qbft client")
	genesis = cfg.Genesis
	genesisBlockTime = cfg.GenesisBlockTime
	applyConfig(cfg, sub)
	if !cfg.Minerstart {
		qbftlog.Info("This node only sync block")
		c := drivers.NewBaseClient(cfg)
		client := &Client{BaseClient: c}
		c.SetChild(client)
		return client
	}
	//init rand
	ttypes.Init()

	signType, ok := ttypes.SignMap[signName.Load().(string)]
	if !ok {
		qbftlog.Error("invalid sign name")
		return nil
	}

	ttypes.CryptoName = types.GetSignName("", signType)
	cr, err := crypto.Load(ttypes.CryptoName, -1)
	if err != nil {
		qbftlog.Error("load qbft crypto fail", "err", err)
		return nil
	}
	ttypes.ConsensusCrypto = cr

	if UseAggSig() {
		_, err = crypto.ToAggregate(ttypes.ConsensusCrypto)
		if err != nil {
			qbftlog.Error("qbft crypto not support aggregate signature", "name", ttypes.CryptoName)
			return nil
		}
	}

	genDoc, err := ttypes.GenesisDocFromFile(genesisFile)
	if err != nil {
		qbftlog.Error("load genesis file fail", "error", err)
		return nil
	}
	privValidator := ttypes.LoadPrivValidatorFS(privFile)
	if privValidator == nil {
		qbftlog.Error("load priv_validator file fail")
		return nil
	}

	qbftlog.Info("show qbft info", "version", qbftVersion, "sign", ttypes.CryptoName, "useAggSig", UseAggSig(),
		"detachExec", DetachExec(), "genesisFile", genesisFile, "privFile", privFile)

	ttypes.InitMessageMap()

	//采用context来统一管理所有服务
	ctx, stop := context.WithCancel(context.Background())

	c := drivers.NewBaseClient(cfg)
	client := &Client{
		BaseClient:    c,
		genesisDoc:    genDoc,
		privValidator: privValidator,
		privKey:       privValidator.PrivKey,
		csStore:       NewConsensusStore(),
		txsAvailable:  make(chan int64, 1),
		ctx:           ctx,
		cancel:        stop,
	}
	c.SetChild(client)
	return client
}

// GenesisState returns the Node's GenesisState.
func (client *Client) GenesisState() *State {
	state, err := MakeGenesisState(client.genesisDoc)
	if err != nil {
		qbftlog.Error("GenesisState", "err", err)
		return nil
	}
	return &state
}

// Close TODO:may need optimize
func (client *Client) Close() {
	client.BaseClient.Close()
	if client.cancel != nil {
		client.cancel()
	}
	if client.node != nil {
		client.node.Stop()
	}
	qbftlog.Info("consensus qbft closed")
}

// SetQueueClient ...
func (client *Client) SetQueueClient(q queue.Client) {
	client.InitClient(q, func() {
		//call init block
		client.InitBlock()
	})

	go client.EventLoop()
	if !client.IsMining() {
		qbftlog.Info("enter sync mode")
		return
	}
	go client.StartConsensus()
}

// StartConsensus a routine that make the consensus start
func (client *Client) StartConsensus() {
	//进入共识前先同步到最大高度
	hint := time.NewTicker(5 * time.Second)
	defer hint.Stop()

	beg := time.Now()
OuterLoop:
	for fastSync {
		select {
		case <-client.ctx.Done():
			qbftlog.Info("StartConsensus quit")
			return
		case <-hint.C:
			qbftlog.Info("Still catching up max height......", "Height", client.GetCurrentHeight(), "cost", time.Since(beg))
		default:
			if client.IsCaughtUp() {
				qbftlog.Info("This node has caught up max height")
				break OuterLoop
			}
			time.Sleep(time.Second)
		}
	}

	// load state
	var state State
	if client.GetCurrentHeight() == 0 {
		genState := client.GenesisState()
		if genState == nil {
			panic("StartConsensus GenesisState fail")
		}
		state = genState.Copy()
	} else if client.GetCurrentHeight() <= client.csStore.LoadStateHeight() {
		stoState := client.csStore.LoadStateFromStore()
		if stoState == nil {
			panic("StartConsensus LoadStateFromStore fail")
		}
		state = LoadState(stoState)
		qbftlog.Info("load state from store")
	} else {
		height := client.GetCurrentHeight()
		blkState := client.LoadBlockState(height)
		if blkState == nil {
			panic("StartConsensus LoadBlockState fail")
		}
		state = LoadState(blkState)
		qbftlog.Info("load state from block")
		//save initial state in store
		blkCommit := client.LoadBlockCommit(height)
		if blkCommit == nil {
			panic("StartConsensus LoadBlockCommit fail")
		}
		err := client.csStore.SaveConsensusState(height-1, blkState, blkCommit)
		if err != nil {
			panic(fmt.Sprintf("StartConsensus SaveConsensusState fail: %v", err))
		}
		qbftlog.Info("save state from block")
	}

	// start
	qbftlog.Info("show state info",
		"privValidator", fmt.Sprintf("%X", ttypes.Fingerprint(client.privValidator.GetAddress())),
		"state", state)
	// Log whether this node is a validator or an observer
	if state.Validators.HasAddress(client.privValidator.GetAddress()) {
		qbftlog.Info("This node is a validator")
	} else {
		qbftlog.Info("This node is not a validator")
	}

	stateDB := NewStateDB(client, state)

	// make block executor for consensus and blockchain reactors to execute blocks
	blockExec := NewBlockExecutor(stateDB)

	// Make ConsensusReactor
	csState := NewConsensusState(client, state, blockExec)
	// reset height, round, state begin at newheigt,0,0
	client.privValidator.ResetLastHeight(state.LastBlockHeight)
	csState.SetPrivValidator(client.privValidator)

	client.csState = csState

	// Create & add listener
	protocol, listeningAddress := "tcp", fmt.Sprintf(":%d", port)
	node := NewNode(validatorNodes, protocol, listeningAddress, client.privKey, state.ChainID, qbftVersion, csState)

	client.node = node
	node.Start()

	go client.CreateBlock()
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
	g.Genesis.Amount = genesisAmount * client.GetAPI().GetConfig().GetCoinPrecision()
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

func (client *Client) getBlockInfoTx(current *types.Block) (*tmtypes.QbftNodeAction, error) {
	//检查第一笔交易
	if len(current.Txs) == 0 {
		return nil, types.ErrEmptyTx
	}
	baseTx := current.Txs[0]

	var valAction tmtypes.QbftNodeAction
	err := types.Decode(baseTx.GetPayload(), &valAction)
	if err != nil {
		return nil, err
	}
	//检查交易类型
	if valAction.GetTy() != tmtypes.QbftNodeActionBlockInfo {
		return nil, ttypes.ErrBaseTxType
	}
	//检查交易内容
	if valAction.GetBlockInfo() == nil {
		return nil, ttypes.ErrBlockInfoTx
	}
	return &valAction, nil
}

// CheckBlock 检查区块
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	cfg := client.GetAPI().GetConfig()
	if current.Block.Difficulty != cfg.GetP(0).PowLimitBits {
		return types.ErrBlockHeaderDifficulty
	}
	valAction, err := client.getBlockInfoTx(current.Block)
	if err != nil {
		return err
	}
	if parent.Height+1 != current.Block.Height {
		return types.ErrBlockHeight
	}
	//判断exec 是否成功
	if current.Receipts[0].Ty != types.ExecOk {
		return ttypes.ErrBaseExecErr
	}
	info := valAction.GetBlockInfo()
	if current.Block.Height > 1 {
		lastValAction, err := client.getBlockInfoTx(parent)
		if err != nil {
			return err
		}
		lastInfo := lastValAction.GetBlockInfo()
		lastProposalBlock := &ttypes.QbftBlock{QbftBlock: lastInfo.GetBlock()}
		if !lastProposalBlock.HashesTo(info.Block.Header.LastBlockID.Hash) {
			return ttypes.ErrLastBlockID
		}
	}
	return nil
}

// ProcEvent reply not support action err
func (client *Client) ProcEvent(msg *queue.Message) bool {
	msg.ReplyErr("Client", types.ErrActionNotSupport)
	return true
}

// CreateBlock trigger consensus forward when tx available
func (client *Client) CreateBlock() {
	ticker := time.NewTicker(time.Duration(timeoutTxAvail.Load().(int32)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-client.ctx.Done():
			qbftlog.Info("CreateBlock quit")
			return
		case <-ticker.C:
			if !client.csState.IsRunning() {
				qbftlog.Info("consensus not running")
				break
			}

			height, err := client.getLastHeight()
			if err != nil {
				qbftlog.Info("getLastHeight fail", "err", err)
				break
			}
			if !client.CheckTxsAvailable(height) {
				break
			}

			if height+1 == client.csState.GetRoundState().Height {
				client.txsAvailable <- height + 1
			}
		}
	}
}

func (client *Client) getLastHeight() (int64, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		return -1, err
	}
	return lastBlock.Height, nil
}

// TxsAvailable check available channel
func (client *Client) TxsAvailable() <-chan int64 {
	return client.txsAvailable
}

// CheckTxsAvailable check whether some new transactions arriving
func (client *Client) CheckTxsAvailable(height int64) bool {
	txs := client.RequestTx(1, nil)
	txs = client.CheckTxDup(txs, height)
	return len(txs) != 0
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

// BuildBlock build a new block
func (client *Client) BuildBlock(height int64) *types.Block {
	lastBlock := client.WaitBlock(height)
	cfg := client.GetAPI().GetConfig()
	txs := client.RequestTx(int(cfg.GetP(lastBlock.Height+1).MaxTxNumber)-1, nil)
	// placeholder
	tx0 := &types.Transaction{}
	txs = append([]*types.Transaction{tx0}, txs...)

	var newblock types.Block
	newblock.ParentHash = lastBlock.Hash(cfg)
	newblock.Height = lastBlock.Height + 1
	client.AddTxsToBlock(&newblock, txs)
	//固定难度
	newblock.Difficulty = cfg.GetP(0).PowLimitBits
	newblock.BlockTime = types.Now().Unix()
	if lastBlock.BlockTime >= newblock.BlockTime {
		// 1秒内产生的多个区块使用相同时间戳
		if !sameBlocktime.Load().(bool) || lastBlock.BlockTime > newblock.BlockTime {
			newblock.BlockTime = lastBlock.BlockTime + 1
		}
	}
	return &newblock
}

// CommitBlock call WriteBlock to commit to chain
func (client *Client) CommitBlock(block *types.Block) {
	err := client.WriteBlock(nil, block)
	if err != nil {
		qbftlog.Warn("CommitBlock fail", "err", err)
		client.WaitBlock(block.Height)
	}
}

// WaitBlock by height
func (client *Client) WaitBlock(height int64) *types.Block {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	beg := time.Now()
	for {
		select {
		case <-client.ctx.Done():
			qbftlog.Info("WaitBlock quit")
			return nil
		case <-ticker.C:
			qbftlog.Info("Still waiting block......", "height", height, "cost", time.Since(beg))
		default:
			newHeight, err := client.getLastHeight()
			if err == nil && newHeight >= height {
				block, err := client.RequestBlock(height)
				if err == nil {
					return block
				}
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// QueryValidatorsByHeight ...
func (client *Client) QueryValidatorsByHeight(height int64) (*tmtypes.QbftNodes, error) {
	if height < 1 {
		return nil, ttypes.ErrHeightLessThanOne
	}
	req := &tmtypes.ReqQbftNodes{Height: height}
	param, err := proto.Marshal(req)
	if err != nil {
		qbftlog.Error("QueryValidatorsByHeight marshal", "err", err)
		return nil, types.ErrInvalidParam
	}
	msg := client.GetQueueClient().NewMessage("execs", types.EventBlockChainQuery,
		&types.ChainExecutor{Driver: "qbftNode", FuncName: "GetQbftNodeByHeight", StateHash: zeroHash[:], Param: param})
	err = client.GetQueueClient().Send(msg, true)
	if err != nil {
		qbftlog.Error("QueryValidatorsByHeight send", "err", err)
		return nil, err
	}
	msg, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		qbftlog.Info("QueryValidatorsByHeight result", "err", err)
		return nil, err
	}
	return msg.GetData().(types.Message).(*tmtypes.QbftNodes), nil
}

// QueryBlockInfoByHeight get blockInfo and block by height
func (client *Client) QueryBlockInfoByHeight(height int64) (*tmtypes.QbftBlockInfo, *types.Block, error) {
	if height < 1 {
		return nil, nil, ttypes.ErrHeightLessThanOne
	}
	block, err := client.RequestBlock(height)
	if err != nil {
		return nil, nil, err
	}
	valAction, err := client.getBlockInfoTx(block)
	if err != nil {
		return nil, nil, err
	}
	return valAction.GetBlockInfo(), block, nil
}

// LoadBlockCommit by height
func (client *Client) LoadBlockCommit(height int64) *tmtypes.QbftCommit {
	blockInfo, _, err := client.QueryBlockInfoByHeight(height)
	if err != nil {
		qbftlog.Error("LoadBlockCommit GetBlockInfo fail", "err", err)
		return nil
	}
	if height > 1 {
		seq, voteType := blockInfo.State.LastSequence, blockInfo.Block.LastCommit.VoteType
		if (seq == 0 && voteType != uint32(ttypes.VoteTypePrecommit)) ||
			(seq > 0 && voteType != uint32(ttypes.VoteTypePrevote)) {
			qbftlog.Error("LoadBlockCommit wrong VoteType", "seq", seq, "voteType", voteType)
			return nil
		}
	}
	return blockInfo.GetBlock().GetLastCommit()
}

// LoadBlockState by height
func (client *Client) LoadBlockState(height int64) *tmtypes.QbftState {
	blockInfo, _, err := client.QueryBlockInfoByHeight(height)
	if err != nil {
		qbftlog.Error("LoadBlockState GetBlockInfo fail", "err", err)
		return nil
	}
	return blockInfo.GetState()
}

// LoadProposalBlock by height
func (client *Client) LoadProposalBlock(height int64) *tmtypes.QbftBlock {
	blockInfo, block, err := client.QueryBlockInfoByHeight(height)
	if err != nil {
		qbftlog.Error("LoadProposal GetBlockInfo fail", "err", err)
		return nil
	}
	proposalBlock := blockInfo.GetBlock()
	proposalBlock.Data = block
	return proposalBlock
}

// Query_IsHealthy query whether consensus is sync
func (client *Client) Query_IsHealthy(req *types.ReqNil) (types.Message, error) {
	isHealthy := false
	if client.IsCaughtUp() {
		rs := client.csState.QueryRoundState()
		if rs != nil && client.GetCurrentHeight() <= rs.Height+1 {
			isHealthy = true
		}
	}
	return &tmtypes.QbftIsHealthy{IsHealthy: isHealthy}, nil
}

// Query_CurrentState query current consensus state
func (client *Client) Query_CurrentState(req *types.ReqNil) (types.Message, error) {
	state := client.csState.QueryState()
	if state == nil {
		return nil, ttypes.ErrConsensusQuery
	}
	return SaveState(*state), nil
}

// Query_NodeInfo query validator node info
func (client *Client) Query_NodeInfo(req *types.ReqNil) (types.Message, error) {
	rs := client.csState.QueryRoundState()
	if rs == nil {
		return nil, ttypes.ErrConsensusQuery
	}
	vals := rs.Validators.Validators
	nodes := make([]*tmtypes.QbftNodeInfo, 0)
	for _, val := range vals {
		if val == nil {
			nodes = append(nodes, &tmtypes.QbftNodeInfo{})
		} else {
			ipstr, idstr := "UNKOWN", "UNKOWN"
			pub, err := ttypes.ConsensusCrypto.PubKeyFromBytes(val.PubKey)
			if err != nil {
				qbftlog.Error("Query_NodeInfo invalid pubkey", "err", err)
			} else {
				id := GenIDByPubKey(pub)
				idstr = string(id)
				if id == client.node.ID {
					ipstr = client.node.IP
				} else {
					ip := client.node.peerSet.GetIP(id)
					if ip == nil {
						qbftlog.Error("Query_NodeInfo nil ip", "id", idstr)
					} else {
						ipstr = ip.String()
					}
				}
			}

			item := &tmtypes.QbftNodeInfo{
				NodeIP:      ipstr,
				NodeID:      idstr,
				Address:     fmt.Sprintf("%X", val.Address),
				PubKey:      fmt.Sprintf("%X", val.PubKey),
				VotingPower: val.VotingPower,
				Accum:       val.Accum,
			}
			nodes = append(nodes, item)
		}
	}
	return &tmtypes.QbftNodeInfoSet{Nodes: nodes}, nil
}

// CmpBestBlock 比较newBlock是不是最优区块
func (client *Client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	return false
}

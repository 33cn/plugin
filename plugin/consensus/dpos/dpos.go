// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package dpos

import (
	"bytes"
	"fmt"
	"github.com/33cn/chain33/common/address"
	"os"
	"time"

	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"github.com/33cn/chain33/util"
	ttypes "github.com/33cn/plugin/plugin/consensus/dpos/types"

	jsonrpc "github.com/33cn/chain33/rpc/jsonclient"
	rpctypes "github.com/33cn/chain33/rpc/types"
	dty "github.com/33cn/plugin/plugin/dapp/dposvote/types"
)

const dposVersion = "0.1.0"

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

	dposDelegateNum      int64 = 3 //委托节点个数，从配置读取，以后可以根据投票结果来定
	dposBlockInterval    int64 = 3 //出块间隔，当前按3s
	dposContinueBlockNum int64 = 6 //一个委托节点当选后，一次性持续出块数量
	dposCycle                  = dposDelegateNum * dposBlockInterval * dposContinueBlockNum
	dposPeriod                 = dposBlockInterval * dposContinueBlockNum
	zeroHash             [32]byte
	dposPort string = "36656"
	rpcAddr string = "http://0.0.0.0:8801"
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
	RpcAddr                   string   `json:"rpcAddr"`
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

	if subcfg.RpcAddr != "" {
		rpcAddr = subcfg.RpcAddr
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
	cr, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		dposlog.Error("NewDPosClient", "err", err)
		return nil
	}

	ttypes.ConsensusCrypto = cr

	//安全连接仍然要使用ed25519
	cr2, err := crypto.New(types.GetSignName("", types.ED25519))
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
	}
	c.SetChild(client)

	client.applyConfig(sub)
	return client
}

// PrivValidator returns the Node's PrivValidator.
// XXX: for convenience only!
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
OuterLoop:
	for !DebugCatchup {
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
	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}
	if block != nil {
		time.Sleep(time.Second * 5)
		cands, err := client.QueryCandidators()
		if err != nil {
			dposlog.Info("QueryCandidators failed", "err", err)
		} else {
			if len(cands) != int(dposDelegateNum) {
				dposlog.Info("QueryCandidators success but no enough candidators", "dposDelegateNum", dposDelegateNum, "candidatorNum", len(cands))
			} else {
				validators := make([]*ttypes.Validator, dposDelegateNum)
				for i, val := range cands {
					// Make validator
					validators[i] = &ttypes.Validator{
						Address: address.PubKeyToAddress(val.Pubkey).Hash160[:],
						PubKey:  val.Pubkey,
					}
				}
				valMgr.Validators = ttypes.NewValidatorSet(validators)
				dposlog.Info("QueryCandidators success and update validator set", "old validators", valMgrTmp.Validators.String(), "new validators", valMgr.Validators.String())
			}
		}
	}

	dposlog.Info("StartConsensus", "validators", valMgr.Validators)
	// Log whether this node is a delegator or an observer
	if valMgr.Validators.HasAddress(client.privValidator.GetAddress()) {
		dposlog.Info("This node is a delegator")
		client.isDelegator = true
	} else {
		dposlog.Info("This node is not a delegator")
	}

	// Make ConsensusReactor
	csState := NewConsensusState(client, valMgr)

	client.csState = csState

	csState.SetPrivValidator(client.privValidator, client.ValidatorIndex())

	// Create & add listener
	protocol, listeningAddress := "tcp", "0.0.0.0:" + dposPort
	node := NewNode(validatorNodes, protocol, listeningAddress, client.privKey, valMgr.ChainID, dposVersion, csState)

	client.node = node

	// 对于受托节点，才需要初始化区块，启动共识相关程序等,后续支持投票要做成动态切换的。
	if client.isDelegator {
		client.InitBlock()
		node.Start()
	}

	go client.MonitorCandidators()
	//go client.CreateBlock()
}

// GetGenesisBlockTime ...
func (client *Client) GetGenesisBlockTime() int64 {
	return genesisBlockTime
}

// CreateGenesisTx ...
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte("coins")
	tx.To = genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = 1e8 * types.Coin
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
	txs := client.RequestTx(int(types.GetP(lastBlock.Height+1).MaxTxNumber), nil)
	if len(txs) == 0 {
		block := client.GetCurrentBlock()
		if createEmptyBlocks {
			emptyBlock := &types.Block{}
			emptyBlock.StateHash = block.StateHash
			emptyBlock.ParentHash = block.Hash()
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
	newblock.ParentHash = lastBlock.Hash()
	newblock.Height = lastBlock.Height + 1
	client.AddTxsToBlock(&newblock, txs)
	//
	newblock.Difficulty = types.GetP(0).PowLimitBits
	newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
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

func (client *Client)QueryCandidators()([]*dty.Candidator, error) {
	var params rpctypes.Query4Jrpc
	params.Execer = dty.DPosX

	req := &dty.CandidatorQuery{
		TopN: int32(dposDelegateNum),
	}
	params.FuncName = dty.FuncNameQueryCandidatorByTopN
	params.Payload = types.MustPBToJSON(req)
	var res dty.CandidatorReply
	ctx := jsonrpc.NewRPCCtx(rpcAddr, "Chain33.Query", params, &res)

	result, err := ctx.RunResult()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return nil, err
	}
	res = *result.(*dty.CandidatorReply)
	return res.GetCandidators(), nil
}

func (client *Client)MonitorCandidators() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case <- ticker.C:
			dposlog.Info("Monitor Candidators")
			block, err := client.RequestLastBlock()
			if err != nil {
				panic(err)
			}

			if block != nil {
				cands, err := client.QueryCandidators()
				if err != nil {
					dposlog.Info("Query Candidators failed", "err", err)
				} else {
					if len(cands) != int(dposDelegateNum) {
						dposlog.Info("QueryCandidators success but no enough candidators", "dposDelegateNum", dposDelegateNum, "candidatorNum", len(cands))
					} else {
						validators := make([]*ttypes.Validator, dposDelegateNum)
						for i, val := range cands {
							// Make validator
							validators[i] = &ttypes.Validator{
								Address: address.PubKeyToAddress(val.Pubkey).Hash160[:],
								PubKey:  val.Pubkey,
							}
						}

						validatorSet := ttypes.NewValidatorSet(validators)
						dposlog.Info("QueryCandidators success and update validator set")
						if !client.isValidatorSetSame(validatorSet, client.csState.validatorMgr.Validators){
							dposlog.Info("ValidatorSet from contract is changed, so stop the node and restart the consensus.")
							client.node.Stop()
							time.Sleep(time.Second * 3)
							go client.StartConsensus()
						} else {
							dposlog.Info("ValidatorSet from contract is the same,no change.")
						}
					}
				}
			}

		}
	}
}

func (client *Client)isValidatorSetSame(v1, v2 *ttypes.ValidatorSet) bool {
	if v1 == nil || v2 == nil || len(v1.Validators) != len(v2.Validators){
		return false
	}

	for i := 0; i < len(v1.Validators); i++ {
		if !bytes.Equal(v1.Validators[i].PubKey, v2.Validators[i].PubKey){
			return false
		}
	}

	return true
}
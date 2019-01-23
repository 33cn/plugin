// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package para

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/33cn/chain33/common/log/log15"
	//"github.com/33cn/chain33/common"
	"encoding/hex"

	"github.com/33cn/chain33/client/api"
	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/crypto"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/rpc/grpcclient"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	paracross "github.com/33cn/plugin/plugin/dapp/paracross/types"
	pt "github.com/33cn/plugin/plugin/dapp/paracross/types"
)

const (
	addAct int64 = 1 //add para block action
	delAct int64 = 2 //reference blockstore.go, del para block action

	paraCrossTxCount = 2 //current only support 2 txs for cross

	minBlockNum = 6 //min block number startHeight before lastHeight in mainchain
)

var (
	plog                     = log.New("module", "para")
	grpcSite                 = "localhost:8802"
	genesisBlockTime   int64 = 1514533390
	startHeight        int64     //parachain sync from startHeight in mainchain
	blockSec           int64 = 5 //write block interval, second
	emptyBlockInterval int64 = 4 //write empty block every interval blocks in mainchain
	zeroHash           [32]byte
	grpcRecSize        = 30 * 1024 * 1024 //the size should be limited in server
	//current miner tx take any privatekey for unify all nodes sign purpose, and para chain is free
	minerPrivateKey               = "6da92a632ab7deb67d38c0f6560bcfed28167998f6496db64c258d5e8393a81b"
	searchHashMatchDepth    int32 = 100
	mainBlockHashForkHeight int64 = types.MaxHeight //calc block hash fork height in main chain
)

func init() {
	drivers.Reg("para", New)
	drivers.QueryData.Register("para", &client{})
}

type client struct {
	*drivers.BaseClient
	grpcClient      types.Chain33Client
	paraClient      paracross.ParacrossClient
	execAPI         api.ExecutorAPI
	isCaughtUp      bool
	commitMsgClient *commitMsgClient
	authAccount     string
	privateKey      crypto.PrivKey
	wg              sync.WaitGroup
	subCfg          *subConfig
}

type subConfig struct {
	WriteBlockSeconds           int64  `json:"writeBlockSeconds,omitempty"`
	ParaRemoteGrpcClient        string `json:"paraRemoteGrpcClient,omitempty"`
	StartHeight                 int64  `json:"startHeight,omitempty"`
	EmptyBlockInterval          int64  `json:"emptyBlockInterval,omitempty"`
	AuthAccount                 string `json:"authAccount,omitempty"`
	WaitBlocks4CommitMsg        int32  `json:"waitBlocks4CommitMsg,omitempty"`
	SearchHashMatchedBlockDepth int32  `json:"searchHashMatchedBlockDepth,omitempty"`
	GenesisAmount               int64  `json:"genesisAmount,omitempty"`
}

// New function to init paracross env
func New(cfg *types.Consensus, sub []byte) queue.Module {
	c := drivers.NewBaseClient(cfg)
	var subcfg subConfig
	if sub != nil {
		types.MustDecode(sub, &subcfg)
	}
	if subcfg.GenesisAmount <= 0 {
		subcfg.GenesisAmount = 1e8
	}
	if subcfg.ParaRemoteGrpcClient != "" {
		grpcSite = subcfg.ParaRemoteGrpcClient
	}
	if subcfg.StartHeight > 0 {
		startHeight = subcfg.StartHeight
	}
	if subcfg.WriteBlockSeconds > 0 {
		blockSec = subcfg.WriteBlockSeconds
	}
	if subcfg.EmptyBlockInterval > 0 {
		emptyBlockInterval = subcfg.EmptyBlockInterval
	}
	if subcfg.SearchHashMatchedBlockDepth > 0 {
		searchHashMatchDepth = subcfg.SearchHashMatchedBlockDepth
	}

	pk, err := hex.DecodeString(minerPrivateKey)
	if err != nil {
		panic(err)
	}
	secp, err := crypto.New(types.GetSignName("", types.SECP256K1))
	if err != nil {
		panic(err)
	}
	priKey, err := secp.PrivKeyFromBytes(pk)
	if err != nil {
		panic(err)
	}

	plog.Debug("New Para consensus client")

	//msgRecvOp := grpc.WithMaxMsgSize(grpcRecSize)
	//conn, err := grpc.Dial(grpcSite, grpc.WithInsecure(), msgRecvOp)
	//if err != nil {
	//	panic(err)
	//}
	//_ = types.NewChain33Client(conn)

	//grpcClient := types.NewChain33Client(conn)
	grpcCli, err := grpcclient.NewMainChainClient(grpcSite)
	if err != nil {
		panic(err)
	}

	para := &client{
		BaseClient:  c,
		grpcClient:  grpcCli,
		authAccount: subcfg.AuthAccount,
		privateKey:  priKey,
		isCaughtUp:  false,
		subCfg:      &subcfg,
	}
	if subcfg.WaitBlocks4CommitMsg < 2 {
		panic("config WaitBlocks4CommitMsg should not less 2")
	}
	para.commitMsgClient = &commitMsgClient{
		paraClient:      para,
		waitMainBlocks:  subcfg.WaitBlocks4CommitMsg,
		commitMsgNotify: make(chan int64, 1),
		delMsgNotify:    make(chan int64, 1),
		mainBlockAdd:    make(chan *types.BlockDetail, 1),
		quit:            make(chan struct{}),
	}
	c.SetChild(para)
	return para
}

//para 不检查任何的交易
func (client *client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	err := checkMinerTx(current)
	return err
}

func (client *client) Close() {
	client.BaseClient.Close()
	close(client.commitMsgClient.quit)
	client.wg.Wait()
	plog.Info("consensus para closed")
}

func (client *client) SetQueueClient(c queue.Client) {
	plog.Info("Enter SetQueueClient method of Para consensus")
	client.InitClient(c, func() {
		client.InitBlock()
	})
	go client.EventLoop()

	client.wg.Add(1)
	go client.commitMsgClient.handler()
	go client.CreateBlock()
}

func (client *client) InitBlock() {
	var err error
	mainBlockHashForkHeight, err = client.GetBlockHashForkHeightOnMainChain()
	if err != nil {
		panic(err)
	}

	client.execAPI = api.New(client.BaseClient.GetAPI(), client.grpcClient)

	block, err := client.RequestLastBlock()
	if err != nil {
		panic(err)
	}

	if block == nil {
		startSeq := client.GetStartSeq(startHeight)
		// 创世区块
		newblock := &types.Block{}
		newblock.Height = 0
		newblock.BlockTime = genesisBlockTime
		newblock.ParentHash = zeroHash[:]
		newblock.MainHash = zeroHash[:]
		tx := client.CreateGenesisTx()
		newblock.Txs = tx
		newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
		client.WriteBlock(zeroHash[:], newblock, startSeq-1)
	} else {
		client.SetCurrentBlock(block)
	}
}

// GetStartSeq get startSeq in mainchain
func (client *client) GetStartSeq(height int64) int64 {
	if height == 0 {
		return 0
	}

	lastHeight, err := client.GetLastHeightOnMainChain()
	if err != nil {
		panic(err)
	}
	if lastHeight < height {
		panic(fmt.Sprintf("lastHeight(%d) less than startHeight(%d) in mainchain", lastHeight, height))
	}

	hint := time.NewTicker(5 * time.Second)
	for lastHeight < height+minBlockNum {
		select {
		case <-hint.C:
			plog.Info("Waiting lastHeight increase......", "lastHeight", lastHeight, "startHeight", height)
		default:
			lastHeight, err = client.GetLastHeightOnMainChain()
			if err != nil {
				panic(err)
			}
			time.Sleep(time.Second)
		}
	}
	hint.Stop()
	plog.Info(fmt.Sprintf("lastHeight more than %d blocks after startHeight", minBlockNum), "lastHeight", lastHeight, "startHeight", height)

	seq, err := client.GetSeqByHeightOnMainChain(height)
	if err != nil {
		panic(err)
	}
	plog.Info("the start sequence in mainchain", "startHeight", height, "startSeq", seq)
	return seq
}

func (client *client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte(types.ExecName(cty.CoinsX))
	tx.To = client.Cfg.Genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = client.subCfg.GenesisAmount * types.Coin
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

func (client *client) ProcEvent(msg queue.Message) bool {
	return false
}

//1. 如果涉及跨链合约，如果有超过两条平行链的交易被判定为失败，交易组会执行不成功,也不PACK。（这样的情况下，主链交易一定会执行不成功）
//2. 如果不涉及跨链合约，那么交易组没有任何规定，可以是20比，10条链。 如果主链交易有失败，平行链也不会执行
//3. 如果交易组有一个ExecOk,主链上的交易都是ok的，可以全部打包
//4. 如果全部是ExecPack，有两种情况，一是交易组所有交易都是平行链交易，另一是主链有交易失败而打包了的交易，需要检查LogErr，如果有错，全部不打包
func calcParaCrossTxGroup(tx *types.Transaction, main *types.BlockDetail, index int) ([]*types.Transaction, int) {
	var headIdx int

	for i := index; i >= 0; i-- {
		if bytes.Equal(tx.Header, main.Block.Txs[i].Hash()) {
			headIdx = i
			break
		}
	}

	endIdx := headIdx + int(tx.GroupCount)
	for i := headIdx; i < endIdx; i++ {
		if types.IsMyParaExecName(string(main.Block.Txs[i].Execer)) {
			continue
		}
		if main.Receipts[i].Ty == types.ExecOk {
			return main.Block.Txs[headIdx:endIdx], endIdx
		}

		for _, log := range main.Receipts[i].Logs {
			if log.Ty == types.TyLogErr {
				return nil, endIdx
			}
		}
	}
	//全部是平行链交易 或主链执行非失败的tx
	return main.Block.Txs[headIdx:endIdx], endIdx
}

func (client *client) FilterTxsForPara(main *types.BlockDetail) []*types.Transaction {
	var txs []*types.Transaction
	for i := 0; i < len(main.Block.Txs); i++ {
		tx := main.Block.Txs[i]
		if types.IsMyParaExecName(string(tx.Execer)) {
			if tx.GroupCount >= paraCrossTxCount {
				mainTxs, endIdx := calcParaCrossTxGroup(tx, main, i)
				txs = append(txs, mainTxs...)
				i = endIdx - 1
				continue
			}
			txs = append(txs, tx)
		}
	}
	return txs
}

//get the last sequence in parachain
func (client *client) GetLastSeq() (int64, error) {
	blockedSeq, err := client.GetAPI().GetLastBlockSequence()
	if err != nil {
		return -2, err
	}
	return blockedSeq.Data, nil
}

func (client *client) GetBlockedSeq(hash []byte) (int64, error) {
	//from blockchain db
	blockedSeq, err := client.GetAPI().GetSequenceByHash(&types.ReqHash{Hash: hash})
	if err != nil {
		return -2, err
	}
	return blockedSeq.Data, nil

}

func (client *client) GetBlockByHeight(height int64) (*types.Block, error) {
	//from blockchain db
	blockDetails, err := client.GetAPI().GetBlocks(&types.ReqBlocks{Start: height, End: height})
	if err != nil {
		plog.Error("paracommitmsg get node status block count fail")
		return nil, err
	}
	if 1 != int64(len(blockDetails.Items)) {
		plog.Error("paracommitmsg get node status block count fail")
		return nil, types.ErrInvalidParam
	}
	return blockDetails.Items[0].Block, nil
}

func (client *client) getLastBlockInfo() (int64, *types.Block, []byte, int64, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, nil, -2, err
	}
	blockedSeq, err := client.GetBlockedSeq(lastBlock.Hash())
	if err != nil {
		plog.Error("Parachain GetBlockedSeq fail", "err", err)
		return -2, nil, nil, -2, err
	}

	if lastBlock.Height > 0 {
		miner, err := getMinerTxInfo(lastBlock)
		if err != nil {
			return -2, nil, nil, -2, err
		}
		return blockedSeq, lastBlock, miner.MainBlockHash, miner.MainBlockHeight, nil
	}

	//sequence in main chain start from 0
	seq := blockedSeq
	if seq == -1 {
		seq = 0
	}
	main, err := client.GetBlockOnMainBySeq(seq)
	if err != nil {
		return -2, nil, nil, -2, err
	}
	return blockedSeq, lastBlock, main.Seq.Hash, main.Detail.Block.Height, nil

}

func (client *client) GetBlockHashForkHeightOnMainChain() (int64, error) {
	ret, err := client.grpcClient.GetFork(context.Background(), &types.ReqKey{Key: []byte("ForkBlockHash")})
	if err != nil {
		plog.Error("para get rpc ForkBlockHash fail", "err", err.Error())
		return -1, err
	}

	return ret.Data, nil
}

func (client *client) GetLastHeightOnMainChain() (int64, error) {
	header, err := client.grpcClient.GetLastHeader(context.Background(), &types.ReqNil{})
	if err != nil {
		plog.Error("GetLastHeightOnMainChain", "Error", err.Error())
		return -1, err
	}
	return header.Height, nil
}

func (client *client) GetLastSeqOnMainChain() (int64, error) {
	seq, err := client.grpcClient.GetLastBlockSequence(context.Background(), &types.ReqNil{})
	if err != nil {
		plog.Error("GetLastSeqOnMainChain", "Error", err.Error())
		return -1, err
	}
	//the reflect checked in grpcHandle
	return seq.Data, nil
}

func (client *client) GetSeqByHeightOnMainChain(height int64) (int64, error) {
	hash, err := client.GetHashByHeightOnMainChain(height)
	if err != nil {
		return -1, err
	}
	seq, err := client.GetSeqByHashOnMainChain(hash)
	return seq, err
}

func (client *client) GetHashByHeightOnMainChain(height int64) ([]byte, error) {
	reply, err := client.grpcClient.GetBlockHash(context.Background(), &types.ReqInt{Height: height})
	if err != nil {
		plog.Error("GetHashByHeightOnMainChain", "Error", err.Error())
		return nil, err
	}
	return reply.Hash, nil
}

func (client *client) GetSeqByHashOnMainChain(hash []byte) (int64, error) {
	seq, err := client.grpcClient.GetSequenceByHash(context.Background(), &types.ReqHash{Hash: hash})
	if err != nil {
		plog.Error("GetSeqByHashOnMainChain", "Error", err.Error())
		return -1, err
	}
	//the reflect checked in grpcHandle
	return seq.Data, nil
}

func (client *client) GetBlockOnMainBySeq(seq int64) (*types.BlockSeq, error) {
	blockSeq, err := client.grpcClient.GetBlockBySeq(context.Background(), &types.Int64{Data: seq})
	if err != nil {
		plog.Error("Not found block on main", "seq", seq)
		return nil, err
	}

	hash := blockSeq.Detail.Block.HashByForkHeight(mainBlockHashForkHeight)
	if !bytes.Equal(blockSeq.Seq.Hash, hash) {
		plog.Error("para compare ForkBlockHash fail", "forkHeight", mainBlockHashForkHeight,
			"seqHash", hex.EncodeToString(blockSeq.Seq.Hash), "calcHash", hex.EncodeToString(hash))
		return nil, types.ErrBlockHashNoMatch
	}

	return blockSeq, nil
}

// preBlockHash to identify the same main node
func (client *client) RequestTx(currSeq int64, preMainBlockHash []byte) ([]*types.Transaction, *types.BlockSeq, error) {
	plog.Debug("Para consensus RequestTx")
	lastSeq, err := client.GetLastSeqOnMainChain()
	if err != nil {
		return nil, nil, err
	}
	plog.Info("RequestTx", "LastMainSeq", lastSeq, "CurrSeq", currSeq)
	if lastSeq >= currSeq {
		blockSeq, err := client.GetBlockOnMainBySeq(currSeq)
		if err != nil {
			return nil, nil, err
		}
		//genesis block start with seq=-1 not check
		if currSeq == 0 ||
			(bytes.Equal(preMainBlockHash, blockSeq.Detail.Block.ParentHash) && blockSeq.Seq.Type == addAct) ||
			(bytes.Equal(preMainBlockHash, blockSeq.Seq.Hash) && blockSeq.Seq.Type == delAct) {

			txs := client.FilterTxsForPara(blockSeq.Detail)
			plog.Info("GetCurrentSeq", "Len of txs", len(txs), "seqTy", blockSeq.Seq.Type)

			if lastSeq-currSeq > emptyBlockInterval {
				client.isCaughtUp = false
			} else {
				client.isCaughtUp = true
			}

			if client.authAccount != "" {
				client.commitMsgClient.onMainBlockAdded(blockSeq.Detail)
			}

			return txs, blockSeq, nil
		}
		//not consistent case be processed at below
		plog.Error("RequestTx", "preMainHash", hex.EncodeToString(preMainBlockHash), "currSeq preMainHash", hex.EncodeToString(blockSeq.Detail.Block.ParentHash),
			"currSeq mainHash", hex.EncodeToString(blockSeq.Seq.Hash), "curr seq", currSeq, "ty", blockSeq.Seq.Type, "currSeq Mainheight", blockSeq.Detail.Block.Height)
		return nil, nil, paracross.ErrParaCurHashNotMatch
	}
	//lastSeq < CurrSeq case:
	//lastSeq = currSeq-1, main node not update
	if lastSeq+1 == currSeq {
		plog.Debug("Waiting new sequence from main chain")
		return nil, nil, paracross.ErrParaWaitingNewSeq
	}

	// 1. lastSeq < currSeq-1
	// 2. lastSeq >= currSeq and seq not consistent or fork case
	return nil, nil, paracross.ErrParaCurHashNotMatch
}

//genesis block scenario,  new main node's blockHash as preMainHash, genesis sequence+1 as currSeq
// for genesis seq=-1 scenario, mainHash not care, as the 0 seq instead of -1
// not seq=-1 scenario, mainHash needed
func (client *client) syncFromGenesisBlock() (int64, []byte, error) {
	lastSeq, _, lastSeqMainHash, _, err := client.getLastBlockInfo()
	if err != nil {
		plog.Error("Parachain getLastBlockInfo fail", "err", err)
		return -2, nil, err
	}
	plog.Info("syncFromGenesisBlock sync from height 0")
	return lastSeq + 1, lastSeqMainHash, nil
}

// search base on para block but not last MainBlockHash, last MainBlockHash can not back tracing
func (client *client) switchHashMatchedBlock(currSeq int64) (int64, []byte, error) {
	lastBlock, err := client.RequestLastBlock()
	if err != nil {
		plog.Error("Parachain RequestLastBlock fail", "err", err)
		return -2, nil, err
	}

	if lastBlock.Height == 0 {
		return client.syncFromGenesisBlock()
	}

	depth := searchHashMatchDepth
	for height := lastBlock.Height; height > 0 && depth > 0; height-- {
		block, err := client.GetBlockByHeight(height)
		if err != nil {
			return -2, nil, err
		}
		//当前block结构已经有mainHash和MainHeight但是从blockchain获取的block还没有写入，以后如果获取到，可以替换从minerTx获取
		miner, err := getMinerTxInfo(block)
		if err != nil {
			return -2, nil, err
		}
		plog.Info("switchHashMatchedBlock", "lastParaBlock height", miner.Height, "mainHeight",
			miner.MainBlockHeight, "mainHash", hex.EncodeToString(miner.MainBlockHash))
		mainSeq, err := client.GetSeqByHashOnMainChain(miner.MainBlockHash)
		if err != nil {
			depth--
			if depth == 0 {
				plog.Error("switchHashMatchedBlock depth overflow", "last info:mainHeight", miner.MainBlockHeight,
					"mainHash", hex.EncodeToString(miner.MainBlockHash), "search startHeight", lastBlock.Height, "curHeight", miner.Height,
					"search depth", searchHashMatchDepth)
				panic("search HashMatchedBlock overflow, re-setting search depth and restart to try")
			}
			if height == 1 {
				plog.Error("switchHashMatchedBlock search to height=1 not found", "lastBlockHeight", lastBlock.Height,
					"height1 mainHash", hex.EncodeToString(miner.MainBlockHash))
				err = client.removeBlocks(0)
				if err != nil {
					return currSeq, nil, nil
				}
				return client.syncFromGenesisBlock()

			}
			continue
		}

		//remove fail, the para chain may be remove part, set the preMainBlockHash to nil, to match nothing, force to search from last
		err = client.removeBlocks(height)
		if err != nil {
			return currSeq, nil, nil
		}

		plog.Info("switchHashMatchedBlock succ", "currHeight", height, "initHeight", lastBlock.Height,
			"new currSeq", mainSeq+1, "new preMainBlockHash", hex.EncodeToString(miner.MainBlockHash))
		return mainSeq + 1, miner.MainBlockHash, nil
	}
	return -2, nil, paracross.ErrParaCurHashNotMatch
}

func (client *client) removeBlocks(endHeight int64) error {
	for {
		lastBlock, err := client.RequestLastBlock()
		if err != nil {
			plog.Error("Parachain RequestLastBlock fail", "err", err)
			return err
		}
		if lastBlock.Height == endHeight {
			return nil
		}

		blockedSeq, err := client.GetBlockedSeq(lastBlock.Hash())
		if err != nil {
			plog.Error("Parachain GetBlockedSeq fail", "err", err)
			return err
		}

		err = client.DelBlock(lastBlock, blockedSeq)
		if err != nil {
			plog.Error("Parachain GetBlockedSeq fail", "err", err)
			return err
		}
		plog.Info("Parachain removeBlocks succ", "localParaHeight", lastBlock.Height, "blockedSeq", blockedSeq)
	}
}

//正常情况下，打包交易
func (client *client) CreateBlock() {
	incSeqFlag := true
	//system startup, take the last added block's seq is ok
	currSeq, _, lastSeqMainHash, _, err := client.getLastBlockInfo()
	if err != nil {
		plog.Error("Parachain getLastBlockInfo fail", "err", err.Error())
		return
	}
	for {
		//should be lastSeq but not LastBlockSeq as del block case the seq is not equal
		lastSeq, err := client.GetLastSeq()
		if err != nil {
			plog.Error("Parachain GetLastSeq fail", "err", err.Error())
			time.Sleep(time.Second)
			continue
		}

		if incSeqFlag || currSeq == lastSeq {
			currSeq++
		}

		txs, blockOnMain, err := client.RequestTx(currSeq, lastSeqMainHash)
		if err != nil {
			incSeqFlag = false
			if err == paracross.ErrParaCurHashNotMatch {
				newSeq, newSeqMainHash, err := client.switchHashMatchedBlock(currSeq)
				if err == nil {
					currSeq = newSeq
					lastSeqMainHash = newSeqMainHash
					continue
				}
			}
			time.Sleep(time.Second * time.Duration(blockSec))
			continue
		}

		lastSeqMainHeight := blockOnMain.Detail.Block.Height
		lastSeqMainHash = blockOnMain.Seq.Hash
		if blockOnMain.Seq.Type == delAct {
			lastSeqMainHash = blockOnMain.Detail.Block.ParentHash
		}

		_, lastBlock, lastBlockMainHash, lastBlockMainHeight, err := client.getLastBlockInfo()
		if err != nil {
			plog.Error("Parachain getLastBlockInfo fail", "err", err)
			time.Sleep(time.Second)
			continue
		}

		plog.Info("Parachain process block", "lastBlockSeq", lastSeq, "curSeq", currSeq,
			"currSeqMainHeight", lastSeqMainHeight, "currSeqMainHash", common.ToHex(lastSeqMainHash),
			"lastBlockMainHeight", lastBlockMainHeight, "lastBlockMainHash", common.ToHex(lastBlockMainHash), "seqTy", blockOnMain.Seq.Type)

		if blockOnMain.Seq.Type == delAct {
			if len(txs) == 0 {
				if lastSeqMainHeight > lastBlockMainHeight {
					incSeqFlag = true
					continue
				}
				plog.Info("Delete empty block")
			}
			err := client.DelBlock(lastBlock, currSeq)
			incSeqFlag = false
			if err != nil {
				plog.Error(fmt.Sprintf("********************err:%v", err.Error()))
			}
		} else if blockOnMain.Seq.Type == addAct {
			if len(txs) == 0 {
				if lastSeqMainHeight-lastBlockMainHeight < emptyBlockInterval {
					incSeqFlag = true
					continue
				}
				plog.Info("Create empty block")
			}
			err := client.createBlock(lastBlock, txs, currSeq, blockOnMain)
			incSeqFlag = false
			if err != nil {
				plog.Error(fmt.Sprintf("********************err:%v", err.Error()))
			}
		} else {
			plog.Error("Incorrect sequence type")
			incSeqFlag = false
		}
		if client.isCaughtUp {
			time.Sleep(time.Second * time.Duration(blockSec))
		}
	}
}

// miner tx need all para node create, but not all node has auth account, here just not sign to keep align
func (client *client) addMinerTx(preStateHash []byte, block *types.Block, main *types.BlockSeq) error {
	status := &pt.ParacrossNodeStatus{
		Title:           types.GetTitle(),
		Height:          block.Height,
		PreBlockHash:    block.ParentHash,
		PreStateHash:    preStateHash,
		MainBlockHash:   main.Seq.Hash,
		MainBlockHeight: main.Detail.Block.Height,
	}
	tx, err := paracross.CreateRawMinerTx(status)
	if err != nil {
		return err
	}
	tx.Sign(types.SECP256K1, client.privateKey)
	block.Txs = append([]*types.Transaction{tx}, block.Txs...)
	return nil
}

func (client *client) createBlock(lastBlock *types.Block, txs []*types.Transaction, seq int64, mainBlock *types.BlockSeq) error {
	var newblock types.Block
	plog.Debug(fmt.Sprintf("the len txs is: %v", len(txs)))

	newblock.ParentHash = lastBlock.Hash()
	newblock.Height = lastBlock.Height + 1
	newblock.Txs = txs
	err := client.addMinerTx(lastBlock.StateHash, &newblock, mainBlock)
	if err != nil {
		return err
	}
	//挖矿固定难度
	newblock.Difficulty = types.GetP(0).PowLimitBits
	newblock.TxHash = merkle.CalcMerkleRoot(newblock.Txs)
	newblock.BlockTime = mainBlock.Detail.Block.BlockTime
	newblock.MainHash = mainBlock.Seq.Hash
	newblock.MainHeight = mainBlock.Detail.Block.Height

	err = client.WriteBlock(lastBlock.StateHash, &newblock, seq)

	plog.Debug("para create new Block", "newblock.ParentHash", common.ToHex(newblock.ParentHash),
		"newblock.Height", newblock.Height, "newblock.TxHash", common.ToHex(newblock.TxHash),
		"newblock.BlockTime", newblock.BlockTime, "sequence", seq)
	return err
}

// 向blockchain写区块
func (client *client) WriteBlock(prev []byte, paraBlock *types.Block, seq int64) error {
	//共识模块不执行block，统一由blockchain模块执行block并做去重的处理，返回执行后的blockdetail
	blockDetail := &types.BlockDetail{Block: paraBlock}

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blockDetail, Sequence: seq}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventAddParaChainBlockDetail, parablockDetail)
	client.GetQueueClient().Send(msg, true)
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blkdetail := resp.GetData().(*types.BlockDetail)
	if blkdetail == nil {
		return errors.New("block detail is nil")
	}

	client.SetCurrentBlock(blkdetail.Block)

	if client.authAccount != "" {
		client.commitMsgClient.onBlockAdded(blkdetail.Block.Height)
	}

	return nil
}

// 向blockchain删区块
func (client *client) DelBlock(block *types.Block, seq int64) error {
	plog.Debug("delete block in parachain")
	start := block.Height
	if start == 0 {
		panic("Parachain attempt to Delete GenesisBlock !")
	}
	msg := client.GetQueueClient().NewMessage("blockchain", types.EventGetBlocks, &types.ReqBlocks{Start: start, End: start, IsDetail: true, Pid: []string{""}})
	client.GetQueueClient().Send(msg, true)
	resp, err := client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}
	blocks := resp.GetData().(*types.BlockDetails)

	parablockDetail := &types.ParaChainBlockDetail{Blockdetail: blocks.Items[0], Sequence: seq}
	msg = client.GetQueueClient().NewMessage("blockchain", types.EventDelParaChainBlockDetail, parablockDetail)
	client.GetQueueClient().Send(msg, true)
	resp, err = client.GetQueueClient().Wait(msg)
	if err != nil {
		return err
	}

	if resp.GetData().(*types.Reply).IsOk {
		if client.authAccount != "" {
			client.commitMsgClient.onBlockDeleted(blocks.Items[0].Block.Height)
		}
	} else {
		reply := resp.GetData().(*types.Reply)
		return errors.New(string(reply.GetMsg()))
	}
	return nil
}

//IsCaughtUp 是否追上最新高度,
func (client *client) Query_IsCaughtUp(req *types.ReqNil) (types.Message, error) {
	if client == nil {
		return nil, fmt.Errorf("%s", "client not bind message queue.")
	}
	return &types.IsCaughtUp{Iscaughtup: client.isCaughtUp}, nil
}

func checkMinerTx(current *types.BlockDetail) error {
	//检查第一个笔交易的execs, 以及执行状态
	if len(current.Block.Txs) == 0 {
		return types.ErrEmptyTx
	}
	baseTx := current.Block.Txs[0]
	//判断交易类型和执行情况
	var action paracross.ParacrossAction
	err := types.Decode(baseTx.GetPayload(), &action)
	if err != nil {
		return err
	}
	if action.GetTy() != paracross.ParacrossActionMiner {
		return paracross.ErrParaMinerTxType
	}
	//判断交易执行是否OK
	if action.GetMiner() == nil {
		return paracross.ErrParaEmptyMinerTx
	}

	//判断exec 是否成功
	if current.Receipts[0].Ty != types.ExecOk {
		return paracross.ErrParaMinerExecErr
	}
	return nil
}

func getMinerTxInfo(block *types.Block) (*paracross.ParacrossNodeStatus, error) {
	baseTx := block.Txs[0]
	//判断交易类型和执行情况
	var action paracross.ParacrossAction
	err := types.Decode(baseTx.GetPayload(), &action)
	if err != nil {
		return nil, err
	}
	if action.GetTy() != paracross.ParacrossActionMiner {
		return nil, paracross.ErrParaMinerTxType
	}

	return action.GetMiner().Status, nil
}

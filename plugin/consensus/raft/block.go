// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"github.com/coreos/etcd/snap"
)

func init() {
	drivers.Reg("raft", NewRaftCluster)
	drivers.QueryData.Register("raft", &Client{})
}

// Client Raft implementation
type Client struct {
	*drivers.BaseClient
	proposeC    chan<- BlockInfo
	commitC     <-chan *BlockInfo
	errorC      <-chan error
	snapshotter *snap.Snapshotter
	validatorC  <-chan bool
	ctx         context.Context
	cancel      context.CancelFunc
	blockInfo   *BlockInfo
	mtx         sync.Mutex
}

// NewBlockstore create Raft Client
func NewBlockstore(ctx context.Context, cfg *types.Consensus, snapshotter *snap.Snapshotter, proposeC chan<- BlockInfo, commitC <-chan *BlockInfo, errorC <-chan error, validatorC <-chan bool, cancel context.CancelFunc) *Client {
	c := drivers.NewBaseClient(cfg)
	client := &Client{BaseClient: c, proposeC: proposeC, snapshotter: snapshotter, validatorC: validatorC, commitC: commitC, errorC: errorC, ctx: ctx, cancel: cancel}
	c.SetChild(client)
	return client
}

// GetGenesisBlockTime get genesis blocktime
func (client *Client) GetGenesisBlockTime() int64 {
	return genesisBlockTime
}

// CreateGenesisTx get genesis tx
func (client *Client) CreateGenesisTx() (ret []*types.Transaction) {
	var tx types.Transaction
	tx.Execer = []byte(cty.CoinsX)
	tx.To = genesis
	//gen payload
	g := &cty.CoinsAction_Genesis{}
	g.Genesis = &types.AssetsGenesis{}
	g.Genesis.Amount = 1e8 * types.Coin
	tx.Payload = types.Encode(&cty.CoinsAction{Value: g, Ty: cty.CoinsActionGenesis})
	ret = append(ret, &tx)
	return
}

// ProcEvent method
func (client *Client) ProcEvent(msg *queue.Message) bool {
	return false
}

// CheckBlock method
func (client *Client) CheckBlock(parent *types.Block, current *types.BlockDetail) error {
	cfg := client.GetAPI().GetConfig()
	if current.Block.Difficulty != cfg.GetP(0).PowLimitBits {
		return types.ErrBlockHeaderDifficulty
	}
	return nil
}

func (client *Client) getSnapshot() ([]byte, error) {
	return json.Marshal(client.GetCurrentInfo())
}

func (client *Client) recoverFromSnapshot(snapshot []byte) error {
	var info *BlockInfo
	if err := json.Unmarshal(snapshot, info); err != nil {
		return err
	}
	client.SetCurrentInfo(info)
	return nil
}

// SetQueueClient method
func (client *Client) SetQueueClient(c queue.Client) {
	rlog.Info("Enter SetQueue method of raft consensus")
	client.InitClient(c, func() {
		client.InitBlock()
	})
	go client.EventLoop()
	go client.readCommits(client.commitC, client.errorC)
	go client.pollingTask()
}

// Close method
func (client *Client) Close() {
	client.cancel()
	rlog.Info("consensus raft closed")
}

// CreateBlock method
func (client *Client) CreateBlock() {
	retry := 0
	count := int64(0)
	cfg := client.GetAPI().GetConfig()
	//打包区块前先同步到最大高度
	for {
		if client.IsCaughtUp() {
			rlog.Info("Leader has caught up the max height")
			break
		}
		time.Sleep(time.Second)
		retry++
		if retry >= 600 {
			panic("Leader encounter problem, exit.")
		}
	}
	curBlock, err := client.RequestLastBlock()
	if err != nil {
		rlog.Error("Leader RequestLastBlock fail", "err", err)
		panic(err)
	}
	curInfo := &BlockInfo{
		Height: curBlock.Height,
		Hash:   common.ToHex(curBlock.Hash(cfg)),
	}
	client.SetCurrentInfo(curInfo)

	ticker := time.NewTicker(time.Duration(writeBlockSeconds) * time.Second)
	hint := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	defer hint.Stop()
	for {
		select {
		case <-client.ctx.Done():
			return
		case <-hint.C:
			rlog.Info("==================This is Leader node=====================")
		case <-ticker.C:
			//如果leader节点突然挂了，不是打包节点，需要退出
			if !mux.Load().(bool) {
				rlog.Warn("Not the Leader node anymore")
				return
			}

			lastBlock, err := client.RequestLastBlock()
			if err != nil {
				rlog.Error("Leader RequestLastBlock fail", "err", err)
				break
			}
			if client.GetCurrentInfoHeight() != lastBlock.Height {
				rlog.Info("Leader wait commit blockInfo", "infoHeight", client.GetCurrentInfoHeight(),
					"blockHeight", lastBlock.Height)
				break
			}

			txs := client.RequestTx(int(cfg.GetP(lastBlock.Height+1).MaxTxNumber), nil)
			if len(txs) == 0 {
				count++
				//not create empty block when emptyBlockInterval is 0
				if emptyBlockInterval == 0 || count < emptyBlockInterval/writeBlockSeconds {
					break
				}
				//create empty block every no tx in emptyBlockInterval seconds
				rlog.Info("Leader create empty block")
			}

			var newblock types.Block
			newblock.ParentHash = lastBlock.Hash(cfg)
			newblock.Height = lastBlock.Height + 1
			client.AddTxsToBlock(&newblock, txs)
			//需要首先对交易进行排序然后再计算TxHash
			if cfg.IsFork(newblock.Height, "ForkRootHash") {
				newblock.Txs = types.TransactionSort(newblock.Txs)
			}
			newblock.TxHash = merkle.CalcMerkleRoot(cfg, newblock.Height, newblock.Txs)
			//固定难度
			newblock.Difficulty = cfg.GetP(0).PowLimitBits
			newblock.BlockTime = types.Now().Unix()
			if lastBlock.BlockTime >= newblock.BlockTime {
				newblock.BlockTime = lastBlock.BlockTime + 1
			}
			err = client.WriteBlock(lastBlock.StateHash, &newblock)
			if err != nil {
				rlog.Error("Leader WriteBlock fail", "err", err)
				break
			}

			info := BlockInfo{
				Height: newblock.Height,
				Hash:   common.ToHex(newblock.Hash(cfg)),
			}
			client.propose(info)
			count = 0
		}

	}
}

// 向raft底层发送BlockInfo
func (client *Client) propose(info BlockInfo) {
	client.proposeC <- info
}

// 从receive channel中读leader发来的block
func (client *Client) readCommits(commitC <-chan *BlockInfo, errorC <-chan error) {
	var data *BlockInfo
	var ok bool
	for {
		select {
		case data, ok = <-commitC:
			if !ok || data == nil {
				continue
			}
			rlog.Info("Commit blockInfo", "height", data.Height, "blockhash", data.Hash)
			client.SetCurrentInfo(data)

		case err, ok := <-errorC:
			if ok {
				panic(err)
			}
		case <-client.ctx.Done():
			return
		}
	}
}

//轮询任务，去检测本机器是否为validator节点，如果是，则执行打包任务
func (client *Client) pollingTask() {
	for {
		select {
		case <-client.ctx.Done():
			return
		case value, ok := <-client.validatorC:
			if ok && !value {
				rlog.Debug("================I'm not the validator node=============")
				leader := mux.Load().(bool)
				if leader {
					isLeader = false
					mux.Store(isLeader)
				}
			} else if ok && !mux.Load().(bool) && value {
				isLeader = true
				mux.Store(isLeader)
				go client.CreateBlock()
			} else if !ok {
				break
			}
		}
	}
}

//比较newBlock是不是最优区块
func (client *Client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	return false
}

// BlockInfo struct
type BlockInfo struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
}

// SetCurrentInfo ...
func (client *Client) SetCurrentInfo(info *BlockInfo) {
	client.mtx.Lock()
	defer client.mtx.Unlock()
	client.blockInfo = info
}

// GetCurrentInfo ...
func (client *Client) GetCurrentInfo() *BlockInfo {
	client.mtx.Lock()
	defer client.mtx.Unlock()
	return client.blockInfo
}

// GetCurrentInfoHeight ...
func (client *Client) GetCurrentInfoHeight() int64 {
	client.mtx.Lock()
	defer client.mtx.Unlock()
	return client.blockInfo.Height
}

// CheckBlockInfo check corresponding block
func (client *Client) CheckBlockInfo(info *BlockInfo) bool {
	retry := 0
	factor := 1
	for {
		lastBlock, err := client.RequestLastBlock()
		if err == nil && lastBlock.Height >= info.Height {
			break
		}
		retry++
		time.Sleep(500 * time.Millisecond)
		if retry >= 30*factor {
			rlog.Info(fmt.Sprintf("CheckBlockInfo wait %d seconds", retry/2), "height", info.Height)
			factor = factor * 2
		}
	}
	block, err := client.RequestBlock(info.Height)
	if err != nil {
		rlog.Error("CheckBlockInfo RequestBlock fail", "err", err)
		return false
	}
	cfg := client.GetAPI().GetConfig()
	if common.ToHex(block.Hash(cfg)) != info.Hash {
		rlog.Error("CheckBlockInfo hash not equal", "blockHash", common.ToHex(block.Hash(cfg)),
			"infoHash", info.Hash)
		return false
	}
	return true
}

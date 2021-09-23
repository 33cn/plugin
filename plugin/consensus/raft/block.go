// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package raft

import (
	"context"
	"time"

	"github.com/33cn/chain33/common"
	"github.com/33cn/chain33/common/merkle"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	cty "github.com/33cn/chain33/system/dapp/coins/types"
	"github.com/33cn/chain33/types"
	"github.com/coreos/etcd/snap"
	"github.com/golang/protobuf/proto"
)

func init() {
	drivers.Reg("raft", NewRaftCluster)
	drivers.QueryData.Register("raft", &Client{})
}

// Client Raft implementation
type Client struct {
	*drivers.BaseClient
	proposeC    chan<- *types.Block
	commitC     <-chan *types.Block
	errorC      <-chan error
	snapshotter *snap.Snapshotter
	validatorC  <-chan bool
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBlockstore create Raft Client
func NewBlockstore(ctx context.Context, cfg *types.Consensus, snapshotter *snap.Snapshotter, proposeC chan<- *types.Block, commitC <-chan *types.Block, errorC <-chan error, validatorC <-chan bool, cancel context.CancelFunc) *Client {
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
	return proto.Marshal(client.GetCurrentBlock())
}

func (client *Client) recoverFromSnapshot(snapshot []byte) error {
	block := &types.Block{}
	if err := proto.Unmarshal(snapshot, block); err != nil {
		return err
	}
	return nil
}

// SetQueueClient method
func (client *Client) SetQueueClient(c queue.Client) {
	rlog.Info("Enter SetQueue method of raft consensus")
	client.InitClient(c, func() {
		client.InitBlock()
	})
	go client.EventLoop()
	if !client.IsMining() {
		rlog.Info("enter sync mode")
		return
	}
	go client.readCommits(client.commitC, client.errorC)
	go client.pollingTask()
}

// Close method
func (client *Client) Close() {
	if client.cancel != nil {
		client.cancel()
	}
	rlog.Info("consensus raft closed")
}

// CreateBlock method
func (client *Client) CreateBlock() {
	//打包区块前先同步到最大高度
	tocker := time.NewTicker(30 * time.Second)
	beg := time.Now()
OuterLoop:
	for {
		select {
		case <-tocker.C:
			rlog.Info("Still catching up max height......", "Height", client.GetCurrentHeight(), "cost", time.Since(beg))
		default:
			if client.IsCaughtUp() {
				rlog.Info("Leader has caught up max height")
				break OuterLoop
			}
			time.Sleep(time.Second)
		}
	}
	tocker.Stop()

	count := int64(0)
	cfg := client.GetAPI().GetConfig()
	hint := time.NewTicker(30 * time.Second)
	ticker := time.NewTicker(time.Duration(writeBlockSeconds) * time.Second)
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
				rlog.Warn("Not Leader node anymore")
				return
			}

			lastBlock, err := client.RequestLastBlock()
			if err != nil {
				rlog.Error("Leader RequestLastBlock fail", "err", err)
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
			pblock := newblock.Clone()
			client.propose(pblock)
			rlog.Info("Leader propose block", "height", pblock.Height, "blockhash", common.ToHex(pblock.Hash(cfg)),
				"txhash", common.ToHex(pblock.TxHash))
			count = 0
		}

	}
}

// 向raft底层发送BlockInfo
func (client *Client) propose(block *types.Block) {
	client.proposeC <- block
}

// 从receive channel中读leader发来的block
func (client *Client) readCommits(commitC <-chan *types.Block, errorC <-chan error) {
	var data *types.Block
	var ok bool
	cfg := client.GetAPI().GetConfig()
	for {
		select {
		case data, ok = <-commitC:
			if !ok || data == nil {
				break
			}
			lastBlock, err := client.RequestLastBlock()
			if err != nil {
				rlog.Error("RequestLastBlock fail", "err", err)
				break
			}
			if lastBlock.Height >= data.Height {
				rlog.Info("already has block", "height", data.Height)
				break
			}
			rlog.Info("Write block", "height", data.Height, "blockhash", common.ToHex(data.Hash(cfg)),
				"txhash", common.ToHex(data.TxHash))
			err = client.WriteBlock(nil, data)
			if err != nil {
				rlog.Error("WriteBlock fail", "err", err)
				break
			}
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
				leader := mux.Load().(bool)
				if leader {
					rlog.Info("================Change to follower node=============")
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

//CmpBestBlock 比较newBlock是不是最优区块
func (client *Client) CmpBestBlock(newBlock *types.Block, cmpBlock *types.Block) bool {
	return false
}

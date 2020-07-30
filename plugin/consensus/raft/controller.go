// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package raft

import (
	"context"
	"strings"
	"sync/atomic"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	drivers "github.com/33cn/chain33/system/consensus"
	"github.com/33cn/chain33/types"
	"github.com/coreos/etcd/raft/raftpb"
)

var (
	rlog                    = log.New("module", "raft")
	genesis                 string
	genesisBlockTime        int64
	defaultSnapCount        uint64 = 10000
	snapshotCatchUpEntriesN uint64 = 10000
	writeBlockSeconds       int64  = 1
	heartbeatTick                  = 1
	emptyBlockInterval      int64
	isLeader                = false
	mux                     atomic.Value
	confChangeC             chan raftpb.ConfChange
)

type subConfig struct {
	Genesis            string `json:"genesis"`
	GenesisBlockTime   int64  `json:"genesisBlockTime"`
	NodeID             int64  `json:"nodeID"`
	PeersURL           string `json:"peersURL"`
	RaftAPIPort        int64  `json:"raftAPIPort"`
	IsNewJoinNode      bool   `json:"isNewJoinNode"`
	ReadOnlyPeersURL   string `json:"readOnlyPeersURL"`
	AddPeersURL        string `json:"addPeersURL"`
	DefaultSnapCount   int64  `json:"defaultSnapCount"`
	WriteBlockSeconds  int64  `json:"writeBlockSeconds"`
	HeartbeatTick      int32  `json:"heartbeatTick"`
	EmptyBlockInterval int64  `json:"emptyBlockInterval"`
}

func init() {
	mux.Store(isLeader)
}

// NewRaftCluster create raft cluster
func NewRaftCluster(cfg *types.Consensus, sub []byte) queue.Module {
	genesis = cfg.Genesis
	genesisBlockTime = cfg.GenesisBlockTime
	if !cfg.Minerstart {
		rlog.Info("node only sync block")
		c := drivers.NewBaseClient(cfg)
		client := &Client{BaseClient: c}
		c.SetChild(client)
		return client
	}
	rlog.Info("Start to create raft cluster")
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
	if int(subcfg.NodeID) == 0 || strings.Compare(subcfg.PeersURL, "") == 0 {
		rlog.Error("Please check whether the configuration of nodeId and peersURL is empty!")
		//TODO 当传入的参数异常时，返回给主函数的是个nil,这时候需要做异常处理
		return nil
	}
	// 默认10000个Entry打一个snapshot
	if subcfg.DefaultSnapCount > 0 {
		defaultSnapCount = uint64(subcfg.DefaultSnapCount)
		snapshotCatchUpEntriesN = uint64(subcfg.DefaultSnapCount)
	}
	// write block interval in second
	if subcfg.WriteBlockSeconds > 0 {
		writeBlockSeconds = subcfg.WriteBlockSeconds
	}
	// raft leader sends heartbeat messages every HeartbeatTick ticks
	if subcfg.HeartbeatTick > 0 {
		heartbeatTick = int(subcfg.HeartbeatTick)
	}
	// write empty block interval in second
	if subcfg.EmptyBlockInterval > 0 {
		emptyBlockInterval = subcfg.EmptyBlockInterval
	}

	var b *Client
	getSnapshot := func() ([]byte, error) { return b.getSnapshot() }
	// raft集群的建立,1. 初始化两条channel： propose channel用于客户端和raft底层交互, commit channel用于获取commit消息
	// 2. raft集群中的节点之间建立http连接
	peers := strings.Split(subcfg.PeersURL, ",")
	if len(peers) == 1 && peers[0] == "" {
		peers = []string{}
	}
	readOnlyPeers := strings.Split(subcfg.ReadOnlyPeersURL, ",")
	if len(readOnlyPeers) == 1 && readOnlyPeers[0] == "" {
		readOnlyPeers = []string{}
	}
	addPeers := strings.Split(subcfg.AddPeersURL, ",")
	if len(addPeers) == 1 && addPeers[0] == "" {
		addPeers = []string{}
	}
	//采用context来统一管理所有服务
	ctx, stop := context.WithCancel(context.Background())
	// propose channel
	proposeC := make(chan *types.Block)
	confChangeC = make(chan raftpb.ConfChange)
	commitC, errorC, snapshotterReady, validatorC := NewRaftNode(ctx, int(subcfg.NodeID), subcfg.IsNewJoinNode, peers, readOnlyPeers, addPeers, getSnapshot, proposeC, confChangeC)
	//启动raft删除节点操作监听
	go serveHTTPRaftAPI(ctx, int(subcfg.RaftAPIPort), confChangeC, errorC)
	// 监听commit channel,取block
	b = NewBlockstore(ctx, cfg, <-snapshotterReady, proposeC, commitC, errorC, validatorC, stop)
	return b
}

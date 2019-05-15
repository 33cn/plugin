// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import (
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	"github.com/33cn/chain33/types"
)

var (
	plog                    = log.New("module", "Pbft")
	genesis                 string
	genesisBlockTime        int64
)

type subConfig struct {
	Genesis          string `json:"genesis"`
	GenesisBlockTime int64  `json:"genesisBlockTime"`
	IsNode           bool   `json:"isNode"`        // 是否参与共识
	NodeID           uint64 `json:"nodeID"`        // 作为共识节点的ID
	PeersURL         string `json:"peersURL"`      // 所有共识节点的IP
	ClientID         uint64 `json:"clientID"`      // 作为客户端的ID
	ClientAddr       string `json:"clientAddr"`    // 所有客户端的IP
	PrimaryID        uint64 `json:"primaryID"`     //主节点ID
	F                uint64 `json:"f"`             // 网络最大容错数
	N                uint64 `json:"N"`             // 网络最多节点数
	K                uint64 `json:"K"`             // 日志周期
	LogMultiplier    uint64 `json:"logMultiplier"` // 常数因子
	Byzantine        bool   `json:"byzantine"`     // 节点是否拜占庭
}

// NewPbftNode 产生一个新的PBFT节点
func NewPbftNode(cfg *types.Consensus, sub []byte) queue.Module {
	plog.Info("Start to creat PBFT node")
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
	if strings.Compare(subcfg.PeersURL, "") == 0 {
		plog.Error("Please check whether the configuration of PeersURL is empty!")
		return nil
	}
	peers := subcfg.PeersURL

	if subcfg.NodeID <= 0 {
		plog.Error("Please check whether the configuration of NodeID is right")
		if subcfg.PrimaryID <= 0 || subcfg.F < 0 || subcfg.N <= 0 || subcfg.K <= 0 || subcfg.LogMultiplier <= 1 {
			plog.Error("Please check whether the ID and paras is below 0 or wrong")
			return nil
		}
	}

	var c *Client
	requestChan, dataChan, isClient, address := NewReplica(subcfg.NodeID, peers, subcfg.PrimaryID, subcfg.F, subcfg.N, subcfg.K, subcfg.LogMultiplier, subcfg.Byzantine)
	c = NewBlockstore(cfg, requestChan, dataChan, isClient, address)
	return c
}

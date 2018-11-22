// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pbft

import (
	"strings"

	log "github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/chain33/queue"
	pb "github.com/33cn/chain33/types"
)

var (
	plog             = log.New("module", "Pbft")
	genesis          string
	genesisBlockTime int64
	clientAddr       string
)

type subConfig struct {
	Genesis          string `json:"genesis"`
	GenesisBlockTime int64  `json:"genesisBlockTime"`
	NodeID           int64  `json:"nodeID"`
	PeersURL         string `json:"peersURL"`
	ClientAddr       string `json:"clientAddr"`
}

// NewPbft create pbft cluster
func NewPbft(cfg *pb.Consensus, sub []byte) queue.Module {
	plog.Info("start to creat pbft node")
	var subcfg subConfig
	if sub != nil {
		pb.MustDecode(sub, &subcfg)
	}

	if subcfg.Genesis != "" {
		genesis = subcfg.Genesis
	}
	if subcfg.GenesisBlockTime > 0 {
		genesisBlockTime = subcfg.GenesisBlockTime
	}
	if int(subcfg.NodeID) == 0 || strings.Compare(subcfg.PeersURL, "") == 0 || strings.Compare(subcfg.ClientAddr, "") == 0 {
		plog.Error("The nodeId, peersURL or clientAddr is empty!")
		return nil
	}
	clientAddr = subcfg.ClientAddr

	var c *Client
	replyChan, requestChan, isPrimary := NewReplica(uint32(subcfg.NodeID), subcfg.PeersURL, subcfg.ClientAddr)
	c = NewBlockstore(cfg, replyChan, requestChan, isPrimary)
	return c
}

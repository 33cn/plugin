// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package neutrino integrate btc light client neutrino
package neutrino

import (
	"context"
	"encoding/json"
	"github.com/33cn/chain33/types"
	"github.com/lightninglabs/neutrino/headerfs"
	"sync"
	"time"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lightninglabs/neutrino"
)

var log = log15.New("module", "neutrino")

var _ lightclient.Lighter = &neutrinoClient{}

func init() {

	lightclient.Register("neutrino", newClient)
}

func newClient() lightclient.Lighter {
	return &neutrinoClient{}
}

type neutrinoClient struct {
	ctx            context.Context
	chain33Api     client.QueueProtocolAPI
	cfg            config
	commitAddr     string
	neutrinoCfg    neutrino.Config
	neutrinoCS     *neutrino.ChainService
	bestBlock      *headerfs.BlockStamp
	lock           sync.Mutex
	chain33FeeRate int64
}

// Init init client context
func (n *neutrinoClient) Init(ctx context.Context, api client.QueueProtocolAPI, cfg *lightclient.Config) error {

	n.ctx = ctx
	n.chain33Api = api
	n.chain33FeeRate = 100000

	subCfg, _ := json.Marshal(cfg.Neutrino)
	types.MustDecode(subCfg, &n.cfg)

	if n.cfg.BlockCacheSize < 1024*1024 {
		n.cfg.BlockCacheSize = defalutBlockCacheSize
	}

	if n.cfg.MaxPeer < 1 {
		n.cfg.MaxPeer = 8
	}

	neutrinoCfg, err := initNeutrinoConfig(api, n.cfg)
	if err != nil {
		log.Error("Init", "initNeutrinoConfig error", err)
		return err
	}
	n.neutrinoCfg = neutrinoCfg

	cs, err := neutrino.NewChainService(neutrinoCfg)
	if err != nil {
		log.Error("Init", "NewChainService error", err)
		_ = neutrinoCfg.Database.Close()
		return err
	}
	n.neutrinoCS = cs
	return nil

}

// Start starting routine
func (n *neutrinoClient) Start() {

	if err := n.neutrinoCS.Start(); err != nil {
		log.Error("Start", "neutrinoCS start error", err)
		_ = n.neutrinoCfg.Database.Close()
		return
	}

	newRGBX().init(n)
	go n.cleanUp()
}

func (n *neutrinoClient) cleanUp() {

	for {

		select {

		case <-n.ctx.Done():
			if err := n.neutrinoCS.Stop(); err != nil {
				log.Error("cleanUp Unable to stop neutrino server", "err", err)
			}
			if err := n.neutrinoCfg.Database.Close(); err != nil {
				log.Error("cleanUp Unable to close neutrino db", "err", err)
			}
		}
	}

}

func (n *neutrinoClient) handleBestBlock() {

	ticker := time.NewTicker(time.Second * 30)
	for {

		select {

		case <-n.ctx.Done():
			return
		case <-ticker.C:

			blk, err := n.neutrinoCS.BestBlock()
			if err != nil {
				log.Error("handleBestBlock", "err", err)
				continue
			}
			n.setBestBlock(blk)
		}
	}
}

func (n *neutrinoClient) getBestBlock() *headerfs.BlockStamp {
	n.lock.Lock()
	defer n.lock.Unlock()
	return n.bestBlock
}

func (n *neutrinoClient) setBestBlock(blk *headerfs.BlockStamp) {
	n.lock.Lock()
	defer n.lock.Unlock()
	if blk != nil {
		n.bestBlock = blk
	}
}

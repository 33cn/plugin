// Copyright Fuzamei Corp. 2018 All Rights Reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package neutrino integrate btc light client neutrino
package neutrino

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/33cn/chain33/types"
	"github.com/lightninglabs/neutrino/headerfs"

	"github.com/33cn/chain33/client"
	"github.com/33cn/chain33/common/log/log15"
	"github.com/33cn/plugin/plugin/dapp/lightclient/rpc/lightclient"
	_ "github.com/btcsuite/btcwallet/walletdb/bdb"
	"github.com/lightninglabs/neutrino"
)

var log = log15.New("module", "lightclient.neutrino")

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
	bw             *btcWallet
	bestBlock      *headerfs.BlockStamp
	lock           sync.Mutex
	chain33FeeRate int64
}

// Init init client context
func (n *neutrinoClient) Init(ctx context.Context, api client.QueueProtocolAPI, cfg *lightclient.Config) error {

	n.ctx = ctx
	n.chain33Api = api
	n.chain33FeeRate = 100000
	n.commitAddr = cfg.CommitAddr
	subCfg, _ := json.Marshal(cfg.Neutrino)
	types.MustDecode(subCfg, &n.cfg)

	err := n.initNeutrinoConfig(api)
	if err != nil {
		log.Error("Init", "initNeutrinoConfig error", err)
		return err
	}

	cs, err := neutrino.NewChainService(n.neutrinoCfg)
	if err != nil {
		log.Error("Init", "NewChainService error", err)
		_ = n.neutrinoCfg.Database.Close()
		return err
	}
	n.neutrinoCS = cs
	bw, err := newBtcWallet(n)
	if err != nil {
		log.Error("Init", "newBtcWallet error", err)
		return err
	}
	n.bw = bw
	return nil

}

// Start starting routine
func (n *neutrinoClient) Start() {

	if !n.cfg.IsOfficialNode {
		return
	}
	if err := n.neutrinoCS.Start(); err != nil {
		log.Error("Start", "neutrinoCS start error", err)
		_ = n.neutrinoCfg.Database.Close()
		panic(err)
	}
	if err := n.bw.start(); err != nil {
		log.Error("Start", "btcwallet start error", err)
		n.bw.stop()
		panic(err)
	}

	go n.handleBestBlock()
	go n.cleanUp()
	newRGBX().init(n)
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
			n.bw.stop()
		}
	}

}

func (n *neutrinoClient) handleBestBlock() {

	interval := time.Duration(n.cfg.BtcBlockInterval) / 3
	ticker := time.NewTicker(time.Second * interval)
	for {

		select {

		case <-n.ctx.Done():
			return
		case <-ticker.C:

			bestBlock := n.getBestBlock()
			blk, err := n.neutrinoCS.BestBlock()
			if err != nil {
				log.Error("handleBestBlock", "err", err)
				continue
			}
			if bestBlock == nil || bestBlock.Height < blk.Height {
				log.Debug("handleBestBlock", "height", blk.Height, "hash", blk.Hash.String())
				n.setBestBlock(blk)
			}
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
